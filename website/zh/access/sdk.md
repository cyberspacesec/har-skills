---
title: Go SDK 接入
---

# Go SDK 接入

需要把 HAR 分析嵌入自己的 Go 程序时，直接用根包 SDK：40 个模块、70+ 方法，零运行时依赖。

## 导入与依赖

```go
import har "github.com/cyberspacesec/har-skills"
```

::: tip 零运行时依赖
SDK 运行时不依赖任何第三方库（仅 `encoding/json`、`net/http` 等标准库）。`testify` 仅用于测试，不会进入你的构建产物。go.mod 干净，适合嵌入对依赖敏感的项目。
:::

## 最小示例

`ParseHarFile` → `Statistics` → `SecurityAudit`，三行覆盖最常见的「解析 → 概览 → 审计」：

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

    stats := h.Statistics()           // *HarStatistics：条目数、传输大小、状态码分布…
    fmt.Printf("entries=%d  size=%d\n", stats.EntryCount, stats.TotalTransferSize)

    report := h.SecurityAudit()       // *SecurityReport
    fmt.Printf("security score=%d/100, findings=%d\n", report.Score, len(report.Findings))
}
```

常用入口一览：

| 函数 | 入参 | 返回 | 适用 |
|------|------|------|------|
| `ParseHarFile(path)` | 文件路径 | `*Har, error` | 最常用，标准解析 |
| `ParseHarFileAuto(path)` | 文件路径 | `*Har, error` | 自动识别 gzip |
| `ParseHar(bytes)` | `[]byte` | `*Har, error` | 已在内存中的数据 |
| `ParseHarFromReader(r)` | `io.Reader` | `*Har, error` | 流式来源（网络、管道） |
| `Parse(bytes, opts...)` | `[]byte` + 选项 | `HARProvider, error` | 需要选解析策略时 |

## 双 API 风格

SDK 同时提供结构体式与函数式选项两种写法，按场景选用。

### 结构体式：`Filter(FilterOptions{...})`

适合配置需要复用、或来自配置文件的场景：

```go
opts := har.FilterOptions{
    Method:      "GET",
    StatusCode:  200,
    URLPattern:  "api/users",
}
result := h.Filter(opts)              // *FilterResult
fmt.Println(result.Count())
```

### 函数式选项：`FilterWith(WithFilterStatusCode(404))`

适合一行表达、可链式拼装的场景：

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

::: tip 两套等价
`Filter(FilterOptions{Method:"GET", StatusCode:404})` 与 `FilterWith(WithFilterMethod("GET"), WithFilterStatusCode(404))` 语义完全等价，内部走同一套过滤逻辑。函数式选项只是更易链式与默认值共存。
:::

更多过滤与链式用法见 [过滤与链式结果](../sdk/filtering.md)，函数式选项全表见 [函数式选项](../sdk/functional-options.md)。

## 四种解析策略

通过 `Parse()` + 选项选择不同解析策略，返回统一接口 `HARProvider`：

| 策略 | 触发选项 | 特点 | 适用 |
|------|----------|------|------|
| standard | 默认 | 一次解析全部到 `*Har` | 通用、文件不大 |
| optimized | `WithMemoryOptimized()` | 紧凑结构体，内存占用低 | 大文件、常驻分析 |
| lazy | `WithLazyLoading()` | 字段按需解码 | 只读少量字段的大文件 |
| streaming | `NewStreamingParserFromReader(r)` | 逐条迭代，不全量驻留 | 超大文件、ETL |

```go
// 标准解析
provider, err := har.Parse(data)

// 内存优化解析
provider, err := har.Parse(data, har.WithMemoryOptimized())

// 懒加载解析
provider, err := har.Parse(data, har.WithLazyLoading())

// 预设组合：OptFast / OptMemoryEfficient / OptLenient
provider, err := har.ParseFile("large.har", har.OptMemoryEfficient...)
```

::: warning 流式不返回完整对象
`WithStreaming()` 不能直接返回完整 `HARProvider`——流式要求逐条消费。请改用 `NewStreamingParserFromReader(r, opts...)` 拿到 `EntryIterator`，在循环里处理每条 entry。详见 [流式解析原理](../internals/streaming.md)。
:::

策略选型深入对比见 [四种解析策略](../sdk/parsing-strategies.md)。

## HARProvider 接口

所有解析策略返回 `HARProvider`，面向抽象编程，运行期再决定具体实现：

```go
func analyze(p har.HARProvider) {
    fmt.Println(p.GetVersion())
    for _, e := range p.GetEntries() {
        fmt.Println(e.GetRequest().GetURL())
    }
}

// 任一策略产物都能传入
p, _ := har.Parse(data, har.WithLazyLoading()...)
analyze(p)
```

需要全量 `*Har` API（如 `SecurityAudit`、`Statistics`）时，用 `.ToStandard()` 转回标准形态：

```go
provider, _ := har.ParseFile("big.har", har.OptMemoryEfficient...)
h := provider.ToStandard()           // *Har，可调用全部方法
report := h.SecurityAudit()
```

`ToStandard()` 是幂等的：标准 `*Har` 调用它返回自身，optimized/lazy 实现各自做转换。详见 [Provider 接口](../sdk/providers.md)。

## 下一步

- [数据结构](../sdk/data-structures.md) —— `Har`/`Entries`/`Request`/`Response` 字段图
- [四种解析策略](../sdk/parsing-strategies.md) —— standard/optimized/lazy/streaming 深入
- [Provider 接口](../sdk/providers.md) —— `HARProvider` 全方法
- [函数式选项](../sdk/functional-options.md) —— `WithFilter*`/`WithReplay*`/`WithConvert*` 全表
- [API 速查](../sdk/api-reference.md) —— 70+ 方法索引
