package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// Client 通用 HTTP 客户端
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Headers    map[string]string
}

// NewClient 创建一个新的 HTTP 客户端
func NewClient(baseURL string, timeout int) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
		Headers: make(map[string]string),
	}
}

// SetHeader 设置默认请求头
func (c *Client) SetHeader(key, value string) {
	c.Headers[key] = value
}

// SetHeaders 批量设置默认请求头
func (c *Client) SetHeaders(headers map[string]string) {
	for k, v := range headers {
		c.Headers[k] = v
	}
}

// Get 发送 GET 请求
// ctx: 上下文
// path: 请求路径（相对于 BaseURL）
// params: URL 查询参数（可选）
// result: 响应结果会被解析到这个对象中
func (c *Client) Get(ctx context.Context, path string, params map[string]string, result interface{}) error {
	return c.DoRequest(ctx, http.MethodGet, path, params, nil, result)
}

// Post 发送 POST 请求
// ctx: 上下文
// path: 请求路径（相对于 BaseURL）
// body: 请求体（会被序列化为 JSON）
// result: 响应结果会被解析到这个对象中
func (c *Client) Post(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.DoRequest(ctx, http.MethodPost, path, nil, body, result)
}

// DoRequest 执行 HTTP 请求（通用方法）
// ctx: 上下文
// method: HTTP 方法（GET, POST, PUT, DELETE 等）
// path: 请求路径（相对于 BaseURL）
// params: URL 查询参数（可选，主要用于 GET 请求）
// body: 请求体（可选，会被序列化为 JSON）
// result: 响应结果会被解析到这个对象中
func (c *Client) DoRequest(ctx context.Context, method, path string, params map[string]string, body interface{}, result interface{}) error {
	// 构建完整 URL
	url := c.BaseURL + path

	// 处理请求体
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body failed: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	// 设置默认请求头
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}

	// 如果有请求体，设置 Content-Type
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// 添加查询参数
	if len(params) > 0 {
		q := req.URL.Query()
		for k, v := range params {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	// 记录请求日志
	//logx.WithContext(ctx).Infof("HTTP Request: %s %s", method, req.URL.String())

	// 发送请求
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body failed: %w", err)
	}

	// 记录响应日志
	//logx.WithContext(ctx).Infof("HTTP Response: status=%d", resp.StatusCode)

	// 检查 HTTP 状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("http request failed with status code %d: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应体
	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response body failed: %w, body: %s", err, string(respBody))
		}
	}

	return nil
}

// DownloadFile 下载文件
// ctx: 上下文
// path: 请求路径（相对于 BaseURL）
// body: 请求体（可选，会被序列化为 JSON）
// savePath: 文件保存路径
func (c *Client) DownloadFile(ctx context.Context, path string, body interface{}, savePath string) error {
	// 构建完整 URL
	url := c.BaseURL + path

	// 处理请求体
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body failed: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, reqBody)
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	// 设置默认请求头
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}

	// 如果有请求体，设置 Content-Type
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// 记录请求日志
	logx.WithContext(ctx).Infof("HTTP Download Request: %s", req.URL.String())

	// 发送请求
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	// 检查 HTTP 状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("http request failed with status code %d", resp.StatusCode)
	}

	// 创建文件
	file, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("create file failed: %w", err)
	}
	defer file.Close()

	// 写入文件
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("write file failed: %w", err)
	}

	logx.WithContext(ctx).Infof("File downloaded successfully: %s", savePath)
	return nil
}
