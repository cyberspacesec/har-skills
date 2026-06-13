package har

import (
	"regexp"
	"sort"
	"strings"
	"time"
)

// TimingPhases represents detailed timing breakdown for a request.
type TimingPhases struct {
	DNS     time.Duration
	Connect time.Duration
	SSL     time.Duration
	Send    time.Duration
	Wait    time.Duration
	Receive time.Duration
	Blocked time.Duration
}

// WaterfallEntry represents a single entry in a waterfall timeline.
type WaterfallEntry struct {
	Index      int           // entry index
	URL        string        // request URL
	Method     string        // HTTP method
	StatusCode int           // response status
	StartTime  time.Duration // offset from first request
	EndTime    time.Duration // offset from first request
	Duration   time.Duration // total time
	Phases     TimingPhases  // detailed timing phases
	Depth      int           // depth for waterfall visualization
}

// SLARule defines a service-level agreement threshold.
type SLARule struct {
	Name       string        // rule name
	URLPattern string        // regex pattern to match URLs (empty = all)
	Method     string        // HTTP method filter (empty = all)
	MaxTime    time.Duration // maximum allowed time
}

// SLAResult records the outcome of checking one entry against one SLA rule.
type SLAResult struct {
	Rule      SLARule
	Entry     Entries
	Actual    time.Duration
	Passed    bool
	Overshoot time.Duration // how much over the limit
}

// ConcurrencyPoint represents the number of active requests at a moment.
type ConcurrencyPoint struct {
	Time          time.Duration // offset from first request
	ActiveCount   int           // number of concurrent requests
	ActiveEntries []int         // entry indices
}

// PageTimingMetrics aggregates page-level timing metrics.
type PageTimingMetrics struct {
	TTFB              time.Duration // Time to First Byte (of first document)
	DOMContentLoaded  time.Duration
	OnLoad            time.Duration
	TotalTime         time.Duration
	DNSLookup         time.Duration // total DNS time
	ConnectTime       time.Duration // total connect time
	SSLTime           time.Duration // total SSL time
}

// msToDuration converts a float64 millisecond value (where -1 means unknown)
// to a time.Duration. Unknown values are returned as 0.
func msToDuration(ms float64) time.Duration {
	if ms < 0 {
		return 0
	}
	return time.Duration(ms * float64(time.Millisecond))
}

// Waterfall computes a waterfall timeline with start/end offsets relative
// to the first request's StartedDateTime. Depth is computed based on
// overlapping time ranges: if a request starts while another is in
// progress, its depth increments.
func (h *Har) Waterfall() []WaterfallEntry {
	if h == nil || len(h.Log.Entries) == 0 {
		return nil
	}

	entries := h.Log.Entries
	baseTime := entries[0].StartedDateTime
	// Find the earliest start time as baseline
	for _, e := range entries {
		if e.StartedDateTime.Before(baseTime) {
			baseTime = e.StartedDateTime
		}
	}

	result := make([]WaterfallEntry, len(entries))
	for i, e := range entries {
		start := e.StartedDateTime.Sub(baseTime)
		dur := msToDuration(e.Time)
		end := start + dur

		result[i] = WaterfallEntry{
			Index:      i,
			URL:        e.Request.URL,
			Method:     e.Request.Method,
			StatusCode: e.Response.Status,
			StartTime:  start,
			EndTime:    end,
			Duration:   dur,
			Phases: TimingPhases{
				Blocked: msToDuration(e.Timings.Blocked),
				DNS:     msToDuration(e.Timings.DNS),
				Connect: msToDuration(e.Timings.Connect),
				SSL:     msToDuration(e.Timings.Ssl),
				Send:    msToDuration(e.Timings.Send),
				Wait:    msToDuration(e.Timings.Wait),
				Receive: msToDuration(e.Timings.Receive),
			},
			Depth: 0,
		}
	}

	// Compute depth: for each entry, find how many previous entries overlap.
	// An entry's depth is the maximum depth among overlapping entries + 1,
	// so that no two overlapping entries share the same depth.
	for i := range result {
		depth := 0
		for j := 0; j < i; j++ {
			// j overlaps with i if j's end > i's start
			if result[j].EndTime > result[i].StartTime {
				if result[j].Depth+1 > depth {
					depth = result[j].Depth + 1
				}
			}
		}
		result[i].Depth = depth
	}

	return result
}

// CriticalPath identifies the critical rendering path. It starts with the
// first document request, then follows blocking CSS, blocking JS (without
// async/defer), and fonts referenced by CSS. Images and async scripts are
// NOT considered critical.
func (h *Har) CriticalPath() []WaterfallEntry {
	if h == nil || len(h.Log.Entries) == 0 {
		return nil
	}

	waterfall := h.Waterfall()
	entries := h.Log.Entries

	var critical []WaterfallEntry
	criticalSet := make(map[int]bool)

	// First entry is always in the critical path
	critical = append(critical, waterfall[0])
	criticalSet[0] = true

	// Walk remaining entries and apply heuristics
	for i := 1; i < len(entries); i++ {
		e := entries[i]
		if isCriticalResource(e) {
			criticalSet[i] = true
			critical = append(critical, waterfall[i])
		}
	}

	return critical
}

// isCriticalResource determines if a request is on the critical rendering
// path using heuristics based on content type and script attributes.
func isCriticalResource(e Entries) bool {
	mimeType := strings.ToLower(e.Response.Content.MimeType)

	// CSS resources are render-blocking
	if strings.Contains(mimeType, "text/css") {
		return true
	}

	// JS resources without async/defer are parser-blocking
	if strings.Contains(mimeType, "javascript") {
		return !hasAsyncOrDefer(e)
	}

	// Font resources referenced by CSS (font/* or application/font-*)
	if strings.Contains(mimeType, "font/") ||
		strings.Contains(mimeType, "application/font") ||
		strings.Contains(mimeType, "application/x-font") {
		return true
	}

	// Common web font formats by extension
	url := strings.ToLower(e.Request.URL)
	if strings.HasSuffix(url, ".woff") ||
		strings.HasSuffix(url, ".woff2") ||
		strings.HasSuffix(url, ".ttf") ||
		strings.HasSuffix(url, ".otf") ||
		strings.HasSuffix(url, ".eot") {
		return true
	}

	return false
}

// hasAsyncOrDefer checks if a script request has async or defer attributes.
// These can be indicated via custom fields or request headers.
func hasAsyncOrDefer(e Entries) bool {
	// Check request headers for async/defer hints (some tools add these)
	for _, h := range e.Request.Headers {
		name := strings.ToLower(h.Name)
		if name == "x-script-async" || name == "x-script-defer" {
			return true
		}
	}

	// Check CustomFields for async/defer flags
	if e.CustomFields != nil {
		if v, ok := e.CustomFields["async"]; ok {
			if b, ok := v.(bool); ok && b {
				return true
			}
		}
		if v, ok := e.CustomFields["defer"]; ok {
			if b, ok := v.(bool); ok && b {
				return true
			}
		}
	}

	return false
}

// SLACheck checks requests against SLA thresholds. For each rule, every
// matching entry is evaluated. A rule with an empty URLPattern and Method
// matches all entries.
func (h *Har) SLACheck(rules []SLARule) []SLAResult {
	if h == nil || len(rules) == 0 {
		return nil
	}

	var results []SLAResult

	for _, rule := range rules {
		var urlRe *regexp.Regexp
		if rule.URLPattern != "" {
			var err error
			urlRe, err = regexp.Compile(rule.URLPattern)
			if err != nil {
				// Skip invalid regex patterns
				continue
			}
		}

		for _, entry := range h.Log.Entries {
			// Method filter
			if rule.Method != "" && entry.Request.Method != rule.Method {
				continue
			}

			// URL pattern filter
			if urlRe != nil && !urlRe.MatchString(entry.Request.URL) {
				continue
			}

			actual := msToDuration(entry.Time)
			passed := actual <= rule.MaxTime
			var overshoot time.Duration
			if !passed {
				overshoot = actual - rule.MaxTime
			}

			results = append(results, SLAResult{
				Rule:      rule,
				Entry:     entry,
				Actual:    actual,
				Passed:    passed,
				Overshoot: overshoot,
			})
		}
	}

	return results
}

// ConcurrencyTimeline computes concurrent request count over time. It
// samples at each request start and end event.
func (h *Har) ConcurrencyTimeline() []ConcurrencyPoint {
	if h == nil || len(h.Log.Entries) == 0 {
		return nil
	}

	entries := h.Log.Entries
	baseTime := entries[0].StartedDateTime
	for _, e := range entries {
		if e.StartedDateTime.Before(baseTime) {
			baseTime = e.StartedDateTime
		}
	}

	// Build events: +1 at start, -1 at end
	type event struct {
		time  time.Duration
		delta int
		idx   int
	}

	var events []event
	for i, e := range entries {
		start := e.StartedDateTime.Sub(baseTime)
		dur := msToDuration(e.Time)
		end := start + dur

		events = append(events, event{time: start, delta: 1, idx: i})
		events = append(events, event{time: end, delta: -1, idx: i})
	}

	// Sort events by time; for same time, process starts before ends
	// so that concurrent starts are counted together.
	sort.Slice(events, func(i, j int) bool {
		if events[i].time != events[j].time {
			return events[i].time < events[j].time
		}
		// starts (+1) before ends (-1)
		return events[i].delta > events[j].delta
	})

	// Sweep through events, collecting unique concurrency points
	var points []ConcurrencyPoint
	activeCount := 0
	var activeSet []int

	for _, ev := range events {
		if ev.delta > 0 {
			activeCount++
			activeSet = append(activeSet, ev.idx)
		} else {
			activeCount--
			// Remove from activeSet
			for k, v := range activeSet {
				if v == ev.idx {
					activeSet = append(activeSet[:k], activeSet[k+1:]...)
					break
				}
			}
		}

		indices := make([]int, len(activeSet))
		copy(indices, activeSet)
		points = append(points, ConcurrencyPoint{
			Time:          ev.time,
			ActiveCount:   activeCount,
			ActiveEntries: indices,
		})
	}

	return points
}

// PageTimingMetrics aggregates page-level metrics from the first page and
// overall entry data.
func (h *Har) PageTimingMetrics() *PageTimingMetrics {
	if h == nil {
		return &PageTimingMetrics{}
	}

	m := &PageTimingMetrics{}

	if len(h.Log.Entries) == 0 {
		return m
	}

	// Find baseline (earliest request)
	baseTime := h.Log.Entries[0].StartedDateTime
	for _, e := range h.Log.Entries {
		if e.StartedDateTime.Before(baseTime) {
			baseTime = e.StartedDateTime
		}
	}

	// Find the last end time for TotalTime
	var lastEnd time.Time
	for _, e := range h.Log.Entries {
		end := e.StartedDateTime.Add(msToDuration(e.Time))
		if end.After(lastEnd) {
			lastEnd = end
		}
	}
	m.TotalTime = lastEnd.Sub(baseTime)

	// TTFB: wait time of the first (document) entry
	firstEntry := h.Log.Entries[0]
	m.TTFB = msToDuration(firstEntry.Timings.Blocked) +
		msToDuration(firstEntry.Timings.DNS) +
		msToDuration(firstEntry.Timings.Connect) +
		msToDuration(firstEntry.Timings.Ssl) +
		msToDuration(firstEntry.Timings.Send) +
		msToDuration(firstEntry.Timings.Wait)

	// Aggregate DNS, Connect, SSL across all entries
	var totalDNS, totalConnect, totalSSL time.Duration
	for _, e := range h.Log.Entries {
		totalDNS += msToDuration(e.Timings.DNS)
		totalConnect += msToDuration(e.Timings.Connect)
		totalSSL += msToDuration(e.Timings.Ssl)
	}
	m.DNSLookup = totalDNS
	m.ConnectTime = totalConnect
	m.SSLTime = totalSSL

	// DOMContentLoaded and OnLoad from the first page
	if len(h.Log.Pages) > 0 {
		page := h.Log.Pages[0]
		if page.PageTimings.OnContentLoad > 0 {
			m.DOMContentLoaded = time.Duration(page.PageTimings.OnContentLoad * float64(time.Millisecond))
		}
		if page.PageTimings.OnLoad > 0 {
			m.OnLoad = time.Duration(page.PageTimings.OnLoad * float64(time.Millisecond))
		}
	}

	return m
}

// ConnectionReuse finds entries sharing the same connection ID. It returns
// a map of connection ID to entry indices.
func (h *Har) ConnectionReuse() map[string][]int {
	result := make(map[string][]int)

	if h == nil {
		return result
	}

	for i, e := range h.Log.Entries {
		if e.Connection != "" {
			result[e.Connection] = append(result[e.Connection], i)
		}
	}

	return result
}
