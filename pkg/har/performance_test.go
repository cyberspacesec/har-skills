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
	e1.SetTimings(50, 100, 200, 10, 2000, 500, 50) // TTFB=2000ms
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
