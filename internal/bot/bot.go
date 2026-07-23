// Package bot 命令路由 + 处理。HTTP 调用主控走 mmwxclient。
package bot

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/mmwx-group/mmwX-tgbot/internal/config"
	"github.com/mmwx-group/mmwX-tgbot/internal/mmwxclient"
)

type Service struct {
	mu          sync.Mutex
	cfg         config.Config
	client      *mmwxclient.Client
	b           *bot.Bot
	cancel      context.CancelFunc
	webSrv      *http.Server
	botUsername string // 由 getMe 获取,用于兑换码文案的 {机器人地址} = https://t.me/<botUsername>

	// Mini App 主题跟随主控「默认主题」的 60s 缓存(见 cachedDefaultTheme)
	themeMu  sync.Mutex
	themeVal string
	themeExp time.Time
}

// botURL 返回机器人的 t.me 链接;拿不到 username 时返回空串。
func (s *Service) botURL() string {
	if s.botUsername == "" {
		return ""
	}
	return "https://t.me/" + s.botUsername
}

func New(cfg config.Config, client *mmwxclient.Client) *Service {
	return &Service{cfg: cfg, client: client}
}

func (s *Service) Start(parent context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, err := bot.New(s.cfg.TGBotToken, bot.WithDefaultHandler(s.defaultHandler))
	if err != nil {
		return err
	}
	registerCommands(b, s)

	ctx, cancel := context.WithCancel(parent)
	s.b = b
	s.cancel = cancel

	// 记录 bot 自己的 @username(用于兑换码文案的 {机器人地址})。失败不致命。
	if me, merr := b.GetMe(ctx); merr == nil && me != nil {
		s.botUsername = me.Username
	} else if merr != nil {
		log.Printf("[mmwX-tgbot] getMe failed (bot url placeholder will be empty): %v", merr)
	}

	s.setMyCommands(ctx, b)
	s.setMenuButton(ctx, b)
	s.startWebApp(ctx)

	go func() {
		log.Printf("[mmwX-tgbot] long-poll started")
		b.Start(ctx)
		log.Printf("[mmwX-tgbot] long-poll stopped")
	}()
	go s.runDailyNotifier(ctx, b)
	go s.runAnnouncementBroadcaster(ctx, b)
	return nil
}

// setMenuButton 把 Mini App 设为聊天菜单按钮(配了 webapp_url 才设)。
func (s *Service) setMenuButton(ctx context.Context, b *bot.Bot) {
	if s.cfg.WebAppURL == "" {
		return
	}
	_, _ = b.SetChatMenuButton(ctx, &bot.SetChatMenuButtonParams{
		MenuButton: &models.MenuButtonWebApp{
			Type:   models.MenuButtonTypeWebApp,
			Text:   "📊 我的面板",
			WebApp: models.WebAppInfo{URL: s.cfg.WebAppURL},
		},
	})
}

func (s *Service) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	s.stopWebApp()
	s.b = nil
}
