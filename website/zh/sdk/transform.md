---
title: 转换与脱敏
titleTemplate: false
---

# 转换与脱敏

SDK 提供两类改写 HAR 的能力：**转换**（`transform.go`）按规则改写请求结构，**脱敏**（`redact.go`）抹掉敏感数据。两者都遵循"克隆语义"——默认返回新的 `*Har`，原对象保持不变，便于在流水线中保留原始快照。

所有示例均可在仓库根目录运行，使用 `testdata/full.har` 或 `testdata/example.har`。

## 转换规则

### TransformRule 结构

一条规则由 `Type` 和若干字段组成。不同 `Type` 读取不同字段，未用到的字段可留空。

```go
type TransformRule struct {
    Type        TransformType // 转换类型
    Pattern     string        // 匹配模式（正则或前缀/主机名/协议）
    Replacement string        // 替换字符串
    HeaderName  string        // 头部名称（用于 Header* 与 QueryParamAdd）
    HeaderValue string        // 头部值（用于 HeaderAdd/HeaderReplace/QueryParamAdd）
}
```

### 十种 TransformType

| 常量 | 用途 | 读取字段 |
| --- | --- | --- |
| `TransformURLRewrite` | 替换 URL 前缀（同时重建 QueryString 与 Host 头） | `Pattern`/`Replacement` |
| `TransformHostReplace` | 按主机名精确匹配后替换 | `Pattern`/`Replacement` |
| `TransformSchemeChange` | 切换协议（如 `http`→`https`） | `Pattern`/`Replacement` |
| `TransformHeaderAdd` | 向请求与响应同时追加头部 | `HeaderName`/`HeaderValue` |
| `TransformHeaderRemove` | 按名称移除请求与响应中的头部 | `HeaderName` |
| `TransformHeaderReplace` | 替换指定名称头部值 | `HeaderName`/`HeaderValue` |
| `TransformQueryParamRemove` | 按名称移除查询参数并重建 URL | `Pattern` |
| `TransformQueryParamAdd` | 追加一个查询参数并重建 URL | `HeaderName`/`HeaderValue` |
| `TransformCookieDomainRewrite` | 重写请求与响应 Cookie 的 Domain 字段 | `Pattern`/`Replacement` |
| `TransformBodyReplace` | 正则替换 PostData.Text（非法正则退化为字符串替换） | `Pattern`/`Replacement` |

### Transform 与 TransformInPlace

`Transform` 先深拷贝再在副本上应用规则，返回新 `*Har`；`TransformInPlace` 直接修改原对象，无返回值。

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
        // 1. 重写 URL 前缀（staging -> prod）
        {Type: har.TransformURLRewrite, Pattern: "http://localhost:8080", Replacement: "https://api.example.com"},
        // 2. 移除敏感头
        {Type: har.TransformHeaderRemove, HeaderName: "Authorization"},
        {Type: har.TransformHeaderRemove, HeaderName: "Cookie"},
        // 3. 添加自定义头
        {Type: har.TransformHeaderAdd, HeaderName: "X-Env", HeaderValue: "production"},
        // 4. 改协议 http -> https
        {Type: har.TransformSchemeChange, Pattern: "http", Replacement: "https"},
        // 5. 移除查询参数（缓存破坏参数）
        {Type: har.TransformQueryParamRemove, Pattern: "_"},
        // 6. 重写 Cookie 域
        {Type: har.TransformCookieDomainRewrite, Pattern: "staging.local", Replacement: "example.com"},
    }

    // 克隆语义：原 h 不变
    transformed := h.Transform(rules)

    // 原地版本：直接修改 h
    // h.TransformInPlace(rules)

    if err := transformed.SaveToFile("transformed.har", true); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

## 便捷方法

对于最常见的三类操作，SDK 提供了封装好的便捷方法，内部仍是构造 `TransformRule` 后调用 `Transform`：

```go
// RewriteURL 替换 URL 前缀，返回新 *Har
prod := h.RewriteURL("http://localhost:8080", "https://api.example.com")

// RemoveHeaders 从请求与响应中移除多个头，返回新 *Har
cleaned := h.RemoveHeaders([]string{"Authorization", "Cookie", "Set-Cookie"})

// AddHeaders 向请求和/或响应添加头，target 取 "request"/"response"/"both"
// 便捷版内部直接追加到对应切片，不经过 TransformInPlace
withEnv := h.AddHeaders(map[string]string{
    "X-Env":   "production",
    "X-Trace": "abc123",
}, "request")
```

| 方法 | 签名 | 说明 |
| --- | --- | --- |
| `RewriteURL` | `(from, to string) *Har` | URL 前缀替换，等价单条 `TransformURLRewrite` |
| `RemoveHeaders` | `(names []string) *Har` | 批量移除头，每个名称生成一条 `TransformHeaderRemove` |
| `AddHeaders` | `(headers map[string]string, target string) *Har` | 按 target 追加到 request/response/both |

## 脱敏 Redact

脱敏用于在分享 HAR 前抹掉密码、令牌、API Key 等敏感字段。与转换一样，`Redact` 返回新 `*Har`（先 `Clone` 再脱敏），`RedactInPlace` 原地修改。

### DefaultRedactOptions 默认清单

`DefaultRedactOptions()` 内置了 HTTP 流量中常见的敏感字段，可直接使用或在其基础上追加：

| 类别 | 默认目标 |
| --- | --- |
| Headers | `Authorization`、`Proxy-Authorization`、`WWW-Authenticate`、`Cookie`、`Set-Cookie`、`X-Api-Key`、`X-Auth-Token`、`X-CSRF-Token` |
| Cookies | `session`、`token`、`auth`、`password`、`secret`、`api_key`、`access_token`、`refresh_token` |
| QueryParams | `password`、`token`、`api_key`、`secret`、`access_token`、`refresh_token`、`private_key`、`client_secret` |
| PostDataFields | 同 QueryParams |

默认替换文本为 `[REDACTED]`，`RedactIPs` 默认 `false`。

### RedactOptions 字段

```go
type RedactOptions struct {
    Headers        []string
    Cookies        []string
    QueryParams    []string
    PostDataFields []string
    Replacement    string                                                   // 默认 "[REDACTED]"
    RedactIPs      bool                                                     // 是否匿名化 ServerIPAddress
    RedactURLs     []RedactURLRule                                          // URL 路径段正则脱敏
    CustomRedactor func(fieldType, name, value string) string              // 自定义回调
}

type RedactURLRule struct {
    Pattern     string // 匹配 URL 路径段的正则
    Replacement string // 替换文本
}
```

`CustomRedactor` 在命中任意目标时被调用，参数 `fieldType` 取值为 `header`/`cookie`/`queryparam`/`postdatafield`，返回值即脱敏后的值。设置后默认的 `[REDACTED]` 替换被覆盖。

### 完整脱敏示例

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

    // 追加自定义头与查询参数
    opts.Headers = append(opts.Headers, "X-Custom-Key", "X-Internal-Token")
    opts.QueryParams = append(opts.QueryParams, "sig")

    // 自定义替换文本
    opts.Replacement = "***REDACTED***"

    // 匿名化服务器 IP（IPv4 末位段置 .0，IPv6 末段置 :0）
    opts.RedactIPs = true

    // 脱敏 URL 路径中的数字 ID 段，例如 /users/12345 -> /users/[id]
    opts.RedactURLs = []har.RedactURLRule{
        {Pattern: `^\d+$`, Replacement: "[id]"},
    }

    // 自定义回调：对 token 做部分遮罩，保留前 4 位
    opts.CustomRedactor = func(fieldType, name, value string) string {
        if name == "token" && len(value) > 4 {
            return value[:4] + "****"
        }
        return "***REDACTED***"
    }

    // 克隆语义：原 h 不变
    redacted := h.Redact(opts)

    if err := redacted.SaveToFile("redacted.har", true); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

### 脱敏覆盖范围

`RedactInPlace` 会遍历每个 Entry 的下列位置：

- 请求头与响应头（按名称匹配，大小写不敏感）
- 请求 Cookie 与响应 Cookie 的 `Value`
- `QueryString` 参数值
- URL 字符串中的查询参数（重新解析 URL 后重写 `RawQuery`）
- URL 路径段（应用 `RedactURLs` 规则）
- PostData：`Params` 字段值 + `Text` 中的 `key=value` 表单或 JSON `"key": "value"` 模式
- `ServerIPAddress`（当 `RedactIPs=true`）

## 设计要点

- **克隆语义**：`Transform`、`RewriteURL`、`RemoveHeaders`、`AddHeaders`、`Redact` 都先 `Clone()` 再改，原始 `*Har` 保持不变，便于对比与回滚。
- **URL 一致性**：URL 类转换在改写后会同步重建 `QueryString` 与 `Host` 头，避免三者不一致。
- **大小写不敏感**：头部与 Cookie 名称匹配均为大小写不敏感，与 HTTP 语义一致。
- **正则容错**：`TransformBodyReplace` 在正则编译失败时退化为普通字符串替换，不会中断整批规则。
