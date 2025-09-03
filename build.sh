#!/bin/bash

# 祖宇字文共享 - 全平台构建脚本 v2.1
echo "🚀 开始构建祖宇字文共享 v2.1 - 全平台版本..."

# 创建构建目录
mkdir -p build
cd build

# 清理之前的构建文件
rm -rf *

echo "📦 正在构建各平台可执行文件..."

# Windows 64位
echo "🖥️  构建 Windows 64位版本..."
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X main.version=v2.1.0" -o zuyu-share-windows-amd64.exe ../main.go

# Windows 32位
echo "🖥️  构建 Windows 32位版本..."
GOOS=windows GOARCH=386 go build -ldflags="-s -w -X main.version=v2.1.0" -o zuyu-share-windows-386.exe ../main.go

# Linux 64位
echo "🐧 构建 Linux 64位版本..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.version=v2.1.0" -o zuyu-share-linux-amd64 ../main.go

# Linux 32位
echo "🐧 构建 Linux 32位版本..."
GOOS=linux GOARCH=386 go build -ldflags="-s -w -X main.version=v2.1.0" -o zuyu-share-linux-386 ../main.go

# Linux ARM64 (飞牛OS、群晖等)
echo "📱 构建 Linux ARM64版本 (飞牛OS/群晖)..."
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w -X main.version=v2.1.0" -o zuyu-share-linux-arm64 ../main.go

# macOS Intel
echo "🍎 构建 macOS Intel版本..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X main.version=v2.1.0" -o zuyu-share-darwin-amd64 ../main.go

# macOS Apple Silicon
echo "🍎 构建 macOS Apple Silicon版本..."
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X main.version=v2.1.0" -o zuyu-share-darwin-arm64 ../main.go

# OpenWrt ARMv7 (玩客云等路由器)
echo "📡 构建 OpenWrt ARMv7版本 (玩客云路由器)..."
GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="-s -w -X main.version=v2.1.0" -o zuyu-share-linux-arm ../main.go

echo "📋 构建完成! 生成的文件:"
ls -la

echo ""
echo "📏 文件大小统计:"
for file in zuyu-share-*; do
    if [ -f "$file" ]; then
        size=$(ls -lh "$file" | awk '{print $5}')
        echo "  $file: $size"
    fi
done

echo ""
echo "✅ 所有平台构建完成!"
echo ""
echo "🎯 使用说明:"
echo "📱 Windows用户: 使用 zuyu-share-windows-amd64.exe"
echo "🐧 Linux用户: 使用 zuyu-share-linux-amd64"  
echo "🍎 macOS用户: 使用 zuyu-share-darwin-amd64 或 zuyu-share-darwin-arm64"
echo "📡 路由器用户: 使用 zuyu-share-linux-arm"
echo "🌟 飞牛OS/群晖: 使用 zuyu-share-linux-arm64"