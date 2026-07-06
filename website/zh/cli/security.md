---
title: 安全与隐私
titleTemplate: false
---

# 安全与隐私

HAR 文件是浏览器的「现场录像」，里面常常混着令牌、Cookie、内网 IP、密钥。在分享或归档前，先做两件事：**审计**（`security`）找出风险点，**脱敏**（`redact`）把敏感值抹掉。这两个命令覆盖了 Level 3 的全部能力。

所有示例都可在仓库根目录直接运行，使用 `testdata/example.har` 或 `testdata/full.har`。

## security — 安全审计

对 HAR 做一次综合体检，输出 **0–100 的安全评分**和按严重性分组的发现清单（HIGH / MEDIUM / LOW / INFO）。评分越低说明风险越密集；发现项会给出 URL、类别、描述与修复建议。

```bash
har -f testdata/full.har security
```

输出形如：

```text
安全审计报告
============
评分: 72/100
发现: 8 个问题

[HIGH] 严重 (2)
------------------------------------------------------------
  1. 缺失 HSTS 头
     URL: http://example.com/login
     类别: security-headers
     描述: 响应未设置 Strict-Transport-Security，存在 SSL 剥离风险
     修复: 在响应头中添加 HSTS，至少 max-age=31536000
  ...
```

### 检查项

`security` 把审计拆成 6 类，可单独开关：

| 检查项 | 关注内容 |
|--------|----------|
| 安全头 | HSTS、Content-Security-Policy、X-Frame-Options、X-Content-Type-Options、Referrer-Policy 等是否缺失或配置薄弱 |
| Cookie 安全 | Secure / HttpOnly / SameSite 属性、会话 Cookie 是否走明文、第三方 Cookie |
| 混合内容 | HTTPS 页面里夹带的 HTTP 子资源（脚本、图片、iframe） |
| 敏感数据 | 响应或表单里出现的密码、令牌、API Key、私钥等 |
| CORS | `Access-Control-Allow-Origin: *` 配合凭据、通配符回显等危险组合 |
| 信息泄露 | `Server`、`X-Powered-By`、堆栈、版本号、内部路径等暴露面 |

### Flags

| Flag | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--check-headers` | bool | `true` | 检查安全头部 |
| `--check-cookies` | bool | `true` | 检查 Cookie 安全性 |
| `--check-mixed-content` | bool | `true` | 检查混合内容 |
| `--check-sensitive-data` | bool | `true` | 检查敏感数据泄露 |
| `--check-cors` | bool | `true` | 检查 CORS 配置 |
| `--check-info-disclosure` | bool | `true` | 检查信息泄露 |
| `--severity` | string | `low` | 最低严重性过滤（`all`/`info`/`low`/`medium`/`high`） |

### 示例

只看高危发现，输出 JSON 方便进 CI：

```bash
har -f testdata/full.har security --severity high --format json -o sec-high.json
```

只跑头部与 CORS 两类检查，快速过一遍：

```bash
har -f testdata/full.har security --check-headers --check-cors
```

关掉全部默认项，只查混合内容（适合迁移 HTTPS 时定位漏网之鱼）：

```bash
har -f testdata/full.har security \
  --check-headers=false --check-cookies=false \
  --check-sensitive-data=false --check-cors=false \
  --check-info-disclosure=false --check-mixed-content
```

`--severity all` 连 INFO 级也显示，适合做完整基线：

```bash
har -f testdata/full.har security --severity all
```

### 实现原理

底层调用 `(*Har).SecurityAudit()`，返回 `*SecurityReport`。报告结构里有 `Score int`、`Findings []SecurityFinding`，每条发现含 `Severity`、`Title`、`Category`、`Description`、`Remedy`、`EntryURL`。CLI 把 6 个 `--check-*` flag 映射成 `SecurityAuditOptions` 的开关位，再按 `--severity` 用 `report.FindBySeverity(sev)` 过滤。文本输出由 `formatSecurityReport` 拼装，JSON/YAML 直接序列化整个 `SecurityReport`。

## redact — 脱敏

把敏感值替换成占位符，**输出一个新的 HAR 文件**（除非加 `--in-place`）。原始文件不被改动，适合在保留结构的前提下安全分享。

```bash
har -f testdata/full.har redact -o redacted.har
```

### 默认脱敏目标

`--defaults`（默认开启）会自动覆盖以下目标：

| 位置 | 目标 |
|------|------|
| 请求头 | `Authorization`、`Proxy-Authorization`、`X-Api-Key`、`X-Auth-Token` |
| Cookie | 名称为 `session`、`token`、`auth`、`password` 的 Cookie |
| 查询参数 | `password`、`token`、`api_key`、`secret`、`access_token` |
| POST 字段 | `password`、`secret`、`token` |

需要额外覆盖时，用 `--header` / `--cookie` / `--query-param` / `--post-field` 追加，不影响默认集。

### Flags

| Flag | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--defaults` | bool | `true` | 启用默认脱敏规则集 |
| `--header` | stringSlice | `[]` | 额外脱敏的请求头字段名 |
| `--cookie` | stringSlice | `[]` | 额外脱敏的 Cookie 名称 |
| `--query-param` | stringSlice | `[]` | 额外脱敏的查询参数名 |
| `--post-field` | stringSlice | `[]` | 额外脱敏的 POST 字段名 |
| `--replacement` | string | `[REDACTED]` | 替换占位文本 |
| `--redact-ips` | bool | `false` | 匿名化 IP 地址 |
| `--in-place` | bool | `false` | 原地改写输入文件 |

### 示例

在默认集之外，额外抹掉自定义头 `X-Custom-Key` 和 Cookie `session_id`：

```bash
har -f testdata/full.har redact \
  --header X-Custom-Key \
  --cookie session_id \
  -o redacted.har
```

把替换文本改成 `***`，并匿名化所有 IP（适合发给外包排查）：

```bash
har -f testdata/full.har redact --redact-ips --replacement "***" -o clean.har
```

直接改写原文件（慎用，会覆盖；建议先备份）：

```bash
har -f testdata/full.har redact --in-place
```

只脱敏查询参数，关掉默认集：

```bash
har -f testdata/full.har redact \
  --defaults=false \
  --query-param sig --query-param nonce \
  -o params-only.har
```

脱敏后顺手做一次校验，确认文件仍合规：

```bash
har -f redacted.har validate --strict
```

::: warning redact 不等于加密
`redact` 是**结构化抹除**，不是加密或哈希。被替换的值无法还原，但 HAR 的请求/响应结构保持完整，便于对方继续做性能或缓存分析。对极敏感场景，建议 `redact` 之后再 `validate`，并人工抽查几条。
:::

### 实现原理

底层调用 `(*Har).Redact(opts)`，返回一个新的 `*Har`（原对象不变）。`opts` 来自 `har.DefaultRedactOptions()`，CLI 把 `--header` 等 stringSlice 合并进对应字段，`--replacement` 写入 `Replacement`，`--redact-ips` 打开 `RedactIPs`。脱敏遍历每条 entry 的 headers、cookies、queryString、postData.params，命中目标名即替换为占位文本；IP 匿名化走单独的地址改写逻辑。`--in-place` 时把结果 `ToJSON(true)` 写回原路径，否则写到 `-o` 指定的文件（缺省输出到 stdout）。

## 小结

| 命令 | 用途 | 是否改文件 |
|------|------|-----------|
| `security` | 审计风险、给评分 | 否，只读 |
| `redact` | 抹除敏感值 | 是，输出新文件或 `--in-place` |

典型链路：先 `security` 找出风险点与泄漏面，再 `redact` 把值抹干净，最后 `validate` 复核。完整工作流见 [安全审计工作流](../workflows/security-audit.md)。
