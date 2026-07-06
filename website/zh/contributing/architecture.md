---
title: 架构总览
---

# 架构总览

HAR Skills 是一个面向 AI Agent 的 HAR 分析工具箱，由三个清晰分层的子系统构成：SDK 根包、CLI、以及二者之间的内部胶水层。理解这一分层与几条贯穿全局的设计模式，是参与贡献、扩展能力或嵌入自研工具的前提。

## 分层总览

自顶向下的调用链：CLI 命令层调用内部胶水层，胶水层调用 SDK 根包，根包最终落地到 HAR 规范。

```
   ┌─────────────────────────────────────────────────────────────┐
   │  CLI 命令层   cmd/har/cmd        24 个 Cobra 子命令 + root  │
   │  (info / list / find / security / waterfall / replay ...)   │
   └──────────────────────────┬──────────────────────────────────┘
                              │ 调用
   ┌──────────────────────────┴──────────────────────────────────┐
   │  内部胶水层  cmd/har/internal   loader（载入 HAR）           │
   │                                  output（格式化 / 写出）     │
   └──────────────────────────┬──────────────────────────────────┘
                              │ 调用
   ┌──────────────────────────┴──────────────────────────────────┐
   │  SDK 根包    github.com/cyberspacesec/har-skills (package har)│
   │  41 个 .go 模块 · 4 种解析实现 · HARProvider 接口统一        │
   └──────────────────────────┬──────────────────────────────────┘
                              │ 实现
   ┌──────────────────────────┴──────────────────────────────────┐
   │  HAR 规范    1.1 / 1.2 / 非官方 1.3   (JSON-based)           │
   └─────────────────────────────────────────────────────────────┘
```

## 包结构

### SDK 根包（`github.com/cyberspacesec/har-skills`，`package har`）

仓库根目录下的 41 个 `.go` 模块文件承载了全部 SDK 能力：解析（4 种策略）、过滤、转换、脱敏、导出、差异、合并、拆分、校验、安全审计、缓存 / Cookie / 性能 / 瀑布图分析、构建器、统计、重放等。该包对外暴露 `*Har` 与 `HARProvider` 两套 API，运行时零外部依赖（`testify` 仅用于测试）。

### CLI（`cmd/har/`）

- `cmd/har/main.go`：入口，仅 `cmd.Execute()`。
- `cmd/har/cmd/`：24 个 Cobra 子命令 + `root.go`（全局 flag、Viper 配置）。每个命令一个文件，如 `info.go`、`find.go`、`security.go`、`waterfall.go`、`replay.go`。
- `cmd/har/internal/`：CLI 与 SDK 之间的胶水层。
  - `loader.go`：统一 HAR 载入逻辑——`LoadHar` / `LoadHarFromPath` / `LoadHarFromStdin` / `LoadHarFromArg`，处理 `-f` 文件、`-` 标准输入、`HAR_FILE` 环境变量。
  - `output.go`：统一格式化与写出——`OutputFormat`、`GetFormat`、`GetOutputPath`、`WriteOutput`、`WriteToFileOrStdout`，以及 `FormatBytes` / `FormatDuration` 等展示辅助。

## 关键设计模式

### 1. 克隆语义

所有 `*Har` 上的变换方法都返回一个新的 `*Har` 实例（内部克隆后修改），原始对象保持不变。例外是名字以 `InPlace` 结尾的方法，它们就地修改。这一约定让链式组合天然安全，无需担心副作用。

```go
redacted := h.Redact(opts)          // 新实例，h 不变
cleaned := h.RemoveHeaders([...])   // 新实例
merged := har.Merge(h1, h2, h3)     // 新实例
```

### 2. FilterResult 链式

`Filter` / `Find*` 系列返回 `*FilterResult`，它自身提供排序、限量、再过滤、转回 `*Har` 的链式方法，支持流式表达查询意图。

```go
result := h.FilterWith(
    har.WithFilterMethod("GET"),
    har.WithFilterStatusCodeRange(200, 299),
).SortByDurationDesc().Limit(10)

entries := result.GetAll()   // []Entries
subHar   := result.ToHar()   // 转回 *Har，可继续转换 / 导出
```

### 3. HARProvider 接口统一 4 种实现

标准、内存优化、懒加载、流式 4 种解析策略都实现 `HARProvider` 接口（以及配套的 `EntryProvider` / `RequestProvider` / `ResponseProvider` 等 10 个 Provider 接口）。`ParseFile` / `Parse` 通过函数式选项选择实现，统一返回 `HARProvider`，调用方代码与具体策略解耦。需要完整 `*Har` API 时用 `.ToStandard()` 转换。

`interfaces.go` 中定义的 10 个 Provider 接口构成完整的对象图契约：

| 接口 | 职责 |
| --- | --- |
| `HARProvider` | HAR 顶层对象：版本 / 创建者 / 浏览器 / 页面 / 条目，可 `ToStandard()` |
| `EntryProvider` | 单个请求条目：开始时间 / 时长 / 请求 / 响应 / 计时 / 页面引用 |
| `RequestProvider` | 请求：方法 / URL / HTTP 版本 / 头 / Cookie / 查询参数 / PostData |
| `ResponseProvider` | 响应：状态 / 状态文本 / 头 / 内容 / 重定向 URL |
| `TimingsProvider` | 计时：blocked / dns / connect / send / wait / receive / ssl |
| `PageProvider` | 页面：开始时间 / ID / 标题 / 页面计时 |
| `ContentProvider` | 响应内容：大小 / MIME / 文本 / 编码 / 压缩 |
| `CookieProvider` | Cookie：名 / 值 / 路径 / 域 / 过期 / 安全属性 |
| `HeaderProvider` | 头：名 / 值 |
| `PostDataProvider` | 请求体：MIME / 参数 / 文本 |

```go
provider, _ := har.ParseFile(path, har.WithMemoryOptimized())
for _, ep := range provider.GetEntries() {
    entry := ep.ToStandard()   // EntryProvider → Entries
    // ...通用处理
}
h := provider.ToStandard()     // HARProvider → *Har
```

#### 解析分发机制

`ParseFile` / `Parse` 接收 `...Option`，在内部把选项归约成 `options` 结构，再据此选择具体实现。分发逻辑大致如下：

```
ParseFile(path, opts...)
   │
   ├─ options.Apply(opts)        ──► options{useLazyLoading, useMemoryOptimized, useStreaming, lenient, ...}
   │
   ├─ useStreaming?  ──► StreamingHar（token 推进，逐条 Decode）
   ├─ useLazyLoading? ──► LazyHar（首屏只读元信息，entry 按需加载）
   ├─ useMemoryOptimized? ──► MemoryOptimizedHar（紧凑结构，复用缓冲）
   └─ 否则           ──► StandardHar（json.Unmarshal 全量）
   │
   └─ 返回 HARProvider
```

### 4. 双 API 风格

同一能力提供两种调用风格，按场景择优：

- **结构体式**：`h.Filter(FilterOptions{Method: "GET"})`，适合静态、完整配置。
- **函数式选项式**：`h.FilterWith(WithFilterMethod("GET"), WithFilterURL("api"))`，适合可选字段多、需要动态拼装的场景。

两套 API 最终落到同一份 `FilterOptions`，行为完全等价。

#### FilterResult 方法清单

`*FilterResult` 提供的链式与终止方法（实现于 `filter.go`）：

| 方法 | 类别 | 说明 |
| --- | --- | --- |
| `GetAll()` / `First()` / `Last()` / `At(i)` | 终止 | 取出底层 `[]Entries` 或单条 |
| `Count()` | 终止 | 结果数量 |
| `SortByTime()` / `SortByDuration()` / `SortByDurationDesc()` | 链式 | 按时间排序 |
| `SortBySize()` / `SortBySizeDesc()` | 链式 | 按响应大小排序 |
| `Limit(n)` / `Offset(n)` | 链式 | 分页 |
| `Chain(options)` | 链式 | 在当前结果上再过滤一轮 |
| `ToHar()` | 终止 | 转回 `*Har`，可继续转换 / 导出 / 脱敏 |

### 5. 渐进式披露

`CLAUDE.md` 本身就是渐进式披露的体现：从 Level 1 基础操作到 Level 5 转换导出，按需展开。文档站沿用同样的层级化导航结构，让 AI Agent 与人类读者都能从最简用法逐步深入到实现原理。

## Go 版本与依赖

- **Go 版本**：1.19+（见 `go.mod`）。
- **CLI 依赖**：`spf13/cobra` v1.8.0（命令框架）、`spf13/viper` v1.18.2（配置 / 环境变量）。
- **SDK 运行时依赖**：零。仅标准库。
- **测试依赖**：`stretchr/testify` v1.8.4，仅用于测试，不进入运行时。

## 模块文件 → 职责速查表

根包 41 个模块文件按职能归类，便于快速定位实现：

| 文件 | 一句话职责 |
| --- | --- |
| `har.go` | 核心类型定义：`Har`、`Entries`、`Request`、`Response`、`Creator`、`Browser`、`Page` 等结构体 |
| `interfaces.go` | 10 个 Provider 接口：`HARProvider` / `EntryProvider` / `RequestProvider` / `ResponseProvider` / `TimingsProvider` 等 |
| `parser.go` | `*Har` 解析入口族：`ParseHar*` / `ParseHarFile*` / `ParseHarEnhanced` / `ParseHarLenient` / `ParseHarWithWarnings` |
| `reader.go` | Reader / 自动检测入口：`ParseHarFromReader` / `ParseHarFileAuto` / `ParseHarFileGzipped` |
| `parse.go` | `HARProvider` 入口族：`Parse` / `ParseFile` + 函数式选项分发 |
| `options.go` | 解析选项：`Option`、`options`、`WithLenient` / `WithMemoryOptimized` / `WithLazyLoading` / `OptFast` 等 |
| `standard_impl.go` | 标准实现的 `HARProvider` |
| `memory.go` | 内存优化实现的 `HARProvider` |
| `lazy.go` + `lazy_impl.go` | 懒加载实现的 `HARProvider` 与按需加载机制 |
| `streaming.go` | 流式解析：`StreamingHar`、`StreamingEntryIterator`、`NewStreamingParserFromFile` |
| `optimized_impl.go` | 优化实现共用逻辑 |
| `filter.go` | 过滤引擎：`Filter`、`FilterWith`、`FindBy*`、`FindErrors`、`FindSlowRequests`、`FilterResult` 链式方法 |
| `functional_options.go` | 函数式选项：`FilterOption`、`WithFilter*` 系列 |
| `converter.go` | 格式转换：`Convert`、`ConvertFormat`（CSV/Markdown/JSON/YAML/XML…）、`ConvertOptions` |
| `export.go` | 导出：`ToCurl` / `ToWget` / `ToPythonRequests` / `ToPostmanCollection` / `ToXML` / `ToYAML` / `ToJSON` 等 |
| `transform.go` | 请求变换：`RewriteURL`、`RemoveHeaders`、`AddHeaders`、`Transform`（`TransformRule`） |
| `redact.go` | 脱敏：`Redact`、`DefaultRedactOptions`（默认目标与替换文本） |
| `builder.go` | 构建器：`HarBuilder` / `EntryBuilder`，链式构造 HAR |
| `generator.go` | 高层生成 API：`NewHar`、`SetCreator`、`AddPage`、`AddEntry`、`SaveToFile` |
| `merge.go` | 合并：`Merge` |
| `diff.go` | 差异：`Diff`、`DiffResult`、`DefaultDiffOptions` |
| `dedup.go` | 去重：`Dedup`，三策略（exact / pattern / content-hash） |
| `split.go`（在相关模块中） | 拆分：按 page / domain / time / size / status / method |
| `statistics.go` | 统计：`Statistics`、`Summary`、`TimingStatistics`、`DomainSummary`、`StatusCodeDistribution` 等 |
| `security.go` | 安全审计：`SecurityAudit`、`SecurityReport`、`FindBySeverity` |
| `cookie.go` | Cookie 分析：`CookieAudit`、`CookieEvolution` |
| `cache.go` | 缓存分析：`CacheAnalysis`、`CacheReport` |
| `performance.go` | 性能评分：`PerformanceScore`、`PerformanceReport`、`Grade()` |
| `timeline.go` | 时间线与并发分析 |
| `content.go` | 内容类型分析 |
| `index.go` | 索引：内存索引与多种查询 |
| `extensions.go` | 扩展字段保真（custom fields） |
| `format.go` | 输出格式化辅助 |
| `decode.go` | 响应体解码（base64 / chunked 等） |
| `http_body.go` | HTTP body 处理 |
| `replay.go` | 请求重放：`Replay`、重放选项与结果 |
| `errors.go` | 增强错误体系：`HarError`、`ErrorCode`、`*FileSystemError` 等分类错误 |
| `validator.go` + `validator_ext.go` | HAR 规范校验：`ValidateHarFile`、`ValidateStrict`、`ValidateTimingsConsistency` |
| `util.go` | 通用工具函数 |
| `doc.go` | 包级 godoc 文档 |

> CLI 侧的 24 个命令文件位于 `cmd/har/cmd/`，每个文件对应一个子命令，命名即职责（`info.go` → `info` 命令，依此类推）。

## 典型调用链

以 `har -f capture.har find --slow 1000 --format json` 为例，从命令行到 SDK 的完整调用链：

```
用户命令  har -f capture.har find --slow 1000 --format json
   │
   ├─ cmd/har/cmd/find.go      解析 flags（--slow=1000），构造过滤条件
   │
   ├─ cmd/har/internal/loader.go  LoadHar(cmd, args)
   │     ├─ 读取 -f 路径（或 - stdin / HAR_FILE 环境变量）
   │     ├─ har.ParseHarFile(path) → *Har
   │     └─ 返回 *Har
   │
   ├─ SDK 根包 filter.go
   │     └─ h.FindSlowRequests(1000) → *FilterResult
   │
   ├─ cmd/har/internal/output.go  GetFormat / GetOutputPath
   │     └─ WriteOutput(cmd, result, textFn, csvFn)
   │           ├─ format=json → json.MarshalIndent
   │           ├─ format=csv  → csvFunc()
   │           └─ format=text → textFunc()
   │
   └─ stdout（或 -o 指定文件）
```

这条链路体现了分层的核心收益：命令层只关心「把 flag 翻译成 SDK 调用」，胶水层只关心「载入与格式化」，SDK 只关心「业务逻辑」。任一层都可独立替换——例如把 CLI 换成 MCP 封装或 AI Agent Skill，SDK 与胶水层无需改动。

## 扩展点

贡献者最常触碰的扩展入口：

- **新增 CLI 命令**：在 `cmd/har/cmd/` 加一个 `.go` 文件，实现 `cobra.Command`，用 `internal.LoadHar` 取 `*Har`，用 `internal.WriteOutput` 写结果，最后在 `root.go` 注册。
- **新增过滤条件**：在 `filter.go` 的 `FilterOptions` 加字段，在过滤引擎中处理，并提供对应的 `WithFilter*` 函数式选项（`functional_options.go`）与便捷方法 `FindBy*`。
- **新增导出格式**：在 `export.go` 加 `ToXxx` 方法，并在 `converter.go` 的 `ConvertFormat` 常量与 switch 分支中登记。
- **新增解析策略**：实现 `HARProvider` 接口（及必要的子接口），在 `ParseFile` 的分发逻辑中加入选择分支，并补 `Option` 与 `With*` 函数式选项。
- **新增校验规则**：在 `validator.go` / `validator_ext.go` 中加入检查函数，并在 `ValidateHarFile` / `ValidateStrict` 中调用。

所有扩展点都遵循同一原则：**接口先行、克隆语义、双 API 风格、文档同步**。

## 下一步

- 想深入 4 种解析策略的实现取舍，前往 [四种解析策略](/zh/sdk/parsing-strategies) 与 [实现原理](/zh/internals/memory-optimized)。
- 想了解 Provider 接口设计，前往 [Provider 接口](/zh/sdk/providers)。
- 准备提交代码，请先阅读 [贡献指南](/zh/contributing/)。
