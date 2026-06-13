package har

import "time"

// 实现标准HAR类型的接口

// GetVersion 实现HARProvider接口
func (h *Har) GetVersion() string {
	return h.Log.Version
}

// GetCreator 实现HARProvider接口
func (h *Har) GetCreator() Creator {
	return h.Log.Creator
}

// GetBrowser 实现HARProvider接口
func (h *Har) GetBrowser() Browser {
	return h.Log.Browser
}

// GetEntries 实现HARProvider接口
func (h *Har) GetEntries() []EntryProvider {
	providers := make([]EntryProvider, len(h.Log.Entries))
	for i := range h.Log.Entries {
		providers[i] = &h.Log.Entries[i]
	}
	return providers
}

// GetPages 实现HARProvider接口
func (h *Har) GetPages() []PageProvider {
	providers := make([]PageProvider, len(h.Log.Pages))
	for i := range h.Log.Pages {
		providers[i] = &h.Log.Pages[i]
	}
	return providers
}

// ToStandard 实现HARProvider接口
func (h *Har) ToStandard() *Har {
	return h
}

// Entries 接口实现

// GetStartedDateTime 实现EntryProvider接口
func (e *Entries) GetStartedDateTime() time.Time {
	return e.StartedDateTime
}

// GetTime 实现EntryProvider接口
func (e *Entries) GetTime() float64 {
	return e.Time
}

// GetRequest 实现EntryProvider接口
func (e *Entries) GetRequest() RequestProvider {
	return &e.Request
}

// GetResponse 实现EntryProvider接口
func (e *Entries) GetResponse() ResponseProvider {
	return &e.Response
}

// GetTimings 实现EntryProvider接口
func (e *Entries) GetTimings() TimingsProvider {
	return &e.Timings
}

// GetPageref 实现EntryProvider接口
func (e *Entries) GetPageref() string {
	return e.Pageref
}

// ToStandard 实现EntryProvider接口
func (e *Entries) ToStandard() Entries {
	return *e
}

// Request 接口实现

// GetMethod 实现RequestProvider接口
func (r *Request) GetMethod() string {
	return r.Method
}

// GetURL 实现RequestProvider接口
func (r *Request) GetURL() string {
	return r.URL
}

// GetHTTPVersion 实现RequestProvider接口
func (r *Request) GetHTTPVersion() string {
	return r.HTTPVersion
}

// GetHeaders 实现RequestProvider接口
func (r *Request) GetHeaders() []HeaderProvider {
	providers := make([]HeaderProvider, len(r.Headers))
	for i := range r.Headers {
		providers[i] = &r.Headers[i]
	}
	return providers
}

// GetCookies 实现RequestProvider接口
func (r *Request) GetCookies() []CookieProvider {
	providers := make([]CookieProvider, len(r.Cookies))
	for i := range r.Cookies {
		providers[i] = &r.Cookies[i]
	}
	return providers
}

// GetBodySize 实现RequestProvider接口
func (r *Request) GetBodySize() int {
	return r.BodySize
}

// GetHeadersSize 实现RequestProvider接口
func (r *Request) GetHeadersSize() int {
	return r.HeadersSize
}

// GetQueryString 实现RequestProvider接口
func (r *Request) GetQueryString() []QueryString {
	return r.QueryString
}

// GetPostData 实现RequestProvider接口
func (r *Request) GetPostData() *PostData {
	return r.PostData
}

// ToStandard 实现RequestProvider接口
func (r *Request) ToStandard() Request {
	return *r
}

// Response 接口实现

// GetStatus 实现ResponseProvider接口
func (r *Response) GetStatus() int {
	return r.Status
}

// GetStatusText 实现ResponseProvider接口
func (r *Response) GetStatusText() string {
	return r.StatusText
}

// GetHTTPVersion 实现ResponseProvider接口
func (r *Response) GetHTTPVersion() string {
	return r.HTTPVersion
}

// GetHeaders 实现ResponseProvider接口
func (r *Response) GetHeaders() []HeaderProvider {
	providers := make([]HeaderProvider, len(r.Headers))
	for i := range r.Headers {
		providers[i] = &r.Headers[i]
	}
	return providers
}

// GetCookies 实现ResponseProvider接口
func (r *Response) GetCookies() []CookieProvider {
	providers := make([]CookieProvider, len(r.Cookies))
	for i := range r.Cookies {
		providers[i] = &r.Cookies[i]
	}
	return providers
}

// GetContent 实现ResponseProvider接口
func (r *Response) GetContent() ContentProvider {
	return &r.Content
}

// GetBodySize 实现ResponseProvider接口
func (r *Response) GetBodySize() int {
	return r.BodySize
}

// GetHeadersSize 实现ResponseProvider接口
func (r *Response) GetHeadersSize() int {
	return r.HeadersSize
}

// ToStandard 实现ResponseProvider接口
func (r *Response) ToStandard() Response {
	return *r
}

// Headers 接口实现

// GetName 实现HeaderProvider接口
func (h *Headers) GetName() string {
	return h.Name
}

// GetValue 实现HeaderProvider接口
func (h *Headers) GetValue() string {
	return h.Value
}

// ToStandard 实现HeaderProvider接口
func (h *Headers) ToStandard() Headers {
	return *h
}

// Cookie 接口实现

// GetName 实现CookieProvider接口
func (c *Cookie) GetName() string {
	return c.Name
}

// GetValue 实现CookieProvider接口
func (c *Cookie) GetValue() string {
	return c.Value
}

// GetDomain 实现CookieProvider接口
func (c *Cookie) GetDomain() string {
	return c.Domain
}

// GetPath 实现CookieProvider接口
func (c *Cookie) GetPath() string {
	return c.Path
}

// GetExpires 实现CookieProvider接口
func (c *Cookie) GetExpires() time.Time {
	return c.Expires
}

// IsHTTPOnly 实现CookieProvider接口
func (c *Cookie) IsHTTPOnly() bool {
	return c.HTTPOnly
}

// IsSecure 实现CookieProvider接口
func (c *Cookie) IsSecure() bool {
	return c.Secure
}

// GetSameSite 实现CookieProvider接口
func (c *Cookie) GetSameSite() string {
	return c.SameSite
}

// ToStandard 实现CookieProvider接口
func (c *Cookie) ToStandard() Cookie {
	return *c
}

// Content 接口实现

// GetSize 实现ContentProvider接口
func (c *Content) GetSize() int {
	return c.Size
}

// GetMimeType 实现ContentProvider接口
func (c *Content) GetMimeType() string {
	return c.MimeType
}

// GetText 实现ContentProvider接口
func (c *Content) GetText() string {
	return c.Text
}

// GetEncoding 实现ContentProvider接口
func (c *Content) GetEncoding() string {
	return c.Encoding
}

// GetCompression 实现ContentProvider接口
func (c *Content) GetCompression() int {
	return c.Compression
}

// ToStandard 实现ContentProvider接口
func (c *Content) ToStandard() Content {
	return *c
}

// Timings 接口实现

// GetBlocked 实现TimingsProvider接口
func (t *Timings) GetBlocked() float64 {
	return t.Blocked
}

// GetDNS 实现TimingsProvider接口
func (t *Timings) GetDNS() float64 {
	return t.DNS
}

// GetConnect 实现TimingsProvider接口
func (t *Timings) GetConnect() float64 {
	return t.Connect
}

// GetSend 实现TimingsProvider接口
func (t *Timings) GetSend() float64 {
	return t.Send
}

// GetWait 实现TimingsProvider接口
func (t *Timings) GetWait() float64 {
	return t.Wait
}

// GetReceive 实现TimingsProvider接口
func (t *Timings) GetReceive() float64 {
	return t.Receive
}

// GetSSL 实现TimingsProvider接口
func (t *Timings) GetSSL() float64 {
	return t.Ssl
}

// ToStandard 实现TimingsProvider接口
func (t *Timings) ToStandard() Timings {
	return *t
}

// Pages 接口实现

// GetID 实现PageProvider接口
func (p *Pages) GetID() string {
	return p.ID
}

// GetTitle 实现PageProvider接口
func (p *Pages) GetTitle() string {
	return p.Title
}

// GetStartedDateTime 实现PageProvider接口
func (p *Pages) GetStartedDateTime() time.Time {
	return p.StartedDateTime
}

// GetPageTimings 实现PageProvider接口
func (p *Pages) GetPageTimings() PageTimingsProvider {
	return &p.PageTimings
}

// ToStandard 实现PageProvider接口
func (p *Pages) ToStandard() Pages {
	return *p
}

// PageTimings 接口实现

// GetOnContentLoad 实现PageTimingsProvider接口
func (pt *PageTimings) GetOnContentLoad() float64 {
	return pt.OnContentLoad
}

// GetOnLoad 实现PageTimingsProvider接口
func (pt *PageTimings) GetOnLoad() float64 {
	return pt.OnLoad
}

// ToStandard 实现PageTimingsProvider接口
func (pt *PageTimings) ToStandard() PageTimings {
	return *pt
}
