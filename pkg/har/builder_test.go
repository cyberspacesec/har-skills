package har

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

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
		w.Write([]byte(`{"status": "ok"}`))
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
	if err == nil {
		t.Error("Expected error for nil HAR")
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
		w.Write([]byte(`{"hello": "world"}`))
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
