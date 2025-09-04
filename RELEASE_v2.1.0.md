# 祖宇字文共享 v2.1.0 发布说明

## 🎉 重大更新：完美支持飞牛OS

本版本专门优化了飞牛OS等NAS系统的支持，解决了Docker架构不匹配问题，并新增多项实用功能。

---

## ✨ 新功能亮点

### 🌐 智能协议检测
- **域名访问**: 自动使用HTTPS协议
- **IP访问**: 自动使用HTTP协议  
- **协议智能**: 根据访问方式自动适配

### 📱 移动端体验升级
- **交互优化**: 点击空白处可收起菜单
- **响应式设计**: 完美适配各种屏幕尺寸
- **触摸友好**: 优化移动设备操作体验

### 🔄 网络状态可视化
- **局域网标识**: IP访问显示"局域网"
- **广域网标识**: 域名访问显示"广域网"
- **实时更新**: 动态显示网络连接状态

### 🐳 Docker完美支持
- **架构修复**: 彻底解决飞牛OS x86_64架构不匹配问题
- **容器优化**: 16MB轻量级镜像，启动迅速
- **自动重启**: 支持系统重启后自动恢复

---

## 🔧 技术改进

### ⚡ 性能优化
- **Go 1.21**: 最新Go语言版本，性能提升
- **内存优化**: 运行时内存占用仅10-15MB
- **启动加速**: 程序启动时间 < 1秒

### 🛠️ 稳定性增强
- **多层级IP获取**: 路由器环境IP检测更准确
- **错误处理**: 完善的异常处理机制
- **日志优化**: 详细的运行状态日志

### 📊 功能完善
- **二维码动态加载**: 解决Go模板安全限制
- **WebSocket优化**: 实时同步更稳定
- **跨域支持**: CORS配置更完善

---

## 📦 支持平台 (8个平台)

| 平台 | 架构 | 文件名 | 适用场景 |
|------|------|--------|----------|
| **Windows** | x64 | `zuyu-share-windows-amd64.exe` | Windows 10/11 |
| **Windows** | x86 | `zuyu-share-windows-386.exe` | 旧版Windows |
| **Linux** | x64 | `zuyu-share-linux-amd64` | 服务器/桌面 |
| **Linux** | x86 | `zuyu-share-linux-386` | 32位Linux |
| **Linux** | ARM64 | `zuyu-share-linux-arm64` | **飞牛OS/群晖** ⭐ |
| **macOS** | Intel | `zuyu-share-darwin-amd64` | Intel Mac |
| **macOS** | Apple Silicon | `zuyu-share-darwin-arm64` | M1/M2 Mac |
| **OpenWrt** | ARMv7 | `zuyu-share-linux-arm` | 路由器/玩客云 |

---

## 🎯 特别优化

### 🌟 飞牛OS专项支持
- ✅ **架构问题**: 完全解决x86_64架构不匹配
- ✅ **容器稳定**: 不再出现重启循环
- ✅ **部署简单**: 一键导入Docker镜像
- ✅ **性能优化**: 针对NAS系统优化

### 🔧 路由器环境优化
- ✅ **IP检测**: 多种策略获取正确局域网IP
- ✅ **资源占用**: 超低内存和CPU占用
- ✅ **自启动**: 支持开机自动启动
- ✅ **OpenWrt**: 完美支持OpenWrt系统

---

## 📥 下载说明

### 推荐下载

| 用户类型 | 推荐版本 | 下载 |
|----------|----------|------|
| **飞牛OS用户** | ARM64版本 | `zuyu-share-v2.1.0-flynas-arm64.tar.gz` |
| **Windows用户** | 64位版本 | `zuyu-share-v2.1.0-windows-amd64.zip` |
| **Linux用户** | 64位版本 | `zuyu-share-v2.1.0-linux-amd64.tar.gz` |
| **macOS用户** | 通用版本 | `zuyu-share-v2.1.0-darwin-universal.tar.gz` |

### 快速部署

#### 飞牛OS Docker部署
```bash
# 解压并导入
tar -xzf zuyu-share-v2.1.0-flynas-arm64.tar.gz
docker run -d --name zuyu-share --restart unless-stopped -p 9405:9405 zuyu-share

# 访问服务
http://飞牛主机IP:9405
```

#### Windows直接运行
```bash
# 解压运行
unzip zuyu-share-v2.1.0-windows-amd64.zip
zuyu-share-windows-amd64.exe

# 访问服务
http://localhost:9405
```

---

## 🛡️ 安全特性

- **局域网专用**: 数据不经过外网，保护隐私
- **跨域安全**: 支持CORS但限制安全范围  
- **实时加密**: WebSocket连接安全可靠
- **无外部依赖**: 单文件运行，无需额外组件

---

## 🚀 使用场景

### 🏠 家庭环境
- 家庭成员间快速文字、文件共享
- 手机与电脑无缝传输
- 多设备实时同步

### 🏢 办公环境  
- 团队内部临时文件分享
- 会议记录实时同步
- 局域网安全传输

### 🖥️ NAS部署
- 飞牛OS/群晖系统集成
- Docker容器化部署
- 自动备份和恢复

---

## 🔄 升级说明

### 从v2.0升级
1. 停止旧版本程序
2. 下载新版本可执行文件
3. 保留原有配置文件
4. 启动新版本即可

### Docker用户升级
```bash
# 停止旧容器
docker stop zuyu-share && docker rm zuyu-share

# 部署新版本
docker run -d --name zuyu-share --restart unless-stopped -p 9405:9405 zuyu-share:v2.1.0
```

---

## 🐛 问题修复

- 修复飞牛OS Docker架构不匹配问题
- 修复路由器环境IP获取不准确
- 修复移动端菜单交互问题
- 修复二维码在某些环境下无法显示
- 修复WebSocket连接偶尔中断

---

## 🙏 致谢

感谢所有用户的反馈和建议，特别是飞牛OS用户帮助测试和优化！

---

## 📞 技术支持

- **GitHub Issues**: [提交问题](../../issues)
- **GitHub Discussions**: [讨论交流](../../discussions)

---

**🎉 享受更稳定、更快速的局域网共享体验！**