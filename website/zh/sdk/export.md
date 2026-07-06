---
title: 导出能力
titleTemplate: false
---

# 导出能力

SDK 的导出能力分布在四个模块：`converter.go`（表格类格式 CSV/Markdown/HTML/Text）、`export.go`（重放命令与结构化格式 cURL/Wget/Python/Postman/XML）、`format.go`（YAML）、`util.go`（JSON 与文件写入）。所有方法都挂在 `*Har` 上，部分也提供单条目 `*Entries` 版本。

## Convert：表格类格式

`Convert(format, opts)` 将 HAR 转为表格文本，适合做报告或电子表格导入。`ConvertFormat` 常量定义在 `converter.go`，`FormatYAML` 定义在 `format.go` 但与前三者同属 `ConvertFormat` 类型。

```go
const (
    FormatCSV      ConvertFormat = "csv"
    FormatMarkdown ConvertFormat = "markdown"
    FormatHTML     ConvertFormat = "html"
    FormatText     ConvertFormat = "text"
)
const FormatYAML ConvertFormat = "yaml"
```

### ConvertOptions 字段裁剪

通过布尔开关决定导出哪些列，未启用的字段不会出现在结果里：

```go
type ConvertOptions struct {
    IncludeURL         bool
    IncludeMethod      bool
    IncludeStatus      bool
    IncludeContentType bool
    IncludeSize        bool
    IncludeTime        bool
    IncludeTimings     bool   // 阻塞/DNS/连接/发送/等待/接收 六列
    IncludeHeaders     bool   // 请求头 + 响应头
    IncludeDateTime    bool
    IncludePostData    bool   // POST 类型 + 文本
    IncludeQueryString bool
    Headers            []string       // 自定义表头（覆盖默认）
    Filter             *FilterOptions // 转换前先过滤
}
```

`DefaultConvertOptions()` 默认开启 URL/Method/Status/ContentType/Size/Time/DateTime，其余关闭。

### Convert 与 ConvertWith

`Convert(format, opts)` 接收结构体选项；`ConvertWith(format, opts...)` 接收函数式选项，更适合链式调用：

```go
package main

import (
    "fmt"

    har "github.com/cyberspacesec/har-skills"
)

func main() {
    h, err := har.ParseHarFile("testdata/full.har")
    if err != nil {
        panic(err)
    }

    // 结构体选项：导出含 timings 的 CSV
    opts := har.DefaultConvertOptions()
    opts.IncludeTimings = true
    opts.IncludeHeaders = false
    csv, err := h.Convert(har.FormatCSV, opts)
    if err != nil {
        panic(err)
    }
    fmt.Println(csv)

    // 函数式选项：导出 Markdown
    md, err := h.ConvertWith(har.FormatMarkdown,
        har.WithConvertIncludeTimings(true),
        har.WithConvertIncludeURL(true),
        har.WithConvertIncludeMethod(true),
        har.WithConvertIncludeStatus(true),
    )
    if err != nil {
        panic(err)
    }
    fmt.Println(md)
}
```

可用的 `WithConvert*` 选项：`IncludeHeaders/Timings/Bodies/Cookies/QueryStrings/Status/Size/URL/Method/Time/MimeType`、`Headers`、`Filter`。

## 重放命令导出

这三个方法返回 `string`，把 HAR 条目转成可在终端直接执行的命令脚本：

```go
curl    := h.ToCurl()              // curl -H '...' --data '...' 'URL'
wget    := h.ToWget()              // wget --header='...' --post-data='...' -qO- 'URL'
python  := h.ToPythonRequests()    // import requests + requests.get/post(...)
```

生成时的小细节：

- `ToCurl` 跳过 `Host` 头（curl 自动添加）；非 GET 方法加 `-X`；检测到 `Accept-Encoding: gzip/deflate` 时追加 `--compressed`；HTTPS 且响应有错误时追加 `-k`。
- `ToWget` 跳过 `Host`；非 GET 用 `--method=`；HTTPS 加 `--no-check-certificate`；默认 `-qO-` 输出到 stdout。
- `ToPythonRequests` 输出 `import requests` 头部，逐条目生成 `headers = {...}` 与 `response = requests.<method>(...)`。

每个方法在 `*Entries` 上也有同名版本，只导出单条目：

```go
first := &h.Log.Entries[0]
fmt.Println(first.ToCurl())
fmt.Println(first.ToPythonRequests())
```

## 结构化格式导出

### Postman Collection v2.1

`ToPostmanCollection()` 返回 `([]byte, error)`，生成符合 Postman v2.1 schema 的 JSON，可直接导入 Postman。`SaveAsPostmanCollection(path)` 是其写文件便捷方法。

```go
data, err := h.ToPostmanCollection()
if err != nil {
    panic(err)
}
// 直接写文件
if err := h.SaveAsPostmanCollection("collection.json"); err != nil {
    panic(err)
}
```

### XML

`ToXML()` 返回 `(string, error)`，输出带 `<?xml ...?>` 头的标准 XML。内部用 `encoding/xml` 的结构体映射，覆盖 version/creator/entries/request/response/headers/content/postData。

```go
xmlStr, err := h.ToXML()
if err != nil {
    panic(err)
}
if err := h.SaveAsXML("capture.xml"); err != nil {
    panic(err)
}
```

### YAML

`ToYAML()` 返回 `(string, error)`。实现不依赖外部 YAML 库——先 `ToJSON(true)` 再走内置 JSON→YAML 转换器，对字符串特殊字符做了转义。

```go
yamlStr, err := h.ToYAML()
if err != nil {
    panic(err)
}
if err := h.SaveAsYAML("capture.yaml"); err != nil {
    panic(err)
}
```

### JSON

`ToJSON(indent bool)` 返回 `([]byte, error)`，是其他结构化格式的基础。`indent=true` 时输出带缩进的 JSON。

```go
data, err := h.ToJSON(true) // 缩进
if err != nil {
    panic(err)
}
```

## 写文件

`util.go` 与各导出模块提供了写文件的便捷方法：

| 方法 | 签名 | 说明 |
| --- | --- | --- |
| `SaveToFile` | `(filePath string, indent bool) error` | 写 JSON，控制缩进 |
| `SaveToFileGzipped` | `(filePath string, indent bool) error` | 写 gzip 压缩 JSON |
| `SaveToWriter` | `(w io.Writer, indent bool) error` | 写 JSON 到任意 Writer |
| `SaveAsPostmanCollection` | `(filePath string) error` | 写 Postman v2.1 JSON |
| `SaveAsXML` | `(filePath string) error` | 写 XML |
| `SaveAsYAML` | `(filePath string) error` | 写 YAML |

## 流式导出 ConvertTo

`ConvertTo(format, w, opts)` 把转换结果直接写入 `io.Writer`，避免在内存中持有完整字符串，适合大文件导出到文件或 HTTP 响应体：

```go
package main

import (
    "os"

    har "github.com/cyberspacesec/har-skills"
)

func main() {
    h, err := har.ParseHarFile("testdata/large.har")
    if err != nil {
        panic(err)
    }

    f, err := os.Create("report.csv")
    if err != nil {
        panic(err)
    }
    defer f.Close()

    opts := har.DefaultConvertOptions()
    opts.IncludeTimings = true

    // 流式写入，不生成中间字符串
    if err := h.ConvertTo(har.FormatCSV, f, opts); err != nil {
        panic(err)
    }
}
```

`ConvertTo` 支持 `FormatYAML`、`FormatCSV`、`FormatMarkdown`、`FormatHTML`、`FormatText`，其它值会退化为带缩进的 JSON 输出。它会对 `nil` writer 做检查并返回 `*HarError`。

## 综合示例：从 HAR 生成 cURL 重放脚本

```go
package main

import (
    "os"

    har "github.com/cyberspacesec/har-skills"
)

func main() {
    h, err := har.ParseHarFile("testdata/full.har")
    if err != nil {
        panic(err)
    }

    // 只重放 API 请求
    api := h.FindByDomain("api.example.com")
    replays := api.ToHar().ToCurl()

    if err := os.WriteFile("replay.sh", []byte("#!/bin/bash\n\n"+replays+"\n"), 0644); err != nil {
        panic(err)
    }
    _ = os.Chmod("replay.sh", 0755)
}
```
