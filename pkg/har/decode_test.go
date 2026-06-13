package har

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/base64"
	"strings"
	"testing"
)

func TestDecodeContentPlainText(t *testing.T) {
	content := &Content{
		Size:     12,
		MimeType: "text/plain",
		Text:     "Hello World!",
	}

	data, err := content.DecodeContent()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if string(data) != "Hello World!" {
		t.Errorf("Expected 'Hello World!', got '%s'", string(data))
	}
}

func TestDecodeContentBase64(t *testing.T) {
	original := "Hello World!"
	encoded := base64.StdEncoding.EncodeToString([]byte(original))

	content := &Content{
		Size:     len(encoded),
		MimeType: "text/plain",
		Text:     encoded,
		Encoding: "base64",
	}

	data, err := content.DecodeContent()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if string(data) != original {
		t.Errorf("Expected '%s', got '%s'", original, string(data))
	}
}

func TestDecodeContentNil(t *testing.T) {
	var content *Content
	_, err := content.DecodeContent()
	if err == nil {
		t.Error("Expected error for nil content")
	}
}

func TestDecodeContentEmpty(t *testing.T) {
	content := &Content{
		Size:     0,
		MimeType: "text/plain",
		Text:     "",
	}

	data, err := content.DecodeContent()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if data != nil {
		t.Errorf("Expected nil for empty content, got %v", data)
	}
}

func TestDecodeEntryContent(t *testing.T) {
	entry := &Entries{
		Response: Response{
			Content: Content{
				Size:     5,
				MimeType: "text/plain",
				Text:     "hello",
			},
		},
	}

	data, err := entry.DecodeContent()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if string(data) != "hello" {
		t.Errorf("Expected 'hello', got '%s'", string(data))
	}
}

func TestDecodeAllContent(t *testing.T) {
	h := NewHar()

	e1 := h.AddEntry("GET", "https://example.com/1", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e1.SetResponseContentText("response 1")

	e2 := h.AddEntry("GET", "https://example.com/2", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")
	e2.SetResponseContentText("response 2")

	results, err := h.DecodeAllContent()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	if string(results[0]) != "response 1" {
		t.Errorf("Expected 'response 1', got '%s'", string(results[0]))
	}

	if string(results[1]) != "response 2" {
		t.Errorf("Expected 'response 2', got '%s'", string(results[1]))
	}
}

func TestIsBase64Encoded(t *testing.T) {
	content := &Content{
		Encoding: "base64",
		Text:     "SGVsbG8=",
	}

	if !content.IsBase64Encoded() {
		t.Error("Expected content to be base64 encoded")
	}

	content2 := &Content{
		Text: "plain text",
	}

	if content2.IsBase64Encoded() {
		t.Error("Expected content not to be base64 encoded")
	}
}

func TestIsCompressed(t *testing.T) {
	tests := []struct {
		name     string
		entry    *Entries
		expected bool
	}{
		{
			name: "gzip encoding",
			entry: &Entries{
				Response: Response{
					Headers: []Headers{
						{Name: "Content-Encoding", Value: "gzip"},
					},
				},
			},
			expected: true,
		},
		{
			name: "deflate encoding",
			entry: &Entries{
				Response: Response{
					Headers: []Headers{
						{Name: "Content-Encoding", Value: "deflate"},
					},
				},
			},
			expected: true,
		},
		{
			name: "br (brotli) encoding",
			entry: &Entries{
				Response: Response{
					Headers: []Headers{
						{Name: "Content-Encoding", Value: "br"},
					},
				},
			},
			expected: true,
		},
		{
			name: "zstd encoding",
			entry: &Entries{
				Response: Response{
					Headers: []Headers{
						{Name: "Content-Encoding", Value: "zstd"},
					},
				},
			},
			expected: true,
		},
		{
			name: "no content-encoding",
			entry: &Entries{
				Response: Response{
					Headers: []Headers{
						{Name: "Content-Type", Value: "text/html"},
					},
				},
			},
			expected: false,
		},
		{
			name: "identity encoding (not compressed)",
			entry: &Entries{
				Response: Response{
					Headers: []Headers{
						{Name: "Content-Encoding", Value: "identity"},
					},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.entry.IsCompressed() != tt.expected {
				t.Errorf("IsCompressed() = %v, expected %v", tt.entry.IsCompressed(), tt.expected)
			}
		})
	}
}

func TestGetContentEncoding(t *testing.T) {
	entry := &Entries{
		Response: Response{
			Headers: []Headers{
				{Name: "Content-Encoding", Value: "gzip"},
			},
		},
	}

	if encoding := entry.GetContentEncoding(); encoding != "gzip" {
		t.Errorf("Expected 'gzip', got '%s'", encoding)
	}
}

func TestDecodeEntryText(t *testing.T) {
	entry := &Entries{
		Response: Response{
			Content: Content{
				Text: "hello world",
			},
		},
	}

	text, err := entry.DecodeEntryText()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if text != "hello world" {
		t.Errorf("Expected 'hello world', got '%s'", text)
	}
}

func TestDecodeGzipContent(t *testing.T) {
	// Create gzip compressed data
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	_, err := writer.Write([]byte("compressed content"))
	if err != nil {
		t.Fatalf("Failed to create gzip data: %v", err)
	}
	writer.Close()

	// Encode as base64
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())

	content := &Content{
		Size:     len(buf.Bytes()),
		MimeType: "text/plain",
		Text:     encoded,
		Encoding: "base64",
	}

	data, err := content.DecodeContent()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if string(data) != "compressed content" {
		t.Errorf("Expected 'compressed content', got '%s'", string(data))
	}
}

func TestIsGzipData(t *testing.T) {
	tests := []struct {
		data     []byte
		expected bool
	}{
		{[]byte{0x1f, 0x8b, 0x08, 0x00}, true},
		{[]byte{0x00, 0x01, 0x02, 0x03}, false},
		{[]byte{0x1f}, false},
		{[]byte{}, false},
	}

	for _, tt := range tests {
		result := isGzipData(tt.data)
		if result != tt.expected {
			t.Errorf("isGzipData(%v) = %v, expected %v", tt.data, result, tt.expected)
		}
	}
}

func TestIsDeflateData(t *testing.T) {
	tests := []struct {
		data     []byte
		expected bool
	}{
		{[]byte{0x78, 0x9c, 0x01, 0x00}, true},  // default compression
		{[]byte{0x78, 0x01, 0x01, 0x00}, true},  // no compression
		{[]byte{0x78, 0x5e, 0x01, 0x00}, true},  // best speed
		{[]byte{0x78, 0xda, 0x01, 0x00}, true},  // best compression
		{[]byte{0x00, 0x01, 0x02, 0x03}, false}, // not deflate
		{[]byte{0x78}, false},                    // too short
		{[]byte{}, false},                        // empty
	}

	for _, tt := range tests {
		result := isDeflateData(tt.data)
		if result != tt.expected {
			t.Errorf("isDeflateData(%v) = %v, expected %v", tt.data, result, tt.expected)
		}
	}
}

// Test for base64 URL-safe encoding
func TestDecodeContentBase64URLSafe(t *testing.T) {
	original := "Hello World!"
	encoded := base64.URLEncoding.EncodeToString([]byte(original))

	content := &Content{
		Size:     len(encoded),
		MimeType: "text/plain",
		Text:     encoded,
		Encoding: "base64",
	}

	data, err := content.DecodeContent()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(string(data), original) {
		t.Errorf("Expected decoded content to contain '%s', got '%s'", original, string(data))
	}
}

// --- New tests for enhanced decode functionality ---

func TestDecompressByEncodingGzip(t *testing.T) {
	original := []byte("hello gzip world")

	// Compress with gzip
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	if _, err := writer.Write(original); err != nil {
		t.Fatalf("Failed to compress: %v", err)
	}
	writer.Close()

	decompressed, err := DecompressByEncoding(buf.Bytes(), "gzip")
	if err != nil {
		t.Fatalf("DecompressByEncoding gzip failed: %v", err)
	}

	if string(decompressed) != string(original) {
		t.Errorf("Expected '%s', got '%s'", string(original), string(decompressed))
	}
}

func TestDecompressByEncodingDeflate(t *testing.T) {
	original := []byte("hello deflate world")

	// Compress with zlib (deflate wrapped in zlib format)
	var buf bytes.Buffer
	writer := zlib.NewWriter(&buf)
	if _, err := writer.Write(original); err != nil {
		t.Fatalf("Failed to compress: %v", err)
	}
	writer.Close()

	decompressed, err := DecompressByEncoding(buf.Bytes(), "deflate")
	if err != nil {
		t.Fatalf("DecompressByEncoding deflate failed: %v", err)
	}

	if string(decompressed) != string(original) {
		t.Errorf("Expected '%s', got '%s'", string(original), string(decompressed))
	}
}

func TestDecompressByEncodingBrotli(t *testing.T) {
	data := []byte("some data")

	_, err := DecompressByEncoding(data, "br")
	if err == nil {
		t.Error("Expected error for brotli encoding, got nil")
	}

	harErr, ok := err.(*HarError)
	if !ok {
		t.Fatalf("Expected *HarError, got %T", err)
	}

	if harErr.Code != ErrCodeUnsupported {
		t.Errorf("Expected ErrCodeUnsupported, got %d", harErr.Code)
	}

	if !strings.Contains(harErr.Message, "brotli") || !strings.Contains(harErr.Message, "标准库") {
		t.Errorf("Error message should mention brotli and standard library, got: %s", harErr.Message)
	}
}

func TestDecompressByEncodingZstd(t *testing.T) {
	data := []byte("some data")

	_, err := DecompressByEncoding(data, "zstd")
	if err == nil {
		t.Error("Expected error for zstd encoding, got nil")
	}

	harErr, ok := err.(*HarError)
	if !ok {
		t.Fatalf("Expected *HarError, got %T", err)
	}

	if harErr.Code != ErrCodeUnsupported {
		t.Errorf("Expected ErrCodeUnsupported, got %d", harErr.Code)
	}

	if !strings.Contains(harErr.Message, "zstd") || !strings.Contains(harErr.Message, "标准库") {
		t.Errorf("Error message should mention zstd and standard library, got: %s", harErr.Message)
	}
}

func TestDecompressByEncodingUnknown(t *testing.T) {
	data := []byte("some data")

	_, err := DecompressByEncoding(data, "unknown-encoding")
	if err == nil {
		t.Error("Expected error for unknown encoding, got nil")
	}

	harErr, ok := err.(*HarError)
	if !ok {
		t.Fatalf("Expected *HarError, got %T", err)
	}

	if harErr.Code != ErrCodeUnsupported {
		t.Errorf("Expected ErrCodeUnsupported, got %d", harErr.Code)
	}
}

func TestDecompressByEncodingIdentity(t *testing.T) {
	data := []byte("identity data")

	result, err := DecompressByEncoding(data, "identity")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if string(result) != string(data) {
		t.Errorf("Expected '%s', got '%s'", string(data), string(result))
	}
}

func TestDecompressByEncodingEmpty(t *testing.T) {
	result, err := DecompressByEncoding([]byte{}, "gzip")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected empty result, got %d bytes", len(result))
	}
}

func TestDecompressByEncodingEmptyEncoding(t *testing.T) {
	data := []byte("plain data")

	result, err := DecompressByEncoding(data, "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if string(result) != string(data) {
		t.Errorf("Expected '%s', got '%s'", string(data), string(result))
	}
}

func TestDecompressByEncodingInvalidGzipData(t *testing.T) {
	// Not valid gzip data but claiming to be gzip
	data := []byte("this is not gzip data")

	_, err := DecompressByEncoding(data, "gzip")
	if err == nil {
		t.Error("Expected error for invalid gzip data, got nil")
	}
}

func TestDecompressByEncodingMultiEncoding(t *testing.T) {
	data := []byte("some data")

	_, err := DecompressByEncoding(data, "gzip, deflate")
	if err == nil {
		t.Error("Expected error for multi-encoding, got nil")
	}

	harErr, ok := err.(*HarError)
	if !ok {
		t.Fatalf("Expected *HarError, got %T", err)
	}

	if harErr.Code != ErrCodeUnsupported {
		t.Errorf("Expected ErrCodeUnsupported, got %d", harErr.Code)
	}

	if !strings.Contains(harErr.Message, "多重编码") {
		t.Errorf("Error message should mention multi-encoding, got: %s", harErr.Message)
	}
}

func TestDecompressWithEncoding(t *testing.T) {
	original := []byte("test with encoding header")

	// Compress with gzip
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	if _, err := writer.Write(original); err != nil {
		t.Fatalf("Failed to compress: %v", err)
	}
	writer.Close()

	decompressed, err := DecompressWithEncoding(buf.Bytes(), "gzip")
	if err != nil {
		t.Fatalf("DecompressWithEncoding failed: %v", err)
	}

	if string(decompressed) != string(original) {
		t.Errorf("Expected '%s', got '%s'", string(original), string(decompressed))
	}
}

func TestCompressContentGzip(t *testing.T) {
	original := []byte("compress me with gzip")

	compressed, err := CompressContent(original, "gzip")
	if err != nil {
		t.Fatalf("CompressContent gzip failed: %v", err)
	}

	// Verify it's actually gzip data
	if !isGzipData(compressed) {
		t.Error("Compressed data should have gzip magic bytes")
	}

	// Decompress and verify
	reader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	decompressed, err := readAll(reader)
	if err != nil {
		t.Fatalf("Failed to decompress: %v", err)
	}

	if string(decompressed) != string(original) {
		t.Errorf("Expected '%s', got '%s'", string(original), string(decompressed))
	}
}

func TestCompressContentDeflate(t *testing.T) {
	original := []byte("compress me with deflate")

	compressed, err := CompressContent(original, "deflate")
	if err != nil {
		t.Fatalf("CompressContent deflate failed: %v", err)
	}

	// Verify it's actually deflate data
	if !isDeflateData(compressed) {
		t.Error("Compressed data should have deflate/zlib header bytes")
	}

	// Decompress and verify
	reader, err := zlib.NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatalf("Failed to create zlib reader: %v", err)
	}
	defer reader.Close()

	decompressed, err := readAll(reader)
	if err != nil {
		t.Fatalf("Failed to decompress: %v", err)
	}

	if string(decompressed) != string(original) {
		t.Errorf("Expected '%s', got '%s'", string(original), string(decompressed))
	}
}

func TestCompressContentBrotli(t *testing.T) {
	data := []byte("some data")

	_, err := CompressContent(data, "br")
	if err == nil {
		t.Error("Expected error for brotli compression, got nil")
	}

	harErr, ok := err.(*HarError)
	if !ok {
		t.Fatalf("Expected *HarError, got %T", err)
	}

	if harErr.Code != ErrCodeUnsupported {
		t.Errorf("Expected ErrCodeUnsupported, got %d", harErr.Code)
	}
}

func TestCompressContentZstd(t *testing.T) {
	data := []byte("some data")

	_, err := CompressContent(data, "zstd")
	if err == nil {
		t.Error("Expected error for zstd compression, got nil")
	}

	harErr, ok := err.(*HarError)
	if !ok {
		t.Fatalf("Expected *HarError, got %T", err)
	}

	if harErr.Code != ErrCodeUnsupported {
		t.Errorf("Expected ErrCodeUnsupported, got %d", harErr.Code)
	}
}

func TestCompressContentUnknown(t *testing.T) {
	data := []byte("some data")

	_, err := CompressContent(data, "unknown")
	if err == nil {
		t.Error("Expected error for unknown compression, got nil")
	}

	harErr, ok := err.(*HarError)
	if !ok {
		t.Fatalf("Expected *HarError, got %T", err)
	}

	if harErr.Code != ErrCodeUnsupported {
		t.Errorf("Expected ErrCodeUnsupported, got %d", harErr.Code)
	}
}

func TestCompressContentEmpty(t *testing.T) {
	result, err := CompressContent([]byte{}, "gzip")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected empty result, got %d bytes", len(result))
	}
}

func TestDecompressIfNeededErrorOnCorruptGzip(t *testing.T) {
	// Create a byte slice that starts with gzip magic but is corrupt
	corruptGzip := []byte{0x1f, 0x8b, 0x08, 0x00, 0xFF, 0xFF, 0xFF, 0xFF}

	_, err := decompressIfNeeded(corruptGzip, "text/plain")
	if err == nil {
		t.Error("Expected error for corrupt gzip data, got nil")
	}

	harErr, ok := err.(*HarError)
	if !ok {
		t.Fatalf("Expected *HarError, got %T", err)
	}

	if harErr.Code != ErrCodeInvalidFormat {
		t.Errorf("Expected ErrCodeInvalidFormat, got %d", harErr.Code)
	}
}

func TestDecompressIfNeededErrorOnCorruptDeflate(t *testing.T) {
	// Create a byte slice that starts with zlib header but is corrupt
	corruptDeflate := []byte{0x78, 0x9c, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}

	_, err := decompressIfNeeded(corruptDeflate, "text/plain")
	if err == nil {
		t.Error("Expected error for corrupt deflate data, got nil")
	}

	harErr, ok := err.(*HarError)
	if !ok {
		t.Fatalf("Expected *HarError, got %T", err)
	}

	if harErr.Code != ErrCodeInvalidFormat {
		t.Errorf("Expected ErrCodeInvalidFormat, got %d", harErr.Code)
	}
}

func TestDecompressIfNeededPlainText(t *testing.T) {
	data := []byte("plain text data")

	result, err := decompressIfNeeded(data, "text/plain")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if string(result) != string(data) {
		t.Errorf("Expected '%s', got '%s'", string(data), string(result))
	}
}

func TestDecompressIfNeededEmpty(t *testing.T) {
	result, err := decompressIfNeeded([]byte{}, "text/plain")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected empty result, got %d bytes", len(result))
	}
}

func TestCompressDecompressRoundTripGzip(t *testing.T) {
	original := []byte("round trip test with gzip: Hello World! 12345")

	compressed, err := CompressContent(original, "gzip")
	if err != nil {
		t.Fatalf("CompressContent failed: %v", err)
	}

	decompressed, err := DecompressByEncoding(compressed, "gzip")
	if err != nil {
		t.Fatalf("DecompressByEncoding failed: %v", err)
	}

	if string(decompressed) != string(original) {
		t.Errorf("Round trip failed: expected '%s', got '%s'", string(original), string(decompressed))
	}
}

func TestCompressDecompressRoundTripDeflate(t *testing.T) {
	original := []byte("round trip test with deflate: Hello World! 12345")

	compressed, err := CompressContent(original, "deflate")
	if err != nil {
		t.Fatalf("CompressContent failed: %v", err)
	}

	decompressed, err := DecompressByEncoding(compressed, "deflate")
	if err != nil {
		t.Fatalf("DecompressByEncoding failed: %v", err)
	}

	if string(decompressed) != string(original) {
		t.Errorf("Round trip failed: expected '%s', got '%s'", string(original), string(decompressed))
	}
}

func TestDecompressByEncodingCaseInsensitive(t *testing.T) {
	original := []byte("case insensitive test")

	// Compress with gzip
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	if _, err := writer.Write(original); err != nil {
		t.Fatalf("Failed to compress: %v", err)
	}
	writer.Close()

	// Test with uppercase
	decompressed, err := DecompressByEncoding(buf.Bytes(), "GZIP")
	if err != nil {
		t.Fatalf("DecompressByEncoding GZIP failed: %v", err)
	}

	if string(decompressed) != string(original) {
		t.Errorf("Expected '%s', got '%s'", string(original), string(decompressed))
	}

	// Test with mixed case
	decompressed, err = DecompressByEncoding(buf.Bytes(), "Gzip")
	if err != nil {
		t.Fatalf("DecompressByEncoding Gzip failed: %v", err)
	}

	if string(decompressed) != string(original) {
		t.Errorf("Expected '%s', got '%s'", string(original), string(decompressed))
	}
}

func TestCompressContentCaseInsensitive(t *testing.T) {
	original := []byte("case test")

	_, err := CompressContent(original, "GZIP")
	if err != nil {
		t.Fatalf("CompressContent GZIP failed: %v", err)
	}

	_, err = CompressContent(original, "Gzip")
	if err != nil {
		t.Fatalf("CompressContent Gzip failed: %v", err)
	}
}

// helper to avoid importing io in test file
func readAll(r interface{ Read([]byte) (int, error) }) ([]byte, error) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r)
	return buf.Bytes(), err
}
