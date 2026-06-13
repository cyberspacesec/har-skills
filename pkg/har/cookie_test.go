package har

import (
	"strings"
	"testing"
	"time"
)

func createTestHarForCookies() *Har {
	h := NewHar()
	h.SetCreator("test", "1.0")

	// Entry 1: HTTPS page with various cookies
	e1 := h.AddEntry("GET", "https://example.com/page", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e1.SetResponseContent(1024, "text/html")
	// Session cookie without Secure on HTTPS -> HIGH
	e1.AddCookie("sessionid", "abc123")
	// Expired cookie
	e1.AddResponseCookie("expired_cookie", "val")
	h.Log.Entries[0].Response.Cookies[0].Expires = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	// Good cookie with Secure and HttpOnly
	goodCookie := Cookie{
		Name:     "good_cookie",
		Value:    "goodvalue",
		Secure:   true,
		HTTPOnly: true,
		SameSite: "Strict",
		Domain:   "example.com",
	}
	h.Log.Entries[0].Response.Cookies = append(h.Log.Entries[0].Response.Cookies, goodCookie)

	// Entry 2: HTTP page with SameSite=None without Secure
	e2 := h.AddEntry("GET", "http://example.com/other", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")
	e2.SetResponseContent(512, "text/html")
	samesiteNoneCookie := Cookie{
		Name:     "tracking",
		Value:    "track123",
		SameSite: "None",
		Secure:   false,
		Domain:   "ads.other.com",
	}
	e2.Request.Cookies = append(e2.Request.Cookies, samesiteNoneCookie)

	// Entry 3: Oversized cookie
	e3 := h.AddEntry("GET", "https://example.com/big", "HTTP/1.1", "")
	e3.SetResponseStatus(200, "OK")
	bigValue := strings.Repeat("x", 5000)
	bigCookie := Cookie{
		Name:   "big_cookie",
		Value:  bigValue,
		Secure: true,
	}
	e3.Request.Cookies = append(e3.Request.Cookies, bigCookie)

	// Entry 4: Duplicate sessionid cookie (same name as entry 1)
	e4 := h.AddEntry("GET", "https://example.com/another", "HTTP/1.1", "")
	e4.SetResponseStatus(200, "OK")
	e4.AddCookie("sessionid", "different_value")

	return h
}

func TestCookieAudit(t *testing.T) {
	h := createTestHarForCookies()
	report := h.CookieAudit()

	if report == nil {
		t.Fatal("Expected non-nil report")
	}

	if report.TotalCookies == 0 {
		t.Error("Expected TotalCookies > 0")
	}

	if report.UniqueCookies == 0 {
		t.Error("Expected UniqueCookies > 0")
	}

	// Check for expired cookie finding
	expiredFindings := report.FindByCategory("expired")
	if len(expiredFindings) == 0 {
		t.Error("Expected expired cookie findings")
	}

	// Check for session cookie finding
	sessionFindings := report.FindByCategory("session")
	if len(sessionFindings) == 0 {
		t.Error("Expected session cookie findings")
	}

	// Check for insecure finding (sessionid without Secure on HTTPS)
	insecureFindings := report.FindByCategory("insecure")
	if len(insecureFindings) == 0 {
		t.Error("Expected insecure cookie findings")
	}

	// Check for SameSite=None without Secure
	samesiteFindings := report.FindByCategory("samesite")
	if len(samesiteFindings) == 0 {
		t.Error("Expected SameSite findings")
	}

	// Check for third-party cookie
	tpFindings := report.FindByCategory("third-party")
	if len(tpFindings) == 0 {
		t.Error("Expected third-party cookie findings")
	}

	// Check for oversized cookie
	oversizedFindings := report.FindByCategory("oversized")
	if len(oversizedFindings) == 0 {
		t.Error("Expected oversized cookie findings")
	}

	// Check for duplicate cookies
	dupFindings := report.FindByCategory("duplicate")
	if len(dupFindings) == 0 {
		t.Error("Expected duplicate cookie findings")
	}
}

func TestCookieAuditHasHighSeverity(t *testing.T) {
	h := createTestHarForCookies()
	report := h.CookieAudit()

	// We have SameSite=None without Secure and insecure cookie on HTTPS -> HIGH
	if !report.HasHighSeverity() {
		t.Error("Expected HasHighSeverity to be true")
	}
}

func TestCookieAuditNoHighSeverity(t *testing.T) {
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	safeCookie := Cookie{
		Name:     "safe",
		Value:    "val",
		Secure:   true,
		HTTPOnly: true,
		SameSite: "Strict",
		Domain:   "example.com",
	}
	h.Log.Entries[0].Request.Cookies = append(h.Log.Entries[0].Request.Cookies, safeCookie)

	report := h.CookieAudit()
	if report.HasHighSeverity() {
		t.Error("Expected HasHighSeverity to be false for safe cookies")
	}
}

func TestCookieAuditNil(t *testing.T) {
	var h *Har
	report := h.CookieAudit()
	if report == nil {
		t.Error("Expected non-nil report for nil HAR")
		return
	}
	if report.TotalCookies != 0 {
		t.Errorf("Expected 0 total cookies, got %d", report.TotalCookies)
	}
}

func TestCookieAuditEmpty(t *testing.T) {
	h := NewHar()
	report := h.CookieAudit()
	if report.TotalCookies != 0 {
		t.Errorf("Expected 0 total cookies, got %d", report.TotalCookies)
	}
}

func TestCookieFindByCategory(t *testing.T) {
	report := &CookieAuditReport{
		Findings: []CookieFinding{
			{Category: "expired", CookieName: "a"},
			{Category: "expired", CookieName: "b"},
			{Category: "samesite", CookieName: "c"},
		},
	}

	expired := report.FindByCategory("expired")
	if len(expired) != 2 {
		t.Errorf("Expected 2 expired findings, got %d", len(expired))
	}

	samesite := report.FindByCategory("samesite")
	if len(samesite) != 1 {
		t.Errorf("Expected 1 samesite finding, got %d", len(samesite))
	}

	none := report.FindByCategory("nonexistent")
	if len(none) != 0 {
		t.Errorf("Expected 0 findings for nonexistent category, got %d", len(none))
	}
}

func TestCookieEvolution(t *testing.T) {
	h := NewHar()
	h.SetCreator("test", "1.0")

	// Entry 1: initial cookie
	e1 := h.AddEntry("GET", "https://example.com/", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	c1 := Cookie{Name: "session", Value: "v1", Secure: false, HTTPOnly: false, SameSite: "", Domain: "example.com"}
	e1.Request.Cookies = append(e1.Request.Cookies, c1)

	// Entry 2: cookie changes
	e2 := h.AddEntry("GET", "https://example.com/page2", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")
	c2 := Cookie{Name: "session", Value: "v2", Secure: true, HTTPOnly: true, SameSite: "Lax", Domain: "example.com"}
	e2.Request.Cookies = append(e2.Request.Cookies, c2)

	evolution := h.CookieEvolution()

	sessionEvo, ok := evolution["session"]
	if !ok {
		t.Fatal("Expected 'session' in evolution map")
	}

	if len(sessionEvo) != 2 {
		t.Fatalf("Expected 2 evolution entries, got %d", len(sessionEvo))
	}

	// First observation
	if sessionEvo[0].Value != "v1" {
		t.Errorf("Expected first value 'v1', got '%s'", sessionEvo[0].Value)
	}
	if sessionEvo[0].Secure {
		t.Error("Expected first observation Secure=false")
	}

	// Second observation
	if sessionEvo[1].Value != "v2" {
		t.Errorf("Expected second value 'v2', got '%s'", sessionEvo[1].Value)
	}
	if !sessionEvo[1].Secure {
		t.Error("Expected second observation Secure=true")
	}
	if sessionEvo[1].SameSite != "Lax" {
		t.Errorf("Expected SameSite='Lax', got '%s'", sessionEvo[1].SameSite)
	}
}

func TestCookieEvolutionNil(t *testing.T) {
	var h *Har
	evolution := h.CookieEvolution()
	if len(evolution) != 0 {
		t.Error("Expected empty evolution map for nil HAR")
	}
}

func TestCookieEvolutionEmpty(t *testing.T) {
	h := NewHar()
	evolution := h.CookieEvolution()
	if len(evolution) != 0 {
		t.Error("Expected empty evolution map for empty HAR")
	}
}

func TestIsSessionLikeName(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"sessionid", true},
		{"SESSION_ID", true},
		{"sid", true},
		{"JSESSIONID", true},
		{"PHPSESSID", true},
		{"token", true},
		{"auth", true},
		{"theme", false},
		{"lang", false},
		{"", false},
	}

	for _, tt := range tests {
		result := isSessionLikeName(tt.name)
		if result != tt.expected {
			t.Errorf("isSessionLikeName(%q) = %v, expected %v", tt.name, result, tt.expected)
		}
	}
}

func TestIsHTTPS(t *testing.T) {
	if !isHTTPS("https://example.com") {
		t.Error("Expected HTTPS URL to return true")
	}
	if isHTTPS("http://example.com") {
		t.Error("Expected HTTP URL to return false")
	}
	if isHTTPS("not-a-url") {
		t.Error("Expected invalid URL to return false")
	}
}

func TestIsThirdParty(t *testing.T) {
	// Same domain
	if isThirdParty("example.com", "https://example.com/page") {
		t.Error("Expected same domain to not be third-party")
	}

	// Subdomain
	if isThirdParty("example.com", "https://sub.example.com/page") {
		t.Error("Expected subdomain to not be third-party")
	}

	// Different domain
	if !isThirdParty("ads.other.com", "https://example.com/page") {
		t.Error("Expected different domain to be third-party")
	}

	// Domain with leading dot
	if isThirdParty(".example.com", "https://example.com/page") {
		t.Error("Expected .example.com to match example.com")
	}
}

func TestCookieSecureAndHttpOnlyCounts(t *testing.T) {
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")

	c1 := Cookie{Name: "a", Value: "1", Secure: true, HTTPOnly: true, SameSite: "Strict"}
	c2 := Cookie{Name: "b", Value: "2", Secure: true, HTTPOnly: false, SameSite: "Lax"}
	c3 := Cookie{Name: "c", Value: "3", Secure: false, HTTPOnly: true, SameSite: ""}
	h.Log.Entries[0].Request.Cookies = append(h.Log.Entries[0].Request.Cookies, c1, c2, c3)

	report := h.CookieAudit()

	if report.SecureCount != 2 {
		t.Errorf("Expected SecureCount=2, got %d", report.SecureCount)
	}
	if report.HttpOnlyCount != 2 {
		t.Errorf("Expected HttpOnlyCount=2, got %d", report.HttpOnlyCount)
	}
	if report.SameSiteCount != 2 {
		t.Errorf("Expected SameSiteCount=2, got %d", report.SameSiteCount)
	}
	if report.TotalCookies != 3 {
		t.Errorf("Expected TotalCookies=3, got %d", report.TotalCookies)
	}
	if report.UniqueCookies != 3 {
		t.Errorf("Expected UniqueCookies=3, got %d", report.UniqueCookies)
	}
}
