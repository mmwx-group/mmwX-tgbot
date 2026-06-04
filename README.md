# mmwX-tgbot

妙妙屋X 配套 Telegram bot — 独立项目,通过 HTTP 调主控 `/api/admin/tgbot/*` 域。

## 功能

**用户**
- `/start <code>` — 用邀请码注册新账号或绑定已有账号
- `/me` — 我的账号信息
- `/sub` — 拿订阅链接
- `/traffic` — 流量统计
- `/nodes` — 节点列表 + 服务器在线状态
- `/unbind` — 解绑 TG(`/unbind UNBIND` 二次确认)

**管理员**(配置 `admin_tg_ids` 名单)
- `/admin_invite list | revoke <code>` — 邀请码列表/撤销(完整 CRUD 走主控 web UI `/tg-bot-invites`)
- `/admin_user <username>` — 查指定用户

## 架构

```
TG user → mmwX-tgbot → mmwx 主控
                         │
                         └─ Authorization: Bearer <MMWX_API_TOKEN>
                            POST /api/admin/tgbot/bind
                            POST /api/admin/tgbot/unbind
                            GET  /api/admin/tgbot/user-by-tg?tg_id=
                            GET  /api/admin/tgbot/user-summary?username=
                            GET  /api/admin/tgbot/user-subscriptions?username=
                            GET  /api/admin/tgbot/user-nodes?username=
                            GET/POST /api/admin/tgbot/invites
```

- bot 持单一 admin token,内部按 TG ID 分发权限
- 无本地 db:所有 TG ↔ username 映射、邀请码、审计都在 mmwx 主控
- 默认 long-poll(零部署依赖);多步注册对话用 in-memory 状态机(10 分钟超时)

## 配置

复制 `config.example.yaml` → `config.yaml` 并填:

```yaml
mmwx_url: https://mmw.2ha.me
mmwx_api_token: <从主控系统设置取>
tg_bot_token: <@BotFather 给的>
admin_tg_ids: [123456789]
```

> 订阅链接基址固定为 `mmwx_url` + `/x`,无需单独配置。

也可全用环境变量(`MMWX_TGBOT_<UPPER_SNAKE>`):

```bash
MMWX_TGBOT_MMWX_URL=https://mmw.2ha.me
MMWX_TGBOT_MMWX_API_TOKEN=...
MMWX_TGBOT_TG_BOT_TOKEN=...
MMWX_TGBOT_ADMIN_TG_IDS=123,456
```

## 部署

### 方式 A:一键 SSH

```bash
TGBOT_SSH_HOST=root@your-host TGBOT_SSH_PORT=22 ./deploy.sh
# 默认部署到 /usr/local/bin/mmwx-tgbot,自带回滚
```

systemd unit 示例 `/etc/systemd/system/mmwx-tgbot.service`:

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

### 方式 B:Docker

```bash
docker compose up -d
```

或 docker run:

```bash
docker run -d --name mmwx-tgbot \
  -e MMWX_TGBOT_MMWX_URL=https://mmw.2ha.me \
  -e MMWX_TGBOT_MMWX_API_TOKEN=... \
  -e MMWX_TGBOT_TG_BOT_TOKEN=... \
  -e MMWX_TGBOT_ADMIN_TG_IDS=123 \
  --restart unless-stopped \
  ghcr.io/iluobei/mmwx-tgbot:latest
```

### 方式 C:本地交叉编译

```bash
./build.sh -a   # 全平台
ls build/
```

## 开发

```bash
go build ./...
go vet ./...
go test ./...    # 暂无单元测试

# 本地起 + 接本地 mmwx
go run ./cmd/mmwx-tgbot -c ./config.yaml
```

## 许可证

跟 mmwx 同 license。
