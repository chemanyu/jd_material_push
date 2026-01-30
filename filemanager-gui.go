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
	// 设置日志输出到文件（用于调试）
	logFile, err := os.OpenFile("filemanager-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		defer logFile.Close()
		log.SetOutput(logFile)
	}

	// 捕获 panic 并记录到日志
	defer func() {
		if r := recover(); r != nil {
			log.Printf("程序崩溃: %v", r)
			fmt.Printf("程序崩溃: %v\n请查看 filemanager-debug.log 文件\n", r)
			time.Sleep(5 * time.Second) // 给用户时间看到错误
		}
	}()

	log.Println("程序启动...")
	flag.Parse()

	var c config.Config

	// 尝试从嵌入的文件加载配置
	log.Println("加载配置文件...")
	if err := conf.LoadFromYamlBytes(configContent, &c); err != nil {
		log.Printf("从嵌入文件加载配置失败: %v，尝试从文件系统加载", err)
		// 如果失败，从文件系统加载
		if err := conf.Load(*configFile, &c); err != nil {
			log.Fatalf("加载配置文件失败: %v", err)
		}
	}
	log.Println("配置文件加载成功")

	// 使用随机可用端口
	log.Println("申请端口...")
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalf("申请端口失败: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	log.Printf("使用端口: %d", port)

	// 启动后端服务
	log.Println("启动后端服务...")
	server := rest.MustNewServer(rest.RestConf{
		Host: "127.0.0.1",
		Port: port,
	})

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	// 在后台启动服务器
	go func() {
		log.Println("后端服务开始监听...")
		server.Start()
	}()

	// 等待服务器启动
	time.Sleep(500 * time.Millisecond)
	log.Println("后端服务已启动")

	// 创建并显示窗口
	url := fmt.Sprintf("http://127.0.0.1:%d", port)
	log.Printf("准备打开窗口: %s", url)

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
	if w == nil {
		log.Fatal("创建 WebView2 窗口失败，可能缺少 WebView2 运行时")
	}
	defer w.Destroy()
	w.SetTitle("文件管理器")
	w.SetSize(1000, 700, webview2.HintNone)
	log.Println("WebView2 窗口创建成功")

	// 绑定文件夹选择函数
	log.Println("绑定 JavaScript 函数...")
	err = w.Bind("selectFolder", func() string {
		log.Println("用户点击了选择文件夹按钮")
		// 使用 Windows 原生文件夹选择对话框
		selectedPath, err := folder.SelectFolder()
		if err != nil {
			log.Printf("文件夹选择失败: %v", err)
			return "{\"error\": \"" + err.Error() + "\"}"
		}

		if selectedPath == "" {
			log.Println("用户取消了文件夹选择")
			// 用户取消选择
			return "{\"cancelled\": true}"
		}

		log.Printf("用户选择了文件夹: %s", selectedPath)
		// 扫描文件夹并返回文件列表
		files := scanFolder(selectedPath)
		result := map[string]interface{}{
			"path":  selectedPath,
			"files": files,
		}
		jsonData, _ := json.Marshal(result)
		return string(jsonData)
	})
	if err != nil {
		log.Fatalf("绑定函数失败: %v", err)
	}
	log.Println("函数绑定成功")

	log.Printf("导航到页面: %s", url)
	w.Navigate(url)

	log.Println("进入主循环...")
	w.Run()

	log.Println("窗口已关闭，停止服务器...")
	// 窗口关闭后停止服务器
	server.Stop()
	log.Println("程序正常退出")
}

// scanFolder 扫描文件夹并返回文件信息
func scanFolder(folderPath string) []FileInfo {
	log.Printf("开始扫描文件夹: %s", folderPath)
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
