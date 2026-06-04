#!/bin/bash
# 多平台交叉编译 mmwX-tgbot。
#
#   ./build.sh           # 编译当前架构
#   ./build.sh -a        # 全平台 (linux amd64/arm64 + darwin amd64/arm64 + windows amd64)
set -euo pipefail

BUILD_DIR="build"
PKG="github.com/iluobei/mmwX-tgbot/cmd/mmwx-tgbot"
VERSION="${VERSION:-dev}"
LDFLAGS="-s -w -X 'main.version=${VERSION}'"

ALL=0
while [[ $# -gt 0 ]]; do
    case "$1" in
        -a|--all) ALL=1; shift ;;
        -h|--help)
            sed -n '2,7p' "$0"; exit 0 ;;
        *) echo "unknown arg: $1" >&2; exit 1 ;;
    esac
done

mkdir -p "$BUILD_DIR"
rm -f "$BUILD_DIR"/*

build_one() {
    local os=$1 arch=$2 out=$3
    echo "[build] $os/$arch → $BUILD_DIR/$out"
    CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -trimpath -ldflags="$LDFLAGS" -o "$BUILD_DIR/$out" "$PKG"
}

if [[ $ALL -eq 0 ]]; then
    CGO_ENABLED=0 go build -trimpath -ldflags="$LDFLAGS" -o "$BUILD_DIR/mmwx-tgbot" "$PKG"
    ls -lh "$BUILD_DIR"
    exit 0
fi

build_one linux   amd64  mmwx-tgbot-linux-amd64
build_one linux   arm64  mmwx-tgbot-linux-arm64
build_one darwin  amd64  mmwx-tgbot-darwin-amd64
build_one darwin  arm64  mmwx-tgbot-darwin-arm64
build_one windows amd64  mmwx-tgbot-windows-amd64.exe
ls -lh "$BUILD_DIR"
