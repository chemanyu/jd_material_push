package logic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"jd_material_push/internal/svc"
	"jd_material_push/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadFilesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadFilesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadFilesLogic {
	return &UploadFilesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadFilesLogic) UploadFiles(req *types.UploadRequest) (resp *types.UploadResponse, err error) {
	resp = &types.UploadResponse{
		Code:    200,
		Message: "success",
		Data:    []types.UploadResult{},
	}

	// 从内存中获取京橙平台的 Cookie
	cookie, err := l.svcCtx.CookieManager.GetCookie()
	if err != nil {
		l.Errorf("获取京橙平台 Cookie 失败: %v", err)
		resp.Code = 500
		resp.Message = fmt.Sprintf("获取京橙平台 Cookie 失败: %v", err)
		return resp, nil
	}

	// 读取文件夹下的所有文件
	files, err := os.ReadDir(req.FolderPath)
	if err != nil {
		l.Errorf("读取文件夹失败: %v", err)
		resp.Code = 500
		resp.Message = fmt.Sprintf("读取文件夹失败: %v", err)
		return resp, nil
	}

	// 收集需要上传的文件
	var filesToUpload []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fileName := file.Name()
		// 跳过隐藏文件
		if len(fileName) > 0 && fileName[0] == '.' {
			continue
		}
		filePath := filepath.Join(req.FolderPath, fileName)
		filesToUpload = append(filesToUpload, filePath)
	}

	if len(filesToUpload) == 0 {
		resp.Message = "没有找到可上传的文件"
		return resp, nil
	}

	l.Infof("准备上传 %d 个文件", len(filesToUpload))

	// 使用协程并发上传文件
	var wg sync.WaitGroup
	var mu sync.Mutex
	maxConcurrent := 10 // 最大并发数
	semaphore := make(chan struct{}, maxConcurrent)

	for _, filePath := range filesToUpload {
		wg.Add(1)
		go func(fp string) {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			fileName := filepath.Base(fp)
			result := l.uploadSingleFile(fp, fileName, cookie)

			// 安全地添加到结果列表
			mu.Lock()
			resp.Data = append(resp.Data, result)
			mu.Unlock()
		}(filePath)
	}

	// 等待所有上传完成
	wg.Wait()
	l.Infof("所有文件上传完成，成功: %d, 总数: %d", countSuccessful(resp.Data), len(resp.Data))

	return resp, nil
}

// countSuccessful 统计成功上传的文件数量
func countSuccessful(results []types.UploadResult) int {
	count := 0
	for _, r := range results {
		if r.Success {
			count++
		}
	}
	return count
}

// uploadSingleFile 上传单个文件到京橙平台
func (l *UploadFilesLogic) uploadSingleFile(filePath, fileName, cookie string) types.UploadResult {
	result := types.UploadResult{
		FileName: fileName,
		Success:  false,
	}

	// 读取文件
	fileData, err := os.Open(filePath)
	if err != nil {
		result.ErrorMsg = fmt.Sprintf("打开文件失败: %v", err)
		l.Errorf("打开文件失败 %s: %v", fileName, err)
		return result
	}
	defer fileData.Close()

	// 获取文件信息
	fileInfo, err := fileData.Stat()
	if err != nil {
		result.ErrorMsg = fmt.Sprintf("获取文件信息失败: %v", err)
		l.Errorf("获取文件信息失败 %s: %v", fileName, err)
		return result
	}
	result.FileSize = fileInfo.Size()

	// 创建 multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加表单字段
	_ = writer.WriteField("systemCode", "jdOrange")
	_ = writer.WriteField("businessCode", "伙伴计划--美数科技")

	// 添加文件
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		result.ErrorMsg = fmt.Sprintf("创建文件表单失败: %v", err)
		l.Errorf("创建文件表单失败 %s: %v", fileName, err)
		return result
	}

	_, err = io.Copy(part, fileData)
	if err != nil {
		result.ErrorMsg = fmt.Sprintf("复制文件数据失败: %v", err)
		l.Errorf("复制文件数据失败 %s: %v", fileName, err)
		return result
	}

	err = writer.Close()
	if err != nil {
		result.ErrorMsg = fmt.Sprintf("关闭 writer 失败: %v", err)
		l.Errorf("关闭 writer 失败 %s: %v", fileName, err)
		return result
	}

	// 创建 HTTP 请求
	apiURL := "https://dlupload.jd.com/common/upload/uploadFile"
	httpReq, err := http.NewRequest("POST", apiURL, body)
	if err != nil {
		result.ErrorMsg = fmt.Sprintf("创建请求失败: %v", err)
		l.Errorf("创建请求失败 %s: %v", fileName, err)
		return result
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	httpReq.Header.Set("Cookie", cookie)

	// 发送请求
	client := &http.Client{}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		result.ErrorMsg = fmt.Sprintf("发送请求失败: %v", err)
		l.Errorf("发送请求失败 %s: %v", fileName, err)
		return result
	}
	defer httpResp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		result.ErrorMsg = fmt.Sprintf("读取响应失败: %v", err)
		l.Errorf("读取响应失败 %s: %v", fileName, err)
		return result
	}

	// 解析响应
	var jcResp types.JingchengUploadResponse
	if err := json.Unmarshal(respBody, &jcResp); err != nil {
		result.ErrorMsg = fmt.Sprintf("解析响应失败: %v, 响应内容: %s", err, string(respBody))
		l.Errorf("解析响应失败 %s: %v, 响应: %s", fileName, err, string(respBody))
		return result
	}

	// 检查响应状态
	if jcResp.Code != 200 {
		result.ErrorMsg = fmt.Sprintf("上传失败: %s", jcResp.Message)
		l.Errorf("上传失败 %s: code=%d, message=%s", fileName, jcResp.Code, jcResp.Message)
		return result
	}

	// 上传成功
	result.Success = true
	result.URL = jcResp.Result.URL
	result.LocalURL = jcResp.Result.LocalURL
	l.Infof("上传成功 %s", fileName)

	return result
}
