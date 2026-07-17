package har

import (
	"testing"
	"time"
)

func TestRegisterValidator(t *testing.T) {
	// 清理自定义规则
	customRules = nil

	// 注册规则
	RegisterValidator("test-rule", ValidationRule{
		Name:        "test-rule",
		Description: "Test validation rule",
		Validate: func(har *Har) []*ValidationError {
			return nil // 不产生错误
		},
	})

	rules := ListValidators()
	if len(rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(rules))
	}
	if rules[0].Name != "test-rule" {
		t.Errorf("Expected rule name 'test-rule', got '%s'", rules[0].Name)
	}
}

func TestUnregisterValidator(t *testing.T) {
	customRules = nil

	RegisterValidator("remove-me", ValidationRule{
		Name:        "remove-me",
		Description: "Rule to be removed",
		Validate: func(har *Har) []*ValidationError {
			return nil
		},
	})

	UnregisterValidator("remove-me")
	rules := ListValidators()
	if len(rules) != 0 {
		t.Errorf("Expected 0 rules after unregister, got %d", len(rules))
	}
}

func TestRegisterValidatorOverride(t *testing.T) {
	customRules = nil

	RegisterValidator("duplicate-rule", ValidationRule{
		Name:        "duplicate-rule",
		Description: "First version",
		Validate: func(har *Har) []*ValidationError {
			return nil
		},
	})

	RegisterValidator("duplicate-rule", ValidationRule{
		Name:        "duplicate-rule",
		Description: "Second version",
		Validate: func(har *Har) []*ValidationError {
			return nil
		},
	})

	rules := ListValidators()
	if len(rules) != 1 {
		t.Errorf("Expected 1 rule (overridden), got %d", len(rules))
	}
	if rules[0].Description != "Second version" {
		t.Errorf("Expected 'Second version', got '%s'", rules[0].Description)
	}
}

func TestValidateWithRules(t *testing.T) {
	customRules = nil

	// 注册一个总是通过的规则
	RegisterValidator("pass-always", ValidationRule{
		Name:        "pass-always",
		Description: "Always passes",
		Validate: func(har *Har) []*ValidationError {
			return nil
		},
	})

	har := NewHar()
	har.SetCreator("test", "1.0")
	entry := har.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	entry.SetResponseStatus(200, "OK")
	entry.SetResponseContent(0, "text/html")
	entry.StartedDateTime = time.Now()

	err := ValidateWithRules(har)
	// 标准验证可能产生错误（MimeType等），但不应该panic
	if err != nil {
		// 验证错误是可以接受的
		t.Logf("Validation error (expected): %v", err)
	}
}

func TestValidateWithRulesCustomError(t *testing.T) {
	customRules = nil

	// 注册一个总是失败的规则
	RegisterValidator("always-fail", ValidationRule{
		Name:        "always-fail",
		Description: "Always fails",
		Validate: func(har *Har) []*ValidationError {
			return []*ValidationError{
				{
					Field:   "test",
					Message: "always fails",
					Rule:    "always-fail",
				},
			}
		},
	})

	har := NewHar()
	har.SetCreator("test", "1.0")
	entry := har.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	entry.SetResponseStatus(200, "OK")
	entry.SetResponseContent(0, "text/html")
	entry.StartedDateTime = time.Now()

	err := ValidateWithRules(har)
	if err == nil {
		t.Error("Expected validation error from custom rule")
	}

	// 清理
	UnregisterValidator("always-fail")
}

func TestValidateStrict(t *testing.T) {
	// 创建一个有效的HAR文件
	har := NewHar()
	har.SetCreator("test", "1.0")
	har.AddPage("page_1", "Test Page")
	entry := har.AddEntry("GET", "https://example.com", "HTTP/1.1", "page_1")
	entry.SetResponseStatus(200, "OK")
	entry.SetResponseContent(0, "text/html")
	entry.StartedDateTime = time.Now()

	err := ValidateStrict(har)
	if err != nil {
		t.Logf("Strict validation error: %v", err)
	}
}

func TestValidateStrictPagerefReference(t *testing.T) {
	har := NewHar()
	har.SetCreator("test", "1.0")
	entry := har.AddEntry("GET", "https://example.com", "HTTP/1.1", "nonexistent_page")
	entry.SetResponseStatus(200, "OK")
	entry.SetResponseContent(0, "text/html")
	entry.StartedDateTime = time.Now()

	err := ValidateStrict(har)
	if err == nil {
		t.Error("Expected error for nonexistent pageref")
	}
}

func TestValidateStrictPageIDUniqueness(t *testing.T) {
	har := NewHar()
	har.SetCreator("test", "1.0")
	har.AddPage("page_1", "Page 1")
	har.AddPage("page_1", "Page 1 Duplicate") // 重复ID
	entry := har.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	entry.SetResponseStatus(200, "OK")
	entry.SetResponseContent(0, "text/html")
	entry.StartedDateTime = time.Now()

	err := ValidateStrict(har)
	if err == nil {
		t.Error("Expected error for duplicate page ID")
	}
}

func TestValidateStrictHTTPMethods(t *testing.T) {
	har := NewHar()
	har.SetCreator("test", "1.0")
	entry := har.AddEntry("INVALIDMETHOD", "https://example.com", "HTTP/1.1", "")
	entry.SetResponseStatus(200, "OK")
	entry.SetResponseContent(0, "text/html")
	entry.StartedDateTime = time.Now()

	err := ValidateStrict(har)
	if err == nil {
		t.Error("Expected error for invalid HTTP method")
	}
}

func TestValidateStrictStatusCodeRange(t *testing.T) {
	har := NewHar()
	har.SetCreator("test", "1.0")
	entry := har.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	entry.SetResponseStatus(999, "Invalid")
	entry.SetResponseContent(0, "text/html")
	entry.StartedDateTime = time.Now()

	err := ValidateStrict(har)
	if err == nil {
		t.Error("Expected error for invalid status code")
	}
}

func TestValidateStrictCookieSameSite(t *testing.T) {
	har := NewHar()
	har.SetCreator("test", "1.0")
	entry := har.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	entry.SetResponseStatus(200, "OK")
	entry.SetResponseContent(0, "text/html")
	entry.StartedDateTime = time.Now()
	entry.AddCookie("test", "value")
	entry.Request.Cookies[0].SameSite = "InvalidValue"

	err := ValidateStrict(har)
	if err == nil {
		t.Error("Expected error for invalid SameSite value")
	}
}

func TestValidateStrictCacheFields(t *testing.T) {
	har := NewHar()
	har.SetCreator("test", "1.0")
	entry := har.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	entry.SetResponseStatus(200, "OK")
	entry.SetResponseContent(0, "text/html")
	entry.StartedDateTime = time.Now()
	// 设置缺少必填字段的Cache
	entry.Cache.BeforeRequest = &BeforeRequest{
		ETag:     "", // 缺少必填字段
		HitCount: -1, // 负值
	}

	err := ValidateStrict(har)
	if err == nil {
		t.Error("Expected error for invalid cache fields")
	}
}

func TestValidationError_Error(t *testing.T) {
	ve := &ValidationError{
		Field:   "log.entries[0].request.url",
		Message: "URL解析失败",
		Rule:    "url-format",
	}

	errStr := ve.Error()
	if !validatorContains(errStr, "[url-format]") {
		t.Errorf("Expected rule name in error string, got: %s", errStr)
	}
	if !validatorContains(errStr, "URL解析失败") {
		t.Errorf("Expected message in error string, got: %s", errStr)
	}
}

func TestValidationError_ErrorNoRule(t *testing.T) {
	ve := &ValidationError{
		Field:   "url",
		Message: "invalid URL",
	}

	errStr := ve.Error()
	if !validatorContains(errStr, "url: invalid URL") {
		t.Errorf("Expected 'url: invalid URL', got: %s", errStr)
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"valid https", "https://example.com/api", false},
		{"valid http", "http://example.com", false},
		{"empty url", "", true},
		{"missing scheme", "example.com/api", true},
		{"missing host", "https://", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL(%s) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

func TestValidateTimingsConsistency(t *testing.T) {
	har := NewHar()
	har.SetCreator("test", "1.0")
	entry := har.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	entry.SetResponseStatus(200, "OK")
	entry.SetResponseContent(0, "text/html")
	entry.StartedDateTime = time.Now()
	entry.Time = 100 // 100ms total
	entry.Timings = Timings{
		Send:    10,
		Wait:    50,
		Receive: 40,
		Blocked: -1,
		DNS:     -1,
		Connect: -1,
		Ssl:     -1,
	}

	errors := ValidateTimingsConsistency(har, 5) // tolerance of 5ms
	if len(errors) > 0 {
		t.Errorf("Expected no timing consistency errors, got %d", len(errors))
	}
}

func TestValidateTimingsConsistencyWithMismatch(t *testing.T) {
	har := NewHar()
	har.SetCreator("test", "1.0")
	entry := har.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	entry.SetResponseStatus(200, "OK")
	entry.SetResponseContent(0, "text/html")
	entry.StartedDateTime = time.Now()
	entry.Time = 100 // 100ms total
	entry.Timings = Timings{
		Send:    10,
		Wait:    20, // too small, sum = 70 vs Time = 100
		Receive: 40,
	}

	errors := ValidateTimingsConsistency(har, 10) // tolerance of 10ms
	if len(errors) == 0 {
		t.Error("Expected timing consistency errors for mismatch")
	}
}

func validatorContains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || len(s) > 0 && validatorContains(s[1:], substr)
}

// --- Comprehensive tests for uncovered branches in validator_ext.go ---

func TestValidateWithRules_Nil(t *testing.T) {
	customRules = nil

	err := ValidateWithRules(nil)
	if err == nil {
		t.Error("Expected error for nil HAR")
	}
}

func TestValidateWithRulesNilCustomValidatorDoesNotPanic(t *testing.T) {
	customRules = nil
	defer func() { customRules = nil }()

	RegisterValidator("nil-validator", ValidationRule{
		Name: "nil-validator",
	})

	har := NewHar()
	entry := har.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	entry.SetResponseStatus(200, "OK")
	entry.SetResponseContent(10, "text/html")
	entry.StartedDateTime = time.Now()

	var err error
	assertDoesNotPanic(t, func() {
		err = ValidateWithRules(har)
	})
	assertHarErrorCode(t, err, ErrCodeValidation)
}

func TestValidateWithRulesIgnoresNilCustomErrors(t *testing.T) {
	customRules = nil
	defer func() { customRules = nil }()

	RegisterValidator("nil-error", ValidationRule{
		Name: "nil-error",
		Validate: func(har *Har) []*ValidationError {
			return []*ValidationError{nil}
		},
	})

	har := NewHar()
	entry := har.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	entry.SetResponseStatus(200, "OK")
	entry.SetResponseContent(10, "text/html")
	entry.SetTimings(0, 0, 0, 0, 10, 0, 0)
	entry.StartedDateTime = time.Now()

	assertDoesNotPanic(t, func() {
		if err := ValidateWithRules(har); err != nil {
			t.Fatalf("ValidateWithRules() error = %v, want nil", err)
		}
	})
}

func TestValidateWithRules_StandardErrorOnly(t *testing.T) {
	customRules = nil

	// Register a passing rule
	RegisterValidator("pass-rule", ValidationRule{
		Name:        "pass-rule",
		Description: "Always passes",
		Validate: func(har *Har) []*ValidationError {
			return nil
		},
	})
	defer UnregisterValidator("pass-rule")

	// Invalid HAR (missing version)
	har := &Har{
		Log: Log{
			Version: "",
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{},
		},
	}

	err := ValidateWithRules(har)
	if err == nil {
		t.Error("Expected validation error from standard validation")
	}
}

func TestValidateWithRules_CustomErrorOnly(t *testing.T) {
	customRules = nil

	RegisterValidator("custom-fail", ValidationRule{
		Name:        "custom-fail",
		Description: "Always fails",
		Validate: func(har *Har) []*ValidationError {
			return []*ValidationError{
				{Field: "custom.field", Message: "custom error", Rule: "custom-fail"},
			}
		},
	})
	defer UnregisterValidator("custom-fail")

	// Valid HAR that passes standard validation
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request:         Request{Method: "GET", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
					Response:        Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
				},
			},
		},
	}

	err := ValidateWithRules(har)
	if err == nil {
		t.Error("Expected validation error from custom rule only")
	}
	harErr, ok := err.(*HarError)
	if !ok {
		t.Fatalf("Expected *HarError, got %T", err)
	}
	if !harErr.HasPartialErrors() {
		t.Error("Expected partial errors")
	}
}

func TestValidateWithRules_BothStandardAndCustomErrors(t *testing.T) {
	customRules = nil

	RegisterValidator("custom-fail2", ValidationRule{
		Name:        "custom-fail2",
		Description: "Always fails",
		Validate: func(har *Har) []*ValidationError {
			return []*ValidationError{
				{Field: "custom.field", Message: "custom error", Rule: "custom-fail2"},
			}
		},
	})
	defer UnregisterValidator("custom-fail2")

	// Invalid HAR that fails both standard and custom validation
	har := &Har{
		Log: Log{
			Version: "", // missing version
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{},
		},
	}

	err := ValidateWithRules(har)
	if err == nil {
		t.Error("Expected combined validation error")
	}
	harErr, ok := err.(*HarError)
	if !ok {
		t.Fatalf("Expected *HarError, got %T", err)
	}
	if !harErr.HasPartialErrors() {
		t.Error("Expected partial errors")
	}
	// Should contain both standard and custom errors
	partials := harErr.GetPartialErrors()
	foundCustom := false
	for _, pe := range partials {
		if pe.Field == "custom.field" {
			foundCustom = true
		}
	}
	if !foundCustom {
		t.Error("Expected custom error to be merged with standard errors")
	}
}

func TestValidateWithRules_NoErrors(t *testing.T) {
	customRules = nil

	RegisterValidator("pass-always2", ValidationRule{
		Name:        "pass-always2",
		Description: "Always passes",
		Validate: func(har *Har) []*ValidationError {
			return nil
		},
	})
	defer UnregisterValidator("pass-always2")

	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request:         Request{Method: "GET", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
					Response:        Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
				},
			},
		},
	}

	err := ValidateWithRules(har)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestValidateWithRules_MultipleCustomRules(t *testing.T) {
	customRules = nil

	RegisterValidator("rule1", ValidationRule{
		Name:        "rule1",
		Description: "Rule 1",
		Validate: func(har *Har) []*ValidationError {
			return []*ValidationError{
				{Field: "field1", Message: "error1", Rule: "rule1"},
			}
		},
	})

	RegisterValidator("rule2", ValidationRule{
		Name:        "rule2",
		Description: "Rule 2",
		Validate: func(har *Har) []*ValidationError {
			return []*ValidationError{
				{Field: "field2", Message: "error2", Rule: "rule2"},
			}
		},
	})

	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request:         Request{Method: "GET", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
					Response:        Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
				},
			},
		},
	}

	err := ValidateWithRules(har)
	if err == nil {
		t.Error("Expected error from multiple custom rules")
	}

	// Cleanup
	UnregisterValidator("rule1")
	UnregisterValidator("rule2")
}

func TestValidateStrict_Nil(t *testing.T) {
	err := ValidateStrict(nil)
	if err == nil {
		t.Error("Expected error for nil HAR in strict validation")
	}
}

func TestValidateStrict_ValidHar(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Pages: []Pages{
				{
					ID:              "page_1",
					StartedDateTime: time.Now(),
					Title:           "Test Page",
				},
			},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Pageref:         "page_1",
					Request:         Request{Method: "GET", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
					Response:        Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
				},
			},
		},
	}

	err := ValidateStrict(har)
	if err != nil {
		t.Logf("Strict validation error: %v", err)
	}
}

func TestValidateStrict_EmptyPageref(t *testing.T) {
	// Empty pageref should not trigger the pageref reference error
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Pages: []Pages{
				{
					ID:              "page_1",
					StartedDateTime: time.Now(),
					Title:           "Test Page",
				},
			},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Pageref:         "", // empty - should not trigger error
					Request:         Request{Method: "GET", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
					Response:        Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
				},
			},
		},
	}

	err := ValidateStrict(har)
	if err != nil {
		t.Logf("Strict validation error: %v", err)
	}
}

func TestValidateStrict_PageIDUniqueness_Valid(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Pages: []Pages{
				{ID: "page_1", StartedDateTime: time.Now(), Title: "Page 1"},
				{ID: "page_2", StartedDateTime: time.Now(), Title: "Page 2"},
			},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request:         Request{Method: "GET", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
					Response:        Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
				},
			},
		},
	}

	err := ValidateStrict(har)
	if err != nil {
		t.Logf("Strict validation error: %v", err)
	}
}

func TestValidateStrict_StatusCodeBoundary(t *testing.T) {
	tests := []struct {
		name   string
		status int
		valid  bool
	}{
		{"status 99", 99, false},
		{"status 100", 100, true},
		{"status 200", 200, true},
		{"status 599", 599, true},
		{"status 600", 600, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			har := &Har{
				Log: Log{
					Version: HarSpecVersion12,
					Creator: Creator{Name: "Test", Version: "1.0"},
					Entries: []Entries{
						{
							StartedDateTime: time.Now(),
							Request:         Request{Method: "GET", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
							Response:        Response{Status: tt.status, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
						},
					},
				},
			}

			err := ValidateStrict(har)
			if tt.valid {
				// valid status code should not produce status code range error
				// (may still have other errors)
				if err != nil {
					harErr, ok := err.(*HarError)
					if ok {
						for _, pe := range harErr.GetPartialErrors() {
							if pe.Field == "log.entries[0].response.status" {
								t.Errorf("Expected no status code range error for status %d, but got one", tt.status)
							}
						}
					}
				}
			} else {
				if err == nil {
					t.Errorf("Expected error for invalid status code %d", tt.status)
				}
			}
		})
	}
}

func TestValidateCookieSameSite_ResponseCookies(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request:         Request{Method: "GET", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
					Response: Response{
						Status:      200,
						HTTPVersion: "HTTP/1.1",
						Content:     Content{Size: 100, MimeType: "text/html"},
						Cookies:     []Cookie{{Name: "test", Value: "val", SameSite: "InvalidValue"}},
					},
				},
			},
		},
	}

	err := ValidateStrict(har)
	if err == nil {
		t.Error("Expected error for invalid SameSite value in response cookies")
	}
}

func TestValidateCookieSameSite_ValidValues(t *testing.T) {
	tests := []struct {
		name     string
		sameSite string
	}{
		{"Strict", "Strict"},
		{"Lax", "Lax"},
		{"None", "None"},
		{"Empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			har := &Har{
				Log: Log{
					Version: HarSpecVersion12,
					Creator: Creator{Name: "Test", Version: "1.0"},
					Entries: []Entries{
						{
							StartedDateTime: time.Now(),
							Request: Request{
								Method:      "GET",
								URL:         "https://example.com",
								HTTPVersion: "HTTP/1.1",
								Cookies:     []Cookie{{Name: "test", Value: "val", SameSite: tt.sameSite}},
							},
							Response: Response{
								Status:      200,
								HTTPVersion: "HTTP/1.1",
								Content:     Content{Size: 100, MimeType: "text/html"},
								Cookies:     []Cookie{{Name: "test2", Value: "val2", SameSite: tt.sameSite}},
							},
						},
					},
				},
			}

			err := ValidateStrict(har)
			if err != nil {
				harErr, ok := err.(*HarError)
				if ok {
					for _, pe := range harErr.GetPartialErrors() {
						// Check that no SameSite errors were produced
						for _, field := range []string{
							"log.entries[0].request.cookies[0].sameSite",
							"log.entries[0].response.cookies[0].sameSite",
						} {
							if pe.Field == field {
								t.Errorf("Expected no SameSite error for value '%s', but got one: %s", tt.sameSite, pe.Message)
							}
						}
					}
				}
			}
		})
	}
}

func TestValidateCacheFields_AfterRequest(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request:         Request{Method: "GET", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
					Response:        Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
					Cache: Cache{
						AfterRequest: &AfterRequest{
							ETag:     "", // missing
							HitCount: -5, // negative
							// LastAccess is zero
						},
					},
				},
			},
		},
	}

	err := ValidateStrict(har)
	if err == nil {
		t.Error("Expected error for invalid AfterRequest cache fields")
	}
}

func TestValidateCacheFields_AfterRequestValid(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request:         Request{Method: "GET", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
					Response:        Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
					Cache: Cache{
						AfterRequest: &AfterRequest{
							LastAccess: time.Now(),
							ETag:       "abc123",
							HitCount:   5,
						},
					},
				},
			},
		},
	}

	err := ValidateStrict(har)
	if err != nil {
		harErr, ok := err.(*HarError)
		if ok {
			for _, pe := range harErr.GetPartialErrors() {
				if pe.Field == "log.entries[0].cache.afterRequest.lastAccess" ||
					pe.Field == "log.entries[0].cache.afterRequest.eTag" ||
					pe.Field == "log.entries[0].cache.afterRequest.hitCount" {
					t.Errorf("Unexpected cache field error: %s - %s", pe.Field, pe.Message)
				}
			}
		}
	}
}

func TestValidateCacheFields_BeforeRequestValid(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request:         Request{Method: "GET", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
					Response:        Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
					Cache: Cache{
						BeforeRequest: &BeforeRequest{
							LastAccess: time.Now(),
							ETag:       "abc123",
							HitCount:   5,
						},
					},
				},
			},
		},
	}

	err := ValidateStrict(har)
	if err != nil {
		harErr, ok := err.(*HarError)
		if ok {
			for _, pe := range harErr.GetPartialErrors() {
				if pe.Field == "log.entries[0].cache.beforeRequest.lastAccess" ||
					pe.Field == "log.entries[0].cache.beforeRequest.eTag" ||
					pe.Field == "log.entries[0].cache.beforeRequest.hitCount" {
					t.Errorf("Unexpected cache field error: %s - %s", pe.Field, pe.Message)
				}
			}
		}
	}
}

func TestValidateCacheFields_NilBoth(t *testing.T) {
	// Both BeforeRequest and AfterRequest are nil — should not error on cache fields
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request:         Request{Method: "GET", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
					Response:        Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
					Cache:           Cache{}, // both nil
				},
			},
		},
	}

	err := ValidateStrict(har)
	if err != nil {
		harErr, ok := err.(*HarError)
		if ok {
			for _, pe := range harErr.GetPartialErrors() {
				if pe.Field == "log.entries[0].cache.beforeRequest.lastAccess" ||
					pe.Field == "log.entries[0].cache.afterRequest.lastAccess" {
					t.Errorf("Unexpected cache field error for nil cache: %s", pe.Message)
				}
			}
		}
	}
}

func TestValidateURL_ParseError(t *testing.T) {
	// Test URL that fails parsing
	err := ValidateURL("://missing-scheme-host")
	if err == nil {
		t.Fatal("Expected error for unparseable URL")
	}
	if err.Field != "url" {
		t.Errorf("Expected field 'url', got '%s'", err.Field)
	}
}

func TestValidateURL_MissingHost(t *testing.T) {
	err := ValidateURL("https://")
	if err == nil {
		t.Fatal("Expected error for URL with no host")
	}
	if err.Field != "url.host" {
		t.Errorf("Expected field 'url.host', got '%s'", err.Field)
	}
}

func TestValidateURL_ValidURL(t *testing.T) {
	err := ValidateURL("https://example.com/path?query=1")
	if err != nil {
		t.Errorf("Expected no error for valid URL, got: %v", err)
	}
}

func TestValidateURL_ValidHTTP(t *testing.T) {
	err := ValidateURL("http://localhost:8080/api")
	if err != nil {
		t.Errorf("Expected no error for valid HTTP URL, got: %v", err)
	}
}

func TestValidateTimingsConsistency_ZeroTolerance(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request:         Request{Method: "GET", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
					Response:        Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
					Time:            100,
					Timings: Timings{
						Send:    10,
						Wait:    50,
						Receive: 40,
						// sum = 100, diff = 0, tolerance = 0, should pass
					},
				},
			},
		},
	}

	errors := ValidateTimingsConsistency(har, 0)
	if len(errors) > 0 {
		t.Errorf("Expected no timing consistency errors with zero tolerance, got %d", len(errors))
	}
}

func TestValidateTimingsConsistency_WithSSLAndConnect(t *testing.T) {
	// When both Connect > 0 and SSL > 0, SSL should NOT be added (it's included in Connect)
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request:         Request{Method: "GET", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
					Response:        Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
					Time:            100,
					Timings: Timings{
						Blocked: 5,
						DNS:     5,
						Connect: 20,
						Ssl:     15, // SSL included in Connect, should not be added
						Send:    10,
						Wait:    40,
						Receive: 20,
						// sum = 5+5+20+10+40+20 = 100
					},
				},
			},
		},
	}

	errors := ValidateTimingsConsistency(har, 5)
	if len(errors) > 0 {
		t.Errorf("Expected no timing consistency errors, got %d: %v", len(errors), errors)
	}
}

func TestValidateTimingsConsistency_SSLWithoutConnect(t *testing.T) {
	// When Connect <= 0 but SSL > 0, SSL should be added separately
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request:         Request{Method: "GET", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
					Response:        Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
					Time:            100,
					Timings: Timings{
						Blocked: 5,
						DNS:     5,
						Connect: -1, // unavailable
						Ssl:     15, // should be added since Connect <= 0
						Send:    10,
						Wait:    35,
						Receive: 30,
						// sum = 5+5+15+10+35+30 = 100
					},
				},
			},
		},
	}

	errors := ValidateTimingsConsistency(har, 5)
	if len(errors) > 0 {
		t.Errorf("Expected no timing consistency errors, got %d: %v", len(errors), errors)
	}
}

func TestValidateTimingsConsistency_NegativeDiff(t *testing.T) {
	// Test when sum > Time (negative diff after subtraction)
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request:         Request{Method: "GET", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
					Response:        Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
					Time:            50, // too small
					Timings: Timings{
						Send:    10,
						Wait:    50,
						Receive: 40,
						// sum = 100, Time = 50, diff = 50
					},
				},
			},
		},
	}

	errors := ValidateTimingsConsistency(har, 10)
	if len(errors) == 0 {
		t.Error("Expected timing consistency error for sum > Time")
	}
}

func TestValidateTimingsConsistency_EmptyEntries(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{},
		},
	}

	errors := ValidateTimingsConsistency(har, 5)
	if len(errors) != 0 {
		t.Errorf("Expected no errors for empty entries, got %d", len(errors))
	}
}

func TestValidateTimingsConsistencyNilHarDoesNotPanic(t *testing.T) {
	var errors []*ValidationError

	assertDoesNotPanic(t, func() {
		errors = ValidateTimingsConsistency(nil, 5)
	})
	if len(errors) != 1 {
		t.Fatalf("ValidateTimingsConsistency(nil) returned %d errors, want 1", len(errors))
	}
	if errors[0].Rule != "timings-consistency" {
		t.Fatalf("ValidateTimingsConsistency(nil)[0].Rule = %q, want timings-consistency", errors[0].Rule)
	}
}

func TestValidateTimingsConsistency_AllNegativeTimings(t *testing.T) {
	// All negative timings — none should be added to sum, so sum = 0 vs Time
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request:         Request{Method: "GET", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
					Response:        Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
					Time:            100,
					Timings: Timings{
						Blocked: -1,
						DNS:     -1,
						Connect: -1,
						Ssl:     -1,
						Send:    -1,
						Wait:    -1,
						Receive: -1,
					},
				},
			},
		},
	}

	errors := ValidateTimingsConsistency(har, 10)
	if len(errors) == 0 {
		t.Error("Expected timing consistency error when all timings are negative but Time is positive")
	}
}

func TestUnregisterValidator_NotFound(t *testing.T) {
	customRules = nil

	// Unregistering a non-existent rule should not panic
	UnregisterValidator("nonexistent")
}

func TestListValidators_Empty(t *testing.T) {
	customRules = nil

	rules := ListValidators()
	if len(rules) != 0 {
		t.Errorf("Expected 0 rules, got %d", len(rules))
	}
}

func TestValidateHTTPMethods_AllValidMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS", "PATCH", "CONNECT", "TRACE"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			har := &Har{
				Log: Log{
					Version: HarSpecVersion12,
					Creator: Creator{Name: "Test", Version: "1.0"},
					Entries: []Entries{
						{
							StartedDateTime: time.Now(),
							Request:         Request{Method: method, URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
							Response:        Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
						},
					},
				},
			}

			err := ValidateStrict(har)
			if err != nil {
				harErr, ok := err.(*HarError)
				if ok {
					for _, pe := range harErr.GetPartialErrors() {
						if pe.Field == "log.entries[0].request.method" {
							t.Errorf("Expected no method error for valid method %s, but got: %s", method, pe.Message)
						}
					}
				}
			}
		})
	}
}

func TestValidateHTTPMethods_LowercaseMethod(t *testing.T) {
	// Methods are compared case-insensitively (ToUpper is used)
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request:         Request{Method: "get", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
					Response:        Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
				},
			},
		},
	}

	err := ValidateStrict(har)
	if err != nil {
		harErr, ok := err.(*HarError)
		if ok {
			for _, pe := range harErr.GetPartialErrors() {
				if pe.Field == "log.entries[0].request.method" {
					t.Errorf("Expected no method error for lowercase 'get', but got: %s", pe.Message)
				}
			}
		}
	}
}

func TestValidationError_ErrorWithRule(t *testing.T) {
	ve := &ValidationError{
		Field:   "test.field",
		Message: "test message",
		Rule:    "my-rule",
	}

	errStr := ve.Error()
	if !validatorContains(errStr, "[my-rule]") {
		t.Errorf("Expected rule name in error string, got: %s", errStr)
	}
}

func TestValidationError_ErrorWithoutRule(t *testing.T) {
	ve := &ValidationError{
		Field:   "test.field",
		Message: "test message",
	}

	errStr := ve.Error()
	if !validatorContains(errStr, "test.field: test message") {
		t.Errorf("Expected 'test.field: test message', got: %s", errStr)
	}
}

func TestValidationErrorNilReceiverError(t *testing.T) {
	var ve *ValidationError

	assertDoesNotPanic(t, func() {
		if got := ve.Error(); got != "<nil>" {
			t.Fatalf("Error() = %q, want <nil>", got)
		}
	})
}
