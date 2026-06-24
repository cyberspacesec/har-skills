package har

import (
	"fmt"
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

// ========== Additional coverage tests ==========

func TestSplitByDomainNilHar(t *testing.T) {
	var h *Har
	result := h.SplitByDomain()
	if len(result) != 0 {
		t.Errorf("Expected empty map for nil HAR, got %d entries", len(result))
	}
}

func TestSplitByDomainEmptyDomain(t *testing.T) {
	h := NewHar()
	// Add entry with invalid/empty URL that will yield empty domain from extractDomain
	e := h.AddEntry("GET", "", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")

	result := h.SplitByDomain()
	// extractDomain("") returns "" since url.Parse("") gives Host=""
	if _, ok := result[""]; !ok {
		t.Error("Expected empty-string domain key for entry with empty URL")
	}
}

func TestSplitByDomainMixedDomains(t *testing.T) {
	h := NewHar()
	e1 := h.AddEntry("GET", "https://api.example.com/v1/users", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e2 := h.AddEntry("GET", "https://cdn.example.com/style.css", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")
	e3 := h.AddEntry("POST", "https://api.example.com/v1/data", "HTTP/1.1", "")
	e3.SetResponseStatus(201, "Created")

	result := h.SplitByDomain()
	if len(result) != 2 {
		t.Errorf("Expected 2 domains, got %d", len(result))
	}
	if len(result["api.example.com"].Log.Entries) != 2 {
		t.Errorf("Expected 2 entries for api.example.com, got %d", len(result["api.example.com"].Log.Entries))
	}
	if len(result["cdn.example.com"].Log.Entries) != 1 {
		t.Errorf("Expected 1 entry for cdn.example.com, got %d", len(result["cdn.example.com"].Log.Entries))
	}
}

func TestSplitByDomainPreservesMetadata(t *testing.T) {
	h := NewHar()
	h.SetCreator("test-creator", "2.0")
	h.Log.Version = "1.2"
	e1 := h.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")

	result := h.SplitByDomain()
	if result["example.com"].Log.Version != "1.2" {
		t.Errorf("Expected version '1.2', got '%s'", result["example.com"].Log.Version)
	}
	if result["example.com"].Log.Creator.Name != "test-creator" {
		t.Errorf("Expected creator name 'test-creator', got '%s'", result["example.com"].Log.Creator.Name)
	}
}

func TestSplitByTimeRangeNilHar(t *testing.T) {
	var h *Har
	result := h.SplitByTimeRange(1 * time.Hour)
	if result != nil {
		t.Errorf("Expected nil for nil HAR, got %v", result)
	}
}

func TestSplitByTimeRangeEmptyEntries(t *testing.T) {
	h := NewHar()
	result := h.SplitByTimeRange(1 * time.Hour)
	if result != nil {
		t.Errorf("Expected nil for empty entries, got %v", result)
	}
}

func TestSplitByTimeRangeAllInOneInterval(t *testing.T) {
	h := NewHar()
	e1 := h.AddEntry("GET", "https://example.com/1", "HTTP/1.1", "")
	e1.StartedDateTime = time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	e2 := h.AddEntry("GET", "https://example.com/2", "HTTP/1.1", "")
	e2.StartedDateTime = time.Date(2024, 1, 1, 10, 0, 30, 0, time.UTC)
	e3 := h.AddEntry("GET", "https://example.com/3", "HTTP/1.1", "")
	e3.StartedDateTime = time.Date(2024, 1, 1, 10, 0, 59, 0, time.UTC)

	// All entries within 1 minute, using 1-hour interval
	result := h.SplitByTimeRange(1 * time.Hour)
	if len(result) != 1 {
		t.Errorf("Expected 1 group (all fit in one interval), got %d", len(result))
	}
	if len(result[0].Log.Entries) != 3 {
		t.Errorf("Expected 3 entries in the single group, got %d", len(result[0].Log.Entries))
	}
}

func TestSplitByTimeRangePreservesMetadata(t *testing.T) {
	h := NewHar()
	h.SetCreator("test-creator", "1.0")
	h.Log.Version = "1.1"
	e1 := h.AddEntry("GET", "https://example.com/1", "HTTP/1.1", "")
	e1.StartedDateTime = time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	e2 := h.AddEntry("GET", "https://example.com/2", "HTTP/1.1", "")
	e2.StartedDateTime = time.Date(2024, 1, 1, 11, 30, 0, 0, time.UTC)

	result := h.SplitByTimeRange(1 * time.Hour)
	if result[0].Log.Version != "1.1" {
		t.Errorf("Expected version '1.1', got '%s'", result[0].Log.Version)
	}
	if result[0].Log.Creator.Name != "test-creator" {
		t.Errorf("Expected creator 'test-creator', got '%s'", result[0].Log.Creator.Name)
	}
}

func TestSplitByTimeRangeMultipleIntervals(t *testing.T) {
	h := NewHar()
	for i := 0; i < 6; i++ {
		e := h.AddEntry("GET", fmt.Sprintf("https://example.com/%d", i), "HTTP/1.1", "")
		e.StartedDateTime = time.Date(2024, 1, 1, 10, i*30, 0, 0, time.UTC) // 30 min intervals
	}

	// Split by 1-hour intervals
	result := h.SplitByTimeRange(1 * time.Hour)
	// Entries at 10:00, 10:30 -> group1
	// Entries at 11:00, 11:30 -> group2
	// Entries at 12:00, 12:30 -> group3
	if len(result) != 3 {
		t.Errorf("Expected 3 groups, got %d", len(result))
	}
	for i, group := range result {
		if len(group.Log.Entries) != 2 {
			t.Errorf("Expected 2 entries in group %d, got %d", i, len(group.Log.Entries))
		}
	}
}

func TestSplitBySizeNilHar(t *testing.T) {
	var h *Har
	result := h.SplitBySize(10)
	if result != nil {
		t.Errorf("Expected nil for nil HAR, got %v", result)
	}
}

func TestSplitBySizeEmptyEntries(t *testing.T) {
	h := NewHar()
	result := h.SplitBySize(5)
	if result != nil {
		t.Errorf("Expected nil for empty entries, got %v", result)
	}
}

func TestSplitBySizeNegativeMaxEntries(t *testing.T) {
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/1", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	result := h.SplitBySize(-1)
	if result != nil {
		t.Errorf("Expected nil for negative maxEntries, got %v", result)
	}
}

func TestSplitBySizeExactFit(t *testing.T) {
	h := NewHar()
	for i := 0; i < 6; i++ {
		e := h.AddEntry("GET", fmt.Sprintf("https://example.com/%d", i), "HTTP/1.1", "")
		e.SetResponseStatus(200, "OK")
	}

	// 6 entries, max 3 per group => exactly 2 groups of 3
	result := h.SplitBySize(3)
	if len(result) != 2 {
		t.Errorf("Expected 2 groups for exact fit, got %d", len(result))
	}
	if len(result[0].Log.Entries) != 3 {
		t.Errorf("Expected 3 entries in first group, got %d", len(result[0].Log.Entries))
	}
	if len(result[1].Log.Entries) != 3 {
		t.Errorf("Expected 3 entries in second group, got %d", len(result[1].Log.Entries))
	}
}

func TestSplitBySizePreservesMetadata(t *testing.T) {
	h := NewHar()
	h.SetCreator("test-creator", "3.0")
	h.Log.Version = "1.3"
	for i := 0; i < 3; i++ {
		e := h.AddEntry("GET", fmt.Sprintf("https://example.com/%d", i), "HTTP/1.1", "")
		e.SetResponseStatus(200, "OK")
	}

	result := h.SplitBySize(2)
	if result[0].Log.Version != "1.3" {
		t.Errorf("Expected version '1.3', got '%s'", result[0].Log.Version)
	}
	if result[0].Log.Creator.Name != "test-creator" {
		t.Errorf("Expected creator 'test-creator', got '%s'", result[0].Log.Creator.Name)
	}
}

func TestSplitByStatusCodeNilHar(t *testing.T) {
	var h *Har
	result := h.SplitByStatusCode()
	if len(result) != 0 {
		t.Errorf("Expected empty map for nil HAR, got %d entries", len(result))
	}
}

func TestSplitByStatusCodeUnusualStatusCodes(t *testing.T) {
	h := NewHar()

	// Status code 100 (1xx - falls into default case)
	e1 := h.AddEntry("GET", "https://example.com/continue", "HTTP/1.1", "")
	e1.SetResponseStatus(100, "Continue")

	// Status code 600 (falls into default case)
	e2 := h.AddEntry("GET", "https://example.com/weird", "HTTP/1.1", "")
	e2.SetResponseStatus(600, "Unknown")

	result := h.SplitByStatusCode()

	if _, ok := result["1xx"]; !ok {
		t.Error("Expected '1xx' group for status 100")
	}
	if _, ok := result["6xx"]; !ok {
		t.Error("Expected '6xx' group for status 600")
	}
	if len(result["1xx"].Log.Entries) != 1 {
		t.Errorf("Expected 1 entry in '1xx' group, got %d", len(result["1xx"].Log.Entries))
	}
	if len(result["6xx"].Log.Entries) != 1 {
		t.Errorf("Expected 1 entry in '6xx' group, got %d", len(result["6xx"].Log.Entries))
	}
}

func TestSplitByStatusCodeMixedGroups(t *testing.T) {
	h := NewHar()
	e1 := h.AddEntry("GET", "https://example.com/ok1", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e2 := h.AddEntry("GET", "https://example.com/ok2", "HTTP/1.1", "")
	e2.SetResponseStatus(204, "No Content")
	e3 := h.AddEntry("GET", "https://example.com/redirect", "HTTP/1.1", "")
	e3.SetResponseStatus(301, "Moved")
	e4 := h.AddEntry("GET", "https://example.com/notfound", "HTTP/1.1", "")
	e4.SetResponseStatus(404, "Not Found")
	e5 := h.AddEntry("GET", "https://example.com/error", "HTTP/1.1", "")
	e5.SetResponseStatus(500, "Error")

	result := h.SplitByStatusCode()
	if len(result) != 4 {
		t.Errorf("Expected 4 groups, got %d", len(result))
	}
	if len(result["2xx"].Log.Entries) != 2 {
		t.Errorf("Expected 2 entries in 2xx group, got %d", len(result["2xx"].Log.Entries))
	}
	if len(result["3xx"].Log.Entries) != 1 {
		t.Errorf("Expected 1 entry in 3xx group, got %d", len(result["3xx"].Log.Entries))
	}
	if len(result["4xx"].Log.Entries) != 1 {
		t.Errorf("Expected 1 entry in 4xx group, got %d", len(result["4xx"].Log.Entries))
	}
	if len(result["5xx"].Log.Entries) != 1 {
		t.Errorf("Expected 1 entry in 5xx group, got %d", len(result["5xx"].Log.Entries))
	}
}

func TestSplitByStatusCodePreservesMetadata(t *testing.T) {
	h := NewHar()
	h.SetCreator("test-creator", "4.0")
	h.Log.Version = "1.4"
	e1 := h.AddEntry("GET", "https://example.com/ok", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")

	result := h.SplitByStatusCode()
	if result["2xx"].Log.Version != "1.4" {
		t.Errorf("Expected version '1.4', got '%s'", result["2xx"].Log.Version)
	}
	if result["2xx"].Log.Creator.Name != "test-creator" {
		t.Errorf("Expected creator 'test-creator', got '%s'", result["2xx"].Log.Creator.Name)
	}
}

func TestSplitByMethodNilHar(t *testing.T) {
	var h *Har
	result := h.SplitByMethod()
	if len(result) != 0 {
		t.Errorf("Expected empty map for nil HAR, got %d entries", len(result))
	}
}

func TestSplitByMethodMultipleMethods(t *testing.T) {
	h := NewHar()

	e1 := h.AddEntry("GET", "https://example.com/1", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e2 := h.AddEntry("POST", "https://example.com/2", "HTTP/1.1", "")
	e2.SetResponseStatus(201, "Created")
	e3 := h.AddEntry("PUT", "https://example.com/3", "HTTP/1.1", "")
	e3.SetResponseStatus(200, "OK")
	e4 := h.AddEntry("DELETE", "https://example.com/4", "HTTP/1.1", "")
	e4.SetResponseStatus(204, "No Content")
	e5 := h.AddEntry("GET", "https://example.com/5", "HTTP/1.1", "")
	e5.SetResponseStatus(200, "OK")

	result := h.SplitByMethod()
	if len(result) != 4 {
		t.Errorf("Expected 4 method groups, got %d", len(result))
	}
	if len(result["GET"].Log.Entries) != 2 {
		t.Errorf("Expected 2 GET entries, got %d", len(result["GET"].Log.Entries))
	}
	if len(result["POST"].Log.Entries) != 1 {
		t.Errorf("Expected 1 POST entry, got %d", len(result["POST"].Log.Entries))
	}
	if len(result["PUT"].Log.Entries) != 1 {
		t.Errorf("Expected 1 PUT entry, got %d", len(result["PUT"].Log.Entries))
	}
	if len(result["DELETE"].Log.Entries) != 1 {
		t.Errorf("Expected 1 DELETE entry, got %d", len(result["DELETE"].Log.Entries))
	}
}

func TestSplitByMethodPreservesMetadata(t *testing.T) {
	h := NewHar()
	h.SetCreator("test-creator", "5.0")
	h.Log.Version = "1.5"
	e1 := h.AddEntry("GET", "https://example.com/1", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")

	result := h.SplitByMethod()
	if result["GET"].Log.Version != "1.5" {
		t.Errorf("Expected version '1.5', got '%s'", result["GET"].Log.Version)
	}
	if result["GET"].Log.Creator.Name != "test-creator" {
		t.Errorf("Expected creator 'test-creator', got '%s'", result["GET"].Log.Creator.Name)
	}
}

func TestMergeWithOptionsNoSort(t *testing.T) {
	har1 := createHarForMerge("a")
	har2 := createHarForMerge("b")

	opts := MergeOptions{SortByTime: false, Deduplicate: false}
	result := MergeWithOptions(opts, har1, har2)
	if len(result.Log.Entries) != 4 {
		t.Errorf("Expected 4 entries, got %d", len(result.Log.Entries))
	}
}

func TestMergeWithOptionsFirstHarNil(t *testing.T) {
	// When nil is the first element, it should be skipped
	har1 := createHarForMerge("a")
	result := MergeWithOptions(DefaultMergeOptions(), nil, har1)

	if len(result.Log.Entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(result.Log.Entries))
	}
	if result.Log.Creator.Name != har1.Log.Creator.Name {
		t.Errorf("Expected metadata from first non-nil HAR, got creator %q", result.Log.Creator.Name)
	}
}

func TestMergeWithNilFirstHar(t *testing.T) {
	// Test that when nil is the first argument, metadata is copied from the first non-nil HAR
	har1 := createHarForMerge("a")
	result := Merge(nil, har1)
	if len(result.Log.Entries) != 2 {
		t.Errorf("Expected 2 entries (nil skipped), got %d", len(result.Log.Entries))
	}
	if result.Log.Creator.Name != har1.Log.Creator.Name {
		t.Errorf("Expected metadata from first non-nil HAR, got creator %q", result.Log.Creator.Name)
	}
}

func TestSplitByPageNoPageref(t *testing.T) {
	h := NewHar()
	e1 := h.AddEntry("GET", "https://example.com/orphan", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	// Pageref is empty string by default

	result := h.SplitByPage()
	if _, ok := result[""]; !ok {
		t.Error("Expected empty-string key for entries with no pageref")
	}
	if len(result[""].Log.Entries) != 1 {
		t.Errorf("Expected 1 entry in empty-string group, got %d", len(result[""].Log.Entries))
	}
}

func TestDeduplicateEntriesOlderWins(t *testing.T) {
	har1 := NewHar()
	e1 := har1.AddEntry("GET", "https://example.com/test", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e1.StartedDateTime = time.Date(2024, 1, 1, 10, 1, 0, 0, time.UTC) // newer

	har2 := NewHar()
	e2 := har2.AddEntry("GET", "https://example.com/test", "HTTP/1.1", "")
	e2.SetResponseStatus(404, "Not Found")
	e2.StartedDateTime = time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC) // older

	opts := MergeOptions{SortByTime: false, Deduplicate: true}
	result := MergeWithOptions(opts, har1, har2)

	if len(result.Log.Entries) != 1 {
		t.Errorf("Expected 1 entry after dedup, got %d", len(result.Log.Entries))
	}
	// The newer entry (status 200) should be kept
	if result.Log.Entries[0].Response.Status != 200 {
		t.Errorf("Expected status 200 (newer entry kept), got %d", result.Log.Entries[0].Response.Status)
	}
}
