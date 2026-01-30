#!/bin/bash

echo "===================================="
echo "正在构建多平台可执行文件..."
echo "===================================="

# 构建 Windows 版本
echo ""
echo "[1/4] 编译 Windows 版本..."
export GOOS=windows
export GOARCH=amd64
export CGO_ENABLED=0
go build -ldflags="-s -w" -o filemanager.exe filemanager.go

if [ $? -ne 0 ]; then
    echo ""
    echo "❌ Windows 版本构建失败！"
    exit 1
fi

# 构建 macOS 版本
echo "[2/4] 编译 macOS 版本..."
export GOOS=darwin
export GOARCH=amd64
go build -ldflags="-s -w" -o filemanager-mac filemanager.go

if [ $? -ne 0 ]; then
    echo ""
    echo "❌ macOS 版本构建失败！"
    exit 1
fi

# 设置发布路径
RELEASE_PATH="./release"

# 创建发布文件夹
echo "[3/4] 创建发布文件夹..."
rm -rf "$RELEASE_PATH"
mkdir -p "$RELEASE_PATH/etc"
mkdir -p "$RELEASE_PATH/static"

# 复制文件
echo "[4/4] 复制必要文件..."
cp filemanager.exe "$RELEASE_PATH/"
cp filemanager-mac "$RELEASE_PATH/"
chmod +x "$RELEASE_PATH/filemanager-mac"
cp etc/filemanager-api.yaml "$RELEASE_PATH/etc/"
cp static/index.html "$RELEASE_PATH/static/"

# 创建启动说明
cat > "$RELEASE_PATH/使用说明.txt" << EOF
文件管理器使用说明

【Windows 用户】
1. 双击运行 filemanager.exe
2. 程序会自动打开浏览器窗口
3. 在输入框中输入文件夹路径，例如：C:\Users
4. 点击"获取文件列表"按钮查看文件

【macOS 用户】
1. 打开终端，进入此文件夹
2. 运行命令：./filemanager-mac -f etc/filemanager-api.yaml
3. 程序会自动打开浏览器窗口
4. 在输入框中输入文件夹路径，例如：/Users/username/Documents
5. 点击"获取文件列表"按钮查看文件

注意：关闭浏览器不会关闭程序，需要在命令行窗口按 Ctrl+C 退出
EOF

echo ""
echo "===================================="
echo "✅ 构建完成！"
echo "===================================="
echo ""
echo "可执行文件位置:"
echo "  Windows: $RELEASE_PATH/filemanager.exe"
echo "  macOS:   $RELEASE_PATH/filemanager-mac"
echo ""
echo "您可以将桌面上的 release 文件夹打包分享给其他人使用"
echo ""
