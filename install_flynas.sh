#!/bin/bash

# 祖宇字文共享 - 飞牛OS一键安装脚本 v2.1
# 支持：飞牛OS、群晖NAS等 x86_64/ARM64 设备

echo "🚀 祖宇字文共享 - 飞牛OS一键安装脚本 v2.1"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 检查是否为root用户
if [ "$EUID" -ne 0 ]; then 
    echo "❌ 请使用 sudo 运行此脚本"
    exit 1
fi

# 检查网络连接
echo "📡 检查网络连接..."
if ! ping -c 1 github.com >/dev/null 2>&1; then
    echo "❌ 网络连接失败，请检查网络后重试"
    exit 1
fi
echo "✅ 网络连接正常"

# 检测系统架构
ARCH=$(uname -m)
echo "🔍 检测到系统架构: $ARCH"

# 确定下载文件名和Docker镜像
if [[ "$ARCH" == "x86_64" ]]; then
    BINARY_NAME="zuyu-share-linux-amd64"
    DOCKER_IMAGE_URL="https://github.com/lifuaini2019/lan-share/releases/latest/download/zuyu-share-flynas-x86_64.tar"
    echo "💻 目标文件: $BINARY_NAME (x86_64)"
elif [[ "$ARCH" == "aarch64" ]] || [[ "$ARCH" == "arm64" ]]; then
    BINARY_NAME="zuyu-share-linux-arm64"
    DOCKER_IMAGE_URL="https://github.com/lifuaini2019/lan-share/releases/latest/download/zuyu-share-flynas-arm64.tar"
    echo "📱 目标文件: $BINARY_NAME (ARM64)"
else
    echo "❌ 不支持的架构: $ARCH"
    echo "支持的架构: x86_64, aarch64/arm64"
    exit 1
fi

# 显示安装选项
echo ""
echo "🎯 请选择安装方式："
echo "  1) Docker 容器部署 (推荐)"
echo "  2) 直接二进制部署"
echo ""
read -p "请输入选择 (1 或 2): " INSTALL_TYPE

case $INSTALL_TYPE in
    1)
        echo "🐳 选择 Docker 容器部署"
        
        # 检查Docker是否安装
        if ! command -v docker >/dev/null 2>&1; then
            echo "❌ Docker 未安装，请先安装 Docker"
            exit 1
        fi
        
        # 创建Docker数据目录
        DOCKER_DATA_DIR="/opt/zuyu-share-data"
        mkdir -p $DOCKER_DATA_DIR
        echo "📁 创建数据目录: $DOCKER_DATA_DIR"
        
        # 下载Docker镜像
        echo "⬇️ 正在下载Docker镜像..."
        cd $DOCKER_DATA_DIR
        
        if command -v wget >/dev/null 2>&1; then
            wget -O zuyu-share-docker.tar "$DOCKER_IMAGE_URL"
        elif command -v curl >/dev/null 2>&1; then
            curl -L -o zuyu-share-docker.tar "$DOCKER_IMAGE_URL"
        else
            echo "❌ 需要 wget 或 curl 命令，请先安装"
            exit 1
        fi
        
        # 检查下载是否成功
        if [ ! -f "zuyu-share-docker.tar" ] || [ ! -s "zuyu-share-docker.tar" ]; then
            echo "❌ Docker镜像下载失败，请检查网络连接"
            exit 1
        fi
        
        # 导入Docker镜像
        echo "📦 导入Docker镜像..."
        docker load -i zuyu-share-docker.tar
        
        # 停止已存在的容器
        echo "🛑 停止已存在的容器..."
        docker stop zuyu-share 2>/dev/null || true
        docker rm zuyu-share 2>/dev/null || true
        
        # 创建配置文件
        echo "📝 创建配置文件..."
        tee $DOCKER_DATA_DIR/templates_config.json > /dev/null << 'EOF'
{
  "categories": {
    "home": {
      "icon": "🏠",
      "name": "共享文字",
      "templates": [
        {
          "title": "欢迎使用",
          "content": "欢迎使用祖宇字文共享系统！您可以在这里快速分享文字内容。"
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
        
        # 启动Docker容器
        echo "🚀 启动Docker容器..."
        docker run -d \
            --name zuyu-share \
            --restart unless-stopped \
            -p 9405:9405 \
            -v $DOCKER_DATA_DIR:/app/data \
            zuyu-share:latest
        
        # 检查容器状态
        sleep 3
        if docker ps | grep -q zuyu-share; then
            echo "✅ Docker容器启动成功"
        else
            echo "❌ Docker容器启动失败，查看日志:"
            docker logs zuyu-share
            exit 1
        fi
        
        INSTALL_PATH="Docker容器"
        ;;
    
    2)
        echo "📦 选择直接二进制部署"
        
        # 创建安装目录
        INSTALL_DIR="/opt/zuyu-share"
        echo "📁 创建安装目录: $INSTALL_DIR"
        mkdir -p $INSTALL_DIR
        cd $INSTALL_DIR
        
        # 下载二进制文件
        echo "⬇️ 正在从GitHub下载最新版本..."
        DOWNLOAD_URL="https://github.com/lifuaini2019/lan-share/releases/latest/download/$BINARY_NAME"
        
        if command -v wget >/dev/null 2>&1; then
            wget -O zuyu-share "$DOWNLOAD_URL"
        elif command -v curl >/dev/null 2>&1; then
            curl -L -o zuyu-share "$DOWNLOAD_URL"
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
        chmod +x zuyu-share
        
        # 创建必要目录
        echo "📁 创建必要目录..."
        mkdir -p templates static
        
        # 下载前端文件
        echo "⬇️ 下载前端文件..."
        if command -v wget >/dev/null 2>&1; then
            wget -O templates/index.html "https://github.com/lifuaini2019/lan-share/raw/main/templates/index.html" || echo "⚠️ 前端页面下载失败"
            wget -O static/style.css "https://github.com/lifuaini2019/lan-share/raw/main/static/style.css" || echo "⚠️ 样式文件下载失败"
        elif command -v curl >/dev/null 2>&1; then
            curl -L -o templates/index.html "https://github.com/lifuaini2019/lan-share/raw/main/templates/index.html" || echo "⚠️ 前端页面下载失败"
            curl -L -o static/style.css "https://github.com/lifuaini2019/lan-share/raw/main/static/style.css" || echo "⚠️ 样式文件下载失败"
        fi
        
        # 创建配置文件
        echo "📝 创建配置文件..."
        tee templates_config.json > /dev/null << 'EOF'
{
  "categories": {
    "home": {
      "icon": "🏠",
      "name": "共享文字",
      "templates": [
        {
          "title": "欢迎使用",
          "content": "欢迎使用祖宇字文共享系统！您可以在这里快速分享文字内容。"
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
        
        # 创建systemd服务
        echo "🔧 创建系统服务..."
        tee /etc/systemd/system/zuyu-share.service > /dev/null << EOF
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
KillMode=mixed
KillSignal=SIGINT
TimeoutStopSec=5

[Install]
WantedBy=multi-user.target
EOF
        
        # 启用并启动服务
        systemctl daemon-reload
        systemctl enable zuyu-share
        systemctl start zuyu-share
        
        # 检查服务状态
        sleep 2
        if systemctl is-active --quiet zuyu-share; then
            echo "✅ 系统服务启动成功"
        else
            echo "❌ 系统服务启动失败，查看日志:"
            journalctl -u zuyu-share --no-pager -n 10
            exit 1
        fi
        
        INSTALL_PATH="/opt/zuyu-share"
        ;;
    
    *)
        echo "❌ 无效选择"
        exit 1
        ;;
esac

# 获取本机IP
LOCAL_IP=$(ip route get 8.8.8.8 2>/dev/null | head -1 | awk '{print $7}')
if [ -z "$LOCAL_IP" ]; then
    LOCAL_IP=$(hostname -I 2>/dev/null | awk '{print $1}')
fi
if [ -z "$LOCAL_IP" ]; then
    LOCAL_IP="<本机IP>"
fi

# 检查端口是否正常监听
sleep 2
if netstat -ln 2>/dev/null | grep -q ":9405" || ss -ln 2>/dev/null | grep -q ":9405"; then
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

if [ "$INSTALL_TYPE" = "1" ]; then
    echo "🐳 Docker 管理命令："
    echo "   查看容器状态: docker ps | grep zuyu-share"
    echo "   查看容器日志: docker logs zuyu-share -f"
    echo "   停止容器: docker stop zuyu-share"
    echo "   启动容器: docker start zuyu-share"
    echo "   重启容器: docker restart zuyu-share"
    echo "   删除容器: docker stop zuyu-share && docker rm zuyu-share"
    echo ""
    echo "📁 数据目录: /opt/zuyu-share-data"
else
    echo "🔧 服务管理命令："
    echo "   启动服务: sudo systemctl start zuyu-share"
    echo "   停止服务: sudo systemctl stop zuyu-share"
    echo "   重启服务: sudo systemctl restart zuyu-share"
    echo "   查看状态: sudo systemctl status zuyu-share"
    echo "   查看日志: sudo journalctl -u zuyu-share -f"
    echo ""
    echo "📁 程序目录: /opt/zuyu-share"
fi

echo "🌟 享受飞牛OS上的局域网共享便利！"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"