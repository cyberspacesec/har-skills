package har

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Cover NewStreamingParser NewStreamingHarFromBytes-error branch
// (lines 128-130): valid JSON (passes validateInput) but the structure
// cannot unmarshal into Har (log.entries is the wrong type), so
// NewStreamingHarFromBytes fails and the error is propagated.
func TestCovNewStreamingParser_UnmarshalError(t *testing.T) {
	// Valid JSON object, but "log" is a string instead of an object ->
	// json.Unmarshal into Har fails.
	bad := []byte(`{"log":"not-an-object"}`)
	it, err := NewStreamingParser(bad)
	if err == nil {
		t.Fatal("expected error from NewStreamingHarFromBytes, got nil")
	}
	assert.Nil(t, it)
	// The error is wrapped by WrapJSONUnmarshalError -> ErrCodeJSONParse.
	assertHarErrorCode(t, err, ErrCodeJSONParse)
}

// Cover NewStreamingParser validateInput-empty branch (lines 122-124)
// and happy path.
func TestCovNewStreamingParser_EmptyAndHappy(t *testing.T) {
	// empty input
	it, err := NewStreamingParser([]byte{})
	if err == nil {
		t.Fatal("expected error for empty input, got nil")
	}
	assert.Nil(t, it)
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)

	// not JSON
	it, err = NewStreamingParser([]byte("not json at all"))
	if err == nil {
		t.Fatal("expected error for non-JSON input, got nil")
	}
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)

	// happy path
	good := []byte(`{"log":{"version":"1.2","creator":{"name":"x","version":"1"},"entries":[]}}`)
	it, err = NewStreamingParser(good)
	if err != nil {
		t.Fatalf("unexpected error on happy path: %v", err)
	}
	if assert.NotNil(t, it) {
		// Drain the iterator to be safe.
		for it.Next() {
			_ = it.Entry()
		}
	}
}
