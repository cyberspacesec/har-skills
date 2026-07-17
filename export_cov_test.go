package har

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Cover Har.ToXML nil-receiver branch (lines 502-504) and happy path.
//
// Note on lines 569-571 (xml.MarshalIndent error branch): ToXML manually
// constructs a HARXML value whose fields are all concrete, marshalable
// types (string/float64/slices of structs). There is no way to inject an
// un-marshalable value through the public API, so that defensive branch
// is unreachable by black-box testing and is intentionally not covered.
func TestCovToXML_Branches(t *testing.T) {
	// nil receiver
	var h *Har
	out, err := h.ToXML()
	if err == nil {
		t.Fatal("expected error for nil HAR, got nil")
	}
	assert.Equal(t, "", out)
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)

	// happy path with headers, post data, and response content
	hh := NewHar()
	ee := hh.AddEntry("POST", "https://example.com/api", "HTTP/1.1", "")
	ee.SetResponseStatus(200, "OK")
	ee.SetResponseContent(11, "application/json")
	ee.Response.Content.Text = `{"ok":true}`
	ee.AddRequestHeader("Content-Type", "application/json")
	ee.AddResponseHeader("Content-Type", "application/json")
	ee.SetPostData("application/json", `{"k":"v"}`)

	out, err = hh.ToXML()
	if !assert.NoError(t, err) {
		return
	}
	if out == "" {
		t.Fatal("expected non-empty XML output")
	}
	assert.Contains(t, out, `<?xml`)
	assert.Contains(t, out, "<har>")
	assert.Contains(t, out, "example.com/api")
	assert.Contains(t, out, "postData")
}
