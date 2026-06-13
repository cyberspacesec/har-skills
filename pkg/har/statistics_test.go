package har

import (
	"testing"
	"time"
)

func createTestHarForStats() *Har {
	h := NewHar()
	h.SetCreator("test", "1.0")
	h.SetBrowser("Chrome", "100.0")

	// Entry 1: GET, 200, fast
	e1 := h.AddEntry("GET", "https://example.com/api/users", "HTTP/1.1", "page_1")
	e1.SetResponseStatus(200, "OK")
	e1.SetResponseContent(1024, "application/json")
	e1.SetTimings(10, 5, 15, 2, 50, 30, 8)
	e1.SetServerIP("1.2.3.4")
	e1.StartedDateTime = time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

	// Entry 2: POST, 201, medium
	e2 := h.AddEntry("POST", "https://example.com/api/users", "HTTP/1.1", "page_1")
	e2.SetResponseStatus(201, "Created")
	e2.SetResponseContent(512, "application/json")
	e2.SetTimings(5, 3, 10, 3, 100, 50, 5)
	e2.SetServerIP("1.2.3.4")
	e2.StartedDateTime = time.Date(2024, 1, 1, 10, 0, 1, 0, time.UTC)

	// Entry 3: GET, 404, slow
	e3 := h.AddEntry("GET", "https://other.com/page", "HTTP/1.1", "page_1")
	e3.SetResponseStatus(404, "Not Found")
	e3.SetResponseContent(256, "text/html")
	e3.SetTimings(20, 10, 25, 5, 200, 100, 10)
	e3.StartedDateTime = time.Date(2024, 1, 1, 10, 0, 2, 0, time.UTC)

	// Entry 4: GET, 301 redirect
	e4 := h.AddEntry("GET", "https://example.com/old", "HTTP/1.1", "page_1")
	e4.SetResponseStatus(301, "Moved Permanently")
	e4.SetResponseContent(128, "text/html")
	e4.SetTimings(2, 1, 5, 1, 20, 10, 0)
	e4.StartedDateTime = time.Date(2024, 1, 1, 10, 0, 3, 0, time.UTC)

	// Entry 5: GET, 500 error
	e5 := h.AddEntry("GET", "https://api.example.com/broken", "HTTP/1.1", "")
	e5.SetResponseStatus(500, "Internal Server Error")
	e5.SetResponseContent(64, "application/json")
	e5.SetTimings(15, 8, 20, 4, 300, 150, 12)
	e5.StartedDateTime = time.Date(2024, 1, 1, 10, 0, 4, 0, time.UTC)

	return h
}

func TestStatistics(t *testing.T) {
	h := createTestHarForStats()
	stats := h.Statistics()

	if stats.TotalRequests != 5 {
		t.Errorf("Expected TotalRequests=5, got %d", stats.TotalRequests)
	}

	if stats.ErrorCount != 2 { // 404 + 500
		t.Errorf("Expected ErrorCount=2, got %d", stats.ErrorCount)
	}

	if stats.RedirectCount != 1 { // 301
		t.Errorf("Expected RedirectCount=1, got %d", stats.RedirectCount)
	}

	if stats.Methods["GET"] != 4 {
		t.Errorf("Expected GET count=4, got %d", stats.Methods["GET"])
	}

	if stats.Methods["POST"] != 1 {
		t.Errorf("Expected POST count=1, got %d", stats.Methods["POST"])
	}

	if stats.StatusCodes[200] != 1 {
		t.Errorf("Expected 200 count=1, got %d", stats.StatusCodes[200])
	}

	if stats.StatusCodes[404] != 1 {
		t.Errorf("Expected 404 count=1, got %d", stats.StatusCodes[404])
	}

	if stats.MaxTime <= 0 {
		t.Errorf("Expected MaxTime > 0, got %f", stats.MaxTime)
	}

	if stats.MinTime <= 0 {
		t.Errorf("Expected MinTime > 0, got %f", stats.MinTime)
	}

	if stats.AvgTime <= 0 {
		t.Errorf("Expected AvgTime > 0, got %f", stats.AvgTime)
	}

	if stats.TotalUncompressed <= 0 {
		t.Errorf("Expected TotalUncompressed > 0, got %d", stats.TotalUncompressed)
	}
}

func TestStatisticsNil(t *testing.T) {
	var h *Har
	stats := h.Statistics()
	if stats.TotalRequests != 0 {
		t.Errorf("Expected 0 requests for nil HAR, got %d", stats.TotalRequests)
	}
}

func TestStatisticsEmpty(t *testing.T) {
	h := NewHar()
	stats := h.Statistics()
	if stats.TotalRequests != 0 {
		t.Errorf("Expected 0 requests for empty HAR, got %d", stats.TotalRequests)
	}
}

func TestTimingStatistics(t *testing.T) {
	h := createTestHarForStats()
	ts := h.TimingStatistics()

	if ts.AvgWait <= 0 {
		t.Errorf("Expected AvgWait > 0, got %f", ts.AvgWait)
	}

	if ts.MaxWait <= 0 {
		t.Errorf("Expected MaxWait > 0, got %f", ts.MaxWait)
	}

	if ts.MaxWait < ts.AvgWait {
		t.Errorf("MaxWait (%f) should be >= AvgWait (%f)", ts.MaxWait, ts.AvgWait)
	}
}

func TestDomainSummary(t *testing.T) {
	h := createTestHarForStats()
	ds := h.DomainSummary()

	if len(ds) == 0 {
		t.Error("Expected non-empty domain summary")
	}

	if _, ok := ds["example.com"]; !ok {
		t.Error("Expected 'example.com' in domain summary")
	}

	if ds["example.com"].RequestCount != 3 {
		t.Errorf("Expected 3 requests for example.com, got %d", ds["example.com"].RequestCount)
	}
}

func TestStatusCodeDistribution(t *testing.T) {
	h := createTestHarForStats()
	dist := h.StatusCodeDistribution()

	if dist[200] != 1 {
		t.Errorf("Expected 200 count=1, got %d", dist[200])
	}
	if dist[404] != 1 {
		t.Errorf("Expected 404 count=1, got %d", dist[404])
	}
	if dist[500] != 1 {
		t.Errorf("Expected 500 count=1, got %d", dist[500])
	}
}

func TestMethodDistribution(t *testing.T) {
	h := createTestHarForStats()
	dist := h.MethodDistribution()

	if dist["GET"] != 4 {
		t.Errorf("Expected GET count=4, got %d", dist["GET"])
	}
	if dist["POST"] != 1 {
		t.Errorf("Expected POST count=1, got %d", dist["POST"])
	}
}

func TestContentTypeDistribution(t *testing.T) {
	h := createTestHarForStats()
	dist := h.ContentTypeDistribution()

	if dist["application/json"] != 3 {
		t.Errorf("Expected application/json count=3, got %d", dist["application/json"])
	}
	if dist["text/html"] != 2 {
		t.Errorf("Expected text/html count=2, got %d", dist["text/html"])
	}
}

func TestSlowestRequests(t *testing.T) {
	h := createTestHarForStats()
	slowest := h.SlowestRequests(3)

	if len(slowest) != 3 {
		t.Errorf("Expected 3 slowest, got %d", len(slowest))
	}

	// 最慢的应该在第一个
	if slowest[0].Time < slowest[1].Time {
		t.Errorf("Expected sorted by time descending")
	}
}

func TestFastestRequests(t *testing.T) {
	h := createTestHarForStats()
	fastest := h.FastestRequests(2)

	if len(fastest) != 2 {
		t.Errorf("Expected 2 fastest, got %d", len(fastest))
	}

	if fastest[0].Time > fastest[1].Time {
		t.Errorf("Expected sorted by time ascending")
	}
}

func TestLargestResponses(t *testing.T) {
	h := createTestHarForStats()
	largest := h.LargestResponses(2)

	if len(largest) != 2 {
		t.Errorf("Expected 2 largest, got %d", len(largest))
	}

	if largest[0].Response.Content.Size < largest[1].Response.Content.Size {
		t.Errorf("Expected sorted by size descending")
	}
}

func TestSummary(t *testing.T) {
	h := createTestHarForStats()
	summary := h.Summary()

	if summary == "" {
		t.Error("Expected non-empty summary")
	}

	if !statsContains(summary, "5") {
		t.Error("Summary should contain total request count")
	}
}

func TestPercentile(t *testing.T) {
	tests := []struct {
		values   []float64
		p        int
		expected float64
	}{
		{[]float64{1, 2, 3, 4, 5}, 50, 3},
		{[]float64{1, 2, 3, 4, 5}, 0, 1},
		{[]float64{1, 2, 3, 4, 5}, 100, 5},
		{[]float64{}, 50, 0},
	}

	for _, tt := range tests {
		result := percentile(tt.values, tt.p)
		if len(tt.values) == 0 {
			if result != 0 {
				t.Errorf("percentile(%v, %d) = %f, expected 0", tt.values, tt.p, result)
			}
			continue
		}
		if result != tt.expected {
			t.Errorf("percentile(%v, %d) = %f, expected %f", tt.values, tt.p, result, tt.expected)
		}
	}
}

func statsContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && statsContainsSubstr(s, substr))
}

func statsContainsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
