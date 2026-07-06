---
title: 函数式选项
titleTemplate: false
---

# 函数式选项

har-skills 用 Go 函数式选项（functional options）模式统一了配置入口。`parse.go` 定义 `Option` 类型，`options.go` 提供 `With*` 工厂函数和预设组合。`functional_options.go` 把同一思路推广到 Filter/Replay/Convert/Diff/Merge/Builder，让所有可配置行为都支持两种风格：结构体式与函数式。

## 两种配置风格

SDK 同时保留两条配置路径：

- **结构体式**：`ParseOptions`（`errors.go`）+ `DefaultParseOptions()`。字段直接公开，适合一次性、字段少的场景。
- **函数式式**：`Option`（`parse.go`）+ `With*`（`options.go`）。可变参数，默认值清晰，组合性强。

```go
// 结构体式
opts := har.DefaultParseOptions()
opts.SkipValidation = true
opts.Lenient = true
h, _ := har.ParseHarWithOptions(data, opts)

// 函数式（推荐）
h, _ := har.ParseFile("capture.har",
    har.WithSkipValidation(),
    har.WithLenient(),
)
```

两者背后是同一份配置：函数式 `Option` 内部写一个未导出的 `options` 结构体，最后用 `toParseOptions()` 转成结构体式。函数式不会引入额外开销，只带来更好的可读性与可组合性。

## Option 类型与 With* 函数

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

每个 `With*` 返回一个闭包，修改内部 `options` 的某个字段。`applyOptions(opts ...Option)` 从 `defaultOptions` 出发，依次应用所有闭包：

```go
func applyOptions(opts ...Option) options {
    o := defaultOptions
    for _, opt := range opts {
        opt(&o)
    }
    return o
}
```

各选项的含义：

| 选项 | 作用 | 默认 |
|------|------|------|
| `WithLenient` | 宽松解析，尽量解析有效部分，跳过格式错误 | false |
| `WithSkipValidation` | 跳过 HAR 规范校验 | false |
| `WithCollectWarnings` | 收集警告而非失败（配合 `WithMaxWarnings`） | false |
| `WithMaxWarnings(n)` | 最大警告数，超过则停止 | 100 |
| `WithMemoryOptimized` | 用 optimized 策略 | false |
| `WithLazyLoading` | 用 lazy 策略 | false |
| `WithStreaming` | 用 streaming（`Parse` 会返回 UnsupportedError，需用 `NewStreamingParser`） | false |
| `WithHarVersion(v)` | 指定 HAR 版本，关闭自动检测 | "1.2" |
| `WithAutoDetectVersion(b)` | 是否自动检测版本 | true |

## 预设组合

`options.go` 预定义了四组常用组合，都是 `[]Option`：

```go
var (
    OptMemoryEfficient = []Option{WithMemoryOptimized(), WithSkipValidation()}
    OptFast            = []Option{WithSkipValidation()}
    OptLenient         = []Option{WithLenient(), WithCollectWarnings()}
    OptPerformance     = []Option{WithMemoryOptimized(), WithSkipValidation(), WithLazyLoading()}
)
```

用 `...` 展开传入即可：

```go
// 快速解析，跳过校验
p, err := har.ParseFile("capture.har", har.OptFast...)

// 内存高效：optimized + 跳过校验
p, err = har.ParseFile("large.har", har.OptMemoryEfficient...)

// 高性能：optimized + 跳过校验 + 懒加载
p, err = har.ParseFile("huge.har", har.OptPerformance...)

// 宽松解析：容忍格式问题并收集警告
p, err = har.ParseFile("messy.har", har.OptLenient...)
```

预设是普通切片，可以混用或追加：

```go
p, err := har.ParseFile("c.har", append(har.OptFast, har.WithLazyLoading())...)
```

## Parse / ParseFile 统一入口

`Parse([]byte, opts...)` 与 `ParseFile(path, opts...)` 都收 `...Option`，根据选项分发到四种策略（详见 [解析策略](./parsing-strategies)）：

```go
// standard + 跳过校验
p, _ := har.ParseFile("a.har", har.WithSkipValidation())

// optimized
p, _ = har.ParseFile("b.har", har.WithMemoryOptimized())

// lazy
p, _ = har.ParseFile("c.har", har.WithLazyLoading())

// 三者返回都是 HARProvider，需要 *Har 时 ToStandard()
h := p.ToStandard()
```

## 函数式不只用于解析

`functional_options.go` 把同样的双风格（结构体 + 函数式选项）推广到其余子系统。每个子系统有自己的 `*Option` 类型和 `With*` 工厂。

### Filter

结构体式 `FilterOptions`（`filter.go`）字段：`URL`、`Method`、`StatusCode`、`StatusCodeMin/Max`、`ContentType`、`StartTime/EndTime`、`MinDuration/MaxDuration`、`ResourceType`、`HasError`、`HeaderName/HeaderValue`、`RespHeaderName/RespHeaderValue`、`UseRegex`。

```go
// 结构体式
result := h.Filter(har.FilterOptions{
    Method:    "GET",
    StatusCode: 200,
})

// 函数式
result = h.FilterWith(
    har.WithFilterMethod("GET"),
    har.WithFilterStatusCode(200),
)
```

`With*` 工厂（`functional_options.go`）：`WithFilterURL`、`WithFilterMethod`、`WithFilterStatusCode`、`WithFilterStatusCodeRange`、`WithFilterContentType`、`WithFilterTimeRange`、`WithFilterDuration`、`WithFilterResourceType`、`WithFilterHasError`、`WithFilterHeader`、`WithFilterResponseHeader`、`WithFilterRegex`。

`WithFilterRegex()` 是开关型：单独调用即把 URL 匹配切到正则模式（无需传参）。更多过滤用法见 [过滤与链式结果](./filtering)。

### Replay

```go
// 函数式回放选项
result, err := h.Replay(
    har.WithReplayTimeout(10*time.Second),
    har.WithReplaySkipSSLVerify(true),
    har.WithReplayFollowRedirects(false),
    har.WithReplayMaxRedirects(0),
    har.WithReplayOverrideHeader("Authorization", "Bearer token"),
)
```

`With*`：`WithReplayTimeout`、`WithReplayFollowRedirects`、`WithReplayMaxRedirects`、`WithReplaySkipSSLVerify`、`WithReplayOverrideHeader`、`WithReplayTransport`。

### Convert

```go
out, err := h.Convert(
    har.WithConvertIncludeHeaders(true),
    har.WithConvertIncludeBodies(false),
    har.WithConvertHeaders([]string{"Content-Type", "Authorization"}),
)
```

`With*`：`WithConvertIncludeHeaders`、`WithConvertIncludeTimings`、`WithConvertIncludeBodies`、`WithConvertIncludeCookies`、`WithConvertIncludeQueryStrings`、`WithConvertIncludeStatus`、`WithConvertIncludeSize`、`WithConvertIncludeURL`、`WithConvertIncludeMethod`、`WithConvertIncludeTime`、`WithConvertIncludeMimeType`、`WithConvertHeaders`、`WithConvertFilter`。

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

## 双风格对比

下面用 Filter 同时展示两种写法，做完全等价的过滤——找出 GET 且状态码 200 的请求，按耗时降序取前 10：

```go
// 结构体式
result1 := h.Filter(har.FilterOptions{
    Method:    "GET",
    StatusCode: 200,
}).SortByDurationDesc().Limit(10)

// 函数式
result2 := h.FilterWith(
    har.WithFilterMethod("GET"),
    har.WithFilterStatusCode(200),
).SortByDurationDesc().Limit(10)
```

结构体式适合"字段较多、一次性写完"的本地配置；函数式适合"作为库 API、调用方按需追加"的场景。对公开 API，函数式更友好——新增选项不会破坏既有调用方，而结构体加字段是破坏性变更。

## 何时该用哪种

- **写示例 / 脚本**：结构体式直观，字段一目了然。
- **设计库 API**：函数式。可变参数 + 默认值，向前兼容。
- **复用配置**：把 `[]Option` 存成变量或预设（像 `OptFast`），多处复用。
- **条件拼装**：函数式天然支持按条件追加：

```go
opts := []har.Option{har.WithSkipValidation()}
if memConstraint {
    opts = append(opts, har.WithMemoryOptimized())
}
p, _ := har.ParseFile("c.har", opts...)
```

## 下一步

- 过滤的完整链式 API 与快捷方法，见 [过滤与链式结果](./filtering)。
- 解析策略如何随 `With*` 切换，见 [解析策略](./parsing-strategies)。
