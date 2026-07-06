---
title: Transform & Redact
titleTemplate: false
---

# Transform & Redact

The SDK provides two ways to rewrite a HAR: **transform** (`transform.go`) applies rule-based edits to request structure, while **redact** (`redact.go`) scrubs sensitive data. Both follow a "clone semantics" — they return a new `*Har` by default and leave the original untouched, so you can keep the raw snapshot in a pipeline.

All examples run from the repo root against `testdata/full.har` or `testdata/example.har`.

## Transform Rules

### The TransformRule struct

A rule is a `Type` plus a few fields. Different `Type` values read different fields; unused ones can stay empty.

```go
type TransformRule struct {
    Type        TransformType // transform type
    Pattern     string        // match pattern (regex, prefix, host, or scheme)
    Replacement string        // replacement string
    HeaderName  string        // header name (Header* and QueryParamAdd)
    HeaderValue string        // header value (HeaderAdd/HeaderReplace/QueryParamAdd)
}
```

### The ten TransformType values

| Constant | Purpose | Fields read |
| --- | --- | --- |
| `TransformURLRewrite` | Replace URL prefix (also rebuilds QueryString and Host header) | `Pattern`/`Replacement` |
| `TransformHostReplace` | Replace the host when it exactly matches | `Pattern`/`Replacement` |
| `TransformSchemeChange` | Switch the scheme (e.g. `http`->`https`) | `Pattern`/`Replacement` |
| `TransformHeaderAdd` | Append a header to both request and response | `HeaderName`/`HeaderValue` |
| `TransformHeaderRemove` | Remove a header by name from request and response | `HeaderName` |
| `TransformHeaderReplace` | Replace the value of a named header | `HeaderName`/`HeaderValue` |
| `TransformQueryParamRemove` | Remove a query parameter by name and rebuild the URL | `Pattern` |
| `TransformQueryParamAdd` | Append a query parameter and rebuild the URL | `HeaderName`/`HeaderValue` |
| `TransformCookieDomainRewrite` | Rewrite the Domain field of request and response cookies | `Pattern`/`Replacement` |
| `TransformBodyReplace` | Regex-replace PostData.Text (falls back to plain string replace on invalid regex) | `Pattern`/`Replacement` |

### Transform and TransformInPlace

`Transform` deep-clones then applies the rules to the clone, returning a new `*Har`. `TransformInPlace` mutates the receiver directly and returns nothing.

```go
package main

import (
    "fmt"
    "os"

    har "github.com/cyberspacesec/har-skills"
)

func main() {
    h, err := har.ParseHarFile("testdata/full.har")
    if err != nil {
        panic(err)
    }

    rules := []har.TransformRule{
        // 1. Rewrite URL prefix (staging -> prod)
        {Type: har.TransformURLRewrite, Pattern: "http://localhost:8080", Replacement: "https://api.example.com"},
        // 2. Remove sensitive headers
        {Type: har.TransformHeaderRemove, HeaderName: "Authorization"},
        {Type: har.TransformHeaderRemove, HeaderName: "Cookie"},
        // 3. Add a custom header
        {Type: har.TransformHeaderAdd, HeaderName: "X-Env", HeaderValue: "production"},
        // 4. Change scheme http -> https
        {Type: har.TransformSchemeChange, Pattern: "http", Replacement: "https"},
        // 5. Remove a query parameter (cache buster)
        {Type: har.TransformQueryParamRemove, Pattern: "_"},
        // 6. Rewrite cookie domain
        {Type: har.TransformCookieDomainRewrite, Pattern: "staging.local", Replacement: "example.com"},
    }

    // Clone semantics: h is unchanged
    transformed := h.Transform(rules)

    // In-place version: mutates h directly
    // h.TransformInPlace(rules)

    if err := transformed.SaveToFile("transformed.har", true); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

## Convenience methods

For the three most common operations the SDK wraps `Transform` with a single-rule shortcut:

```go
// RewriteURL replaces a URL prefix and returns a new *Har
prod := h.RewriteURL("http://localhost:8080", "https://api.example.com")

// RemoveHeaders drops multiple headers from request and response
cleaned := h.RemoveHeaders([]string{"Authorization", "Cookie", "Set-Cookie"})

// AddHeaders appends headers to request and/or response; target is "request"/"response"/"both"
// This convenience appends directly to the slices, bypassing TransformInPlace.
withEnv := h.AddHeaders(map[string]string{
    "X-Env":   "production",
    "X-Trace": "abc123",
}, "request")
```

| Method | Signature | Notes |
| --- | --- | --- |
| `RewriteURL` | `(from, to string) *Har` | URL prefix replacement, equivalent to a single `TransformURLRewrite` |
| `RemoveHeaders` | `(names []string) *Har` | Batch header removal, one `TransformHeaderRemove` per name |
| `AddHeaders` | `(headers map[string]string, target string) *Har` | Appends to request/response/both per `target` |

## Redact

Redaction scrubs passwords, tokens, and API keys before sharing a HAR. Like transform, `Redact` returns a new `*Har` (it `Clone`s first) and `RedactInPlace` mutates in place.

### DefaultRedactOptions targets

`DefaultRedactOptions()` ships with sensible defaults for common sensitive fields — use it as-is or append to it:

| Category | Default targets |
| --- | --- |
| Headers | `Authorization`, `Proxy-Authorization`, `WWW-Authenticate`, `Cookie`, `Set-Cookie`, `X-Api-Key`, `X-Auth-Token`, `X-CSRF-Token` |
| Cookies | `session`, `token`, `auth`, `password`, `secret`, `api_key`, `access_token`, `refresh_token` |
| QueryParams | `password`, `token`, `api_key`, `secret`, `access_token`, `refresh_token`, `private_key`, `client_secret` |
| PostDataFields | same as QueryParams |

The default replacement text is `[REDACTED]` and `RedactIPs` defaults to `false`.

### RedactOptions fields

```go
type RedactOptions struct {
    Headers        []string
    Cookies        []string
    QueryParams    []string
    PostDataFields []string
    Replacement    string                                                   // default "[REDACTED]"
    RedactIPs      bool                                                     // anonymize ServerIPAddress
    RedactURLs     []RedactURLRule                                          // URL path-segment redaction
    CustomRedactor func(fieldType, name, value string) string              // custom callback
}

type RedactURLRule struct {
    Pattern     string // regex matching a URL path segment
    Replacement string // replacement text
}
```

`CustomRedactor` is invoked whenever any target is hit. Its `fieldType` argument is one of `header`/`cookie`/`queryparam`/`postdatafield`, and its return value replaces the original. When set, it overrides the default `[REDACTED]` replacement.

### Full redaction example

```go
package main

import (
    "fmt"
    "os"

    har "github.com/cyberspacesec/har-skills"
)

func main() {
    h, err := har.ParseHarFile("testdata/full.har")
    if err != nil {
        panic(err)
    }

    opts := har.DefaultRedactOptions()

    // Append custom headers and query params
    opts.Headers = append(opts.Headers, "X-Custom-Key", "X-Internal-Token")
    opts.QueryParams = append(opts.QueryParams, "sig")

    // Custom replacement text
    opts.Replacement = "***REDACTED***"

    // Anonymize server IP (IPv4 last octet -> .0, IPv6 last segment -> :0)
    opts.RedactIPs = true

    // Redact numeric ID path segments, e.g. /users/12345 -> /users/[id]
    opts.RedactURLs = []har.RedactURLRule{
        {Pattern: `^\d+$`, Replacement: "[id]"},
    }

    // Custom callback: partially mask tokens, keeping the first 4 chars
    opts.CustomRedactor = func(fieldType, name, value string) string {
        if name == "token" && len(value) > 4 {
            return value[:4] + "****"
        }
        return "***REDACTED***"
    }

    // Clone semantics: h is unchanged
    redacted := h.Redact(opts)

    if err := redacted.SaveToFile("redacted.har", true); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

### Redaction coverage

`RedactInPlace` walks each Entry and scrubs the following locations:

- Request and response headers (matched by name, case-insensitive)
- `Value` of request and response cookies
- `QueryString` parameter values
- Query parameters embedded in the URL string (parses the URL and rewrites `RawQuery`)
- URL path segments (via `RedactURLs` rules)
- PostData: `Params` field values plus `Text` matching `key=value` form bodies or JSON `"key": "value"` patterns
- `ServerIPAddress` (when `RedactIPs=true`)

## Design notes

- **Clone semantics**: `Transform`, `RewriteURL`, `RemoveHeaders`, `AddHeaders`, and `Redact` all `Clone()` first, so the original `*Har` is preserved for diffing or rollback.
- **URL consistency**: URL-rewriting transforms rebuild `QueryString` and the `Host` header alongside the URL so the three never drift apart.
- **Case-insensitive matching**: header and cookie names match case-insensitively, consistent with HTTP semantics.
- **Regex tolerance**: `TransformBodyReplace` falls back to plain string replacement when the regex fails to compile, so a bad pattern never aborts the whole rule batch.
