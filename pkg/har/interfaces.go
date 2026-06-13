package har

import "time"

// HARProvider 定义所有HAR实现应实现的接口
type HARProvider interface {
	// GetVersion 获取HAR版本
	GetVersion() string

	// GetCreator 获取创建者信息
	GetCreator() Creator

	// GetBrowser 获取浏览器信息
	GetBrowser() Browser

	// GetEntries 获取所有条目
	GetEntries() []EntryProvider

	// GetPages 获取所有页面
	GetPages() []PageProvider

	// ToStandard 转换为标准HAR对象
	ToStandard() *Har
}

// EntryProvider 定义单个Entry的接口
type EntryProvider interface {
	// GetStartedDateTime 获取开始时间
	GetStartedDateTime() time.Time

	// GetTime 获取总时长
	GetTime() float64

	// GetRequest 获取请求信息
	GetRequest() RequestProvider

	// GetResponse 获取响应信息
	GetResponse() ResponseProvider

	// GetTimings 获取计时信息
	GetTimings() TimingsProvider

	// GetPageref 获取页面引用
	GetPageref() string

	// ToStandard 转换为标准Entry对象
	ToStandard() Entries
}

// RequestProvider 定义请求的接口
type RequestProvider interface {
	// GetMethod 获取HTTP方法
	GetMethod() string

	// GetURL 获取URL
	GetURL() string

	// GetHTTPVersion 获取HTTP版本
	GetHTTPVersion() string

	// GetHeaders 获取头部信息
	GetHeaders() []HeaderProvider

	// GetCookies 获取Cookie信息
	GetCookies() []CookieProvider

	// GetQueryString 获取查询参数
	GetQueryString() []QueryString

	// GetPostData 获取POST数据
	GetPostData() *PostData

	// GetBodySize 获取请求体大小
	GetBodySize() int

	// GetHeadersSize 获取头部大小
	GetHeadersSize() int

	// ToStandard 转换为标准Request对象
	ToStandard() Request
}

// ResponseProvider 定义响应的接口
type ResponseProvider interface {
	// GetStatus 获取状态码
	GetStatus() int

	// GetStatusText 获取状态文本
	GetStatusText() string

	// GetHTTPVersion 获取HTTP版本
	GetHTTPVersion() string

	// GetHeaders 获取头部信息
	GetHeaders() []HeaderProvider

	// GetCookies 获取Cookie信息
	GetCookies() []CookieProvider

	// GetContent 获取内容
	GetContent() ContentProvider

	// GetBodySize 获取响应体大小
	GetBodySize() int

	// GetHeadersSize 获取头部大小
	GetHeadersSize() int

	// ToStandard 转换为标准Response对象
	ToStandard() Response
}

// HeaderProvider 定义HTTP头部的接口
type HeaderProvider interface {
	// GetName 获取名称
	GetName() string

	// GetValue 获取值
	GetValue() string

	// ToStandard 转换为标准Header对象
	ToStandard() Headers
}

// CookieProvider 定义Cookie的接口
type CookieProvider interface {
	// GetName 获取名称
	GetName() string

	// GetValue 获取值
	GetValue() string

	// GetDomain 获取域
	GetDomain() string

	// GetPath 获取路径
	GetPath() string

	// GetExpires 获取过期时间
	GetExpires() time.Time

	// IsHTTPOnly 是否为HTTPOnly
	IsHTTPOnly() bool

	// IsSecure 是否为Secure
	IsSecure() bool

	// GetSameSite 获取SameSite值
	GetSameSite() string

	// ToStandard 转换为标准Cookie对象
	ToStandard() Cookie
}

// ContentProvider 定义内容的接口
type ContentProvider interface {
	// GetSize 获取大小
	GetSize() int

	// GetMimeType 获取MIME类型
	GetMimeType() string

	// GetText 获取文本内容（如果有）
	GetText() string

	// GetEncoding 获取编码（如果有）
	GetEncoding() string

	// GetCompression 获取压缩节省字节数
	GetCompression() int

	// ToStandard 转换为标准Content对象
	ToStandard() Content
}

// TimingsProvider 定义计时信息的接口
type TimingsProvider interface {
	// GetBlocked 获取被阻塞时间
	GetBlocked() float64

	// GetDNS 获取DNS解析时间
	GetDNS() float64

	// GetConnect 获取连接时间
	GetConnect() float64

	// GetSend 获取发送时间
	GetSend() float64

	// GetWait 获取等待时间
	GetWait() float64

	// GetReceive 获取接收时间
	GetReceive() float64

	// GetSSL 获取SSL握手时间
	GetSSL() float64

	// ToStandard 转换为标准Timings对象
	ToStandard() Timings
}

// PageProvider 定义页面的接口
type PageProvider interface {
	// GetID 获取ID
	GetID() string

	// GetTitle 获取标题
	GetTitle() string

	// GetStartedDateTime 获取开始时间
	GetStartedDateTime() time.Time

	// GetPageTimings 获取页面计时信息
	GetPageTimings() PageTimingsProvider

	// ToStandard 转换为标准Page对象
	ToStandard() Pages
}

// PageTimingsProvider 定义页面计时信息的接口
type PageTimingsProvider interface {
	// GetOnContentLoad 获取内容加载时间
	GetOnContentLoad() float64

	// GetOnLoad 获取页面加载时间
	GetOnLoad() float64

	// ToStandard 转换为标准PageTimings对象
	ToStandard() PageTimings
}
