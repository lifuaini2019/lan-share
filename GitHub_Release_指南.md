# 📦 GitHub Release 创建指南

## 🎯 Release 创建步骤

### 1. 访问GitHub仓库
- 打开：https://github.com/lifuaini2019/lan-share
- 点击右侧的 "Releases" 或访问：https://github.com/lifuaini2019/lan-share/releases

### 2. 创建新Release
- 点击 "Create a new release" 按钮
- 或者点击 "Draft a new release"

### 3. 填写Release信息

#### Tag版本
```
v2.1.0
```

#### Release标题
```
🚀 祖宇字文共享 v2.1.0 - 智能局域网检测版本
```

#### Release描述
```markdown
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
curl -fsSL https://raw.githubusercontent.com/lifuaini2019/lan-share/main/install_wankeyun.sh | sudo bash
```

#### 飞牛OS一键安装
```bash
curl -fsSL https://raw.githubusercontent.com/lifuaini2019/lan-share/main/install_flynas.sh | sudo bash
```

### 🎯 使用场景

**典型场景**: 用户在家中局域网环境，习惯性通过绑定的域名访问服务，系统检测到后会智能提示：

> "检测到您在局域网环境，是否切换到局域网地址以获得更快的访问速度？"

**优势**:
- ⚡ **更快速度**: 局域网访问不经过公网，延迟更低
- 🛡️ **更稳定**: 避免外网网络波动影响  
- 🔒 **更安全**: 数据传输不经过公网

### 🐳 Docker镜像

专为飞牛OS优化的Docker镜像已包含在仓库中：
- **文件**: `zuyu-share-flynas-arm64.tar` (8.2MB)
- **架构**: ARM64
- **使用**: 下载后使用 `docker load` 导入

### 📖 详细文档

- [智能局域网检测功能说明](智能局域网检测功能说明.md)
- [一键安装指南](一键安装指南.md)
- [程序上传使用指南](程序上传使用指南.md)
- [快速使用说明](快速使用说明.md)

### 🆘 故障排除

详细的故障排除指南请参考仓库中的相关文档。

---

**🎉 享受更智能、更便捷的局域网共享体验！**

### 📦 下载说明

由于GitHub对单个文件大小的限制，编译好的程序文件已直接包含在仓库的 `build/` 目录中。您可以：

1. **Clone整个仓库**: `git clone https://github.com/lifuaini2019/lan-share.git`
2. **直接下载单个文件**: 点击仓库中 `build/` 目录下的对应文件
3. **使用一键安装脚本**: 自动下载和配置

### 🔗 快速链接

- [源码仓库](https://github.com/lifuaini2019/lan-share)
- [一键安装 - 玩客云](https://raw.githubusercontent.com/lifuaini2019/lan-share/main/install_wankeyun.sh)
- [一键安装 - 飞牛OS](https://raw.githubusercontent.com/lifuaini2019/lan-share/main/install_flynas.sh)
```

---

## 🎯 Release创建注意事项

### ✅ 要做的事情
1. **选择正确的标签**: 使用 `v2.1.0`
2. **设为最新版本**: 勾选 "Set as the latest release"
3. **不上传额外文件**: 程序文件已在仓库的 `build/` 目录中
4. **预览发布**: 使用 "Preview" 功能检查格式

### ❌ 不要做的事情
1. **不要上传大文件**: GitHub有100MB单文件限制
2. **不要重复上传**: 程序文件已在仓库中
3. **不要设为预发布**: 除非是测试版本

### 📝 发布后的操作
1. **更新README**: 确保README中的下载链接正确
2. **测试下载**: 验证所有下载链接可用
3. **宣传推广**: 在相关社区分享新版本
```
