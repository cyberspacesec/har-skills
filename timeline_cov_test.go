package har

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Cover Waterfall earliest-baseTime branch (lines 90-93): the second
// entry starts earlier than the first, so baseTime is updated.
func TestCovWaterfall_EarliestBaseTime(t *testing.T) {
	h := NewHar()
	e1 := h.AddEntry("GET", "https://example.com/first", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e1.StartedDateTime = time.Date(2024, 1, 1, 12, 0, 5, 0, time.UTC) // later
	e1.Time = 100

	e2 := h.AddEntry("GET", "https://example.com/second", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")
	e2.StartedDateTime = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC) // earlier
	e2.Time = 50

	wf := h.Waterfall()
	if !assert.NotNil(t, wf) {
		return
	}
	assert.Len(t, wf, 2)
	// The earliest entry (e2, index 1) should have StartTime == 0.
	assert.Equal(t, time.Duration(0), wf[1].StartTime)
	// e1 (index 0) starts 5s after e2.
	assert.Equal(t, 5*time.Second, wf[0].StartTime)
}

// Cover Waterfall nil/empty branch (lines 83-85).
func TestCovWaterfall_NilEmpty(t *testing.T) {
	var h *Har
	assert.Nil(t, h.Waterfall())

	empty := NewHar()
	assert.Nil(t, empty.Waterfall())
}

// Cover isCriticalResource CSS, JS-without-async/defer, font-mime and
// font-extension branches (lines 179-203).
func TestCovIsCriticalResource_Branches(t *testing.T) {
	// CSS -> true (line 179-181)
	e := Entries{Response: Response{Content: Content{MimeType: "text/css"}}}
	assert.True(t, isCriticalResource(e))

	// JavaScript without async/defer -> true (lines 184-186)
	e = Entries{
		Request:  Request{URL: "https://example.com/app.js"},
		Response: Response{Content: Content{MimeType: "application/javascript"}},
	}
	assert.True(t, isCriticalResource(e))

	// JavaScript WITH async -> false (hasAsyncOrDefer true)
	e = Entries{
		Request: Request{
			URL:     "https://example.com/app.js",
			Headers: []Headers{{Name: "X-Script-Async", Value: "true"}},
		},
		Response: Response{Content: Content{MimeType: "application/javascript"}},
	}
	assert.False(t, isCriticalResource(e))

	// font/* MIME -> true (lines 189-193)
	e = Entries{Response: Response{Content: Content{MimeType: "font/woff2"}}}
	assert.True(t, isCriticalResource(e))

	// application/x-font-* MIME -> true
	e = Entries{Response: Response{Content: Content{MimeType: "application/x-font-ttf"}}}
	assert.True(t, isCriticalResource(e))

	// Font by extension (.woff2) -> true (lines 197-202)
	e = Entries{Request: Request{URL: "https://example.com/font.woff2"}}
	assert.True(t, isCriticalResource(e))

	// Font by extension (.eot)
	e = Entries{Request: Request{URL: "https://example.com/font.EOT"}}
	// URL is lowercased, so .EOT becomes .eot
	assert.True(t, isCriticalResource(e))

	// Non-critical (image) -> false (line 205)
	e = Entries{Request: Request{URL: "https://example.com/logo.png"}, Response: Response{Content: Content{MimeType: "image/png"}}}
	assert.False(t, isCriticalResource(e))
}

// Cover hasAsyncOrDefer CustomFields branches (lines 220-229): both the
// async-true and defer-true sub-branches, plus the non-bool / missing
// sub-branches.
func TestCovHasAsyncOrDefer_CustomFields(t *testing.T) {
	// async=true -> true (lines 221-225)
	e := Entries{CustomFields: CustomFields{"async": true}}
	assert.True(t, hasAsyncOrDefer(e))

	// defer=true -> true (lines 226-230)
	e = Entries{CustomFields: CustomFields{"defer": true}}
	assert.True(t, hasAsyncOrDefer(e))

	// async present but not bool -> false
	e = Entries{CustomFields: CustomFields{"async": "yes"}}
	assert.False(t, hasAsyncOrDefer(e))

	// defer present but false -> false
	e = Entries{CustomFields: CustomFields{"defer": false}}
	assert.False(t, hasAsyncOrDefer(e))

	// CustomFields nil + no headers -> false
	e = Entries{}
	assert.False(t, hasAsyncOrDefer(e))

	// Header-based async hint (lines 212-217)
	e = Entries{Request: Request{Headers: []Headers{{Name: "X-Script-Defer", Value: "1"}}}}
	assert.True(t, hasAsyncOrDefer(e))
}

// Cover SLACheck branches: invalid-regex skip (lines 250-254), method
// filter (lines 258-260), URL-pattern filter (lines 263-265), passed
// (line 269) and overshoot (lines 271-273).
func TestCovSLACheck_Branches(t *testing.T) {
	h := NewHar()
	e1 := h.AddEntry("GET", "https://api.example.com/v1/users", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e1.Time = 500 // 500ms

	e2 := h.AddEntry("POST", "https://api.example.com/v1/login", "HTTP/1.1", "")
	e2.SetResponseStatus(200, "OK")
	e2.Time = 3000 // 3000ms -> overshoots 2000ms

	results := h.SLACheck([]SLARule{
		{Name: "invalid", URLPattern: "(invalid regex", MaxTime: 100 * time.Millisecond},                    // skipped (lines 250-254)
		{Name: "api", URLPattern: "api.example.com", Method: "GET", MaxTime: 1000 * time.Millisecond},       // matches e1, passes
		{Name: "api-post", URLPattern: "api.example.com", Method: "POST", MaxTime: 1000 * time.Millisecond}, // matches e2, fails w/ overshoot
		{Name: "all", MaxTime: 1 * time.Millisecond},                                                        // matches both, both fail
	})

	// Invalid regex rule produces no results; verify the api rule passed.
	var apiResult, postResult *SLAResult
	for i := range results {
		switch results[i].Rule.Name {
		case "api":
			apiResult = &results[i]
		case "api-post":
			postResult = &results[i]
		}
	}
	if assert.NotNil(t, apiResult) {
		assert.True(t, apiResult.Passed)
		assert.Equal(t, time.Duration(0), apiResult.Overshoot)
	}
	if assert.NotNil(t, postResult) {
		assert.False(t, postResult.Passed)
		assert.Equal(t, 2000*time.Millisecond, postResult.Overshoot)
	}
}

// Cover SLACheck nil/empty-rules branch (lines 240-242).
func TestCovSLACheck_NilEmpty(t *testing.T) {
	var h *Har
	assert.Nil(t, h.SLACheck(nil))

	empty := NewHar()
	assert.Nil(t, empty.SLACheck(nil))
}

// Cover ConcurrencyTimeline earliest-baseTime branch (lines 297-300) and
// the event-sweep including activeSet removal (lines 335-357).
func TestCovConcurrencyTimeline_Branches(t *testing.T) {
	h := NewHar()
	e1 := h.AddEntry("GET", "https://example.com/a", "HTTP/1.1", "")
	e1.StartedDateTime = time.Date(2024, 1, 1, 12, 0, 5, 0, time.UTC) // later
	e1.Time = 100

	e2 := h.AddEntry("GET", "https://example.com/b", "HTTP/1.1", "")
	e2.StartedDateTime = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC) // earlier
	e2.Time = 50

	pts := h.ConcurrencyTimeline()
	if !assert.NotNil(t, pts) {
		return
	}
	assert.NotEmpty(t, pts)
}

// Cover ConcurrencyTimeline nil/empty branch (lines 291-293).
func TestCovConcurrencyTimeline_NilEmpty(t *testing.T) {
	var h *Har
	assert.Nil(t, h.ConcurrencyTimeline())

	empty := NewHar()
	assert.Nil(t, empty.ConcurrencyTimeline())
}

// Cover PageTimingMetrics earliest-baseTime branch (lines 377-380) and
// the page-timings branch (lines 414-422).
func TestCovPageTimingMetrics_Branches(t *testing.T) {
	h := NewHar()
	h.AddPage("page_1", "Home")

	e1 := h.AddEntry("GET", "https://example.com/", "HTTP/1.1", "")
	e1.StartedDateTime = time.Date(2024, 1, 1, 12, 0, 5, 0, time.UTC) // later
	e1.Timings = Timings{Blocked: 10, DNS: 20, Connect: 30, Ssl: 40, Send: 5, Wait: 100, Receive: 10}
	e1.Time = 215

	e2 := h.AddEntry("GET", "https://example.com/asset.js", "HTTP/1.1", "")
	e2.StartedDateTime = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC) // earlier -> baseTime updated
	e2.Timings = Timings{DNS: 15, Connect: 25, Ssl: 35, Wait: 50}
	e2.Time = 200

	// Set page timings.
	h.Log.Pages[0].PageTimings.OnContentLoad = 800
	h.Log.Pages[0].PageTimings.OnLoad = 1500

	m := h.PageTimingMetrics()
	if !assert.NotNil(t, m) {
		return
	}
	assert.Greater(t, m.TotalTime, time.Duration(0))
	assert.Greater(t, m.TTFB, time.Duration(0))
	assert.Greater(t, m.DNSLookup, time.Duration(0))
	assert.Equal(t, 800*time.Millisecond, m.DOMContentLoaded)
	assert.Equal(t, 1500*time.Millisecond, m.OnLoad)
}

// Cover PageTimingMetrics nil-receiver (lines 365-367) and empty-entries
// (lines 371-373).
func TestCovPageTimingMetrics_NilEmpty(t *testing.T) {
	var h *Har
	m := h.PageTimingMetrics()
	if assert.NotNil(t, m) {
		assert.Equal(t, time.Duration(0), m.TotalTime)
	}

	empty := NewHar()
	m2 := empty.PageTimingMetrics()
	if assert.NotNil(t, m2) {
		assert.Equal(t, time.Duration(0), m2.TotalTime)
	}
}

// Cover CriticalPath branches (lines 146-171).
func TestCovCriticalPath_Branches(t *testing.T) {
	// nil/empty
	var h *Har
	assert.Nil(t, h.CriticalPath())
	empty := NewHar()
	assert.Nil(t, empty.CriticalPath())

	// With a CSS + an async JS + a font.
	hh := NewHar()
	doc := hh.AddEntry("GET", "https://example.com/", "HTTP/1.1", "")
	doc.Response.Content.MimeType = "text/html"

	css := hh.AddEntry("GET", "https://example.com/style.css", "HTTP/1.1", "")
	css.Response.Content.MimeType = "text/css"

	asyncJS := hh.AddEntry("GET", "https://example.com/tracking.js", "HTTP/1.1", "")
	asyncJS.Response.Content.MimeType = "application/javascript"
	asyncJS.Request.Headers = []Headers{{Name: "X-Script-Async", Value: "true"}}

	font := hh.AddEntry("GET", "https://example.com/icon.woff2", "HTTP/1.1", "")
	font.Response.Content.MimeType = "application/octet-stream"

	cp := hh.CriticalPath()
	if !assert.NotNil(t, cp) {
		return
	}
	// First entry (doc) is always critical; CSS and font are critical;
	// async JS is NOT critical. So at least 3 critical entries (doc, css, font).
	assert.GreaterOrEqual(t, len(cp), 3)
}
