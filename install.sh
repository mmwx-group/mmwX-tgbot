#!/usr/bin/env bash
# mmwX-tgbot 一键安装脚本(交互式)
#
#   curl -fsSL https://raw.githubusercontent.com/mmwx-group/mmwX-tgbot/main/install.sh | sudo bash
#   # 或下载后:sudo bash install.sh
#
# 流程:交互收集配置 → 下载对应架构二进制(GitHub Release)→ 写 config.yaml → 装 systemd 服务 → 启动。
set -euo pipefail

REPO="mmwx-group/mmwX-tgbot"
BIN_PATH="/usr/local/bin/mmwx-tgbot"
CONFIG_DIR="/etc/mmwx-tgbot"
CONFIG_FILE="$CONFIG_DIR/config.yaml"
SERVICE="mmwx-tgbot"
SERVICE_FILE="/etc/systemd/system/$SERVICE.service"

# ---- 颜色 ----
if [[ -t 1 ]]; then
  R=$'\e[0m'; B=$'\e[1m'; G=$'\e[32m'; Y=$'\e[33m'; C=$'\e[36m'; RED=$'\e[31m'
else R=; B=; G=; Y=; C=; RED=; fi
info(){ echo "${C}▶${R} $*"; }
ok(){ echo "${G}✅${R} $*"; }
warn(){ echo "${Y}⚠${R}  $*"; }
err(){ echo "${RED}✖${R} $*" >&2; }

[[ $EUID -eq 0 ]] || { err "请用 root 运行:sudo bash install.sh"; exit 1; }

# ---- 下载工具 ----
if command -v curl >/dev/null 2>&1; then DLO(){ curl -fsSL -o "$1" "$2"; }
elif command -v wget >/dev/null 2>&1; then DLO(){ wget -qO "$1" "$2"; }
else err "需要 curl 或 wget"; exit 1; fi

# ---- 架构检测 ----
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$(uname -m)" in
  x86_64|amd64) ARCH=amd64 ;;
  aarch64|arm64) ARCH=arm64 ;;
  *) err "不支持的架构:$(uname -m)"; exit 1 ;;
esac
[[ "$OS" == "linux" ]] || warn "当前系统 $OS 非 linux,systemd 步骤可能不适用"
ASSET="mmwx-tgbot-${OS}-${ARCH}"

# 交互输入从 /dev/tty 读,兼容 `curl ... | sudo bash` 管道执行
TTY=/dev/tty
ask(){ # ask VAR "提示" "默认值" "必填(1/0)"
  local _var=$1 _prompt=$2 _def=${3:-} _req=${4:-0} _val
  while :; do
    if [[ -n "$_def" ]]; then read -rp "  $_prompt [$_def]: " _val <"$TTY"; _val=${_val:-$_def}
    else read -rp "  $_prompt: " _val <"$TTY"; fi
    [[ -z "$_val" && "$_req" == "1" ]] && { warn "不能为空"; continue; }
    break
  done
  printf -v "$_var" '%s' "$_val"
}

echo
echo "${B}========== mmwX-tgbot 一键安装 ==========${R}"
echo
info "请输入配置(方括号内为默认值,回车采用):"
ask MMWX_URL   "主控地址 mmwx_url"                         "https://mmw.2ha.me" 1
MMWX_URL=${MMWX_URL%/}
ask API_TOKEN  "主控 admin API token (mmwx_api_token)"     ""                   1
ask BOT_TOKEN  "Telegram bot token (tg_bot_token)"         ""                   1
ask ADMIN_IDS  "管理员 TG ID,多个用逗号隔开 (admin_tg_ids)" ""                   1
ask WEBAPP_URL "Mini App 公网地址 webapp_url(需自配 nginx,可留空)" ""           0

# admin_tg_ids → YAML 流式列表 [a, b]
_yaml_ids=""
IFS=',' read -ra _ids <<<"$ADMIN_IDS"
for id in "${_ids[@]}"; do
  id="${id//[[:space:]]/}"; [[ -z "$id" ]] && continue
  [[ "$id" =~ ^[0-9]+$ ]] || { err "管理员 ID 必须为数字:$id"; exit 1; }
  _yaml_ids+="${_yaml_ids:+, }$id"
done
[[ -n "$_yaml_ids" ]] || { err "至少需要一个管理员 ID"; exit 1; }

# ---- 下载二进制 ----
echo
URL="https://github.com/$REPO/releases/latest/download/$ASSET"
info "下载 $ASSET ..."
tmp=$(mktemp)
if ! DLO "$tmp" "$URL"; then
  err "下载失败:$URL"
  err "确认该 Release 资产已发布(仓库需先打 tag 触发 CI 发版)。"
  rm -f "$tmp"; exit 1
fi
install -m 0755 "$tmp" "$BIN_PATH"; rm -f "$tmp"
ok "二进制已安装:$BIN_PATH ($($BIN_PATH -v 2>/dev/null || echo unknown))"

# ---- 写配置 ----
mkdir -p "$CONFIG_DIR"
write_cfg=1
if [[ -f "$CONFIG_FILE" ]]; then
  read -rp "  $CONFIG_FILE 已存在,覆盖?(y/N): " yn <"$TTY"
  [[ "${yn,,}" == "y" ]] || { warn "保留现有配置,不覆盖"; write_cfg=0; }
fi
if [[ "$write_cfg" == "1" ]]; then
  cat >"$CONFIG_FILE" <<EOF
mmwx_url: $MMWX_URL
mmwx_api_token: $API_TOKEN
tg_bot_token: $BOT_TOKEN
admin_tg_ids: [$_yaml_ids]

# Mini App(默认只听本机回环,由 nginx 反代到公网 HTTPS)
webapp_listen: "127.0.0.1:23088"
webapp_url: "${WEBAPP_URL}"
EOF
  chmod 600 "$CONFIG_FILE"
  ok "配置已写入:$CONFIG_FILE"
fi

# ---- systemd 服务 ----
cat >"$SERVICE_FILE" <<EOF
[Unit]
Description=mmwX Telegram bot
After=network.target

[Service]
Type=simple
ExecStart=$BIN_PATH -c $CONFIG_FILE
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
systemctl daemon-reload
systemctl enable "$SERVICE" >/dev/null 2>&1 || true
systemctl restart "$SERVICE"
sleep 2

echo
if systemctl is-active --quiet "$SERVICE"; then
  ok "$SERVICE 已启动并设为开机自启"
else
  err "$SERVICE 启动失败,最近日志:"
  journalctl -u "$SERVICE" --no-pager -n 20 || true
  exit 1
fi

echo
echo "${B}========== 安装完成 ==========${R}"
echo "  配置文件:$CONFIG_FILE"
echo "  查看日志:journalctl -u $SERVICE -f"
echo "  重启服务:systemctl restart $SERVICE"
echo "  修改配置后 systemctl restart $SERVICE 生效"
if [[ -n "$WEBAPP_URL" ]]; then
  echo
  warn "Mini App 还需在 nginx 给该域名加反代到 127.0.0.1:23088(location /app 与 /api/tg-webapp/),见 README《Mini App 的 nginx 反代》。"
fi
