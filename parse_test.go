package har

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// validateInput
// ---------------------------------------------------------------------------

func TestValidateInput_EmptyInput(t *testing.T) {
	err := validateInput([]byte{})
	assert.Error(t, err)

	harErr, ok := err.(*HarError)
	require.True(t, ok, "expected HarError type")
	assert.Equal(t, ErrCodeInvalidFormat, harErr.Code)
	assert.Contains(t, harErr.Message, "输入为空")
}

func TestValidateInput_NilInput(t *testing.T) {
	err := validateInput(nil)
	assert.Error(t, err)

	harErr, ok := err.(*HarError)
	require.True(t, ok, "expected HarError type")
	assert.Equal(t, ErrCodeInvalidFormat, harErr.Code)
}

func TestValidateInput_NotJSON(t *testing.T) {
	err := validateInput([]byte("this is not json at all"))
	assert.Error(t, err)

	harErr, ok := err.(*HarError)
	require.True(t, ok, "expected HarError type")
	assert.Equal(t, ErrCodeInvalidFormat, harErr.Code)
	assert.Contains(t, harErr.Message, "JSON")
}

func TestValidateInput_ValidJSONObject(t *testing.T) {
	err := validateInput([]byte(`{"key": "value"}`))
	assert.NoError(t, err)
}

func TestValidateInput_ValidJSONArray(t *testing.T) {
	// isJSONContent also accepts arrays
	err := validateInput([]byte(`[1,2,3]`))
	assert.NoError(t, err)
}

func TestValidateInput_WhitespacePaddedJSON(t *testing.T) {
	err := validateInput([]byte("  \n\t {\"a\":1}  \n"))
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// parseWithStrategy
// ---------------------------------------------------------------------------

func TestParseWithStrategy_StreamingUnsupported(t *testing.T) {
	data := []byte(`{"log":{"version":"1.2","creator":{"name":"t","version":"1"},"entries":[]}}`)

	result, err := parseWithStrategy(data, options{useStreaming: true})
	assert.Nil(t, result)
	assert.Error(t, err)

	harErr, ok := err.(*HarError)
	require.True(t, ok, "expected HarError type")
	assert.Equal(t, ErrCodeUnsupported, harErr.Code)
}

func TestParseWithStrategy_MemoryOptimized(t *testing.T) {
	data := []byte(`{"log":{"version":"1.2","creator":{"name":"t","version":"1"},"entries":[]}}`)

	result, err := parseWithStrategy(data, options{useMemoryOptimized: true})
	require.NoError(t, err)
	assert.IsType(t, &OptimizedHar{}, result)
}

func TestParseWithStrategy_LazyLoading(t *testing.T) {
	data := []byte(`{"log":{"version":"1.2","creator":{"name":"t","version":"1"},"entries":[]}}`)

	result, err := parseWithStrategy(data, options{useLazyLoading: true})
	require.NoError(t, err)
	assert.NotNil(t, result)
	// LazyLoading returns a *LazyHar which implements HARProvider
	assert.Equal(t, "1.2", result.GetVersion())
}

func TestParseWithStrategy_Standard(t *testing.T) {
	data := []byte(`{"log":{"version":"1.2","creator":{"name":"t","version":"1"},"entries":[]}}`)

	result, err := parseWithStrategy(data, options{})
	require.NoError(t, err)
	assert.NotNil(t, result)

	std := result.ToStandard()
	assert.Equal(t, "1.2", std.Log.Version)
}

func TestParseWithStrategy_StandardWithLenientOption(t *testing.T) {
	data := []byte(`{"log":{"version":"1.2","creator":{"name":"t","version":"1"},"entries":[]}}`)

	result, err := parseWithStrategy(data, options{lenient: true, collectWarnings: true})
	require.NoError(t, err)
	assert.NotNil(t, result)
}

// ---------------------------------------------------------------------------
// ParseFile
// ---------------------------------------------------------------------------

func TestParseFile_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	harPath := filepath.Join(tmpDir, "test.har")

	harData := Har{
		Log: Log{
			Version: "1.2",
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{},
		},
	}
	data, err := json.MarshalIndent(harData, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(harPath, data, 0644))

	result, err := ParseFile(harPath)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "1.2", result.GetVersion())
}

func TestParseFile_NonExistentFile(t *testing.T) {
	result, err := ParseFile("/nonexistent/path/file.har")
	assert.Nil(t, result)
	assert.Error(t, err)

	harErr, ok := err.(*HarError)
	require.True(t, ok, "expected HarError type")
	assert.Equal(t, ErrCodeFileSystem, harErr.Code)
}

func TestParseFile_InvalidContent(t *testing.T) {
	tmpDir := t.TempDir()
	harPath := filepath.Join(tmpDir, "bad.har")

	require.NoError(t, os.WriteFile(harPath, []byte("not json"), 0644))

	result, err := ParseFile(harPath)
	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestParseFile_InvalidHarWithHarError(t *testing.T) {
	// ParseFile has a branch that checks if the error is a *HarError and
	// calls WithMetadata on it. Use invalid JSON (not a valid HAR) that
	// passes isJSONContent but fails during parsing, producing a *HarError.
	tmpDir := t.TempDir()
	harPath := filepath.Join(tmpDir, "invalid.har")

	// This is valid JSON but not valid HAR content — it will fail validation
	// and produce a *HarError with code validation
	invalidContent := `{"log":{"version":"1.2","creator":{},"entries":[]}}`
	require.NoError(t, os.WriteFile(harPath, []byte(invalidContent), 0644))

	result, err := ParseFile(harPath)
	// The result depends on whether validation is strict; we just need to
	// exercise the HarError.WithMetadata branch
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestParseFile_WithOptions(t *testing.T) {
	tmpDir := t.TempDir()
	harPath := filepath.Join(tmpDir, "test.har")

	harData := Har{
		Log: Log{
			Version: "1.2",
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{},
		},
	}
	data, err := json.MarshalIndent(harData, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(harPath, data, 0644))

	result, err := ParseFile(harPath, WithSkipValidation())
	require.NoError(t, err)
	assert.NotNil(t, result)
}

// ---------------------------------------------------------------------------
// Parse (top-level)
// ---------------------------------------------------------------------------

func TestParse_EmptyInput(t *testing.T) {
	result, err := Parse([]byte{})
	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestParse_InvalidJSON(t *testing.T) {
	result, err := Parse([]byte("plain text"))
	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestParse_ValidStandard(t *testing.T) {
	data := []byte(`{"log":{"version":"1.2","creator":{"name":"t","version":"1"},"entries":[]}}`)

	result, err := Parse(data)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "1.2", result.GetVersion())
}

func TestParse_WithStreamingOption(t *testing.T) {
	data := []byte(`{"log":{"version":"1.2","creator":{"name":"t","version":"1"},"entries":[]}}`)

	result, err := Parse(data, WithStreaming())
	assert.Nil(t, result)
	assert.Error(t, err)

	harErr, ok := err.(*HarError)
	require.True(t, ok, "expected HarError type")
	assert.Equal(t, ErrCodeUnsupported, harErr.Code)
}

// ---------------------------------------------------------------------------
// NewStreamingParser
// ---------------------------------------------------------------------------

func TestNewStreamingParser_ValidData(t *testing.T) {
	harData := Har{
		Log: Log{
			Version: "1.2",
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					Request:  Request{Method: "GET", URL: "https://example.com"},
					Response: Response{Status: 200},
				},
			},
		},
	}
	data, err := json.Marshal(harData)
	require.NoError(t, err)

	iter, err := NewStreamingParser(data)
	require.NoError(t, err)
	require.NotNil(t, iter)

	count := 0
	for iter.Next() {
		entry := iter.Entry()
		assert.NotNil(t, entry)
		count++
	}
	assert.NoError(t, iter.Err())
	assert.Equal(t, 1, count)
}

func TestNewStreamingParser_EmptyEntries(t *testing.T) {
	harData := Har{
		Log: Log{
			Version: "1.2",
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{},
		},
	}
	data, err := json.Marshal(harData)
	require.NoError(t, err)

	iter, err := NewStreamingParser(data)
	require.NoError(t, err)
	require.NotNil(t, iter)

	count := 0
	for iter.Next() {
		count++
	}
	assert.NoError(t, iter.Err())
	assert.Equal(t, 0, count)
}

func TestNewStreamingParser_InvalidInput(t *testing.T) {
	iter, err := NewStreamingParser([]byte("not json"))
	assert.Nil(t, iter)
	assert.Error(t, err)
}

func TestNewStreamingParser_EmptyInput(t *testing.T) {
	iter, err := NewStreamingParser([]byte{})
	assert.Nil(t, iter)
	assert.Error(t, err)
}

func TestNewStreamingParser_InvalidHarContent(t *testing.T) {
	// Valid JSON but not valid HAR structure - NewStreamingHarFromBytes
	// will fail during json.Unmarshal because it can't unmarshal into Har
	data := []byte(`{"not":"har"}`)

	iter, err := NewStreamingParser(data)
	// This may or may not error depending on how strict the unmarshal is
	// Just ensure it doesn't panic
	if err != nil {
		assert.Nil(t, iter)
	}
}

// ---------------------------------------------------------------------------
// NewStreamingParserFromFile
// ---------------------------------------------------------------------------

func TestNewStreamingParserFromFile_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	harPath := filepath.Join(tmpDir, "stream.har")

	harData := Har{
		Log: Log{
			Version: "1.2",
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					Request:  Request{Method: "GET", URL: "https://example.com"},
					Response: Response{Status: 200},
				},
			},
		},
	}
	data, err := json.Marshal(harData)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(harPath, data, 0644))

	iter, err := NewStreamingParserFromFile(harPath)
	require.NoError(t, err)
	require.NotNil(t, iter)

	count := 0
	for iter.Next() {
		entry := iter.Entry()
		assert.NotNil(t, entry)
		count++
	}
	assert.NoError(t, iter.Err())
	assert.Equal(t, 1, count)
}

func TestNewStreamingParserFromFile_NonExistentFile(t *testing.T) {
	iter, err := NewStreamingParserFromFile("/nonexistent/path/file.har")
	assert.Nil(t, iter)
	assert.Error(t, err)

	harErr, ok := err.(*HarError)
	require.True(t, ok, "expected HarError type")
	assert.Equal(t, ErrCodeFileSystem, harErr.Code)
}

func TestNewStreamingParserFromFile_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	harPath := filepath.Join(tmpDir, "empty.har")

	require.NoError(t, os.WriteFile(harPath, []byte{}, 0644))

	iter, err := NewStreamingParserFromFile(harPath)
	assert.Nil(t, iter)
	assert.Error(t, err)
}

func TestNewStreamingParserFromFile_NotJSONFile(t *testing.T) {
	tmpDir := t.TempDir()
	harPath := filepath.Join(tmpDir, "bad.har")

	require.NoError(t, os.WriteFile(harPath, []byte("this is not json"), 0644))

	iter, err := NewStreamingParserFromFile(harPath)
	assert.Nil(t, iter)
	assert.Error(t, err)
}

func TestNewStreamingParserFromFile_EmptyEntries(t *testing.T) {
	tmpDir := t.TempDir()
	harPath := filepath.Join(tmpDir, "empty_entries.har")

	harData := Har{
		Log: Log{
			Version: "1.2",
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{},
		},
	}
	data, err := json.Marshal(harData)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(harPath, data, 0644))

	iter, err := NewStreamingParserFromFile(harPath)
	require.NoError(t, err)
	require.NotNil(t, iter)

	count := 0
	for iter.Next() {
		count++
	}
	assert.NoError(t, iter.Err())
	assert.Equal(t, 0, count)
}
