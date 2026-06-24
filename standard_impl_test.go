package har

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Helper: build a fully-populated Har fixture for testing standard_impl methods.
func newTestHar() *Har {
	ts := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	return &Har{
		Log: Log{
			Version: "1.2",
			Creator: Creator{
				Name:    "test-creator",
				Version: "1.0",
				Comment: "creator comment",
			},
			Browser: Browser{
				Name:    "test-browser",
				Version: "2.0",
				Comment: "browser comment",
			},
			Pages: []Pages{
				{
					ID:              "page_1",
					Title:           "Test Page",
					StartedDateTime: ts,
					PageTimings: PageTimings{
						OnContentLoad: 150.5,
						OnLoad:        300.2,
						Comment:       "page timings comment",
					},
					Comment: "page comment",
				},
			},
			Entries: []Entries{
				{
					StartedDateTime: ts,
					Time:            123.45,
					Pageref:         "page_1",
					Request: Request{
						Method:      "GET",
						URL:         "https://example.com/api/users",
						HTTPVersion: "HTTP/1.1",
						Headers: []Headers{
							{Name: "Content-Type", Value: "application/json", Comment: "hdr comment"},
						},
						Cookies: []Cookie{
							{
								Name:     "session_id",
								Value:    "abc123",
								Path:     "/",
								Domain:   ".example.com",
								Expires:  ts.Add(24 * time.Hour),
								HTTPOnly: true,
								Secure:   true,
								SameSite: "Strict",
								Comment:  "cookie comment",
							},
						},
						QueryString: []QueryString{
							{Name: "page", Value: "1", Comment: "qs comment"},
						},
						PostData: &PostData{
							MimeType: "application/x-www-form-urlencoded",
							Text:     "key=value",
							Comment:  "postdata comment",
						},
						HeadersSize: 256,
						BodySize:    128,
						Comment:     "request comment",
					},
					Response: Response{
						Status:      200,
						StatusText:  "OK",
						HTTPVersion: "HTTP/1.1",
						Headers: []Headers{
							{Name: "X-Custom", Value: "resp-val", Comment: "resp hdr comment"},
						},
						Cookies: []Cookie{
							{
								Name:     "track",
								Value:    "xyz789",
								Path:     "/",
								Domain:   ".example.com",
								Expires:  ts.Add(48 * time.Hour),
								HTTPOnly: false,
								Secure:   false,
								SameSite: "Lax",
								Comment:  "resp cookie comment",
							},
						},
						Content: Content{
							Size:        1024,
							MimeType:    "application/json",
							Text:        `{"status":"ok"}`,
							Encoding:    "base64",
							Compression: -256,
							Comment:     "content comment",
						},
						RedirectURL: "https://example.com/redirect",
						HeadersSize: 512,
						BodySize:    2048,
						Comment:     "response comment",
					},
					Timings: Timings{
						Blocked: 10.1,
						DNS:     20.2,
						Connect: 30.3,
						Ssl:     40.4,
						Send:    5.5,
						Wait:    50.5,
						Receive: 15.6,
						Comment: "timings comment",
					},
					Cache: Cache{
						BeforeRequest: &BeforeRequest{
							Expires:    ts,
							LastAccess: ts,
							ETag:       "etag-before",
							HitCount:   3,
						},
						AfterRequest: &AfterRequest{
							Expires:    ts,
							LastAccess: ts,
							ETag:       "etag-after",
							HitCount:   5,
						},
					},
					ServerIPAddress: "10.0.0.1",
					Connection:      "conn-123",
					Comment:         "entry comment",
				},
			},
		},
	}
}

// ---------------------------------------------------------------------------
// Har (HARProvider) tests
// ---------------------------------------------------------------------------

func TestStandardHar_GetVersion(t *testing.T) {
	h := newTestHar()
	assert.Equal(t, "1.2", h.GetVersion())
}

func TestStandardHar_GetCreator(t *testing.T) {
	h := newTestHar()
	creator := h.GetCreator()
	assert.Equal(t, "test-creator", creator.Name)
	assert.Equal(t, "1.0", creator.Version)
	assert.Equal(t, "creator comment", creator.Comment)
}

func TestStandardHar_GetBrowser(t *testing.T) {
	h := newTestHar()
	browser := h.GetBrowser()
	assert.Equal(t, "test-browser", browser.Name)
	assert.Equal(t, "2.0", browser.Version)
	assert.Equal(t, "browser comment", browser.Comment)
}

func TestStandardHar_GetEntries(t *testing.T) {
	h := newTestHar()
	entries := h.GetEntries()
	assert.Len(t, entries, 1)
	// Verify the entry provider returns correct data
	assert.Equal(t, 123.45, entries[0].GetTime())
}

func TestStandardHar_GetPages(t *testing.T) {
	h := newTestHar()
	pages := h.GetPages()
	assert.Len(t, pages, 1)
	assert.Equal(t, "page_1", pages[0].GetID())
}

func TestStandardHar_ToStandard(t *testing.T) {
	h := newTestHar()
	std := h.ToStandard()
	assert.Same(t, h, std) // ToStandard returns the same pointer
}

// ---------------------------------------------------------------------------
// Entries (EntryProvider) tests
// ---------------------------------------------------------------------------

func TestStandardEntries_GetStartedDateTime(t *testing.T) {
	ts := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	e := &Entries{StartedDateTime: ts}
	assert.Equal(t, ts, e.GetStartedDateTime())
}

func TestStandardEntries_GetTime(t *testing.T) {
	e := &Entries{Time: 999.99}
	assert.Equal(t, 999.99, e.GetTime())
}

func TestStandardEntries_GetRequest(t *testing.T) {
	e := &Entries{
		Request: Request{Method: "POST", URL: "https://example.com"},
	}
	rp := e.GetRequest()
	assert.Equal(t, "POST", rp.GetMethod())
	assert.Equal(t, "https://example.com", rp.GetURL())
}

func TestStandardEntries_GetResponse(t *testing.T) {
	e := &Entries{
		Response: Response{Status: 404, StatusText: "Not Found"},
	}
	rp := e.GetResponse()
	assert.Equal(t, 404, rp.GetStatus())
	assert.Equal(t, "Not Found", rp.GetStatusText())
}

func TestStandardEntries_GetTimings(t *testing.T) {
	e := &Entries{
		Timings: Timings{DNS: 42.0, Wait: 100.0},
	}
	tp := e.GetTimings()
	assert.Equal(t, 42.0, tp.GetDNS())
	assert.Equal(t, 100.0, tp.GetWait())
}

func TestStandardEntries_GetPageref(t *testing.T) {
	e := &Entries{Pageref: "page_2"}
	assert.Equal(t, "page_2", e.GetPageref())
}

func TestStandardEntries_GetPageref_Empty(t *testing.T) {
	e := &Entries{}
	assert.Equal(t, "", e.GetPageref())
}

func TestStandardEntries_ToStandard(t *testing.T) {
	e := &Entries{Time: 55.5, Pageref: "ref"}
	std := e.ToStandard()
	assert.Equal(t, 55.5, std.Time)
	assert.Equal(t, "ref", std.Pageref)
	// ToStandard returns a copy, not the same pointer
	assert.NotSame(t, e, &std)
}

// ---------------------------------------------------------------------------
// Request (RequestProvider) tests
// ---------------------------------------------------------------------------

func TestStandardRequest_GetMethod(t *testing.T) {
	r := &Request{Method: "DELETE"}
	assert.Equal(t, "DELETE", r.GetMethod())
}

func TestStandardRequest_GetURL(t *testing.T) {
	r := &Request{URL: "https://api.example.com/v1/resource"}
	assert.Equal(t, "https://api.example.com/v1/resource", r.GetURL())
}

func TestStandardRequest_GetHTTPVersion(t *testing.T) {
	r := &Request{HTTPVersion: "HTTP/2.0"}
	assert.Equal(t, "HTTP/2.0", r.GetHTTPVersion())
}

func TestStandardRequest_GetHeaders(t *testing.T) {
	r := &Request{
		Headers: []Headers{
			{Name: "Accept", Value: "*/*"},
			{Name: "Authorization", Value: "Bearer token123"},
		},
	}
	hp := r.GetHeaders()
	assert.Len(t, hp, 2)
	assert.Equal(t, "Accept", hp[0].GetName())
	assert.Equal(t, "*/*", hp[0].GetValue())
	assert.Equal(t, "Authorization", hp[1].GetName())
	assert.Equal(t, "Bearer token123", hp[1].GetValue())
}

func TestStandardRequest_GetHeaders_Empty(t *testing.T) {
	r := &Request{}
	hp := r.GetHeaders()
	assert.Len(t, hp, 0)
}

func TestStandardRequest_GetCookies(t *testing.T) {
	r := &Request{
		Cookies: []Cookie{
			{Name: "c1", Value: "v1"},
			{Name: "c2", Value: "v2"},
		},
	}
	cp := r.GetCookies()
	assert.Len(t, cp, 2)
	assert.Equal(t, "c1", cp[0].GetName())
	assert.Equal(t, "v1", cp[0].GetValue())
	assert.Equal(t, "c2", cp[1].GetName())
	assert.Equal(t, "v2", cp[1].GetValue())
}

func TestStandardRequest_GetCookies_Empty(t *testing.T) {
	r := &Request{}
	cp := r.GetCookies()
	assert.Len(t, cp, 0)
}

func TestStandardRequest_GetBodySize(t *testing.T) {
	r := &Request{BodySize: 512}
	assert.Equal(t, 512, r.GetBodySize())
}

func TestStandardRequest_GetHeadersSize(t *testing.T) {
	r := &Request{HeadersSize: 1024}
	assert.Equal(t, 1024, r.GetHeadersSize())
}

func TestStandardRequest_GetQueryString(t *testing.T) {
	r := &Request{
		QueryString: []QueryString{
			{Name: "q", Value: "search"},
			{Name: "limit", Value: "10"},
		},
	}
	qs := r.GetQueryString()
	assert.Len(t, qs, 2)
	assert.Equal(t, "q", qs[0].Name)
	assert.Equal(t, "search", qs[0].Value)
	assert.Equal(t, "limit", qs[1].Name)
	assert.Equal(t, "10", qs[1].Value)
}

func TestStandardRequest_GetQueryString_Empty(t *testing.T) {
	r := &Request{}
	qs := r.GetQueryString()
	assert.Len(t, qs, 0)
}

func TestStandardRequest_GetPostData(t *testing.T) {
	pd := &PostData{MimeType: "application/json", Text: `{"key":"val"}`}
	r := &Request{PostData: pd}
	result := r.GetPostData()
	assert.NotNil(t, result)
	assert.Equal(t, "application/json", result.MimeType)
	assert.Equal(t, `{"key":"val"}`, result.Text)
}

func TestStandardRequest_GetPostData_Nil(t *testing.T) {
	r := &Request{}
	assert.Nil(t, r.GetPostData())
}

func TestStandardRequest_ToStandard(t *testing.T) {
	r := &Request{Method: "PUT", URL: "https://example.com/item", BodySize: 99}
	std := r.ToStandard()
	assert.Equal(t, "PUT", std.Method)
	assert.Equal(t, "https://example.com/item", std.URL)
	assert.Equal(t, 99, std.BodySize)
}

// ---------------------------------------------------------------------------
// Response (ResponseProvider) tests
// ---------------------------------------------------------------------------

func TestStandardResponse_GetStatus(t *testing.T) {
	r := &Response{Status: 201}
	assert.Equal(t, 201, r.GetStatus())
}

func TestStandardResponse_GetStatusText(t *testing.T) {
	r := &Response{StatusText: "Created"}
	assert.Equal(t, "Created", r.GetStatusText())
}

func TestStandardResponse_GetHTTPVersion(t *testing.T) {
	r := &Response{HTTPVersion: "HTTP/1.0"}
	assert.Equal(t, "HTTP/1.0", r.GetHTTPVersion())
}

func TestStandardResponse_GetHeaders(t *testing.T) {
	r := &Response{
		Headers: []Headers{
			{Name: "Content-Type", Value: "text/html"},
			{Name: "Set-Cookie", Value: "id=a3fWa"},
		},
	}
	hp := r.GetHeaders()
	assert.Len(t, hp, 2)
	assert.Equal(t, "Content-Type", hp[0].GetName())
	assert.Equal(t, "text/html", hp[0].GetValue())
	assert.Equal(t, "Set-Cookie", hp[1].GetName())
	assert.Equal(t, "id=a3fWa", hp[1].GetValue())
}

func TestStandardResponse_GetHeaders_Empty(t *testing.T) {
	r := &Response{}
	hp := r.GetHeaders()
	assert.Len(t, hp, 0)
}

func TestStandardResponse_GetCookies(t *testing.T) {
	r := &Response{
		Cookies: []Cookie{
			{Name: "rc", Value: "rv"},
		},
	}
	cp := r.GetCookies()
	assert.Len(t, cp, 1)
	assert.Equal(t, "rc", cp[0].GetName())
	assert.Equal(t, "rv", cp[0].GetValue())
}

func TestStandardResponse_GetCookies_Empty(t *testing.T) {
	r := &Response{}
	cp := r.GetCookies()
	assert.Len(t, cp, 0)
}

func TestStandardResponse_GetContent(t *testing.T) {
	r := &Response{
		Content: Content{Size: 2048, MimeType: "text/plain"},
	}
	cp := r.GetContent()
	assert.Equal(t, 2048, cp.GetSize())
	assert.Equal(t, "text/plain", cp.GetMimeType())
}

func TestStandardResponse_GetBodySize(t *testing.T) {
	r := &Response{BodySize: 4096}
	assert.Equal(t, 4096, r.GetBodySize())
}

func TestStandardResponse_GetHeadersSize(t *testing.T) {
	r := &Response{HeadersSize: 800}
	assert.Equal(t, 800, r.GetHeadersSize())
}

func TestStandardResponse_ToStandard(t *testing.T) {
	r := &Response{Status: 301, StatusText: "Moved", BodySize: 0}
	std := r.ToStandard()
	assert.Equal(t, 301, std.Status)
	assert.Equal(t, "Moved", std.StatusText)
	assert.Equal(t, 0, std.BodySize)
}

// ---------------------------------------------------------------------------
// Headers (HeaderProvider) tests
// ---------------------------------------------------------------------------

func TestStandardHeaders_GetName(t *testing.T) {
	h := &Headers{Name: "X-Request-ID"}
	assert.Equal(t, "X-Request-ID", h.GetName())
}

func TestStandardHeaders_GetValue(t *testing.T) {
	h := &Headers{Value: "abc-def-ghi"}
	assert.Equal(t, "abc-def-ghi", h.GetValue())
}

func TestStandardHeaders_ToStandard(t *testing.T) {
	h := &Headers{Name: "Cache-Control", Value: "no-cache", Comment: "cc"}
	std := h.ToStandard()
	assert.Equal(t, "Cache-Control", std.Name)
	assert.Equal(t, "no-cache", std.Value)
	assert.Equal(t, "cc", std.Comment)
}

// ---------------------------------------------------------------------------
// Cookie (CookieProvider) tests
// ---------------------------------------------------------------------------

func TestStandardCookie_GetName(t *testing.T) {
	c := &Cookie{Name: "user_session"}
	assert.Equal(t, "user_session", c.GetName())
}

func TestStandardCookie_GetValue(t *testing.T) {
	c := &Cookie{Value: "secret-token"}
	assert.Equal(t, "secret-token", c.GetValue())
}

func TestStandardCookie_GetDomain(t *testing.T) {
	c := &Cookie{Domain: ".api.example.com"}
	assert.Equal(t, ".api.example.com", c.GetDomain())
}

func TestStandardCookie_GetPath(t *testing.T) {
	c := &Cookie{Path: "/api/v1"}
	assert.Equal(t, "/api/v1", c.GetPath())
}

func TestStandardCookie_GetExpires(t *testing.T) {
	ts := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
	c := &Cookie{Expires: ts}
	assert.Equal(t, ts, c.GetExpires())
}

func TestStandardCookie_IsHTTPOnly(t *testing.T) {
	c := &Cookie{HTTPOnly: true}
	assert.True(t, c.IsHTTPOnly())
	c2 := &Cookie{HTTPOnly: false}
	assert.False(t, c2.IsHTTPOnly())
}

func TestStandardCookie_IsSecure(t *testing.T) {
	c := &Cookie{Secure: true}
	assert.True(t, c.IsSecure())
	c2 := &Cookie{Secure: false}
	assert.False(t, c2.IsSecure())
}

func TestStandardCookie_GetSameSite(t *testing.T) {
	c := &Cookie{SameSite: "None"}
	assert.Equal(t, "None", c.GetSameSite())
}

func TestStandardCookie_GetSameSite_Empty(t *testing.T) {
	c := &Cookie{}
	assert.Equal(t, "", c.GetSameSite())
}

func TestStandardCookie_ToStandard(t *testing.T) {
	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	c := &Cookie{
		Name:     "test",
		Value:    "val",
		Path:     "/",
		Domain:   ".test.com",
		Expires:  ts,
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
	}
	std := c.ToStandard()
	assert.Equal(t, "test", std.Name)
	assert.Equal(t, "val", std.Value)
	assert.Equal(t, "/", std.Path)
	assert.Equal(t, ".test.com", std.Domain)
	assert.Equal(t, ts, std.Expires)
	assert.True(t, std.HTTPOnly)
	assert.False(t, std.Secure)
	assert.Equal(t, "Lax", std.SameSite)
}

// ---------------------------------------------------------------------------
// Content (ContentProvider) tests
// ---------------------------------------------------------------------------

func TestStandardContent_GetSize(t *testing.T) {
	c := &Content{Size: 5000}
	assert.Equal(t, 5000, c.GetSize())
}

func TestStandardContent_GetMimeType(t *testing.T) {
	c := &Content{MimeType: "image/png"}
	assert.Equal(t, "image/png", c.GetMimeType())
}

func TestStandardContent_GetText(t *testing.T) {
	c := &Content{Text: "hello world"}
	assert.Equal(t, "hello world", c.GetText())
}

func TestStandardContent_GetText_Empty(t *testing.T) {
	c := &Content{}
	assert.Equal(t, "", c.GetText())
}

func TestStandardContent_GetEncoding(t *testing.T) {
	c := &Content{Encoding: "base64"}
	assert.Equal(t, "base64", c.GetEncoding())
}

func TestStandardContent_GetEncoding_Empty(t *testing.T) {
	c := &Content{}
	assert.Equal(t, "", c.GetEncoding())
}

func TestStandardContent_GetCompression(t *testing.T) {
	c := &Content{Compression: -500}
	assert.Equal(t, -500, c.GetCompression())
}

func TestStandardContent_GetCompression_Zero(t *testing.T) {
	c := &Content{}
	assert.Equal(t, 0, c.GetCompression())
}

func TestStandardContent_ToStandard(t *testing.T) {
	c := &Content{Size: 100, MimeType: "text/html", Text: "<html>", Encoding: "utf-8", Compression: 50}
	std := c.ToStandard()
	assert.Equal(t, 100, std.Size)
	assert.Equal(t, "text/html", std.MimeType)
	assert.Equal(t, "<html>", std.Text)
	assert.Equal(t, "utf-8", std.Encoding)
	assert.Equal(t, 50, std.Compression)
}

// ---------------------------------------------------------------------------
// Timings (TimingsProvider) tests
// ---------------------------------------------------------------------------

func TestStandardTimings_GetBlocked(t *testing.T) {
	tm := &Timings{Blocked: 11.1}
	assert.Equal(t, 11.1, tm.GetBlocked())
}

func TestStandardTimings_GetDNS(t *testing.T) {
	tm := &Timings{DNS: 22.2}
	assert.Equal(t, 22.2, tm.GetDNS())
}

func TestStandardTimings_GetConnect(t *testing.T) {
	tm := &Timings{Connect: 33.3}
	assert.Equal(t, 33.3, tm.GetConnect())
}

func TestStandardTimings_GetSend(t *testing.T) {
	tm := &Timings{Send: 4.4}
	assert.Equal(t, 4.4, tm.GetSend())
}

func TestStandardTimings_GetWait(t *testing.T) {
	tm := &Timings{Wait: 55.5}
	assert.Equal(t, 55.5, tm.GetWait())
}

func TestStandardTimings_GetReceive(t *testing.T) {
	tm := &Timings{Receive: 66.6}
	assert.Equal(t, 66.6, tm.GetReceive())
}

func TestStandardTimings_GetSSL(t *testing.T) {
	tm := &Timings{Ssl: 77.7}
	assert.Equal(t, 77.7, tm.GetSSL())
}

func TestStandardTimings_GetSSL_Zero(t *testing.T) {
	tm := &Timings{}
	assert.Equal(t, 0.0, tm.GetSSL())
}

func TestStandardTimings_ToStandard(t *testing.T) {
	tm := &Timings{Blocked: 1, DNS: 2, Connect: 3, Ssl: 4, Send: 5, Wait: 6, Receive: 7}
	std := tm.ToStandard()
	assert.Equal(t, 1.0, std.Blocked)
	assert.Equal(t, 2.0, std.DNS)
	assert.Equal(t, 3.0, std.Connect)
	assert.Equal(t, 4.0, std.Ssl)
	assert.Equal(t, 5.0, std.Send)
	assert.Equal(t, 6.0, std.Wait)
	assert.Equal(t, 7.0, std.Receive)
}

// ---------------------------------------------------------------------------
// Pages (PageProvider) tests
// ---------------------------------------------------------------------------

func TestStandardPages_GetID(t *testing.T) {
	p := &Pages{ID: "page_abc"}
	assert.Equal(t, "page_abc", p.GetID())
}

func TestStandardPages_GetTitle(t *testing.T) {
	p := &Pages{Title: "My Page Title"}
	assert.Equal(t, "My Page Title", p.GetTitle())
}

func TestStandardPages_GetStartedDateTime(t *testing.T) {
	ts := time.Date(2024, 3, 10, 12, 0, 0, 0, time.UTC)
	p := &Pages{StartedDateTime: ts}
	assert.Equal(t, ts, p.GetStartedDateTime())
}

func TestStandardPages_GetPageTimings(t *testing.T) {
	p := &Pages{
		PageTimings: PageTimings{OnContentLoad: 200.0, OnLoad: 500.0},
	}
	ptp := p.GetPageTimings()
	assert.Equal(t, 200.0, ptp.GetOnContentLoad())
	assert.Equal(t, 500.0, ptp.GetOnLoad())
}

func TestStandardPages_ToStandard(t *testing.T) {
	ts := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	p := &Pages{ID: "p1", Title: "T1", StartedDateTime: ts}
	std := p.ToStandard()
	assert.Equal(t, "p1", std.ID)
	assert.Equal(t, "T1", std.Title)
	assert.Equal(t, ts, std.StartedDateTime)
}

// ---------------------------------------------------------------------------
// PageTimings (PageTimingsProvider) tests
// ---------------------------------------------------------------------------

func TestStandardPageTimings_GetOnContentLoad(t *testing.T) {
	pt := &PageTimings{OnContentLoad: 350.7}
	assert.Equal(t, 350.7, pt.GetOnContentLoad())
}

func TestStandardPageTimings_GetOnLoad(t *testing.T) {
	pt := &PageTimings{OnLoad: 800.9}
	assert.Equal(t, 800.9, pt.GetOnLoad())
}

func TestStandardPageTimings_ToStandard(t *testing.T) {
	pt := &PageTimings{OnContentLoad: 100.0, OnLoad: 200.0, Comment: "pt comment"}
	std := pt.ToStandard()
	assert.Equal(t, 100.0, std.OnContentLoad)
	assert.Equal(t, 200.0, std.OnLoad)
	assert.Equal(t, "pt comment", std.Comment)
}

// ---------------------------------------------------------------------------
// Integration: use the full test fixture to verify providers interact correctly
// ---------------------------------------------------------------------------

func TestStandardHar_FullFixture_Integration(t *testing.T) {
	h := newTestHar()

	// HARProvider
	assert.Equal(t, "1.2", h.GetVersion())
	assert.Equal(t, "test-creator", h.GetCreator().Name)
	assert.Equal(t, "test-browser", h.GetBrowser().Name)

	// EntryProvider
	entries := h.GetEntries()
	assert.Len(t, entries, 1)
	entry := entries[0]
	assert.Equal(t, time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC), entry.GetStartedDateTime())
	assert.Equal(t, 123.45, entry.GetTime())
	assert.Equal(t, "page_1", entry.GetPageref())

	// Entry -> RequestProvider
	req := entry.GetRequest()
	assert.Equal(t, "GET", req.GetMethod())
	assert.Equal(t, "https://example.com/api/users", req.GetURL())
	assert.Equal(t, "HTTP/1.1", req.GetHTTPVersion())
	assert.Equal(t, 256, req.GetHeadersSize())
	assert.Equal(t, 128, req.GetBodySize())

	// Request -> HeaderProvider
	reqHeaders := req.GetHeaders()
	assert.Len(t, reqHeaders, 1)
	assert.Equal(t, "Content-Type", reqHeaders[0].GetName())
	assert.Equal(t, "application/json", reqHeaders[0].GetValue())

	// Request -> CookieProvider
	reqCookies := req.GetCookies()
	assert.Len(t, reqCookies, 1)
	assert.Equal(t, "session_id", reqCookies[0].GetName())
	assert.Equal(t, "abc123", reqCookies[0].GetValue())
	assert.Equal(t, ".example.com", reqCookies[0].GetDomain())
	assert.Equal(t, "/", reqCookies[0].GetPath())
	assert.True(t, reqCookies[0].IsHTTPOnly())
	assert.True(t, reqCookies[0].IsSecure())
	assert.Equal(t, "Strict", reqCookies[0].GetSameSite())

	// Request -> QueryString
	qs := req.GetQueryString()
	assert.Len(t, qs, 1)
	assert.Equal(t, "page", qs[0].Name)
	assert.Equal(t, "1", qs[0].Value)

	// Request -> PostData
	pd := req.GetPostData()
	assert.NotNil(t, pd)
	assert.Equal(t, "application/x-www-form-urlencoded", pd.MimeType)
	assert.Equal(t, "key=value", pd.Text)

	// Entry -> ResponseProvider
	resp := entry.GetResponse()
	assert.Equal(t, 200, resp.GetStatus())
	assert.Equal(t, "OK", resp.GetStatusText())
	assert.Equal(t, "HTTP/1.1", resp.GetHTTPVersion())
	assert.Equal(t, 512, resp.GetHeadersSize())
	assert.Equal(t, 2048, resp.GetBodySize())

	// Response -> HeaderProvider
	respHeaders := resp.GetHeaders()
	assert.Len(t, respHeaders, 1)
	assert.Equal(t, "X-Custom", respHeaders[0].GetName())
	assert.Equal(t, "resp-val", respHeaders[0].GetValue())

	// Response -> CookieProvider
	respCookies := resp.GetCookies()
	assert.Len(t, respCookies, 1)
	assert.Equal(t, "track", respCookies[0].GetName())
	assert.Equal(t, "xyz789", respCookies[0].GetValue())
	assert.False(t, respCookies[0].IsHTTPOnly())
	assert.False(t, respCookies[0].IsSecure())
	assert.Equal(t, "Lax", respCookies[0].GetSameSite())

	// Response -> ContentProvider
	content := resp.GetContent()
	assert.Equal(t, 1024, content.GetSize())
	assert.Equal(t, "application/json", content.GetMimeType())
	assert.Equal(t, `{"status":"ok"}`, content.GetText())
	assert.Equal(t, "base64", content.GetEncoding())
	assert.Equal(t, -256, content.GetCompression())

	// Entry -> TimingsProvider
	timings := entry.GetTimings()
	assert.Equal(t, 10.1, timings.GetBlocked())
	assert.Equal(t, 20.2, timings.GetDNS())
	assert.Equal(t, 30.3, timings.GetConnect())
	assert.Equal(t, 40.4, timings.GetSSL())
	assert.Equal(t, 5.5, timings.GetSend())
	assert.Equal(t, 50.5, timings.GetWait())
	assert.Equal(t, 15.6, timings.GetReceive())

	// PageProvider
	pages := h.GetPages()
	assert.Len(t, pages, 1)
	page := pages[0]
	assert.Equal(t, "page_1", page.GetID())
	assert.Equal(t, "Test Page", page.GetTitle())
	assert.Equal(t, time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC), page.GetStartedDateTime())

	// Page -> PageTimingsProvider
	pt := page.GetPageTimings()
	assert.Equal(t, 150.5, pt.GetOnContentLoad())
	assert.Equal(t, 300.2, pt.GetOnLoad())
}

// ---------------------------------------------------------------------------
// Interface compliance checks (compile-time verification)
// ---------------------------------------------------------------------------

func TestStandardHar_ImplementsHARProvider(t *testing.T) {
	var _ HARProvider = (*Har)(nil)
}

func TestStandardEntries_ImplementsEntryProvider(t *testing.T) {
	var _ EntryProvider = (*Entries)(nil)
}

func TestStandardRequest_ImplementsRequestProvider(t *testing.T) {
	var _ RequestProvider = (*Request)(nil)
}

func TestStandardResponse_ImplementsResponseProvider(t *testing.T) {
	var _ ResponseProvider = (*Response)(nil)
}

func TestStandardHeaders_ImplementsHeaderProvider(t *testing.T) {
	var _ HeaderProvider = (*Headers)(nil)
}

func TestStandardCookie_ImplementsCookieProvider(t *testing.T) {
	var _ CookieProvider = (*Cookie)(nil)
}

func TestStandardContent_ImplementsContentProvider(t *testing.T) {
	var _ ContentProvider = (*Content)(nil)
}

func TestStandardTimings_ImplementsTimingsProvider(t *testing.T) {
	var _ TimingsProvider = (*Timings)(nil)
}

func TestStandardPages_ImplementsPageProvider(t *testing.T) {
	var _ PageProvider = (*Pages)(nil)
}

func TestStandardPageTimings_ImplementsPageTimingsProvider(t *testing.T) {
	var _ PageTimingsProvider = (*PageTimings)(nil)
}

// ---------------------------------------------------------------------------
// Edge cases: zero-value structs
// ---------------------------------------------------------------------------

func TestStandardEntries_ZeroValue(t *testing.T) {
	e := &Entries{}
	assert.Equal(t, time.Time{}, e.GetStartedDateTime())
	assert.Equal(t, 0.0, e.GetTime())
	assert.Equal(t, "", e.GetPageref())
}

func TestStandardRequest_ZeroValue(t *testing.T) {
	r := &Request{}
	assert.Equal(t, "", r.GetMethod())
	assert.Equal(t, "", r.GetURL())
	assert.Equal(t, "", r.GetHTTPVersion())
	assert.Equal(t, 0, r.GetBodySize())
	assert.Equal(t, 0, r.GetHeadersSize())
	assert.Empty(t, r.GetHeaders())
	assert.Empty(t, r.GetCookies())
	assert.Empty(t, r.GetQueryString())
	assert.Nil(t, r.GetPostData())
}

func TestStandardResponse_ZeroValue(t *testing.T) {
	r := &Response{}
	assert.Equal(t, 0, r.GetStatus())
	assert.Equal(t, "", r.GetStatusText())
	assert.Equal(t, "", r.GetHTTPVersion())
	assert.Equal(t, 0, r.GetBodySize())
	assert.Equal(t, 0, r.GetHeadersSize())
	assert.Empty(t, r.GetHeaders())
	assert.Empty(t, r.GetCookies())
}

func TestStandardCookie_ZeroValue(t *testing.T) {
	c := &Cookie{}
	assert.Equal(t, "", c.GetName())
	assert.Equal(t, "", c.GetValue())
	assert.Equal(t, "", c.GetDomain())
	assert.Equal(t, "", c.GetPath())
	assert.Equal(t, time.Time{}, c.GetExpires())
	assert.False(t, c.IsHTTPOnly())
	assert.False(t, c.IsSecure())
	assert.Equal(t, "", c.GetSameSite())
}

func TestStandardContent_ZeroValue(t *testing.T) {
	c := &Content{}
	assert.Equal(t, 0, c.GetSize())
	assert.Equal(t, "", c.GetMimeType())
	assert.Equal(t, "", c.GetText())
	assert.Equal(t, "", c.GetEncoding())
	assert.Equal(t, 0, c.GetCompression())
}

func TestStandardTimings_ZeroValue(t *testing.T) {
	tm := &Timings{}
	assert.Equal(t, 0.0, tm.GetBlocked())
	assert.Equal(t, 0.0, tm.GetDNS())
	assert.Equal(t, 0.0, tm.GetConnect())
	assert.Equal(t, 0.0, tm.GetSend())
	assert.Equal(t, 0.0, tm.GetWait())
	assert.Equal(t, 0.0, tm.GetReceive())
	assert.Equal(t, 0.0, tm.GetSSL())
}

func TestStandardPages_ZeroValue(t *testing.T) {
	p := &Pages{}
	assert.Equal(t, "", p.GetID())
	assert.Equal(t, "", p.GetTitle())
	assert.Equal(t, time.Time{}, p.GetStartedDateTime())
}

func TestStandardPageTimings_ZeroValue(t *testing.T) {
	pt := &PageTimings{}
	assert.Equal(t, 0.0, pt.GetOnContentLoad())
	assert.Equal(t, 0.0, pt.GetOnLoad())
}

func TestStandardHar_ZeroValue(t *testing.T) {
	h := &Har{}
	assert.Equal(t, "", h.GetVersion())
	assert.Empty(t, h.GetEntries())
	assert.Empty(t, h.GetPages())
}

// ---------------------------------------------------------------------------
// ToStandard returns copies (value semantics)
// ---------------------------------------------------------------------------

func TestStandardEntries_ToStandard_IsCopy(t *testing.T) {
	e := &Entries{Time: 100.0, Pageref: "ref1"}
	std := e.ToStandard()
	std.Time = 999.0
	std.Pageref = "changed"
	// Original should be unaffected
	assert.Equal(t, 100.0, e.Time)
	assert.Equal(t, "ref1", e.Pageref)
}

func TestStandardRequest_ToStandard_IsCopy(t *testing.T) {
	r := &Request{Method: "GET", URL: "https://original.com"}
	std := r.ToStandard()
	std.Method = "POST"
	std.URL = "https://changed.com"
	assert.Equal(t, "GET", r.Method)
	assert.Equal(t, "https://original.com", r.URL)
}

func TestStandardResponse_ToStandard_IsCopy(t *testing.T) {
	r := &Response{Status: 200, StatusText: "OK"}
	std := r.ToStandard()
	std.Status = 500
	std.StatusText = "Error"
	assert.Equal(t, 200, r.Status)
	assert.Equal(t, "OK", r.StatusText)
}

func TestStandardHeaders_ToStandard_IsCopy(t *testing.T) {
	h := &Headers{Name: "X-Orig", Value: "orig"}
	std := h.ToStandard()
	std.Name = "X-Changed"
	std.Value = "changed"
	assert.Equal(t, "X-Orig", h.Name)
	assert.Equal(t, "orig", h.Value)
}

func TestStandardCookie_ToStandard_IsCopy(t *testing.T) {
	c := &Cookie{Name: "orig", Value: "val"}
	std := c.ToStandard()
	std.Name = "changed"
	std.Value = "mod"
	assert.Equal(t, "orig", c.Name)
	assert.Equal(t, "val", c.Value)
}

func TestStandardContent_ToStandard_IsCopy(t *testing.T) {
	c := &Content{Size: 100, MimeType: "text/plain"}
	std := c.ToStandard()
	std.Size = 999
	std.MimeType = "changed"
	assert.Equal(t, 100, c.Size)
	assert.Equal(t, "text/plain", c.MimeType)
}

func TestStandardTimings_ToStandard_IsCopy(t *testing.T) {
	tm := &Timings{DNS: 42.0}
	std := tm.ToStandard()
	std.DNS = 99.0
	assert.Equal(t, 42.0, tm.DNS)
}

func TestStandardPages_ToStandard_IsCopy(t *testing.T) {
	p := &Pages{ID: "orig", Title: "Original"}
	std := p.ToStandard()
	std.ID = "changed"
	std.Title = "Changed"
	assert.Equal(t, "orig", p.ID)
	assert.Equal(t, "Original", p.Title)
}

func TestStandardPageTimings_ToStandard_IsCopy(t *testing.T) {
	pt := &PageTimings{OnContentLoad: 100.0, OnLoad: 200.0}
	std := pt.ToStandard()
	std.OnContentLoad = 999.0
	std.OnLoad = 888.0
	assert.Equal(t, 100.0, pt.OnContentLoad)
	assert.Equal(t, 200.0, pt.OnLoad)
}

// ---------------------------------------------------------------------------
// Multiple entries/pages
// ---------------------------------------------------------------------------

func TestStandardHar_MultipleEntries(t *testing.T) {
	h := &Har{
		Log: Log{
			Version: "1.2",
			Entries: []Entries{
				{Time: 10.0, Pageref: "p1"},
				{Time: 20.0, Pageref: "p2"},
				{Time: 30.0, Pageref: ""},
			},
		},
	}
	entries := h.GetEntries()
	assert.Len(t, entries, 3)
	assert.Equal(t, 10.0, entries[0].GetTime())
	assert.Equal(t, 20.0, entries[1].GetTime())
	assert.Equal(t, 30.0, entries[2].GetTime())
	assert.Equal(t, "p1", entries[0].GetPageref())
	assert.Equal(t, "p2", entries[1].GetPageref())
	assert.Equal(t, "", entries[2].GetPageref())
}

func TestStandardHar_MultiplePages(t *testing.T) {
	h := &Har{
		Log: Log{
			Version: "1.2",
			Pages: []Pages{
				{ID: "pg1", Title: "First"},
				{ID: "pg2", Title: "Second"},
			},
		},
	}
	pages := h.GetPages()
	assert.Len(t, pages, 2)
	assert.Equal(t, "pg1", pages[0].GetID())
	assert.Equal(t, "pg2", pages[1].GetID())
	assert.Equal(t, "First", pages[0].GetTitle())
	assert.Equal(t, "Second", pages[1].GetTitle())
}

// ---------------------------------------------------------------------------
// Cookie Expires edge case
// ---------------------------------------------------------------------------

func TestStandardCookie_GetExpires_ZeroTime(t *testing.T) {
	c := &Cookie{}
	assert.Equal(t, time.Time{}, c.GetExpires())
}

// ---------------------------------------------------------------------------
// Request with multiple headers and cookies
// ---------------------------------------------------------------------------

func TestStandardRequest_MultipleHeadersAndCookies(t *testing.T) {
	r := &Request{
		Headers: []Headers{
			{Name: "H1", Value: "V1"},
			{Name: "H2", Value: "V2"},
			{Name: "H3", Value: "V3"},
		},
		Cookies: []Cookie{
			{Name: "C1", Value: "CV1"},
			{Name: "C2", Value: "CV2"},
		},
	}
	headers := r.GetHeaders()
	assert.Len(t, headers, 3)
	for i, h := range headers {
		assert.Equal(t, r.Headers[i].Name, h.GetName())
		assert.Equal(t, r.Headers[i].Value, h.GetValue())
	}
	cookies := r.GetCookies()
	assert.Len(t, cookies, 2)
	for i, c := range cookies {
		assert.Equal(t, r.Cookies[i].Name, c.GetName())
		assert.Equal(t, r.Cookies[i].Value, c.GetValue())
	}
}

func TestStandardResponse_MultipleHeadersAndCookies(t *testing.T) {
	r := &Response{
		Headers: []Headers{
			{Name: "RH1", Value: "RV1"},
			{Name: "RH2", Value: "RV2"},
		},
		Cookies: []Cookie{
			{Name: "RC1", Value: "RCV1"},
			{Name: "RC2", Value: "RCV2"},
			{Name: "RC3", Value: "RCV3"},
		},
	}
	headers := r.GetHeaders()
	assert.Len(t, headers, 2)
	cookies := r.GetCookies()
	assert.Len(t, cookies, 3)
}

func TestStandardHarNilReceiverMethods(t *testing.T) {
	var h *Har

	assert.Equal(t, "", h.GetVersion())
	assert.Equal(t, Creator{}, h.GetCreator())
	assert.Equal(t, Browser{}, h.GetBrowser())
	assert.Nil(t, h.GetEntries())
	assert.Nil(t, h.GetPages())
	assert.Nil(t, h.ToStandard())
}

func TestStandardProviderNilReceiverMethods(t *testing.T) {
	var entry *Entries
	assert.True(t, entry.GetStartedDateTime().IsZero())
	assert.Equal(t, 0.0, entry.GetTime())
	assert.Nil(t, entry.GetRequest())
	assert.Nil(t, entry.GetResponse())
	assert.Nil(t, entry.GetTimings())
	assert.Equal(t, "", entry.GetPageref())
	assert.Equal(t, Entries{}, entry.ToStandard())

	var request *Request
	assert.Equal(t, "", request.GetMethod())
	assert.Equal(t, "", request.GetURL())
	assert.Equal(t, "", request.GetHTTPVersion())
	assert.Nil(t, request.GetHeaders())
	assert.Nil(t, request.GetCookies())
	assert.Nil(t, request.GetQueryString())
	assert.Nil(t, request.GetPostData())
	assert.Equal(t, 0, request.GetBodySize())
	assert.Equal(t, 0, request.GetHeadersSize())
	assert.Equal(t, Request{}, request.ToStandard())

	var response *Response
	assert.Equal(t, 0, response.GetStatus())
	assert.Equal(t, "", response.GetStatusText())
	assert.Equal(t, "", response.GetHTTPVersion())
	assert.Nil(t, response.GetHeaders())
	assert.Nil(t, response.GetCookies())
	assert.Nil(t, response.GetContent())
	assert.Equal(t, 0, response.GetBodySize())
	assert.Equal(t, 0, response.GetHeadersSize())
	assert.Equal(t, Response{}, response.ToStandard())

	var header *Headers
	assert.Equal(t, "", header.GetName())
	assert.Equal(t, "", header.GetValue())
	assert.Equal(t, Headers{}, header.ToStandard())

	var cookie *Cookie
	assert.Equal(t, "", cookie.GetName())
	assert.Equal(t, "", cookie.GetValue())
	assert.Equal(t, "", cookie.GetDomain())
	assert.Equal(t, "", cookie.GetPath())
	assert.True(t, cookie.GetExpires().IsZero())
	assert.False(t, cookie.IsHTTPOnly())
	assert.False(t, cookie.IsSecure())
	assert.Equal(t, "", cookie.GetSameSite())
	assert.Equal(t, Cookie{}, cookie.ToStandard())

	var content *Content
	assert.Equal(t, 0, content.GetSize())
	assert.Equal(t, "", content.GetMimeType())
	assert.Equal(t, "", content.GetText())
	assert.Equal(t, "", content.GetEncoding())
	assert.Equal(t, 0, content.GetCompression())
	assert.Equal(t, Content{}, content.ToStandard())

	var timings *Timings
	assert.Equal(t, 0.0, timings.GetBlocked())
	assert.Equal(t, 0.0, timings.GetDNS())
	assert.Equal(t, 0.0, timings.GetConnect())
	assert.Equal(t, 0.0, timings.GetSend())
	assert.Equal(t, 0.0, timings.GetWait())
	assert.Equal(t, 0.0, timings.GetReceive())
	assert.Equal(t, 0.0, timings.GetSSL())
	assert.Equal(t, Timings{}, timings.ToStandard())

	var page *Pages
	assert.Equal(t, "", page.GetID())
	assert.Equal(t, "", page.GetTitle())
	assert.True(t, page.GetStartedDateTime().IsZero())
	assert.Nil(t, page.GetPageTimings())
	assert.Equal(t, Pages{}, page.ToStandard())

	var pageTimings *PageTimings
	assert.Equal(t, 0.0, pageTimings.GetOnContentLoad())
	assert.Equal(t, 0.0, pageTimings.GetOnLoad())
	assert.Equal(t, PageTimings{}, pageTimings.ToStandard())
}
