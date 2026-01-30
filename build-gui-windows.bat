@echo off
chcp 65001 >nul
echo ====================================
echo 正在构建 Windows GUI 版本...
echo ====================================

REM 设置发布路径
set RELEASE_PATH=.\release

REM 创建发布文件夹
echo.
echo 创建发布文件夹...
if exist "%RELEASE_PATH%" rd /s /q "%RELEASE_PATH%"
mkdir "%RELEASE_PATH%\etc"
mkdir "%RELEASE_PATH%\static"

REM 构建 Windows GUI 版本
echo.
echo [1/1] 编译 Windows GUI 版本...
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=1
REM 使用 -H=windowsgui 隐藏命令行窗口
go build -ldflags="-s -w -H=windowsgui" -o "%RELEASE_PATH%\filemanager-gui.exe" filemanager-gui.go

if %errorlevel% equ 0 (
    set WINDOWS_GUI_SUCCESS=1
    echo ✓ Windows GUI 版本编译成功
) else (
    set WINDOWS_GUI_SUCCESS=0
    echo ✗ Windows GUI 版本编译失败
)

REM 复制配置文件
echo.
echo 复制配置文件...
copy etc\filemanager-api.yaml "%RELEASE_PATH%\etc\" >nul
copy static\index.html "%RELEASE_PATH%\static\" >nul

REM 创建使用说明
echo.
echo 创建使用说明...
(
echo 文件管理器使用说明
echo.
echo ==================================================
echo Windows GUI 版本 ^(独立窗口^)
echo ==================================================
echo.
echo 【使用方法】
echo 1. 双击运行 filemanager-gui.exe
echo 2. 会弹出一个独立的小窗口
echo 3. 在输入框中输入文件夹路径，例如：C:\Users
echo 4. 点击 获取文件列表 按钮查看文件
echo 5. 关闭窗口即可退出程序
echo.
echo ==================================================
echo 注意事项
echo ==================================================
echo - 关闭窗口即可退出程序
echo - 可能需要安装 WebView2 运行时
echo - 首次运行可能需要允许防火墙访问
echo.
echo ==================================================
echo 系统要求
echo ==================================================
echo - Windows 10/11 64位
echo - WebView2 运行时 Windows 11 已内置
echo - 如果无法运行，请访问以下链接下载 WebView2
echo   https://developer.microsoft.com/microsoft-edge/webview2/
) > "%RELEASE_PATH%\使用说明.txt"

REM 显示构建结果
echo.
echo ====================================
echo ✓ 构建完成！
echo ====================================
echo.
echo 可执行文件位置: %RELEASE_PATH%
echo.
if %WINDOWS_GUI_SUCCESS% equ 1 (
    echo ✓ Windows GUI 版本构建成功：
    echo   filemanager-gui.exe
) else (
    echo ✗ 构建失败，请检查错误信息
)
echo.
echo 您可以将 release 文件夹打包分享给其他人使用
echo.
pause
