#!/usr/bin/env bash
# DMHL 全平台打包：Linux 便携 tar.gz + deb；Windows 单 exe + NSIS 脚本
set -e
cd "$(dirname "$0")/.."

VERSION="${1:-0.2.0}"
mkdir -p dist

echo "==> 编译 Go 后端 (Linux amd64)"
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o dist/dmhl-server-linux-amd64 ./cmd/server

echo "==> 编译 Go 后端 (Windows amd64)"
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o dist/dmhl-server-windows-amd64.exe ./cmd/server

echo "==> Linux 便携 tar.gz"
tar -czf "dist/dmhl-${VERSION}-linux-portable.tar.gz" -C dist dmhl-server-linux-amd64

echo "==> Linux deb 包"
PKG="dist/dmhl-${VERSION}-amd64.deb"
STAGE=$(mktemp -d)
mkdir -p "$STAGE/DEBIAN" "$STAGE/usr/bin" "$STAGE/usr/share/applications" "$STAGE/usr/share/icons/hicolor/256x256/apps"
cp dist/dmhl-server-linux-amd64 "$STAGE/usr/bin/dmhl"
chmod +x "$STAGE/usr/bin/dmhl"
cp assets/linux/dmhl.png "$STAGE/usr/share/icons/hicolor/256x256/apps/dmhl.png"
cat > "$STAGE/usr/share/applications/dmhl.desktop" <<EOF
[Desktop Entry]
Name=DMHL
Comment=Darling Minecraft Hello Launcher
Exec=dmhl
Icon=dmhl
Type=Application
Categories=Game;
Terminal=false
EOF
cat > "$STAGE/DEBIAN/control" <<EOF
Package: dmhl
Version: $VERSION
Section: games
Priority: optional
Architecture: amd64
Depends: libc6
Maintainer: DarlingZeroTwo-Cloud
Description: Darling Minecraft Hello Launcher
 Flutter UI + Go backend Minecraft launcher.
EOF
dpkg-deb --build --root-owner-group "$STAGE" "$PKG" 2>/dev/null || dpkg-deb -b "$STAGE" "$PKG"
rm -rf "$STAGE"

echo "==> Windows NSIS 安装版脚本"
cat > dist/dmhl-installer.nsi <<'NSIS'
!include "MUI2.nsh"
Name "DMHL"
OutFile "dmhl-setup.exe"
InstallDir "$PROGRAMFILES64\DMHL"
Page directory
Page instfiles
Section "Install"
  SetOutPath "$INSTDIR"
  File "dmhl-server-windows-amd64.exe"
  CreateDirectory "$SMPROGRAMS\DMHL"
  CreateShortCut "$SMPROGRAMS\DMHL\DMHL.lnk" "$INSTDIR\dmhl-server-windows-amd64.exe"
SectionEnd
NSIS

echo "==> 完成"
ls -lh dist/
