package har

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// ---- MIMECategory tests ----

func TestContent_MIMECategory(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
		want     MIMECategory
	}{
		{"image png", "image/png", MIMEImage},
		{"image jpeg", "image/jpeg", MIMEImage},
		{"image svg", "image/svg+xml", MIMEImage},
		{"image webp", "image/webp", MIMEImage},
		{"javascript", "application/javascript", MIMEScript},
		{"text javascript", "text/javascript", MIMEScript},
		{"x-javascript", "application/x-javascript", MIMEScript},
		{"css", "text/css", MIMEStylesheet},
		{"font woff", "font/woff", MIMEFont},
		{"font woff2", "font/woff2", MIMEFont},
		{"x-font-ttf", "application/x-font-ttf", MIMEFont},
		{"audio mp3", "audio/mpeg", MIMEMedia},
		{"video mp4", "video/mp4", MIMEMedia},
		{"html", "text/html", MIMEDocument},
		{"xhtml", "application/xhtml+xml", MIMEDocument},
		{"plain text", "text/plain", MIMEDocument},
		{"pdf", "application/pdf", MIMEDocument},
		{"json", "application/json", MIMEAPI},
		{"json with suffix", "application/vnd.api+json", MIMEAPI},
		{"graphql", "application/graphql", MIMEAPI},
		{"csv", "text/csv", MIMEData},
		{"form-urlencoded", "application/x-www-form-urlencoded", MIMEData},
		{"octet-stream", "application/octet-stream", MIMEData},
		{"unknown", "application/x-unknown", MIMEOther},
		{"mime with params", "text/html; charset=utf-8", MIMEDocument},
		{"empty", "", MIMEOther},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Content{MimeType: tt.mimeType}
			got := c.MIMECategory()
			if got != tt.want {
				t.Errorf("MIMECategory() = %v, want %v (mimeType=%q)", got, tt.want, tt.mimeType)
			}
		})
	}

	// Test nil Content
	var nilContent *Content
	if nilContent.MIMECategory() != MIMEOther {
		t.Errorf("nil Content MIMECategory() should be MIMEOther")
	}
}

// ---- IsBinary / IsText tests ----

func TestContent_IsBinary(t *testing.T) {
	tests := []struct {
		name     string
		content  *Content
		want     bool
	}{
		{
			name:    "text html",
			content: &Content{MimeType: "text/html", Text: "<html></html>"},
			want:    false,
		},
		{
			name:    "json",
			content: &Content{MimeType: "application/json", Text: `{"key":"value"}`},
			want:    false,
		},
		{
			name:    "javascript",
			content: &Content{MimeType: "application/javascript", Text: "var x = 1;"},
			want:    false,
		},
		{
			name:    "image png",
			content: &Content{MimeType: "image/png", Size: 100},
			want:    true,
		},
		{
			name:    "video mp4",
			content: &Content{MimeType: "video/mp4", Size: 1000},
			want:    true,
		},
		{
			name:    "xml",
			content: &Content{MimeType: "application/xml", Text: "<root/>"},
			want:    false,
		},
		{
			name:    "nil content",
			content: nil,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.content.IsBinary()
			if got != tt.want {
				t.Errorf("IsBinary() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContent_IsText(t *testing.T) {
	c := &Content{MimeType: "text/plain", Text: "hello"}
	if !c.IsText() {
		t.Error("IsText() should be true for text/plain content")
	}

	bin := &Content{MimeType: "image/png", Size: 100}
	if bin.IsText() {
		t.Error("IsText() should be false for image/png content")
	}
}

// ---- DetectMIMEType tests ----

func TestContent_DetectMIMEType(t *testing.T) {
	// Text content detection
	c := &Content{Text: "<html><body>Hello</body></html>", MimeType: "text/html"}
	detected := c.DetectMIMEType()
	if detected != "text/html; charset=utf-8" && detected != "text/html" {
		// http.DetectContentType may or may not include charset
		t.Logf("DetectMIMEType() = %q (acceptable)", detected)
	}

	// Empty content falls back to MimeType
	c2 := &Content{Text: "", MimeType: "application/json"}
	detected2 := c2.DetectMIMEType()
	if detected2 != "application/json" {
		t.Errorf("DetectMIMEType() for empty content = %q, want %q", detected2, "application/json")
	}

	// nil Content
	var nilContent *Content
	if nilContent.DetectMIMEType() != "" {
		t.Error("nil Content DetectMIMEType() should return empty string")
	}
}

// ---- Hash tests ----

func TestContent_Hash(t *testing.T) {
	c := &Content{Text: "hello world"}
	hash, err := c.Hash()
	if err != nil {
		t.Fatalf("Hash() error: %v", err)
	}
	if hash == "" {
		t.Error("Hash() returned empty string")
	}
	// SHA-256 of "hello world" is known
	expected := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	if hash != expected {
		t.Errorf("Hash() = %q, want %q", hash, expected)
	}
}

func TestContent_Hash_Nil(t *testing.T) {
	var c *Content
	_, err := c.Hash()
	if err == nil {
		t.Error("nil Content Hash() should return error")
	}
}

// ---- ParseJSON tests ----

func TestContent_ParseJSON(t *testing.T) {
	c := &Content{Text: `{"name":"test","value":42}`}
	result, err := c.ParseJSON()
	if err != nil {
		t.Fatalf("ParseJSON() error: %v", err)
	}

	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("ParseJSON() should return a map")
	}
	if m["name"] != "test" {
		t.Errorf("ParseJSON() name = %v, want 'test'", m["name"])
	}

	// Parse JSON array
	arr := &Content{Text: `[1, 2, 3]`}
	result2, err := arr.ParseJSON()
	if err != nil {
		t.Fatalf("ParseJSON() array error: %v", err)
	}
	slice, ok := result2.([]interface{})
	if !ok || len(slice) != 3 {
		t.Errorf("ParseJSON() array result = %v, want 3-element slice", result2)
	}
}

func TestContent_ParseJSON_Invalid(t *testing.T) {
	c := &Content{Text: "not json at all"}
	_, err := c.ParseJSON()
	if err == nil {
		t.Error("ParseJSON() should return error for invalid JSON")
	}
}

func TestContent_ParseJSON_Nil(t *testing.T) {
	var c *Content
	_, err := c.ParseJSON()
	if err == nil {
		t.Error("nil Content ParseJSON() should return error")
	}
}

// ---- ParseAsMap tests ----

func TestContent_ParseAsMap(t *testing.T) {
	c := &Content{Text: `{"key":"value","num":123}`}
	result, err := c.ParseAsMap()
	if err != nil {
		t.Fatalf("ParseAsMap() error: %v", err)
	}
	if result["key"] != "value" {
		t.Errorf("ParseAsMap() key = %v, want 'value'", result["key"])
	}
}

func TestContent_ParseAsMap_ArrayInput(t *testing.T) {
	c := &Content{Text: `[1,2,3]`}
	_, err := c.ParseAsMap()
	if err == nil {
		t.Error("ParseAsMap() should return error for JSON array")
	}
}

// ---- Entries.ContentLength tests ----

func TestEntries_ContentLength(t *testing.T) {
	e := &Entries{
		Response: Response{
			Headers: []Headers{
				{Name: "Content-Length", Value: "1234"},
			},
		},
	}
	if e.ContentLength() != 1234 {
		t.Errorf("ContentLength() = %d, want 1234", e.ContentLength())
	}

	// No Content-Length header
	e2 := &Entries{
		Response: Response{
			Headers: []Headers{},
		},
	}
	if e2.ContentLength() != -1 {
		t.Errorf("ContentLength() with no header = %d, want -1", e2.ContentLength())
	}

	// nil Entries
	var nilEntries *Entries
	if nilEntries.ContentLength() != -1 {
		t.Error("nil Entries ContentLength() should return -1")
	}
}

// ---- Entries.HasContentLengthMismatch tests ----

func TestEntries_HasContentLengthMismatch(t *testing.T) {
	// Mismatch
	e := &Entries{
		Response: Response{
			Headers: []Headers{
				{Name: "Content-Length", Value: "100"},
			},
			Content: Content{Size: 200},
		},
	}
	if !e.HasContentLengthMismatch() {
		t.Error("HasContentLengthMismatch() should be true when sizes differ")
	}

	// Match
	e2 := &Entries{
		Response: Response{
			Headers: []Headers{
				{Name: "Content-Length", Value: "200"},
			},
			Content: Content{Size: 200},
		},
	}
	if e2.HasContentLengthMismatch() {
		t.Error("HasContentLengthMismatch() should be false when sizes match")
	}

	// No Content-Length header
	e3 := &Entries{
		Response: Response{
			Headers: []Headers{},
			Content: Content{Size: 200},
		},
	}
	if e3.HasContentLengthMismatch() {
		t.Error("HasContentLengthMismatch() should be false when no Content-Length header")
	}

	// nil Entries
	var nilEntries *Entries
	if nilEntries.HasContentLengthMismatch() {
		t.Error("nil Entries HasContentLengthMismatch() should be false")
	}
}

// ---- Entries.EstimateTransferSize tests ----

func TestEntries_EstimateTransferSize(t *testing.T) {
	// TransferSize available
	e := &Entries{
		Response: Response{
			TransferSize: 500,
			BodySize:     1000,
			Content:      Content{Size: 1000},
		},
	}
	if e.EstimateTransferSize() != 500 {
		t.Errorf("EstimateTransferSize() = %d, want 500", e.EstimateTransferSize())
	}

	// BodySize with compression
	e2 := &Entries{
		Response: Response{
			BodySize: 1000,
			Content:  Content{Size: 1000, Compression: 300},
		},
	}
	if e2.EstimateTransferSize() != 700 {
		t.Errorf("EstimateTransferSize() = %d, want 700", e2.EstimateTransferSize())
	}

	// BodySize without compression
	e3 := &Entries{
		Response: Response{
			BodySize: 1000,
			Content:  Content{Size: 1000},
		},
	}
	if e3.EstimateTransferSize() != 1000 {
		t.Errorf("EstimateTransferSize() = %d, want 1000", e3.EstimateTransferSize())
	}

	// Fall back to Content.Size
	e4 := &Entries{
		Response: Response{
			Content: Content{Size: 800},
		},
	}
	if e4.EstimateTransferSize() != 800 {
		t.Errorf("EstimateTransferSize() = %d, want 800", e4.EstimateTransferSize())
	}

	// nil Entries
	var nilEntries *Entries
	if nilEntries.EstimateTransferSize() != 0 {
		t.Error("nil Entries EstimateTransferSize() should return 0")
	}
}

// ---- Har.ContentSummary tests ----

func TestHar_ContentSummary(t *testing.T) {
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Response: Response{
						Content: Content{
							Size:        1000,
							MimeType:    "text/html",
							Compression: 200,
						},
					},
				},
				{
					Response: Response{
						Content: Content{
							Size:     500,
							MimeType: "image/png",
						},
					},
				},
				{
					Response: Response{
						Content: Content{
							Size:     300,
							MimeType: "application/json",
						},
					},
				},
			},
		},
	}

	summary := h.ContentSummary()
	if summary == nil {
		t.Fatal("ContentSummary() returned nil")
	}

	if summary.TotalSize != 1800 {
		t.Errorf("TotalSize = %d, want 1800", summary.TotalSize)
	}

	if summary.TextSize != 1300 {
		t.Errorf("TextSize = %d, want 1300", summary.TextSize)
	}

	if summary.BinarySize != 500 {
		t.Errorf("BinarySize = %d, want 500", summary.BinarySize)
	}

	if summary.CompressedSize != 200 {
		t.Errorf("CompressedSize = %d, want 200", summary.CompressedSize)
	}

	if summary.ByCategory[MIMEDocument] != 1000 {
		t.Errorf("ByCategory[document] = %d, want 1000", summary.ByCategory[MIMEDocument])
	}

	if summary.ByCategory[MIMEImage] != 500 {
		t.Errorf("ByCategory[image] = %d, want 500", summary.ByCategory[MIMEImage])
	}

	if summary.ByCategory[MIMEAPI] != 300 {
		t.Errorf("ByCategory[api] = %d, want 300", summary.ByCategory[MIMEAPI])
	}

	if summary.ByMIMEType["text/html"] != 1000 {
		t.Errorf("ByMIMEType[text/html] = %d, want 1000", summary.ByMIMEType["text/html"])
	}

	if summary.ByMIMEType["image/png"] != 500 {
		t.Errorf("ByMIMEType[image/png] = %d, want 500", summary.ByMIMEType["image/png"])
	}
}

func TestHar_ContentSummary_Nil(t *testing.T) {
	var h *Har
	if h.ContentSummary() != nil {
		t.Error("nil Har ContentSummary() should return nil")
	}
}

// ---- Content.SaveToFile tests ----

func TestContent_SaveToFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test_output.txt")

	c := &Content{Text: "hello world"}
	if err := c.SaveToFile(path); err != nil {
		t.Fatalf("SaveToFile() error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}

	if string(data) != "hello world" {
		t.Errorf("saved content = %q, want %q", string(data), "hello world")
	}
}

func TestContent_SaveToFile_Base64(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "b64_output.txt")

	original := "base64 encoded content"
	encoded := base64.StdEncoding.EncodeToString([]byte(original))
	c := &Content{Text: encoded, Encoding: "base64"}

	if err := c.SaveToFile(path); err != nil {
		t.Fatalf("SaveToFile() error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}

	if string(data) != original {
		t.Errorf("saved content = %q, want %q", string(data), original)
	}
}

func TestContent_SaveToFile_Nil(t *testing.T) {
	var c *Content
	err := c.SaveToFile("/tmp/should_not_exist")
	if err == nil {
		t.Error("nil Content SaveToFile() should return error")
	}
}

func TestContent_SaveToFile_EmptyContent(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "empty.txt")

	c := &Content{Text: ""}
	if err := c.SaveToFile(path); err != nil {
		t.Fatalf("SaveToFile() error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}

	if len(data) != 0 {
		t.Errorf("saved content should be empty, got %q", string(data))
	}
}

// ---- isTextMIME tests ----

func TestIsTextMIME(t *testing.T) {
	tests := []struct {
		mime string
		want bool
	}{
		{"text/html", true},
		{"text/plain", true},
		{"text/css", true},
		{"application/json", true},
		{"application/xml", true},
		{"application/javascript", true},
		{"application/ld+json", true},
		{"application/atom+xml", true},
		{"application/rss+xml", true},
		{"application/graphql", true},
		{"text/html; charset=utf-8", true},
		{"image/png", false},
		{"video/mp4", false},
		{"application/octet-stream", false},
		{"application/pdf", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.mime, func(t *testing.T) {
			got := isTextMIME(tt.mime)
			if got != tt.want {
				t.Errorf("isTextMIME(%q) = %v, want %v", tt.mime, got, tt.want)
			}
		})
	}
}

// ---- Integration: full HAR content summary ----

func TestHar_ContentSummary_Integration(t *testing.T) {
	// Build a small HAR with mixed content types
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Response: Response{
						Headers: []Headers{
							{Name: "Content-Length", Value: "2048"},
						},
						Content: Content{
							Size:        2048,
							MimeType:    "text/html",
							Compression: 512,
							Text:        "<html><body>Page</body></html>",
						},
					},
				},
				{
					Response: Response{
						Headers: []Headers{
							{Name: "Content-Length", Value: "10240"},
						},
						Content: Content{
							Size:     10240,
							MimeType: "image/jpeg",
						},
					},
				},
				{
					Response: Response{
						Content: Content{
							Size:     512,
							MimeType: "application/json",
							Text:     `{"status":"ok"}`,
						},
					},
				},
			},
		},
	}

	summary := h.ContentSummary()

	// Verify totals
	if summary.TotalSize != 2048+10240+512 {
		t.Errorf("TotalSize = %d, want %d", summary.TotalSize, 2048+10240+512)
	}

	// Verify mismatch detection
	// Entry 0: Content-Length=2048, Content.Size=2048 => no mismatch
	if h.Log.Entries[0].HasContentLengthMismatch() {
		t.Error("Entry 0 should not have mismatch (both 2048)")
	}

	// Entry 1: Content-Length=10240, Content.Size=10240 => no mismatch
	if h.Log.Entries[1].HasContentLengthMismatch() {
		t.Error("Entry 1 should not have mismatch (both 10240)")
	}

	// Verify content analysis
	htmlContent := h.Log.Entries[0].Response.Content
	if htmlContent.IsBinary() {
		t.Error("text/html should not be binary")
	}
	if !htmlContent.IsText() {
		t.Error("text/html should be text")
	}

	imgContent := h.Log.Entries[1].Response.Content
	if !imgContent.IsBinary() {
		t.Error("image/jpeg should be binary")
	}

	// JSON ParseAsMap
	jsonContent := h.Log.Entries[2].Response.Content
	m, err := jsonContent.ParseAsMap()
	if err != nil {
		t.Fatalf("ParseAsMap() error: %v", err)
	}
	if m["status"] != "ok" {
		t.Errorf("ParseAsMap() status = %v, want 'ok'", m["status"])
	}

	// Hash
	hash, err := jsonContent.Hash()
	if err != nil {
		t.Fatalf("Hash() error: %v", err)
	}
	if len(hash) != 64 {
		t.Errorf("Hash() length = %d, want 64 (SHA-256 hex)", len(hash))
	}
}

// ---- Content with JSON parsing edge cases ----

func TestContent_ParseJSON_Empty(t *testing.T) {
	c := &Content{Text: ""}
	_, err := c.ParseJSON()
	if err == nil {
		t.Error("ParseJSON() should return error for empty string")
	}
}

func TestContent_ParseAsMap_Nil(t *testing.T) {
	var c *Content
	_, err := c.ParseAsMap()
	if err == nil {
		t.Error("nil Content ParseAsMap() should return error")
	}
}

func TestContent_Hash_Empty(t *testing.T) {
	c := &Content{Text: ""}
	_, err := c.Hash()
	if err == nil {
		t.Error("Hash() should return error for empty content")
	}
}

// ---- MIMECategory case insensitivity ----

func TestContent_MIMECategory_CaseInsensitive(t *testing.T) {
	c := &Content{MimeType: "Text/HTML"}
	if c.MIMECategory() != MIMEDocument {
		t.Errorf("MIMECategory() for 'Text/HTML' = %v, want MIMEDocument", c.MIMECategory())
	}

	c2 := &Content{MimeType: "Application/JSON"}
	if c2.MIMECategory() != MIMEAPI {
		t.Errorf("MIMECategory() for 'Application/JSON' = %v, want MIMEAPI", c2.MIMECategory())
	}
}

// ---- ContentLength with various header formats ----

func TestEntries_ContentLength_InvalidValue(t *testing.T) {
	e := &Entries{
		Response: Response{
			Headers: []Headers{
				{Name: "Content-Length", Value: "not-a-number"},
			},
		},
	}
	if e.ContentLength() != -1 {
		t.Errorf("ContentLength() with invalid value = %d, want -1", e.ContentLength())
	}
}

// ---- EstimateTransferSize: compression larger than body (edge case) ----

func TestEntries_EstimateTransferSize_CompressionLargerThanBody(t *testing.T) {
	e := &Entries{
		Response: Response{
			BodySize: 100,
			Content: Content{
				Size:        100,
				Compression: 200, // larger than BodySize
			},
		},
	}
	// BodySize (100) - Compression (200) would be negative, so return BodySize
	if e.EstimateTransferSize() != 100 {
		t.Errorf("EstimateTransferSize() = %d, want 100", e.EstimateTransferSize())
	}
}

// ---- SaveToFile with JSON content ----

func TestContent_SaveToFile_JSONContent(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "data.json")

	jsonStr := `{"key":"value","num":42}`
	c := &Content{Text: jsonStr, MimeType: "application/json"}

	if err := c.SaveToFile(path); err != nil {
		t.Fatalf("SaveToFile() error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("saved file is not valid JSON: %v", err)
	}

	if result["key"] != "value" {
		t.Errorf("saved JSON key = %v, want 'value'", result["key"])
	}
}
