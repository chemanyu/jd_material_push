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
京东素材批量上传工具使用说明

==================================================
启动程序
==================================================

【Windows 用户】
1. 双击运行 filemanager-gui.exe
2. 会弹出一个独立的应用窗口

【macOS 用户】
1. 双击运行  filemanager-gui-mac
2. 会弹出一个独立的应用窗口
3. 首次运行可能需要在"系统偏好设置 > 安全性与隐私"中允许运行

==================================================
使用步骤
==================================================

1. 选择素材文件夹
   - 点击"选择素材文件夹"按钮
   - 选择包含要上传素材的文件夹（支持图片、视频等格式）
   - 文件列表会自动显示所选文件夹中的所有文件

2. 选择投放媒体
   - 点击"选择投放媒体"按钮
   - 勾选一个或多个投放平台：
     · 巨量引擎、巨量星图
     · 快手磁力智投、快手磁力聚星
     · 百度营销、广点通
     · B站、趣头条

3. 选择素材品类
   - 点击"选择素材品类"按钮
   - 勾选一个或多个品类：
     · 如：美妆护肤、食品饮料、家用电器等

4. 输入投放文案
   - 在"投放文案"输入框中填写文案内容
   - 默认使用"使用媒体平台推荐文案"

5. 上传并提交
   - 确认所有信息无误后，点击"上传并提交素材"按钮
   - 等待上传完成，会显示上传结果

==================================================
注意事项
==================================================
- 请确保已配置 etc/filemanager-api.yaml 中的京东 API 相关参数
- 素材文件夹中不要包含隐藏文件（如 .DS_Store）
- 上传前请确保网络连接正常
- 使用 Fyne GUI 框架，无需额外安装组件
- 关闭窗口即可退出程序
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
