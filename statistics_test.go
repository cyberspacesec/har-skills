package har

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

func TestSummaryNilHar(t *testing.T) {
	var h *Har
	summary := h.Summary()

	assert.Contains(t, summary, "HAR 文件摘要")
	assert.Contains(t, summary, "总请求数: 0")
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

// --- DomainSummary uncovered branches ---

func TestDomainSummaryNilHar(t *testing.T) {
	var h *Har
	ds := h.DomainSummary()
	assert.NotNil(t, ds)
	assert.Equal(t, 0, len(ds))
}

func TestDomainSummaryEmptyDomain(t *testing.T) {
	// Cover the case where extractDomain returns ""
	h := NewHar()
	e := h.AddEntry("GET", "not-a-valid-url", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.SetResponseContent(100, "text/html")
	ds := h.DomainSummary()
	// The entry has an invalid URL so extractDomain should return ""
	_, hasExample := ds[""]
	assert.False(t, hasExample, "empty domain should be skipped")
}

func TestDomainSummaryWithBodySizeZero(t *testing.T) {
	// Cover the entry.Response.BodySize <= 0 branch (default is -1 from NewHar)
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/page", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.SetResponseContent(100, "text/html")
	// BodySize defaults to -1 from AddEntry, which should NOT be counted
	ds := h.DomainSummary()
	assert.NotNil(t, ds)
	if d, ok := ds["example.com"]; ok {
		assert.Equal(t, int64(0), d.TotalTransferred, "BodySize <= 0 should not be counted")
	}
}

func TestDomainSummaryWithPositiveBodySize(t *testing.T) {
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/page", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.SetResponseContent(100, "text/html")
	e.Response.BodySize = 500
	ds := h.DomainSummary()
	if d, ok := ds["example.com"]; ok {
		assert.Equal(t, int64(500), d.TotalTransferred)
	}
}

func TestDomainSummaryMultipleEntriesSameDomain(t *testing.T) {
	h := NewHar()
	e1 := h.AddEntry("GET", "https://example.com/a", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e1.Response.BodySize = 100
	e1.Time = 100

	e2 := h.AddEntry("GET", "https://example.com/b", "HTTP/1.1", "")
	e2.SetResponseStatus(500, "Error")
	e2.Response.BodySize = 200
	e2.Time = 200

	ds := h.DomainSummary()
	d := ds["example.com"]
	assert.Equal(t, 2, d.RequestCount)
	assert.Equal(t, int64(300), d.TotalTransferred)
	assert.Equal(t, 1, d.ErrorCount)
	assert.Greater(t, d.AvgTime, 0.0)
}

// --- StatusCodeDistribution uncovered branches ---

func TestStatusCodeDistributionNilHar(t *testing.T) {
	var h *Har
	dist := h.StatusCodeDistribution()
	assert.NotNil(t, dist)
	assert.Equal(t, 0, len(dist))
}

// --- MethodDistribution uncovered branches ---

func TestMethodDistributionNilHar(t *testing.T) {
	var h *Har
	dist := h.MethodDistribution()
	assert.NotNil(t, dist)
	assert.Equal(t, 0, len(dist))
}

// --- ContentTypeDistribution uncovered branches ---

func TestContentTypeDistributionNilHar(t *testing.T) {
	var h *Har
	dist := h.ContentTypeDistribution()
	assert.NotNil(t, dist)
	assert.Equal(t, 0, len(dist))
}

func TestContentTypeDistributionEmptyMimeType(t *testing.T) {
	// Cover the branch where contentType == "" (empty MimeType)
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/page", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	// Deliberately don't set content type -> MimeType is ""
	dist := h.ContentTypeDistribution()
	assert.NotNil(t, dist)
	_, hasEmpty := dist[""]
	assert.False(t, hasEmpty, "empty content type should be skipped")
}

func TestContentTypeDistributionWithCharset(t *testing.T) {
	// Cover the branch where content type has "; charset=utf-8" suffix
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/page", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.Response.Content.MimeType = "text/html; charset=utf-8"
	dist := h.ContentTypeDistribution()
	assert.Equal(t, 1, dist["text/html"])
	_, hasOriginal := dist["text/html; charset=utf-8"]
	assert.False(t, hasOriginal, "charset should be stripped")
}

func TestContentTypeDistributionSemicolonNoCharset(t *testing.T) {
	// Cover the branch with semicolon but not charset
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.Response.Content.MimeType = "application/json; boundary=something"
	dist := h.ContentTypeDistribution()
	assert.Equal(t, 1, dist["application/json"])
}

// --- SlowestRequests uncovered branches ---

func TestSlowestRequestsNilHar(t *testing.T) {
	var h *Har
	result := h.SlowestRequests(5)
	assert.Nil(t, result)
}

func TestSlowestRequestsZeroN(t *testing.T) {
	h := createTestHarForStats()
	result := h.SlowestRequests(0)
	assert.Nil(t, result)
}

func TestSlowestRequestsNegativeN(t *testing.T) {
	h := createTestHarForStats()
	result := h.SlowestRequests(-1)
	assert.Nil(t, result)
}

func TestSlowestRequestsNLargerThanEntries(t *testing.T) {
	// Cover the branch where n > len(entries)
	h := createTestHarForStats()
	result := h.SlowestRequests(100)
	assert.Equal(t, 5, len(result))
}

// --- FastestRequests uncovered branches ---

func TestFastestRequestsNilHar(t *testing.T) {
	var h *Har
	result := h.FastestRequests(5)
	assert.Nil(t, result)
}

func TestFastestRequestsZeroN(t *testing.T) {
	h := createTestHarForStats()
	result := h.FastestRequests(0)
	assert.Nil(t, result)
}

func TestFastestRequestsNegativeN(t *testing.T) {
	h := createTestHarForStats()
	result := h.FastestRequests(-1)
	assert.Nil(t, result)
}

func TestFastestRequestsNLargerThanEntries(t *testing.T) {
	// Cover the branch where n > len(entries)
	h := createTestHarForStats()
	result := h.FastestRequests(100)
	assert.Equal(t, 5, len(result))
}

// --- LargestResponses uncovered branches ---

func TestLargestResponsesNilHar(t *testing.T) {
	var h *Har
	result := h.LargestResponses(5)
	assert.Nil(t, result)
}

func TestLargestResponsesZeroN(t *testing.T) {
	h := createTestHarForStats()
	result := h.LargestResponses(0)
	assert.Nil(t, result)
}

func TestLargestResponsesNegativeN(t *testing.T) {
	h := createTestHarForStats()
	result := h.LargestResponses(-1)
	assert.Nil(t, result)
}

func TestLargestResponsesNLargerThanEntries(t *testing.T) {
	// Cover the branch where n > len(entries)
	h := createTestHarForStats()
	result := h.LargestResponses(100)
	assert.Equal(t, 5, len(result))
}

// --- extractDomain uncovered branches ---

func TestExtractDomainInvalidURL(t *testing.T) {
	// Cover the err != nil branch
	result := extractDomain("://invalid")
	assert.Equal(t, "", result)
}

func TestExtractDomainEmptyURL(t *testing.T) {
	result := extractDomain("")
	assert.Equal(t, "", result)
}

func TestExtractDomainWithPort(t *testing.T) {
	result := extractDomain("https://example.com:8080/path")
	assert.Equal(t, "example.com:8080", result)
}

func TestExtractDomainSimple(t *testing.T) {
	result := extractDomain("https://www.example.com/path?q=1")
	assert.Equal(t, "www.example.com", result)
}

// --- Statistics additional uncovered branches ---

func TestStatisticsWithNegativeBodySize(t *testing.T) {
	// Cover the BodySize > 0 check (negative body size should not be counted)
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.SetResponseContent(100, "application/json")
	e.SetTimings(10, 5, 15, 2, 50, 30, 8)
	e.Response.BodySize = -1 // default from AddEntry

	stats := h.Statistics()
	assert.Equal(t, int64(0), stats.TotalTransferred, "Negative BodySize should not be counted")
}

func TestStatisticsWithNegativeContentSize(t *testing.T) {
	// Cover the Content.Size > 0 check
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.Response.Content.Size = -1
	e.Response.Content.MimeType = "text/html"
	e.SetTimings(10, 5, 15, 2, 50, 30, 8)

	stats := h.Statistics()
	assert.Equal(t, int64(0), stats.TotalUncompressed, "Negative Content.Size should not be counted")
}

func TestStatisticsWithPositiveBodySize(t *testing.T) {
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.SetResponseContent(100, "application/json")
	e.SetTimings(10, 5, 15, 2, 50, 30, 8)
	e.Response.BodySize = 500

	stats := h.Statistics()
	assert.Equal(t, int64(500), stats.TotalTransferred)
	assert.Equal(t, int64(100), stats.TotalUncompressed)
}

func TestStatisticsTimingsAllNegative(t *testing.T) {
	// Cover the branch where all timing values are negative/zero (not > 0)
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.SetResponseContent(100, "application/json")
	// Use negative timings (default from AddEntry is -1 for all)
	e.Timings = Timings{Blocked: -1, DNS: -1, Connect: -1, Send: -1, Wait: -1, Receive: -1, Ssl: -1}
	e.Time = 0

	stats := h.Statistics()
	// All timing min/avg should be 0 since no positive values
	assert.Equal(t, float64(0), stats.TimingsSummary.MinBlocked)
	assert.Equal(t, float64(0), stats.TimingsSummary.MinDNS)
	assert.Equal(t, float64(0), stats.TimingsSummary.MinConnect)
	assert.Equal(t, float64(0), stats.TimingsSummary.MinSend)
	assert.Equal(t, float64(0), stats.TimingsSummary.MinWait)
	assert.Equal(t, float64(0), stats.TimingsSummary.MinReceive)
	assert.Equal(t, float64(0), stats.TimingsSummary.MinSSL)
}

func TestStatisticsTimingsMixed(t *testing.T) {
	// Test with some positive and some negative timing values
	h := NewHar()
	e1 := h.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e1.SetResponseContent(100, "application/json")
	e1.Timings = Timings{Blocked: 10, DNS: -1, Connect: 20, Send: -1, Wait: 50, Receive: 30, Ssl: -1}
	e1.Time = 10 + 20 + 50 + 30

	e2 := h.AddEntry("GET", "https://example.com/api2", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")
	e2.SetResponseContent(100, "application/json")
	e2.Timings = Timings{Blocked: 5, DNS: -1, Connect: 15, Send: -1, Wait: 100, Receive: 60, Ssl: -1}
	e2.Time = 5 + 15 + 100 + 60

	stats := h.Statistics()
	assert.Equal(t, float64(10), stats.TimingsSummary.MaxBlocked)
	assert.Equal(t, float64(5), stats.TimingsSummary.MinBlocked)
	assert.Equal(t, float64(20), stats.TimingsSummary.MaxConnect)
	assert.Equal(t, float64(15), stats.TimingsSummary.MinConnect)
	assert.Equal(t, float64(100), stats.TimingsSummary.MaxWait)
	assert.Equal(t, float64(50), stats.TimingsSummary.MinWait)
	// DNS and SSL have no positive values
	assert.Equal(t, float64(0), stats.TimingsSummary.MaxDNS)
	assert.Equal(t, float64(0), stats.TimingsSummary.MinDNS)
	assert.Equal(t, float64(0), stats.TimingsSummary.MaxSSL)
	assert.Equal(t, float64(0), stats.TimingsSummary.MinSSL)
}

func TestStatisticsWithContentTypeWithCharset(t *testing.T) {
	// Cover the content type semicolon stripping in Statistics()
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.Response.Content.MimeType = "text/html; charset=utf-8"
	e.SetTimings(10, 5, 15, 2, 50, 30, 8)

	stats := h.Statistics()
	assert.Equal(t, 1, stats.ContentTypes["text/html"])
}

func TestStatisticsWithEmptyContentType(t *testing.T) {
	// Cover the contentType != "" branch
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	// MimeType is empty by default
	e.SetTimings(10, 5, 15, 2, 50, 30, 8)

	stats := h.Statistics()
	assert.Equal(t, 0, len(stats.ContentTypes))
}

func TestStatisticsTimeRangeAndPercentiles(t *testing.T) {
	// Test TotalTime, StartTime, EndTime, and percentile calculations
	h := NewHar()
	e1 := h.AddEntry("GET", "https://example.com/a", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e1.SetResponseContent(100, "text/html")
	e1.SetTimings(10, 5, 15, 2, 50, 30, 8)
	e1.StartedDateTime = time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

	e2 := h.AddEntry("GET", "https://example.com/b", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")
	e2.SetResponseContent(100, "text/html")
	e2.SetTimings(5, 3, 10, 1, 20, 10, 4)
	e2.StartedDateTime = time.Date(2024, 1, 1, 10, 0, 1, 0, time.UTC)

	stats := h.Statistics()
	assert.False(t, stats.StartTime.IsZero())
	assert.False(t, stats.EndTime.IsZero())
	assert.Greater(t, stats.TotalTime, float64(0))
	assert.Greater(t, stats.MedianTime, float64(0))
	assert.Greater(t, stats.P95Time, float64(0))
	assert.Greater(t, stats.P99Time, float64(0))
}

func TestStatisticsEntryWithInvalidURL(t *testing.T) {
	// Cover the case where extractDomain returns "" in Statistics()
	h := NewHar()
	e := h.AddEntry("GET", "not-a-url", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.SetResponseContent(100, "text/html")
	e.SetTimings(10, 5, 15, 2, 50, 30, 8)

	stats := h.Statistics()
	assert.Equal(t, 0, len(stats.Domains), "Invalid URLs should not add domains")
}

func TestStatisticsRedirectCount(t *testing.T) {
	// Ensure the 3xx redirect counting branch is exercised
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/old", "HTTP/1.1", "")
	e.SetResponseStatus(302, "Found")
	e.SetResponseContent(0, "text/html")
	e.SetTimings(1, 1, 1, 1, 10, 5, 1)

	stats := h.Statistics()
	assert.Equal(t, 1, stats.RedirectCount)
	assert.Equal(t, 0, stats.ErrorCount)
}

func TestStatisticsSingleEntry(t *testing.T) {
	// Test with a single entry for edge cases in percentile
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.SetResponseContent(100, "application/json")
	e.SetTimings(10, 5, 15, 2, 50, 30, 8)

	stats := h.Statistics()
	assert.Equal(t, 1, stats.TotalRequests)
	assert.Equal(t, stats.MaxTime, stats.MinTime)
	assert.Equal(t, stats.MaxTime, stats.MedianTime)
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
