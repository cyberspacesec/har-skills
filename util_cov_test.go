package har

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Cover Har.Clone Marshal-error branch (lines 26-29 -> returns nil).
// Injecting a func value into Response.Error (any, json:"_error") makes
// json.Marshal fail, so Clone returns nil.
func TestCovHarClone_MarshalError(t *testing.T) {
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.Response.Error = func() {} // unsupported by json.Marshal

	clone := h.Clone()
	assert.Nil(t, clone)
}

// Cover Har.Clone nil-receiver branch (lines 22-24) and happy path.
func TestCovHarClone_NilAndHappy(t *testing.T) {
	var h *Har
	assert.Nil(t, h.Clone())

	// happy path
	hh := NewHar()
	ee := hh.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	ee.SetResponseStatus(200, "OK")
	ee.AddRequestHeader("X-Test", "1")
	clone := hh.Clone()
	if assert.NotNil(t, clone) {
		assert.Equal(t, hh.Log.Entries[0].Request.URL, clone.Log.Entries[0].Request.URL)
		// Mutating clone must not affect original (deep copy).
		clone.Log.Entries[0].Request.URL = "https://changed.example.com"
		assert.Equal(t, "https://example.com", hh.Log.Entries[0].Request.URL)
	}
}

// Cover Har.SaveToFileGzipped Create-error branch (lines 142-145).
// os.Create on a path inside a non-existent directory fails.
func TestCovHarSaveToFileGzipped_CreateError(t *testing.T) {
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")

	badPath := "/nonexistent_dir_xyz_cov/deep/out.har.gz"
	err := h.SaveToFileGzipped(badPath, true)
	if err == nil {
		t.Fatal("expected error from os.Create failure, got nil")
	}
	assertHarErrorCode(t, err, ErrCodeFileSystem)
}

// Cover Har.SaveToFileGzipped nil-receiver branch (lines 133-135) and
// happy path (lines 142-146).
func TestCovHarSaveToFileGzipped_NilAndHappy(t *testing.T) {
	var h *Har
	err := h.SaveToFileGzipped("/tmp/whatever_cov.har.gz", true)
	if err == nil {
		t.Fatal("expected error for nil HAR, got nil")
	}
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)

	// happy path
	hh := NewHar()
	ee := hh.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	ee.SetResponseStatus(200, "OK")
	path := t.TempDir() + "/out.har.gz"
	if err := hh.SaveToFileGzipped(path, false); err != nil {
		t.Fatalf("unexpected error on happy path: %v", err)
	}
	assert.FileExists(t, path)
}

// Cover Har.SaveToFileGzipped ToJSON-error branch (lines 138-140):
// inject a func value into Response.Error (any, json:"_error") to make
// ToJSON fail; SaveToFileGzipped propagates the error before reaching
// os.Create.
func TestCovHarSaveToFileGzipped_ToJSONError(t *testing.T) {
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.Response.Error = func() {} // unsupported by json.Marshal

	err := h.SaveToFileGzipped(t.TempDir()+"/should-not-exist.har.gz", true)
	if err == nil {
		t.Fatal("expected error from ToJSON failure, got nil")
	}
	assertHarErrorCode(t, err, ErrCodeJSONParse)
}

// Cover Har.Clone Unmarshal-error branch (lines 32-34).
//
// In practice this branch is unreachable: any data that json.Marshal
// produces successfully can be json.Unmarshal'd back into *Har, so the
// Marshal-error test above (lines 26-29) returns nil before reaching the
// Unmarshal call. This test documents that the Marshal-error path is the
// effective guard and that the Unmarshal branch is defensive only.
func TestCovHarClone_UnmarshalErrorUnreachable(t *testing.T) {
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.Response.Error = func() {}
	// Returns nil via the Marshal-error branch (lines 27-29); the Unmarshal
	// branch (lines 32-34) is not reached.
	assert.Nil(t, h.Clone())
}

// Cover isNilWriter non-pointer-value-kind branch (lines 175-176).
// A concrete value type (struct) implementing io.Writer has reflect.Kind
// Struct, which is not in the chan/func/interface/map/pointer/slice set,
// so isNilWriter returns false.
type valueWriter struct{}

func (valueWriter) Write(p []byte) (int, error) { return len(p), nil }

func TestCovIsNilWriter_ValueKind(t *testing.T) {
	// concrete value type -> false (lines 175-176)
	assert.False(t, isNilWriter(valueWriter{}))

	// nil interface -> true (lines 167-169)
	assert.True(t, isNilWriter(nil))

	// typed-nil pointer -> true (lines 173-175)
	var buf *bytes.Buffer
	assert.True(t, isNilWriter(buf))

	// non-nil pointer -> false
	assert.False(t, isNilWriter(&bytes.Buffer{}))
}
