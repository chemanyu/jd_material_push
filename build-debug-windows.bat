@echo off
chcp 65001 >nul
echo ====================================
echo 正在构建 Windows 调试版本...
echo ====================================

REM 设置发布路径
set RELEASE_PATH=.\release-debug

REM 创建发布文件夹
echo.
echo 创建发布文件夹...
if exist "%RELEASE_PATH%" rd /s /q "%RELEASE_PATH%"
mkdir "%RELEASE_PATH%\etc"
mkdir "%RELEASE_PATH%\static"

REM 构建 Windows 调试版本（显示控制台窗口）
echo.
echo [1/1] 编译 Windows 调试版本（带控制台窗口）...
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=1
REM 注意：不使用 -H=windowsgui，这样会显示控制台窗口
go build -o "%RELEASE_PATH%\filemanager-gui-debug.exe" filemanager-gui.go

if %errorlevel% equ 0 (
    echo ✓ Windows 调试版本编译成功
    echo.
    echo 输出文件: %RELEASE_PATH%\filemanager-gui-debug.exe
    echo.
    echo 请在命令行中运行此程序以查看错误信息：
    echo cd %RELEASE_PATH%
    echo filemanager-gui-debug.exe
) else (
    echo ✗ Windows 调试版本编译失败
    pause
    exit /b 1
)

REM 复制必需的文件
echo.
echo 复制配置文件和静态文件...
copy "etc\filemanager-api.yaml" "%RELEASE_PATH%\etc\" >nul
copy "static\index.html" "%RELEASE_PATH%\static\" >nul

echo.
echo ====================================
echo 构建完成！
echo ====================================
echo.
echo 调试步骤：
echo 1. 打开命令提示符（cmd）
echo 2. cd %RELEASE_PATH%
echo 3. filemanager-gui-debug.exe
echo 4. 查看控制台输出的错误信息
echo.
pause
