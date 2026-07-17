package har

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Cover Har.ToYAML ToJSON-error branch (lines 25-27): inject an
// un-marshalable value (func) into Response.Error (any, json:"_error")
// so that ToJSON fails and ToYAML propagates the error.
func TestCovToYAML_ToJSONError(t *testing.T) {
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.Response.Error = func() {} // unsupported by json.Marshal

	out, err := h.ToYAML()
	if err == nil {
		t.Fatal("expected error from ToJSON failure in ToYAML, got nil")
	}
	if out != "" {
		t.Fatalf("expected empty string on error, got %q", out)
	}
	assertHarErrorCode(t, err, ErrCodeJSONParse)
}

// Cover Har.ToYAML nil-receiver branch (line 19-21) and happy path.
func TestCovToYAML_NilAndHappy(t *testing.T) {
	// nil receiver
	var h *Har
	out, err := h.ToYAML()
	if err == nil {
		t.Fatal("expected error for nil HAR, got nil")
	}
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
	assert.Equal(t, "", out)

	// happy path
	hh := NewHar()
	ee := hh.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	ee.SetResponseStatus(200, "OK")
	out, err = hh.ToYAML()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == "" {
		t.Fatal("expected non-empty YAML output")
	}
}

// Cover ConvertTo error branch (lines 195-197): when the chosen format's
// underlying call returns an error. We trigger it by using FormatYAML with
// a HAR whose ToJSON fails (func in Response.Error), so ToYAML returns
// error and ConvertTo propagates it (line 195-197).
func TestCovConvertTo_ErrorBranch(t *testing.T) {
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.Response.Error = func() {}

	var buf bytes.Buffer
	err := h.ConvertTo(FormatYAML, &buf, DefaultConvertOptions())
	if err == nil {
		t.Fatal("expected error from ConvertTo (YAML/ToJSON failure), got nil")
	}
	assertHarErrorCode(t, err, ErrCodeJSONParse)
}

// Cover ConvertTo nil-har branch (line 171-173) and nil-writer branch
// (lines 174-176).
func TestCovConvertTo_NilBranches(t *testing.T) {
	var h *Har
	err := h.ConvertTo(FormatYAML, &bytes.Buffer{}, DefaultConvertOptions())
	if err == nil {
		t.Fatal("expected error for nil HAR, got nil")
	}
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)

	hh := NewHar()
	err = hh.ConvertTo(FormatYAML, nil, DefaultConvertOptions())
	if err == nil {
		t.Fatal("expected error for nil writer, got nil")
	}
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
}

// Cover ConvertTo default (JSON) branch (lines 186-192) and the
// writeAllToWriter success path (line 199).
func TestCovConvertTo_DefaultJSONBranch(t *testing.T) {
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")

	// Unknown format falls into default -> ToJSON + writeAllToWriter (line 191).
	var buf bytes.Buffer
	err := h.ConvertTo(ConvertFormat("unknown-format"), &buf, DefaultConvertOptions())
	if err != nil {
		t.Fatalf("unexpected error on default JSON branch: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected non-empty output on default JSON branch")
	}
}
