package har

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// --- HarBuilder: AddPage (0% coverage) ---

func TestBuilderAddPage(t *testing.T) {
	h := NewHarBuilder().
		AddPage("page_1", "Home").
		AddPage("page_2", "About").
		Build()

	if len(h.Log.Pages) != 2 {
		t.Fatalf("Expected 2 pages, got %d", len(h.Log.Pages))
	}
	if h.Log.Pages[0].ID != "page_1" {
		t.Errorf("Expected page ID 'page_1', got %s", h.Log.Pages[0].ID)
	}
	if h.Log.Pages[0].Title != "Home" {
		t.Errorf("Expected page title 'Home', got %s", h.Log.Pages[0].Title)
	}
	if h.Log.Pages[1].ID != "page_2" {
		t.Errorf("Expected page ID 'page_2', got %s", h.Log.Pages[1].ID)
	}
	if h.Log.Pages[1].Title != "About" {
		t.Errorf("Expected page title 'About', got %s", h.Log.Pages[1].Title)
	}
}

// --- HarBuilder: AddEntryWithHTTPVersion (0% coverage) ---

func TestBuilderAddEntryWithHTTPVersion(t *testing.T) {
	h := NewHarBuilder().
		AddEntryWithHTTPVersion("GET", "https://example.com", "HTTP/2.0").
		WithResponseStatus(200, "OK").
		EndEntry().
		Build()

	if len(h.Log.Entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(h.Log.Entries))
	}
	entry := h.Log.Entries[0]
	if entry.Request.HTTPVersion != "HTTP/2.0" {
		t.Errorf("Expected request HTTP version 'HTTP/2.0', got %s", entry.Request.HTTPVersion)
	}
	if entry.Response.HTTPVersion != "HTTP/2.0" {
		t.Errorf("Expected response HTTP version 'HTTP/2.0', got %s", entry.Response.HTTPVersion)
	}
}

// --- HarBuilder: AddEntryForPage (0% coverage) ---

func TestBuilderAddEntryForPage(t *testing.T) {
	h := NewHarBuilder().
		AddPage("page_1", "Home").
		AddEntryForPage("GET", "https://example.com", "page_1").
		WithResponseStatus(200, "OK").
		EndEntry().
		Build()

	if len(h.Log.Entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(h.Log.Entries))
	}
	entry := h.Log.Entries[0]
	if entry.Pageref != "page_1" {
		t.Errorf("Expected pageref 'page_1', got %s", entry.Pageref)
	}
}

// --- HarBuilder: AddEntryFromHTTP nil request (branch coverage) ---

func TestBuilderAddEntryFromHTTPNilRequest(t *testing.T) {
	h := NewHarBuilder().
		AddEntryFromHTTP(nil, nil, 100*time.Millisecond).
		Build()

	if len(h.Log.Entries) != 0 {
		t.Errorf("Expected 0 entries for nil request, got %d", len(h.Log.Entries))
	}
}

// --- HarBuilder: AddEntryFromHTTP with request body (branch coverage) ---

func TestBuilderAddEntryFromHTTPWithRequestBody(t *testing.T) {
	body := strings.NewReader(`{"key": "value"}`)
	req, _ := http.NewRequest("POST", "https://example.com/api", body)
	req.Header.Set("Content-Type", "application/json")

	h := NewHarBuilder().
		AddEntryFromHTTP(req, nil, 50*time.Millisecond).
		Build()

	if len(h.Log.Entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(h.Log.Entries))
	}
	entry := h.Log.Entries[0]

	if entry.Request.PostData == nil {
		t.Fatal("Expected PostData to be set for POST request with body")
	}
	if entry.Request.PostData.MimeType != "application/json" {
		t.Errorf("Expected PostData mimeType 'application/json', got %s", entry.Request.PostData.MimeType)
	}
	if entry.Request.PostData.Text != `{"key": "value"}` {
		t.Errorf("Expected PostData text '{\"key\": \"value\"}', got %s", entry.Request.PostData.Text)
	}
	if entry.Request.BodySize != len(`{"key": "value"}`) {
		t.Errorf("Expected BodySize %d, got %d", len(`{"key": "value"}`), entry.Request.BodySize)
	}
}

// --- HarBuilder: AddEntryFromHTTP with request cookies ---

func TestBuilderAddEntryFromHTTPWithRequestCookies(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://example.com/api", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "abc123", Path: "/", Domain: "example.com", HttpOnly: true, Secure: true})

	h := NewHarBuilder().
		AddEntryFromHTTP(req, nil, 50*time.Millisecond).
		Build()

	if len(h.Log.Entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(h.Log.Entries))
	}
	entry := h.Log.Entries[0]

	if len(entry.Request.Cookies) != 1 {
		t.Fatalf("Expected 1 request cookie, got %d", len(entry.Request.Cookies))
	}
	cookie := entry.Request.Cookies[0]
	if cookie.Name != "session" {
		t.Errorf("Expected cookie name 'session', got %s", cookie.Name)
	}
	if cookie.Value != "abc123" {
		t.Errorf("Expected cookie value 'abc123', got %s", cookie.Value)
	}
	// Note: Go's http.Request.Cookies() strips Path, Domain, HttpOnly, Secure
	// because cookies in request headers don't carry these attributes.
	// Only Set-Cookie response headers contain these.
}

// --- HarBuilder: AddEntryFromHTTP with response cookies and body ---

func TestBuilderAddEntryFromHTTPWithResponseBodyAndCookies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Set-Cookie", "tracking=xyz789; Path=/; HttpOnly")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"hello": "world"}`))
	}))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL+"/api", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	h := NewHarBuilder().
		AddEntryFromHTTP(req, resp, 50*time.Millisecond).
		Build()

	if len(h.Log.Entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(h.Log.Entries))
	}
	entry := h.Log.Entries[0]

	// Check response body
	if entry.Response.Content.Size != len(`{"hello": "world"}`) {
		t.Errorf("Expected content size %d, got %d", len(`{"hello": "world"}`), entry.Response.Content.Size)
	}
	if entry.Response.Content.Text != `{"hello": "world"}` {
		t.Errorf("Expected content text, got %s", entry.Response.Content.Text)
	}
	if entry.Response.BodySize != len(`{"hello": "world"}`) {
		t.Errorf("Expected body size %d, got %d", len(`{"hello": "world"}`), entry.Response.BodySize)
	}

	// Check response headers
	if len(entry.Response.Headers) == 0 {
		t.Error("Expected response headers to be populated")
	}

	// Check response cookies
	if len(entry.Response.Cookies) == 0 {
		t.Error("Expected response cookies to be populated")
	}
}

// --- HarBuilder: AddEntryFromHTTP with empty body (branch: len(bodyBytes) == 0) ---

func TestBuilderAddEntryFromHTTPEmptyRequestBody(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://example.com/api", nil)

	h := NewHarBuilder().
		AddEntryFromHTTP(req, nil, 50*time.Millisecond).
		Build()

	if len(h.Log.Entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(h.Log.Entries))
	}
	entry := h.Log.Entries[0]

	// GET with nil body should not set PostData
	if entry.Request.PostData != nil {
		t.Error("Expected PostData to be nil for GET request with no body")
	}
}

// --- HarBuilder: AddEntryFromHTTP with nil response body ---

func TestBuilderAddEntryFromHTTPNilResponseBody(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://example.com/api", nil)
	// resp with nil Body
	resp := &http.Response{
		StatusCode: 204,
		Status:     "204 No Content",
		Proto:      "HTTP/1.1",
		Header:     http.Header{},
		Body:       http.NoBody,
	}

	h := NewHarBuilder().
		AddEntryFromHTTP(req, resp, 10*time.Millisecond).
		Build()

	if len(h.Log.Entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(h.Log.Entries))
	}
	entry := h.Log.Entries[0]

	if entry.Response.Status != 204 {
		t.Errorf("Expected status 204, got %d", entry.Response.Status)
	}
	if entry.Response.StatusText != "204 No Content" {
		t.Errorf("Expected status text '204 No Content', got %s", entry.Response.StatusText)
	}
}

// --- HarBuilder: BuildAndSave (0% coverage) ---

func TestBuilderBuildAndSave(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "har-builder-save-*.har")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	err = NewHarBuilder().
		SetCreator("test", "1.0").
		AddEntry("GET", "https://example.com").
		WithResponseStatus(200, "OK").
		EndEntry().
		BuildAndSave(tmpPath, true)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify the file was written and is valid JSON
	data, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}
	if len(data) == 0 {
		t.Error("Expected non-empty file output")
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Errorf("Saved file is not valid JSON: %v", err)
	}
}

func TestBuilderBuildAndSaveNoIndent(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "har-builder-save-*.har")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	err = NewHarBuilder().
		SetCreator("test", "1.0").
		AddEntry("GET", "https://example.com").
		WithResponseStatus(200, "OK").
		EndEntry().
		BuildAndSave(tmpPath, false)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	data, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}
	if len(data) == 0 {
		t.Error("Expected non-empty file output")
	}
}

func TestBuilderBuildAndSaveInvalidPath(t *testing.T) {
	err := NewHarBuilder().
		SetCreator("test", "1.0").
		AddEntry("GET", "https://example.com").
		WithResponseStatus(200, "OK").
		EndEntry().
		BuildAndSave("/nonexistent/dir/that/does/not/exist/file.har", true)

	if err == nil {
		t.Error("Expected error when saving to invalid path")
	}
}

// --- EntryBuilder: WithHTTPVersion (0% coverage) ---

func TestBuilderWithHTTPVersion(t *testing.T) {
	h := NewHarBuilder().
		AddEntry("GET", "https://example.com").
		WithHTTPVersion("HTTP/2.0").
		WithResponseStatus(200, "OK").
		EndEntry().
		Build()

	entry := h.Log.Entries[0]
	if entry.Request.HTTPVersion != "HTTP/2.0" {
		t.Errorf("Expected request HTTP version 'HTTP/2.0', got %s", entry.Request.HTTPVersion)
	}
	if entry.Response.HTTPVersion != "HTTP/2.0" {
		t.Errorf("Expected response HTTP version 'HTTP/2.0', got %s", entry.Response.HTTPVersion)
	}
}

// --- EntryBuilder: WithStartedDateTime (0% coverage) ---

func TestBuilderWithStartedDateTime(t *testing.T) {
	ts := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	h := NewHarBuilder().
		AddEntry("GET", "https://example.com").
		WithStartedDateTime(ts).
		WithResponseStatus(200, "OK").
		EndEntry().
		Build()

	entry := h.Log.Entries[0]
	if !entry.StartedDateTime.Equal(ts) {
		t.Errorf("Expected started date time %v, got %v", ts, entry.StartedDateTime)
	}
}

// --- EntryBuilder: WithPageref (0% coverage) ---

func TestBuilderWithPageref(t *testing.T) {
	h := NewHarBuilder().
		AddPage("page_1", "Home").
		AddEntry("GET", "https://example.com").
		WithPageref("page_1").
		WithResponseStatus(200, "OK").
		EndEntry().
		Build()

	entry := h.Log.Entries[0]
	if entry.Pageref != "page_1" {
		t.Errorf("Expected pageref 'page_1', got %s", entry.Pageref)
	}
}

// --- EntryBuilder: WithConnection (0% coverage) ---

func TestBuilderWithConnection(t *testing.T) {
	h := NewHarBuilder().
		AddEntry("GET", "https://example.com").
		WithConnection("conn-123").
		WithResponseStatus(200, "OK").
		EndEntry().
		Build()

	entry := h.Log.Entries[0]
	if entry.Connection != "conn-123" {
		t.Errorf("Expected connection 'conn-123', got %s", entry.Connection)
	}
}

// --- EntryBuilder: WithPostDataParams (0% coverage) ---

func TestBuilderWithPostDataParams(t *testing.T) {
	params := []Param{
		{Name: "username", Value: "john"},
		{Name: "password", Value: "secret"},
	}
	h := NewHarBuilder().
		AddEntry("POST", "https://example.com/login").
		WithPostDataParams("application/x-www-form-urlencoded", params).
		WithResponseStatus(200, "OK").
		EndEntry().
		Build()

	entry := h.Log.Entries[0]
	if entry.Request.PostData == nil {
		t.Fatal("Expected PostData to be set")
	}
	if entry.Request.PostData.MimeType != "application/x-www-form-urlencoded" {
		t.Errorf("Expected mimeType 'application/x-www-form-urlencoded', got %s", entry.Request.PostData.MimeType)
	}
	if len(entry.Request.PostData.Params) != 2 {
		t.Fatalf("Expected 2 params, got %d", len(entry.Request.PostData.Params))
	}
	if entry.Request.PostData.Params[0].Name != "username" {
		t.Errorf("Expected param name 'username', got %s", entry.Request.PostData.Params[0].Name)
	}
	if entry.Request.PostData.Params[1].Value != "secret" {
		t.Errorf("Expected param value 'secret', got %s", entry.Request.PostData.Params[1].Value)
	}
}

// --- EntryBuilder: WithResponseContentText (0% coverage) ---

func TestBuilderWithResponseContentText(t *testing.T) {
	h := NewHarBuilder().
		AddEntry("GET", "https://example.com/api").
		WithResponseContentText(100, "application/json", `{"status": "ok"}`).
		EndEntry().
		Build()

	entry := h.Log.Entries[0]
	if entry.Response.Content.Size != 100 {
		t.Errorf("Expected content size 100, got %d", entry.Response.Content.Size)
	}
	if entry.Response.Content.MimeType != "application/json" {
		t.Errorf("Expected content mimeType 'application/json', got %s", entry.Response.Content.MimeType)
	}
	if entry.Response.Content.Text != `{"status": "ok"}` {
		t.Errorf("Expected content text '{\"status\": \"ok\"}', got %s", entry.Response.Content.Text)
	}
}

// --- EntryBuilder: WithCache (0% coverage) ---

func TestBuilderWithCache(t *testing.T) {
	cache := Cache{
		BeforeRequest: &BeforeRequest{
			ETag:       "abc123",
			HitCount:   2,
			LastAccess: time.Now(),
		},
		AfterRequest: &AfterRequest{
			ETag:       "def456",
			HitCount:   3,
			LastAccess: time.Now(),
		},
	}
	h := NewHarBuilder().
		AddEntry("GET", "https://example.com").
		WithCache(cache).
		WithResponseStatus(200, "OK").
		EndEntry().
		Build()

	entry := h.Log.Entries[0]
	if entry.Cache.BeforeRequest == nil {
		t.Fatal("Expected BeforeRequest to be set")
	}
	if entry.Cache.BeforeRequest.ETag != "abc123" {
		t.Errorf("Expected BeforeRequest ETag 'abc123', got %s", entry.Cache.BeforeRequest.ETag)
	}
	if entry.Cache.AfterRequest == nil {
		t.Fatal("Expected AfterRequest to be set")
	}
	if entry.Cache.AfterRequest.ETag != "def456" {
		t.Errorf("Expected AfterRequest ETag 'def456', got %s", entry.Cache.AfterRequest.ETag)
	}
}

// --- Recorder: SetBrowser (0% coverage) ---

func TestRecorderSetBrowser(t *testing.T) {
	recorder := NewRecorder().
		SetBrowser("Chrome", "120.0")

	h := recorder.ToHar()
	if h.Log.Browser.Name != "Chrome" {
		t.Errorf("Expected browser name 'Chrome', got %s", h.Log.Browser.Name)
	}
	if h.Log.Browser.Version != "120.0" {
		t.Errorf("Expected browser version '120.0', got %s", h.Log.Browser.Version)
	}
}

// --- Recorder: SaveToFile (0% coverage) ---

func TestRecorderSaveToFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "har-recorder-save-*.har")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	recorder := NewRecorder().SetCreator("test-recorder", "1.0")
	recorder.CaptureEntry(Entries{
		Request:  Request{Method: "GET", URL: "https://example.com"},
		Response: Response{Status: 200, StatusText: "OK"},
	})

	err = recorder.SaveToFile(tmpPath)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	data, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}
	if len(data) == 0 {
		t.Error("Expected non-empty file output")
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Errorf("Saved file is not valid JSON: %v", err)
	}
}

// --- Recorder: ToJSON (0% coverage) ---

func TestRecorderToJSON(t *testing.T) {
	recorder := NewRecorder().SetCreator("test-recorder", "1.0")
	recorder.CaptureEntry(Entries{
		Request:  Request{Method: "GET", URL: "https://example.com"},
		Response: Response{Status: 200, StatusText: "OK"},
	})

	// With indent
	jsonData, err := recorder.ToJSON(true)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(jsonData) == 0 {
		t.Error("Expected non-empty JSON output with indent")
	}

	// Without indent
	jsonData, err = recorder.ToJSON(false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(jsonData) == 0 {
		t.Error("Expected non-empty JSON output without indent")
	}
}

// --- WriteToWriter: error path (writer fails) ---

func TestWriteToWriterWriteError(t *testing.T) {
	h := NewHarBuilder().
		SetCreator("test", "1.0").
		AddEntry("GET", "https://example.com").
		WithResponseStatus(200, "OK").
		EndEntry().
		Build()

	// Use a writer that always fails
	errWriter := &failingWriter{err: errors.New("write failure")}
	err := WriteToWriter(h, errWriter, false)
	if err == nil {
		t.Error("Expected error from failing writer")
	}
	assertHarErrorCode(t, err, ErrCodeFileSystem)
	if !errors.Is(err, errWriter.err) {
		t.Fatalf("expected wrapped write error, got %v", err)
	}
}

func TestWriteToWriterShortWrite(t *testing.T) {
	h := NewHarBuilder().
		SetCreator("test", "1.0").
		AddEntry("GET", "https://example.com").
		WithResponseStatus(200, "OK").
		EndEntry().
		Build()

	err := WriteToWriter(h, &shortWriter{}, false)
	assertHarErrorCode(t, err, ErrCodeFileSystem)
	if !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("expected wrapped short write error, got %v", err)
	}
}

// failingWriter is a test helper that always fails on Write
type failingWriter struct {
	err error
}

func (w *failingWriter) Write(p []byte) (int, error) {
	return 0, w.err
}

// --- WriteEntriesToWriter: nil HAR (error path) ---

func TestWriteEntriesToWriterNil(t *testing.T) {
	var buf bytes.Buffer
	err := WriteEntriesToWriter(nil, &buf)
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
}

func TestWriteEntriesToWriterNilWriter(t *testing.T) {
	h := NewHar()
	h.AddEntry("GET", "https://example.com", "HTTP/1.1", "")

	var nilWriter io.Writer
	var typedNilBuffer *bytes.Buffer
	tests := []struct {
		name   string
		writer io.Writer
	}{
		{name: "nil interface", writer: nilWriter},
		{name: "typed nil", writer: typedNilBuffer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WriteEntriesToWriter(h, tt.writer)
			assertHarErrorCode(t, err, ErrCodeInvalidFormat)
		})
	}
}

// --- WriteEntriesToWriter: normal operation ---

func TestWriteEntriesToWriter(t *testing.T) {
	h := NewHarBuilder().
		AddEntry("GET", "https://example.com/1").
		WithResponseStatus(200, "OK").
		EndEntry().
		AddEntry("GET", "https://example.com/2").
		WithResponseStatus(404, "Not Found").
		EndEntry().
		Build()

	var buf bytes.Buffer
	err := WriteEntriesToWriter(h, &buf)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Expected non-empty output")
	}

	// Verify we can decode two JSON values from the stream.
	decoder := json.NewDecoder(&buf)
	count := 0
	for {
		var entry Entries
		err := decoder.Decode(&entry)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Failed to decode entry %d: %v", count, err)
		}
		count++
	}
	if count != 2 {
		t.Errorf("Expected 2 entries, got %d", count)
	}
}

// --- WriteEntriesToWriter: encoder error path ---

func TestWriteEntriesToWriterEncoderError(t *testing.T) {
	h := NewHarBuilder().
		AddEntry("GET", "https://example.com").
		WithResponseStatus(200, "OK").
		EndEntry().
		Build()

	errWriter := &failingWriter{err: errors.New("encode failure")}
	err := WriteEntriesToWriter(h, errWriter)
	if err == nil {
		t.Error("Expected error from failing writer during encoding")
	}
	assertHarErrorCode(t, err, ErrCodeFileSystem)
	if !errors.Is(err, errWriter.err) {
		t.Fatalf("expected wrapped encode error, got %v", err)
	}
}

func TestWriteEntriesToWriterShortWrite(t *testing.T) {
	h := NewHarBuilder().
		AddEntry("GET", "https://example.com").
		WithResponseStatus(200, "OK").
		EndEntry().
		Build()

	err := WriteEntriesToWriter(h, &shortWriter{})
	assertHarErrorCode(t, err, ErrCodeFileSystem)
	if !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("expected wrapped short write error, got %v", err)
	}
}

func TestWriteEntriesToWriterMarshalError(t *testing.T) {
	h := NewHar()
	entry := h.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	entry.SetCustomField("_bad", func() {})

	var buf bytes.Buffer
	err := WriteEntriesToWriter(h, &buf)
	assertHarErrorCode(t, err, ErrCodeJSONParse)
}

// --- ReadEntriesFromReader (0% coverage) ---

func TestReadEntriesFromReader(t *testing.T) {
	// Build some JSON lines
	h := NewHarBuilder().
		AddEntry("GET", "https://example.com/1").
		WithResponseStatus(200, "OK").
		EndEntry().
		AddEntry("POST", "https://example.com/2").
		WithResponseStatus(201, "Created").
		EndEntry().
		Build()

	var buf bytes.Buffer
	if err := WriteEntriesToWriter(h, &buf); err != nil {
		t.Fatalf("Failed to write entries: %v", err)
	}

	entries, err := ReadEntriesFromReader(&buf)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(entries))
	}

	if entries[0].Request.Method != "GET" {
		t.Errorf("Expected first entry method GET, got %s", entries[0].Request.Method)
	}
	if entries[0].Response.Status != 200 {
		t.Errorf("Expected first entry status 200, got %d", entries[0].Response.Status)
	}
	if entries[1].Request.Method != "POST" {
		t.Errorf("Expected second entry method POST, got %s", entries[1].Request.Method)
	}
	if entries[1].Response.Status != 201 {
		t.Errorf("Expected second entry status 201, got %d", entries[1].Response.Status)
	}
}

func TestReadEntriesFromReaderNilReader(t *testing.T) {
	var nilReader io.Reader
	var typedNilBuffer *bytes.Buffer
	tests := []struct {
		name   string
		reader io.Reader
	}{
		{name: "nil interface", reader: nilReader},
		{name: "typed nil", reader: typedNilBuffer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries, err := ReadEntriesFromReader(tt.reader)
			assertHarErrorCode(t, err, ErrCodeInvalidFormat)
			if entries != nil {
				t.Fatalf("expected nil entries for nil reader, got %#v", entries)
			}
		})
	}
}

func TestReadEntriesFromReaderEmpty(t *testing.T) {
	entries, err := ReadEntriesFromReader(strings.NewReader(""))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries from empty reader, got %d", len(entries))
	}
}

func TestReadEntriesFromReaderInvalidJSON(t *testing.T) {
	entries, err := ReadEntriesFromReader(strings.NewReader("not valid json\n"))
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
	assertHarErrorCode(t, err, ErrCodeJSONParse)
	// Partial entries should still be returned (none decoded before error)
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries before error, got %d", len(entries))
	}
}

func TestReadEntriesFromReaderPartialOnInvalidJSON(t *testing.T) {
	entry := Entries{
		Request: Request{
			Method:      "GET",
			URL:         "https://example.com/ok",
			HTTPVersion: "HTTP/1.1",
		},
		Response: Response{
			Status:     200,
			StatusText: "OK",
		},
	}

	validLine, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("Failed to marshal entry: %v", err)
	}

	entries, err := ReadEntriesFromReader(strings.NewReader(string(validLine) + "\nnot valid json\n"))
	if err == nil {
		t.Fatal("Expected error for invalid JSON after valid entry")
	}
	assertHarErrorCode(t, err, ErrCodeJSONParse)
	if len(entries) != 1 {
		t.Fatalf("Expected 1 partial entry before error, got %d", len(entries))
	}
	if entries[0].Request.URL != "https://example.com/ok" {
		t.Fatalf("Unexpected partial entry URL: %s", entries[0].Request.URL)
	}
}

func TestHarBuilderNilReceiverMethods(t *testing.T) {
	var builder *HarBuilder

	if got := builder.SetVersion("1.2"); got != nil {
		t.Fatalf("SetVersion() = %#v, want nil", got)
	}
	if got := builder.SetCreator("test", "1.0"); got != nil {
		t.Fatalf("SetCreator() = %#v, want nil", got)
	}
	if got := builder.SetBrowser("browser", "1.0"); got != nil {
		t.Fatalf("SetBrowser() = %#v, want nil", got)
	}
	if got := builder.SetComment("comment"); got != nil {
		t.Fatalf("SetComment() = %#v, want nil", got)
	}
	if got := builder.AddPage("page", "Page"); got != nil {
		t.Fatalf("AddPage() = %#v, want nil", got)
	}
	if got := builder.AddEntry("GET", "https://example.com"); got != nil {
		t.Fatalf("AddEntry() = %#v, want nil", got)
	}
	if got := builder.AddEntryWithHTTPVersion("GET", "https://example.com", "HTTP/2.0"); got != nil {
		t.Fatalf("AddEntryWithHTTPVersion() = %#v, want nil", got)
	}
	if got := builder.AddEntryForPage("GET", "https://example.com", "page"); got != nil {
		t.Fatalf("AddEntryForPage() = %#v, want nil", got)
	}
	if got := builder.AddEntryFromHTTP(nil, nil, time.Second); got != nil {
		t.Fatalf("AddEntryFromHTTP() = %#v, want nil", got)
	}
	if got := builder.Build(); got != nil {
		t.Fatalf("Build() = %#v, want nil", got)
	}

	data, err := builder.BuildJSON(false)
	if data != nil {
		t.Fatalf("BuildJSON() data = %#v, want nil", data)
	}
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)

	err = builder.BuildAndSave("unused.har", false)
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
}

func TestHarBuilderZeroValueMethods(t *testing.T) {
	var builder HarBuilder

	har := builder.
		SetCreator("test", "1.0").
		AddPage("page", "Page").
		AddEntry("GET", "https://example.com").
		WithResponseStatus(200, "OK").
		EndEntry().
		Build()

	if har == nil {
		t.Fatal("Build() returned nil for zero-value builder")
	}
	if har.Log.Creator.Name != "test" {
		t.Fatalf("creator = %q, want test", har.Log.Creator.Name)
	}
	if len(har.Log.Pages) != 1 || len(har.Log.Entries) != 1 {
		t.Fatalf("pages/entries = %d/%d, want 1/1", len(har.Log.Pages), len(har.Log.Entries))
	}
}

func TestEntryBuilderNilReceiverMethods(t *testing.T) {
	var entryBuilder *EntryBuilder

	if got := entryBuilder.WithHTTPVersion("HTTP/2.0"); got != nil {
		t.Fatalf("WithHTTPVersion() = %#v, want nil", got)
	}
	if got := entryBuilder.WithStartedDateTime(time.Now()); got != nil {
		t.Fatalf("WithStartedDateTime() = %#v, want nil", got)
	}
	if got := entryBuilder.WithPageref("page"); got != nil {
		t.Fatalf("WithPageref() = %#v, want nil", got)
	}
	if got := entryBuilder.WithServerIP("127.0.0.1"); got != nil {
		t.Fatalf("WithServerIP() = %#v, want nil", got)
	}
	if got := entryBuilder.WithConnection("conn"); got != nil {
		t.Fatalf("WithConnection() = %#v, want nil", got)
	}
	if got := entryBuilder.WithComment("comment"); got != nil {
		t.Fatalf("WithComment() = %#v, want nil", got)
	}
	if got := entryBuilder.AddRequestHeader("A", "B"); got != nil {
		t.Fatalf("AddRequestHeader() = %#v, want nil", got)
	}
	if got := entryBuilder.AddResponseHeader("A", "B"); got != nil {
		t.Fatalf("AddResponseHeader() = %#v, want nil", got)
	}
	if got := entryBuilder.AddCookie("a", "b"); got != nil {
		t.Fatalf("AddCookie() = %#v, want nil", got)
	}
	if got := entryBuilder.AddResponseCookie("a", "b"); got != nil {
		t.Fatalf("AddResponseCookie() = %#v, want nil", got)
	}
	if got := entryBuilder.AddQueryParam("a", "b"); got != nil {
		t.Fatalf("AddQueryParam() = %#v, want nil", got)
	}
	if got := entryBuilder.WithPostData("text/plain", "body"); got != nil {
		t.Fatalf("WithPostData() = %#v, want nil", got)
	}
	if got := entryBuilder.WithPostDataParams("application/x-www-form-urlencoded", nil); got != nil {
		t.Fatalf("WithPostDataParams() = %#v, want nil", got)
	}
	if got := entryBuilder.WithResponseStatus(200, "OK"); got != nil {
		t.Fatalf("WithResponseStatus() = %#v, want nil", got)
	}
	if got := entryBuilder.WithResponseContent(10, "text/plain"); got != nil {
		t.Fatalf("WithResponseContent() = %#v, want nil", got)
	}
	if got := entryBuilder.WithResponseContentText(10, "text/plain", "body"); got != nil {
		t.Fatalf("WithResponseContentText() = %#v, want nil", got)
	}
	if got := entryBuilder.WithTimings(0, 0, 0, 0, 0, 0, 0); got != nil {
		t.Fatalf("WithTimings() = %#v, want nil", got)
	}
	if got := entryBuilder.WithCache(Cache{}); got != nil {
		t.Fatalf("WithCache() = %#v, want nil", got)
	}
	if got := entryBuilder.WithInitiator("script", "https://example.com/app.js", 1); got != nil {
		t.Fatalf("WithInitiator() = %#v, want nil", got)
	}
	if got := entryBuilder.WithPriority("high"); got != nil {
		t.Fatalf("WithPriority() = %#v, want nil", got)
	}
	if got := entryBuilder.WithResourceType("xhr"); got != nil {
		t.Fatalf("WithResourceType() = %#v, want nil", got)
	}
	if got := entryBuilder.EndEntry(); got != nil {
		t.Fatalf("EndEntry() = %#v, want nil", got)
	}
}

func TestRecorderNilReceiverMethods(t *testing.T) {
	var recorder *Recorder

	if got := recorder.SetCreator("test", "1.0"); got != nil {
		t.Fatalf("SetCreator() = %#v, want nil", got)
	}
	if got := recorder.SetBrowser("browser", "1.0"); got != nil {
		t.Fatalf("SetBrowser() = %#v, want nil", got)
	}
	if got := recorder.Capture(nil, nil, time.Second); got != nil {
		t.Fatalf("Capture() = %#v, want nil", got)
	}
	if got := recorder.CaptureEntry(Entries{}); got != nil {
		t.Fatalf("CaptureEntry() = %#v, want nil", got)
	}
	if got := recorder.EntryCount(); got != 0 {
		t.Fatalf("EntryCount() = %d, want 0", got)
	}
	if got := recorder.ToHar(); got != nil {
		t.Fatalf("ToHar() = %#v, want nil", got)
	}

	data, err := recorder.ToJSON(false)
	if data != nil {
		t.Fatalf("ToJSON() data = %#v, want nil", data)
	}
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)

	err = recorder.SaveToFile("unused.har")
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
}

func TestRecorderZeroValueMethods(t *testing.T) {
	var recorder Recorder

	recorder.SetCreator("test", "1.0").CaptureEntry(Entries{
		Request: Request{Method: "GET", URL: "https://example.com"},
	})

	if recorder.EntryCount() != 1 {
		t.Fatalf("EntryCount() = %d, want 1", recorder.EntryCount())
	}
	har := recorder.ToHar()
	if har == nil {
		t.Fatal("ToHar() returned nil for zero-value recorder")
	}
	if har.Log.Creator.Name != "test" {
		t.Fatalf("creator = %q, want test", har.Log.Creator.Name)
	}
}

// --- ToJSONLines: nil HAR (branch coverage) ---

func TestToJSONLinesNil(t *testing.T) {
	var h *Har
	lines, err := h.ToJSONLines()
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
	if lines != "" {
		t.Errorf("Expected empty string for nil HAR, got %q", lines)
	}
}

// --- ToJSONLines: with entries (additional coverage) ---

func TestToJSONLinesWithEntries(t *testing.T) {
	h := NewHarBuilder().
		AddEntry("GET", "https://example.com/1").
		WithResponseStatus(200, "OK").
		EndEntry().
		AddEntry("GET", "https://example.com/2").
		WithResponseStatus(404, "Not Found").
		EndEntry().
		Build()

	lines, err := h.ToJSONLines()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if lines == "" {
		t.Error("Expected non-empty JSON lines output")
	}

	// Each line should be valid JSON
	lineCount := 0
	for _, line := range strings.Split(strings.TrimSpace(lines), "\n") {
		if line == "" {
			continue
		}
		var entry Entries
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Errorf("Failed to unmarshal line %d: %v", lineCount, err)
		}
		lineCount++
	}
	if lineCount != 2 {
		t.Errorf("Expected 2 JSON lines, got %d", lineCount)
	}
}

// --- EntryBuilder chaining with all methods ---

func TestBuilderEntryChainingAllMethods(t *testing.T) {
	ts := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	params := []Param{{Name: "field1", Value: "val1"}}
	cache := Cache{
		BeforeRequest: &BeforeRequest{ETag: "etag-before", HitCount: 1, LastAccess: ts},
		AfterRequest:  &AfterRequest{ETag: "etag-after", HitCount: 2, LastAccess: ts},
	}

	h := NewHarBuilder().
		SetVersion("1.2").
		SetCreator("test", "1.0").
		SetBrowser("Chrome", "100.0").
		SetComment("chaining test").
		AddPage("page_1", "Home").
		AddEntryForPage("GET", "https://example.com", "page_1").
		WithHTTPVersion("HTTP/1.0").
		WithStartedDateTime(ts).
		WithPageref("page_1").
		WithConnection("conn-456").
		WithServerIP("10.0.0.1").
		WithComment("entry comment").
		AddRequestHeader("Accept", "text/html").
		AddResponseHeader("Content-Type", "text/html").
		AddCookie("session", "val").
		AddResponseCookie("tracking", "xyz").
		AddQueryParam("q", "test").
		WithPostDataParams("multipart/form-data", params).
		WithResponseStatus(200, "OK").
		WithResponseContentText(42, "text/html", "<html>hello</html>").
		WithTimings(1, 2, 3, 4, 5, 6, 7).
		WithCache(cache).
		WithInitiator("parser", "https://example.com/index.html", 10).
		WithPriority("High").
		WithResourceType("document").
		EndEntry().
		AddEntryWithHTTPVersion("POST", "https://example.com/api", "HTTP/2.0").
		WithPostData("application/json", `{"key":"val"}`).
		WithResponseStatus(201, "Created").
		WithResponseContent(256, "application/json").
		EndEntry().
		Build()

	if len(h.Log.Pages) != 1 {
		t.Errorf("Expected 1 page, got %d", len(h.Log.Pages))
	}
	if len(h.Log.Entries) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(h.Log.Entries))
	}

	entry := h.Log.Entries[0]
	if entry.Request.HTTPVersion != "HTTP/1.0" {
		t.Errorf("Expected HTTP/1.0, got %s", entry.Request.HTTPVersion)
	}
	if entry.Response.HTTPVersion != "HTTP/1.0" {
		t.Errorf("Expected response HTTP/1.0, got %s", entry.Response.HTTPVersion)
	}
	if !entry.StartedDateTime.Equal(ts) {
		t.Errorf("Expected started date time %v, got %v", ts, entry.StartedDateTime)
	}
	if entry.Pageref != "page_1" {
		t.Errorf("Expected pageref 'page_1', got %s", entry.Pageref)
	}
	if entry.Connection != "conn-456" {
		t.Errorf("Expected connection 'conn-456', got %s", entry.Connection)
	}
	if entry.ServerIPAddress != "10.0.0.1" {
		t.Errorf("Expected server IP '10.0.0.1', got %s", entry.ServerIPAddress)
	}
	if entry.Comment != "entry comment" {
		t.Errorf("Expected comment 'entry comment', got %s", entry.Comment)
	}
	if entry.Request.PostData == nil {
		t.Fatal("Expected PostData to be set")
	}
	if len(entry.Request.PostData.Params) != 1 {
		t.Errorf("Expected 1 PostData param, got %d", len(entry.Request.PostData.Params))
	}
	if entry.Response.Content.Text != "<html>hello</html>" {
		t.Errorf("Expected response content text '<html>hello</html>', got %s", entry.Response.Content.Text)
	}
	if entry.Cache.BeforeRequest == nil || entry.Cache.BeforeRequest.ETag != "etag-before" {
		t.Error("Expected Cache.BeforeRequest to be set with ETag 'etag-before'")
	}
	if entry.Initiator.Type != "parser" {
		t.Errorf("Expected initiator type 'parser', got %s", entry.Initiator.Type)
	}
	if entry.Priority != "High" {
		t.Errorf("Expected priority 'High', got %s", entry.Priority)
	}
	if entry.ResourceType != "document" {
		t.Errorf("Expected resource type 'document', got %s", entry.ResourceType)
	}

	// Second entry
	entry2 := h.Log.Entries[1]
	if entry2.Request.HTTPVersion != "HTTP/2.0" {
		t.Errorf("Expected HTTP/2.0, got %s", entry2.Request.HTTPVersion)
	}
	if entry2.Request.PostData == nil {
		t.Fatal("Expected PostData to be set on second entry")
	}
	if entry2.Request.PostData.MimeType != "application/json" {
		t.Errorf("Expected mimeType 'application/json', got %s", entry2.Request.PostData.MimeType)
	}
}

// --- Existing tests (preserved) ---

func TestHarBuilderBasic(t *testing.T) {
	har := NewHarBuilder().
		SetVersion("1.2").
		SetCreator("test", "1.0").
		SetBrowser("Chrome", "100.0").
		SetComment("test comment").
		Build()

	if har.Log.Version != "1.2" {
		t.Errorf("Expected version 1.2, got %s", har.Log.Version)
	}

	if har.Log.Creator.Name != "test" {
		t.Errorf("Expected creator name 'test', got %s", har.Log.Creator.Name)
	}

	if har.Log.Browser.Name != "Chrome" {
		t.Errorf("Expected browser name 'Chrome', got %s", har.Log.Browser.Name)
	}

	if har.Log.Comment != "test comment" {
		t.Errorf("Expected comment 'test comment', got %s", har.Log.Comment)
	}
}

func TestEntryBuilder(t *testing.T) {
	har := NewHarBuilder().
		SetCreator("test", "1.0").
		AddEntry("GET", "https://example.com/api/users").
		AddRequestHeader("Accept", "application/json").
		AddRequestHeader("Authorization", "Bearer token123").
		AddCookie("session", "abc123").
		AddQueryParam("page", "1").
		AddQueryParam("limit", "10").
		WithResponseStatus(200, "OK").
		WithResponseContent(1024, "application/json").
		AddResponseHeader("Content-Type", "application/json").
		AddResponseCookie("tracking", "xyz789").
		WithTimings(10, 5, 15, 2, 50, 30, 8).
		WithServerIP("1.2.3.4").
		WithComment("API call").
		EndEntry().
		Build()

	if len(har.Log.Entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(har.Log.Entries))
	}

	entry := har.Log.Entries[0]

	if entry.Request.Method != "GET" {
		t.Errorf("Expected GET, got %s", entry.Request.Method)
	}

	if entry.Request.URL != "https://example.com/api/users" {
		t.Errorf("Unexpected URL: %s", entry.Request.URL)
	}

	if len(entry.Request.Headers) != 2 {
		t.Errorf("Expected 2 request headers, got %d", len(entry.Request.Headers))
	}

	if len(entry.Request.Cookies) != 1 {
		t.Errorf("Expected 1 request cookie, got %d", len(entry.Request.Cookies))
	}

	if len(entry.Request.QueryString) != 2 {
		t.Errorf("Expected 2 query params, got %d", len(entry.Request.QueryString))
	}

	if entry.Response.Status != 200 {
		t.Errorf("Expected status 200, got %d", entry.Response.Status)
	}

	if entry.ServerIPAddress != "1.2.3.4" {
		t.Errorf("Expected server IP '1.2.3.4', got %s", entry.ServerIPAddress)
	}

	if entry.Comment != "API call" {
		t.Errorf("Expected comment 'API call', got %s", entry.Comment)
	}
}

func TestEntryBuilderWithPostData(t *testing.T) {
	har := NewHarBuilder().
		AddEntry("POST", "https://example.com/api/users").
		WithPostData("application/json", `{"name": "test"}`).
		WithResponseStatus(201, "Created").
		EndEntry().
		Build()

	entry := har.Log.Entries[0]

	if entry.Request.PostData == nil {
		t.Fatal("Expected PostData to be set")
	}

	if entry.Request.PostData.MimeType != "application/json" {
		t.Errorf("Expected mimeType 'application/json', got %s", entry.Request.PostData.MimeType)
	}

	if entry.Request.PostData.Text != `{"name": "test"}` {
		t.Errorf("Unexpected PostData text: %s", entry.Request.PostData.Text)
	}
}

func TestHarBuilderMultipleEntries(t *testing.T) {
	har := NewHarBuilder().
		AddEntry("GET", "https://example.com/1").
		WithResponseStatus(200, "OK").
		EndEntry().
		AddEntry("GET", "https://example.com/2").
		WithResponseStatus(404, "Not Found").
		EndEntry().
		AddEntry("POST", "https://example.com/3").
		WithResponseStatus(201, "Created").
		EndEntry().
		Build()

	if len(har.Log.Entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(har.Log.Entries))
	}

	if har.Log.Entries[0].Response.Status != 200 {
		t.Errorf("Expected first entry status 200, got %d", har.Log.Entries[0].Response.Status)
	}

	if har.Log.Entries[1].Response.Status != 404 {
		t.Errorf("Expected second entry status 404, got %d", har.Log.Entries[1].Response.Status)
	}
}

func TestHarBuilderBuildJSON(t *testing.T) {
	jsonData, err := NewHarBuilder().
		SetCreator("test", "1.0").
		AddEntry("GET", "https://example.com").
		WithResponseStatus(200, "OK").
		EndEntry().
		BuildJSON(true)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("Expected non-empty JSON output")
	}
}

func TestRecorder(t *testing.T) {
	recorder := NewRecorder().
		SetCreator("recorder-test", "1.0")

	if recorder.EntryCount() != 0 {
		t.Errorf("Expected 0 entries, got %d", recorder.EntryCount())
	}

	// Add an entry manually
	recorder.CaptureEntry(Entries{
		Request: Request{
			Method: "GET",
			URL:    "https://example.com/test",
		},
		Response: Response{
			Status:     200,
			StatusText: "OK",
		},
	})

	if recorder.EntryCount() != 1 {
		t.Errorf("Expected 1 entry, got %d", recorder.EntryCount())
	}

	har := recorder.ToHar()
	if len(har.Log.Entries) != 1 {
		t.Errorf("Expected 1 entry in HAR, got %d", len(har.Log.Entries))
	}
}

func TestRecorderWithHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	recorder := NewRecorder()

	req, _ := http.NewRequest("GET", server.URL+"/test", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	recorder.Capture(req, resp, 100*time.Millisecond)

	if recorder.EntryCount() != 1 {
		t.Errorf("Expected 1 entry, got %d", recorder.EntryCount())
	}
}

func TestWriteToWriter(t *testing.T) {
	har := NewHarBuilder().
		SetCreator("test", "1.0").
		AddEntry("GET", "https://example.com").
		WithResponseStatus(200, "OK").
		EndEntry().
		Build()

	var buf bytes.Buffer
	err := WriteToWriter(har, &buf, false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Expected non-empty output")
	}
}

func TestWriteToWriterNil(t *testing.T) {
	var buf bytes.Buffer
	err := WriteToWriter(nil, &buf, false)
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
}

func TestWriteToWriterNilWriter(t *testing.T) {
	har := NewHar()
	har.SetCreator("test", "1.0")

	var nilWriter io.Writer
	var typedNilBuffer *bytes.Buffer
	tests := []struct {
		name   string
		writer io.Writer
	}{
		{name: "nil interface", writer: nilWriter},
		{name: "typed nil", writer: typedNilBuffer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WriteToWriter(har, tt.writer, false)
			assertHarErrorCode(t, err, ErrCodeInvalidFormat)
		})
	}
}

func TestToJSONLines(t *testing.T) {
	h := NewHar()
	e1 := h.AddEntry("GET", "https://example.com/1", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e2 := h.AddEntry("GET", "https://example.com/2", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")

	lines, err := h.ToJSONLines()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if lines == "" {
		t.Error("Expected non-empty JSON lines output")
	}
}

func TestAddEntryFromHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"hello": "world"}`))
	}))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL+"/api", nil)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	har := NewHarBuilder().
		SetCreator("test", "1.0").
		AddEntryFromHTTP(req, resp, 50*time.Millisecond).
		Build()

	if len(har.Log.Entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(har.Log.Entries))
	}

	entry := har.Log.Entries[0]

	if entry.Request.Method != "GET" {
		t.Errorf("Expected GET method, got %s", entry.Request.Method)
	}

	if entry.Response.Status != 200 {
		t.Errorf("Expected status 200, got %d", entry.Response.Status)
	}
}

func TestEntryBuilderWithInitiator(t *testing.T) {
	har := NewHarBuilder().
		AddEntry("GET", "https://example.com/script.js").
		WithInitiator("script", "https://example.com/index.html", 42).
		WithPriority("High").
		WithResourceType("script").
		WithResponseStatus(200, "OK").
		EndEntry().
		Build()

	entry := har.Log.Entries[0]

	if entry.Initiator.Type != "script" {
		t.Errorf("Expected initiator type 'script', got %s", entry.Initiator.Type)
	}

	if entry.Priority != "High" {
		t.Errorf("Expected priority 'High', got %s", entry.Priority)
	}

	if entry.ResourceType != "script" {
		t.Errorf("Expected resource type 'script', got %s", entry.ResourceType)
	}
}
