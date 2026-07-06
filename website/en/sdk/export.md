---
title: Export
titleTemplate: false
---

# Export

The SDK's export surface spans four modules: `converter.go` (tabular formats CSV/Markdown/HTML/Text), `export.go` (replay commands and structured formats cURL/Wget/Python/Postman/XML), `format.go` (YAML), and `util.go` (JSON and file writing). All methods hang off `*Har`, and some have single-entry `*Entries` counterparts.

## Convert: tabular formats

`Convert(format, opts)` turns the HAR into tabular text, suitable for reports or spreadsheet import. `ConvertFormat` constants are defined in `converter.go`; `FormatYAML` lives in `format.go` but shares the same `ConvertFormat` type.

```go
const (
    FormatCSV      ConvertFormat = "csv"
    FormatMarkdown ConvertFormat = "markdown"
    FormatHTML     ConvertFormat = "html"
    FormatText     ConvertFormat = "text"
)
const FormatYAML ConvertFormat = "yaml"
```

### ConvertOptions field selection

Boolean toggles decide which columns are exported; disabled fields do not appear in the output:

```go
type ConvertOptions struct {
    IncludeURL         bool
    IncludeMethod      bool
    IncludeStatus      bool
    IncludeContentType bool
    IncludeSize        bool
    IncludeTime        bool
    IncludeTimings     bool   // six columns: blocked/DNS/connect/send/wait/receive
    IncludeHeaders     bool   // request + response headers
    IncludeDateTime    bool
    IncludePostData    bool   // POST type + text
    IncludeQueryString bool
    Headers            []string       // custom headers (overrides defaults)
    Filter             *FilterOptions // filter before converting
}
```

`DefaultConvertOptions()` enables URL/Method/Status/ContentType/Size/Time/DateTime by default and leaves the rest off.

### Convert and ConvertWith

`Convert(format, opts)` takes a struct; `ConvertWith(format, opts...)` takes functional options, which is nicer for chaining:

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

    // Struct options: CSV with timings
    opts := har.DefaultConvertOptions()
    opts.IncludeTimings = true
    opts.IncludeHeaders = false
    csv, err := h.Convert(har.FormatCSV, opts)
    if err != nil {
        panic(err)
    }
    fmt.Println(csv)

    // Functional options: Markdown
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

Available `WithConvert*` options: `IncludeHeaders/Timings/Bodies/Cookies/QueryStrings/Status/Size/URL/Method/Time/MimeType`, `Headers`, `Filter`.

## Replay command export

These three methods return `string`, producing scripts you can run in a terminal:

```go
curl    := h.ToCurl()              // curl -H '...' --data '...' 'URL'
wget    := h.ToWget()              // wget --header='...' --post-data='...' -qO- 'URL'
python  := h.ToPythonRequests()    // import requests + requests.get/post(...)
```

Generation details:

- `ToCurl` skips the `Host` header (curl adds it); non-GET methods get `-X`; `Accept-Encoding: gzip/deflate` triggers `--compressed`; HTTPS with a response error triggers `-k`.
- `ToWget` skips `Host`; non-GET uses `--method=`; HTTPS gets `--no-check-certificate`; defaults to `-qO-` (stdout).
- `ToPythonRequests` emits `import requests`, then per entry a `headers = {...}` block and `response = requests.<method>(...)`.

Each method also has a same-named `*Entries` form for a single entry:

```go
first := &h.Log.Entries[0]
fmt.Println(first.ToCurl())
fmt.Println(first.ToPythonRequests())
```

## Structured format export

### Postman Collection v2.1

`ToPostmanCollection()` returns `([]byte, error)`, producing JSON that conforms to the Postman v2.1 schema and imports directly into Postman. `SaveAsPostmanCollection(path)` is the file-writing convenience.

```go
data, err := h.ToPostmanCollection()
if err != nil {
    panic(err)
}
// Write straight to disk
if err := h.SaveAsPostmanCollection("collection.json"); err != nil {
    panic(err)
}
```

### XML

`ToXML()` returns `(string, error)` and emits standard XML with an `<?xml ...?>` header. It uses `encoding/xml` struct mapping covering version/creator/entries/request/response/headers/content/postData.

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

`ToYAML()` returns `(string, error)`. The implementation depends on no external YAML library — it calls `ToJSON(true)` then runs a built-in JSON-to-YAML converter with special-character escaping for strings.

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

`ToJSON(indent bool)` returns `([]byte, error)` and underpins the other structured formats. With `indent=true` the output is pretty-printed.

```go
data, err := h.ToJSON(true) // pretty-printed
if err != nil {
    panic(err)
}
```

## Writing files

`util.go` and the export modules provide file-writing conveniences:

| Method | Signature | Notes |
| --- | --- | --- |
| `SaveToFile` | `(filePath string, indent bool) error` | Write JSON, control indentation |
| `SaveToFileGzipped` | `(filePath string, indent bool) error` | Write gzip-compressed JSON |
| `SaveToWriter` | `(w io.Writer, indent bool) error` | Write JSON to any Writer |
| `SaveAsPostmanCollection` | `(filePath string) error` | Write Postman v2.1 JSON |
| `SaveAsXML` | `(filePath string) error` | Write XML |
| `SaveAsYAML` | `(filePath string) error` | Write YAML |

## Streaming export with ConvertTo

`ConvertTo(format, w, opts)` writes the converted output directly to an `io.Writer`, avoiding a full in-memory string — ideal for exporting large files to disk or an HTTP response body:

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

    // Streamed write, no intermediate string
    if err := h.ConvertTo(har.FormatCSV, f, opts); err != nil {
        panic(err)
    }
}
```

`ConvertTo` accepts `FormatYAML`, `FormatCSV`, `FormatMarkdown`, `FormatHTML`, and `FormatText`; any other value falls back to indented JSON. It nil-checks the writer and returns a `*HarError` on failure.

## Putting it together: generate a cURL replay script from a HAR

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

    // Replay only API requests
    api := h.FindByDomain("api.example.com")
    replays := api.ToHar().ToCurl()

    if err := os.WriteFile("replay.sh", []byte("#!/bin/bash\n\n"+replays+"\n"), 0644); err != nil {
        panic(err)
    }
    _ = os.Chmod("replay.sh", 0755)
}
```
