package har

import (
	"bytes"
	"errors"
	"fmt"
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

func TestToHTTPRequestPostDataNoText(t *testing.T) {
	// PostData exists but Text is empty — body should be nil
	entry := &Entries{
		Request: Request{
			Method:      "POST",
			URL:         "https://example.com/api",
			HTTPVersion: "HTTP/1.1",
			PostData: &PostData{
				MimeType: "application/json",
				Text:     "",
			},
		},
	}

	req, err := entry.ToHTTPRequest()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if req.Body != nil {
		t.Error("Expected nil body when PostData.Text is empty")
	}

	// Content-Type should still be set from PostData.MimeType
	if req.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type to be set, got %q", req.Header.Get("Content-Type"))
	}
}

func TestToHTTPRequestNoPostData(t *testing.T) {
	// No PostData at all
	entry := &Entries{
		Request: Request{
			Method:      "GET",
			URL:         "https://example.com/api",
			HTTPVersion: "HTTP/1.1",
		},
	}

	req, err := entry.ToHTTPRequest()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if req.Body != nil {
		t.Error("Expected nil body when no PostData")
	}
}

func TestToHTTPRequestWithCookies(t *testing.T) {
	entry := &Entries{
		Request: Request{
			Method:      "GET",
			URL:         "https://example.com/api",
			HTTPVersion: "HTTP/1.1",
			Cookies: []Cookie{
				{Name: "session", Value: "abc", Path: "/", Domain: "example.com", HTTPOnly: true, Secure: true},
				{Name: "token", Value: "xyz"},
			},
		},
	}

	req, err := entry.ToHTTPRequest()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	cookies := req.Cookies()
	if len(cookies) != 2 {
		t.Errorf("Expected 2 cookies, got %d", len(cookies))
	}
}

func TestToHTTPRequestPostDataNoMimeType(t *testing.T) {
	// PostData with Text but no MimeType — Content-Type should not be set
	entry := &Entries{
		Request: Request{
			Method:      "POST",
			URL:         "https://example.com/api",
			HTTPVersion: "HTTP/1.1",
			PostData: &PostData{
				Text: "raw body data",
			},
		},
	}

	req, err := entry.ToHTTPRequest()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if req.Header.Get("Content-Type") != "" {
		t.Errorf("Expected no Content-Type when MimeType is empty, got %q", req.Header.Get("Content-Type"))
	}

	body, err := readRequestBody(req)
	if err != nil {
		t.Fatalf("Failed to read body: %v", err)
	}
	if body != "raw body data" {
		t.Errorf("Expected body 'raw body data', got %q", body)
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

// ===================== NEW COMPREHENSIVE TESTS =====================

// --- Replay: error path when ToHTTPRequest fails ---

func TestReplayToHTTPRequestError(t *testing.T) {
	entry := &Entries{
		Request: Request{
			Method: "GET",
			URL:    "://invalid-url",
		},
	}

	opts := DefaultReplayOptions()
	result, err := entry.Replay(opts)
	if err == nil {
		t.Error("Expected error when ToHTTPRequest fails")
	}
	if result != nil {
		t.Error("Expected nil result when ToHTTPRequest fails")
	}
}

// --- Replay: network error path (connection refused) ---

func TestReplayNetworkError(t *testing.T) {
	entry := &Entries{
		Request: Request{
			Method: "GET",
			URL:    "http://127.0.0.1:1/unreachable", // port 1 should be refused
		},
	}

	opts := DefaultReplayOptions()
	opts.Timeout = 2 * time.Second

	result, err := entry.Replay(opts)
	if err == nil {
		t.Error("Expected error for unreachable server")
	}
	if result == nil {
		t.Fatal("Expected non-nil result even on error")
	}
	if result.Entry == nil {
		t.Error("Expected Entry to be set in error result")
	}
	if result.Error == nil {
		t.Error("Expected Error to be set in error result")
	}
	if result.Duration == 0 {
		t.Error("Expected non-zero duration even on error")
	}
}

// --- Replay: override headers ---

func TestReplayOverrideHeaders(t *testing.T) {
	var receivedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.WriteHeader(200)
	}))
	defer server.Close()

	entry := &Entries{
		Request: Request{
			Method:      "GET",
			URL:         server.URL + "/test",
			HTTPVersion: "HTTP/1.1",
			Headers: []Headers{
				{Name: "Authorization", Value: "Bearer old-token"},
			},
		},
	}

	opts := DefaultReplayOptions()
	opts.OverrideHeaders = map[string]string{
		"Authorization": "Bearer new-token",
	}

	result, err := entry.Replay(opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Response.StatusCode != 200 {
		t.Errorf("Expected 200, got %d", result.Response.StatusCode)
	}
	if receivedAuth != "Bearer new-token" {
		t.Errorf("Expected overridden auth header, got %q", receivedAuth)
	}
}

// --- Replay: POST with body ---

func TestReplayPostWithBody(t *testing.T) {
	var receivedMethod, receivedBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		bodyBytes, _ := io.ReadAll(r.Body)
		receivedBody = string(bodyBytes)
		w.WriteHeader(201)
	}))
	defer server.Close()

	entry := &Entries{
		Request: Request{
			Method:      "POST",
			URL:         server.URL + "/submit",
			HTTPVersion: "HTTP/1.1",
			PostData: &PostData{
				MimeType: "application/json",
				Text:     `{"key":"value"}`,
			},
		},
	}

	opts := DefaultReplayOptions()
	result, err := entry.Replay(opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Response.StatusCode != 201 {
		t.Errorf("Expected 201, got %d", result.Response.StatusCode)
	}
	if receivedMethod != "POST" {
		t.Errorf("Expected POST, got %s", receivedMethod)
	}
	if receivedBody != `{"key":"value"}` {
		t.Errorf("Expected body content, got %q", receivedBody)
	}
}

// --- ReplayAll: nil Har ---

func TestReplayAllNilHar(t *testing.T) {
	var h *Har
	opts := DefaultReplayOptions()
	_, err := h.ReplayAll(opts)
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
}

func TestReplayAllInvalidRequestKeepsResult(t *testing.T) {
	h := NewHar()
	h.AddEntry("GET", "://invalid-url", "HTTP/1.1", "")

	results, err := h.ReplayAll(DefaultReplayOptions())
	assertHarErrorCode(t, err, ErrCodeInvalidValue)
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0] == nil {
		t.Fatal("Expected non-nil result for invalid request")
	}
	if results[0].Index != 0 {
		t.Errorf("Expected result index 0, got %d", results[0].Index)
	}
	if results[0].Error == nil {
		t.Error("Expected result to retain replay error")
	}
}

// --- ReplayAll: mixed success and failure ---

func TestReplayAllMixedResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer server.Close()

	h := NewHar()
	h.AddEntry("GET", server.URL+"/ok", "HTTP/1.1", "")
	h.AddEntry("GET", "http://127.0.0.1:1/fail", "HTTP/1.1", "") // unreachable

	opts := DefaultReplayOptions()
	opts.Timeout = 2 * time.Second

	results, err := h.ReplayAll(opts)
	if err == nil {
		t.Error("Expected error from the unreachable request")
	}
	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}
	// First should succeed
	if results[0].Error != nil {
		t.Errorf("Expected first result to succeed, got error: %v", results[0].Error)
	}
	if results[0].Index != 0 {
		t.Errorf("Expected index 0, got %d", results[0].Index)
	}
	// Second should fail
	if results[1].Error == nil {
		t.Error("Expected second result to fail")
	}
	if results[1].Index != 1 {
		t.Errorf("Expected index 1, got %d", results[1].Index)
	}
}

// --- ReplayAll: verify index assignment ---

func TestReplayAllIndexAssignment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer server.Close()

	h := NewHar()
	h.AddEntry("GET", server.URL+"/1", "HTTP/1.1", "")
	h.AddEntry("GET", server.URL+"/2", "HTTP/1.1", "")
	h.AddEntry("GET", server.URL+"/3", "HTTP/1.1", "")

	opts := DefaultReplayOptions()
	opts.Timeout = 5 * time.Second

	results, err := h.ReplayAll(opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	for i, r := range results {
		if r.Index != i {
			t.Errorf("Result %d has Index=%d, expected %d", i, r.Index, i)
		}
	}
}

// --- ReplaySelective: basic with matching filter ---

func TestReplaySelectiveBasic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer server.Close()

	h := NewHar()
	h.AddEntry("GET", server.URL+"/api/users", "HTTP/1.1", "")
	h.AddEntry("POST", server.URL+"/api/submit", "HTTP/1.1", "")
	h.AddEntry("GET", server.URL+"/other/page", "HTTP/1.1", "")

	opts := DefaultReplayOptions()
	opts.Timeout = 5 * time.Second

	filterOpts := FilterOptions{
		URL: "/api/",
	}

	results, err := h.ReplaySelective(opts, filterOpts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results matching /api/, got %d", len(results))
	}
	for _, r := range results {
		if r.Response == nil || r.Response.StatusCode != 200 {
			t.Error("Expected successful response for matched entries")
		}
	}
}

// --- ReplaySelective: no matching entries returns nil ---

func TestReplaySelectiveNoMatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer server.Close()

	h := NewHar()
	h.AddEntry("GET", server.URL+"/api/users", "HTTP/1.1", "")

	opts := DefaultReplayOptions()
	filterOpts := FilterOptions{
		URL: "/nonexistent/",
	}

	results, err := h.ReplaySelective(opts, filterOpts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if results != nil {
		t.Errorf("Expected nil results when no entries match, got %d results", len(results))
	}
}

// --- ReplaySelective: with method filter ---

func TestReplaySelectiveMethodFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer server.Close()

	h := NewHar()
	h.AddEntry("GET", server.URL+"/1", "HTTP/1.1", "")
	h.AddEntry("POST", server.URL+"/2", "HTTP/1.1", "")
	h.AddEntry("GET", server.URL+"/3", "HTTP/1.1", "")

	opts := DefaultReplayOptions()
	opts.Timeout = 5 * time.Second

	filterOpts := FilterOptions{
		Method: "POST",
	}

	results, err := h.ReplaySelective(opts, filterOpts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result for POST filter, got %d", len(results))
	}
}

// --- ReplaySelective: error path ---

func TestReplaySelectiveError(t *testing.T) {
	h := NewHar()
	h.AddEntry("GET", "http://127.0.0.1:1/fail", "HTTP/1.1", "") // unreachable

	opts := DefaultReplayOptions()
	opts.Timeout = 2 * time.Second

	filterOpts := FilterOptions{
		URL: "/fail",
	}

	results, err := h.ReplaySelective(opts, filterOpts)
	if err == nil {
		t.Error("Expected error from unreachable server")
	}
	if results == nil {
		t.Fatal("Expected results even when requests fail")
	}
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0].Error == nil {
		t.Error("Expected error on the result entry")
	}
}

func TestReplaySelectiveNilHar(t *testing.T) {
	var h *Har

	results, err := h.ReplaySelective(DefaultReplayOptions(), FilterOptions{})
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
	if results != nil {
		t.Fatalf("Expected nil results for nil HAR, got %d", len(results))
	}
}

func TestReplaySelectiveInvalidRequestKeepsResult(t *testing.T) {
	h := NewHar()
	h.AddEntry("GET", "://invalid-url", "HTTP/1.1", "")

	results, err := h.ReplaySelective(DefaultReplayOptions(), FilterOptions{})
	assertHarErrorCode(t, err, ErrCodeInvalidValue)
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0] == nil {
		t.Fatal("Expected non-nil result for invalid request")
	}
	if results[0].Index != 0 {
		t.Errorf("Expected result index 0, got %d", results[0].Index)
	}
	if results[0].Error == nil {
		t.Error("Expected result to retain replay error")
	}
}

// --- ReplayResultsToHar: with successful responses ---

func TestReplayResultsToHarWithResponses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	entry := &Entries{
		Request: Request{
			Method:      "GET",
			URL:         server.URL + "/test",
			HTTPVersion: "HTTP/1.1",
		},
	}

	opts := DefaultReplayOptions()
	opts.Timeout = 5 * time.Second

	result, err := entry.Replay(opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	har := ReplayResultsToHar([]*ReplayResult{result})
	if har == nil {
		t.Fatal("Expected non-nil Har")
	}
	if len(har.Log.Entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(har.Log.Entries))
	}
	// The entry should have response data from the server
	if har.Log.Entries[0].Response.Status != 200 {
		t.Errorf("Expected status 200, got %d", har.Log.Entries[0].Response.Status)
	}
}

// --- ReplayResultsToHar: with nil results ---

func TestReplayResultsToHarNilResult(t *testing.T) {
	har := ReplayResultsToHar([]*ReplayResult{nil})
	if har == nil {
		t.Fatal("Expected non-nil Har")
	}
	if len(har.Log.Entries) != 0 {
		t.Errorf("Expected 0 entries for nil result, got %d", len(har.Log.Entries))
	}
}

// --- ReplayResultsToHar: with error result (no response, but has Entry) ---

func TestReplayResultsToHarErrorResult(t *testing.T) {
	entry := &Entries{
		Request: Request{
			Method: "GET",
			URL:    "http://127.0.0.1:1/unreachable",
		},
	}

	result := &ReplayResult{
		Entry:    entry,
		Error:    errors.New("connection refused"),
		Duration: 100 * time.Millisecond,
	}

	har := ReplayResultsToHar([]*ReplayResult{result})
	if har == nil {
		t.Fatal("Expected non-nil Har")
	}
	if len(har.Log.Entries) != 1 {
		t.Fatalf("Expected 1 entry for error result with Entry, got %d", len(har.Log.Entries))
	}
	// Should contain original entry since Response is nil but Entry is not
	if har.Log.Entries[0].Request.URL != "http://127.0.0.1:1/unreachable" {
		t.Errorf("Expected original entry URL, got %s", har.Log.Entries[0].Request.URL)
	}
}

// --- ReplayResultsToHar: mixed nil and non-nil results ---

func TestReplayResultsToHarMixed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer server.Close()

	entry := &Entries{
		Request: Request{
			Method:      "GET",
			URL:         server.URL + "/test",
			HTTPVersion: "HTTP/1.1",
		},
	}

	opts := DefaultReplayOptions()
	opts.Timeout = 5 * time.Second
	result, _ := entry.Replay(opts)

	har := ReplayResultsToHar([]*ReplayResult{nil, result, nil})
	if len(har.Log.Entries) != 1 {
		t.Errorf("Expected 1 entry (nil results skipped), got %d", len(har.Log.Entries))
	}
}

// --- ReplayResultsToHar: creator is set ---

func TestReplayResultsToHarCreator(t *testing.T) {
	har := ReplayResultsToHar([]*ReplayResult{})
	if har.Log.Creator.Name != "go-har-replay" {
		t.Errorf("Expected creator name 'go-har-replay', got %q", har.Log.Creator.Name)
	}
	if har.Log.Creator.Version != "1.0" {
		t.Errorf("Expected creator version '1.0', got %q", har.Log.Creator.Version)
	}
}

// --- ReadBody: nil entry ---

func TestReadBodyNilEntry(t *testing.T) {
	data, err := ReadBody(nil)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if data != nil {
		t.Errorf("Expected nil for nil entry, got %v", data)
	}
}

// --- ReadBody: nil PostData ---

func TestReadBodyNilPostData(t *testing.T) {
	entry := &Entries{
		Request: Request{
			Method: "GET",
			URL:    "https://example.com",
		},
	}

	data, err := ReadBody(entry)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if data != nil {
		t.Errorf("Expected nil for nil PostData, got %v", data)
	}
}

// --- ReadBody: with PostData text ---

func TestReadBodyWithText(t *testing.T) {
	entry := &Entries{
		Request: Request{
			Method: "POST",
			URL:    "https://example.com/api",
			PostData: &PostData{
				MimeType: "application/json",
				Text:     `{"hello":"world"}`,
			},
		},
	}

	data, err := ReadBody(entry)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if string(data) != `{"hello":"world"}` {
		t.Errorf("Expected body text, got %q", string(data))
	}
}

// --- ReadBody: with empty PostData text ---

func TestReadBodyEmptyText(t *testing.T) {
	entry := &Entries{
		Request: Request{
			Method: "POST",
			URL:    "https://example.com/api",
			PostData: &PostData{
				MimeType: "text/plain",
				Text:     "",
			},
		},
	}

	data, err := ReadBody(entry)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("Expected empty bytes for empty text, got %q", string(data))
	}
}

// --- createHTTPClient: default transport ---

func TestCreateHTTPClientDefaultTransport(t *testing.T) {
	opts := DefaultReplayOptions()
	client := createHTTPClient(opts)

	if client.Timeout != 30*time.Second {
		t.Errorf("Expected 30s timeout, got %v", client.Timeout)
	}
	if client.Transport == nil {
		t.Error("Expected non-nil transport with default options")
	}
	if client.CheckRedirect == nil {
		t.Error("Expected non-nil CheckRedirect when default MaxRedirects > 0")
	}
}

// --- createHTTPClient: custom transport ---

func TestCreateHTTPClientCustomTransport(t *testing.T) {
	customTransport := &http.Transport{}
	opts := ReplayOptions{
		Timeout:         10 * time.Second,
		FollowRedirects: true,
		Transport:       customTransport,
	}

	client := createHTTPClient(opts)
	if client.Transport != customTransport {
		t.Error("Expected custom transport to be used")
	}
}

func TestCreateHTTPClientTypedNilTransportUsesDefault(t *testing.T) {
	var typedNilTransport *mockRoundTripper
	opts := ReplayOptions{
		Timeout:         10 * time.Second,
		FollowRedirects: true,
		Transport:       typedNilTransport,
	}

	client := createHTTPClient(opts)
	if client.Transport == nil {
		t.Fatal("Expected default transport for typed-nil transport")
	}
	if _, ok := client.Transport.(*http.Transport); !ok {
		t.Fatalf("Expected default *http.Transport, got %T", client.Transport)
	}
}

// --- createHTTPClient: no follow redirects ---

func TestCreateHTTPClientNoFollowRedirects(t *testing.T) {
	opts := ReplayOptions{
		Timeout:         5 * time.Second,
		FollowRedirects: false,
	}

	client := createHTTPClient(opts)
	if client.CheckRedirect == nil {
		t.Error("Expected non-nil CheckRedirect when FollowRedirects is false")
	}

	// Test that the redirect policy returns ErrUseLastResponse
	req := httptest.NewRequest("GET", "/redirect", nil)
	err := client.CheckRedirect(req, []*http.Request{})
	if err != http.ErrUseLastResponse {
		t.Errorf("Expected ErrUseLastResponse, got %v", err)
	}
}

// --- createHTTPClient: follow redirects with MaxRedirects ---

func TestCreateHTTPClientWithMaxRedirects(t *testing.T) {
	opts := ReplayOptions{
		Timeout:         5 * time.Second,
		FollowRedirects: true,
		MaxRedirects:    3,
	}

	client := createHTTPClient(opts)
	if client.CheckRedirect == nil {
		t.Error("Expected non-nil CheckRedirect when MaxRedirects > 0")
	}

	// Test within limit
	req := httptest.NewRequest("GET", "/redirect", nil)
	via := make([]*http.Request, 2) // 2 < 3
	err := client.CheckRedirect(req, via)
	if err != nil {
		t.Errorf("Expected nil error within redirect limit, got %v", err)
	}

	// Test at limit
	via = make([]*http.Request, 3) // 3 >= 3
	err = client.CheckRedirect(req, via)
	if err == nil {
		t.Error("Expected error when at redirect limit")
	}
	harErr := assertHarErrorCode(t, err, ErrCodeInvalidValue)
	if harErr.Metadata["maxRedirects"] != 3 {
		t.Errorf("Expected maxRedirects metadata 3, got %v", harErr.Metadata["maxRedirects"])
	}
}

// --- createHTTPClient: SkipSSLVerify ---

func TestCreateHTTPClientSkipSSLVerify(t *testing.T) {
	opts := ReplayOptions{
		Timeout:         5 * time.Second,
		SkipSSLVerify:   true,
		FollowRedirects: true,
		MaxRedirects:    0, // no max redirects => CheckRedirect is nil
	}

	client := createHTTPClient(opts)
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Expected *http.Transport")
	}
	if !transport.TLSClientConfig.InsecureSkipVerify {
		t.Error("Expected InsecureSkipVerify to be true")
	}
}

// --- createHTTPClient: follow redirects with MaxRedirects=0 => no redirect checker ---

func TestCreateHTTPClientFollowNoMaxRedirects(t *testing.T) {
	opts := ReplayOptions{
		Timeout:         5 * time.Second,
		FollowRedirects: true,
		MaxRedirects:    0,
	}

	client := createHTTPClient(opts)
	if client.CheckRedirect != nil {
		t.Error("Expected nil CheckRedirect when FollowRedirects=true and MaxRedirects=0")
	}
}

// --- HTTPResponseToEntries: nil response ---

func TestHTTPResponseToEntriesNilResponse(t *testing.T) {
	entry := HTTPResponseToEntries(&Entries{}, nil, 100*time.Millisecond)
	if entry != nil {
		t.Error("Expected nil for nil response")
	}
}

func TestHTTPResponseToEntriesNilRequest(t *testing.T) {
	resp := &http.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Proto:      "HTTP/1.1",
		Header:     http.Header{"Content-Type": []string{"text/plain"}},
		Body:       io.NopCloser(strings.NewReader("ok")),
	}

	entry := HTTPResponseToEntries(nil, resp, 10*time.Millisecond)
	if entry == nil {
		t.Fatal("Expected non-nil entry")
	}
	if entry.Request.Method != "" || entry.Request.URL != "" || len(entry.Request.Headers) != 0 {
		t.Errorf("Expected zero-value request, got %#v", entry.Request)
	}
	if entry.Response.Status != 200 {
		t.Errorf("Expected status 200, got %d", entry.Response.Status)
	}
}

// --- HTTPResponseToEntries: response with no body ---

func TestHTTPResponseToEntriesNoBody(t *testing.T) {
	resp := &http.Response{
		StatusCode: 204,
		Status:     "204 No Content",
		Proto:      "HTTP/1.1",
		Header:     http.Header{},
		Body:       nil,
	}

	origEntry := &Entries{
		Request: Request{
			Method: "DELETE",
			URL:    "https://example.com/resource",
		},
	}

	entry := HTTPResponseToEntries(origEntry, resp, 50*time.Millisecond)
	if entry == nil {
		t.Fatal("Expected non-nil entry")
	}
	if entry.Response.Status != 204 {
		t.Errorf("Expected status 204, got %d", entry.Response.Status)
	}
	if entry.Response.BodySize != -1 {
		t.Errorf("Expected BodySize -1 for no body, got %d", entry.Response.BodySize)
	}
}

// --- HTTPResponseToEntries: response with cookies ---

func TestHTTPResponseToEntriesWithCookies(t *testing.T) {
	resp := &http.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Proto:      "HTTP/1.1",
		Header: http.Header{
			"Set-Cookie": []string{"session=abc123; Path=/; HttpOnly", "token=xyz; Secure"},
		},
		Body: io.NopCloser(strings.NewReader("")),
	}

	origEntry := &Entries{
		Request: Request{
			Method: "GET",
			URL:    "https://example.com/test",
		},
	}

	entry := HTTPResponseToEntries(origEntry, resp, 10*time.Millisecond)
	if entry == nil {
		t.Fatal("Expected non-nil entry")
	}
	if len(entry.Response.Cookies) == 0 {
		t.Error("Expected cookies in response")
	}
}

// --- HTTPResponseToEntries: response with multiple header values ---

func TestHTTPResponseToEntriesMultipleHeaderValues(t *testing.T) {
	resp := &http.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Proto:      "HTTP/1.1",
		Header: http.Header{
			"Set-Cookie":   []string{"a=1", "b=2"},
			"Content-Type": []string{"text/html"},
		},
		Body: io.NopCloser(strings.NewReader("<html></html>")),
	}

	origEntry := &Entries{
		Request: Request{Method: "GET", URL: "https://example.com/"},
	}

	entry := HTTPResponseToEntries(origEntry, resp, 20*time.Millisecond)
	if entry == nil {
		t.Fatal("Expected non-nil entry")
	}

	// Check that multiple Set-Cookie headers result in multiple Headers entries
	cookieHeaders := 0
	for _, h := range entry.Response.Headers {
		if h.Name == "Set-Cookie" {
			cookieHeaders++
		}
	}
	if cookieHeaders != 2 {
		t.Errorf("Expected 2 Set-Cookie headers, got %d", cookieHeaders)
	}

	// Verify body content
	if entry.Response.Content.Text != "<html></html>" {
		t.Errorf("Expected body text, got %q", entry.Response.Content.Text)
	}
	expectedSize := len("<html></html>")
	if entry.Response.Content.Size != expectedSize {
		t.Errorf("Expected content size %d, got %d", expectedSize, entry.Response.Content.Size)
	}
	if entry.Response.BodySize != expectedSize {
		t.Errorf("Expected body size %d, got %d", expectedSize, entry.Response.BodySize)
	}
}

// --- HTTPResponseToEntries: preserves request from original entry ---

func TestHTTPResponseToEntriesPreservesRequest(t *testing.T) {
	resp := &http.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Proto:      "HTTP/1.1",
		Header:     http.Header{"Content-Type": []string{"text/plain"}},
		Body:       io.NopCloser(strings.NewReader("ok")),
	}

	origEntry := &Entries{
		Request: Request{
			Method: "POST",
			URL:    "https://example.com/api",
		},
	}

	entry := HTTPResponseToEntries(origEntry, resp, 50*time.Millisecond)
	if entry.Request.Method != "POST" {
		t.Errorf("Expected request method POST, got %s", entry.Request.Method)
	}
	if entry.Request.URL != "https://example.com/api" {
		t.Errorf("Expected request URL preserved, got %s", entry.Request.URL)
	}
}

// --- BuildQueryStringFromURL: invalid URL ---

func TestBuildQueryStringFromURLInvalidURL(t *testing.T) {
	params := BuildQueryStringFromURL("://invalid")
	if params != nil {
		t.Errorf("Expected nil for invalid URL, got %v", params)
	}
}

// --- BuildQueryStringFromURL: URL with no query params ---

func TestBuildQueryStringFromURLNoParams(t *testing.T) {
	params := BuildQueryStringFromURL("https://example.com/api")
	if len(params) != 0 {
		t.Errorf("Expected 0 params for URL without query string, got %d", len(params))
	}
}

// --- BuildQueryStringFromURL: URL with multi-value params ---

func TestBuildQueryStringFromURLMultiValue(t *testing.T) {
	params := BuildQueryStringFromURL("https://example.com/api?tag=a&tag=b&id=1")
	if len(params) != 3 {
		t.Errorf("Expected 3 params (multi-value), got %d", len(params))
	}

	// Verify the multi-value param
	tagCount := 0
	for _, p := range params {
		if p.Name == "tag" {
			tagCount++
		}
	}
	if tagCount != 2 {
		t.Errorf("Expected 2 'tag' params, got %d", tagCount)
	}
}

// --- ParseResponseHeaders: empty string ---

func TestParseResponseHeadersEmpty(t *testing.T) {
	headers := ParseResponseHeaders("")
	if len(headers) != 0 {
		t.Errorf("Expected 0 headers for empty string, got %d", len(headers))
	}
}

// --- ParseResponseHeaders: blank lines ---

func TestParseResponseHeadersBlankLines(t *testing.T) {
	headers := ParseResponseHeaders("\n\n\n")
	if len(headers) != 0 {
		t.Errorf("Expected 0 headers for blank lines, got %d", len(headers))
	}
}

// --- ParseResponseHeaders: line without colon ---

func TestParseResponseHeadersNoColon(t *testing.T) {
	headers := ParseResponseHeaders("JustAKeyNoColon")
	if len(headers) != 0 {
		t.Errorf("Expected 0 headers for line without colon, got %d", len(headers))
	}
}

// --- ParseResponseHeaders: trimming whitespace ---

func TestParseResponseHeadersTrimming(t *testing.T) {
	headers := ParseResponseHeaders("  Content-Type :   application/json  ")
	if len(headers) != 1 {
		t.Fatalf("Expected 1 header, got %d", len(headers))
	}
	if headers[0].Name != "Content-Type" {
		t.Errorf("Expected trimmed name 'Content-Type', got %q", headers[0].Name)
	}
	if headers[0].Value != "application/json" {
		t.Errorf("Expected trimmed value 'application/json', got %q", headers[0].Value)
	}
}

// --- ParseResponseHeaders: colon in value ---

func TestParseResponseHeadersColonInValue(t *testing.T) {
	headers := ParseResponseHeaders("Content-Type:text/html;charset=utf-8")
	if len(headers) != 1 {
		t.Fatalf("Expected 1 header, got %d", len(headers))
	}
	if headers[0].Value != "text/html;charset=utf-8" {
		t.Errorf("Expected value with colon preserved, got %q", headers[0].Value)
	}
}

// --- FormatBytes: all ranges including GB ---

func TestFormatBytesAllRanges(t *testing.T) {
	tests := []struct {
		size     int
		expected string
	}{
		{0, "0 B"},
		{1, "1 B"},
		{512, "512 B"},
		{1023, "1023 B"},
		{1024, "1.00 KB"},
		{1536, "1.50 KB"},
		{1048576, "1.00 MB"},
		{1572864, "1.50 MB"},
		{1073741824, "1.00 GB"},
		{1610612736, "1.50 GB"},
	}

	for _, tt := range tests {
		result := FormatBytes(tt.size)
		if result != tt.expected {
			t.Errorf("FormatBytes(%d) = %q, expected %q", tt.size, result, tt.expected)
		}
	}
}

// --- WriteRequestToWriter: nil entry ---

func TestWriteRequestToWriterNilEntry(t *testing.T) {
	var buf bytes.Buffer
	err := WriteRequestToWriter(nil, &buf)
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
}

func TestWriteRequestToWriterNilWriter(t *testing.T) {
	entry := &Entries{
		Request: Request{
			Method:      "GET",
			URL:         "https://example.com/api",
			HTTPVersion: "HTTP/1.1",
		},
	}

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
			err := WriteRequestToWriter(entry, tt.writer)
			assertHarErrorCode(t, err, ErrCodeInvalidFormat)
		})
	}
}

func TestWriteRequestToWriterWriteError(t *testing.T) {
	entry := &Entries{
		Request: Request{
			Method:      "GET",
			URL:         "https://example.com/api",
			HTTPVersion: "HTTP/1.1",
		},
	}

	errWriter := &failingWriter{err: errors.New("request write failure")}
	err := WriteRequestToWriter(entry, errWriter)
	if err == nil {
		t.Fatal("Expected write error")
	}
	assertHarErrorCode(t, err, ErrCodeFileSystem)
	if !errors.Is(err, errWriter.err) {
		t.Fatalf("expected wrapped request write failure, got %v", err)
	}
}

func TestWriteRequestToWriterShortWrite(t *testing.T) {
	entry := &Entries{
		Request: Request{
			Method:      "GET",
			URL:         "https://example.com/api",
			HTTPVersion: "HTTP/1.1",
		},
	}

	err := WriteRequestToWriter(entry, &shortWriter{})
	assertHarErrorCode(t, err, ErrCodeFileSystem)
	if !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("expected wrapped short write error, got %v", err)
	}
}

// --- WriteRequestToWriter: with PostData ---

func TestWriteRequestToWriterWithPostData(t *testing.T) {
	entry := &Entries{
		Request: Request{
			Method:      "POST",
			URL:         "https://example.com/api",
			HTTPVersion: "HTTP/1.1",
			Headers: []Headers{
				{Name: "Content-Type", Value: "application/json"},
			},
			PostData: &PostData{
				MimeType: "application/json",
				Text:     `{"key":"value"}`,
			},
		},
	}

	var buf bytes.Buffer
	err := WriteRequestToWriter(entry, &buf)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "POST") {
		t.Error("Expected POST method in output")
	}
	if !strings.Contains(output, `{"key":"value"}`) {
		t.Error("Expected PostData text in output")
	}
	if !strings.Contains(output, "Content-Type: application/json") {
		t.Error("Expected Content-Type header in output")
	}
}

// --- WriteRequestToWriter: no PostData ---

func TestWriteRequestToWriterNoPostData(t *testing.T) {
	entry := &Entries{
		Request: Request{
			Method:      "GET",
			URL:         "https://example.com/api",
			HTTPVersion: "HTTP/1.1",
			Headers:     []Headers{},
		},
	}

	var buf bytes.Buffer
	err := WriteRequestToWriter(entry, &buf)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "GET") {
		t.Error("Expected GET method in output")
	}
	// Should not have body content
	lines := strings.Split(output, "\n")
	if len(lines) < 2 {
		t.Error("Expected at least request line and blank line")
	}
}

// --- WriteRequestToWriter: PostData with empty text (should not write body) ---

func TestWriteRequestToWriterEmptyPostData(t *testing.T) {
	entry := &Entries{
		Request: Request{
			Method:      "POST",
			URL:         "https://example.com/api",
			HTTPVersion: "HTTP/1.1",
			Headers:     []Headers{},
			PostData: &PostData{
				MimeType: "text/plain",
				Text:     "",
			},
		},
	}

	var buf bytes.Buffer
	err := WriteRequestToWriter(entry, &buf)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "text/plain") {
		t.Error("PostData with empty text should not write body content")
	}
}

// --- CloneEntry: with PostData and Params ---

func TestCloneEntryWithPostDataParams(t *testing.T) {
	entry := &Entries{
		Request: Request{
			Method: "POST",
			URL:    "https://example.com/api",
			Headers: []Headers{
				{Name: "Content-Type", Value: "application/x-www-form-urlencoded"},
			},
			Cookies: []Cookie{
				{Name: "session", Value: "abc"},
			},
			QueryString: []QueryString{
				{Name: "id", Value: "123"},
			},
			PostData: &PostData{
				MimeType: "application/x-www-form-urlencoded",
				Params: []Param{
					{Name: "field1", Value: "value1"},
					{Name: "field2", Value: "value2"},
				},
				Text: "field1=value1&field2=value2",
			},
		},
		Response: Response{
			Status:     200,
			StatusText: "OK",
			Headers: []Headers{
				{Name: "Content-Type", Value: "application/json"},
			},
			Cookies: []Cookie{
				{Name: "tracking", Value: "xyz"},
			},
		},
	}

	cloned := CloneEntry(entry)

	// Verify deep copy of PostData Params
	if len(cloned.Request.PostData.Params) != 2 {
		t.Fatalf("Expected 2 PostData params, got %d", len(cloned.Request.PostData.Params))
	}
	cloned.Request.PostData.Params[0].Value = "modified"
	if entry.Request.PostData.Params[0].Value == "modified" {
		t.Error("Modifying cloned PostData.Params should not affect original")
	}

	// Verify deep copy of Request.Headers
	cloned.Request.Headers[0].Value = "modified"
	if entry.Request.Headers[0].Value == "modified" {
		t.Error("Modifying cloned Request.Headers should not affect original")
	}

	// Verify deep copy of Request.Cookies
	cloned.Request.Cookies[0].Value = "modified"
	if entry.Request.Cookies[0].Value == "modified" {
		t.Error("Modifying cloned Request.Cookies should not affect original")
	}

	// Verify deep copy of Request.QueryString
	cloned.Request.QueryString[0].Value = "modified"
	if entry.Request.QueryString[0].Value == "modified" {
		t.Error("Modifying cloned Request.QueryString should not affect original")
	}

	// Verify deep copy of Response.Headers
	cloned.Response.Headers[0].Value = "modified"
	if entry.Response.Headers[0].Value == "modified" {
		t.Error("Modifying cloned Response.Headers should not affect original")
	}

	// Verify deep copy of Response.Cookies
	cloned.Response.Cookies[0].Value = "modified"
	if entry.Response.Cookies[0].Value == "modified" {
		t.Error("Modifying cloned Response.Cookies should not affect original")
	}
}

// --- CloneEntry: PostData without Params ---

func TestCloneEntryPostDataNoParams(t *testing.T) {
	entry := &Entries{
		Request: Request{
			Method: "POST",
			URL:    "https://example.com/api",
			PostData: &PostData{
				MimeType: "application/json",
				Text:     `{"key":"value"}`,
			},
		},
	}

	cloned := CloneEntry(entry)
	if cloned.Request.PostData == nil {
		t.Fatal("Expected PostData to be cloned")
	}
	if cloned.Request.PostData.Text != `{"key":"value"}` {
		t.Errorf("Expected PostData.Text to be preserved, got %q", cloned.Request.PostData.Text)
	}

	// Modify clone's PostData and verify original is unchanged
	cloned.Request.PostData.Text = "modified"
	if entry.Request.PostData.Text == "modified" {
		t.Error("Modifying cloned PostData.Text should not affect original")
	}
}

// --- CloneEntry: with empty slices ---

func TestCloneEntryEmptySlices(t *testing.T) {
	entry := &Entries{
		Request: Request{
			Method:      "GET",
			URL:         "https://example.com",
			Headers:     []Headers{},
			Cookies:     []Cookie{},
			QueryString: []QueryString{},
		},
		Response: Response{
			Status:     200,
			StatusText: "OK",
			Headers:    []Headers{},
			Cookies:    []Cookie{},
		},
	}

	cloned := CloneEntry(entry)
	if cloned == nil {
		t.Fatal("Expected non-nil cloned entry")
	}
	if len(cloned.Request.Headers) != 0 {
		t.Errorf("Expected empty headers, got %d", len(cloned.Request.Headers))
	}
}

// --- Replay: complete round-trip with httptest including response body ---

func TestReplayCompleteRoundTrip(t *testing.T) {
	var receivedMethod, receivedURL, receivedCT string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedURL = r.URL.Path
		receivedCT = r.Header.Get("Content-Type")
		w.Header().Set("X-Custom", "response-value")
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(200)
		fmt.Fprint(w, "hello from server")
	}))
	defer server.Close()

	entry := &Entries{
		Request: Request{
			Method:      "PUT",
			URL:         server.URL + "/update",
			HTTPVersion: "HTTP/1.1",
			Headers: []Headers{
				{Name: "Content-Type", Value: "application/json"},
			},
			PostData: &PostData{
				MimeType: "application/json",
				Text:     `{"update":true}`,
			},
		},
	}

	opts := DefaultReplayOptions()
	result, err := entry.Replay(opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if receivedMethod != "PUT" {
		t.Errorf("Expected PUT, got %s", receivedMethod)
	}
	if receivedURL != "/update" {
		t.Errorf("Expected /update, got %s", receivedURL)
	}
	if receivedCT != "application/json" {
		t.Errorf("Expected application/json content type, got %s", receivedCT)
	}
	if result.Response.StatusCode != 200 {
		t.Errorf("Expected 200, got %d", result.Response.StatusCode)
	}
	if result.Response.Header.Get("X-Custom") != "response-value" {
		t.Errorf("Expected X-Custom header, got %q", result.Response.Header.Get("X-Custom"))
	}
}

// --- Replay + HTTPResponseToEntries + ReplayResultsToHar: full integration ---

func TestReplayFullIntegration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		fmt.Fprint(w, `{"result":"success"}`)
	}))
	defer server.Close()

	h := NewHar()
	h.AddEntry("GET", server.URL+"/api/test", "HTTP/1.1", "")
	h.AddEntry("GET", server.URL+"/api/other", "HTTP/1.1", "")

	opts := DefaultReplayOptions()
	opts.Timeout = 5 * time.Second

	// ReplayAll
	results, err := h.ReplayAll(opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	// Convert to Har
	replayHar := ReplayResultsToHar(results)
	if len(replayHar.Log.Entries) != 2 {
		t.Fatalf("Expected 2 entries in replay HAR, got %d", len(replayHar.Log.Entries))
	}

	// Verify the entries have proper response data
	for i, e := range replayHar.Log.Entries {
		if e.Response.Status != 200 {
			t.Errorf("Entry %d: Expected status 200, got %d", i, e.Response.Status)
		}
		if e.Response.Content.Text != `{"result":"success"}` {
			t.Errorf("Entry %d: Unexpected body: %q", i, e.Response.Content.Text)
		}
	}
}

// --- ReplayAll: empty Har ---

func TestReplayAllEmptyHar(t *testing.T) {
	h := NewHar()
	opts := DefaultReplayOptions()

	results, err := h.ReplayAll(opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results for empty Har, got %d", len(results))
	}
}

// --- ReplayResultsToHar: result with both Response and Entry (Response takes precedence) ---

func TestReplayResultsToHarWithBothResponseAndEntry(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		fmt.Fprint(w, `{"from":"server"}`)
	}))
	defer server.Close()

	entry := &Entries{
		Request: Request{
			Method:      "GET",
			URL:         server.URL + "/test",
			HTTPVersion: "HTTP/1.1",
		},
	}

	opts := DefaultReplayOptions()
	opts.Timeout = 5 * time.Second

	result, err := entry.Replay(opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	har := ReplayResultsToHar([]*ReplayResult{result})
	if len(har.Log.Entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(har.Log.Entries))
	}

	// When Response is present, HTTPResponseToEntries is used, which reads the body
	if har.Log.Entries[0].Response.Content.Text != `{"from":"server"}` {
		t.Errorf("Expected server response body, got %q", har.Log.Entries[0].Response.Content.Text)
	}
}

// --- Replay: test with SkipSSLVerify and httptest TLS server ---

func TestReplaySkipSSLVerify(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		fmt.Fprint(w, "secure ok")
	}))
	defer server.Close()

	entry := &Entries{
		Request: Request{
			Method:      "GET",
			URL:         server.URL + "/secure",
			HTTPVersion: "HTTP/1.1",
		},
	}

	opts := DefaultReplayOptions()
	opts.SkipSSLVerify = true
	opts.Timeout = 5 * time.Second

	result, err := entry.Replay(opts)
	if err != nil {
		t.Fatalf("Unexpected error with SkipSSLVerify: %v", err)
	}
	if result.Response.StatusCode != 200 {
		t.Errorf("Expected 200, got %d", result.Response.StatusCode)
	}
}

// --- Replay: test redirect following ---

func TestReplayFollowRedirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/final", http.StatusFound)
			return
		}
		w.WriteHeader(200)
		fmt.Fprint(w, "final destination")
	}))
	defer server.Close()

	entry := &Entries{
		Request: Request{
			Method:      "GET",
			URL:         server.URL + "/redirect",
			HTTPVersion: "HTTP/1.1",
		},
	}

	opts := DefaultReplayOptions()
	opts.FollowRedirects = true
	opts.Timeout = 5 * time.Second

	result, err := entry.Replay(opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	// After following redirect, should end up at the final destination
	if result.Response.StatusCode != 200 {
		t.Errorf("Expected 200 after redirect, got %d", result.Response.StatusCode)
	}
}

// --- Replay: max redirect error remains typed after net/http wrapping ---

func TestReplayMaxRedirectsReturnsHarError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/redirect", http.StatusFound)
	}))
	defer server.Close()

	entry := &Entries{
		Request: Request{
			Method:      "GET",
			URL:         server.URL + "/redirect",
			HTTPVersion: "HTTP/1.1",
		},
	}

	opts := DefaultReplayOptions()
	opts.FollowRedirects = true
	opts.MaxRedirects = 1
	opts.Timeout = 5 * time.Second

	result, err := entry.Replay(opts)
	harErr := assertHarErrorCode(t, err, ErrCodeInvalidValue)
	if harErr.Metadata["maxRedirects"] != 1 {
		t.Errorf("Expected maxRedirects metadata 1, got %v", harErr.Metadata["maxRedirects"])
	}
	if result == nil {
		t.Fatal("Expected replay result on redirect error")
	}
	resultErr := assertHarErrorCode(t, result.Error, ErrCodeInvalidValue)
	if resultErr != harErr {
		t.Error("Expected returned error and result error to be the same HAR error")
	}
}

// --- Replay: test no follow redirects ---

func TestReplayNoFollowRedirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/final", http.StatusFound)
			return
		}
		w.WriteHeader(200)
	}))
	defer server.Close()

	entry := &Entries{
		Request: Request{
			Method:      "GET",
			URL:         server.URL + "/redirect",
			HTTPVersion: "HTTP/1.1",
		},
	}

	opts := DefaultReplayOptions()
	opts.FollowRedirects = false
	opts.Timeout = 5 * time.Second

	result, err := entry.Replay(opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	// Without following redirects, should get 302
	if result.Response.StatusCode != http.StatusFound {
		t.Errorf("Expected 302 without following redirects, got %d", result.Response.StatusCode)
	}
}
