package har

import (
	"testing"
)

func createHarForDiff1() *Har {
	h := NewHar()
	h.SetCreator("test", "1.0")

	e1 := h.AddEntry("GET", "https://example.com/api/users", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e1.SetResponseContent(1024, "application/json")

	e2 := h.AddEntry("POST", "https://example.com/api/users", "HTTP/1.1", "")
	e2.SetResponseStatus(201, "Created")
	e2.SetResponseContent(512, "application/json")

	e3 := h.AddEntry("GET", "https://example.com/static/style.css", "HTTP/1.1", "")
	e3.SetResponseStatus(200, "OK")
	e3.SetResponseContent(2048, "text/css")

	return h
}

func createHarForDiff2() *Har {
	h := NewHar()
	h.SetCreator("test", "1.0")

	// 相同的请求
	e1 := h.AddEntry("GET", "https://example.com/api/users", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e1.SetResponseContent(1024, "application/json")

	// 修改的请求：状态码从201变为200
	e2 := h.AddEntry("POST", "https://example.com/api/users", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")
	e2.SetResponseContent(512, "application/json")

	// 删除了 style.css，新增了 script.js
	e4 := h.AddEntry("GET", "https://example.com/static/script.js", "HTTP/1.1", "")
	e4.SetResponseStatus(200, "OK")
	e4.SetResponseContent(4096, "application/javascript")

	return h
}

func TestDiff(t *testing.T) {
	har1 := createHarForDiff1()
	har2 := createHarForDiff2()

	diff := Diff(har1, har2, DefaultDiffOptions())

	if !diff.HasChanges() {
		t.Error("Expected changes between the two HAR files")
	}

	if len(diff.Modified) == 0 {
		t.Error("Expected modified entries")
	}

	// POST /api/users should be modified (status 201 -> 200)
	found := false
	for _, m := range diff.Modified {
		if m.Method == "POST" && m.URL == "https://example.com/api/users" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected POST /api/users to be modified")
	}
}

func TestDiffNil(t *testing.T) {
	har1 := createHarForDiff1()

	// Both nil
	diff := Diff(nil, nil, DefaultDiffOptions())
	if diff.HasChanges() {
		t.Error("Expected no changes for nil vs nil")
	}

	// First nil
	diff = Diff(nil, har1, DefaultDiffOptions())
	if len(diff.Added) != 3 {
		t.Errorf("Expected 3 added entries, got %d", len(diff.Added))
	}

	// Second nil
	diff = Diff(har1, nil, DefaultDiffOptions())
	if len(diff.Removed) != 3 {
		t.Errorf("Expected 3 removed entries, got %d", len(diff.Removed))
	}
}

func TestDiffIdentical(t *testing.T) {
	har1 := createHarForDiff1()
	har2 := createHarForDiff1()

	diff := Diff(har1, har2, DefaultDiffOptions())

	if diff.HasChanges() {
		t.Error("Expected no changes for identical HAR files")
	}

	if diff.Unchanged != 3 {
		t.Errorf("Expected 3 unchanged entries, got %d", diff.Unchanged)
	}
}

func TestDiffReport(t *testing.T) {
	har1 := createHarForDiff1()
	har2 := createHarForDiff2()

	diff := Diff(har1, har2, DefaultDiffOptions())

	// Test text report
	textReport := diff.Report(FormatText)
	if textReport == "" {
		t.Error("Expected non-empty text report")
	}

	// Test markdown report
	mdReport := diff.Report(FormatMarkdown)
	if mdReport == "" {
		t.Error("Expected non-empty markdown report")
	}

	// Test CSV report
	csvReport := diff.Report(FormatCSV)
	if csvReport == "" {
		t.Error("Expected non-empty CSV report")
	}
}

func TestDiffTotalChanges(t *testing.T) {
	har1 := createHarForDiff1()
	har2 := createHarForDiff2()

	diff := Diff(har1, har2, DefaultDiffOptions())
	total := diff.TotalChanges()

	if total != len(diff.Added)+len(diff.Removed)+len(diff.Modified) {
		t.Errorf("TotalChanges should equal sum of Added, Removed, Modified")
	}
}

func TestDiffWithOptions(t *testing.T) {
	har1 := createHarForDiff1()
	har2 := createHarForDiff2()

	// With include body comparison
	opts := DefaultDiffOptions()
	opts.IncludeBody = true

	diff := Diff(har1, har2, opts)
	if diff == nil {
		t.Error("Expected non-nil diff result")
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://example.com/api?b=2&a=1", "https://example.com/api?a=1&b=2"},
		{"https://example.com/api", "https://example.com/api"},
		{"invalid-url", "invalid-url"},
	}

	for _, tt := range tests {
		result := normalizeURL(tt.input, nil)
		if tt.input == "invalid-url" {
			// For invalid URLs, just check it doesn't crash
			continue
		}
		if result != tt.expected {
			t.Errorf("normalizeURL(%s) = %s, expected %s", tt.input, result, tt.expected)
		}
	}
}
