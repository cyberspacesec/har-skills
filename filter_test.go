package har

import (
	"testing"
	"time"
)

func createHarForFilter() *Har {
	h := NewHar()
	h.SetCreator("test", "1.0")

	e1 := h.AddEntry("GET", "https://example.com/api/users", "HTTP/1.1", "page_1")
	e1.SetResponseStatus(200, "OK")
	e1.SetResponseContent(1024, "application/json")
	e1.AddRequestHeader("Accept", "application/json")
	e1.AddResponseHeader("Content-Type", "application/json")
	e1.ServerIPAddress = "1.2.3.4"
	e1.Connection = "conn1"

	e2 := h.AddEntry("POST", "https://other.com/api/data", "HTTP/1.1", "page_1")
	e2.SetResponseStatus(404, "Not Found")
	e2.SetResponseContent(512, "text/html")
	e2.AddRequestHeader("Content-Type", "application/json")
	e2.AddCookie("session", "abc123")

	e3 := h.AddEntry("GET", "https://example.com/static/style.css", "HTTP/1.1", "page_2")
	e3.SetResponseStatus(301, "Moved Permanently")
	e3.SetResponseContent(128, "text/css")
	e3.ServerIPAddress = "1.2.3.4"

	e4 := h.AddEntry("GET", "https://example.com/api/error", "HTTP/1.1", "")
	e4.SetResponseStatus(500, "Internal Server Error")
	e4.SetResponseContent(64, "application/json")
	e4.AddResponseHeader("X-Error", "true")

	e5 := h.AddEntry("GET", "https://cdn.example.com/script.js", "HTTP/1.1", "")
	e5.SetResponseStatus(200, "OK")
	e5.SetResponseContent(4096, "application/javascript")
	e5.AddCookie("session", "xyz789")
	e5.ServerIPAddress = "5.6.7.8"

	return h
}

func TestFilterNilHarMethods(t *testing.T) {
	var h *Har

	results := []*FilterResult{
		h.Filter(FilterOptions{}),
		h.FindByURL("example.com", false),
		h.FindByMethod("GET"),
		h.FindByStatusCode(200),
		h.FindErrors(),
		h.FindByTimeRange(time.Now(), time.Now()),
		h.FindByContentType("application/json"),
		h.FindSlowRequests(100),
		h.FindByDomain("example.com"),
		h.FindByHeader("Accept", "application/json"),
		h.FindByResponseHeader("Content-Type", "application/json"),
		h.FindByCookie("session"),
		h.FindByStatusCodeRange(200, 299),
		h.FindRedirects(),
		h.FindCacheHits(),
		h.FindByResourceType("xhr"),
		h.FindByServerIP("127.0.0.1"),
		h.FindByConnection("conn"),
	}

	for i, result := range results {
		if result == nil {
			t.Fatalf("result %d is nil", i)
		}
		if result.Count() != 0 {
			t.Fatalf("result %d count = %d, want 0", i, result.Count())
		}
	}
}

func TestFilterResultNilReceiverMethods(t *testing.T) {
	var result *FilterResult

	if result.Count() != 0 {
		t.Fatalf("Count() = %d, want 0", result.Count())
	}
	if result.First() != nil {
		t.Fatal("First() should return nil")
	}
	if result.Last() != nil {
		t.Fatal("Last() should return nil")
	}
	if result.At(0) != nil {
		t.Fatal("At() should return nil")
	}
	if result.SortByTime() != nil {
		t.Fatal("SortByTime() should return nil")
	}
	if result.SortByDuration() != nil {
		t.Fatal("SortByDuration() should return nil")
	}
	if result.SortByDurationDesc() != nil {
		t.Fatal("SortByDurationDesc() should return nil")
	}
	if result.SortBySize() != nil {
		t.Fatal("SortBySize() should return nil")
	}
	if result.SortBySizeDesc() != nil {
		t.Fatal("SortBySizeDesc() should return nil")
	}
	if result.Limit(1) != nil {
		t.Fatal("Limit() should return nil")
	}
	if result.Offset(1) != nil {
		t.Fatal("Offset() should return nil")
	}

	chained := result.Chain(FilterOptions{Method: "GET"})
	if chained == nil {
		t.Fatal("Chain() should return an empty result")
	}
	if chained.Count() != 0 {
		t.Fatalf("Chain().Count() = %d, want 0", chained.Count())
	}

	har := result.ToHar()
	if har == nil {
		t.Fatal("ToHar() should return an empty HAR")
	}
	if len(har.Log.Entries) != 0 {
		t.Fatalf("ToHar() entries = %d, want 0", len(har.Log.Entries))
	}
}

func TestFilterResultNegativeLimitAndOffset(t *testing.T) {
	h := createHarForFilter()

	limited := h.Filter(FilterOptions{}).Limit(-1)
	if limited.Count() != 0 {
		t.Fatalf("Limit(-1).Count() = %d, want 0", limited.Count())
	}

	offset := h.Filter(FilterOptions{}).Offset(-1)
	if offset.Count() != len(h.Log.Entries) {
		t.Fatalf("Offset(-1).Count() = %d, want %d", offset.Count(), len(h.Log.Entries))
	}
}

func TestFilterResultLast(t *testing.T) {
	h := createHarForFilter()
	result := h.FindByMethod("GET")

	last := result.Last()
	if last == nil {
		t.Fatal("Expected non-nil last entry")
	}

	if last.Request.URL != "https://cdn.example.com/script.js" {
		t.Errorf("Expected last entry to be script.js, got %s", last.Request.URL)
	}
}

func TestFilterResultAt(t *testing.T) {
	h := createHarForFilter()
	result := h.FindByMethod("GET")

	first := result.At(0)
	if first == nil {
		t.Fatal("Expected non-nil entry at index 0")
	}

	outOfRange := result.At(100)
	if outOfRange != nil {
		t.Error("Expected nil for out of range index")
	}
}

func TestFilterResultSortByTime(t *testing.T) {
	h := createHarForFilter()

	// Set specific times
	h.Log.Entries[0].StartedDateTime = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	h.Log.Entries[1].StartedDateTime = time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	h.Log.Entries[2].StartedDateTime = time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC)

	result := h.Filter(FilterOptions{}).SortByTime()
	if result.At(0).StartedDateTime.After(result.At(1).StartedDateTime) {
		t.Error("Expected entries sorted by time ascending")
	}
}

func TestFilterResultSortByDuration(t *testing.T) {
	h := createHarForFilter()
	h.Log.Entries[0].Time = 100
	h.Log.Entries[1].Time = 50
	h.Log.Entries[2].Time = 200

	result := h.Filter(FilterOptions{}).SortByDuration()
	if result.At(0).Time > result.At(1).Time {
		t.Error("Expected entries sorted by duration ascending")
	}
}

func TestFilterResultSortByDurationDesc(t *testing.T) {
	h := createHarForFilter()
	h.Log.Entries[0].Time = 100
	h.Log.Entries[1].Time = 50
	h.Log.Entries[2].Time = 200

	result := h.Filter(FilterOptions{}).SortByDurationDesc()
	if result.At(0).Time < result.At(1).Time {
		t.Error("Expected entries sorted by duration descending")
	}
}

func TestFilterResultSortBySize(t *testing.T) {
	h := createHarForFilter()
	result := h.Filter(FilterOptions{}).SortBySize()
	if result.At(0).Response.Content.Size > result.At(1).Response.Content.Size {
		t.Error("Expected entries sorted by size ascending")
	}
}

func TestFilterResultSortBySizeDesc(t *testing.T) {
	h := createHarForFilter()
	result := h.Filter(FilterOptions{}).SortBySizeDesc()
	if result.At(0).Response.Content.Size < result.At(1).Response.Content.Size {
		t.Error("Expected entries sorted by size descending")
	}
}

func TestFilterResultLimit(t *testing.T) {
	h := createHarForFilter()
	result := h.Filter(FilterOptions{}).Limit(2)
	if result.Count() != 2 {
		t.Errorf("Expected 2 entries after limit, got %d", result.Count())
	}
}

func TestFilterResultOffset(t *testing.T) {
	h := createHarForFilter()
	result := h.Filter(FilterOptions{}).Offset(3)
	if result.Count() != 2 {
		t.Errorf("Expected 2 entries after offset, got %d", result.Count())
	}
}

func TestFilterResultChain(t *testing.T) {
	h := createHarForFilter()

	// First filter by method, then chain by status code
	result := h.FindByMethod("GET").Chain(FilterOptions{
		StatusCode: 200,
	})

	if result.Count() != 2 {
		t.Errorf("Expected 2 GET requests with status 200, got %d", result.Count())
	}
}

func TestFindByDomain(t *testing.T) {
	h := createHarForFilter()
	result := h.FindByDomain("example.com")

	if result.Count() != 3 {
		t.Errorf("Expected 3 entries for example.com, got %d", result.Count())
	}
}

func TestFindByHeader(t *testing.T) {
	h := createHarForFilter()
	result := h.FindByHeader("Accept", "application/json")

	if result.Count() != 1 {
		t.Errorf("Expected 1 entry with Accept header, got %d", result.Count())
	}
}

func TestFindByResponseHeader(t *testing.T) {
	h := createHarForFilter()
	result := h.FindByResponseHeader("X-Error", "true")

	if result.Count() != 1 {
		t.Errorf("Expected 1 entry with X-Error header, got %d", result.Count())
	}
}

func TestFindByCookie(t *testing.T) {
	h := createHarForFilter()
	result := h.FindByCookie("session")

	if result.Count() != 2 {
		t.Errorf("Expected 2 entries with session cookie, got %d", result.Count())
	}
}

func TestFindByStatusCodeRange(t *testing.T) {
	h := createHarForFilter()
	result := h.FindByStatusCodeRange(200, 299)

	if result.Count() != 2 {
		t.Errorf("Expected 2 entries with 2xx status, got %d", result.Count())
	}
}

func TestFindRedirects(t *testing.T) {
	h := createHarForFilter()
	result := h.FindRedirects()

	if result.Count() != 1 {
		t.Errorf("Expected 1 redirect, got %d", result.Count())
	}
}

func TestFindCacheHits(t *testing.T) {
	h := createHarForFilter()
	result := h.FindCacheHits()

	// No cache data set, so should be 0
	if result.Count() != 0 {
		t.Errorf("Expected 0 cache hits, got %d", result.Count())
	}
}

func TestFindByServerIP(t *testing.T) {
	h := createHarForFilter()
	result := h.FindByServerIP("1.2.3.4")

	if result.Count() != 2 {
		t.Errorf("Expected 2 entries for IP 1.2.3.4, got %d", result.Count())
	}
}

func TestFindByConnection(t *testing.T) {
	h := createHarForFilter()
	result := h.FindByConnection("conn1")

	if result.Count() != 1 {
		t.Errorf("Expected 1 entry for conn1, got %d", result.Count())
	}
}

func TestFindByResourceType(t *testing.T) {
	h := createHarForFilter()
	h.Log.Entries[0].ResourceType = "xhr"
	h.Log.Entries[2].ResourceType = "stylesheet"

	result := h.FindByResourceType("xhr")
	if result.Count() != 1 {
		t.Errorf("Expected 1 xhr entry, got %d", result.Count())
	}
}

func TestFilterResultLimitMoreThanCount(t *testing.T) {
	h := createHarForFilter()
	result := h.Filter(FilterOptions{}).Limit(100)
	if result.Count() != 5 {
		t.Errorf("Expected 5 entries (limit > count), got %d", result.Count())
	}
}

func TestFilterResultOffsetMoreThanCount(t *testing.T) {
	h := createHarForFilter()
	result := h.Filter(FilterOptions{}).Offset(100)
	if result.Count() != 0 {
		t.Errorf("Expected 0 entries (offset > count), got %d", result.Count())
	}
}

func TestChainedFilterOperations(t *testing.T) {
	h := createHarForFilter()

	result := h.Filter(FilterOptions{}).
		SortByDurationDesc().
		Limit(3).
		Offset(1)

	if result.Count() != 2 {
		t.Errorf("Expected 2 entries after chained operations, got %d", result.Count())
	}
}

// --- Tests for previously uncovered functions and branches ---

// createHarForFilterComprehensive creates a HAR with entries covering
// many different filter branches: various URLs, methods, status codes,
// content types, times, durations, cookies, headers, cache data, etc.
func createHarForFilterComprehensive() *Har {
	h := NewHar()
	h.SetCreator("test", "1.0")

	// Entry 0: GET, 200, application/json, slow request
	e0 := h.AddEntry("GET", "https://api.example.com/v1/users", "HTTP/1.1", "page_1")
	e0.SetResponseStatus(200, "OK")
	e0.SetResponseContent(2048, "application/json")
	e0.AddRequestHeader("Accept", "application/json")
	e0.AddRequestHeader("Authorization", "Bearer token123")
	e0.AddResponseHeader("Content-Type", "application/json")
	e0.AddResponseHeader("X-Request-Id", "req-001")
	e0.ServerIPAddress = "10.0.0.1"
	e0.Connection = "conn-a"
	e0.Time = 1500.0 // slow request (ms)
	e0.StartedDateTime = time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)
	e0.AddCookie("session", "abc123")
	e0.AddCookie("theme", "dark")

	// Entry 1: POST, 201, application/json
	e1 := h.AddEntry("POST", "https://api.example.com/v1/users", "HTTP/1.1", "page_1")
	e1.SetResponseStatus(201, "Created")
	e1.SetResponseContent(512, "application/json")
	e1.AddRequestHeader("Content-Type", "application/json")
	e1.AddResponseHeader("Content-Type", "application/json; charset=utf-8")
	e1.Time = 200.0
	e1.StartedDateTime = time.Date(2024, 6, 15, 10, 1, 0, 0, time.UTC)
	e1.AddCookie("session", "def456")

	// Entry 2: GET, 301 redirect, text/html (no MimeType set, only header)
	e2 := h.AddEntry("GET", "https://example.com/old-page", "HTTP/1.1", "page_2")
	e2.SetResponseStatus(301, "Moved Permanently")
	// Set MimeType to empty so content-type fallback to header is tested
	e2.SetResponseContent(256, "")
	e2.AddResponseHeader("Content-Type", "text/html; charset=utf-8")
	e2.Time = 50.0
	e2.StartedDateTime = time.Date(2024, 6, 15, 10, 2, 0, 0, time.UTC)

	// Entry 3: GET, 404, application/json, error
	e3 := h.AddEntry("GET", "https://api.example.com/v1/missing", "HTTP/1.1", "")
	e3.SetResponseStatus(404, "Not Found")
	e3.SetResponseContent(128, "application/json")
	e3.AddResponseHeader("X-Error", "true")
	e3.Time = 80.0
	e3.StartedDateTime = time.Date(2024, 6, 15, 10, 3, 0, 0, time.UTC)

	// Entry 4: GET, 500, application/json, server error
	e4 := h.AddEntry("GET", "https://api.example.com/v1/broken", "HTTP/1.1", "")
	e4.SetResponseStatus(500, "Internal Server Error")
	e4.SetResponseContent(64, "application/json")
	e4.Time = 3000.0 // very slow
	e4.StartedDateTime = time.Date(2024, 6, 15, 10, 4, 0, 0, time.UTC)

	// Entry 5: GET, 200, text/css, with cache hit (BeforeRequest)
	e5 := h.AddEntry("GET", "https://cdn.example.com/style.css", "HTTP/1.1", "")
	e5.SetResponseStatus(200, "OK")
	e5.SetResponseContent(4096, "text/css")
	e5.AddResponseHeader("Content-Type", "text/css")
	e5.Time = 30.0
	e5.StartedDateTime = time.Date(2024, 6, 15, 10, 5, 0, 0, time.UTC)
	e5.Cache.BeforeRequest = &BeforeRequest{
		ETag:       "abc",
		HitCount:   3,
		LastAccess: time.Date(2024, 6, 15, 9, 0, 0, 0, time.UTC),
	}
	e5.ServerIPAddress = "10.0.0.2"

	// Entry 6: GET, 200, application/javascript, with cache hit (AfterRequest)
	e6 := h.AddEntry("GET", "https://cdn.example.com/app.js", "HTTP/1.1", "")
	e6.SetResponseStatus(200, "OK")
	e6.SetResponseContent(8192, "application/javascript")
	e6.Time = 45.0
	e6.StartedDateTime = time.Date(2024, 6, 15, 10, 6, 0, 0, time.UTC)
	e6.Cache.AfterRequest = &AfterRequest{
		ETag:       "def",
		HitCount:   1,
		LastAccess: time.Date(2024, 6, 15, 9, 30, 0, 0, time.UTC),
	}

	// Entry 7: GET, 200, image/png, with response cookie
	e7 := h.AddEntry("GET", "https://img.example.com/logo.png", "HTTP/1.1", "")
	e7.SetResponseStatus(200, "OK")
	e7.SetResponseContent(16384, "image/png")
	e7.Time = 100.0
	e7.StartedDateTime = time.Date(2024, 6, 15, 10, 7, 0, 0, time.UTC)
	e7.Response.Cookies = append(e7.Response.Cookies, Cookie{Name: "tracking", Value: "xyz"})

	return h
}

// --- FindByURL tests ---

func TestFindByURL_Substring(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.FindByURL("api.example.com", false)
	// Should match entries 0, 1, 3, 4 (all api.example.com URLs)
	if result.Count() != 4 {
		t.Errorf("Expected 4 entries matching 'api.example.com', got %d", result.Count())
	}
}

func TestFindByURL_SubstringNoMatch(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.FindByURL("nonexistent.example.com", false)
	if result.Count() != 0 {
		t.Errorf("Expected 0 entries, got %d", result.Count())
	}
}

func TestFindByURL_Regex(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.FindByURL(`^https://api\.example\.com/v1/`, true)
	// Should match entries 0, 1, 3, 4
	if result.Count() != 4 {
		t.Errorf("Expected 4 entries matching regex, got %d", result.Count())
	}
}

func TestFindByURL_RegexNoMatch(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.FindByURL(`^https://nonexistent\.com`, true)
	if result.Count() != 0 {
		t.Errorf("Expected 0 entries, got %d", result.Count())
	}
}

func TestFindByURL_InvalidRegex(t *testing.T) {
	h := createHarForFilterComprehensive()
	// Invalid regex should not panic; it compiles with error so the regex
	// branch is skipped (err != nil), and the URL filter becomes a no-op,
	// meaning all entries pass the URL check
	result := h.FindByURL("[invalid(regex", true)
	// Since the regex compile fails, the URL filter doesn't exclude anything,
	// so all entries should be returned
	if result.Count() != h.Filter(FilterOptions{}).Count() {
		t.Errorf("Expected all entries when regex is invalid, got %d", result.Count())
	}
}

// --- FindByStatusCode tests ---

func TestFindByStatusCode(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.FindByStatusCode(200)
	if result.Count() != 4 {
		t.Errorf("Expected 4 entries with status 200, got %d", result.Count())
	}
}

func TestFindByStatusCode_NoMatch(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.FindByStatusCode(403)
	if result.Count() != 0 {
		t.Errorf("Expected 0 entries with status 403, got %d", result.Count())
	}
}

func TestFindByStatusCode_404(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.FindByStatusCode(404)
	if result.Count() != 1 {
		t.Errorf("Expected 1 entry with status 404, got %d", result.Count())
	}
}

// --- FindErrors tests ---

func TestFindErrors(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.FindErrors()
	// Entries 3 (404) and 4 (500)
	if result.Count() != 2 {
		t.Errorf("Expected 2 error entries, got %d", result.Count())
	}
	for _, e := range result.Entries {
		if e.Response.Status < 400 || e.Response.Status >= 600 {
			t.Errorf("Expected error status (4xx/5xx), got %d", e.Response.Status)
		}
	}
}

func TestFindErrors_NoErrors(t *testing.T) {
	h := NewHar()
	e := h.AddEntry("GET", "https://ok.example.com", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	result := h.FindErrors()
	if result.Count() != 0 {
		t.Errorf("Expected 0 error entries, got %d", result.Count())
	}
}

// --- FindByTimeRange tests ---

func TestFindByTimeRange(t *testing.T) {
	h := createHarForFilterComprehensive()
	start := time.Date(2024, 6, 15, 10, 1, 0, 0, time.UTC)
	end := time.Date(2024, 6, 15, 10, 4, 0, 0, time.UTC)
	result := h.FindByTimeRange(start, end)
	// Entries 1, 2, 3, 4 have times 10:01, 10:02, 10:03, 10:04
	// EndTime is inclusive (After check), so entry 4 at 10:04 is included
	if result.Count() != 4 {
		t.Errorf("Expected 4 entries in time range, got %d", result.Count())
	}
}

func TestFindByTimeRange_NoMatch(t *testing.T) {
	h := createHarForFilterComprehensive()
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
	result := h.FindByTimeRange(start, end)
	if result.Count() != 0 {
		t.Errorf("Expected 0 entries in future time range, got %d", result.Count())
	}
}

func TestFindByTimeRange_ExactBoundary(t *testing.T) {
	h := createHarForFilterComprehensive()
	// Use exact start time of entry 0
	start := time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)
	result := h.FindByTimeRange(start, end)
	// Entry 0 is at exactly 10:00:00, so it should be included
	// (not before start, not after end)
	if result.Count() != 1 {
		t.Errorf("Expected 1 entry at exact boundary, got %d", result.Count())
	}
}

// --- FindByContentType tests ---

func TestFindByContentType_MimeType(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.FindByContentType("application/json")
	// Entries 0, 1, 3, 4 have application/json MimeType
	if result.Count() != 4 {
		t.Errorf("Expected 4 entries with application/json, got %d", result.Count())
	}
}

func TestFindByContentType_HeaderFallback(t *testing.T) {
	h := createHarForFilterComprehensive()
	// Entry 2 has empty MimeType but Content-Type: text/html header
	result := h.FindByContentType("text/html")
	if result.Count() != 1 {
		t.Errorf("Expected 1 entry with text/html (via header), got %d", result.Count())
	}
}

func TestFindByContentType_NoMatch(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.FindByContentType("application/xml")
	if result.Count() != 0 {
		t.Errorf("Expected 0 entries with application/xml, got %d", result.Count())
	}
}

func TestFindByContentType_CaseInsensitive(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.FindByContentType("Application/JSON")
	// Case-insensitive match on MimeType
	if result.Count() != 4 {
		t.Errorf("Expected 4 entries with case-insensitive match, got %d", result.Count())
	}
}

// --- FindSlowRequests tests ---

func TestFindSlowRequests(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.FindSlowRequests(1000.0)
	// Entry 0 (1500ms) and entry 4 (3000ms)
	if result.Count() != 2 {
		t.Errorf("Expected 2 slow requests, got %d", result.Count())
	}
}

func TestFindSlowRequests_AllSlow(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.FindSlowRequests(1.0)
	// All entries are >= 1ms
	if result.Count() != 8 {
		t.Errorf("Expected all 8 entries to be slow, got %d", result.Count())
	}
}

func TestFindSlowRequests_NoneSlow(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.FindSlowRequests(5000.0)
	// No entries >= 5000ms
	if result.Count() != 0 {
		t.Errorf("Expected 0 slow requests, got %d", result.Count())
	}
}

// --- FindByCookie tests (response cookie path) ---

func TestFindByCookie_ResponseCookie(t *testing.T) {
	h := createHarForFilterComprehensive()
	// Entry 7 has a response cookie named "tracking"
	result := h.FindByCookie("tracking")
	if result.Count() != 1 {
		t.Errorf("Expected 1 entry with response cookie 'tracking', got %d", result.Count())
	}
	if result.First().Request.URL != "https://img.example.com/logo.png" {
		t.Errorf("Expected logo.png entry, got %s", result.First().Request.URL)
	}
}

func TestFindByCookie_RequestCookie(t *testing.T) {
	h := createHarForFilterComprehensive()
	// Entries 0 and 1 have request cookie "session"
	result := h.FindByCookie("session")
	if result.Count() != 2 {
		t.Errorf("Expected 2 entries with request cookie 'session', got %d", result.Count())
	}
}

func TestFindByCookie_NotFound(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.FindByCookie("nonexistent_cookie")
	if result.Count() != 0 {
		t.Errorf("Expected 0 entries, got %d", result.Count())
	}
}

// --- FindCacheHits tests (with actual cache data) ---

func TestFindCacheHits_BeforeRequest(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.FindCacheHits()
	// Entry 5 has BeforeRequest with HitCount=3, entry 6 has AfterRequest with HitCount=1
	if result.Count() != 2 {
		t.Errorf("Expected 2 cache hits, got %d", result.Count())
	}
}

func TestFindCacheHits_EmptyCache(t *testing.T) {
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	// No cache data set
	result := h.FindCacheHits()
	if result.Count() != 0 {
		t.Errorf("Expected 0 cache hits with empty cache, got %d", result.Count())
	}
}

func TestFindCacheHits_ZeroHitCount(t *testing.T) {
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	// BeforeRequest exists but HitCount is 0
	e.Cache.BeforeRequest = &BeforeRequest{
		ETag:       "abc",
		HitCount:   0,
		LastAccess: time.Now(),
	}
	result := h.FindCacheHits()
	if result.Count() != 0 {
		t.Errorf("Expected 0 cache hits with HitCount=0, got %d", result.Count())
	}
}

func TestFindCacheHits_BothCacheStates(t *testing.T) {
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.Cache.BeforeRequest = &BeforeRequest{
		ETag:       "abc",
		HitCount:   2,
		LastAccess: time.Now(),
	}
	e.Cache.AfterRequest = &AfterRequest{
		ETag:       "def",
		HitCount:   3,
		LastAccess: time.Now(),
	}
	result := h.FindCacheHits()
	if result.Count() != 1 {
		t.Errorf("Expected 1 cache hit entry, got %d", result.Count())
	}
}

// --- First tests ---

func TestFilterResultFirst(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.FindByMethod("POST")
	first := result.First()
	if first == nil {
		t.Fatal("Expected non-nil first entry")
	}
	if first.Request.URL != "https://api.example.com/v1/users" {
		t.Errorf("Expected first POST entry URL to be api users, got %s", first.Request.URL)
	}
}

func TestFilterResultFirst_Empty(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.FindByMethod("DELETE")
	first := result.First()
	if first != nil {
		t.Error("Expected nil for empty filter result First()")
	}
}

// --- Last tests (empty case) ---

func TestFilterResultLast_Empty(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.FindByMethod("DELETE")
	last := result.Last()
	if last != nil {
		t.Error("Expected nil for empty filter result Last()")
	}
}

// --- ToHar tests ---

func TestFilterResultToHar(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.FindByMethod("POST")
	newHar := result.ToHar()
	if newHar == nil {
		t.Fatal("Expected non-nil Har from ToHar()")
	}
	if len(newHar.Log.Entries) != 1 {
		t.Errorf("Expected 1 entry in new Har, got %d", len(newHar.Log.Entries))
	}
	if newHar.Log.Entries[0].Request.Method != "POST" {
		t.Errorf("Expected POST method in new Har entry, got %s", newHar.Log.Entries[0].Request.Method)
	}
}

func TestFilterResultToHar_Empty(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.FindByMethod("DELETE")
	newHar := result.ToHar()
	if newHar == nil {
		t.Fatal("Expected non-nil Har from ToHar() even for empty result")
	}
	if len(newHar.Log.Entries) != 0 {
		t.Errorf("Expected 0 entries in new Har, got %d", len(newHar.Log.Entries))
	}
}

func TestFilterResultToHar_MultipleEntries(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.FindByContentType("application/json")
	newHar := result.ToHar()
	if len(newHar.Log.Entries) != 4 {
		t.Errorf("Expected 4 entries in new Har, got %d", len(newHar.Log.Entries))
	}
}

// --- matchesFilter comprehensive branch tests ---

func TestMatchesFilter_URLSubstring(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.Filter(FilterOptions{URL: "cdn.example.com"})
	if result.Count() != 2 {
		t.Errorf("Expected 2 entries with 'cdn.example.com', got %d", result.Count())
	}
}

func TestMatchesFilter_URLRegex(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.Filter(FilterOptions{URL: `\.css$`, UseRegex: true})
	if result.Count() != 1 {
		t.Errorf("Expected 1 entry ending with .css, got %d", result.Count())
	}
}

func TestMatchesFilter_URLRegexInvalid(t *testing.T) {
	h := createHarForFilterComprehensive()
	// Invalid regex: compile error, so URL filter is effectively skipped
	result := h.Filter(FilterOptions{URL: "[invalid", UseRegex: true})
	// All entries should pass since regex compilation fails
	allCount := h.Filter(FilterOptions{}).Count()
	if result.Count() != allCount {
		t.Errorf("Expected %d entries (regex compile failure = no URL filter), got %d", allCount, result.Count())
	}
}

func TestMatchesFilter_Method(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.Filter(FilterOptions{Method: "POST"})
	if result.Count() != 1 {
		t.Errorf("Expected 1 POST entry, got %d", result.Count())
	}
}

func TestMatchesFilter_MethodNoMatch(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.Filter(FilterOptions{Method: "PUT"})
	if result.Count() != 0 {
		t.Errorf("Expected 0 PUT entries, got %d", result.Count())
	}
}

func TestMatchesFilter_StatusCode(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.Filter(FilterOptions{StatusCode: 201})
	if result.Count() != 1 {
		t.Errorf("Expected 1 entry with status 201, got %d", result.Count())
	}
}

func TestMatchesFilter_StatusCodeMin(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.Filter(FilterOptions{StatusCodeMin: 400})
	// Entries 3 (404), 4 (500)
	if result.Count() != 2 {
		t.Errorf("Expected 2 entries with status >= 400, got %d", result.Count())
	}
}

func TestMatchesFilter_StatusCodeMax(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.Filter(FilterOptions{StatusCodeMax: 301})
	// Entries 0 (200), 1 (201), 2 (301), 5 (200), 6 (200), 7 (200) = 6
	if result.Count() != 6 {
		t.Errorf("Expected 6 entries with status <= 301, got %d", result.Count())
	}
}

func TestMatchesFilter_StatusCodeRange(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.Filter(FilterOptions{StatusCodeMin: 200, StatusCodeMax: 299})
	// Entries 0 (200), 1 (201), 5 (200), 6 (200), 7 (200) = 5
	if result.Count() != 5 {
		t.Errorf("Expected 5 entries with 2xx status, got %d", result.Count())
	}
}

func TestMatchesFilter_ContentTypeViaMimeType(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.Filter(FilterOptions{ContentType: "text/css"})
	// Entry 5 has MimeType text/css
	if result.Count() != 1 {
		t.Errorf("Expected 1 entry with text/css MimeType, got %d", result.Count())
	}
}

func TestMatchesFilter_ContentTypeViaHeader(t *testing.T) {
	h := createHarForFilterComprehensive()
	// Entry 2 has empty MimeType but Content-Type: text/html header
	result := h.Filter(FilterOptions{ContentType: "text/html"})
	if result.Count() != 1 {
		t.Errorf("Expected 1 entry with text/html via header, got %d", result.Count())
	}
}

func TestMatchesFilter_ContentTypeNoMatch(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.Filter(FilterOptions{ContentType: "application/xml"})
	if result.Count() != 0 {
		t.Errorf("Expected 0 entries with application/xml, got %d", result.Count())
	}
}

func TestMatchesFilter_StartTime(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.Filter(FilterOptions{
		StartTime: time.Date(2024, 6, 15, 10, 4, 0, 0, time.UTC),
	})
	// Entries at 10:04, 10:05, 10:06, 10:07 = entries 4, 5, 6, 7
	if result.Count() != 4 {
		t.Errorf("Expected 4 entries at or after 10:04, got %d", result.Count())
	}
}

func TestMatchesFilter_EndTime(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.Filter(FilterOptions{
		EndTime: time.Date(2024, 6, 15, 10, 1, 0, 0, time.UTC),
	})
	// Entries at 10:00, 10:01 = entries 0, 1
	if result.Count() != 2 {
		t.Errorf("Expected 2 entries at or before 10:01, got %d", result.Count())
	}
}

func TestMatchesFilter_MinDuration(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.Filter(FilterOptions{MinDuration: 1000.0})
	// Entry 0 (1500ms), entry 4 (3000ms)
	if result.Count() != 2 {
		t.Errorf("Expected 2 entries with duration >= 1000ms, got %d", result.Count())
	}
}

func TestMatchesFilter_MaxDuration(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.Filter(FilterOptions{MaxDuration: 100.0})
	// Entry 2 (50ms), entry 3 (80ms), entry 5 (30ms), entry 6 (45ms), entry 7 (100ms)
	if result.Count() != 5 {
		t.Errorf("Expected 5 entries with duration <= 100ms, got %d", result.Count())
	}
}

func TestMatchesFilter_DurationRange(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.Filter(FilterOptions{MinDuration: 50.0, MaxDuration: 200.0})
	// Entry 1 (200ms), 2 (50ms), 3 (80ms), 7 (100ms)
	if result.Count() != 4 {
		t.Errorf("Expected 4 entries with 50ms <= duration <= 200ms, got %d", result.Count())
	}
}

func TestMatchesFilter_ResourceType(t *testing.T) {
	h := createHarForFilterComprehensive()
	h.Log.Entries[0].ResourceType = "xhr"
	h.Log.Entries[5].ResourceType = "stylesheet"
	result := h.Filter(FilterOptions{ResourceType: "xhr"})
	if result.Count() != 1 {
		t.Errorf("Expected 1 xhr entry, got %d", result.Count())
	}
}

func TestMatchesFilter_ResourceTypeNoMatch(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.Filter(FilterOptions{ResourceType: "font"})
	if result.Count() != 0 {
		t.Errorf("Expected 0 font entries, got %d", result.Count())
	}
}

func TestMatchesFilter_HasError(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.Filter(FilterOptions{HasError: true})
	// Entries 3 (404), 4 (500)
	if result.Count() != 2 {
		t.Errorf("Expected 2 error entries, got %d", result.Count())
	}
}

func TestMatchesFilter_HasError_ExcludesNonErrors(t *testing.T) {
	h := NewHar()
	e := h.AddEntry("GET", "https://ok.example.com", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e2 := h.AddEntry("GET", "https://redirect.example.com", "HTTP/1.1", "")
	e2.SetResponseStatus(302, "Found")
	result := h.Filter(FilterOptions{HasError: true})
	if result.Count() != 0 {
		t.Errorf("Expected 0 error entries (200 and 302 are not errors), got %d", result.Count())
	}
}

func TestMatchesFilter_HasError_ExcludesStatusBelow400(t *testing.T) {
	h := NewHar()
	// Status 399 is not an error
	e := h.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	e.SetResponseStatus(399, "")
	// Status 600 is not an error per the code's definition (>= 600 is excluded)
	e2 := h.AddEntry("GET", "https://example.com/2", "HTTP/1.1", "")
	e2.SetResponseStatus(600, "")
	result := h.Filter(FilterOptions{HasError: true})
	if result.Count() != 0 {
		t.Errorf("Expected 0 error entries (399 and 600 are outside 400-599), got %d", result.Count())
	}
}

func TestMatchesFilter_HeaderNameOnly(t *testing.T) {
	h := createHarForFilterComprehensive()
	// Find entries that have an Authorization header (any value)
	result := h.Filter(FilterOptions{HeaderName: "Authorization"})
	if result.Count() != 1 {
		t.Errorf("Expected 1 entry with Authorization header, got %d", result.Count())
	}
}

func TestMatchesFilter_HeaderNameAndValue(t *testing.T) {
	h := createHarForFilterComprehensive()
	// Find entries with Authorization header containing "Bearer"
	result := h.Filter(FilterOptions{HeaderName: "Authorization", HeaderValue: "Bearer"})
	if result.Count() != 1 {
		t.Errorf("Expected 1 entry with Authorization: Bearer, got %d", result.Count())
	}
}

func TestMatchesFilter_HeaderNameValueNoMatch(t *testing.T) {
	h := createHarForFilterComprehensive()
	// Find entries with Authorization header containing "Basic"
	result := h.Filter(FilterOptions{HeaderName: "Authorization", HeaderValue: "Basic"})
	if result.Count() != 0 {
		t.Errorf("Expected 0 entries with Authorization: Basic, got %d", result.Count())
	}
}

func TestMatchesFilter_HeaderNameNotFound(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.Filter(FilterOptions{HeaderName: "X-Custom-Header"})
	if result.Count() != 0 {
		t.Errorf("Expected 0 entries with X-Custom-Header, got %d", result.Count())
	}
}

func TestMatchesFilter_RespHeaderNameOnly(t *testing.T) {
	h := createHarForFilterComprehensive()
	// Find entries with X-Error response header
	result := h.Filter(FilterOptions{RespHeaderName: "X-Error"})
	if result.Count() != 1 {
		t.Errorf("Expected 1 entry with X-Error response header, got %d", result.Count())
	}
}

func TestMatchesFilter_RespHeaderNameAndValue(t *testing.T) {
	h := createHarForFilterComprehensive()
	// Find entries with Content-Type response header containing "text/css"
	result := h.Filter(FilterOptions{RespHeaderName: "Content-Type", RespHeaderValue: "text/css"})
	if result.Count() != 1 {
		t.Errorf("Expected 1 entry with Content-Type: text/css response header, got %d", result.Count())
	}
}

func TestMatchesFilter_RespHeaderNameValueNoMatch(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.Filter(FilterOptions{RespHeaderName: "Content-Type", RespHeaderValue: "application/xml"})
	if result.Count() != 0 {
		t.Errorf("Expected 0 entries with Content-Type: application/xml, got %d", result.Count())
	}
}

func TestMatchesFilter_RespHeaderNameNotFound(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.Filter(FilterOptions{RespHeaderName: "X-Nonexistent"})
	if result.Count() != 0 {
		t.Errorf("Expected 0 entries with X-Nonexistent response header, got %d", result.Count())
	}
}

// Test combined filters to exercise multiple matchesFilter branches together
func TestMatchesFilter_CombinedURLAndMethod(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.Filter(FilterOptions{
		URL:    "api.example.com",
		Method: "POST",
	})
	if result.Count() != 1 {
		t.Errorf("Expected 1 POST api.example.com entry, got %d", result.Count())
	}
}

func TestMatchesFilter_CombinedStatusCodeAndContentType(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.Filter(FilterOptions{
		StatusCode:  200,
		ContentType: "application/json",
	})
	// Entry 0 is 200 + application/json
	if result.Count() != 1 {
		t.Errorf("Expected 1 entry with 200 + json, got %d", result.Count())
	}
}

func TestMatchesFilter_CombinedTimeRangeAndDuration(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.Filter(FilterOptions{
		StartTime:   time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC),
		EndTime:     time.Date(2024, 6, 15, 10, 5, 0, 0, time.UTC),
		MinDuration: 1000.0,
	})
	// Entry 0 (10:00, 1500ms) and entry 4 (10:04, 3000ms) are in range and slow
	if result.Count() != 2 {
		t.Errorf("Expected 2 entries in time range and slow, got %d", result.Count())
	}
}

func TestMatchesFilter_EmptyOptions(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.Filter(FilterOptions{})
	if result.Count() != 8 {
		t.Errorf("Expected all 8 entries with empty filter options, got %d", result.Count())
	}
}

// Test that HeaderValue match works case-sensitively (Contains, not EqualFold)
func TestMatchesFilter_HeaderValueCaseSensitive(t *testing.T) {
	h := createHarForFilterComprehensive()
	// Header name is case-insensitive (EqualFold), but value uses Contains
	result := h.Filter(FilterOptions{HeaderName: "authorization", HeaderValue: "Bearer"})
	if result.Count() != 1 {
		t.Errorf("Expected 1 entry (case-insensitive header name), got %d", result.Count())
	}
}

// Test that RespHeaderName is case-insensitive
func TestMatchesFilter_RespHeaderNameCaseInsensitive(t *testing.T) {
	h := createHarForFilterComprehensive()
	result := h.Filter(FilterOptions{RespHeaderName: "content-type", RespHeaderValue: "text/css"})
	if result.Count() != 1 {
		t.Errorf("Expected 1 entry (case-insensitive resp header name), got %d", result.Count())
	}
}
