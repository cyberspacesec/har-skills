package har

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Cover ParseHarWithLazyLoading JSON-unmarshal-error branch (lines 164-166):
// valid JSON content (passes validateInput) but not unmarshalable into LazyHar
// struct because the "log.entries" field is the wrong type.
func TestCovParseHarWithLazyLoading_UnmarshalError(t *testing.T) {
	// This is valid JSON (passes isJSONContent) but "entries" being a
	// string instead of an array makes json.Unmarshal fail.
	bad := []byte(`{"log":{"version":"1.2","creator":{"name":"x","version":"1"},"pages":[],"entries":"not-an-array"}}`)
	lh, err := ParseHarWithLazyLoading(bad)
	if err == nil {
		t.Fatal("expected error from invalid entries type, got nil")
	}
	assert.Nil(t, lh)
	assertHarErrorCode(t, err, ErrCodeJSONParse)
}

// Cover ParseHarWithLazyLoading validateInput-empty branch (line 158-160)
// and happy path.
func TestCovParseHarWithLazyLoading_EmptyAndHappy(t *testing.T) {
	// empty input -> validateInput error
	lh, err := ParseHarWithLazyLoading([]byte{})
	if err == nil {
		t.Fatal("expected error for empty input, got nil")
	}
	assert.Nil(t, lh)
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)

	// not JSON
	lh, err = ParseHarWithLazyLoading([]byte("plain text not json"))
	if err == nil {
		t.Fatal("expected error for non-JSON input, got nil")
	}
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)

	// happy path: minimal valid HAR
	good := []byte(`{"log":{"version":"1.2","creator":{"name":"x","version":"1"},"pages":[],"entries":[]}}`)
	lh, err = ParseHarWithLazyLoading(good)
	if err != nil {
		t.Fatalf("unexpected error on happy path: %v", err)
	}
	if assert.NotNil(t, lh) {
		assert.Equal(t, "1.2", lh.Log.Version)
	}
}
