---
title: 示例代码集
---

# 示例代码集

`examples/` 目录下收录了 7 个独立可运行的 Go 示例，每个示例都是一个 `package main`，覆盖从最基础的 HAR 解析到性能基准、错误处理、可视化报告生成的完整场景。这些示例既是上手 SDK 的最佳起点，也是真实业务场景的可复用模板。

所有示例都可以直接用仓库 `testdata/` 目录下的真实 HAR 文件运行，无需额外准备数据。

## 示例总览

### 1. 基础解析 `examples/main.go`

最简起步示例，演示两种最常用的解析入口：从字节切片解析、从文件路径解析。

| 项 | 内容 |
| --- | --- |
| 演示能力 | HAR 文件解析、基础输出 |
| 关键 API | `har.ParseHar`、`har.ParseHarFile` |
| 源码路径 | `examples/main.go` |
| 运行方式 | `cd examples && go run main.go` |

### 2. 进阶用法 `examples/advanced/main.go`

一次性演示过滤、转换、创建三个核心能力，是理解 SDK「读 → 过滤 → 转换 → 写」全链路最完整的单文件示例。

| 项 | 内容 |
| --- | --- |
| 演示能力 | 过滤（方法/内容类型/慢请求/错误）、格式转换（CSV/Markdown）、从零构建 HAR |
| 关键 API | `FindByMethod`、`Filter(FilterOptions{})`、`FindSlowRequests`、`FindErrors`、`Convert`、`NewHar`、`AddEntry`、`SaveToFile` |
| 源码路径 | `examples/advanced/main.go` |
| 运行方式 | `cd examples/advanced && go run main.go` |

### 3. 迷你 CLI `examples/cli-tool/main.go`

不依赖任何 CLI 框架，仅用标准库 `flag` 从零构建一个 6 命令的 HAR 分析工具（info / list / find / headers / timing / extract），是学习如何把 SDK 嵌入自研工具的范本。

| 项 | 内容 |
| --- | --- |
| 演示能力 | 命令行参数解析、`HARProvider` 通用处理、`ToStandard()` 遍历、JSON/CSV/Text 多格式输出 |
| 关键 API | `har.ParseFile`、`har.WithMemoryOptimized`、`HARProvider.GetEntries`、`EntryProvider.ToStandard` |
| 源码路径 | `examples/cli-tool/main.go` |
| 运行方式 | `cd examples/cli-tool && go run main.go -file ../../testdata/example.har -cmd info` |

### 4. 增强错误处理 `examples/error_handling/main.go`

演示 SDK 的增强错误体系：分类错误码、文件系统错误判定、宽松模式部分解析、警告收集与嵌套部分错误。

| 项 | 内容 |
| --- | --- |
| 演示能力 | 详细错误信息、文件系统错误识别、宽松解析返回部分结果、警告收集 |
| 关键 API | `ParseHarFileEnhanced`、`*HarError`（`GetCode`/`IsFileSystemError`/`HasPartialErrors`/`GetPartialErrors`）、`ParseHarFileLenient`、`ParseHarFileWithWarnings` |
| 源码路径 | `examples/error_handling/main.go` |
| 运行方式 | `cd examples/error_handling && go run main.go` |

### 5. 性能与解析模式 `examples/performance/main.go`

对 4 种解析模式（标准 / 内存优化 / 懒加载 / 流式）做内存与耗时基准，并演示函数式选项与 `HARProvider` 的通用消费方式。

| 项 | 内容 |
| --- | --- |
| 演示能力 | 4 种解析模式基准对比、懒加载触发时机、流式遍历、函数式选项 |
| 关键 API | `har.ParseFile`、`har.WithMemoryOptimized`、`har.WithLazyLoading`、`har.OptFast`、`har.NewStreamingParserFromFile`、`HARProvider` |
| 源码路径 | `examples/performance/main.go` |
| 运行方式 | `cd examples/performance && go run main.go -file ../../testdata/large.har -mode all` |

### 6. 深度统计 `examples/statistics/main.go`

用自定义 `Stats` 结构对 HAR 做深度统计：P95、中位数、成功率、缓存命中率、按域名 / 内容类型聚合、慢请求与大响应 Top 榜。

| 项 | 内容 |
| --- | --- |
| 演示能力 | 自定义统计结构、百分位计算、按域名性能聚合、按内容类型分组 |
| 关键 API | `har.ParseFile`、`har.WithMemoryOptimized`、`HARProvider.GetEntries`、`EntryProvider.ToStandard` |
| 源码路径 | `examples/statistics/main.go` |
| 运行方式 | `cd examples/statistics && go run main.go -file ../../testdata/example.har` |

### 7. 可视化报告 `examples/visualization/main.go`

把 HAR 转成带 Chart.js 图表与彩色瀑布图的独立 HTML 报告，无需前端框架，开箱即用。

| 项 | 内容 |
| --- | --- |
| 演示能力 | 生成独立 HTML 报告、Chart.js 饼图、彩色时间线瀑布图 |
| 关键 API | `har.ParseFile`、`HARProvider.GetEntries`、`EntryProvider.ToStandard`、`Timings` 字段访问 |
| 源码路径 | `examples/visualization/main.go` |
| 运行方式 | `cd examples/visualization && go run main.go -file ../../testdata/example.har -output har_viz` |

## 常用示例核心代码

下面给出 3 个最常用示例的核心片段，方便快速复制上手。

### 基础解析

从字节或文件路径两种方式拿到 `*Har`，是最常见的入口。源码见 `examples/main.go`。

```go
package main

import (
	"fmt"
	"os"

	har "github.com/cyberspacesec/har-skills"
)

func main() {
	harFilePath := "./testdata/example.har"

	// 方式一：从字节切片解析
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

	// 方式二：直接给定文件路径解析
	harFile002, err := har.ParseHarFile(harFilePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(harFile002)
}
```

### 过滤与转换

`FilterResult` 支持链式调用，过滤结果可通过 `ToHar()` 转回 `*Har` 再做格式转换。源码见 `examples/advanced/main.go`。

```go
package main

import (
	"fmt"

	har "github.com/cyberspacesec/har-skills"
)

func main() {
	harFile, err := har.ParseHarFile("../../testdata/example.har")
	if err != nil {
		fmt.Println("解析HAR文件失败:", err)
		return
	}

	// 过滤所有 POST 请求
	postRequests := harFile.FindByMethod("POST")
	fmt.Printf("找到 %d 个POST请求\n", postRequests.Count())

	// 过滤所有图片请求
	imageRequests := harFile.Filter(har.FilterOptions{
		ContentType: "image/",
	})
	fmt.Printf("找到 %d 个图片请求\n", imageRequests.Count())

	// 过滤所有慢请求（超过 500ms）
	slowRequests := harFile.FindSlowRequests(500)
	fmt.Printf("找到 %d 个慢请求（>500ms）\n", slowRequests.Count())

	// 查找错误请求
	errorRequests := harFile.FindErrors()
	fmt.Printf("找到 %d 个错误请求\n", errorRequests.Count())

	// 将过滤结果转换为 CSV
	if slowRequests.Count() > 0 {
		options := har.DefaultConvertOptions()
		options.IncludeTimings = true

		csvData, err := slowRequests.ToHar().Convert(har.FormatCSV, options)
		if err != nil {
			fmt.Println("转换为CSV失败:", err)
		} else {
			fmt.Println("慢请求的CSV格式:")
			fmt.Println(csvData)
		}
	}
}
```

### 创建 HAR 文件

`NewHar()` 返回一个可链式配置的 `*Har`，通过 `AddEntry` 链式 builder 写入请求 / 响应 / 计时，最后 `SaveToFile` 落盘。源码见 `examples/advanced/main.go`。

```go
package main

import (
	"fmt"

	har "github.com/cyberspacesec/har-skills"
)

func main() {
	newHar := har.NewHar()
	newHar.SetCreator("go-har-example", "1.0")

	// 添加页面
	page := newHar.AddPage("page1", "示例页面")
	page.SetPageTimings(100, 300)

	// 添加请求/响应条目（链式 builder）
	entry := newHar.AddEntry("GET", "https://example.com/api/data", "HTTP/1.1", "page1")
	entry.AddRequestHeader("Accept", "application/json")
	entry.AddRequestHeader("User-Agent", "go-har/1.0")

	entry.SetResponseStatus(200, "OK")
	entry.AddResponseHeader("Content-Type", "application/json")
	entry.SetResponseContent(1024, "application/json")
	entry.SetTimings(10, 20, 30, 5, 50, 30, 25)

	// 保存为文件（indent=true 美化输出）
	newHarPath := "./generated.har"
	if err := newHar.SaveToFile(newHarPath, true); err != nil {
		fmt.Println("保存HAR文件失败:", err)
		return
	}
	fmt.Printf("成功创建并保存HAR文件到 %s\n", newHarPath)
}
```

## 用 testdata 运行

仓库 `testdata/` 目录提供了多份真实 HAR 文件，可直接喂给上述示例：

| 文件 | 大小 | 适用场景 |
| --- | --- | --- |
| `testdata/example.har` | 小 | 基础解析、过滤、统计、可视化 |
| `testdata/full.har` | 中 | 全字段覆盖、complex_test 场景 |
| `testdata/large.har` | ~1.5 MB | 性能基准、流式 / 懒加载对比 |
| `testdata/minimal_valid.har` | 极小 | 最小合规校验 |
| `testdata/invalid_date.har` | 小 | 宽松解析、错误处理示例 |
| `testdata/v1.1.har` | 小 | HAR 1.1 版本自动检测 |

运行示例时，把示例中相对路径指向 `testdata/` 即可。例如从仓库根目录运行性能基准：

```bash
cd examples/performance
go run main.go -file ../../testdata/large.har -mode all
```

运行可视化报告生成：

```bash
cd examples/visualization
go run main.go -file ../../testdata/example.har -output har_viz
# 生成的 har_viz/visualization.html 可直接用浏览器打开
```

## 下一步

- 想看完整的过滤 / 链式结果用法，前往 [过滤与链式结果](/zh/sdk/filtering)。
- 想了解 4 种解析模式的取舍，前往 [四种解析策略](/zh/sdk/parsing-strategies)。
- 想把这些示例组合成真实工作流，前往 [示例与工作流](/zh/workflows/security-audit)。
