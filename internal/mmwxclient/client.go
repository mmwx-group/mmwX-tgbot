// Package mmwxclient HTTP 客户端,调主控 /api/admin/tgbot/* 端点。
//
// 鉴权:Authorization: Bearer <MMWX_API_TOKEN>(admin 级 token,bot 内部按 tg_id 分发权限)。
// 所有错误透传(网络/状态码/JSON 解析失败均返回 wrapped error)。
package mmwxclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	apiToken   string
	httpClient *http.Client
}

// New 构造 client。base 形如 "https://mmw.domain.com",token 是 admin api token。
func New(base, token string, timeoutSeconds int) *Client {
	if timeoutSeconds <= 0 {
		timeoutSeconds = 8
	}
	return &Client{
		baseURL:  base,
		apiToken: token,
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
		},
	}
}

// doRequest 发起请求并把 response body 反序列化到 out(nil 则忽略)。
//
// 行为:
//   - 网络/解析错误:返回 wrapped error
//   - HTTP 4xx/5xx:尝试解出 {"error": "..."},否则用原始 body
//   - 2xx 但 body 含 {"error": ...}:也视为错误
func (c *Client) doRequest(ctx context.Context, method, path string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal body: %w", err)
		}
		reader = bytes.NewReader(buf)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("User-Agent", "mmwX-tgbot/0.1")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http do: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errEnv struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		if jerr := json.Unmarshal(bodyBytes, &errEnv); jerr == nil {
			if errEnv.Error != "" {
				return fmt.Errorf("HTTP %d: %s", resp.StatusCode, errEnv.Error)
			}
			if errEnv.Message != "" {
				return fmt.Errorf("HTTP %d: %s", resp.StatusCode, errEnv.Message)
			}
		}
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(bodyBytes), 256))
	}

	if out != nil {
		if err := json.Unmarshal(bodyBytes, out); err != nil {
			return fmt.Errorf("unmarshal response: %w (body: %s)", err, truncate(string(bodyBytes), 256))
		}
	}
	return nil
}

func (c *Client) get(ctx context.Context, path string, q url.Values, out any) error {
	if len(q) > 0 {
		path = path + "?" + q.Encode()
	}
	return c.doRequest(ctx, http.MethodGet, path, nil, out)
}

func (c *Client) post(ctx context.Context, path string, body, out any) error {
	return c.doRequest(ctx, http.MethodPost, path, body, out)
}

// SubscriptionUsage is the aggregate usage advertised by an MMWX
// subscription through the standard Subscription-Userinfo response header.
type SubscriptionUsage struct {
	Upload   int64
	Download int64
	Total    int64
}

func (u SubscriptionUsage) Used() int64 {
	return u.Upload + u.Download
}

func (u SubscriptionUsage) Remaining() int64 {
	remaining := u.Total - u.Used()
	if remaining < 0 {
		return 0
	}
	return remaining
}

// GetSubscriptionUsage reads aggregate quota/usage without parsing the YAML
// subscription body. This is used for the administrator global traffic view.
func (c *Client) GetSubscriptionUsage(ctx context.Context, combinedCode string) (*SubscriptionUsage, error) {
	combinedCode = strings.TrimSpace(combinedCode)
	if combinedCode == "" {
		return nil, fmt.Errorf("subscription code is empty")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.baseURL+"/x/"+url.PathEscape(combinedCode), nil)
	if err != nil {
		return nil, fmt.Errorf("new subscription request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("User-Agent", "mmwX-tgbot/0.1")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("subscription request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("subscription HTTP %d", resp.StatusCode)
	}

	raw := resp.Header.Get("Subscription-Userinfo")
	if raw == "" {
		return nil, fmt.Errorf("Subscription-Userinfo header is missing")
	}
	values := map[string]int64{}
	for _, field := range strings.Split(raw, ";") {
		parts := strings.SplitN(strings.TrimSpace(field), "=", 2)
		if len(parts) != 2 {
			continue
		}
		n, parseErr := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
		if parseErr != nil || n < 0 {
			continue
		}
		values[strings.ToLower(strings.TrimSpace(parts[0]))] = n
	}
	usage := &SubscriptionUsage{
		Upload:   values["upload"],
		Download: values["download"],
		Total:    values["total"],
	}
	if usage.Total <= 0 {
		return nil, fmt.Errorf("invalid subscription total quota")
	}
	return usage, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
