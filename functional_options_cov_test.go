package har

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Cover isNilReplayTransport default-kind (non-nil concrete) branch
// (lines 176-177): a non-nil concrete value whose reflect.Kind is Struct
// (not in the chan/func/interface/map/pointer/slice set) falls through to
// `return false`. We use a custom value-type RoundTripper because
// http.Transport's RoundTrip has a pointer receiver, so the value type
// does not implement http.RoundTripper.
type valueRoundTripper struct{}

func (valueRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, nil
}

func TestCovIsNilReplayTransport_NonNilValueKind(t *testing.T) {
	var rt valueRoundTripper // value type, Kind Struct, non-nil
	assert.False(t, isNilReplayTransport(rt))
}

// Cover isNilReplayTransport nil-interface branch (lines 168-169).
func TestCovIsNilReplayTransport_NilInterface(t *testing.T) {
	var rt http.RoundTripper // nil interface
	assert.True(t, isNilReplayTransport(rt))
}

// Cover isNilReplayTransport typed-nil pointer branch (lines 173-175):
// a (*http.Transport)(nil) has Kind Pointer and IsNil()==true.
func TestCovIsNilReplayTransport_TypedNilPointer(t *testing.T) {
	var p *http.Transport // typed nil pointer
	assert.True(t, isNilReplayTransport(p))
}

// Cover WithReplayTransport not setting Transport for a typed-nil pointer
// (integration path through the functional option).
func TestCovWithReplayTransport_TypedNilIgnored(t *testing.T) {
	var p *http.Transport
	opts := NewReplayOptions(WithReplayTransport(p))
	assert.Nil(t, opts.Transport)

	// A real transport is accepted.
	rt := &http.Transport{}
	opts2 := NewReplayOptions(WithReplayTransport(rt))
	if assert.NotNil(t, opts2.Transport) {
		assert.Same(t, rt, opts2.Transport)
	}
}
