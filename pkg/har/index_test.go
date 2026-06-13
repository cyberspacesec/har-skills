package har

import (
	"testing"
	"time"
)

func createHarForIndex() *Har {
	h := NewHar()
	h.SetCreator("test", "1.0")

	t1 := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC)
	t3 := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	e1 := h.AddEntry("GET", "https://example.com/api/users", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e1.SetResponseContent(1024, "application/json")
	e1.StartedDateTime = t1

	e2 := h.AddEntry("POST", "https://example.com/api/create", "HTTP/1.1", "")
	e2.SetResponseStatus(201, "Created")
	e2.SetResponseContent(128, "application/json")
	e2.StartedDateTime = t2

	e3 := h.AddEntry("GET", "https://cdn.example.com/script.js", "HTTP/1.1", "")
	e3.SetResponseStatus(200, "OK")
	e3.SetResponseContent(4096, "application/javascript")
	e3.StartedDateTime = t3

	e4 := h.AddEntry("GET", "https://example.com/api/users", "HTTP/1.1", "")
	e4.SetResponseStatus(304, "Not Modified")
	e4.SetResponseContent(0, "application/json")
	e4.StartedDateTime = t3

	e5 := h.AddEntry("GET", "https://other.com/page", "HTTP/1.1", "")
	e5.SetResponseStatus(404, "Not Found")
	e5.SetResponseContent(512, "text/html")
	e5.StartedDateTime = t1

	return h
}

func TestBuildIndex(t *testing.T) {
	h := createHarForIndex()
	idx := h.BuildIndex()

	if idx == nil {
		t.Fatal("Expected non-nil index")
	}

	if idx.Size() != 5 {
		t.Errorf("Expected 5 entries in index, got %d", idx.Size())
	}
}

func TestIndexByURL(t *testing.T) {
	h := createHarForIndex()
	idx := h.BuildIndex()

	entries := idx.ByURL("https://example.com/api/users")
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries for /api/users URL, got %d", len(entries))
	}
}

func TestIndexByURLNotFound(t *testing.T) {
	h := createHarForIndex()
	idx := h.BuildIndex()

	entries := idx.ByURL("https://nonexistent.com/api")
	if entries != nil {
		t.Errorf("Expected nil for non-existent URL, got %d entries", len(entries))
	}
}

func TestIndexByMethod(t *testing.T) {
	h := createHarForIndex()
	idx := h.BuildIndex()

	getEntries := idx.ByMethod("GET")
	if len(getEntries) != 4 {
		t.Errorf("Expected 4 GET entries, got %d", len(getEntries))
	}

	postEntries := idx.ByMethod("POST")
	if len(postEntries) != 1 {
		t.Errorf("Expected 1 POST entry, got %d", len(postEntries))
	}
}

func TestIndexByStatus(t *testing.T) {
	h := createHarForIndex()
	idx := h.BuildIndex()

	status200 := idx.ByStatus(200)
	if len(status200) != 2 {
		t.Errorf("Expected 2 entries with status 200, got %d", len(status200))
	}

	status404 := idx.ByStatus(404)
	if len(status404) != 1 {
		t.Errorf("Expected 1 entry with status 404, got %d", len(status404))
	}
}

func TestIndexByDomain(t *testing.T) {
	h := createHarForIndex()
	idx := h.BuildIndex()

	exampleEntries := idx.ByDomain("example.com")
	if len(exampleEntries) != 3 {
		t.Errorf("Expected 3 entries for example.com, got %d", len(exampleEntries))
	}

	cdnEntries := idx.ByDomain("cdn.example.com")
	if len(cdnEntries) != 1 {
		t.Errorf("Expected 1 entry for cdn.example.com, got %d", len(cdnEntries))
	}
}

func TestIndexByMimeType(t *testing.T) {
	h := createHarForIndex()
	idx := h.BuildIndex()

	jsonEntries := idx.ByMimeType("application/json")
	if len(jsonEntries) != 3 {
		t.Errorf("Expected 3 entries for application/json, got %d", len(jsonEntries))
	}
}

func TestIndexByURLPattern(t *testing.T) {
	h := createHarForIndex()
	idx := h.BuildIndex()

	entries := idx.ByURLPattern(`example\.com/api`)
	if len(entries) < 2 {
		t.Errorf("Expected at least 2 entries matching URL pattern, got %d", len(entries))
	}
}

func TestIndexByURLPatternInvalid(t *testing.T) {
	h := createHarForIndex()
	idx := h.BuildIndex()

	entries := idx.ByURLPattern(`[invalid regex`)
	if entries != nil {
		t.Error("Expected nil for invalid regex pattern")
	}
}

func TestIndexByTimeRange(t *testing.T) {
	h := createHarForIndex()
	idx := h.BuildIndex()

	start := time.Date(2024, 1, 1, 10, 30, 0, 0, time.UTC)
	end := time.Date(2024, 1, 1, 11, 30, 0, 0, time.UTC)

	entries := idx.ByTimeRange(start, end)
	// Only the 11:00 entry should match
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry in time range, got %d", len(entries))
	}
}

func TestIndexByTimeRangeInclusive(t *testing.T) {
	h := createHarForIndex()
	idx := h.BuildIndex()

	start := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	entries := idx.ByTimeRange(start, end)
	// All entries should match (inclusive range)
	if len(entries) != 5 {
		t.Errorf("Expected 5 entries in inclusive time range, got %d", len(entries))
	}
}

func TestIndexSize(t *testing.T) {
	h := createHarForIndex()
	idx := h.BuildIndex()

	if idx.Size() != 5 {
		t.Errorf("Expected index size 5, got %d", idx.Size())
	}
}

func TestIndexSizeEmpty(t *testing.T) {
	h := NewHar()
	idx := h.BuildIndex()

	if idx.Size() != 0 {
		t.Errorf("Expected index size 0 for empty Har, got %d", idx.Size())
	}
}

func TestIndexStats(t *testing.T) {
	h := createHarForIndex()
	idx := h.BuildIndex()

	stats := idx.Stats()

	if stats.UniqueURLs != 4 {
		t.Errorf("Expected 4 unique URLs, got %d", stats.UniqueURLs)
	}

	if stats.UniqueDomains != 3 {
		t.Errorf("Expected 3 unique domains, got %d", stats.UniqueDomains)
	}

	// Check status codes are sorted
	if len(stats.StatusCodes) < 3 {
		t.Errorf("Expected at least 3 status codes, got %d", len(stats.StatusCodes))
	}
	for i := 1; i < len(stats.StatusCodes); i++ {
		if stats.StatusCodes[i] < stats.StatusCodes[i-1] {
			t.Error("Status codes should be sorted")
		}
	}

	// Check methods are sorted
	if len(stats.Methods) != 2 {
		t.Errorf("Expected 2 methods, got %d", len(stats.Methods))
	}
	if stats.Methods[0] > stats.Methods[1] {
		t.Error("Methods should be sorted")
	}
}

func TestIndexNilHar(t *testing.T) {
	var h *Har
	idx := h.BuildIndex()

	if idx == nil {
		t.Fatal("Expected non-nil index even for nil Har")
	}

	if idx.Size() != 0 {
		t.Errorf("Expected 0 size for nil Har index, got %d", idx.Size())
	}
}

func TestIndexByURLPatternAll(t *testing.T) {
	h := createHarForIndex()
	idx := h.BuildIndex()

	// Match all HTTPS URLs
	entries := idx.ByURLPattern(`^https://`)
	if len(entries) != 5 {
		t.Errorf("Expected 5 entries matching https:// pattern, got %d", len(entries))
	}
}
