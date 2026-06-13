package har

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"time"
)

// 错误定义
var (
	// ErrInvalidHar 表示HAR对象缺少必要字段
	ErrInvalidHar = NewValidationError("HAR对象缺少必要字段", "")

	// ErrInvalidURL 表示HAR条目中的URL无效
	ErrInvalidURL = NewValidationError("HAR条目中包含无效URL", "")

	// ErrNotJsonContent 表示内容不是JSON格式
	ErrNotJsonContent = NewInvalidFormatError("内容不是JSON格式")
)

// ParseHarFile 解析HAR格式的文件
//
// ParseHarFile是ParseHar的便捷包装，它会先读取文件内容再进行解析。
// 该函数遵循错误处理最佳实践，将所有错误转换为HarError类型，便于统一处理。
//
// 示例:
//
//	har, err := ParseHarFile("example.har")
//	if err != nil {
//	    log.Fatalf("解析HAR文件失败: %v", err)
//	}
func ParseHarFile(harFilePath string) (*Har, error) {
	harFileBytes, err := os.ReadFile(harFilePath)
	if err != nil {
		return nil, NewFileSystemError(fmt.Sprintf("无法读取文件 '%s'", harFilePath), err)
	}
	return ParseHar(harFileBytes)
}

// ParseHar 解析HAR格式的字节数据
//
// ParseHar函数将HAR格式的字节数据解析为Har结构体对象。
// 该函数会进行完整的验证，确保HAR对象满足规范要求。
//
// 示例:
//
//	harBytes, _ := ioutil.ReadFile("example.har")
//	har, err := ParseHar(harBytes)
//	if err != nil {
//	    log.Fatalf("解析HAR数据失败: %v", err)
//	}
func ParseHar(harFileBytes []byte) (*Har, error) {
	// 检查输入是否为空
	if len(harFileBytes) == 0 {
		return nil, NewInvalidFormatError("输入为空")
	}

	// 检查是否是JSON格式
	if !isJSONContent(harFileBytes) {
		return nil, ErrNotJsonContent
	}

	// 解析JSON
	har := new(Har)
	err := json.Unmarshal(harFileBytes, har)
	if err != nil {
		return nil, WrapJSONUnmarshalError(err)
	}

	// 验证HAR对象
	if err := ValidateHarFile(har); err != nil {
		return nil, err
	}

	return har, nil
}

// Har 表示HTTP归档(HAR)文件的主结构
//
// Har结构是HAR格式的根对象，包含一个Log字段。
// 所有HAR数据都包含在Log字段中。
type Har struct {
	Log          Log          `json:"log"` // HAR日志对象
	CustomFields CustomFields `json:"-"`
}

// Log 表示HAR日志对象
//
// Log包含HAR数据的主要部分，包括版本、创建者信息、
// 页面信息和HTTP条目数据。
type Log struct {
	Version      string       `json:"version"`           // HAR规范版本
	Creator      Creator      `json:"creator"`           // 创建工具信息
	Browser      Browser      `json:"browser,omitempty"` // 浏览器信息（可选）
	Pages        []Pages      `json:"pages,omitempty"`   // 页面信息
	Entries      []Entries    `json:"entries"`           // HTTP请求/响应条目
	Comment      string       `json:"comment,omitempty"` // 可选注释
	CustomFields CustomFields `json:"-"`
}

// Creator 表示创建HAR文件的工具信息
type Creator struct {
	Name    string `json:"name"`              // 创建工具名称
	Version string `json:"version"`           // 创建工具版本
	Comment string `json:"comment,omitempty"` // 可选注释
}

// Browser 表示浏览器信息
type Browser struct {
	Name    string `json:"name"`              // 浏览器名称
	Version string `json:"version"`           // 浏览器版本
	Comment string `json:"comment,omitempty"` // 可选注释
}

// PageTimings 表示页面加载计时
type PageTimings struct {
	OnContentLoad float64 `json:"onContentLoad"`     // DOMContentLoaded事件触发时间(ms)
	OnLoad        float64 `json:"onLoad"`            // load事件触发时间(ms)
	Comment       string  `json:"comment,omitempty"` // 可选注释
}

// Pages 表示HAR文件中的页面信息
type Pages struct {
	StartedDateTime time.Time    `json:"startedDateTime"`       // 页面加载开始时间
	ID              string       `json:"id"`                    // 页面唯一标识
	Title           string       `json:"title"`                 // 页面标题
	PageTimings     PageTimings  `json:"pageTimings"`           // 页面加载计时
	Comment         string       `json:"comment,omitempty"`     // 可选注释
	CustomFields    CustomFields `json:"-"`
}

// Headers 表示HTTP头部
type Headers struct {
	Name    string `json:"name"`              // 头部名称
	Value   string `json:"value"`             // 头部值
	Comment string `json:"comment,omitempty"` // 可选注释
}

// QueryString 表示URL查询参数
type QueryString struct {
	Name    string `json:"name"`              // 参数名称
	Value   string `json:"value"`             // 参数值
	Comment string `json:"comment,omitempty"` // 可选注释
}

// Cookie 表示HTTP Cookie
type Cookie struct {
	Name         string       `json:"name"`               // Cookie名称
	Value        string       `json:"value"`              // Cookie值
	Path         string       `json:"path,omitempty"`     // Cookie路径
	Domain       string       `json:"domain,omitempty"`   // Cookie域
	Expires      time.Time    `json:"expires,omitempty"`  // 过期时间
	HTTPOnly     bool         `json:"httpOnly,omitempty"` // 是否为HttpOnly
	Secure       bool         `json:"secure,omitempty"`   // 是否为Secure
	SameSite     string       `json:"sameSite,omitempty"` // SameSite策略
	Comment      string       `json:"comment,omitempty"`  // 可选注释
	CustomFields CustomFields `json:"-"`
}

// PostData 表示HTTP请求的POST数据
type PostData struct {
	MimeType     string       `json:"mimeType"`             // MIME类型
	Params       []Param      `json:"params,omitempty"`     // 参数列表（表单提交时使用）
	Text         string       `json:"text,omitempty"`       // 请求体文本内容
	Comment      string       `json:"comment,omitempty"`    // 可选注释
	CustomFields CustomFields `json:"-"`
}

// Param 表示POST请求中的表单参数
type Param struct {
	Name         string       `json:"name"`                    // 参数名称
	Value        string       `json:"value,omitempty"`         // 参数值
	FileName     string       `json:"fileName,omitempty"`      // 文件名（用于文件上传）
	ContentType  string       `json:"contentType,omitempty"`   // 内容类型
	Comment      string       `json:"comment,omitempty"`       // 可选注释
	CustomFields CustomFields `json:"-"`
}

// Content 表示HTTP响应内容
type Content struct {
	Size         int          `json:"size"`                      // 内容大小(字节)
	MimeType     string       `json:"mimeType"`                  // MIME类型
	Compression  int          `json:"compression,omitempty"`     // 压缩节省字节数(可选)
	Text         string       `json:"text,omitempty"`            // 文本内容(可选)
	Encoding     string       `json:"encoding,omitempty"`        // 编码方式(可选，如base64)
	Comment      string       `json:"comment,omitempty"`         // 可选注释
	CustomFields CustomFields `json:"-"`
}

// Request 表示HTTP请求
type Request struct {
	Method       string        `json:"method"`                 // HTTP方法(GET, POST等)
	URL          string        `json:"url"`                    // 请求URL
	HTTPVersion  string        `json:"httpVersion"`            // HTTP版本
	Cookies      []Cookie      `json:"cookies"`                // Cookie列表
	Headers      []Headers     `json:"headers"`                // 头部列表
	QueryString  []QueryString `json:"queryString"`            // 查询参数
	PostData     *PostData     `json:"postData,omitempty"`     // POST数据(可选)
	HeadersSize  int           `json:"headersSize"`            // 头部大小(字节)
	BodySize     int           `json:"bodySize"`               // 请求体大小(字节)
	Comment      string        `json:"comment,omitempty"`      // 可选注释
	CustomFields CustomFields  `json:"-"`
}

// Response 表示HTTP响应
type Response struct {
	Status       int          `json:"status"`                  // 状态码
	StatusText   string       `json:"statusText"`              // 状态描述
	HTTPVersion  string       `json:"httpVersion"`             // HTTP版本
	Cookies      []Cookie     `json:"cookies"`                 // Cookie列表
	Headers      []Headers    `json:"headers"`                 // 头部列表
	Content      Content      `json:"content"`                 // 响应内容
	RedirectURL  string       `json:"redirectURL"`             // 重定向URL
	HeadersSize  int          `json:"headersSize"`             // 头部大小(字节)
	BodySize     int          `json:"bodySize"`                // 响应体大小(字节)
	TransferSize int          `json:"_transferSize,omitempty"` // 传输大小(Chrome扩展)
	Error        any          `json:"_error,omitempty"`        // 错误信息(Chrome扩展)
	Comment      string       `json:"comment,omitempty"`       // 可选注释
	CustomFields CustomFields `json:"-"`
}

// BeforeRequest 表示请求前的缓存状态
type BeforeRequest struct {
	Expires      time.Time    `json:"expires,omitempty"` // 过期时间
	LastAccess   time.Time    `json:"lastAccess"`        // 最后访问时间
	ETag         string       `json:"eTag"`              // ETag
	HitCount     int          `json:"hitCount"`          // 命中次数
	Comment      string       `json:"comment,omitempty"` // 注释（可选）
	CustomFields CustomFields `json:"-"`
}

// AfterRequest 表示请求后的缓存状态
type AfterRequest struct {
	Expires      time.Time    `json:"expires,omitempty"` // 过期时间
	LastAccess   time.Time    `json:"lastAccess"`        // 最后访问时间
	ETag         string       `json:"eTag"`              // ETag
	HitCount     int          `json:"hitCount"`          // 命中次数
	Comment      string       `json:"comment,omitempty"` // 注释（可选）
	CustomFields CustomFields `json:"-"`
}

// Cache 表示HTTP缓存信息
type Cache struct {
	BeforeRequest *BeforeRequest `json:"beforeRequest,omitempty"` // 请求前缓存状态
	AfterRequest  *AfterRequest  `json:"afterRequest,omitempty"`  // 请求后缓存状态
	Comment       string         `json:"comment,omitempty"`       // 注释
	CustomFields  CustomFields   `json:"-"`
}

// Timings 表示HTTP请求/响应过程中的时间指标
type Timings struct {
	Blocked         float64      `json:"blocked"`                     // 阻塞时间(ms)
	DNS             float64      `json:"dns"`                         // DNS解析时间(ms)
	Connect         float64      `json:"connect"`                     // TCP连接时间(ms)
	Ssl             float64      `json:"ssl"`                         // SSL/TLS协商时间(ms)
	Send            float64      `json:"send"`                        // 发送请求时间(ms)
	Wait            float64      `json:"wait"`                        // 等待响应时间(ms)
	Receive         float64      `json:"receive"`                     // 接收响应时间(ms)
	BlockedQueueing float64      `json:"_blocked_queueing,omitempty"` // 排队阻塞时间(Chrome扩展, ms)
	BlockedProxy    float64      `json:"_blocked_proxy,omitempty"`    // 代理阻塞时间(Chrome扩展, ms)
	Comment         string       `json:"comment,omitempty"`           // 可选注释
	CustomFields    CustomFields `json:"-"`
}

// Entries 表示HAR文件中的单个HTTP请求/响应条目
type Entries struct {
	StartedDateTime time.Time    `json:"startedDateTime"`           // 请求开始时间
	Time            float64      `json:"time"`                      // 总耗时(ms)
	Request         Request      `json:"request"`                   // 请求信息
	Response        Response     `json:"response"`                  // 响应信息
	Cache           Cache        `json:"cache"`                     // 缓存信息
	Timings         Timings      `json:"timings"`                   // 详细计时
	Pageref         string       `json:"pageref,omitempty"`         // 关联的页面ID
	ServerIPAddress string       `json:"serverIPAddress,omitempty"` // 服务器IP
	Connection      string       `json:"connection,omitempty"`      // 连接ID
	Initiator       Initiator    `json:"_initiator,omitempty"`      // 请求发起者(Chrome扩展)
	Priority        string       `json:"_priority,omitempty"`       // 请求优先级(Chrome扩展)
	ResourceType    string       `json:"_resourceType,omitempty"`   // 资源类型(Chrome扩展)
	Comment         string       `json:"comment,omitempty"`         // 可选注释
	CustomFields    CustomFields `json:"-"`
}

// Initiator 表示请求发起者(Chrome DevTools扩展)
type Initiator struct {
	Type       string `json:"type"`       // 发起类型
	URL        string `json:"url"`        // 发起URL
	LineNumber int    `json:"lineNumber"` // 代码行号
	Stack      Stack  `json:"stack"`      // 调用栈
}

// Stack 表示调用栈(Chrome DevTools扩展)
type Stack struct {
	CallFrames []CallFrame `json:"callFrames"` // 调用帧
	Parent     Parent      `json:"parent"`     // 父级调用栈
}

// Parent 表示父级调用栈(Chrome DevTools扩展)
type Parent struct {
	Parent      *Parent     `json:"parent"`      // 嵌套父级
	Description string      `json:"description"` // 描述
	CallFrames  []CallFrame `json:"callFrames"`  // 调用帧
	ParentID    ParentID    `json:"parentId"`    // 父级ID
}

// ParentID 表示父级ID(Chrome DevTools扩展)
type ParentID struct {
	ID         string `json:"id"`         // ID
	DebuggerID string `json:"debuggerId"` // 调试器ID
}

// CallFrame 表示调用帧(Chrome DevTools扩展)
type CallFrame struct {
	FunctionName string `json:"functionName"` // 函数名
	ScriptID     string `json:"scriptId"`     // 脚本ID
	URL          string `json:"url"`          // URL
	LineNumber   int    `json:"lineNumber"`   // 行号
	ColumnNumber int    `json:"columnNumber"` // 列号
}


// IsValidURL 检查URL是否有效
//
// 该函数检查给定的URL字符串是否符合URL规范。
// 返回true表示URL有效，false表示无效。
func IsValidURL(rawURL string) bool {
	_, err := url.Parse(rawURL)
	return err == nil
}
