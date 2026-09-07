package cookie

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"jd_material_push/internal/account"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
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

// cookieEntry 单个账号类型的 Cookie 缓存
type cookieEntry struct {
	cookie     string
	lastUpdate time.Time
}

// Manager Cookie 管理器，按账号类型分别维护 Cookie
type Manager struct {
	cookies map[account.Type]*cookieEntry
	mu      sync.RWMutex
	stopCh  chan struct{}
}

// NewManager 创建 Cookie 管理器
func NewManager() *Manager {
	m := &Manager{
		cookies: make(map[account.Type]*cookieEntry),
		stopCh:  make(chan struct{}),
	}

	// 首次获取各账号类型的 Cookie
	for _, t := range account.Ordered() {
		if err := m.fetchCookie(t); err != nil {
			logx.Errorf("初始化获取 %s Cookie 失败: %v", account.Name(t), err)
			// 仅广义新用有内置的默认 Cookie 兜底
			if t == account.Xinyong {
				m.setCookie(t, DefaultCookie)
				logx.Infof("已设置 %s 默认 Cookie，长度: %d", account.Name(t), len(DefaultCookie))
			}
		}
	}

	// 启动定时刷新
	go m.autoRefresh()

	return m
}

// GetCookie 获取指定账号类型的当前 Cookie
func (m *Manager) GetCookie(t account.Type) (string, error) {
	if _, ok := account.Get(t); !ok {
		return "", fmt.Errorf("未知的账号类型: %s", t)
	}

	m.mu.RLock()
	entry := m.cookies[t]
	m.mu.RUnlock()

	if entry == nil || entry.cookie == "" {
		// 缓存为空时立即补一次，避免必须等到下个刷新周期
		if err := m.fetchCookie(t); err != nil {
			return "", fmt.Errorf("%s Cookie 未初始化: %w", account.Name(t), err)
		}
		m.mu.RLock()
		entry = m.cookies[t]
		m.mu.RUnlock()
	}

	if entry == nil || entry.cookie == "" {
		return "", fmt.Errorf("%s Cookie 未初始化", account.Name(t))
	}

	return entry.cookie, nil
}

// setCookie 更新指定账号类型的 Cookie 缓存
func (m *Manager) setCookie(t account.Type, cookie string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cookies[t] = &cookieEntry{
		cookie:     cookie,
		lastUpdate: time.Now(),
	}
}

// fetchCookie 从接口获取指定账号类型的 Cookie
func (m *Manager) fetchCookie(t account.Type) error {
	plat, ok := account.Get(t)
	if !ok {
		return fmt.Errorf("未知的账号类型: %s", t)
	}

	// 创建请求
	req, err := http.NewRequest(http.MethodGet, plat.CookieURL, nil)
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
	m.setCookie(t, cookieResp.Data)

	logx.Infof("成功获取 %s Cookie，长度: %d", account.Name(t), len(cookieResp.Data))
	return nil
}

// autoRefresh 自动刷新所有账号类型的 Cookie
func (m *Manager) autoRefresh() {
	ticker := time.NewTicker(RefreshInterval)
	defer ticker.Stop()

	types := account.Ordered()
	retryWaits := make(map[account.Type]time.Duration, len(types))
	for _, t := range types {
		retryWaits[t] = InitialRetryWait
	}

	for {
		select {
		case <-ticker.C:
			for _, t := range types {
				retryWait := retryWaits[t]
				if err := m.fetchCookie(t); err != nil {
					logx.Errorf("刷新 %s Cookie 失败: %v，将在 %v 后重试", account.Name(t), err, retryWait)
					// 失败后快速重试
					time.AfterFunc(retryWait, func() {
						if err := m.fetchCookie(t); err != nil {
							logx.Errorf("重试获取 %s Cookie 失败: %v", account.Name(t), err)
						}
					})
					// 指数退避，最多等待 5 分钟
					retryWait *= 2
					if retryWait > 5*time.Minute {
						retryWait = 5 * time.Minute
					}
					retryWaits[t] = retryWait
				} else {
					// 成功后重置重试等待时间
					retryWaits[t] = InitialRetryWait
				}
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

// GetLastUpdateTime 获取指定账号类型上次更新时间
func (m *Manager) GetLastUpdateTime(t account.Type) time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if entry := m.cookies[t]; entry != nil {
		return entry.lastUpdate
	}
	return time.Time{}
}
