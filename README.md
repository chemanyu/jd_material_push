# 文件管理服务

这是一个基于 go-zero 框架开发的文件管理服务，提供获取本地文件夹中所有文件的 API 接口，并附带图形化界面。

## 项目结构

```
.
├── filemanager.api           # API 定义文件
├── filemanager.go            # 主程序入口
├── build-windows.sh          # macOS/Linux 构建脚本
├── build-windows.bat         # Windows 构建脚本
├── etc/
│   └── filemanager-api.yaml  # 配置文件
├── static/
│   └── index.html            # 图形化界面
└── internal/
    ├── config/               # 配置结构体
    ├── handler/              # HTTP 处理器
    ├── logic/                # 业务逻辑
    ├── svc/                  # 服务上下文
    └── types/                # 类型定义
```

## 功能说明

### 图形界面
程序启动后会自动打开浏览器窗口，提供友好的图形化操作界面：
- 输入文件夹路径查询文件列表
- 显示文件/文件夹图标
- 显示文件大小、修改时间等详细信息
- 自动统计文件和文件夹数量

### API 接口

**获取文件列表**
- 接口路径: `GET /api/files`
- 请求参数:
  - `path` (string): 要查询的文件夹路径
- 响应格式:
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "name": "example.txt",
      "path": "/path/to/folder/example.txt",
      "size": 1024,
      "isDir": false,
      "modTime": "2026-01-30T10:30:00Z"
    }
  ]
}
```

## 使用方法

### 方式一：直接运行（开发模式）

#### 1. 安装依赖

```bash
go mod tidy
```

#### 2. 启动服务

```bash
go run filemanager.go -f etc/filemanager-api.yaml
```

程序会自动打开浏览器窗口，访问地址：http://localhost:8888

### 方式二：打包成 .exe 文件（分发给他人）

#### macOS/Linux 系统构建：

```bash
./build-windows.sh
```

#### Windows 系统构建：

```cmd
build-windows.bat
```

构建完成后，会在 `release` 文件夹中生成以下文件：
- `filemanager.exe` - 可执行文件
- `etc/filemanager-api.yaml` - 配置文件
- `static/index.html` - 界面文件
- `使用说明.txt` - 使用说明

**将整个 `release` 文件夹打包发送给他人，双击 `filemanager.exe` 即可使用！**

### 方式三：手动构建

```bash
# 构建 Windows 版本
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o filemanager.exe filemanager.go

# 构建 macOS 版本
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o filemanager-mac filemanager.go

# 构建 Linux 版本
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o filemanager-linux filemanager.go
```

## 配置说明

配置文件 `etc/filemanager-api.yaml`:
- `Name`: 服务名称
- `Host`: 监听地址 (0.0.0.0 表示监听所有网卡)
- `Port`: 监听端口 (默认 8888)
- `Timeout`: 请求超时时间(毫秒)

## 使用说明

1. 运行程序后，会自动打开浏览器窗口
2. 在输入框中输入要查询的文件夹路径，例如：
   - Windows: `C:\Users\YourName\Documents`
   - macOS/Linux: `/Users/username/Documents`
3. 点击"获取文件列表"按钮
4. 界面会显示该文件夹下的所有文件和文件夹

## 错误码说明

- `200`: 成功
- `400`: 参数错误（路径为空或不是目录）
- `404`: 路径不存在
- `500`: 服务器内部错误
# jd_material_push
