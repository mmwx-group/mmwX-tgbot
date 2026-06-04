// Package config 加载 mmwX-tgbot 配置。
//
// 加载顺序:
//  1. CLI -c <path> 指定的 yaml 文件(默认 ./config.yaml)
//  2. 环境变量覆盖(yaml 字段名转大写 + 替换 . 为 _,加 MMWX_TGBOT_ 前缀)
//
// 示例:
//
//	mmwx_url: https://mmw.domain.com
//	mmwx_api_token: bfd6ac59-...
//	tg_bot_token: 1234:xxx
//	admin_tg_ids: [12345, 67890]
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	// 主控 mmwx 地址(http(s)://host[:port],bot 用 admin token 调 /api/admin/tgbot/*)
	MMWXURL string `yaml:"mmwx_url"`
	// admin 级 API token,从主控系统设置或 system_config.api_token 取
	MMWXAPIToken string `yaml:"mmwx_api_token"`
	// TG bot token(@BotFather)
	TGBotToken string `yaml:"tg_bot_token"`
	// 谁能用 /admin_* 命令(独立 mmwx 角色,避免每次反查 db)
	AdminTGIDs []int64 `yaml:"admin_tg_ids"`
	// 调主控的 HTTP 超时(秒),默认 8
	HTTPTimeoutSeconds int `yaml:"http_timeout_seconds"`
	// Mini App HTTP 服务监听地址,默认 127.0.0.1:23088(只听回环,由前置 nginx 反代到公网 HTTPS)
	WebAppListen string `yaml:"webapp_listen"`
	// Mini App 公网 HTTPS 地址(nginx 暴露的,如 https://mmw-tgapp.domain.com/app)。
	// 非空时 bot 启动会把它设为 TG 菜单按钮;空则不设按钮(仅本地起服务)。
	WebAppURL string `yaml:"webapp_url"`
	// 调试:允许从 ?initData= 读取(本地浏览器预览用)。生产务必关闭(默认 false),
	// 仅 Telegram 注入的 initData 才可信;开启会让带 initData 的 URL 可被分享重放。
	WebAppDevPreview bool `yaml:"webapp_dev_preview"`
}

func Load(path string) (Config, error) {
	c := Config{HTTPTimeoutSeconds: 8, WebAppListen: "127.0.0.1:23088"}
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return c, fmt.Errorf("read config: %w", err)
		}
		if err := yaml.Unmarshal(data, &c); err != nil {
			return c, fmt.Errorf("parse yaml: %w", err)
		}
	}
	applyEnvOverrides(&c)
	if err := c.Validate(); err != nil {
		return c, err
	}
	return c, nil
}

func applyEnvOverrides(c *Config) {
	if v := os.Getenv("MMWX_TGBOT_MMWX_URL"); v != "" {
		c.MMWXURL = v
	}
	if v := os.Getenv("MMWX_TGBOT_MMWX_API_TOKEN"); v != "" {
		c.MMWXAPIToken = v
	}
	if v := os.Getenv("MMWX_TGBOT_TG_BOT_TOKEN"); v != "" {
		c.TGBotToken = v
	}
	if v := os.Getenv("MMWX_TGBOT_ADMIN_TG_IDS"); v != "" {
		c.AdminTGIDs = parseInt64List(v)
	}
	if v := os.Getenv("MMWX_TGBOT_HTTP_TIMEOUT_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			c.HTTPTimeoutSeconds = n
		}
	}
	if v := os.Getenv("MMWX_TGBOT_WEBAPP_LISTEN"); v != "" {
		c.WebAppListen = v
	}
	if v := os.Getenv("MMWX_TGBOT_WEBAPP_URL"); v != "" {
		c.WebAppURL = v
	}
	if v := os.Getenv("MMWX_TGBOT_WEBAPP_DEV_PREVIEW"); v == "1" || v == "true" {
		c.WebAppDevPreview = true
	}
}

func parseInt64List(s string) []int64 {
	var out []int64
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if n, err := strconv.ParseInt(part, 10, 64); err == nil {
			out = append(out, n)
		}
	}
	return out
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.MMWXURL) == "" {
		return fmt.Errorf("mmwx_url 必填")
	}
	if strings.TrimSpace(c.MMWXAPIToken) == "" {
		return fmt.Errorf("mmwx_api_token 必填")
	}
	if strings.TrimSpace(c.TGBotToken) == "" {
		return fmt.Errorf("tg_bot_token 必填")
	}
	return nil
}

// IsAdmin 判断 tg_id 是否在 admin 名单。
func (c Config) IsAdmin(tgID int64) bool {
	for _, id := range c.AdminTGIDs {
		if id == tgID {
			return true
		}
	}
	return false
}
