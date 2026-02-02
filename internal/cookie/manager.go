package cookie

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	CookieAPIURL     = "https://rta.zhltech.net/guangyixinmedia/report/jingcheng/cookie"
	RefreshInterval  = 30 * time.Minute // 每30分钟刷新一次
	InitialRetryWait = 10 * time.Second // 初始重试等待时间

	// DefaultCookie 默认 Cookie，当接口获取失败时使用
	DefaultCookie = "__jdu=1740564634752297692458; shshshfpa=1f665855-d40d-4573-911e-1817e2ec065a-1740564636; shshshfpx=1f665855-d40d-4573-911e-1817e2ec065a-1740564636; mba_muid=1740564634752297692458; x-rp-evtoken=mGW9U4qbzsaBdCMe70m9pJbXbFXe6C8Jsw3mKiteWCHGLPn9LYcODlL17XjUR_UaWKgtKm3eF7qnYvc5KxUGpw%3D%3D; webp=1; visitkey=4831893135550762302; __wga=1758537421714.1758537421714.1758537421714.1758537421714.1.1; jdd69fo72b8lfeoe=P7NFMPEEEQ67KDC3QBNPJ6H7WGBY5MCBVTVXAM3OIM2E2NEXEANYMYQWZ5SLM7FUGYNCNOYOEL4KZMS4PEIFMY74SE; focus-login-switch=saas; cn=1; user-key=eea59422-e9ef-4976-9a35-194857cc0799; ceshi3.com=000; app_id=jdsaas; me_fp=fd24709377ac93f9bbbf30ad594895f7; 3AB9D23F7A4B3C9B=BZGES66BKGLDCNSLHTOXJZ7MJQDWBY6YJS3ECPTGI4IMWM6ICIG4W7GVCSKVRUQQ5IUIXAMLW3EECRVYLFESMNVRII; __jdv=209449046|direct|-|none|-|1768790198414; me_js_token=jdd03BZGES66BKGLDCNSLHTOXJZ7MJQDWBY6YJS3ECPTGI4IMWM6ICIG4W7GVCSKVRUQQ5IUIXAMLW3EECRVYLFESMNVRIIAAAAM4A6U2Q3IAAAAACPX2U2U77IPEK4X; me_saas_userInfo=U2FsdGVkX1/Grztr/QtZzOfIoUxcjY8YTZQ7piipEHlJ3qCVTrOqGU6qAoRspMg29qzEOd0WRUYtRQTARugIM5I2tT5duiYK96EoP3HYogU=; focus-token-type=3; focus-team-id=$z9tKQlZAxTS0xOn1WWPz3; 3AB9D23F7A4B3CSS=jdd03BZGES66BKGLDCNSLHTOXJZ7MJQDWBY6YJS3ECPTGI4IMWM6ICIG4W7GVCSKVRUQQ5IUIXAMLW3EECRVYLFESMNVRIIAAAAM4BCQTRHYAAAAACOZANRHTCVGSLYX; shshshfpb=BApXWcq-pC_lAkPZFJkqrxyPXcWBZKreABgIXQRlB9xJ1MiDTo4G28XS-ii2sZNVxIOUOs_OCgH6hR1c; TrackID=1PNlUsnH4YqoknM3IUfnlDeCJfmPJgeaipvixrDCOy8mwGH5rwCaxkxagZs_iuRQqzmVmmc7USbBj2vMpFLz7nzXfEqEiaZibzL10fT-xkN_ibMTN624jzZVq_LBZkSbJhMzbV34HnqlsxU5EXyYm3Q; light_key=AASBKE7rOxgWQziEhC_QY6yawCa-wXDgLajck1ZHNsGbo6nBEu46npFrzqIya8YWmCgq_r8H4TWoWQiNHo-WLnH4clEg1w; pinId=C4k73xwU8lX3I4S8SxuoI4N_gKxl9BLfnRytSPX6E9U; pin=%E4%BC%99%E4%BC%B4%E8%AE%A1%E5%88%92--%E7%BE%8E%E6%95%B0%E7%A7%91%E6%8A%80; unick=14pvh8vw2t7f9i; _tp=Zy0qZGk%2F4LH%2B9ChrciUT9XK0E%2B8dfHfw1Sa7EXGdnhfoQ4EfVxjVOq9LyhsnctfEnfX%2FpjfOsyO4wSl1GGI1Oc6dDxrwtTNjzuCmxS990X0%3D; _pst=%E4%BC%99%E4%BC%B4%E8%AE%A1%E5%88%92--%E7%BE%8E%E6%95%B0%E7%A7%91%E6%8A%80; __jdc=246305811; __jda=246305811.1740564634752297692458.1740564634.1769669793.1769751643.111; thor=B34E7578B0967B8420965A11126652FF608A80FA2406D997FE6015FC3945F1B73AE9A16C7726A2A93E40D9BD3E3E03C4677F083E637DF96A1EE8367F22608B0E926BC73FD64E7486ABEECBC4A3941795237CEA9409A501106FE18173BE840F46AAF7D4F42752FD575EB29D7BB78F7811171F85900437B60D174E91C1B1C54D69192C9ECB231BC2C54D841A20D2C25291; flash=3_gSa1Lw9jX-BTBN06oQjrqOMqCqCbv7mJu475MDnwKRZ-d8nVJC0HzTUovguJwJi2-ccf1b5qD8JV0ro-vGyVE5E_Oj_XnuDKvZ2NhsGwZAOx93Xo--QLRrV4ka8XM-GBsuXwLsU0umb-L5HwIny9_6VD5vYWGw9ge583mJ3mPspqsa98NzQ3y9TJu6rQmjcaKck*; sdtoken=AAbEsBpEIOVjqTAKCQtvQu1741jRdGHelQ2fIad1T8vMMpUfeYVhKq9SIXWbqfx4yKvvn34kc3s3aB_bgHJG81zliiMElnd4WdIZVtRydv6zybhDvCf-MbRmnIVTL314TAHwXqIqqupsYatc13B1F7WG6EIFyJhLCGoXwPhzdFozkNHccR4"
)

// CookieResponse 接口返回数据结构
type CookieResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

// Manager Cookie 管理器
type Manager struct {
	cookie     string
	lastUpdate time.Time
	mu         sync.RWMutex
	stopCh     chan struct{}
}

// NewManager 创建 Cookie 管理器
func NewManager() *Manager {
	m := &Manager{
		stopCh: make(chan struct{}),
	}

	// 首次获取 Cookie
	if err := m.fetchCookie(); err != nil {
		logx.Errorf("初始化获取 Cookie 失败: %v，使用默认 Cookie", err)
		// 使用默认 Cookie
		m.mu.Lock()
		m.cookie = DefaultCookie
		m.lastUpdate = time.Now()
		m.mu.Unlock()
		logx.Infof("已设置默认 Cookie，长度: %d", len(DefaultCookie))
	}

	// 启动定时刷新
	go m.autoRefresh()

	return m
}

// GetCookie 获取当前 Cookie
func (m *Manager) GetCookie() (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.cookie == "" {
		return "", fmt.Errorf("Cookie 未初始化")
	}

	return m.cookie, nil
}

// fetchCookie 从接口获取 Cookie
func (m *Manager) fetchCookie() error {
	// 创建请求
	req, err := http.NewRequest(http.MethodGet, CookieAPIURL, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置 HTTP 基本认证
	req.SetBasicAuth("guangyixin", "*~je,R#(anqAD")

	// 发送请求
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求 Cookie 接口失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	var cookieResp CookieResponse
	if err := json.Unmarshal(body, &cookieResp); err != nil {
		return fmt.Errorf("解析响应失败: %w, 响应: %s", err, string(body))
	}

	if cookieResp.Code != 200 {
		return fmt.Errorf("获取 Cookie 失败: code=%d, message=%s", cookieResp.Code, cookieResp.Message)
	}

	if cookieResp.Data == "" {
		return fmt.Errorf("Cookie 数据为空")
	}

	// 更新 Cookie
	m.mu.Lock()
	m.cookie = cookieResp.Data
	m.lastUpdate = time.Now()
	m.mu.Unlock()

	logx.Infof("成功获取 Cookie，长度: %d", len(cookieResp.Data))
	return nil
}

// autoRefresh 自动刷新 Cookie
func (m *Manager) autoRefresh() {
	ticker := time.NewTicker(RefreshInterval)
	defer ticker.Stop()

	retryWait := InitialRetryWait

	for {
		select {
		case <-ticker.C:
			if err := m.fetchCookie(); err != nil {
				logx.Errorf("刷新 Cookie 失败: %v，将在 %v 后重试", err, retryWait)
				// 失败后快速重试
				time.AfterFunc(retryWait, func() {
					if err := m.fetchCookie(); err != nil {
						logx.Errorf("重试获取 Cookie 失败: %v", err)
					}
				})
				// 指数退避，最多等待 5 分钟
				retryWait *= 2
				if retryWait > 5*time.Minute {
					retryWait = 5 * time.Minute
				}
			} else {
				// 成功后重置重试等待时间
				retryWait = InitialRetryWait
			}
		case <-m.stopCh:
			logx.Info("Cookie 管理器已停止")
			return
		}
	}
}

// Stop 停止自动刷新
func (m *Manager) Stop() {
	close(m.stopCh)
}

// GetLastUpdateTime 获取上次更新时间
func (m *Manager) GetLastUpdateTime() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastUpdate
}
