package har

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// createTestOptimizedHarJSON builds a minimal valid HAR JSON byte slice for testing.
func createTestOptimizedHarJSON() []byte {
	har := Har{
		Log: Log{
			Version: "1.2",
			Creator: Creator{
				Name:    "TestCreator",
				Version: "1.0",
			},
			Entries: []Entries{
				{
					StartedDateTime: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					Time:            150.5,
					Request: Request{
						Method:      "GET",
						URL:         "https://example.com/api/users?id=42",
						HTTPVersion: "HTTP/1.1",
						Headers: []Headers{
							{Name: "Content-Type", Value: "application/json"},
							{Name: "Accept", Value: "*/*"},
						},
						QueryString: []QueryString{
							{Name: "id", Value: "42"},
						},
						HeadersSize: 200,
						BodySize:    0,
					},
					Response: Response{
						Status:      200,
						StatusText:  "OK",
						HTTPVersion: "HTTP/1.1",
						Headers: []Headers{
							{Name: "Content-Type", Value: "application/json"},
							{Name: "X-Custom", Value: "value1"},
						},
						Content: Content{
							Size:     256,
							MimeType: "application/json",
							Text:     `{"name":"test"}`,
							Comment:  "test content comment",
						},
						RedirectURL:  "",
						HeadersSize:  150,
						BodySize:     256,
						TransferSize: 300,
					},
					Cache: Cache{
						Comment: "cache comment",
					},
					Timings: Timings{
						Blocked: 10.0,
						DNS:     5.0,
						Connect: 15.0,
						Send:    3.0,
						Wait:    100.0,
						Receive: 17.5,
						Ssl:     8.0,
					},
					ServerIPAddress: "93.184.216.34",
					Connection:      "keep-alive",
					Pageref:         "page_0",
				},
				{
					StartedDateTime: time.Date(2024, 1, 15, 10, 30, 1, 0, time.UTC),
					Time:            80.0,
					Request: Request{
						Method:      "POST",
						URL:         "https://example.com/api/users",
						HTTPVersion: "HTTP/1.1",
						Headers: []Headers{
							{Name: "Content-Type", Value: "application/json"},
						},
						QueryString: []QueryString{},
						PostData: &PostData{
							MimeType: "application/json",
							Text:     `{"name":"new"}`,
						},
						HeadersSize: 180,
						BodySize:    50,
					},
					Response: Response{
						Status:      201,
						StatusText:  "Created",
						HTTPVersion: "HTTP/1.1",
						Headers: []Headers{
							{Name: "Content-Type", Value: "application/json"},
						},
						Content: Content{
							Size:     64,
							MimeType: "application/json",
							Text:     `{"id":99}`,
						},
						HeadersSize: 120,
						BodySize:    64,
					},
					Timings: Timings{
						Send: 2.0,
						Wait: 60.0,
						Receive: 18.0,
					},
				},
				{
					StartedDateTime: time.Date(2024, 1, 15, 10, 30, 2, 0, time.UTC),
					Time:            40.0,
					Request: Request{
						Method:      "GET",
						URL:         "https://other.com/page",
						HTTPVersion: "HTTP/2.0",
						Headers: []Headers{
							{Name: "Accept", Value: "text/html"},
						},
						HeadersSize: 100,
						BodySize:    0,
					},
					Response: Response{
						Status:      404,
						StatusText:  "Not Found",
						HTTPVersion: "HTTP/2.0",
						Content: Content{
							Size:     0,
							MimeType: "text/html",
						},
						HeadersSize: 80,
						BodySize:    0,
					},
					Timings: Timings{
						Send:    1.0,
						Wait:    30.0,
						Receive: 9.0,
					},
				},
			},
		},
	}

	data, err := json.Marshal(har)
	if err != nil {
		panic("failed to marshal test HAR: " + err.Error())
	}
	return data
}

// writeTestHarFile writes a test HAR file to a temporary directory and returns the path.
func writeTestHarFile(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.har")
	data := createTestOptimizedHarJSON()
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write test HAR file: %v", err)
	}
	return path
}

// ---------------------------------------------------------------------------
// ParseHarFileOptimized / ParseHarOptimized
// ---------------------------------------------------------------------------

func TestParseHarOptimized(t *testing.T) {
	data := createTestOptimizedHarJSON()

	optHar, err := ParseHarOptimized(data)
	if err != nil {
		t.Fatalf("ParseHarOptimized returned error: %v", err)
	}
	if optHar == nil {
		t.Fatal("ParseHarOptimized returned nil")
	}

	// Verify basic fields
	if optHar.Log.Version != "1.2" {
		t.Errorf("Version = %q, want %q", optHar.Log.Version, "1.2")
	}
	if optHar.Log.Creator.Name != "TestCreator" {
		t.Errorf("Creator.Name = %q, want %q", optHar.Log.Creator.Name, "TestCreator")
	}
	if len(optHar.Log.Entries) != 3 {
		t.Fatalf("len(Entries) = %d, want 3", len(optHar.Log.Entries))
	}

	// Spot-check first entry
	e0 := optHar.Log.Entries[0]
	if e0.Request.Method != MethodGET {
		t.Errorf("Entries[0].Request.Method = %v, want MethodGET", e0.Request.Method)
	}
	if e0.Request.URL != "https://example.com/api/users?id=42" {
		t.Errorf("Entries[0].Request.URL = %q, want %q", e0.Request.URL, "https://example.com/api/users?id=42")
	}
	if e0.Response.Status != 200 {
		t.Errorf("Entries[0].Response.Status = %d, want 200", e0.Response.Status)
	}
}

func TestParseHarOptimizedEmptyInput(t *testing.T) {
	_, err := ParseHarOptimized([]byte{})
	if err == nil {
		t.Error("expected error for empty input, got nil")
	}
}

func TestParseHarOptimizedInvalidJSON(t *testing.T) {
	_, err := ParseHarOptimized([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestParseHarFileOptimized(t *testing.T) {
	path := writeTestHarFile(t)

	optHar, err := ParseHarFileOptimized(path)
	if err != nil {
		t.Fatalf("ParseHarFileOptimized returned error: %v", err)
	}
	if optHar == nil {
		t.Fatal("ParseHarFileOptimized returned nil")
	}
	if optHar.Log.Version != "1.2" {
		t.Errorf("Version = %q, want %q", optHar.Log.Version, "1.2")
	}
	if len(optHar.Log.Entries) != 3 {
		t.Errorf("len(Entries) = %d, want 3", len(optHar.Log.Entries))
	}
}

func TestParseHarFileOptimizedNoSuchFile(t *testing.T) {
	_, err := ParseHarFileOptimized("/nonexistent/path/file.har")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

// ---------------------------------------------------------------------------
// ToOptimizedHar
// ---------------------------------------------------------------------------

func TestToOptimizedHar(t *testing.T) {
	standard := createFullTestHar()

	optimized := ToOptimizedHar(standard)

	// Version and Creator
	if optimized.Log.Version != standard.Log.Version {
		t.Errorf("Version = %q, want %q", optimized.Log.Version, standard.Log.Version)
	}
	if optimized.Log.Creator.Name != standard.Log.Creator.Name {
		t.Errorf("Creator.Name = %q, want %q", optimized.Log.Creator.Name, standard.Log.Creator.Name)
	}
	if optimized.Log.Creator.Version != standard.Log.Creator.Version {
		t.Errorf("Creator.Version = %q, want %q", optimized.Log.Creator.Version, standard.Log.Creator.Version)
	}

	// Pages
	if len(optimized.Log.Pages) != len(standard.Log.Pages) {
		t.Errorf("len(Pages) = %d, want %d", len(optimized.Log.Pages), len(standard.Log.Pages))
	}

	// Entries
	if len(optimized.Log.Entries) != len(standard.Log.Entries) {
		t.Fatalf("len(Entries) = %d, want %d", len(optimized.Log.Entries), len(standard.Log.Entries))
	}

	// Check first entry fields
	oe0 := optimized.Log.Entries[0]
	se0 := standard.Log.Entries[0]

	// Request method should be enum
	if oe0.Request.Method != MethodGET {
		t.Errorf("Method = %v, want MethodGET", oe0.Request.Method)
	}
	if oe0.Request.URL != se0.Request.URL {
		t.Errorf("URL = %q, want %q", oe0.Request.URL, se0.Request.URL)
	}
	if oe0.Request.HTTPVersion != se0.Request.HTTPVersion {
		t.Errorf("HTTPVersion = %q, want %q", oe0.Request.HTTPVersion, se0.Request.HTTPVersion)
	}

	// Headers should be in map
	if v, ok := oe0.Request.Headers["Accept"]; !ok || v != "application/json" {
		t.Errorf("Request.Headers[Accept] = %q, ok=%v, want %q, ok=true", v, ok, "application/json")
	}
	if v, ok := oe0.Request.Headers["User-Agent"]; !ok || v != "Go-HAR Test" {
		t.Errorf("Request.Headers[User-Agent] = %q, ok=%v, want %q, ok=true", v, ok, "Go-HAR Test")
	}

	// QueryString should be in map
	if v, ok := oe0.Request.QueryString["id"]; !ok || v != "12345" {
		t.Errorf("Request.QueryString[id] = %q, ok=%v, want %q, ok=true", v, ok, "12345")
	}
	if v, ok := oe0.Request.QueryString["format"]; !ok || v != "json" {
		t.Errorf("Request.QueryString[format] = %q, ok=%v, want %q, ok=true", v, ok, "json")
	}

	// Response
	if oe0.Response.Status != se0.Response.Status {
		t.Errorf("Response.Status = %d, want %d", oe0.Response.Status, se0.Response.Status)
	}
	if oe0.Response.StatusText != se0.Response.StatusText {
		t.Errorf("Response.StatusText = %q, want %q", oe0.Response.StatusText, se0.Response.StatusText)
	}

	// Response headers
	if v, ok := oe0.Response.Headers["Content-Type"]; !ok || v != "application/json" {
		t.Errorf("Response.Headers[Content-Type] = %q, ok=%v, want %q, ok=true", v, ok, "application/json")
	}

	// ServerIPAddress and Connection
	if oe0.ServerIP == nil || *oe0.ServerIP != "192.168.1.1" {
		t.Errorf("ServerIP = %v, want pointer to %q", oe0.ServerIP, "192.168.1.1")
	}
	if oe0.Connection == nil || *oe0.Connection != "close" {
		t.Errorf("Connection = %v, want pointer to %q", oe0.Connection, "close")
	}

	// Pageref
	if oe0.PageRef == nil || *oe0.PageRef != "page_1" {
		t.Errorf("PageRef = %v, want pointer to %q", oe0.PageRef, "page_1")
	}

	// Sizes
	if oe0.Request.HeadersSize == nil || *oe0.Request.HeadersSize != 150 {
		t.Errorf("Request.HeadersSize = %v, want pointer to 150", oe0.Request.HeadersSize)
	}
	if oe0.Response.HeadersSize == nil || *oe0.Response.HeadersSize != 120 {
		t.Errorf("Response.HeadersSize = %v, want pointer to 120", oe0.Response.HeadersSize)
	}
	if oe0.Response.BodySize == nil || *oe0.Response.BodySize != 1024 {
		t.Errorf("Response.BodySize = %v, want pointer to 1024", oe0.Response.BodySize)
	}
	if oe0.Response.TransferSize == nil || *oe0.Response.TransferSize != 1144 {
		t.Errorf("Response.TransferSize = %v, want pointer to 1144", oe0.Response.TransferSize)
	}

	// Content
	if oe0.Response.Content == nil {
		t.Fatal("Response.Content is nil, want non-nil")
	}
	if oe0.Response.Content.Size != 1024 {
		t.Errorf("Content.Size = %d, want 1024", oe0.Response.Content.Size)
	}
	if oe0.Response.Content.MimeType != "application/json" {
		t.Errorf("Content.MimeType = %q, want %q", oe0.Response.Content.MimeType, "application/json")
	}

	// Timings
	if oe0.Timings.Blocked == nil || *oe0.Timings.Blocked != 12.5 {
		t.Errorf("Timings.Blocked = %v, want pointer to 12.5", oe0.Timings.Blocked)
	}
	if oe0.Timings.DNS == nil || *oe0.Timings.DNS != 10.0 {
		t.Errorf("Timings.DNS = %v, want pointer to 10.0", oe0.Timings.DNS)
	}
	if oe0.Timings.Send == nil || *oe0.Timings.Send != 5.5 {
		t.Errorf("Timings.Send = %v, want pointer to 5.5", oe0.Timings.Send)
	}
	if oe0.Timings.Wait == nil || *oe0.Timings.Wait != 75.25 {
		t.Errorf("Timings.Wait = %v, want pointer to 75.25", oe0.Timings.Wait)
	}

	// Cache: an empty Cache{} has all zero fields, so convertToOptimizedEntry
	// will set it to nil. This is expected behavior - only non-trivial caches
	// are preserved. Verify the expectation.
	if oe0.Cache != nil {
		t.Errorf("Cache = %+v, want nil (empty Cache{} has all zero fields, not preserved)", oe0.Cache)
	}
}

// createFullTestHar creates a standard Har struct with full data for testing.
func createFullTestHar() *Har {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	return &Har{
		Log: Log{
			Version: "1.2",
			Creator: Creator{
				Name:    "Go-HAR Test",
				Version: "1.0",
			},
			Pages: []Pages{
				{
					StartedDateTime: now,
					ID:              "page_1",
					Title:           "Test Page",
					PageTimings: PageTimings{
						OnContentLoad: 150.5,
						OnLoad:        250.75,
					},
				},
			},
			Entries: []Entries{
				{
					Pageref:         "page_1",
					StartedDateTime: now,
					Time:            350.25,
					Request: Request{
						Method:      "GET",
						URL:         "https://example.com/test",
						HTTPVersion: "HTTP/1.1",
						Headers: []Headers{
							{Name: "Accept", Value: "application/json"},
							{Name: "User-Agent", Value: "Go-HAR Test"},
						},
						QueryString: []QueryString{
							{Name: "id", Value: "12345"},
							{Name: "format", Value: "json"},
						},
						Cookies: []Cookie{
							{
								Name:     "session",
								Value:    "abc123",
								Path:     "/",
								Domain:   "example.com",
								HTTPOnly: true,
								Secure:   true,
							},
						},
						HeadersSize: 150,
						BodySize:    0,
					},
					Response: Response{
						Status:      200,
						StatusText:  "OK",
						HTTPVersion: "HTTP/1.1",
						Headers: []Headers{
							{Name: "Content-Type", Value: "application/json"},
							{Name: "Cache-Control", Value: "no-cache"},
						},
						Content: Content{
							Size:     1024,
							MimeType: "application/json",
						},
						RedirectURL:  "",
						HeadersSize:  120,
						BodySize:     1024,
						TransferSize: 1144,
					},
					Cache: Cache{},
					Timings: Timings{
						Blocked: 12.5,
						DNS:     10.0,
						Connect: 25.5,
						Send:    5.5,
						Wait:    75.25,
						Receive: 15.75,
						Ssl:     20.0,
					},
					ServerIPAddress: "192.168.1.1",
					Connection:      "close",
				},
			},
		},
	}
}

// ---------------------------------------------------------------------------
// OptimizedHar.ToStandardHar - round-trip conversion
// ---------------------------------------------------------------------------

func TestOptimizedHarRoundTrip(t *testing.T) {
	original := createFullTestHar()

	// Standard -> Optimized -> Standard
	optimized := ToOptimizedHar(original)
	roundTripped := optimized.ToStandardHar()

	// Version
	if roundTripped.Log.Version != original.Log.Version {
		t.Errorf("Version: got %q, want %q", roundTripped.Log.Version, original.Log.Version)
	}

	// Creator
	if roundTripped.Log.Creator.Name != original.Log.Creator.Name {
		t.Errorf("Creator.Name: got %q, want %q", roundTripped.Log.Creator.Name, original.Log.Creator.Name)
	}

	// Pages
	if len(roundTripped.Log.Pages) != len(original.Log.Pages) {
		t.Fatalf("Pages length: got %d, want %d", len(roundTripped.Log.Pages), len(original.Log.Pages))
	}
	if roundTripped.Log.Pages[0].ID != original.Log.Pages[0].ID {
		t.Errorf("Pages[0].ID: got %q, want %q", roundTripped.Log.Pages[0].ID, original.Log.Pages[0].ID)
	}

	// Entries count
	if len(roundTripped.Log.Entries) != len(original.Log.Entries) {
		t.Fatalf("Entries length: got %d, want %d", len(roundTripped.Log.Entries), len(original.Log.Entries))
	}

	// Entry fields
	oe := roundTripped.Log.Entries[0]
	se := original.Log.Entries[0]

	if oe.Request.Method != se.Request.Method {
		t.Errorf("Request.Method: got %q, want %q", oe.Request.Method, se.Request.Method)
	}
	if oe.Request.URL != se.Request.URL {
		t.Errorf("Request.URL: got %q, want %q", oe.Request.URL, se.Request.URL)
	}
	if oe.Response.Status != se.Response.Status {
		t.Errorf("Response.Status: got %d, want %d", oe.Response.Status, se.Response.Status)
	}
	if oe.Response.StatusText != se.Response.StatusText {
		t.Errorf("Response.StatusText: got %q, want %q", oe.Response.StatusText, se.Response.StatusText)
	}

	// Timings
	if oe.Timings.Send != se.Timings.Send {
		t.Errorf("Timings.Send: got %v, want %v", oe.Timings.Send, se.Timings.Send)
	}
	if oe.Timings.Wait != se.Timings.Wait {
		t.Errorf("Timings.Wait: got %v, want %v", oe.Timings.Wait, se.Timings.Wait)
	}

	// ServerIPAddress
	if oe.ServerIPAddress != se.ServerIPAddress {
		t.Errorf("ServerIPAddress: got %q, want %q", oe.ServerIPAddress, se.ServerIPAddress)
	}

	// Connection
	if oe.Connection != se.Connection {
		t.Errorf("Connection: got %q, want %q", oe.Connection, se.Connection)
	}

	// Pageref
	if oe.Pageref != se.Pageref {
		t.Errorf("Pageref: got %q, want %q", oe.Pageref, se.Pageref)
	}

	// Request sizes
	if oe.Request.HeadersSize != se.Request.HeadersSize {
		t.Errorf("Request.HeadersSize: got %d, want %d", oe.Request.HeadersSize, se.Request.HeadersSize)
	}
	if oe.Request.BodySize != se.Request.BodySize {
		t.Errorf("Request.BodySize: got %d, want %d", oe.Request.BodySize, se.Request.BodySize)
	}

	// Response sizes
	if oe.Response.HeadersSize != se.Response.HeadersSize {
		t.Errorf("Response.HeadersSize: got %d, want %d", oe.Response.HeadersSize, se.Response.HeadersSize)
	}
	if oe.Response.BodySize != se.Response.BodySize {
		t.Errorf("Response.BodySize: got %d, want %d", oe.Response.BodySize, se.Response.BodySize)
	}
}

// ---------------------------------------------------------------------------
// OptimizedEntries.ToStandard - ServerIPAddress, Connection preserved
// ---------------------------------------------------------------------------

func TestOptimizedEntriesToStandard_ServerIPAndConnection(t *testing.T) {
	serverIP := "10.0.0.1"
	conn := "12345"
	pageRef := "page_abc"

	entry := OptimizedEntries{
		StartedDateTime: time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC),
		Time:            100.0,
		Request: OptimizedRequest{
			Method: MethodGET,
			URL:    "https://example.com/",
		},
		Response: OptimizedResponse{
			Status:     200,
			StatusText: "OK",
		},
		ServerIP:   &serverIP,
		Connection: &conn,
		PageRef:    &pageRef,
	}

	standard := entry.ToStandard()

	if standard.ServerIPAddress != serverIP {
		t.Errorf("ServerIPAddress = %q, want %q", standard.ServerIPAddress, serverIP)
	}
	if standard.Connection != conn {
		t.Errorf("Connection = %q, want %q", standard.Connection, conn)
	}
	if standard.Pageref != pageRef {
		t.Errorf("Pageref = %q, want %q", standard.Pageref, pageRef)
	}
}

func TestOptimizedEntriesToStandard_NilOptionalFields(t *testing.T) {
	entry := OptimizedEntries{
		StartedDateTime: time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC),
		Time:            50.0,
		Request: OptimizedRequest{
			Method: MethodGET,
			URL:    "https://example.com/",
		},
		Response: OptimizedResponse{
			Status:     200,
			StatusText: "OK",
		},
		// ServerIP, Connection, PageRef, Cache all nil
	}

	standard := entry.ToStandard()

	if standard.ServerIPAddress != "" {
		t.Errorf("ServerIPAddress = %q, want empty", standard.ServerIPAddress)
	}
	if standard.Connection != "" {
		t.Errorf("Connection = %q, want empty", standard.Connection)
	}
	if standard.Pageref != "" {
		t.Errorf("Pageref = %q, want empty", standard.Pageref)
	}
}

func TestOptimizedEntriesToStandard_Cache(t *testing.T) {
	cache := Cache{Comment: "cached"}
	entry := OptimizedEntries{
		StartedDateTime: time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC),
		Time:            50.0,
		Request: OptimizedRequest{
			Method: MethodGET,
			URL:    "https://example.com/",
		},
		Response: OptimizedResponse{
			Status:     200,
			StatusText: "OK",
		},
		Cache: &cache,
	}

	standard := entry.ToStandard()

	if standard.Cache.Comment != "cached" {
		t.Errorf("Cache.Comment = %q, want %q", standard.Cache.Comment, "cached")
	}
}

// ---------------------------------------------------------------------------
// OptimizedContent.ToStandard - Text, Encoding, Comment preserved
// ---------------------------------------------------------------------------

func TestOptimizedContentToStandard_AllFields(t *testing.T) {
	text := "response body text"
	encoding := "base64"
	comment := "content comment"

	content := &OptimizedContent{
		Size:     500,
		MimeType: "application/json",
		Text:     &text,
		Encoding: &encoding,
		Comment:  &comment,
	}

	standard := content.ToStandard()

	if standard.Size != 500 {
		t.Errorf("Size = %d, want 500", standard.Size)
	}
	if standard.MimeType != "application/json" {
		t.Errorf("MimeType = %q, want %q", standard.MimeType, "application/json")
	}
	if standard.Text != text {
		t.Errorf("Text = %q, want %q", standard.Text, text)
	}
	if standard.Encoding != encoding {
		t.Errorf("Encoding = %q, want %q", standard.Encoding, encoding)
	}
	if standard.Comment != comment {
		t.Errorf("Comment = %q, want %q", standard.Comment, comment)
	}
}

func TestOptimizedContentToStandard_NilFields(t *testing.T) {
	content := &OptimizedContent{
		Size:     100,
		MimeType: "text/html",
		// Text, Encoding, Comment are nil
	}

	standard := content.ToStandard()

	if standard.Text != "" {
		t.Errorf("Text = %q, want empty string for nil Text", standard.Text)
	}
	if standard.Encoding != "" {
		t.Errorf("Encoding = %q, want empty string for nil Encoding", standard.Encoding)
	}
	if standard.Comment != "" {
		t.Errorf("Comment = %q, want empty string for nil Comment", standard.Comment)
	}
}

func TestOptimizedContentToStandard_EmptyStrings(t *testing.T) {
	empty := ""
	content := &OptimizedContent{
		Size:     0,
		MimeType: "",
		Text:     &empty,
		Encoding: &empty,
		Comment:  &empty,
	}

	standard := content.ToStandard()

	// Pointers to empty strings should still result in empty strings, not loss
	if standard.Text != "" {
		t.Errorf("Text = %q, want empty", standard.Text)
	}
	if standard.Encoding != "" {
		t.Errorf("Encoding = %q, want empty", standard.Encoding)
	}
	if standard.Comment != "" {
		t.Errorf("Comment = %q, want empty", standard.Comment)
	}
}

// Content field preservation through full round-trip (Standard -> Optimized -> Standard)
func TestOptimizedContentRoundTrip(t *testing.T) {
	text := "some body content"
	encoding := "base64"
	comment := "a comment"

	original := Content{
		Size:     42,
		MimeType: "text/plain",
		Text:     text,
		Encoding: encoding,
		Comment:  comment,
	}

	// Build a standard Har containing this content
	standardHar := &Har{
		Log: Log{
			Version: "1.2",
			Creator: Creator{Name: "test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					Time:            10,
					Request: Request{
						Method: "GET",
						URL:    "https://example.com/",
					},
					Response: Response{
						Status:     200,
						StatusText: "OK",
						Content:    original,
					},
					Timings: Timings{Send: 1, Wait: 5, Receive: 4},
				},
			},
		},
	}

	optimized := ToOptimizedHar(standardHar)
	roundTripped := optimized.ToStandardHar()
	rtContent := roundTripped.Log.Entries[0].Response.Content

	if rtContent.Text != text {
		t.Errorf("Content.Text: got %q, want %q", rtContent.Text, text)
	}
	if rtContent.Encoding != encoding {
		t.Errorf("Content.Encoding: got %q, want %q", rtContent.Encoding, encoding)
	}
	if rtContent.Comment != comment {
		t.Errorf("Content.Comment: got %q, want %q", rtContent.Comment, comment)
	}
	if rtContent.Size != original.Size {
		t.Errorf("Content.Size: got %d, want %d", rtContent.Size, original.Size)
	}
	if rtContent.MimeType != original.MimeType {
		t.Errorf("Content.MimeType: got %q, want %q", rtContent.MimeType, original.MimeType)
	}
}

// ---------------------------------------------------------------------------
// OptimizedRequest.ToStandard - QueryString and PostData preserved
// ---------------------------------------------------------------------------

func TestOptimizedRequestToStandard_QueryStringAndPostData(t *testing.T) {
	postData := &PostData{
		MimeType: "application/json",
		Text:     `{"key":"value"}`,
	}

	req := OptimizedRequest{
		Method:      MethodPOST,
		URL:         "https://example.com/submit?flag=yes",
		HTTPVersion: "HTTP/1.1",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		QueryString: map[string]string{
			"flag": "yes",
			"page": "1",
		},
		PostData: postData,
	}

	standard := req.ToStandard()

	// Method
	if standard.Method != "POST" {
		t.Errorf("Method = %q, want %q", standard.Method, "POST")
	}

	// URL
	if standard.URL != "https://example.com/submit?flag=yes" {
		t.Errorf("URL = %q, want %q", standard.URL, "https://example.com/submit?flag=yes")
	}

	// PostData should be preserved
	if standard.PostData == nil {
		t.Fatal("PostData is nil, want non-nil")
	}
	if standard.PostData.MimeType != "application/json" {
		t.Errorf("PostData.MimeType = %q, want %q", standard.PostData.MimeType, "application/json")
	}
	if standard.PostData.Text != `{"key":"value"}` {
		t.Errorf("PostData.Text = %q, want %q", standard.PostData.Text, `{"key":"value"}`)
	}

	// QueryString should have 2 entries
	if len(standard.QueryString) != 2 {
		t.Fatalf("len(QueryString) = %d, want 2", len(standard.QueryString))
	}

	// Verify query string values are present (map iteration order is non-deterministic)
	qsMap := map[string]string{}
	for _, qs := range standard.QueryString {
		qsMap[qs.Name] = qs.Value
	}
	if v, ok := qsMap["flag"]; !ok || v != "yes" {
		t.Errorf("QueryString flag = %q, ok=%v, want %q, ok=true", v, ok, "yes")
	}
	if v, ok := qsMap["page"]; !ok || v != "1" {
		t.Errorf("QueryString page = %q, ok=%v, want %q, ok=true", v, ok, "1")
	}
}

func TestOptimizedRequestToStandard_NilPostData(t *testing.T) {
	req := OptimizedRequest{
		Method: MethodGET,
		URL:    "https://example.com/",
	}

	standard := req.ToStandard()

	if standard.PostData != nil {
		t.Errorf("PostData = %+v, want nil", standard.PostData)
	}
}

func TestOptimizedRequestToStandard_EmptyQueryString(t *testing.T) {
	req := OptimizedRequest{
		Method:      MethodGET,
		URL:         "https://example.com/",
		QueryString: map[string]string{},
	}

	standard := req.ToStandard()

	if len(standard.QueryString) != 0 {
		t.Errorf("len(QueryString) = %d, want 0", len(standard.QueryString))
	}
}

func TestOptimizedRequestToStandard_Sizes(t *testing.T) {
	headerSize := 250
	bodySize := 80

	req := OptimizedRequest{
		Method:      MethodPOST,
		URL:         "https://example.com/",
		HeadersSize: &headerSize,
		BodySize:    &bodySize,
	}

	standard := req.ToStandard()

	if standard.HeadersSize != headerSize {
		t.Errorf("HeadersSize = %d, want %d", standard.HeadersSize, headerSize)
	}
	if standard.BodySize != bodySize {
		t.Errorf("BodySize = %d, want %d", standard.BodySize, bodySize)
	}
}

// Full round-trip for request query string and post data
func TestOptimizedRequestRoundTrip(t *testing.T) {
	standardHar := &Har{
		Log: Log{
			Version: "1.2",
			Creator: Creator{Name: "test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					Time:            10,
					Request: Request{
						Method:      "POST",
						URL:         "https://example.com/api?token=abc",
						HTTPVersion: "HTTP/1.1",
						QueryString: []QueryString{
							{Name: "token", Value: "abc"},
						},
						PostData: &PostData{
							MimeType: "application/json",
							Text:     `{"hello":"world"}`,
						},
						HeadersSize: 200,
						BodySize:    30,
					},
					Response: Response{
						Status:     200,
						StatusText: "OK",
						Content: Content{
							Size:     10,
							MimeType: "text/plain",
						},
					},
					Timings: Timings{Send: 1, Wait: 5, Receive: 4},
				},
			},
		},
	}

	optimized := ToOptimizedHar(standardHar)
	roundTripped := optimized.ToStandardHar()

	rtReq := roundTripped.Log.Entries[0].Request

	// QueryString
	if len(rtReq.QueryString) != 1 {
		t.Fatalf("len(QueryString) = %d, want 1", len(rtReq.QueryString))
	}
	if rtReq.QueryString[0].Name != "token" || rtReq.QueryString[0].Value != "abc" {
		t.Errorf("QueryString[0] = {%q, %q}, want {token, abc}", rtReq.QueryString[0].Name, rtReq.QueryString[0].Value)
	}

	// PostData
	if rtReq.PostData == nil {
		t.Fatal("PostData is nil after round trip")
	}
	if rtReq.PostData.MimeType != "application/json" {
		t.Errorf("PostData.MimeType = %q, want %q", rtReq.PostData.MimeType, "application/json")
	}
	if rtReq.PostData.Text != `{"hello":"world"}` {
		t.Errorf("PostData.Text = %q, want %q", rtReq.PostData.Text, `{"hello":"world"}`)
	}

	// Sizes
	if rtReq.HeadersSize != 200 {
		t.Errorf("HeadersSize = %d, want 200", rtReq.HeadersSize)
	}
	if rtReq.BodySize != 30 {
		t.Errorf("BodySize = %d, want 30", rtReq.BodySize)
	}
}

// ---------------------------------------------------------------------------
// SearchByURL, SearchByMethod, SearchByStatusCode
// ---------------------------------------------------------------------------

func TestSearchByURL(t *testing.T) {
	data := createTestOptimizedHarJSON()
	optHar, err := ParseHarOptimized(data)
	if err != nil {
		t.Fatalf("ParseHarOptimized error: %v", err)
	}

	// Search for "example.com"
	results := optHar.SearchByURL("example.com")
	if len(results) != 2 {
		t.Errorf("SearchByURL(example.com) returned %d results, want 2", len(results))
	}

	// Search for "other.com"
	results = optHar.SearchByURL("other.com")
	if len(results) != 1 {
		t.Errorf("SearchByURL(other.com) returned %d results, want 1", len(results))
	}
	if len(results) > 0 && results[0].Request.URL != "https://other.com/page" {
		t.Errorf("SearchByURL(other.com) URL = %q, want %q", results[0].Request.URL, "https://other.com/page")
	}

	// Search for non-existent pattern
	results = optHar.SearchByURL("nonexistent.example")
	if len(results) != 0 {
		t.Errorf("SearchByURL(nonexistent.example) returned %d results, want 0", len(results))
	}
}

func TestSearchByMethod(t *testing.T) {
	data := createTestOptimizedHarJSON()
	optHar, err := ParseHarOptimized(data)
	if err != nil {
		t.Fatalf("ParseHarOptimized error: %v", err)
	}

	// GET entries
	getResults := optHar.SearchByMethod(MethodGET)
	if len(getResults) != 2 {
		t.Errorf("SearchByMethod(GET) returned %d results, want 2", len(getResults))
	}

	// POST entries
	postResults := optHar.SearchByMethod(MethodPOST)
	if len(postResults) != 1 {
		t.Errorf("SearchByMethod(POST) returned %d results, want 1", len(postResults))
	}
	if len(postResults) > 0 && postResults[0].Request.URL != "https://example.com/api/users" {
		t.Errorf("SearchByMethod(POST) URL = %q, want %q", postResults[0].Request.URL, "https://example.com/api/users")
	}

	// Unknown method
	unknownResults := optHar.SearchByMethod(MethodDELETE)
	if len(unknownResults) != 0 {
		t.Errorf("SearchByMethod(DELETE) returned %d results, want 0", len(unknownResults))
	}
}

func TestSearchByStatusCode(t *testing.T) {
	data := createTestOptimizedHarJSON()
	optHar, err := ParseHarOptimized(data)
	if err != nil {
		t.Fatalf("ParseHarOptimized error: %v", err)
	}

	// 200 status
	results200 := optHar.SearchByStatusCode(200)
	if len(results200) != 1 {
		t.Errorf("SearchByStatusCode(200) returned %d results, want 1", len(results200))
	}

	// 201 status
	results201 := optHar.SearchByStatusCode(201)
	if len(results201) != 1 {
		t.Errorf("SearchByStatusCode(201) returned %d results, want 1", len(results201))
	}

	// 404 status
	results404 := optHar.SearchByStatusCode(404)
	if len(results404) != 1 {
		t.Errorf("SearchByStatusCode(404) returned %d results, want 1", len(results404))
	}

	// 500 status - no match
	results500 := optHar.SearchByStatusCode(500)
	if len(results500) != 0 {
		t.Errorf("SearchByStatusCode(500) returned %d results, want 0", len(results500))
	}
}

// ---------------------------------------------------------------------------
// GetRequestHeaderValue / GetResponseHeaderValue
// ---------------------------------------------------------------------------

func TestGetRequestHeaderValue(t *testing.T) {
	req := &OptimizedRequest{
		Method: MethodGET,
		URL:    "https://example.com/",
		Headers: map[string]string{
			"Content-Type": "application/json",
			"Accept":       "text/html",
			"X-Request-Id": "abc-123",
		},
	}

	// Existing header
	val, ok := req.GetRequestHeaderValue("Content-Type")
	if !ok || val != "application/json" {
		t.Errorf("GetRequestHeaderValue(Content-Type) = %q, ok=%v, want %q, ok=true", val, ok, "application/json")
	}

	// Case-sensitive lookup (headers are stored as-is)
	val, ok = req.GetRequestHeaderValue("content-type")
	if ok {
		t.Errorf("GetRequestHeaderValue(content-type) = %q, ok=%v, want ok=false (case-sensitive)", val, ok)
	}

	// Missing header
	val, ok = req.GetRequestHeaderValue("Authorization")
	if ok || val != "" {
		t.Errorf("GetRequestHeaderValue(Authorization) = %q, ok=%v, want empty and ok=false", val, ok)
	}
}

func TestGetResponseHeaderValue(t *testing.T) {
	resp := &OptimizedResponse{
		Status:     200,
		StatusText: "OK",
		Headers: map[string]string{
			"Content-Type":  "text/html; charset=utf-8",
			"X-Custom":      "value1",
			"Cache-Control": "no-cache",
		},
	}

	// Existing headers
	val, ok := resp.GetResponseHeaderValue("Content-Type")
	if !ok || val != "text/html; charset=utf-8" {
		t.Errorf("GetResponseHeaderValue(Content-Type) = %q, ok=%v, want %q, ok=true", val, ok, "text/html; charset=utf-8")
	}

	val, ok = resp.GetResponseHeaderValue("X-Custom")
	if !ok || val != "value1" {
		t.Errorf("GetResponseHeaderValue(X-Custom) = %q, ok=%v, want %q, ok=true", val, ok, "value1")
	}

	// Missing header
	val, ok = resp.GetResponseHeaderValue("Set-Cookie")
	if ok || val != "" {
		t.Errorf("GetResponseHeaderValue(Set-Cookie) = %q, ok=%v, want empty and ok=false", val, ok)
	}
}

// ---------------------------------------------------------------------------
// HTTPMethod.String()
// ---------------------------------------------------------------------------

func TestHTTPMethodString(t *testing.T) {
	tests := []struct {
		method   HTTPMethod
		expected string
	}{
		{MethodGET, "GET"},
		{MethodPOST, "POST"},
		{MethodPUT, "PUT"},
		{MethodDELETE, "DELETE"},
		{MethodHEAD, "HEAD"},
		{MethodOPTIONS, "OPTIONS"},
		{MethodPATCH, "PATCH"},
		{MethodCONNECT, "CONNECT"},
		{MethodTRACE, "TRACE"},
		{MethodUnknown, "UNKNOWN"},
		{HTTPMethod(255), "UNKNOWN"}, // out-of-range value
	}

	for _, tt := range tests {
		got := tt.method.String()
		if got != tt.expected {
			t.Errorf("HTTPMethod(%d).String() = %q, want %q", tt.method, got, tt.expected)
		}
	}
}

// ---------------------------------------------------------------------------
// ParseMethod
// ---------------------------------------------------------------------------

func TestParseMethod(t *testing.T) {
	tests := []struct {
		input    string
		expected HTTPMethod
	}{
		{"GET", MethodGET},
		{"POST", MethodPOST},
		{"PUT", MethodPUT},
		{"DELETE", MethodDELETE},
		{"HEAD", MethodHEAD},
		{"OPTIONS", MethodOPTIONS},
		{"PATCH", MethodPATCH},
		{"CONNECT", MethodCONNECT},
		{"TRACE", MethodTRACE},
		{"get", MethodGET},     // case-insensitive
		{"Post", MethodPOST},   // case-insensitive
		{"UNKNOWN", MethodUnknown},
		{"", MethodUnknown},
		{"INVALID", MethodUnknown},
	}

	for _, tt := range tests {
		got := ParseMethod(tt.input)
		if got != tt.expected {
			t.Errorf("ParseMethod(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

// ---------------------------------------------------------------------------
// OptimizedResponse.ToStandard - Content, sizes, headers
// ---------------------------------------------------------------------------

func TestOptimizedResponseToStandard(t *testing.T) {
	text := "hello world"
	encoding := "utf-8"
	comment := "resp comment"
	headerSize := 300
	bodySize := 100

	resp := OptimizedResponse{
		Status:      200,
		StatusText:  "OK",
		HTTPVersion: "HTTP/1.1",
		Headers: map[string]string{
			"Content-Type": "text/plain",
		},
		Content: &OptimizedContent{
			Size:     11,
			MimeType: "text/plain",
			Text:     &text,
			Encoding: &encoding,
			Comment:  &comment,
		},
		RedirectURL:  "/redirect",
		HeadersSize:  &headerSize,
		BodySize:     &bodySize,
	}

	standard := resp.ToStandard()

	if standard.Status != 200 {
		t.Errorf("Status = %d, want 200", standard.Status)
	}
	if standard.StatusText != "OK" {
		t.Errorf("StatusText = %q, want %q", standard.StatusText, "OK")
	}
	if standard.HTTPVersion != "HTTP/1.1" {
		t.Errorf("HTTPVersion = %q, want %q", standard.HTTPVersion, "HTTP/1.1")
	}
	if standard.RedirectURL != "/redirect" {
		t.Errorf("RedirectURL = %q, want %q", standard.RedirectURL, "/redirect")
	}
	if standard.HeadersSize != 300 {
		t.Errorf("HeadersSize = %d, want 300", standard.HeadersSize)
	}
	if standard.BodySize != 100 {
		t.Errorf("BodySize = %d, want 100", standard.BodySize)
	}
	if standard.Content.Text != text {
		t.Errorf("Content.Text = %q, want %q", standard.Content.Text, text)
	}
	if standard.Content.Encoding != encoding {
		t.Errorf("Content.Encoding = %q, want %q", standard.Content.Encoding, encoding)
	}
	if standard.Content.Comment != comment {
		t.Errorf("Content.Comment = %q, want %q", standard.Content.Comment, comment)
	}
}

func TestOptimizedResponseToStandard_NilContent(t *testing.T) {
	resp := OptimizedResponse{
		Status:     404,
		StatusText: "Not Found",
	}

	standard := resp.ToStandard()

	// When Content is nil, the resulting Content should be zero-valued
	if standard.Content.Size != 0 {
		t.Errorf("Content.Size = %d, want 0 for nil Content", standard.Content.Size)
	}
	if standard.Content.MimeType != "" {
		t.Errorf("Content.MimeType = %q, want empty for nil Content", standard.Content.MimeType)
	}
}

// ---------------------------------------------------------------------------
// OptimizedTimings.ToStandard
// ---------------------------------------------------------------------------

func TestOptimizedTimingsToStandard(t *testing.T) {
	blocked := 10.0
	dns := 5.0
	connect := 15.0
	send := 3.0
	wait := 100.0
	receive := 17.5
	ssl := 8.0

	timings := OptimizedTimings{
		Blocked: &blocked,
		DNS:     &dns,
		Connect: &connect,
		Send:    &send,
		Wait:    &wait,
		Receive: &receive,
		Ssl:     &ssl,
	}

	standard := timings.ToStandard()

	if standard.Blocked != blocked {
		t.Errorf("Blocked = %v, want %v", standard.Blocked, blocked)
	}
	if standard.DNS != dns {
		t.Errorf("DNS = %v, want %v", standard.DNS, dns)
	}
	if standard.Connect != connect {
		t.Errorf("Connect = %v, want %v", standard.Connect, connect)
	}
	if standard.Send != send {
		t.Errorf("Send = %v, want %v", standard.Send, send)
	}
	if standard.Wait != wait {
		t.Errorf("Wait = %v, want %v", standard.Wait, wait)
	}
	if standard.Receive != receive {
		t.Errorf("Receive = %v, want %v", standard.Receive, receive)
	}
	if standard.Ssl != ssl {
		t.Errorf("Ssl = %v, want %v", standard.Ssl, ssl)
	}
}

func TestOptimizedTimingsToStandard_NilFields(t *testing.T) {
	// All nil timings
	timings := OptimizedTimings{}

	standard := timings.ToStandard()

	// When nil, ToStandard returns -1 for each field
	if standard.Blocked != -1 {
		t.Errorf("Blocked = %v, want -1 for nil", standard.Blocked)
	}
	if standard.DNS != -1 {
		t.Errorf("DNS = %v, want -1 for nil", standard.DNS)
	}
	if standard.Connect != -1 {
		t.Errorf("Connect = %v, want -1 for nil", standard.Connect)
	}
	if standard.Send != -1 {
		t.Errorf("Send = %v, want -1 for nil", standard.Send)
	}
	if standard.Wait != -1 {
		t.Errorf("Wait = %v, want -1 for nil", standard.Wait)
	}
	if standard.Receive != -1 {
		t.Errorf("Receive = %v, want -1 for nil", standard.Receive)
	}
	if standard.Ssl != -1 {
		t.Errorf("Ssl = %v, want -1 for nil", standard.Ssl)
	}
}

// ---------------------------------------------------------------------------
// OptimizedContent interface methods (GetSize, GetMimeType, GetText, GetEncoding)
// ---------------------------------------------------------------------------

func TestOptimizedContentInterfaceMethods(t *testing.T) {
	text := "body content"
	encoding := "base64"

	content := &OptimizedContent{
		Size:     42,
		MimeType: "text/plain",
		Text:     &text,
		Encoding: &encoding,
	}

	if content.GetSize() != 42 {
		t.Errorf("GetSize() = %d, want 42", content.GetSize())
	}
	if content.GetMimeType() != "text/plain" {
		t.Errorf("GetMimeType() = %q, want %q", content.GetMimeType(), "text/plain")
	}
	if content.GetText() != text {
		t.Errorf("GetText() = %q, want %q", content.GetText(), text)
	}
	if content.GetEncoding() != encoding {
		t.Errorf("GetEncoding() = %q, want %q", content.GetEncoding(), encoding)
	}
	if content.GetCompression() != 0 {
		t.Errorf("GetCompression() = %d, want 0 (not tracked)", content.GetCompression())
	}
}

func TestOptimizedContentNilPointerMethods(t *testing.T) {
	content := &OptimizedContent{
		Size:     0,
		MimeType: "",
		// Text, Encoding, Comment all nil
	}

	if content.GetText() != "" {
		t.Errorf("GetText() = %q, want empty for nil Text", content.GetText())
	}
	if content.GetEncoding() != "" {
		t.Errorf("GetEncoding() = %q, want empty for nil Encoding", content.GetEncoding())
	}
}

// ---------------------------------------------------------------------------
// OptimizedTimings interface getters
// ---------------------------------------------------------------------------

func TestOptimizedTimingsGetters(t *testing.T) {
	blocked := 12.5
	dns := 6.0
	connect := 20.0
	send := 3.0
	wait := 50.0
	receive := 10.0
	ssl := 15.0

	timings := OptimizedTimings{
		Blocked: &blocked,
		DNS:     &dns,
		Connect: &connect,
		Send:    &send,
		Wait:    &wait,
		Receive: &receive,
		Ssl:     &ssl,
	}

	if timings.GetBlocked() != 12.5 {
		t.Errorf("GetBlocked() = %v, want 12.5", timings.GetBlocked())
	}
	if timings.GetDNS() != 6.0 {
		t.Errorf("GetDNS() = %v, want 6.0", timings.GetDNS())
	}
	if timings.GetConnect() != 20.0 {
		t.Errorf("GetConnect() = %v, want 20.0", timings.GetConnect())
	}
	if timings.GetSend() != 3.0 {
		t.Errorf("GetSend() = %v, want 3.0", timings.GetSend())
	}
	if timings.GetWait() != 50.0 {
		t.Errorf("GetWait() = %v, want 50.0", timings.GetWait())
	}
	if timings.GetReceive() != 10.0 {
		t.Errorf("GetReceive() = %v, want 10.0", timings.GetReceive())
	}
	if timings.GetSSL() != 15.0 {
		t.Errorf("GetSSL() = %v, want 15.0", timings.GetSSL())
	}
}

func TestOptimizedTimingsNilGetters(t *testing.T) {
	timings := OptimizedTimings{}

	if timings.GetBlocked() != -1 {
		t.Errorf("GetBlocked() = %v, want -1 for nil", timings.GetBlocked())
	}
	if timings.GetDNS() != -1 {
		t.Errorf("GetDNS() = %v, want -1 for nil", timings.GetDNS())
	}
	if timings.GetConnect() != -1 {
		t.Errorf("GetConnect() = %v, want -1 for nil", timings.GetConnect())
	}
	if timings.GetSend() != -1 {
		t.Errorf("GetSend() = %v, want -1 for nil", timings.GetSend())
	}
	if timings.GetWait() != -1 {
		t.Errorf("GetWait() = %v, want -1 for nil", timings.GetWait())
	}
	if timings.GetReceive() != -1 {
		t.Errorf("GetReceive() = %v, want -1 for nil", timings.GetReceive())
	}
	if timings.GetSSL() != -1 {
		t.Errorf("GetSSL() = %v, want -1 for nil", timings.GetSSL())
	}
}

// ---------------------------------------------------------------------------
// OptimizedRequest interface methods (GetMethod, GetURL, etc.)
// ---------------------------------------------------------------------------

func TestOptimizedRequestMethod(t *testing.T) {
	req := &OptimizedRequest{
		Method:      MethodPOST,
		URL:         "https://example.com/api",
		HTTPVersion: "HTTP/2.0",
	}

	if req.GetMethod() != "POST" {
		t.Errorf("GetMethod() = %q, want %q", req.GetMethod(), "POST")
	}
	if req.GetURL() != "https://example.com/api" {
		t.Errorf("GetURL() = %q, want %q", req.GetURL(), "https://example.com/api")
	}
	if req.GetHTTPVersion() != "HTTP/2.0" {
		t.Errorf("GetHTTPVersion() = %q, want %q", req.GetHTTPVersion(), "HTTP/2.0")
	}
}

func TestOptimizedRequestGetBodySizeAndGetHeadersSize(t *testing.T) {
	headersSize := 500
	bodySize := 1024

	req := &OptimizedRequest{
		HeadersSize: &headersSize,
		BodySize:    &bodySize,
	}

	if req.GetHeadersSize() != 500 {
		t.Errorf("GetHeadersSize() = %d, want 500", req.GetHeadersSize())
	}
	if req.GetBodySize() != 1024 {
		t.Errorf("GetBodySize() = %d, want 1024", req.GetBodySize())
	}

	// Nil sizes
	req2 := &OptimizedRequest{}
	if req2.GetHeadersSize() != 0 {
		t.Errorf("GetHeadersSize() = %d, want 0 for nil", req2.GetHeadersSize())
	}
	if req2.GetBodySize() != 0 {
		t.Errorf("GetBodySize() = %d, want 0 for nil", req2.GetBodySize())
	}
}

func TestOptimizedRequestGetQueryString(t *testing.T) {
	req := &OptimizedRequest{
		QueryString: map[string]string{
			"q":     "golang",
			"page":  "1",
			"limit": "10",
		},
	}

	qs := req.GetQueryString()
	if len(qs) != 3 {
		t.Fatalf("len(GetQueryString()) = %d, want 3", len(qs))
	}

	// Verify all values are present (order not guaranteed)
	m := map[string]string{}
	for _, item := range qs {
		m[item.Name] = item.Value
	}
	if m["q"] != "golang" {
		t.Errorf("QueryString q = %q, want %q", m["q"], "golang")
	}
	if m["page"] != "1" {
		t.Errorf("QueryString page = %q, want %q", m["page"], "1")
	}
	if m["limit"] != "10" {
		t.Errorf("QueryString limit = %q, want %q", m["limit"], "10")
	}
}

func TestOptimizedRequestGetPostData(t *testing.T) {
	pd := &PostData{
		MimeType: "application/x-www-form-urlencoded",
		Text:     "key=val",
	}

	req := &OptimizedRequest{
		PostData: pd,
	}

	got := req.GetPostData()
	if got == nil {
		t.Fatal("GetPostData() returned nil, want non-nil")
	}
	if got.MimeType != "application/x-www-form-urlencoded" {
		t.Errorf("GetPostData().MimeType = %q, want %q", got.MimeType, "application/x-www-form-urlencoded")
	}

	// Nil PostData
	req2 := &OptimizedRequest{}
	if req2.GetPostData() != nil {
		t.Error("GetPostData() should return nil when PostData is nil")
	}
}

// ---------------------------------------------------------------------------
// OptimizedResponse interface methods
// ---------------------------------------------------------------------------

func TestOptimizedResponseGetSizeMethods(t *testing.T) {
	headersSize := 400
	bodySize := 2048

	resp := &OptimizedResponse{
		Status:      200,
		StatusText:  "OK",
		HTTPVersion: "HTTP/1.1",
		HeadersSize: &headersSize,
		BodySize:    &bodySize,
	}

	if resp.GetStatus() != 200 {
		t.Errorf("GetStatus() = %d, want 200", resp.GetStatus())
	}
	if resp.GetStatusText() != "OK" {
		t.Errorf("GetStatusText() = %q, want %q", resp.GetStatusText(), "OK")
	}
	if resp.GetHTTPVersion() != "HTTP/1.1" {
		t.Errorf("GetHTTPVersion() = %q, want %q", resp.GetHTTPVersion(), "HTTP/1.1")
	}
	if resp.GetHeadersSize() != 400 {
		t.Errorf("GetHeadersSize() = %d, want 400", resp.GetHeadersSize())
	}
	if resp.GetBodySize() != 2048 {
		t.Errorf("GetBodySize() = %d, want 2048", resp.GetBodySize())
	}

	// Nil sizes
	resp2 := &OptimizedResponse{}
	if resp2.GetHeadersSize() != 0 {
		t.Errorf("GetHeadersSize() = %d, want 0 for nil", resp2.GetHeadersSize())
	}
	if resp2.GetBodySize() != 0 {
		t.Errorf("GetBodySize() = %d, want 0 for nil", resp2.GetBodySize())
	}
}

func TestOptimizedResponseGetContent(t *testing.T) {
	content := &OptimizedContent{
		Size:     100,
		MimeType: "text/plain",
	}

	resp := &OptimizedResponse{
		Content: content,
	}

	got := resp.GetContent()
	if got == nil {
		t.Fatal("GetContent() returned nil, want non-nil")
	}
	if got.GetSize() != 100 {
		t.Errorf("GetContent().GetSize() = %d, want 100", got.GetSize())
	}

	// Nil content
	resp2 := &OptimizedResponse{}
	if resp2.GetContent() != nil {
		t.Error("GetContent() should return nil when Content is nil")
	}
}

// ---------------------------------------------------------------------------
// OptimizedHar interface methods (GetVersion, GetCreator, GetBrowser, GetEntries, GetPages)
// ---------------------------------------------------------------------------

func TestOptimizedHarInterfaceMethods(t *testing.T) {
	data := createTestOptimizedHarJSON()
	optHar, err := ParseHarOptimized(data)
	if err != nil {
		t.Fatalf("ParseHarOptimized error: %v", err)
	}

	if optHar.GetVersion() != "1.2" {
		t.Errorf("GetVersion() = %q, want %q", optHar.GetVersion(), "1.2")
	}
	if optHar.GetCreator().Name != "TestCreator" {
		t.Errorf("GetCreator().Name = %q, want %q", optHar.GetCreator().Name, "TestCreator")
	}
	// Browser is always empty for OptimizedHar
	if optHar.GetBrowser().Name != "" {
		t.Errorf("GetBrowser().Name = %q, want empty (not tracked)", optHar.GetBrowser().Name)
	}

	entries := optHar.GetEntries()
	if len(entries) != 3 {
		t.Errorf("len(GetEntries()) = %d, want 3", len(entries))
	}

	pages := optHar.GetPages()
	if len(pages) != 0 {
		t.Errorf("len(GetPages()) = %d, want 0", len(pages))
	}
}

// ---------------------------------------------------------------------------
// OptimizedEntries interface methods
// ---------------------------------------------------------------------------

func TestOptimizedEntriesInterfaceMethods(t *testing.T) {
	data := createTestOptimizedHarJSON()
	optHar, err := ParseHarOptimized(data)
	if err != nil {
		t.Fatalf("ParseHarOptimized error: %v", err)
	}

	entries := optHar.Log.Entries
	if len(entries) == 0 {
		t.Fatal("no entries in parsed HAR")
	}

	e0 := &entries[0]

	if e0.GetStartedDateTime().IsZero() {
		t.Error("GetStartedDateTime() returned zero time")
	}
	if e0.GetTime() != 150.5 {
		t.Errorf("GetTime() = %v, want 150.5", e0.GetTime())
	}
	if e0.GetPageref() != "page_0" {
		t.Errorf("GetPageref() = %q, want %q", e0.GetPageref(), "page_0")
	}

	req := e0.GetRequest()
	if req.GetMethod() != "GET" {
		t.Errorf("GetRequest().GetMethod() = %q, want %q", req.GetMethod(), "GET")
	}

	resp := e0.GetResponse()
	if resp.GetStatus() != 200 {
		t.Errorf("GetResponse().GetStatus() = %d, want 200", resp.GetStatus())
	}

	timings := e0.GetTimings()
	if timings.GetSend() != 3.0 {
		t.Errorf("GetTimings().GetSend() = %v, want 3.0", timings.GetSend())
	}
}
