package har

import (
	"testing"
)

// --- WithCollectWarnings ---

func TestWithCollectWarnings(t *testing.T) {
	opts := applyOptions(WithCollectWarnings())
	if !opts.collectWarnings {
		t.Error("Expected collectWarnings to be true")
	}
	if opts.lenient {
		t.Error("Expected lenient to remain false (default)")
	}
}

func TestWithCollectWarnings_DefaultIsFalse(t *testing.T) {
	opts := applyOptions()
	if opts.collectWarnings {
		t.Error("Expected collectWarnings to be false by default")
	}
}

func TestWithCollectWarnings_IntegratedWithLenient(t *testing.T) {
	// Often used together with WithLenient, as in OptLenient preset
	opts := applyOptions(WithLenient(), WithCollectWarnings())
	if !opts.lenient {
		t.Error("Expected lenient to be true")
	}
	if !opts.collectWarnings {
		t.Error("Expected collectWarnings to be true")
	}
}

// --- WithMaxWarnings ---

func TestWithMaxWarnings(t *testing.T) {
	opts := applyOptions(WithMaxWarnings(50))
	if opts.maxWarnings != 50 {
		t.Errorf("Expected maxWarnings to be 50, got %d", opts.maxWarnings)
	}
}

func TestWithMaxWarnings_DefaultIs100(t *testing.T) {
	opts := applyOptions()
	if opts.maxWarnings != 100 {
		t.Errorf("Expected default maxWarnings to be 100, got %d", opts.maxWarnings)
	}
}

func TestWithMaxWarnings_Zero(t *testing.T) {
	opts := applyOptions(WithMaxWarnings(0))
	if opts.maxWarnings != 0 {
		t.Errorf("Expected maxWarnings to be 0, got %d", opts.maxWarnings)
	}
}

func TestWithMaxWarnings_LargeValue(t *testing.T) {
	opts := applyOptions(WithMaxWarnings(10000))
	if opts.maxWarnings != 10000 {
		t.Errorf("Expected maxWarnings to be 10000, got %d", opts.maxWarnings)
	}
}

func TestWithMaxWarnings_NegativeValue(t *testing.T) {
	// Negative values are technically allowed by the option
	opts := applyOptions(WithMaxWarnings(-1))
	if opts.maxWarnings != -1 {
		t.Errorf("Expected maxWarnings to be -1, got %d", opts.maxWarnings)
	}
}

func TestWithMaxWarnings_CombinedWithCollectWarnings(t *testing.T) {
	opts := applyOptions(WithCollectWarnings(), WithMaxWarnings(25))
	if !opts.collectWarnings {
		t.Error("Expected collectWarnings to be true")
	}
	if opts.maxWarnings != 25 {
		t.Errorf("Expected maxWarnings to be 25, got %d", opts.maxWarnings)
	}
}

// --- WithStreaming ---

func TestWithStreaming(t *testing.T) {
	opts := applyOptions(WithStreaming())
	if !opts.useStreaming {
		t.Error("Expected useStreaming to be true")
	}
}

func TestWithStreaming_DefaultIsFalse(t *testing.T) {
	opts := applyOptions()
	if opts.useStreaming {
		t.Error("Expected useStreaming to be false by default")
	}
}

func TestWithStreaming_CombinedWithOtherOptions(t *testing.T) {
	opts := applyOptions(WithStreaming(), WithMemoryOptimized(), WithSkipValidation())
	if !opts.useStreaming {
		t.Error("Expected useStreaming to be true")
	}
	if !opts.useMemoryOptimized {
		t.Error("Expected useMemoryOptimized to be true")
	}
	if !opts.skipValidation {
		t.Error("Expected skipValidation to be true")
	}
}

// --- applyOptions general behavior ---

func TestApplyOptions_NoOptions(t *testing.T) {
	opts := applyOptions()
	// Should match defaultOptions
	if opts.lenient != false {
		t.Error("Expected default lenient to be false")
	}
	if opts.skipValidation != false {
		t.Error("Expected default skipValidation to be false")
	}
	if opts.collectWarnings != false {
		t.Error("Expected default collectWarnings to be false")
	}
	if opts.maxWarnings != 100 {
		t.Errorf("Expected default maxWarnings to be 100, got %d", opts.maxWarnings)
	}
	if opts.useMemoryOptimized != false {
		t.Error("Expected default useMemoryOptimized to be false")
	}
	if opts.useLazyLoading != false {
		t.Error("Expected default useLazyLoading to be false")
	}
	if opts.useStreaming != false {
		t.Error("Expected default useStreaming to be false")
	}
	if opts.harVersion != HarSpecVersion12 {
		t.Errorf("Expected default harVersion to be %q, got %q", HarSpecVersion12, opts.harVersion)
	}
	if opts.autoDetectVersion != true {
		t.Error("Expected default autoDetectVersion to be true")
	}
}

func TestApplyOptions_MultipleOptions(t *testing.T) {
	opts := applyOptions(
		WithLenient(),
		WithCollectWarnings(),
		WithMaxWarnings(10),
		WithStreaming(),
		WithMemoryOptimized(),
	)
	if !opts.lenient {
		t.Error("Expected lenient to be true")
	}
	if !opts.collectWarnings {
		t.Error("Expected collectWarnings to be true")
	}
	if opts.maxWarnings != 10 {
		t.Errorf("Expected maxWarnings to be 10, got %d", opts.maxWarnings)
	}
	if !opts.useStreaming {
		t.Error("Expected useStreaming to be true")
	}
	if !opts.useMemoryOptimized {
		t.Error("Expected useMemoryOptimized to be true")
	}
}

func TestApplyOptions_LastOptionWins(t *testing.T) {
	// If same option is applied multiple times, last one wins
	opts := applyOptions(WithMaxWarnings(10), WithMaxWarnings(20))
	if opts.maxWarnings != 20 {
		t.Errorf("Expected maxWarnings to be 20 (last option wins), got %d", opts.maxWarnings)
	}
}

// --- toParseOptions ---

func TestToParseOptions(t *testing.T) {
	opts := applyOptions(
		WithLenient(),
		WithCollectWarnings(),
		WithMaxWarnings(42),
		WithSkipValidation(),
	)
	parseOpts := opts.toParseOptions()

	if !parseOpts.Lenient {
		t.Error("Expected Lenient to be true")
	}
	if !parseOpts.SkipValidation {
		t.Error("Expected SkipValidation to be true")
	}
	if !parseOpts.CollectWarnings {
		t.Error("Expected CollectWarnings to be true")
	}
	if parseOpts.MaxWarnings != 42 {
		t.Errorf("Expected MaxWarnings to be 42, got %d", parseOpts.MaxWarnings)
	}
}

func TestToParseOptions_Defaults(t *testing.T) {
	opts := applyOptions()
	parseOpts := opts.toParseOptions()

	if parseOpts.Lenient {
		t.Error("Expected default Lenient to be false")
	}
	if parseOpts.SkipValidation {
		t.Error("Expected default SkipValidation to be false")
	}
	if parseOpts.CollectWarnings {
		t.Error("Expected default CollectWarnings to be false")
	}
	if parseOpts.MaxWarnings != 100 {
		t.Errorf("Expected default MaxWarnings to be 100, got %d", parseOpts.MaxWarnings)
	}
}

// --- Preset option combinations ---

func TestOptLenient(t *testing.T) {
	opts := applyOptions(OptLenient...)
	if !opts.lenient {
		t.Error("OptLenient should set lenient to true")
	}
	if !opts.collectWarnings {
		t.Error("OptLenient should set collectWarnings to true")
	}
}

func TestOptPerformance(t *testing.T) {
	opts := applyOptions(OptPerformance...)
	if !opts.useMemoryOptimized {
		t.Error("OptPerformance should set useMemoryOptimized to true")
	}
	if !opts.skipValidation {
		t.Error("OptPerformance should set skipValidation to true")
	}
	if !opts.useLazyLoading {
		t.Error("OptPerformance should set useLazyLoading to true")
	}
}

// --- WithHarVersion edge cases ---

func TestWithHarVersion_InvalidVersion(t *testing.T) {
	// Invalid version should be ignored, keeping the default
	opts := applyOptions(WithHarVersion("0.9"))
	if opts.harVersion != HarSpecVersion12 {
		t.Errorf("Expected default version for invalid input, got %q", opts.harVersion)
	}
	if !opts.autoDetectVersion {
		t.Error("Expected autoDetectVersion to remain true when invalid version is provided")
	}
}

func TestWithHarVersion_ValidVersion(t *testing.T) {
	opts := applyOptions(WithHarVersion(HarSpecVersion12))
	if opts.harVersion != HarSpecVersion12 {
		t.Errorf("Expected version %q, got %q", HarSpecVersion12, opts.harVersion)
	}
	if opts.autoDetectVersion {
		t.Error("Expected autoDetectVersion to be false when explicit version is set")
	}
}

func TestWithAutoDetectVersion(t *testing.T) {
	opts := applyOptions(WithAutoDetectVersion(false))
	if opts.autoDetectVersion {
		t.Error("Expected autoDetectVersion to be false")
	}
}

func TestWithAutoDetectVersion_Enable(t *testing.T) {
	opts := applyOptions(WithAutoDetectVersion(true))
	if !opts.autoDetectVersion {
		t.Error("Expected autoDetectVersion to be true")
	}
}

// --- WithLazyLoading ---

func TestWithLazyLoading(t *testing.T) {
	opts := applyOptions(WithLazyLoading())
	if !opts.useLazyLoading {
		t.Error("Expected useLazyLoading to be true")
	}
}

// --- WithMemoryOptimized ---

func TestWithMemoryOptimized(t *testing.T) {
	opts := applyOptions(WithMemoryOptimized())
	if !opts.useMemoryOptimized {
		t.Error("Expected useMemoryOptimized to be true")
	}
}

// --- WithSkipValidation ---

func TestWithSkipValidation(t *testing.T) {
	opts := applyOptions(WithSkipValidation())
	if !opts.skipValidation {
		t.Error("Expected skipValidation to be true")
	}
}

// --- WithLenient ---

func TestWithLenient(t *testing.T) {
	opts := applyOptions(WithLenient())
	if !opts.lenient {
		t.Error("Expected lenient to be true")
	}
}

// --- OptMemoryEfficient preset ---

func TestOptMemoryEfficient(t *testing.T) {
	opts := applyOptions(OptMemoryEfficient...)
	if !opts.useMemoryOptimized {
		t.Error("OptMemoryEfficient should set useMemoryOptimized to true")
	}
	if !opts.skipValidation {
		t.Error("OptMemoryEfficient should set skipValidation to true")
	}
}

// --- OptFast preset ---

func TestOptFast(t *testing.T) {
	opts := applyOptions(OptFast...)
	if !opts.skipValidation {
		t.Error("OptFast should set skipValidation to true")
	}
}
