package har

import "time"

// 实现标准HAR类型的接口

// GetVersion 实现HARProvider接口
func (h *Har) GetVersion() string {
	if h == nil {
		return ""
	}
	return h.Log.Version
}

// GetCreator 实现HARProvider接口
func (h *Har) GetCreator() Creator {
	if h == nil {
		return Creator{}
	}
	return h.Log.Creator
}

// GetBrowser 实现HARProvider接口
func (h *Har) GetBrowser() Browser {
	if h == nil {
		return Browser{}
	}
	return h.Log.Browser
}

// GetEntries 实现HARProvider接口
func (h *Har) GetEntries() []EntryProvider {
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
func (h *Har) GetPages() []PageProvider {
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
func (h *Har) ToStandard() *Har {
	if h == nil {
		return nil
	}
	return h
}

// Entries 接口实现

// GetStartedDateTime 实现EntryProvider接口
func (e *Entries) GetStartedDateTime() time.Time {
	if e == nil {
		return time.Time{}
	}
	return e.StartedDateTime
}

// GetTime 实现EntryProvider接口
func (e *Entries) GetTime() float64 {
	if e == nil {
		return 0
	}
	return e.Time
}

// GetRequest 实现EntryProvider接口
func (e *Entries) GetRequest() RequestProvider {
	if e == nil {
		return nil
	}
	return &e.Request
}

// GetResponse 实现EntryProvider接口
func (e *Entries) GetResponse() ResponseProvider {
	if e == nil {
		return nil
	}
	return &e.Response
}

// GetTimings 实现EntryProvider接口
func (e *Entries) GetTimings() TimingsProvider {
	if e == nil {
		return nil
	}
	return &e.Timings
}

// GetPageref 实现EntryProvider接口
func (e *Entries) GetPageref() string {
	if e == nil {
		return ""
	}
	return e.Pageref
}

// ToStandard 实现EntryProvider接口
func (e *Entries) ToStandard() Entries {
	if e == nil {
		return Entries{}
	}
	return *e
}

// Request 接口实现

// GetMethod 实现RequestProvider接口
func (r *Request) GetMethod() string {
	if r == nil {
		return ""
	}
	return r.Method
}

// GetURL 实现RequestProvider接口
func (r *Request) GetURL() string {
	if r == nil {
		return ""
	}
	return r.URL
}

// GetHTTPVersion 实现RequestProvider接口
func (r *Request) GetHTTPVersion() string {
	if r == nil {
		return ""
	}
	return r.HTTPVersion
}

// GetHeaders 实现RequestProvider接口
func (r *Request) GetHeaders() []HeaderProvider {
	if r == nil {
		return nil
	}
	providers := make([]HeaderProvider, len(r.Headers))
	for i := range r.Headers {
		providers[i] = &r.Headers[i]
	}
	return providers
}

// GetCookies 实现RequestProvider接口
func (r *Request) GetCookies() []CookieProvider {
	if r == nil {
		return nil
	}
	providers := make([]CookieProvider, len(r.Cookies))
	for i := range r.Cookies {
		providers[i] = &r.Cookies[i]
	}
	return providers
}

// GetBodySize 实现RequestProvider接口
func (r *Request) GetBodySize() int {
	if r == nil {
		return 0
	}
	return r.BodySize
}

// GetHeadersSize 实现RequestProvider接口
func (r *Request) GetHeadersSize() int {
	if r == nil {
		return 0
	}
	return r.HeadersSize
}

// GetQueryString 实现RequestProvider接口
func (r *Request) GetQueryString() []QueryString {
	if r == nil {
		return nil
	}
	return r.QueryString
}

// GetPostData 实现RequestProvider接口
func (r *Request) GetPostData() *PostData {
	if r == nil {
		return nil
	}
	return r.PostData
}

// ToStandard 实现RequestProvider接口
func (r *Request) ToStandard() Request {
	if r == nil {
		return Request{}
	}
	return *r
}

// Response 接口实现

// GetStatus 实现ResponseProvider接口
func (r *Response) GetStatus() int {
	if r == nil {
		return 0
	}
	return r.Status
}

// GetStatusText 实现ResponseProvider接口
func (r *Response) GetStatusText() string {
	if r == nil {
		return ""
	}
	return r.StatusText
}

// GetHTTPVersion 实现ResponseProvider接口
func (r *Response) GetHTTPVersion() string {
	if r == nil {
		return ""
	}
	return r.HTTPVersion
}

// GetHeaders 实现ResponseProvider接口
func (r *Response) GetHeaders() []HeaderProvider {
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
func (r *Response) GetCookies() []CookieProvider {
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
func (r *Response) GetContent() ContentProvider {
	if r == nil {
		return nil
	}
	return &r.Content
}

// GetBodySize 实现ResponseProvider接口
func (r *Response) GetBodySize() int {
	if r == nil {
		return 0
	}
	return r.BodySize
}

// GetHeadersSize 实现ResponseProvider接口
func (r *Response) GetHeadersSize() int {
	if r == nil {
		return 0
	}
	return r.HeadersSize
}

// ToStandard 实现ResponseProvider接口
func (r *Response) ToStandard() Response {
	if r == nil {
		return Response{}
	}
	return *r
}

// Headers 接口实现

// GetName 实现HeaderProvider接口
func (h *Headers) GetName() string {
	if h == nil {
		return ""
	}
	return h.Name
}

// GetValue 实现HeaderProvider接口
func (h *Headers) GetValue() string {
	if h == nil {
		return ""
	}
	return h.Value
}

// ToStandard 实现HeaderProvider接口
func (h *Headers) ToStandard() Headers {
	if h == nil {
		return Headers{}
	}
	return *h
}

// Cookie 接口实现

// GetName 实现CookieProvider接口
func (c *Cookie) GetName() string {
	if c == nil {
		return ""
	}
	return c.Name
}

// GetValue 实现CookieProvider接口
func (c *Cookie) GetValue() string {
	if c == nil {
		return ""
	}
	return c.Value
}

// GetDomain 实现CookieProvider接口
func (c *Cookie) GetDomain() string {
	if c == nil {
		return ""
	}
	return c.Domain
}

// GetPath 实现CookieProvider接口
func (c *Cookie) GetPath() string {
	if c == nil {
		return ""
	}
	return c.Path
}

// GetExpires 实现CookieProvider接口
func (c *Cookie) GetExpires() time.Time {
	if c == nil {
		return time.Time{}
	}
	return c.Expires
}

// IsHTTPOnly 实现CookieProvider接口
func (c *Cookie) IsHTTPOnly() bool {
	if c == nil {
		return false
	}
	return c.HTTPOnly
}

// IsSecure 实现CookieProvider接口
func (c *Cookie) IsSecure() bool {
	if c == nil {
		return false
	}
	return c.Secure
}

// GetSameSite 实现CookieProvider接口
func (c *Cookie) GetSameSite() string {
	if c == nil {
		return ""
	}
	return c.SameSite
}

// ToStandard 实现CookieProvider接口
func (c *Cookie) ToStandard() Cookie {
	if c == nil {
		return Cookie{}
	}
	return *c
}

// Content 接口实现

// GetSize 实现ContentProvider接口
func (c *Content) GetSize() int {
	if c == nil {
		return 0
	}
	return c.Size
}

// GetMimeType 实现ContentProvider接口
func (c *Content) GetMimeType() string {
	if c == nil {
		return ""
	}
	return c.MimeType
}

// GetText 实现ContentProvider接口
func (c *Content) GetText() string {
	if c == nil {
		return ""
	}
	return c.Text
}

// GetEncoding 实现ContentProvider接口
func (c *Content) GetEncoding() string {
	if c == nil {
		return ""
	}
	return c.Encoding
}

// GetCompression 实现ContentProvider接口
func (c *Content) GetCompression() int {
	if c == nil {
		return 0
	}
	return c.Compression
}

// ToStandard 实现ContentProvider接口
func (c *Content) ToStandard() Content {
	if c == nil {
		return Content{}
	}
	return *c
}

// Timings 接口实现

// GetBlocked 实现TimingsProvider接口
func (t *Timings) GetBlocked() float64 {
	if t == nil {
		return 0
	}
	return t.Blocked
}

// GetDNS 实现TimingsProvider接口
func (t *Timings) GetDNS() float64 {
	if t == nil {
		return 0
	}
	return t.DNS
}

// GetConnect 实现TimingsProvider接口
func (t *Timings) GetConnect() float64 {
	if t == nil {
		return 0
	}
	return t.Connect
}

// GetSend 实现TimingsProvider接口
func (t *Timings) GetSend() float64 {
	if t == nil {
		return 0
	}
	return t.Send
}

// GetWait 实现TimingsProvider接口
func (t *Timings) GetWait() float64 {
	if t == nil {
		return 0
	}
	return t.Wait
}

// GetReceive 实现TimingsProvider接口
func (t *Timings) GetReceive() float64 {
	if t == nil {
		return 0
	}
	return t.Receive
}

// GetSSL 实现TimingsProvider接口
func (t *Timings) GetSSL() float64 {
	if t == nil {
		return 0
	}
	return t.Ssl
}

// ToStandard 实现TimingsProvider接口
func (t *Timings) ToStandard() Timings {
	if t == nil {
		return Timings{}
	}
	return *t
}

// Pages 接口实现

// GetID 实现PageProvider接口
func (p *Pages) GetID() string {
	if p == nil {
		return ""
	}
	return p.ID
}

// GetTitle 实现PageProvider接口
func (p *Pages) GetTitle() string {
	if p == nil {
		return ""
	}
	return p.Title
}

// GetStartedDateTime 实现PageProvider接口
func (p *Pages) GetStartedDateTime() time.Time {
	if p == nil {
		return time.Time{}
	}
	return p.StartedDateTime
}

// GetPageTimings 实现PageProvider接口
func (p *Pages) GetPageTimings() PageTimingsProvider {
	if p == nil {
		return nil
	}
	return &p.PageTimings
}

// ToStandard 实现PageProvider接口
func (p *Pages) ToStandard() Pages {
	if p == nil {
		return Pages{}
	}
	return *p
}

// PageTimings 接口实现

// GetOnContentLoad 实现PageTimingsProvider接口
func (pt *PageTimings) GetOnContentLoad() float64 {
	if pt == nil {
		return 0
	}
	return pt.OnContentLoad
}

// GetOnLoad 实现PageTimingsProvider接口
func (pt *PageTimings) GetOnLoad() float64 {
	if pt == nil {
		return 0
	}
	return pt.OnLoad
}

// ToStandard 实现PageTimingsProvider接口
func (pt *PageTimings) ToStandard() PageTimings {
	if pt == nil {
		return PageTimings{}
	}
	return *pt
}
