# 祖宇字文共享系统 v1.0.15

## 项目简介

祖宇字文共享系统是一个专为局域网环境设计的文字共享工具，支持多设备间实时同步文字内容。该系统采用Go语言开发，具有响应式设计，适配手机端和电脑端访问。

## 核心功能

- ✅ **实时文字共享**：支持多设备间实时同步文字内容
- ✅ **模板管理**：预设常用回复模板，快速复制使用
- ✅ **智能局域网检测**：自动识别局域网环境并提示切换
- ✅ **响应式设计**：支持电脑端和手机端访问
- ✅ **调试工具**：完整的调试和测试页面
- 🔄 **文件传输**：支持局域网内文件快速传输（开发中）
- 🔄 **二维码分享**：自动生成访问二维码（开发中）

## 快速开始

### 1. 本地运行

```bash
# 直接运行Go程序
go run main_simple.go

# 或者构建后运行
go build -o zuyu-share main_simple.go
./zuyu-share
```

### 2. 访问系统

- **本地访问**: http://localhost:9405
- **局域网访问**: http://[本机IP]:9405

### 3. 玩客云部署

使用提供的一键部署脚本：

```bash
# 为玩客云编译ARM版本
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -o zuyu-share-openwrt-armv7 main_simple.go

# 运行一键部署脚本
./1.sh
```

## 系统架构

### 技术栈
- **后端**: Go 1.21+ (标准库)
- **前端**: HTML5/CSS3/JavaScript
- **通信**: HTTP REST API
- **部署**: 支持Linux ARM (玩客云)

### 项目结构
```
├── main_simple.go          # 简化版主程序
├── main.go                 # 完整版主程序（需要外部依赖）
├── templates_config.json   # 模板配置文件
├── templates/              # HTML模板目录
│   ├── index.html          # 主页面
│   ├── debug_lan_detection.html  # 调试页面
│   ├── test-domain.html    # 域名测试页面
│   └── ...
├── static/                 # 静态资源目录
│   └── style.css           # 样式文件
├── 1.sh                    # 玩客云一键部署脚本
└── README.md               # 项目说明
```

## 功能说明

### 智能局域网检测

系统会自动检测用户的网络环境：

1. **检测条件**：
   - 通过域名访问（非IP地址）
   - 客户端在局域网环境中
   - 服务器地址可获取

2. **显示提示**：满足条件时自动显示切换到局域网地址的提示

3. **优势**：局域网访问速度更快，延迟更低

### 调试工具

- `/debug-lan-detection` - 局域网检测调试页面
- `/test-domain` - 域名测试页面
- `/advanced-debug` - 增强调试工具（完整版）
- `/smart-detection-help` - 智能检测帮助页面

## API接口

### 消息管理
- `GET /api/messages` - 获取消息列表
- `POST /api/messages` - 发送新消息
- `DELETE /api/messages/{id}` - 删除指定消息

### 模板管理
- `GET /api/templates` - 获取模板配置

### 网络检测
- `GET /api/lan-check` - 检查局域网环境

## 部署说明

### 玩客云部署

1. **编译ARM版本**：
   ```bash
   CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -o zuyu-share-openwrt-armv7 main_simple.go
   ```

2. **运行部署脚本**：
   ```bash
   ./1.sh
   ```

3. **访问地址**：
   - 局域网：http://[玩客云IP]:9405
   - 域名：http://2lan.lifu.eu.org

### 手动部署

1. 上传程序文件到服务器
2. 设置执行权限：`chmod +x zuyu-share-openwrt-armv7`
3. 后台运行：`nohup ./zuyu-share-openwrt-armv7 > app.log 2>&1 &`
4. 检查运行状态：`netstat -ln | grep 9405`

## 版本信息

- **当前版本**: v1.0.15
- **发布日期**: 2025年1月
- **更新内容**: 
  - 智能局域网检测功能
  - 简化版本实现（无外部依赖）
  - 完整的调试工具集
  - 响应式设计优化

## 开发说明

### 两个版本

1. **main_simple.go** - 简化版本
   - 仅使用Go标准库
   - 核心功能完整
   - 易于部署和维护

2. **main.go** - 完整版本
   - 包含WebSocket实时通信
   - 文件上传功能
   - 二维码生成
   - 需要外部依赖包

### 扩展开发

如需添加新功能，建议：
1. 先在简化版本中实现基础功能
2. 在完整版本中添加高级特性
3. 保持两个版本的API兼容性

## 故障排除

### 常见问题

1. **端口被占用**
   ```bash
   # 检查端口占用
   netstat -ln | grep 9405
   # 杀死占用进程
   lsof -ti:9405 | xargs kill -9
   ```

2. **权限问题**
   ```bash
   chmod +x zuyu-share-openwrt-armv7
   ```

3. **网络检测失败**
   - 检查防火墙设置
   - 确认网络连接正常
   - 使用调试页面排查

### 日志查看

```bash
# 查看运行日志
tail -f app.log

# 查看系统日志
journalctl -f
```

## 许可证

本项目采用 MIT 许可证。

## 联系方式

如有问题或建议，请通过以下方式联系：
- 项目地址：[GitHub仓库地址]
- 邮箱：[联系邮箱]

---

*祖宇字文共享系统 - 让局域网文字共享更简单*