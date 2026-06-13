package har

import (
	"strings"
	"testing"
	"time"
)

func TestSecurityMissingHeaders(t *testing.T) {
	har := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						URL:    "https://example.com/",
						Method: "GET",
						Headers: []Headers{
							{Name: "Host", Value: "example.com"},
						},
					},
					Response: Response{
						Headers: []Headers{
							{Name: "Content-Type", Value: "text/html"},
						},
					},
				},
			},
		},
	}

	report := har.SecurityAudit()
	findings := report.FindByCategory("headers")

	// Should find all 7 missing security headers
	if len(findings) < 7 {
		t.Errorf("expected at least 7 header findings, got %d", len(findings))
	}

	// Check for specific headers
	titles := make(map[string]bool)
	for _, f := range findings {
		titles[f.Title] = true
	}

	expectedTitles := []string{
		"Missing Strict-Transport-Security header",
		"Missing Content-Security-Policy header",
		"Missing X-Content-Type-Options header",
		"Missing X-Frame-Options header",
		"Missing X-XSS-Protection header",
		"Missing Referrer-Policy header",
		"Missing Permissions-Policy header",
	}
	for _, title := range expectedTitles {
		if !titles[title] {
			t.Errorf("expected finding: %s", title)
		}
	}

	// HSTS should be HIGH on HTTPS
	hstsFindings := filterFindings(findings, func(f SecurityFinding) bool {
		return strings.Contains(f.Title, "Strict-Transport-Security")
	})
	if len(hstsFindings) > 0 && hstsFindings[0].Severity != "high" {
		t.Errorf("expected HSTS finding to be HIGH on HTTPS, got %s", hstsFindings[0].Severity)
	}
}

func TestSecurityMissingHeadersHTTP(t *testing.T) {
	har := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						URL:    "http://example.com/",
						Method: "GET",
					},
					Response: Response{
						Headers: []Headers{},
					},
				},
			},
		},
	}

	report := har.SecurityAudit()
	findings := report.FindByCategory("headers")

	// HSTS should be INFO on HTTP
	hstsFindings := filterFindings(findings, func(f SecurityFinding) bool {
		return strings.Contains(f.Title, "Strict-Transport-Security")
	})
	if len(hstsFindings) == 0 {
		t.Fatal("expected HSTS finding")
	}
	if hstsFindings[0].Severity != "info" {
		t.Errorf("expected HSTS finding to be INFO on HTTP, got %s", hstsFindings[0].Severity)
	}
}

func TestSecurityInsecureCookies(t *testing.T) {
	har := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						URL: "https://example.com/",
						Cookies: []Cookie{
							{Name: "sessionid", Value: "abc123", Secure: false, HTTPOnly: false, SameSite: "none"},
						},
					},
					Response: Response{
						Headers: []Headers{},
						Cookies: []Cookie{
							{Name: "auth_token", Value: "xyz789", Secure: false, HTTPOnly: false, SameSite: "none"},
						},
					},
				},
			},
		},
	}

	report := har.SecurityAuditWithOptions(SecurityAuditOptions{
		CheckCookies: true,
	})
	findings := report.FindByCategory("cookies")

	// Should find: Secure flag missing, HttpOnly missing (session name), SameSite=None without Secure
	// Plus session cookie without expiration (INFO) for each cookie
	// Request cookie (sessionid): no Secure (HIGH), no HttpOnly (MEDIUM), SameSite=None without Secure (HIGH), no Expires (INFO)
	// Response cookie (auth_token): no Secure (HIGH), no HttpOnly (MEDIUM), SameSite=None without Secure (HIGH), no Expires (INFO)
	if len(findings) < 6 {
		t.Errorf("expected at least 6 cookie findings, got %d", len(findings))
	}

	// Check for specific issues
	hasNoSecure := false
	hasNoHttpOnly := false
	hasSameSiteNoneNoSecure := false
	for _, f := range findings {
		if strings.Contains(f.Title, "without Secure flag") {
			hasNoSecure = true
			if f.Severity != "high" {
				t.Errorf("expected no-Secure finding to be HIGH, got %s", f.Severity)
			}
		}
		if strings.Contains(f.Title, "without HttpOnly") {
			hasNoHttpOnly = true
			if f.Severity != "medium" {
				t.Errorf("expected no-HttpOnly finding to be MEDIUM, got %s", f.Severity)
			}
		}
		if strings.Contains(f.Title, "SameSite=None") {
			hasSameSiteNoneNoSecure = true
			if f.Severity != "high" {
				t.Errorf("expected SameSite=None without Secure finding to be HIGH, got %s", f.Severity)
			}
		}
	}

	if !hasNoSecure {
		t.Error("expected finding for cookie without Secure flag")
	}
	if !hasNoHttpOnly {
		t.Error("expected finding for session cookie without HttpOnly flag")
	}
	if !hasSameSiteNoneNoSecure {
		t.Error("expected finding for SameSite=None without Secure flag")
	}
}

func TestSecurityMixedContent(t *testing.T) {
	har := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						URL: "https://example.com/",
					},
					Response: Response{
						Headers: []Headers{},
					},
				},
				{
					Request: Request{
						URL: "http://example.com/insecure.js",
					},
					Response: Response{
						Headers: []Headers{},
					},
				},
				{
					Request: Request{
						URL: "https://example.com/secure.js",
					},
					Response: Response{
						Headers: []Headers{},
					},
				},
			},
		},
	}

	report := har.SecurityAuditWithOptions(SecurityAuditOptions{
		CheckMixedContent: true,
	})
	findings := report.FindByCategory("mixed-content")

	if len(findings) != 1 {
		t.Fatalf("expected 1 mixed-content finding, got %d", len(findings))
	}

	if findings[0].Severity != "high" {
		t.Errorf("expected mixed content finding to be HIGH, got %s", findings[0].Severity)
	}
	if findings[0].EntryIndex != 1 {
		t.Errorf("expected EntryIndex 1, got %d", findings[0].EntryIndex)
	}
}

func TestSecurityMixedContentNoHTTPPage(t *testing.T) {
	har := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						URL: "http://example.com/",
					},
					Response: Response{
						Headers: []Headers{},
					},
				},
				{
					Request: Request{
						URL: "http://example.com/insecure.js",
					},
					Response: Response{
						Headers: []Headers{},
					},
				},
			},
		},
	}

	report := har.SecurityAuditWithOptions(SecurityAuditOptions{
		CheckMixedContent: true,
	})
	findings := report.FindByCategory("mixed-content")

	// No mixed content since page itself is HTTP
	if len(findings) != 0 {
		t.Errorf("expected 0 mixed-content findings for HTTP page, got %d", len(findings))
	}
}

func TestSecuritySensitiveDataInURLs(t *testing.T) {
	har := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						URL: "https://example.com/api?password=secret&api_key=abc123",
						QueryString: []QueryString{
							{Name: "password", Value: "secret"},
							{Name: "api_key", Value: "abc123"},
							{Name: "page", Value: "1"},
						},
					},
					Response: Response{
						Headers: []Headers{},
					},
				},
			},
		},
	}

	report := har.SecurityAuditWithOptions(SecurityAuditOptions{
		CheckSensitiveData: true,
	})
	findings := report.FindByCategory("sensitive-data")

	if len(findings) != 2 {
		t.Fatalf("expected 2 sensitive-data findings, got %d", len(findings))
	}

	for _, f := range findings {
		if f.Severity != "high" {
			t.Errorf("expected sensitive data finding to be HIGH, got %s", f.Severity)
		}
	}

	names := make(map[string]bool)
	for _, f := range findings {
		names[f.Description] = true
	}
	if !anyContains(names, "password") {
		t.Error("expected finding for 'password' parameter")
	}
	if !anyContains(names, "api_key") {
		t.Error("expected finding for 'api_key' parameter")
	}
}

func TestSecurityCORSMisconfiguration(t *testing.T) {
	// Test: ACAO: * + ACAC: true -> HIGH
	har := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						URL: "https://api.example.com/data",
						Headers: []Headers{
							{Name: "Authorization", Value: "Bearer token123"},
						},
					},
					Response: Response{
						Headers: []Headers{
							{Name: "Access-Control-Allow-Origin", Value: "*"},
							{Name: "Access-Control-Allow-Credentials", Value: "true"},
						},
					},
				},
			},
		},
	}

	report := har.SecurityAuditWithOptions(SecurityAuditOptions{
		CheckCORS: true,
	})
	findings := report.FindByCategory("cors")

	if len(findings) == 0 {
		t.Fatal("expected at least 1 CORS finding")
	}

	// Should find the HIGH severity credentials+wildcard issue
	hasHighCORS := false
	for _, f := range findings {
		if f.Severity == "high" && strings.Contains(f.Title, "credentials") {
			hasHighCORS = true
		}
	}
	if !hasHighCORS {
		t.Error("expected HIGH severity CORS finding for wildcard origin with credentials")
	}
}

func TestSecurityCORSPermissiveWithAuth(t *testing.T) {
	har := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						URL: "https://api.example.com/data",
						Headers: []Headers{
							{Name: "Authorization", Value: "Bearer token123"},
						},
					},
					Response: Response{
						Headers: []Headers{
							{Name: "Access-Control-Allow-Origin", Value: "*"},
						},
					},
				},
			},
		},
	}

	report := har.SecurityAuditWithOptions(SecurityAuditOptions{
		CheckCORS: true,
	})
	findings := report.FindByCategory("cors")

	if len(findings) != 1 {
		t.Fatalf("expected 1 CORS finding, got %d", len(findings))
	}

	if findings[0].Severity != "medium" {
		t.Errorf("expected MEDIUM severity for permissive CORS with auth, got %s", findings[0].Severity)
	}
}

func TestSecurityInfoDisclosureHeaders(t *testing.T) {
	har := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						URL: "https://example.com/",
					},
					Response: Response{
						Headers: []Headers{
							{Name: "Server", Value: "Apache/2.4.41"},
							{Name: "X-Powered-By", Value: "Express"},
							{Name: "X-AspNet-Version", Value: "4.0.30319"},
							{Name: "X-Generator", Value: "Hugo 0.80"},
						},
					},
				},
			},
		},
	}

	report := har.SecurityAuditWithOptions(SecurityAuditOptions{
		CheckInfoDisclosure: true,
	})
	findings := report.FindByCategory("info-disclosure")

	if len(findings) != 4 {
		t.Fatalf("expected 4 info-disclosure findings, got %d", len(findings))
	}

	// Check severities
	for _, f := range findings {
		if strings.Contains(f.Title, "Server header") && f.Severity != "low" {
			t.Errorf("expected Server header finding to be LOW, got %s", f.Severity)
		}
		if strings.Contains(f.Title, "X-Powered-By") && f.Severity != "low" {
			t.Errorf("expected X-Powered-By finding to be LOW, got %s", f.Severity)
		}
		if strings.Contains(f.Title, "X-AspNet-Version") && f.Severity != "low" {
			t.Errorf("expected X-AspNet-Version finding to be LOW, got %s", f.Severity)
		}
		if strings.Contains(f.Title, "X-Generator") && f.Severity != "info" {
			t.Errorf("expected X-Generator finding to be INFO, got %s", f.Severity)
		}
	}
}

func TestSecurityReportScoring(t *testing.T) {
	// Create a HAR with multiple issues to test scoring
	har := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						URL: "https://example.com/?password=secret",
						QueryString: []QueryString{
							{Name: "password", Value: "secret"},
						},
						Cookies: []Cookie{
							{Name: "sessionid", Value: "abc", Secure: false, SameSite: "none"},
						},
					},
					Response: Response{
						Headers: []Headers{
							{Name: "Server", Value: "nginx/1.18.0"},
						},
						Cookies: []Cookie{
							{Name: "auth_token", Value: "xyz", Secure: false, HTTPOnly: false},
						},
					},
				},
				{
					Request: Request{
						URL: "http://example.com/insecure.js",
					},
					Response: Response{
						Headers: []Headers{},
					},
				},
			},
		},
	}

	report := har.SecurityAudit()

	// Compute expected score: start at 100, subtract per finding
	highCount := len(report.FindBySeverity("high"))
	mediumCount := len(report.FindBySeverity("medium"))
	lowCount := len(report.FindBySeverity("low"))
	infoCount := len(report.FindBySeverity("info"))

	expectedScore := 100 - highCount*10 - mediumCount*5 - lowCount*2 - infoCount*0
	if expectedScore < 0 {
		expectedScore = 0
	}

	if report.Score != expectedScore {
		t.Errorf("expected score %d, got %d (high=%d, medium=%d, low=%d, info=%d)",
			expectedScore, report.Score, highCount, mediumCount, lowCount, infoCount)
	}

	if report.Score > 100 || report.Score < 0 {
		t.Errorf("score %d is out of range [0, 100]", report.Score)
	}
}

func TestSecurityReportScoringMinimum(t *testing.T) {
	// Create many HIGH findings to force score to 0
	findings := make([]SecurityFinding, 15)
	for i := range findings {
		findings[i] = SecurityFinding{Severity: "high", Category: "test", Title: "test"}
	}

	score := computeScore(findings)
	if score != 0 {
		t.Errorf("expected score to be clamped to 0, got %d", score)
	}
}

func TestSecurityCustomCheck(t *testing.T) {
	customCheck := func(h *Har) []SecurityFinding {
		var findings []SecurityFinding
		for i, entry := range h.Log.Entries {
			if strings.HasPrefix(entry.Request.Method, "P") {
				findings = append(findings, SecurityFinding{
					Severity:    "medium",
					Category:    "custom",
					Title:       "POST-like method detected",
					Description: "Entry uses a POST-like method: " + entry.Request.Method,
					EntryIndex:  i,
					EntryURL:    entry.Request.URL,
					Remedy:      "Review POST-like requests for CSRF protection",
				})
			}
		}
		return findings
	}

	har := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						URL:    "https://example.com/api",
						Method: "POST",
					},
					Response: Response{
						Headers: []Headers{},
					},
				},
				{
					Request: Request{
						URL:    "https://example.com/page",
						Method: "GET",
					},
					Response: Response{
						Headers: []Headers{},
					},
				},
			},
		},
	}

	report := har.SecurityAuditWithOptions(SecurityAuditOptions{
		CustomChecks: []SecurityCheckFunc{customCheck},
	})

	customFindings := report.FindByCategory("custom")
	if len(customFindings) != 1 {
		t.Fatalf("expected 1 custom finding, got %d", len(customFindings))
	}
	if customFindings[0].Title != "POST-like method detected" {
		t.Errorf("unexpected custom finding title: %s", customFindings[0].Title)
	}
}

func TestDefaultSecurityAuditOptions(t *testing.T) {
	opts := DefaultSecurityAuditOptions()

	if !opts.CheckSecurityHeaders {
		t.Error("expected CheckSecurityHeaders to be true")
	}
	if !opts.CheckCookies {
		t.Error("expected CheckCookies to be true")
	}
	if !opts.CheckMixedContent {
		t.Error("expected CheckMixedContent to be true")
	}
	if !opts.CheckSensitiveData {
		t.Error("expected CheckSensitiveData to be true")
	}
	if !opts.CheckCORS {
		t.Error("expected CheckCORS to be true")
	}
	if !opts.CheckInfoDisclosure {
		t.Error("expected CheckInfoDisclosure to be true")
	}
}

func TestSecurityReportMethods(t *testing.T) {
	report := &SecurityReport{
		Findings: []SecurityFinding{
			{Severity: "high", Category: "cookies", Title: "A"},
			{Severity: "high", Category: "headers", Title: "B"},
			{Severity: "medium", Category: "cookies", Title: "C"},
			{Severity: "low", Category: "headers", Title: "D"},
			{Severity: "info", Category: "info-disclosure", Title: "E"},
		},
		Score:     73,
		CheckedAt: time.Now(),
	}

	if !report.HasHighSeverity() {
		t.Error("expected HasHighSeverity to return true")
	}
	if !report.HasMediumSeverity() {
		t.Error("expected HasMediumSeverity to return true")
	}

	highFindings := report.FindBySeverity("high")
	if len(highFindings) != 2 {
		t.Errorf("expected 2 high findings, got %d", len(highFindings))
	}

	cookieFindings := report.FindByCategory("cookies")
	if len(cookieFindings) != 2 {
		t.Errorf("expected 2 cookie findings, got %d", len(cookieFindings))
	}

	summary := report.Summary()
	if !strings.Contains(summary, "73/100") {
		t.Errorf("summary should contain score, got: %s", summary)
	}
	if !strings.Contains(summary, "2 high") {
		t.Errorf("summary should contain high count, got: %s", summary)
	}
	if !strings.Contains(summary, "1 medium") {
		t.Errorf("summary should contain medium count, got: %s", summary)
	}
}

func TestSecurityAuditEmptyHAR(t *testing.T) {
	har := &Har{
		Log: Log{
			Entries: []Entries{},
		},
	}

	report := har.SecurityAudit()
	if report.Score != 100 {
		t.Errorf("expected score 100 for empty HAR, got %d", report.Score)
	}
	if len(report.Findings) != 0 {
		t.Errorf("expected 0 findings for empty HAR, got %d", len(report.Findings))
	}
}

func TestSecurityAuditAllPresent(t *testing.T) {
	// HAR with all security headers present, secure cookies, no issues
	har := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						URL: "https://example.com/",
						Cookies: []Cookie{
							{Name: "sessionid", Value: "abc", Secure: true, HTTPOnly: true, SameSite: "Strict", Expires: time.Now().Add(24 * time.Hour)},
						},
					},
					Response: Response{
						Headers: []Headers{
							{Name: "Strict-Transport-Security", Value: "max-age=31536000"},
							{Name: "Content-Security-Policy", Value: "default-src 'self'"},
							{Name: "X-Content-Type-Options", Value: "nosniff"},
							{Name: "X-Frame-Options", Value: "DENY"},
							{Name: "X-XSS-Protection", Value: "1; mode=block"},
							{Name: "Referrer-Policy", Value: "strict-origin-when-cross-origin"},
							{Name: "Permissions-Policy", Value: "camera=()"},
						},
						Cookies: []Cookie{
							{Name: "auth", Value: "xyz", Secure: true, HTTPOnly: true, SameSite: "Lax", Expires: time.Now().Add(24 * time.Hour)},
						},
					},
				},
			},
		},
	}

	report := har.SecurityAudit()
	// Should have no high or medium findings
	if report.HasHighSeverity() {
		t.Error("expected no high severity findings for well-configured HAR")
	}
	if report.HasMediumSeverity() {
		t.Error("expected no medium severity findings for well-configured HAR")
	}
}

// Helper functions

func filterFindings(findings []SecurityFinding, pred func(SecurityFinding) bool) []SecurityFinding {
	var result []SecurityFinding
	for _, f := range findings {
		if pred(f) {
			result = append(result, f)
		}
	}
	return result
}

func anyContains(m map[string]bool, substr string) bool {
	for k := range m {
		if strings.Contains(k, substr) {
			return true
		}
	}
	return false
}
