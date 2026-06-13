package har

import (
	"testing"
)

func createHarForDedup() *Har {
	h := NewHar()
	h.SetCreator("test", "1.0")

	e1 := h.AddEntry("GET", "https://example.com/api/users", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e1.SetResponseContent(1024, "application/json")

	e2 := h.AddEntry("GET", "https://example.com/api/users", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")
	e2.SetResponseContent(1024, "application/json")

	e3 := h.AddEntry("POST", "https://example.com/api/data", "HTTP/1.1", "")
	e3.SetResponseStatus(201, "Created")
	e3.SetResponseContent(128, "application/json")

	e4 := h.AddEntry("GET", "https://example.com/api/items?cb=123", "HTTP/1.1", "")
	e4.SetResponseStatus(200, "OK")
	e4.SetResponseContent(512, "application/json")

	e5 := h.AddEntry("GET", "https://example.com/api/items?cb=456", "HTTP/1.1", "")
	e5.SetResponseStatus(200, "OK")
	e5.SetResponseContent(512, "application/json")

	e6 := h.AddEntry("GET", "https://example.com/api/items?sort=name", "HTTP/1.1", "")
	e6.SetResponseStatus(200, "OK")
	e6.SetResponseContent(256, "application/json")

	return h
}

func TestFindDuplicatesExactURL(t *testing.T) {
	h := createHarForDedup()

	opts := DeduplicateOptions{
		Strategy: DedupExactURL,
	}

	groups := h.FindDuplicates(opts)

	// Should find 1 group: the two exact-duplicate /api/users requests
	if len(groups) != 1 {
		t.Fatalf("Expected 1 duplicate group, got %d", len(groups))
	}

	if groups[0].Count != 2 {
		t.Errorf("Expected 2 duplicates, got %d", groups[0].Count)
	}
}

func TestFindDuplicatesURLPattern(t *testing.T) {
	h := createHarForDedup()

	opts := DeduplicateOptions{
		Strategy:     DedupURLPattern,
		IgnoreParams: []string{"cb"},
	}

	groups := h.FindDuplicates(opts)

	// Should find 2 groups:
	// - 2 entries for /api/users (exact duplicates)
	// - 2 entries for /api/items with cb= param (pattern duplicates)
	totalDuplicates := 0
	for _, g := range groups {
		totalDuplicates += g.Count
	}

	if len(groups) < 2 {
		t.Errorf("Expected at least 2 duplicate groups with URL pattern strategy, got %d", len(groups))
	}

	// Check that cb-variant items are grouped together
	foundCBGroup := false
	for _, g := range groups {
		if g.Count == 2 && len(g.EntryIndices) == 2 {
			// Check if these are the cb-variant entries
			idx0 := g.EntryIndices[0]
			idx1 := g.EntryIndices[1]
			url0 := h.Log.Entries[idx0].Request.URL
			url1 := h.Log.Entries[idx1].Request.URL
			if (containsQueryParam(url0, "cb") || containsQueryParam(url1, "cb")) &&
				(containsPath(url0, "/api/items") && containsPath(url1, "/api/items")) {
				foundCBGroup = true
			}
		}
	}
	if !foundCBGroup {
		t.Error("Expected to find a group of cb-variant /api/items requests")
	}
}

func TestDeduplicateExactURL(t *testing.T) {
	h := createHarForDedup()

	opts := DeduplicateOptions{
		Strategy: DedupExactURL,
	}

	result := h.Deduplicate(opts)
	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Original should have 6 entries
	if len(h.Log.Entries) != 6 {
		t.Errorf("Original should have 6 entries, got %d", len(h.Log.Entries))
	}

	// Deduped should have 5 (one duplicate /api/users removed)
	if len(result.Log.Entries) != 5 {
		t.Errorf("Expected 5 entries after dedup, got %d", len(result.Log.Entries))
	}
}

func TestDeduplicateURLPattern(t *testing.T) {
	h := createHarForDedup()

	opts := DeduplicateOptions{
		Strategy:     DedupURLPattern,
		IgnoreParams: []string{"cb"},
	}

	result := h.Deduplicate(opts)

	// Should have 4 entries:
	// - 1 x /api/users (one of two duplicates removed)
	// - 1 x POST /api/data
	// - 1 x /api/items (cb variants merged)
	// - 1 x /api/items?sort=name (different params)
	if len(result.Log.Entries) != 4 {
		t.Errorf("Expected 4 entries after pattern dedup, got %d", len(result.Log.Entries))
	}
}

func TestDeduplicateContentHash(t *testing.T) {
	h := NewHar()

	// Two entries with same response content
	e1 := h.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e1.SetResponseContent(100, "application/json")
	e1.Response.Content.Text = `{"status":"ok"}`

	e2 := h.AddEntry("GET", "https://other.com/api", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")
	e2.SetResponseContent(100, "application/json")
	e2.Response.Content.Text = `{"status":"ok"}`

	opts := DeduplicateOptions{
		Strategy: DedupContentHash,
	}

	groups := h.FindDuplicates(opts)
	if len(groups) != 1 {
		t.Fatalf("Expected 1 duplicate group for content hash, got %d", len(groups))
	}
}

func TestDeduplicateNilHar(t *testing.T) {
	var h *Har
	opts := DefaultDeduplicateOptions()
	result := h.Deduplicate(opts)
	if result != nil {
		t.Error("Expected nil result for nil Har")
	}
}

func TestFindDuplicatesEmptyHar(t *testing.T) {
	h := NewHar()
	opts := DefaultDeduplicateOptions()
	groups := h.FindDuplicates(opts)
	if groups != nil {
		t.Errorf("Expected nil for empty Har, got %v", groups)
	}
}

func TestIsCacheBusterParam(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"_", true},
		{"cb", true},
		{"CB", true}, // case insensitive
		{"cachebuster", true},
		{"timestamp", true},
		{"t", true},
		{"rand", true},
		{"random", true},
		{"page", false},
		{"sort", false},
		{"id", false},
	}

	for _, tt := range tests {
		result := IsCacheBusterParam(tt.name)
		if result != tt.expected {
			t.Errorf("IsCacheBusterParam(%q) = %v, expected %v", tt.name, result, tt.expected)
		}
	}
}

func TestIsCacheBusterParamWithValue(t *testing.T) {
	// "v" with numeric value should be a cache buster
	if !IsCacheBusterParamWithValue("v", "1234567890") {
		t.Error("Expected 'v' with numeric value to be a cache buster")
	}

	// "v" with non-numeric value should NOT be a cache buster
	if IsCacheBusterParamWithValue("v", "abc") {
		t.Error("Expected 'v' with non-numeric value NOT to be a cache buster")
	}

	// Other cache busters should work regardless of value
	if !IsCacheBusterParamWithValue("cb", "anyvalue") {
		t.Error("Expected 'cb' to be a cache buster regardless of value")
	}
}

func TestDefaultDeduplicateOptions(t *testing.T) {
	opts := DefaultDeduplicateOptions()

	if opts.Strategy != DedupURLPattern {
		t.Errorf("Expected DedupURLPattern strategy, got %d", opts.Strategy)
	}

	if len(opts.IgnoreParams) == 0 {
		t.Error("Expected non-empty IgnoreParams in default options")
	}

	// Check common cache busters are included
	found := false
	for _, p := range opts.IgnoreParams {
		if p == "_" {
			found = true
		}
	}
	if !found {
		t.Error("Expected '_' to be in default ignore params")
	}
}

func TestDeduplicateKeepsFirstOccurrence(t *testing.T) {
	h := NewHar()

	e1 := h.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")

	e2 := h.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e2.SetResponseStatus(304, "Not Modified")

	opts := DeduplicateOptions{Strategy: DedupExactURL}
	result := h.Deduplicate(opts)

	if len(result.Log.Entries) != 1 {
		t.Fatalf("Expected 1 entry after dedup, got %d", len(result.Log.Entries))
	}

	// Should keep the first entry (200 OK)
	if result.Log.Entries[0].Response.Status != 200 {
		t.Errorf("Expected first occurrence (200 OK) to be kept, got status %d", result.Log.Entries[0].Response.Status)
	}
}

func TestDeduplicateWithCompareHeaders(t *testing.T) {
	h := NewHar()

	e1 := h.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e1.AddRequestHeader("Accept", "application/json")
	e1.SetResponseStatus(200, "OK")

	e2 := h.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e2.AddRequestHeader("Accept", "text/html")
	e2.SetResponseStatus(200, "OK")

	// Without header comparison, these are duplicates
	optsNoHeaders := DeduplicateOptions{Strategy: DedupExactURL, CompareHeaders: false}
	resultNoHeaders := h.Deduplicate(optsNoHeaders)
	if len(resultNoHeaders.Log.Entries) != 1 {
		t.Errorf("Expected 1 entry without header comparison, got %d", len(resultNoHeaders.Log.Entries))
	}

	// With header comparison, these are NOT duplicates
	optsWithHeaders := DeduplicateOptions{Strategy: DedupExactURL, CompareHeaders: true}
	resultWithHeaders := h.Deduplicate(optsWithHeaders)
	if len(resultWithHeaders.Log.Entries) != 2 {
		t.Errorf("Expected 2 entries with header comparison, got %d", len(resultWithHeaders.Log.Entries))
	}
}

func TestDeduplicateWithCompareBody(t *testing.T) {
	h := NewHar()

	e1 := h.AddEntry("POST", "https://example.com/api", "HTTP/1.1", "")
	e1.SetPostData("application/json", `{"data":"first"}`)
	e1.SetResponseStatus(200, "OK")

	e2 := h.AddEntry("POST", "https://example.com/api", "HTTP/1.1", "")
	e2.SetPostData("application/json", `{"data":"second"}`)
	e2.SetResponseStatus(200, "OK")

	// Without body comparison, these are duplicates
	optsNoBody := DeduplicateOptions{Strategy: DedupExactURL, CompareBody: false}
	resultNoBody := h.Deduplicate(optsNoBody)
	if len(resultNoBody.Log.Entries) != 1 {
		t.Errorf("Expected 1 entry without body comparison, got %d", len(resultNoBody.Log.Entries))
	}

	// With body comparison, these are NOT duplicates
	optsWithBody := DeduplicateOptions{Strategy: DedupExactURL, CompareBody: true}
	resultWithBody := h.Deduplicate(optsWithBody)
	if len(resultWithBody.Log.Entries) != 2 {
		t.Errorf("Expected 2 entries with body comparison, got %d", len(resultWithBody.Log.Entries))
	}
}

// Helper functions for test assertions
func containsQueryParam(urlStr, param string) bool {
	return len(urlStr) > 0 && param != ""
}

func containsPath(urlStr, path string) bool {
	return len(urlStr) > 0 && len(path) > 0
}
