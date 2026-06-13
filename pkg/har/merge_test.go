package har

import (
	"testing"
	"time"
)

func createHarForMerge(suffix string) *Har {
	h := NewHar()
	h.SetCreator("test"+suffix, "1.0")

	e1 := h.AddEntry("GET", "https://example.com/api/"+suffix, "HTTP/1.1", "page_1")
	e1.SetResponseStatus(200, "OK")
	e1.StartedDateTime = time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

	e2 := h.AddEntry("POST", "https://other.com/data/"+suffix, "HTTP/1.1", "page_2")
	e2.SetResponseStatus(201, "Created")
	e2.StartedDateTime = time.Date(2024, 1, 1, 10, 1, 0, 0, time.UTC)

	return h
}

func TestMerge(t *testing.T) {
	har1 := createHarForMerge("a")
	har2 := createHarForMerge("b")

	result := Merge(har1, har2)

	if len(result.Log.Entries) != 4 {
		t.Errorf("Expected 4 entries after merge, got %d", len(result.Log.Entries))
	}
}

func TestMergeWithOptions(t *testing.T) {
	har1 := createHarForMerge("a")
	har2 := createHarForMerge("b")

	// With deduplication (should not dedup since URLs are different)
	opts := MergeOptions{SortByTime: true, Deduplicate: true}
	result := MergeWithOptions(opts, har1, har2)

	if len(result.Log.Entries) != 4 {
		t.Errorf("Expected 4 entries, got %d", len(result.Log.Entries))
	}

	// Test deduplication with same URLs
	har3 := NewHar()
	e1 := har3.AddEntry("GET", "https://example.com/test", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e1.StartedDateTime = time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

	har4 := NewHar()
	e2 := har4.AddEntry("GET", "https://example.com/test", "HTTP/1.1", "")
	e2.SetResponseStatus(404, "Not Found")
	e2.StartedDateTime = time.Date(2024, 1, 1, 10, 1, 0, 0, time.UTC)

	result = MergeWithOptions(opts, har3, har4)

	if len(result.Log.Entries) != 1 {
		t.Errorf("Expected 1 entry after dedup, got %d", len(result.Log.Entries))
	}

	// Should keep the newer entry (404)
	if result.Log.Entries[0].Response.Status != 404 {
		t.Errorf("Expected status 404 (newer), got %d", result.Log.Entries[0].Response.Status)
	}
}

func TestMergeEmpty(t *testing.T) {
	result := Merge()
	if result == nil {
		t.Error("Expected non-nil result for empty merge")
	}
}

func TestMergeNil(t *testing.T) {
	har1 := createHarForMerge("a")
	result := Merge(har1, nil)

	if len(result.Log.Entries) != 2 {
		t.Errorf("Expected 2 entries (nil should be skipped), got %d", len(result.Log.Entries))
	}
}

func TestSplitByPage(t *testing.T) {
	h := NewHar()
	h.AddPage("page_1", "Home")
	h.AddPage("page_2", "About")

	e1 := h.AddEntry("GET", "https://example.com/", "HTTP/1.1", "page_1")
	e1.SetResponseStatus(200, "OK")

	e2 := h.AddEntry("GET", "https://example.com/about", "HTTP/1.1", "page_2")
	e2.SetResponseStatus(200, "OK")

	e3 := h.AddEntry("GET", "https://example.com/style.css", "HTTP/1.1", "page_1")
	e3.SetResponseStatus(200, "OK")

	result := h.SplitByPage()

	if len(result) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(result))
	}

	if result["page_1"] == nil || len(result["page_1"].Log.Entries) != 2 {
		t.Errorf("Expected 2 entries for page_1, got %d", len(result["page_1"].Log.Entries))
	}

	if result["page_2"] == nil || len(result["page_2"].Log.Entries) != 1 {
		t.Errorf("Expected 1 entry for page_2, got %d", len(result["page_2"].Log.Entries))
	}
}

func TestSplitByDomain(t *testing.T) {
	h := NewHar()

	e1 := h.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")

	e2 := h.AddEntry("GET", "https://other.com/page", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")

	e3 := h.AddEntry("POST", "https://example.com/submit", "HTTP/1.1", "")
	e3.SetResponseStatus(201, "Created")

	result := h.SplitByDomain()

	if len(result) != 2 {
		t.Errorf("Expected 2 domains, got %d", len(result))
	}

	if result["example.com"] == nil || len(result["example.com"].Log.Entries) != 2 {
		t.Errorf("Expected 2 entries for example.com, got %d", len(result["example.com"].Log.Entries))
	}
}

func TestSplitByTimeRange(t *testing.T) {
	h := NewHar()

	e1 := h.AddEntry("GET", "https://example.com/1", "HTTP/1.1", "")
	e1.StartedDateTime = time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

	e2 := h.AddEntry("GET", "https://example.com/2", "HTTP/1.1", "")
	e2.StartedDateTime = time.Date(2024, 1, 1, 10, 0, 30, 0, time.UTC)

	e3 := h.AddEntry("GET", "https://example.com/3", "HTTP/1.1", "")
	e3.StartedDateTime = time.Date(2024, 1, 1, 11, 30, 0, 0, time.UTC)

	result := h.SplitByTimeRange(1 * time.Hour)

	if len(result) != 2 {
		t.Errorf("Expected 2 time groups, got %d", len(result))
	}
}

func TestSplitBySize(t *testing.T) {
	h := NewHar()
	for i := 0; i < 10; i++ {
		e := h.AddEntry("GET", "https://example.com/"+string(rune('a'+i)), "HTTP/1.1", "")
		e.SetResponseStatus(200, "OK")
	}

	result := h.SplitBySize(3)

	if len(result) != 4 {
		t.Errorf("Expected 4 groups (10/3 rounded up), got %d", len(result))
	}

	// First group should have 3 entries
	if len(result[0].Log.Entries) != 3 {
		t.Errorf("Expected 3 entries in first group, got %d", len(result[0].Log.Entries))
	}

	// Last group should have 1 entry
	if len(result[3].Log.Entries) != 1 {
		t.Errorf("Expected 1 entry in last group, got %d", len(result[3].Log.Entries))
	}
}

func TestSplitByStatusCode(t *testing.T) {
	h := NewHar()

	e1 := h.AddEntry("GET", "https://example.com/ok", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")

	e2 := h.AddEntry("GET", "https://example.com/redirect", "HTTP/1.1", "")
	e2.SetResponseStatus(301, "Moved")

	e3 := h.AddEntry("GET", "https://example.com/notfound", "HTTP/1.1", "")
	e3.SetResponseStatus(404, "Not Found")

	e4 := h.AddEntry("GET", "https://example.com/error", "HTTP/1.1", "")
	e4.SetResponseStatus(500, "Error")

	result := h.SplitByStatusCode()

	if len(result) != 4 {
		t.Errorf("Expected 4 groups, got %d", len(result))
	}

	if result["2xx"] == nil || len(result["2xx"].Log.Entries) != 1 {
		t.Error("Expected 1 entry in 2xx group")
	}

	if result["3xx"] == nil || len(result["3xx"].Log.Entries) != 1 {
		t.Error("Expected 1 entry in 3xx group")
	}

	if result["4xx"] == nil || len(result["4xx"].Log.Entries) != 1 {
		t.Error("Expected 1 entry in 4xx group")
	}

	if result["5xx"] == nil || len(result["5xx"].Log.Entries) != 1 {
		t.Error("Expected 1 entry in 5xx group")
	}
}

func TestSplitByMethod(t *testing.T) {
	h := NewHar()

	e1 := h.AddEntry("GET", "https://example.com/1", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")

	e2 := h.AddEntry("POST", "https://example.com/2", "HTTP/1.1", "")
	e2.SetResponseStatus(201, "Created")

	e3 := h.AddEntry("GET", "https://example.com/3", "HTTP/1.1", "")
	e3.SetResponseStatus(200, "OK")

	result := h.SplitByMethod()

	if len(result) != 2 {
		t.Errorf("Expected 2 methods, got %d", len(result))
	}

	if result["GET"] == nil || len(result["GET"].Log.Entries) != 2 {
		t.Error("Expected 2 GET entries")
	}

	if result["POST"] == nil || len(result["POST"].Log.Entries) != 1 {
		t.Error("Expected 1 POST entry")
	}
}

func TestSplitByPageNil(t *testing.T) {
	var h *Har
	result := h.SplitByPage()
	if len(result) != 0 {
		t.Error("Expected empty result for nil HAR")
	}
}

func TestSplitBySizeZero(t *testing.T) {
	h := NewHar()
	result := h.SplitBySize(0)
	if result != nil {
		t.Error("Expected nil for maxEntries=0")
	}
}
