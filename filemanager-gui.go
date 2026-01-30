package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"jd_material_push/internal/config"
	"jd_material_push/internal/handler"
	"jd_material_push/internal/svc"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
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

	// 创建 Fyne 应用
	myApp := app.New()
	myWindow := myApp.NewWindow("文件管理器")
	myWindow.Resize(fyne.NewSize(600, 400))

	// 创建界面元素
	var fileList *widget.List
	var selectedPath string
	var fileInfos []FileInfo

	// 文件列表
	fileList = widget.NewList(
		func() int {
			return len(fileInfos)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			if id < len(fileInfos) {
				fileInfo := fileInfos[id]
				icon := "📄"
				if fileInfo.IsDir {
					icon = "📁"
				}
				label.SetText(fmt.Sprintf("%s %s", icon, fileInfo.Name))
			}
		},
	)

	// 路径标签
	pathLabel := widget.NewLabel("请选择文件夹")

	// 选择文件夹按钮
	selectBtn := widget.NewButton("选择文件夹", func() {
		log.Println("用户点击了选择文件夹按钮")
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				log.Printf("选择文件夹出错: %v", err)
				dialog.ShowError(err, myWindow)
				return
			}
			if uri == nil {
				log.Println("用户取消了选择")
				return
			}

			selectedPath = uri.Path()
			log.Printf("用户选择了文件夹: %s", selectedPath)
			pathLabel.SetText(selectedPath)

			// 扫描文件夹
			fileInfos = scanFolder(selectedPath)
			fileList.Refresh()

			log.Printf("扫描到 %d 个文件/文件夹", len(fileInfos))
		}, myWindow)
	})

	// 提交按钮
	submitBtn := widget.NewButton("提交文件列表", func() {
		if selectedPath == "" {
			dialog.ShowInformation("提示", "请先选择文件夹", myWindow)
			return
		}
		log.Printf("提交文件列表，共 %d 个文件", len(fileInfos))
		dialog.ShowInformation("成功", fmt.Sprintf("已扫描 %d 个文件/文件夹", len(fileInfos)), myWindow)
	})

	// 布局
	content := container.NewBorder(
		container.NewVBox(pathLabel, selectBtn),
		submitBtn,
		nil,
		nil,
		fileList,
	)

	myWindow.SetContent(content)

	// 关闭时停止服务器
	myWindow.SetOnClosed(func() {
		log.Println("窗口已关闭，停止服务器...")
		server.Stop()
		log.Println("程序正常退出")
	})

	log.Println("显示窗口...")
	myWindow.ShowAndRun()
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
