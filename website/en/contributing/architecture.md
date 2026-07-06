---
title: Architecture Overview
---

# Architecture Overview

HAR Skills is an AI-agent-oriented HAR analysis toolkit built from three cleanly separated subsystems: the SDK root package, the CLI, and a thin internal glue layer between them. Understanding this layering — plus a handful of design patterns that run through the whole codebase — is the prerequisite for contributing, extending capabilities, or embedding the SDK into your own tools.

## Layered Overview

The top-down call chain: the CLI command layer calls the internal glue layer, which calls the SDK root package, which ultimately grounds out in the HAR spec.

```
   ┌─────────────────────────────────────────────────────────────┐
   │  CLI Layer      cmd/har/cmd        24 Cobra subcommands      │
   │  (info / list / find / security / waterfall / replay ...)    │
   │                                      + root.go (global flags)│
   └──────────────────────────┬──────────────────────────────────┘
                              │ calls
   ┌──────────────────────────┴──────────────────────────────────┐
   │  Internal Glue  cmd/har/internal   loader (load HAR)         │
   │                                    output (format / write)   │
   └──────────────────────────┬──────────────────────────────────┘
                              │ calls
   ┌──────────────────────────┴──────────────────────────────────┐
   │  SDK Root Package  github.com/cyberspacesec/har-skills       │
   │  (package har)  41 .go modules · 4 parsing impls             │
   │                 unified by the HARProvider interface         │
   └──────────────────────────┬──────────────────────────────────┘
                              │ implements
   ┌──────────────────────────┴──────────────────────────────────┐
   │  HAR Spec    1.1 / 1.2 / unofficial 1.3   (JSON-based)       │
   └─────────────────────────────────────────────────────────────┘
```

## Package Structure

### SDK Root Package (`github.com/cyberspacesec/har-skills`, `package har`)

The 41 `.go` module files at the repository root carry all SDK capabilities: parsing (4 strategies), filtering, transformation, redaction, export, diff, merge, split, validation, security audit, cache / cookie / performance / waterfall analysis, builders, statistics, replay, and more. The package exposes two API surfaces — `*Har` and `HARProvider` — and has zero runtime dependencies (`testify` is test-only).

### CLI (`cmd/har/`)

- `cmd/har/main.go`: the entry point — just `cmd.Execute()`.
- `cmd/har/cmd/`: 24 Cobra subcommands plus `root.go` (global flags, Viper config). One file per command: `info.go`, `find.go`, `security.go`, `waterfall.go`, `replay.go`, and so on.
- `cmd/har/internal/`: the glue layer between the CLI and the SDK.
  - `loader.go`: unified HAR loading — `LoadHar` / `LoadHarFromPath` / `LoadHarFromStdin` / `LoadHarFromArg`, handling the `-f` file path, `-` for stdin, and the `HAR_FILE` environment variable.
  - `output.go`: unified formatting and writing — `OutputFormat`, `GetFormat`, `GetOutputPath`, `WriteOutput`, `WriteToFileOrStdout`, plus display helpers like `FormatBytes` and `FormatDuration`.

## Key Design Patterns

### 1. Clone Semantics

Every transformation method on `*Har` returns a new `*Har` instance (clone-then-modify internally); the original is left untouched. The only exceptions are methods whose names end with `InPlace`, which mutate in place. This convention makes chained composition inherently safe — no surprising side effects.

```go
redacted := h.Redact(opts)          // new instance; h unchanged
cleaned := h.RemoveHeaders([...])   // new instance
merged := har.Merge(h1, h2, h3)     // new instance
```

### 2. FilterResult Chaining

`Filter` / `Find*` return a `*FilterResult` that itself offers sorting, limiting, re-filtering, and conversion back to `*Har` — letting you express a query intent as a fluent pipeline.

```go
result := h.FilterWith(
    har.WithFilterMethod("GET"),
    har.WithFilterStatusCodeRange(200, 299),
).SortByDurationDesc().Limit(10)

entries := result.GetAll()   // []Entries
subHar   := result.ToHar()   // back to *Har for further convert / export
```

### 3. HARProvider Unifies 4 Implementations

The four parsing strategies — standard, memory-optimized, lazy, and streaming — all implement the `HARProvider` interface (alongside 10 companion Provider interfaces: `EntryProvider` / `RequestProvider` / `ResponseProvider` / `TimingsProvider`, etc.). `ParseFile` / `Parse` select an implementation via functional options and uniformly return `HARProvider`, decoupling caller code from the concrete strategy. Use `.ToStandard()` when you need the full `*Har` API.

The 10 Provider interfaces defined in `interfaces.go` form the complete object-graph contract:

| Interface | Responsibility |
| --- | --- |
| `HARProvider` | Top-level HAR object: version / creator / browser / pages / entries; `ToStandard()` |
| `EntryProvider` | A single request entry: start time / duration / request / response / timings / page ref |
| `RequestProvider` | Request: method / URL / HTTP version / headers / cookies / query params / PostData |
| `ResponseProvider` | Response: status / status text / headers / content / redirect URL |
| `TimingsProvider` | Timings: blocked / dns / connect / send / wait / receive / ssl |
| `PageProvider` | Page: start time / id / title / page timings |
| `ContentProvider` | Response content: size / mime / text / encoding / compression |
| `CookieProvider` | Cookie: name / value / path / domain / expiry / secure attributes |
| `HeaderProvider` | Header: name / value |
| `PostDataProvider` | Request body: mime / params / text |

```go
provider, _ := har.ParseFile(path, har.WithMemoryOptimized())
for _, ep := range provider.GetEntries() {
    entry := ep.ToStandard()   // EntryProvider → Entries
    // ...generic processing
}
h := provider.ToStandard()     // HARProvider → *Har
```

#### Parse Dispatch Mechanism

`ParseFile` / `Parse` accept `...Option`, reduce them internally into an `options` struct, and select the concrete implementation accordingly. The dispatch logic is roughly:

```
ParseFile(path, opts...)
   │
   ├─ options.Apply(opts)        ──► options{useLazyLoading, useMemoryOptimized, useStreaming, lenient, ...}
   │
   ├─ useStreaming?       ──► StreamingHar (token-advance, Decode per entry)
   ├─ useLazyLoading?     ──► LazyHar       (read metadata first; load entries on demand)
   ├─ useMemoryOptimized? ──► MemoryOptimizedHar (compact structs, buffer reuse)
   └─ otherwise           ──► StandardHar   (json.Unmarshal, full)
   │
   └─ returns HARProvider
```

### 4. Dual API Style

The same capability is exposed in two call styles; pick by scenario:

- **Struct style**: `h.Filter(FilterOptions{Method: "GET"})` — good for static, fully-specified configs.
- **Functional-option style**: `h.FilterWith(WithFilterMethod("GET"), WithFilterURL("api"))` — good when many fields are optional or must be assembled dynamically.

Both styles funnel into the same `FilterOptions` and are fully equivalent in behavior.

#### FilterResult Method Catalog

The chainable and terminal methods on `*FilterResult` (implemented in `filter.go`):

| Method | Category | Description |
| --- | --- | --- |
| `GetAll()` / `First()` / `Last()` / `At(i)` | Terminal | Extract the underlying `[]Entries` or a single entry |
| `Count()` | Terminal | Number of results |
| `SortByTime()` / `SortByDuration()` / `SortByDurationDesc()` | Chain | Sort by time |
| `SortBySize()` / `SortBySizeDesc()` | Chain | Sort by response size |
| `Limit(n)` / `Offset(n)` | Chain | Pagination |
| `Chain(options)` | Chain | Re-filter on top of the current result |
| `ToHar()` | Terminal | Convert back to `*Har` for further convert / export / redact |

### 5. Progressive Disclosure

`CLAUDE.md` is itself an instance of progressive disclosure: from Level 1 basic operations to Level 5 transformation and export, expanding on demand. This docs site follows the same hierarchical navigation, letting both AI agents and human readers move from the simplest usage down into implementation principles.

## Go Version & Dependencies

- **Go version**: 1.19+ (see `go.mod`).
- **CLI dependencies**: `spf13/cobra` v1.8.0 (command framework), `spf13/viper` v1.18.2 (config / env vars).
- **SDK runtime dependencies**: zero. Standard library only.
- **Test dependencies**: `stretchr/testify` v1.8.4, test-only — never enters runtime.

## Module File → Responsibility Cheat Sheet

The 41 root-package module files grouped by function, for quick orientation:

| File | One-line responsibility |
| --- | --- |
| `har.go` | Core types: `Har`, `Entries`, `Request`, `Response`, `Creator`, `Browser`, `Page`, etc. |
| `interfaces.go` | The 10 Provider interfaces: `HARProvider` / `EntryProvider` / `RequestProvider` / `ResponseProvider` / `TimingsProvider`, ... |
| `parser.go` | `*Har` parsing entry family: `ParseHar*` / `ParseHarFile*` / `ParseHarEnhanced` / `ParseHarLenient` / `ParseHarWithWarnings` |
| `reader.go` | Reader / auto-detect entries: `ParseHarFromReader` / `ParseHarFileAuto` / `ParseHarFileGzipped` |
| `parse.go` | `HARProvider` entry family: `Parse` / `ParseFile` + functional-option dispatch |
| `options.go` | Parse options: `Option`, `options`, `WithLenient` / `WithMemoryOptimized` / `WithLazyLoading` / `OptFast`, ... |
| `standard_impl.go` | The standard `HARProvider` implementation |
| `memory.go` | The memory-optimized `HARProvider` implementation |
| `lazy.go` + `lazy_impl.go` | The lazy-loading `HARProvider` and on-demand loading mechanics |
| `streaming.go` | Streaming parsing: `StreamingHar`, `StreamingEntryIterator`, `NewStreamingParserFromFile` |
| `optimized_impl.go` | Shared logic for optimized implementations |
| `filter.go` | Filter engine: `Filter`, `FilterWith`, `FindBy*`, `FindErrors`, `FindSlowRequests`, `FilterResult` chain methods |
| `functional_options.go` | Functional options: `FilterOption`, the `WithFilter*` family |
| `converter.go` | Format conversion: `Convert`, `ConvertFormat` (CSV/Markdown/JSON/YAML/XML…), `ConvertOptions` |
| `export.go` | Export: `ToCurl` / `ToWget` / `ToPythonRequests` / `ToPostmanCollection` / `ToXML` / `ToYAML` / `ToJSON`, ... |
| `transform.go` | Request transformation: `RewriteURL`, `RemoveHeaders`, `AddHeaders`, `Transform` (`TransformRule`) |
| `redact.go` | Redaction: `Redact`, `DefaultRedactOptions` (default targets and replacement text) |
| `builder.go` | Builders: `HarBuilder` / `EntryBuilder`, chain-based HAR construction |
| `generator.go` | High-level generation API: `NewHar`, `SetCreator`, `AddPage`, `AddEntry`, `SaveToFile` |
| `merge.go` | Merge: `Merge` |
| `diff.go` | Diff: `Diff`, `DiffResult`, `DefaultDiffOptions` |
| `dedup.go` | Deduplication: `Dedup`, three strategies (exact / pattern / content-hash) |
| `split.go` (in related modules) | Split: by page / domain / time / size / status / method |
| `statistics.go` | Statistics: `Statistics`, `Summary`, `TimingStatistics`, `DomainSummary`, `StatusCodeDistribution`, ... |
| `security.go` | Security audit: `SecurityAudit`, `SecurityReport`, `FindBySeverity` |
| `cookie.go` | Cookie analysis: `CookieAudit`, `CookieEvolution` |
| `cache.go` | Cache analysis: `CacheAnalysis`, `CacheReport` |
| `performance.go` | Performance scoring: `PerformanceScore`, `PerformanceReport`, `Grade()` |
| `timeline.go` | Timeline and concurrency analysis |
| `content.go` | Content-type analysis |
| `index.go` | Indexing: in-memory index and multiple lookup modes |
| `extensions.go` | Custom-field fidelity (extension fields) |
| `format.go` | Output formatting helpers |
| `decode.go` | Response-body decoding (base64 / chunked, etc.) |
| `http_body.go` | HTTP body handling |
| `replay.go` | Request replay: `Replay`, replay options and results |
| `errors.go` | Enhanced error system: `HarError`, `ErrorCode`, classified errors like `*FileSystemError` |
| `validator.go` + `validator_ext.go` | HAR spec validation: `ValidateHarFile`, `ValidateStrict`, `ValidateTimingsConsistency` |
| `util.go` | General utility functions |
| `doc.go` | Package-level godoc |

> The 24 CLI command files live in `cmd/har/cmd/`, one file per subcommand — the filename is the responsibility (`info.go` → the `info` command, and so on).

## Typical Call Chain

Taking `har -f capture.har find --slow 1000 --format json` as an example, here is the full chain from the command line down to the SDK:

```
user command  har -f capture.har find --slow 1000 --format json
   │
   ├─ cmd/har/cmd/find.go      parse flags (--slow=1000), build filter conditions
   │
   ├─ cmd/har/internal/loader.go  LoadHar(cmd, args)
   │     ├─ read -f path (or - for stdin / HAR_FILE env var)
   │     ├─ har.ParseHarFile(path) → *Har
   │     └─ return *Har
   │
   ├─ SDK root package filter.go
   │     └─ h.FindSlowRequests(1000) → *FilterResult
   │
   ├─ cmd/har/internal/output.go  GetFormat / GetOutputPath
   │     └─ WriteOutput(cmd, result, textFn, csvFn)
   │           ├─ format=json → json.MarshalIndent
   │           ├─ format=csv  → csvFunc()
   │           └─ format=text → textFunc()
   │
   └─ stdout (or the file given via -o)
```

This chain shows the core payoff of the layering: the command layer only cares about "translate flags into an SDK call," the glue layer only cares about "load and format," and the SDK only cares about "business logic." Any layer can be swapped independently — for example, replacing the CLI with an MCP wrapper or an AI Agent Skill requires no changes to the SDK or the glue layer.

## Extension Points

The extension entry points contributors touch most often:

- **New CLI command**: add a `.go` file in `cmd/har/cmd/`, implement a `cobra.Command`, obtain the `*Har` via `internal.LoadHar`, write output via `internal.WriteOutput`, and register it in `root.go`.
- **New filter condition**: add a field to `FilterOptions` in `filter.go`, handle it in the filter engine, and provide a matching `WithFilter*` functional option (`functional_options.go`) plus a `FindBy*` convenience method.
- **New export format**: add a `ToXxx` method in `export.go`, and register it in the `ConvertFormat` constants and switch branches of `converter.go`.
- **New parsing strategy**: implement the `HARProvider` interface (and necessary sub-interfaces), add a selection branch in `ParseFile`'s dispatch logic, and add an `Option` + `With*` functional option.
- **New validation rule**: add a check function in `validator.go` / `validator_ext.go`, and invoke it from `ValidateHarFile` / `ValidateStrict`.

All extension points follow one principle: **interfaces first, clone semantics, dual API style, docs in lockstep**.

## Next Steps

- For the implementation trade-offs of the 4 parsing strategies, see [Parsing Strategies](/en/sdk/parsing-strategies) and the [Internals](/en/internals/memory-optimized) section.
- For the Provider interface design, see [Provider Interfaces](/en/sdk/providers).
- Before submitting code, read the [Contributing Guide](/en/contributing/).
