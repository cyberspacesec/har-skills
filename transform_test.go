package har

import (
	"strings"
	"testing"
)

func createHarForTransform() *Har {
	h := NewHar()
	h.SetCreator("test", "1.0")

	e1 := h.AddEntry("GET", "http://localhost:8080/api/users", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e1.SetResponseContent(1024, "application/json")
	e1.AddRequestHeader("Accept", "application/json")
	e1.AddRequestHeader("Host", "localhost:8080")
	e1.AddResponseHeader("Content-Type", "application/json")

	e2 := h.AddEntry("POST", "http://localhost:8080/api/data", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")
	e2.SetResponseContent(512, "text/html")
	e2.AddRequestHeader("Content-Type", "application/json")
	e2.AddRequestHeader("Host", "localhost:8080")
	e2.AddCookie("session", "abc123")

	e3 := h.AddEntry("GET", "https://example.com/api/items?page=1", "HTTP/1.1", "")
	e3.SetResponseStatus(200, "OK")
	e3.SetResponseContent(256, "application/json")
	e3.AddRequestHeader("Host", "example.com")

	return h
}

func TestTransformURLRewrite(t *testing.T) {
	h := createHarForTransform()

	result := h.RewriteURL("http://localhost:8080", "https://prod.example.com")
	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Verify original is not modified
	if h.Log.Entries[0].Request.URL != "http://localhost:8080/api/users" {
		t.Error("Original should not be modified")
	}

	// Verify transformed
	if result.Log.Entries[0].Request.URL != "https://prod.example.com/api/users" {
		t.Errorf("Expected URL to be rewritten, got %s", result.Log.Entries[0].Request.URL)
	}
	if result.Log.Entries[1].Request.URL != "https://prod.example.com/api/data" {
		t.Errorf("Expected URL to be rewritten, got %s", result.Log.Entries[1].Request.URL)
	}
	// Third entry should be unchanged
	if result.Log.Entries[2].Request.URL != "https://example.com/api/items?page=1" {
		t.Errorf("Third entry URL should be unchanged, got %s", result.Log.Entries[2].Request.URL)
	}
}

func TestTransformURLRewriteUpdatesHostHeader(t *testing.T) {
	h := createHarForTransform()

	result := h.RewriteURL("http://localhost:8080", "https://prod.example.com")

	// Check Host header updated
	found := false
	for _, hdr := range result.Log.Entries[0].Request.Headers {
		if hdr.Name == "Host" {
			if hdr.Value != "prod.example.com" {
				t.Errorf("Expected Host header to be 'prod.example.com', got '%s'", hdr.Value)
			}
			found = true
		}
	}
	if !found {
		t.Error("Expected Host header to be found")
	}
}

func TestTransformInPlace(t *testing.T) {
	h := createHarForTransform()

	rules := []TransformRule{
		{
			Type:        TransformURLRewrite,
			Pattern:     "http://localhost:8080",
			Replacement: "https://prod.example.com",
		},
	}

	h.TransformInPlace(rules)

	if h.Log.Entries[0].Request.URL != "https://prod.example.com/api/users" {
		t.Errorf("Expected URL to be rewritten in place, got %s", h.Log.Entries[0].Request.URL)
	}
}

func TestTransformHostReplace(t *testing.T) {
	h := createHarForTransform()

	result := h.Transform([]TransformRule{
		{
			Type:        TransformHostReplace,
			Pattern:     "localhost:8080",
			Replacement: "prod.example.com",
		},
	})

	if result.Log.Entries[0].Request.URL != "http://prod.example.com/api/users" {
		t.Errorf("Expected host to be replaced, got %s", result.Log.Entries[0].Request.URL)
	}
}

func TestTransformSchemeChange(t *testing.T) {
	h := NewHar()
	e := h.AddEntry("GET", "http://example.com/api", "HTTP/1.1", "")
	e.AddRequestHeader("Host", "example.com")

	result := h.Transform([]TransformRule{
		{
			Type:        TransformSchemeChange,
			Pattern:     "http",
			Replacement: "https",
		},
	})

	if result.Log.Entries[0].Request.URL != "https://example.com/api" {
		t.Errorf("Expected scheme to be changed, got %s", result.Log.Entries[0].Request.URL)
	}
}

func TestTransformHeaderAdd(t *testing.T) {
	h := createHarForTransform()

	result := h.Transform([]TransformRule{
		{
			Type:        TransformHeaderAdd,
			HeaderName:  "X-Custom",
			HeaderValue: "test-value",
		},
	})

	// Check request header added
	found := false
	for _, hdr := range result.Log.Entries[0].Request.Headers {
		if hdr.Name == "X-Custom" && hdr.Value == "test-value" {
			found = true
		}
	}
	if !found {
		t.Error("Expected X-Custom header to be added to request")
	}

	// Check response header added
	found = false
	for _, hdr := range result.Log.Entries[0].Response.Headers {
		if hdr.Name == "X-Custom" && hdr.Value == "test-value" {
			found = true
		}
	}
	if !found {
		t.Error("Expected X-Custom header to be added to response")
	}
}

func TestTransformHeaderRemove(t *testing.T) {
	h := createHarForTransform()

	result := h.RemoveHeaders([]string{"Host"})

	for _, entry := range result.Log.Entries {
		for _, hdr := range entry.Request.Headers {
			if hdr.Name == "Host" {
				t.Error("Host header should have been removed")
			}
		}
		for _, hdr := range entry.Response.Headers {
			if hdr.Name == "Host" {
				t.Error("Host header should have been removed from response")
			}
		}
	}
}

func TestTransformHeaderReplace(t *testing.T) {
	h := createHarForTransform()

	result := h.Transform([]TransformRule{
		{
			Type:        TransformHeaderReplace,
			HeaderName:  "Accept",
			HeaderValue: "text/html",
		},
	})

	for _, hdr := range result.Log.Entries[0].Request.Headers {
		if hdr.Name == "Accept" {
			if hdr.Value != "text/html" {
				t.Errorf("Expected Accept header value to be 'text/html', got '%s'", hdr.Value)
			}
		}
	}
}

func TestTransformQueryParamRemove(t *testing.T) {
	h := NewHar()
	_ = h.AddEntry("GET", "https://example.com/api?page=1&cb=123&sort=name", "HTTP/1.1", "")

	result := h.Transform([]TransformRule{
		{
			Type:    TransformQueryParamRemove,
			Pattern: "cb",
		},
	})

	if len(result.Log.Entries[0].Request.QueryString) != 2 {
		t.Errorf("Expected 2 query params after removal, got %d", len(result.Log.Entries[0].Request.QueryString))
	}

	for _, qs := range result.Log.Entries[0].Request.QueryString {
		if qs.Name == "cb" {
			t.Error("cb query param should have been removed")
		}
	}
}

func TestTransformQueryParamAdd(t *testing.T) {
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/api?page=1", "HTTP/1.1", "")
	_ = e

	result := h.Transform([]TransformRule{
		{
			Type:        TransformQueryParamAdd,
			HeaderName:  "sort",
			HeaderValue: "name",
		},
	})

	if len(result.Log.Entries[0].Request.QueryString) != 2 {
		t.Errorf("Expected 2 query params after add, got %d", len(result.Log.Entries[0].Request.QueryString))
	}

	found := false
	for _, qs := range result.Log.Entries[0].Request.QueryString {
		if qs.Name == "sort" && qs.Value == "name" {
			found = true
		}
	}
	if !found {
		t.Error("Expected sort=name query param to be added")
	}
}

func TestTransformCookieDomainRewrite(t *testing.T) {
	h := NewHar()
	e := h.AddEntry("GET", "https://old.example.com/api", "HTTP/1.1", "")
	e.AddCookie("session", "abc123")
	e.Request.Cookies[0].Domain = "old.example.com"
	e.AddResponseCookie("tracking", "xyz")
	e.Response.Cookies[0].Domain = "old.example.com"

	result := h.Transform([]TransformRule{
		{
			Type:        TransformCookieDomainRewrite,
			Pattern:     "old.example.com",
			Replacement: "new.example.com",
		},
	})

	if result.Log.Entries[0].Request.Cookies[0].Domain != "new.example.com" {
		t.Errorf("Expected request cookie domain to be rewritten, got %s", result.Log.Entries[0].Request.Cookies[0].Domain)
	}
	if result.Log.Entries[0].Response.Cookies[0].Domain != "new.example.com" {
		t.Errorf("Expected response cookie domain to be rewritten, got %s", result.Log.Entries[0].Response.Cookies[0].Domain)
	}
}

func TestTransformBodyReplace(t *testing.T) {
	h := NewHar()
	e := h.AddEntry("POST", "https://example.com/api", "HTTP/1.1", "")
	e.SetPostData("application/json", `{"name":"old","value":"test"}`)

	result := h.Transform([]TransformRule{
		{
			Type:        TransformBodyReplace,
			Pattern:     "old",
			Replacement: "new",
		},
	})

	if !strings.Contains(result.Log.Entries[0].Request.PostData.Text, "new") {
		t.Error("Expected body to contain 'new'")
	}
	if strings.Contains(result.Log.Entries[0].Request.PostData.Text, "old") {
		t.Error("Expected body not to contain 'old'")
	}
}

func TestTransformBodyReplaceRegex(t *testing.T) {
	h := NewHar()
	e := h.AddEntry("POST", "https://example.com/api", "HTTP/1.1", "")
	e.SetPostData("application/json", `{"token":"abc123def","key":"xyz"}`)

	result := h.Transform([]TransformRule{
		{
			Type:        TransformBodyReplace,
			Pattern:     `"token":"[^"]*"`,
			Replacement: `"token":"REDACTED"`,
		},
	})

	if !strings.Contains(result.Log.Entries[0].Request.PostData.Text, `"token":"REDACTED"`) {
		t.Error("Expected token to be redacted")
	}
	if !strings.Contains(result.Log.Entries[0].Request.PostData.Text, `"key":"xyz"`) {
		t.Error("Expected other fields to remain unchanged")
	}
}

func TestTransformNilHar(t *testing.T) {
	var h *Har
	result := h.Transform(nil)
	if result != nil {
		t.Error("Expected nil result for nil Har")
	}
}

func TestAddHeadersToRequest(t *testing.T) {
	h := createHarForTransform()

	headers := map[string]string{
		"X-Request-ID":  "abc-123",
		"Authorization": "Bearer token",
	}
	result := h.AddHeaders(headers, "request")

	for _, entry := range result.Log.Entries {
		for _, hdr := range entry.Request.Headers {
			if hdr.Name == "X-Request-ID" {
				if hdr.Value != "abc-123" {
					t.Errorf("Expected X-Request-ID header value 'abc-123', got '%s'", hdr.Value)
				}
			}
		}
		// Response should not have the header
		for _, hdr := range entry.Response.Headers {
			if hdr.Name == "X-Request-ID" {
				t.Error("X-Request-ID should not be added to response when target is 'request'")
			}
		}
	}
}

func TestAddHeadersToResponse(t *testing.T) {
	h := createHarForTransform()

	headers := map[string]string{
		"X-Custom": "test",
	}
	result := h.AddHeaders(headers, "response")

	for _, entry := range result.Log.Entries {
		// Request should not have the header
		for _, hdr := range entry.Request.Headers {
			if hdr.Name == "X-Custom" {
				t.Error("X-Custom should not be added to request when target is 'response'")
			}
		}
		// Response should have the header
		found := false
		for _, hdr := range entry.Response.Headers {
			if hdr.Name == "X-Custom" && hdr.Value == "test" {
				found = true
			}
		}
		if !found {
			t.Error("X-Custom header should be added to response")
		}
	}
}

func TestAddHeadersToBoth(t *testing.T) {
	h := createHarForTransform()

	headers := map[string]string{
		"X-Both": "test",
	}
	result := h.AddHeaders(headers, "both")

	for _, entry := range result.Log.Entries {
		reqFound := false
		respFound := false
		for _, hdr := range entry.Request.Headers {
			if hdr.Name == "X-Both" {
				reqFound = true
			}
		}
		for _, hdr := range entry.Response.Headers {
			if hdr.Name == "X-Both" {
				respFound = true
			}
		}
		if !reqFound || !respFound {
			t.Error("X-Both header should be added to both request and response")
		}
	}
}

func TestTransformDoesNotModifyOriginal(t *testing.T) {
	h := createHarForTransform()
	originalURL := h.Log.Entries[0].Request.URL

	_ = h.RewriteURL("http://localhost:8080", "https://prod.example.com")

	if h.Log.Entries[0].Request.URL != originalURL {
		t.Errorf("Original HAR should not be modified, got %s", h.Log.Entries[0].Request.URL)
	}
}

func TestTransformMultipleRules(t *testing.T) {
	h := NewHar()
	e := h.AddEntry("GET", "http://example.com/api?cb=123", "HTTP/1.1", "")
	e.AddRequestHeader("Host", "example.com")
	e.AddRequestHeader("X-Debug", "true")

	rules := []TransformRule{
		{Type: TransformSchemeChange, Pattern: "http", Replacement: "https"},
		{Type: TransformHeaderRemove, HeaderName: "X-Debug"},
		{Type: TransformQueryParamRemove, Pattern: "cb"},
	}

	result := h.Transform(rules)

	if !strings.HasPrefix(result.Log.Entries[0].Request.URL, "https://") {
		t.Errorf("Expected https scheme, got %s", result.Log.Entries[0].Request.URL)
	}

	for _, hdr := range result.Log.Entries[0].Request.Headers {
		if hdr.Name == "X-Debug" {
			t.Error("X-Debug header should have been removed")
		}
	}

	for _, qs := range result.Log.Entries[0].Request.QueryString {
		if qs.Name == "cb" {
			t.Error("cb query param should have been removed")
		}
	}
}

// --- Additional tests for uncovered branches ---

func TestTransformInPlaceNilHar(t *testing.T) {
	// Covers TransformInPlace branch: nil Har
	var h *Har
	h.TransformInPlace([]TransformRule{{Type: TransformSchemeChange, Pattern: "http", Replacement: "https"}})
	// Should not panic
}

func TestTransformInPlaceEmptyRules(t *testing.T) {
	// Covers TransformInPlace branch: empty rules
	h := createHarForTransform()
	originalURL := h.Log.Entries[0].Request.URL
	h.TransformInPlace([]TransformRule{})
	if h.Log.Entries[0].Request.URL != originalURL {
		t.Errorf("empty rules should not modify Har, got: %s", h.Log.Entries[0].Request.URL)
	}
}

func TestTransformBodyReplaceNilPostData(t *testing.T) {
	// Covers applyBodyReplace branch: nil PostData
	h := NewHar()
	e := h.AddEntry("POST", "https://example.com/api", "HTTP/1.1", "")
	e.Request.PostData = nil // explicitly nil

	result := h.Transform([]TransformRule{
		{Type: TransformBodyReplace, Pattern: "old", Replacement: "new"},
	})

	if result.Log.Entries[0].Request.PostData != nil {
		t.Error("nil PostData should remain nil after body replace")
	}
}

func TestTransformBodyReplaceEmptyPattern(t *testing.T) {
	// Covers applyBodyReplace branch: empty pattern
	h := NewHar()
	e := h.AddEntry("POST", "https://example.com/api", "HTTP/1.1", "")
	e.SetPostData("application/json", `{"key":"value"}`)

	result := h.Transform([]TransformRule{
		{Type: TransformBodyReplace, Pattern: "", Replacement: "new"},
	})

	// Empty pattern should return early, body unchanged
	if result.Log.Entries[0].Request.PostData.Text != `{"key":"value"}` {
		t.Errorf("empty pattern should not modify body, got: %s", result.Log.Entries[0].Request.PostData.Text)
	}
}

func TestTransformBodyReplaceInvalidRegex(t *testing.T) {
	// Covers applyBodyReplace branch: invalid regex pattern falls back to plain string replace
	h := NewHar()
	e := h.AddEntry("POST", "https://example.com/api", "HTTP/1.1", "")
	e.SetPostData("text/plain", "replace-me-with-plain")

	result := h.Transform([]TransformRule{
		{Type: TransformBodyReplace, Pattern: "[invalid", Replacement: "good"},
	})

	// Invalid regex should fall back to strings.ReplaceAll
	if result.Log.Entries[0].Request.PostData.Text != "replace-me-with-plain" {
		// The pattern "[invalid" is not in the text, so no replacement happens
		t.Errorf("invalid regex fallback should still try plain replace, got: %s", result.Log.Entries[0].Request.PostData.Text)
	}
}

func TestTransformBodyReplaceInvalidRegexWithMatch(t *testing.T) {
	// Covers applyBodyReplace branch: invalid regex falls back to plain string replace with actual match
	h := NewHar()
	e := h.AddEntry("POST", "https://example.com/api", "HTTP/1.1", "")
	e.SetPostData("text/plain", "hello [invalid world")

	result := h.Transform([]TransformRule{
		{Type: TransformBodyReplace, Pattern: "[invalid", Replacement: "valid"},
	})

	// Invalid regex "[invalid" is treated as literal string
	if result.Log.Entries[0].Request.PostData.Text != "hello valid world" {
		t.Errorf("invalid regex should fall back to plain string replace, got: %s", result.Log.Entries[0].Request.PostData.Text)
	}
}

func TestUpdateHostHeaderNoExistingHost(t *testing.T) {
	// Covers updateHostHeader branch: no Host header present, adds one
	h := NewHar()
	e := h.AddEntry("GET", "http://localhost:8080/api", "HTTP/1.1", "")
	// Remove any auto-added Host header
	e.Request.Headers = nil

	result := h.RewriteURL("http://localhost:8080", "https://newhost.example.com")

	// Host header should be added
	found := false
	for _, hdr := range result.Log.Entries[0].Request.Headers {
		if hdr.Name == "Host" {
			if hdr.Value != "newhost.example.com" {
				t.Errorf("Expected Host header 'newhost.example.com', got '%s'", hdr.Value)
			}
			found = true
		}
	}
	if !found {
		t.Error("Host header should have been added when missing")
	}
}

func TestRebuildURLFromQueryStringParseError(t *testing.T) {
	// Covers rebuildURLFromQueryString branch: url.Parse error
	// We need to trigger rebuildURLFromQueryString with an unparseable URL.
	// applyQueryParamRemove calls rebuildURLFromQueryString.
	h := NewHar()
	e := h.AddEntry("GET", "://bad-url?cb=123", "HTTP/1.1", "")
	// Manually set QueryString to trigger the rebuild path
	e.Request.QueryString = []QueryString{
		{Name: "cb", Value: "123"},
	}

	// This should not panic even though the URL is unparseable
	result := h.Transform([]TransformRule{
		{Type: TransformQueryParamRemove, Pattern: "cb"},
	})

	// URL should remain unchanged since url.Parse fails
	if result.Log.Entries[0].Request.URL != "://bad-url?cb=123" {
		t.Errorf("unparseable URL should remain unchanged on rebuild failure, got: %s", result.Log.Entries[0].Request.URL)
	}
}

func TestTransformQueryParamAddEmptyQueryString(t *testing.T) {
	// Covers applyQueryParamAdd branch: empty QueryString parsed from URL
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/api?existing=1", "HTTP/1.1", "")
	// Clear QueryString so it gets re-parsed from URL
	e.Request.QueryString = nil

	result := h.Transform([]TransformRule{
		{Type: TransformQueryParamAdd, HeaderName: "newparam", HeaderValue: "newval"},
	})

	found := false
	for _, qs := range result.Log.Entries[0].Request.QueryString {
		if qs.Name == "newparam" && qs.Value == "newval" {
			found = true
		}
	}
	if !found {
		t.Error("newparam should have been added")
	}

	// existing param should also be present
	foundExisting := false
	for _, qs := range result.Log.Entries[0].Request.QueryString {
		if qs.Name == "existing" && qs.Value == "1" {
			foundExisting = true
		}
	}
	if !foundExisting {
		t.Error("existing param should still be present")
	}
}

func TestTransformQueryParamRemoveEmptyQueryString(t *testing.T) {
	// Covers applyQueryParamRemove branch: empty QueryString parsed from URL
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/api?cb=123&page=1", "HTTP/1.1", "")
	// Clear QueryString so it gets re-parsed from URL
	e.Request.QueryString = nil

	result := h.Transform([]TransformRule{
		{Type: TransformQueryParamRemove, Pattern: "cb"},
	})

	for _, qs := range result.Log.Entries[0].Request.QueryString {
		if qs.Name == "cb" {
			t.Error("cb query param should have been removed")
		}
	}

	foundPage := false
	for _, qs := range result.Log.Entries[0].Request.QueryString {
		if qs.Name == "page" && qs.Value == "1" {
			foundPage = true
		}
	}
	if !foundPage {
		t.Error("page query param should still be present")
	}
}

func TestTransformCloneFailureReturnsNil(t *testing.T) {
	h := NewHar()
	h.Log.CustomFields = CustomFields{
		"_bad": func() {},
	}

	result := h.Transform([]TransformRule{
		{Type: TransformHeaderAdd, HeaderName: "X-Test", HeaderValue: "1"},
	})
	if result != nil {
		t.Errorf("Transform() should return nil when Clone() fails, got %#v", result)
	}

	if result := h.AddHeaders(map[string]string{"X-Test": "1"}, "request"); result != nil {
		t.Errorf("AddHeaders() should return nil when Clone() fails, got %#v", result)
	}
}

func TestAddHeadersNilHar(t *testing.T) {
	var h *Har
	if result := h.AddHeaders(map[string]string{"X-Test": "1"}, "request"); result != nil {
		t.Errorf("nil HAR AddHeaders() should return nil, got %#v", result)
	}
}

func TestTransformNilEntryHelpersDoNotPanic(t *testing.T) {
	rule := TransformRule{
		Type:        TransformURLRewrite,
		Pattern:     "http://example.com",
		Replacement: "https://example.com",
		HeaderName:  "X-Test",
		HeaderValue: "value",
	}

	tests := []struct {
		name string
		fn   func()
	}{
		{"applyRules", func() { applyRules(nil, []TransformRule{rule}) }},
		{"applyURLRewrite", func() { applyURLRewrite(nil, rule) }},
		{"applyHostReplace", func() { applyHostReplace(nil, rule) }},
		{"applySchemeChange", func() { applySchemeChange(nil, rule) }},
		{"applyHeaderAdd", func() { applyHeaderAdd(nil, rule) }},
		{"applyHeaderRemove", func() { applyHeaderRemove(nil, rule) }},
		{"applyHeaderReplace", func() { applyHeaderReplace(nil, rule) }},
		{"applyQueryParamRemove", func() { applyQueryParamRemove(nil, rule) }},
		{"applyQueryParamAdd", func() { applyQueryParamAdd(nil, rule) }},
		{"applyCookieDomainRewrite", func() { applyCookieDomainRewrite(nil, rule) }},
		{"applyBodyReplace", func() { applyBodyReplace(nil, rule) }},
		{"updateHostHeader", func() { updateHostHeader(nil, "example.com") }},
		{"rebuildURLFromQueryString", func() { rebuildURLFromQueryString(nil) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertDoesNotPanic(t, tt.fn)
		})
	}
}
