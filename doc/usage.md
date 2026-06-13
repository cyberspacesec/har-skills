# Go-HAR 使用文档

Go-HAR 是一个强大而灵活的 Go 语言 HTTP Archive (HAR) 文件解析和处理库。本文档将帮助您了解如何使用 Go-HAR 及其提供的所有功能。

## 功能概述

Go-HAR 提供以下主要功能：

- **标准解析** - 解析 HAR 文件并将其加载到内存中
- **内存优化解析** - 使用优化的数据结构减少内存占用
- **懒加载** - 延迟加载大型内容，减少初始内存占用
- **流式解析** - 一次处理一个条目，适用于处理超大 HAR 文件
- **增强的错误处理** - 提供详细的错误和警告信息
- **转换功能** - 在不同解析模式之间转换
- **过滤和搜索** - 高效查找和过滤 HAR 数据

## 安装

使用 `go get` 命令添加 Go-HAR 到您的项目：

```bash
go get github.com/cyberspacesec/har-skills
```

## 基本用法

### 解析 HAR 文件

最基本的用法是将 HAR 文件解析为内存中的结构：

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/cyberspacesec/har-skills"
)

func main() {
    // 解析 HAR 文件
    harData, err := har.ParseHarFile("example.har")
    if err != nil {
        log.Fatalf("无法解析 HAR 文件: %v", err)
    }
    
    // 访问 HAR 数据
    fmt.Printf("HAR 版本: %s\n", harData.Log.Version)
    fmt.Printf("条目数量: %d\n", len(harData.Log.Entries))
    
    // 遍历所有请求
    for i, entry := range harData.Log.Entries {
        fmt.Printf("请求 #%d: %s %s\n", i+1, entry.Request.Method, entry.Request.URL)
    }
}
```

### 使用函数选项 API（推荐）

Go-HAR 提供了一个灵活的函数选项 API，允许您自定义解析行为：

```go
package main

import (
    "log"
    
    "github.com/cyberspacesec/har-skills"
)

func main() {
    // 使用选项解析 HAR 文件
    harData, err := har.ParseHarFile("large.har", 
        har.WithMemoryOptimized(),  // 使用内存优化模式
        har.WithValidation(false),  // 跳过验证以提高性能
    )
    if err != nil {
        log.Fatalf("无法解析 HAR 文件: %v", err)
    }
    
    // 使用接口处理数据
    ProcessEntries(harData)
}

// 使用接口允许任何 HAR 实现
func ProcessEntries(harData har.HARProvider) {
    for _, entry := range harData.GetEntries() {
        // 处理每个条目
        _ = entry.GetRequest().GetURL()
    }
}
```

## 高级用法

### 内存优化

对于大型 HAR 文件，内存优化模式可以显著减少内存使用：

```go
// 使用内存优化模式
harData, err := har.ParseHarFile("large.har", har.WithMemoryOptimized())
if err != nil {
    log.Fatalf("无法解析 HAR 文件: %v", err)
}

// 接口保持一致，使用方式相同
for _, entry := range harData.GetEntries() {
    fmt.Printf("URL: %s\n", entry.GetRequest().GetURL())
}
```

内存优化模式使用以下技术减少内存占用：
- 使用 map 代替数组存储头部和查询参数
- 使用指针表示可选字段
- 使用枚举代替字符串存储 HTTP 方法

### 懒加载

对于包含大型响应内容的 HAR 文件，懒加载模式可以延迟加载内容：

```go
// 使用懒加载模式
harData, err := har.ParseHarFile("large_content.har", har.WithLazyLoading())
if err != nil {
    log.Fatalf("无法解析 HAR 文件: %v", err)
}

// 基本信息直接可用
for _, entry := range harData.GetEntries() {
    resp := entry.GetResponse()
    fmt.Printf("状态码: %d, 内容大小: %d\n", 
        resp.GetStatus(), 
        resp.GetContent().GetSize())
    
    // 内容仅在需要时加载
    if resp.GetStatus() == 200 {
        content := resp.GetContent()
        text := content.GetText() // 此时才加载内容
        fmt.Printf("内容长度: %d\n", len(text))
    }
}
```

### 流式解析

对于超大 HAR 文件，流式解析可以一次处理一个条目：

```go
// 创建流式解析器
iterator, err := har.NewStreamingParserFromFile("huge.har")
if err != nil {
    log.Fatalf("无法创建流式解析器: %v", err)
}
defer iterator.Close()

// 逐个条目处理
for iterator.Next() {
    entry := iterator.Entry()
    fmt.Printf("处理请求: %s\n", entry.GetRequest().GetURL())
    
    // 处理完毕后，条目将被释放
}

// 检查是否有错误发生
if err := iterator.Error(); err != nil {
    log.Fatalf("流式解析出错: %v", err)
}
```

### 增强的错误处理

对于格式可能不完全符合标准的 HAR 文件，使用增强的错误处理：

```go
// 使用增强的错误处理
result, err := har.ParseHarFileWithWarnings("problematic.har")
if err != nil {
    log.Fatalf("解析完全失败: %v", err)
} else {
    // 解析成功，但可能有警告
    harData := result.HAR
    
    if len(result.Warnings) > 0 {
        fmt.Printf("解析成功，但有 %d 个警告:\n", len(result.Warnings))
        for i, w := range result.Warnings {
            fmt.Printf("警告 #%d: %s\n", i+1, w.Error())
        }
    }
    
    // 继续使用 harData
    fmt.Printf("成功解析 %d 个条目\n", len(harData.GetEntries()))
}
```

### 过滤和搜索

Go-HAR 提供高效的过滤和搜索功能：

```go
// 使用过滤器找到所有 POST 请求
postRequests := har.Filter(harData, func(entry har.EntryProvider) bool {
    return entry.GetRequest().GetMethod() == har.MethodPOST
})

// 找到所有状态码为 404 的响应
notFoundResponses := har.Filter(harData, func(entry har.EntryProvider) bool {
    return entry.GetResponse().GetStatus() == 404
})

// 搜索特定的 URL 模式
apiCalls := har.Filter(harData, func(entry har.EntryProvider) bool {
    return strings.Contains(entry.GetRequest().GetURL(), "/api/v1/")
})

// 组合过滤条件
slowApiCalls := har.Filter(harData, func(entry har.EntryProvider) bool {
    return strings.Contains(entry.GetRequest().GetURL(), "/api/") && 
           entry.GetTime() > 1000 // 超过1秒的请求
})
```

### 转换功能

在不同的解析模式之间转换：

```go
// 从标准模式转换为内存优化模式
standardHar, _ := har.ParseHarFile("example.har")
optimizedHar := har.ToOptimized(standardHar)

// 从内存优化模式转换为标准模式
optimizedHar, _ := har.ParseHarFile("example.har", har.WithMemoryOptimized())
standardHar := optimizedHar.ToStandard()

// 从任何模式转换为标准模式（通过接口）
func ConvertToStandard(provider har.HARProvider) *har.Har {
    return provider.ToStandard()
}
```

## 实用工具

Go-HAR 还提供了几个实用工具，帮助您处理 HAR 文件：

### 统计分析工具

```go
// 分析 HAR 文件并生成统计信息
stats := har.AnalyzeStatistics(harData)

fmt.Printf("总请求数: %d\n", stats.TotalRequests)
fmt.Printf("平均响应时间: %.2fms\n", stats.AverageResponseTime)
fmt.Printf("最慢的请求: %s (%.2fms)\n", stats.SlowestRequest.URL, stats.SlowestRequest.Time)
fmt.Printf("最大响应: %s (%d 字节)\n", stats.LargestResponse.URL, stats.LargestResponse.Size)
```

### 可视化工具

```go
// 创建瀑布图
waterfall := har.CreateWaterfall(harData)
err := waterfall.SaveAsHTML("waterfall.html")
if err != nil {
    log.Fatalf("无法保存瀑布图: %v", err)
}

// 创建性能图表
perfChart := har.CreatePerformanceChart(harData)
err = perfChart.SaveAsHTML("performance.html")
if err != nil {
    log.Fatalf("无法保存性能图表: %v", err)
}
```

### 命令行工具

Go-HAR 还提供了命令行工具：

```bash
# 显示 HAR 文件基本信息
go-har info example.har

# 列出所有请求
go-har list example.har

# 查找特定请求
go-har find example.har --url "/api"

# 显示请求头部
go-har headers example.har --url "/login"

# 分析时间
go-har timing example.har --sort-by time

# 提取内容
go-har extract example.har --url "/api/data" --output data.json
```

## 参考

### 主要接口

Go-HAR 设计基于接口，主要接口包括：

- `HARProvider` - HAR 文件的主接口
- `EntryProvider` - 单个 HTTP 请求/响应条目
- `RequestProvider` - HTTP 请求
- `ResponseProvider` - HTTP 响应
- `ContentProvider` - 响应内容
- `HeadersProvider` - HTTP 头部
- `TimingsProvider` - 请求时间信息

### 函数选项

可用的函数选项包括：

- `WithMemoryOptimized()` - 使用内存优化模式
- `WithLazyLoading()` - 使用懒加载模式
- `WithValidation(bool)` - 启用或禁用验证
- `WithWarnings()` - 收集警告而不是返回错误
- `WithCacheEnabled(bool)` - 控制内容缓存
- `WithMaxContentSize(int)` - 限制内容大小
- `WithIgnoreFields([]string)` - 忽略特定字段

## 结论

Go-HAR 提供了灵活而强大的 API，用于处理各种规模的 HAR 文件。通过选择合适的解析模式和利用提供的接口，您可以高效地处理和分析 HTTP 存档数据，无论其大小或复杂性如何。

如需更多详细信息，请参考代码文档和提供的示例。 