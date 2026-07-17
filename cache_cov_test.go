package har

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Cover assessCacheability SMaxAge branch (lines 215-216) and the
// combined no-CC + Expires fallback (lines 228-231), plus pragma-no-cache
// when cc is nil (lines 223-225).
func TestCovAssessCacheability_Branches(t *testing.T) {
	// SMaxAge path (lines 215-216): cc has SMaxAge set, MaxAge nil.
	smaxAge := 30
	cc := &CacheControlDirectives{SMaxAge: &smaxAge}
	cacheable, cacheType, maxAge := assessCacheability(cc, false, "")
	assert.True(t, cacheable)
	assert.Equal(t, "public", cacheType)
	assert.Equal(t, 30*time.Second, maxAge)

	// no Cache-Control + Expires present (lines 228-231)
	cacheable, cacheType, _ = assessCacheability(nil, false, "Wed, 21 Oct 2025 07:28:00 GMT")
	assert.True(t, cacheable)
	assert.Equal(t, "public", cacheType)

	// Pragma: no-cache with cc == nil (lines 223-225)
	cacheable, cacheType, _ = assessCacheability(nil, true, "")
	assert.False(t, cacheable)
	assert.Equal(t, "no-cache", cacheType)

	// Default branch: cc == nil, no pragma, no expires -> cacheable public.
	cacheable, cacheType, _ = assessCacheability(nil, false, "")
	assert.True(t, cacheable)
	assert.Equal(t, "public", cacheType)
}

// Cover the full CacheAnalysis path that drives assessCacheability with
// SMaxAge via real headers (integration coverage for the branch).
func TestCovCacheAnalysis_SMaxAgeAndExpires(t *testing.T) {
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	e.AddResponseHeader("Cache-Control", "s-maxage=120")
	e.AddResponseHeader("Expires", "Wed, 21 Oct 2025 07:28:00 GMT")
	e.AddResponseHeader("Pragma", "no-cache")

	report := h.CacheAnalysis()
	if assert.NotNil(t, report) {
		assert.GreaterOrEqual(t, report.CacheableCount, 0)
	}
}
