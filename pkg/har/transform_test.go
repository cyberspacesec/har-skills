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
		"X-Request-ID": "abc-123",
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
