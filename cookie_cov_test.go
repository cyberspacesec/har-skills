package har

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Cover isHTTPS error branch (lines 259-261): url.Parse fails on a
// control-character-laden string, causing err != nil and return false.
func TestCovIsHTTPS_ParseError(t *testing.T) {
	// A string with a raw control character makes url.Parse return an error.
	bad := "https://example.com\x7f"
	assert.False(t, isHTTPS(bad))
}

// Cover isHTTPS success branches.
func TestCovIsHTTPS_Schemes(t *testing.T) {
	assert.True(t, isHTTPS("https://example.com/path"))
	assert.False(t, isHTTPS("http://example.com/path"))
	assert.False(t, isHTTPS("ftp://example.com"))
	// Empty string parses fine with empty scheme -> not https.
	assert.False(t, isHTTPS(""))
}
