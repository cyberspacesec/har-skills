package har

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// alwaysFailWriter always returns an error on Write.
type alwaysFailWriter struct{ err error }

func (w *alwaysFailWriter) Write(p []byte) (int, error) { return 0, w.err }

// Cover writeCSVToWriter header-write error branch (lines 105-107).
//
// csv.Writer.Write buffers via a bufio.Writer (default 4096 bytes). To make
// the per-Write call itself return a non-nil error (rather than deferring
// the error to Flush), we pass a huge custom header (>4096 bytes via
// options.Headers) so that bufio flushes immediately during Write(headers)
// and surfaces the underlying writer's error. The writer always fails.
func TestCovWriteCSVToWriter_HeaderWriteError(t *testing.T) {
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")

	bigHeader := strings.Repeat("h", 5000) // > bufio buffer -> immediate flush
	opts := ConvertOptions{Headers: []string{bigHeader}, IncludeURL: true}

	w := &alwaysFailWriter{err: errors.New("header write boom")}
	err := writeCSVToWriter(w, h.Log.Entries, opts)
	if err == nil {
		t.Fatal("expected error from writer.Write(headers), got nil")
	}
	assertHarErrorCode(t, err, ErrCodeFileSystem)
}

// Cover writeCSVToWriter data-row write error branch (lines 112-114).
//
// A huge URL field (>4096 bytes) forces bufio to flush during
// writer.Write(row), surfacing the always-failing writer's error.
func TestCovWriteCSVToWriter_DataRowWriteError(t *testing.T) {
	h := NewHar()
	e := h.AddEntry("POST", "https://example.com/", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	// Make the URL field huge so writer.Write(row) flushes immediately.
	e.Request.URL = "https://example.com/" + strings.Repeat("x", 5000)

	w := &alwaysFailWriter{err: errors.New("row write boom")}
	err := writeCSVToWriter(w, h.Log.Entries, DefaultConvertOptions())
	if err == nil {
		t.Fatal("expected error from writer.Write(row), got nil")
	}
	assertHarErrorCode(t, err, ErrCodeFileSystem)
}

// Cover writeCSVToWriter Flush error branch (lines 118-120): small rows
// that stay buffered, so the underlying error only surfaces on Flush via
// writer.Error().
func TestCovWriteCSVToWriter_FlushError(t *testing.T) {
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")

	w := &alwaysFailWriter{err: errors.New("flush boom")}
	err := writeCSVToWriter(w, h.Log.Entries, DefaultConvertOptions())
	if err == nil {
		t.Fatal("expected error from Flush/ writer.Error(), got nil")
	}
	assertHarErrorCode(t, err, ErrCodeFileSystem)
}

// Cover writeCSVToWriter nil-writer branch (lines 97-99 via isNilWriter).
func TestCovWriteCSVToWriter_NilWriter(t *testing.T) {
	err := writeCSVToWriter(nil, nil, DefaultConvertOptions())
	if err == nil {
		t.Fatal("expected error for nil writer, got nil")
	}
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
}

// Cover writeCSVToWriter happy path (returns nil) and convertToCSV happy
// path (lines 89-93 success side).
//
// Note on lines 89-91 (convertToCSV error branch): convertToCSV uses an
// internal bytes.Buffer whose Write never fails, so writeCSVToWriter can
// only fail via isNilWriter (buf is non-nil) or csv errors (buf.Write never
// fails, so csv.Writer.Write never errors). That branch is therefore
// unreachable by black-box testing; the success side of lines 89/93 is
// covered here.
func TestCovConvertToCSV_HappyPath(t *testing.T) {
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")

	out, err := h.Convert(FormatCSV, DefaultConvertOptions())
	if err != nil {
		t.Fatalf("Convert(FormatCSV) unexpected error: %v", err)
	}
	if out == "" {
		t.Fatal("expected non-empty CSV output")
	}

	// Also exercise the other convert formats for branch breadth.
	for _, fmtName := range []ConvertFormat{FormatMarkdown, FormatHTML, FormatText} {
		if _, err := h.Convert(fmtName, DefaultConvertOptions()); err != nil {
			t.Fatalf("Convert(%s) unexpected error: %v", fmtName, err)
		}
	}
}

// Smoke-check helpers are wired.
func TestCovConverter_Helpers(t *testing.T) {
	assert.NotNil(t, DefaultConvertOptions())
	// ensure alwaysFailWriter satisfies io.Writer.
	var _ io.Writer = &alwaysFailWriter{}
}
