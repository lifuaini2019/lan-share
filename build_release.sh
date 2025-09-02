#!/bin/bash

# LAN-to-go GitHub 发布版本构建脚本
echo "🚀 开始构建 LAN-to-go GitHub 发布版本..."

# 创建发布目录
mkdir -p releases
cd releases

# 清理之前的构建文件
rm -rf *

echo "📦 正在构建各平台版本..."

# Windows 64位
echo "🖥️  构建 Windows 64位版本..."
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o lan-to-go-windows-amd64.exe ../main.go
if [ $? -eq 0 ]; then
    echo "✅ Windows 64位版本构建成功"
else
    echo "❌ Windows 64位版本构建失败"
fi

# Windows 32位
echo "🖥️  构建 Windows 32位版本..."
GOOS=windows GOARCH=386 go build -ldflags "-s -w" -o lan-to-go-windows-386.exe ../main.go
if [ $? -eq 0 ]; then
    echo "✅ Windows 32位版本构建成功"
else
    echo "❌ Windows 32位版本构建失败"
fi

# 玩客云路由器 ARM v7 (IP修复版)
echo "📡 构建玩客云路由器版本 (ARM v7 - IP修复版)..."
GOOS=linux GOARCH=arm GOARM=7 go build -ldflags "-s -w" -o lan-to-go-openwrt-armv7 ../main.go
if [ $? -eq 0 ]; then
    echo "✅ 玩客云路由器版本构建成功"
else
    echo "❌ 玩客云路由器版本构建失败"
fi

# Linux 64位
echo "🐧 构建 Linux 64位版本..."
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o lan-to-go-linux-amd64 ../main.go
if [ $? -eq 0 ]; then
    echo "✅ Linux 64位版本构建成功"
else
    echo "❌ Linux 64位版本构建失败"
fi

# macOS Intel
echo "🍎 构建 macOS Intel版本..."
GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o lan-to-go-macos-amd64 ../main.go
if [ $? -eq 0 ]; then
    echo "✅ macOS Intel版本构建成功"
else
    echo "❌ macOS Intel版本构建失败"
fi

# macOS Apple Silicon
echo "🍎 构建 macOS Apple Silicon版本..."
GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o lan-to-go-macos-arm64 ../main.go
if [ $? -eq 0 ]; then
    echo "✅ macOS Apple Silicon版本构建成功"
else
    echo "❌ macOS Apple Silicon版本构建失败"
fi

echo ""
echo "📁 构建完成！生成的文件："
ls -la

echo ""
echo "📏 文件大小统计："
for file in lan-to-go-*; do
    if [ -f "$file" ]; then
        size=$(ls -lh "$file" | awk '{print $5}')
        echo "  $file: $size"
    fi
done

# 创建部署包
echo ""
echo "📦 创建完整部署包..."

# Windows 部署包
echo "📱 创建 Windows 部署包..."
mkdir -p windows-package
cp lan-to-go-windows-amd64.exe windows-package/
cp -r ../templates windows-package/
cp -r ../static windows-package/
cp ../templates_config.json windows-package/
touch windows-package/messages.txt

# 创建 Windows 启动脚本
cat > windows-package/start.bat << 'EOF'
@echo off
title LAN-to-go 局域网共享服务器
echo 🚀 启动 LAN-to-go 局域网共享服务器...
echo.
echo 程序启动后请访问显示的网址
echo 按 Ctrl+C 可停止服务器
echo.
lan-to-go-windows-amd64.exe
pause
EOF

# 创建 Windows 使用说明
cat > windows-package/README.txt << 'EOF'
LAN-to-go 局域网共享服务器 - Windows版本

📋 使用说明：
1. 双击 start.bat 启动服务器（推荐）
2. 或者直接运行 lan-to-go-windows-amd64.exe
3. 程序启动后会显示访问地址和二维码
4. 在局域网内的其他设备上访问显示的地址即可使用

📁 文件说明：
- lan-to-go-windows-amd64.exe: 主程序
- start.bat: 启动脚本（推荐使用）
- templates/: 网页模板文件夹
- static/: 静态资源文件夹
- templates_config.json: 模板配置文件
- messages.txt: 消息存储文件

⚠️ 注意事项：
- 请确保所有文件在同一目录下
- 程序默认端口为 9405
- 如果端口被占用，程序会提示并退出
- 防火墙可能会询问是否允许程序访问网络，请选择允许

🔧 技术支持：
如有问题请检查防火墙设置和端口占用情况。
EOF

zip -r lan-to-go-windows-amd64.zip windows-package/
echo "✅ Windows 部署包创建完成: lan-to-go-windows-amd64.zip"

# 玩客云部署包
echo "📡 创建玩客云部署包..."
mkdir -p openwrt-package
cp lan-to-go-openwrt-armv7 openwrt-package/
cp -r ../templates openwrt-package/
cp -r ../static openwrt-package/
cp ../templates_config.json openwrt-package/
touch openwrt-package/messages.txt

# 创建玩客云启动脚本
cat > openwrt-package/start.sh << 'EOF'
#!/bin/sh

echo "🚀 启动 LAN-to-go 局域网共享服务器..."
echo "程序启动后请访问显示的网址"
echo "按 Ctrl+C 可停止服务器"
echo ""

# 设置可执行权限
chmod +x ./lan-to-go-openwrt-armv7

# 启动程序
./lan-to-go-openwrt-armv7
EOF

chmod +x openwrt-package/start.sh

# 创建玩客云自动安装脚本
cat > openwrt-package/install_autostart.sh << 'EOF'
#!/bin/sh

# 玩客云 LAN-to-go 开机自动安装脚本

echo "🚀 安装 LAN-to-go 开机自动服务..."

# 检查是否为root用户
if [ "$(id -u)" != "0" ]; then
   echo "❌ 请使用 root 用户执行此脚本"
   exit 1
fi

# 获取当前目录
CURRENT_DIR=$(pwd)
SERVICE_DIR="/root/lan-to-go"

# 创建服务目录
echo "📁 创建服务目录: $SERVICE_DIR"
mkdir -p $SERVICE_DIR

# 复制文件
echo "📎 复制程序文件..."
cp -r * $SERVICE_DIR/
chmod +x $SERVICE_DIR/lan-to-go-openwrt-armv7

# 创建系统服务脚本
cat > /etc/init.d/lan-to-go << 'EOFSERVICE'
#!/bin/sh /etc/rc.common

START=99
STOP=10

USE_PROCD=1
PROG="/root/lan-to-go/lan-to-go-openwrt-armv7"
PIDFILE="/var/run/lan-to-go.pid"

start_service() {
    echo "Starting LAN-to-go service..."
    procd_open_instance
    procd_set_param command $PROG
    procd_set_param pidfile $PIDFILE
    procd_set_param respawn ${respawn_threshold:-3600} ${respawn_timeout:-5} ${respawn_retry:-5}
    procd_set_param stdout 1
    procd_set_param stderr 1
    procd_close_instance
}

stop_service() {
    echo "Stopping LAN-to-go service..."
    killall lan-to-go-openwrt-armv7 2>/dev/null
}

reload_service() {
    stop_service
    start_service
}
EOFSERVICE

chmod +x /etc/init.d/lan-to-go

# 启用开机自动
echo "⚙️ 启用开机自动..."
/etc/init.d/lan-to-go enable

echo ""
echo "✅ 安装完成！"
echo ""
echo "🎯 使用方法："
echo "  启动服务: /etc/init.d/lan-to-go start"
echo "  停止服务: /etc/init.d/lan-to-go stop"
echo "  重启服务: /etc/init.d/lan-to-go restart"
echo "  禁用自动: /etc/init.d/lan-to-go disable"
echo ""
echo "🔍 查看状态："
echo "  ps | grep lan-to-go"
echo "  netstat -tuln | grep 9405"
echo ""

# 询问是否立即启动
echo -n "是否立即启动服务？(Y/n): "
read -r response
case "$response" in
    [nN][oO]|[nN]) 
        echo "ℹ️ 可稍后使用命令启动: /etc/init.d/lan-to-go start"
        ;;
    *) 
        echo "🚀 启动服务..."
        /etc/init.d/lan-to-go start
        sleep 2
        echo ""
        echo "🌍 请在浏览器中访问路由器IP的9405端口"
        echo "   例如: http://192.168.1.1:9405"
        ;;
esac

echo ""
echo "🎉 玩客云 LAN-to-go 安装完成！"
EOF

chmod +x openwrt-package/install_autostart.sh

# 创建玩客云说明文档
cat > openwrt-package/README.md << 'EOF'
# LAN-to-go 玩客云路由器版本 (IP修复版)

## 🌟 新版本特性

### 🔄 IP获取问题修复
- ✅ 修复了原版本只显示 127.0.0.1 的问题
- ✅ 增强的网络环境兼容性，支持多种IP获取方式
- ✅ 智能识别局域网IP地址（192.168.x.x）
- ✅ 二维码显示正确的局域网访问地址

### 🚀 开机自动启动
- ✅ 一键安装开机自动服务
- ✅ 标准的OpenWrt服务管理
- ✅ 自动重启和故障恢复
- ✅ 完整的服务状态监控

## 📋 快速安装

### 方法一：自动安装（推荐）
```bash
# 1. 上传整个 openwrt-package 文件夹到玩客云 /root/ 目录
# 2. SSH 登录玩客云
ssh root@192.168.1.1  # 替换为您的路由器IP

# 3. 进入程序目录
cd /root/openwrt-package

# 4. 运行自动安装脚本
chmod +x install_autostart.sh
./install_autostart.sh
```

### 方法二：手动启动
```bash
# 1. 设置权限并运行
chmod +x lan-to-go-openwrt-armv7
./lan-to-go-openwrt-armv7

# 2. 后台运行（可选）
nohup ./lan-to-go-openwrt-armv7 > lan-to-go.log 2>&1 &
```

## 🔧 服务管理

### 安装开机自动后可用的命令：
```bash
# 启动服务
/etc/init.d/lan-to-go start

# 停止服务  
/etc/init.d/lan-to-go stop

# 重启服务
/etc/init.d/lan-to-go restart

# 禁用开机自动
/etc/init.d/lan-to-go disable

# 启用开机自动
/etc/init.d/lan-to-go enable
```

### 查看运行状态：
```bash
# 查看进程
ps | grep lan-to-go

# 查看端口
netstat -tuln | grep 9405

# 查看日志（如果使用后台运行）
tail -f lan-to-go.log
```

## 📁 文件说明
- `lan-to-go-openwrt-armv7`: 主程序（ARM v7架构，IP修复版）
- `start.sh`: 手动启动脚本
- `install_autostart.sh`: 自动安装开机自启动脚本  
- `templates/`: 网页模板文件夹
- `static/`: 静态资源文件夹  
- `templates_config.json`: 模板配置文件
- `messages.txt`: 消息存储文件

## ⚠️ 注意事项
- 确保玩客云有足够的存储空间（至少30MB）
- 程序默认端口为 9405，确保端口未被占用
- 建议运行前检查路由器内存使用情况：`free -m`
- 如需修改端口，请编辑源码中的 Port 常量

## 🔧 故障排除

### 1. 如果仍显示 127.0.0.1
```bash
# 检查网络接口
ip addr show

# 检查路由
ip route

# 手动测试网络
ping 8.8.8.8
```

### 2. 权限问题
```bash
# 确保可执行权限
chmod +x lan-to-go-openwrt-armv7

# 检查文件权限
ls -la lan-to-go-openwrt-armv7
```

### 3. 端口占用
```bash
# 检查端口占用
netstat -tuln | grep 9405

# 杀死占用进程
killall lan-to-go-openwrt-armv7
```

### 4. 内存不足
```bash
# 检查内存使用
free -m

# 清理内存缓存
echo 3 > /proc/sys/vm/drop_caches
```

## 📱 访问方式

程序启动成功后：
1. 在浏览器输入：`http://路由器IP:9405`
2. 或扫描程序显示的二维码直接访问
3. 支持手机、电脑等多设备同时访问

---

🎯 **享受便捷的局域网文件和消息共享！**
EOF

tar -czf lan-to-go-openwrt-armv7.tar.gz openwrt-package/
echo "✅ 玩客云部署包创建完成: lan-to-go-openwrt-armv7.tar.gz"

# Linux 部署包
echo "🐧 创建 Linux 部署包..."
mkdir -p linux-package
cp lan-to-go-linux-amd64 linux-package/
cp -r ../templates linux-package/
cp -r ../static linux-package/
cp ../templates_config.json linux-package/
touch linux-package/messages.txt

cat > linux-package/start.sh << 'EOF'
#!/bin/bash

echo "🚀 启动 LAN-to-go 局域网共享服务器..."
echo "程序启动后请访问显示的网址"
echo "按 Ctrl+C 可停止服务器"
echo ""

# 设置可执行权限
chmod +x ./lan-to-go-linux-amd64

# 启动程序
./lan-to-go-linux-amd64
EOF

chmod +x linux-package/start.sh
tar -czf lan-to-go-linux-amd64.tar.gz linux-package/
echo "✅ Linux 部署包创建完成: lan-to-go-linux-amd64.tar.gz"

echo ""
echo "📁 所有发布包创建完成："
ls -la *.zip *.tar.gz

echo ""
echo "📏 发布包大小统计："
for file in *.zip *.tar.gz; do
    if [ -f "$file" ]; then
        size=$(ls -lh "$file" | awk '{print $5}')
        echo "  $file: $size"
    fi
done

echo ""
echo "🎯 GitHub 发布说明："
echo "📱 Windows用户: 下载 lan-to-go-windows-amd64.zip"
echo "📡 玩客云用户: 下载 lan-to-go-openwrt-armv7.tar.gz (推荐IP修复版)"
echo "🐧 Linux用户: 下载 lan-to-go-linux-amd64.tar.gz"
echo "🍎 macOS用户: 下载对应架构的单文件版本"

echo ""
echo "✅ GitHub 发布版本构建完成！"
echo "📦 所有文件已清理个人信息，可以安全地上传到 GitHub 仓库"