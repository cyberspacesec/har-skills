---
title: File Operations
titleTemplate: false
---

# File Operations

The 4 Level 2 commands work "between HAR files": compare, merge, split, and validate. `diff` and `merge` take positional args (they bypass `-f`); `split` and `validate` still use `--file`. Some produce new HAR files (merge/split write output), others only reports (diff/validate).

Every example below runs from the repository root against `testdata/example.har` or `testdata/full.har`.

## diff — Compare Two HAR Files

Find requests added, removed, and modified between two HAR files. Takes positional args `<file1> <file2>`, **exactly two** (`cobra.ExactArgs(2)`), and bypasses `-f`.

```bash
har diff testdata/full.har testdata/large.har
```

Match by URL (default is "index + URL" pairing) and compare response bodies:

```bash
har diff testdata/full.har testdata/large.har --compare-by-url --include-body
```

Ignore Cookie and Date header differences:

```bash
har diff a.har b.har --ignore-headers Cookie,Date
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--ignore-headers` | stringSlice | `[]` | Header field names to ignore (comma-separated) |
| `--ignore-timings` | bool | `true` | Ignore timing differences |
| `--ignore-dates` | bool | `true` | Ignore date differences |
| `--include-body` | bool | `false` | Compare response bodies |
| `--compare-by-url` | bool | `false` | Match by URL (default: index + URL) |

::: tip What is ignored by default
`--ignore-timings` and `--ignore-dates` both default to `true` — timestamps and per-phase timings of two captures are almost never identical, so ignoring them by default keeps the focus on "did the requests themselves change". Pass `--ignore-timings=false` explicitly when you do want timing differences.
:::

### How it works

Both files are loaded via `internal.LoadHarFromArg` (gzip auto-detected); options start from `har.DefaultDiffOptions()` and are overridden by each flag; `har.Diff(har1, har2, options)` returns a `*DiffResult`. Output goes through `internal.WriteOutput`: text via `diffResult.Report(har.FormatText)`, CSV via `diffResult.Report(har.FormatCSV)`, JSON serializes the whole `*DiffResult`.

## merge — Merge Multiple HAR Files

Merge entries from multiple HAR files into one, keeping the first file's version and creator info. Takes positional args `<file1> [file2...]`, **at least one** (`cobra.MinimumNArgs(1)`), and bypasses `-f`.

```bash
har merge part1.har part2.har part3.har
```

Merge and deduplicate by Method+URL (keep the newest):

```bash
har merge a.har b.har --deduplicate -o merged.har
```

Don't sort by time (keep the raw concatenation order of the inputs):

```bash
har merge a.har b.har --sort-by-time=false -o raw.har
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--sort-by-time` | bool | `true` | Sort merged entries by startedDateTime |
| `--deduplicate` | bool | `false` | Deduplicate (by Method+URL, keep newest) |

::: tip Output is JSON
`merge` serializes the merged `*Har` straight to JSON (it skips `internal.WriteOutput`'s multi-format branch), so `--format` has no effect on it; use `-o` for a file or leave it empty for stdout. Merge notices (entry counts, etc.) go to stderr.
:::

### How it works

Each positional arg is loaded by `internal.LoadHarFromArg` into `[]*har.Har`; options are packed into `har.MergeOptions{SortByTime, Deduplicate}`; `har.MergeWithOptions(options, hars...)` returns the merged `*Har`, which is then `json.MarshalIndent`'d and written. `--deduplicate` keys on Method+URL, keeping the entry with the latest timestamp.

## split — Split a HAR File

Break a large HAR into smaller files by page reference, domain, time interval, entry count, status code range, or HTTP method. `--by` is **required**.

Split by domain, filename prefix `by-domain`:

```bash
har -f testdata/full.har split --by domain -o by-domain
```

Split by time interval, one file per 30 minutes:

```bash
har -f testdata/full.har split --by time --interval 30m
```

Split by entry count, one file per 50 entries:

```bash
har -f testdata/full.har split --by size --max-entries 50
```

Split by status code range:

```bash
har -f testdata/full.har split --by status
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--by` | string | `""` | Split mode (`page`/`domain`/`time`/`size`/`status`/`method`), required |
| `--interval` | duration | `1h` | Time interval (with `--by time`) |
| `--max-entries` | int | `100` | Max entries per group (with `--by size`) |
| `--output`/`-o` | string | `split` | Output filename prefix (local flag) |

::: tip Filename rules
`split`'s `-o` is a command-local flag (default `split`), distinct from the global `--output`: it is a **filename prefix**, not a single output file. Generated files are named `<prefix>_<kind>_<key>.har` — e.g. `by-domain_domain_api.example.com.har`; ordered splits (`time`/`size`) use a three-digit sequence: `split_time_001.har`. Special characters in `key` (`/ \ : * ? " < > |` and space) are replaced with `_` to avoid invalid filenames. After each file is written, its path and entry count are printed to stderr.
:::

### How it works

Empty `--by` errors out immediately; otherwise `internal.LoadHar` loads the file and `--by` dispatches to `splitByPage` / `splitByDomain` / `splitByTime` / `splitBySize` / `splitByStatus` / `splitByMethod`, each calling the matching `(*Har).SplitByPage()` / `SplitByDomain()` / `SplitByTimeRange(interval)` / `SplitBySize(maxEntries)` / `SplitByStatusCode()` / `SplitByMethod()`. Map-shaped results (page/domain/status/method) go through `writeSplitMap` (named by key); slice-shaped results (time/size) go through `writeSplitSlice` (named by sequence); all are `json.MarshalIndent`'d to disk.

## validate — Validate HAR Spec Compliance

Check a HAR file against the spec: standard mode checks basic structure and required fields; strict mode additionally checks cross-references, HTTP methods, and status-code ranges; a timing-consistency check verifies that the `Time` field matches the sum of the `Timings` phases.

```bash
har -f testdata/full.har validate
```

Strict mode:

```bash
har -f testdata/full.har validate --strict
```

Custom timing tolerance of 5 ms:

```bash
har -f testdata/full.har validate --timings-tolerance 5
```

Strictest (tolerance 0, requiring the timings sum to equal Time exactly):

```bash
har -f testdata/full.har validate --strict --timings-tolerance 0
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--strict` | bool | `false` | Enable strict validation (cross-references / methods / status-code ranges) |
| `--timings-tolerance` | float64 | `10` | Timing consistency tolerance (ms); `0` means exact |

::: tip Tolerance of 0
`--timings-tolerance` is a float64 defaulting to 10 ms. Setting it to `0` requires the sum of the `Timings` phases to **exactly equal** `entry.Time` — capture-tool floating-point error usually trips this check, which is why 10 ms of slack is the default. Only pass `0` when you want absolute strictness.
:::

### How it works

After `internal.LoadHar` loads the file, `--strict` true calls `har.ValidateStrict(h)`, otherwise `har.ValidateHarFile(h)`; the returned `*HarError` is split into `[]*ValidationError` by `collectErrors`. Timing consistency always runs `har.ValidateTimingsConsistency(h, tolerance)` (enabled when tolerance ≥ 0), and its results are merged with the above. Text output is rendered by `formatValidateText`: no errors prints `✓ Valid`; with errors, each is listed as `[Rule] Field: Message`.

## Summary

| Command | Input style | Produces new HAR? |
|---------|-------------|-------------------|
| `diff` | Positional ×2 | No (report) |
| `merge` | Positional ≥1 | Yes (merged HAR) |
| `split` | `--file` | Yes (multiple HARs) |
| `validate` | `--file` | No (report) |

`diff`/`merge` take positional args and bypass `-f`; `split`/`validate` use `--file` and also accept stdin. For security audits or redaction, move on to [Security & Privacy](./security.md).
