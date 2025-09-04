# 使用多阶段构建，针对飞牛OS优化
FROM --platform=linux/arm64 alpine:3.18

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 设置时区为中国
RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
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
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9405/ || exit 1

# 暴露端口
EXPOSE 9405

# 设置环境变量
ENV GIN_MODE=release
ENV TZ=Asia/Shanghai

# 启动命令
CMD ["/app/zuyu-share"]
