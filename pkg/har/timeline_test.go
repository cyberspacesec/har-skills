package har

import (
	"testing"
	"time"
)

// createTestHarForTimeline creates a HAR with overlapping requests for
// timeline/waterfall testing.
func createTestHarForTimeline() *Har {
	h := NewHar()
	h.SetCreator("test", "1.0")
	h.SetBrowser("Chrome", "100.0")

	baseTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

	// Entry 0: HTML document (first request) - 0ms start, 200ms total
	e0 := h.AddEntry("GET", "https://example.com/", "HTTP/1.1", "page_1")
	e0.SetResponseStatus(200, "OK")
	e0.SetResponseContent(4096, "text/html")
	e0.SetTimings(10, 5, 20, 3, 100, 50, 8)
	e0.StartedDateTime = baseTime

	// Entry 1: CSS (blocking) - starts at 150ms, 100ms total
	e1 := h.AddEntry("GET", "https://example.com/style.css", "HTTP/1.1", "page_1")
	e1.SetResponseStatus(200, "OK")
	e1.SetResponseContent(2048, "text/css")
	e1.SetTimings(5, 2, 10, 2, 40, 30, 5)
	e1.StartedDateTime = baseTime.Add(150 * time.Millisecond)

	// Entry 2: JS (blocking, no async/defer) - starts at 200ms, 80ms total
	e2 := h.AddEntry("GET", "https://example.com/app.js", "HTTP/1.1", "page_1")
	e2.SetResponseStatus(200, "OK")
	e2.SetResponseContent(1024, "application/javascript")
	e2.SetTimings(3, 1, 8, 2, 35, 25, 4)
	e2.StartedDateTime = baseTime.Add(200 * time.Millisecond)

	// Entry 3: JS async script - starts at 200ms, 60ms total
	e3 := h.AddEntry("GET", "https://example.com/async-lib.js", "HTTP/1.1", "page_1")
	e3.SetResponseStatus(200, "OK")
	e3.SetResponseContent(512, "application/javascript")
	e3.SetTimings(2, 1, 5, 1, 25, 20, 3)
	e3.StartedDateTime = baseTime.Add(200 * time.Millisecond)
	// Mark as async via request header
	e3.AddRequestHeader("X-Script-Async", "true")

	// Entry 4: Image (not critical) - starts at 250ms, 120ms total
	e4 := h.AddEntry("GET", "https://example.com/logo.png", "HTTP/1.1", "page_1")
	e4.SetResponseStatus(200, "OK")
	e4.SetResponseContent(8192, "image/png")
	e4.SetTimings(5, 2, 15, 2, 50, 40, 6)
	e4.StartedDateTime = baseTime.Add(250 * time.Millisecond)

	// Entry 5: Font - starts at 300ms, 70ms total
	e5 := h.AddEntry("GET", "https://example.com/fonts/roboto.woff2", "HTTP/1.1", "page_1")
	e5.SetResponseStatus(200, "OK")
	e5.SetResponseContent(256, "font/woff2")
	e5.SetTimings(2, 1, 10, 2, 30, 20, 5)
	e5.StartedDateTime = baseTime.Add(300 * time.Millisecond)

	// Entry 6: Another CSS - starts at 160ms, 90ms total (overlaps with entry 1)
	e6 := h.AddEntry("GET", "https://cdn.example.com/theme.css", "HTTP/1.1", "page_1")
	e6.SetResponseStatus(200, "OK")
	e6.SetResponseContent(1024, "text/css")
	e6.SetTimings(4, 1, 12, 2, 35, 30, 6)
	e6.StartedDateTime = baseTime.Add(160 * time.Millisecond)

	// Add connection IDs for connection reuse testing
	// Access entries directly from the slice to avoid stale pointer issues
	h.Log.Entries[0].Connection = "conn-1"
	h.Log.Entries[1].Connection = "conn-1"
	h.Log.Entries[2].Connection = "conn-2"
	h.Log.Entries[3].Connection = "conn-2"
	h.Log.Entries[4].Connection = "conn-3"
	h.Log.Entries[5].Connection = "conn-1"
	h.Log.Entries[6].Connection = "conn-2"

	// Add a page with timings
	page := h.AddPage("page_1", "Example Page")
	page.SetPageTimings(350, 500)
	page.StartedDateTime = baseTime

	return h
}

func TestWaterfall(t *testing.T) {
	h := createTestHarForTimeline()
	wf := h.Waterfall()

	if len(wf) != 7 {
		t.Fatalf("Expected 7 waterfall entries, got %d", len(wf))
	}

	// First entry starts at offset 0
	if wf[0].StartTime != 0 {
		t.Errorf("Expected first entry StartTime=0, got %v", wf[0].StartTime)
	}

	// Check that durations are computed correctly
	// Entry 0 total = 10+5+20+3+100+50 = 188ms (note: SSL not summed separately per HAR spec)
	if wf[0].Duration != 188*time.Millisecond {
		t.Errorf("Expected entry 0 Duration=188ms, got %v", wf[0].Duration)
	}

	// Check URL
	if wf[0].URL != "https://example.com/" {
		t.Errorf("Expected entry 0 URL=https://example.com/, got %s", wf[0].URL)
	}

	// Check Method
	if wf[0].Method != "GET" {
		t.Errorf("Expected entry 0 Method=GET, got %s", wf[0].Method)
	}

	// Check StatusCode
	if wf[0].StatusCode != 200 {
		t.Errorf("Expected entry 0 StatusCode=200, got %d", wf[0].StatusCode)
	}

	// Check that entry 1 start offset is relative to first request
	if wf[1].StartTime != 150*time.Millisecond {
		t.Errorf("Expected entry 1 StartTime=150ms, got %v", wf[1].StartTime)
	}

	// Check depth: entry 0 should be depth 0
	if wf[0].Depth != 0 {
		t.Errorf("Expected entry 0 Depth=0, got %d", wf[0].Depth)
	}

	// Entry 1 starts at 150ms, entry 0 ends at ~188ms, so they overlap -> depth >= 1
	if wf[1].Depth < 1 {
		t.Errorf("Expected entry 1 Depth >= 1 (overlaps with entry 0), got %d", wf[1].Depth)
	}

	// Check timing phases
	if wf[0].Phases.DNS != 5*time.Millisecond {
		t.Errorf("Expected entry 0 Phases.DNS=5ms, got %v", wf[0].Phases.DNS)
	}
	if wf[0].Phases.Connect != 20*time.Millisecond {
		t.Errorf("Expected entry 0 Phases.Connect=20ms, got %v", wf[0].Phases.Connect)
	}
	if wf[0].Phases.SSL != 8*time.Millisecond {
		t.Errorf("Expected entry 0 Phases.SSL=8ms, got %v", wf[0].Phases.SSL)
	}
	if wf[0].Phases.Send != 3*time.Millisecond {
		t.Errorf("Expected entry 0 Phases.Send=3ms, got %v", wf[0].Phases.Send)
	}
	if wf[0].Phases.Wait != 100*time.Millisecond {
		t.Errorf("Expected entry 0 Phases.Wait=100ms, got %v", wf[0].Phases.Wait)
	}
	if wf[0].Phases.Receive != 50*time.Millisecond {
		t.Errorf("Expected entry 0 Phases.Receive=50ms, got %v", wf[0].Phases.Receive)
	}
}

func TestWaterfallNil(t *testing.T) {
	var h *Har
	wf := h.Waterfall()
	if wf != nil {
		t.Errorf("Expected nil waterfall for nil HAR, got %v", wf)
	}
}

func TestWaterfallEmpty(t *testing.T) {
	h := NewHar()
	wf := h.Waterfall()
	if wf != nil {
		t.Errorf("Expected nil waterfall for empty HAR, got %v", wf)
	}
}

func TestCriticalPath(t *testing.T) {
	h := createTestHarForTimeline()
	cp := h.CriticalPath()

	// First entry (HTML document) is always critical
	if len(cp) == 0 {
		t.Fatal("Expected non-empty critical path")
	}

	// Entry 0 (HTML) must be in critical path
	foundHTML := false
	for _, e := range cp {
		if e.URL == "https://example.com/" {
			foundHTML = true
			break
		}
	}
	if !foundHTML {
		t.Error("Expected HTML document in critical path")
	}

	// CSS entries should be in critical path
	foundCSS := false
	for _, e := range cp {
		if e.URL == "https://example.com/style.css" {
			foundCSS = true
			break
		}
	}
	if !foundCSS {
		t.Error("Expected CSS in critical path")
	}

	// Blocking JS (without async/defer) should be in critical path
	foundBlockingJS := false
	for _, e := range cp {
		if e.URL == "https://example.com/app.js" {
			foundBlockingJS = true
			break
		}
	}
	if !foundBlockingJS {
		t.Error("Expected blocking JS in critical path")
	}

	// Async JS should NOT be in critical path
	foundAsyncJS := false
	for _, e := range cp {
		if e.URL == "https://example.com/async-lib.js" {
			foundAsyncJS = true
			break
		}
	}
	if foundAsyncJS {
		t.Error("Async JS should NOT be in critical path")
	}

	// Images should NOT be in critical path
	foundImage := false
	for _, e := range cp {
		if e.URL == "https://example.com/logo.png" {
			foundImage = true
			break
		}
	}
	if foundImage {
		t.Error("Image should NOT be in critical path")
	}

	// Font should be in critical path
	foundFont := false
	for _, e := range cp {
		if e.URL == "https://example.com/fonts/roboto.woff2" {
			foundFont = true
			break
		}
	}
	if !foundFont {
		t.Error("Expected font in critical path")
	}
}

func TestCriticalPathNil(t *testing.T) {
	var h *Har
	cp := h.CriticalPath()
	if cp != nil {
		t.Errorf("Expected nil critical path for nil HAR, got %v", cp)
	}
}

func TestCriticalPathEmpty(t *testing.T) {
	h := NewHar()
	cp := h.CriticalPath()
	if cp != nil {
		t.Errorf("Expected nil critical path for empty HAR, got %v", cp)
	}
}

func TestSLACheck(t *testing.T) {
	h := createTestHarForTimeline()

	rules := []SLARule{
		{
			Name:       "All requests under 200ms",
			URLPattern: "",
			Method:     "",
			MaxTime:    200 * time.Millisecond,
		},
		{
			Name:       "CSS under 100ms",
			URLPattern: "\\.css$",
			Method:     "GET",
			MaxTime:    100 * time.Millisecond,
		},
		{
			Name:       "GET requests under 150ms",
			URLPattern: "",
			Method:     "GET",
			MaxTime:    150 * time.Millisecond,
		},
	}

	results := h.SLACheck(rules)

	// Should have results for each matching entry per rule
	if len(results) == 0 {
		t.Fatal("Expected non-empty SLA results")
	}

	// Count passed and failed
	var passed, failed int
	for _, r := range results {
		if r.Passed {
			passed++
		} else {
			failed++
		}
	}

	// There should be at least some failures (entry 0 is ~188ms, entry 4 is ~120ms)
	if passed == 0 {
		t.Error("Expected at least one passed SLA check")
	}

	// Verify overshoot calculation for failed results
	for _, r := range results {
		if !r.Passed {
			if r.Overshoot != r.Actual-r.Rule.MaxTime {
				t.Errorf("Overshoot mismatch: got %v, expected %v (actual=%v, max=%v)",
					r.Overshoot, r.Actual-r.Rule.MaxTime, r.Actual, r.Rule.MaxTime)
			}
		} else {
			if r.Overshoot != 0 {
				t.Errorf("Passed result should have 0 overshoot, got %v", r.Overshoot)
			}
		}
	}
}

func TestSLACheckNil(t *testing.T) {
	var h *Har
	results := h.SLACheck([]SLARule{
		{Name: "test", MaxTime: 100 * time.Millisecond},
	})
	if results != nil {
		t.Errorf("Expected nil SLA results for nil HAR, got %v", results)
	}
}

func TestSLACheckEmptyRules(t *testing.T) {
	h := createTestHarForTimeline()
	results := h.SLACheck(nil)
	if results != nil {
		t.Errorf("Expected nil SLA results for empty rules, got %v", results)
	}
}

func TestConcurrencyTimeline(t *testing.T) {
	h := createTestHarForTimeline()
	ct := h.ConcurrencyTimeline()

	if len(ct) == 0 {
		t.Fatal("Expected non-empty concurrency timeline")
	}

	// First point should have ActiveCount=1 (first request started)
	if ct[0].ActiveCount != 1 {
		t.Errorf("Expected first concurrency point ActiveCount=1, got %d", ct[0].ActiveCount)
	}

	// There should be at least one point with ActiveCount >= 2 (overlapping requests)
	maxActive := 0
	for _, p := range ct {
		if p.ActiveCount > maxActive {
			maxActive = p.ActiveCount
		}
	}
	if maxActive < 2 {
		t.Errorf("Expected at least one point with ActiveCount >= 2, max was %d", maxActive)
	}

	// Last point should have ActiveCount=0 (all requests ended)
	if ct[len(ct)-1].ActiveCount != 0 {
		t.Errorf("Expected last concurrency point ActiveCount=0, got %d", ct[len(ct)-1].ActiveCount)
	}

	// All entries should be tracked in active sets at some point
	allIndicesSeen := make(map[int]bool)
	for _, p := range ct {
		for _, idx := range p.ActiveEntries {
			allIndicesSeen[idx] = true
		}
	}
	for i := 0; i < 7; i++ {
		if !allIndicesSeen[i] {
			t.Errorf("Entry %d never appeared in active entries", i)
		}
	}
}

func TestConcurrencyTimelineNil(t *testing.T) {
	var h *Har
	ct := h.ConcurrencyTimeline()
	if ct != nil {
		t.Errorf("Expected nil concurrency timeline for nil HAR, got %v", ct)
	}
}

func TestConcurrencyTimelineEmpty(t *testing.T) {
	h := NewHar()
	ct := h.ConcurrencyTimeline()
	if ct != nil {
		t.Errorf("Expected nil concurrency timeline for empty HAR, got %v", ct)
	}
}

func TestPageTimingMetrics(t *testing.T) {
	h := createTestHarForTimeline()
	m := h.PageTimingMetrics()

	if m == nil {
		t.Fatal("Expected non-nil PageTimingMetrics")
	}

	// TTFB should be > 0 (sum of first entry phases before receive)
	if m.TTFB == 0 {
		t.Error("Expected TTFB > 0")
	}

	// DOMContentLoaded should come from page timings (350ms)
	if m.DOMContentLoaded != 350*time.Millisecond {
		t.Errorf("Expected DOMContentLoaded=350ms, got %v", m.DOMContentLoaded)
	}

	// OnLoad should come from page timings (500ms)
	if m.OnLoad != 500*time.Millisecond {
		t.Errorf("Expected OnLoad=500ms, got %v", m.OnLoad)
	}

	// TotalTime should be > 0
	if m.TotalTime == 0 {
		t.Error("Expected TotalTime > 0")
	}

	// DNS, Connect, SSL should be > 0
	if m.DNSLookup == 0 {
		t.Error("Expected DNSLookup > 0")
	}
	if m.ConnectTime == 0 {
		t.Error("Expected ConnectTime > 0")
	}
	if m.SSLTime == 0 {
		t.Error("Expected SSLTime > 0")
	}
}

func TestPageTimingMetricsNil(t *testing.T) {
	var h *Har
	m := h.PageTimingMetrics()
	if m.TTFB != 0 || m.TotalTime != 0 {
		t.Errorf("Expected zero metrics for nil HAR")
	}
}

func TestPageTimingMetricsEmpty(t *testing.T) {
	h := NewHar()
	m := h.PageTimingMetrics()
	if m.TTFB != 0 || m.TotalTime != 0 {
		t.Errorf("Expected zero metrics for empty HAR")
	}
}

func TestConnectionReuse(t *testing.T) {
	h := createTestHarForTimeline()
	reuse := h.ConnectionReuse()

	// There should be 3 connection IDs
	if len(reuse) != 3 {
		t.Errorf("Expected 3 connection IDs, got %d", len(reuse))
	}

	// conn-1 should have entries 0, 1, 5
	if len(reuse["conn-1"]) != 3 {
		t.Errorf("Expected conn-1 to have 3 entries, got %d", len(reuse["conn-1"]))
	}

	// conn-2 should have entries 2, 3, 6
	if len(reuse["conn-2"]) != 3 {
		t.Errorf("Expected conn-2 to have 3 entries, got %d", len(reuse["conn-2"]))
	}

	// conn-3 should have entry 4
	if len(reuse["conn-3"]) != 1 {
		t.Errorf("Expected conn-3 to have 1 entry, got %d", len(reuse["conn-3"]))
	}
}

func TestConnectionReuseNil(t *testing.T) {
	var h *Har
	reuse := h.ConnectionReuse()
	if len(reuse) != 0 {
		t.Errorf("Expected empty connection reuse for nil HAR, got %d entries", len(reuse))
	}
}

func TestConnectionReuseNoConnections(t *testing.T) {
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/", "HTTP/1.1", "")
	// No connection ID set
	e.SetResponseStatus(200, "OK")
	reuse := h.ConnectionReuse()
	if len(reuse) != 0 {
		t.Errorf("Expected empty connection reuse when no connections set, got %d", len(reuse))
	}
}

func TestMsToDuration(t *testing.T) {
	tests := []struct {
		ms       float64
		expected time.Duration
	}{
		{0, 0},
		{100, 100 * time.Millisecond},
		{1.5, 1500 * time.Microsecond},
		{-1, 0},     // unknown/invalid
		{-5, 0},     // unknown/invalid
		{0.001, time.Microsecond},
	}

	for _, tt := range tests {
		result := msToDuration(tt.ms)
		if result != tt.expected {
			t.Errorf("msToDuration(%f) = %v, expected %v", tt.ms, result, tt.expected)
		}
	}
}

func TestWaterfallNegativeTimings(t *testing.T) {
	// Test with entries that have -1 timings (unknown values)
	h := NewHar()
	baseTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

	e := h.AddEntry("GET", "https://example.com/", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.StartedDateTime = baseTime
	// Leave timings as default (-1 values from AddEntry)
	// Override specific known values
	e.Timings.Send = 5
	e.Timings.Wait = 100
	e.Timings.Receive = 20
	e.Time = 125 // send+wait+receive

	wf := h.Waterfall()
	if len(wf) != 1 {
		t.Fatalf("Expected 1 waterfall entry, got %d", len(wf))
	}

	// Negative timings should result in 0 Duration phases
	if wf[0].Phases.DNS != 0 {
		t.Errorf("Expected DNS=0 for -1 value, got %v", wf[0].Phases.DNS)
	}
	if wf[0].Phases.Connect != 0 {
		t.Errorf("Expected Connect=0 for -1 value, got %v", wf[0].Phases.Connect)
	}
	if wf[0].Phases.SSL != 0 {
		t.Errorf("Expected SSL=0 for -1 value, got %v", wf[0].Phases.SSL)
	}
	// Known timings should be correctly converted
	if wf[0].Phases.Send != 5*time.Millisecond {
		t.Errorf("Expected Send=5ms, got %v", wf[0].Phases.Send)
	}
	if wf[0].Phases.Wait != 100*time.Millisecond {
		t.Errorf("Expected Wait=100ms, got %v", wf[0].Phases.Wait)
	}
	if wf[0].Phases.Receive != 20*time.Millisecond {
		t.Errorf("Expected Receive=20ms, got %v", wf[0].Phases.Receive)
	}
}