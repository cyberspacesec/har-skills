package har

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"math"
	"os"
	"path/filepath"
	"testing"
)

// This file adds coverage-focused tests (prefixed "Cov") for reader.go.
//
// It targets the previously-uncovered branches in:
//   - ParseHarFromReader / ParseHarFromReaderWithOptions (non-*HarError fallback)
//   - ParseHarFileGzipped (parse failure, gzip close error paths)
//   - SaveToFileGzipped (ToJSON error path)
//   - writeGzippedDataToFile (nil file, gzip writer close failure)
//   - detectGzipMagicBytes (read error, close error paths)
//
// Several branches in reader.go are structurally unreachable without
// modifying source:
//   - lines 28, 48: ParseHar/ParseHarWithOptions only ever return *HarError,
//     so the "else" WrapJSONUnmarshalError branch can never execute.
//   - lines 75-77, 235-237: the deferred file.Close() failure when err==nil
//     cannot be triggered because os.File.Close() does not fail on a healthy fd.
//   - lines 86-89, 99-101: gzip.Reader.Close() never returns an error after a
//     successful read, so these close-failure branches are dead code.
//   - lines 201-203: the deferred gzWriter.Close() branch requires gzipClosed==false
//     AND err==nil, which cannot occur given the control flow.
//
// The tests below cover everything that is reachable from the public and
// internal (same-package) API.

// ---------- ParseHarFileGzipped: ParseHarFromReader failure (lines 94-96) ----------

// CovParseHarFileGzipped_InvalidJSONContent covers the branch where the file
// is a valid gzip stream but its decompressed content is not valid HAR JSON.
// gzip.NewReader succeeds, but ParseHarFromReader fails -> lines 94-96.
func TestCovParseHarFileGzipped_InvalidJSONContent(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "invalid.har.gz")

	f, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	gzWriter := gzip.NewWriter(f)
	if _, err := gzWriter.Write([]byte("not valid json at all")); err != nil {
		t.Fatalf("failed to write gzipped data: %v", err)
	}
	if err := gzWriter.Close(); err != nil {
		t.Fatalf("failed to close gzip writer: %v", err)
	}
	f.Close()

	_, err = ParseHarFileGzipped(filePath)
	if err == nil {
		t.Fatal("expected error for gzipped file with invalid JSON content, got nil")
	}
}

// CovParseHarFileGzipped_ValidJSONButInvalidHAR covers the branch where the
// decompressed content is valid JSON but fails HAR validation, still
// exercising the ParseHarFromReader failure path (lines 94-96).
func TestCovParseHarFileGzipped_ValidJSONButInvalidHAR(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "invalid_har.har.gz")

	f, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	gzWriter := gzip.NewWriter(f)
	// Valid JSON object but missing required HAR fields -> validation error.
	if _, err := gzWriter.Write([]byte(`{"log":{}}`)); err != nil {
		t.Fatalf("failed to write gzipped data: %v", err)
	}
	if err := gzWriter.Close(); err != nil {
		t.Fatalf("failed to close gzip writer: %v", err)
	}
	f.Close()

	_, err = ParseHarFileGzipped(filePath)
	if err == nil {
		t.Fatal("expected error for gzipped file with invalid HAR content, got nil")
	}
}

// ---------- SaveToFileGzipped: ToJSON error path (lines 180-182) ----------

// CovSaveToFileGzipped_ToJSONError covers the ToJSON failure branch by
// constructing a Har whose Response.Error field is set to math.NaN(), which
// json.Marshal cannot encode. This makes Har.ToJSON fail -> lines 180-182.
func TestCovSaveToFileGzipped_ToJSONError(t *testing.T) {
	h, err := ParseHar([]byte(testHarJSON))
	if err != nil {
		t.Fatalf("ParseHar returned error: %v", err)
	}

	// Poison the response error with NaN, which json.Marshal rejects with
	// "json: unsupported value: NaN".
	if len(h.Log.Entries) > 0 {
		h.Log.Entries[0].Response.Error = math.NaN()
	}

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "nan.har.gz")

	err = SaveToFileGzipped(h, filePath, true)
	if err == nil {
		t.Fatal("expected error for HAR with unmarshalable field, got nil")
	}
}

// ---------- writeGzippedDataToFile: nil file (lines 192-194) ----------

// CovWriteGzippedDataToFile_NilFile covers the file==nil guard -> lines 192-194.
func TestCovWriteGzippedDataToFile_NilFile(t *testing.T) {
	err := writeGzippedDataToFile(nil, "whatever", []byte("{}"))
	if err == nil {
		t.Fatal("expected error for nil file, got nil")
	}
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
}

// ---------- writeGzippedDataToFile: explicit gzWriter.Close failure (lines 215-217) ----------

// failingOnTrailerWriter is an io.WriteCloser that accepts writes up to a
// byte limit, then returns an error. The limit is chosen so that the gzip
// header + compressed payload fit (so gzWriter.Write succeeds), but the gzip
// trailer flush during gzWriter.Close() overflows the limit and fails.
type failingOnTrailerWriter struct {
	buf     bytes.Buffer
	written int
	limit   int
}

func (w *failingOnTrailerWriter) Write(p []byte) (int, error) {
	if w.written+len(p) > w.limit {
		remain := w.limit - w.written
		if remain > 0 {
			n, _ := w.buf.Write(p[:remain])
			w.written += n
		}
		return 0, errors.New("simulated write failure on trailer flush")
	}
	n, err := w.buf.Write(p)
	w.written += n
	return n, err
}

func (w *failingOnTrailerWriter) Close() error { return nil }

// CovWriteGzippedDataToFile_GzipCloseFailure covers the explicit
// gzWriter.Close() error branch -> lines 215-217.
func TestCovWriteGzippedDataToFile_GzipCloseFailure(t *testing.T) {
	data := []byte(testHarJSON)
	// gzip header (10 bytes) + compressed data (~tens of bytes) + trailer (8 bytes).
	// A limit of 25 lets the header+payload through but fails the trailer flush
	// during Close (confirmed empirically with the stdlib gzip writer).
	fw := &failingOnTrailerWriter{limit: 25}

	err := writeGzippedDataToFile(fw, "simulated.gz", data)
	if err == nil {
		t.Fatal("expected error from gzWriter.Close failure, got nil")
	}
	assertHarErrorCode(t, err, ErrCodeFileSystem)
}

// ---------- writeGzippedDataToFile: gzWriter.Write failure (lines 210-212) ----------
//
// Already covered by existing tests indirectly, but we add an explicit case
// for completeness using a writer that fails immediately.

// immediateFailWriter fails every Write call.
type immediateFailWriter struct{}

func (immediateFailWriter) Write(p []byte) (int, error) {
	return 0, errors.New("simulated immediate write failure")
}

func (immediateFailWriter) Close() error { return nil }

// CovWriteGzippedDataToFile_WriteFailure covers the gzWriter.Write error
// branch -> lines 210-212.
func TestCovWriteGzippedDataToFile_WriteFailure(t *testing.T) {
	err := writeGzippedDataToFile(immediateFailWriter{}, "simulated.gz", []byte(testHarJSON))
	if err == nil {
		t.Fatal("expected error from gzWriter.Write failure, got nil")
	}
	assertHarErrorCode(t, err, ErrCodeFileSystem)
}

// ---------- writeGzippedDataToFile: deferred file.Close failure (lines 205-207) ----------
//
// The deferred file.Close() failure (when err==nil) is reachable via a
// WriteCloser whose Close returns an error after a successful gzip write.

// closeFailWriter accepts all writes but Close returns an error.
type closeFailWriter struct {
	buf bytes.Buffer
}

func (w *closeFailWriter) Write(p []byte) (int, error) { return w.buf.Write(p) }
func (w *closeFailWriter) Close() error                { return errors.New("simulated close failure") }

// CovWriteGzippedDataToFile_FileCloseFailure covers the deferred file.Close()
// error branch -> lines 205-207.
func TestCovWriteGzippedDataToFile_FileCloseFailure(t *testing.T) {
	err := writeGzippedDataToFile(&closeFailWriter{}, "simulated.gz", []byte(testHarJSON))
	if err == nil {
		t.Fatal("expected error from deferred file.Close failure, got nil")
	}
	assertHarErrorCode(t, err, ErrCodeFileSystem)
}

// ---------- detectGzipMagicBytes: read error (lines 242-244) ----------

// CovDetectGzipMagicBytes_ReadErrorOnDirectory covers the file.Read error
// branch (non-EOF). Opening a directory succeeds, but reading it returns an
// "is a directory" error which is not io.EOF -> lines 242-244.
func TestCovDetectGzipMagicBytes_ReadErrorOnDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	isGz, err := detectGzipMagicBytes(tmpDir)
	if err == nil {
		t.Fatal("expected error when reading a directory, got nil")
	}
	assertHarErrorCode(t, err, ErrCodeFileSystem)
	if isGz {
		t.Error("expected isGz=false when reading a directory")
	}
}

// ---------- detectGzipMagicBytes: boundary lengths ----------

// CovDetectGzipMagicBytes_ShortInputs covers the n<2 boundary conditions
// (1-byte and 2-byte non-magic files) to ensure the magic-byte comparison
// is exercised without panicking.
func TestCovDetectGzipMagicBytes_ShortInputs(t *testing.T) {
	tmpDir := t.TempDir()

	cases := []struct {
		name    string
		content []byte
		wantGz  bool
	}{
		{"one_byte_no_magic", []byte{0x00}, false},
		{"two_bytes_no_magic", []byte{0x1f, 0x00}, false},
		{"two_bytes_first_only", []byte{0x1f, 0x00}, false},
		{"two_bytes_second_only", []byte{0x00, 0x8b}, false},
		{"exact_magic", []byte{0x1f, 0x8b}, true},
		{"exact_magic_plus_extra", []byte{0x1f, 0x8b, 0x08}, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := filepath.Join(tmpDir, tc.name)
			if err := os.WriteFile(p, tc.content, 0644); err != nil {
				t.Fatalf("failed to write temp file: %v", err)
			}
			isGz, err := detectGzipMagicBytes(p)
			if err != nil {
				t.Fatalf("detectGzipMagicBytes returned error: %v", err)
			}
			if isGz != tc.wantGz {
				t.Errorf("expected isGz=%v, got %v", tc.wantGz, isGz)
			}
		})
	}
}

// ---------- ParseHarFromReader / ParseHarFromReaderWithOptions: non-*HarError fallback ----------
//
// Lines 28 and 48 expect ParseHar / ParseHarWithOptions to return a non-*HarError.
// In the current implementation both functions only ever return *HarError
// (every error constructor returns *HarError), so these branches are dead code.
// We assert this invariant so the test documents the actual behavior: if a
// future change makes ParseHar return a plain error, the type assertion will
// still pass through and these tests will exercise lines 28/48.

// CovParseHarFromReader_AlwaysHarError documents that every ParseHar error is
// a *HarError, confirming line 28 is currently unreachable.
func TestCovParseHarFromReader_AlwaysHarError(t *testing.T) {
	inputs := []string{
		"",
		"not json",
		`{"log":{}}`,
		`{"log":{"version":123,"creator":"not-an-object"}}`,
	}
	for _, in := range inputs {
		_, err := ParseHarFromReader(stringReader(in))
		if err == nil {
			t.Errorf("expected error for input %q, got nil", in)
			continue
		}
		if _, ok := err.(*HarError); !ok {
			// If this ever fires, line 28 becomes reachable and this test
			// should be updated to assert the WrapJSONUnmarshalError path.
			t.Errorf("expected *HarError for input %q, got %T (this means line 28 is now reachable)", in, err)
		}
	}
}

// CovParseHarFromReaderWithOptions_AlwaysHarError documents that every
// ParseHarWithOptions error is a *HarError, confirming line 48 is unreachable.
func TestCovParseHarFromReaderWithOptions_AlwaysHarError(t *testing.T) {
	// With SkipValidation:true, {"log":{}} parses successfully (no error), so
	// it is excluded here. The remaining inputs all produce errors.
	inputs := []string{
		"",
		"not json",
		`{"log":{"version":123,"creator":"not-an-object"}}`,
	}
	for _, in := range inputs {
		_, err := ParseHarFromReaderWithOptions(stringReader(in), ParseOptions{SkipValidation: true})
		if err == nil {
			t.Errorf("expected error for input %q, got nil", in)
			continue
		}
		if _, ok := err.(*HarError); !ok {
			t.Errorf("expected *HarError for input %q, got %T (this means line 48 is now reachable)", in, err)
		}
	}
}

// stringReader is a small helper to avoid importing strings in multiple
// places; it mirrors strings.NewReader.
func stringReader(s string) io.Reader {
	return bytes.NewReader([]byte(s))
}
