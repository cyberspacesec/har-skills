---
title: Diff · Merge · Split
titleTemplate: false
---

# Diff · Merge · Split

These three modules manage relationships between HAR files: `diff.go` compares two captures, `merge.go` combines several, and the `Split*` family partitions one by a chosen dimension. They compose naturally — diff to see what changed, merge to assemble a full picture, then split to distribute by domain.

## Diff: comparison

### The package-level Diff function

`Diff(har1, har2, opts)` is a package-level function that returns a `*HarDiff` and leaves both inputs untouched. It matches entries by a `Method + URL` key by default.

```go
func Diff(har1, har2 *Har, options DiffOptions) *HarDiff
```

### DiffOptions

```go
type DiffOptions struct {
    IgnoreHeaders []string // header names to ignore (case-insensitive)
    IgnoreTimings bool     // ignore timing differences (default true)
    IgnoreDates   bool     // ignore date differences (default true)
    IgnoreCache   bool     // ignore cache differences (default true)
    IgnoreComment bool     // ignore comment differences
    NormalizeURL  bool     // normalize URL (sort query params)
    CompareByURL  bool     // match by URL only (default is Method+URL)
    IncludeBody   bool     // compare response bodies
}
```

`DefaultDiffOptions()` turns on `IgnoreTimings`/`IgnoreDates`/`IgnoreCache` because these fields almost always differ between two captures and would otherwise generate noise.

### HarDiff result

```go
type HarDiff struct {
    Added     []DiffEntry     // requests in har2 but not har1
    Removed   []DiffEntry     // requests in har1 but not har2
    Modified  []ModifiedEntry // requests whose fields changed
    Unchanged int             // count of unchanged requests
}
```

`ModifiedEntry.Changes` is a `[]FieldChange`, each recording a `Field` (e.g. `response.status`, `request.headers.Authorization`), `OldValue`, and `NewValue`.

### Reading and reporting

```go
func (d *HarDiff) HasChanges() bool       // any change at all
func (d *HarDiff) TotalChanges() int      // Added+Removed+Modified
func (d *HarDiff) Report(format ConvertFormat) string // text/markdown/csv report
```

### Full diff example

```go
package main

import (
    "fmt"

    har "github.com/cyberspacesec/har-skills"
)

func main() {
    before, _ := har.ParseHarFile("before.har")
    after, _ := har.ParseHarFile("after.har")

    opts := har.DefaultDiffOptions()
    opts.IgnoreHeaders = []string{"Date", "X-Request-Id"} // ignore volatile headers
    opts.IncludeBody = true                                // compare bodies

    diff := har.Diff(before, after, opts)
    fmt.Printf("Total: %d (added %d, removed %d, modified %d, unchanged %d)\n",
        diff.TotalChanges(), len(diff.Added), len(diff.Removed), len(diff.Modified), diff.Unchanged)

    // Markdown report
    fmt.Println(diff.Report(har.FormatMarkdown))
}
```

### Functional DiffWith

`DiffWith(har1, har2, opts...)` takes `DiffOption` values for chaining:

```go
diff := har.DiffWith(before, after,
    har.WithDiffIgnoreHeaders("Date", "X-Request-Id"),
    har.WithDiffIgnoreTimings(true),
    har.WithDiffIncludeBody(true),
    har.WithDiffCompareByURL(true),
)
```

Available `WithDiff*` options: `IgnoreHeaders`, `IgnoreTimings`, `IgnoreDates`, `IgnoreCache`, `IgnoreComment`, `NormalizeURL`, `CompareByURL`, `IncludeBody`.

## Merge: combine

### Merge and MergeWithOptions

```go
func Merge(hars ...*Har) *Har                                       // default options
func MergeWithOptions(options MergeOptions, hars ...*Har) *Har      // custom options
```

The result inherits the `Version`, `Creator`, and `Browser` of the **first non-nil HAR**. Pages and Entries are appended directly.

### MergeOptions

```go
type MergeOptions struct {
    SortByTime  bool // sort merged entries by StartedDateTime (default true)
    Deduplicate bool // dedup by Method+URL, keeping the newest (later StartedDateTime)
}
```

The dedup key is `Method + " " + URL`; on collision the entry with the later `StartedDateTime` wins, modeling "same endpoint captured multiple times, keep the latest snapshot".

### Functional MergeWith

`MergeWith(opts...)` returns a merge function you can inject into a pipeline:

```go
mergeFn := har.MergeWith(
    har.WithMergeSortByTime(true),
    har.WithMergeDeduplicate(true),
)
merged := mergeFn(part1, part2, part3)
```

### Example

```go
package main

import (
    "github.com/cyberspacesec/har-skills"
)

func main() {
    a, _ := har.ParseHarFile("part1.har")
    b, _ := har.ParseHarFile("part2.har")
    c, _ := har.ParseHarFile("part3.har")

    // Default merge: sorted by time
    merged := har.Merge(a, b, c)
    _ = merged.SaveToFile("merged.har", true)

    // Deduplicating merge: keep newest per Method+URL
    deduped := har.MergeWithOptions(har.MergeOptions{SortByTime: true, Deduplicate: true}, a, b, c)
    _ = deduped.SaveToFile("deduped.har", true)
}
```

## Split: partition

Six `Split*` methods hang off `*Har`. Each returns new `*Har` slices or maps whose entries inherit the original `Version` and `Creator`.

### SplitByPage and SplitByDomain

Return `map[string]*Har` keyed by `pageref` or domain. Entries with no `pageref` land under the empty-string key.

```go
byPage := h.SplitByPage()       // map[pageref]*Har
byDomain := h.SplitByDomain()   // map[domain]*Har

for domain, sub := range byDomain {
    _ = sub.SaveToFile("by-domain/"+domain+".har", true)
}
```

### SplitByTimeRange and SplitBySize

Return `[]*Har` slices. `SplitByTimeRange(interval)` sorts by time first, then starts a new group once `interval` elapses. `SplitBySize(maxEntries)` chunks by a fixed entry count.

```go
byTime := h.SplitByTimeRange(time.Hour)    // one group per hour
bySize := h.SplitBySize(50)                 // 50 entries per chunk
```

### SplitByStatusCode and SplitByMethod

Return `map[string]*Har`. `SplitByStatusCode` groups into `2xx/3xx/4xx/5xx` (others fall into `Nxx`); `SplitByMethod` groups by HTTP method.

```go
byStatus := h.SplitByStatusCode()  // map["2xx"]*Har, map["4xx"]*Har, ...
byMethod := h.SplitByMethod()      // map["GET"]*Har, map["POST"]*Har, ...
```

### Full split example

```go
package main

import (
    "fmt"
    "os"
    "path/filepath"
    "time"

    har "github.com/cyberspacesec/har-skills"
)

func main() {
    h, _ := har.ParseHarFile("testdata/large.har")

    // Split by domain and write to disk
    for domain, sub := range h.SplitByDomain() {
        dir := "out/by-domain"
        _ = os.MkdirAll(dir, 0755)
        path := filepath.Join(dir, domain+".har")
        if err := sub.SaveToFile(path, true); err != nil {
            fmt.Fprintln(os.Stderr, err)
            continue
        }
    }

    // Split by time, numbered filenames
    for i, chunk := range h.SplitByTimeRange(30 * time.Minute) {
        path := fmt.Sprintf("out/by-time/part-%03d.har", i)
        _ = chunk.SaveToFile(path, true)
    }

    // Split by size, for size-limited upload systems
    for i, chunk := range h.SplitBySize(100) {
        path := fmt.Sprintf("out/by-size/chunk-%03d.har", i)
        _ = chunk.SaveToFile(path, true)
    }
}
```

## Combined workflow

```go
// 1. Merge three captures and dedup
merged := har.MergeWithOptions(
    har.MergeOptions{SortByTime: true, Deduplicate: true},
    a, b, c,
)

// 2. Redact sensitive data
safe := merged.Redact(har.DefaultRedactOptions())

// 3. Split by domain and hand each team its slice
for domain, sub := range safe.SplitByDomain() {
    _ = sub.SaveToFile("dist/"+domain+".har", true)
}
```

## Design notes

- **Inputs are immutable**: `Diff` never modifies either input. `Merge`/`Split*` produce new objects; `SplitBySize`/`SplitByTimeRange` `copy` their entry slices to avoid aliasing.
- **Metadata inheritance**: every split shard inherits the original `Version` and `Creator`, so the resulting files remain valid HAR.
- **Match key**: `Diff` matches on `Method + URL` by default; `NormalizeURL` sorts query params first so parameter ordering doesn't produce spurious diffs.
- **Newest wins on dedup**: `MergeWithOptions` with `Deduplicate` keeps the entry with the later `StartedDateTime` on a key collision, matching the "refresh snapshot" intuition.
