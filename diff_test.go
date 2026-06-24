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

func TestHarDiffNilReceiver(t *testing.T) {
	var diff *HarDiff

	assertDoesNotPanic(t, func() {
		if diff.HasChanges() {
			t.Error("Expected nil diff to have no changes")
		}
	})

	assertDoesNotPanic(t, func() {
		if total := diff.TotalChanges(); total != 0 {
			t.Errorf("Expected nil diff to have 0 total changes, got %d", total)
		}
	})

	for _, format := range []ConvertFormat{FormatText, FormatMarkdown, FormatCSV, ConvertFormat("unknown")} {
		assertDoesNotPanic(t, func() {
			report := diff.Report(format)
			if report == "" {
				t.Errorf("Expected non-empty report for format %q", format)
			}
		})
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

// --- Comprehensive branch coverage tests for diff.go ---

func TestDiffEntryKeyNormalizeURL(t *testing.T) {
	// Cover the NormalizeURL=true branch in entryKey
	h1 := NewHar()
	e1 := h1.AddEntry("GET", "https://example.com/api?b=2&a=1", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")

	h2 := NewHar()
	e2 := h2.AddEntry("GET", "https://example.com/api?a=1&b=2", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")

	opts := DefaultDiffOptions()
	opts.NormalizeURL = true

	diff := Diff(h1, h2, opts)
	// With URL normalization, these should be considered the same entry
	if diff.HasChanges() {
		t.Errorf("Expected no changes with NormalizeURL=true, got %d changes", diff.TotalChanges())
	}
	if diff.Unchanged != 1 {
		t.Errorf("Expected 1 unchanged entry, got %d", diff.Unchanged)
	}
}

func TestDiffEntryKeyNoNormalizeURL(t *testing.T) {
	// Cover the NormalizeURL=false branch in entryKey (default)
	// Different query param order => different key => added+removed
	h1 := NewHar()
	e1 := h1.AddEntry("GET", "https://example.com/api?b=2&a=1", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")

	h2 := NewHar()
	e2 := h2.AddEntry("GET", "https://example.com/api?a=1&b=2", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")

	opts := DefaultDiffOptions()
	opts.NormalizeURL = false

	diff := Diff(h1, h2, opts)
	// Without URL normalization, different query order means different keys
	if !diff.HasChanges() {
		t.Error("Expected changes without NormalizeURL when query param order differs")
	}
}

func TestDiffBuildEntryMapDuplicateKeys(t *testing.T) {
	// Cover the duplicate key handling in buildEntryMap
	// When two entries have the same method+URL, the second gets a suffix
	h1 := NewHar()
	e1 := h1.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e2 := h1.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e2.SetResponseStatus(404, "Not Found")

	h2 := NewHar()
	e3 := h2.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e3.SetResponseStatus(200, "OK")
	e4 := h2.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e4.SetResponseStatus(404, "Not Found")

	diff := Diff(h1, h2, DefaultDiffOptions())
	// Both duplicate entries should be matched properly
	if diff.HasChanges() {
		t.Errorf("Expected no changes for identical HAR files with duplicate keys, got %d changes", diff.TotalChanges())
	}
	if diff.Unchanged != 2 {
		t.Errorf("Expected 2 unchanged entries, got %d", diff.Unchanged)
	}
}

func TestDiffCompareEntriesIgnoreTimingsFalse(t *testing.T) {
	// Cover the IgnoreTimings=false branch in compareEntries
	h1 := NewHar()
	e1 := h1.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e1.Time = 100.0

	h2 := NewHar()
	e2 := h2.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")
	e2.Time = 200.0

	opts := DefaultDiffOptions()
	opts.IgnoreTimings = false

	diff := Diff(h1, h2, opts)
	if len(diff.Modified) == 0 {
		t.Error("Expected modified entries when timings differ and IgnoreTimings=false")
	}

	found := false
	for _, m := range diff.Modified {
		for _, c := range m.Changes {
			if c.Field == "time" {
				found = true
				if c.OldValue != 100.0 || c.NewValue != 200.0 {
					t.Errorf("Expected time change 100 -> 200, got %v -> %v", c.OldValue, c.NewValue)
				}
			}
		}
	}
	if !found {
		t.Error("Expected 'time' field change in modified entries")
	}
}

func TestDiffCompareEntriesStatusTextChange(t *testing.T) {
	// Cover statusText comparison branch
	h1 := NewHar()
	e1 := h1.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")

	h2 := NewHar()
	e2 := h2.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "Custom Status")

	diff := Diff(h1, h2, DefaultDiffOptions())
	found := false
	for _, m := range diff.Modified {
		for _, c := range m.Changes {
			if c.Field == "response.statusText" {
				found = true
			}
		}
	}
	if !found {
		t.Error("Expected 'response.statusText' field change")
	}
}

func TestDiffCompareEntriesContentTypeChange(t *testing.T) {
	// Cover response.content.mimeType comparison branch
	h1 := NewHar()
	e1 := h1.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e1.SetResponseContent(1024, "application/json")

	h2 := NewHar()
	e2 := h2.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")
	e2.SetResponseContent(1024, "text/html")

	diff := Diff(h1, h2, DefaultDiffOptions())
	found := false
	for _, m := range diff.Modified {
		for _, c := range m.Changes {
			if c.Field == "response.content.mimeType" {
				found = true
				if c.OldValue != "application/json" || c.NewValue != "text/html" {
					t.Errorf("Expected mimeType change application/json -> text/html, got %v -> %v", c.OldValue, c.NewValue)
				}
			}
		}
	}
	if !found {
		t.Error("Expected 'response.content.mimeType' field change")
	}
}

func TestDiffCompareEntriesContentSizeChange(t *testing.T) {
	// Cover response.content.size comparison branch
	h1 := NewHar()
	e1 := h1.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e1.SetResponseContent(1024, "application/json")

	h2 := NewHar()
	e2 := h2.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")
	e2.SetResponseContent(2048, "application/json")

	diff := Diff(h1, h2, DefaultDiffOptions())
	found := false
	for _, m := range diff.Modified {
		for _, c := range m.Changes {
			if c.Field == "response.content.size" {
				found = true
				if c.OldValue != 1024 || c.NewValue != 2048 {
					t.Errorf("Expected size change 1024 -> 2048, got %v -> %v", c.OldValue, c.NewValue)
				}
			}
		}
	}
	if !found {
		t.Error("Expected 'response.content.size' field change")
	}
}

func TestDiffCompareEntriesIncludeBody(t *testing.T) {
	// Cover the IncludeBody=true branch in compareEntries
	h1 := NewHar()
	e1 := h1.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e1.Response.Content.Text = `{"version":1}`

	h2 := NewHar()
	e2 := h2.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")
	e2.Response.Content.Text = `{"version":2}`

	opts := DefaultDiffOptions()
	opts.IncludeBody = true

	diff := Diff(h1, h2, opts)
	found := false
	for _, m := range diff.Modified {
		for _, c := range m.Changes {
			if c.Field == "response.content.text" {
				found = true
			}
		}
	}
	if !found {
		t.Error("Expected 'response.content.text' field change when IncludeBody=true")
	}
}

func TestDiffCompareEntriesIncludeBodySameText(t *testing.T) {
	// Cover the IncludeBody=true branch where body text is the same (no change)
	h1 := NewHar()
	e1 := h1.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e1.Response.Content.Text = `{"same":"body"}`

	h2 := NewHar()
	e2 := h2.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")
	e2.Response.Content.Text = `{"same":"body"}`

	opts := DefaultDiffOptions()
	opts.IncludeBody = true

	diff := Diff(h1, h2, opts)
	if diff.HasChanges() {
		t.Error("Expected no changes when bodies are identical")
	}
}

func TestDiffCompareHeadersAddedHeader(t *testing.T) {
	// Cover the "header added" branch in compareHeaders (header in map2 but not map1)
	h1 := NewHar()
	e1 := h1.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e1.AddRequestHeader("Accept", "application/json")

	h2 := NewHar()
	e2 := h2.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")
	e2.AddRequestHeader("Accept", "application/json")
	e2.AddRequestHeader("X-Custom", "value")

	diff := Diff(h1, h2, DefaultDiffOptions())
	found := false
	for _, m := range diff.Modified {
		for _, c := range m.Changes {
			if c.Field == "request.headers.X-Custom" {
				found = true
				if c.OldValue != nil {
					t.Errorf("Expected OldValue nil for added header, got %v", c.OldValue)
				}
				if c.NewValue != "value" {
					t.Errorf("Expected NewValue 'value' for added header, got %v", c.NewValue)
				}
			}
		}
	}
	if !found {
		t.Error("Expected added header 'X-Custom' in changes")
	}
}

func TestDiffCompareHeadersRemovedHeader(t *testing.T) {
	// Cover the "header removed" branch in compareHeaders (header in map1 but not map2)
	h1 := NewHar()
	e1 := h1.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e1.AddRequestHeader("Accept", "application/json")
	e1.AddRequestHeader("X-Debug", "true")

	h2 := NewHar()
	e2 := h2.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")
	e2.AddRequestHeader("Accept", "application/json")

	diff := Diff(h1, h2, DefaultDiffOptions())
	found := false
	for _, m := range diff.Modified {
		for _, c := range m.Changes {
			if c.Field == "request.headers.X-Debug" {
				found = true
				if c.OldValue != "true" {
					t.Errorf("Expected OldValue 'true' for removed header, got %v", c.OldValue)
				}
				if c.NewValue != nil {
					t.Errorf("Expected NewValue nil for removed header, got %v", c.NewValue)
				}
			}
		}
	}
	if !found {
		t.Error("Expected removed header 'X-Debug' in changes")
	}
}

func TestDiffCompareHeadersValueChanged(t *testing.T) {
	// Cover the "header value changed" branch in compareHeaders
	h1 := NewHar()
	e1 := h1.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e1.AddRequestHeader("Accept", "application/json")

	h2 := NewHar()
	e2 := h2.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")
	e2.AddRequestHeader("Accept", "text/html")

	diff := Diff(h1, h2, DefaultDiffOptions())
	found := false
	for _, m := range diff.Modified {
		for _, c := range m.Changes {
			if c.Field == "request.headers.Accept" {
				found = true
				if c.OldValue != "application/json" || c.NewValue != "text/html" {
					t.Errorf("Expected Accept change application/json -> text/html, got %v -> %v", c.OldValue, c.NewValue)
				}
			}
		}
	}
	if !found {
		t.Error("Expected changed header 'Accept' in changes")
	}
}

func TestDiffCompareHeadersResponseHeaders(t *testing.T) {
	// Cover response header comparison branch in compareEntries
	h1 := NewHar()
	e1 := h1.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e1.AddResponseHeader("X-Response", "old")

	h2 := NewHar()
	e2 := h2.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")
	e2.AddResponseHeader("X-Response", "new")

	diff := Diff(h1, h2, DefaultDiffOptions())
	found := false
	for _, m := range diff.Modified {
		for _, c := range m.Changes {
			if c.Field == "response.headers.X-Response" {
				found = true
				if c.OldValue != "old" || c.NewValue != "new" {
					t.Errorf("Expected X-Response change old -> new, got %v -> %v", c.OldValue, c.NewValue)
				}
			}
		}
	}
	if !found {
		t.Error("Expected changed response header 'X-Response' in changes")
	}
}

func TestDiffHeadersToMapWithIgnoreSet(t *testing.T) {
	// Cover the ignore set filtering in headersToMap
	h1 := NewHar()
	e1 := h1.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e1.AddRequestHeader("Authorization", "Bearer token1")
	e1.AddRequestHeader("Accept", "application/json")

	h2 := NewHar()
	e2 := h2.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")
	e2.AddRequestHeader("Authorization", "Bearer token2")
	e2.AddRequestHeader("Accept", "application/json")

	opts := DefaultDiffOptions()
	opts.IgnoreHeaders = []string{"Authorization"}

	diff := Diff(h1, h2, opts)
	// Authorization header should be ignored, so no changes expected
	if diff.HasChanges() {
		t.Error("Expected no changes when Authorization header is ignored")
	}
}

func TestDiffHeadersToMapCaseInsensitiveIgnore(t *testing.T) {
	// Cover case-insensitive matching in headersToMap ignore set
	h1 := NewHar()
	e1 := h1.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e1.AddRequestHeader("authorization", "Bearer token1")

	h2 := NewHar()
	e2 := h2.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")
	e2.AddRequestHeader("authorization", "Bearer token2")

	opts := DefaultDiffOptions()
	opts.IgnoreHeaders = []string{"Authorization"} // different case

	diff := Diff(h1, h2, opts)
	if diff.HasChanges() {
		t.Error("Expected no changes when header is ignored with different case")
	}
}

func TestDiffReportDefaultFormat(t *testing.T) {
	// Cover the default format branch in Report (falls through to text)
	har1 := createHarForDiff1()
	har2 := createHarForDiff2()

	diff := Diff(har1, har2, DefaultDiffOptions())

	// Use an unrecognized format to hit the default branch
	report := diff.Report(ConvertFormat("unknown"))
	if report == "" {
		t.Error("Expected non-empty report for default format fallback")
	}
	// The default fallback is text format, should contain the text report header
	if len(report) == 0 {
		t.Error("Expected non-empty text report for unknown format")
	}
}

func TestDiffReportTextWithAllSections(t *testing.T) {
	// Cover added + modified sections in text report (removed needs diff setup)
	h1 := NewHar()
	e1 := h1.AddEntry("GET", "https://example.com/old", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e1.AddRequestHeader("X-Old", "value")

	h2 := NewHar()
	e2 := h2.AddEntry("GET", "https://example.com/new", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")
	e3 := h2.AddEntry("GET", "https://example.com/old", "HTTP/1.1", "")
	e3.SetResponseStatus(404, "Not Found")

	diff := Diff(h1, h2, DefaultDiffOptions())
	report := diff.Report(FormatText)

	if !containsSubstring(report, "新增请求") {
		t.Error("Expected '新增请求' section in text report")
	}
	if len(diff.Removed) > 0 {
		// Only check removed section if there are removed entries
		if !containsSubstring(report, "删除请求") {
			t.Error("Expected '删除请求' section in text report")
		}
	}
	if !containsSubstring(report, "修改请求") {
		t.Error("Expected '修改请求' section in text report")
	}
}

func TestDiffReportMarkdownWithAllSections(t *testing.T) {
	// Cover all three sections in markdown report
	h1 := NewHar()
	e1 := h1.AddEntry("GET", "https://example.com/old", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")

	h2 := NewHar()
	e2 := h2.AddEntry("GET", "https://example.com/new", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")
	e3 := h2.AddEntry("GET", "https://example.com/old", "HTTP/1.1", "")
	e3.SetResponseStatus(404, "Not Found")

	diff := Diff(h1, h2, DefaultDiffOptions())
	report := diff.Report(FormatMarkdown)

	if !containsSubstring(report, "## 新增请求") {
		t.Error("Expected '## 新增请求' section in markdown report")
	}
	if len(diff.Removed) > 0 {
		if !containsSubstring(report, "## 删除请求") {
			t.Error("Expected '## 删除请求' section in markdown report")
		}
	}
}

func TestDiffReportCSV(t *testing.T) {
	// Cover CSV report with added, removed, and modified entries
	h1 := NewHar()
	e1 := h1.AddEntry("GET", "https://example.com/old", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")

	h2 := NewHar()
	e2 := h2.AddEntry("GET", "https://example.com/new", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")
	e3 := h2.AddEntry("GET", "https://example.com/old", "HTTP/1.1", "")
	e3.SetResponseStatus(404, "Not Found")

	diff := Diff(h1, h2, DefaultDiffOptions())
	report := diff.Report(FormatCSV)

	if !containsSubstring(report, "added,") {
		t.Error("Expected 'added,' row in CSV report")
	}
	if len(diff.Removed) > 0 {
		if !containsSubstring(report, "removed,") {
			t.Error("Expected 'removed,' row in CSV report")
		}
	}
	if !containsSubstring(report, "modified,") {
		t.Error("Expected 'modified,' row in CSV report")
	}
}

func TestDiffReportTextEmpty(t *testing.T) {
	// Cover text report when there are no changes at all
	h1 := NewHar()
	h2 := NewHar()

	diff := Diff(h1, h2, DefaultDiffOptions())
	report := diff.Report(FormatText)
	if report == "" {
		t.Error("Expected non-empty report even for empty diff")
	}
}

func TestDiffReportMarkdownEmpty(t *testing.T) {
	// Cover markdown report with no added/removed/modified sections
	h1 := NewHar()
	h2 := NewHar()

	diff := Diff(h1, h2, DefaultDiffOptions())
	report := diff.Report(FormatMarkdown)
	if report == "" {
		t.Error("Expected non-empty report even for empty diff")
	}
}

func TestDiffCompareByURL(t *testing.T) {
	// Cover CompareByURL option - entries matched by URL instead of index
	h1 := NewHar()
	e1 := h1.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e2 := h1.AddEntry("GET", "https://example.com/other", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")

	h2 := NewHar()
	e3 := h2.AddEntry("GET", "https://example.com/other", "HTTP/1.1", "")
	e3.SetResponseStatus(200, "OK")
	e4 := h2.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e4.SetResponseStatus(404, "Not Found")

	opts := DefaultDiffOptions()
	opts.CompareByURL = true

	diff := Diff(h1, h2, opts)
	// /api should be modified (200 -> 404), /other should be unchanged
	if len(diff.Modified) == 0 {
		t.Error("Expected modified entries when CompareByURL=true")
	}
}

func TestDiffNilHarWithEntries(t *testing.T) {
	// Cover nil har1 with entries in har2 (Added entries get Index)
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")

	diff := Diff(nil, h, DefaultDiffOptions())
	if len(diff.Added) != 1 {
		t.Fatalf("Expected 1 added entry, got %d", len(diff.Added))
	}
	if diff.Added[0].Index != 0 {
		t.Errorf("Expected index 0, got %d", diff.Added[0].Index)
	}

	// Cover nil har2 with entries in har1 (Removed entries get Index)
	diff2 := Diff(h, nil, DefaultDiffOptions())
	if len(diff2.Removed) != 1 {
		t.Fatalf("Expected 1 removed entry, got %d", len(diff2.Removed))
	}
	if diff2.Removed[0].Index != 0 {
		t.Errorf("Expected index 0, got %d", diff2.Removed[0].Index)
	}
}

// Helper to check substring presence
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsSubstr(s, substr)))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
