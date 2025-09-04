#!/bin/bash

# 祖宇字文共享 - 玩客云一键安装脚本 v2.1
# 支持：玩客云、路由器等 ARMv7 设备

echo "🚀 祖宇字文共享 - 玩客云一键安装脚本 v2.1"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 检查网络连接
echo "📡 检查网络连接..."
if ! ping -c 1 github.com >/dev/null 2>&1; then
    echo "❌ 网络连接失败，请检查网络后重试"
    exit 1
fi
echo "✅ 网络连接正常"

# 创建安装目录
INSTALL_DIR="/opt/zuyu-share"
echo "📁 创建安装目录: $INSTALL_DIR"
sudo mkdir -p $INSTALL_DIR
cd $INSTALL_DIR

# 检测系统架构
ARCH=$(uname -m)
echo "🔍 检测到系统架构: $ARCH"

# 确定下载文件名
if [[ "$ARCH" == "armv7l" ]] || [[ "$ARCH" == "arm" ]]; then
    BINARY_NAME="zuyu-share-linux-arm"
    echo "📱 目标文件: $BINARY_NAME (ARMv7)"
elif [[ "$ARCH" == "aarch64" ]] || [[ "$ARCH" == "arm64" ]]; then
    BINARY_NAME="zuyu-share-linux-arm64"
    echo "📱 目标文件: $BINARY_NAME (ARM64)"
elif [[ "$ARCH" == "x86_64" ]]; then
    BINARY_NAME="zuyu-share-linux-amd64"
    echo "💻 目标文件: $BINARY_NAME (x86_64)"
else
    echo "❌ 不支持的架构: $ARCH"
    exit 1
fi

# 下载最新版本
echo "⬇️ 正在从GitHub下载最新版本..."
DOWNLOAD_URL="https://github.com/username/LAN-Share-Go/releases/latest/download/$BINARY_NAME"

if command -v wget >/dev/null 2>&1; then
    sudo wget -O zuyu-share "$DOWNLOAD_URL"
elif command -v curl >/dev/null 2>&1; then
    sudo curl -L -o zuyu-share "$DOWNLOAD_URL"
else
    echo "❌ 需要 wget 或 curl 命令，请先安装"
    exit 1
fi

# 检查下载是否成功
if [ ! -f "zuyu-share" ] || [ ! -s "zuyu-share" ]; then
    echo "❌ 下载失败，请检查网络连接或GitHub仓库地址"
    exit 1
fi

# 设置执行权限
echo "🔧 设置执行权限..."
sudo chmod +x zuyu-share

# 创建配置文件
echo "📝 创建配置文件..."
sudo tee templates_config.json > /dev/null << 'EOF'
{
  "categories": {
    "home": {
      "icon": "🏠",
      "name": "共享文字",
      "templates": [
        {
          "title": "欢迎使用",
          "content": "欢迎使用祖宇字文共享系统！"
        },
        {
          "title": "使用提示",
          "content": "💡 小提示：电脑端按回车键可快速提交内容，支持多设备实时同步。"
        }
      ]
    },
    "presale": {
      "icon": "💬",
      "name": "售前问题",
      "templates": []
    },
    "express": {
      "icon": "📦",
      "name": "快递问题",
      "templates": []
    },
    "aftersale": {
      "icon": "🛠️",
      "name": "售后问题",
      "templates": []
    },
    "purchase": {
      "icon": "🛒",
      "name": "购买链接",
      "templates": []
    },
    "repair": {
      "icon": "🔧",
      "name": "维修问题",
      "templates": []
    },
    "settings": {
      "icon": "⚙️",
      "name": "系统设置",
      "templates": []
    }
  }
}
EOF

# 创建系统服务文件
echo "🔧 创建系统服务..."
sudo tee /etc/init.d/zuyu-share > /dev/null << 'EOF'
#!/bin/sh /etc/rc.common

START=95
STOP=10

USE_PROCD=1
PROG=/opt/zuyu-share/zuyu-share
PIDFILE=/var/run/zuyu-share.pid

start_service() {
    procd_open_instance
    procd_set_param command $PROG
    procd_set_param pidfile $PIDFILE
    procd_set_param respawn
    procd_set_param stdout 1
    procd_set_param stderr 1
    procd_close_instance
}

stop_service() {
    [ -f $PIDFILE ] && kill $(cat $PIDFILE)
}
EOF

# 设置服务权限
sudo chmod +x /etc/init.d/zuyu-share

# 启用开机自启动
echo "🔄 启用开机自启动..."
if command -v systemctl >/dev/null 2>&1; then
    # systemd 系统
    sudo tee /etc/systemd/system/zuyu-share.service > /dev/null << EOF
[Unit]
Description=Zuyu Share Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/zuyu-share
ExecStart=/opt/zuyu-share/zuyu-share
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
EOF
    
    sudo systemctl daemon-reload
    sudo systemctl enable zuyu-share
elif command -v rc-update >/dev/null 2>&1; then
    # OpenRC 系统 (OpenWrt)
    sudo rc-update add zuyu-share default
else
    echo "⚠️ 无法自动配置开机自启动，请手动配置"
fi

# 获取本机IP
LOCAL_IP=$(ip route get 8.8.8.8 | head -1 | awk '{print $7}')
if [ -z "$LOCAL_IP" ]; then
    LOCAL_IP=$(hostname -I | awk '{print $1}')
fi

# 启动服务
echo "🚀 启动服务..."
if command -v systemctl >/dev/null 2>&1; then
    sudo systemctl start zuyu-share
    sleep 2
    if sudo systemctl is-active --quiet zuyu-share; then
        echo "✅ 服务启动成功"
    else
        echo "❌ 服务启动失败，尝试手动启动..."
        cd /opt/zuyu-share
        sudo nohup ./zuyu-share > /var/log/zuyu-share.log 2>&1 &
        sleep 2
    fi
else
    cd /opt/zuyu-share
    sudo nohup ./zuyu-share > /var/log/zuyu-share.log 2>&1 &
    sleep 2
fi

# 检查端口是否正常监听
if netstat -ln | grep -q ":9405"; then
    echo "✅ 端口 9405 监听正常"
else
    echo "⚠️ 端口 9405 未监听，请检查服务状态"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🎉 安装完成！"
echo ""
echo "📱 访问地址："
echo "   局域网访问: http://$LOCAL_IP:9405"
echo "   本机访问: http://127.0.0.1:9405"
echo ""
echo "🔧 服务管理命令："
if command -v systemctl >/dev/null 2>&1; then
    echo "   启动服务: sudo systemctl start zuyu-share"
    echo "   停止服务: sudo systemctl stop zuyu-share"
    echo "   重启服务: sudo systemctl restart zuyu-share"
    echo "   查看状态: sudo systemctl status zuyu-share"
    echo "   查看日志: sudo journalctl -u zuyu-share -f"
else
    echo "   启动服务: sudo /etc/init.d/zuyu-share start"
    echo "   停止服务: sudo /etc/init.d/zuyu-share stop"
    echo "   重启服务: sudo /etc/init.d/zuyu-share restart"
    echo "   查看日志: sudo tail -f /var/log/zuyu-share.log"
fi
echo ""
echo "📁 程序目录: /opt/zuyu-share"
echo "🌟 享受局域网共享的便利！"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"