<div align="center">

# ❄️ mmwX-tgbot

**妙妙屋X 配套 Telegram 机器人** — 命令交互 + 每日通知 + 免登录 Mini App,一个二进制全搞定。

[![Go](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Build](https://github.com/mmwx-group/mmwX-tgbot/actions/workflows/build.yml/badge.svg)](https://github.com/mmwx-group/mmwX-tgbot/actions/workflows/build.yml)
[![Docker](https://img.shields.io/badge/ghcr.io-mmwx--group%2Fmmwx--tgbot-2496ED?logo=docker&logoColor=white)](https://github.com/mmwx-group/mmwX-tgbot/pkgs/container/mmwx-tgbot)
[![License](https://img.shields.io/badge/license-mmwx-blue)](#-许可证)

</div>

---

## 🧭 这是什么

独立部署的 Telegram bot,**自身无数据库**。所有数据(账号、套餐、流量、订阅、兑换码)都在 [妙妙屋X](https://mmw.domain.com) 主控,bot 通过 `Authorization: Bearer <admin token>` 调主控的 `/api/admin/tgbot/*` 域,只做 Telegram 这一层交互前端。

```
                       ┌──────────────────────────────┐
   Telegram 用户 ──────▶│         mmwX-tgbot           │
   (命令 / Mini App)    │  long-poll  +  HTTP(:23088)  │
                       └───────────────┬──────────────┘
                                       │ Bearer <MMWX_API_TOKEN>
                                       ▼
                       ┌──────────────────────────────┐
                       │   妙妙屋X 主控  mmw.domain.com│
                       │   /api/admin/tgbot/*         │
                       └──────────────────────────────┘
```

---

## ✨ 功能一览

<table>
<tr><td width="33%" valign="top">

### 👤 用户命令
| 命令 | 说明 |
|---|---|
| `/start <code>` | 兑换码注册 / 绑定 |
| `/me` | 我的账号信息 |
| `/sub` | 订阅链接 |
| `/traffic` | 流量统计 |
| `/nodes` | 我的节点 + 在线状态 |
| `/notify` | 通知开关 on/off/status |
| `/unbind` | 解绑 TG(二次确认) |
| `/help` | 帮助 |

</td><td width="33%" valign="top">

### 🛡️ 管理员命令
| 命令 | 说明 |
|---|---|
| `/admin_invite list` | 列兑换码 |
| `/admin_invite create` | 按钮交互生成 |
| `/admin_invite revoke <code>` | 撤销 |
| `/admin_user <username>` | 查指定用户 |

> 仅 `admin_tg_ids` 名单可用 · 限流 5 次/分

</td><td width="33%" valign="top">

### 🔔 每日通知
开通后(`/notify on`)bot 每天 **20:00** 推送:
- 📊 当日流量概况
- ⏰ 套餐剩 **7/3/1 天**到期提醒

opt-in,默认关闭,开关存主控。

</td></tr>
</table>

---

## 📱 Mini App(免登录面板)

点机器人菜单按钮「📊 我的面板」打开,**无需账号密码** —— Telegram 用 bot token 给 `initData` 签名,bot 校验后即可信任其中的 telegram_id,反查账号。

<table>
<tr>
<td width="25%" align="center">🏠<br><b>主页</b><br><sub>账号 · 流量 · 订阅 · 每日曲线 · 兑换续期</sub></td>
<td width="25%" align="center">📈<br><b>流量</b><br><sub>各节点已用流量排行</sub></td>
<td width="25%" align="center">📡<br><b>状态</b><br><sub>节点在线/离线/外部「?」</sub></td>
<td width="25%" align="center">🎟️<br><b>兑换码</b><br><sub>(仅管理员)生成/查看/撤销</sub></td>
</tr>
</table>

- 🆕 **未注册**用户:在面板内输入「兑换码 + 用户名 + 密码」注册并绑定。
- ♻️ **已注册**用户:输入兑换码**续期**(延长到期时间)。
- 🔐 **越权防护**:管理端点服务端强制校验 `IsAdmin(tgID)`,身份只来自签名后的 initData,前端无法指定查谁。
- 🎨 主题取自妙妙屋X 前端(陶土橙 `#d97757`,跟随 Telegram 明暗),顶部带特效「X」logo。

> ⚠️ Mini App 是一个本地 HTTP 服务(默认 `127.0.0.1:23088`),需前置 **nginx 反代到公网 HTTPS** 才能被 Telegram 加载,见 [部署](#-部署)。

---

## ⚙️ 配置

复制 `config.example.yaml` → `config.yaml`:

```yaml
mmwx_url: https://mmw.domain.com         # 主控地址(订阅链接基址固定为 mmwx_url + /x)
mmwx_api_token: <从主控系统设置取>      # admin 级 API token
tg_bot_token: <@BotFather 给的>
admin_tg_ids: [123456789]                  # 谁能用 /admin_* 与 Mini App 邀请码页
webapp_listen: "127.0.0.1:23088"           # Mini App 本地监听(默认只听回环)
webapp_url: "https://mmw-tgapp.domain.com/app"       # Mini App 公网地址;留空则不设菜单按钮
webapp_dev_preview: false                  # 调试用,生产保持 false(见下)
```

> 🔒 **安全默认**:Mini App 只听 `127.0.0.1`(靠 nginx 反代,不裸奔公网);API 端点 per-IP 限流 60 次/分;`webapp_dev_preview=false` 时仅信任 Telegram 注入的 initData(关闭 `?initData=` URL 预览,防 URL 泄漏重放)。

也可全用环境变量(`MMWX_TGBOT_<UPPER_SNAKE>`):

| 环境变量 | 对应字段 |
|---|---|
| `MMWX_TGBOT_MMWX_URL` | `mmwx_url` |
| `MMWX_TGBOT_MMWX_API_TOKEN` | `mmwx_api_token` |
| `MMWX_TGBOT_TG_BOT_TOKEN` | `tg_bot_token` |
| `MMWX_TGBOT_ADMIN_TG_IDS` | `admin_tg_ids`(逗号分隔) |
| `MMWX_TGBOT_WEBAPP_LISTEN` | `webapp_listen` |
| `MMWX_TGBOT_WEBAPP_URL` | `webapp_url` |
| `MMWX_TGBOT_HTTP_TIMEOUT_SECONDS` | `http_timeout_seconds`(默认 8) |

---

## 🚀 部署

### 方式 0 · 一键安装(推荐)🎉

在目标服务器上交互式安装(收集配置 → 下载二进制 → 装 systemd → 启动):

```bash
curl -fsSL https://raw.githubusercontent.com/mmwx-group/mmwX-tgbot/main/install.sh | sudo bash
```

依次输入 `mmwx_url`、`mmwx_api_token`、`tg_bot_token`、`admin_tg_ids`(`webapp_url` 可留空、配好 nginx 后再补)。配置在 `/etc/mmwx-tgbot/config.yaml`,服务名 `mmwx-tgbot`。

**更新到最新版**(复用现有配置,不交互):

```bash
curl -fsSL https://raw.githubusercontent.com/mmwx-group/mmwX-tgbot/main/install.sh | sudo bash -s update
```

### 方式 A · 编译后 SSH 推送(开发者,自带备份回滚)

```bash
TGBOT_SSH_HOST=root@your-host TGBOT_SSH_PORT=22 ./deploy.sh
```

systemd unit `/etc/systemd/system/mmwx-tgbot.service`:

```ini
[Unit]
Description=mmwX Telegram bot
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/mmwx-tgbot -c /etc/mmwx-tgbot/config.yaml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### 方式 B · Docker

```bash
docker compose up -d
```

或:

```bash
docker run -d --name mmwx-tgbot --restart unless-stopped \
  -p 127.0.0.1:23088:23088 \
  -e MMWX_TGBOT_MMWX_URL=https://mmw.domain.com \
  -e MMWX_TGBOT_MMWX_API_TOKEN=... \
  -e MMWX_TGBOT_TG_BOT_TOKEN=... \
  -e MMWX_TGBOT_ADMIN_TG_IDS=123 \
  -e MMWX_TGBOT_WEBAPP_LISTEN=:23088 \
  -e MMWX_TGBOT_WEBAPP_URL=https://mmw-tgbot.domain.com/app \
  ghcr.io/mmwx-group/mmwx-tgbot:latest
```

### 方式 C · 本地交叉编译

```bash
./build.sh -a   # linux/darwin/windows × amd64/arm64 → build/
```

### 🌐 Mini App 的 nginx 反代

bot 的 `127.0.0.1:23088` 只听本地,在主控域名(如 `mmw.domain.com`)的 nginx server 块里、`location /` **之前**加(`X-Real-IP` 用于限流准确性):

```nginx
location = /app {
    proxy_pass http://127.0.0.1:23088;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
}
location /api/tg-webapp/ {
    proxy_pass http://127.0.0.1:23088;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
}
```

`nginx -t && nginx -s reload` 后,把 `webapp_url` 设为 `https://你的域名/app` 重启 bot,菜单按钮即出现。

---

## 🛠️ 开发

```bash
go vet ./...
go build ./...
go run ./cmd/mmwx-tgbot -c ./config.yaml   # 本地起,接主控
```

**发布**(需 `gh` 已登录):`bash scripts/release.sh [patch|minor|major|vX.Y.Z]` —— 自动 bump 版本、写更新日志、打 tag、推送、建 Release;CI 随后自动打包二进制 + 推 Docker 镜像。

<details>
<summary>📂 项目结构</summary>

```
cmd/mmwx-tgbot/      入口:加载配置 → 起 long-poll + Mini App HTTP → 优雅关闭
internal/
  config/            YAML + 环境变量配置
  mmwxclient/        调主控 /api/admin/tgbot/* 的 HTTP 客户端
  bot/               命令路由 + 处理 + 每日通知 + Mini App(webapp.go / webapp_page.go)
```
</details>

<details>
<summary>🔌 调用的主控接口(参考)</summary>

| 端点 | 用途 |
|---|---|
| `POST /api/admin/tgbot/bind` · `redeem` · `unbind` | 注册绑定 / 续期 / 解绑 |
| `GET  /api/admin/tgbot/user-by-tg` · `user-summary` | TG→账号反查 / 账号摘要 |
| `GET  /api/admin/tgbot/user-subscriptions` · `user-nodes` | 订阅 / 节点+在线状态 |
| `GET/POST /api/admin/tgbot/invites` · `invites/revoke` | 兑换码 列/建/撤 |
| `POST /api/admin/tgbot/notify` · `GET notify-digest` | 通知开关 / 每日推送名单 |
| `GET  /api/admin/tgbot/user-daily-traffic` | 每日流量曲线 |
| `GET  /api/admin/packages` · `/api/admin/traffic/user-nodes` | 套餐列表 / 各节点已用流量 |

</details>

---

## 📜 更新日志

<details>
<summary>更新日志</summary>
### v0.0.9 (2026-07-10)
- 🌈增加复制兑换码按钮

### v0.0.8 (2026-06-05)
- 🌈切换logo

### v0.0.7 (2026-06-05)
- 🌈支持客户端选择与兑换码次数设置

### v0.0.6 (2026-06-05)
- 🌈支持客户端选择与兑换码次数设置

### v0.0.5 (2026-06-04)
- 🌈beta 0.0.5

### v0.0.4 (2026-06-04)
- 🌈beta0.0.3
- 🌈beta0.0.4

### v0.0.3 (2026-06-04)
- 🌈beta0.0.3

### v0.0.2 (2026-06-04)
- 🌈beta

### v0.0.1 (2026-06-04)
- 🌈beta

### v0.1.0 (2026-06-04)
- Create LICENSE
- 🌈beta


</details>

## 📄 许可证

跟妙妙屋X 同 license。

<div align="center"><sub>Made with ❄️ by <a href="https://github.com/mmwx-group">mmwx-group</a></sub></div>
