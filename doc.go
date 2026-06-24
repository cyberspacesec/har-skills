// Package har provides a comprehensive SDK for parsing, creating, and manipulating
// HAR (HTTP Archive) files in Go. It implements the HAR specification (versions 1.1, 1.2,
// and unofficial 1.3) and offers a wide range of features for working with HTTP traffic data.
//
// The HAR format is a JSON-based archive format for logging HTTP transactions. It is
// commonly used by web browsers and debugging tools to export network traffic data.
// See https://w3c.github.io/web-performance/specs/HAR/Overview.html for the specification.
//
// # Quick Start
//
// Parsing a HAR file:
//
//	har, err := har.ParseFile("example.har")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Entries: %d\n", har.GetEntryCount())
//
// Creating a new HAR file:
//
//	h := har.NewHar()
//	h.SetCreator("my-app", "1.0")
//	h.AddEntry("GET", "https://example.com/api", "HTTP/2.0", "")
//	h.SaveToFile("output.har", true)
//
// Filtering entries:
//
//	// Struct-based API
//	result := harObj.Filter(har.FilterOptions{StatusCode: 404})
//	fmt.Printf("404 errors: %d\n", result.Count())
//
//	// Functional options API
//	result = harObj.FilterWith(
//	    har.WithFilterStatusCode(404),
//	)
//
// Converting to different formats:
//
//	// Struct-based API
//	text, _ := harObj.Convert(har.FormatMarkdown, har.DefaultConvertOptions())
//
//	// Functional options API
//	text, _ = harObj.ConvertWith(har.FormatMarkdown,
//	    har.WithConvertIncludeHeaders(true),
//	)
//
// Using the builder pattern:
//
//	h := har.NewHarBuilder().
//	    SetCreator("my-tool", "1.0").
//	    AddEntry("GET", "https://example.com/api").
//	    WithResponseStatus(200, "OK").
//	    WithResponseContent(42, "application/json").
//	    EndEntry().
//	    Build()
//
// # Features
//
// This package provides the following capabilities:
//
//   - Parsing: Parse HAR files from bytes, files, readers, gzip-compressed files,
//     or auto-detect format. Supports lenient, enhanced, and streaming modes.
//   - Creation: Build HAR files programmatically using direct struct manipulation,
//     the builder pattern (HarBuilder/EntryBuilder), or the Recorder for live
//     HTTP capture.
//   - Streaming: Process large HAR files entry-by-entry via EntryIterator without
//     loading the entire file into memory.
//   - Lazy loading: Parse large HAR files with LazyHar, which defers loading of
//     response content until explicitly requested.
//   - Memory optimization: Use OptimizedHar with HTTPMethod enum and compact
//     representations to reduce memory footprint.
//   - Filtering: Filter entries by URL, method, status code, content type, time range,
//     duration, headers, cookies, resource type, and more. Supports regex matching
//     and chained filters.
//   - Diff: Compare two HAR files and identify added, removed, and modified requests
//     with configurable comparison options.
//   - Merge and Split: Combine multiple HAR files with deduplication and time sorting.
//     Split by page, domain, time range, size, status code, or HTTP method.
//   - Statistics: Compute request counts, timing summaries (avg/min/max/median/P95/P99),
//     domain distribution, content type distribution, and more.
//   - Validation: Validate HAR files against the spec with standard or strict mode.
//     Register custom validation rules for domain-specific checks.
//   - Replay: Re-execute recorded HTTP requests with configurable timeout, redirects,
//     SSL verification, and header overrides.
//   - Decode: Automatically decode base64-encoded and gzip/deflate-compressed response
//     content.
//   - Export formats: Convert HAR data to CSV, Markdown, HTML, plain text, YAML, JSON,
//     XML, cURL commands, wget commands, Python requests code, and Postman Collections.
//   - Functional options: A modern options API using functional option pattern for
//     parsing, filtering, converting, diffing, merging, and building.
//
// # API Styles
//
// This package offers two complementary API styles:
//
// Legacy struct-based API uses configuration structs passed directly:
//
//	har, err := ParseHarWithOptions(data, ParseOptions{Lenient: true})
//	result := harObj.Filter(FilterOptions{Method: "GET"})
//
// Functional options API uses composable option functions for a more fluent style:
//
//	har, err := Parse(data, WithLenient(), WithSkipValidation())
//	result := harObj.FilterWith(WithFilterMethod("GET"))
//
// Both styles are fully supported and can be mixed. The functional options API is
// recommended for new code.
//
// # HAR Spec and Version Support
//
// The package supports HAR specification versions 1.1, 1.2, and the unofficial 1.3.
// By default, HAR 1.2 is assumed. Version auto-detection is enabled by default when
// using the functional options API.
//
// Key spec points:
//   - The "log" object is the root container; "entries" is an array of request/response pairs.
//   - Required fields: log.version, log.creator, request.method, request.url, response.status.
//   - The "time" field in entries is the total elapsed time in milliseconds.
//   - Negative values in timings indicate that the phase is not applicable.
//   - Content.encoding, when present, must be "base64".
//   - Custom fields prefixed with "_" (e.g., "_initiator", "_priority") are extensions
//     commonly added by browser DevTools and are preserved by this package.
package har
