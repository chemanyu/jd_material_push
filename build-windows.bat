@echo off
chcp 65001
echo ====================================
echo 正在构建 Windows 可执行文件...
echo ====================================

:: 设置环境变量
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=0

:: 构建可执行文件
echo.
echo [1/3] 编译 Go 代码...
go build -ldflags="-s -w" -o filemanager.exe filemanager.go

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo ❌ 构建失败！
    pause
    exit /b 1
)

:: 设置发布路径
set RELEASE_PATH=%USERPROFILE%\Desktop\release

:: 创建发布文件夹
echo [2/3] 创建发布文件夹...
if exist "%RELEASE_PATH%" (
    rmdir /s /q "%RELEASE_PATH%"
)
mkdir "%RELEASE_PATH%"
mkdir "%RELEASE_PATH%\etc"
mkdir "%RELEASE_PATH%\static"

:: 复制文件
echo [3/3] 复制必要文件...
copy filemanager.exe "%RELEASE_PATH%\"
copy etc\filemanager-api.yaml "%RELEASE_PATH%\etc\"
copy static\index.html "%RELEASE_PATH%\static\"

:: 创建启动说明
echo 文件管理器使用说明 > "%RELEASE_PATH%\使用说明.txt"
echo. >> "%RELEASE_PATH%\使用说明.txt"
echo 1. 双击运行 filemanager.exe >> "%RELEASE_PATH%\使用说明.txt"
echo 2. 程序会自动打开浏览器窗口 >> "%RELEASE_PATH%\使用说明.txt"
echo 3. 在输入框中输入文件夹路径，例如：C:\Users >> "%RELEASE_PATH%\使用说明.txt"
echo 4. 点击"获取文件列表"按钮查看文件 >> "%RELEASE_PATH%\使用说明.txt"
echo. >> "%RELEASE_PATH%\使用说明.txt"
echo 注意：关闭浏览器不会关闭程序，需要在命令行窗口按 Ctrl+C 退出 >> "%RELEASE_PATH%\使用说明.txt"

echo.
echo ====================================
echo ✅ 构建完成！
echo ====================================
echo.
echo 可执行文件位置: %RELEASE_PATH%\filemanager.exe
echo 您可以将桌面上的 release 文件夹打包分享给其他人使用
echo.
pause
