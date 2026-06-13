package har

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testHarJSON is a valid minimal HAR document used across tests.
const testHarJSON = `{
  "log": {
    "version": "1.2",
    "creator": {
      "name": "Go-HAR Test",
      "version": "1.0"
    },
    "browser": {
      "name": "TestBrowser",
      "version": "2.0"
    },
    "entries": [
      {
        "startedDateTime": "2024-01-15T10:30:00.000Z",
        "time": 150.5,
        "request": {
          "method": "GET",
          "url": "https://example.com/api",
          "httpVersion": "HTTP/1.1",
          "headers": [],
          "cookies": [],
          "queryString": [],
          "headersSize": 200,
          "bodySize": 0
        },
        "response": {
          "status": 200,
          "statusText": "OK",
          "httpVersion": "HTTP/1.1",
          "headers": [],
          "cookies": [],
          "content": {
            "size": 1234,
            "mimeType": "application/json"
          },
          "redirectURL": "",
          "headersSize": 300,
          "bodySize": 1234
        },
        "cache": {},
        "timings": {
          "blocked": 5,
          "dns": 10,
          "connect": 20,
          "send": 1,
          "wait": 100,
          "receive": 14.5,
          "ssl": 15
        }
      }
    ]
  }
}`

// ---------- ParseHarFromReader ----------

func TestParseHarFromReader(t *testing.T) {
	r := strings.NewReader(testHarJSON)
	har, err := ParseHarFromReader(r)
	if err != nil {
		t.Fatalf("ParseHarFromReader returned error: %v", err)
	}
	if har.Log.Version != "1.2" {
		t.Errorf("expected version 1.2, got %s", har.Log.Version)
	}
	if har.Log.Creator.Name != "Go-HAR Test" {
		t.Errorf("expected creator name 'Go-HAR Test', got %s", har.Log.Creator.Name)
	}
	if har.Log.Browser.Name != "TestBrowser" {
		t.Errorf("expected browser name 'TestBrowser', got %s", har.Log.Browser.Name)
	}
	if len(har.Log.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(har.Log.Entries))
	}
	if har.Log.Entries[0].Request.Method != "GET" {
		t.Errorf("expected method GET, got %s", har.Log.Entries[0].Request.Method)
	}
}

func TestParseHarFromReader_InvalidJSON(t *testing.T) {
	r := strings.NewReader("not json at all")
	_, err := ParseHarFromReader(r)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestParseHarFromReader_Empty(t *testing.T) {
	r := strings.NewReader("")
	_, err := ParseHarFromReader(r)
	if err == nil {
		t.Fatal("expected error for empty input, got nil")
	}
}

// ---------- ParseHarFromReaderWithOptions ----------

func TestParseHarFromReaderWithOptions(t *testing.T) {
	r := strings.NewReader(testHarJSON)
	opts := DefaultParseOptions()
	har, err := ParseHarFromReaderWithOptions(r, opts)
	if err != nil {
		t.Fatalf("ParseHarFromReaderWithOptions returned error: %v", err)
	}
	if har.Log.Version != "1.2" {
		t.Errorf("expected version 1.2, got %s", har.Log.Version)
	}
}

func TestParseHarFromReaderWithOptions_SkipValidation(t *testing.T) {
	r := strings.NewReader(testHarJSON)
	opts := ParseOptions{SkipValidation: true}
	har, err := ParseHarFromReaderWithOptions(r, opts)
	if err != nil {
		t.Fatalf("ParseHarFromReaderWithOptions with SkipValidation returned error: %v", err)
	}
	if har == nil {
		t.Fatal("expected non-nil HAR")
	}
}

// ---------- ParseFromReader ----------

func TestParseFromReader(t *testing.T) {
	r := strings.NewReader(testHarJSON)
	provider, err := ParseFromReader(r)
	if err != nil {
		t.Fatalf("ParseFromReader returned error: %v", err)
	}
	if provider.GetVersion() != "1.2" {
		t.Errorf("expected version 1.2, got %s", provider.GetVersion())
	}
	if provider.GetCreator().Name != "Go-HAR Test" {
		t.Errorf("expected creator name 'Go-HAR Test', got %s", provider.GetCreator().Name)
	}
}

func TestParseFromReader_WithSkipValidation(t *testing.T) {
	r := strings.NewReader(testHarJSON)
	provider, err := ParseFromReader(r, WithSkipValidation())
	if err != nil {
		t.Fatalf("ParseFromReader with SkipValidation returned error: %v", err)
	}
	if provider == nil {
		t.Fatal("expected non-nil HARProvider")
	}
}

func TestParseFromReader_InvalidInput(t *testing.T) {
	r := strings.NewReader("not json")
	_, err := ParseFromReader(r)
	if err == nil {
		t.Fatal("expected error for invalid input, got nil")
	}
}

// ---------- ParseHarFileGzipped ----------

func TestParseHarFileGzipped(t *testing.T) {
	// Create a temporary gzipped HAR file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.har.gz")

	// Write gzipped HAR data
	f, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	gzWriter := gzip.NewWriter(f)
	if _, err := gzWriter.Write([]byte(testHarJSON)); err != nil {
		t.Fatalf("failed to write gzipped data: %v", err)
	}
	gzWriter.Close()
	f.Close()

	// Parse it
	har, err := ParseHarFileGzipped(filePath)
	if err != nil {
		t.Fatalf("ParseHarFileGzipped returned error: %v", err)
	}
	if har.Log.Version != "1.2" {
		t.Errorf("expected version 1.2, got %s", har.Log.Version)
	}
	if har.Log.Browser.Name != "TestBrowser" {
		t.Errorf("expected browser name 'TestBrowser', got %s", har.Log.Browser.Name)
	}
}

func TestParseHarFileGzipped_FileNotFound(t *testing.T) {
	_, err := ParseHarFileGzipped("/nonexistent/path/test.har.gz")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestParseHarFileGzipped_NotGzipped(t *testing.T) {
	// Create a temp file with plain JSON (not gzipped)
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.har.gz")

	if err := os.WriteFile(filePath, []byte(testHarJSON), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	// Should fail because the file is not actually gzipped
	_, err := ParseHarFileGzipped(filePath)
	if err == nil {
		t.Fatal("expected error for non-gzipped file, got nil")
	}
}

// ---------- ParseHarFileAuto ----------

func TestParseHarFileAuto_PlainHar(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.har")

	if err := os.WriteFile(filePath, []byte(testHarJSON), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	har, err := ParseHarFileAuto(filePath)
	if err != nil {
		t.Fatalf("ParseHarFileAuto returned error: %v", err)
	}
	if har.Log.Version != "1.2" {
		t.Errorf("expected version 1.2, got %s", har.Log.Version)
	}
}

func TestParseHarFileAuto_GzippedByExtension(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.har.gz")

	f, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	gzWriter := gzip.NewWriter(f)
	if _, err := gzWriter.Write([]byte(testHarJSON)); err != nil {
		t.Fatalf("failed to write gzipped data: %v", err)
	}
	gzWriter.Close()
	f.Close()

	har, err := ParseHarFileAuto(filePath)
	if err != nil {
		t.Fatalf("ParseHarFileAuto returned error: %v", err)
	}
	if har.Log.Version != "1.2" {
		t.Errorf("expected version 1.2, got %s", har.Log.Version)
	}
}

func TestParseHarFileAuto_GzippedByMagicBytes(t *testing.T) {
	// File with no .har.gz extension but contains gzip magic bytes
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_har_data")

	f, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	gzWriter := gzip.NewWriter(f)
	if _, err := gzWriter.Write([]byte(testHarJSON)); err != nil {
		t.Fatalf("failed to write gzipped data: %v", err)
	}
	gzWriter.Close()
	f.Close()

	har, err := ParseHarFileAuto(filePath)
	if err != nil {
		t.Fatalf("ParseHarFileAuto returned error: %v", err)
	}
	if har.Log.Version != "1.2" {
		t.Errorf("expected version 1.2, got %s", har.Log.Version)
	}
}

func TestParseHarFileAuto_GzipExtension(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.har.gzip")

	f, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	gzWriter := gzip.NewWriter(f)
	if _, err := gzWriter.Write([]byte(testHarJSON)); err != nil {
		t.Fatalf("failed to write gzipped data: %v", err)
	}
	gzWriter.Close()
	f.Close()

	har, err := ParseHarFileAuto(filePath)
	if err != nil {
		t.Fatalf("ParseHarFileAuto returned error: %v", err)
	}
	if har.Log.Version != "1.2" {
		t.Errorf("expected version 1.2, got %s", har.Log.Version)
	}
}

func TestParseHarFileAuto_FileNotFound(t *testing.T) {
	_, err := ParseHarFileAuto("/nonexistent/path/test.har")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

// ---------- NewStreamingParserFromReader ----------

func TestNewStreamingParserFromReader(t *testing.T) {
	r := strings.NewReader(testHarJSON)
	iterator, err := NewStreamingParserFromReader(r)
	if err != nil {
		t.Fatalf("NewStreamingParserFromReader returned error: %v", err)
	}
	defer iterator.Close()

	count := 0
	for iterator.Next() {
		entry := iterator.Entry()
		count++
		if entry.Request.URL != "https://example.com/api" {
			t.Errorf("expected URL 'https://example.com/api', got %s", entry.Request.URL)
		}
	}
	if err := iterator.Err(); err != nil {
		t.Fatalf("iterator error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 entry, got %d", count)
	}
}

func TestNewStreamingParserFromReader_InvalidInput(t *testing.T) {
	r := strings.NewReader("not json")
	_, err := NewStreamingParserFromReader(r)
	if err == nil {
		t.Fatal("expected error for invalid input, got nil")
	}
}

func TestNewStreamingParserFromReader_EmptyInput(t *testing.T) {
	r := strings.NewReader("")
	_, err := NewStreamingParserFromReader(r)
	if err == nil {
		t.Fatal("expected error for empty input, got nil")
	}
}

// ---------- SaveToFileGzipped ----------

func TestSaveToFileGzipped(t *testing.T) {
	// First, parse a valid HAR
	har, err := ParseHar([]byte(testHarJSON))
	if err != nil {
		t.Fatalf("ParseHar returned error: %v", err)
	}

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "output.har.gz")

	// Save as gzipped
	if err := SaveToFileGzipped(har, filePath, true); err != nil {
		t.Fatalf("SaveToFileGzipped returned error: %v", err)
	}

	// Verify the file exists and is gzipped
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if len(data) < 2 {
		t.Fatal("output file is too small")
	}
	// Check gzip magic bytes
	if data[0] != 0x1f || data[1] != 0x8b {
		t.Errorf("expected gzip magic bytes 0x1f 0x8b, got 0x%02x 0x%02x", data[0], data[1])
	}

	// Verify we can read it back
	har2, err := ParseHarFileGzipped(filePath)
	if err != nil {
		t.Fatalf("ParseHarFileGzipped returned error: %v", err)
	}
	if har2.Log.Version != har.Log.Version {
		t.Errorf("round-trip version mismatch: expected %s, got %s", har.Log.Version, har2.Log.Version)
	}
	if len(har2.Log.Entries) != len(har.Log.Entries) {
		t.Errorf("round-trip entries count mismatch: expected %d, got %d", len(har.Log.Entries), len(har2.Log.Entries))
	}
}

func TestSaveToFileGzipped_NoIndent(t *testing.T) {
	har, err := ParseHar([]byte(testHarJSON))
	if err != nil {
		t.Fatalf("ParseHar returned error: %v", err)
	}

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "output_noindent.har.gz")

	if err := SaveToFileGzipped(har, filePath, false); err != nil {
		t.Fatalf("SaveToFileGzipped returned error: %v", err)
	}

	// Verify we can read it back
	har2, err := ParseHarFileGzipped(filePath)
	if err != nil {
		t.Fatalf("ParseHarFileGzipped returned error: %v", err)
	}
	if har2.Log.Version != har.Log.Version {
		t.Errorf("round-trip version mismatch: expected %s, got %s", har.Log.Version, har2.Log.Version)
	}
}

func TestSaveToFileGzipped_NilHar(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "nil.har.gz")

	err := SaveToFileGzipped(nil, filePath, true)
	if err == nil {
		t.Fatal("expected error for nil HAR, got nil")
	}
}

// ---------- WriteToWriter ----------

func TestWriteToWriter_Reader(t *testing.T) {
	har, err := ParseHar([]byte(testHarJSON))
	if err != nil {
		t.Fatalf("ParseHar returned error: %v", err)
	}

	var buf bytes.Buffer
	if err := WriteToWriter(har, &buf, true); err != nil {
		t.Fatalf("WriteToWriter returned error: %v", err)
	}

	// Verify output is valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	// Verify we can parse it back as HAR
	har2, err := ParseHar(buf.Bytes())
	if err != nil {
		t.Fatalf("ParseHar on WriteToWriter output returned error: %v", err)
	}
	if har2.Log.Version != har.Log.Version {
		t.Errorf("round-trip version mismatch: expected %s, got %s", har.Log.Version, har2.Log.Version)
	}
}

func TestWriteToWriter_Reader_NoIndent(t *testing.T) {
	har, err := ParseHar([]byte(testHarJSON))
	if err != nil {
		t.Fatalf("ParseHar returned error: %v", err)
	}

	var buf bytes.Buffer
	if err := WriteToWriter(har, &buf, false); err != nil {
		t.Fatalf("WriteToWriter returned error: %v", err)
	}

	// Compact JSON should not have leading spaces after { or [
	output := buf.String()
	if strings.Contains(output, "\n  ") {
		t.Error("expected compact JSON (no indentation), but found indentation")
	}
}

func TestWriteToWriter_Reader_NilHar(t *testing.T) {
	var buf bytes.Buffer
	err := WriteToWriter(nil, &buf, true)
	if err == nil {
		t.Fatal("expected error for nil HAR, got nil")
	}
}

// ---------- Streaming entries from file-based streaming ----------

func TestStreamingHarFromFile_Entries(t *testing.T) {
	// Use the testdata minimal.har file
	testFilePath := filepath.Join("testdata", "minimal.har")
	if _, err := os.Stat(testFilePath); os.IsNotExist(err) {
		t.Skip("testdata/minimal.har not found")
	}

	streamingHar, err := NewStreamingHarFromFile(testFilePath)
	if err != nil {
		t.Fatalf("NewStreamingHarFromFile returned error: %v", err)
	}
	defer streamingHar.Close()

	// Test basic info
	if streamingHar.GetVersion() == "" {
		t.Error("expected non-empty version")
	}

	// Test iteration over entries (minimal.har has 0 entries)
	iterator := streamingHar.Entries()
	defer iterator.Close()

	count := 0
	for iterator.Next() {
		count++
	}
	if err := iterator.Err(); err != nil {
		t.Errorf("iterator error: %v", err)
	}
	// minimal.har has empty entries array, so count should be 0
}

func TestStreamingHarFromBytes_Browser(t *testing.T) {
	data := []byte(testHarJSON)
	streamingHar, err := NewStreamingHarFromBytes(data)
	if err != nil {
		t.Fatalf("NewStreamingHarFromBytes returned error: %v", err)
	}
	defer streamingHar.Close()

	browser := streamingHar.GetBrowser()
	if browser.Name != "TestBrowser" {
		t.Errorf("expected browser name 'TestBrowser', got '%s'", browser.Name)
	}
	if browser.Version != "2.0" {
		t.Errorf("expected browser version '2.0', got '%s'", browser.Version)
	}
}

func TestStreamingHarFromFile_WithEntries(t *testing.T) {
	// Create a temp HAR file with entries
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "streaming_test.har")

	harJSON := `{
	  "log": {
	    "version": "1.2",
	    "creator": {"name": "Test", "version": "1.0"},
	    "entries": [
	      {
	        "startedDateTime": "2024-01-15T10:30:00.000Z",
	        "time": 100,
	        "request": {
	          "method": "GET",
	          "url": "https://example.com/page1",
	          "httpVersion": "HTTP/1.1",
	          "headers": [],
	          "cookies": [],
	          "queryString": [],
	          "headersSize": 100,
	          "bodySize": 0
	        },
	        "response": {
	          "status": 200,
	          "statusText": "OK",
	          "httpVersion": "HTTP/1.1",
	          "headers": [],
	          "cookies": [],
	          "content": {"size": 500, "mimeType": "text/html"},
	          "redirectURL": "",
	          "headersSize": 200,
	          "bodySize": 500
	        },
	        "cache": {},
	        "timings": {"blocked": 0, "dns": 0, "connect": 0, "send": 0, "wait": 50, "receive": 50, "ssl": 0}
	      },
	      {
	        "startedDateTime": "2024-01-15T10:30:01.000Z",
	        "time": 200,
	        "request": {
	          "method": "POST",
	          "url": "https://example.com/api",
	          "httpVersion": "HTTP/1.1",
	          "headers": [],
	          "cookies": [],
	          "queryString": [],
	          "headersSize": 150,
	          "bodySize": 50
	        },
	        "response": {
	          "status": 201,
	          "statusText": "Created",
	          "httpVersion": "HTTP/1.1",
	          "headers": [],
	          "cookies": [],
	          "content": {"size": 100, "mimeType": "application/json"},
	          "redirectURL": "",
	          "headersSize": 200,
	          "bodySize": 100
	        },
	        "cache": {},
	        "timings": {"blocked": 0, "dns": 0, "connect": 0, "send": 1, "wait": 150, "receive": 49, "ssl": 0}
	      }
	    ]
	  }
	}`

	if err := os.WriteFile(filePath, []byte(harJSON), 0644); err != nil {
		t.Fatalf("failed to write temp HAR file: %v", err)
	}

	streamingHar, err := NewStreamingHarFromFile(filePath)
	if err != nil {
		t.Fatalf("NewStreamingHarFromFile returned error: %v", err)
	}
	defer streamingHar.Close()

	iterator := streamingHar.Entries()
	defer iterator.Close()

	count := 0
	var urls []string
	for iterator.Next() {
		entry := iterator.Entry()
		count++
		urls = append(urls, entry.Request.URL)
	}
	if err := iterator.Err(); err != nil {
		t.Fatalf("iterator error: %v", err)
	}

	if count != 2 {
		t.Errorf("expected 2 entries, got %d", count)
	}
	if len(urls) >= 2 {
		if urls[0] != "https://example.com/page1" {
			t.Errorf("expected first URL 'https://example.com/page1', got %s", urls[0])
		}
		if urls[1] != "https://example.com/api" {
			t.Errorf("expected second URL 'https://example.com/api', got %s", urls[1])
		}
	}
}

// ---------- Round-trip gzip save/load ----------

func TestRoundTripGzipped(t *testing.T) {
	// Parse -> SaveToFileGzipped -> ParseHarFileGzipped -> compare
	har, err := ParseHar([]byte(testHarJSON))
	if err != nil {
		t.Fatalf("ParseHar returned error: %v", err)
	}

	tmpDir := t.TempDir()
	gzPath := filepath.Join(tmpDir, "roundtrip.har.gz")

	// Save with indent
	if err := SaveToFileGzipped(har, gzPath, true); err != nil {
		t.Fatalf("SaveToFileGzipped returned error: %v", err)
	}

	// Load back via auto-detect
	har2, err := ParseHarFileAuto(gzPath)
	if err != nil {
		t.Fatalf("ParseHarFileAuto returned error: %v", err)
	}

	if !har.Equals(har2) {
		t.Error("round-trip HAR objects are not equal")
	}
}

// ---------- Helper function tests ----------

func TestIsGzippedByExtension(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"test.har.gz", true},
		{"test.har.gzip", true},
		{"test.har", false},
		{"test.json", false},
		{"/some/path/test.har.gz", true},
		{"test.gz", false},
	}

	for _, tc := range tests {
		result := isGzippedByExtension(tc.path)
		if result != tc.expected {
			t.Errorf("isGzippedByExtension(%q) = %v, expected %v", tc.path, result, tc.expected)
		}
	}
}

func TestDetectGzipMagicBytes(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a gzipped file
	gzPath := filepath.Join(tmpDir, "test.gz")
	f, _ := os.Create(gzPath)
	gzWriter := gzip.NewWriter(f)
	_, _ = gzWriter.Write([]byte("hello"))
	gzWriter.Close()
	f.Close()

	isGz, err := detectGzipMagicBytes(gzPath)
	if err != nil {
		t.Fatalf("detectGzipMagicBytes returned error: %v", err)
	}
	if !isGz {
		t.Error("expected gzipped file to be detected as gzip")
	}

	// Create a plain text file
	plainPath := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(plainPath, []byte("plain text"), 0644)

	isGz, err = detectGzipMagicBytes(plainPath)
	if err != nil {
		t.Fatalf("detectGzipMagicBytes returned error: %v", err)
	}
	if isGz {
		t.Error("expected plain text file not to be detected as gzip")
	}
}

func TestParseHarFromReader_WithFileReader(t *testing.T) {
	// Use an actual file as the io.Reader
	testFilePath := filepath.Join("testdata", "minimal.har")
	if _, err := os.Stat(testFilePath); os.IsNotExist(err) {
		t.Skip("testdata/minimal.har not found")
	}

	f, err := os.Open(testFilePath)
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer f.Close()

	har, err := ParseHarFromReader(f)
	if err != nil {
		t.Fatalf("ParseHarFromReader from file returned error: %v", err)
	}
	if har.Log.Version == "" {
		t.Error("expected non-empty version")
	}
}
