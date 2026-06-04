// mmwX-tgbot 入口。
//
// 用法:
//
//	mmwx-tgbot               # 加载 ./config.yaml(默认)
//	mmwx-tgbot -c /etc/mmwx-tgbot/config.yaml
//
// 也可全部用环境变量,见 config.example.yaml 注释。
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mmwx-group/mmwX-tgbot/internal/bot"
	"github.com/mmwx-group/mmwX-tgbot/internal/config"
	"github.com/mmwx-group/mmwX-tgbot/internal/mmwxclient"
)

var version = "dev"

func main() {
	configPath := flag.String("c", "config.yaml", "config 文件路径")
	showVersion := flag.Bool("v", false, "显示版本号")
	flag.Parse()

	if *showVersion {
		log.Printf("mmwX-tgbot %s", version)
		return
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	client := mmwxclient.New(cfg.MMWXURL, cfg.MMWXAPIToken, cfg.HTTPTimeoutSeconds)
	service := bot.New(cfg, client)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := service.Start(ctx); err != nil {
		log.Fatalf("bot 启动失败: %v", err)
	}
	log.Printf("mmwX-tgbot %s started (mmwx_url=%s, admins=%d)",
		version, cfg.MMWXURL, len(cfg.AdminTGIDs))

	<-ctx.Done()
	log.Printf("收到信号,正在关闭...")
	service.Stop()
	os.Exit(0)
}
