package har

import (
	"testing"
	"time"
)

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
