---
title: 差异·合并·拆分
titleTemplate: false
---

# 差异·合并·拆分

这三个模块处理多 HAR 文件的关系：`diff.go` 比较差异、`merge.go` 合并多个 HAR、`Split*` 系列按维度拆分单个 HAR。它们常配合使用——先 diff 看变化、再 merge 凑齐数据、最后 split 按域分发。

## Diff：差异比较

### 包级函数 Diff

`Diff(har1, har2, opts)` 是包级函数，返回 `*HarDiff`，不修改输入。默认按 `Method + URL` 构建键来匹配条目。

```go
func Diff(har1, har2 *Har, options DiffOptions) *HarDiff
```

### DiffOptions

```go
type DiffOptions struct {
    IgnoreHeaders []string // 忽略的头部名（大小写不敏感）
    IgnoreTimings bool     // 忽略时间差异（默认 true）
    IgnoreDates   bool     // 忽略日期差异（默认 true）
    IgnoreCache   bool     // 忽略缓存差异（默认 true）
    IgnoreComment bool     // 忽略注释差异
    NormalizeURL  bool     // URL 归一化（排序查询参数）
    CompareByURL  bool     // 仅按 URL 匹配（默认按 Method+URL）
    IncludeBody   bool     // 比较响应体内容
}
```

`DefaultDiffOptions()` 默认开启 `IgnoreTimings`/`IgnoreDates`/`IgnoreCache`，因为这三类字段在两次抓取间几乎必然不同，关掉后会产生大量噪声。

### HarDiff 结果

```go
type HarDiff struct {
    Added     []DiffEntry     // 新增的请求（har2 有，har1 无）
    Removed   []DiffEntry     // 删除的请求（har1 有，har2 无）
    Modified  []ModifiedEntry // 修改的请求（字段有变化）
    Unchanged int             // 未变更的请求数
}
```

`ModifiedEntry.Changes` 是 `[]FieldChange`，每条记录 `Field`（如 `response.status`、`request.headers.Authorization`）、`OldValue`、`NewValue`。

### 读取与报告

```go
func (d *HarDiff) HasChanges() bool       // 是否有任意变更
func (d *HarDiff) TotalChanges() int      // Added+Removed+Modified 总数
func (d *HarDiff) Report(format ConvertFormat) string // text/markdown/csv 报告
```

### 完整 diff 示例

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
    opts.IgnoreHeaders = []string{"Date", "X-Request-Id"} // 忽略易变头
    opts.IncludeBody = true                                // 比较响应体

    diff := har.Diff(before, after, opts)
    fmt.Printf("总变更: %d (新增 %d, 删除 %d, 修改 %d, 未变 %d)\n",
        diff.TotalChanges(), len(diff.Added), len(diff.Removed), len(diff.Modified), diff.Unchanged)

    // 生成 Markdown 报告
    fmt.Println(diff.Report(har.FormatMarkdown))
}
```

### 函数式 DiffWith

`DiffWith(har1, har2, opts...)` 接收 `DiffOption`，可链式指定：

```go
diff := har.DiffWith(before, after,
    har.WithDiffIgnoreHeaders("Date", "X-Request-Id"),
    har.WithDiffIgnoreTimings(true),
    har.WithDiffIncludeBody(true),
    har.WithDiffCompareByURL(true),
)
```

可用的 `WithDiff*` 选项：`IgnoreHeaders`、`IgnoreTimings`、`IgnoreDates`、`IgnoreCache`、`IgnoreComment`、`NormalizeURL`、`CompareByURL`、`IncludeBody`。

## Merge：合并

### Merge 与 MergeWithOptions

```go
func Merge(hars ...*Har) *Har                                       // 默认选项
func MergeWithOptions(options MergeOptions, hars ...*Har) *Har      // 自定义选项
```

合并后产物继承**第一个非 nil HAR** 的 `Version`、`Creator`、`Browser` 元信息，Pages 与 Entries 直接追加。

### MergeOptions

```go
type MergeOptions struct {
    SortByTime  bool // 合并后按 StartedDateTime 排序（默认 true）
    Deduplicate bool // 按 Method+URL 去重，保留最新的（StartedDateTime 更晚者）
}
```

去重键是 `Method + " " + URL`，命中时保留 `StartedDateTime` 更晚的条目，模拟"同接口多次抓取，保留最新快照"的场景。

### 函数式 MergeWith

`MergeWith(opts...)` 返回一个合并函数，便于注入到 pipeline：

```go
mergeFn := har.MergeWith(
    har.WithMergeSortByTime(true),
    har.WithMergeDeduplicate(true),
)
merged := mergeFn(part1, part2, part3)
```

### 示例

```go
package main

import (
    "github.com/cyberspacesec/har-skills"
)

func main() {
    a, _ := har.ParseHarFile("part1.har")
    b, _ := har.ParseHarFile("part2.har")
    c, _ := har.ParseHarFile("part3.har")

    // 默认合并：按时间排序
    merged := har.Merge(a, b, c)
    _ = merged.SaveToFile("merged.har", true)

    // 去重合并：同 Method+URL 只保留最新
    deduped := har.MergeWithOptions(har.MergeOptions{SortByTime: true, Deduplicate: true}, a, b, c)
    _ = deduped.SaveToFile("deduped.har", true)
}
```

## Split：拆分

六个 `Split*` 方法都挂在 `*Har` 上，返回新 `*Har` 切片或映射，每个分片都继承原 HAR 的 `Version` 与 `Creator`。

### SplitByPage 与 SplitByDomain

返回 `map[string]*Har`，键分别是 `pageref` 与域名。没有 `pageref` 的条目归入空字符串键。

```go
byPage := h.SplitByPage()       // map[pageref]*Har
byDomain := h.SplitByDomain()   // map[domain]*Har

for domain, sub := range byDomain {
    _ = sub.SaveToFile("by-domain/"+domain+".har", true)
}
```

### SplitByTimeRange 与 SplitBySize

返回 `[]*Har` 切片。`SplitByTimeRange(interval)` 先按时间排序，每达到 `interval` 切出一组；`SplitBySize(maxEntries)` 按固定条目数切块。

```go
byTime := h.SplitByTimeRange(time.Hour)    // 每小时一组
bySize := h.SplitBySize(50)                 // 每 50 条一组
```

### SplitByStatusCode 与 SplitByMethod

返回 `map[string]*Har`。`SplitByStatusCode` 按 `2xx/3xx/4xx/5xx` 分组，其它按 `Nxx` 归类；`SplitByMethod` 按请求方法分组。

```go
byStatus := h.SplitByStatusCode()  // map["2xx"]*Har, map["4xx"]*Har, ...
byMethod := h.SplitByMethod()      // map["GET"]*Har, map["POST"]*Har, ...
```

### 完整拆分示例

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

    // 按域名拆分并写盘
    for domain, sub := range h.SplitByDomain() {
        dir := "out/by-domain"
        _ = os.MkdirAll(dir, 0755)
        path := filepath.Join(dir, domain+".har")
        if err := sub.SaveToFile(path, true); err != nil {
            fmt.Fprintln(os.Stderr, err)
            continue
        }
    }

    // 按时间拆分，文件名带序号
    for i, chunk := range h.SplitByTimeRange(30 * time.Minute) {
        path := fmt.Sprintf("out/by-time/part-%03d.har", i)
        _ = chunk.SaveToFile(path, true)
    }

    // 按大小拆分，便于上传有体积限制的系统
    for i, chunk := range h.SplitBySize(100) {
        path := fmt.Sprintf("out/by-size/chunk-%03d.har", i)
        _ = chunk.SaveToFile(path, true)
    }
}
```

## 组合工作流

```go
// 1. 合并三份抓取并去重
merged := har.MergeWithOptions(
    har.MergeOptions{SortByTime: true, Deduplicate: true},
    a, b, c,
)

// 2. 脱敏
safe := merged.Redact(har.DefaultRedactOptions())

// 3. 按域名拆分，分发给各团队
for domain, sub := range safe.SplitByDomain() {
    _ = sub.SaveToFile("dist/"+domain+".har", true)
}
```

## 设计要点

- **不可变输入**：`Diff` 不修改任一输入；`Merge`/`Split*` 产生的子 HAR 是新对象，但 Entries 切片在 `SplitBySize`/`SplitByTimeRange` 中对原切片做了 `copy` 以避免别名。
- **元信息继承**：所有拆分产物都继承原 HAR 的 `Version` 与 `Creator`，保证拆分后的文件仍是合法 HAR。
- **匹配键**：`Diff` 默认用 `Method + URL` 匹配；开启 `NormalizeURL` 会先排序查询参数，避免参数顺序导致的假性差异。
- **去重保留最新**：`MergeWithOptions` 的 `Deduplicate` 在同键冲突时保留 `StartedDateTime` 更晚者，符合"刷新快照"的直觉。
