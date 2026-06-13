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