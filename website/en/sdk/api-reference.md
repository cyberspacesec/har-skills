---
title: API Reference
titleTemplate: false
---

# API Reference

A categorized quick-reference for the exported functions and methods of the root package `har` (`github.com/cyberspacesec/har-skills`). All signatures mirror the source. Items marked "package-level" are package functions; the rest are methods on `*Har` (or the relevant type). `HARProvider` is the unified return interface of the parse functions — call `.ToStandard()` to obtain a `*Har` and access the full API.

## Parsing

| Function / Method | Signature essentials | Returns |
| --- | --- | --- |
| `ParseHarFile` | `ParseHarFile(path string)` · package | `(*Har, error)` |
| `ParseHar` | `ParseHar(data []byte)` · package | `(*Har, error)` |
| `ParseHarFromReader` | `ParseHarFromReader(r io.Reader)` · package | `(*Har, error)` |
| `ParseHarFileAuto` | `ParseHarFileAuto(path)` auto-detects gzip · package | `(*Har, error)` |
| `Parse` | `Parse(data []byte, opts ...Option)` functional · package | `(HARProvider, error)` |
| `ParseFile` | `ParseFile(path string, opts ...Option)` · package | `(HARProvider, error)` |
| `ParseHarWithOptions` | `ParseHarWithOptions(data, options ParseOptions)` | `(*Har, error)` |
| `ParseHarLenient` | `ParseHarLenient(data)` lenient mode · package | `(*Har, error)` |
| `ParseHarWithWarnings` | `ParseHarWithWarnings(data)` returns warnings · package | `(*Result, error)` |
| `ParseHarFileOptimized` | `ParseHarFileOptimized(path)` memory-optimized · package | `(*OptimizedHar, error)` |
| `ParseHarWithLazyLoading` | `ParseHarWithLazyLoading(data)` lazy · package | `(*LazyHar, error)` |
| `NewStreamingParser` | `NewStreamingParser(data, opts...)` · package | `(EntryIterator, error)` |
| `NewStreamingParserFromFile` | `NewStreamingParserFromFile(path, opts...)` · package | `(EntryIterator, error)` |

Functional options (`Option`): `WithLenient`, `WithSkipValidation`, `WithCollectWarnings`, `WithMemoryOptimized`, `WithLazyLoading`, `WithStreaming`, `WithHarVersion`, `WithAutoDetectVersion`.

## Statistics

| Method | Signature essentials | Returns |
| --- | --- | --- |
| `Statistics` | `h.Statistics()` | `*HarStatistics` |
| `Summary` | `h.Summary()` | `string` |
| `TimingStatistics` | `h.TimingStatistics()` | `*TimingsSummary` |
| `DomainSummary` | `h.DomainSummary()` | `map[string]*DomainStats` |
| `StatusCodeDistribution` | `h.StatusCodeDistribution()` | `map[int]int` |
| `MethodDistribution` | `h.MethodDistribution()` | `map[string]int` |
| `ContentTypeDistribution` | `h.ContentTypeDistribution()` | `map[string]int` |
| `SlowestRequests` | `h.SlowestRequests(n int)` | `[]Entries` |
| `FastestRequests` | `h.FastestRequests(n int)` | `[]Entries` |
| `LargestResponses` | `h.LargestResponses(n int)` | `[]Entries` |

## Security

| Method | Signature essentials | Returns |
| --- | --- | --- |
| `SecurityAudit` | `h.SecurityAudit()` | `*SecurityReport` |
| `SecurityAuditWithOptions` | `h.SecurityAuditWithOptions(opts SecurityAuditOptions)` | `*SecurityReport` |
| `CookieAudit` | `h.CookieAudit()` | `*CookieAuditReport` |
| `CookieEvolution` | `h.CookieEvolution()` | `map[string][]CookieEvolutionEntry` |
| `CacheAnalysis` | `h.CacheAnalysis()` | `*CacheReport` |
| `ParseCacheControl` | `ParseCacheControl(value string)` · package | `*CacheControlDirectives` |
| `PerformanceScore` | `h.PerformanceScore()` | `*PerformanceReport` (has `Grade()`) |

`SecurityReport` helpers: `HasHighSeverity()`, `HasMediumSeverity()`, `FindByCategory(cat)`, `FindBySeverity(sev)`, `Summary()`.

## Filtering

| Method | Signature essentials | Returns |
| --- | --- | --- |
| `Filter` | `h.Filter(options FilterOptions)` | `*FilterResult` |
| `FilterWith` | `h.FilterWith(opts ...FilterOption)` functional | `*FilterResult` |
| `FindErrors` | `h.FindErrors()` 4xx/5xx | `*FilterResult` |
| `FindRedirects` | `h.FindRedirects()` 3xx | `*FilterResult` |
| `FindSlowRequests` | `h.FindSlowRequests(minDuration float64)` in ms | `*FilterResult` |
| `FindCacheHits` | `h.FindCacheHits()` | `*FilterResult` |
| `FindByURL` | `h.FindByURL(urlStr string, useRegex bool)` | `*FilterResult` |
| `FindByMethod` | `h.FindByMethod(method string)` | `*FilterResult` |
| `FindByStatusCode` | `h.FindByStatusCode(code int)` | `*FilterResult` |
| `FindByStatusCodeRange` | `h.FindByStatusCodeRange(min, max int)` | `*FilterResult` |
| `FindByDomain` | `h.FindByDomain(domain string)` | `*FilterResult` |
| `FindByContentType` | `h.FindByContentType(contentType string)` | `*FilterResult` |
| `FindByResourceType` | `h.FindByResourceType(resourceType string)` | `*FilterResult` |
| `FindByHeader` | `h.FindByHeader(name, value string)` request header | `*FilterResult` |
| `FindByResponseHeader` | `h.FindByResponseHeader(name, value string)` | `*FilterResult` |
| `FindByCookie` | `h.FindByCookie(name string)` | `*FilterResult` |
| `FindByTimeRange` | `h.FindByTimeRange(start, end time.Time)` | `*FilterResult` |
| `FindByServerIP` | `h.FindByServerIP(ip string)` | `*FilterResult` |
| `FindByConnection` | `h.FindByConnection(connectionID string)` | `*FilterResult` |

`FilterResult` chaining: `Count()`, `First()`, `Last()`, `At(i)`, `SortByTime()`, `SortByDuration()`, `SortByDurationDesc()`, `SortBySize()`, `SortBySizeDesc()`, `Limit(n)`, `Offset(n)`, `Chain(opts)`, `ToHar()`, `GetAll()`.

## Transform & Redact

| Method | Signature essentials | Returns |
| --- | --- | --- |
| `Transform` | `h.Transform(rules []TransformRule)` clone + transform | `*Har` |
| `TransformInPlace` | `h.TransformInPlace(rules)` in place | none |
| `RewriteURL` | `h.RewriteURL(from, to string)` | `*Har` |
| `RemoveHeaders` | `h.RemoveHeaders(names []string)` | `*Har` |
| `AddHeaders` | `h.AddHeaders(headers map[string]string, target string)` | `*Har` |
| `Redact` | `h.Redact(opts RedactOptions)` clone + redact | `*Har` |
| `RedactInPlace` | `h.RedactInPlace(opts)` in place | none |
| `DefaultRedactOptions` | `DefaultRedactOptions()` · package | `RedactOptions` |

The ten `TransformType` values: `TransformURLRewrite`, `TransformHostReplace`, `TransformSchemeChange`, `TransformHeaderAdd`, `TransformHeaderRemove`, `TransformHeaderReplace`, `TransformQueryParamRemove`, `TransformQueryParamAdd`, `TransformCookieDomainRewrite`, `TransformBodyReplace`.

## Export

| Method | Signature essentials | Returns |
| --- | --- | --- |
| `Convert` | `h.Convert(format ConvertFormat, options ConvertOptions)` | `(string, error)` |
| `ConvertWith` | `h.ConvertWith(format, opts ...ConvertOption)` functional | `(string, error)` |
| `ConvertTo` | `h.ConvertTo(format, w io.Writer, options)` streaming | `error` |
| `ToCurl` | `h.ToCurl()` / `e.ToCurl()` | `string` |
| `ToWget` | `h.ToWget()` / `e.ToWget()` | `string` |
| `ToPythonRequests` | `h.ToPythonRequests()` / `e.ToPythonRequests()` | `string` |
| `ToPostmanCollection` | `h.ToPostmanCollection()` Postman v2.1 | `([]byte, error)` |
| `ToXML` | `h.ToXML()` | `(string, error)` |
| `ToYAML` | `h.ToYAML()` | `(string, error)` |
| `ToJSON` | `h.ToJSON(indent bool)` | `([]byte, error)` |

`ConvertFormat` constants: `FormatCSV`, `FormatMarkdown`, `FormatHTML`, `FormatText`, `FormatYAML`, `FormatJSON`. File-writing conveniences: `SaveToFile`, `SaveToFileGzipped`, `SaveToWriter`, `SaveAsPostmanCollection`, `SaveAsXML`, `SaveAsYAML`.

## Diff · Merge · Split

| Function / Method | Signature essentials | Returns |
| --- | --- | --- |
| `Diff` | `Diff(har1, har2 *Har, options DiffOptions)` · package | `*HarDiff` |
| `DiffWith` | `DiffWith(har1, har2, opts ...DiffOption)` functional · package | `*HarDiff` |
| `Merge` | `Merge(hars ...*Har)` · package | `*Har` |
| `MergeWithOptions` | `MergeWithOptions(options MergeOptions, hars ...*Har)` · package | `*Har` |
| `MergeWith` | `MergeWith(opts ...MergeOption) func(hars ...*Har) *Har` · package | merge func |
| `SplitByPage` | `h.SplitByPage()` | `map[string]*Har` |
| `SplitByDomain` | `h.SplitByDomain()` | `map[string]*Har` |
| `SplitByTimeRange` | `h.SplitByTimeRange(interval time.Duration)` | `[]*Har` |
| `SplitBySize` | `h.SplitBySize(maxEntries int)` | `[]*Har` |
| `SplitByStatusCode` | `h.SplitByStatusCode()` | `map[string]*Har` |
| `SplitByMethod` | `h.SplitByMethod()` | `map[string]*Har` |

`HarDiff` methods: `HasChanges()`, `TotalChanges()`, `Report(format ConvertFormat)`. Fields: `Added`/`Removed` `[]DiffEntry`, `Modified` `[]ModifiedEntry`, `Unchanged int`.

## Builder & Recorder

| Function / Method | Signature essentials | Returns |
| --- | --- | --- |
| `NewHar` | `NewHar()` · package | `*Har` |
| `NewHarBuilder` | `NewHarBuilder()` · package | `*HarBuilder` |
| `NewRecorder` | `NewRecorder()` · package | `*Recorder` |
| `HarBuilder.AddEntry` | `b.AddEntry(method, url string)` | `*EntryBuilder` |
| `HarBuilder.AddEntryFromHTTP` | `b.AddEntryFromHTTP(req *http.Request, resp *http.Response, duration time.Duration)` | `*HarBuilder` |
| `HarBuilder.Build` | `b.Build()` | `*Har` |
| `HarBuilder.BuildJSON` | `b.BuildJSON(indent bool)` | `([]byte, error)` |
| `HarBuilder.BuildAndSave` | `b.BuildAndSave(filePath string, indent bool)` | `error` |
| `Recorder.Capture` | `r.Capture(req, resp, duration)` | `*Recorder` |
| `Recorder.CaptureEntry` | `r.CaptureEntry(entry Entries)` | `*Recorder` |
| `Recorder.SaveToFile` | `r.SaveToFile(path string)` | `error` |
| `Recorder.ToJSON` | `r.ToJSON(indent bool)` | `([]byte, error)` |

`EntryBuilder` chain: `WithHTTPVersion`, `WithStartedDateTime`, `WithPageref`, `WithServerIP`, `WithConnection`, `WithComment`, `AddRequestHeader`, `AddResponseHeader`, `AddCookie`, `AddResponseCookie`, `AddQueryParam`, `WithPostData`, `WithPostDataParams`, `WithResponseStatus`, `WithResponseContent`.

## Validation

| Function / Method | Signature essentials | Returns |
| --- | --- | --- |
| `ValidateHarFile` | `ValidateHarFile(h *Har)` · package | `error` |
| `ValidateStrict` | `ValidateStrict(h *Har)` · package | `error` |
| `ValidateTimingsConsistency` | `ValidateTimingsConsistency(h *Har, tolerance float64)` · package | `error` |
| `IsValidHarVersion` | `IsValidHarVersion(version string)` · package | `bool` |
| `DetectHarVersion` | `DetectHarVersion(h *Har)` · package | `string` |

## Decode & Compression

| Method | Signature essentials | Returns |
| --- | --- | --- |
| `DecodeAllContent` | `h.DecodeAllContent()` decode all response bodies | `([][]byte, error)` |
| `DecodeContent` | `e.DecodeContent()` entry / `c.DecodeContent()` Content | `([]byte, error)` |
| `DecodeEntryText` | `e.DecodeEntryText()` decode to text | `(string, error)` |
| `DecompressByEncoding` | `DecompressByEncoding(data, encoding)` · package | `([]byte, error)` |
| `DecompressWithEncoding` | `DecompressWithEncoding(data, contentEncoding)` · package | `([]byte, error)` |
| `CompressContent` | `CompressContent(data, encoding)` · package | `([]byte, error)` |
| `IsCompressed` | `e.IsCompressed()` | `bool` |
| `GetContentEncoding` | `e.GetContentEncoding()` | `string` |
| `IsBase64Encoded` | `c.IsBase64Encoded()` | `bool` |

## Index

| Method | Signature essentials | Returns |
| --- | --- | --- |
| `BuildIndex` | `h.BuildIndex()` build in-memory index | `*HarIndex` |
| `HarIndex.ByURL` | `idx.ByURL(urlStr string)` exact match | `[]*Entries` |
| `HarIndex.ByMethod` | `idx.ByMethod(method string)` | `[]*Entries` |
| `HarIndex.ByStatus` | `idx.ByStatus(code int)` | `[]*Entries` |
| `HarIndex.ByDomain` | `idx.ByDomain(domain string)` | `[]*Entries` |
| `HarIndex.ByMimeType` | `idx.ByMimeType(mime string)` | `[]*Entries` |
| `HarIndex.ByURLPattern` | `idx.ByURLPattern(pattern string)` regex | `[]*Entries` |
| `HarIndex.ByTimeRange` | `idx.ByTimeRange(start, end time.Time)` | `[]*Entries` |
| `HarIndex.Size` | `idx.Size()` | `int` |
| `HarIndex.Stats` | `idx.Stats()` | `IndexStats` |

## Timeline

| Method | Signature essentials | Returns |
| --- | --- | --- |
| `Waterfall` | `h.Waterfall()` | `[]WaterfallEntry` |
| `CriticalPath` | `h.CriticalPath()` critical rendering path | `[]WaterfallEntry` |
| `SLACheck` | `h.SLACheck(rules []SLARule)` | `[]SLAResult` |
| `ConcurrencyTimeline` | `h.ConcurrencyTimeline()` | `[]ConcurrencyPoint` |
| `PageTimingMetrics` | `h.PageTimingMetrics()` | `*PageTimingMetrics` |
| `ConnectionReuse` | `h.ConnectionReuse()` | `map[string][]int` |

## Deduplication

| Method | Signature essentials | Returns |
| --- | --- | --- |
| `FindDuplicates` | `h.FindDuplicates(opts DeduplicateOptions)` | `[]DuplicateGroup` |
| `Deduplicate` | `h.Deduplicate(opts DeduplicateOptions)` returns new `*Har` | `*Har` |
| `IsCacheBusterParam` | `IsCacheBusterParam(name string)` · package | `bool` |
| `IsCacheBusterParamWithValue` | `IsCacheBusterParamWithValue(name, value string)` · package | `bool` |
| `DefaultDeduplicateOptions` | `DefaultDeduplicateOptions()` · package | `DeduplicateOptions` |

`DeduplicateOptions.Strategy` values: `exact`, `pattern`, `content-hash`; configurable via `IgnoreParams`, `CompareHeaders`, `CompareBody`.

## Content

| Method | Signature essentials | Returns |
| --- | --- | --- |
| `ContentSummary` | `h.ContentSummary()` | `*ContentSummary` |
| `ParseJSON` | `c.ParseJSON()` parse response body as `interface{}` | `(interface{}, error)` |
| `ParseAsMap` | `c.ParseAsMap()` | `(map[string]interface{}, error)` |
| `SaveToFile` | `c.SaveToFile(path string)` save response body | `error` |
| `DetectMIMEType` | `c.DetectMIMEType()` | `string` |
| `Hash` | `c.Hash()` content hash | `(string, error)` |
| `MIMECategory` | `c.MIMECategory()` | `MIMECategory` |
| `IsBinary` / `IsText` | `c.IsBinary()` / `c.IsText()` | `bool` |

## Replay

| Method | Signature essentials | Returns |
| --- | --- | --- |
| `ToHTTPRequest` | `e.ToHTTPRequest()` convert back to `*http.Request` | `(*http.Request, error)` |
| `Replay` | `e.Replay(options ReplayOptions)` single replay | `(*ReplayResult, error)` |
| `ReplayAll` | `h.ReplayAll(options)` full replay | `([]*ReplayResult, error)` |
| `ReplaySelective` | `h.ReplaySelective(options, filterOptions)` | `([]*ReplayResult, error)` |
| `ReplayResultsToHar` | `ReplayResultsToHar(results)` · package | `*Har` |
| `HTTPResponseToEntries` | `HTTPResponseToEntries(req, resp, duration)` · package | `*Entries` |

## Utility methods

| Method | Signature essentials | Returns |
| --- | --- | --- |
| `Clone` | `h.Clone()` deep copy | `*Har` |
| `Walk` | `h.Walk(fn func(*Entries) error)` iterate | `error` |
| `GetEntryCount` | `h.GetEntryCount()` | `int` |
| `GetUniqueDomains` | `h.GetUniqueDomains()` sorted | `[]string` |
| `Equals` | `h.Equals(other *Har)` | `bool` |
| `GetHeader` | `r.GetHeader(name)` / `resp.GetHeader(name)` | `string` |
| `HasHeader` | `r.HasHeader(name)` / `resp.HasHeader(name)` | `bool` |
| `GetCookie` | `r.GetCookie(name)` / `resp.GetCookie(name)` | `*Cookie` |
| `GetResponseBody` | `e.GetResponseBody()` auto base64-decodes | `([]byte, error)` |
| `GetRequestBody` | `e.GetRequestBody()` | `[]byte` |
| `GetSize` | `e.GetSize()` total size | `int` |
| `GetDomain` | `e.GetDomain()` | `string` |
| `GetURL` | `e.GetURL()` | `*url.URL` |
| `GetElapsedTime` | `e.GetElapsedTime()` | `time.Duration` |
| `IsError` / `IsRedirect` / `IsSuccess` | `e.IsError()` etc. | `bool` |
| `BuildQueryStringFromURL` | `BuildQueryStringFromURL(rawURL)` · package | `[]QueryString` |
| `ParseResponseHeaders` | `ParseResponseHeaders(headerStr)` · package | `[]Headers` |
| `EstimateHeaderSize` | `EstimateHeaderSize(headers)` · package | `int` |
| `FormatBytes` | `FormatBytes(size)` · package | `string` |
| `ReadBody` | `ReadBody(entry *Entries)` · package | `([]byte, error)` |
| `CloneEntry` | `CloneEntry(entry *Entries)` · package | `*Entries` |
