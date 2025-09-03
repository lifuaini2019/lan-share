@echo off
echo 🚀 开始构建 LAN-to-go 多平台版本...

:: 创建构建目录
if not exist build mkdir build
cd build

:: 清理之前的构建文件
del /Q *

echo 📦 正在构建各平台版本...

:: Windows 64位 (exe)
echo 🖥️  构建 Windows 64位版本...
set GOOS=windows
set GOARCH=amd64
go build -ldflags "-s -w" -o lan-to-go-windows-amd64.exe ../main.go
if %errorlevel%==0 (
    echo ✅ Windows 64位版本构建成功: lan-to-go-windows-amd64.exe
) else (
    echo ❌ Windows 64位版本构建失败
)

:: Windows 32位 (exe)
echo 🖥️  构建 Windows 32位版本...
set GOOS=windows
set GOARCH=386
go build -ldflags "-s -w" -o lan-to-go-windows-386.exe ../main.go
if %errorlevel%==0 (
    echo ✅ Windows 32位版本构建成功: lan-to-go-windows-386.exe
) else (
    echo ❌ Windows 32位版本构建失败
)

:: 玩客云路由器 ARM v7
echo 📡 构建玩客云路由器版本 (ARM v7)...
set GOOS=linux
set GOARCH=arm
set GOARM=7
go build -ldflags "-s -w" -o lan-to-go-openwrt-armv7 ../main.go
if %errorlevel%==0 (
    echo ✅ 玩客云路由器版本构建成功: lan-to-go-openwrt-armv7
) else (
    echo ❌ 玩客云路由器版本构建失败
)

:: Linux 64位
echo 🐧 构建 Linux 64位版本...
set GOOS=linux
set GOARCH=amd64
go build -ldflags "-s -w" -o lan-to-go-linux-amd64 ../main.go
if %errorlevel%==0 (
    echo ✅ Linux 64位版本构建成功: lan-to-go-linux-amd64
) else (
    echo ❌ Linux 64位版本构建失败
)

echo.
echo 📁 构建完成！生成的文件：
dir /B

echo.
echo 🎯 部署说明：
echo 📱 Windows电脑: 使用 lan-to-go-windows-amd64.exe
echo 📡 玩客云路由器: 使用 lan-to-go-openwrt-armv7
echo 🐧 Linux服务器: 使用 lan-to-go-linux-amd64

echo.
echo ⚠️  注意事项：
echo 1. 运行程序前请确保 templates/ 和 static/ 目录存在
echo 2. 玩客云路由器可能需要设置可执行权限: chmod +x lan-to-go-openwrt-armv7
echo 3. 程序默认端口为 9405，请确保防火墙允许该端口

echo.
echo ✅ 构建脚本执行完成！
pause