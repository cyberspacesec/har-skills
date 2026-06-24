package har

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseHar(t *testing.T) {
	harFileContent, err := os.ReadFile("testdata/example.har")
	assert.Nil(t, err)
	har, err := ParseHar(harFileContent)
	assert.Nil(t, err)
	t.Log(har)
}

func TestParseHarFile(t *testing.T) {
	har, err := ParseHarFile("testdata/example.har")
	assert.Nil(t, err)
	t.Log(har)
}

// --- ParseHar comprehensive tests ---

func TestParseHarEmptyInput(t *testing.T) {
	// Cover the len(harFileBytes) == 0 branch
	har, err := ParseHar([]byte{})
	assert.Nil(t, har)
	assert.NotNil(t, err)
	assert.Equal(t, ErrCodeInvalidFormat, err.(*HarError).Code)
	assert.Contains(t, err.Error(), "输入为空")
}

func TestParseHarNilInput(t *testing.T) {
	// Cover nil/empty byte slice
	har, err := ParseHar(nil)
	assert.Nil(t, har)
	assert.NotNil(t, err)
	assert.Equal(t, ErrCodeInvalidFormat, err.(*HarError).Code)
}

func TestParseHarNotJSON(t *testing.T) {
	// Cover the !isJSONContent branch -> ErrNotJsonContent
	har, err := ParseHar([]byte("this is not json"))
	assert.Nil(t, har)
	assert.NotNil(t, err)
	assert.Equal(t, ErrCodeInvalidFormat, err.(*HarError).Code)
	assert.Contains(t, err.Error(), "JSON")
}

func TestParseHarNotJSONButArray(t *testing.T) {
	// Valid JSON array is not a valid HAR object (won't unmarshal into Har struct)
	har, err := ParseHar([]byte("[1, 2, 3]"))
	assert.Nil(t, har)
	assert.NotNil(t, err)
	// Should be a JSON parse error since unmarshal into Har fails
	assert.Equal(t, ErrCodeJSONParse, err.(*HarError).Code)
}

func TestParseHarInvalidJSONStructure(t *testing.T) {
	// Valid JSON but wrong structure for HAR
	har, err := ParseHar([]byte(`{"foo": "bar"}`))
	assert.Nil(t, har)
	assert.NotNil(t, err)
	// This should fail validation since log.entries would be nil and other required fields missing
}

func TestParseHarValidationFailure(t *testing.T) {
	// Cover the ValidateHarFile error path in ParseHar
	// missing_required.har has missing required fields
	harFileContent, err := os.ReadFile("testdata/missing_required.har")
	assert.Nil(t, err)
	h, err := ParseHar(harFileContent)
	assert.Nil(t, h)
	assert.NotNil(t, err)
}

func TestParseHarInvalidHARFile(t *testing.T) {
	// Test with invalid.har which is missing fields
	harFileContent, err := os.ReadFile("testdata/invalid.har")
	assert.Nil(t, err)
	har, err := ParseHar(harFileContent)
	assert.Nil(t, har)
	assert.NotNil(t, err)
}

// --- ParseHarFile comprehensive tests ---

func TestParseHarFileNonExistent(t *testing.T) {
	// Cover the os.ReadFile error path
	har, err := ParseHarFile("testdata/nonexistent.har")
	assert.Nil(t, har)
	assert.NotNil(t, err)
	assert.Equal(t, ErrCodeFileSystem, err.(*HarError).Code)
	assert.Contains(t, err.Error(), "无法读取文件")
}

func TestParseHarFileNotJSON(t *testing.T) {
	// Cover the path where file exists but is not JSON
	har, err := ParseHarFile("testdata/not_json.har")
	assert.Nil(t, har)
	assert.NotNil(t, err)
}

// --- Har SaveToFile tests ---

func TestHarSaveToFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "output.har")

	h, err := ParseHarFile("testdata/example.har")
	assert.Nil(t, err)

	err = h.SaveToFile(filePath, true)
	assert.Nil(t, err)

	parsed, err := ParseHarFile(filePath)
	assert.Nil(t, err)
	assert.Equal(t, h.GetEntryCount(), parsed.GetEntryCount())
	assert.Equal(t, h.Log.Creator.Name, parsed.Log.Creator.Name)
}

func TestHarSaveToFileNilHar(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "nil.har")

	var h *Har
	err := h.SaveToFile(filePath, true)
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)

	_, statErr := os.Stat(filePath)
	assert.True(t, os.IsNotExist(statErr))
}

func TestHarSaveToFileInvalidPath(t *testing.T) {
	h := NewHar()
	err := h.SaveToFile("/nonexistent/directory/deep/file.har", true)
	assertHarErrorCode(t, err, ErrCodeFileSystem)
}

func TestHarToJSON(t *testing.T) {
	h, err := ParseHarFile("testdata/example.har")
	assert.Nil(t, err)

	data, err := h.ToJSON(true)
	assert.Nil(t, err)
	assert.Contains(t, string(data), "\n")
	assert.Contains(t, string(data), `"log"`)

	data, err = h.ToJSON(false)
	assert.Nil(t, err)
	assert.NotContains(t, string(data), "\n")
	assert.Contains(t, string(data), `"log"`)
}

func TestHarToJSONNilHar(t *testing.T) {
	var h *Har
	data, err := h.ToJSON(true)
	assert.Nil(t, data)
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
}

// --- IsValidURL comprehensive tests ---

func TestIsValidURLValid(t *testing.T) {
	assert.True(t, IsValidURL("https://example.com"))
	assert.True(t, IsValidURL("http://localhost:8080/path?q=1"))
	assert.True(t, IsValidURL("ftp://files.example.com"))
	assert.True(t, IsValidURL("/relative/path"))
}

func TestIsValidURLInvalid(t *testing.T) {
	// url.Parse is very lenient; only truly unparseable strings return errors.
	// Control characters cause url.Parse to fail.
	assert.False(t, IsValidURL("://missing-scheme"))
	assert.False(t, IsValidURL(string([]byte{0x00})))
}

func TestIsValidURLEmpty(t *testing.T) {
	// Empty string parses successfully in url.Parse
	assert.True(t, IsValidURL(""))
}

func TestIsValidURLWithFragment(t *testing.T) {
	assert.True(t, IsValidURL("https://example.com/page#section"))
}

func TestIsValidURLWithUserInfo(t *testing.T) {
	assert.True(t, IsValidURL("https://user:pass@example.com/path"))
}
