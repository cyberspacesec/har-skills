package har

import (
	"net/http"
	"testing"
	"time"
)

func assertDoesNotPanic(t *testing.T, fn func()) {
	t.Helper()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	fn()
}

// ========== FilterOption 测试 ==========

func TestNewFilterOptions(t *testing.T) {
	opts := NewFilterOptions(
		WithFilterURL("example.com"),
		WithFilterMethod("GET"),
		WithFilterStatusCode(200),
		WithFilterRegex(),
	)

	if opts.URL != "example.com" {
		t.Errorf("Expected URL 'example.com', got '%s'", opts.URL)
	}
	if opts.Method != "GET" {
		t.Errorf("Expected Method 'GET', got '%s'", opts.Method)
	}
	if opts.StatusCode != 200 {
		t.Errorf("Expected StatusCode 200, got %d", opts.StatusCode)
	}
	if !opts.UseRegex {
		t.Error("Expected UseRegex true")
	}
}

func TestFunctionalOptionsIgnoreNilOptions(t *testing.T) {
	assertDoesNotPanic(t, func() {
		opts := NewFilterOptions(nil, WithFilterMethod("POST"))
		if opts.Method != "POST" {
			t.Errorf("Expected Method POST, got %s", opts.Method)
		}
	})

	assertDoesNotPanic(t, func() {
		opts := NewReplayOptions(nil, WithReplayTimeout(time.Second))
		if opts.Timeout != time.Second {
			t.Errorf("Expected timeout 1s, got %v", opts.Timeout)
		}
		if !opts.FollowRedirects {
			t.Error("Expected default FollowRedirects true")
		}
	})

	assertDoesNotPanic(t, func() {
		opts := NewConvertOptions(nil, WithConvertIncludeHeaders(true))
		if !opts.IncludeHeaders {
			t.Error("Expected IncludeHeaders true")
		}
		if !opts.IncludeURL {
			t.Error("Expected default IncludeURL true")
		}
	})

	assertDoesNotPanic(t, func() {
		opts := NewDiffOptions(nil, WithDiffIncludeBody(true))
		if !opts.IncludeBody {
			t.Error("Expected IncludeBody true")
		}
		if !opts.IgnoreTimings {
			t.Error("Expected default IgnoreTimings true")
		}
	})

	assertDoesNotPanic(t, func() {
		opts := NewMergeOptions(nil, WithMergeDeduplicate(true))
		if !opts.Deduplicate {
			t.Error("Expected Deduplicate true")
		}
		if !opts.SortByTime {
			t.Error("Expected default SortByTime true")
		}
	})

	assertDoesNotPanic(t, func() {
		builder := NewHarBuilderWithOptions(nil, WithBuilderCreator("nil-safe", "1.0"))
		har := builder.Build()
		if har.Log.Creator.Name != "nil-safe" {
			t.Errorf("Expected creator nil-safe, got %s", har.Log.Creator.Name)
		}
	})
}

func TestFilterWithOptions(t *testing.T) {
	har := createTestHarForOpts()
	result := har.FilterWith(
		WithFilterMethod("GET"),
	)
	if result.Count() != 1 {
		t.Errorf("Expected 1 result, got %d", result.Count())
	}
}

func TestFilterOptionStatusCodeRange(t *testing.T) {
	opts := NewFilterOptions(
		WithFilterStatusCodeRange(200, 299),
	)
	if opts.StatusCodeMin != 200 || opts.StatusCodeMax != 299 {
		t.Errorf("Expected range 200-299, got %d-%d", opts.StatusCodeMin, opts.StatusCodeMax)
	}
}

func TestFilterOptionContentType(t *testing.T) {
	opts := NewFilterOptions(
		WithFilterContentType("text/html"),
	)
	if opts.ContentType != "text/html" {
		t.Errorf("Expected ContentType 'text/html', got '%s'", opts.ContentType)
	}
}

func TestFilterOptionTimeRange(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
	opts := NewFilterOptions(
		WithFilterTimeRange(start, end),
	)
	if !opts.StartTime.Equal(start) || !opts.EndTime.Equal(end) {
		t.Error("Time range not set correctly")
	}
}

func TestFilterOptionDuration(t *testing.T) {
	opts := NewFilterOptions(
		WithFilterDuration(100, 500),
	)
	if opts.MinDuration != 100 || opts.MaxDuration != 500 {
		t.Errorf("Expected duration 100-500, got %v-%v", opts.MinDuration, opts.MaxDuration)
	}
}

func TestFilterOptionHasError(t *testing.T) {
	opts := NewFilterOptions(WithFilterHasError())
	if !opts.HasError {
		t.Error("Expected HasError true")
	}
}

func TestFilterOptionHeader(t *testing.T) {
	opts := NewFilterOptions(
		WithFilterHeader("Content-Type", "application/json"),
	)
	if opts.HeaderName != "Content-Type" || opts.HeaderValue != "application/json" {
		t.Error("Header filter not set correctly")
	}
}

func TestFilterOptionResponseHeader(t *testing.T) {
	opts := NewFilterOptions(
		WithFilterResponseHeader("X-Custom", "value"),
	)
	if opts.RespHeaderName != "X-Custom" || opts.RespHeaderValue != "value" {
		t.Error("Response header filter not set correctly")
	}
}

// ========== ReplayOption 测试 ==========

func TestNewReplayOptions(t *testing.T) {
	opts := NewReplayOptions(
		WithReplayTimeout(10*time.Second),
		WithReplayFollowRedirects(false),
		WithReplayMaxRedirects(5),
		WithReplaySkipSSLVerify(true),
		WithReplayOverrideHeader("Authorization", "Bearer token"),
	)

	if opts.Timeout != 10*time.Second {
		t.Errorf("Expected timeout 10s, got %v", opts.Timeout)
	}
	if opts.FollowRedirects {
		t.Error("Expected FollowRedirects false")
	}
	if opts.MaxRedirects != 5 {
		t.Errorf("Expected MaxRedirects 5, got %d", opts.MaxRedirects)
	}
	if !opts.SkipSSLVerify {
		t.Error("Expected SkipSSLVerify true")
	}
	if opts.OverrideHeaders["Authorization"] != "Bearer token" {
		t.Error("Override header not set correctly")
	}
}

func TestDefaultReplayOptionsFunctional(t *testing.T) {
	opts := NewReplayOptions()
	if opts.Timeout != 30*time.Second {
		t.Errorf("Expected default timeout 30s, got %v", opts.Timeout)
	}
	if !opts.FollowRedirects {
		t.Error("Expected default FollowRedirects true")
	}
}

// ========== ConvertOption 测试 ==========

func TestNewConvertOptions(t *testing.T) {
	opts := NewConvertOptions(
		WithConvertIncludeHeaders(true),
		WithConvertIncludeTimings(true),
		WithConvertIncludeStatus(false),
		WithConvertIncludeURL(true),
	)

	if !opts.IncludeHeaders {
		t.Error("Expected IncludeHeaders true")
	}
	if !opts.IncludeTimings {
		t.Error("Expected IncludeTimings true")
	}
	if opts.IncludeStatus {
		t.Error("Expected IncludeStatus false")
	}
	if !opts.IncludeURL {
		t.Error("Expected IncludeURL true")
	}
}

func TestConvertWithOption(t *testing.T) {
	har := createTestHarForOpts()
	result, err := har.ConvertWith(FormatText,
		WithConvertIncludeHeaders(true),
		WithConvertIncludeMethod(true),
	)
	if err != nil {
		t.Fatalf("ConvertWith failed: %v", err)
	}
	if result == "" {
		t.Error("Expected non-empty result")
	}
}

// ========== DiffOption 测试 ==========

func TestNewDiffOptions(t *testing.T) {
	opts := NewDiffOptions(
		WithDiffIgnoreHeaders("Date", "X-Request-ID"),
		WithDiffIgnoreTimings(false),
		WithDiffNormalizeURL(true),
		WithDiffIncludeBody(true),
	)

	if len(opts.IgnoreHeaders) != 2 {
		t.Errorf("Expected 2 ignored headers, got %d", len(opts.IgnoreHeaders))
	}
	if opts.IgnoreTimings {
		t.Error("Expected IgnoreTimings false")
	}
	if !opts.NormalizeURL {
		t.Error("Expected NormalizeURL true")
	}
	if !opts.IncludeBody {
		t.Error("Expected IncludeBody true")
	}
}

func TestDiffWith(t *testing.T) {
	har1 := createTestHarForOpts()
	har2 := createTestHarForOpts()
	diff := DiffWith(har1, har2,
		WithDiffIgnoreTimings(true),
		WithDiffIgnoreDates(true),
	)
	if diff.HasChanges() {
		t.Error("Expected no changes for identical HARs")
	}

	assertDoesNotPanic(t, func() {
		diff = DiffWith(har1, har2, nil, WithDiffIgnoreDates(true))
		if diff == nil {
			t.Fatal("Expected non-nil diff")
		}
	})
}

// ========== MergeOption 测试 ==========

func TestNewMergeOptions(t *testing.T) {
	opts := NewMergeOptions(
		WithMergeSortByTime(false),
		WithMergeDeduplicate(true),
	)

	if opts.SortByTime {
		t.Error("Expected SortByTime false")
	}
	if !opts.Deduplicate {
		t.Error("Expected Deduplicate true")
	}
}

func TestMergeWith(t *testing.T) {
	har1 := createTestHarForOpts()
	har2 := createTestHarForOpts()

	mergeFunc := MergeWith(
		WithMergeSortByTime(true),
		WithMergeDeduplicate(true),
	)
	result := mergeFunc(har1, har2)

	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if len(result.Log.Entries) == 0 {
		t.Error("Expected entries in merged HAR")
	}

	assertDoesNotPanic(t, func() {
		mergeFunc = MergeWith(nil, WithMergeDeduplicate(true))
		result = mergeFunc(har1)
		if result == nil {
			t.Fatal("Expected non-nil result")
		}
	})
}

// ========== HarBuilderOption 测试 ==========

func TestNewHarBuilderWithOptions(t *testing.T) {
	builder := NewHarBuilderWithOptions(
		WithBuilderVersion("1.1"),
		WithBuilderCreator("test-tool", "2.0"),
		WithBuilderBrowser("Chrome", "120.0"),
		WithBuilderComment("Test HAR"),
	)

	har := builder.Build()
	if har.Log.Version != "1.1" {
		t.Errorf("Expected version '1.1', got '%s'", har.Log.Version)
	}
	if har.Log.Creator.Name != "test-tool" {
		t.Errorf("Expected creator 'test-tool', got '%s'", har.Log.Creator.Name)
	}
	if har.Log.Browser.Name != "Chrome" {
		t.Errorf("Expected browser 'Chrome', got '%s'", har.Log.Browser.Name)
	}
}

// ========== FilterOption: WithFilterResourceType 测试 ==========

func TestFilterOptionResourceType(t *testing.T) {
	opts := NewFilterOptions(
		WithFilterResourceType("script"),
	)
	if opts.ResourceType != "script" {
		t.Errorf("Expected ResourceType 'script', got '%s'", opts.ResourceType)
	}
}

func TestFilterOptionResourceTypeEmpty(t *testing.T) {
	opts := NewFilterOptions(
		WithFilterResourceType(""),
	)
	if opts.ResourceType != "" {
		t.Errorf("Expected ResourceType '', got '%s'", opts.ResourceType)
	}
}

// ========== ReplayOption: WithReplayTransport 测试 ==========

func TestReplayOptionTransport(t *testing.T) {
	opts := NewReplayOptions(
		WithReplayTransport(nil),
	)
	if opts.Transport != nil {
		t.Error("Expected Transport nil for nil replay transport")
	}
}

type mockRoundTripper struct{}

func (m *mockRoundTripper) RoundTrip(_ *http.Request) (*http.Response, error) {
	return nil, nil
}

func TestReplayOptionTransportAssignsRoundTripper(t *testing.T) {
	transport := &mockRoundTripper{}
	opts := NewReplayOptions(
		WithReplayTransport(transport),
	)
	if opts.Transport != transport {
		t.Error("Expected WithReplayTransport to assign custom RoundTripper")
	}
}

func TestReplayOptionTransportTypedNil(t *testing.T) {
	var transport *mockRoundTripper
	opts := NewReplayOptions(
		WithReplayTransport(transport),
	)
	if opts.Transport != nil {
		t.Error("Expected Transport nil for typed-nil replay transport")
	}
}

// ========== ReplayOption: ReplayAllWith 测试 ==========

func TestReplayAllWith(t *testing.T) {
	_ = createTestHarForOpts() // verify helper works
	// Use dry-run style: ReplayAllWith just delegates to ReplayAll with the
	// assembled options. We verify it constructs options correctly by checking
	// that calling it with custom timeout does not panic.
	// The actual replay will fail because example.com isn't reachable in tests,
	// but we can still verify the option construction path.
	opts := NewReplayOptions(
		WithReplayTimeout(1*time.Nanosecond),
		WithReplaySkipSSLVerify(true),
	)
	if opts.Timeout != 1*time.Nanosecond {
		t.Errorf("Expected timeout 1ns, got %v", opts.Timeout)
	}
	if !opts.SkipSSLVerify {
		t.Error("Expected SkipSSLVerify true")
	}

	// Also verify ReplayAllWith exists and compiles by calling it with a HAR
	// that has no entries — should succeed with empty results.
	emptyHar := NewHar()
	emptyHar.SetCreator("test", "1.0")
	results, err := emptyHar.ReplayAllWith(
		WithReplayTimeout(1 * time.Second),
	)
	if err != nil {
		t.Errorf("ReplayAllWith on empty HAR should not error, got: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 replay results for empty HAR, got %d", len(results))
	}
}

// ========== ConvertOption 测试: WithConvertIncludeBodies ==========

func TestConvertOptionIncludeBodies(t *testing.T) {
	opts := NewConvertOptions(
		WithConvertIncludeBodies(true),
	)
	if !opts.IncludePostData {
		t.Error("Expected IncludePostData true after WithConvertIncludeBodies(true)")
	}

	opts2 := NewConvertOptions(
		WithConvertIncludeBodies(false),
	)
	if opts2.IncludePostData {
		t.Error("Expected IncludePostData false after WithConvertIncludeBodies(false)")
	}
}

// ========== ConvertOption 测试: WithConvertIncludeCookies ==========

func TestConvertOptionIncludeCookies(t *testing.T) {
	// WithConvertIncludeCookies has an empty body (cookies are handled via
	// headers). Verify it doesn't panic and doesn't change other fields.
	opts := NewConvertOptions(
		WithConvertIncludeCookies(true),
	)
	// No specific field to check — the function is a no-op.
	// Just verify it doesn't panic and defaults are preserved.
	if opts.IncludeHeaders {
		t.Error("Expected IncludeHeaders to remain at default (false)")
	}
}

// ========== ConvertOption 测试: WithConvertIncludeQueryStrings ==========

func TestConvertOptionIncludeQueryStrings(t *testing.T) {
	opts := NewConvertOptions(
		WithConvertIncludeQueryStrings(true),
	)
	if !opts.IncludeQueryString {
		t.Error("Expected IncludeQueryString true")
	}

	opts2 := NewConvertOptions(
		WithConvertIncludeQueryStrings(false),
	)
	if opts2.IncludeQueryString {
		t.Error("Expected IncludeQueryString false")
	}
}

// ========== ConvertOption 测试: WithConvertIncludeSize ==========

func TestConvertOptionIncludeSize(t *testing.T) {
	opts := NewConvertOptions(
		WithConvertIncludeSize(true),
	)
	if !opts.IncludeSize {
		t.Error("Expected IncludeSize true")
	}

	opts2 := NewConvertOptions(
		WithConvertIncludeSize(false),
	)
	if opts2.IncludeSize {
		t.Error("Expected IncludeSize false")
	}
}

// ========== ConvertOption 测试: WithConvertIncludeTime ==========

func TestConvertOptionIncludeTime(t *testing.T) {
	opts := NewConvertOptions(
		WithConvertIncludeTime(true),
	)
	if !opts.IncludeTime {
		t.Error("Expected IncludeTime true")
	}

	opts2 := NewConvertOptions(
		WithConvertIncludeTime(false),
	)
	if opts2.IncludeTime {
		t.Error("Expected IncludeTime false")
	}
}

// ========== ConvertOption 测试: WithConvertIncludeMimeType ==========

func TestConvertOptionIncludeMimeType(t *testing.T) {
	opts := NewConvertOptions(
		WithConvertIncludeMimeType(true),
	)
	if !opts.IncludeContentType {
		t.Error("Expected IncludeContentType true")
	}

	opts2 := NewConvertOptions(
		WithConvertIncludeMimeType(false),
	)
	if opts2.IncludeContentType {
		t.Error("Expected IncludeContentType false")
	}
}

// ========== ConvertOption 测试: WithConvertHeaders ==========

func TestConvertOptionHeaders(t *testing.T) {
	customHeaders := []string{"Method", "URL", "Status"}
	opts := NewConvertOptions(
		WithConvertHeaders(customHeaders),
	)
	if len(opts.Headers) != 3 {
		t.Fatalf("Expected 3 headers, got %d", len(opts.Headers))
	}
	if opts.Headers[0] != "Method" || opts.Headers[1] != "URL" || opts.Headers[2] != "Status" {
		t.Errorf("Headers not set correctly: %v", opts.Headers)
	}
}

func TestConvertOptionHeadersEmpty(t *testing.T) {
	opts := NewConvertOptions(
		WithConvertHeaders(nil),
	)
	if opts.Headers != nil {
		t.Errorf("Expected nil Headers, got %v", opts.Headers)
	}
}

// ========== ConvertOption 测试: WithConvertFilter ==========

func TestConvertOptionFilter(t *testing.T) {
	filter := FilterOptions{
		Method:     "GET",
		StatusCode: 200,
	}
	opts := NewConvertOptions(
		WithConvertFilter(filter),
	)
	if opts.Filter == nil {
		t.Fatal("Expected Filter to be set, got nil")
	}
	if opts.Filter.Method != "GET" {
		t.Errorf("Expected Filter.Method 'GET', got '%s'", opts.Filter.Method)
	}
	if opts.Filter.StatusCode != 200 {
		t.Errorf("Expected Filter.StatusCode 200, got %d", opts.Filter.StatusCode)
	}
}

func TestConvertOptionFilterIntegration(t *testing.T) {
	// Test that ConvertWith + WithConvertFilter produces filtered output
	h := createTestHarForOpts()
	filter := FilterOptions{Method: "GET"}
	result, err := h.ConvertWith(FormatText,
		WithConvertFilter(filter),
	)
	if err != nil {
		t.Fatalf("ConvertWith with filter failed: %v", err)
	}
	if result == "" {
		t.Error("Expected non-empty result from ConvertWith with filter")
	}
}

// ========== DiffOption 测试: WithDiffIgnoreCache ==========

func TestDiffOptionIgnoreCache(t *testing.T) {
	opts := NewDiffOptions(
		WithDiffIgnoreCache(true),
	)
	if !opts.IgnoreCache {
		t.Error("Expected IgnoreCache true")
	}

	// Default is already true, verify explicit false
	opts2 := NewDiffOptions(
		WithDiffIgnoreCache(false),
	)
	if opts2.IgnoreCache {
		t.Error("Expected IgnoreCache false")
	}
}

// ========== DiffOption 测试: WithDiffIgnoreComment ==========

func TestDiffOptionIgnoreComment(t *testing.T) {
	opts := NewDiffOptions(
		WithDiffIgnoreComment(true),
	)
	if !opts.IgnoreComment {
		t.Error("Expected IgnoreComment true")
	}

	opts2 := NewDiffOptions(
		WithDiffIgnoreComment(false),
	)
	if opts2.IgnoreComment {
		t.Error("Expected IgnoreComment false")
	}
}

// ========== DiffOption 测试: WithDiffCompareByURL ==========

func TestDiffOptionCompareByURL(t *testing.T) {
	opts := NewDiffOptions(
		WithDiffCompareByURL(true),
	)
	if !opts.CompareByURL {
		t.Error("Expected CompareByURL true")
	}

	opts2 := NewDiffOptions(
		WithDiffCompareByURL(false),
	)
	if opts2.CompareByURL {
		t.Error("Expected CompareByURL false")
	}
}

func TestDiffOptionCompareByURLIntegration(t *testing.T) {
	har1 := createTestHarForOpts()
	har2 := createTestHarForOpts()
	diff := DiffWith(har1, har2,
		WithDiffCompareByURL(true),
		WithDiffIgnoreCache(true),
		WithDiffIgnoreComment(true),
	)
	if diff.HasChanges() {
		t.Error("Expected no changes for identical HARs with CompareByURL")
	}
}

// ========== 辅助函数 ==========

func createTestHarForOpts() *Har {
	h := NewHar()
	h.SetCreator("test", "1.0")
	entry := h.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	entry.AddRequestHeader("Accept", "application/json")
	entry.SetResponseStatus(200, "OK")
	entry.SetResponseContent(100, "application/json")
	entry.SetResponseContentText(`{"status":"ok"}`)
	return h
}
