package logic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

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

	// 构建 applyAttr
	columnEnum := buildColumnEnum(req.MediaList, req.CategoryList)
	applyAttr := map[string]interface{}{
		"diyColumns": []map[string]interface{}{
			{
				"isRequired": true,
				"columnType": 2,
				"length":     30,
				"isMultiple": 1,
				"label":      "投放媒体",
				"value":      req.MediaList,
				"columnEnum": columnEnum["media"],
				"key":        "media",
			},
			{
				"isRequired": true,
				"columnType": 2,
				"length":     30,
				"isMultiple": 1,
				"label":      "素材所属品类",
				"value":      req.CategoryList,
				"columnEnum": columnEnum["category"],
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
				"value":      req.ReleaseCopy,
			},
		},
	}
	log.Println("media:", req.MediaList)
	log.Println("category:", req.CategoryList)

	applyAttrJSON, _ := json.Marshal(applyAttr)

	// 构建请求参数
	param := map[string]interface{}{
		"funName": "extAddMaterial",
		"param": map[string]interface{}{
			"isApproval":   1,
			"materialList": req.MaterialList,
			"systemCode":   "jdOrange",
			"businessCode": "伙伴计划--美数科技",
			"applyAttr":    string(applyAttrJSON),
		},
		"loginType": "3",
	}

	bodyJSON, _ := json.Marshal(param)

	// 获取 Cookie
	cookie, err := l.svcCtx.CookieManager.GetCookie()
	if err != nil {
		return nil, fmt.Errorf("获取 Cookie 失败: %w", err)
	}

	// 构建表单数据
	formData := url.Values{}
	formData.Set("appid", "materialCenter")
	formData.Set("functionId", "material_center_api")
	formData.Set("_", "1770018106")
	formData.Set("loginType", "3")
	formData.Set("body", string(bodyJSON))

	// 创建请求
	apiURL := "https://api.m.jd.com/?functionId=material_center_api&appid=materialCenter"
	httpReq, err := http.NewRequest("POST", apiURL, bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return nil, err
	}

	// 设置请求头
	httpReq.Header.Set("Cookie", cookie)
	httpReq.Header.Set("Origin", "https://jcheng.jd.com")
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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

// buildColumnEnum 构建 columnEnum 数据
func buildColumnEnum(mediaList, categoryList []string) map[string]interface{} {
	// 投放媒体选项
	mediaOptions := []map[string]string{
		{"label": "巨量引擎", "value": "jlyq"},
		{"label": "巨量星图", "value": "jlxt"},
		{"label": "快手磁力智投", "value": "ksclzt"},
		{"label": "快手磁力聚星", "value": "kscljx"},
		{"label": "百度营销", "value": "bdyx"},
		{"label": "广点通", "value": "gdt"},
		{"label": "B站", "value": "bz"},
		{"label": "趣头条", "value": "qtt"},
	}

	// 素材品类选项
	categoryOptions := []map[string]string{
		{"label": "本地生活/旅游出行", "value": "4938"},
		{"label": "家庭清洁/纸品", "value": "15901"},
		{"label": "鲜花/奢侈品", "value": "1672"},
		{"label": "数码", "value": "652"},
		{"label": "家用电器", "value": "737"},
		{"label": "食品饮料", "value": "1320"},
		{"label": "厨具", "value": "6196"},
		{"label": "美妆护肤", "value": "1316"},
		{"label": "手机通讯", "value": "9987"},
		{"label": "服饰内衣", "value": "1315"},
		{"label": "生活日用", "value": "1620"},
		{"label": "个人护理", "value": "16750"},
		{"label": "鞋靴", "value": "11729"},
		{"label": "电脑、办公", "value": "670"},
		{"label": "运动户外", "value": "1318"},
		{"label": "生鲜", "value": "12218"},
		{"label": "母婴", "value": "1319"},
	}

	return map[string]interface{}{
		"media":    mediaOptions,
		"category": categoryOptions,
	}
}
