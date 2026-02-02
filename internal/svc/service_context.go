// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"jd_material_push/internal/config"
	"jd_material_push/internal/cookie"
)

type ServiceContext struct {
	Config        config.Config
	CookieManager *cookie.Manager
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化 Cookie 管理器
	cookieMgr := cookie.NewManager()

	return &ServiceContext{
		Config:        c,
		CookieManager: cookieMgr,
	}
}
