package har

import (
	"testing"
	"time"
)

func createTestHarForCache() *Har {
	h := NewHar()
	h.SetCreator("test", "1.0")

	// Entry 1: Cacheable with max-age
	e1 := h.AddEntry("GET", "https://example.com/static.js", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e1.SetResponseContent(1024, "application/javascript")
	e1.AddResponseHeader("Cache-Control", "public, max-age=3600")
	e1.AddResponseHeader("ETag", "\"abc123\"")

	// Entry 2: No-store
	e2 := h.AddEntry("GET", "https://example.com/api/data", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")
	e2.SetResponseContent(512, "application/json")
	e2.AddResponseHeader("Cache-Control", "no-store")

	// Entry 3: Private with max-age
	e3 := h.AddEntry("GET", "https://example.com/user/profile", "HTTP/1.1", "")
	e3.SetResponseStatus(200, "OK")
	e3.SetResponseContent(256, "text/html")
	e3.AddResponseHeader("Cache-Control", "private, max-age=600")
	e3.AddResponseHeader("Vary", "Accept-Encoding, Cookie")
	e3.AddResponseHeader("Last-Modified", "Mon, 01 Jan 2024 00:00:00 GMT")

	// Entry 4: No-cache
	e4 := h.AddEntry("GET", "https://example.com/api/refresh", "HTTP/1.1", "")
	e4.SetResponseStatus(200, "OK")
	e4.SetResponseContent(128, "application/json")
	e4.AddResponseHeader("Cache-Control", "no-cache")

	// Entry 5: Cacheable with Age header
	e5 := h.AddEntry("GET", "https://example.com/img/logo.png", "HTTP/1.1", "")
	e5.SetResponseStatus(200, "OK")
	e5.SetResponseContent(2048, "image/png")
	e5.AddResponseHeader("Cache-Control", "public, max-age=86400")
	e5.AddResponseHeader("Age", "1200")

	// Entry 6: No cache headers at all (default cacheable)
	e6 := h.AddEntry("GET", "https://example.com/style.css", "HTTP/1.1", "")
	e6.SetResponseStatus(200, "OK")
	e6.SetResponseContent(512, "text/css")

	// Entry 7: Pragma: no-cache
	e7 := h.AddEntry("GET", "https://example.com/legacy", "HTTP/1.1", "")
	e7.SetResponseStatus(200, "OK")
	e7.SetResponseContent(64, "text/html")
	e7.AddResponseHeader("Pragma", "no-cache")

	return h
}

func TestCacheAnalysis(t *testing.T) {
	h := createTestHarForCache()
	report := h.CacheAnalysis()

	if report == nil {
		t.Fatal("Expected non-nil report")
	}

	if len(report.Assessments) != 7 {
		t.Errorf("Expected 7 assessments, got %d", len(report.Assessments))
	}

	// Entry 0: public, max-age=3600 -> cacheable
	a0 := report.Assessments[0]
	if !a0.Cacheable {
		t.Error("Expected entry 0 to be cacheable")
	}
	if a0.CacheType != "public" {
		t.Errorf("Expected cache type 'public', got '%s'", a0.CacheType)
	}
	if a0.MaxAge != 3600*time.Second {
		t.Errorf("Expected max-age 3600s, got %v", a0.MaxAge)
	}
	if !a0.HasETag {
		t.Error("Expected HasETag=true")
	}

	// Entry 1: no-store -> not cacheable
	a1 := report.Assessments[1]
	if a1.Cacheable {
		t.Error("Expected entry 1 to not be cacheable")
	}
	if a1.CacheType != "no-store" {
		t.Errorf("Expected cache type 'no-store', got '%s'", a1.CacheType)
	}

	// Entry 2: private, max-age=600 -> cacheable (private)
	a2 := report.Assessments[2]
	if !a2.Cacheable {
		t.Error("Expected entry 2 to be cacheable (private)")
	}
	if a2.CacheType != "private" {
		t.Errorf("Expected cache type 'private', got '%s'", a2.CacheType)
	}
	if !a2.HasVary {
		t.Error("Expected HasVary=true")
	}
	if len(a2.VaryHeaders) != 2 {
		t.Errorf("Expected 2 Vary headers, got %d", len(a2.VaryHeaders))
	}
	if !a2.HasLastModified {
		t.Error("Expected HasLastModified=true")
	}

	// Entry 3: no-cache -> not cacheable
	a3 := report.Assessments[3]
	if a3.Cacheable {
		t.Error("Expected entry 3 to not be cacheable")
	}
	if a3.CacheType != "no-cache" {
		t.Errorf("Expected cache type 'no-cache', got '%s'", a3.CacheType)
	}

	// Entry 4: public, max-age=86400, Age=1200
	a4 := report.Assessments[4]
	if !a4.Cacheable {
		t.Error("Expected entry 4 to be cacheable")
	}
	if a4.Age != 1200*time.Second {
		t.Errorf("Expected Age=1200s, got %v", a4.Age)
	}

	// Entry 5: no cache headers -> default cacheable
	a5 := report.Assessments[5]
	if !a5.Cacheable {
		t.Error("Expected entry 5 to be cacheable (default)")
	}

	// Entry 6: Pragma: no-cache -> not cacheable
	a6 := report.Assessments[6]
	if a6.Cacheable {
		t.Error("Expected entry 6 to not be cacheable (Pragma: no-cache)")
	}
	if a6.CacheType != "no-cache" {
		t.Errorf("Expected cache type 'no-cache', got '%s'", a6.CacheType)
	}
}

func TestCacheEfficiency(t *testing.T) {
	h := createTestHarForCache()
	report := h.CacheAnalysis()

	// 7 entries: 3 cacheable (0, 2, 4, 5) = 4, 3 non-cacheable (1, 3, 6)
	if report.CacheableCount != 4 {
		t.Errorf("Expected CacheableCount=4, got %d", report.CacheableCount)
	}
	if report.NonCacheableCount != 3 {
		t.Errorf("Expected NonCacheableCount=3, got %d", report.NonCacheableCount)
	}

	expectedEfficiency := float64(4) / float64(7) * 100
	if report.CacheEfficiency < expectedEfficiency-0.01 || report.CacheEfficiency > expectedEfficiency+0.01 {
		t.Errorf("Expected CacheEfficiency=%.2f, got %.2f", expectedEfficiency, report.CacheEfficiency)
	}
}

func TestCacheReportFindByURL(t *testing.T) {
	h := createTestHarForCache()
	report := h.CacheAnalysis()

	found := report.FindByURL("https://example.com/static.js")
	if found == nil {
		t.Error("Expected to find assessment for static.js")
	}
	if found != nil && !found.HasETag {
		t.Error("Expected HasETag=true for static.js")
	}

	notFound := report.FindByURL("https://example.com/nonexistent")
	if notFound != nil {
		t.Error("Expected nil for nonexistent URL")
	}
}

func TestCacheReportNonCacheableEntries(t *testing.T) {
	h := createTestHarForCache()
	report := h.CacheAnalysis()

	nonCacheable := report.NonCacheableEntries()
	if len(nonCacheable) != 3 {
		t.Errorf("Expected 3 non-cacheable entries, got %d", len(nonCacheable))
	}

	for _, nc := range nonCacheable {
		if nc.Cacheable {
			t.Error("Non-cacheable entry should not have Cacheable=true")
		}
	}
}

func TestCacheAnalysisNil(t *testing.T) {
	var h *Har
	report := h.CacheAnalysis()
	if report == nil {
		t.Error("Expected non-nil report for nil HAR")
	}
	if len(report.Assessments) != 0 {
		t.Errorf("Expected 0 assessments for nil HAR, got %d", len(report.Assessments))
	}
}

func TestCacheAnalysisEmpty(t *testing.T) {
	h := NewHar()
	report := h.CacheAnalysis()
	if len(report.Assessments) != 0 {
		t.Errorf("Expected 0 assessments for empty HAR, got %d", len(report.Assessments))
	}
}

func TestParseCacheControl(t *testing.T) {
	tests := []struct {
		value    string
		expected CacheControlDirectives
	}{
		{
			value: "public, max-age=3600",
			expected: CacheControlDirectives{
				Public: true,
				MaxAge: intPtr(3600),
			},
		},
		{
			value: "no-store",
			expected: CacheControlDirectives{
				NoStore: true,
			},
		},
		{
			value: "private, no-cache, must-revalidate",
			expected: CacheControlDirectives{
				Private:        true,
				NoCache:        true,
				MustRevalidate: true,
			},
		},
		{
			value: "public, s-maxage=600, stale-while-revalidate=30, stale-if-error=60",
			expected: CacheControlDirectives{
				Public:               true,
				SMaxAge:              intPtr(600),
				StaleWhileRevalidate: intPtr(30),
				StaleIfError:         intPtr(60),
			},
		},
		{
			value: "no-cache, proxy-revalidate",
			expected: CacheControlDirectives{
				NoCache:         true,
				ProxyRevalidate: true,
			},
		},
		{
			value: "",
			expected: CacheControlDirectives{},
		},
	}

	for _, tt := range tests {
		result := ParseCacheControl(tt.value)

		if result.NoCache != tt.expected.NoCache {
			t.Errorf("ParseCacheControl(%q): NoCache = %v, expected %v", tt.value, result.NoCache, tt.expected.NoCache)
		}
		if result.NoStore != tt.expected.NoStore {
			t.Errorf("ParseCacheControl(%q): NoStore = %v, expected %v", tt.value, result.NoStore, tt.expected.NoStore)
		}
		if result.Private != tt.expected.Private {
			t.Errorf("ParseCacheControl(%q): Private = %v, expected %v", tt.value, result.Private, tt.expected.Private)
		}
		if result.Public != tt.expected.Public {
			t.Errorf("ParseCacheControl(%q): Public = %v, expected %v", tt.value, result.Public, tt.expected.Public)
		}
		if result.MustRevalidate != tt.expected.MustRevalidate {
			t.Errorf("ParseCacheControl(%q): MustRevalidate = %v, expected %v", tt.value, result.MustRevalidate, tt.expected.MustRevalidate)
		}
		if result.ProxyRevalidate != tt.expected.ProxyRevalidate {
			t.Errorf("ParseCacheControl(%q): ProxyRevalidate = %v, expected %v", tt.value, result.ProxyRevalidate, tt.expected.ProxyRevalidate)
		}

		// Compare pointer fields
		if !intPtrEqual(result.MaxAge, tt.expected.MaxAge) {
			t.Errorf("ParseCacheControl(%q): MaxAge = %v, expected %v", tt.value, result.MaxAge, tt.expected.MaxAge)
		}
		if !intPtrEqual(result.SMaxAge, tt.expected.SMaxAge) {
			t.Errorf("ParseCacheControl(%q): SMaxAge = %v, expected %v", tt.value, result.SMaxAge, tt.expected.SMaxAge)
		}
		if !intPtrEqual(result.StaleWhileRevalidate, tt.expected.StaleWhileRevalidate) {
			t.Errorf("ParseCacheControl(%q): StaleWhileRevalidate = %v, expected %v", tt.value, result.StaleWhileRevalidate, tt.expected.StaleWhileRevalidate)
		}
		if !intPtrEqual(result.StaleIfError, tt.expected.StaleIfError) {
			t.Errorf("ParseCacheControl(%q): StaleIfError = %v, expected %v", tt.value, result.StaleIfError, tt.expected.StaleIfError)
		}
	}
}

func intPtrEqual(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func TestParseVary(t *testing.T) {
	result := parseVary("Accept-Encoding, Cookie")
	if len(result) != 2 {
		t.Fatalf("Expected 2 vary headers, got %d", len(result))
	}
	if result[0] != "Accept-Encoding" {
		t.Errorf("Expected 'Accept-Encoding', got '%s'", result[0])
	}
	if result[1] != "Cookie" {
		t.Errorf("Expected 'Cookie', got '%s'", result[1])
	}

	// Single value
	single := parseVary("Accept")
	if len(single) != 1 || single[0] != "Accept" {
		t.Errorf("Expected ['Accept'], got %v", single)
	}

	// Empty
	empty := parseVary("")
	if len(empty) != 0 {
		t.Errorf("Expected empty, got %v", empty)
	}
}

func TestCacheWithExpiresHeader(t *testing.T) {
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/page", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.AddResponseHeader("Expires", "Mon, 01 Jan 2030 00:00:00 GMT")

	report := h.CacheAnalysis()
	if len(report.Assessments) != 1 {
		t.Fatalf("Expected 1 assessment, got %d", len(report.Assessments))
	}

	// With Expires header and no Cache-Control, should be cacheable
	if !report.Assessments[0].Cacheable {
		t.Error("Expected entry with Expires header to be cacheable")
	}
}
