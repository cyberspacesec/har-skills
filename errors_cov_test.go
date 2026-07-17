package har

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Cover HarError.Unwrap nil-receiver branch (lines 79-81).
func TestCovHarErrorUnwrap_NilReceiver(t *testing.T) {
	var e *HarError
	// nil receiver -> return nil (lines 79-81)
	assert.Nil(t, e.Unwrap())
}

// Cover HarError.Unwrap non-nil with wrapped error (line 82).
func TestCovHarErrorUnwrap_WithErr(t *testing.T) {
	base := errors.New("root cause")
	e := NewFileSystemError("fs failure", base)
	// errors.Is reaches the wrapped error via Unwrap.
	assert.True(t, errors.Is(e, base))
}

// Cover HarError.Unwrap non-nil with no wrapped error (line 82 returns nil).
func TestCovHarErrorUnwrap_NoErr(t *testing.T) {
	e := NewInvalidFormatError("bad format")
	assert.Nil(t, e.Unwrap())
}
