package har

import (
	"strings"
	"testing"
)

func TestRedactDefaultRemovesAuthHeaders(t *testing.T) {
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						Method: "GET",
						URL:    "https://example.com/api",
						Headers: []Headers{
							{Name: "Authorization", Value: "Bearer secret-token"},
							{Name: "X-Api-Key", Value: "my-api-key"},
							{Name: "Content-Type", Value: "application/json"},
						},
					},
					Response: Response{
						Status: 200,
						Headers: []Headers{
							{Name: "Set-Cookie", Value: "session=abc123"},
							{Name: "Content-Type", Value: "application/json"},
						},
					},
				},
			},
		},
	}

	opts := DefaultRedactOptions()
	result := h.Redact(opts)

	// Original should be unchanged
	if h.Log.Entries[0].Request.Headers[0].Value != "Bearer secret-token" {
		t.Error("original Har was modified by Redact()")
	}

	// Redacted values
	reqHeaders := result.Log.Entries[0].Request.Headers
	if reqHeaders[0].Value != "[REDACTED]" {
		t.Errorf("Authorization header not redacted, got: %s", reqHeaders[0].Value)
	}
	if reqHeaders[1].Value != "[REDACTED]" {
		t.Errorf("X-Api-Key header not redacted, got: %s", reqHeaders[1].Value)
	}
	if reqHeaders[2].Value != "application/json" {
		t.Errorf("Content-Type header should not be redacted, got: %s", reqHeaders[2].Value)
	}

	respHeaders := result.Log.Entries[0].Response.Headers
	if respHeaders[0].Value != "[REDACTED]" {
		t.Errorf("Set-Cookie header not redacted, got: %s", respHeaders[0].Value)
	}
	if respHeaders[1].Value != "application/json" {
		t.Errorf("Content-Type response header should not be redacted, got: %s", respHeaders[1].Value)
	}
}

func TestRedactCustomReplacement(t *testing.T) {
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						URL: "https://example.com",
						Headers: []Headers{
							{Name: "Authorization", Value: "Bearer token123"},
						},
					},
					Response: Response{Status: 200},
				},
			},
		},
	}

	opts := RedactOptions{
		Headers:     []string{"Authorization"},
		Replacement: "***",
	}
	result := h.Redact(opts)

	if result.Log.Entries[0].Request.Headers[0].Value != "***" {
		t.Errorf("expected ***, got: %s", result.Log.Entries[0].Request.Headers[0].Value)
	}
}

func TestRedactCookieRedaction(t *testing.T) {
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						URL: "https://example.com",
						Cookies: []Cookie{
							{Name: "session", Value: "abc123"},
							{Name: "preferences", Value: "dark-mode"},
						},
					},
					Response: Response{
						Status: 200,
						Cookies: []Cookie{
							{Name: "access_token", Value: "tok-xyz"},
							{Name: "theme", Value: "light"},
						},
					},
				},
			},
		},
	}

	opts := RedactOptions{
		Cookies:     []string{"session", "access_token"},
		Replacement: "[REDACTED]",
	}
	result := h.Redact(opts)

	reqCookies := result.Log.Entries[0].Request.Cookies
	if reqCookies[0].Value != "[REDACTED]" {
		t.Errorf("session cookie not redacted, got: %s", reqCookies[0].Value)
	}
	if reqCookies[1].Value != "dark-mode" {
		t.Errorf("preferences cookie should not be redacted, got: %s", reqCookies[1].Value)
	}

	respCookies := result.Log.Entries[0].Response.Cookies
	if respCookies[0].Value != "[REDACTED]" {
		t.Errorf("access_token cookie not redacted, got: %s", respCookies[0].Value)
	}
	if respCookies[1].Value != "light" {
		t.Errorf("theme cookie should not be redacted, got: %s", respCookies[1].Value)
	}
}

func TestRedactQueryParamRedaction(t *testing.T) {
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						URL: "https://example.com/api?token=secret123&page=1",
						QueryString: []QueryString{
							{Name: "token", Value: "secret123"},
							{Name: "page", Value: "1"},
						},
					},
					Response: Response{Status: 200},
				},
			},
		},
	}

	opts := RedactOptions{
		QueryParams: []string{"token"},
		Replacement: "[REDACTED]",
	}
	result := h.Redact(opts)

	qs := result.Log.Entries[0].Request.QueryString
	if qs[0].Value != "[REDACTED]" {
		t.Errorf("token query param not redacted, got: %s", qs[0].Value)
	}
	if qs[1].Value != "1" {
		t.Errorf("page query param should not be redacted, got: %s", qs[1].Value)
	}

	// URL should also have the token redacted
	resultURL := result.Log.Entries[0].Request.URL
	if !strings.Contains(resultURL, "token=[REDACTED]") {
		t.Errorf("URL query param not redacted, got: %s", resultURL)
	}
	if !strings.Contains(resultURL, "page=1") {
		t.Errorf("URL non-sensitive query param changed, got: %s", resultURL)
	}
}

func TestRedactPostDataRedaction(t *testing.T) {
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						URL:    "https://example.com/login",
						Method: "POST",
						PostData: &PostData{
							MimeType: "application/x-www-form-urlencoded",
							Params: []Param{
								{Name: "password", Value: "hunter2"},
								{Name: "username", Value: "alice"},
							},
							Text: "password=hunter2&username=alice",
						},
					},
					Response: Response{Status: 200},
				},
			},
		},
	}

	opts := RedactOptions{
		PostDataFields: []string{"password"},
		Replacement:    "[REDACTED]",
	}
	result := h.Redact(opts)

	// Params
	params := result.Log.Entries[0].Request.PostData.Params
	if params[0].Value != "[REDACTED]" {
		t.Errorf("password POST param not redacted, got: %s", params[0].Value)
	}
	if params[1].Value != "alice" {
		t.Errorf("username POST param should not be redacted, got: %s", params[1].Value)
	}

	// Text body
	text := result.Log.Entries[0].Request.PostData.Text
	if !strings.Contains(text, "password=[REDACTED]") {
		t.Errorf("POST body password not redacted, got: %s", text)
	}
	if !strings.Contains(text, "username=alice") {
		t.Errorf("POST body username changed, got: %s", text)
	}
}

func TestRedactPostDataJSONBody(t *testing.T) {
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						URL:    "https://example.com/api",
						Method: "POST",
						PostData: &PostData{
							MimeType: "application/json",
							Text:     `{"password": "hunter2", "username": "alice", "secret": "s3cret"}`,
						},
					},
					Response: Response{Status: 200},
				},
			},
		},
	}

	opts := RedactOptions{
		PostDataFields: []string{"password", "secret"},
		Replacement:    "[REDACTED]",
	}
	result := h.Redact(opts)

	text := result.Log.Entries[0].Request.PostData.Text
	if !strings.Contains(text, `"password": "[REDACTED]"`) {
		t.Errorf("JSON password not redacted, got: %s", text)
	}
	if !strings.Contains(text, `"secret": "[REDACTED]"`) {
		t.Errorf("JSON secret not redacted, got: %s", text)
	}
	if !strings.Contains(text, `"username": "alice"`) {
		t.Errorf("JSON username changed, got: %s", text)
	}
}

func TestRedactIPAnonymization(t *testing.T) {
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						URL: "https://example.com",
					},
					Response:          Response{Status: 200},
					ServerIPAddress: "192.168.1.42",
				},
			},
		},
	}

	opts := RedactOptions{
		RedactIPs:  true,
		Replacement: "[REDACTED]",
	}
	result := h.Redact(opts)

	if result.Log.Entries[0].ServerIPAddress != "192.168.1.0" {
		t.Errorf("IP not anonymized, got: %s", result.Log.Entries[0].ServerIPAddress)
	}

	// Original should be unchanged
	if h.Log.Entries[0].ServerIPAddress != "192.168.1.42" {
		t.Error("original Har was modified")
	}
}

func TestRedactIPAnonymizationIPv6(t *testing.T) {
	result := anonymizeIP("::1")
	if result != "::0" {
		t.Errorf("IPv6 loopback not anonymized, got: %s", result)
	}
}

func TestRedactReturnsClone(t *testing.T) {
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						URL: "https://example.com",
						Headers: []Headers{
							{Name: "Authorization", Value: "Bearer token"},
						},
					},
					Response: Response{Status: 200},
				},
			},
		},
	}

	opts := DefaultRedactOptions()
	result := h.Redact(opts)

	// Verify original is unchanged
	if h.Log.Entries[0].Request.Headers[0].Value != "Bearer token" {
		t.Error("Redact() modified the original Har")
	}

	// Verify clone is redacted
	if result.Log.Entries[0].Request.Headers[0].Value != "[REDACTED]" {
		t.Error("Redact() did not redact the clone")
	}

	// Verify they are different objects
	if &h.Log.Entries[0] == &result.Log.Entries[0] {
		t.Error("Redact() returned the same object, not a clone")
	}
}

func TestRedactInPlaceModifiesOriginal(t *testing.T) {
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						URL: "https://example.com",
						Headers: []Headers{
							{Name: "Authorization", Value: "Bearer token"},
						},
					},
					Response: Response{Status: 200},
				},
			},
		},
	}

	opts := RedactOptions{
		Headers:     []string{"Authorization"},
		Replacement: "[REDACTED]",
	}
	h.RedactInPlace(opts)

	if h.Log.Entries[0].Request.Headers[0].Value != "[REDACTED]" {
		t.Errorf("RedactInPlace() did not modify original, got: %s", h.Log.Entries[0].Request.Headers[0].Value)
	}
}

func TestRedactCustomRedactor(t *testing.T) {
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						URL: "https://example.com/api?token=secret&page=2",
						Headers: []Headers{
							{Name: "Authorization", Value: "Bearer tok"},
						},
						Cookies: []Cookie{
							{Name: "session", Value: "sess123"},
						},
						QueryString: []QueryString{
							{Name: "token", Value: "secret"},
						},
					},
					Response: Response{Status: 200},
				},
			},
		},
	}

	opts := RedactOptions{
		Headers:     []string{"Authorization"},
		Cookies:     []string{"session"},
		QueryParams: []string{"token"},
		CustomRedactor: func(fieldType string, name string, value string) string {
			return fieldType + ":" + name + ":custom"
		},
		Replacement: "[REDACTED]",
	}
	result := h.Redact(opts)

	// Header
	headerVal := result.Log.Entries[0].Request.Headers[0].Value
	if headerVal != "header:Authorization:custom" {
		t.Errorf("custom redactor not used for header, got: %s", headerVal)
	}

	// Cookie
	cookieVal := result.Log.Entries[0].Request.Cookies[0].Value
	if cookieVal != "cookie:session:custom" {
		t.Errorf("custom redactor not used for cookie, got: %s", cookieVal)
	}

	// Query string param
	qsVal := result.Log.Entries[0].Request.QueryString[0].Value
	if qsVal != "queryparam:token:custom" {
		t.Errorf("custom redactor not used for query param, got: %s", qsVal)
	}

	// URL should also use custom redactor
	resultURL := result.Log.Entries[0].Request.URL
	if !strings.Contains(resultURL, "token=queryparam:token:custom") {
		t.Errorf("custom redactor not used in URL, got: %s", resultURL)
	}
}

func TestRedactEmptyOptions(t *testing.T) {
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						URL:    "https://example.com/api?token=secret",
						Method: "GET",
						Headers: []Headers{
							{Name: "Authorization", Value: "Bearer token"},
							{Name: "Content-Type", Value: "application/json"},
						},
						Cookies: []Cookie{
							{Name: "session", Value: "abc"},
						},
						QueryString: []QueryString{
							{Name: "token", Value: "secret"},
						},
					},
					Response: Response{
						Status: 200,
						Headers: []Headers{
							{Name: "Content-Type", Value: "application/json"},
						},
					},
					ServerIPAddress: "10.0.0.1",
				},
			},
		},
	}

	opts := RedactOptions{} // empty options
	result := h.Redact(opts)

	// Nothing should be redacted since no fields are specified
	if result.Log.Entries[0].Request.Headers[0].Value != "Bearer token" {
		t.Error("empty options redacted Authorization header")
	}
	if result.Log.Entries[0].Request.Headers[1].Value != "application/json" {
		t.Error("empty options changed Content-Type header")
	}
	if result.Log.Entries[0].Request.Cookies[0].Value != "abc" {
		t.Error("empty options redacted cookie")
	}
	if result.Log.Entries[0].Request.QueryString[0].Value != "secret" {
		t.Error("empty options redacted query param")
	}
	if result.Log.Entries[0].ServerIPAddress != "10.0.0.1" {
		t.Error("empty options changed IP (RedactIPs is false)")
	}
}

func TestRedactCaseInsensitive(t *testing.T) {
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						URL: "https://example.com",
						Headers: []Headers{
							{Name: "authorization", Value: "Bearer token"}, // lowercase
							{Name: "AUTHORIZATION", Value: "Bearer token"}, // uppercase
						},
					},
					Response: Response{Status: 200},
				},
			},
		},
	}

	opts := RedactOptions{
		Headers:     []string{"Authorization"},
		Replacement: "[REDACTED]",
	}
	result := h.Redact(opts)

	if result.Log.Entries[0].Request.Headers[0].Value != "[REDACTED]" {
		t.Errorf("lowercase authorization not redacted, got: %s", result.Log.Entries[0].Request.Headers[0].Value)
	}
	if result.Log.Entries[0].Request.Headers[1].Value != "[REDACTED]" {
		t.Errorf("uppercase AUTHORIZATION not redacted, got: %s", result.Log.Entries[0].Request.Headers[1].Value)
	}
}

func TestRedactURLPathRules(t *testing.T) {
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						URL: "https://example.com/api/users/12345/profile",
					},
					Response: Response{Status: 200},
				},
			},
		},
	}

	opts := RedactOptions{
		RedactURLs: []RedactURLRule{
			{Pattern: `^\d+$`, Replacement: "[ID]"},
		},
		Replacement: "[REDACTED]",
	}
	result := h.Redact(opts)

	resultURL := result.Log.Entries[0].Request.URL
	if !strings.Contains(resultURL, "users") && !strings.Contains(resultURL, "profile") {
		t.Errorf("URL path structure lost, got: %s", resultURL)
	}
	// The brackets in [ID] get URL-encoded by url.String()
	if !strings.Contains(resultURL, "[ID]") && !strings.Contains(resultURL, "%5BID%5D") {
		t.Errorf("URL path segment not redacted, got: %s", resultURL)
	}
}

func TestAnonymizeIP(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"192.168.1.42", "192.168.1.0"},
		{"10.0.0.255", "10.0.0.0"},
		{"127.0.0.1", "127.0.0.0"},
		{"not-an-ip", "not-an-ip"},
	}

	for _, tt := range tests {
		result := anonymizeIP(tt.input)
		if result != tt.expected {
			t.Errorf("anonymizeIP(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestRedactDefaultOptionsHasExpectedFields(t *testing.T) {
	opts := DefaultRedactOptions()

	if opts.Replacement != "[REDACTED]" {
		t.Errorf("default replacement should be [REDACTED], got: %s", opts.Replacement)
	}
	if opts.RedactIPs != false {
		t.Error("default RedactIPs should be false")
	}
	if len(opts.Headers) == 0 {
		t.Error("default Headers should not be empty")
	}
	if len(opts.Cookies) == 0 {
		t.Error("default Cookies should not be empty")
	}
	if len(opts.QueryParams) == 0 {
		t.Error("default QueryParams should not be empty")
	}
	if len(opts.PostDataFields) == 0 {
		t.Error("default PostDataFields should not be empty")
	}
}
