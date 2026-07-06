---
title: 文件操作
titleTemplate: false
---

# 文件操作

Level 2 的 4 个命令面向「多份 HAR 之间」的工作：对比、合并、拆分、验证。`diff` 与 `merge` 直接吃位置参数（不走 `-f`），`split` 与 `validate` 仍走 `--file`。它们既可能改文件（合并/拆分会写出新 HAR），也可能只产出报告（对比/验证）。

所有示例都可在仓库根目录直接运行，使用 `testdata/example.har` 或 `testdata/full.har`。

## diff — 对比两个 HAR

找出两个 HAR 之间新增、删除、修改的请求。位置参数 `<file1> <file2>`，**精确两个**（`cobra.ExactArgs(2)`），不走 `-f`。

```bash
har diff testdata/full.har testdata/large.har
```

按 URL 匹配（默认按「索引 + URL」配对），并比较响应体：

```bash
har diff testdata/full.har testdata/large.har --compare-by-url --include-body
```

忽略 Cookie 与 Date 头的差异：

```bash
har diff a.har b.har --ignore-headers Cookie,Date
```

### Flags

| Flag | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--ignore-headers` | stringSlice | `[]` | 忽略的头部字段名（逗号分隔） |
| `--ignore-timings` | bool | `true` | 忽略时间差异 |
| `--ignore-dates` | bool | `true` | 忽略日期差异 |
| `--include-body` | bool | `false` | 比较响应体内容 |
| `--compare-by-url` | bool | `false` | 按 URL 匹配（默认按索引+URL） |

::: tip 默认忽略什么
`--ignore-timings` 与 `--ignore-dates` 默认都为 `true`——两条抓包的时间戳与各阶段耗时几乎不可能一致，默认忽略它们才能聚焦「请求本身有没有变」。需要看时间差异时显式传 `--ignore-timings=false`。
:::

### 实现原理

两个文件经 `internal.LoadHarFromArg` 分别加载（gzip 自动检测）；选项走 `har.DefaultDiffOptions()` 再用各 flag 覆盖；调用 `har.Diff(har1, har2, options)` 得到 `*DiffResult`。输出经 `internal.WriteOutput`：文本走 `diffResult.Report(har.FormatText)`，CSV 走 `diffResult.Report(har.FormatCSV)`，JSON 序列化整个 `*DiffResult`。

## merge — 合并多 HAR

把多个 HAR 的条目合并到一个 HAR，沿用第一个文件的版本与创建者信息。位置参数 `<file1> [file2...]`，**至少一个**（`cobra.MinimumNArgs(1)`），不走 `-f`。

```bash
har merge part1.har part2.har part3.har
```

合并并按 Method+URL 去重（保留最新的）：

```bash
har merge a.har b.har --deduplicate -o merged.har
```

不按时间排序（保留各文件原始拼接顺序）：

```bash
har merge a.har b.har --sort-by-time=false -o raw.har
```

### Flags

| Flag | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--sort-by-time` | bool | `true` | 合并后按 startedDateTime 排序 |
| `--deduplicate` | bool | `false` | 去重（按 Method+URL，保留最新） |

::: tip 输出走 JSON
`merge` 直接把合并后的 `*Har` 序列化为 JSON 写出（不经 `internal.WriteOutput` 的多格式分支），因此 `--format` 对它无效；用 `-o` 指定文件，或留空走 stdout。合并信息（条目数等）打到 stderr。
:::

### 实现原理

每个位置参数经 `internal.LoadHarFromArg` 加载成 `[]*har.Har`；选项包成 `har.MergeOptions{SortByTime, Deduplicate}`；调用 `har.MergeWithOptions(options, hars...)` 得到合并后的 `*Har`，再 `json.MarshalIndent` 写出。`--deduplicate` 以 Method+URL 为键去重，保留时间最新的那条。

## split — 拆分 HAR

按页面引用、域名、时间间隔、条目数、状态码范围或 HTTP 方法把一个大 HAR 拆成多个小文件。`--by` 是**必需** flag。

按域名拆分，文件名前缀 `by-domain`：

```bash
har -f testdata/full.har split --by domain -o by-domain
```

按时间间隔拆分，每 30 分钟一个文件：

```bash
har -f testdata/full.har split --by time --interval 30m
```

按条目数拆分，每 50 条一个文件：

```bash
har -f testdata/full.har split --by size --max-entries 50
```

按状态码范围拆分：

```bash
har -f testdata/full.har split --by status
```

### Flags

| Flag | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--by` | string | `""` | 拆分方式（`page`/`domain`/`time`/`size`/`status`/`method`），必需 |
| `--interval` | duration | `1h` | 时间间隔（配合 `--by time`） |
| `--max-entries` | int | `100` | 每组最大条目数（配合 `--by size`） |
| `--output`/`-o` | string | `split` | 输出文件名前缀（本地 flag） |

::: tip 文件命名规则
`split` 的 `-o` 是命令本地 flag（默认 `split`），与全局 `--output` 不同：它作**文件名前缀**，不指单个输出文件。生成的文件按 `<prefix>_<kind>_<key>.har` 命名——例如 `by-domain_domain_api.example.com.har`；`time`/`size` 这类有序拆分用三位序号：`split_time_001.har`。`key` 里的特殊字符（`/ \ : * ? " < > |` 与空格）会被替换为 `_`，避免非法文件名。每个文件写入后会在 stderr 打印路径与条目数。
:::

### 实现原理

`--by` 为空直接报错退出；否则 `internal.LoadHar` 加载，按 `--by` 分派到 `splitByPage` / `splitByDomain` / `splitByTime` / `splitBySize` / `splitByStatus` / `splitByMethod`，分别调 `(*Har).SplitByPage()` / `SplitByDomain()` / `SplitByTimeRange(interval)` / `SplitBySize(maxEntries)` / `SplitByStatusCode()` / `SplitByMethod()`。map 形结果（page/domain/status/method）走 `writeSplitMap`（按 key 命名），slice 形结果（time/size）走 `writeSplitSlice`（按序号命名），统一 `json.MarshalIndent` 写盘。

## validate — 验证 HAR 规范

检查 HAR 文件是否符合规范：标准模式查基本结构与必填字段；严格模式额外查交叉引用、HTTP 方法、状态码范围；时间一致性校验查 `Time` 字段与 `Timings` 各阶段之和是否吻合。

```bash
har -f testdata/full.har validate
```

严格模式：

```bash
har -f testdata/full.har validate --strict
```

自定义时间容差 5 毫秒：

```bash
har -f testdata/full.har validate --timings-tolerance 5
```

最严格（容差 0，要求 timings 之和与 Time 完全一致）：

```bash
har -f testdata/full.har validate --strict --timings-tolerance 0
```

### Flags

| Flag | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--strict` | bool | `false` | 启用严格验证（交叉引用 / 方法 / 状态码范围） |
| `--timings-tolerance` | float64 | `10` | 时间一致性容差（毫秒），`0` 表示严格一致 |

::: tip 容差为 0
`--timings-tolerance` 是 float64，默认 10 毫秒。设为 `0` 表示要求 `Timings` 各阶段之和与 `entry.Time` **严格相等**——抓包工具的浮点误差通常会让该检查报错，故默认留了 10 毫秒缓冲。需要绝对严格时再传 `0`。
:::

### 实现原理

`internal.LoadHar` 加载后，`--strict` 为 true 走 `har.ValidateStrict(h)`，否则走 `har.ValidateHarFile(h)`；二者返回的 `*HarError` 经 `collectErrors` 拆成 `[]*ValidationError`。时间一致性始终跑 `har.ValidateTimingsConsistency(h, tolerance)`（容差 ≥ 0 时启用），结果与上面的错误合并。文本输出由 `formatValidateText` 渲染：无错打印 `✓ Valid`，有错则逐条列出 `[Rule] Field: Message`。

## 小结

| 命令 | 输入方式 | 是否产出新 HAR |
|------|----------|----------------|
| `diff` | 位置参数 ×2 | 否（报告） |
| `merge` | 位置参数 ≥1 | 是（合并 HAR） |
| `split` | `--file` | 是（多个 HAR） |
| `validate` | `--file` | 否（报告） |

`diff`/`merge` 走位置参数、不走 `-f`；`split`/`validate` 走 `--file`，也支持 stdin。需要安全审计或脱敏时进 [安全与隐私](./security.md)。
