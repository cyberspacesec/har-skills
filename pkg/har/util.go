package har

import (
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

// ========== Har 方法 ==========

// Clone 深拷贝整个Har对象
// 对所有切片（Pages、Entries、Headers、Cookies、QueryString、PostData.Params）进行深拷贝，
// 确保修改克隆对象不会影响原始对象。
func (h *Har) Clone() *Har {
	if h == nil {
		return nil
	}

	data, err := json.Marshal(h)
	if err != nil {
		return nil
	}

	clone := &Har{}
	if err := json.Unmarshal(data, clone); err != nil {
		return nil
	}
	return clone
}

// GetEntryCount 返回HAR中的条目数量
func (h *Har) GetEntryCount() int {
	if h == nil {
		return 0
	}
	return len(h.Log.Entries)
}

// Walk 遍历所有条目，对每个条目调用访问函数
// 如果访问函数返回错误，则停止遍历并返回该错误。
func (h *Har) Walk(fn func(*Entries) error) error {
	if h == nil {
		return nil
	}
	for i := range h.Log.Entries {
		if err := fn(&h.Log.Entries[i]); err != nil {
			return err
		}
	}
	return nil
}

// GetUniqueDomains 返回所有条目URL中唯一域名的排序列表
func (h *Har) GetUniqueDomains() []string {
	if h == nil {
		return nil
	}

	domainSet := make(map[string]bool)
	for _, entry := range h.Log.Entries {
		domain := extractDomain(entry.Request.URL)
		if domain != "" {
			domainSet[domain] = true
		}
	}

	domains := make([]string, 0, len(domainSet))
	for d := range domainSet {
		domains = append(domains, d)
	}
	sort.Strings(domains)
	return domains
}

// Equals 比较两个HAR对象是否相等
// 比较版本、创建者、浏览器、条目数量以及每个条目的方法+URL+状态码。
func (h *Har) Equals(other *Har) bool {
	if h == nil && other == nil {
		return true
	}
	if h == nil || other == nil {
		return false
	}

	// 比较版本
	if h.Log.Version != other.Log.Version {
		return false
	}

	// 比较创建者
	if h.Log.Creator.Name != other.Log.Creator.Name || h.Log.Creator.Version != other.Log.Creator.Version {
		return false
	}

	// 比较浏览器
	if h.Log.Browser.Name != other.Log.Browser.Name || h.Log.Browser.Version != other.Log.Browser.Version {
		return false
	}

	// 比较条目数量
	if len(h.Log.Entries) != len(other.Log.Entries) {
		return false
	}

	// 比较每个条目的方法、URL和状态码
	for i := range h.Log.Entries {
		if h.Log.Entries[i].Request.Method != other.Log.Entries[i].Request.Method {
			return false
		}
		if h.Log.Entries[i].Request.URL != other.Log.Entries[i].Request.URL {
			return false
		}
		if h.Log.Entries[i].Response.Status != other.Log.Entries[i].Response.Status {
			return false
		}
	}

	return true
}

// SaveToFileGzipped 将HAR保存为gzip压缩文件
func (h *Har) SaveToFileGzipped(filePath string, indent bool) error {
	if h == nil {
		return NewInvalidFormatError("HAR对象为空")
	}

	f, err := os.Create(filePath)
	if err != nil {
		return NewFileSystemError(fmt.Sprintf("无法创建文件 '%s'", filePath), err)
	}
	defer f.Close()

	gzWriter := gzip.NewWriter(f)
	defer gzWriter.Close()

	return h.SaveToWriter(gzWriter, indent)
}

// SaveToWriter 将HAR JSON写入io.Writer
func (h *Har) SaveToWriter(w io.Writer, indent bool) error {
	if h == nil {
		return NewInvalidFormatError("HAR对象为空")
	}

	data, err := h.ToJSON(indent)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

// ========== Entries 方法 ==========

// IsError 判断响应是否为错误（状态码 >= 400）
func (e *Entries) IsError() bool {
	if e == nil {
		return false
	}
	return e.Response.Status >= 400
}

// IsRedirect 判断响应是否为重定向（状态码为3xx）
func (e *Entries) IsRedirect() bool {
	if e == nil {
		return false
	}
	return e.Response.Status >= 300 && e.Response.Status < 400
}

// IsSuccess 判断响应是否为成功（状态码为2xx）
func (e *Entries) IsSuccess() bool {
	if e == nil {
		return false
	}
	return e.Response.Status >= 200 && e.Response.Status < 300
}

// GetElapsedTime 将Time字段（毫秒）转换为time.Duration
func (e *Entries) GetElapsedTime() time.Duration {
	if e == nil {
		return 0
	}
	return time.Duration(e.Time * float64(time.Millisecond))
}

// GetURL 解析并返回请求URL
func (e *Entries) GetURL() *url.URL {
	if e == nil {
		return nil
	}
	u, err := url.Parse(e.Request.URL)
	if err != nil {
		return nil
	}
	return u
}

// GetDomain 从请求URL中提取域名
func (e *Entries) GetDomain() string {
	if e == nil {
		return ""
	}
	return extractDomain(e.Request.URL)
}

// GetSize 计算条目的总大小（请求头 + 请求体 + 响应头 + 响应体）
// 对于大小为-1的字段（表示未知），按0计算。
func (e *Entries) GetSize() int {
	if e == nil {
		return 0
	}

	reqHeadersSize := e.Request.HeadersSize
	if reqHeadersSize < 0 {
		reqHeadersSize = 0
	}
	reqBodySize := e.Request.BodySize
	if reqBodySize < 0 {
		reqBodySize = 0
	}
	respHeadersSize := e.Response.HeadersSize
	if respHeadersSize < 0 {
		respHeadersSize = 0
	}
	respBodySize := e.Response.BodySize
	if respBodySize < 0 {
		respBodySize = 0
	}

	return reqHeadersSize + reqBodySize + respHeadersSize + respBodySize
}

// GetRequestBody 获取请求体字节数据（从PostData.Text获取）
func (e *Entries) GetRequestBody() []byte {
	if e == nil || e.Request.PostData == nil {
		return nil
	}
	return []byte(e.Request.PostData.Text)
}

// GetResponseBody 获取解码后的响应体
// 如果内容使用base64编码，则自动解码；否则直接返回文本字节数据。
func (e *Entries) GetResponseBody() ([]byte, error) {
	if e == nil {
		return nil, nil
	}

	text := e.Response.Content.Text
	if text == "" {
		return []byte{}, nil
	}

	// 如果编码为base64，进行解码
	if strings.EqualFold(e.Response.Content.Encoding, "base64") {
		data, err := base64.StdEncoding.DecodeString(text)
		if err != nil {
			return nil, fmt.Errorf("base64解码响应体失败: %w", err)
		}
		return data, nil
	}

	return []byte(text), nil
}

// ========== Request 方法 ==========

// GetHeader 获取指定名称的第一个请求头值（不区分大小写）
func (r *Request) GetHeader(name string) string {
	if r == nil {
		return ""
	}
	for _, h := range r.Headers {
		if strings.EqualFold(h.Name, name) {
			return h.Value
		}
	}
	return ""
}

// GetHeaderValues 获取指定名称的所有请求头值（不区分大小写）
func (r *Request) GetHeaderValues(name string) []string {
	if r == nil {
		return nil
	}
	var values []string
	for _, h := range r.Headers {
		if strings.EqualFold(h.Name, name) {
			values = append(values, h.Value)
		}
	}
	return values
}

// GetCookie 根据名称获取请求Cookie（区分大小写）
func (r *Request) GetCookie(name string) *Cookie {
	if r == nil {
		return nil
	}
	for i := range r.Cookies {
		if r.Cookies[i].Name == name {
			return &r.Cookies[i]
		}
	}
	return nil
}

// HasHeader 检查指定名称的请求头是否存在（不区分大小写）
func (r *Request) HasHeader(name string) bool {
	if r == nil {
		return false
	}
	for _, h := range r.Headers {
		if strings.EqualFold(h.Name, name) {
			return true
		}
	}
	return false
}

// ========== Response 方法 ==========

// GetHeader 获取指定名称的第一个响应头值（不区分大小写）
func (r *Response) GetHeader(name string) string {
	if r == nil {
		return ""
	}
	for _, h := range r.Headers {
		if strings.EqualFold(h.Name, name) {
			return h.Value
		}
	}
	return ""
}

// GetHeaderValues 获取指定名称的所有响应头值（不区分大小写）
func (r *Response) GetHeaderValues(name string) []string {
	if r == nil {
		return nil
	}
	var values []string
	for _, h := range r.Headers {
		if strings.EqualFold(h.Name, name) {
			values = append(values, h.Value)
		}
	}
	return values
}

// GetCookie 根据名称获取响应Cookie（区分大小写）
func (r *Response) GetCookie(name string) *Cookie {
	if r == nil {
		return nil
	}
	for i := range r.Cookies {
		if r.Cookies[i].Name == name {
			return &r.Cookies[i]
		}
	}
	return nil
}

// HasHeader 检查指定名称的响应头是否存在（不区分大小写）
func (r *Response) HasHeader(name string) bool {
	if r == nil {
		return false
	}
	for _, h := range r.Headers {
		if strings.EqualFold(h.Name, name) {
			return true
		}
	}
	return false
}

// GetContentType 获取Content-Type响应头（不区分大小写）
func (r *Response) GetContentType() string {
	return r.GetHeader("Content-Type")
}

// ========== Content 方法 ==========

// EncodeContent 使用base64编码二进制数据并设置相应字段
func (c *Content) EncodeContent(data []byte, mimeType string) {
	if c == nil {
		return
	}
	c.Text = base64.StdEncoding.EncodeToString(data)
	c.Encoding = "base64"
	c.MimeType = mimeType
	c.Size = len(data)
}

// SetText 设置文本内容并更新大小
func (c *Content) SetText(text string) {
	if c == nil {
		return
	}
	c.Text = text
	c.Size = len(text)
}
