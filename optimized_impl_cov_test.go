package har

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Cover OptimizedHar.GetBrowser nil-receiver branch (lines 25-27).
func TestCovOptimizedHarGetBrowser_NilReceiver(t *testing.T) {
	var h *OptimizedHar
	b := h.GetBrowser()
	assert.Equal(t, Browser{}, b)
}

// Cover OptimizedHar.GetBrowser happy path (line 28).
func TestCovOptimizedHarGetBrowser_HappyPath(t *testing.T) {
	h := &OptimizedHar{}
	h.Log.Browser = Browser{Name: "Chrome", Version: "120"}
	b := h.GetBrowser()
	assert.Equal(t, "Chrome", b.Name)
	assert.Equal(t, "120", b.Version)
}

// Smoke-cover the other nil-receiver getters on OptimizedHar that share the
// same defensive pattern, to keep the file well-exercised.
func TestCovOptimizedHar_NilReceivers(t *testing.T) {
	var h *OptimizedHar
	assert.Equal(t, "", h.GetVersion())
	assert.Equal(t, Creator{}, h.GetCreator())
	assert.Nil(t, h.GetEntries())
	assert.Nil(t, h.GetPages())
	assert.Nil(t, h.ToStandard())
}
