#!/usr/bin/env bash
# DMHL 打包脚本：Linux 便携 tar.gz + deb 骨架；Windows 单 exe
set -e
cd "$(dirname "$0")/.."

VERSION="${1:-0.1.0}"
mkdir -p dist

echo "==> 编译 Go 后端"
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o dist/dmhl-server-linux-amd64 ./cmd/server
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o dist/dmhl-server-windows-amd64.exe ./cmd/server

echo "==> 打 Linux 便携包"
tar -czf "dist/dmhl-${VERSION}-linux-portable.tar.gz" -C dist dmhl-server-linux-amd64

echo "==> 完成"
ls -lh dist/
