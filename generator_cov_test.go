package har

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Cover Har.SetVersion nil-receiver branch (lines 39-40, returns nil).
func TestCovSetVersion_NilReceiver(t *testing.T) {
	var h *Har
	// Should not panic; returns nil.
	result := h.SetVersion("1.3")
	assert.Nil(t, result)
}

// Cover Har.SetVersion normal path (lines 42-43).
func TestCovSetVersion_HappyPath(t *testing.T) {
	h := NewHar()
	got := h.SetVersion("1.3")
	if !assert.Same(t, h, got) {
		t.Fatal("SetVersion should return the same receiver")
	}
	assert.Equal(t, "1.3", h.Log.Version)
}

// Cover Entries.SetConnection nil-receiver branch (lines 279-280).
func TestCovSetConnection_NilReceiver(t *testing.T) {
	var e *Entries
	result := e.SetConnection("conn-1")
	assert.Nil(t, result)
}

// Cover Entries.SetConnection normal path (lines 281-282).
func TestCovSetConnection_HappyPath(t *testing.T) {
	e := Entries{}
	got := e.SetConnection("conn-1")
	if !assert.Same(t, &e, got) {
		t.Fatal("SetConnection should return the same receiver")
	}
	assert.Equal(t, "conn-1", e.Connection)
}

// Cover Entries.SetPageref nil-receiver branch (lines 288-289).
func TestCovSetPageref_NilReceiver(t *testing.T) {
	var e *Entries
	result := e.SetPageref("page_1")
	assert.Nil(t, result)
}

// Cover Entries.SetPageref normal path (lines 290-291).
func TestCovSetPageref_HappyPath(t *testing.T) {
	e := Entries{}
	got := e.SetPageref("page_1")
	if !assert.Same(t, &e, got) {
		t.Fatal("SetPageref should return the same receiver")
	}
	assert.Equal(t, "page_1", e.Pageref)
}

// Cover Har.SaveToFile write-error branch (lines 326-328).
// Writing to a path inside a non-existent directory makes os.WriteFile fail,
// which is wrapped into ErrCodeFileSystem.
func TestCovHarSaveToFile_WriteError(t *testing.T) {
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")

	badPath := "/nonexistent_dir_xyz_cov/deep/path/out.har"
	err := h.SaveToFile(badPath, true)
	if err == nil {
		t.Fatal("expected error writing to non-existent path, got nil")
	}
	assertHarErrorCode(t, err, ErrCodeFileSystem)
}

// Cover Har.SaveToFile ToJSON-error branch (lines 323-325): inject a func
// value into Response.Error (any, json:"_error") to make ToJSON fail;
// SaveToFile propagates the error before reaching os.WriteFile.
func TestCovHarSaveToFile_ToJSONError(t *testing.T) {
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.Response.Error = func() {} // unsupported by json.Marshal

	err := h.SaveToFile(t.TempDir()+"/should-not-exist.har", true)
	if err == nil {
		t.Fatal("expected error from ToJSON failure, got nil")
	}
	assertHarErrorCode(t, err, ErrCodeJSONParse)
}

// Cover Har.SaveToFile nil-receiver branch (lines 318-320) and happy path.
func TestCovHarSaveToFile_NilAndHappy(t *testing.T) {
	var h *Har
	err := h.SaveToFile("/tmp/whatever_cov.har", true)
	if err == nil {
		t.Fatal("expected error for nil HAR, got nil")
	}
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)

	// happy path
	hh := NewHar()
	ee := hh.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	ee.SetResponseStatus(200, "OK")
	path := t.TempDir() + "/out.har"
	if err := hh.SaveToFile(path, false); err != nil {
		t.Fatalf("unexpected error on happy path: %v", err)
	}
	assert.FileExists(t, path)
}
