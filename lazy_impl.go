package har

import "time"

// LazyHar 接口实现

// GetVersion 实现HARProvider接口
func (h *LazyHar) GetVersion() string {
	if h == nil {
		return ""
	}
	return h.Log.Version
}

// GetCreator 实现HARProvider接口
func (h *LazyHar) GetCreator() Creator {
	if h == nil {
		return Creator{}
	}
	return h.Log.Creator
}

// GetBrowser 实现HARProvider接口
func (h *LazyHar) GetBrowser() Browser {
	if h == nil {
		return Browser{}
	}
	return h.Log.Browser
}

// GetEntries 实现HARProvider接口
func (h *LazyHar) GetEntries() []EntryProvider {
	if h == nil {
		return nil
	}
	providers := make([]EntryProvider, len(h.Log.Entries))
	for i := range h.Log.Entries {
		providers[i] = &h.Log.Entries[i]
	}
	return providers
}

// GetPages 实现HARProvider接口
func (h *LazyHar) GetPages() []PageProvider {
	if h == nil {
		return nil
	}
	providers := make([]PageProvider, len(h.Log.Pages))
	for i := range h.Log.Pages {
		providers[i] = &h.Log.Pages[i]
	}
	return providers
}

// ToStandard 实现HARProvider接口
func (h *LazyHar) ToStandard() *Har {
	if h == nil {
		return nil
	}
	result := &Har{
		Log: Log{
			Version: h.Log.Version,
			Creator: h.Log.Creator,
			Browser: h.Log.Browser,
			Pages:   h.Log.Pages,
			Entries: make([]Entries, len(h.Log.Entries)),
		},
	}
	for i := range h.Log.Entries {
		result.Log.Entries[i] = h.Log.Entries[i].ToStandard()
	}
	return result
}

// LazyEntries 接口实现

// GetStartedDateTime 实现EntryProvider接口
func (e *LazyEntries) GetStartedDateTime() time.Time {
	if e == nil {
		return time.Time{}
	}
	return e.StartedDateTime
}

// GetTime 实现EntryProvider接口
func (e *LazyEntries) GetTime() float64 {
	if e == nil {
		return 0
	}
	return e.Time
}

// GetRequest 实现EntryProvider接口
func (e *LazyEntries) GetRequest() RequestProvider {
	if e == nil {
		return nil
	}
	return &e.Request
}

// GetResponse 实现EntryProvider接口
func (e *LazyEntries) GetResponse() ResponseProvider {
	if e == nil {
		return nil
	}
	return &e.Response
}

// GetTimings 实现EntryProvider接口
func (e *LazyEntries) GetTimings() TimingsProvider {
	if e == nil {
		return nil
	}
	return &e.Timings
}

// GetPageref 实现EntryProvider接口
func (e *LazyEntries) GetPageref() string {
	if e == nil {
		return ""
	}
	return e.Pageref
}

// ToStandard 实现EntryProvider接口
func (e *LazyEntries) ToStandard() Entries {
	if e == nil {
		return Entries{}
	}
	return Entries{
		StartedDateTime: e.StartedDateTime,
		Time:            e.Time,
		Request:         e.Request,
		Response:        e.Response.ToStandard(),
		Cache:           e.Cache,
		Timings:         e.Timings,
		Pageref:         e.Pageref,
		ServerIPAddress: e.ServerIPAddress,
		Connection:      e.Connection,
		Comment:         e.Comment,
	}
}

// LazyResponse 接口实现

// GetStatus 实现ResponseProvider接口
func (r *LazyResponse) GetStatus() int {
	if r == nil {
		return 0
	}
	return r.Status
}

// GetStatusText 实现ResponseProvider接口
func (r *LazyResponse) GetStatusText() string {
	if r == nil {
		return ""
	}
	return r.StatusText
}

// GetHTTPVersion 实现ResponseProvider接口
func (r *LazyResponse) GetHTTPVersion() string {
	if r == nil {
		return ""
	}
	return r.HTTPVersion
}

// GetHeaders 实现ResponseProvider接口
func (r *LazyResponse) GetHeaders() []HeaderProvider {
	if r == nil {
		return nil
	}
	providers := make([]HeaderProvider, len(r.Headers))
	for i := range r.Headers {
		providers[i] = &r.Headers[i]
	}
	return providers
}

// GetCookies 实现ResponseProvider接口
func (r *LazyResponse) GetCookies() []CookieProvider {
	if r == nil {
		return nil
	}
	providers := make([]CookieProvider, len(r.Cookies))
	for i := range r.Cookies {
		providers[i] = &r.Cookies[i]
	}
	return providers
}

// GetContent 实现ResponseProvider接口
func (r *LazyResponse) GetContent() ContentProvider {
	if r == nil {
		return nil
	}
	// LazyContent 指针不直接实现 ContentProvider 接口
	// 因此需要创建一个包装器
	return &lazyContentWrapper{content: r.Content}
}

// GetBodySize 实现ResponseProvider接口
func (r *LazyResponse) GetBodySize() int {
	if r == nil {
		return 0
	}
	return r.BodySize
}

// GetHeadersSize 实现ResponseProvider接口
func (r *LazyResponse) GetHeadersSize() int {
	if r == nil {
		return 0
	}
	return r.HeadersSize
}

// ToStandard 实现ResponseProvider接口
func (r *LazyResponse) ToStandard() Response {
	if r == nil {
		return Response{}
	}
	var content Content

	// 创建标准Content对象，保留所有字段
	if r.Content != nil {
		content = r.Content.ToStandard()
	}

	return Response{
		Status:       r.Status,
		StatusText:   r.StatusText,
		HTTPVersion:  r.HTTPVersion,
		Cookies:      r.Cookies,
		Headers:      r.Headers,
		RedirectURL:  r.RedirectURL,
		HeadersSize:  r.HeadersSize,
		BodySize:     r.BodySize,
		Content:      content,
		TransferSize: r.TransferSize,
		Error:        r.Error,
	}
}

// lazyContentWrapper 是 LazyContent 的包装器
// 实现 ContentProvider 接口
type lazyContentWrapper struct {
	content *LazyContent
}

// GetSize 实现 ContentProvider 接口
func (w *lazyContentWrapper) GetSize() int {
	if w == nil || w.content == nil {
		return 0
	}
	return w.content.Size
}

// GetMimeType 实现 ContentProvider 接口
func (w *lazyContentWrapper) GetMimeType() string {
	if w == nil || w.content == nil {
		return ""
	}
	return w.content.MimeType
}

// GetText 实现 ContentProvider 接口
func (w *lazyContentWrapper) GetText() string {
	if w == nil || w.content == nil {
		return ""
	}
	// LazyContent.GetText() returns (*string, error) - defined in lazy.go
	text, err := w.content.GetText()
	if err != nil || text == nil {
		return ""
	}
	return *text
}

// GetEncoding 实现 ContentProvider 接口
func (w *lazyContentWrapper) GetEncoding() string {
	if w == nil || w.content == nil {
		return ""
	}

	// 确保内容已加载
	_ = w.content.Load()
	if w.content.Encoding == nil {
		return ""
	}
	return *w.content.Encoding
}

// GetCompression 实现 ContentProvider 接口
func (w *lazyContentWrapper) GetCompression() int {
	if w == nil || w.content == nil {
		return 0
	}
	return w.content.Compression
}

// ToStandard 实现 ContentProvider 接口
func (w *lazyContentWrapper) ToStandard() Content {
	if w == nil || w.content == nil {
		return Content{}
	}

	return w.content.ToStandard()
}

// LazyContent 接口实现

// GetSize 实现ContentProvider接口
func (c *LazyContent) GetSize() int {
	if c == nil {
		return 0
	}
	return c.Size
}

// GetMimeType 实现ContentProvider接口
func (c *LazyContent) GetMimeType() string {
	if c == nil {
		return ""
	}
	return c.MimeType
}

// GetEncoding 实现ContentProvider接口
func (c *LazyContent) GetEncoding() string {
	if c == nil {
		return ""
	}
	_ = c.Load()
	if c.Encoding == nil {
		return ""
	}
	return *c.Encoding
}

// GetCompression 实现ContentProvider接口
func (c *LazyContent) GetCompression() int {
	if c == nil {
		return 0
	}
	return c.Compression
}

// ToStandard 实现ContentProvider接口
func (c *LazyContent) ToStandard() Content {
	if c == nil {
		return Content{}
	}
	_ = c.Load()
	content := Content{
		Size:        c.Size,
		MimeType:    c.MimeType,
		Compression: c.Compression,
		Comment:     c.Comment,
	}
	if c.Text != nil {
		content.Text = *c.Text
	}
	if c.Encoding != nil {
		content.Encoding = *c.Encoding
	}
	return content
}
