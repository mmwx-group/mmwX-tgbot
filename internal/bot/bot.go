// Package bot 命令路由 + 处理。HTTP 调用主控走 mmwxclient。
package bot

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/iluobei/mmwX-tgbot/internal/config"
	"github.com/iluobei/mmwX-tgbot/internal/mmwxclient"
)

type Service struct {
	mu     sync.Mutex
	cfg    config.Config
	client *mmwxclient.Client
	b      *bot.Bot
	cancel context.CancelFunc
	webSrv *http.Server
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

	s.setMyCommands(ctx, b)
	s.setMenuButton(ctx, b)
	s.startWebApp(ctx)

	go func() {
		log.Printf("[mmwX-tgbot] long-poll started")
		b.Start(ctx)
		log.Printf("[mmwX-tgbot] long-poll stopped")
	}()
	go s.runDailyNotifier(ctx, b)
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
