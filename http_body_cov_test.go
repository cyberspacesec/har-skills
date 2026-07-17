package har

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Cover readAndCloseResponseBody nil-reader branch (lines 9-11).
func TestCovReadAndCloseResponseBody_NilReader(t *testing.T) {
	// A nil io.ReadCloser interface -> isNilReader true -> returns zeros.
	var rc io.ReadCloser
	body, readErr, closeErr := readAndCloseResponseBody(rc)
	assert.Nil(t, body)
	assert.Nil(t, readErr)
	assert.Nil(t, closeErr)
}

// Cover readAndCloseResponseBody happy path (lines 13-15) with a normal
// reader.
func TestCovReadAndCloseResponseBody_HappyPath(t *testing.T) {
	rc := io.NopCloser(newStrReader("hello body"))
	body, readErr, closeErr := readAndCloseResponseBody(rc)
	assert.Equal(t, "hello body", string(body))
	assert.Nil(t, readErr)
	assert.Nil(t, closeErr)
}

// Cover responseBodyErrorMessage read+close both-error branch (lines 23-25).
func TestCovResponseBodyErrorMessage_BothErrors(t *testing.T) {
	readErr := errors.New("read boom")
	closeErr := errors.New("close boom")
	msg := responseBodyErrorMessage(readErr, closeErr)
	if msg == "" {
		t.Fatal("expected non-empty message for both errors")
	}
	// Must mention both errors.
	assert.Contains(t, msg, "read boom")
	assert.Contains(t, msg, "close boom")
}

// Cover responseBodyErrorMessage read-only-error branch (lines 26-27).
func TestCovResponseBodyErrorMessage_ReadOnly(t *testing.T) {
	readErr := errors.New("read boom")
	msg := responseBodyErrorMessage(readErr, nil)
	assert.Contains(t, msg, "failed to read response body")
	assert.Contains(t, msg, "read boom")
	assert.NotContains(t, msg, "close")
}

// Cover responseBodyErrorMessage close-only-error branch (lines 28-29).
func TestCovResponseBodyErrorMessage_CloseOnly(t *testing.T) {
	closeErr := errors.New("close boom")
	msg := responseBodyErrorMessage(nil, closeErr)
	assert.Contains(t, msg, "failed to close response body")
	assert.Contains(t, msg, "close boom")
	assert.NotContains(t, msg, "read response")
}

// Cover responseBodyErrorMessage no-error branch (lines 19-21).
func TestCovResponseBodyErrorMessage_NoErrors(t *testing.T) {
	assert.Equal(t, "", responseBodyErrorMessage(nil, nil))
}

// Cover readAndCloseResponseBody with a reader whose Read and Close both
// fail, exercising the full triple return (lines 13-15) plus the
// downstream responseBodyErrorMessage both-error path end-to-end.
func TestCovReadAndCloseResponseBody_FailingReadAndClose(t *testing.T) {
	rc := &failingReadCloserCov{}
	_, readErr, closeErr := readAndCloseResponseBody(rc)
	// readErr/closeErr surface the underlying errors.
	assert.Error(t, readErr)
	assert.Error(t, closeErr)
	// body may be nil or partial; the important part is both errors set.
	msg := responseBodyErrorMessage(readErr, closeErr)
	assert.Contains(t, msg, "read failure")
	assert.Contains(t, msg, "close failure")
}

// failingReadCloserCov is a ReadCloser whose Read returns an error and whose
// Close also returns an error. (Renamed to avoid clashing with the
// failingReadCloser type defined in http_body_test.go.)
type failingReadCloserCov struct{}

func (f *failingReadCloserCov) Read(p []byte) (int, error) {
	return 0, errors.New("read failure")
}

func (f *failingReadCloserCov) Close() error {
	return errors.New("close failure")
}

// strReader wraps a string as a ReadCloser for happy-path tests.
type strReader struct {
	s   string
	pos int
}

func newStrReader(s string) *strReader { return &strReader{s: s} }

func (s *strReader) Read(p []byte) (int, error) {
	if s.pos >= len(s.s) {
		return 0, io.EOF
	}
	n := copy(p, s.s[s.pos:])
	s.pos += n
	return n, nil
}

func (s *strReader) Close() error { return nil }
