package bot

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mmwx-group/mmwX-tgbot/internal/mmwxclient"
)

// Telegram Mini App 后端。bot 起一个 HTTP 服务(默认 :8088,前置 nginx 反代到公网 HTTPS):
//   GET /app              → 返回单页前端
//   GET /api/tg-webapp/me → 校验 Telegram initData(用 bot 自己的 token)→ 反查账号 → 聚合返回
// 免登录:initData 由 Telegram 用 bot token 签名,校验通过即可信任其中的 telegram_id。

// startWebApp 启动 Mini App HTTP 服务。listen 为空则不启动。
func (s *Service) startWebApp(ctx context.Context) {
	if strings.TrimSpace(s.cfg.WebAppListen) == "" {
		return
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/app", s.webAppPage)
	mux.HandleFunc("/api/tg-webapp/logo-light", s.webAppLogoLight)
	mux.HandleFunc("/api/tg-webapp/logo-dark", s.webAppLogoDark)
	// API 端点统一 per-IP 限流(60 次/分)。
	mux.HandleFunc("/api/tg-webapp/me", webRL(s.webAppMe))
	mux.HandleFunc("/api/tg-webapp/register", webRL(s.webAppRegister))
	mux.HandleFunc("/api/tg-webapp/redeem", webRL(s.webAppRedeem))
	mux.HandleFunc("/api/tg-webapp/admin/invites", webRL(s.webAppAdminInvites))
	mux.HandleFunc("/api/tg-webapp/admin/invite-create", webRL(s.webAppAdminInviteCreate))
	mux.HandleFunc("/api/tg-webapp/admin/invite-revoke", webRL(s.webAppAdminInviteRevoke))
	mux.HandleFunc("/api/tg-webapp/admin/invite-delete", webRL(s.webAppAdminInviteDelete))

	srv := &http.Server{Addr: s.cfg.WebAppListen, Handler: mux}
	s.webSrv = srv
	go func() {
		log.Printf("[mmwX-tgbot] mini app HTTP listening on %s", s.cfg.WebAppListen)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[mmwX-tgbot] mini app server error: %v", err)
		}
	}()
}

func (s *Service) stopWebApp() {
	if s.webSrv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = s.webSrv.Shutdown(ctx)
		s.webSrv = nil
	}
}

// validateInitData 按 Telegram 规范校验 initData,返回可信的 telegram_id 和 @handle。
// secret = HMAC_SHA256(key="WebAppData", msg=botToken);hash = HMAC_SHA256(key=secret, msg=data_check_string)。
func validateInitData(initData, botToken string) (int64, string, error) {
	if initData == "" || botToken == "" {
		return 0, "", errors.New("empty init data or token")
	}
	parsed, err := url.ParseQuery(initData)
	if err != nil {
		return 0, "", err
	}
	hash := parsed.Get("hash")
	if hash == "" {
		return 0, "", errors.New("missing hash")
	}

	pairs := make([]string, 0, len(parsed))
	for k, vs := range parsed {
		if k == "hash" {
			continue
		}
		pairs = append(pairs, k+"="+vs[0])
	}
	sort.Strings(pairs)
	dcs := strings.Join(pairs, "\n")

	secret := hmacSum([]byte("WebAppData"), []byte(botToken))
	calc := hex.EncodeToString(hmacSum(secret, []byte(dcs)))
	if !hmac.Equal([]byte(calc), []byte(hash)) {
		return 0, "", errors.New("bad signature")
	}

	// 时效:拒绝超过 24h 的 initData,防重放
	if ad := parsed.Get("auth_date"); ad != "" {
		if sec, e := strconv.ParseInt(ad, 10, 64); e == nil {
			if time.Since(time.Unix(sec, 0)) > 24*time.Hour {
				return 0, "", errors.New("init data expired")
			}
		}
	}

	var u struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
	}
	if e := json.Unmarshal([]byte(parsed.Get("user")), &u); e != nil || u.ID == 0 {
		return 0, "", errors.New("no user id")
	}
	return u.ID, u.Username, nil
}

func hmacSum(key, msg []byte) []byte {
	m := hmac.New(sha256.New, key)
	m.Write(msg)
	return m.Sum(nil)
}

// webAppMe 校验 initData → 反查账号 → 聚合账号/流量/节点/订阅。
func (s *Service) webAppMe(w http.ResponseWriter, r *http.Request) {
	initData := r.Header.Get("X-Telegram-Init-Data")
	if initData == "" {
		initData = r.URL.Query().Get("initData")
	}
	tgID, handle, err := validateInitData(initData, s.cfg.TGBotToken)
	if err != nil {
		writeJSONResp(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
		return
	}

	ctx := r.Context()
	info, err := s.client.UserByTG(ctx, tgID)
	if err != nil {
		writeJSONResp(w, http.StatusBadGateway, map[string]any{"error": "查询失败"})
		return
	}
	if !info.Bound {
		// 管理员未绑定 → 自动绑到主控管理员账号(妙妙屋X 默认单管理员),再重查。
		if s.cfg.IsAdmin(tgID) {
			if _, berr := s.client.BindAdmin(ctx, tgID, handle); berr == nil {
				if re, rerr := s.client.UserByTG(ctx, tgID); rerr == nil {
					info = re
				}
			}
		}
		if !info.Bound {
			writeJSONResp(w, http.StatusOK, map[string]any{"bound": false})
			return
		}
	}
	username := info.Username

	resp := map[string]any{"bound": true, "is_admin": s.cfg.IsAdmin(tgID)}

	// 账号 + 流量 + 套餐
	if summary, err := s.client.UserSummary(ctx, username); err == nil {
		resp["account"] = map[string]any{
			"username":  summary.Username,
			"role":      summary.Role,
			"is_active": summary.IsActive,
			"email":     summary.Email,
		}
		pkgName, limitGB := "", 0.0
		if summary.Package != nil {
			if v, ok := summary.Package["name"].(string); ok {
				pkgName = v
			}
			if v, ok := summary.Package["traffic_limit_gb"].(float64); ok {
				limitGB = v
			}
		}
		traffic := map[string]any{
			"package_name": pkgName,
			"limit_gb":     limitGB,
			"cycle_used":   summary.Traffic.CycleUplink + summary.Traffic.CycleDownlink,
			"total_up":     summary.Traffic.TotalUplink,
			"total_down":   summary.Traffic.TotalDownlink,
		}
		if summary.PackageEndDate != "" {
			traffic["end_date"] = summary.PackageEndDate
			if d, ok := daysUntil(summary.PackageEndDate); ok {
				traffic["days_left"] = d
			}
		}
		resp["traffic"] = traffic
	}

	// 套餐周期内每日用量(首页曲线)
	history := []map[string]any{}
	if hist, err := s.client.UserDailyTraffic(ctx, username); err == nil {
		for _, d := range hist {
			history = append(history, map[string]any{"date": d.Date, "used_gb": d.UsedGB})
		}
	}
	resp["history"] = history

	// 各节点已用流量(主页用)
	nodes := []map[string]any{}
	if items, err := s.client.UserNodeTraffic(ctx, username); err == nil {
		for _, n := range items {
			nodes = append(nodes, map[string]any{
				"name": n.NodeName,
				"used": n.Uplink + n.Downlink,
			})
		}
	}
	resp["nodes"] = nodes

	// 各节点在线状态(状态页用)。server_status 为空 = 外部/未托管节点,无状态 → unknown。
	nodeStatus := []map[string]any{}
	if nr, err := s.client.UserNodes(ctx, username); err == nil {
		for _, n := range nr.Nodes {
			st := "offline"
			if n.ServerOnline {
				st = "online"
			} else if strings.TrimSpace(n.ServerStatus) == "" {
				st = "unknown"
			}
			nodeStatus = append(nodeStatus, map[string]any{
				"name": n.Name, "protocol": n.Protocol, "status": st,
			})
		}
	}
	resp["node_status"] = nodeStatus

	// 订阅(默认订阅在前)
	base := strings.TrimRight(s.cfg.MMWXURL, "/") + "/x"
	subs := []map[string]any{}
	if sr, err := s.client.UserSubscriptions(ctx, username); err == nil {
		if sr.DefaultSubscription != nil {
			subs = append(subs, map[string]any{
				"name": sr.DefaultSubscription.Name, "default": true,
				"url": base + "/" + sr.DefaultSubscription.CombinedCode,
			})
		}
		for _, sf := range sr.Subscriptions {
			subs = append(subs, map[string]any{
				"name": sf.Name, "default": false,
				"url": base + "/" + sf.CombinedCode,
			})
		}
	}
	resp["subscriptions"] = subs

	// 管理员账号(无套餐、无个人订阅)→ 用系统订阅列表第一个订阅 + 其节点。
	// 订阅链接取第一个订阅;流量页/状态页用该订阅的全部节点(用量按节点名从已用流量合并,无则 0)。
	if s.cfg.IsAdmin(tgID) && len(subs) == 0 {
		if sv, err := s.client.GetAdminSubview(ctx, username); err == nil && sv != nil {
			if sv.Subscription != nil {
				resp["subscriptions"] = []map[string]any{{
					"name": sv.Subscription.Name, "default": false,
					"url": base + "/" + sv.Subscription.CombinedCode,
				}}
			}
			usedByName := map[string]int64{}
			for _, n := range nodes {
				if nm, ok := n["name"].(string); ok {
					if u, ok2 := n["used"].(int64); ok2 {
						usedByName[nm] = u
					}
				}
			}
			nn := []map[string]any{}
			ns := []map[string]any{}
			for _, sn := range sv.Nodes {
				nn = append(nn, map[string]any{"name": sn.Name, "used": usedByName[sn.Name]})
				ns = append(ns, map[string]any{"name": sn.Name, "protocol": sn.Protocol, "status": sn.Status})
			}
			resp["nodes"] = nn
			resp["node_status"] = ns
		}
	}

	writeJSONResp(w, http.StatusOK, resp)
}

// webAppRegister 未绑定用户在 Mini App 内用「邀请码+用户名+密码」注册并绑定 TG。
func (s *Service) webAppRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONResp(w, http.StatusMethodNotAllowed, map[string]any{"error": "method"})
		return
	}
	initData := r.Header.Get("X-Telegram-Init-Data")
	tgID, handle, err := validateInitData(initData, s.cfg.TGBotToken)
	if err != nil {
		writeJSONResp(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
		return
	}
	var body struct {
		Code     string `json:"code"`
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONResp(w, http.StatusBadRequest, map[string]any{"error": "invalid body"})
		return
	}

	resp, err := s.client.Bind(r.Context(), mmwxclient.BindRequest{
		Code:           strings.ToUpper(strings.TrimSpace(body.Code)),
		TelegramID:     tgID,
		TelegramHandle: handle,
		Username:       strings.TrimSpace(body.Username),
		Email:          strings.TrimSpace(body.Email),
		Password:       body.Password,
	})
	if err != nil {
		writeJSONResp(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	writeJSONResp(w, http.StatusOK, map[string]any{"success": true, "username": resp.Username})
}

// webAppRedeem 已绑定用户用兑换码续期。
func (s *Service) webAppRedeem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONResp(w, http.StatusMethodNotAllowed, map[string]any{"error": "method"})
		return
	}
	tgID, _, err := validateInitData(r.Header.Get("X-Telegram-Init-Data"), s.cfg.TGBotToken)
	if err != nil {
		writeJSONResp(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
		return
	}
	var body struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Code) == "" {
		writeJSONResp(w, http.StatusBadRequest, map[string]any{"error": "请输入兑换码"})
		return
	}
	res, err := s.client.Redeem(r.Context(), strings.ToUpper(strings.TrimSpace(body.Code)), tgID)
	if err != nil {
		writeJSONResp(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	writeJSONResp(w, http.StatusOK, map[string]any{
		"success": true, "username": res.Username, "end_date": res.EndDate, "package_name": res.PackageName,
	})
}

// adminTGID 校验 initData 且要求是管理员;失败时已写好响应并返回 ok=false。
// 授权严格服务端判定(从签名校验出的 tgID),不信任任何前端标志。
func (s *Service) adminTGID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	tgID, _, err := validateInitData(r.Header.Get("X-Telegram-Init-Data"), s.cfg.TGBotToken)
	if err != nil {
		writeJSONResp(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
		return 0, false
	}
	if !s.cfg.IsAdmin(tgID) {
		writeJSONResp(w, http.StatusForbidden, map[string]any{"error": "forbidden"})
		return 0, false
	}
	return tgID, true
}

// webAppAdminInvites GET 列邀请码 + 套餐(供生成表单选套餐)。
func (s *Service) webAppAdminInvites(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.adminTGID(w, r); !ok {
		return
	}
	ctx := r.Context()
	invites, err := s.client.ListInvites(ctx, 50)
	if err != nil {
		writeJSONResp(w, http.StatusBadGateway, map[string]any{"error": err.Error()})
		return
	}
	pkgs, _ := s.client.ListPackages(ctx)
	// 兑换码「复制文案」用:主控可配的模板 + 占位符取值。模板取失败不阻塞列表(前端退化为只复制码)。
	tpl, _ := s.client.GetRedeemTemplate(ctx)
	writeJSONResp(w, http.StatusOK, map[string]any{
		"invites":         invites,
		"packages":        pkgs,
		"redeem_template": tpl,
		"master_url":      s.cfg.MMWXURL,
		"bot_url":         s.botURL(),
	})
}

// webAppAdminInviteCreate POST 创建邀请码。
func (s *Service) webAppAdminInviteCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONResp(w, http.StatusMethodNotAllowed, map[string]any{"error": "method"})
		return
	}
	if _, ok := s.adminTGID(w, r); !ok {
		return
	}
	var body struct {
		PackageID      *int64 `json:"package_id"`
		DurationMonths int    `json:"duration_months"`
		MaxUses        int    `json:"max_uses"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONResp(w, http.StatusBadRequest, map[string]any{"error": "invalid body"})
		return
	}
	maxUses := body.MaxUses
	if maxUses < 1 {
		maxUses = 1
	}
	// 兑换码统一 kind=new;使用次数可设(>1 = 多用户共用一个码注册)。
	req := mmwxclient.CreateInviteRequest{
		Kind:           "new",
		MaxUses:        maxUses,
		PackageID:      body.PackageID,
		DurationMonths: body.DurationMonths,
	}
	code, err := s.client.CreateInvite(r.Context(), req)
	if err != nil {
		writeJSONResp(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	writeJSONResp(w, http.StatusOK, map[string]any{"success": true, "code": code})
}

// webAppAdminInviteRevoke POST 撤销邀请码。
func (s *Service) webAppAdminInviteRevoke(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONResp(w, http.StatusMethodNotAllowed, map[string]any{"error": "method"})
		return
	}
	if _, ok := s.adminTGID(w, r); !ok {
		return
	}
	var body struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Code) == "" {
		writeJSONResp(w, http.StatusBadRequest, map[string]any{"error": "code 必填"})
		return
	}
	if err := s.client.RevokeInvite(r.Context(), strings.TrimSpace(body.Code)); err != nil {
		writeJSONResp(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	writeJSONResp(w, http.StatusOK, map[string]any{"success": true})
}

// webAppAdminInviteDelete POST 硬删除邀请码(仅限已不可用的)。
func (s *Service) webAppAdminInviteDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONResp(w, http.StatusMethodNotAllowed, map[string]any{"error": "method"})
		return
	}
	if _, ok := s.adminTGID(w, r); !ok {
		return
	}
	var body struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Code) == "" {
		writeJSONResp(w, http.StatusBadRequest, map[string]any{"error": "code 必填"})
		return
	}
	if err := s.client.DeleteInvite(r.Context(), strings.TrimSpace(body.Code)); err != nil {
		writeJSONResp(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	writeJSONResp(w, http.StatusOK, map[string]any{"success": true})
}

// per-IP 固定窗口限流(60 次/分),给 Mini App API 端点用。
var (
	webRLMu  sync.Mutex
	webRLMap = map[string]*rlEntry{}
)

func webAllow(ip string) bool {
	webRLMu.Lock()
	defer webRLMu.Unlock()
	now := time.Now()
	e := webRLMap[ip]
	if e == nil || now.Sub(e.windowStart) >= time.Minute {
		webRLMap[ip] = &rlEntry{count: 1, windowStart: now}
		return true
	}
	e.count++
	return e.count <= 60
}

func clientIP(r *http.Request) string {
	if v := strings.TrimSpace(r.Header.Get("X-Real-IP")); v != "" {
		return v
	}
	if v := r.Header.Get("X-Forwarded-For"); v != "" {
		return strings.TrimSpace(strings.SplitN(v, ",", 2)[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func webRL(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !webAllow(clientIP(r)) {
			writeJSONResp(w, http.StatusTooManyRequests, map[string]any{"error": "请求过于频繁,请稍后再试"})
			return
		}
		h(w, r)
	}
}

func writeJSONResp(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
