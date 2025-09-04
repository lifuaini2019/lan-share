#!/bin/bash

# 祖宇字文共享 - 飞牛OS Docker镜像构建脚本 v2.1

echo "🐳 开始构建飞牛OS专用Docker镜像..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 检测系统架构
SYSTEM_ARCH=$(uname -m)
echo "🔍 检测到构建系统架构: $SYSTEM_ARCH"

# 确保构建目录存在
mkdir -p docker_build
cd docker_build

# 清理旧文件
rm -rf *

# 根据系统架构选择相应的二进制文件
if [[ "$SYSTEM_ARCH" == "x86_64" ]]; then
    TARGET_ARCH="amd64"
    BINARY_SOURCE="../build/zuyu-share-linux-amd64"
    DOCKER_PLATFORM="linux/amd64"
    echo "💻 构建 x86_64 架构镜像"
elif [[ "$SYSTEM_ARCH" == "aarch64" ]] || [[ "$SYSTEM_ARCH" == "arm64" ]]; then
    TARGET_ARCH="arm64"
    BINARY_SOURCE="../build/zuyu-share-linux-arm64"
    DOCKER_PLATFORM="linux/arm64"
    echo "📱 构建 ARM64 架构镜像"
else
    echo "❌ 不支持的构建架构: $SYSTEM_ARCH"
    exit 1
fi

# 检查二进制文件是否存在
if [ ! -f "$BINARY_SOURCE" ]; then
    echo "❌ 二进制文件不存在: $BINARY_SOURCE"
    echo "请先运行构建脚本生成二进制文件"
    exit 1
fi

# 复制二进制文件
echo "📦 复制二进制文件..."
cp "$BINARY_SOURCE" zuyu-share
chmod +x zuyu-share

# 创建模板目录结构
echo "📁 创建目录结构..."
mkdir -p templates static

# 复制模板文件
echo "📝 复制模板文件..."
cp ../templates/index.html templates/
cp ../static/style.css static/

# 创建默认配置文件
echo "⚙️ 创建默认配置文件..."
cat > templates_config.json << 'EOF'
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
      "templates": [
        {
          "title": "在线客服",
          "content": "我们一直在线，如果没能第一时间回复您，说明我们正在忙碌中。有任何机器方面的问题直接咨询即可，我看到消息会马上回复您的。"
        }
      ]
    },
    "express": {
      "icon": "📦",
      "name": "快递问题",
      "templates": []
    },
    "aftersale": {
      "icon": "🛠️",
      "name": "售后问题",
      "templates": [
        {
          "title": "保修说明",
          "content": "我们提供完善的保修服务：主机主板保修一年，打印头保修半年，电源附加线等保修一个月。"
        }
      ]
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
      "templates": [
        {
          "title": "数据管理",
          "content": "使用下方的导入导出功能管理模板数据，支持全部或指定栏目的数据备份和恢复。"
        }
      ]
    }
  }
}
EOF

# 创建 Dockerfile
echo "🐳 创建 Dockerfile..."
cat > Dockerfile << EOF
# 使用多阶段构建，针对飞牛OS优化
FROM --platform=$DOCKER_PLATFORM alpine:3.18

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 设置时区为中国
RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \\
    echo "Asia/Shanghai" > /etc/timezone

# 创建应用目录
WORKDIR /app

# 复制二进制文件和资源
COPY zuyu-share /app/zuyu-share
COPY templates/ /app/templates/
COPY static/ /app/static/
COPY templates_config.json /app/templates_config.json

# 创建数据目录（用于持久化存储）
RUN mkdir -p /app/data

# 设置权限
RUN chmod +x /app/zuyu-share

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \\
    CMD wget --no-verbose --tries=1 --spider http://localhost:9405/ || exit 1

# 暴露端口
EXPOSE 9405

# 设置环境变量
ENV GIN_MODE=release
ENV TZ=Asia/Shanghai

# 启动命令
CMD ["/app/zuyu-share"]
EOF

# 构建Docker镜像
echo "🔨 构建Docker镜像..."
docker build --platform=$DOCKER_PLATFORM -t zuyu-share:latest .

if [ $? -ne 0 ]; then
    echo "❌ Docker镜像构建失败"
    exit 1
fi

# 验证镜像
echo "🔍 验证镜像信息..."
docker inspect zuyu-share:latest --format='{{.Architecture}}'

# 导出镜像文件
echo "📦 导出镜像文件..."
IMAGE_FILE="zuyu-share-flynas-${TARGET_ARCH}.tar"
docker save -o "../$IMAGE_FILE" zuyu-share:latest

if [ $? -eq 0 ]; then
    echo "✅ 镜像导出成功: $IMAGE_FILE"
    
    # 显示文件大小
    FILE_SIZE=$(ls -lh "../$IMAGE_FILE" | awk '{print $5}')
    echo "📏 镜像大小: $FILE_SIZE"
else
    echo "❌ 镜像导出失败"
    exit 1
fi

# 创建部署说明
echo "📝 创建部署说明..."
cat > "../flynas_docker_deploy_${TARGET_ARCH}.md" << EOF
# 飞牛OS Docker部署说明

## 镜像信息
- 文件名: $IMAGE_FILE
- 架构: $TARGET_ARCH
- 大小: $FILE_SIZE
- 构建时间: $(date '+%Y-%m-%d %H:%M:%S')

## 部署步骤

### 1. 上传镜像文件
将 \`$IMAGE_FILE\` 上传到飞牛OS主机

### 2. SSH连接飞牛OS
\`\`\`bash
ssh root@<飞牛OS_IP>
\`\`\`

### 3. 导入Docker镜像
\`\`\`bash
docker load -i $IMAGE_FILE
\`\`\`

### 4. 创建数据目录
\`\`\`bash
mkdir -p /opt/zuyu-share-data
\`\`\`

### 5. 运行容器
\`\`\`bash
docker run -d \\
    --name zuyu-share \\
    --restart unless-stopped \\
    -p 9405:9405 \\
    -v /opt/zuyu-share-data:/app/data \\
    zuyu-share:latest
\`\`\`

### 6. 验证部署
\`\`\`bash
# 查看容器状态
docker ps | grep zuyu-share

# 查看容器日志
docker logs zuyu-share

# 测试服务
curl http://localhost:9405
\`\`\`

## 访问服务
- 局域网访问: http://<飞牛OS_IP>:9405
- 本机访问: http://127.0.0.1:9405

## 管理命令
\`\`\`bash
# 停止容器
docker stop zuyu-share

# 启动容器
docker start zuyu-share

# 重启容器
docker restart zuyu-share

# 查看日志
docker logs zuyu-share -f

# 进入容器
docker exec -it zuyu-share sh

# 删除容器
docker stop zuyu-share && docker rm zuyu-share

# 删除镜像
docker rmi zuyu-share:latest
\`\`\`

## 数据持久化
- 配置文件: /opt/zuyu-share-data/templates_config.json
- 消息数据: /opt/zuyu-share-data/messages.txt
- 数据会在容器重启后保持

## 故障排除
1. 如果容器无法启动，检查端口9405是否被占用
2. 如果无法访问，检查防火墙设置
3. 查看容器日志获取详细错误信息

## 升级方法
1. 停止并删除旧容器
2. 导入新的镜像文件
3. 使用相同的运行命令启动新容器
4. 数据会自动保留
EOF

# 清理构建文件
cd ..
rm -rf docker_build

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🎉 飞牛OS Docker镜像构建完成！"
echo ""
echo "📦 生成的文件："
echo "   镜像文件: $IMAGE_FILE"
echo "   部署说明: flynas_docker_deploy_${TARGET_ARCH}.md"
echo ""
echo "🚀 快速部署命令："
echo "   docker load -i $IMAGE_FILE"
echo "   docker run -d --name zuyu-share --restart unless-stopped -p 9405:9405 zuyu-share:latest"
echo ""
echo "🌟 享受飞牛OS上的容器化部署体验！"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"