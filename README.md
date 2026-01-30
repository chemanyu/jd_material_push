# 文件管理服务

这是一个基于 go-zero 框架开发的文件管理服务，提供获取本地文件夹中所有文件的 API 接口。

## 项目结构

```
.
├── filemanager.api           # API 定义文件
├── filemanager.go            # 主程序入口
├── etc/
│   └── filemanager-api.yaml  # 配置文件
└── internal/
    ├── config/               # 配置结构体
    ├── handler/              # HTTP 处理器
    ├── logic/                # 业务逻辑
    ├── svc/                  # 服务上下文
    └── types/                # 类型定义
```

## 功能说明

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

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 启动服务

```bash
go run filemanager.go -f etc/filemanager-api.yaml
```

### 3. 测试接口

```bash
# 获取指定文件夹中的文件列表
curl "http://localhost:8888/api/files?path=/Users/chemanyu/go-zero/jd_material_push"
```

## 配置说明

配置文件 `etc/filemanager-api.yaml`:
- `Name`: 服务名称
- `Host`: 监听地址 (0.0.0.0 表示监听所有网卡)
- `Port`: 监听端口 (默认 8888)
- `Timeout`: 请求超时时间(毫秒)

## 错误码说明

- `200`: 成功
- `400`: 参数错误（路径为空或不是目录）
- `404`: 路径不存在
- `500`: 服务器内部错误
# jd_material_push
