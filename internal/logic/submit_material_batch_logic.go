package logic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"jd_material_push/internal/account"
	"jd_material_push/internal/svc"
	"jd_material_push/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubmitMaterialBatchLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSubmitMaterialBatchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitMaterialBatchLogic {
	return &SubmitMaterialBatchLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SubmitMaterialBatchLogic) SubmitMaterialBatch(req *types.SubmitMaterialBatchRequest) (resp *types.SubmitMaterialResponse, err error) {
	// 检查素材列表
	if len(req.MaterialList) == 0 {
		return &types.SubmitMaterialResponse{
			Code:    400,
			Message: "素材列表不能为空",
			Result:  false,
		}, nil
	}

	if len(req.MaterialList) > 20 {
		return &types.SubmitMaterialResponse{
			Code:    400,
			Message: "单次最多提交20个素材",
			Result:  false,
		}, nil
	}

	// 解析账号类型，决定提交到哪个平台的素材中心
	accountType := account.Parse(req.AccountType)
	plat := account.MustGet(accountType)

	l.Infof("投放媒体: %v, 素材品类: %v", req.MediaList, req.CategoryList)

	applyAttrJSON, _ := json.Marshal(buildApplyAttr(plat, req.MediaList, req.CategoryList, req.ReleaseCopy))

	// 构建请求参数
	param := map[string]interface{}{
		"funName": "extAddMaterial",
		"param": map[string]interface{}{
			"isApproval":   1,
			"materialList": req.MaterialList,
			"systemCode":   plat.SystemCode,
			"businessCode": plat.BusinessCode,
			"applyAttr":    string(applyAttrJSON),
		},
		"loginType": "3",
	}

	bodyJSON, _ := json.Marshal(param)

	// 获取 Cookie（按账号类型区分：广义新用 / 复投-低活）
	cookie, err := l.svcCtx.CookieManager.GetCookie(accountType)
	if err != nil {
		return nil, fmt.Errorf("获取 %s Cookie 失败: %w", plat.Name, err)
	}
	l.Infof("本次提交使用账号类型: %s (systemCode=%s, businessCode=%s)", plat.Name, plat.SystemCode, plat.BusinessCode)

	// 构建表单数据，参数全部放在 body 里（与浏览器抓包一致）
	formData := url.Values{}
	formData.Set("appid", "materialCenter")
	formData.Set("functionId", "material_center_api")
	formData.Set("_", strconv.FormatInt(time.Now().UnixMilli(), 10))
	formData.Set("loginType", "3")
	formData.Set("body", string(bodyJSON))

	// 创建请求
	apiURL := "https://api.m.jd.com/"
	httpReq, err := http.NewRequest("POST", apiURL, bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return nil, err
	}

	// 设置请求头，与浏览器抓包保持一致（缺 UA / x-requested-with 等头会被风控拦截）
	httpReq.Header.Set("Cookie", cookie)
	httpReq.Header.Set("Origin", plat.Origin)
	if plat.Referer != "" {
		httpReq.Header.Set("Referer", plat.Referer)
	}
	if plat.RefererPage != "" {
		httpReq.Header.Set("x-referer-page", plat.RefererPage)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json, text/plain, */*")
	httpReq.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	httpReq.Header.Set("User-Agent", account.UserAgent)
	httpReq.Header.Set("x-requested-with", "XMLHttpRequest")
	httpReq.Header.Set("x-rp-client", "h5_1.0.0")

	// 发送请求
	client := &http.Client{}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	logx.Infof("素材批量提交响应: %s", string(respBody))

	// 解析响应
	var submitResp types.SubmitMaterialResponse
	if err := json.Unmarshal(respBody, &submitResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v, 响应内容: %s", err, string(respBody))
	}

	return &submitResp, nil
}

// buildApplyAttr 构建 applyAttr，columnEnum 必须是该平台素材中心的完整可选项
func buildApplyAttr(plat account.Platform, mediaList, categoryList []string, releaseCopy string) map[string]interface{} {
	return map[string]interface{}{
		"diyColumns": []map[string]interface{}{
			{
				"isRequired": true,
				"columnType": 2,
				"length":     30,
				"isMultiple": 1,
				"label":      "投放媒体",
				"value":      mediaList,
				"columnEnum": toColumnEnum(plat.MediaOptions),
				"key":        "media",
			},
			{
				"isRequired": true,
				"columnType": 2,
				"length":     30,
				"isMultiple": 1,
				"label":      "素材所属品类",
				"value":      categoryList,
				"columnEnum": toColumnEnum(plat.CategoryOptions),
				"key":        "cate",
			},
			{
				"isRequired": true,
				"columnType": 3,
				"length":     30,
				"isMultiple": 2,
				"label":      "投放文案",
				"columnEnum": []map[string]string{{"value": "使用媒体平台推荐文案"}},
				"key":        "release",
				"value":      releaseCopy,
			},
		},
	}
}

// toColumnEnum 把平台可选项转成素材中心 applyAttr 需要的 columnEnum 结构
func toColumnEnum(options []account.Option) []map[string]string {
	columnEnum := make([]map[string]string, 0, len(options))
	for _, opt := range options {
		columnEnum = append(columnEnum, map[string]string{
			"label": opt.Label,
			"value": opt.Value,
		})
	}
	return columnEnum
}
