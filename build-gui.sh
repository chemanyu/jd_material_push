#!/bin/bash

echo "===================================="
echo "正在构建 GUI 版本（独立窗口）..."
echo "===================================="

# 设置发布路径
RELEASE_PATH="/Users/chemanyu/Desktop/release"

# 创建发布文件夹
echo ""
echo "创建发布文件夹..."
rm -rf "$RELEASE_PATH"
mkdir -p "$RELEASE_PATH/etc"

# 构建 Windows GUI 版本
echo ""
echo "[1/2] 编译 Windows GUI 版本..."
export GOOS=windows
export GOARCH=amd64
export CGO_ENABLED=1
export CC=x86_64-w64-mingw32-gcc
go build -ldflags="-s -w" -o "$RELEASE_PATH/filemanager-gui.exe" filemanager-gui.go

WINDOWS_GUI_SUCCESS=$?

# 构建 macOS GUI 版本
echo "[2/2] 编译 macOS GUI 版本..."
export GOOS=darwin
export GOARCH=amd64
export CGO_ENABLED=1
export CC=gcc
go build -ldflags="-s -w" -o "$RELEASE_PATH/filemanager-gui-mac" filemanager-gui.go

MACOS_GUI_SUCCESS=$?

# 设置 macOS GUI 可执行权限
if [ $MACOS_GUI_SUCCESS -eq 0 ]; then
    chmod +x "$RELEASE_PATH/filemanager-gui-mac"
fi

# 复制配置文件
echo ""
echo "复制配置文件..."

cp etc/filemanager-api.yaml "$RELEASE_PATH/etc/"

# 创建启动说明
cat > "$RELEASE_PATH/使用说明.txt" << EOF
文件管理器使用说明

==================================================
独立窗口版本（无需浏览器）
==================================================

【Windows 用户】
1. 双击运行 filemanager-gui.exe
2. 会弹出一个独立的小窗口
3. 在输入框中输入文件夹路径，例如：C:\Users
4. 点击"获取文件列表"按钮查看文件
5. 关闭窗口即可退出程序

【macOS 用户】
1. 打开终端，进入此文件夹
2. 运行命令：./filemanager-gui-mac
3. 会弹出一个独立的小窗口
4. 在输入框中输入文件夹路径，例如：/Users/username/Documents
5. 点击"获取文件列表"按钮查看文件
6. 关闭窗口即可退出程序

==================================================
注意事项
==================================================
- 关闭窗口即可退出程序
- 使用 Fyne GUI 框架，无需额外安装组件
- macOS 版本首次运行可能需要在"系统偏好设置 > 安全性与隐私"中允许运行
EOF

echo ""
echo "===================================="
echo "✅ 构建完成！"
echo "===================================="
echo ""
echo "可执行文件位置: $RELEASE_PATH"
echo ""
if [ $WINDOWS_GUI_SUCCESS -eq 0 ] && [ $MACOS_GUI_SUCCESS -eq 0 ]; then
    echo "✅ 所有版本构建成功："
    echo "  Windows: filemanager-gui.exe"
    echo "  macOS:   filemanager-gui-mac"
elif [ $WINDOWS_GUI_SUCCESS -eq 0 ]; then
    echo "✅ Windows GUI 版本构建成功"
    echo "⚠️  macOS GUI 版本构建失败"
elif [ $MACOS_GUI_SUCCESS -eq 0 ]; then
    echo "✅ macOS GUI 版本构建成功"
    echo "⚠️  Windows GUI 版本构建失败（可能需要 MinGW）"
else
    echo "❌ 构建失败"
fi
echo ""
echo "您可以将桌面上的 release 文件夹打包分享给其他人使用"
echo ""
