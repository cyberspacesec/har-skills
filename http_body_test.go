package har

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestAddEntryFromHTTPSkipsRequestBodyOnReadError(t *testing.T) {
	reqBody := &failingReadCloser{
		data:    []byte("partial request"),
		readErr: errors.New("request read failure"),
	}
	req, err := http.NewRequest("POST", "https://example.com/api", reqBody)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "text/plain")

	h := NewHarBuilder().
		AddEntryFromHTTP(req, nil, time.Millisecond).
		Build()

	if len(h.Log.Entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(h.Log.Entries))
	}
	entry := h.Log.Entries[0]
	if entry.Request.PostData != nil {
		t.Fatalf("Expected PostData to remain nil after read error, got %#v", entry.Request.PostData)
	}
	if entry.Request.BodySize != -1 {
		t.Fatalf("Expected request BodySize to remain unknown, got %d", entry.Request.BodySize)
	}
}

func TestAddEntryFromHTTPRecordsResponseBodyReadError(t *testing.T) {
	req, err := http.NewRequest("GET", "https://example.com/api", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	readErr := errors.New("response read failure")
	resp := &http.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Proto:      "HTTP/1.1",
		Header:     http.Header{"Content-Type": []string{"text/plain"}},
		Body:       &failingReadCloser{data: []byte("partial response"), readErr: readErr},
	}

	h := NewHarBuilder().
		AddEntryFromHTTP(req, resp, time.Millisecond).
		Build()

	if len(h.Log.Entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(h.Log.Entries))
	}
	entry := h.Log.Entries[0]
	if entry.Response.Content.Text != "" || entry.Response.BodySize != -1 {
		t.Fatalf("Expected response body capture to remain unset, got content=%q bodySize=%d", entry.Response.Content.Text, entry.Response.BodySize)
	}
	if entry.Response.Error == nil {
		t.Fatal("Expected response read error to be recorded")
	}
	if !strings.Contains(entry.Response.Error.(string), readErr.Error()) {
		t.Fatalf("Expected recorded error to include read failure, got %v", entry.Response.Error)
	}
	if !resp.Body.(*failingReadCloser).closed {
		t.Fatal("Expected response body to be closed after read error")
	}
}

func TestAddEntryFromHTTPRecordsResponseBodyCloseErrorWithContent(t *testing.T) {
	req, err := http.NewRequest("GET", "https://example.com/api", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	closeErr := errors.New("response close failure")
	resp := &http.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Proto:      "HTTP/1.1",
		Header:     http.Header{"Content-Type": []string{"text/plain"}},
		Body:       &failingReadCloser{data: []byte("complete response"), closeErr: closeErr},
	}

	h := NewHarBuilder().
		AddEntryFromHTTP(req, resp, time.Millisecond).
		Build()

	entry := h.Log.Entries[0]
	if entry.Response.Content.Text != "complete response" {
		t.Fatalf("Expected response content despite close error, got %q", entry.Response.Content.Text)
	}
	if entry.Response.BodySize != len("complete response") {
		t.Fatalf("Expected response BodySize %d, got %d", len("complete response"), entry.Response.BodySize)
	}
	if entry.Response.Error == nil {
		t.Fatal("Expected response close error to be recorded")
	}
	if !strings.Contains(entry.Response.Error.(string), closeErr.Error()) {
		t.Fatalf("Expected recorded error to include close failure, got %v", entry.Response.Error)
	}
}

func TestHTTPResponseToEntriesRecordsBodyReadError(t *testing.T) {
	readErr := errors.New("replay read failure")
	resp := &http.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Proto:      "HTTP/1.1",
		Header:     http.Header{"Content-Type": []string{"text/plain"}},
		Body:       &failingReadCloser{data: []byte("partial replay"), readErr: readErr},
	}

	entry := HTTPResponseToEntries(nil, resp, time.Millisecond)
	if entry == nil {
		t.Fatal("Expected entry")
	}
	if entry.Response.Content.Text != "" || entry.Response.BodySize != -1 {
		t.Fatalf("Expected response body capture to remain unset, got content=%q bodySize=%d", entry.Response.Content.Text, entry.Response.BodySize)
	}
	if entry.Response.Error == nil {
		t.Fatal("Expected response read error to be recorded")
	}
	if !strings.Contains(entry.Response.Error.(string), readErr.Error()) {
		t.Fatalf("Expected recorded error to include read failure, got %v", entry.Response.Error)
	}
	if !resp.Body.(*failingReadCloser).closed {
		t.Fatal("Expected response body to be closed after read error")
	}
}

func TestHTTPResponseToEntriesRecordsBodyCloseErrorWithContent(t *testing.T) {
	closeErr := errors.New("replay close failure")
	resp := &http.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Proto:      "HTTP/1.1",
		Header:     http.Header{"Content-Type": []string{"text/plain"}},
		Body:       &failingReadCloser{data: []byte("complete replay"), closeErr: closeErr},
	}

	entry := HTTPResponseToEntries(nil, resp, time.Millisecond)
	if entry.Response.Content.Text != "complete replay" {
		t.Fatalf("Expected response content despite close error, got %q", entry.Response.Content.Text)
	}
	if entry.Response.BodySize != len("complete replay") {
		t.Fatalf("Expected response BodySize %d, got %d", len("complete replay"), entry.Response.BodySize)
	}
	if entry.Response.Error == nil {
		t.Fatal("Expected response close error to be recorded")
	}
	if !strings.Contains(entry.Response.Error.(string), closeErr.Error()) {
		t.Fatalf("Expected recorded error to include close failure, got %v", entry.Response.Error)
	}
}

func TestHTTPConversionTypedNilBodiesDoNotPanic(t *testing.T) {
	var typedNilBody *failingReadCloser
	req, err := http.NewRequest("POST", "https://example.com/api", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Body = typedNilBody
	resp := &http.Response{
		StatusCode: 204,
		Status:     "204 No Content",
		Proto:      "HTTP/1.1",
		Header:     http.Header{},
		Body:       typedNilBody,
	}

	assertDoesNotPanic(t, func() {
		h := NewHarBuilder().AddEntryFromHTTP(req, resp, time.Millisecond).Build()
		if len(h.Log.Entries) != 1 {
			t.Fatalf("Expected 1 entry, got %d", len(h.Log.Entries))
		}
	})
	assertDoesNotPanic(t, func() {
		if entry := HTTPResponseToEntries(nil, resp, time.Millisecond); entry == nil {
			t.Fatal("Expected entry")
		}
	})
}

func TestAddEntryFromHTTPNilRequestURLDoesNotPanic(t *testing.T) {
	req := &http.Request{
		Method: "GET",
		Proto:  "HTTP/1.1",
		Header: http.Header{},
	}

	assertDoesNotPanic(t, func() {
		h := NewHarBuilder().AddEntryFromHTTP(req, nil, time.Millisecond).Build()
		if len(h.Log.Entries) != 1 {
			t.Fatalf("Expected 1 entry, got %d", len(h.Log.Entries))
		}
		if h.Log.Entries[0].Request.URL != "" {
			t.Fatalf("Expected empty request URL, got %q", h.Log.Entries[0].Request.URL)
		}
	})
}

type failingReadCloser struct {
	data     []byte
	readErr  error
	closeErr error
	closed   bool
}

func (r *failingReadCloser) Read(p []byte) (int, error) {
	if len(r.data) > 0 {
		n := copy(p, r.data)
		r.data = r.data[n:]
		if len(r.data) == 0 && r.readErr != nil {
			return n, r.readErr
		}
		return n, nil
	}
	if r.readErr != nil {
		err := r.readErr
		r.readErr = nil
		return 0, err
	}
	return 0, io.EOF
}

func (r *failingReadCloser) Close() error {
	r.closed = true
	return r.closeErr
}
