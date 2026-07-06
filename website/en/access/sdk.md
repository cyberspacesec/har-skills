---
title: Go SDK Access
---

# Go SDK Access

When you need to embed HAR analysis in your own Go program, use the root-package SDK directly: 40 modules, 70+ methods, zero runtime dependencies.

## Import and dependencies

```go
import har "github.com/cyberspacesec/har-skills"
```

::: tip Zero runtime dependencies
The SDK has no third-party runtime deps (only stdlib `encoding/json`, `net/http`, etc.). `testify` is test-only and never enters your build. A clean go.mod makes it safe to embed in dependency-sensitive projects.
:::

## Minimal example

`ParseHarFile` → `Statistics` → `SecurityAudit` — three lines cover the common "parse → overview → audit" flow:

```go
package main

import (
    "fmt"

    har "github.com/cyberspacesec/har-skills"
)

func main() {
    h, err := har.ParseHarFile("capture.har")
    if err != nil {
        panic(err)
    }

    stats := h.Statistics()           // *HarStatistics: entry count, transfer size, status distribution…
    fmt.Printf("entries=%d  size=%d\n", stats.EntryCount, stats.TotalTransferSize)

    report := h.SecurityAudit()       // *SecurityReport
    fmt.Printf("security score=%d/100, findings=%d\n", report.Score, len(report.Findings))
}
```

Common entry points:

| Function | Input | Returns | Use when |
|----------|-------|---------|----------|
| `ParseHarFile(path)` | file path | `*Har, error` | default, standard parse |
| `ParseHarFileAuto(path)` | file path | `*Har, error` | auto-detects gzip |
| `ParseHar(bytes)` | `[]byte` | `*Har, error` | data already in memory |
| `ParseHarFromReader(r)` | `io.Reader` | `*Har, error` | network/pipe source |
| `Parse(bytes, opts...)` | `[]byte` + options | `HARProvider, error` | when you need a strategy |

## Two API styles

The SDK offers both a struct-style and a functional-options style — pick per task.

### Struct style: `Filter(FilterOptions{...})`

Good when config is reused or loaded from a file:

```go
opts := har.FilterOptions{
    Method:      "GET",
    StatusCode:  200,
    URLPattern:  "api/users",
}
result := h.Filter(opts)              // *FilterResult
fmt.Println(result.Count())
```

### Functional options: `FilterWith(WithFilterStatusCode(404))`

Good for one-liners and chaining:

```go
result := h.FilterWith(
    har.WithFilterMethod("GET"),
    har.WithFilterStatusCode(404),
    har.WithFilterRegex(),
).SortByDurationDesc().Limit(10)

for _, e := range result.GetAll() {
    fmt.Println(e.Request.URL)
}
```

::: tip The two are equivalent
`Filter(FilterOptions{Method:"GET", StatusCode:404})` and `FilterWith(WithFilterMethod("GET"), WithFilterStatusCode(404))` are semantically identical — same filtering engine underneath. Functional options just chain better and coexist with defaults.
:::

See [Filtering & Chaining](../sdk/filtering.md) for more, and [Functional Options](../sdk/functional-options.md) for the full option table.

## Four parsing strategies

Pick a strategy via `Parse()` + options; all return the unified `HARProvider` interface:

| Strategy | Trigger option | Trait | Use when |
|----------|----------------|-------|----------|
| standard | default | parses everything into a `*Har` at once | general purpose, modest files |
| optimized | `WithMemoryOptimized()` | compact structs, lower footprint | large files, long-running analysis |
| lazy | `WithLazyLoading()` | fields decoded on demand | large files where few fields are read |
| streaming | `NewStreamingParserFromReader(r)` | iterate entry by entry, no full resident | huge files, ETL |

```go
// standard
provider, err := har.Parse(data)

// memory-optimized
provider, err := har.Parse(data, har.WithMemoryOptimized())

// lazy
provider, err := har.Parse(data, har.WithLazyLoading())

// preset combos: OptFast / OptMemoryEfficient / OptLenient
provider, err := har.ParseFile("large.har", har.OptMemoryEfficient...)
```

::: warning Streaming does not return a full object
`WithStreaming()` cannot return a complete `HARProvider` directly — streaming demands per-entry consumption. Use `NewStreamingParserFromReader(r, opts...)` to get an `EntryIterator` and process entries in a loop. See [Streaming Parsing](../internals/streaming.md).
:::

For a deep comparison see [Parsing Strategies](../sdk/parsing-strategies.md).

## The HARProvider interface

Every strategy returns `HARProvider`, so you program to the abstraction and decide the implementation at runtime:

```go
func analyze(p har.HARProvider) {
    fmt.Println(p.GetVersion())
    for _, e := range p.GetEntries() {
        fmt.Println(e.GetRequest().GetURL())
    }
}

// any strategy's product works
p, _ := har.Parse(data, har.WithLazyLoading()...)
analyze(p)
```

When you need the full `*Har` API (e.g. `SecurityAudit`, `Statistics`), convert with `.ToStandard()`:

```go
provider, _ := har.ParseFile("big.har", har.OptMemoryEfficient...)
h := provider.ToStandard()           // *Har, full API available
report := h.SecurityAudit()
```

`ToStandard()` is idempotent: a standard `*Har` returns itself; optimized/lazy implementations each perform the conversion. See [Provider Interfaces](../sdk/providers.md).

## Next steps

- [Data Structures](../sdk/data-structures.md) — field map for `Har`/`Entries`/`Request`/`Response`
- [Parsing Strategies](../sdk/parsing-strategies.md) — standard/optimized/lazy/streaming in depth
- [Provider Interfaces](../sdk/providers.md) — full `HARProvider` method list
- [Functional Options](../sdk/functional-options.md) — the full `WithFilter*`/`WithReplay*`/`WithConvert*` table
- [API Reference](../sdk/api-reference.md) — index of all 70+ methods
