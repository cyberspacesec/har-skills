package har

import (
	"testing"
)

func createTestHarForPerformance() *Har {
	h := NewHar()
	h.SetCreator("test", "1.0")

	// Fast TTFB, small size, good caching, compressed
	e1 := h.AddEntry("GET", "https://example.com/", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e1.SetResponseContent(5000, "text/html")
	e1.SetTimings(5, 10, 15, 2, 80, 30, 5) // TTFB=80ms
	e1.AddResponseHeader("Cache-Control", "public, max-age=3600")
	e1.AddResponseHeader("Content-Encoding", "gzip")

	// CSS resource, compressed
	e2 := h.AddEntry("GET", "https://example.com/style.css", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")
	e2.SetResponseContent(3000, "text/css")
	e2.SetTimings(1, 2, 3, 1, 40, 20, 2)
	e2.AddResponseHeader("Cache-Control", "public, max-age=86400")
	e2.AddResponseHeader("Content-Encoding", "gzip")

	// JS resource, compressed
	e3 := h.AddEntry("GET", "https://example.com/app.js", "HTTP/1.1", "")
	e3.SetResponseStatus(200, "OK")
	e3.SetResponseContent(8000, "application/javascript")
	e3.SetTimings(1, 2, 3, 1, 60, 25, 2)
	e3.AddResponseHeader("Cache-Control", "public, max-age=3600")
	e3.AddResponseHeader("Content-Encoding", "br")

	// JSON API, not cacheable, not compressed
	e4 := h.AddEntry("GET", "https://example.com/api/data", "HTTP/1.1", "")
	e4.SetResponseStatus(200, "OK")
	e4.SetResponseContent(2000, "application/json")
	e4.SetTimings(1, 2, 3, 1, 100, 15, 2)
	e4.AddResponseHeader("Cache-Control", "no-store")

	return h
}

func createSlowHarForPerformance() *Har {
	h := NewHar()
	h.SetCreator("test", "1.0")

	// Slow TTFB, large resources, no caching, no compression
	// First entry with high TTFB
	e1 := h.AddEntry("GET", "https://slow.example.com/", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e1.SetResponseContent(500000, "text/html")
	e1.SetTimings(50, 100, 200, 10, 2000, 500, 50)        // TTFB=2000ms
	e1.StartedDateTime = e1.StartedDateTime.Add(-5 * 1e9) // 5s ago

	// Many uncompressed resources
	for i := 0; i < 55; i++ {
		e := h.AddEntry("GET", "https://slow.example.com/resource", "HTTP/1.1", "")
		e.SetResponseStatus(200, "OK")
		e.SetResponseContent(100000, "application/javascript")
		e.SetTimings(10, 20, 30, 5, 500, 100, 10)
	}

	return h
}

func TestPerformanceScoreGood(t *testing.T) {
	h := createTestHarForPerformance()
	report := h.PerformanceScore()

	if report == nil {
		t.Fatal("Expected non-nil report")
	}

	// Good HAR should have a reasonable score
	if report.OverallScore < 50 {
		t.Errorf("Expected good HAR to score >= 50, got %.1f", report.OverallScore)
	}

	// Should have categories
	if len(report.Categories) != 6 {
		t.Errorf("Expected 6 categories, got %d", len(report.Categories))
	}
}

func TestPerformanceScoreSlow(t *testing.T) {
	h := createSlowHarForPerformance()
	report := h.PerformanceScore()

	if report.OverallScore > 50 {
		t.Errorf("Expected slow HAR to score < 50, got %.1f", report.OverallScore)
	}

	// Should have recommendations
	if len(report.Recommendations) == 0 {
		t.Error("Expected recommendations for slow HAR")
	}
}

func TestPerformanceGrade(t *testing.T) {
	tests := []struct {
		score    float64
		expected string
	}{
		{95, "A"},
		{90, "A"},
		{85, "B"},
		{70, "B"},
		{60, "C"},
		{50, "C"},
		{30, "D"},
		{0, "D"},
	}

	for _, tt := range tests {
		report := &PerformanceReport{OverallScore: tt.score}
		grade := report.Grade()
		if grade != tt.expected {
			t.Errorf("Grade() for score %.0f = %s, expected %s", tt.score, grade, tt.expected)
		}
	}
}

func TestPerformanceCategoryByName(t *testing.T) {
	report := &PerformanceReport{
		Categories: []PerformanceCategory{
			{Name: "TTFB", Score: 90},
			{Name: "Compression", Score: 50},
		},
	}

	cat := report.CategoryByName("TTFB")
	if cat == nil {
		t.Fatal("Expected to find TTFB category")
	}
	if cat.Score != 90 {
		t.Errorf("Expected score 90, got %.1f", cat.Score)
	}

	notFound := report.CategoryByName("NonExistent")
	if notFound != nil {
		t.Error("Expected nil for nonexistent category")
	}
}

func TestPerformanceReportNilReceiverMethods(t *testing.T) {
	var report *PerformanceReport

	if got := report.CategoryByName("TTFB"); got != nil {
		t.Fatalf("CategoryByName() = %#v, want nil", got)
	}
	if got := report.Grade(); got != "D" {
		t.Fatalf("Grade() = %q, want D", got)
	}
}

func TestPerformanceScoreNil(t *testing.T) {
	var h *Har
	report := h.PerformanceScore()
	if report == nil {
		t.Fatal("Expected non-nil report for nil HAR")
	}
	if report.OverallScore != 100 {
		t.Errorf("Expected 100 for nil HAR, got %.1f", report.OverallScore)
	}
}

func TestPerformanceScoreHelpersNilReceiverDoNotPanic(t *testing.T) {
	var h *Har

	tests := []struct {
		name     string
		score    func() PerformanceCategory
		wantName string
		want     float64
	}{
		{
			name:     "ttfb",
			score:    h.scoreTTFB,
			wantName: "Time to First Byte",
			want:     100,
		},
		{
			name:     "load time",
			score:    h.scoreTotalLoadTime,
			wantName: "Total Load Time",
			want:     100,
		},
		{
			name:     "request count",
			score:    h.scoreRequestCount,
			wantName: "Request Count",
			want:     100,
		},
		{
			name:     "transfer size",
			score:    h.scoreTransferSize,
			wantName: "Transfer Size",
			want:     100,
		},
		{
			name:     "compression",
			score:    h.scoreCompression,
			wantName: "Compression",
			want:     100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cat PerformanceCategory
			assertDoesNotPanic(t, func() {
				cat = tt.score()
			})
			if cat.Name != tt.wantName {
				t.Fatalf("Name = %q, want %q", cat.Name, tt.wantName)
			}
			if cat.Score != tt.want {
				t.Fatalf("Score = %.1f, want %.1f", cat.Score, tt.want)
			}
		})
	}
}

func TestScoreCacheEfficiencyNilReportDoesNotPanic(t *testing.T) {
	var h *Har
	var cat PerformanceCategory

	assertDoesNotPanic(t, func() {
		cat = h.scoreCacheEfficiency(nil)
	})

	if cat.Name != "Cache Efficiency" {
		t.Fatalf("Name = %q, want Cache Efficiency", cat.Name)
	}
	if cat.Score != 0 {
		t.Fatalf("Score = %.1f, want 0.0", cat.Score)
	}
	if len(cat.Findings) != 1 {
		t.Fatalf("Findings length = %d, want 1", len(cat.Findings))
	}
	if cat.Findings[0].Title != "Low cache efficiency" {
		t.Fatalf("Finding title = %q, want Low cache efficiency", cat.Findings[0].Title)
	}
}

func TestPerformanceScoreEmpty(t *testing.T) {
	h := NewHar()
	report := h.PerformanceScore()
	if report == nil {
		t.Fatal("Expected non-nil report for empty HAR")
	}
	if report.OverallScore != 100 {
		t.Errorf("Expected 100 for empty HAR, got %.1f", report.OverallScore)
	}
}

func TestTTFBScoring(t *testing.T) {
	// Very fast TTFB
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.SetTimings(1, 2, 3, 1, 50, 20, 2)
	report := h.PerformanceScore()
	cat := report.CategoryByName("Time to First Byte")
	if cat == nil {
		t.Fatal("Expected TTFB category")
	}
	if cat.Score != 100 {
		t.Errorf("Expected score 100 for TTFB < 200ms, got %.1f", cat.Score)
	}
}

func TestCompressionScoring(t *testing.T) {
	// All text resources compressed
	h := NewHar()
	e1 := h.AddEntry("GET", "https://example.com/app.js", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e1.SetResponseContent(1000, "application/javascript")
	e1.AddResponseHeader("Content-Encoding", "gzip")

	e2 := h.AddEntry("GET", "https://example.com/style.css", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")
	e2.SetResponseContent(1000, "text/css")
	e2.AddResponseHeader("Content-Encoding", "br")

	report := h.PerformanceScore()
	cat := report.CategoryByName("Compression")
	if cat == nil {
		t.Fatal("Expected Compression category")
	}
	if cat.Score != 100 {
		t.Errorf("Expected score 100 for all compressed, got %.1f", cat.Score)
	}
	if len(cat.Findings) != 0 {
		t.Errorf("Expected 0 findings for all compressed, got %d", len(cat.Findings))
	}
}

func TestCompressionNoCompression(t *testing.T) {
	// No text resources compressed
	h := NewHar()
	e1 := h.AddEntry("GET", "https://example.com/app.js", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e1.SetResponseContent(1000, "application/javascript")

	e2 := h.AddEntry("GET", "https://example.com/style.css", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")
	e2.SetResponseContent(1000, "text/css")

	report := h.PerformanceScore()
	cat := report.CategoryByName("Compression")
	if cat == nil {
		t.Fatal("Expected Compression category")
	}
	if cat.Score != 0 {
		t.Errorf("Expected score 0 for no compression, got %.1f", cat.Score)
	}
	if len(cat.Findings) != 2 {
		t.Errorf("Expected 2 findings, got %d", len(cat.Findings))
	}
}

func TestCacheEfficiencyScoring(t *testing.T) {
	// All cacheable
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/static.js", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.SetResponseContent(1000, "application/javascript")
	e.AddResponseHeader("Cache-Control", "public, max-age=3600")

	report := h.PerformanceScore()
	cat := report.CategoryByName("Cache Efficiency")
	if cat == nil {
		t.Fatal("Expected Cache Efficiency category")
	}
	if cat.Score != 100 {
		t.Errorf("Expected score 100 for all cacheable, got %.1f", cat.Score)
	}
}

func TestRecommendations(t *testing.T) {
	h := createSlowHarForPerformance()
	report := h.PerformanceScore()

	// Should generate at least some recommendations
	if len(report.Recommendations) == 0 {
		t.Error("Expected recommendations for slow HAR")
	}

	// Check for expected recommendation content
	foundCompression := false
	foundRequests := false
	foundTTFB := false

	for _, rec := range report.Recommendations {
		if rec == "Enable compression for text resources" {
			foundCompression = true
		}
		if rec == "Reduce the number of HTTP requests" {
			foundRequests = true
		}
		if rec == "Optimize server response time" {
			foundTTFB = true
		}
	}

	if !foundCompression {
		t.Error("Expected 'Enable compression for text resources' recommendation")
	}
	if !foundRequests {
		t.Error("Expected 'Reduce the number of HTTP requests' recommendation")
	}
	if !foundTTFB {
		t.Error("Expected 'Optimize server response time' recommendation")
	}
}

func TestRequestCountScoring(t *testing.T) {
	// Few requests
	h := NewHar()
	for i := 0; i < 5; i++ {
		e := h.AddEntry("GET", "https://example.com/page", "HTTP/1.1", "")
		e.SetResponseStatus(200, "OK")
		e.SetResponseContent(100, "text/html")
	}

	report := h.PerformanceScore()
	cat := report.CategoryByName("Request Count")
	if cat == nil {
		t.Fatal("Expected Request Count category")
	}
	if cat.Score != 100 {
		t.Errorf("Expected score 100 for <10 requests, got %.1f", cat.Score)
	}
}

func TestTransferSizeScoring(t *testing.T) {
	// Small transfer size
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.SetResponseContent(100, "text/html")

	report := h.PerformanceScore()
	cat := report.CategoryByName("Transfer Size")
	if cat == nil {
		t.Fatal("Expected Transfer Size category")
	}
	if cat.Score != 100 {
		t.Errorf("Expected score 100 for small transfer, got %.1f", cat.Score)
	}
}

func TestScoreTTFBNegativeWaitTime(t *testing.T) {
	// When Timings.Wait is negative, it should fall back to entry.Time
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.SetResponseContent(100, "text/html")
	e.Timings.Wait = -1 // Negative wait time
	e.Time = 50         // Fallback value (under 200ms)

	cat := h.scoreTTFB()
	if cat.Score != 100 {
		t.Errorf("Expected score 100 for TTFB < 200ms (fallback to entry.Time), got %.1f", cat.Score)
	}
}

func TestScoreTTFBBetween200And500(t *testing.T) {
	// TTFB between 200ms and 500ms should score 80
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.SetResponseContent(100, "text/html")
	e.Timings.Wait = 300

	cat := h.scoreTTFB()
	if cat.Score != 80 {
		t.Errorf("Expected score 80 for TTFB 300ms, got %.1f", cat.Score)
	}
}

func TestScoreTTFBBetween500And1000(t *testing.T) {
	// TTFB between 500ms and 1000ms should score 50
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.SetResponseContent(100, "text/html")
	e.Timings.Wait = 700

	cat := h.scoreTTFB()
	if cat.Score != 50 {
		t.Errorf("Expected score 50 for TTFB 700ms, got %.1f", cat.Score)
	}
}

func TestScoreTTFBAbove1000(t *testing.T) {
	// TTFB >= 1000ms should score 20 and include a finding
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.SetResponseContent(100, "text/html")
	e.Timings.Wait = 1500

	cat := h.scoreTTFB()
	if cat.Score != 20 {
		t.Errorf("Expected score 20 for TTFB 1500ms, got %.1f", cat.Score)
	}
	if len(cat.Findings) != 1 {
		t.Fatalf("Expected 1 finding for slow TTFB, got %d", len(cat.Findings))
	}
	if cat.Findings[0].Title != "Slow server response time" {
		t.Errorf("Expected 'Slow server response time' finding, got %s", cat.Findings[0].Title)
	}
}

func TestScoreTTFBZeroOrNegative(t *testing.T) {
	// TTFB <= 0 should score 100
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.SetResponseContent(100, "text/html")
	e.Timings.Wait = 0

	cat := h.scoreTTFB()
	if cat.Score != 100 {
		t.Errorf("Expected score 100 for TTFB 0ms, got %.1f", cat.Score)
	}
}

func TestScoreTotalLoadTimeBetween1And3Sec(t *testing.T) {
	// Total load time between 1s and 3s should score 80
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.SetResponseContent(100, "text/html")
	e.Timings.Wait = 100
	e.Time = 2000 // 2s total time

	cat := h.scoreTotalLoadTime()
	if cat.Score != 80 {
		t.Errorf("Expected score 80 for total load time 2000ms, got %.1f", cat.Score)
	}
}

func TestScoreTotalLoadTimeBetween3And5Sec(t *testing.T) {
	// Total load time between 3s and 5s should score 50
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.SetResponseContent(100, "text/html")
	e.Timings.Wait = 100
	e.Time = 4000 // 4s total time

	cat := h.scoreTotalLoadTime()
	if cat.Score != 50 {
		t.Errorf("Expected score 50 for total load time 4000ms, got %.1f", cat.Score)
	}
}

func TestScoreTotalLoadTimeAbove5Sec(t *testing.T) {
	// Total load time >= 5s should score 20 and include a finding
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.SetResponseContent(100, "text/html")
	e.Timings.Wait = 100
	e.Time = 7000 // 7s total time

	cat := h.scoreTotalLoadTime()
	if cat.Score != 20 {
		t.Errorf("Expected score 20 for total load time 7000ms, got %.1f", cat.Score)
	}
	if len(cat.Findings) != 1 {
		t.Fatalf("Expected 1 finding for high total load time, got %d", len(cat.Findings))
	}
	if cat.Findings[0].Title != "High total load time" {
		t.Errorf("Expected 'High total load time' finding, got %s", cat.Findings[0].Title)
	}
}

func TestScoreTotalLoadTimeZeroOrNegative(t *testing.T) {
	// Total load time <= 0 should score 100
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.SetResponseContent(100, "text/html")
	e.Timings.Wait = 0
	e.Time = 0

	cat := h.scoreTotalLoadTime()
	if cat.Score != 100 {
		t.Errorf("Expected score 100 for zero total load time, got %.1f", cat.Score)
	}
}

func TestScoreRequestCountBetween10And30(t *testing.T) {
	// Request count between 10 and 30 should score 80
	h := NewHar()
	for i := 0; i < 20; i++ {
		e := h.AddEntry("GET", "https://example.com/page", "HTTP/1.1", "")
		e.SetResponseStatus(200, "OK")
		e.SetResponseContent(100, "text/html")
	}

	cat := h.scoreRequestCount()
	if cat.Score != 80 {
		t.Errorf("Expected score 80 for 20 requests, got %.1f", cat.Score)
	}
}

func TestScoreRequestCountBetween30And50(t *testing.T) {
	// Request count between 30 and 50 should score 50
	h := NewHar()
	for i := 0; i < 40; i++ {
		e := h.AddEntry("GET", "https://example.com/page", "HTTP/1.1", "")
		e.SetResponseStatus(200, "OK")
		e.SetResponseContent(100, "text/html")
	}

	cat := h.scoreRequestCount()
	if cat.Score != 50 {
		t.Errorf("Expected score 50 for 40 requests, got %.1f", cat.Score)
	}
}

func TestScoreRequestCountAbove50(t *testing.T) {
	// Request count >= 50 should score 20 and include a finding
	h := NewHar()
	for i := 0; i < 55; i++ {
		e := h.AddEntry("GET", "https://example.com/page", "HTTP/1.1", "")
		e.SetResponseStatus(200, "OK")
		e.SetResponseContent(100, "text/html")
	}

	cat := h.scoreRequestCount()
	if cat.Score != 20 {
		t.Errorf("Expected score 20 for 55 requests, got %.1f", cat.Score)
	}
	if len(cat.Findings) != 1 {
		t.Fatalf("Expected 1 finding for too many requests, got %d", len(cat.Findings))
	}
	if cat.Findings[0].Title != "Too many HTTP requests" {
		t.Errorf("Expected 'Too many HTTP requests' finding, got %s", cat.Findings[0].Title)
	}
}

func TestScoreTransferSizeBetween500KBAnd1MB(t *testing.T) {
	// Transfer size between 500KB and 1MB should score 80
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	// 700KB in body size
	e.Response.BodySize = 700 * 1024
	e.Response.Content.Size = 700 * 1024
	e.Response.Content.MimeType = "text/html"

	cat := h.scoreTransferSize()
	if cat.Score != 80 {
		t.Errorf("Expected score 80 for ~700KB transfer, got %.1f", cat.Score)
	}
}

func TestScoreTransferSizeBetween1MBAnd3MB(t *testing.T) {
	// Transfer size between 1MB and 3MB should score 50
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	// 2MB in body size
	e.Response.BodySize = 2 * 1024 * 1024
	e.Response.Content.Size = 2 * 1024 * 1024
	e.Response.Content.MimeType = "text/html"

	cat := h.scoreTransferSize()
	if cat.Score != 50 {
		t.Errorf("Expected score 50 for ~2MB transfer, got %.1f", cat.Score)
	}
}

func TestScoreTransferSizeAbove3MB(t *testing.T) {
	// Transfer size >= 3MB should score 20 and include a finding
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	// 4MB in body size
	e.Response.BodySize = 4 * 1024 * 1024
	e.Response.Content.Size = 4 * 1024 * 1024
	e.Response.Content.MimeType = "text/html"

	cat := h.scoreTransferSize()
	if cat.Score != 20 {
		t.Errorf("Expected score 20 for ~4MB transfer, got %.1f", cat.Score)
	}
	if len(cat.Findings) != 1 {
		t.Fatalf("Expected 1 finding for large transfer size, got %d", len(cat.Findings))
	}
	if cat.Findings[0].Title != "Large total transfer size" {
		t.Errorf("Expected 'Large total transfer size' finding, got %s", cat.Findings[0].Title)
	}
}

func TestScoreCacheEfficiencyLow(t *testing.T) {
	// Cache efficiency < 50% should include a finding
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.SetResponseContent(100, "text/html")
	e.AddResponseHeader("Cache-Control", "no-store")

	cacheReport := h.CacheAnalysis()
	cat := h.scoreCacheEfficiency(cacheReport)
	if cat.Score >= 50 {
		t.Errorf("Expected score < 50 for non-cacheable resources, got %.1f", cat.Score)
	}
	if len(cat.Findings) == 0 {
		t.Error("Expected findings for low cache efficiency")
	}
	found := false
	for _, f := range cat.Findings {
		if f.Title == "Low cache efficiency" {
			found = true
		}
	}
	if !found {
		t.Error("Expected 'Low cache efficiency' finding")
	}
}

func TestScoreCacheEfficiencyHigh(t *testing.T) {
	// Cache efficiency >= 50% should NOT include a finding
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.SetResponseContent(100, "text/html")
	e.AddResponseHeader("Cache-Control", "public, max-age=3600")

	cacheReport := h.CacheAnalysis()
	cat := h.scoreCacheEfficiency(cacheReport)
	if cat.Score < 50 {
		t.Errorf("Expected score >= 50 for cacheable resources, got %.1f", cat.Score)
	}
	if len(cat.Findings) != 0 {
		t.Errorf("Expected no findings for high cache efficiency, got %d", len(cat.Findings))
	}
}

func TestGenerateRecommendationsNoCompressionRec(t *testing.T) {
	// When uncompressedCount <= 3, the compression recommendation should NOT be generated
	// Create a HAR where only 2 text resources are uncompressed
	h := NewHar()
	e1 := h.AddEntry("GET", "https://example.com/", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e1.SetResponseContent(100, "text/html")
	e1.AddResponseHeader("Content-Encoding", "gzip")
	e1.AddResponseHeader("Cache-Control", "public, max-age=3600")

	e2 := h.AddEntry("GET", "https://example.com/app.js", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")
	e2.SetResponseContent(100, "application/javascript")
	// Not compressed, not cacheable

	report := h.PerformanceScore()

	for _, rec := range report.Recommendations {
		if rec == "Enable compression for text resources" {
			t.Error("Should not generate compression recommendation when only 2 resources are uncompressed")
		}
	}
}

func TestGenerateRecommendationsCacheHeaderNotNeeded(t *testing.T) {
	// When cacheCat.Score >= 60, the cache header recommendation should NOT be generated
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.SetResponseContent(100, "text/html")
	e.AddResponseHeader("Content-Encoding", "gzip")
	e.AddResponseHeader("Cache-Control", "public, max-age=3600")

	report := h.PerformanceScore()

	for _, rec := range report.Recommendations {
		if rec == "Add proper cache headers" {
			t.Error("Should not generate cache header recommendation when cache score is >= 60")
		}
	}
}

func TestGenerateRecommendationsOptimizeSizeNotNeeded(t *testing.T) {
	// When transferSizeCat.Score > 50, the size optimization recommendation should NOT be generated
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.SetResponseContent(100, "text/html")
	e.AddResponseHeader("Content-Encoding", "gzip")
	e.AddResponseHeader("Cache-Control", "public, max-age=3600")

	report := h.PerformanceScore()

	for _, rec := range report.Recommendations {
		if rec == "Optimize resource sizes" {
			t.Error("Should not generate size optimization recommendation when transfer size score > 50")
		}
	}
}

func TestGenerateRecommendationsTTFBNotNeeded(t *testing.T) {
	// When ttfbCat.Score > 50, the TTFB optimization recommendation should NOT be generated
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.SetResponseContent(100, "text/html")
	e.SetTimings(1, 2, 3, 1, 80, 20, 2) // Fast TTFB
	e.AddResponseHeader("Content-Encoding", "gzip")
	e.AddResponseHeader("Cache-Control", "public, max-age=3600")

	report := h.PerformanceScore()

	for _, rec := range report.Recommendations {
		if rec == "Optimize server response time" {
			t.Error("Should not generate TTFB optimization recommendation when TTFB score > 50")
		}
	}
}
