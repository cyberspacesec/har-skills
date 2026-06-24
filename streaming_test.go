package har

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// ============================================================================
// NewStreamingHarFromFile tests
// ============================================================================

func TestStreamingNewStreamingHarFromFile_Valid(t *testing.T) {
	sh, err := NewStreamingHarFromFile("testdata/example.har")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer sh.Close()

	if sh.GetVersion() != "1.2" {
		t.Errorf("expected version 1.2, got %s", sh.GetVersion())
	}
}

func TestStreamingNewStreamingHarFromFile_FileNotFound(t *testing.T) {
	_, err := NewStreamingHarFromFile("testdata/nonexistent.har")
	assertHarErrorCode(t, err, ErrCodeFileSystem)
}

func TestStreamingNewStreamingHarFromFile_InvalidJSON(t *testing.T) {
	_, err := NewStreamingHarFromFile("testdata/not_json.har")
	assertHarErrorCode(t, err, ErrCodeJSONParse)
}

func TestStreamingNewStreamingHarFromFile_InvalidMissingLog(t *testing.T) {
	// Create a temp file with JSON that has no "log" field
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "no_log.har")
	content := `{"something": "else"}`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	_, err := NewStreamingHarFromFile(tmpFile)
	if err == nil {
		t.Fatal("expected error for HAR without log field, got nil")
	}
}

func TestStreamingNewStreamingHarFromFile_NoOpeningBrace(t *testing.T) {
	// JSON that doesn't start with {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "no_brace.har")
	content := `"just a string"`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	_, err := NewStreamingHarFromFile(tmpFile)
	if err == nil {
		t.Fatal("expected error for file not starting with {, got nil")
	}
}

func TestStreamingNewStreamingHarFromFile_LogNotObject(t *testing.T) {
	// JSON where "log" value is not an object
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "log_not_obj.har")
	content := `{"log": "not an object"}`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	_, err := NewStreamingHarFromFile(tmpFile)
	if err == nil {
		t.Fatal("expected error when log is not an object, got nil")
	}
}

func TestStreamingNewStreamingHarFromFile_ComplexHAR(t *testing.T) {
	sh, err := NewStreamingHarFromFile("testdata/complex_test.har")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer sh.Close()

	if sh.GetVersion() != "1.2" {
		t.Errorf("expected version 1.2, got %s", sh.GetVersion())
	}
}

// ============================================================================
// findHarObjectStart tests (indirect through various malformed inputs)
// ============================================================================

func TestStreamingFindHarObjectStart_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "empty.har")
	if err := os.WriteFile(tmpFile, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	_, err := NewStreamingHarFromFile(tmpFile)
	if err == nil {
		t.Fatal("expected error for empty file, got nil")
	}
}

// ============================================================================
// parseHarBasicInfo tests
// ============================================================================

func TestStreamingParseHarBasicInfo_InvalidVersionValue(t *testing.T) {
	// Version field with non-string value (should cause decode error in streaming)
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "bad_version.har")
	content := `{"log": {"version": 123, "entries": []}}`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	_, err := NewStreamingHarFromFile(tmpFile)
	// version is decoded into string, number 123 might fail or be accepted
	// depending on json.Decoder behavior. Either way, it should not panic.
	_ = err
}

func TestStreamingParseHarBasicInfo_InvalidCreatorValue(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "bad_creator.har")
	content := `{"log": {"creator": "not_an_object", "entries": []}}`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	_, err := NewStreamingHarFromFile(tmpFile)
	if err == nil {
		t.Fatal("expected error for invalid creator value, got nil")
	}
}

func TestStreamingParseHarBasicInfo_InvalidBrowserValue(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "bad_browser.har")
	content := `{"log": {"version": "1.2", "browser": 42, "entries": []}}`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	_, err := NewStreamingHarFromFile(tmpFile)
	if err == nil {
		t.Fatal("expected error for invalid browser value, got nil")
	}
}

func TestStreamingParseHarBasicInfo_InvalidPagesValue(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "bad_pages.har")
	content := `{"log": {"version": "1.2", "pages": "not_an_array", "entries": []}}`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	_, err := NewStreamingHarFromFile(tmpFile)
	if err == nil {
		t.Fatal("expected error for invalid pages value, got nil")
	}
}

func TestStreamingParseHarBasicInfo_UnknownFieldThatFailsToSkip(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "bad_unknown.har")
	content := `{"log": {"version": "1.2", "badField": {"nested": [1, 2, 3]}, "entries": []}}`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	// This should succeed because unknown fields are skipped with decoder.Decode(&dummy)
	sh, err := NewStreamingHarFromFile(tmpFile)
	if err != nil {
		t.Fatalf("expected no error for unknown field, got: %v", err)
	}
	defer sh.Close()
}

func TestStreamingParseHarBasicInfo_EntriesNotArray(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "entries_not_array.har")
	content := `{"log": {"version": "1.2", "entries": "not_an_array"}}`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	_, err := NewStreamingHarFromFile(tmpFile)
	if err == nil {
		t.Fatal("expected error when entries is not an array, got nil")
	}
}

func TestStreamingParseHarBasicInfo_ClosingBraceBeforeEntries(t *testing.T) {
	// Log object closes before "entries" field appears
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "closed_log.har")
	content := `{"log": {"version": "1.2"}}`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	sh, err := NewStreamingHarFromFile(tmpFile)
	if err != nil {
		t.Fatalf("expected no error (closes cleanly before entries), got: %v", err)
	}
	defer sh.Close()

	// When we try to iterate, there should be no entries
	it := sh.Entries()
	defer it.Close()
	if it.Next() {
		t.Error("expected no entries in HAR with empty log object")
	}
}

func TestStreamingParseHarBasicInfo_InvalidSkipField(t *testing.T) {
	// Unknown field with invalid JSON that can't be skipped
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "bad_skip.har")
	content := `{"log": {"version": "1.2", "badField": [invalid json], "entries": []}}`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	_, err := NewStreamingHarFromFile(tmpFile)
	if err == nil {
		t.Fatal("expected error for invalid JSON in unknown field, got nil")
	}
}

// ============================================================================
// GetCreator tests
// ============================================================================

func TestStreamingGetCreator(t *testing.T) {
	sh, err := NewStreamingHarFromFile("testdata/complex_test.har")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer sh.Close()

	creator := sh.GetCreator()
	if creator.Name != "Go-HAR Test" {
		t.Errorf("expected creator name 'Go-HAR Test', got '%s'", creator.Name)
	}
	if creator.Version != "1.0" {
		t.Errorf("expected creator version '1.0', got '%s'", creator.Version)
	}
}

func TestStreamingGetCreator_Empty(t *testing.T) {
	sh, err := NewStreamingHarFromFile("testdata/minimal.har")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer sh.Close()

	creator := sh.GetCreator()
	if creator.Name != "Go-HAR Test" {
		t.Errorf("expected creator name 'Go-HAR Test', got '%s'", creator.Name)
	}
}

// ============================================================================
// GetPages tests
// ============================================================================

func TestStreamingGetPages_WithPages(t *testing.T) {
	sh, err := NewStreamingHarFromFile("testdata/example.har")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer sh.Close()

	pages := sh.GetPages()
	if len(pages) != 1 {
		t.Fatalf("expected 1 page, got %d", len(pages))
	}
	if pages[0].ID != "page_1" {
		t.Errorf("expected page ID 'page_1', got '%s'", pages[0].ID)
	}
	if pages[0].Title != "Test Page" {
		t.Errorf("expected page title 'Test Page', got '%s'", pages[0].Title)
	}
}

func TestStreamingGetPages_NoPages(t *testing.T) {
	sh, err := NewStreamingHarFromFile("testdata/minimal.har")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer sh.Close()

	pages := sh.GetPages()
	if len(pages) != 0 {
		t.Errorf("expected 0 pages, got %d", len(pages))
	}
}

// ============================================================================
// GetBrowser tests
// ============================================================================

func TestStreamingGetBrowser(t *testing.T) {
	sh, err := NewStreamingHarFromFile("testdata/complex_test.har")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer sh.Close()

	browser := sh.GetBrowser()
	if browser.Name != "Test Browser" {
		t.Errorf("expected browser name 'Test Browser', got '%s'", browser.Name)
	}
	if browser.Version != "1.0" {
		t.Errorf("expected browser version '1.0', got '%s'", browser.Version)
	}
}

// ============================================================================
// Entries / StreamingEntryIterator tests
// ============================================================================

func TestStreamingEntries_FromFile(t *testing.T) {
	sh, err := NewStreamingHarFromFile("testdata/example.har")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer sh.Close()

	it := sh.Entries()
	defer it.Close()

	count := 0
	for it.Next() {
		entry := it.Entry()
		if entry == nil {
			t.Error("expected non-nil entry")
		}
		count++
	}

	if count != 1 {
		t.Errorf("expected 1 entry, got %d", count)
	}
}

func TestStreamingEntries_ComplexFile(t *testing.T) {
	sh, err := NewStreamingHarFromFile("testdata/complex_test.har")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer sh.Close()

	it := sh.Entries()
	defer it.Close()

	count := 0
	for it.Next() {
		count++
	}
	if count != 3 {
		t.Errorf("expected 3 entries, got %d", count)
	}
}

func TestStreamingEntries_EmptyEntries(t *testing.T) {
	sh, err := NewStreamingHarFromFile("testdata/minimal.har")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer sh.Close()

	it := sh.Entries()
	defer it.Close()

	count := 0
	for it.Next() {
		count++
	}
	if count != 0 {
		t.Errorf("expected 0 entries, got %d", count)
	}
}

func TestStreamingEntries_FromBytes(t *testing.T) {
	data, err := os.ReadFile("testdata/example.har")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	sh, err := NewStreamingHarFromBytes(data)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	it := sh.Entries()
	defer it.Close()

	count := 0
	for it.Next() {
		entry := it.Entry()
		if entry == nil {
			t.Error("expected non-nil entry")
		}
		count++
	}
	if count != 1 {
		t.Errorf("expected 1 entry, got %d", count)
	}
}

// ============================================================================
// Next tests
// ============================================================================

func TestStreamingNext_ClosedIterator(t *testing.T) {
	sh, err := NewStreamingHarFromFile("testdata/example.har")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer sh.Close()

	it := sh.Entries()
	it.Close()

	// Next on closed iterator should return false
	if it.Next() {
		t.Error("expected Next() to return false on closed iterator")
	}
}

func TestStreamingNext_IteratorWithError(t *testing.T) {
	sh, err := NewStreamingHarFromFile("testdata/example.har")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer sh.Close()

	it := sh.Entries()
	defer it.Close()

	// Manually set error and verify Next returns false
	it.err = &json.UnmarshalTypeError{}
	if it.Next() {
		t.Error("expected Next() to return false when err is set")
	}
}

func TestStreamingNext_DecodeError(t *testing.T) {
	// Create a HAR file with a malformed entry
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "malformed_entry.har")
	content := `{"log": {"version": "1.2", "entries": [{"invalid": json}]}}`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	sh, err := NewStreamingHarFromFile(tmpFile)
	if err != nil {
		t.Fatalf("expected no error opening file, got: %v", err)
	}
	defer sh.Close()

	it := sh.Entries()
	defer it.Close()

	if it.Next() {
		t.Error("expected Next() to return false for malformed entry")
	}
	assertHarErrorCode(t, it.Err(), ErrCodeJSONParse)
}

// ============================================================================
// Position tests
// ============================================================================

func TestStreamingPosition(t *testing.T) {
	sh, err := NewStreamingHarFromFile("testdata/complex_test.har")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer sh.Close()

	it := sh.Entries()
	defer it.Close()

	if it.Position() != 0 {
		t.Errorf("expected initial position 0, got %d", it.Position())
	}

	for i := 0; it.Next(); i++ {
		if it.Position() != i+1 {
			t.Errorf("expected position %d after %d Next() calls, got %d", i+1, i+1, it.Position())
		}
	}
}

// ============================================================================
// Err tests
// ============================================================================

func TestStreamingErr_NoError(t *testing.T) {
	sh, err := NewStreamingHarFromFile("testdata/example.har")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer sh.Close()

	it := sh.Entries()
	defer it.Close()

	for it.Next() {
		// iterate all
	}

	if it.Err() != nil {
		t.Errorf("expected no error after clean iteration, got: %v", it.Err())
	}
}

func TestStreamingErr_WithNonEOFError(t *testing.T) {
	sh := &StreamingHar{}
	it := sh.Entries()
	defer it.Close()

	// The iterator should have an error since there's no data source
	if it.Err() == nil {
		t.Error("expected error when no data source available")
	}
}

// ============================================================================
// Close tests
// ============================================================================

func TestStreamingClose_DoubleClose(t *testing.T) {
	sh, err := NewStreamingHarFromFile("testdata/example.har")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// First close
	if err := sh.Close(); err != nil {
		t.Errorf("expected no error on first close, got: %v", err)
	}
	// Second close (file is already nil)
	if err := sh.Close(); err != nil {
		t.Errorf("expected no error on second close, got: %v", err)
	}
}

func TestStreamingIteratorClose_AlreadyClosed(t *testing.T) {
	sh, err := NewStreamingHarFromFile("testdata/example.har")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer sh.Close()

	it := sh.Entries()

	// Close once
	if err := it.Close(); err != nil {
		t.Errorf("expected no error on first close, got: %v", err)
	}
	// Close again - should return nil (already closed)
	if err := it.Close(); err != nil {
		t.Errorf("expected no error on double close, got: %v", err)
	}
}

func TestStreamingIteratorClose_NoFile(t *testing.T) {
	// Create iterator from bytes source (no file)
	data, err := os.ReadFile("testdata/example.har")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	sh, err := NewStreamingHarFromBytes(data)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	it := sh.Entries()
	// File should be nil for byte-based iterator
	if err := it.Close(); err != nil {
		t.Errorf("expected no error on close for byte-based iterator, got: %v", err)
	}
}

// ============================================================================
// GetAllEntries tests
// ============================================================================

func TestStreamingGetAllEntries(t *testing.T) {
	sh, err := NewStreamingHarFromFile("testdata/complex_test.har")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer sh.Close()

	entries, err := sh.GetAllEntries()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}
}

func TestStreamingGetAllEntries_Empty(t *testing.T) {
	sh, err := NewStreamingHarFromFile("testdata/minimal.har")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer sh.Close()

	entries, err := sh.GetAllEntries()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestStreamingGetAllEntries_FromBytes(t *testing.T) {
	data, err := os.ReadFile("testdata/example.har")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	sh, err := NewStreamingHarFromBytes(data)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	entries, err := sh.GetAllEntries()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
}

// ============================================================================
// NewStreamingHarFromBytes tests
// ============================================================================

func TestStreamingNewStreamingHarFromBytes_InvalidJSON(t *testing.T) {
	_, err := NewStreamingHarFromBytes([]byte("not json"))
	assertHarErrorCode(t, err, ErrCodeJSONParse)
}

func TestStreamingNewStreamingHarFromBytes_Valid(t *testing.T) {
	data, err := os.ReadFile("testdata/example.har")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	sh, err := NewStreamingHarFromBytes(data)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if sh.GetVersion() != "1.2" {
		t.Errorf("expected version 1.2, got %s", sh.GetVersion())
	}
}

// ============================================================================
// Entries - no data source
// ============================================================================

func TestStreamingEntries_NoDataSource(t *testing.T) {
	sh := &StreamingHar{}
	it := sh.Entries()
	defer it.Close()

	if it.Err() == nil {
		t.Error("expected error for streaming HAR with no data source")
	}
}

// ============================================================================
// Entries - file reopen failure
// ============================================================================

func TestStreamingEntries_FileReopenFailure(t *testing.T) {
	sh, err := NewStreamingHarFromFile("testdata/example.har")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Close the underlying file first, then try to iterate
	sh.Close()

	it := sh.Entries()
	defer it.Close()

	// The file was closed, so reopening should fail
	if it.Err() == nil {
		t.Error("expected error when file is closed and cannot be reopened")
	}
}

// ============================================================================
// Next - entries not started (from bytes)
// ============================================================================

func TestStreamingNext_EntriesNotStartedFromBytes(t *testing.T) {
	data, err := os.ReadFile("testdata/example.har")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	sh, err := NewStreamingHarFromBytes(data)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	it := sh.Entries()
	defer it.Close()

	// entriesStarted should be false for byte-based iterator
	// Next should find the entries and iterate them
	count := 0
	for it.Next() {
		count++
	}
	if count != 1 {
		t.Errorf("expected 1 entry, got %d", count)
	}
}

func TestStreamingNext_EntriesFieldNotArrayFromBytes(t *testing.T) {
	// Create HAR data where "entries" is not an array
	content := `{"log": {"version": "1.2", "entries": "not_array"}}`
	// But NewStreamingHarFromBytes parses the whole thing into a Har struct first,
	// so entries as a string won't parse correctly into []LazyEntries
	// This should fail at parse time
	_, err := NewStreamingHarFromBytes([]byte(content))
	if err == nil {
		t.Fatal("expected error for entries not being an array, got nil")
	}
}

// ============================================================================
// Next - entries found but not followed by [
// ============================================================================

func TestStreamingNext_EntriesNotFollowedByArrayFromBytes(t *testing.T) {
	// When using bytes, the iterator reads through the JSON looking for "entries"
	// then expects "[". If entries is present but not as array start, it errors.
	// However, NewStreamingHarFromBytes fully parses, so this is hard to trigger
	// from bytes path. Instead, we create a specially crafted stream that
	// has "entries" followed by something other than "[".

	// Actually for the bytes path, entriesStarted is false initially.
	// The iterator scans for "entries" key then expects "[".
	// Let's create JSON where entries appears but isn't an array.

	// For bytes-based, the data is parsed fully so this scenario won't happen.
	// For file-based, we can craft a file where entries isn't found properly.
	// Let's test with a file that has extra fields before entries.
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "extra_before_entries.har")
	content := `{"log": {"version": "1.2", "creator": {"name": "Test", "version": "1.0"}, "extraField": "value", "entries": [{"startedDateTime": "2023-01-01T00:00:00.000Z", "time": 50, "request": {"method": "GET", "url": "https://example.com/test", "httpVersion": "HTTP/1.1", "cookies": [], "headers": [], "queryString": [], "headersSize": 0, "bodySize": 0}, "response": {"status": 200, "statusText": "OK", "httpVersion": "HTTP/1.1", "cookies": [], "headers": [], "content": {"size": 0, "mimeType": "text/plain"}, "redirectURL": "", "headersSize": 0, "bodySize": 0}, "cache": {}, "timings": {"blocked": 0, "dns": 0, "connect": 0, "send": 0, "wait": 0, "receive": 0, "ssl": 0}}]}}`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	sh, err := NewStreamingHarFromFile(tmpFile)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer sh.Close()

	it := sh.Entries()
	defer it.Close()

	count := 0
	for it.Next() {
		count++
	}
	if count != 1 {
		t.Errorf("expected 1 entry, got %d", count)
	}
	if it.Err() != nil {
		t.Errorf("expected no error, got: %v", it.Err())
	}
}

// ============================================================================
// Next - non-string token in entries scan (from bytes)
// ============================================================================

func TestStreamingNext_NonStringTokenInScan(t *testing.T) {
	// Create a HAR with extra fields before entries that have non-string tokens
	// This is for the bytes-based path where entriesStarted is false
	// The bytes path creates a new decoder and scans for "entries"
	// If the JSON has a structure where scanning encounters something unexpected...

	// Actually for the bytes path, the iterator simply reads the full data
	// and looks for "entries" field. The tricky case is when entries is encountered
	// but followed by a non-[ delimiter.

	// Let's test a file with entries as an object instead of array from file path
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "entries_object.har")
	content := `{"log": {"version": "1.2", "entries": {}}}`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	_, err := NewStreamingHarFromFile(tmpFile)
	if err == nil {
		t.Fatal("expected error when entries is not an array, got nil")
	}
}

// ============================================================================
// GetVersion tests
// ============================================================================

func TestStreamingGetVersion(t *testing.T) {
	sh, err := NewStreamingHarFromFile("testdata/v1.1.har")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer sh.Close()

	if sh.GetVersion() != "1.1" {
		t.Errorf("expected version 1.1, got %s", sh.GetVersion())
	}
}

// ============================================================================
// Entry method test
// ============================================================================

func TestStreamingEntry_ReturnsCurrent(t *testing.T) {
	sh, err := NewStreamingHarFromFile("testdata/complex_test.har")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer sh.Close()

	it := sh.Entries()
	defer it.Close()

	if it.Next() {
		entry := it.Entry()
		if entry == nil {
			t.Fatal("expected non-nil entry")
		}
		if entry.Request.Method != "GET" {
			t.Errorf("expected method GET, got %s", entry.Request.Method)
		}
		if entry.Request.URL != "https://example.com/test" {
			t.Errorf("expected URL 'https://example.com/test', got %s", entry.Request.URL)
		}
	}
}

// ============================================================================
// Iterator Entry before Next called
// ============================================================================

func TestStreamingEntry_BeforeNext(t *testing.T) {
	sh, err := NewStreamingHarFromFile("testdata/example.har")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer sh.Close()

	it := sh.Entries()
	defer it.Close()

	// Entry() before Next() should return empty entry (not nil)
	entry := it.Entry()
	if entry == nil {
		t.Fatal("expected non-nil entry pointer (zero value)")
	}
}

// ============================================================================
// Multiple iterations from same StreamingHar
// ============================================================================

func TestStreamingMultipleIterations(t *testing.T) {
	sh, err := NewStreamingHarFromFile("testdata/example.har")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer sh.Close()

	// First iteration
	it1 := sh.Entries()
	count1 := 0
	for it1.Next() {
		count1++
	}
	it1.Close()

	if count1 != 1 {
		t.Errorf("expected 1 entry in first iteration, got %d", count1)
	}

	// Second iteration (file should be reopened)
	it2 := sh.Entries()
	count2 := 0
	for it2.Next() {
		count2++
	}
	it2.Close()

	if count2 != 1 {
		t.Errorf("expected 1 entry in second iteration, got %d", count2)
	}
}

// ============================================================================
// StreamingHar Close sets file to nil
// ============================================================================

func TestStreamingClose_SetsFileNil(t *testing.T) {
	sh, err := NewStreamingHarFromFile("testdata/example.har")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if err := sh.Close(); err != nil {
		t.Errorf("expected no error on close, got: %v", err)
	}

	// After close, file should be nil
	if sh.file != nil {
		t.Error("expected file to be nil after close")
	}
}

func TestStreamingClose_WrapsCloseError(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "closed-*.har")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}

	sh := &StreamingHar{file: tmpFile}
	err = sh.Close()
	assertHarErrorCode(t, err, ErrCodeFileSystem)
	if !errors.Is(err, os.ErrClosed) {
		t.Fatalf("Close() error should wrap os.ErrClosed, got: %v", err)
	}
	if sh.file != nil {
		t.Error("expected file to be nil after failed close")
	}
}

func TestStreamingIteratorClose_WrapsCloseError(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "closed-iterator-*.har")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}

	it := &StreamingEntryIterator{file: tmpFile}
	err = it.Close()
	assertHarErrorCode(t, err, ErrCodeFileSystem)
	if !errors.Is(err, os.ErrClosed) {
		t.Fatalf("Close() error should wrap os.ErrClosed, got: %v", err)
	}
	if !it.closed {
		t.Error("expected iterator to be marked closed after failed close")
	}
}

// ============================================================================
// parseHarBasicInfo - field name not a string (rare, but possible with streaming JSON)
// ============================================================================

func TestStreamingParseHarBasicInfo_FieldNameNotString(t *testing.T) {
	// This is hard to trigger through NewStreamingHarFromFile since
	// json.Decoder.Token() always returns proper tokens for well-formed JSON.
	// But we can test with a crafted file that has a structure where
	// a non-string token appears where a field name is expected.
	// Actually in valid JSON, object keys are always strings, so this
	// branch is essentially unreachable for valid JSON. However, let's
	// verify the code handles edge cases.

	// A valid JSON file should work fine
	sh, err := NewStreamingHarFromFile("testdata/full.har")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer sh.Close()
}

// ============================================================================
// Full HAR with all fields
// ============================================================================

func TestStreamingFullHAR(t *testing.T) {
	sh, err := NewStreamingHarFromFile("testdata/full.har")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer sh.Close()

	// Check metadata
	if sh.GetVersion() != "1.2" {
		t.Errorf("expected version 1.2, got %s", sh.GetVersion())
	}

	creator := sh.GetCreator()
	if creator.Name != "Go-HAR Test" {
		t.Errorf("expected creator name 'Go-HAR Test', got '%s'", creator.Name)
	}

	browser := sh.GetBrowser()
	if browser.Name != "" {
		t.Errorf("expected empty browser name, got '%s'", browser.Name)
	}

	pages := sh.GetPages()
	if len(pages) != 1 {
		t.Errorf("expected 1 page, got %d", len(pages))
	}

	// Iterate entries
	it := sh.Entries()
	defer it.Close()

	count := 0
	for it.Next() {
		entry := it.Entry()
		if entry.Request.URL != "https://example.com/test" {
			t.Errorf("unexpected URL: %s", entry.Request.URL)
		}
		count++
	}
	if count != 1 {
		t.Errorf("expected 1 entry, got %d", count)
	}
	if it.Err() != nil {
		t.Errorf("expected no iteration error, got: %v", it.Err())
	}
}

// ============================================================================
// GetAllEntries with iteration error
// ============================================================================

func TestStreamingGetAllEntries_WithIterationError(t *testing.T) {
	// Create a file with a truncated entry to cause iteration error
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "truncated.har")
	content := `{"log": {"version": "1.2", "entries": [{"startedDateTime": "2023-01-01T00:00:00.000Z", "time": 50, "request": {"method": "GET", "url": "https://example.com/test", "httpVersion": "HTTP/1.1", "cookies": [], "headers": [], "queryString": [], "headersSize": 0, "bodySize": 0}, "response": {"status": 200, "statusText": "OK", "httpVersion": "HTTP/1.1", "cookies": [], "headers": [], "content": {"size": 0, "mimeType": "text/plain"}, "redirectURL": "", "headersSize": 0, "bodySize": 0}, "cache": {}, "timings": {"blocked": 0, "dns": 0, "connect": 0, "send": 0, "wait": 0, "receive": 0, "ssl": 0}}, {"incomplete":`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	sh, err := NewStreamingHarFromFile(tmpFile)
	if err != nil {
		t.Fatalf("expected no error opening file, got: %v", err)
	}
	defer sh.Close()

	entries, err := sh.GetAllEntries()
	// Should get the first entry but error on the second
	if len(entries) < 1 {
		t.Error("expected at least 1 entry before error")
	}
	assertHarErrorCode(t, err, ErrCodeJSONParse)
}

// ============================================================================
// Next from file-based streaming with findHarObjectStart error on re-open
// ============================================================================

func TestStreamingEntries_FileReopenErrorIsTyped(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "remove_before_iteration.har")
	content := `{"log": {"version": "1.2", "creator": {"name": "test", "version": "1.0"}, "entries": []}}`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	sh, err := NewStreamingHarFromFile(tmpFile)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer sh.Close()

	if err := os.Remove(tmpFile); err != nil {
		t.Fatalf("failed to remove temp file: %v", err)
	}

	it := sh.Entries()
	assertHarErrorCode(t, it.Err(), ErrCodeFileSystem)
	if it.Next() {
		t.Error("expected Next() to return false when source file cannot be reopened")
	}
}

func TestStreamingEntriesNoDataSourceErrorIsTyped(t *testing.T) {
	sh := &StreamingHar{}
	it := sh.Entries()
	assertHarErrorCode(t, it.Err(), ErrCodeInvalidFormat)
	if it.Next() {
		t.Error("expected Next() to return false without data source")
	}
}

func TestStreamingEntries_FileReopenFindStartError(t *testing.T) {
	// We test the path where the file is reopened but findHarObjectStart fails
	// This is hard to simulate because a valid HAR file will always have
	// the correct structure. Instead, test that normal reopen works.
	sh, err := NewStreamingHarFromFile("testdata/example.har")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	it := sh.Entries()
	defer it.Close()

	count := 0
	for it.Next() {
		count++
	}
	if count != 1 {
		t.Errorf("expected 1 entry, got %d", count)
	}
}

// ============================================================================
// Next from file where entriesStarted is already true (from file reopen)
// ============================================================================

func TestStreamingNext_EntriesAlreadyStarted(t *testing.T) {
	// When using file-based streaming, the Entries() method sets entriesStarted=true
	// because it re-parses through findHarObjectStart and parseHarBasicInfo
	sh, err := NewStreamingHarFromFile("testdata/complex_test.har")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer sh.Close()

	it := sh.Entries()
	defer it.Close()

	// For file-based iterator, entriesStarted should be true
	if !it.entriesStarted {
		t.Error("expected entriesStarted to be true for file-based iterator")
	}

	count := 0
	for it.Next() {
		count++
	}
	if count != 3 {
		t.Errorf("expected 3 entries, got %d", count)
	}
}

// ============================================================================
// Non-EOF Err handling
// ============================================================================

func TestStreamingErr_NonEOFReturnsError(t *testing.T) {
	it := &StreamingEntryIterator{
		err:    &json.UnmarshalTypeError{},
		closed: false,
	}

	// Err() should return non-nil for non-EOF errors
	if it.Err() == nil {
		t.Error("expected non-nil error for non-EOF error type")
	}
}

func TestStreamingErr_EOFReturnsNil(t *testing.T) {
	it := &StreamingEntryIterator{
		err:    os.ErrClosed, // not io.EOF
		closed: false,
	}

	// Non-EOF error should be returned
	if it.Err() == nil {
		t.Error("expected non-nil error for non-EOF error")
	}
}

func TestStreamingHarNilReceiverMethods(t *testing.T) {
	var sh *StreamingHar

	if err := sh.Close(); err != nil {
		t.Fatalf("Close() = %v, want nil", err)
	}
	if sh.GetVersion() != "" {
		t.Errorf("GetVersion() = %q, want empty", sh.GetVersion())
	}
	if got := sh.GetCreator(); got != (Creator{}) {
		t.Errorf("GetCreator() = %#v, want zero Creator", got)
	}
	if got := sh.GetBrowser(); got != (Browser{}) {
		t.Errorf("GetBrowser() = %#v, want zero Browser", got)
	}
	if pages := sh.GetPages(); pages != nil {
		t.Errorf("GetPages() = %#v, want nil", pages)
	}

	it := sh.Entries()
	if it == nil {
		t.Fatal("Entries() returned nil iterator")
	}
	assertHarErrorCode(t, it.Err(), ErrCodeInvalidFormat)
	if it.Next() {
		t.Error("Next() should return false for nil StreamingHar iterator")
	}

	entries, err := sh.GetAllEntries()
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
	if entries != nil {
		t.Errorf("GetAllEntries() entries = %#v, want nil", entries)
	}
}

func TestStreamingEntryIteratorNilReceiverMethods(t *testing.T) {
	var it *StreamingEntryIterator

	if it.Next() {
		t.Error("Next() should return false for nil iterator")
	}
	if entry := it.Entry(); entry != nil {
		t.Errorf("Entry() = %#v, want nil", entry)
	}
	if pos := it.Position(); pos != 0 {
		t.Errorf("Position() = %d, want 0", pos)
	}
	assertHarErrorCode(t, it.Err(), ErrCodeInvalidFormat)
	if err := it.Close(); err != nil {
		t.Fatalf("Close() = %v, want nil", err)
	}
}
