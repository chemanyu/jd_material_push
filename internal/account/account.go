package account

import "github.com/zeromicro/go-zero/core/logx"

// Type 账号类型，不同账号类型对应不同的 Cookie 接口和素材中心业务空间
type Type string

const (
	// Xinyong 广义新用（京橙 jcheng.jd.com）
	Xinyong Type = "xinyong"
	// LowActive 复投-低活（京东联盟 union.jd.com）
	LowActive Type = "lowactive"

	// Default 未指定账号类型时的默认值
	Default = Xinyong
)

// UserAgent 与浏览器抓包保持一致的 UA，京东接口风控会校验该头
const UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/151.0.0.0 Safari/537.36"

// Option 下拉/多选项，对应素材中心 applyAttr 里的 columnEnum
type Option struct {
	Label string
	Value string
}

// Platform 账号类型对应的平台参数
type Platform struct {
	Name            string   // 中文名称
	CookieURL       string   // Cookie 获取接口
	SystemCode      string   // 素材中心 systemCode
	BusinessCode    string   // 素材中心 businessCode
	Origin          string   // 素材中心站点 Origin
	Referer         string   // 素材中心站点 Referer，为空则不设置该请求头
	RefererPage     string   // 素材管理页地址，对应 x-referer-page 头，为空则不设置
	MediaOptions    []Option // 投放媒体可选项（各平台不同）
	CategoryOptions []Option // 素材所属品类可选项（各平台不同）
}

// ordered 账号类型的展示顺序
var ordered = []Type{Xinyong, LowActive}

var platforms = map[Type]Platform{
	Xinyong: {
		Name:            "广义新用",
		CookieURL:       "https://rta.zhltech.net/guangyixinmedia/report/jingcheng/cookie",
		SystemCode:      "jdOrange",
		BusinessCode:    "伙伴计划--美数科技",
		Origin:          "https://jcheng.jd.com",
		MediaOptions:    jdOrangeMediaOptions,
		CategoryOptions: jdOrangeCategoryOptions,
	},
	LowActive: {
		Name:            "复投-低活",
		CookieURL:       "https://rta.zhltech.net/guangyixinmedia/report/jingcheng/futou/cookie",
		SystemCode:      "union",
		BusinessCode:    "美数科技-复投",
		Origin:          "https://union.jd.com",
		Referer:         "https://union.jd.com/",
		RefererPage:     "https://union.jd.com/userTask/materialManage",
		MediaOptions:    unionMediaOptions,
		CategoryOptions: unionCategoryOptions,
	},
}

// 京橙（广义新用）投放媒体
var jdOrangeMediaOptions = []Option{
	{"巨量引擎", "jlyq"},
	{"巨量星图", "jlxt"},
	{"快手磁力智投", "ksclzt"},
	{"快手磁力聚星", "kscljx"},
	{"百度营销", "bdyx"},
	{"广点通", "gdt"},
	{"B站", "bz"},
	{"趣头条", "qtt"},
}

// 京橙（广义新用）素材所属品类
var jdOrangeCategoryOptions = []Option{
	{"本地生活/旅游出行", "4938"},
	{"家庭清洁/纸品", "15901"},
	{"鲜花/奢侈品", "1672"},
	{"数码", "652"},
	{"家用电器", "737"},
	{"食品饮料", "1320"},
	{"厨具", "6196"},
	{"美妆护肤", "1316"},
	{"手机通讯", "9987"},
	{"服饰内衣", "1315"},
	{"生活日用", "1620"},
	{"个人护理", "16750"},
	{"鞋靴", "11729"},
	{"电脑、办公", "670"},
	{"运动户外", "1318"},
	{"生鲜", "12218"},
	{"母婴", "1319"},
}

// 京东联盟（复投-低活）投放媒体
var unionMediaOptions = []Option{
	{"巨量引擎", "jlyq"},
	{"巨量星图", "jlxt"},
	{"快手磁力智投", "ksclzt"},
	{"快手磁力聚星", "kscljx"},
	{"百度营销", "bdyx"},
	{"广点通", "gdt"},
	{"B站", "bz"},
	{"DSP", "dsp"},
	{"OPPO", "oppo"},
	{"VIVO", "vivo"},
	{"小米", "xm"},
	{"微博", "wb"},
}

// 京东联盟（复投-低活）素材所属品类
var unionCategoryOptions = []Option{
	{"本地生活/旅游出行", "4938"},
	{"家庭清洁/纸品", "15901"},
	{"鲜花/奢侈品", "1672"},
	{"数码", "652"},
	{"家用电器", "737"},
	{"食品饮料", "1320"},
	{"厨具", "6196"},
	{"美妆护肤", "1316"},
	{"手机通讯", "9987"},
	{"服饰内衣", "1315"},
	{"生活日用", "1620"},
	{"个人护理", "16750"},
	{"鞋靴", "11729"},
	{"电脑、办公", "670"},
	{"运动户外", "1318"},
	{"生鲜", "12218"},
	{"母婴", "1319"},
	{"收纳用品", "35025"},
	{"酒类", "12259"},
	{"医疗保健", "9192"},
	{"居家布艺", "34767"},
	{"家居饰品", "35338"},
	{"珠宝首饰", "6144"},
	{"京东服务", "15980"},
	{"箱包皮具", "17329"},
	{"宠物生活", "6994"},
	{"家装建材", "9855"},
	{"医药", "13314"},
	{"玩具乐器", "6233"},
	{"钟表眼镜", "5025"},
	{"二手商品", "13765"},
	{"教育培训", "13678"},
	{"营养保健", "27983"},
	{"图书", "1713"},
	{"床上用品", "34675"},
	{"汽车用品", "6728"},
	{"传统滋补", "27546"},
	{"家具", "9847"},
	{"摩托车/电动车", "27508"},
	{"工业品", "14065"},
	{"拍卖", "12650"},
	{"宠物健康", "31443"},
	{"农资园艺", "12473"},
	{"文娱", "4053"},
	{"家纺", "15248"},
	{"汽车", "12379"},
	{"健康服务", "27156"},
	{"邮币", "13887"},
	{"外卖美食", "31061"},
	{"数字内容", "5272"},
	{"元器件", "33458"},
	{"民俗/非遗", "18528"},
	{"卖家服务", "9669"},
	{"（京喜通）方便速食", "18000"},
	{"艺术品", "15126"},
	{"院端医疗器械", "27338"},
	{"公益", "30746"},
	{"京东金融", "15298"},
	{"二手回收", "19882"},
	// 平台该项名称末尾带一个空格（抓包 form-urlencoded 里是 "非药健康+"，+ 即空格）
	{"非药健康 ", "13720"},
}

// Ordered 返回所有账号类型（按展示顺序）
func Ordered() []Type {
	return append([]Type(nil), ordered...)
}

// Get 返回账号类型对应的平台参数
func Get(t Type) (Platform, bool) {
	p, ok := platforms[t]
	return p, ok
}

// MustGet 返回账号类型对应的平台参数，未知类型回退到默认类型
func MustGet(t Type) Platform {
	if p, ok := platforms[t]; ok {
		return p
	}
	return platforms[Default]
}

// Name 返回账号类型的中文名称
func Name(t Type) string {
	if p, ok := platforms[t]; ok {
		return p.Name
	}
	return string(t)
}

// Parse 解析账号类型，空值或未知值回退到默认类型（广义新用）
func Parse(s string) Type {
	t := Type(s)
	if _, ok := platforms[t]; ok {
		return t
	}
	if s != "" {
		logx.Errorf("未知的账号类型: %s，回退到 %s", s, Name(Default))
	}
	return Default
}
