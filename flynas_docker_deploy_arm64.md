# 飞牛OS Docker部署说明

## 镜像信息
- 文件名: zuyu-share-flynas-arm64.tar
- 架构: arm64
- 大小: 8.2M
- 构建时间: 2025-09-04 10:51:45

## 部署步骤

### 1. 上传镜像文件
将 `zuyu-share-flynas-arm64.tar` 上传到飞牛OS主机

### 2. SSH连接飞牛OS
```bash
ssh root@<飞牛OS_IP>
```

### 3. 导入Docker镜像
```bash
docker load -i zuyu-share-flynas-arm64.tar
```

### 4. 创建数据目录
```bash
mkdir -p /opt/zuyu-share-data
```

### 5. 运行容器
```bash
docker run -d \
    --name zuyu-share \
    --restart unless-stopped \
    -p 9405:9405 \
    -v /opt/zuyu-share-data:/app/data \
    zuyu-share:latest
```

### 6. 验证部署
```bash
# 查看容器状态
docker ps | grep zuyu-share

# 查看容器日志
docker logs zuyu-share

# 测试服务
curl http://localhost:9405
```

## 访问服务
- 局域网访问: http://<飞牛OS_IP>:9405
- 本机访问: http://127.0.0.1:9405

## 管理命令
```bash
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
```

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
