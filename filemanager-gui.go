package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"jd_material_push/internal/account"
	"jd_material_push/internal/config"
	"jd_material_push/internal/handler"
	"jd_material_push/internal/svc"
	"jd_material_push/internal/types"

	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

const (
	IconFolder = "[DIR] "
	IconFile   = "[FILE] "
)

//go:embed etc/filemanager-api.yaml
var configContent []byte

//go:embed fonts/NotoSansSC-Regular.ttf
var chineseFont []byte

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

	// 设置自定义主题以支持中文字体（必须在创建任何 widget 之前）
	log.Println("加载中文字体...")
	customTheme := newChineseTheme()
	myApp.Settings().SetTheme(customTheme)
	log.Println("主题设置完成")

	myWindow := myApp.NewWindow("文件管理器")
	log.Println("创建窗口成功")
	myWindow.Resize(fyne.NewSize(800, 600))

	// 创建界面元素
	var fileList *widget.List
	var selectedPath string
	var fileInfos []FileInfo
	var selectedMedia []string
	var selectedCategories []string

	// 账号类型选项（决定使用哪个平台的 Cookie、素材中心业务空间和可选项）
	var accountTypeLabels []string
	accountTypeValues := map[string]string{}
	for _, t := range account.Ordered() {
		label := account.Name(t)
		accountTypeLabels = append(accountTypeLabels, label)
		accountTypeValues[label] = string(t)
	}
	selectedAccountType := accountTypeValues[accountTypeLabels[0]]

	// 选中的媒体显示标签 - 使用多行富文本显示
	selectedMediaLabel := widget.NewRichTextFromMarkdown("**未选择**")
	selectedMediaLabel.Wrapping = fyne.TextWrapWord

	// 选中的品类显示标签 - 使用多行富文本显示
	selectedCategoryLabel := widget.NewRichTextFromMarkdown("**未选择**")
	selectedCategoryLabel.Wrapping = fyne.TextWrapWord

	// currentPlatform 返回当前账号类型对应的平台参数（可选项按平台区分）
	currentPlatform := func() account.Platform {
		return account.MustGet(account.Type(selectedAccountType))
	}

	// refreshSelectedLabel 按平台可选项的顺序刷新已选项展示
	refreshSelectedLabel := func(label *widget.RichText, options []account.Option, selected []string) {
		if len(selected) == 0 {
			label.ParseMarkdown("**未选择**")
			return
		}
		mdText := fmt.Sprintf("**已选择 %d 项：**\n", len(selected))
		i := 0
		for _, opt := range options {
			for _, sel := range selected {
				if opt.Value == sel {
					i++
					mdText += fmt.Sprintf("%d. %s\n", i, opt.Label)
					break
				}
			}
		}
		label.ParseMarkdown(mdText)
	}

	accountTypeRadio := widget.NewRadioGroup(accountTypeLabels, func(label string) {
		if label == "" {
			return
		}
		value := accountTypeValues[label]
		if value == selectedAccountType {
			return
		}
		selectedAccountType = value
		// 各平台的媒体/品类枚举不同，切换后已选项失效，需要重新选择
		selectedMedia = nil
		selectedCategories = nil
		selectedMediaLabel.ParseMarkdown("**未选择**")
		selectedCategoryLabel.ParseMarkdown("**未选择**")
		log.Printf("选择账号类型: %s (%s)，已清空投放媒体和素材品类选择", label, selectedAccountType)
	})
	accountTypeRadio.Horizontal = true
	accountTypeRadio.SetSelected(accountTypeLabels[0])

	// 投放文案输入框
	releaseCopyEntry := widget.NewEntry()
	releaseCopyEntry.SetPlaceHolder("请输入投放文案")
	releaseCopyEntry.SetText("使用媒体平台推荐文案")

	// 限制投放文案输入框的高度
	releaseCopyContainer := container.NewVBox(releaseCopyEntry)
	releaseCopyContainer.Resize(fyne.NewSize(0, 60)) // 限制高度为60像素

	// showOptionDialog 弹出多选对话框，可选项来自当前账号类型对应的平台
	showOptionDialog := func(title, tip string, options []account.Option, selected []string, minHeight float32, onConfirm func([]string)) {
		tempSelected := make(map[string]bool)
		for _, val := range selected {
			tempSelected[val] = true
		}

		content := container.NewVBox(
			widget.NewLabel(tip),
			widget.NewSeparator(),
		)
		// 按平台可选项的顺序创建复选框
		for _, opt := range options {
			value := opt.Value
			check := widget.NewCheck(opt.Label, func(checked bool) {
				if checked {
					tempSelected[value] = true
				} else {
					delete(tempSelected, value)
				}
			})
			check.Checked = tempSelected[value]
			content.Add(check)
		}

		// 创建带滚动的容器，设置最小尺寸
		scrollContent := container.NewVScroll(content)
		scrollContent.SetMinSize(fyne.NewSize(400, minHeight))

		dialog.ShowCustomConfirm(title, "确定", "取消",
			scrollContent,
			func(confirmed bool) {
				if !confirmed {
					return
				}
				// 按平台可选项的顺序回填，保证提交顺序稳定
				result := []string{}
				for _, opt := range options {
					if tempSelected[opt.Value] {
						result = append(result, opt.Value)
					}
				}
				onConfirm(result)
			}, myWindow)
	}

	// 投放媒体选择对话框
	selectMediaBtn := widget.NewButton("选择投放媒体", func() {
		options := currentPlatform().MediaOptions
		showOptionDialog("选择投放媒体", "请选择一个或多个投放媒体平台：", options, selectedMedia, 300,
			func(result []string) {
				selectedMedia = result
				refreshSelectedLabel(selectedMediaLabel, options, selectedMedia)
			})
	})

	// 素材品类选择对话框
	selectCategoryBtn := widget.NewButton("选择素材品类", func() {
		options := currentPlatform().CategoryOptions
		showOptionDialog("选择素材品类", "请选择一个或多个素材品类：", options, selectedCategories, 400,
			func(result []string) {
				selectedCategories = result
				refreshSelectedLabel(selectedCategoryLabel, options, selectedCategories)
			})
	})

	// 文件列表
	fileList = widget.NewList(
		func() int {
			return len(fileInfos)
		},
		func() fyne.CanvasObject {
			// 使用深色文本的 Label
			label := canvas.NewText("", color.RGBA{R: 0, G: 0, B: 0, A: 255})
			label.TextSize = 20
			label.TextStyle = fyne.TextStyle{Bold: true}
			return label
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			label := obj.(*canvas.Text)
			if id < len(fileInfos) {
				fileInfo := fileInfos[id]
				icon := IconFile
				if fileInfo.IsDir {
					icon = IconFolder
				}
				label.Text = fmt.Sprintf("%s%s", icon, fileInfo.Name)
				label.Refresh()
			}
		},
	)

	// 路径标签 - 使用深色文本
	pathLabel := canvas.NewText("请选择文件夹", color.RGBA{R: 0, G: 0, B: 0, A: 255})
	pathLabel.TextSize = 14
	pathLabel.TextStyle = fyne.TextStyle{Bold: true}

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
			pathLabel.Text = selectedPath
			pathLabel.Refresh()
			fileInfos = scanFolder(selectedPath)
			fileList.Refresh()

			log.Printf("扫描到 %d 个文件/文件夹", len(fileInfos))
		}, myWindow)
	})

	// 提交按钮
	submitBtn := widget.NewButton("上传并提交素材", func() {
		if selectedPath == "" {
			dialog.ShowInformation("提示", "请先选择文件夹", myWindow)
			return
		}
		if len(selectedMedia) == 0 {
			dialog.ShowInformation("提示", "请选择投放媒体", myWindow)
			return
		}
		if len(selectedCategories) == 0 {
			dialog.ShowInformation("提示", "请选择素材品类", myWindow)
			return
		}
		if releaseCopyEntry.Text == "" {
			dialog.ShowInformation("提示", "请输入投放文案", myWindow)
			return
		}
		if selectedAccountType == "" {
			dialog.ShowInformation("提示", "请选择账号类型", myWindow)
			return
		}

		log.Printf("开始上传并提交素材，共 %d 个文件，账号类型: %s", len(fileInfos), selectedAccountType)

		// 显示进度对话框
		progressDialog := dialog.NewCustomWithoutButtons("上传中",
			widget.NewProgressBarInfinite(),
			myWindow)
		progressDialog.Show()

		// 在后台上传并提交
		go func() {
			result := uploadAndSubmitMaterial(selectedPath, port, selectedMedia, selectedCategories, releaseCopyEntry.Text, selectedAccountType)

			// 关闭进度对话框并在主线程显示结果
			progressDialog.Hide()
			showUploadResultDialog(result, myWindow)
		}()
	})

	// 布局
	formContent := container.NewVBox(
		widget.NewLabelWithStyle("账号类型:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		accountTypeRadio,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("投放媒体:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewPadded(selectedMediaLabel),
		selectMediaBtn,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("素材品类:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewPadded(selectedCategoryLabel),
		selectCategoryBtn,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("投放文案:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		releaseCopyContainer,
	)

	// 给表单内容添加滚动支持
	formScroll := container.NewVScroll(formContent)
	formScroll.SetMinSize(fyne.NewSize(0, 350)) // 增加最小高度，确保所有选项可见

	content := container.NewBorder(
		container.NewVBox(pathLabel, selectBtn, widget.NewSeparator(), formScroll),
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
		// 过滤掉 .DS_Store 和其他隐藏文件
		if entry.Name() == ".DS_Store" || (len(entry.Name()) > 0 && entry.Name()[0] == '.') {
			log.Printf("跳过隐藏文件: %s", entry.Name())
			continue
		}

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

// 自定义主题以支持中文字体
type chineseTheme struct {
	fyne.Theme
}

// 创建使用嵌入中文字体的主题
func newChineseTheme() fyne.Theme {
	return &chineseTheme{
		Theme: theme.DefaultTheme(),
	}
}

// 重写 Font 方法，为所有文本样式返回中文字体
func (ct *chineseTheme) Font(style fyne.TextStyle) fyne.Resource {
	// NotoSansSC 是可变字体，支持粗细变化，可以处理所有样式
	// 对于 Monospace 等宽字体，也使用中文字体以保证中文显示正常
	return fyne.NewStaticResource("NotoSansSC-Regular.ttf", chineseFont)
}

// uploadAndSubmitMaterial 上传文件并提交素材到京橙平台（批量上传+批量提交）
func uploadAndSubmitMaterial(folderPath string, port int, mediaList, categoryList []string, releaseCopy, accountType string) string {
	log.Printf("开始上传文件夹: %s, 账号类型: %s", folderPath, accountType)

	// 第一步：扫描文件夹获取所有文件
	fileInfos := scanFolder(folderPath)
	if len(fileInfos) == 0 {
		return "# ⚠️ 上传失败\n\n没有找到任何文件"
	}

	// 只处理非目录文件
	var files []FileInfo
	for _, f := range fileInfos {
		if !f.IsDir {
			files = append(files, f)
		}
	}

	if len(files) == 0 {
		return "# ⚠️ 上传失败\n\n没有找到任何可上传的文件"
	}

	log.Printf("找到 %d 个文件，开始上传...", len(files))

	// 第二步：调用一次上传接口，后端会处理文件夹中的所有文件
	reqBody := types.UploadRequest{
		FolderPath:  folderPath,
		AccountType: accountType,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Sprintf("# ⚠️ 上传失败\n\n序列化请求失败: %v", err)
	}

	url := fmt.Sprintf("http://127.0.0.1:%d/api/upload", port)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Sprintf("# ⚠️ 上传失败\n\n发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	var uploadResp types.UploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return fmt.Sprintf("# ⚠️ 上传失败\n\n解析响应失败: %v", err)
	}

	uploadResults := uploadResp.Data

	// 统计上传结果
	successCount := 0
	failCount := 0
	var successResults []types.UploadResult
	var resultDetails string

	for _, result := range uploadResults {
		if result.Success {
			successCount++
			successResults = append(successResults, result)
			sizeStr := formatFileSize(result.FileSize)
			resultDetails += fmt.Sprintf("### ✅ %s\n", result.FileName)
			resultDetails += fmt.Sprintf("- **大小:** %s\n", sizeStr)
		} else {
			failCount++
			resultDetails += fmt.Sprintf("### ❌ %s\n", result.FileName)
			resultDetails += fmt.Sprintf("- **错误:** %s\n\n", result.ErrorMsg)
		}
	}

	// 第三步：批量提交素材（每20条一批）
	var submitResults []types.SubmitMaterialResponse
	batchSize := 20

	if len(successResults) > 0 {
		log.Printf("开始批量提交素材，共 %d 个成功文件", len(successResults))

		for i := 0; i < len(successResults); i += batchSize {
			end := i + batchSize
			if end > len(successResults) {
				end = len(successResults)
			}
			batch := successResults[i:end]

			log.Printf("提交批次 %d: %d-%d/%d", i/batchSize+1, i+1, end, len(successResults))

			// 提交这一批素材
			submitResp := submitMaterialBatch(batch, mediaList, categoryList, releaseCopy, port, accountType)
			submitResults = append(submitResults, submitResp)
		}
	}

	// 统计提交结果
	submitSuccessCount := 0
	submitFailCount := 0
	var submitDetails string

	for idx, submitResp := range submitResults {
		if submitResp.Code == 200 && submitResp.Result {
			submitSuccessCount++
			submitDetails += fmt.Sprintf("### ✅ 批次 %d\n", idx+1)
			submitDetails += fmt.Sprintf("- **状态:** 提交成功\n")
			submitDetails += fmt.Sprintf("- **信息:** %s\n\n", submitResp.Message)
		} else {
			submitFailCount++
			submitDetails += fmt.Sprintf("### ❌ 批次 %d\n", idx+1)
			submitDetails += fmt.Sprintf("- **状态:** 提交失败\n")
			submitDetails += fmt.Sprintf("- **信息:** %s\n\n", submitResp.Message)
		}
	}

	// 构建最终汇总
	summary := fmt.Sprintf("# 📤 上传完成\n\n"+
		"## 📊 统计信息\n"+
		"- **扫描文件:** %d 个\n"+
		"- **成功上传:** %d 个文件\n"+
		"- **失败上传:** %d 个文件\n"+
		"- **提交批次:** %d 批（每批最多%d个）\n"+
		"- **成功批次:** %d 批\n"+
		"- **失败批次:** %d 批\n\n",
		len(files), successCount, failCount,
		len(submitResults), batchSize,
		submitSuccessCount, submitFailCount)

	log.Println(summary)
	return summary
}

// videoExts 素材中心按视频（materialType=2）处理的扩展名
var videoExts = map[string]bool{
	".mp4": true, ".avi": true, ".mov": true,
	".mkv": true, ".webm": true, ".flv": true,
	".wmv": true, ".m4v": true, ".mpeg": true, ".mpg": true,
}

// materialTypeByExt 按扩展名判断素材类型，1=图片 2=视频
func materialTypeByExt(fileName string) int {
	if videoExts[strings.ToLower(filepath.Ext(fileName))] {
		return 2
	}
	return 1
}

// submitMaterialBatch 批量提交素材到素材中心
func submitMaterialBatch(uploadResults []types.UploadResult, mediaList, categoryList []string, releaseCopy string, port int, accountType string) types.SubmitMaterialResponse {
	// 构建素材列表
	var materialList []types.MaterialItem
	for _, result := range uploadResults {
		if result.Success {
			// 优先用上传接口返回的类型，没返回时才按扩展名兜底
			materialType := result.FileType
			if materialType == 0 {
				materialType = materialTypeByExt(result.FileName)
			}

			materialList = append(materialList, types.MaterialItem{
				MaterialName: result.FileName,
				MaterialSize: result.FileSize,
				MaterialType: materialType,
				URL:          result.URL,
				LocalURL:     result.LocalURL,
			})
		}
	}

	if len(materialList) == 0 {
		return types.SubmitMaterialResponse{
			Code:    400,
			Message: "没有可提交的素材",
			Result:  false,
		}
	}

	// 构建请求
	submitURL := fmt.Sprintf("http://127.0.0.1:%d/api/submit-material-batch", port)
	submitReq := map[string]interface{}{
		"materialList": materialList,
		"mediaList":    mediaList,
		"categoryList": categoryList,
		"releaseCopy":  releaseCopy,
		"accountType":  accountType,
	}

	submitData, err := json.Marshal(submitReq)
	if err != nil {
		log.Printf("序列化提交请求失败: %v", err)
		return types.SubmitMaterialResponse{
			Code:    500,
			Message: fmt.Sprintf("序列化提交请求失败: %v", err),
			Result:  false,
		}
	}

	submitResp, err := http.Post(submitURL, "application/json", bytes.NewBuffer(submitData))
	if err != nil {
		log.Printf("发送提交请求失败: %v", err)
		return types.SubmitMaterialResponse{
			Code:    500,
			Message: fmt.Sprintf("发送提交请求失败: %v", err),
			Result:  false,
		}
	}
	defer submitResp.Body.Close()

	var materialResp types.SubmitMaterialResponse
	if err := json.NewDecoder(submitResp.Body).Decode(&materialResp); err != nil {
		log.Printf("解析提交响应失败: %v", err)
		return types.SubmitMaterialResponse{
			Code:    500,
			Message: fmt.Sprintf("解析提交响应失败: %v", err),
			Result:  false,
		}
	}

	return materialResp
}

// formatFileSize 格式化文件大小
func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// showUploadResultDialog 显示上传结果对话框
func showUploadResultDialog(content string, window fyne.Window) {
	// 创建富文本显示，支持Markdown格式
	resultText := widget.NewRichTextFromMarkdown(content)
	resultText.Wrapping = fyne.TextWrapWord

	// 创建滚动容器
	scroll := container.NewScroll(resultText)
	scroll.SetMinSize(fyne.NewSize(700, 500))

	// 创建对话框
	d := dialog.NewCustom("📊 上传结果", "关闭", scroll, window)
	d.Resize(fyne.NewSize(750, 550))
	d.Show()
}
