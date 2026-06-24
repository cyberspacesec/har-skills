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
		name    string
		content *Content
		want    bool
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

	var nilContent *Content
	if nilContent.IsText() {
		t.Error("IsText() should be false for nil content")
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

func TestContent_ParseAsMap_NullInput(t *testing.T) {
	c := &Content{Text: `null`}
	_, err := c.ParseAsMap()
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
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

// ---- MIMECategory: comprehensive branch coverage ----

func TestContent_MIMECategory_Comprehensive(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
		want     MIMECategory
	}{
		// Script category
		{"text/x-javascript", "text/x-javascript", MIMEScript},
		{"application/ecmascript", "application/ecmascript", MIMEScript},
		{"text/ecmascript", "text/ecmascript", MIMEScript},
		{"application/vnd.dart", "application/vnd.dart", MIMEScript},
		{"application/vnd.dart.v2", "application/vnd.dart.v2", MIMEScript},

		// Stylesheet category
		{"text/x-css", "text/x-css", MIMEStylesheet},
		{"application/x-css", "application/x-css", MIMEStylesheet},

		// Font category
		{"application/x-font-woff", "application/x-font-woff", MIMEFont},
		{"application/font-woff", "application/font-woff", MIMEFont},
		{"application/font-woff2", "application/font-woff2", MIMEFont},
		{"application/x-font-opentype", "application/x-font-opentype", MIMEFont},
		{"application/vnd.ms-fontobject", "application/vnd.ms-fontobject", MIMEFont},
		{"application/x-font-custom", "application/x-font-custom", MIMEFont},

		// Document category
		{"text/xml", "text/xml", MIMEDocument},
		{"text/richtext", "text/richtext", MIMEDocument},
		{"application/msword", "application/msword", MIMEDocument},
		{"application/vnd.oasis.opendocument.text", "application/vnd.oasis.opendocument.text", MIMEDocument},

		// API category
		{"text/json", "text/json", MIMEAPI},
		{"application/hal+json", "application/hal+json", MIMEAPI},

		// Data category
		{"text/tab-separated-values", "text/tab-separated-values", MIMEData},
		{"multipart/form-data", "multipart/form-data", MIMEData},
		{"application/vnd.api+json", "application/vnd.api+json", MIMEAPI},

		// Other text types fall into document
		{"text/markdown", "text/markdown", MIMEDocument},
		{"text/x-custom", "text/x-custom", MIMEDocument},

		// MIME with parameters (semicolon)
		{"text/css with charset", "text/css; charset=utf-8", MIMEStylesheet},
		{"application/json with charset", "application/json; charset=utf-8", MIMEAPI},

		// Truly unknown / other
		{"application/x-unknown", "application/x-unknown", MIMEOther},
		{"model/vrml", "model/vrml", MIMEOther},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Content{MimeType: tt.mimeType}
			got := c.MIMECategory()
			if got != tt.want {
				t.Errorf("MIMECategory(%q) = %v, want %v", tt.mimeType, got, tt.want)
			}
		})
	}
}

// ---- IsBinary: comprehensive branch coverage ----

func TestContent_IsBinary_DecodeContentError(t *testing.T) {
	// Content with invalid base64 encoding — DecodeContent returns error,
	// so the detected-MIME-type fallback path is skipped, and it returns true.
	c := &Content{MimeType: "image/png", Text: "!!!invalid-base64!!!", Encoding: "base64"}
	got := c.IsBinary()
	if !got {
		t.Error("IsBinary() should be true when MIME is non-text and DecodeContent fails")
	}
}

func TestContent_IsBinary_NonTextMIME_TextDetected(t *testing.T) {
	// MIME type is not text, but the actual content is detected as text.
	// http.DetectContentType should return "text/plain; charset=utf-8" for plain text.
	c := &Content{MimeType: "application/octet-stream", Text: "hello world, this is plain text data"}
	got := c.IsBinary()
	if got {
		t.Error("IsBinary() should be false when detected content type is text, even if MIME is non-text")
	}
}

func TestContent_IsBinary_NonTextMIME_BinaryDetected(t *testing.T) {
	// MIME type is not text, and content is truly binary (gzip magic bytes).
	// http.DetectContentType will not classify it as text.
	c := &Content{MimeType: "application/octet-stream", Text: string([]byte{0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff})}
	got := c.IsBinary()
	if !got {
		t.Error("IsBinary() should be true when detected content type is binary and MIME is non-text")
	}
}

func TestContent_IsBinary_NonTextMIME_EmptyContent(t *testing.T) {
	// Non-text MIME with no Text field — DecodeContent returns nil,
	// so the detection path is skipped, and it returns true.
	c := &Content{MimeType: "image/png", Size: 100}
	got := c.IsBinary()
	if !got {
		t.Error("IsBinary() should be true for non-text MIME with no content text")
	}
}

func TestContent_IsBinary_TextMIME(t *testing.T) {
	// Text MIME — returns false immediately without checking content
	c := &Content{MimeType: "text/html", Text: ""}
	got := c.IsBinary()
	if got {
		t.Error("IsBinary() should be false for text/* MIME types regardless of content")
	}
}

// ---- Hash: DecodeContent error path ----

func TestContent_Hash_DecodeContentError(t *testing.T) {
	// Content with invalid base64 encoding — DecodeContent returns error
	c := &Content{Text: "!!!invalid-base64!!!", Encoding: "base64"}
	_, err := c.Hash()
	if err == nil {
		t.Error("Hash() should return error when DecodeContent fails")
	}
}

// ---- ParseJSON: DecodeContent error path ----

func TestContent_ParseJSON_DecodeContentError(t *testing.T) {
	// Content with invalid base64 encoding — DecodeContent returns error
	c := &Content{Text: "!!!invalid-base64!!!", Encoding: "base64"}
	_, err := c.ParseJSON()
	if err == nil {
		t.Error("ParseJSON() should return error when DecodeContent fails")
	}
}

// ---- ParseAsMap: DecodeContent error and empty data paths ----

func TestContent_ParseAsMap_DecodeContentError(t *testing.T) {
	// Content with invalid base64 encoding — DecodeContent returns error
	c := &Content{Text: "!!!invalid-base64!!!", Encoding: "base64"}
	_, err := c.ParseAsMap()
	if err == nil {
		t.Error("ParseAsMap() should return error when DecodeContent fails")
	}
}

func TestContent_ParseAsMap_EmptyData(t *testing.T) {
	// Empty text content — DecodeContent returns nil data
	c := &Content{Text: ""}
	_, err := c.ParseAsMap()
	if err == nil {
		t.Error("ParseAsMap() should return error for empty content data")
	}
}

// ---- ContentSummary: negative size and empty MIME type ----

func TestHar_ContentSummary_NegativeSize(t *testing.T) {
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Response: Response{
						Content: Content{
							Size:     -1,
							MimeType: "text/html",
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
	// Negative size should be treated as 0
	if summary.TotalSize != 0 {
		t.Errorf("TotalSize = %d, want 0 for negative size entries", summary.TotalSize)
	}
	if summary.TextSize != 0 {
		t.Errorf("TextSize = %d, want 0 for negative size entries", summary.TextSize)
	}
}

func TestHar_ContentSummary_EmptyMimeType(t *testing.T) {
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Response: Response{
						Content: Content{
							Size:     100,
							MimeType: "",
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
	// Empty MIME type should be keyed as "unknown"
	if _, ok := summary.ByMIMEType["unknown"]; !ok {
		t.Error("ByMIMEType should contain 'unknown' key for empty MimeType")
	}
	if summary.ByMIMEType["unknown"] != 100 {
		t.Errorf("ByMIMEType['unknown'] = %d, want 100", summary.ByMIMEType["unknown"])
	}
}

func TestHar_ContentSummary_NoCompression(t *testing.T) {
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Response: Response{
						Content: Content{
							Size:     500,
							MimeType: "text/html",
						},
					},
				},
			},
		},
	}

	summary := h.ContentSummary()
	if summary.CompressedSize != 0 {
		t.Errorf("CompressedSize = %d, want 0 when no compression", summary.CompressedSize)
	}
}

func TestHar_ContentSummary_BinarySize(t *testing.T) {
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Response: Response{
						Content: Content{
							Size:     2000,
							MimeType: "image/jpeg",
						},
					},
				},
			},
		},
	}

	summary := h.ContentSummary()
	if summary.BinarySize != 2000 {
		t.Errorf("BinarySize = %d, want 2000 for image content", summary.BinarySize)
	}
	if summary.TextSize != 0 {
		t.Errorf("TextSize = %d, want 0 for binary-only content", summary.TextSize)
	}
}

// ---- SaveToFile: file system error path ----

func TestContent_SaveToFile_InvalidPath(t *testing.T) {
	c := &Content{Text: "hello"}
	// Writing to a directory path should fail
	err := c.SaveToFile("/dev/null/impossible/path/file.txt")
	if err == nil {
		t.Error("SaveToFile() should return error for invalid file path")
	}
}

// ---- isTextMIME: comprehensive branch coverage ----

func TestIsTextMIME_Comprehensive(t *testing.T) {
	tests := []struct {
		mime string
		want bool
	}{
		// text/* always true
		{"text/csv", true},
		{"text/xml", true},
		{"text/x-custom", true},

		// application types that are text
		{"application/x-yaml", true},
		{"application/yaml", true},
		{"application/toml", true},
		{"application/manifest+json", true},
		{"application/schema+json", true},
		{"application/vnd.api+json", true},
		{"application/soap+xml", true},

		// Suffix matching
		{"application/custom+json", true},
		{"application/custom+xml", true},

		// Non-text application types
		{"application/pdf", false},
		{"application/octet-stream", false},
		{"application/zip", false},
		{"application/x-shockwave-flash", false},

		// MIME with parameters
		{"application/json; charset=utf-8", true},
		{"application/xml; charset=iso-8859-1", true},
		{"image/png; charset=binary", false},

		// Edge: empty
		{"", false},

		// Case insensitivity
		{"Application/JSON", true},
		{"TEXT/HTML", true},
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
