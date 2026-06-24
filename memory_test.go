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
			Browser: Browser{
				Name:    "TestBrowser",
				Version: "2.0",
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
						Send:    2.0,
						Wait:    60.0,
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
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
}

func TestParseHarOptimizedInvalidJSON(t *testing.T) {
	_, err := ParseHarOptimized([]byte("not json"))
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
}

func TestParseHarOptimizedMalformedJSON(t *testing.T) {
	_, err := ParseHarOptimized([]byte(`{"log":}`))
	assertHarErrorCode(t, err, ErrCodeJSONParse)
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
	assertHarErrorCode(t, err, ErrCodeFileSystem)
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

func TestToOptimizedHarNil(t *testing.T) {
	if optimized := ToOptimizedHar(nil); optimized != nil {
		t.Fatalf("ToOptimizedHar(nil) = %#v, want nil", optimized)
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
		{"get", MethodGET},   // case-insensitive
		{"Post", MethodPOST}, // case-insensitive
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
		RedirectURL: "/redirect",
		HeadersSize: &headerSize,
		BodySize:    &bodySize,
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
	if optHar.GetBrowser().Name != "TestBrowser" {
		t.Errorf("GetBrowser().Name = %q, want %q", optHar.GetBrowser().Name, "TestBrowser")
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

// ---------------------------------------------------------------------------
// OptimizedHar.GetPages — with pages present
// ---------------------------------------------------------------------------

func TestOptimizedHarGetPages_WithPages(t *testing.T) {
	optHar := &OptimizedHar{}
	optHar.Log.Version = "1.2"
	optHar.Log.Creator = Creator{Name: "test", Version: "1.0"}
	optHar.Log.Browser = Browser{Name: "browser", Version: "2.0"}
	optHar.Log.Pages = []Pages{
		{
			ID:              "page_1",
			Title:           "First Page",
			StartedDateTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			PageTimings: PageTimings{
				OnContentLoad: 100.0,
				OnLoad:        200.0,
			},
		},
		{
			ID:              "page_2",
			Title:           "Second Page",
			StartedDateTime: time.Date(2024, 1, 1, 0, 1, 0, 0, time.UTC),
			PageTimings: PageTimings{
				OnContentLoad: 150.0,
				OnLoad:        300.0,
			},
		},
	}
	optHar.Log.Entries = []OptimizedEntries{}

	pages := optHar.GetPages()
	if len(pages) != 2 {
		t.Fatalf("len(GetPages()) = %d, want 2", len(pages))
	}

	// Verify the PageProvider interface methods work
	if pages[0].GetID() != "page_1" {
		t.Errorf("Pages[0].GetID() = %q, want %q", pages[0].GetID(), "page_1")
	}
	if pages[0].GetTitle() != "First Page" {
		t.Errorf("Pages[0].GetTitle() = %q, want %q", pages[0].GetTitle(), "First Page")
	}
	if pages[1].GetID() != "page_2" {
		t.Errorf("Pages[1].GetID() = %q, want %q", pages[1].GetID(), "page_2")
	}
	if pages[1].GetTitle() != "Second Page" {
		t.Errorf("Pages[1].GetTitle() = %q, want %q", pages[1].GetTitle(), "Second Page")
	}

	// Verify page timings through the provider interface
	pt0 := pages[0].GetPageTimings()
	if pt0.GetOnContentLoad() != 100.0 {
		t.Errorf("Pages[0].GetPageTimings().GetOnContentLoad() = %v, want 100.0", pt0.GetOnContentLoad())
	}
	if pt0.GetOnLoad() != 200.0 {
		t.Errorf("Pages[0].GetPageTimings().GetOnLoad() = %v, want 200.0", pt0.GetOnLoad())
	}
}

// ---------------------------------------------------------------------------
// OptimizedHar.ToStandard — interface method (not ToStandardHar)
// ---------------------------------------------------------------------------

func TestOptimizedHarToStandard(t *testing.T) {
	pageRef := "page_1"
	serverIP := "10.0.0.1"
	conn := "conn-1"
	headerSize := 200
	bodySize := 50

	optHar := &OptimizedHar{}
	optHar.Log.Version = "1.2"
	optHar.Log.Creator = Creator{Name: "test", Version: "1.0"}
	optHar.Log.Browser = Browser{Name: "browser", Version: "2.0"}
	optHar.Log.Pages = []Pages{
		{ID: "page_1", Title: "Test Page"},
	}
	optHar.Log.Entries = []OptimizedEntries{
		{
			StartedDateTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Time:            100.0,
			Request: OptimizedRequest{
				Method:      MethodGET,
				URL:         "https://example.com/test",
				HTTPVersion: "HTTP/1.1",
				Headers:     map[string]string{"Accept": "text/html"},
				Cookies: []Cookie{
					{Name: "session", Value: "abc", Domain: "example.com", Path: "/"},
				},
				QueryString: map[string]string{"q": "hello"},
				HeadersSize: &headerSize,
				BodySize:    &bodySize,
			},
			Response: OptimizedResponse{
				Status:      200,
				StatusText:  "OK",
				HTTPVersion: "HTTP/1.1",
				Headers:     map[string]string{"Content-Type": "text/html"},
				Cookies: []Cookie{
					{Name: "visited", Value: "true", Domain: "example.com"},
				},
				Content: &OptimizedContent{
					Size:     100,
					MimeType: "text/html",
				},
				HeadersSize: &headerSize,
				BodySize:    &bodySize,
			},
			PageRef:    &pageRef,
			ServerIP:   &serverIP,
			Connection: &conn,
			Timings:    OptimizedTimings{},
		},
	}

	// Call ToStandard via the HARProvider interface
	var provider HARProvider = optHar
	standard := provider.ToStandard()

	if standard == nil {
		t.Fatal("ToStandard() returned nil")
	}
	if standard.Log.Version != "1.2" {
		t.Errorf("Version = %q, want %q", standard.Log.Version, "1.2")
	}
	if standard.Log.Browser.Name != "browser" {
		t.Errorf("Browser.Name = %q, want %q", standard.Log.Browser.Name, "browser")
	}
	if len(standard.Log.Pages) != 1 {
		t.Errorf("len(Pages) = %d, want 1", len(standard.Log.Pages))
	}
	if standard.Log.Pages[0].ID != "page_1" {
		t.Errorf("Pages[0].ID = %q, want %q", standard.Log.Pages[0].ID, "page_1")
	}
	if len(standard.Log.Entries) != 1 {
		t.Fatalf("len(Entries) = %d, want 1", len(standard.Log.Entries))
	}
	if standard.Log.Entries[0].Request.Method != "GET" {
		t.Errorf("Method = %q, want %q", standard.Log.Entries[0].Request.Method, "GET")
	}
	if standard.Log.Entries[0].Pageref != "page_1" {
		t.Errorf("Pageref = %q, want %q", standard.Log.Entries[0].Pageref, "page_1")
	}
	if standard.Log.Entries[0].ServerIPAddress != "10.0.0.1" {
		t.Errorf("ServerIPAddress = %q, want %q", standard.Log.Entries[0].ServerIPAddress, "10.0.0.1")
	}
	if standard.Log.Entries[0].Connection != "conn-1" {
		t.Errorf("Connection = %q, want %q", standard.Log.Entries[0].Connection, "conn-1")
	}
}

// ---------------------------------------------------------------------------
// OptimizedEntries.GetPageref — nil PageRef branch
// ---------------------------------------------------------------------------

func TestOptimizedEntriesGetPageref_Nil(t *testing.T) {
	entry := OptimizedEntries{
		StartedDateTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Time:            50.0,
		// PageRef is nil
	}

	if entry.GetPageref() != "" {
		t.Errorf("GetPageref() = %q, want empty string when PageRef is nil", entry.GetPageref())
	}
}

// ---------------------------------------------------------------------------
// OptimizedRequest.GetMethod — all HTTP method branches
// ---------------------------------------------------------------------------

func TestOptimizedRequestGetMethod_AllBranches(t *testing.T) {
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
		{HTTPMethod(99), "UNKNOWN"}, // out-of-range
	}

	for _, tt := range tests {
		req := &OptimizedRequest{Method: tt.method}
		got := req.GetMethod()
		if got != tt.expected {
			t.Errorf("OptimizedRequest{Method:%v}.GetMethod() = %q, want %q", tt.method, got, tt.expected)
		}
	}
}

// ---------------------------------------------------------------------------
// OptimizedRequest.GetHeaders — with and without headers
// ---------------------------------------------------------------------------

func TestOptimizedRequestGetHeaders(t *testing.T) {
	req := &OptimizedRequest{
		Method: MethodGET,
		URL:    "https://example.com/",
		Headers: map[string]string{
			"Content-Type": "application/json",
			"Accept":       "text/html",
			"X-Custom":     "value1",
		},
	}

	headers := req.GetHeaders()
	if len(headers) != 3 {
		t.Fatalf("len(GetHeaders()) = %d, want 3", len(headers))
	}

	// Verify all header values are accessible via HeaderProvider interface
	headerMap := map[string]string{}
	for _, h := range headers {
		headerMap[h.GetName()] = h.GetValue()
	}
	if headerMap["Content-Type"] != "application/json" {
		t.Errorf("Headers[Content-Type] = %q, want %q", headerMap["Content-Type"], "application/json")
	}
	if headerMap["Accept"] != "text/html" {
		t.Errorf("Headers[Accept] = %q, want %q", headerMap["Accept"], "text/html")
	}
	if headerMap["X-Custom"] != "value1" {
		t.Errorf("Headers[X-Custom] = %q, want %q", headerMap["X-Custom"], "value1")
	}
}

func TestOptimizedRequestGetHeaders_Empty(t *testing.T) {
	req := &OptimizedRequest{
		Method:  MethodGET,
		URL:     "https://example.com/",
		Headers: map[string]string{},
	}

	headers := req.GetHeaders()
	if len(headers) != 0 {
		t.Errorf("len(GetHeaders()) = %d, want 0 for empty headers map", len(headers))
	}
}

func TestOptimizedRequestGetHeaders_Nil(t *testing.T) {
	req := &OptimizedRequest{
		Method: MethodGET,
		URL:    "https://example.com/",
		// Headers is nil (zero value)
	}

	headers := req.GetHeaders()
	if len(headers) != 0 {
		t.Errorf("len(GetHeaders()) = %d, want 0 for nil headers map", len(headers))
	}
}

// ---------------------------------------------------------------------------
// OptimizedRequest.GetCookies — with and without cookies
// ---------------------------------------------------------------------------

func TestOptimizedRequestGetCookies(t *testing.T) {
	req := &OptimizedRequest{
		Method: MethodGET,
		URL:    "https://example.com/",
		Cookies: []Cookie{
			{Name: "session", Value: "abc123", Domain: "example.com", Path: "/", HTTPOnly: true, Secure: true},
			{Name: "lang", Value: "en", Domain: "example.com", Path: "/"},
		},
	}

	cookies := req.GetCookies()
	if len(cookies) != 2 {
		t.Fatalf("len(GetCookies()) = %d, want 2", len(cookies))
	}

	// Verify CookieProvider interface methods
	cookieMap := map[string]CookieProvider{}
	for _, c := range cookies {
		cookieMap[c.GetName()] = c
	}
	if c, ok := cookieMap["session"]; !ok {
		t.Error("cookie 'session' not found")
	} else {
		if c.GetValue() != "abc123" {
			t.Errorf("session cookie value = %q, want %q", c.GetValue(), "abc123")
		}
		if c.GetDomain() != "example.com" {
			t.Errorf("session cookie domain = %q, want %q", c.GetDomain(), "example.com")
		}
		if c.GetPath() != "/" {
			t.Errorf("session cookie path = %q, want %q", c.GetPath(), "/")
		}
		if !c.IsHTTPOnly() {
			t.Error("session cookie should be HTTPOnly")
		}
		if !c.IsSecure() {
			t.Error("session cookie should be Secure")
		}
	}
	if c, ok := cookieMap["lang"]; !ok {
		t.Error("cookie 'lang' not found")
	} else if c.GetValue() != "en" {
		t.Errorf("lang cookie value = %q, want %q", c.GetValue(), "en")
	}
}

func TestOptimizedRequestGetCookies_Empty(t *testing.T) {
	req := &OptimizedRequest{
		Method:  MethodGET,
		URL:     "https://example.com/",
		Cookies: []Cookie{},
	}

	cookies := req.GetCookies()
	if len(cookies) != 0 {
		t.Errorf("len(GetCookies()) = %d, want 0 for empty cookies", len(cookies))
	}
}

func TestOptimizedRequestGetCookies_Nil(t *testing.T) {
	req := &OptimizedRequest{
		Method: MethodGET,
		URL:    "https://example.com/",
		// Cookies is nil (zero value)
	}

	cookies := req.GetCookies()
	if len(cookies) != 0 {
		t.Errorf("len(GetCookies()) = %d, want 0 for nil cookies", len(cookies))
	}
}

// ---------------------------------------------------------------------------
// OptimizedResponse.GetHeaders — with and without headers
// ---------------------------------------------------------------------------

func TestOptimizedResponseGetHeaders(t *testing.T) {
	resp := &OptimizedResponse{
		Status:     200,
		StatusText: "OK",
		Headers: map[string]string{
			"Content-Type":  "text/html; charset=utf-8",
			"Cache-Control": "no-cache",
			"Set-Cookie":    "id=a3fWa; Expires=Thu, 21 Oct 2021 07:28:00 GMT",
		},
	}

	headers := resp.GetHeaders()
	if len(headers) != 3 {
		t.Fatalf("len(GetHeaders()) = %d, want 3", len(headers))
	}

	headerMap := map[string]string{}
	for _, h := range headers {
		headerMap[h.GetName()] = h.GetValue()
	}
	if headerMap["Content-Type"] != "text/html; charset=utf-8" {
		t.Errorf("Headers[Content-Type] = %q, want %q", headerMap["Content-Type"], "text/html; charset=utf-8")
	}
	if headerMap["Cache-Control"] != "no-cache" {
		t.Errorf("Headers[Cache-Control] = %q, want %q", headerMap["Cache-Control"], "no-cache")
	}
	if headerMap["Set-Cookie"] != "id=a3fWa; Expires=Thu, 21 Oct 2021 07:28:00 GMT" {
		t.Errorf("Headers[Set-Cookie] = %q, want correct Set-Cookie value", headerMap["Set-Cookie"])
	}
}

func TestOptimizedResponseGetHeaders_Empty(t *testing.T) {
	resp := &OptimizedResponse{
		Status:  200,
		Headers: map[string]string{},
	}

	headers := resp.GetHeaders()
	if len(headers) != 0 {
		t.Errorf("len(GetHeaders()) = %d, want 0 for empty headers map", len(headers))
	}
}

func TestOptimizedResponseGetHeaders_Nil(t *testing.T) {
	resp := &OptimizedResponse{
		Status: 200,
		// Headers is nil
	}

	headers := resp.GetHeaders()
	if len(headers) != 0 {
		t.Errorf("len(GetHeaders()) = %d, want 0 for nil headers map", len(headers))
	}
}

// ---------------------------------------------------------------------------
// OptimizedResponse.GetCookies — with and without cookies
// ---------------------------------------------------------------------------

func TestOptimizedResponseGetCookies(t *testing.T) {
	resp := &OptimizedResponse{
		Status:     200,
		StatusText: "OK",
		Cookies: []Cookie{
			{Name: "session_id", Value: "xyz789", Domain: "example.com", HTTPOnly: true, Secure: true},
			{Name: "tracking", Value: "enabled", Domain: "example.com", Path: "/", SameSite: "Lax"},
		},
	}

	cookies := resp.GetCookies()
	if len(cookies) != 2 {
		t.Fatalf("len(GetCookies()) = %d, want 2", len(cookies))
	}

	// Verify CookieProvider interface methods
	cookieMap := map[string]CookieProvider{}
	for _, c := range cookies {
		cookieMap[c.GetName()] = c
	}
	if c, ok := cookieMap["session_id"]; !ok {
		t.Error("cookie 'session_id' not found")
	} else {
		if c.GetValue() != "xyz789" {
			t.Errorf("session_id cookie value = %q, want %q", c.GetValue(), "xyz789")
		}
		if !c.IsHTTPOnly() {
			t.Error("session_id cookie should be HTTPOnly")
		}
		if !c.IsSecure() {
			t.Error("session_id cookie should be Secure")
		}
	}
	if c, ok := cookieMap["tracking"]; !ok {
		t.Error("cookie 'tracking' not found")
	} else {
		if c.GetSameSite() != "Lax" {
			t.Errorf("tracking cookie SameSite = %q, want %q", c.GetSameSite(), "Lax")
		}
	}
}

func TestOptimizedResponseGetCookies_Empty(t *testing.T) {
	resp := &OptimizedResponse{
		Status:  200,
		Cookies: []Cookie{},
	}

	cookies := resp.GetCookies()
	if len(cookies) != 0 {
		t.Errorf("len(GetCookies()) = %d, want 0 for empty cookies", len(cookies))
	}
}

func TestOptimizedResponseGetCookies_Nil(t *testing.T) {
	resp := &OptimizedResponse{
		Status: 200,
		// Cookies is nil
	}

	cookies := resp.GetCookies()
	if len(cookies) != 0 {
		t.Errorf("len(GetCookies()) = %d, want 0 for nil cookies", len(cookies))
	}
}

// ---------------------------------------------------------------------------
// convertToOptimizedEntry — uncovered branches
// The uncovered branches are in:
//   - entry.Request.HeadersSize == 0 (not setting pointer)
//   - entry.Request.BodySize == 0 (not setting pointer)
//   - entry.Response.HeadersSize == 0 (not setting pointer)
//   - entry.Response.BodySize == 0 (not setting pointer)
//   - entry.Response.TransferSize == 0 (not setting pointer)
//   - entry.Response.Content with various zero/empty fields
//   - entry.Timings with various zero fields
//   - entry.Cache with empty Cache{}
//   - entry.Pageref == ""
//   - entry.ServerIPAddress == ""
//   - entry.Connection == ""
// ---------------------------------------------------------------------------

func TestConvertToOptimizedEntry_AllZeroOptionalFields(t *testing.T) {
	// Test entry where all optional fields are zero/empty — verifies the
	// branches that skip setting pointers when values are 0 or empty.
	entry := Entries{
		StartedDateTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Time:            10.0,
		Request: Request{
			Method:      "GET",
			URL:         "https://example.com/",
			HTTPVersion: "HTTP/1.1",
			Headers:     []Headers{},
			QueryString: []QueryString{},
			HeadersSize: 0, // zero — should NOT set HeadersSize pointer
			BodySize:    0, // zero — should NOT set BodySize pointer
		},
		Response: Response{
			Status:      200,
			StatusText:  "OK",
			HTTPVersion: "HTTP/1.1",
			Headers:     []Headers{},
			// Content with all zero/empty fields — should NOT create OptimizedContent
			Content: Content{
				Size:     0,
				MimeType: "",
				Text:     "",
				Encoding: "",
				Comment:  "",
			},
			RedirectURL:  "",
			HeadersSize:  0, // zero — should NOT set HeadersSize pointer
			BodySize:     0, // zero — should NOT set BodySize pointer
			TransferSize: 0, // zero — should NOT set TransferSize pointer
		},
		Cache: Cache{}, // empty — should NOT set Cache pointer
		Timings: Timings{
			Blocked:         0, // zero — should NOT set pointer
			DNS:             0,
			Connect:         0,
			Send:            0,
			Wait:            0,
			Receive:         0,
			Ssl:             0,
			BlockedQueueing: 0,
			BlockedProxy:    0,
		},
		// Pageref, ServerIPAddress, Connection all empty
	}

	optimized := convertToOptimizedEntry(entry)

	// Request sizes should be nil (zero was not set)
	if optimized.Request.HeadersSize != nil {
		t.Errorf("Request.HeadersSize = %v, want nil for zero value", optimized.Request.HeadersSize)
	}
	if optimized.Request.BodySize != nil {
		t.Errorf("Request.BodySize = %v, want nil for zero value", optimized.Request.BodySize)
	}

	// Response sizes should be nil
	if optimized.Response.HeadersSize != nil {
		t.Errorf("Response.HeadersSize = %v, want nil for zero value", optimized.Response.HeadersSize)
	}
	if optimized.Response.BodySize != nil {
		t.Errorf("Response.BodySize = %v, want nil for zero value", optimized.Response.BodySize)
	}
	if optimized.Response.TransferSize != nil {
		t.Errorf("Response.TransferSize = %v, want nil for zero value", optimized.Response.TransferSize)
	}

	// Content should be nil when all fields are zero/empty
	if optimized.Response.Content != nil {
		t.Errorf("Response.Content = %+v, want nil when all content fields are zero/empty", optimized.Response.Content)
	}

	// Cache should be nil when empty
	if optimized.Cache != nil {
		t.Errorf("Cache = %+v, want nil for empty Cache", optimized.Cache)
	}

	// Timings should all be nil
	if optimized.Timings.Blocked != nil {
		t.Errorf("Timings.Blocked = %v, want nil for zero value", optimized.Timings.Blocked)
	}
	if optimized.Timings.DNS != nil {
		t.Errorf("Timings.DNS = %v, want nil for zero value", optimized.Timings.DNS)
	}
	if optimized.Timings.Connect != nil {
		t.Errorf("Timings.Connect = %v, want nil for zero value", optimized.Timings.Connect)
	}
	if optimized.Timings.Send != nil {
		t.Errorf("Timings.Send = %v, want nil for zero value", optimized.Timings.Send)
	}
	if optimized.Timings.Wait != nil {
		t.Errorf("Timings.Wait = %v, want nil for zero value", optimized.Timings.Wait)
	}
	if optimized.Timings.Receive != nil {
		t.Errorf("Timings.Receive = %v, want nil for zero value", optimized.Timings.Receive)
	}
	if optimized.Timings.Ssl != nil {
		t.Errorf("Timings.Ssl = %v, want nil for zero value", optimized.Timings.Ssl)
	}
	if optimized.Timings.BlockedQueueing != nil {
		t.Errorf("Timings.BlockedQueueing = %v, want nil for zero value", optimized.Timings.BlockedQueueing)
	}
	if optimized.Timings.BlockedProxy != nil {
		t.Errorf("Timings.BlockedProxy = %v, want nil for zero value", optimized.Timings.BlockedProxy)
	}

	// Optional string fields should be nil when empty
	if optimized.PageRef != nil {
		t.Errorf("PageRef = %v, want nil for empty Pageref", optimized.PageRef)
	}
	if optimized.ServerIP != nil {
		t.Errorf("ServerIP = %v, want nil for empty ServerIPAddress", optimized.ServerIP)
	}
	if optimized.Connection != nil {
		t.Errorf("Connection = %v, want nil for empty Connection", optimized.Connection)
	}
}

func TestConvertToOptimizedEntry_NonZeroOptionalFields(t *testing.T) {
	// Test entry where all optional fields are non-zero — verifies the
	// branches that DO set pointers when values are non-zero.
	entry := Entries{
		StartedDateTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Time:            10.0,
		Request: Request{
			Method:      "POST",
			URL:         "https://example.com/api",
			HTTPVersion: "HTTP/1.1",
			Headers:     []Headers{},
			QueryString: []QueryString{},
			HeadersSize: 300, // non-zero — should set HeadersSize pointer
			BodySize:    100, // non-zero — should set BodySize pointer
		},
		Response: Response{
			Status:      201,
			StatusText:  "Created",
			HTTPVersion: "HTTP/1.1",
			Headers:     []Headers{},
			Content: Content{
				Size:     50,
				MimeType: "application/json",
				Text:     "hello",
				Encoding: "utf-8",
				Comment:  "test comment",
			},
			RedirectURL:  "/new",
			HeadersSize:  250, // non-zero
			BodySize:     50,  // non-zero
			TransferSize: 100, // non-zero
		},
		Cache: Cache{
			Comment: "cache hit",
		},
		Timings: Timings{
			Blocked:         1.0,
			DNS:             2.0,
			Connect:         3.0,
			Send:            4.0,
			Wait:            5.0,
			Receive:         6.0,
			Ssl:             7.0,
			BlockedQueueing: 8.0,
			BlockedProxy:    9.0,
		},
		Pageref:         "page_abc",
		ServerIPAddress: "192.168.1.1",
		Connection:      "keep-alive",
	}

	optimized := convertToOptimizedEntry(entry)

	// Request sizes should be set
	if optimized.Request.HeadersSize == nil || *optimized.Request.HeadersSize != 300 {
		t.Errorf("Request.HeadersSize = %v, want pointer to 300", optimized.Request.HeadersSize)
	}
	if optimized.Request.BodySize == nil || *optimized.Request.BodySize != 100 {
		t.Errorf("Request.BodySize = %v, want pointer to 100", optimized.Request.BodySize)
	}

	// Response sizes should be set
	if optimized.Response.HeadersSize == nil || *optimized.Response.HeadersSize != 250 {
		t.Errorf("Response.HeadersSize = %v, want pointer to 250", optimized.Response.HeadersSize)
	}
	if optimized.Response.BodySize == nil || *optimized.Response.BodySize != 50 {
		t.Errorf("Response.BodySize = %v, want pointer to 50", optimized.Response.BodySize)
	}
	if optimized.Response.TransferSize == nil || *optimized.Response.TransferSize != 100 {
		t.Errorf("Response.TransferSize = %v, want pointer to 100", optimized.Response.TransferSize)
	}

	// Content should be set
	if optimized.Response.Content == nil {
		t.Fatal("Response.Content is nil, want non-nil")
	}
	if optimized.Response.Content.Size != 50 {
		t.Errorf("Content.Size = %d, want 50", optimized.Response.Content.Size)
	}
	if optimized.Response.Content.MimeType != "application/json" {
		t.Errorf("Content.MimeType = %q, want %q", optimized.Response.Content.MimeType, "application/json")
	}
	if optimized.Response.Content.Text == nil || *optimized.Response.Content.Text != "hello" {
		t.Errorf("Content.Text = %v, want pointer to %q", optimized.Response.Content.Text, "hello")
	}
	if optimized.Response.Content.Encoding == nil || *optimized.Response.Content.Encoding != "utf-8" {
		t.Errorf("Content.Encoding = %v, want pointer to %q", optimized.Response.Content.Encoding, "utf-8")
	}
	if optimized.Response.Content.Comment == nil || *optimized.Response.Content.Comment != "test comment" {
		t.Errorf("Content.Comment = %v, want pointer to %q", optimized.Response.Content.Comment, "test comment")
	}

	// Cache should be set
	if optimized.Cache == nil || optimized.Cache.Comment != "cache hit" {
		t.Errorf("Cache = %+v, want non-nil with Comment = %q", optimized.Cache, "cache hit")
	}

	// Timings should all be set
	if optimized.Timings.Blocked == nil || *optimized.Timings.Blocked != 1.0 {
		t.Errorf("Timings.Blocked = %v, want pointer to 1.0", optimized.Timings.Blocked)
	}
	if optimized.Timings.DNS == nil || *optimized.Timings.DNS != 2.0 {
		t.Errorf("Timings.DNS = %v, want pointer to 2.0", optimized.Timings.DNS)
	}
	if optimized.Timings.Connect == nil || *optimized.Timings.Connect != 3.0 {
		t.Errorf("Timings.Connect = %v, want pointer to 3.0", optimized.Timings.Connect)
	}
	if optimized.Timings.Send == nil || *optimized.Timings.Send != 4.0 {
		t.Errorf("Timings.Send = %v, want pointer to 4.0", optimized.Timings.Send)
	}
	if optimized.Timings.Wait == nil || *optimized.Timings.Wait != 5.0 {
		t.Errorf("Timings.Wait = %v, want pointer to 5.0", optimized.Timings.Wait)
	}
	if optimized.Timings.Receive == nil || *optimized.Timings.Receive != 6.0 {
		t.Errorf("Timings.Receive = %v, want pointer to 6.0", optimized.Timings.Receive)
	}
	if optimized.Timings.Ssl == nil || *optimized.Timings.Ssl != 7.0 {
		t.Errorf("Timings.Ssl = %v, want pointer to 7.0", optimized.Timings.Ssl)
	}
	if optimized.Timings.BlockedQueueing == nil || *optimized.Timings.BlockedQueueing != 8.0 {
		t.Errorf("Timings.BlockedQueueing = %v, want pointer to 8.0", optimized.Timings.BlockedQueueing)
	}
	if optimized.Timings.BlockedProxy == nil || *optimized.Timings.BlockedProxy != 9.0 {
		t.Errorf("Timings.BlockedProxy = %v, want pointer to 9.0", optimized.Timings.BlockedProxy)
	}

	// Optional string fields should be set
	if optimized.PageRef == nil || *optimized.PageRef != "page_abc" {
		t.Errorf("PageRef = %v, want pointer to %q", optimized.PageRef, "page_abc")
	}
	if optimized.ServerIP == nil || *optimized.ServerIP != "192.168.1.1" {
		t.Errorf("ServerIP = %v, want pointer to %q", optimized.ServerIP, "192.168.1.1")
	}
	if optimized.Connection == nil || *optimized.Connection != "keep-alive" {
		t.Errorf("Connection = %v, want pointer to %q", optimized.Connection, "keep-alive")
	}
}

func TestConvertToOptimizedEntry_ContentOnlyMimeType(t *testing.T) {
	// Content with only MimeType set (Size=0, Text="", Encoding="", Comment="")
	// Should still create OptimizedContent because MimeType != ""
	entry := Entries{
		StartedDateTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Time:            10.0,
		Request: Request{
			Method: "GET",
			URL:    "https://example.com/",
		},
		Response: Response{
			Status:     200,
			StatusText: "OK",
			Content: Content{
				Size:     0,
				MimeType: "text/html", // only this is non-zero/empty
				Text:     "",
				Encoding: "",
				Comment:  "",
			},
		},
		Timings: Timings{},
	}

	optimized := convertToOptimizedEntry(entry)

	if optimized.Response.Content == nil {
		t.Fatal("Response.Content is nil, want non-nil when MimeType is set")
	}
	if optimized.Response.Content.MimeType != "text/html" {
		t.Errorf("Content.MimeType = %q, want %q", optimized.Response.Content.MimeType, "text/html")
	}
	if optimized.Response.Content.Size != 0 {
		t.Errorf("Content.Size = %d, want 0", optimized.Response.Content.Size)
	}
	if optimized.Response.Content.Text != nil {
		t.Errorf("Content.Text = %v, want nil for empty text", optimized.Response.Content.Text)
	}
	if optimized.Response.Content.Encoding != nil {
		t.Errorf("Content.Encoding = %v, want nil for empty encoding", optimized.Response.Content.Encoding)
	}
	if optimized.Response.Content.Comment != nil {
		t.Errorf("Content.Comment = %v, want nil for empty comment", optimized.Response.Content.Comment)
	}
}

func TestConvertToOptimizedEntry_ContentOnlySize(t *testing.T) {
	// Content with only Size set (MimeType="", Text="", etc.)
	// Should still create OptimizedContent because Size != 0
	entry := Entries{
		StartedDateTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Time:            10.0,
		Request: Request{
			Method: "GET",
			URL:    "https://example.com/",
		},
		Response: Response{
			Status:     200,
			StatusText: "OK",
			Content: Content{
				Size: 256, // only this is non-zero
			},
		},
		Timings: Timings{},
	}

	optimized := convertToOptimizedEntry(entry)

	if optimized.Response.Content == nil {
		t.Fatal("Response.Content is nil, want non-nil when Size is non-zero")
	}
	if optimized.Response.Content.Size != 256 {
		t.Errorf("Content.Size = %d, want 256", optimized.Response.Content.Size)
	}
}

func TestConvertToOptimizedEntry_CacheWithBeforeRequest(t *testing.T) {
	// Cache with BeforeRequest set — should create Cache pointer
	entry := Entries{
		StartedDateTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Time:            10.0,
		Request: Request{
			Method: "GET",
			URL:    "https://example.com/",
		},
		Response: Response{
			Status:     200,
			StatusText: "OK",
		},
		Cache: Cache{
			BeforeRequest: &BeforeRequest{
				Expires: time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
			},
		},
		Timings: Timings{},
	}

	optimized := convertToOptimizedEntry(entry)

	if optimized.Cache == nil {
		t.Fatal("Cache is nil, want non-nil when BeforeRequest is set")
	}
	if optimized.Cache.BeforeRequest == nil {
		t.Error("Cache.BeforeRequest is nil, want non-nil")
	}
}

func TestConvertToOptimizedEntry_CacheWithAfterRequest(t *testing.T) {
	// Cache with AfterRequest set — should create Cache pointer
	entry := Entries{
		StartedDateTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Time:            10.0,
		Request: Request{
			Method: "GET",
			URL:    "https://example.com/",
		},
		Response: Response{
			Status:     200,
			StatusText: "OK",
		},
		Cache: Cache{
			AfterRequest: &AfterRequest{
				LastAccess: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		Timings: Timings{},
	}

	optimized := convertToOptimizedEntry(entry)

	if optimized.Cache == nil {
		t.Fatal("Cache is nil, want non-nil when AfterRequest is set")
	}
	if optimized.Cache.AfterRequest == nil {
		t.Error("Cache.AfterRequest is nil, want non-nil")
	}
}

// ---------------------------------------------------------------------------
// convertToStandardEntry — uncovered branches
// The uncovered branches are in:
//   - entry.Request.HeadersSize == nil
//   - entry.Request.BodySize == nil
//   - entry.Response.HeadersSize == nil
//   - entry.Response.BodySize == nil
//   - entry.Response.TransferSize == nil
//   - entry.Response.Content == nil
//   - entry.Timings.* == nil (all timing fields)
//   - entry.Cache == nil
//   - entry.PageRef == nil
//   - entry.ServerIP == nil
//   - entry.Connection == nil
// ---------------------------------------------------------------------------

func TestConvertToStandardEntry_AllNilOptionalFields(t *testing.T) {
	// Entry where all optional pointer fields are nil
	entry := OptimizedEntries{
		StartedDateTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Time:            10.0,
		Request: OptimizedRequest{
			Method:      MethodGET,
			URL:         "https://example.com/",
			HTTPVersion: "HTTP/1.1",
			Headers:     map[string]string{},
			QueryString: map[string]string{},
			// HeadersSize, BodySize are nil
		},
		Response: OptimizedResponse{
			Status:      200,
			StatusText:  "OK",
			HTTPVersion: "HTTP/1.1",
			Headers:     map[string]string{},
			// Content, HeadersSize, BodySize, TransferSize are nil
		},
		// Cache, PageRef, ServerIP, Connection are nil
		Timings: OptimizedTimings{
			// All timing fields are nil
		},
	}

	standard := convertToStandardEntry(entry)

	// Request sizes should be 0 (nil pointers)
	if standard.Request.HeadersSize != 0 {
		t.Errorf("Request.HeadersSize = %d, want 0 for nil pointer", standard.Request.HeadersSize)
	}
	if standard.Request.BodySize != 0 {
		t.Errorf("Request.BodySize = %d, want 0 for nil pointer", standard.Request.BodySize)
	}

	// Response sizes should be 0
	if standard.Response.HeadersSize != 0 {
		t.Errorf("Response.HeadersSize = %d, want 0 for nil pointer", standard.Response.HeadersSize)
	}
	if standard.Response.BodySize != 0 {
		t.Errorf("Response.BodySize = %d, want 0 for nil pointer", standard.Response.BodySize)
	}
	if standard.Response.TransferSize != 0 {
		t.Errorf("Response.TransferSize = %d, want 0 for nil pointer", standard.Response.TransferSize)
	}

	// Content should be zero-valued
	if standard.Response.Content.Size != 0 || standard.Response.Content.MimeType != "" {
		t.Errorf("Response.Content = %+v, want zero-valued for nil Content pointer", standard.Response.Content)
	}

	// Cache should be zero-valued
	if standard.Cache.Comment != "" || standard.Cache.BeforeRequest != nil || standard.Cache.AfterRequest != nil {
		t.Errorf("Cache = %+v, want zero-valued for nil Cache pointer", standard.Cache)
	}

	// Optional string fields should be empty
	if standard.Pageref != "" {
		t.Errorf("Pageref = %q, want empty for nil pointer", standard.Pageref)
	}
	if standard.ServerIPAddress != "" {
		t.Errorf("ServerIPAddress = %q, want empty for nil pointer", standard.ServerIPAddress)
	}
	if standard.Connection != "" {
		t.Errorf("Connection = %q, want empty for nil pointer", standard.Connection)
	}

	// Timings should all be 0 (nil pointers — not -1, that's the ToStandard path)
	if standard.Timings.Blocked != 0 {
		t.Errorf("Timings.Blocked = %v, want 0 for nil pointer", standard.Timings.Blocked)
	}
	if standard.Timings.DNS != 0 {
		t.Errorf("Timings.DNS = %v, want 0 for nil pointer", standard.Timings.DNS)
	}
	if standard.Timings.Connect != 0 {
		t.Errorf("Timings.Connect = %v, want 0 for nil pointer", standard.Timings.Connect)
	}
	if standard.Timings.Send != 0 {
		t.Errorf("Timings.Send = %v, want 0 for nil pointer", standard.Timings.Send)
	}
	if standard.Timings.Wait != 0 {
		t.Errorf("Timings.Wait = %v, want 0 for nil pointer", standard.Timings.Wait)
	}
	if standard.Timings.Receive != 0 {
		t.Errorf("Timings.Receive = %v, want 0 for nil pointer", standard.Timings.Receive)
	}
	if standard.Timings.Ssl != 0 {
		t.Errorf("Timings.Ssl = %v, want 0 for nil pointer", standard.Timings.Ssl)
	}
	if standard.Timings.BlockedQueueing != 0 {
		t.Errorf("Timings.BlockedQueueing = %v, want 0 for nil pointer", standard.Timings.BlockedQueueing)
	}
	if standard.Timings.BlockedProxy != 0 {
		t.Errorf("Timings.BlockedProxy = %v, want 0 for nil pointer", standard.Timings.BlockedProxy)
	}
}

func TestConvertToStandardEntry_AllNonNilOptionalFields(t *testing.T) {
	headerSize := 300
	bodySize := 100
	transferSize := 150
	text := "response body"
	encoding := "gzip"
	comment := "content comment"
	blocked := 1.0
	dns := 2.0
	connect := 3.0
	send := 4.0
	wait := 5.0
	receive := 6.0
	ssl := 7.0
	blockedQueueing := 8.0
	blockedProxy := 9.0
	pageRef := "page_xyz"
	serverIP := "10.0.0.5"
	connection := "conn-99"

	entry := OptimizedEntries{
		StartedDateTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Time:            10.0,
		Request: OptimizedRequest{
			Method:      MethodPOST,
			URL:         "https://example.com/api",
			HTTPVersion: "HTTP/1.1",
			Headers:     map[string]string{"Content-Type": "application/json"},
			QueryString: map[string]string{"id": "42"},
			HeadersSize: &headerSize,
			BodySize:    &bodySize,
		},
		Response: OptimizedResponse{
			Status:      201,
			StatusText:  "Created",
			HTTPVersion: "HTTP/1.1",
			Headers:     map[string]string{"Location": "/api/42"},
			Content: &OptimizedContent{
				Size:     100,
				MimeType: "application/json",
				Text:     &text,
				Encoding: &encoding,
				Comment:  &comment,
			},
			HeadersSize:  &headerSize,
			BodySize:     &bodySize,
			TransferSize: &transferSize,
		},
		Cache: &Cache{Comment: "cached"},
		Timings: OptimizedTimings{
			Blocked:         &blocked,
			DNS:             &dns,
			Connect:         &connect,
			Send:            &send,
			Wait:            &wait,
			Receive:         &receive,
			Ssl:             &ssl,
			BlockedQueueing: &blockedQueueing,
			BlockedProxy:    &blockedProxy,
		},
		PageRef:    &pageRef,
		ServerIP:   &serverIP,
		Connection: &connection,
	}

	standard := convertToStandardEntry(entry)

	// Request sizes
	if standard.Request.HeadersSize != headerSize {
		t.Errorf("Request.HeadersSize = %d, want %d", standard.Request.HeadersSize, headerSize)
	}
	if standard.Request.BodySize != bodySize {
		t.Errorf("Request.BodySize = %d, want %d", standard.Request.BodySize, bodySize)
	}

	// Response sizes
	if standard.Response.HeadersSize != headerSize {
		t.Errorf("Response.HeadersSize = %d, want %d", standard.Response.HeadersSize, headerSize)
	}
	if standard.Response.BodySize != bodySize {
		t.Errorf("Response.BodySize = %d, want %d", standard.Response.BodySize, bodySize)
	}
	if standard.Response.TransferSize != transferSize {
		t.Errorf("Response.TransferSize = %d, want %d", standard.Response.TransferSize, transferSize)
	}

	// Content
	if standard.Response.Content.Text != text {
		t.Errorf("Content.Text = %q, want %q", standard.Response.Content.Text, text)
	}
	if standard.Response.Content.Encoding != encoding {
		t.Errorf("Content.Encoding = %q, want %q", standard.Response.Content.Encoding, encoding)
	}
	if standard.Response.Content.Comment != comment {
		t.Errorf("Content.Comment = %q, want %q", standard.Response.Content.Comment, comment)
	}

	// Cache
	if standard.Cache.Comment != "cached" {
		t.Errorf("Cache.Comment = %q, want %q", standard.Cache.Comment, "cached")
	}

	// Timings
	if standard.Timings.Blocked != blocked {
		t.Errorf("Timings.Blocked = %v, want %v", standard.Timings.Blocked, blocked)
	}
	if standard.Timings.DNS != dns {
		t.Errorf("Timings.DNS = %v, want %v", standard.Timings.DNS, dns)
	}
	if standard.Timings.Connect != connect {
		t.Errorf("Timings.Connect = %v, want %v", standard.Timings.Connect, connect)
	}
	if standard.Timings.Send != send {
		t.Errorf("Timings.Send = %v, want %v", standard.Timings.Send, send)
	}
	if standard.Timings.Wait != wait {
		t.Errorf("Timings.Wait = %v, want %v", standard.Timings.Wait, wait)
	}
	if standard.Timings.Receive != receive {
		t.Errorf("Timings.Receive = %v, want %v", standard.Timings.Receive, receive)
	}
	if standard.Timings.Ssl != ssl {
		t.Errorf("Timings.Ssl = %v, want %v", standard.Timings.Ssl, ssl)
	}
	if standard.Timings.BlockedQueueing != blockedQueueing {
		t.Errorf("Timings.BlockedQueueing = %v, want %v", standard.Timings.BlockedQueueing, blockedQueueing)
	}
	if standard.Timings.BlockedProxy != blockedProxy {
		t.Errorf("Timings.BlockedProxy = %v, want %v", standard.Timings.BlockedProxy, blockedProxy)
	}

	// Optional string fields
	if standard.Pageref != pageRef {
		t.Errorf("Pageref = %q, want %q", standard.Pageref, pageRef)
	}
	if standard.ServerIPAddress != serverIP {
		t.Errorf("ServerIPAddress = %q, want %q", standard.ServerIPAddress, serverIP)
	}
	if standard.Connection != connection {
		t.Errorf("Connection = %q, want %q", standard.Connection, connection)
	}
}

func TestOptimizedHarNilReceiverMethods(t *testing.T) {
	var optHar *OptimizedHar

	if got := optHar.GetVersion(); got != "" {
		t.Errorf("GetVersion() = %q, want empty", got)
	}
	if got := optHar.GetCreator(); got != (Creator{}) {
		t.Errorf("GetCreator() = %#v, want zero Creator", got)
	}
	if entries := optHar.GetEntries(); entries != nil {
		t.Errorf("GetEntries() = %#v, want nil", entries)
	}
	if pages := optHar.GetPages(); pages != nil {
		t.Errorf("GetPages() = %#v, want nil", pages)
	}
	if standard := optHar.ToStandard(); standard != nil {
		t.Errorf("ToStandard() = %#v, want nil", standard)
	}
	if standard := optHar.ToStandardHar(); standard != nil {
		t.Errorf("ToStandardHar() = %#v, want nil", standard)
	}
	if results := optHar.SearchByURL("example.com"); results != nil {
		t.Errorf("SearchByURL() = %#v, want nil", results)
	}
	if results := optHar.SearchByMethod(MethodGET); results != nil {
		t.Errorf("SearchByMethod() = %#v, want nil", results)
	}
	if results := optHar.SearchByStatusCode(200); results != nil {
		t.Errorf("SearchByStatusCode() = %#v, want nil", results)
	}
}

func TestOptimizedProviderNilReceiverMethods(t *testing.T) {
	var entry *OptimizedEntries
	if !entry.GetStartedDateTime().IsZero() {
		t.Error("GetStartedDateTime() should return zero time for nil entry")
	}
	if entry.GetTime() != 0 {
		t.Errorf("GetTime() = %v, want 0", entry.GetTime())
	}
	if entry.GetRequest() != nil {
		t.Error("GetRequest() should return nil for nil entry")
	}
	if entry.GetResponse() != nil {
		t.Error("GetResponse() should return nil for nil entry")
	}
	if entry.GetTimings() != nil {
		t.Error("GetTimings() should return nil for nil entry")
	}
	if entry.GetPageref() != "" {
		t.Errorf("GetPageref() = %q, want empty", entry.GetPageref())
	}
	if standard := entry.ToStandard(); standard.Request.Method != "" {
		t.Errorf("ToStandard().Request.Method = %q, want empty", standard.Request.Method)
	}

	var request *OptimizedRequest
	if request.GetMethod() != "UNKNOWN" {
		t.Errorf("GetMethod() = %q, want UNKNOWN", request.GetMethod())
	}
	if request.GetURL() != "" || request.GetHTTPVersion() != "" {
		t.Errorf("request URL/version = %q/%q, want empty", request.GetURL(), request.GetHTTPVersion())
	}
	if request.GetHeaders() != nil || request.GetCookies() != nil || request.GetQueryString() != nil {
		t.Error("request collection getters should return nil for nil request")
	}
	if request.GetHeadersSize() != 0 || request.GetBodySize() != 0 {
		t.Errorf("request sizes = %d/%d, want 0/0", request.GetHeadersSize(), request.GetBodySize())
	}
	if request.GetPostData() != nil {
		t.Error("GetPostData() should return nil for nil request")
	}
	if value, ok := request.GetRequestHeaderValue("Accept"); ok || value != "" {
		t.Errorf("GetRequestHeaderValue() = %q, %v; want empty, false", value, ok)
	}
	if standard := request.ToStandard(); standard.Method != "" {
		t.Errorf("ToStandard().Method = %q, want empty", standard.Method)
	}

	var response *OptimizedResponse
	if response.GetStatus() != 0 || response.GetStatusText() != "" || response.GetHTTPVersion() != "" {
		t.Errorf("response status/text/version = %d/%q/%q, want zero", response.GetStatus(), response.GetStatusText(), response.GetHTTPVersion())
	}
	if response.GetHeaders() != nil || response.GetCookies() != nil {
		t.Error("response collection getters should return nil for nil response")
	}
	if response.GetContent() != nil {
		t.Error("GetContent() should return nil for nil response")
	}
	if response.GetHeadersSize() != 0 || response.GetBodySize() != 0 {
		t.Errorf("response sizes = %d/%d, want 0/0", response.GetHeadersSize(), response.GetBodySize())
	}
	if value, ok := response.GetResponseHeaderValue("Content-Type"); ok || value != "" {
		t.Errorf("GetResponseHeaderValue() = %q, %v; want empty, false", value, ok)
	}
	if standard := response.ToStandard(); standard.Status != 0 {
		t.Errorf("ToStandard().Status = %d, want 0", standard.Status)
	}

	var content *OptimizedContent
	if content.GetSize() != 0 || content.GetMimeType() != "" || content.GetText() != "" || content.GetEncoding() != "" {
		t.Error("content getters returned non-zero values")
	}
	if content.GetCompression() != 0 {
		t.Errorf("GetCompression() = %d, want 0", content.GetCompression())
	}
	if standard := content.ToStandard(); standard.Size != 0 || standard.MimeType != "" || standard.Text != "" {
		t.Errorf("ToStandard() = %#v, want zero content", standard)
	}

	var timings *OptimizedTimings
	if timings.GetBlocked() != -1 || timings.GetDNS() != -1 || timings.GetConnect() != -1 ||
		timings.GetSend() != -1 || timings.GetWait() != -1 || timings.GetReceive() != -1 ||
		timings.GetSSL() != -1 {
		t.Error("nil optimized timings getters should return -1")
	}
	if standard := timings.ToStandard(); standard.Blocked != -1 || standard.DNS != -1 ||
		standard.Connect != -1 || standard.Send != -1 || standard.Wait != -1 ||
		standard.Receive != -1 || standard.Ssl != -1 {
		t.Errorf("ToStandard() = %#v, want -1 timings", standard)
	}
}
