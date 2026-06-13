package har

import "time"

// LazyHar 接口实现

// GetVersion 实现HARProvider接口
func (h *LazyHar) GetVersion() string {
	return h.Log.Version
}

// GetCreator 实现HARProvider接口
func (h *LazyHar) GetCreator() Creator {
	return h.Log.Creator
}

// GetBrowser 实现HARProvider接口
func (h *LazyHar) GetBrowser() Browser {
	return Browser{}
}

// GetEntries 实现HARProvider接口
func (h *LazyHar) GetEntries() []EntryProvider {
	providers := make([]EntryProvider, len(h.Log.Entries))
	for i := range h.Log.Entries {
		providers[i] = &h.Log.Entries[i]
	}
	return providers
}

// GetPages 实现HARProvider接口
func (h *LazyHar) GetPages() []PageProvider {
	providers := make([]PageProvider, len(h.Log.Pages))
	for i := range h.Log.Pages {
		providers[i] = &h.Log.Pages[i]
	}
	return providers
}

// ToStandard 实现HARProvider接口
func (h *LazyHar) ToStandard() *Har {
	// 转换为标准HAR对象
	standardHar, _ := h.ToStandardHar()
	return standardHar
}

// LazyEntries 接口实现

// GetStartedDateTime 实现EntryProvider接口
func (e *LazyEntries) GetStartedDateTime() time.Time {
	return e.StartedDateTime
}

// GetTime 实现EntryProvider接口
func (e *LazyEntries) GetTime() float64 {
	return e.Time
}

// GetRequest 实现EntryProvider接口
func (e *LazyEntries) GetRequest() RequestProvider {
	return &e.Request
}

// GetResponse 实现EntryProvider接口
func (e *LazyEntries) GetResponse() ResponseProvider {
	return &e.Response
}

// GetTimings 实现EntryProvider接口
func (e *LazyEntries) GetTimings() TimingsProvider {
	return &e.Timings
}

// GetPageref 实现EntryProvider接口
func (e *LazyEntries) GetPageref() string {
	return e.Pageref
}

// ToStandard 实现EntryProvider接口
func (e *LazyEntries) ToStandard() Entries {
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
	return r.Status
}

// GetStatusText 实现ResponseProvider接口
func (r *LazyResponse) GetStatusText() string {
	return r.StatusText
}

// GetHTTPVersion 实现ResponseProvider接口
func (r *LazyResponse) GetHTTPVersion() string {
	return r.HTTPVersion
}

// GetHeaders 实现ResponseProvider接口
func (r *LazyResponse) GetHeaders() []HeaderProvider {
	providers := make([]HeaderProvider, len(r.Headers))
	for i := range r.Headers {
		providers[i] = &r.Headers[i]
	}
	return providers
}

// GetCookies 实现ResponseProvider接口
func (r *LazyResponse) GetCookies() []CookieProvider {
	providers := make([]CookieProvider, len(r.Cookies))
	for i := range r.Cookies {
		providers[i] = &r.Cookies[i]
	}
	return providers
}

// GetContent 实现ResponseProvider接口
func (r *LazyResponse) GetContent() ContentProvider {
	// LazyContent 指针不直接实现 ContentProvider 接口
	// 因此需要创建一个包装器
	return &lazyContentWrapper{content: r.Content}
}

// GetBodySize 实现ResponseProvider接口
func (r *LazyResponse) GetBodySize() int {
	return r.BodySize
}

// GetHeadersSize 实现ResponseProvider接口
func (r *LazyResponse) GetHeadersSize() int {
	return r.HeadersSize
}

// ToStandard 实现ResponseProvider接口
func (r *LazyResponse) ToStandard() Response {
	var content Content

	// 创建标准Content对象，保留所有字段
	if r.Content != nil {
		content = Content{
			Size:        r.Content.Size,
			MimeType:    r.Content.MimeType,
			Compression: r.Content.Compression,
			Comment:     r.Content.Comment,
		}
		if r.Content.Text != nil {
			content.Text = *r.Content.Text
		}
		if r.Content.Encoding != nil {
			content.Encoding = *r.Content.Encoding
		}
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
	if w.content == nil {
		return 0
	}
	return w.content.Size
}

// GetMimeType 实现 ContentProvider 接口
func (w *lazyContentWrapper) GetMimeType() string {
	if w.content == nil {
		return ""
	}
	return w.content.MimeType
}

// GetText 实现 ContentProvider 接口
func (w *lazyContentWrapper) GetText() string {
	if w.content == nil {
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
	if w.content == nil {
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
	if w.content == nil {
		return 0
	}
	return w.content.Compression
}

// ToStandard 实现 ContentProvider 接口
func (w *lazyContentWrapper) ToStandard() Content {
	if w.content == nil {
		return Content{}
	}

	content := Content{
		Size:        w.content.Size,
		MimeType:    w.content.MimeType,
		Compression: w.content.Compression,
		Comment:     w.content.Comment,
	}
	if w.content.Text != nil {
		content.Text = *w.content.Text
	}
	if w.content.Encoding != nil {
		content.Encoding = *w.content.Encoding
	}
	return content
}

// LazyContent 接口实现

// GetSize 实现ContentProvider接口
func (c *LazyContent) GetSize() int {
	return c.Size
}

// GetMimeType 实现ContentProvider接口
func (c *LazyContent) GetMimeType() string {
	return c.MimeType
}

// GetEncoding 实现ContentProvider接口
func (c *LazyContent) GetEncoding() string {
	_ = c.Load()
	if c.Encoding == nil {
		return ""
	}
	return *c.Encoding
}

// GetCompression 实现ContentProvider接口
func (c *LazyContent) GetCompression() int {
	return c.Compression
}

// ToStandard 实现ContentProvider接口
func (c *LazyContent) ToStandard() Content {
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
