// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"flag"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"jd_material_push/internal/config"
	"jd_material_push/internal/handler"
	"jd_material_push/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/filemanager-api.yaml", "the config file")

// openBrowser 打开浏览器
func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		fmt.Printf("无法自动打开浏览器: %v\n", err)
		fmt.Printf("请手动访问: %s\n", url)
	}
}

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	// 构建访问地址
	url := fmt.Sprintf("http://localhost:%d", c.Port)
	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	fmt.Printf("访问地址: %s\n", url)

	// 延迟1秒后打开浏览器，确保服务已启动
	go func() {
		time.Sleep(1 * time.Second)
		openBrowser(url)
	}()

	server.Start()
}
