package har

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// ============================================================================
// LazyContent.Load tests
// ============================================================================

func TestLazyLoad_NotLoaded(t *testing.T) {
	data := `{"size": 100, "mimeType": "text/plain", "text": "hello world", "encoding": "utf8"}`
	var lc LazyContent
	if err := json.Unmarshal([]byte(data), &lc); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if lc.loaded {
		t.Error("expected loaded to be false after unmarshal")
	}
	if lc.Size != 100 {
		t.Errorf("expected size 100, got %d", lc.Size)
	}
	if lc.MimeType != "text/plain" {
		t.Errorf("expected mimeType text/plain, got %s", lc.MimeType)
	}

	// Now load
	if err := lc.Load(); err != nil {
		t.Fatalf("expected no error on Load, got: %v", err)
	}

	if !lc.loaded {
		t.Error("expected loaded to be true after Load")
	}
	if lc.Text == nil || *lc.Text != "hello world" {
		t.Errorf("expected text 'hello world', got %v", lc.Text)
	}
	if lc.Encoding == nil || *lc.Encoding != "utf8" {
		t.Errorf("expected encoding 'utf8', got %v", lc.Encoding)
	}
}

func TestLazyLoad_AlreadyLoaded(t *testing.T) {
	data := `{"size": 50, "mimeType": "application/json", "text": "{}"}`
	var lc LazyContent
	if err := json.Unmarshal([]byte(data), &lc); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Load first time
	if err := lc.Load(); err != nil {
		t.Fatalf("expected no error on first Load, got: %v", err)
	}

	// Load second time - should return nil error immediately
	if err := lc.Load(); err != nil {
		t.Fatalf("expected no error on second Load, got: %v", err)
	}
}

func TestLazyLoad_EmptyContent(t *testing.T) {
	data := `{"size": 0, "mimeType": ""}`
	var lc LazyContent
	if err := json.Unmarshal([]byte(data), &lc); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if err := lc.Load(); err != nil {
		t.Fatalf("expected no error on Load of empty content, got: %v", err)
	}

	if lc.Text != nil {
		t.Errorf("expected nil text for empty content, got %v", lc.Text)
	}
	if lc.Encoding != nil {
		t.Errorf("expected nil encoding for empty content, got %v", lc.Encoding)
	}
}

func TestLazyLoad_NilContent(t *testing.T) {
	var lc *LazyContent
	err := lc.Load()
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
}

// ============================================================================
// LazyContent.GetText tests
// ============================================================================

func TestLazyGetText_NotLoaded(t *testing.T) {
	data := `{"size": 11, "mimeType": "text/plain", "text": "hello world"}`
	var lc LazyContent
	if err := json.Unmarshal([]byte(data), &lc); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	text, err := lc.GetText()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if text == nil || *text != "hello world" {
		t.Errorf("expected text 'hello world', got %v", text)
	}
}

func TestLazyGetText_AlreadyLoaded(t *testing.T) {
	data := `{"size": 5, "mimeType": "text/plain", "text": "hello"}`
	var lc LazyContent
	if err := json.Unmarshal([]byte(data), &lc); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Pre-load
	if err := lc.Load(); err != nil {
		t.Fatalf("expected no error on Load, got: %v", err)
	}

	// GetText should use already-loaded data
	text, err := lc.GetText()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if text == nil || *text != "hello" {
		t.Errorf("expected text 'hello', got %v", text)
	}
}

func TestLazyGetText_NoText(t *testing.T) {
	data := `{"size": 0, "mimeType": "text/plain"}`
	var lc LazyContent
	if err := json.Unmarshal([]byte(data), &lc); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	text, err := lc.GetText()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if text != nil {
		t.Errorf("expected nil text, got %v", text)
	}
}

func TestLazyGetText_NilContent(t *testing.T) {
	var lc *LazyContent
	text, err := lc.GetText()
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
	if text != nil {
		t.Errorf("expected nil text for nil content, got %v", text)
	}
}

// ============================================================================
// ParseHarWithLazyLoading tests
// ============================================================================

func TestLazyParseHarWithLazyLoading_Valid(t *testing.T) {
	data, err := os.ReadFile("testdata/example.har")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	lh, err := ParseHarWithLazyLoading(data)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if lh.Log.Version != "1.2" {
		t.Errorf("expected version 1.2, got %s", lh.Log.Version)
	}
	if lh.Log.Creator.Name != "Go-HAR Test" {
		t.Errorf("expected creator name 'Go-HAR Test', got '%s'", lh.Log.Creator.Name)
	}
	if len(lh.Log.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(lh.Log.Entries))
	}
}

func TestLazyParseHarWithLazyLoading_InvalidJSON(t *testing.T) {
	_, err := ParseHarWithLazyLoading([]byte("not json"))
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
}

func TestLazyParseHarWithLazyLoading_EmptyInput(t *testing.T) {
	_, err := ParseHarWithLazyLoading([]byte{})
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
}

func TestLazyParseHarWithLazyLoading_Empty(t *testing.T) {
	// Empty JSON object {} unmarshals to an empty LazyHar without error
	// since LazyHar has no required fields in its JSON struct tags
	lh, err := ParseHarWithLazyLoading([]byte("{}"))
	if err != nil {
		t.Fatalf("unexpected error for empty JSON object: %v", err)
	}
	if lh == nil {
		t.Fatal("expected non-nil LazyHar for empty JSON object")
	}
}

func TestLazyParseHarWithLazyLoading_Complex(t *testing.T) {
	data, err := os.ReadFile("testdata/complex_test.har")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	lh, err := ParseHarWithLazyLoading(data)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(lh.Log.Entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(lh.Log.Entries))
	}
	if len(lh.Log.Pages) != 2 {
		t.Errorf("expected 2 pages, got %d", len(lh.Log.Pages))
	}
}

// ============================================================================
// ParseHarFileWithLazyLoading tests
// ============================================================================

func TestLazyParseHarFileWithLazyLoading_Valid(t *testing.T) {
	lh, err := ParseHarFileWithLazyLoading("testdata/example.har")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if lh.Log.Version != "1.2" {
		t.Errorf("expected version 1.2, got %s", lh.Log.Version)
	}
	if len(lh.Log.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(lh.Log.Entries))
	}
}

func TestLazyParseHarFileWithLazyLoading_FileNotFound(t *testing.T) {
	_, err := ParseHarFileWithLazyLoading("testdata/nonexistent.har")
	assertHarErrorCode(t, err, ErrCodeFileSystem)
}

func TestLazyParseHarFileWithLazyLoading_InvalidContent(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.har")
	if err := os.WriteFile(tmpFile, []byte("not json"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	_, err := ParseHarFileWithLazyLoading(tmpFile)
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
}

// ============================================================================
// ToStandardHar tests
// ============================================================================

func TestLazyToStandardHar_Valid(t *testing.T) {
	data, err := os.ReadFile("testdata/example.har")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	lh, err := ParseHarWithLazyLoading(data)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	h, err := lh.ToStandardHar()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if h.Log.Version != "1.2" {
		t.Errorf("expected version 1.2, got %s", h.Log.Version)
	}
	if len(h.Log.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(h.Log.Entries))
	}
	if h.Log.Entries[0].Request.Method != "GET" {
		t.Errorf("expected method GET, got %s", h.Log.Entries[0].Request.Method)
	}
}

func TestLazyToStandardHar_NilHar(t *testing.T) {
	var lh *LazyHar
	h, err := lh.ToStandardHar()
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
	if h != nil {
		t.Errorf("expected nil standard HAR for nil LazyHar, got %#v", h)
	}
}

func TestLazyToStandardHar_Complex(t *testing.T) {
	data, err := os.ReadFile("testdata/complex_test.har")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	lh, err := ParseHarWithLazyLoading(data)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	h, err := lh.ToStandardHar()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(h.Log.Entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(h.Log.Entries))
	}
}

func TestLazyToStandardHar_WithContent(t *testing.T) {
	data, err := os.ReadFile("testdata/complex_test.har")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	lh, err := ParseHarWithLazyLoading(data)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	h, err := lh.ToStandardHar()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// First entry has content with text
	if h.Log.Entries[0].Response.Content.Size != 100 {
		t.Errorf("expected content size 100, got %d", h.Log.Entries[0].Response.Content.Size)
	}
	if h.Log.Entries[0].Response.Content.MimeType != "application/json" {
		t.Errorf("expected mimeType application/json, got %s", h.Log.Entries[0].Response.Content.MimeType)
	}
	if h.Log.Entries[0].Response.Content.Text != `{"message":"Hello, World!"}` {
		t.Errorf("expected content text to be preserved, got %q", h.Log.Entries[0].Response.Content.Text)
	}
	if h.Log.Entries[0].Response.Content.Encoding != "base64" {
		t.Errorf("expected content encoding base64, got %q", h.Log.Entries[0].Response.Content.Encoding)
	}
	if h.Log.Entries[2].Response.Content.Compression != 2000 {
		t.Errorf("expected content compression 2000, got %d", h.Log.Entries[2].Response.Content.Compression)
	}
}

func TestLazyToStandardHar_NilContent(t *testing.T) {
	// Create a LazyHar with nil content in response
	lh := &LazyHar{}
	lh.Log.Version = "1.2"
	lh.Log.Entries = []LazyEntries{
		{
			StartedDateTime: parseTime(t, "2023-01-01T00:00:00.000Z"),
			Time:            50,
			Request: Request{
				Method:      "GET",
				URL:         "https://example.com/test",
				HTTPVersion: "HTTP/1.1",
			},
			Response: LazyResponse{
				Status:      200,
				StatusText:  "OK",
				HTTPVersion: "HTTP/1.1",
				Content:     nil, // nil content
			},
		},
	}

	h, err := lh.ToStandardHar()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(h.Log.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(h.Log.Entries))
	}
}

func TestLazyToStandardHar_ContentLoadError(t *testing.T) {
	// Create a LazyContent with invalid rawData that will fail to load
	lh := &LazyHar{}
	lh.Log.Version = "1.2"
	lh.Log.Entries = []LazyEntries{
		{
			StartedDateTime: parseTime(t, "2023-01-01T00:00:00.000Z"),
			Time:            50,
			Request: Request{
				Method:      "GET",
				URL:         "https://example.com/test",
				HTTPVersion: "HTTP/1.1",
			},
			Response: LazyResponse{
				Status:      200,
				StatusText:  "OK",
				HTTPVersion: "HTTP/1.1",
				Content: &LazyContent{
					Size:     10,
					MimeType: "text/plain",
					rawData:  json.RawMessage(`{invalid json}`),
					loaded:   false,
				},
			},
		},
	}

	_, err := lh.ToStandardHar()
	assertHarErrorCode(t, err, ErrCodeJSONParse)
}

// ============================================================================
// GetEntry tests
// ============================================================================

func TestLazyGetEntry_ValidIndex(t *testing.T) {
	data, err := os.ReadFile("testdata/complex_test.har")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	lh, err := ParseHarWithLazyLoading(data)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	entry, err := lh.GetEntry(0)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if entry.Request.Method != "GET" {
		t.Errorf("expected method GET, got %s", entry.Request.Method)
	}
	if entry.Request.URL != "https://example.com/test" {
		t.Errorf("expected URL 'https://example.com/test', got %s", entry.Request.URL)
	}
}

func TestLazyGetEntry_LastIndex(t *testing.T) {
	data, err := os.ReadFile("testdata/complex_test.har")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	lh, err := ParseHarWithLazyLoading(data)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	entry, err := lh.GetEntry(2)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if entry.Request.URL != "https://example.com/image.jpg" {
		t.Errorf("expected URL 'https://example.com/image.jpg', got %s", entry.Request.URL)
	}
}

func TestLazyGetEntry_NegativeIndex(t *testing.T) {
	lh := &LazyHar{}
	lh.Log.Entries = []LazyEntries{{}}

	_, err := lh.GetEntry(-1)
	if err == nil {
		t.Fatal("expected error for negative index, got nil")
	}
}

func TestLazyGetEntry_NilHar(t *testing.T) {
	var lh *LazyHar
	entry, err := lh.GetEntry(0)
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
	if entry != nil {
		t.Errorf("expected nil entry for nil LazyHar, got %#v", entry)
	}
}

func TestLazyGetEntry_IndexOutOfRange(t *testing.T) {
	lh := &LazyHar{}
	lh.Log.Entries = []LazyEntries{{}}

	_, err := lh.GetEntry(1)
	if err == nil {
		t.Fatal("expected error for out of range index, got nil")
	}
}

func TestLazyGetEntry_EmptyEntries(t *testing.T) {
	lh := &LazyHar{}
	lh.Log.Entries = []LazyEntries{}

	_, err := lh.GetEntry(0)
	if err == nil {
		t.Fatal("expected error for empty entries, got nil")
	}
}

// ============================================================================
// GetEntriesCount tests
// ============================================================================

func TestLazyGetEntriesCount(t *testing.T) {
	data, err := os.ReadFile("testdata/complex_test.har")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	lh, err := ParseHarWithLazyLoading(data)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if count := lh.GetEntriesCount(); count != 3 {
		t.Errorf("expected 3 entries, got %d", count)
	}
}

func TestLazyGetEntriesCount_Empty(t *testing.T) {
	data, err := os.ReadFile("testdata/minimal.har")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	lh, err := ParseHarWithLazyLoading(data)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if count := lh.GetEntriesCount(); count != 0 {
		t.Errorf("expected 0 entries, got %d", count)
	}
}

func TestLazyGetEntriesCount_NoEntries(t *testing.T) {
	lh := &LazyHar{}

	if count := lh.GetEntriesCount(); count != 0 {
		t.Errorf("expected 0 entries for nil entries, got %d", count)
	}
}

func TestLazyGetEntriesCount_NilHar(t *testing.T) {
	var lh *LazyHar
	if count := lh.GetEntriesCount(); count != 0 {
		t.Errorf("expected 0 entries for nil LazyHar, got %d", count)
	}
}

// ============================================================================
// GetResponseContent tests
// ============================================================================

func TestLazyGetResponseContent_ValidIndex(t *testing.T) {
	data, err := os.ReadFile("testdata/complex_test.har")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	lh, err := ParseHarWithLazyLoading(data)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	content, err := lh.GetResponseContent(0)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if content == nil {
		t.Fatal("expected non-nil content")
	}
	if content.MimeType != "application/json" {
		t.Errorf("expected mimeType 'application/json', got '%s'", content.MimeType)
	}
	if content.Size != 100 {
		t.Errorf("expected size 100, got %d", content.Size)
	}
}

func TestLazyGetResponseContent_InvalidIndex(t *testing.T) {
	lh := &LazyHar{}
	lh.Log.Entries = []LazyEntries{}

	_, err := lh.GetResponseContent(0)
	if err == nil {
		t.Fatal("expected error for invalid index, got nil")
	}
}

func TestLazyGetResponseContent_NilHar(t *testing.T) {
	var lh *LazyHar
	content, err := lh.GetResponseContent(0)
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
	if content != nil {
		t.Errorf("expected nil content for nil LazyHar, got %#v", content)
	}
}

func TestLazyGetResponseContent_NilContent(t *testing.T) {
	lh := &LazyHar{}
	lh.Log.Entries = []LazyEntries{
		{
			Response: LazyResponse{
				Content: nil,
			},
		},
	}

	content, err := lh.GetResponseContent(0)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if content != nil {
		t.Errorf("expected nil content, got %v", content)
	}
}

// ============================================================================
// GetResponseText tests
// ============================================================================

func TestLazyGetResponseText_Valid(t *testing.T) {
	data, err := os.ReadFile("testdata/complex_test.har")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	lh, err := ParseHarWithLazyLoading(data)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	text, err := lh.GetResponseText(0)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if text == nil {
		t.Fatal("expected non-nil text")
	}
	if *text != `{"message":"Hello, World!"}` {
		t.Errorf("expected text '{\"message\":\"Hello, World!\"}', got '%s'", *text)
	}
}

func TestLazyGetResponseText_NoText(t *testing.T) {
	// Entry with content but no text field
	data, err := os.ReadFile("testdata/example.har")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	lh, err := ParseHarWithLazyLoading(data)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	text, err := lh.GetResponseText(0)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	// The example.har has content with size 100 but no text field
	if text != nil {
		t.Errorf("expected nil text for content without text, got '%v'", text)
	}
}

func TestLazyGetResponseText_InvalidIndex(t *testing.T) {
	lh := &LazyHar{}
	lh.Log.Entries = []LazyEntries{}

	_, err := lh.GetResponseText(5)
	if err == nil {
		t.Fatal("expected error for invalid index, got nil")
	}
}

func TestLazyGetResponseText_NilHar(t *testing.T) {
	var lh *LazyHar
	text, err := lh.GetResponseText(0)
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
	if text != nil {
		t.Errorf("expected nil text for nil LazyHar, got %#v", text)
	}
}

func TestLazyGetResponseText_NilContent(t *testing.T) {
	lh := &LazyHar{}
	lh.Log.Entries = []LazyEntries{
		{
			Response: LazyResponse{
				Content: nil,
			},
		},
	}

	text, err := lh.GetResponseText(0)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if text != nil {
		t.Errorf("expected nil text for nil content, got %v", text)
	}
}

// ============================================================================
// LazyContent UnmarshalJSON tests
// ============================================================================

func TestLazyContentUnmarshalJSON_InvalidJSON(t *testing.T) {
	var lc LazyContent
	err := json.Unmarshal([]byte(`not valid json`), &lc)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestLazyContentUnmarshalJSON_InvalidFieldType(t *testing.T) {
	var lc LazyContent
	err := json.Unmarshal([]byte(`{"size":"large","mimeType":"text/plain"}`), &lc)
	assertHarErrorCode(t, err, ErrCodeJSONParse)
}

func TestLazyContentUnmarshalJSONDirectInvalidJSON(t *testing.T) {
	var lc LazyContent
	err := lc.UnmarshalJSON([]byte(`not valid json`))
	assertHarErrorCode(t, err, ErrCodeJSONParse)
}

func TestLazyContentUnmarshalJSONNilReceiver(t *testing.T) {
	var lc *LazyContent
	err := lc.UnmarshalJSON([]byte(`{"size":1,"mimeType":"text/plain"}`))
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
}

func TestLazyContentUnmarshalJSON_WithCompression(t *testing.T) {
	data := `{"size": 5000, "mimeType": "image/jpeg", "compression": 2000, "comment": "compressed"}`
	var lc LazyContent
	if err := json.Unmarshal([]byte(data), &lc); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if lc.Compression != 2000 {
		t.Errorf("expected compression 2000, got %d", lc.Compression)
	}
	if lc.Comment != "compressed" {
		t.Errorf("expected comment 'compressed', got '%s'", lc.Comment)
	}
}

// ============================================================================
// LazyContent concurrent access (Load + GetText)
// ============================================================================

func TestLazyContentConcurrentAccess(t *testing.T) {
	data := `{"size": 11, "mimeType": "text/plain", "text": "hello world"}`
	var lc LazyContent
	if err := json.Unmarshal([]byte(data), &lc); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Concurrent GetText calls
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			text, err := lc.GetText()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if text == nil || *text != "hello world" {
				t.Errorf("expected text 'hello world', got %v", text)
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

// ============================================================================
// Full workflow test
// ============================================================================

func TestLazyFullWorkflow(t *testing.T) {
	// Parse file with lazy loading
	lh, err := ParseHarFileWithLazyLoading("testdata/complex_test.har")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Check count
	if count := lh.GetEntriesCount(); count != 3 {
		t.Errorf("expected 3 entries, got %d", count)
	}

	// Get individual entries
	entry0, err := lh.GetEntry(0)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if entry0.Request.Method != "GET" {
		t.Errorf("expected GET, got %s", entry0.Request.Method)
	}

	// Get response content
	content, err := lh.GetResponseContent(0)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if content.MimeType != "application/json" {
		t.Errorf("expected mimeType 'application/json', got '%s'", content.MimeType)
	}

	// Get response text
	text, err := lh.GetResponseText(0)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if text == nil {
		t.Fatal("expected non-nil text")
	}

	// Convert to standard HAR
	h, err := lh.ToStandardHar()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(h.Log.Entries) != 3 {
		t.Errorf("expected 3 entries in standard HAR, got %d", len(h.Log.Entries))
	}
}

// Helper to parse time for test setup
func parseTime(t *testing.T, s string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		parsed, err = time.Parse("2006-01-02T15:04:05.000Z", s)
		if err != nil {
			t.Fatalf("failed to parse time %q: %v", s, err)
		}
	}
	return parsed
}
