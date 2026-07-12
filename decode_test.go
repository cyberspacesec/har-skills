package har

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/base64"
	"errors"
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
		{[]byte{0x78}, false},                   // too short
		{[]byte{}, false},                       // empty
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
	original := []byte("brotli round-trip via DecompressByEncoding 数据")
	compressed, err := CompressContent(original, "br")
	if err != nil {
		t.Fatalf("CompressContent br failed: %v", err)
	}

	decompressed, err := DecompressByEncoding(compressed, "br")
	if err != nil {
		t.Fatalf("DecompressByEncoding br failed: %v", err)
	}

	if string(decompressed) != string(original) {
		t.Errorf("Expected '%s', got '%s'", string(original), string(decompressed))
	}
}

func TestDecompressByEncodingZstd(t *testing.T) {
	original := []byte("zstd round-trip via DecompressByEncoding 数据")
	compressed, err := CompressContent(original, "zstd")
	if err != nil {
		t.Fatalf("CompressContent zstd failed: %v", err)
	}

	decompressed, err := DecompressByEncoding(compressed, "zstd")
	if err != nil {
		t.Fatalf("DecompressByEncoding zstd failed: %v", err)
	}

	if string(decompressed) != string(original) {
		t.Errorf("Expected '%s', got '%s'", string(original), string(decompressed))
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
	original := []byte("multi-encoding round-trip 数据")

	// gzip(deflate(original))
	deflated, err := CompressContent(original, "deflate")
	if err != nil {
		t.Fatalf("CompressContent deflate failed: %v", err)
	}
	gzipWrapped, err := CompressContent(deflated, "gzip")
	if err != nil {
		t.Fatalf("CompressContent gzip failed: %v", err)
	}

	decompressed, err := DecompressByEncoding(gzipWrapped, "gzip, deflate")
	if err != nil {
		t.Fatalf("DecompressByEncoding multi-encoding failed: %v", err)
	}

	if string(decompressed) != string(original) {
		t.Errorf("Expected '%s', got '%s'", string(original), string(decompressed))
	}

	// 损坏的输入：第一层编码解压失败应返回错误（含层号信息）
	_, err = DecompressByEncoding([]byte("not compressed"), "gzip, deflate")
	if err == nil {
		t.Error("Expected error for corrupt multi-encoding input, got nil")
	}
	if !strings.Contains(err.Error(), "第0层") {
		t.Errorf("Error should mention layer index, got: %s", err.Error())
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
	original := []byte("brotli compress round-trip")
	compressed, err := CompressContent(original, "br")
	if err != nil {
		t.Fatalf("CompressContent br failed: %v", err)
	}

	decompressed, err := DecompressByEncoding(compressed, "br")
	if err != nil {
		t.Fatalf("DecompressByEncoding br failed: %v", err)
	}

	if string(decompressed) != string(original) {
		t.Errorf("Expected '%s', got '%s'", string(original), string(decompressed))
	}
}

func TestCompressContentZstd(t *testing.T) {
	original := []byte("zstd compress round-trip")
	compressed, err := CompressContent(original, "zstd")
	if err != nil {
		t.Fatalf("CompressContent zstd failed: %v", err)
	}

	decompressed, err := DecompressByEncoding(compressed, "zstd")
	if err != nil {
		t.Fatalf("DecompressByEncoding zstd failed: %v", err)
	}

	if string(decompressed) != string(original) {
		t.Errorf("Expected '%s', got '%s'", string(original), string(decompressed))
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

// --- Comprehensive branch coverage tests for decode.go ---

func TestDecodeContentBase64WithEmptyText(t *testing.T) {
	// Branch: encoding is "base64" but Text is "" -> falls to else branch returning nil,nil
	content := &Content{
		Size:     0,
		MimeType: "text/plain",
		Text:     "",
		Encoding: "base64",
	}

	data, err := content.DecodeContent()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if data != nil {
		t.Errorf("Expected nil data for base64 encoding with empty text, got %v", data)
	}
}

func TestDecodeContentInvalidBase64(t *testing.T) {
	// Branch: base64 decoding fails both StdEncoding and URLEncoding
	content := &Content{
		Size:     10,
		MimeType: "text/plain",
		Text:     "!!!invalid-base64!!!",
		Encoding: "base64",
	}

	data, err := content.DecodeContent()
	if err == nil {
		t.Error("Expected error for invalid base64, got nil")
	}
	if data != nil {
		t.Errorf("Expected nil data for invalid base64, got %v", data)
	}

	harErr, ok := err.(*HarError)
	if !ok {
		t.Fatalf("Expected *HarError, got %T", err)
	}
	if harErr.Code != ErrCodeInvalidFormat {
		t.Errorf("Expected ErrCodeInvalidFormat, got %d", harErr.Code)
	}
	if !strings.Contains(harErr.Message, "base64") {
		t.Errorf("Error message should mention base64, got: %s", harErr.Message)
	}
}

func TestDecodeContentURLSafeBase64Fallback(t *testing.T) {
	// Branch: StdEncoding fails but URLEncoding succeeds (URL-safe base64 with padding)
	// Create data that contains URL-safe characters that StdEncoding can't handle
	original := "Hello World!"
	// Use URL-safe encoding explicitly
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
	if string(data) != original {
		t.Errorf("Expected '%s', got '%s'", original, string(data))
	}
}

func TestDecodeContentCaseInsensitiveBase64(t *testing.T) {
	// Branch: strings.EqualFold matches "BASE64" (uppercase)
	original := "test data"
	encoded := base64.StdEncoding.EncodeToString([]byte(original))

	content := &Content{
		Size:     len(encoded),
		MimeType: "text/plain",
		Text:     encoded,
		Encoding: "BASE64", // uppercase
	}

	data, err := content.DecodeContent()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if string(data) != original {
		t.Errorf("Expected '%s', got '%s'", original, string(data))
	}
}

func TestDecodeEntryContentNilEntry(t *testing.T) {
	// Branch: nil Entries
	var entry *Entries
	_, err := entry.DecodeContent()
	if err == nil {
		t.Error("Expected error for nil entry, got nil")
	}

	harErr, ok := err.(*HarError)
	if !ok {
		t.Fatalf("Expected *HarError, got %T", err)
	}
	if harErr.Code != ErrCodeInvalidFormat {
		t.Errorf("Expected ErrCodeInvalidFormat, got %d", harErr.Code)
	}
}

func TestDecodeAllContentNilHar(t *testing.T) {
	// Branch: nil Har
	var h *Har
	_, err := h.DecodeAllContent()
	if err == nil {
		t.Error("Expected error for nil HAR, got nil")
	}

	harErr, ok := err.(*HarError)
	if !ok {
		t.Fatalf("Expected *HarError, got %T", err)
	}
	if harErr.Code != ErrCodeInvalidFormat {
		t.Errorf("Expected ErrCodeInvalidFormat, got %d", harErr.Code)
	}
}

func TestDecodeAllContentWithErrors(t *testing.T) {
	// Branch: some entries fail decoding, resulting in partial errors
	h := NewHar()

	// Valid entry
	e1 := h.AddEntry("GET", "https://example.com/1", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e1.SetResponseContentText("valid response")

	// Entry with invalid base64 content
	e2 := h.AddEntry("GET", "https://example.com/2", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")
	e2.Response.Content.Encoding = "base64"
	e2.Response.Content.Text = "!!!invalid!!!"

	results, err := h.DecodeAllContent()
	// Should return partial results with error
	if err == nil {
		t.Error("Expected error for entries with decode failures, got nil")
	}

	// First result should be valid
	if string(results[0]) != "valid response" {
		t.Errorf("Expected 'valid response', got '%s'", string(results[0]))
	}

	// Second result should be nil (failed decode)
	if results[1] != nil {
		t.Errorf("Expected nil for failed decode, got %v", results[1])
	}

	harErr, ok := err.(*HarError)
	if !ok {
		t.Fatalf("Expected *HarError, got %T", err)
	}
	if harErr.Code != ErrCodeInvalidFormat {
		t.Errorf("Expected ErrCodeInvalidFormat, got %d", harErr.Code)
	}
	if !harErr.HasPartialErrors() {
		t.Fatal("Expected DecodeAllContent error to include partial errors")
	}
	partials := harErr.GetPartialErrors()
	if len(partials) != 1 {
		t.Fatalf("Expected 1 partial error, got %d", len(partials))
	}
	partial := partials[0]
	if partial.Code != ErrCodeInvalidFormat {
		t.Errorf("Expected partial ErrCodeInvalidFormat, got %d", partial.Code)
	}
	if partial.Field != "log.entries[1].response.content" {
		t.Errorf("Expected partial field for second entry content, got %q", partial.Field)
	}
	if partial.Metadata["entry_index"] != 1 {
		t.Errorf("Expected partial entry_index metadata 1, got %v", partial.Metadata["entry_index"])
	}
	if harErr.Metadata["error_count"] != 1 {
		t.Errorf("Expected root error_count metadata 1, got %v", harErr.Metadata["error_count"])
	}
}

func TestDecodePartialErrorWrapsGenericError(t *testing.T) {
	root := errors.New("raw decode failure")

	harErr := decodePartialError(3, root)
	if harErr.Code != ErrCodeInvalidFormat {
		t.Errorf("Expected ErrCodeInvalidFormat, got %d", harErr.Code)
	}
	if harErr.Err != root {
		t.Error("Expected generic error to be preserved as wrapped error")
	}
	if harErr.Field != "log.entries[3].response.content" {
		t.Errorf("Expected content field path, got %q", harErr.Field)
	}
	if harErr.Metadata["entry_index"] != 3 {
		t.Errorf("Expected entry_index metadata 3, got %v", harErr.Metadata["entry_index"])
	}
}

func TestDecodeAllContentAllValid(t *testing.T) {
	// Branch: all entries decode successfully, no error returned
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

func TestIsBase64EncodedNil(t *testing.T) {
	// Branch: nil Content returns false
	var content *Content
	if content.IsBase64Encoded() {
		t.Error("Expected false for nil Content")
	}
}

func TestIsBase64EncodedCaseInsensitive(t *testing.T) {
	// Branch: strings.EqualFold matches different cases
	content1 := &Content{Encoding: "BASE64"}
	if !content1.IsBase64Encoded() {
		t.Error("Expected true for uppercase BASE64")
	}

	content2 := &Content{Encoding: "Base64"}
	if !content2.IsBase64Encoded() {
		t.Error("Expected true for mixed case Base64")
	}

	content3 := &Content{Encoding: "not-base64"}
	if content3.IsBase64Encoded() {
		t.Error("Expected false for non-base64 encoding")
	}

	content4 := &Content{Encoding: ""}
	if content4.IsBase64Encoded() {
		t.Error("Expected false for empty encoding")
	}
}

func TestIsCompressedNilEntry(t *testing.T) {
	// Branch: nil Entries returns false
	var entry *Entries
	if entry.IsCompressed() {
		t.Error("Expected false for nil entry")
	}
}

func TestIsCompressedWhitespaceInValue(t *testing.T) {
	// Branch: Content-Encoding value with whitespace
	entry := &Entries{
		Response: Response{
			Headers: []Headers{
				{Name: "Content-Encoding", Value: "  gzip  "},
			},
		},
	}
	if !entry.IsCompressed() {
		t.Error("Expected true for gzip with whitespace")
	}
}

func TestIsCompressedNoHeaders(t *testing.T) {
	// Branch: no headers at all
	entry := &Entries{
		Response: Response{
			Headers: []Headers{},
		},
	}
	if entry.IsCompressed() {
		t.Error("Expected false for entry with no headers")
	}
}

func TestIsCompressedCaseInsensitiveHeaderName(t *testing.T) {
	// Branch: strings.EqualFold matches "content-encoding" in different cases
	entry := &Entries{
		Response: Response{
			Headers: []Headers{
				{Name: "content-encoding", Value: "gzip"},
			},
		},
	}
	if !entry.IsCompressed() {
		t.Error("Expected true for lowercase content-encoding header name")
	}
}

func TestGetContentEncodingNilEntry(t *testing.T) {
	// Branch: nil Entries returns empty string
	var entry *Entries
	if entry.GetContentEncoding() != "" {
		t.Errorf("Expected empty string for nil entry, got '%s'", entry.GetContentEncoding())
	}
}

func TestGetContentEncodingNoHeader(t *testing.T) {
	// Branch: no Content-Encoding header found
	entry := &Entries{
		Response: Response{
			Headers: []Headers{
				{Name: "Content-Type", Value: "text/html"},
			},
		},
	}
	if entry.GetContentEncoding() != "" {
		t.Errorf("Expected empty string, got '%s'", entry.GetContentEncoding())
	}
}

func TestGetContentEncodingNoHeaders(t *testing.T) {
	// Branch: empty headers list
	entry := &Entries{
		Response: Response{
			Headers: []Headers{},
		},
	}
	if entry.GetContentEncoding() != "" {
		t.Errorf("Expected empty string, got '%s'", entry.GetContentEncoding())
	}
}

func TestGetContentEncodingWhitespaceValue(t *testing.T) {
	// Branch: Content-Encoding value with leading/trailing whitespace (TrimSpace)
	entry := &Entries{
		Response: Response{
			Headers: []Headers{
				{Name: "Content-Encoding", Value: "  gzip  "},
			},
		},
	}
	if entry.GetContentEncoding() != "gzip" {
		t.Errorf("Expected 'gzip' (trimmed), got '%s'", entry.GetContentEncoding())
	}
}

func TestDecodeEntryTextNilEntry(t *testing.T) {
	// Branch: nil Entries
	var entry *Entries
	_, err := entry.DecodeEntryText()
	if err == nil {
		t.Error("Expected error for nil entry, got nil")
	}
}

func TestDecodeEntryTextNilData(t *testing.T) {
	// Branch: DecodeContent returns nil data (empty Content)
	entry := &Entries{
		Response: Response{
			Content: Content{
				Size:     0,
				MimeType: "text/plain",
				Text:     "",
			},
		},
	}

	text, err := entry.DecodeEntryText()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if text != "" {
		t.Errorf("Expected empty string for nil data, got '%s'", text)
	}
}

func TestDecodeEntryTextError(t *testing.T) {
	// Branch: DecodeContent returns error
	entry := &Entries{
		Response: Response{
			Content: Content{
				Encoding: "base64",
				Text:     "!!!invalid!!!",
			},
		},
	}

	text, err := entry.DecodeEntryText()
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if text != "" {
		t.Errorf("Expected empty string on error, got '%s'", text)
	}
}

func TestDecompressIfNeededGzipSuccess(t *testing.T) {
	// Branch: successful gzip decompression via decompressIfNeeded
	original := []byte("test data for gzip")
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	if _, err := writer.Write(original); err != nil {
		t.Fatalf("Failed to compress: %v", err)
	}
	writer.Close()

	result, err := decompressIfNeeded(buf.Bytes(), "application/gzip")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if string(result) != string(original) {
		t.Errorf("Expected '%s', got '%s'", string(original), string(result))
	}
}

func TestDecompressIfNeededDeflateSuccess(t *testing.T) {
	// Branch: successful deflate decompression via decompressIfNeeded
	original := []byte("test data for deflate")
	var buf bytes.Buffer
	writer := zlib.NewWriter(&buf)
	if _, err := writer.Write(original); err != nil {
		t.Fatalf("Failed to compress: %v", err)
	}
	writer.Close()

	result, err := decompressIfNeeded(buf.Bytes(), "application/deflate")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if string(result) != string(original) {
		t.Errorf("Expected '%s', got '%s'", string(original), string(result))
	}
}

func TestDecompressIfNeededNeitherGzipNorDeflate(t *testing.T) {
	// Branch: data is neither gzip nor deflate, returned as-is
	data := []byte("just plain text, not compressed")

	result, err := decompressIfNeeded(data, "text/plain")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if string(result) != string(data) {
		t.Errorf("Expected '%s', got '%s'", string(data), string(result))
	}
}

func TestDecompressIfNeededShortData(t *testing.T) {
	// Branch: data too short to be gzip or deflate (< 2 bytes)
	shortData := []byte{0x1f} // only 1 byte, can't match gzip magic
	result, err := decompressIfNeeded(shortData, "text/plain")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if string(result) != string(shortData) {
		t.Errorf("Expected short data returned as-is, got '%s'", string(result))
	}
}

func TestDecompressByEncodingDeflateInvalidData(t *testing.T) {
	// Branch: invalid deflate data
	data := []byte("this is not deflate data")

	_, err := DecompressByEncoding(data, "deflate")
	if err == nil {
		t.Error("Expected error for invalid deflate data, got nil")
	}

	harErr, ok := err.(*HarError)
	if !ok {
		t.Fatalf("Expected *HarError, got %T", err)
	}
	if harErr.Code != ErrCodeInvalidFormat {
		t.Errorf("Expected ErrCodeInvalidFormat, got %d", harErr.Code)
	}
	if !strings.Contains(harErr.Message, "deflate") {
		t.Errorf("Error message should mention deflate, got: %s", harErr.Message)
	}
}

func TestDecompressByEncodingWhitespaceInEncoding(t *testing.T) {
	// Branch: encoding value with leading/trailing whitespace
	original := []byte("test data")
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	if _, err := writer.Write(original); err != nil {
		t.Fatalf("Failed to compress: %v", err)
	}
	writer.Close()

	// Encoding with whitespace should be trimmed and matched
	result, err := DecompressByEncoding(buf.Bytes(), "  gzip  ")
	if err != nil {
		t.Fatalf("Unexpected error for whitespace encoding: %v", err)
	}
	if string(result) != string(original) {
		t.Errorf("Expected '%s', got '%s'", string(original), string(result))
	}
}

func TestCompressContentIdentityNotSupported(t *testing.T) {
	// Branch: "identity" falls to default case in CompressContent (not a compressible encoding)
	data := []byte("identity data")
	_, err := CompressContent(data, "identity")
	if err == nil {
		t.Error("Expected error for identity compression, got nil")
	}

	harErr, ok := err.(*HarError)
	if !ok {
		t.Fatalf("Expected *HarError, got %T", err)
	}
	if harErr.Code != ErrCodeUnsupported {
		t.Errorf("Expected ErrCodeUnsupported, got %d", harErr.Code)
	}
}

func TestCompressContentEmptyDeflate(t *testing.T) {
	// Branch: empty data for deflate compression
	result, err := CompressContent([]byte{}, "deflate")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("Expected empty result, got %d bytes", len(result))
	}
}

func TestDecompressByEncodingCaseInsensitiveDeflate(t *testing.T) {
	// Branch: case-insensitive deflate encoding matching
	original := []byte("case test deflate")
	var buf bytes.Buffer
	writer := zlib.NewWriter(&buf)
	if _, err := writer.Write(original); err != nil {
		t.Fatalf("Failed to compress: %v", err)
	}
	writer.Close()

	// Test uppercase DEFLATE
	decompressed, err := DecompressByEncoding(buf.Bytes(), "DEFLATE")
	if err != nil {
		t.Fatalf("DecompressByEncoding DEFLATE failed: %v", err)
	}
	if string(decompressed) != string(original) {
		t.Errorf("Expected '%s', got '%s'", string(original), string(decompressed))
	}
}

func TestDecompressIfNeededGzipReadAllError(t *testing.T) {
	// Branch: gzip.NewReader succeeds but io.ReadAll fails on corrupt stream
	// Create truncated gzip data: valid header but truncated body
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	if _, err := writer.Write([]byte("some data")); err != nil {
		t.Fatalf("Failed to compress: %v", err)
	}
	writer.Close()

	// Truncate the gzip data so ReadAll will fail
	truncated := buf.Bytes()[:buf.Len()-4]

	_, err := decompressIfNeeded(truncated, "text/plain")
	if err == nil {
		t.Error("Expected error for truncated gzip data, got nil")
	}

	harErr, ok := err.(*HarError)
	if !ok {
		t.Fatalf("Expected *HarError, got %T", err)
	}
	if harErr.Code != ErrCodeInvalidFormat {
		t.Errorf("Expected ErrCodeInvalidFormat, got %d", harErr.Code)
	}
}

func TestDecompressIfNeededDeflateReadAllError(t *testing.T) {
	// Branch: zlib.NewReader succeeds but io.ReadAll fails on corrupt stream
	// Create truncated zlib data: valid header but truncated body
	var buf bytes.Buffer
	writer := zlib.NewWriter(&buf)
	if _, err := writer.Write([]byte("some data")); err != nil {
		t.Fatalf("Failed to compress: %v", err)
	}
	writer.Close()

	// Truncate the zlib data so ReadAll will fail
	truncated := buf.Bytes()[:buf.Len()-4]

	_, err := decompressIfNeeded(truncated, "text/plain")
	if err == nil {
		t.Error("Expected error for truncated deflate data, got nil")
	}

	harErr, ok := err.(*HarError)
	if !ok {
		t.Fatalf("Expected *HarError, got %T", err)
	}
	if harErr.Code != ErrCodeInvalidFormat {
		t.Errorf("Expected ErrCodeInvalidFormat, got %d", harErr.Code)
	}
}

func TestDecodeContentGzipCompressedNonBase64(t *testing.T) {
	// Branch: Text contains gzip data without base64 encoding (else if c.Text != "")
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	if _, err := writer.Write([]byte("plain gzip")); err != nil {
		t.Fatalf("Failed to compress: %v", err)
	}
	writer.Close()

	// Put raw gzip bytes as text (not base64 encoded)
	content := &Content{
		Size:     buf.Len(),
		MimeType: "text/plain",
		Text:     string(buf.Bytes()),
		Encoding: "", // not base64
	}

	data, err := content.DecodeContent()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if string(data) != "plain gzip" {
		t.Errorf("Expected 'plain gzip', got '%s'", string(data))
	}
}

func TestDecodeContentDeflateCompressedNonBase64(t *testing.T) {
	// Branch: Text contains deflate data without base64 encoding
	var buf bytes.Buffer
	writer := zlib.NewWriter(&buf)
	if _, err := writer.Write([]byte("plain deflate")); err != nil {
		t.Fatalf("Failed to compress: %v", err)
	}
	writer.Close()

	// Put raw zlib bytes as text (not base64 encoded)
	content := &Content{
		Size:     buf.Len(),
		MimeType: "text/plain",
		Text:     string(buf.Bytes()),
		Encoding: "", // not base64
	}

	data, err := content.DecodeContent()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if string(data) != "plain deflate" {
		t.Errorf("Expected 'plain deflate', got '%s'", string(data))
	}
}

func TestDecompressByEncodingGzipReadAllError(t *testing.T) {
	// Branch: gzip.NewReader succeeds but io.ReadAll fails in DecompressByEncoding
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	if _, err := writer.Write([]byte("data")); err != nil {
		t.Fatalf("Failed to compress: %v", err)
	}
	writer.Close()

	// Truncate the gzip data
	truncated := buf.Bytes()[:buf.Len()-4]

	_, err := DecompressByEncoding(truncated, "gzip")
	if err == nil {
		t.Error("Expected error for truncated gzip in DecompressByEncoding, got nil")
	}

	harErr, ok := err.(*HarError)
	if !ok {
		t.Fatalf("Expected *HarError, got %T", err)
	}
	if harErr.Code != ErrCodeInvalidFormat {
		t.Errorf("Expected ErrCodeInvalidFormat, got %d", harErr.Code)
	}
}

func TestDecompressByEncodingDeflateReadAllError(t *testing.T) {
	// Branch: zlib.NewReader succeeds but io.ReadAll fails in DecompressByEncoding
	var buf bytes.Buffer
	writer := zlib.NewWriter(&buf)
	if _, err := writer.Write([]byte("data")); err != nil {
		t.Fatalf("Failed to compress: %v", err)
	}
	writer.Close()

	// Truncate the zlib data
	truncated := buf.Bytes()[:buf.Len()-4]

	_, err := DecompressByEncoding(truncated, "deflate")
	if err == nil {
		t.Error("Expected error for truncated deflate in DecompressByEncoding, got nil")
	}

	harErr, ok := err.(*HarError)
	if !ok {
		t.Fatalf("Expected *HarError, got %T", err)
	}
	if harErr.Code != ErrCodeInvalidFormat {
		t.Errorf("Expected ErrCodeInvalidFormat, got %d", harErr.Code)
	}
}

func TestCompressContentWhitespaceEncoding(t *testing.T) {
	// Branch: encoding value with whitespace is trimmed in CompressContent
	data := []byte("test data")

	// Whitespace-padded "gzip" should still work
	compressed, err := CompressContent(data, "  gzip  ")
	if err != nil {
		t.Fatalf("Unexpected error for whitespace encoding: %v", err)
	}
	if !isGzipData(compressed) {
		t.Error("Compressed data should have gzip magic bytes")
	}
}

// helper to avoid importing io in test file
func readAll(r interface{ Read([]byte) (int, error) }) ([]byte, error) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r)
	return buf.Bytes(), err
}
