package har

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Cover Recorder.EntryCount nil-builder path (lines 537-539):
// a nil *Recorder -> ensureBuilder returns nil -> har == nil -> return 0.
func TestCovRecorderEntryCount_NilReceiver(t *testing.T) {
	var r *Recorder
	assert.Equal(t, 0, r.EntryCount())
}

// Cover Recorder.EntryCount normal path (lines 541: return len(entries)).
func TestCovRecorderEntryCount_HappyPath(t *testing.T) {
	r := NewRecorder()
	assert.Equal(t, 0, r.EntryCount())
	r.CaptureEntry(Entries{Request: Request{Method: "GET", URL: "https://example.com"}})
	r.CaptureEntry(Entries{Request: Request{Method: "POST", URL: "https://example.com/api"}})
	assert.Equal(t, 2, r.EntryCount())
}

// Cover WriteToWriter ToJSON-error path (lines 580-582), nil-writer path
// (lines 575-577) and nil-har path (lines 572-574).
func TestCovWriteToWriter_Branches(t *testing.T) {
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")

	// nil har branch (lines 572-574)
	err := WriteToWriter(nil, &bytes.Buffer{}, false)
	if err == nil {
		t.Fatal("expected error for nil har, got nil")
	}
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)

	// nil writer branch (lines 575-577)
	err = WriteToWriter(h, nil, false)
	if err == nil {
		t.Fatal("expected error for nil writer, got nil")
	}
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)

	// ToJSON-error branch (lines 580-582): inject an un-marshalable value
	// into Response.Error (any, json:"_error,omitempty") to force
	// json.MarshalIndent to fail.
	he := NewHar()
	ee := he.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	ee.Response.Error = func() {} // unsupported type for json.Marshal
	err = WriteToWriter(he, &bytes.Buffer{}, false)
	if err == nil {
		t.Fatal("expected error from ToJSON failure, got nil")
	}
	assertHarErrorCode(t, err, ErrCodeJSONParse)

	// success path (covers lines 579, 584)
	var buf bytes.Buffer
	err = WriteToWriter(h, &buf, true)
	if err != nil {
		t.Fatalf("unexpected error on success path: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected non-empty output on success path")
	}
}

// failingWriter is referenced here to ensure the type is available in this
// test file's scope (defined in builder_test.go in the same package).
func TestCovBuilder_FailingWriterRef(t *testing.T) {
	w := &failingWriter{err: errors.New("boom")}
	_, err := w.Write([]byte("x"))
	assert.Error(t, err)
}
