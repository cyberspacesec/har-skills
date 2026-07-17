package har

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Cover anonymizeIP branches: IPv4 (lines 284-289), invalid IP (line 281),
// and the IPv6 last-segment replacement (lines 293-297). Line 299 (return ip
// when no colon) is a defensive fallthrough that is unreachable in practice
// because net.ParseIP'd IPv6 strings always contain a colon; we exercise
// the reachable IPv6 path here.
func TestCovAnonymizeIP_Branches(t *testing.T) {
	// Invalid IP string -> returns original (line 281).
	assert.Equal(t, "not-an-ip", anonymizeIP("not-an-ip"))

	// IPv4 -> last octet zeroed (lines 284-289).
	assert.Equal(t, "192.168.1.0", anonymizeIP("192.168.1.55"))

	// IPv6 -> last hextet zeroed (lines 293-297).
	anon := anonymizeIP("2001:db8::1")
	assert.Contains(t, anon, "2001:db8:")
	assert.True(t, anon == "2001:db8::0" || anon == "2001:db8:0:0:0:0:0:0", "got %q", anon)

	// IPv4 in IPv6 form still maps to To4() != nil path -> but the input
	// string is not dotted-decimal so the Split(".")==4 check fails and we
	// fall through to the IPv6 branch.  "::ffff:192.168.1.55" -> To4()!=nil
	// but has no dotted form in the original string -> falls to IPv6 branch.
	anon2 := anonymizeIP("::ffff:192.168.1.55")
	assert.NotEqual(t, "::ffff:192.168.1.55", anon2)
}

// Cover redactQueryStringSimple non-CustomRedactor replacement branch
// (line 390) and the CustomRedactor branch (line 388). Line 392 (return
// match) is a defensive fallthrough when FindStringSubmatch returns < 3
// parts, which cannot happen for a successful regex match; we cover the
// reachable replacement branches here.
func TestCovRedactQueryStringSimple_Branches(t *testing.T) {
	// Non-CustomRedactor path -> replacement text applied (line 390).
	out := redactQueryStringSimple(
		"https://bad-url/path?token=secret&keep=1",
		[]string{"token"},
		"[REDACTED]",
		RedactOptions{},
		nil, // valueRes
	)
	assert.Contains(t, out, "token=[REDACTED]")
	assert.Contains(t, out, "keep=1")

	// CustomRedactor path (line 388).
	out = redactQueryStringSimple(
		"https://bad-url/path?token=secret",
		[]string{"token"},
		"[REDACTED]",
		RedactOptions{
			CustomRedactor: func(fieldType, name, value string) string {
				return "[CUSTOM:" + name + "]"
			},
		},
		nil, // valueRes
	)
	assert.Contains(t, out, "token=[CUSTOM:token]")
}
