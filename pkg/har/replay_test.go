package har

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestToHTTPRequest(t *testing.T) {
	entry := &Entries{
		Request: Request{
			Method:      "GET",
			URL:         "https://example.com/api/users?id=123",
			HTTPVersion: "HTTP/1.1",
			Headers: []Headers{
				{Name: "Accept", Value: "application/json"},
				{Name: "User-Agent", Value: "Go-HAR-Test"},
			},
			Cookies: []Cookie{
				{Name: "session", Value: "abc123"},
			},
			QueryString: []QueryString{
				{Name: "id", Value: "123"},
			},
		},
	}

	req, err := entry.ToHTTPRequest()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if req.Method != "GET" {
		t.Errorf("Expected method GET, got %s", req.Method)
	}

	if req.URL.String() != "https://example.com/api/users?id=123" {
		t.Errorf("Unexpected URL: %s", req.URL.String())
	}

	if req.Header.Get("Accept") != "application/json" {
		t.Errorf("Expected Accept header, got %s", req.Header.Get("Accept"))
	}
}

func TestToHTTPRequestWithPostData(t *testing.T) {
	entry := &Entries{
		Request: Request{
			Method:      "POST",
			URL:         "https://example.com/api/users",
			HTTPVersion: "HTTP/1.1",
			PostData: &PostData{
				MimeType: "application/json",
				Text:     `{"name": "test"}`,
			},
		},
	}

	req, err := entry.ToHTTPRequest()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if req.Method != "POST" {
		t.Errorf("Expected method POST, got %s", req.Method)
	}

	if req.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type header, got %s", req.Header.Get("Content-Type"))
	}

	body, err := readRequestBody(req)
	if err != nil {
		t.Fatalf("Failed to read body: %v", err)
	}

	if body != `{"name": "test"}` {
		t.Errorf("Expected body content, got %s", body)
	}
}

func TestToHTTPRequestNil(t *testing.T) {
	var entry *Entries
	_, err := entry.ToHTTPRequest()
	if err == nil {
		t.Error("Expected error for nil entry")
	}
}

func TestToHTTPRequestInvalidURL(t *testing.T) {
	entry := &Entries{
		Request: Request{
			Method: "GET",
			URL:    "://invalid-url",
		},
	}

	_, err := entry.ToHTTPRequest()
	if err == nil {
		t.Error("Expected error for invalid URL")
	}
}

func TestReplayWithTestServer(t *testing.T) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	entry := &Entries{
		Request: Request{
			Method:      "GET",
			URL:         server.URL + "/test",
			HTTPVersion: "HTTP/1.1",
			Headers: []Headers{
				{Name: "Accept", Value: "application/json"},
			},
		},
	}

	opts := DefaultReplayOptions()
	opts.Timeout = 5 * time.Second

	result, err := entry.Replay(opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Response == nil {
		t.Fatal("Expected non-nil response")
	}

	if result.Response.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", result.Response.StatusCode)
	}

	if result.Duration == 0 {
		t.Error("Expected non-zero duration")
	}
}

func TestReplayNilEntry(t *testing.T) {
	var entry *Entries
	opts := DefaultReplayOptions()

	_, err := entry.Replay(opts)
	if err == nil {
		t.Error("Expected error for nil entry")
	}
}

func TestReplayAllWithTestServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	h := NewHar()
	e1 := h.AddEntry("GET", server.URL+"/1", "HTTP/1.1", "")
	e2 := h.AddEntry("GET", server.URL+"/2", "HTTP/1.1", "")
	_ = e1
	_ = e2

	opts := DefaultReplayOptions()
	opts.Timeout = 5 * time.Second

	results, err := h.ReplayAll(opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestBuildQueryStringFromURL(t *testing.T) {
	params := BuildQueryStringFromURL("https://example.com/api?name=test&id=123")
	if len(params) != 2 {
		t.Errorf("Expected 2 query params, got %d", len(params))
	}
}

func TestParseResponseHeaders(t *testing.T) {
	headers := ParseResponseHeaders("Content-Type: application/json\nCache-Control: no-cache")
	if len(headers) != 2 {
		t.Errorf("Expected 2 headers, got %d", len(headers))
	}
}

func TestEstimateHeaderSize(t *testing.T) {
	headers := []Headers{
		{Name: "Content-Type", Value: "text/html"},
		{Name: "Content-Length", Value: "100"},
	}
	size := EstimateHeaderSize(headers)
	if size <= 0 {
		t.Errorf("Expected positive size, got %d", size)
	}
}

func TestCloneEntry(t *testing.T) {
	entry := &Entries{
		Request: Request{
			Method: "GET",
			URL:    "https://example.com",
			Headers: []Headers{
				{Name: "Accept", Value: "*/*"},
			},
		},
		Response: Response{
			Status:     200,
			StatusText: "OK",
		},
	}

	cloned := CloneEntry(entry)
	if cloned == nil {
		t.Fatal("Expected non-nil cloned entry")
	}

	if cloned.Request.URL != entry.Request.URL {
		t.Error("Cloned entry should have same URL")
	}

	// Modify clone and check original is unchanged
	cloned.Request.URL = "https://modified.com"
	if entry.Request.URL != "https://example.com" {
		t.Error("Original entry should not be modified")
	}
}

func TestCloneEntryNil(t *testing.T) {
	result := CloneEntry(nil)
	if result != nil {
		t.Error("Expected nil for nil input")
	}
}

func TestWriteRequestToWriter(t *testing.T) {
	entry := &Entries{
		Request: Request{
			Method:      "GET",
			URL:         "https://example.com/api",
			HTTPVersion: "HTTP/1.1",
			Headers: []Headers{
				{Name: "Accept", Value: "application/json"},
			},
		},
	}

	var buf bytes.Buffer
	err := WriteRequestToWriter(entry, &buf)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("Expected non-empty output")
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		size     int
		expected string
	}{
		{100, "100 B"},
		{1024, "1.00 KB"},
		{1048576, "1.00 MB"},
	}

	for _, tt := range tests {
		result := FormatBytes(tt.size)
		if result != tt.expected {
			t.Errorf("FormatBytes(%d) = %s, expected %s", tt.size, result, tt.expected)
		}
	}
}

func TestDefaultReplayOptions(t *testing.T) {
	opts := DefaultReplayOptions()
	if opts.Timeout != 30*time.Second {
		t.Errorf("Expected 30s timeout, got %v", opts.Timeout)
	}
	if !opts.FollowRedirects {
		t.Error("Expected FollowRedirects to be true")
	}
}

func TestHTTPResponseToEntries(t *testing.T) {
	// Create a mock response
	body := `{"status":"ok"}`
	resp := &http.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Proto:      "HTTP/1.1",
		Header: http.Header{
			"Content-Type":   []string{"application/json"},
			"Content-Length": []string{strconv.Itoa(len(body))},
		},
		Body: io.NopCloser(strings.NewReader(body)),
	}

	origEntry := &Entries{
		Request: Request{
			Method: "GET",
			URL:    "https://example.com/test",
		},
	}

	entry := HTTPResponseToEntries(origEntry, resp, 100*time.Millisecond)
	if entry == nil {
		t.Fatal("Expected non-nil entry")
	}

	if entry.Response.Status != 200 {
		t.Errorf("Expected status 200, got %d", entry.Response.Status)
	}

	if entry.Time < 100 {
		t.Errorf("Expected time >= 100ms, got %f", entry.Time)
	}
}

// Helper function to read request body
func readRequestBody(req *http.Request) (string, error) {
	if req.Body == nil {
		return "", nil
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
