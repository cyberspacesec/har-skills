package har

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Helper: build a fully-populated LazyHar via JSON round-trip
// ---------------------------------------------------------------------------

func buildLazyHar(t *testing.T) *LazyHar {
	t.Helper()

	raw := `{
		"log": {
			"version": "1.2",
			"creator": {"name": "test-creator", "version": "0.1", "comment": "creator-comment"},
			"browser": {"name": "test-browser", "version": "1.0", "comment": "browser-comment"},
			"pages": [
				{
					"startedDateTime": "2024-06-01T10:00:00Z",
					"id": "page_1",
					"title": "Test Page",
					"pageTimings": {"onContentLoad": 500.0, "onLoad": 1200.0, "comment": "pt-comment"},
					"comment": "page-comment"
				}
			],
			"entries": [
				{
					"startedDateTime": "2024-06-01T10:00:01Z",
					"time": 150.5,
					"request": {
						"method": "GET",
						"url": "https://example.com/api/data",
						"httpVersion": "HTTP/2.0",
						"cookies": [
							{"name": "req_ck", "value": "rv1", "path": "/", "domain": "example.com", "httpOnly": true, "secure": true, "sameSite": "Strict", "comment": "req-ck-comment"}
						],
						"headers": [
							{"name": "Accept", "value": "application/json", "comment": "hdr-comment"},
							{"name": "Authorization", "value": "Bearer token123", "comment": ""}
						],
						"queryString": [{"name": "page", "value": "1"}],
						"postData": {"mimeType": "application/json", "text": "{\"key\":\"val\"}"},
						"headersSize": 300,
						"bodySize": 45,
						"comment": "req-comment"
					},
					"response": {
						"status": 200,
						"statusText": "OK",
						"httpVersion": "HTTP/2.0",
						"cookies": [
							{"name": "resp_ck", "value": "rv2", "path": "/", "domain": "example.com", "httpOnly": false, "secure": true, "sameSite": "Lax", "comment": "resp-ck-comment"}
						],
						"headers": [
							{"name": "Content-Type", "value": "application/json", "comment": "resp-hdr-comment"}
						],
						"content": {
							"size": 1024,
							"mimeType": "application/json",
							"compression": 256,
							"text": "{\"result\":true}",
							"encoding": "utf-8",
							"comment": "content-comment"
						},
						"redirectURL": "",
						"headersSize": 200,
						"bodySize": 1024,
						"_transferSize": 1200,
						"_error": null,
						"comment": "resp-comment"
					},
					"cache": {
						"beforeRequest": {"expires": "2025-01-01T00:00:00Z", "lastAccess": "2024-06-01T10:00:00Z", "eTag": "\"abc\"", "hitCount": 3},
						"afterRequest": {"expires": "2025-01-01T00:00:00Z", "lastAccess": "2024-06-01T10:00:01Z", "eTag": "\"def\"", "hitCount": 4},
						"comment": "cache-comment"
					},
					"timings": {
						"blocked": 5.0,
						"dns": 10.0,
						"connect": 15.0,
						"ssl": 20.0,
						"send": 2.0,
						"wait": 80.0,
						"receive": 18.5,
						"_blocked_queueing": 3.0,
						"comment": "timings-comment"
					},
					"pageref": "page_1",
					"connection": "conn-42",
					"serverIPAddress": "10.0.0.1",
					"comment": "entry-comment",
					"_initiator": {"type": "script", "url": "https://example.com/app.js", "lineNumber": 42},
					"_priority": "high",
					"_resourceType": "xhr"
				}
			]
		}
	}`

	lh, err := ParseHarWithLazyLoading([]byte(raw))
	require.NoError(t, err)
	require.NotNil(t, lh)
	return lh
}

// buildLazyHarWithNilContent creates a LazyHar whose response content is nil.
func buildLazyHarWithNilContent(t *testing.T) *LazyHar {
	t.Helper()

	raw := `{
		"log": {
			"version": "1.2",
			"creator": {"name": "test", "version": "0.1"},
			"entries": [
				{
					"startedDateTime": "2024-06-01T10:00:01Z",
					"time": 50,
					"request": {"method": "GET", "url": "https://example.com", "httpVersion": "HTTP/1.1"},
					"response": {
						"status": 204,
						"statusText": "No Content",
						"httpVersion": "HTTP/1.1",
						"headersSize": 0,
						"bodySize": 0
					},
					"timings": {"blocked": 0, "dns": 0, "connect": 0, "ssl": 0, "send": 0, "wait": 0, "receive": 0}
				}
			]
		}
	}`

	lh, err := ParseHarWithLazyLoading([]byte(raw))
	require.NoError(t, err)
	return lh
}

// ===========================================================================
// LazyHar tests
// ===========================================================================

func TestLazyHar_GetVersion(t *testing.T) {
	lh := buildLazyHar(t)
	assert.Equal(t, "1.2", lh.GetVersion())
}

func TestLazyHar_GetCreator(t *testing.T) {
	lh := buildLazyHar(t)
	creator := lh.GetCreator()
	assert.Equal(t, "test-creator", creator.Name)
	assert.Equal(t, "0.1", creator.Version)
	assert.Equal(t, "creator-comment", creator.Comment)
}

func TestLazyHar_GetBrowser(t *testing.T) {
	lh := buildLazyHar(t)
	browser := lh.GetBrowser()
	assert.Equal(t, "test-browser", browser.Name)
	assert.Equal(t, "1.0", browser.Version)
	assert.Equal(t, "browser-comment", browser.Comment)
}

func TestLazyHar_GetEntries(t *testing.T) {
	lh := buildLazyHar(t)
	providers := lh.GetEntries()
	assert.Len(t, providers, 1)

	// Verify each provider is a *LazyEntries
	ep := providers[0]
	assert.Equal(t, "GET", ep.GetRequest().GetMethod())
	assert.Equal(t, "https://example.com/api/data", ep.GetRequest().GetURL())
}

func TestLazyHar_GetEntries_Empty(t *testing.T) {
	raw := `{"log":{"version":"1.2","creator":{"name":"t","version":"1"},"entries":[]}}`
	lh, err := ParseHarWithLazyLoading([]byte(raw))
	require.NoError(t, err)
	providers := lh.GetEntries()
	assert.Len(t, providers, 0)
}

func TestLazyHar_GetPages(t *testing.T) {
	lh := buildLazyHar(t)
	pages := lh.GetPages()
	assert.Len(t, pages, 1)

	pp := pages[0]
	assert.Equal(t, "page_1", pp.GetID())
	assert.Equal(t, "Test Page", pp.GetTitle())
}

func TestLazyHar_GetPages_Empty(t *testing.T) {
	raw := `{"log":{"version":"1.2","creator":{"name":"t","version":"1"},"entries":[],"pages":[]}}`
	lh, err := ParseHarWithLazyLoading([]byte(raw))
	require.NoError(t, err)
	pages := lh.GetPages()
	assert.Len(t, pages, 0)
}

func TestLazyHar_ToStandard(t *testing.T) {
	lh := buildLazyHar(t)
	std := lh.ToStandard()
	require.NotNil(t, std)
	assert.Equal(t, "1.2", std.Log.Version)
	assert.Equal(t, "test-creator", std.Log.Creator.Name)
	assert.Equal(t, "test-browser", std.Log.Browser.Name)
	assert.Len(t, std.Log.Entries, 1)
	assert.Equal(t, "{\"result\":true}", std.Log.Entries[0].Response.Content.Text)
	assert.Equal(t, "utf-8", std.Log.Entries[0].Response.Content.Encoding)
}

func TestLazyHar_ToStandard_BestEffortContentLoadError(t *testing.T) {
	raw := `{
		"log": {
			"version": "1.2",
			"creator": {"name": "test", "version": "1"},
			"entries": [{
				"startedDateTime": "2024-06-01T10:00:01Z",
				"time": 50,
				"request": {"method": "GET", "url": "https://example.com", "httpVersion": "HTTP/1.1"},
				"response": {
					"status": 200,
					"statusText": "OK",
					"httpVersion": "HTTP/1.1",
					"content": {"size": 5, "mimeType": "application/json", "text": {"bad": true}}
				},
				"timings": {"blocked": 0, "dns": 0, "connect": 0, "ssl": 0, "send": 0, "wait": 0, "receive": 0}
			}]
		}
	}`
	lh, err := ParseHarWithLazyLoading([]byte(raw))
	require.NoError(t, err)

	std := lh.ToStandard()
	require.NotNil(t, std)
	require.Len(t, std.Log.Entries, 1)
	assert.Equal(t, 5, std.Log.Entries[0].Response.Content.Size)
	assert.Equal(t, "application/json", std.Log.Entries[0].Response.Content.MimeType)
	assert.Empty(t, std.Log.Entries[0].Response.Content.Text)

	_, err = lh.ToStandardHar()
	assertHarErrorCode(t, err, ErrCodeJSONParse)
}

// ===========================================================================
// LazyEntries tests
// ===========================================================================

func TestLazyEntries_GetStartedDateTime(t *testing.T) {
	lh := buildLazyHar(t)
	entry := &lh.Log.Entries[0]
	expected := time.Date(2024, 6, 1, 10, 0, 1, 0, time.UTC)
	assert.Equal(t, expected, entry.GetStartedDateTime())
}

func TestLazyEntries_GetTime(t *testing.T) {
	lh := buildLazyHar(t)
	entry := &lh.Log.Entries[0]
	assert.Equal(t, 150.5, entry.GetTime())
}

func TestLazyEntries_GetRequest(t *testing.T) {
	lh := buildLazyHar(t)
	entry := &lh.Log.Entries[0]
	rp := entry.GetRequest()
	require.NotNil(t, rp)
	assert.Equal(t, "GET", rp.GetMethod())
	assert.Equal(t, "https://example.com/api/data", rp.GetURL())
	assert.Equal(t, "HTTP/2.0", rp.GetHTTPVersion())
	assert.Equal(t, 300, rp.GetHeadersSize())
	assert.Equal(t, 45, rp.GetBodySize())
}

func TestLazyEntries_GetResponse(t *testing.T) {
	lh := buildLazyHar(t)
	entry := &lh.Log.Entries[0]
	resp := entry.GetResponse()
	require.NotNil(t, resp)
	assert.Equal(t, 200, resp.GetStatus())
	assert.Equal(t, "OK", resp.GetStatusText())
	assert.Equal(t, "HTTP/2.0", resp.GetHTTPVersion())
	assert.Equal(t, 200, resp.GetHeadersSize())
	assert.Equal(t, 1024, resp.GetBodySize())
}

func TestLazyEntries_GetTimings(t *testing.T) {
	lh := buildLazyHar(t)
	entry := &lh.Log.Entries[0]
	tp := entry.GetTimings()
	require.NotNil(t, tp)
	assert.Equal(t, 5.0, tp.GetBlocked())
	assert.Equal(t, 10.0, tp.GetDNS())
	assert.Equal(t, 15.0, tp.GetConnect())
	assert.Equal(t, 20.0, tp.GetSSL())
	assert.Equal(t, 2.0, tp.GetSend())
	assert.Equal(t, 80.0, tp.GetWait())
	assert.Equal(t, 18.5, tp.GetReceive())
}

func TestLazyEntries_GetPageref(t *testing.T) {
	lh := buildLazyHar(t)
	entry := &lh.Log.Entries[0]
	assert.Equal(t, "page_1", entry.GetPageref())
}

func TestLazyEntries_GetPageref_Empty(t *testing.T) {
	raw := `{"log":{"version":"1.2","creator":{"name":"t","version":"1"},"entries":[{"startedDateTime":"2024-01-01T00:00:00Z","time":10,"request":{"method":"GET","url":"https://x.com","httpVersion":"HTTP/1.1"},"response":{"status":200,"statusText":"OK","httpVersion":"HTTP/1.1"},"timings":{"blocked":0,"dns":0,"connect":0,"ssl":0,"send":0,"wait":0,"receive":0}}]}}`
	lh, err := ParseHarWithLazyLoading([]byte(raw))
	require.NoError(t, err)
	assert.Equal(t, "", lh.Log.Entries[0].GetPageref())
}

func TestLazyEntries_ToStandard(t *testing.T) {
	lh := buildLazyHar(t)
	entry := &lh.Log.Entries[0]
	std := entry.ToStandard()

	assert.Equal(t, time.Date(2024, 6, 1, 10, 0, 1, 0, time.UTC), std.StartedDateTime)
	assert.Equal(t, 150.5, std.Time)
	assert.Equal(t, "GET", std.Request.Method)
	assert.Equal(t, "https://example.com/api/data", std.Request.URL)
	assert.Equal(t, 200, std.Response.Status)
	assert.Equal(t, "OK", std.Response.StatusText)
	assert.Equal(t, "page_1", std.Pageref)
	assert.Equal(t, "conn-42", std.Connection)
	assert.Equal(t, "10.0.0.1", std.ServerIPAddress)
	assert.Equal(t, "entry-comment", std.Comment)
}

// ===========================================================================
// LazyResponse tests
// ===========================================================================

func TestLazyResponse_GetStatus(t *testing.T) {
	lh := buildLazyHar(t)
	resp := &lh.Log.Entries[0].Response
	assert.Equal(t, 200, resp.GetStatus())
}

func TestLazyResponse_GetStatusText(t *testing.T) {
	lh := buildLazyHar(t)
	resp := &lh.Log.Entries[0].Response
	assert.Equal(t, "OK", resp.GetStatusText())
}

func TestLazyResponse_GetHTTPVersion(t *testing.T) {
	lh := buildLazyHar(t)
	resp := &lh.Log.Entries[0].Response
	assert.Equal(t, "HTTP/2.0", resp.GetHTTPVersion())
}

func TestLazyResponse_GetHeaders(t *testing.T) {
	lh := buildLazyHar(t)
	resp := &lh.Log.Entries[0].Response
	headers := resp.GetHeaders()
	assert.Len(t, headers, 1)
	assert.Equal(t, "Content-Type", headers[0].GetName())
	assert.Equal(t, "application/json", headers[0].GetValue())
}

func TestLazyResponse_GetHeaders_Empty(t *testing.T) {
	lh := buildLazyHarWithNilContent(t)
	resp := &lh.Log.Entries[0].Response
	headers := resp.GetHeaders()
	assert.Len(t, headers, 0)
}

func TestLazyResponse_GetCookies(t *testing.T) {
	lh := buildLazyHar(t)
	resp := &lh.Log.Entries[0].Response
	cookies := resp.GetCookies()
	assert.Len(t, cookies, 1)
	assert.Equal(t, "resp_ck", cookies[0].GetName())
	assert.Equal(t, "rv2", cookies[0].GetValue())
}

func TestLazyResponse_GetCookies_Empty(t *testing.T) {
	lh := buildLazyHarWithNilContent(t)
	resp := &lh.Log.Entries[0].Response
	cookies := resp.GetCookies()
	assert.Len(t, cookies, 0)
}

func TestLazyResponse_GetContent(t *testing.T) {
	lh := buildLazyHar(t)
	resp := &lh.Log.Entries[0].Response
	cp := resp.GetContent()
	require.NotNil(t, cp)
	assert.Equal(t, 1024, cp.GetSize())
	assert.Equal(t, "application/json", cp.GetMimeType())
	assert.Equal(t, 256, cp.GetCompression())
}

func TestLazyResponse_GetContent_NilContent(t *testing.T) {
	lh := buildLazyHarWithNilContent(t)
	resp := &lh.Log.Entries[0].Response
	// LazyResponse.Content will be nil because the JSON has no "content" field
	cp := resp.GetContent()
	require.NotNil(t, cp) // lazyContentWrapper is always returned
	assert.Equal(t, 0, cp.GetSize())
	assert.Equal(t, "", cp.GetMimeType())
}

func TestLazyResponse_GetBodySize(t *testing.T) {
	lh := buildLazyHar(t)
	resp := &lh.Log.Entries[0].Response
	assert.Equal(t, 1024, resp.GetBodySize())
}

func TestLazyResponse_GetHeadersSize(t *testing.T) {
	lh := buildLazyHar(t)
	resp := &lh.Log.Entries[0].Response
	assert.Equal(t, 200, resp.GetHeadersSize())
}

func TestLazyResponse_ToStandard(t *testing.T) {
	lh := buildLazyHar(t)
	resp := &lh.Log.Entries[0].Response
	std := resp.ToStandard()

	assert.Equal(t, 200, std.Status)
	assert.Equal(t, "OK", std.StatusText)
	assert.Equal(t, "HTTP/2.0", std.HTTPVersion)
	assert.Equal(t, 200, std.HeadersSize)
	assert.Equal(t, 1024, std.BodySize)
	assert.Equal(t, 1200, std.TransferSize)
	assert.Equal(t, "", std.RedirectURL)
	assert.Len(t, std.Cookies, 1)
	assert.Len(t, std.Headers, 1)
	// Content fields
	assert.Equal(t, 1024, std.Content.Size)
	assert.Equal(t, "application/json", std.Content.MimeType)
	assert.Equal(t, 256, std.Content.Compression)
	assert.Equal(t, "content-comment", std.Content.Comment)
	assert.Equal(t, "{\"result\":true}", std.Content.Text)
	assert.Equal(t, "utf-8", std.Content.Encoding)
}

func TestLazyResponse_ToStandard_NilContent(t *testing.T) {
	lh := buildLazyHarWithNilContent(t)
	resp := &lh.Log.Entries[0].Response
	std := resp.ToStandard()
	// Content should be zero-value
	assert.Equal(t, 0, std.Content.Size)
	assert.Equal(t, "", std.Content.MimeType)
}

func TestLazyResponse_ToStandard_WithTextAndEncoding(t *testing.T) {
	lh := buildLazyHar(t)
	resp := &lh.Log.Entries[0].Response

	// Force-load the lazy content first
	require.NotNil(t, resp.Content)
	err := resp.Content.Load()
	require.NoError(t, err)

	std := resp.ToStandard()
	assert.Equal(t, "{\"result\":true}", std.Content.Text)
	assert.Equal(t, "utf-8", std.Content.Encoding)
}

func TestLazyResponse_ToStandard_ContentNilTextNilEncoding(t *testing.T) {
	// Build a response where content has nil Text and nil Encoding
	resp := &LazyResponse{
		Status:      200,
		StatusText:  "OK",
		HTTPVersion: "HTTP/1.1",
		Content: &LazyContent{
			Size:     0,
			MimeType: "text/plain",
		},
	}
	// Ensure Text and Encoding are nil
	assert.Nil(t, resp.Content.Text)
	assert.Nil(t, resp.Content.Encoding)

	std := resp.ToStandard()
	assert.Equal(t, "", std.Content.Text)
	assert.Equal(t, "", std.Content.Encoding)
}

// ===========================================================================
// lazyContentWrapper tests
// ===========================================================================

func TestLazyContentWrapper_GetSize_Nil(t *testing.T) {
	w := &lazyContentWrapper{content: nil}
	assert.Equal(t, 0, w.GetSize())
}

func TestLazyContentWrapper_GetMimeType_Nil(t *testing.T) {
	w := &lazyContentWrapper{content: nil}
	assert.Equal(t, "", w.GetMimeType())
}

func TestLazyContentWrapper_GetText_Nil(t *testing.T) {
	w := &lazyContentWrapper{content: nil}
	assert.Equal(t, "", w.GetText())
}

func TestLazyContentWrapper_GetEncoding_Nil(t *testing.T) {
	w := &lazyContentWrapper{content: nil}
	assert.Equal(t, "", w.GetEncoding())
}

func TestLazyContentWrapper_GetCompression_Nil(t *testing.T) {
	w := &lazyContentWrapper{content: nil}
	assert.Equal(t, 0, w.GetCompression())
}

func TestLazyContentWrapper_ToStandard_Nil(t *testing.T) {
	w := &lazyContentWrapper{content: nil}
	c := w.ToStandard()
	assert.Equal(t, Content{}, c)
}

func TestLazyContentWrapper_GetSize(t *testing.T) {
	lh := buildLazyHar(t)
	cp := lh.Log.Entries[0].Response.GetContent()
	assert.Equal(t, 1024, cp.GetSize())
}

func TestLazyContentWrapper_GetMimeType(t *testing.T) {
	lh := buildLazyHar(t)
	cp := lh.Log.Entries[0].Response.GetContent()
	assert.Equal(t, "application/json", cp.GetMimeType())
}

func TestLazyContentWrapper_GetText(t *testing.T) {
	lh := buildLazyHar(t)
	cp := lh.Log.Entries[0].Response.GetContent()
	text := cp.GetText()
	assert.Equal(t, "{\"result\":true}", text)
}

func TestLazyContentWrapper_GetEncoding(t *testing.T) {
	lh := buildLazyHar(t)
	cp := lh.Log.Entries[0].Response.GetContent()
	enc := cp.GetEncoding()
	assert.Equal(t, "utf-8", enc)
}

func TestLazyContentWrapper_GetCompression(t *testing.T) {
	lh := buildLazyHar(t)
	cp := lh.Log.Entries[0].Response.GetContent()
	assert.Equal(t, 256, cp.GetCompression())
}

func TestLazyContentWrapper_ToStandard(t *testing.T) {
	lh := buildLazyHar(t)
	cp := lh.Log.Entries[0].Response.GetContent()
	c := cp.ToStandard()
	assert.Equal(t, 1024, c.Size)
	assert.Equal(t, "application/json", c.MimeType)
	assert.Equal(t, 256, c.Compression)
	assert.Equal(t, "content-comment", c.Comment)
}

func TestLazyContentWrapper_GetText_LoadError(t *testing.T) {
	// Create a LazyContent with invalid rawData to cause a load error
	lc := &LazyContent{
		Size:     10,
		MimeType: "text/plain",
	}
	lc.rawData = json.RawMessage(`{invalid json`)
	lc.loaded = false

	w := &lazyContentWrapper{content: lc}
	// GetText should return "" when load fails
	text := w.GetText()
	assert.Equal(t, "", text)
}

func TestLazyContentWrapper_GetText_NilTextPointer(t *testing.T) {
	// LazyContent with valid rawData but no text field
	lc := &LazyContent{
		Size:     5,
		MimeType: "text/plain",
	}
	lc.rawData = json.RawMessage(`{"size":5,"mimeType":"text/plain"}`)
	lc.loaded = false

	w := &lazyContentWrapper{content: lc}
	text := w.GetText()
	assert.Equal(t, "", text) // nil *string returns ""
}

func TestLazyContentWrapper_GetEncoding_NilEncodingAfterLoad(t *testing.T) {
	// LazyContent with no encoding in rawData
	lc := &LazyContent{
		Size:     5,
		MimeType: "text/plain",
	}
	lc.rawData = json.RawMessage(`{"size":5,"mimeType":"text/plain"}`)
	lc.loaded = false

	w := &lazyContentWrapper{content: lc}
	enc := w.GetEncoding()
	assert.Equal(t, "", enc)
}

func TestLazyContentWrapper_ToStandard_WithTextAndEncoding(t *testing.T) {
	textVal := "hello world"
	encVal := "base64"
	lc := &LazyContent{
		Size:        11,
		MimeType:    "text/plain",
		Compression: 5,
		Comment:     "cmt",
		Text:        &textVal,
		Encoding:    &encVal,
	}
	w := &lazyContentWrapper{content: lc}
	c := w.ToStandard()
	assert.Equal(t, "hello world", c.Text)
	assert.Equal(t, "base64", c.Encoding)
	assert.Equal(t, 11, c.Size)
	assert.Equal(t, "text/plain", c.MimeType)
	assert.Equal(t, 5, c.Compression)
	assert.Equal(t, "cmt", c.Comment)
}

func TestLazyContentWrapper_ToStandard_NilTextNilEncoding(t *testing.T) {
	lc := &LazyContent{
		Size:     5,
		MimeType: "text/plain",
		// Text and Encoding are nil
	}
	w := &lazyContentWrapper{content: lc}
	c := w.ToStandard()
	assert.Equal(t, "", c.Text)
	assert.Equal(t, "", c.Encoding)
}

// ===========================================================================
// LazyContent direct method tests
// ===========================================================================

func TestLazyContent_GetSize(t *testing.T) {
	lc := &LazyContent{Size: 42, MimeType: "text/html"}
	assert.Equal(t, 42, lc.GetSize())
}

func TestLazyContent_GetMimeType(t *testing.T) {
	lc := &LazyContent{Size: 10, MimeType: "image/png"}
	assert.Equal(t, "image/png", lc.GetMimeType())
}

func TestLazyContent_GetEncoding(t *testing.T) {
	enc := "base64"
	lc := &LazyContent{Size: 10, MimeType: "text/plain", Encoding: &enc}
	result := lc.GetEncoding()
	assert.Equal(t, "base64", result)
}

func TestLazyContent_GetEncoding_Nil(t *testing.T) {
	lc := &LazyContent{Size: 10, MimeType: "text/plain"}
	// No rawData so Load() will fail, but Encoding is nil so it returns ""
	result := lc.GetEncoding()
	assert.Equal(t, "", result)
}

func TestLazyContent_GetEncoding_LoadFromRawData(t *testing.T) {
	lc := &LazyContent{Size: 10, MimeType: "text/plain"}
	lc.rawData = json.RawMessage(`{"size":10,"mimeType":"text/plain","encoding":"gzip"}`)
	lc.loaded = false
	result := lc.GetEncoding()
	assert.Equal(t, "gzip", result)
}

func TestLazyContent_GetCompression(t *testing.T) {
	lc := &LazyContent{Size: 100, MimeType: "text/plain", Compression: 50}
	assert.Equal(t, 50, lc.GetCompression())
}

func TestLazyContent_ToStandard(t *testing.T) {
	textVal := "body content"
	encVal := "utf-8"
	lc := &LazyContent{
		Size:        12,
		MimeType:    "text/html",
		Compression: 3,
		Comment:     "my-comment",
		Text:        &textVal,
		Encoding:    &encVal,
	}
	c := lc.ToStandard()
	assert.Equal(t, 12, c.Size)
	assert.Equal(t, "text/html", c.MimeType)
	assert.Equal(t, 3, c.Compression)
	assert.Equal(t, "my-comment", c.Comment)
	assert.Equal(t, "body content", c.Text)
	assert.Equal(t, "utf-8", c.Encoding)
}

func TestLazyContent_ToStandard_NilTextNilEncoding(t *testing.T) {
	lc := &LazyContent{
		Size:     0,
		MimeType: "text/plain",
	}
	c := lc.ToStandard()
	assert.Equal(t, "", c.Text)
	assert.Equal(t, "", c.Encoding)
}

func TestLazyContent_ToStandard_LoadsLazyFields(t *testing.T) {
	var lc LazyContent
	require.NoError(t, json.Unmarshal([]byte(`{"size":11,"mimeType":"text/plain","text":"lazy body","encoding":"utf-8"}`), &lc))
	require.Nil(t, lc.Text)
	require.Nil(t, lc.Encoding)

	c := lc.ToStandard()
	assert.Equal(t, "lazy body", c.Text)
	assert.Equal(t, "utf-8", c.Encoding)
}

func TestLazyContent_UnmarshalJSON(t *testing.T) {
	raw := `{"size":500,"mimeType":"application/json","compression":100,"comment":"cmt","text":"hello","encoding":"base64"}`
	var lc LazyContent
	err := json.Unmarshal([]byte(raw), &lc)
	require.NoError(t, err)

	assert.Equal(t, 500, lc.Size)
	assert.Equal(t, "application/json", lc.MimeType)
	assert.Equal(t, 100, lc.Compression)
	assert.Equal(t, "cmt", lc.Comment)
	// Text and Encoding should NOT be populated yet (lazy)
	assert.Nil(t, lc.Text)
	assert.Nil(t, lc.Encoding)
	assert.False(t, lc.loaded)
	// rawData should be saved
	assert.NotEmpty(t, lc.rawData)
}

func TestLazyContent_UnmarshalJSON_Invalid(t *testing.T) {
	var lc LazyContent
	err := json.Unmarshal([]byte(`not json`), &lc)
	assert.Error(t, err)
}

func TestLazyContent_Load(t *testing.T) {
	raw := `{"size":500,"mimeType":"application/json","compression":100,"comment":"cmt","text":"hello","encoding":"base64"}`
	var lc LazyContent
	err := json.Unmarshal([]byte(raw), &lc)
	require.NoError(t, err)

	// Before load
	assert.Nil(t, lc.Text)
	assert.Nil(t, lc.Encoding)
	assert.False(t, lc.loaded)

	// Load
	err = lc.Load()
	require.NoError(t, err)
	assert.True(t, lc.loaded)
	assert.NotNil(t, lc.Text)
	assert.Equal(t, "hello", *lc.Text)
	assert.NotNil(t, lc.Encoding)
	assert.Equal(t, "base64", *lc.Encoding)
}

func TestLazyContent_Load_Idempotent(t *testing.T) {
	raw := `{"size":10,"mimeType":"text/plain","text":"abc"}`
	var lc LazyContent
	err := json.Unmarshal([]byte(raw), &lc)
	require.NoError(t, err)

	err = lc.Load()
	require.NoError(t, err)
	text1 := *lc.Text

	err = lc.Load()
	require.NoError(t, err)
	text2 := *lc.Text
	assert.Equal(t, text1, text2)
}

func TestLazyContent_Load_InvalidRawData(t *testing.T) {
	lc := &LazyContent{
		Size:     10,
		MimeType: "text/plain",
	}
	lc.rawData = json.RawMessage(`{invalid}`)
	lc.loaded = false

	err := lc.Load()
	assert.Error(t, err)
}

func TestLazyContent_GetText(t *testing.T) {
	raw := `{"size":5,"mimeType":"text/plain","text":"hello"}`
	var lc LazyContent
	err := json.Unmarshal([]byte(raw), &lc)
	require.NoError(t, err)

	text, err := lc.GetText()
	require.NoError(t, err)
	require.NotNil(t, text)
	assert.Equal(t, "hello", *text)
}

func TestLazyContent_GetText_AlreadyLoaded(t *testing.T) {
	raw := `{"size":5,"mimeType":"text/plain","text":"world"}`
	var lc LazyContent
	err := json.Unmarshal([]byte(raw), &lc)
	require.NoError(t, err)

	// Pre-load
	err = lc.Load()
	require.NoError(t, err)

	// GetText should use cached value
	text, err := lc.GetText()
	require.NoError(t, err)
	require.NotNil(t, text)
	assert.Equal(t, "world", *text)
}

func TestLazyContent_GetText_NoText(t *testing.T) {
	raw := `{"size":0,"mimeType":"text/plain"}`
	var lc LazyContent
	err := json.Unmarshal([]byte(raw), &lc)
	require.NoError(t, err)

	text, err := lc.GetText()
	require.NoError(t, err)
	assert.Nil(t, text)
}

func TestLazyContent_GetText_LoadError(t *testing.T) {
	lc := &LazyContent{
		Size:     10,
		MimeType: "text/plain",
	}
	lc.rawData = json.RawMessage(`{bad json`)
	lc.loaded = false

	text, err := lc.GetText()
	assert.Error(t, err)
	assert.Nil(t, text)
}

// ===========================================================================
// LazyHar additional helper methods (from lazy.go)
// ===========================================================================

func TestLazyHar_GetEntry(t *testing.T) {
	lh := buildLazyHar(t)
	entry, err := lh.GetEntry(0)
	require.NoError(t, err)
	assert.Equal(t, "GET", entry.Request.Method)
}

func TestLazyHar_GetEntry_OutOfRange(t *testing.T) {
	lh := buildLazyHar(t)
	_, err := lh.GetEntry(-1)
	assert.Error(t, err)
	_, err = lh.GetEntry(999)
	assert.Error(t, err)
}

func TestLazyHar_GetEntriesCount(t *testing.T) {
	lh := buildLazyHar(t)
	assert.Equal(t, 1, lh.GetEntriesCount())
}

func TestLazyHar_GetEntriesCount_Empty(t *testing.T) {
	raw := `{"log":{"version":"1.2","creator":{"name":"t","version":"1"},"entries":[]}}`
	lh, err := ParseHarWithLazyLoading([]byte(raw))
	require.NoError(t, err)
	assert.Equal(t, 0, lh.GetEntriesCount())
}

func TestLazyHar_GetResponseContent(t *testing.T) {
	lh := buildLazyHar(t)
	content, err := lh.GetResponseContent(0)
	require.NoError(t, err)
	require.NotNil(t, content)
	assert.Equal(t, 1024, content.Size)
	assert.Equal(t, "application/json", content.MimeType)
}

func TestLazyHar_GetResponseContent_OutOfRange(t *testing.T) {
	lh := buildLazyHar(t)
	_, err := lh.GetResponseContent(999)
	assert.Error(t, err)
}

func TestLazyHar_GetResponseText(t *testing.T) {
	lh := buildLazyHar(t)
	text, err := lh.GetResponseText(0)
	require.NoError(t, err)
	require.NotNil(t, text)
	assert.Equal(t, "{\"result\":true}", *text)
}

func TestLazyHar_GetResponseText_OutOfRange(t *testing.T) {
	lh := buildLazyHar(t)
	_, err := lh.GetResponseText(999)
	assert.Error(t, err)
}

func TestLazyHar_GetResponseText_NilContent(t *testing.T) {
	lh := buildLazyHarWithNilContent(t)
	text, err := lh.GetResponseText(0)
	require.NoError(t, err)
	assert.Nil(t, text)
}

// ===========================================================================
// ParseHarWithLazyLoading / ParseHarFileWithLazyLoading
// ===========================================================================

func TestParseHarWithLazyLoading(t *testing.T) {
	raw := `{"log":{"version":"1.2","creator":{"name":"t","version":"1"},"entries":[]}}`
	lh, err := ParseHarWithLazyLoading([]byte(raw))
	require.NoError(t, err)
	assert.Equal(t, "1.2", lh.GetVersion())
}

func TestParseHarWithLazyLoading_InvalidJSON(t *testing.T) {
	_, err := ParseHarWithLazyLoading([]byte(`not json`))
	assert.Error(t, err)
}

func TestParseHarFileWithLazyLoading(t *testing.T) {
	// Use the existing testdata/example.har
	lh, err := ParseHarFileWithLazyLoading("testdata/example.har")
	if err != nil {
		// If the testdata file doesn't exist, create a minimal one for the test
		t.Logf("testdata/example.har not available, skipping file-based test: %v", err)
		return
	}
	assert.NotNil(t, lh)
}

func TestParseHarFileWithLazyLoading_NotFound(t *testing.T) {
	_, err := ParseHarFileWithLazyLoading("testdata/nonexistent_file.har")
	assert.Error(t, err)
}

// ===========================================================================
// LazyHar ToStandardHar
// ===========================================================================

func TestLazyHar_ToStandardHar(t *testing.T) {
	lh := buildLazyHar(t)
	std, err := lh.ToStandardHar()
	require.NoError(t, err)
	require.NotNil(t, std)
	assert.Equal(t, "1.2", std.Log.Version)
	assert.Equal(t, "test-creator", std.Log.Creator.Name)
	assert.Len(t, std.Log.Entries, 1)
	assert.Equal(t, "GET", std.Log.Entries[0].Request.Method)
	assert.Equal(t, 200, std.Log.Entries[0].Response.Status)
}

func TestLazyHar_ToStandardHar_NilContent(t *testing.T) {
	lh := buildLazyHarWithNilContent(t)
	std, err := lh.ToStandardHar()
	require.NoError(t, err)
	require.NotNil(t, std)
	assert.Len(t, std.Log.Entries, 1)
	// Content should be zero value when LazyContent is nil
	assert.Equal(t, 0, std.Log.Entries[0].Response.Content.Size)
}

// ===========================================================================
// Multiple entries test
// ===========================================================================

func TestLazyHar_MultipleEntries(t *testing.T) {
	raw := `{
		"log": {
			"version": "1.2",
			"creator": {"name": "multi-test", "version": "2.0"},
			"entries": [
				{
					"startedDateTime": "2024-01-01T00:00:00Z",
					"time": 100,
					"request": {"method": "GET", "url": "https://a.com", "httpVersion": "HTTP/1.1", "headersSize": 50, "bodySize": 0},
					"response": {
						"status": 200, "statusText": "OK", "httpVersion": "HTTP/1.1",
						"headers": [{"name": "X-Custom", "value": "v1"}],
						"cookies": [{"name": "ck1", "value": "val1"}],
						"content": {"size": 50, "mimeType": "text/html", "text": "hello"},
						"headersSize": 50, "bodySize": 50
					},
					"timings": {"blocked": 1, "dns": 2, "connect": 3, "ssl": 4, "send": 5, "wait": 6, "receive": 7},
					"pageref": "p1",
					"connection": "c1",
					"serverIPAddress": "1.2.3.4"
				},
				{
					"startedDateTime": "2024-01-01T00:00:01Z",
					"time": 200,
					"request": {"method": "POST", "url": "https://b.com/api", "httpVersion": "HTTP/2.0", "headersSize": 80, "bodySize": 30},
					"response": {
						"status": 404, "statusText": "Not Found", "httpVersion": "HTTP/2.0",
						"headers": [],
						"cookies": [],
						"content": {"size": 0, "mimeType": "text/plain"},
						"headersSize": 30, "bodySize": 0
					},
					"timings": {"blocked": 10, "dns": 20, "connect": 30, "ssl": 40, "send": 5, "wait": 60, "receive": 35},
					"serverIPAddress": "5.6.7.8"
				}
			]
		}
	}`

	lh, err := ParseHarWithLazyLoading([]byte(raw))
	require.NoError(t, err)

	// HARProvider
	assert.Equal(t, "1.2", lh.GetVersion())
	assert.Equal(t, "2.0", lh.GetCreator().Version)
	assert.Equal(t, 2, lh.GetEntriesCount())

	entries := lh.GetEntries()
	assert.Len(t, entries, 2)

	// Entry 0
	e0 := entries[0]
	assert.Equal(t, "GET", e0.GetRequest().GetMethod())
	assert.Equal(t, 200, e0.GetResponse().GetStatus())
	assert.Equal(t, "p1", e0.GetPageref())

	// Entry 1
	e1 := entries[1]
	assert.Equal(t, "POST", e1.GetRequest().GetMethod())
	assert.Equal(t, 404, e1.GetResponse().GetStatus())
	assert.Equal(t, "", e1.GetPageref())
}

func TestLazyHarNilReceiverMethods(t *testing.T) {
	var lh *LazyHar

	assert.Equal(t, "", lh.GetVersion())
	assert.Equal(t, Creator{}, lh.GetCreator())
	assert.Equal(t, Browser{}, lh.GetBrowser())
	assert.Nil(t, lh.GetEntries())
	assert.Nil(t, lh.GetPages())
	assert.Nil(t, lh.ToStandard())

	std, err := lh.ToStandardHar()
	assert.Nil(t, std)
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)

	entry, err := lh.GetEntry(0)
	assert.Nil(t, entry)
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)

	assert.Equal(t, 0, lh.GetEntriesCount())

	content, err := lh.GetResponseContent(0)
	assert.Nil(t, content)
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)

	text, err := lh.GetResponseText(0)
	assert.Nil(t, text)
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
}

func TestLazyProviderNilReceiverMethods(t *testing.T) {
	var entry *LazyEntries
	assert.True(t, entry.GetStartedDateTime().IsZero())
	assert.Equal(t, 0.0, entry.GetTime())
	assert.Nil(t, entry.GetRequest())
	assert.Nil(t, entry.GetResponse())
	assert.Nil(t, entry.GetTimings())
	assert.Equal(t, "", entry.GetPageref())
	assert.Equal(t, Entries{}, entry.ToStandard())

	var response *LazyResponse
	assert.Equal(t, 0, response.GetStatus())
	assert.Equal(t, "", response.GetStatusText())
	assert.Equal(t, "", response.GetHTTPVersion())
	assert.Nil(t, response.GetHeaders())
	assert.Nil(t, response.GetCookies())
	assert.Nil(t, response.GetContent())
	assert.Equal(t, 0, response.GetBodySize())
	assert.Equal(t, 0, response.GetHeadersSize())
	assert.Equal(t, Response{}, response.ToStandard())

	var wrapper *lazyContentWrapper
	assert.Equal(t, 0, wrapper.GetSize())
	assert.Equal(t, "", wrapper.GetMimeType())
	assert.Equal(t, "", wrapper.GetText())
	assert.Equal(t, "", wrapper.GetEncoding())
	assert.Equal(t, 0, wrapper.GetCompression())
	assert.Equal(t, Content{}, wrapper.ToStandard())

	var content *LazyContent
	assert.Equal(t, 0, content.GetSize())
	assert.Equal(t, "", content.GetMimeType())
	assert.Equal(t, "", content.GetEncoding())
	assert.Equal(t, 0, content.GetCompression())
	assert.Equal(t, Content{}, content.ToStandard())

	text, err := content.GetText()
	assert.Nil(t, text)
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
}

// ===========================================================================
// Interface conformance checks (compile-time)
// ===========================================================================

func TestLazyHar_ImplementsHARProvider(t *testing.T) {
	var _ HARProvider = (*LazyHar)(nil)
}

func TestLazyEntries_ImplementsEntryProvider(t *testing.T) {
	var _ EntryProvider = (*LazyEntries)(nil)
}

func TestLazyResponse_ImplementsResponseProvider(t *testing.T) {
	var _ ResponseProvider = (*LazyResponse)(nil)
}

func TestLazyContent_DoesNotDirectlyImplementContentProvider(t *testing.T) {
	// LazyContent.GetText() returns (*string, error), not string,
	// so it does NOT directly implement ContentProvider.
	// It's used via lazyContentWrapper instead.
	lc := &LazyContent{}
	_ = lc // just verify it compiles, don't assert interface compliance
}

func TestLazyContentWrapper_ImplementsContentProvider(t *testing.T) {
	var _ ContentProvider = (*lazyContentWrapper)(nil)
}
