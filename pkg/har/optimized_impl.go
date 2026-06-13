package har

import "time"

// OptimizedHar 接口实现

// GetVersion 实现HARProvider接口
func (h *OptimizedHar) GetVersion() string {
	return h.Log.Version
}

// GetCreator 实现HARProvider接口
func (h *OptimizedHar) GetCreator() Creator {
	return h.Log.Creator
}

// GetBrowser 实现HARProvider接口
func (h *OptimizedHar) GetBrowser() Browser {
	return Browser{}
}

// GetEntries 实现HARProvider接口
func (h *OptimizedHar) GetEntries() []EntryProvider {
	providers := make([]EntryProvider, len(h.Log.Entries))
	for i := range h.Log.Entries {
		providers[i] = &h.Log.Entries[i]
	}
	return providers
}

// GetPages 实现HARProvider接口
func (h *OptimizedHar) GetPages() []PageProvider {
	providers := make([]PageProvider, len(h.Log.Pages))
	for i := range h.Log.Pages {
		providers[i] = &h.Log.Pages[i]
	}
	return providers
}

// ToStandard 实现HARProvider接口
func (h *OptimizedHar) ToStandard() *Har {
	// 从优化格式转换为标准格式
	standard := &Har{
		Log: Log{
			Version: h.Log.Version,
			Creator: h.Log.Creator,
			Pages:   h.Log.Pages,
			Entries: make([]Entries, len(h.Log.Entries)),
		},
	}

	// 转换条目
	for i, entry := range h.Log.Entries {
		standard.Log.Entries[i] = entry.ToStandard()
	}

	return standard
}

// OptimizedEntries 接口实现

// GetStartedDateTime 实现EntryProvider接口
func (e *OptimizedEntries) GetStartedDateTime() time.Time {
	return e.StartedDateTime
}

// GetTime 实现EntryProvider接口
func (e *OptimizedEntries) GetTime() float64 {
	return e.Time
}

// GetRequest 实现EntryProvider接口
func (e *OptimizedEntries) GetRequest() RequestProvider {
	return &e.Request
}

// GetResponse 实现EntryProvider接口
func (e *OptimizedEntries) GetResponse() ResponseProvider {
	return &e.Response
}

// GetTimings 实现EntryProvider接口
func (e *OptimizedEntries) GetTimings() TimingsProvider {
	return &e.Timings
}

// GetPageref 实现EntryProvider接口
func (e *OptimizedEntries) GetPageref() string {
	if e.PageRef != nil {
		return *e.PageRef
	}
	return ""
}

// ToStandard 实现EntryProvider接口
func (e *OptimizedEntries) ToStandard() Entries {
	// 转换为标准格式
	entry := Entries{
		StartedDateTime: e.StartedDateTime,
		Time:            e.Time,
		Request:         e.Request.ToStandard(),
		Response:        e.Response.ToStandard(),
		Timings:         e.Timings.ToStandard(),
	}

	// 可选地添加Pageref（如果不为空）
	if e.PageRef != nil {
		entry.Pageref = *e.PageRef
	}

	// 可选地添加ServerIP
	if e.ServerIP != nil {
		entry.ServerIPAddress = *e.ServerIP
	}

	// 可选地添加Connection
	if e.Connection != nil {
		entry.Connection = *e.Connection
	}

	// 转换缓存
	if e.Cache != nil {
		entry.Cache = *e.Cache
	}

	return entry
}

// OptimizedRequest 接口实现

// GetMethod 实现RequestProvider接口
func (r *OptimizedRequest) GetMethod() string {
	// 从HTTPMethod枚举转换为字符串
	switch r.Method {
	case MethodGET:
		return "GET"
	case MethodPOST:
		return "POST"
	case MethodPUT:
		return "PUT"
	case MethodDELETE:
		return "DELETE"
	case MethodHEAD:
		return "HEAD"
	case MethodOPTIONS:
		return "OPTIONS"
	case MethodPATCH:
		return "PATCH"
	case MethodCONNECT:
		return "CONNECT"
	case MethodTRACE:
		return "TRACE"
	default:
		return "UNKNOWN"
	}
}

// GetURL 实现RequestProvider接口
func (r *OptimizedRequest) GetURL() string {
	return r.URL
}

// GetHTTPVersion 实现RequestProvider接口
func (r *OptimizedRequest) GetHTTPVersion() string {
	return r.HTTPVersion
}

// GetHeaders 实现RequestProvider接口
func (r *OptimizedRequest) GetHeaders() []HeaderProvider {
	// 将map转换为切片
	headers := make([]HeaderProvider, 0, len(r.Headers))
	for name, value := range r.Headers {
		header := &Headers{
			Name:  name,
			Value: value,
		}
		headers = append(headers, header)
	}
	return headers
}

// GetCookies 实现RequestProvider接口
func (r *OptimizedRequest) GetCookies() []CookieProvider {
	providers := make([]CookieProvider, len(r.Cookies))
	for i := range r.Cookies {
		providers[i] = &r.Cookies[i]
	}
	return providers
}

// GetBodySize 实现RequestProvider接口
func (r *OptimizedRequest) GetBodySize() int {
	if r.BodySize != nil {
		return *r.BodySize
	}
	return 0
}

// GetHeadersSize 实现RequestProvider接口
func (r *OptimizedRequest) GetHeadersSize() int {
	if r.HeadersSize != nil {
		return *r.HeadersSize
	}
	return 0
}

// GetQueryString 实现RequestProvider接口
func (r *OptimizedRequest) GetQueryString() []QueryString {
	params := make([]QueryString, 0, len(r.QueryString))
	for name, value := range r.QueryString {
		params = append(params, QueryString{
			Name:  name,
			Value: value,
		})
	}
	return params
}

// GetPostData 实现RequestProvider接口
func (r *OptimizedRequest) GetPostData() *PostData {
	return r.PostData
}

// ToStandard 实现RequestProvider接口
func (r *OptimizedRequest) ToStandard() Request {
	// 从优化格式转换为标准格式
	request := Request{
		Method:      r.GetMethod(),
		URL:         r.URL,
		HTTPVersion: r.HTTPVersion,
		PostData:    r.PostData,
	}

	// 转换头部
	for name, value := range r.Headers {
		request.Headers = append(request.Headers, Headers{
			Name:  name,
			Value: value,
		})
	}

	// 转换查询参数
	for name, value := range r.QueryString {
		request.QueryString = append(request.QueryString, QueryString{
			Name:  name,
			Value: value,
		})
	}

	// 转换Cookie
	request.Cookies = make([]Cookie, len(r.Cookies))
	for i, cookie := range r.Cookies {
		request.Cookies[i] = cookie
	}

	// 处理可选字段
	if r.HeadersSize != nil {
		request.HeadersSize = *r.HeadersSize
	}

	if r.BodySize != nil {
		request.BodySize = *r.BodySize
	}

	return request
}

// OptimizedResponse 接口实现

// GetStatus 实现ResponseProvider接口
func (r *OptimizedResponse) GetStatus() int {
	return r.Status
}

// GetStatusText 实现ResponseProvider接口
func (r *OptimizedResponse) GetStatusText() string {
	return r.StatusText
}

// GetHTTPVersion 实现ResponseProvider接口
func (r *OptimizedResponse) GetHTTPVersion() string {
	return r.HTTPVersion
}

// GetHeaders 实现ResponseProvider接口
func (r *OptimizedResponse) GetHeaders() []HeaderProvider {
	// 将map转换为切片
	headers := make([]HeaderProvider, 0, len(r.Headers))
	for name, value := range r.Headers {
		header := &Headers{
			Name:  name,
			Value: value,
		}
		headers = append(headers, header)
	}
	return headers
}

// GetCookies 实现ResponseProvider接口
func (r *OptimizedResponse) GetCookies() []CookieProvider {
	providers := make([]CookieProvider, len(r.Cookies))
	for i := range r.Cookies {
		providers[i] = &r.Cookies[i]
	}
	return providers
}

// GetContent 实现ResponseProvider接口
func (r *OptimizedResponse) GetContent() ContentProvider {
	if r.Content == nil {
		return nil
	}
	return r.Content
}

// GetBodySize 实现ResponseProvider接口
func (r *OptimizedResponse) GetBodySize() int {
	if r.BodySize != nil {
		return *r.BodySize
	}
	return 0
}

// GetHeadersSize 实现ResponseProvider接口
func (r *OptimizedResponse) GetHeadersSize() int {
	if r.HeadersSize != nil {
		return *r.HeadersSize
	}
	return 0
}

// ToStandard 实现ResponseProvider接口
func (r *OptimizedResponse) ToStandard() Response {
	// 从优化格式转换为标准格式
	response := Response{
		Status:      r.Status,
		StatusText:  r.StatusText,
		HTTPVersion: r.HTTPVersion,
		RedirectURL: r.RedirectURL,
	}

	// 处理Content字段
	if r.Content != nil {
		response.Content = r.Content.ToStandard()
	}

	// 转换头部
	for name, value := range r.Headers {
		response.Headers = append(response.Headers, Headers{
			Name:  name,
			Value: value,
		})
	}

	// 转换Cookie
	response.Cookies = make([]Cookie, len(r.Cookies))
	for i, cookie := range r.Cookies {
		response.Cookies[i] = cookie
	}

	// 处理可选字段
	if r.HeadersSize != nil {
		response.HeadersSize = *r.HeadersSize
	}

	if r.BodySize != nil {
		response.BodySize = *r.BodySize
	}

	return response
}

// OptimizedContent 接口实现

// GetSize 实现ContentProvider接口
func (c *OptimizedContent) GetSize() int {
	return c.Size
}

// GetMimeType 实现ContentProvider接口
func (c *OptimizedContent) GetMimeType() string {
	return c.MimeType
}

// GetText 实现ContentProvider接口
func (c *OptimizedContent) GetText() string {
	if c.Text != nil {
		return *c.Text
	}
	return ""
}

// GetEncoding 实现ContentProvider接口
func (c *OptimizedContent) GetEncoding() string {
	if c.Encoding != nil {
		return *c.Encoding
	}
	return ""
}

// GetCompression 实现ContentProvider接口
func (c *OptimizedContent) GetCompression() int {
	return 0 // OptimizedContent doesn't track compression
}

// ToStandard 实现ContentProvider接口
func (c *OptimizedContent) ToStandard() Content {
	content := Content{
		Size:     c.Size,
		MimeType: c.MimeType,
	}
	if c.Text != nil {
		content.Text = *c.Text
	}
	if c.Encoding != nil {
		content.Encoding = *c.Encoding
	}
	if c.Comment != nil {
		content.Comment = *c.Comment
	}
	return content
}

// OptimizedTimings 接口实现

// GetBlocked 实现TimingsProvider接口
func (t *OptimizedTimings) GetBlocked() float64 {
	if t.Blocked != nil {
		return *t.Blocked
	}
	return -1
}

// GetDNS 实现TimingsProvider接口
func (t *OptimizedTimings) GetDNS() float64 {
	if t.DNS != nil {
		return *t.DNS
	}
	return -1
}

// GetConnect 实现TimingsProvider接口
func (t *OptimizedTimings) GetConnect() float64 {
	if t.Connect != nil {
		return *t.Connect
	}
	return -1
}

// GetSend 实现TimingsProvider接口
func (t *OptimizedTimings) GetSend() float64 {
	if t.Send != nil {
		return *t.Send
	}
	return -1
}

// GetWait 实现TimingsProvider接口
func (t *OptimizedTimings) GetWait() float64 {
	if t.Wait != nil {
		return *t.Wait
	}
	return -1
}

// GetReceive 实现TimingsProvider接口
func (t *OptimizedTimings) GetReceive() float64 {
	if t.Receive != nil {
		return *t.Receive
	}
	return -1
}

// GetSSL 实现TimingsProvider接口
func (t *OptimizedTimings) GetSSL() float64 {
	if t.Ssl != nil {
		return *t.Ssl
	}
	return -1
}

// ToStandard 实现TimingsProvider接口
func (t *OptimizedTimings) ToStandard() Timings {
	timings := Timings{}

	// 处理必需字段，避免空指针
	if t.Blocked != nil {
		timings.Blocked = *t.Blocked
	} else {
		timings.Blocked = -1
	}

	if t.Send != nil {
		timings.Send = *t.Send
	} else {
		timings.Send = -1
	}

	if t.Wait != nil {
		timings.Wait = *t.Wait
	} else {
		timings.Wait = -1
	}

	if t.Receive != nil {
		timings.Receive = *t.Receive
	} else {
		timings.Receive = -1
	}

	// 处理可选字段
	if t.DNS != nil {
		timings.DNS = *t.DNS
	} else {
		timings.DNS = -1
	}

	if t.Connect != nil {
		timings.Connect = *t.Connect
	} else {
		timings.Connect = -1
	}

	if t.Ssl != nil {
		timings.Ssl = *t.Ssl
	} else {
		timings.Ssl = -1
	}

	return timings
}
