package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"jd_material_push/internal/config"
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
			Width:  1000,
			Height: 700,
			IconId: 2,
		},
	})
	defer w.Destroy()

	w.Navigate(url)
	w.Run()

	// 窗口关闭后停止服务器
	server.Stop()
}
