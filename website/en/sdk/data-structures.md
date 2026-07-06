---
title: Data Structures
titleTemplate: false
---

# Data Structures

`*Har` is the central type of the entire SDK, defined in `har.go`. Its struct tree maps the HAR 1.2 spec in full while preserving Chrome DevTools extension fields (`_initiator`, `_priority`, `_resourceType`, `_transferSize`, etc.). Understanding this tree is a prerequisite for using all 70+ methods and writing custom analysis logic.

## Type tree overview

```
Har
└── Log
    ├── Version        string            // HAR spec version, e.g. "1.2"
    ├── Creator        Creator           // tool that produced the HAR
    ├── Browser        Browser           // browser info (optional)
    ├── Pages          []Pages           // page info (optional)
    │   ├── StartedDateTime time.Time
    │   ├── ID              string
    │   ├── Title           string
    │   └── PageTimings     PageTimings
    │       ├── OnContentLoad float64    // DOMContentLoaded (ms)
    │       └── OnLoad        float64    // load event (ms)
    ├── Entries        []Entries         // HTTP request/response entries
    └── Comment        string            // optional comment
```

`Har` itself holds only a `Log` field and an unexported `CustomFields`; almost all business data lives inside `Log`. `Log.Entries` is the main object that subsequent filtering, analysis, export, and replay operate on.

## Creator and Browser

These two structs are the simplest: they record who generated the HAR and with which tool or browser.

```go
// Creator holds info about the tool that created the HAR file
type Creator struct {
    Name    string `json:"name"`
    Version string `json:"version"`
    Comment string `json:"comment,omitempty"`
}

// Browser holds browser info
type Browser struct {
    Name    string `json:"name"`
    Version string `json:"version"`
    Comment string `json:"comment,omitempty"`
}
```

Corresponding JSON:

```json
"creator": { "name": "WebInspector", "version": "537.36" },
"browser": { "name": "Chrome", "version": "120.0.0.0" }
```

## Pages and PageTimings

`Pages` describes the metadata of a page load plus two key timing points. A single HAR file can contain multiple pages (e.g. a multi-tab capture).

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
    OnContentLoad float64 `json:"onContentLoad"` // time of DOMContentLoaded (ms)
    OnLoad        float64 `json:"onLoad"`        // time of load event (ms)
    Comment       string  `json:"comment,omitempty"`
}
```

`Entries.Pageref` ties an individual request to its page via `Pages.ID`. This is the data behind the `waterfall --page-timings` command.

## Entries — the core entry

`Entries` is the most frequently used type in the SDK. It corresponds to one complete HTTP transaction: request, response, timings, cache, and Chrome extension metadata.

```go
type Entries struct {
    StartedDateTime time.Time `json:"startedDateTime"`           // request start time
    Time            float64   `json:"time"`                      // total duration (ms)
    Request         Request   `json:"request"`
    Response        Response  `json:"response"`
    Cache           Cache     `json:"cache"`
    Timings         Timings   `json:"timings"`
    Pageref         string    `json:"pageref,omitempty"`         // linked page ID
    ServerIPAddress string    `json:"serverIPAddress,omitempty"` // server IP
    Connection      string    `json:"connection,omitempty"`      // connection ID (reuse analysis)
    Initiator       Initiator `json:"_initiator,omitempty"`      // request initiator (Chrome ext)
    Priority        string    `json:"_priority,omitempty"`       // request priority (Chrome ext)
    ResourceType    string    `json:"_resourceType,omitempty"`   // resource type (Chrome ext)
    Comment         string    `json:"comment,omitempty"`
    CustomFields    CustomFields `json:"-"`
}
```

`_initiator`, `_priority`, and `_resourceType` are Chrome DevTools extension fields present only in browser-exported HAR. The SDK treats them as ordinary fields; capabilities like `find --resource-type`, `connections`, and `FindByResourceType` build on them.

### Request

```go
type Request struct {
    Method       string        `json:"method"`
    URL          string        `json:"url"`
    HTTPVersion  string        `json:"httpVersion"`
    Cookies      []Cookie      `json:"cookies"`
    Headers      []Headers     `json:"headers"`
    QueryString  []QueryString `json:"queryString"`
    PostData     *PostData     `json:"postData,omitempty"` // optional, only for body-bearing requests
    HeadersSize  int           `json:"headersSize"`
    BodySize     int           `json:"bodySize"`
    Comment      string        `json:"comment,omitempty"`
    CustomFields CustomFields  `json:"-"`
}
```

`PostData` is a pointer, expressing "optional value" — this is the very idea the `optimized` strategy generalizes with pointers for optional fields. In the standard implementation `Headers` is a slice; in the optimized implementation it is rewritten as `map[string][]string` to speed up lookup.

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
    TransferSize int          `json:"_transferSize,omitempty"` // Chrome ext
    Error        any          `json:"_error,omitempty"`        // Chrome ext
    Comment      string       `json:"comment,omitempty"`
    CustomFields CustomFields `json:"-"`
}
```

`TransferSize` is the real source of "transfer size" in `find --largest` and the `performance` score; it only exists in browser-exported HAR and may be 0 for tool-exported files.

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

The security attributes of `Cookie` (`HTTPOnly`/`Secure`/`SameSite`) are exactly what the `cookie` command and `CookieAudit()` inspect. Note `SameSite` is a string, not an enum: the spec allows `"Strict"`/`"Lax"`/`"None"`, but it may also be empty.

## PostData / Param

```go
type PostData struct {
    MimeType     string       `json:"mimeType"`
    Params       []Param      `json:"params,omitempty"` // used for form submission
    Text         string       `json:"text,omitempty"`   // request body text
    Comment      string       `json:"comment,omitempty"`
    CustomFields CustomFields `json:"-"`
}

type Param struct {
    Name         string `json:"name"`
    Value        string `json:"value,omitempty"`
    FileName     string `json:"fileName,omitempty"`    // file upload
    ContentType  string `json:"contentType,omitempty"`
    Comment      string `json:"comment,omitempty"`
    CustomFields CustomFields `json:"-"`
}
```

`Params` and `Text` are usually mutually exclusive: `application/x-www-form-urlencoded` uses `Params`, `application/json` uses `Text`. The `redact` command by default redacts `Params` named `password/secret/token`.

## Content

```go
type Content struct {
    Size        int    `json:"size"`                  // content size in bytes (decompressed)
    MimeType    string `json:"mimeType"`
    Compression int    `json:"compression,omitempty"` // bytes saved by compression
    Text        string `json:"text,omitempty"`        // text content
    Encoding    string `json:"encoding,omitempty"`    // encoding, e.g. "base64"
    Comment     string `json:"comment,omitempty"`
    CustomFields CustomFields `json:"-"`
}
```

When `Encoding` is `"base64"`, `Text` is the base64-encoded binary content (e.g. an image). Both the `extract` command and the `lazy` strategy revolve around `Content.Text`: lazy defers its parsing until first access.

## Cache and Timings

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
    BlockedQueueing float64 `json:"_blocked_queueing,omitempty"` // Chrome ext
    BlockedProxy    float64 `json:"_blocked_proxy,omitempty"`    // Chrome ext
    Comment         string  `json:"comment,omitempty"`
    CustomFields    CustomFields `json:"-"`
}
```

::: warning The -1 convention for Timings
The HAR spec mandates that **an unmeasured timing field has the value `-1`**, not `0`. The SDK skips `-1` fields in `timing`, `waterfall`, and `PerformanceScore()`. If you iterate `Timings` yourself, always check `> 0` before summing, otherwise you will get a negative total.
:::

`Cache.BeforeRequest`/`AfterRequest` are both pointers, expressing the common case of "no cache info for this phase". `FindCacheHits()` treats `HitCount > 0` as a cache hit.

## The CustomFields extension mechanism

Nearly every struct carries an unexported `CustomFields CustomFields \`json:"-"\`` field. It is not emitted in JSON serialization, but lets the SDK internals and advanced users attach custom metadata to in-memory objects (e.g. source file path, parse warnings) without polluting spec fields. You can ignore it in normal usage.

## Full structure side-by-side example

Below is a minimal `Entries` shown in both the Go struct and JSON form — a cheat sheet for understanding the whole type tree:

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
        DNS: -1, Connect: -1, Ssl: -1, // unmeasured
        Send: 1.0, Wait: 100.0, Receive: 19.5,
    },
    ServerIPAddress: "10.0.0.1",
    Connection:      "conn-1",
}
```

Corresponding JSON fragment:

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

## Next steps

- For how these structs are stored under different parsing strategies, see [Parsing strategies](./parsing-strategies).
- For treating any implementation through interface abstraction, see [Provider interfaces](./providers).
- For selecting a subset of `Entries`, see [Filtering and chained results](./filtering).
