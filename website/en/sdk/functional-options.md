---
title: Functional Options
titleTemplate: false
---

# Functional Options

har-skills uses Go's functional options pattern to unify configuration entry points. `parse.go` defines the `Option` type; `options.go` provides `With*` factory functions and preset combos. `functional_options.go` extends the same idea to Filter/Replay/Convert/Diff/Merge/Builder, so every configurable behavior supports two styles: struct-style and functional.

## Two configuration styles

The SDK keeps both paths available:

- **Struct-style**: `ParseOptions` (`errors.go`) + `DefaultParseOptions()`. Fields are public; good for one-off, low-field-count cases.
- **Functional-style**: `Option` (`parse.go`) + `With*` (`options.go`). Variadic, clear defaults, highly composable.

```go
// Struct-style
opts := har.DefaultParseOptions()
opts.SkipValidation = true
opts.Lenient = true
h, _ := har.ParseHarWithOptions(data, opts)

// Functional (recommended)
h, _ := har.ParseFile("capture.har",
    har.WithSkipValidation(),
    har.WithLenient(),
)
```

Both back onto the same config: the functional `Option` writes an unexported `options` struct internally, finally converted to the struct form via `toParseOptions()`. Functional style adds no overhead — only better readability and composability.

## The Option type and With* functions

```go
// parse.go
type Option func(*options)

// options.go
func WithLenient() Option
func WithSkipValidation() Option
func WithCollectWarnings() Option
func WithMaxWarnings(max int) Option
func WithMemoryOptimized() Option
func WithLazyLoading() Option
func WithStreaming() Option
func WithHarVersion(version string) Option
func WithAutoDetectVersion(enabled bool) Option
```

Each `With*` returns a closure that mutates a field of the internal `options`. `applyOptions(opts ...Option)` starts from `defaultOptions` and applies every closure in turn:

```go
func applyOptions(opts ...Option) options {
    o := defaultOptions
    for _, opt := range opts {
        opt(&o)
    }
    return o
}
```

What each option means:

| Option | Effect | Default |
|--------|--------|---------|
| `WithLenient` | lenient parsing; parse valid parts, skip format errors | false |
| `WithSkipValidation` | skip HAR spec validation | false |
| `WithCollectWarnings` | collect warnings instead of failing (pair with `WithMaxWarnings`) | false |
| `WithMaxWarnings(n)` | max warnings before stopping | 100 |
| `WithMemoryOptimized` | use the optimized strategy | false |
| `WithLazyLoading` | use the lazy strategy | false |
| `WithStreaming` | use streaming (`Parse` returns UnsupportedError; use `NewStreamingParser`) | false |
| `WithHarVersion(v)` | pin the HAR version, disable auto-detection | "1.2" |
| `WithAutoDetectVersion(b)` | whether to auto-detect the version | true |

## Preset combos

`options.go` predefines four common combos, all `[]Option`:

```go
var (
    OptMemoryEfficient = []Option{WithMemoryOptimized(), WithSkipValidation()}
    OptFast            = []Option{WithSkipValidation()}
    OptLenient         = []Option{WithLenient(), WithCollectWarnings()}
    OptPerformance     = []Option{WithMemoryOptimized(), WithSkipValidation(), WithLazyLoading()}
)
```

Spread them in with `...`:

```go
// Fast parse, skip validation
p, err := har.ParseFile("capture.har", har.OptFast...)

// Memory efficient: optimized + skip validation
p, err = har.ParseFile("large.har", har.OptMemoryEfficient...)

// High performance: optimized + skip validation + lazy
p, err = har.ParseFile("huge.har", har.OptPerformance...)

// Lenient: tolerate format issues and collect warnings
p, err = har.ParseFile("messy.har", har.OptLenient...)
```

Presets are ordinary slices — mix or extend freely:

```go
p, err := har.ParseFile("c.har", append(har.OptFast, har.WithLazyLoading())...)
```

## Parse / ParseFile unified entry

Both `Parse([]byte, opts...)` and `ParseFile(path, opts...)` take `...Option` and dispatch to one of the four strategies (see [Parsing strategies](./parsing-strategies)):

```go
// standard + skip validation
p, _ := har.ParseFile("a.har", har.WithSkipValidation())

// optimized
p, _ = har.ParseFile("b.har", har.WithMemoryOptimized())

// lazy
p, _ = har.ParseFile("c.har", har.WithLazyLoading())

// All three return HARProvider; ToStandard() when you need *Har
h := p.ToStandard()
```

## Functional options are not just for parsing

`functional_options.go` extends the same dual style (struct + functional options) to the other subsystems. Each has its own `*Option` type and `With*` factories.

### Filter

Struct-style `FilterOptions` (`filter.go`) fields: `URL`, `Method`, `StatusCode`, `StatusCodeMin/Max`, `ContentType`, `StartTime/EndTime`, `MinDuration/MaxDuration`, `ResourceType`, `HasError`, `HeaderName/HeaderValue`, `RespHeaderName/RespHeaderValue`, `UseRegex`.

```go
// Struct-style
result := h.Filter(har.FilterOptions{
    Method:    "GET",
    StatusCode: 200,
})

// Functional
result = h.FilterWith(
    har.WithFilterMethod("GET"),
    har.WithFilterStatusCode(200),
)
```

`With*` factories (`functional_options.go`): `WithFilterURL`, `WithFilterMethod`, `WithFilterStatusCode`, `WithFilterStatusCodeRange`, `WithFilterContentType`, `WithFilterTimeRange`, `WithFilterDuration`, `WithFilterResourceType`, `WithFilterHasError`, `WithFilterHeader`, `WithFilterResponseHeader`, `WithFilterRegex`.

`WithFilterRegex()` is a toggle: calling it alone switches URL matching into regex mode (no argument needed). See [Filtering and chained results](./filtering) for more.

### Replay

```go
// Functional replay options
result, err := h.Replay(
    har.WithReplayTimeout(10*time.Second),
    har.WithReplaySkipSSLVerify(true),
    har.WithReplayFollowRedirects(false),
    har.WithReplayMaxRedirects(0),
    har.WithReplayOverrideHeader("Authorization", "Bearer token"),
)
```

`With*`: `WithReplayTimeout`, `WithReplayFollowRedirects`, `WithReplayMaxRedirects`, `WithReplaySkipSSLVerify`, `WithReplayOverrideHeader`, `WithReplayTransport`.

### Convert

```go
out, err := h.Convert(
    har.WithConvertIncludeHeaders(true),
    har.WithConvertIncludeBodies(false),
    har.WithConvertHeaders([]string{"Content-Type", "Authorization"}),
)
```

`With*`: `WithConvertIncludeHeaders`, `WithConvertIncludeTimings`, `WithConvertIncludeBodies`, `WithConvertIncludeCookies`, `WithConvertIncludeQueryStrings`, `WithConvertIncludeStatus`, `WithConvertIncludeSize`, `WithConvertIncludeURL`, `WithConvertIncludeMethod`, `WithConvertIncludeTime`, `WithConvertIncludeMimeType`, `WithConvertHeaders`, `WithConvertFilter`.

### Diff

```go
dr := har.Diff(h1, h2,
    har.WithDiffIgnoreHeaders("Cookie", "Date"),
    har.WithDiffIgnoreTimings(true),
    har.WithDiffIgnoreDates(true),
    har.WithDiffIgnoreCache(true),
    har.WithDiffIgnoreComment(true),
    har.WithDiffNormalizeURL(true),
    har.WithDiffCompareByURL(true),
    har.WithDiffIncludeBody(true),
)
```

### Merge

```go
merged := har.Merge(h1, h2, h3,
    har.WithMergeSortByTime(true),
    har.WithMergeDeduplicate(true),
)
```

### Builder

```go
b := har.NewHarBuilderWithOptions(
    har.WithBuilderVersion("1.2"),
    har.WithBuilderCreator("my-tool", "1.0.0"),
    har.WithBuilderBrowser("Chrome", "120.0"),
    har.WithBuilderComment("generated"),
)
```

## Dual-style comparison

Below, Filter shows both styles doing the exact same thing — find GET requests with status 200, sorted by duration descending, take the top 10:

```go
// Struct-style
result1 := h.Filter(har.FilterOptions{
    Method:    "GET",
    StatusCode: 200,
}).SortByDurationDesc().Limit(10)

// Functional
result2 := h.FilterWith(
    har.WithFilterMethod("GET"),
    har.WithFilterStatusCode(200),
).SortByDurationDesc().Limit(10)
```

Struct-style suits "many fields, written once" local config; functional suits "library API, caller appends as needed". For public APIs, functional is friendlier — adding an option does not break existing callers, whereas adding a struct field is a breaking change.

## When to use which

- **Examples / scripts**: struct-style is direct, fields at a glance.
- **Designing a library API**: functional. Variadic + defaults, forward-compatible.
- **Reusing config**: store `[]Option` as a variable or preset (like `OptFast`) and reuse across sites.
- **Conditional assembly**: functional supports appending by condition naturally:

```go
opts := []har.Option{har.WithSkipValidation()}
if memConstraint {
    opts = append(opts, har.WithMemoryOptimized())
}
p, _ := har.ParseFile("c.har", opts...)
```

## Next steps

- For the full chained API and shortcut methods of filtering, see [Filtering and chained results](./filtering).
- For how parsing strategy switches with `With*`, see [Parsing strategies](./parsing-strategies).
