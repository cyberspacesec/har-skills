---
title: API 速查
titleTemplate: false
---

# API 速查

按功能分类列出根包 `har`（`github.com/cyberspacesec/har-skills`）的主要导出函数与方法。所有签名均与源码对齐；标记为「包级」的是包函数，其余是 `*Har`（或对应类型）的方法。`HARProvider` 是解析函数的统一返回接口，调用 `.ToStandard()` 即可拿到 `*Har` 使用全套 API。

## 解析

| 函数 / 方法 | 签名要点 | 返回值 |
| --- | --- | --- |
| `ParseHarFile` | `ParseHarFile(path string)` · 包级 | `(*Har, error)` |
| `ParseHar` | `ParseHar(data []byte)` · 包级 | `(*Har, error)` |
| `ParseHarFromReader` | `ParseHarFromReader(r io.Reader)` · 包级 | `(*Har, error)` |
| `ParseHarFileAuto` | `ParseHarFileAuto(path)` 自动检测 gzip · 包级 | `(*Har, error)` |
| `Parse` | `Parse(data []byte, opts ...Option)` 函数选项 · 包级 | `(HARProvider, error)` |
| `ParseFile` | `ParseFile(path string, opts ...Option)` · 包级 | `(HARProvider, error)` |
| `ParseHarWithOptions` | `ParseHarWithOptions(data, options ParseOptions)` | `(*Har, error)` |
| `ParseHarLenient` | `ParseHarLenient(data)` 宽松模式 · 包级 | `(*Har, error)` |
| `ParseHarWithWarnings` | `ParseHarWithWarnings(data)` 返回警告 · 包级 | `(*Result, error)` |
| `ParseHarFileOptimized` | `ParseHarFileOptimized(path)` 内存优化 · 包级 | `(*OptimizedHar, error)` |
| `ParseHarWithLazyLoading` | `ParseHarWithLazyLoading(data)` 懒加载 · 包级 | `(*LazyHar, error)` |
| `NewStreamingParser` | `NewStreamingParser(data, opts...)` · 包级 | `(EntryIterator, error)` |
| `NewStreamingParserFromFile` | `NewStreamingParserFromFile(path, opts...)` · 包级 | `(EntryIterator, error)` |

函数选项（`Option`）：`WithLenient`、`WithSkipValidation`、`WithCollectWarnings`、`WithMemoryOptimized`、`WithLazyLoading`、`WithStreaming`、`WithHarVersion`、`WithAutoDetectVersion`。

## 统计

| 方法 | 签名要点 | 返回值 |
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

## 安全

| 方法 | 签名要点 | 返回值 |
| --- | --- | --- |
| `SecurityAudit` | `h.SecurityAudit()` | `*SecurityReport` |
| `SecurityAuditWithOptions` | `h.SecurityAuditWithOptions(opts SecurityAuditOptions)` | `*SecurityReport` |
| `CookieAudit` | `h.CookieAudit()` | `*CookieAuditReport` |
| `CookieEvolution` | `h.CookieEvolution()` | `map[string][]CookieEvolutionEntry` |
| `CacheAnalysis` | `h.CacheAnalysis()` | `*CacheReport` |
| `ParseCacheControl` | `ParseCacheControl(value string)` · 包级 | `*CacheControlDirectives` |
| `PerformanceScore` | `h.PerformanceScore()` | `*PerformanceReport`（含 `Grade()`） |

`SecurityReport` 辅助：`HasHighSeverity()`、`HasMediumSeverity()`、`FindByCategory(cat)`、`FindBySeverity(sev)`、`Summary()`。

## 过滤

| 方法 | 签名要点 | 返回值 |
| --- | --- | --- |
| `Filter` | `h.Filter(options FilterOptions)` | `*FilterResult` |
| `FilterWith` | `h.FilterWith(opts ...FilterOption)` 函数选项 | `*FilterResult` |
| `FindErrors` | `h.FindErrors()` 4xx/5xx | `*FilterResult` |
| `FindRedirects` | `h.FindRedirects()` 3xx | `*FilterResult` |
| `FindSlowRequests` | `h.FindSlowRequests(minDuration float64)` 单位 ms | `*FilterResult` |
| `FindCacheHits` | `h.FindCacheHits()` | `*FilterResult` |
| `FindByURL` | `h.FindByURL(urlStr string, useRegex bool)` | `*FilterResult` |
| `FindByMethod` | `h.FindByMethod(method string)` | `*FilterResult` |
| `FindByStatusCode` | `h.FindByStatusCode(code int)` | `*FilterResult` |
| `FindByStatusCodeRange` | `h.FindByStatusCodeRange(min, max int)` | `*FilterResult` |
| `FindByDomain` | `h.FindByDomain(domain string)` | `*FilterResult` |
| `FindByContentType` | `h.FindByContentType(contentType string)` | `*FilterResult` |
| `FindByResourceType` | `h.FindByResourceType(resourceType string)` | `*FilterResult` |
| `FindByHeader` | `h.FindByHeader(name, value string)` 请求头 | `*FilterResult` |
| `FindByResponseHeader` | `h.FindByResponseHeader(name, value string)` | `*FilterResult` |
| `FindByCookie` | `h.FindByCookie(name string)` | `*FilterResult` |
| `FindByTimeRange` | `h.FindByTimeRange(start, end time.Time)` | `*FilterResult` |
| `FindByServerIP` | `h.FindByServerIP(ip string)` | `*FilterResult` |
| `FindByConnection` | `h.FindByConnection(connectionID string)` | `*FilterResult` |

`FilterResult` 链式方法：`Count()`、`First()`、`Last()`、`At(i)`、`SortByTime()`、`SortByDuration()`、`SortByDurationDesc()`、`SortBySize()`、`SortBySizeDesc()`、`Limit(n)`、`Offset(n)`、`Chain(opts)`、`ToHar()`、`GetAll()`。

## 转换与脱敏

| 方法 | 签名要点 | 返回值 |
| --- | --- | --- |
| `Transform` | `h.Transform(rules []TransformRule)` 克隆+转换 | `*Har` |
| `TransformInPlace` | `h.TransformInPlace(rules)` 原地 | 无 |
| `RewriteURL` | `h.RewriteURL(from, to string)` | `*Har` |
| `RemoveHeaders` | `h.RemoveHeaders(names []string)` | `*Har` |
| `AddHeaders` | `h.AddHeaders(headers map[string]string, target string)` | `*Har` |
| `Redact` | `h.Redact(opts RedactOptions)` 克隆+脱敏 | `*Har` |
| `RedactInPlace` | `h.RedactInPlace(opts)` 原地 | 无 |
| `DefaultRedactOptions` | `DefaultRedactOptions()` · 包级 | `RedactOptions` |

`TransformType` 十种：`TransformURLRewrite`、`TransformHostReplace`、`TransformSchemeChange`、`TransformHeaderAdd`、`TransformHeaderRemove`、`TransformHeaderReplace`、`TransformQueryParamRemove`、`TransformQueryParamAdd`、`TransformCookieDomainRewrite`、`TransformBodyReplace`。

## 导出

| 方法 | 签名要点 | 返回值 |
| --- | --- | --- |
| `Convert` | `h.Convert(format ConvertFormat, options ConvertOptions)` | `(string, error)` |
| `ConvertWith` | `h.ConvertWith(format, opts ...ConvertOption)` 函数选项 | `(string, error)` |
| `ConvertTo` | `h.ConvertTo(format, w io.Writer, options)` 流式 | `error` |
| `ToCurl` | `h.ToCurl()` / `e.ToCurl()` | `string` |
| `ToWget` | `h.ToWget()` / `e.ToWget()` | `string` |
| `ToPythonRequests` | `h.ToPythonRequests()` / `e.ToPythonRequests()` | `string` |
| `ToPostmanCollection` | `h.ToPostmanCollection()` Postman v2.1 | `([]byte, error)` |
| `ToXML` | `h.ToXML()` | `(string, error)` |
| `ToYAML` | `h.ToYAML()` | `(string, error)` |
| `ToJSON` | `h.ToJSON(indent bool)` | `([]byte, error)` |

`ConvertFormat` 常量：`FormatCSV`、`FormatMarkdown`、`FormatHTML`、`FormatText`、`FormatYAML`、`FormatJSON`。写文件便捷方法：`SaveToFile`、`SaveToFileGzipped`、`SaveToWriter`、`SaveAsPostmanCollection`、`SaveAsXML`、`SaveAsYAML`。

## 比对·合并·拆分

| 函数 / 方法 | 签名要点 | 返回值 |
| --- | --- | --- |
| `Diff` | `Diff(har1, har2 *Har, options DiffOptions)` · 包级 | `*HarDiff` |
| `DiffWith` | `DiffWith(har1, har2, opts ...DiffOption)` 函数选项 · 包级 | `*HarDiff` |
| `Merge` | `Merge(hars ...*Har)` · 包级 | `*Har` |
| `MergeWithOptions` | `MergeWithOptions(options MergeOptions, hars ...*Har)` · 包级 | `*Har` |
| `MergeWith` | `MergeWith(opts ...MergeOption) func(hars ...*Har) *Har` · 包级 | 合并函数 |
| `SplitByPage` | `h.SplitByPage()` | `map[string]*Har` |
| `SplitByDomain` | `h.SplitByDomain()` | `map[string]*Har` |
| `SplitByTimeRange` | `h.SplitByTimeRange(interval time.Duration)` | `[]*Har` |
| `SplitBySize` | `h.SplitBySize(maxEntries int)` | `[]*Har` |
| `SplitByStatusCode` | `h.SplitByStatusCode()` | `map[string]*Har` |
| `SplitByMethod` | `h.SplitByMethod()` | `map[string]*Har` |

`HarDiff` 方法：`HasChanges()`、`TotalChanges()`、`Report(format ConvertFormat)`。字段：`Added`/`Removed` `[]DiffEntry`、`Modified` `[]ModifiedEntry`、`Unchanged int`。

## 建造者与录制

| 函数 / 方法 | 签名要点 | 返回值 |
| --- | --- | --- |
| `NewHar` | `NewHar()` · 包级 | `*Har` |
| `NewHarBuilder` | `NewHarBuilder()` · 包级 | `*HarBuilder` |
| `NewRecorder` | `NewRecorder()` · 包级 | `*Recorder` |
| `HarBuilder.AddEntry` | `b.AddEntry(method, url string)` | `*EntryBuilder` |
| `HarBuilder.AddEntryFromHTTP` | `b.AddEntryFromHTTP(req *http.Request, resp *http.Response, duration time.Duration)` | `*HarBuilder` |
| `HarBuilder.Build` | `b.Build()` | `*Har` |
| `HarBuilder.BuildJSON` | `b.BuildJSON(indent bool)` | `([]byte, error)` |
| `HarBuilder.BuildAndSave` | `b.BuildAndSave(filePath string, indent bool)` | `error` |
| `Recorder.Capture` | `r.Capture(req, resp, duration)` | `*Recorder` |
| `Recorder.CaptureEntry` | `r.CaptureEntry(entry Entries)` | `*Recorder` |
| `Recorder.SaveToFile` | `r.SaveToFile(path string)` | `error` |
| `Recorder.ToJSON` | `r.ToJSON(indent bool)` | `([]byte, error)` |

`EntryBuilder` 链式方法：`WithHTTPVersion`、`WithStartedDateTime`、`WithPageref`、`WithServerIP`、`WithConnection`、`WithComment`、`AddRequestHeader`、`AddResponseHeader`、`AddCookie`、`AddResponseCookie`、`AddQueryParam`、`WithPostData`、`WithPostDataParams`、`WithResponseStatus`、`WithResponseContent`。

## 验证

| 函数 / 方法 | 签名要点 | 返回值 |
| --- | --- | --- |
| `ValidateHarFile` | `ValidateHarFile(h *Har)` · 包级 | `error` |
| `ValidateStrict` | `ValidateStrict(h *Har)` · 包级 | `error` |
| `ValidateTimingsConsistency` | `ValidateTimingsConsistency(h *Har, tolerance float64)` · 包级 | `error` |
| `IsValidHarVersion` | `IsValidHarVersion(version string)` · 包级 | `bool` |
| `DetectHarVersion` | `DetectHarVersion(h *Har)` · 包级 | `string` |

## 解码与压缩

| 方法 | 签名要点 | 返回值 |
| --- | --- | --- |
| `DecodeAllContent` | `h.DecodeAllContent()` 解码所有响应体 | `([][]byte, error)` |
| `DecodeContent` | `e.DecodeContent()` 单条目 / `c.DecodeContent()` Content | `([]byte, error)` |
| `DecodeEntryText` | `e.DecodeEntryText()` 解码为文本 | `(string, error)` |
| `DecompressByEncoding` | `DecompressByEncoding(data, encoding)` · 包级 | `([]byte, error)` |
| `DecompressWithEncoding` | `DecompressWithEncoding(data, contentEncoding)` · 包级 | `([]byte, error)` |
| `CompressContent` | `CompressContent(data, encoding)` · 包级 | `([]byte, error)` |
| `IsCompressed` | `e.IsCompressed()` | `bool` |
| `GetContentEncoding` | `e.GetContentEncoding()` | `string` |
| `IsBase64Encoded` | `c.IsBase64Encoded()` | `bool` |

## 索引

| 方法 | 签名要点 | 返回值 |
| --- | --- | --- |
| `BuildIndex` | `h.BuildIndex()` 构建内存索引 | `*HarIndex` |
| `HarIndex.ByURL` | `idx.ByURL(urlStr string)` 精确匹配 | `[]*Entries` |
| `HarIndex.ByMethod` | `idx.ByMethod(method string)` | `[]*Entries` |
| `HarIndex.ByStatus` | `idx.ByStatus(code int)` | `[]*Entries` |
| `HarIndex.ByDomain` | `idx.ByDomain(domain string)` | `[]*Entries` |
| `HarIndex.ByMimeType` | `idx.ByMimeType(mime string)` | `[]*Entries` |
| `HarIndex.ByURLPattern` | `idx.ByURLPattern(pattern string)` 正则 | `[]*Entries` |
| `HarIndex.ByTimeRange` | `idx.ByTimeRange(start, end time.Time)` | `[]*Entries` |
| `HarIndex.Size` | `idx.Size()` | `int` |
| `HarIndex.Stats` | `idx.Stats()` | `IndexStats` |

## 时间线

| 方法 | 签名要点 | 返回值 |
| --- | --- | --- |
| `Waterfall` | `h.Waterfall()` 瀑布流 | `[]WaterfallEntry` |
| `CriticalPath` | `h.CriticalPath()` 关键渲染路径 | `[]WaterfallEntry` |
| `SLACheck` | `h.SLACheck(rules []SLARule)` | `[]SLAResult` |
| `ConcurrencyTimeline` | `h.ConcurrencyTimeline()` | `[]ConcurrencyPoint` |
| `PageTimingMetrics` | `h.PageTimingMetrics()` | `*PageTimingMetrics` |
| `ConnectionReuse` | `h.ConnectionReuse()` 连接复用 | `map[string][]int` |

## 去重

| 方法 | 签名要点 | 返回值 |
| --- | --- | --- |
| `FindDuplicates` | `h.FindDuplicates(opts DeduplicateOptions)` | `[]DuplicateGroup` |
| `Deduplicate` | `h.Deduplicate(opts DeduplicateOptions)` 返回新 `*Har` | `*Har` |
| `IsCacheBusterParam` | `IsCacheBusterParam(name string)` · 包级 | `bool` |
| `IsCacheBusterParamWithValue` | `IsCacheBusterParamWithValue(name, value string)` · 包级 | `bool` |
| `DefaultDeduplicateOptions` | `DefaultDeduplicateOptions()` · 包级 | `DeduplicateOptions` |

`DeduplicateOptions.Strategy` 取值：`exact`、`pattern`、`content-hash`；可配 `IgnoreParams`、`CompareHeaders`、`CompareBody`。

## 内容

| 方法 | 签名要点 | 返回值 |
| --- | --- | --- |
| `ContentSummary` | `h.ContentSummary()` | `*ContentSummary` |
| `ParseJSON` | `c.ParseJSON()` 解析响应体为 `interface{}` | `(interface{}, error)` |
| `ParseAsMap` | `c.ParseAsMap()` | `(map[string]interface{}, error)` |
| `SaveToFile` | `c.SaveToFile(path string)` 保存响应体 | `error` |
| `DetectMIMEType` | `c.DetectMIMEType()` | `string` |
| `Hash` | `c.Hash()` 内容哈希 | `(string, error)` |
| `MIMECategory` | `c.MIMECategory()` | `MIMECategory` |
| `IsBinary` / `IsText` | `c.IsBinary()` / `c.IsText()` | `bool` |

## 重放

| 方法 | 签名要点 | 返回值 |
| --- | --- | --- |
| `ToHTTPRequest` | `e.ToHTTPRequest()` 转回 `*http.Request` | `(*http.Request, error)` |
| `Replay` | `e.Replay(options ReplayOptions)` 单条重放 | `(*ReplayResult, error)` |
| `ReplayAll` | `h.ReplayAll(options)` 全量重放 | `([]*ReplayResult, error)` |
| `ReplaySelective` | `h.ReplaySelective(options, filterOptions)` | `([]*ReplayResult, error)` |
| `ReplayResultsToHar` | `ReplayResultsToHar(results)` · 包级 | `*Har` |
| `HTTPResponseToEntries` | `HTTPResponseToEntries(req, resp, duration)` · 包级 | `*Entries` |

## 工具方法

| 方法 | 签名要点 | 返回值 |
| --- | --- | --- |
| `Clone` | `h.Clone()` 深拷贝 | `*Har` |
| `Walk` | `h.Walk(fn func(*Entries) error)` 遍历 | `error` |
| `GetEntryCount` | `h.GetEntryCount()` | `int` |
| `GetUniqueDomains` | `h.GetUniqueDomains()` 已排序 | `[]string` |
| `Equals` | `h.Equals(other *Har)` | `bool` |
| `GetHeader` | `r.GetHeader(name)` / `resp.GetHeader(name)` | `string` |
| `HasHeader` | `r.HasHeader(name)` / `resp.HasHeader(name)` | `bool` |
| `GetCookie` | `r.GetCookie(name)` / `resp.GetCookie(name)` | `*Cookie` |
| `GetResponseBody` | `e.GetResponseBody()` 自动 base64 解码 | `([]byte, error)` |
| `GetRequestBody` | `e.GetRequestBody()` | `[]byte` |
| `GetSize` | `e.GetSize()` 总大小 | `int` |
| `GetDomain` | `e.GetDomain()` | `string` |
| `GetURL` | `e.GetURL()` | `*url.URL` |
| `GetElapsedTime` | `e.GetElapsedTime()` | `time.Duration` |
| `IsError` / `IsRedirect` / `IsSuccess` | `e.IsError()` 等 | `bool` |
| `BuildQueryStringFromURL` | `BuildQueryStringFromURL(rawURL)` · 包级 | `[]QueryString` |
| `ParseResponseHeaders` | `ParseResponseHeaders(headerStr)` · 包级 | `[]Headers` |
| `EstimateHeaderSize` | `EstimateHeaderSize(headers)` · 包级 | `int` |
| `FormatBytes` | `FormatBytes(size)` · 包级 | `string` |
| `ReadBody` | `ReadBody(entry *Entries)` · 包级 | `([]byte, error)` |
| `CloneEntry` | `CloneEntry(entry *Entries)` · 包级 | `*Entries` |
