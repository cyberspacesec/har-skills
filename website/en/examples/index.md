---
title: Examples
---

# Examples

The `examples/` directory ships 7 standalone, runnable Go examples. Each is a `package main` covering a distinct scenario — from the most basic HAR parsing all the way to performance benchmarking, enhanced error handling, and HTML report generation. They are both the fastest on-ramp to the SDK and reusable templates for real-world tasks.

Every example can run against the real HAR files in the repository's `testdata/` directory — no extra data preparation required.

## Example Overview

### 1. Basic Parsing `examples/main.go`

The minimal getting-started example: two most common parsing entry points — from a byte slice and from a file path.

| Field | Detail |
| --- | --- |
| Demonstrates | HAR file parsing, basic output |
| Key API | `har.ParseHar`, `har.ParseHarFile` |
| Source | `examples/main.go` |
| Run | `cd examples && go run main.go` |

### 2. Advanced Usage `examples/advanced/main.go`

A single file that demonstrates filtering, conversion, and creation in one shot — the most complete walk-through of the SDK's "read → filter → convert → write" pipeline.

| Field | Detail |
| --- | --- |
| Demonstrates | Filtering (method / content type / slow / errors), format conversion (CSV / Markdown), building a HAR from scratch |
| Key API | `FindByMethod`, `Filter(FilterOptions{})`, `FindSlowRequests`, `FindErrors`, `Convert`, `NewHar`, `AddEntry`, `SaveToFile` |
| Source | `examples/advanced/main.go` |
| Run | `cd examples/advanced && go run main.go` |

### 3. Mini CLI `examples/cli-tool/main.go`

Builds a 6-command HAR analysis tool (info / list / find / headers / timing / extract) using only the standard library `flag` package — no CLI framework. A reference for embedding the SDK into your own tooling.

| Field | Detail |
| --- | --- |
| Demonstrates | Argument parsing, generic `HARProvider` consumption, `ToStandard()` iteration, JSON/CSV/Text multi-format output |
| Key API | `har.ParseFile`, `har.WithMemoryOptimized`, `HARProvider.GetEntries`, `EntryProvider.ToStandard` |
| Source | `examples/cli-tool/main.go` |
| Run | `cd examples/cli-tool && go run main.go -file ../../testdata/example.har -cmd info` |

### 4. Enhanced Error Handling `examples/error_handling/main.go`

Exercises the SDK's enhanced error system: classified error codes, file-system error detection, lenient partial parsing, and warning collection with nested partial errors.

| Field | Detail |
| --- | --- |
| Demonstrates | Detailed error info, file-system error classification, lenient parsing returning partial results, warning collection |
| Key API | `ParseHarFileEnhanced`, `*HarError` (`GetCode`/`IsFileSystemError`/`HasPartialErrors`/`GetPartialErrors`), `ParseHarFileLenient`, `ParseHarFileWithWarnings` |
| Source | `examples/error_handling/main.go` |
| Run | `cd examples/error_handling && go run main.go` |

### 5. Performance & Parsing Modes `examples/performance/main.go`

Benchmarks the 4 parsing modes (standard / memory-optimized / lazy / streaming) for memory and time, and demonstrates functional options plus generic `HARProvider` consumption.

| Field | Detail |
| --- | --- |
| Demonstrates | 4-mode benchmark comparison, lazy-load trigger timing, streaming iteration, functional options |
| Key API | `har.ParseFile`, `har.WithMemoryOptimized`, `har.WithLazyLoading`, `har.OptFast`, `har.NewStreamingParserFromFile`, `HARProvider` |
| Source | `examples/performance/main.go` |
| Run | `cd examples/performance && go run main.go -file ../../testdata/large.har -mode all` |

### 6. Deep Statistics `examples/statistics/main.go`

Computes a deep `Stats` struct over a HAR: P95, median, success rate, cache-hit rate, per-domain and per-content-type aggregation, plus top slow requests and largest responses.

| Field | Detail |
| --- | --- |
| Demonstrates | Custom statistics struct, percentile computation, per-domain performance aggregation, content-type grouping |
| Key API | `har.ParseFile`, `har.WithMemoryOptimized`, `HARProvider.GetEntries`, `EntryProvider.ToStandard` |
| Source | `examples/statistics/main.go` |
| Run | `cd examples/statistics && go run main.go -file ../../testdata/example.har` |

### 7. Visualization Report `examples/visualization/main.go`

Turns a HAR into a standalone HTML report with Chart.js charts and a colored waterfall — no front-end framework, ready to open.

| Field | Detail |
| --- | --- |
| Demonstrates | Standalone HTML report generation, Chart.js pie charts, colored timeline waterfall |
| Key API | `har.ParseFile`, `HARProvider.GetEntries`, `EntryProvider.ToStandard`, `Timings` field access |
| Source | `examples/visualization/main.go` |
| Run | `cd examples/visualization && go run main.go -file ../../testdata/example.har -output har_viz` |

## Core Snippets of Common Examples

Below are the core snippets of the 3 most commonly used examples for quick copy-paste onboarding.

### Basic Parsing

Two ways to obtain a `*Har` — from bytes or from a file path — the most common entry points. See `examples/main.go`.

```go
package main

import (
	"fmt"
	"os"

	har "github.com/cyberspacesec/har-skills"
)

func main() {
	harFilePath := "./testdata/example.har"

	// Option 1: parse from a byte slice
	harFileBytes, err := os.ReadFile(harFilePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	harFile, err := har.ParseHar(harFileBytes)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(harFile)

	// Option 2: parse directly from a file path
	harFile002, err := har.ParseHarFile(harFilePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(harFile002)
}
```

### Filtering and Conversion

`FilterResult` is chainable; its results can be turned back into a `*Har` via `ToHar()` for format conversion. See `examples/advanced/main.go`.

```go
package main

import (
	"fmt"

	har "github.com/cyberspacesec/har-skills"
)

func main() {
	harFile, err := har.ParseHarFile("../../testdata/example.har")
	if err != nil {
		fmt.Println("failed to parse HAR:", err)
		return
	}

	// Filter all POST requests
	postRequests := harFile.FindByMethod("POST")
	fmt.Printf("found %d POST requests\n", postRequests.Count())

	// Filter all image requests
	imageRequests := harFile.Filter(har.FilterOptions{
		ContentType: "image/",
	})
	fmt.Printf("found %d image requests\n", imageRequests.Count())

	// Filter all slow requests (> 500ms)
	slowRequests := harFile.FindSlowRequests(500)
	fmt.Printf("found %d slow requests (>500ms)\n", slowRequests.Count())

	// Find error requests
	errorRequests := harFile.FindErrors()
	fmt.Printf("found %d error requests\n", errorRequests.Count())

	// Convert the filtered result to CSV
	if slowRequests.Count() > 0 {
		options := har.DefaultConvertOptions()
		options.IncludeTimings = true

		csvData, err := slowRequests.ToHar().Convert(har.FormatCSV, options)
		if err != nil {
			fmt.Println("CSV conversion failed:", err)
		} else {
			fmt.Println("slow requests in CSV:")
			fmt.Println(csvData)
		}
	}
}
```

### Creating a HAR File

`NewHar()` returns a chainable `*Har`. Use the `AddEntry` chain builder to populate request / response / timings, then `SaveToFile` to persist. See `examples/advanced/main.go`.

```go
package main

import (
	"fmt"

	har "github.com/cyberspacesec/har-skills"
)

func main() {
	newHar := har.NewHar()
	newHar.SetCreator("go-har-example", "1.0")

	// Add a page
	page := newHar.AddPage("page1", "demo page")
	page.SetPageTimings(100, 300)

	// Add a request/response entry (chain builder)
	entry := newHar.AddEntry("GET", "https://example.com/api/data", "HTTP/1.1", "page1")
	entry.AddRequestHeader("Accept", "application/json")
	entry.AddRequestHeader("User-Agent", "go-har/1.0")

	entry.SetResponseStatus(200, "OK")
	entry.AddResponseHeader("Content-Type", "application/json")
	entry.SetResponseContent(1024, "application/json")
	entry.SetTimings(10, 20, 30, 5, 50, 30, 25)

	// Save to file (indent=true for pretty output)
	newHarPath := "./generated.har"
	if err := newHar.SaveToFile(newHarPath, true); err != nil {
		fmt.Println("failed to save HAR:", err)
		return
	}
	fmt.Printf("HAR file created and saved to %s\n", newHarPath)
}
```

## Running with testdata

The repository's `testdata/` directory ships several real HAR files you can feed directly to the examples above:

| File | Size | Suitable for |
| --- | --- | --- |
| `testdata/example.har` | small | basic parsing, filtering, statistics, visualization |
| `testdata/full.har` | medium | full-field coverage, complex_test scenarios |
| `testdata/large.har` | ~1.5 MB | performance benchmarking, streaming / lazy comparison |
| `testdata/minimal_valid.har` | tiny | minimal compliance validation |
| `testdata/invalid_date.har` | small | lenient parsing, error-handling examples |
| `testdata/v1.1.har` | small | HAR 1.1 version auto-detection |

When running an example, point its relative path at `testdata/`. For example, to run the performance benchmark from the repo root:

```bash
cd examples/performance
go run main.go -file ../../testdata/large.har -mode all
```

To generate a visualization report:

```bash
cd examples/visualization
go run main.go -file ../../testdata/example.har -output har_viz
# open har_viz/visualization.html in a browser
```

## Next Steps

- For the full filtering / chained-result API, see [Filtering & Chaining](/en/sdk/filtering).
- For the trade-offs of the 4 parsing modes, see [Parsing Strategies](/en/sdk/parsing-strategies).
- To compose these examples into real workflows, see [Workflows & Examples](/en/workflows/security-audit).
