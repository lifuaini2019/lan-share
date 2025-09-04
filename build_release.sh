#!/bin/bash

# 祖宇字文共享 - 发布构建脚本 v2.1.0
echo "📦 开始构建发布包..."

# 确保构建目录存在
mkdir -p release
cd release

# 清理旧文件
rm -rf *

# 复制构建好的二进制文件
cp ../build/* . 2>/dev/null || echo "⚠️ 未找到构建文件，请先运行 ./build.sh"

# 复制重要文件
cp ../README.md .
cp ../LICENSE .
cp ../RELEASE_v2.1.0.md .
cp ../一键安装指南.md .
cp ../智能局域网检测功能说明.md .

# 复制安装脚本
cp ../install_wankeyun.sh .
cp ../install_flynas.sh .

# 如果有Docker镜像，也复制过来
cp ../zuyu-share-flynas-*.tar . 2>/dev/null || echo "⚠️ 未找到Docker镜像文件"
cp ../flynas_docker_deploy_*.md . 2>/dev/null || echo "⚠️ 未找到Docker部署说明"

echo "✅ 发布包构建完成！"
echo "📁 文件列表："
ls -la

echo ""
echo "🎯 GitHub Release 建议内容："
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
cat << 'EOF'
## 🎉 祖宇字文共享 v2.1.0 - 智能局域网检测版本

### 🌟 重大更新

#### 🧠 智能局域网检测功能
- **自动识别访问环境**: 当用户通过域名访问但实际在局域网环境时，智能提示切换
- **美观提示界面**: 现代化设计的切换提示框，提供清晰的信息展示
- **智能判断逻辑**: 仅在"域名访问+局域网环境"时显示提示，不影响正常使用
- **一键切换**: 流畅的切换体验，1秒内完成地址跳转

#### 📦 部署革新
- **🖥️ 玩客云一键安装**: 自动架构识别，支持ARMv7路由器设备
- **🐳 飞牛OS一键安装**: 支持Docker容器和直接部署两种方式
- **🔧 自动配置**: 自动创建系统服务，设置开机自启动
- **📋 完善文档**: 详细的安装指南和故障排除说明

### 🛠️ 技术亮点

- **新增API**: `/api/lan-check` 接口，返回详细的网络环境信息
- **智能检测**: 页面加载后自动执行网络环境检测
- **日志记录**: 完整的检测日志，便于调试和分析
- **错误处理**: 完善的异常处理机制

### 📱 支持平台

| 平台 | 架构 | 文件名 | 适用场景 |
|------|------|--------|----------|
| **Windows** | x64 | `zuyu-share-windows-amd64.exe` | Windows 10/11 |
| **Windows** | x86 | `zuyu-share-windows-386.exe` | 旧版Windows |
| **Linux** | x64 | `zuyu-share-linux-amd64` | 服务器/桌面 |
| **Linux** | x86 | `zuyu-share-linux-386` | 32位Linux |
| **Linux** | ARM64 | `zuyu-share-linux-arm64` | **飞牛OS/群晖** ⭐ |
| **macOS** | Intel | `zuyu-share-darwin-amd64` | Intel Mac |
| **macOS** | Apple Silicon | `zuyu-share-darwin-arm64` | M1/M2 Mac |
| **OpenWrt** | ARMv7 | `zuyu-share-linux-arm` | **玩客云/路由器** ⭐ |

### 🚀 快速安装

#### 玩客云/路由器一键安装
```bash
curl -fsSL https://raw.githubusercontent.com/username/LAN-Share-Go/main/install_wankeyun.sh | sudo bash
```

#### 飞牛OS一键安装
```bash
curl -fsSL https://raw.githubusercontent.com/username/LAN-Share-Go/main/install_flynas.sh | sudo bash
```

### 🎯 使用场景

**典型场景**: 用户在家中局域网环境，习惯性通过绑定的域名访问服务，系统检测到后会智能提示：

> "检测到您在局域网环境，是否切换到局域网地址以获得更快的访问速度？"

**优势**:
- ⚡ **更快速度**: 局域网访问不经过公网，延迟更低
- 🛡️ **更稳定**: 避免外网网络波动影响  
- 🔒 **更安全**: 数据传输不经过公网

### 📖 详细文档

- [智能局域网检测功能说明](智能局域网检测功能说明.md)
- [一键安装指南](一键安装指南.md)
- [完整版本说明](RELEASE_v2.1.0.md)

### 🆘 故障排除

详细的故障排除指南请参考[一键安装指南](一键安装指南.md)中的相关章节。

---

**🎉 享受更智能、更便捷的局域网共享体验！**
EOF
