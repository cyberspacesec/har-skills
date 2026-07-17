package har

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// writeTempHAR writes content to a temp .har file and returns its path.
func writeTempHAR(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "test.har")
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp HAR: %v", err)
	}
	return p
}

// ============================================================================
// closeStreamingFileAfterError — Close error branch (lines 91-93)
// ============================================================================

func TestCovCloseStreamingFileAfterError_CloseError(t *testing.T) {
	// Use a file that is already closed so file.Close() returns an error
	// (os.ErrClosed), exercising the WithMetadata("closeError", ...) branch.
	f, err := os.CreateTemp(t.TempDir(), "closed-err-*.har")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}

	harErr := NewJSONParseError("parse failed", nil)
	result := closeStreamingFileAfterError(f, harErr)

	if result != harErr {
		t.Fatal("expected the same HarError pointer to be returned")
	}
	if result.Metadata == nil {
		t.Fatal("expected metadata to be populated with closeError")
	}
	closeVal, ok := result.Metadata["closeError"]
	if !ok {
		t.Fatal("expected closeError metadata key")
	}
	if closeVal == "" {
		t.Error("expected non-empty closeError message")
	}
}

func TestCovCloseStreamingFileAfterError_CloseOK(t *testing.T) {
	// Normal case: file.Close() succeeds, no metadata added.
	f, err := os.CreateTemp(t.TempDir(), "ok-close-*.har")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	harErr := NewJSONParseError("parse failed", nil)
	result := closeStreamingFileAfterError(f, harErr)

	if result != harErr {
		t.Fatal("expected the same HarError pointer to be returned")
	}
	if result.Metadata != nil {
		t.Errorf("expected no metadata when Close succeeds, got %#v", result.Metadata)
	}
}

// ============================================================================
// findHarObjectStart — token error after "log" field (lines 117-119)
// ============================================================================

func TestCovFindHarObjectStart_TokenErrorAfterLog(t *testing.T) {
	// Data: {"log":  (truncated right after the "log" key with no value).
	// findHarObjectStart reads "{", then loops to find "log" string,
	// then tries to read the token after "log" -> io.EOF -> lines 117-119.
	p := writeTempHAR(t, `{"log":`)
	_, err := NewStreamingHarFromFile(p)
	if err == nil {
		t.Fatal("expected error for truncated HAR after log key, got nil")
	}
	assertHarErrorCode(t, err, ErrCodeJSONParse)
}

// ============================================================================
// parseHarBasicInfo — field name read error / EOF (lines 130-132)
// ============================================================================

func TestCovParseHarBasicInfo_FieldNameReadError(t *testing.T) {
	// Data: {"log":{  (truncated right after the log object's opening brace).
	// findHarObjectStart succeeds ({, log, {), then parseHarBasicInfo
	// reads the first field name -> io.EOF -> lines 130-132.
	p := writeTempHAR(t, `{"log":{`)
	_, err := NewStreamingHarFromFile(p)
	if err == nil {
		t.Fatal("expected error for truncated HAR after log object brace, got nil")
	}
	assertHarErrorCode(t, err, ErrCodeJSONParse)
}

// ============================================================================
// parseHarBasicInfo — entries array start token read error (lines 162-164)
// ============================================================================

func TestCovParseHarBasicInfo_EntriesStartTokenError(t *testing.T) {
	// Data: {"log":{"version":"1.2","entries":  (truncated after entries key).
	// parseHarBasicInfo reads "version", decodes "1.2", reads "entries" key,
	// then decoder.Token() expecting "[" -> io.EOF -> lines 162-164.
	p := writeTempHAR(t, `{"log":{"version":"1.2","entries":`)
	_, err := NewStreamingHarFromFile(p)
	if err == nil {
		t.Fatal("expected error for truncated HAR after entries key, got nil")
	}
	assertHarErrorCode(t, err, ErrCodeJSONParse)
}

// ============================================================================
// Entries() — findHarObjectStart fails on re-open (lines 265-274)
// ============================================================================

func TestCovEntries_ReopenFindHarObjectStartFails(t *testing.T) {
	// Parse a valid HAR first (opens & validates file), then overwrite the
	// file on disk with content that fails findHarObjectStart on re-open.
	p := writeTempHAR(t, `{"log":{"version":"1.2","entries":[]}}`)

	sh, err := NewStreamingHarFromFile(p)
	if err != nil {
		t.Fatalf("expected no error parsing valid HAR, got: %v", err)
	}
	defer sh.Close()

	// Overwrite with content whose first token is not "{"
	if err := os.WriteFile(p, []byte(`"not an object"`), 0644); err != nil {
		t.Fatalf("failed to overwrite file: %v", err)
	}

	it := sh.Entries()
	defer it.Close()

	assertHarErrorCode(t, it.Err(), ErrCodeJSONParse)
	if it.Next() {
		t.Error("expected Next() to return false when findHarObjectStart fails on re-open")
	}
}

// ============================================================================
// Entries() — parseHarBasicInfo fails on re-open (lines 277-286)
// ============================================================================

func TestCovEntries_ReopenParseHarBasicInfoFails(t *testing.T) {
	// Parse a valid HAR first, then overwrite with content where
	// findHarObjectStart succeeds but parseHarBasicInfo fails (entries not an array).
	p := writeTempHAR(t, `{"log":{"version":"1.2","entries":[]}}`)

	sh, err := NewStreamingHarFromFile(p)
	if err != nil {
		t.Fatalf("expected no error parsing valid HAR, got: %v", err)
	}
	defer sh.Close()

	// Overwrite: findHarObjectStart succeeds ({, log, {), but parseHarBasicInfo
	// reaches "entries" and finds a string instead of "[" -> error.
	if err := os.WriteFile(p, []byte(`{"log":{"version":"1.2","entries":"notarray"}}`), 0644); err != nil {
		t.Fatalf("failed to overwrite file: %v", err)
	}

	it := sh.Entries()
	defer it.Close()

	assertHarErrorCode(t, it.Err(), ErrCodeJSONParse)
	if it.Next() {
		t.Error("expected Next() to return false when parseHarBasicInfo fails on re-open")
	}
}

// ============================================================================
// Next() — token read error during entries scan, bytes path (lines 318-321)
// ============================================================================

func TestCovNext_ScanTokenErrorBytes(t *testing.T) {
	// Construct a StreamingHar directly with data that has no "entries" key.
	// The byte-path iterator (entriesStarted=false) scans tokens for "entries",
	// reaching io.EOF at end of stream -> lines 318-321.
	sh := &StreamingHar{data: []byte(`{"log":{"version":"1.2"}}`)}
	it := sh.Entries()
	defer it.Close()

	if it.Next() {
		t.Error("expected Next() to return false when entries key is absent")
	}
	// err is io.EOF, so Err() returns nil.
	if it.Err() != nil {
		t.Errorf("expected nil Err() for io.EOF scan, got: %v", it.Err())
	}
}

// ============================================================================
// Next() — token read error after finding "entries" key, bytes path (lines 325-328)
// ============================================================================

func TestCovNext_ScanEntriesValueTokenErrorBytes(t *testing.T) {
	// Construct data where "entries" appears as a string value (comment field)
	// immediately followed by end of stream. The scanner finds the "entries"
	// string token, then decoder.Token() to read the array start -> io.EOF
	// -> lines 325-328.
	sh := &StreamingHar{data: []byte(`{"log":{"version":"1.2","comment":"entries"`)}
	it := sh.Entries()
	defer it.Close()

	if it.Next() {
		t.Error("expected Next() to return false when entries value token read errors")
	}
	// err is io.EOF, so Err() returns nil.
	if it.Err() != nil {
		t.Errorf("expected nil Err() for io.EOF after entries token, got: %v", it.Err())
	}
}

// ============================================================================
// Next() — entries field followed by non-"[" token, bytes path (lines 332-335)
// ============================================================================

func TestCovNext_ScanEntriesNotArrayBytes(t *testing.T) {
	// Construct data where "entries" appears as a string value (comment field)
	// followed by a "}" delimiter (not "["). The scanner finds the "entries"
	// string token, reads next token "}", which is not "[" -> lines 332-335.
	sh := &StreamingHar{data: []byte(`{"log":{"version":"1.2","comment":"entries"}}`)}
	it := sh.Entries()
	defer it.Close()

	if it.Next() {
		t.Error("expected Next() to return false when entries is not followed by [")
	}
	assertHarErrorCode(t, it.Err(), ErrCodeInvalidFormat)
}

// ============================================================================
// Err() — io.EOF returns nil (lines 376-378)
// ============================================================================

func TestCovErr_EOFReturnsNil(t *testing.T) {
	it := &StreamingEntryIterator{err: io.EOF}
	if it.Err() != nil {
		t.Errorf("expected nil Err() when err is io.EOF, got: %v", it.Err())
	}
}

func TestCovErr_NilReceiver(t *testing.T) {
	var it *StreamingEntryIterator
	assertHarErrorCode(t, it.Err(), ErrCodeInvalidFormat)
}

// ============================================================================
// wrapStreamingIteratorError — nil / io.EOF passthrough (lines 383-385)
// ============================================================================

func TestCovWrapStreamingIteratorError_NilError(t *testing.T) {
	result := wrapStreamingIteratorError("some message", nil)
	if result != nil {
		t.Errorf("expected nil for nil error, got: %v", result)
	}
}

func TestCovWrapStreamingIteratorError_EOFError(t *testing.T) {
	result := wrapStreamingIteratorError("some message", io.EOF)
	if result != io.EOF {
		t.Errorf("expected io.EOF for io.EOF error, got: %v", result)
	}
}

func TestCovWrapStreamingIteratorError_RealError(t *testing.T) {
	inner := errors.New("boom")
	result := wrapStreamingIteratorError("failed", inner)
	if result == nil {
		t.Fatal("expected non-nil wrapped error")
	}
	assertHarErrorCode(t, result, ErrCodeJSONParse)
	if !errors.Is(result, inner) {
		t.Errorf("expected wrapped error to unwrap to inner, got: %v", result)
	}
}

// ============================================================================
// Extra: ensure direct decoder-based parseHarBasicInfo covers the version
// decode error branch to keep coverage stable.
// ============================================================================

func TestCovParseHarBasicInfo_VersionDecodeError(t *testing.T) {
	// version with an object value fails to decode into string.
	p := writeTempHAR(t, `{"log":{"version":{"x":1},"entries":[]}}`)
	_, err := NewStreamingHarFromFile(p)
	if err == nil {
		t.Fatal("expected error for invalid version value, got nil")
	}
	assertHarErrorCode(t, err, ErrCodeJSONParse)
}

// ============================================================================
// Extra: Next on a bytes-path iterator whose decoder is nil returns false
// (guard branch in Next).
// ============================================================================

func TestCovNext_NilDecoder(t *testing.T) {
	it := &StreamingEntryIterator{}
	if it.Next() {
		t.Error("expected Next() to return false when decoder is nil")
	}
}

// ============================================================================
// Extra: Err() returning a plain (non-EOF) json error still returns it.
// ============================================================================

func TestCovErr_JSONTypeError(t *testing.T) {
	it := &StreamingEntryIterator{err: &json.UnmarshalTypeError{}}
	if it.Err() == nil {
		t.Error("expected non-nil Err() for json.UnmarshalTypeError")
	}
}
