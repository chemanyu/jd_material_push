package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"jd_material_push/internal/config"
	"jd_material_push/internal/folder"
	"jd_material_push/internal/handler"
	"jd_material_push/internal/svc"

	"github.com/jchv/go-webview2"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

//go:embed static/index.html
var staticFiles embed.FS

//go:embed etc/filemanager-api.yaml
var configContent []byte

var configFile = flag.String("f", "etc/filemanager-api.yaml", "the config file")

type FileInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	IsDir   bool   `json:"isDir"`
	ModTime string `json:"modTime"`
}

func main() {
	flag.Parse()

	var c config.Config

	// 尝试从嵌入的文件加载配置
	if err := conf.LoadFromYamlBytes(configContent, &c); err != nil {
		// 如果失败，从文件系统加载
		conf.MustLoad(*configFile, &c)
	}

	// 使用随机可用端口
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	// 启动后端服务
	server := rest.MustNewServer(rest.RestConf{
		Host: "127.0.0.1",
		Port: port,
	})

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	// 在后台启动服务器
	go func() {
		server.Start()
	}()

	// 等待服务器启动
	time.Sleep(500 * time.Millisecond)

	// 创建并显示窗口
	url := fmt.Sprintf("http://127.0.0.1:%d", port)

	w := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug:     false,
		AutoFocus: true,
		WindowOptions: webview2.WindowOptions{
			Title:  "文件管理器",
			Width:  uint(1000),
			Height: uint(700),
			IconId: uint(2),
		},
	})
	defer w.Destroy()

	// 绑定文件夹选择函数
	w.Bind("selectFolder", func() string {
		// 使用 Windows 原生文件夹选择对话框
		selectedPath, err := folder.SelectFolder()
		if err != nil {
			log.Printf("文件夹选择失败: %v", err)
			return "{\"error\": \"" + err.Error() + "\"}"
		}

		if selectedPath == "" {
			// 用户取消选择
			return "{\"cancelled\": true}"
		}

		// 扫描文件夹并返回文件列表
		files := scanFolder(selectedPath)
		result := map[string]interface{}{
			"path":  selectedPath,
			"files": files,
		}
		jsonData, _ := json.Marshal(result)
		return string(jsonData)
	})

	w.Navigate(url)
	w.Run()

	// 窗口关闭后停止服务器
	server.Stop()
}

// scanFolder 扫描文件夹并返回文件信息
func scanFolder(folderPath string) []FileInfo {
	var files []FileInfo

	entries, err := os.ReadDir(folderPath)
	if err != nil {
		log.Printf("读取文件夹失败: %v", err)
		return files
	}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		fullPath := filepath.Join(folderPath, entry.Name())
		fileData := FileInfo{
			Name:    entry.Name(),
			Path:    fullPath,
			Size:    info.Size(),
			IsDir:   entry.IsDir(),
			ModTime: info.ModTime().Format(time.RFC3339),
		}
		files = append(files, fileData)
	}

	return files
}
