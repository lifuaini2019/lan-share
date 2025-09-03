# 祖宇字文共享 - LAN Share Go

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20Linux%20%7C%20macOS%20%7C%20OpenWrt-lightgrey.svg)]()

🌐 现代化的局域网共享服务器，支持实时消息同步、文件共享和模板管理

## ✨ 主要特性

- 🌐 **Web界面**: 现代化响应式设计，完美适配移动端
- 📱 **二维码访问**: 自动生成局域网地址，手机扫码即用
- 🔄 **实时同步**: WebSocket技术，多设备实时消息同步
- 📁 **文件共享**: 拖拽上传，支持16MB以下多种格式
- 💼 **模板系统**: 预设消息模板，客服场景快速响应
- 📊 **导入导出**: 支持TXT/JSON格式的数据管理
- 🎯 **跨平台**: Windows/Linux/macOS/OpenWrt全平台支持
- 🐳 **Docker**: 支持容器化部署

## 🚀 快速开始

### 下载运行

1. 从 [Releases](../../releases) 下载对应平台的可执行文件
2. 解压并运行程序
3. 打开浏览器访问 `http://localhost:9405`
4. 手机扫描二维码即可访问

### Docker 部署

```bash
# 构建镜像
go build -o zuyu-share main.go
docker build -t zuyu-share .

# 运行容器
docker run -d \
  --name zuyu-share \
  --restart unless-stopped \
  -p 9405:9405 \
  zuyu-share
```

## 📦 支持平台

| 平台 | 架构 | 文件名 |
|------|------|--------|
| Windows | x64 | `zuyu-share-windows-amd64.exe` |
| Windows | x86 | `zuyu-share-windows-386.exe` |
| Linux | x64 | `zuyu-share-linux-amd64` |
| Linux | x86 | `zuyu-share-linux-386` |
| Linux | ARM64 | `zuyu-share-linux-arm64` |
| macOS | Intel | `zuyu-share-darwin-amd64` |
| macOS | Apple Silicon | `zuyu-share-darwin-arm64` |
| OpenWrt | ARMv7 | `zuyu-share-linux-arm` |

## 🔧 本地构建

### 环境要求

- Go 1.21+
- Git

### 构建步骤

```bash
# 克隆项目
git clone https://github.com/你的用户名/LAN-Share-Go.git
cd LAN-Share-Go

# 安装依赖
go mod tidy

# 本地运行
go run main.go

# 构建可执行文件
go build -o zuyu-share main.go

# 构建所有平台 (Windows)
./build.bat

# 构建所有平台 (Linux/macOS)
chmod +x build.sh
./build.sh
```

## 🌟 功能亮点

### 智能协议检测
- IP访问自动使用HTTP协议
- 域名访问自动使用HTTPS协议
- 网络类型智能标识（局域网/广域网）

### 移动端优化
- 响应式界面设计
- 触摸友好的操作体验
- 点击空白处收起菜单

### 实时同步
- WebSocket长连接
- 多设备消息实时同步
- 连接状态实时显示

## 🛡️ 技术栈

- **后端**: Go 1.21 + Gin Framework
- **前端**: HTML5 + CSS3 + JavaScript
- **通信**: WebSocket + HTTP RESTful API
- **二维码**: go-qrcode
- **跨域**: Gin CORS

## 📖 使用说明

### 基本功能

1. **发送消息**: 在文本框输入内容，点击发送或按Enter
2. **文件上传**: 拖拽文件到上传区域或点击选择文件
3. **模板使用**: 点击预设模板快速发送常用消息
4. **清空消息**: 点击清空按钮清除所有消息
5. **导入导出**: 管理和备份消息数据

### 高级功能

- **模板管理**: 自定义消息模板，支持分类管理
- **实时同步**: 多个设备同时访问，消息实时同步
- **移动适配**: 完美支持手机、平板等移动设备

## 🔧 配置文件

`templates_config.json` - 模板配置文件
```json
{
  "通用": ["欢迎使用", "谢谢"],
  "问候": ["你好", "早上好", "晚安"]
}
```

## 📊 性能特点

- **内存占用**: 10-15MB
- **启动时间**: < 1秒
- **并发支持**: 1000+ 连接
- **文件大小**: 单文件 < 16MB

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

本项目采用 [MIT](LICENSE) 许可证。

## 🙏 致谢

感谢所有为这个项目做出贡献的开发者们！

---

**让局域网共享更简单！** 🚀