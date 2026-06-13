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
