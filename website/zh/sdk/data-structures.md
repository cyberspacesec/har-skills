---
title: 数据结构
titleTemplate: false
---

# 数据结构

`*Har` 是整个 SDK 的核心类型，在 `har.go` 中定义。它以一棵结构体树完整映射 HAR 1.2 规范，同时保留 Chrome DevTools 扩展字段（`_initiator`、`_priority`、`_resourceType`、`_transferSize` 等）。掌握这棵类型树，是使用所有 70+ 方法和编写自定义分析逻辑的前提。

## 类型树总览

```
Har
└── Log
    ├── Version        string            // HAR 规范版本，如 "1.2"
    ├── Creator        Creator           // 生成 HAR 的工具
    ├── Browser        Browser           // 浏览器信息（可选）
    ├── Pages          []Pages           // 页面信息（可选）
    │   ├── StartedDateTime time.Time
    │   ├── ID              string
    │   ├── Title           string
    │   └── PageTimings     PageTimings
    │       ├── OnContentLoad float64    // DOMContentLoaded（ms）
    │       └── OnLoad        float64    // load 事件（ms）
    ├── Entries        []Entries         // HTTP 请求/响应条目
    └── Comment        string            // 可选注释
```

`Har` 本身只持有一个 `Log` 字段和一个未导出的 `CustomFields`，几乎所有业务数据都在 `Log` 里。`Log.Entries` 是后续过滤、分析、导出、回放的主要操作对象。

## Creator 与 Browser

这两个结构体最简单，用来记录 HAR 是由谁、用什么工具或浏览器生成的。

```go
// Creator 表示创建 HAR 文件的工具信息
type Creator struct {
    Name    string `json:"name"`
    Version string `json:"version"`
    Comment string `json:"comment,omitempty"`
}

// Browser 表示浏览器信息
type Browser struct {
    Name    string `json:"name"`
    Version string `json:"version"`
    Comment string `json:"comment,omitempty"`
}
```

对应 JSON：

```json
"creator": { "name": "WebInspector", "version": "537.36" },
"browser": { "name": "Chrome", "version": "120.0.0.0" }
```

## Pages 与 PageTimings

`Pages` 描述一次页面加载的元信息和两个关键时间点。一个 HAR 文件可以包含多个页面（例如多标签页抓取）。

```go
type Pages struct {
    StartedDateTime time.Time   `json:"startedDateTime"`
    ID              string      `json:"id"`
    Title           string      `json:"title"`
    PageTimings     PageTimings `json:"pageTimings"`
    Comment         string      `json:"comment,omitempty"`
    CustomFields    CustomFields `json:"-"`
}

type PageTimings struct {
    OnContentLoad float64 `json:"onContentLoad"` // DOMContentLoaded 触发时间（ms）
    OnLoad        float64 `json:"onLoad"`        // load 事件触发时间（ms）
    Comment       string  `json:"comment,omitempty"`
}
```

`Entries.Pageref` 通过 `Pages.ID` 把单个请求关联到所属页面，这是 `waterfall --page-timings` 命令背后的数据基础。

## Entries —— 核心条目

`Entries` 是 SDK 中出现频率最高的类型。它对应一条完整的 HTTP 事务：请求、响应、计时、缓存以及 Chrome 扩展元数据。

```go
type Entries struct {
    StartedDateTime time.Time `json:"startedDateTime"`           // 请求开始时间
    Time            float64   `json:"time"`                      // 总耗时（ms）
    Request         Request   `json:"request"`
    Response        Response  `json:"response"`
    Cache           Cache     `json:"cache"`
    Timings         Timings   `json:"timings"`
    Pageref         string    `json:"pageref,omitempty"`         // 关联页面 ID
    ServerIPAddress string    `json:"serverIPAddress,omitempty"` // 服务器 IP
    Connection      string    `json:"connection,omitempty"`      // 连接 ID（连接复用分析）
    Initiator       Initiator `json:"_initiator,omitempty"`      // 请求发起者（Chrome 扩展）
    Priority        string    `json:"_priority,omitempty"`       // 请求优先级（Chrome 扩展）
    ResourceType    string    `json:"_resourceType,omitempty"`   // 资源类型（Chrome 扩展）
    Comment         string    `json:"comment,omitempty"`
    CustomFields    CustomFields `json:"-"`
}
```

`_initiator`、`_priority`、`_resourceType` 是 Chrome DevTools 扩展字段，浏览器导出的 HAR 才有。SDK 把它们当作普通字段处理，`find --resource-type`、`connections`、`FindByResourceType` 等能力都建立在这之上。

### Request

```go
type Request struct {
    Method       string        `json:"method"`
    URL          string        `json:"url"`
    HTTPVersion  string        `json:"httpVersion"`
    Cookies      []Cookie      `json:"cookies"`
    Headers      []Headers     `json:"headers"`
    QueryString  []QueryString `json:"queryString"`
    PostData     *PostData     `json:"postData,omitempty"` // 可选，仅 POST 等带 body 的请求
    HeadersSize  int           `json:"headersSize"`
    BodySize     int           `json:"bodySize"`
    Comment      string        `json:"comment,omitempty"`
    CustomFields CustomFields  `json:"-"`
}
```

`PostData` 是指针类型，表达"可选值"——这正是 `optimized` 策略用指针表达可选值的来源思路。`Headers` 在 standard 实现里是切片，在 optimized 实现里会被改写成 `map[string][]string` 以加速查找。

### Response

```go
type Response struct {
    Status       int          `json:"status"`
    StatusText   string       `json:"statusText"`
    HTTPVersion  string       `json:"httpVersion"`
    Cookies      []Cookie     `json:"cookies"`
    Headers      []Headers    `json:"headers"`
    Content      Content      `json:"content"`
    RedirectURL  string       `json:"redirectURL"`
    HeadersSize  int          `json:"headersSize"`
    BodySize     int          `json:"bodySize"`
    TransferSize int          `json:"_transferSize,omitempty"` // Chrome 扩展
    Error        any          `json:"_error,omitempty"`        // Chrome 扩展
    Comment      string       `json:"comment,omitempty"`
    CustomFields CustomFields `json:"-"`
}
```

`TransferSize` 是 `find --largest`、`performance` 评分中"传输大小"的真实来源；它只存在于浏览器导出的 HAR 中，工具导出的可能为 0。

## Headers / QueryString / Cookie

```go
type Headers struct {
    Name    string `json:"name"`
    Value   string `json:"value"`
    Comment string `json:"comment,omitempty"`
}

type QueryString struct {
    Name    string `json:"name"`
    Value   string `json:"value"`
    Comment string `json:"comment,omitempty"`
}

type Cookie struct {
    Name         string    `json:"name"`
    Value        string    `json:"value"`
    Path         string    `json:"path,omitempty"`
    Domain       string    `json:"domain,omitempty"`
    Expires      time.Time `json:"expires,omitempty"`
    HTTPOnly     bool      `json:"httpOnly,omitempty"`
    Secure       bool      `json:"secure,omitempty"`
    SameSite     string    `json:"sameSite,omitempty"`
    Comment      string    `json:"comment,omitempty"`
    CustomFields CustomFields `json:"-"`
}
```

`Cookie` 的安全属性（`HTTPOnly`/`Secure`/`SameSite`）正是 `cookie` 命令和 `CookieAudit()` 审计的对象。注意 `SameSite` 是字符串而非枚举，规范允许 `"Strict"`/`"Lax"`/`"None"`，但也可能为空。

## PostData / Param

```go
type PostData struct {
    MimeType     string       `json:"mimeType"`
    Params       []Param      `json:"params,omitempty"` // 表单提交时使用
    Text         string       `json:"text,omitempty"`   // 请求体文本
    Comment      string       `json:"comment,omitempty"`
    CustomFields CustomFields `json:"-"`
}

type Param struct {
    Name         string `json:"name"`
    Value        string `json:"value,omitempty"`
    FileName     string `json:"fileName,omitempty"`    // 文件上传
    ContentType  string `json:"contentType,omitempty"`
    Comment      string `json:"comment,omitempty"`
    CustomFields CustomFields `json:"-"`
}
```

`Params` 和 `Text` 通常二选一：`application/x-www-form-urlencoded` 用 `Params`，`application/json` 用 `Text`。`redact` 默认会对 `Params` 中名为 `password/secret/token` 的字段脱敏。

## Content

```go
type Content struct {
    Size        int    `json:"size"`                  // 内容大小（字节，解压后）
    MimeType    string `json:"mimeType"`
    Compression int    `json:"compression,omitempty"` // 压缩节省的字节数
    Text        string `json:"text,omitempty"`        // 文本内容
    Encoding    string `json:"encoding,omitempty"`    // 编码方式，如 "base64"
    Comment     string `json:"comment,omitempty"`
    CustomFields CustomFields `json:"-"`
}
```

`Encoding` 为 `"base64"` 时，`Text` 是二进制内容的 base64 编码（如图片）。`extract` 命令和 `lazy` 策略都围绕 `Content.Text` 工作：lazy 会把它的解析推迟到首次访问时。

## Cache 与 Timings

```go
type Cache struct {
    BeforeRequest *BeforeRequest `json:"beforeRequest,omitempty"`
    AfterRequest  *AfterRequest  `json:"afterRequest,omitempty"`
    Comment       string         `json:"comment,omitempty"`
    CustomFields  CustomFields   `json:"-"`
}

type Timings struct {
    Blocked         float64 `json:"blocked"`
    DNS             float64 `json:"dns"`
    Connect         float64 `json:"connect"`
    Ssl             float64 `json:"ssl"`
    Send            float64 `json:"send"`
    Wait            float64 `json:"wait"`
    Receive         float64 `json:"receive"`
    BlockedQueueing float64 `json:"_blocked_queueing,omitempty"` // Chrome 扩展
    BlockedProxy    float64 `json:"_blocked_proxy,omitempty"`    // Chrome 扩展
    Comment         string  `json:"comment,omitempty"`
    CustomFields    CustomFields `json:"-"`
}
```

::: warning Timings 的 -1 约定
HAR 规范规定：**未测量的时间字段值为 `-1`**，而不是 `0`。SDK 在 `timing`、`waterfall`、`PerformanceScore()` 中都会跳过 `-1` 字段。如果你自行遍历 `Timings`，务必先判断 `> 0` 再累加，否则会得到负数总和。
:::

`Cache.BeforeRequest`/`AfterRequest` 都是指针，表达"该阶段没有缓存信息"的常见情况。`FindCacheHits()` 依据 `HitCount > 0` 判定缓存命中。

## CustomFields 扩展机制

几乎所有结构体都带一个未导出的 `CustomFields CustomFields \`json:"-"\`` 字段。它不在 JSON 序列化中输出，但允许 SDK 内部和高级用户在内存对象上挂载自定义元数据（例如来源文件路径、解析警告等），而不污染规范字段。普通使用中无需关心它。

## 完整结构对照示例

下面是一条最小 `Entries` 在 Go 结构体与 JSON 中的对照，可作为理解整棵类型树的速查：

```go
entry := har.Entries{
    StartedDateTime: time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC),
    Time:            120.5,
    Request: har.Request{
        Method:      "GET",
        URL:         "https://api.example.com/users",
        HTTPVersion: "HTTP/2",
        Headers: []har.Headers{
            {Name: "Authorization", Value: "Bearer ***"},
        },
        QueryString: []har.QueryString{
            {Name: "page", Value: "1"},
        },
        HeadersSize: -1,
        BodySize:    0,
    },
    Response: har.Response{
        Status:      200,
        StatusText:  "OK",
        HTTPVersion: "HTTP/2",
        Content: har.Content{
            Size:     1024,
            MimeType: "application/json",
            Text:     `{"users":[]}`,
        },
        BodySize: 1024,
    },
    Timings: har.Timings{
        DNS: -1, Connect: -1, Ssl: -1, // 未测量
        Send: 1.0, Wait: 100.0, Receive: 19.5,
    },
    ServerIPAddress: "10.0.0.1",
    Connection:      "conn-1",
}
```

对应 JSON 片段：

```json
{
  "startedDateTime": "2024-01-02T03:04:05.000Z",
  "time": 120.5,
  "request": {
    "method": "GET",
    "url": "https://api.example.com/users",
    "httpVersion": "HTTP/2",
    "headers": [{ "name": "Authorization", "value": "Bearer ***" }],
    "queryString": [{ "name": "page", "value": "1" }],
    "headersSize": -1,
    "bodySize": 0
  },
  "response": {
    "status": 200,
    "statusText": "OK",
    "httpVersion": "HTTP/2",
    "content": { "size": 1024, "mimeType": "application/json", "text": "{\"users\":[]}" },
    "bodySize": 1024
  },
  "timings": { "dns": -1, "connect": -1, "ssl": -1, "send": 1, "wait": 100, "receive": 19.5 },
  "serverIPAddress": "10.0.0.1",
  "connection": "conn-1"
}
```

## 下一步

- 想知道这些结构体在不同解析策略下如何被存储，见 [解析策略](./parsing-strategies)。
- 想以接口抽象方式处理任意实现，见 [Provider 接口](./providers)。
- 想从 `Entries` 中筛选子集，见 [过滤与链式结果](./filtering)。
