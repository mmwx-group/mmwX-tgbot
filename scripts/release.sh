#!/bin/bash
# 一键发布脚本(对齐 ../miaomiaowux/scripts/release.sh)
#
# 流程:确定新版本 -> 收集 changelog -> 更新 README 更新日志 -> commit -> tag -> push -> 创建 GitHub Release
# 二进制由 .github/workflows/build.yml 在 tag 推送后自动打包并附到该 Release。
#
# 用法:
#   bash scripts/release.sh            # 在上个 tag 基础上 +patch
#   bash scripts/release.sh minor      # +minor / major / patch
#   bash scripts/release.sh v1.4.0     # 指定具体版本
#
# 依赖:git、gh(已 gh auth login)
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT"

command -v gh >/dev/null 2>&1 || { echo "[ERROR] 需要 GitHub CLI(gh),先 gh auth login"; exit 1; }
git rev-parse --is-inside-work-tree >/dev/null 2>&1 || { echo "[ERROR] 当前目录不是 git 仓库"; exit 1; }

# 工作区需干净(避免把无关改动一起发布)
if [ -n "$(git status --porcelain)" ]; then
  echo "[ERROR] 工作区有未提交改动,请先 commit 或 stash"
  git status --short
  exit 1
fi

# 上一个 tag(可能没有,首发)
PREV_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")

# 计算新版本:参数优先(vX.Y.Z / X.Y.Z),否则在上个 tag 上 bump(默认 patch)
bump() { # bump <prev vX.Y.Z> <part>
  local v=${1#v}; local ma mi pa
  IFS=. read -r ma mi pa <<<"$v"
  ma=${ma:-0}; mi=${mi:-0}; pa=${pa:-0}
  case "$2" in
    major) ma=$((ma + 1)); mi=0; pa=0 ;;
    minor) mi=$((mi + 1)); pa=0 ;;
    *)     pa=$((pa + 1)) ;;
  esac
  echo "v${ma}.${mi}.${pa}"
}

ARG="${1:-patch}"
case "$ARG" in
  patch|minor|major)
    if [ -z "$PREV_TAG" ]; then NEW_TAG="v0.1.0"; else NEW_TAG=$(bump "$PREV_TAG" "$ARG"); fi
    ;;
  v[0-9]*) NEW_TAG="$ARG" ;;
  [0-9]*)  NEW_TAG="v$ARG" ;;
  *) echo "[ERROR] 无法识别的版本参数: $ARG(用 patch/minor/major 或 vX.Y.Z)"; exit 1 ;;
esac

if git rev-parse "$NEW_TAG" >/dev/null 2>&1; then
  echo "[ERROR] tag $NEW_TAG 已存在"; exit 1
fi

# 收集自上个 tag 以来的 commit(排除版本号 commit / merge commit)
if [ -n "$PREV_TAG" ]; then
  RANGE="${PREV_TAG}..HEAD"
else
  RANGE="HEAD"
fi
COMMITS=$(git log $RANGE --pretty=format:"- %s" --no-merges | grep -v "^- v[0-9]" | sort -u || true)
[ -n "$COMMITS" ] || COMMITS="- chore: release ${NEW_TAG}"

echo "=== 发布 ${NEW_TAG}(上个 tag: ${PREV_TAG:-无})==="
echo "$COMMITS"
echo ""

TODAY=$(date +%Y-%m-%d)

# 1) 更新 README 更新日志(插入到 <summary>更新日志</summary> 之后;无标记则跳过)
echo "[1/4] 更新 README 更新日志..."
INSERT_LINE=$(grep -n '<summary>更新日志</summary>' "$PROJECT_ROOT/README.md" | head -1 | cut -d: -f1 || true)
if [ -n "$INSERT_LINE" ]; then
  TMPFILE=$(mktemp)
  {
    echo "### ${NEW_TAG} (${TODAY})"
    echo "$COMMITS"
    echo ""
  } >"$TMPFILE"
  {
    head -n "$INSERT_LINE" "$PROJECT_ROOT/README.md"
    cat "$TMPFILE"
    tail -n +"$((INSERT_LINE + 1))" "$PROJECT_ROOT/README.md"
  } >"$PROJECT_ROOT/README.md.tmp"
  mv "$PROJECT_ROOT/README.md.tmp" "$PROJECT_ROOT/README.md"
  rm -f "$TMPFILE"
  echo "  -> README 已更新"
else
  echo "  -> 未找到「更新日志」标记,跳过 README"
fi

# 2) commit + tag
echo "[2/4] 创建 commit 和 tag..."
git add -A
git commit -m "$NEW_TAG" --no-verify || echo "  (无文件改动,仅打 tag)"
git tag "$NEW_TAG"

# 3) push
echo "[3/4] 推送到远程..."
BRANCH=$(git rev-parse --abbrev-ref HEAD)
git push origin "$BRANCH"
git push origin "$NEW_TAG"

# 4) GitHub Release(二进制由 CI 在 tag 推送后自动构建并附加)
echo "[4/4] 创建 GitHub Release..."
RELEASE_BODY="## 更新日志

### ${NEW_TAG} (${TODAY})
${COMMITS}

---
二进制(linux / darwin / windows × amd64 / arm64)与 Docker 镜像由 GitHub Actions 自动构建。
一键安装:\`curl -fsSL https://raw.githubusercontent.com/mmwx-group/mmwX-tgbot/main/install.sh | sudo bash\`"

gh release create "$NEW_TAG" \
  --title "$NEW_TAG" \
  --notes "$RELEASE_BODY" \
  --generate-notes \
  --latest

echo ""
echo "=== 发布完成! ${NEW_TAG} ==="
echo "  Release: https://github.com/mmwx-group/mmwX-tgbot/releases/tag/${NEW_TAG}"
echo "  GitHub Actions 将自动打包二进制 + 推送 Docker 镜像到该 Release。"
