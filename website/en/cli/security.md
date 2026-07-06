---
title: Security & Privacy
titleTemplate: false
---

# Security & Privacy

A HAR file is a "recording" of a browser session, and it routinely carries tokens, cookies, internal IPs, and keys. Before sharing or archiving one, do two things: **audit** (`security`) to surface the risks, then **redact** (`redact`) to scrub the sensitive values. These two commands cover the entire Level 3 capability set.

Every example below runs from the repository root against `testdata/example.har` or `testdata/full.har`.

## security — Security Audit

Run a full checkup on the HAR. Output is a **0–100 security score** plus a list of findings grouped by severity (HIGH / MEDIUM / LOW / INFO). A lower score means denser risk; each finding carries a URL, category, description, and remedy.

```bash
har -f testdata/full.har security
```

Output looks like:

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

### Check categories

`security` splits the audit into 6 categories you can toggle individually:

| Category | What it looks for |
|----------|-------------------|
| Security headers | HSTS, Content-Security-Policy, X-Frame-Options, X-Content-Type-Options, Referrer-Policy — missing or weak |
| Cookie security | Secure / HttpOnly / SameSite attributes, session cookies over cleartext, third-party cookies |
| Mixed content | HTTP sub-resources (scripts, images, iframes) loaded from HTTPS pages |
| Sensitive data | Passwords, tokens, API keys, private keys appearing in responses or form bodies |
| CORS | `Access-Control-Allow-Origin: *` paired with credentials, wildcard reflections, and other dangerous combos |
| Information disclosure | `Server`, `X-Powered-By`, stack traces, version numbers, internal paths |

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--check-headers` | bool | `true` | Check security headers |
| `--check-cookies` | bool | `true` | Check cookie security |
| `--check-mixed-content` | bool | `true` | Check mixed content |
| `--check-sensitive-data` | bool | `true` | Check sensitive-data leakage |
| `--check-cors` | bool | `true` | Check CORS configuration |
| `--check-info-disclosure` | bool | `true` | Check information disclosure |
| `--severity` | string | `low` | Minimum severity filter (`all`/`info`/`low`/`medium`/`high`) |

### Examples

Show only HIGH findings, emit JSON for CI ingestion:

```bash
har -f testdata/full.har security --severity high --format json -o sec-high.json
```

Run only the headers and CORS checks for a quick pass:

```bash
har -f testdata/full.har security --check-headers --check-cors
```

Turn every default off and check only mixed content (handy when migrating to HTTPS):

```bash
har -f testdata/full.har security \
  --check-headers=false --check-cookies=false \
  --check-sensitive-data=false --check-cors=false \
  --check-info-disclosure=false --check-mixed-content
```

`--severity all` surfaces INFO-level findings too — useful for a full baseline:

```bash
har -f testdata/full.har security --severity all
```

### How it works

Under the hood the CLI calls `(*Har).SecurityAudit()`, which returns a `*SecurityReport`. The report holds `Score int` and `Findings []SecurityFinding`; each finding has `Severity`, `Title`, `Category`, `Description`, `Remedy`, and `EntryURL`. The 6 `--check-*` flags map onto the boolean fields of `SecurityAuditOptions`. `--severity` is applied via `report.FindBySeverity(sev)` to filter the output. Text rendering is handled by `formatSecurityReport`; JSON/YAML simply serialize the whole `SecurityReport`.

## redact — Redaction

Replace sensitive values with a placeholder and **write a new HAR file** (unless `--in-place` is given). The original file is left untouched — ideal for safe sharing while preserving structure.

```bash
har -f testdata/full.har redact -o redacted.har
```

### Default redaction targets

`--defaults` (on by default) automatically covers these targets:

| Location | Targets |
|----------|---------|
| Request headers | `Authorization`, `Proxy-Authorization`, `X-Api-Key`, `X-Auth-Token` |
| Cookies | cookies named `session`, `token`, `auth`, `password` |
| Query params | `password`, `token`, `api_key`, `secret`, `access_token` |
| POST fields | `password`, `secret`, `token` |

To cover more, append with `--header` / `--cookie` / `--query-param` / `--post-field`. These add to the default set; they do not replace it.

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--defaults` | bool | `true` | Enable the default redaction rule set |
| `--header` | stringSlice | `[]` | Extra request-header names to redact |
| `--cookie` | stringSlice | `[]` | Extra cookie names to redact |
| `--query-param` | stringSlice | `[]` | Extra query-parameter names to redact |
| `--post-field` | stringSlice | `[]` | Extra POST field names to redact |
| `--replacement` | string | `[REDACTED]` | Replacement placeholder text |
| `--redact-ips` | bool | `false` | Anonymize IP addresses |
| `--in-place` | bool | `false` | Rewrite the input file in place |

### Examples

Beyond the defaults, also scrub a custom header `X-Custom-Key` and a `session_id` cookie:

```bash
har -f testdata/full.har redact \
  --header X-Custom-Key \
  --cookie session_id \
  -o redacted.har
```

Change the placeholder to `***` and anonymize all IPs (good for handing off to an outside party):

```bash
har -f testdata/full.har redact --redact-ips --replacement "***" -o clean.har
```

Rewrite the original file directly (use with care — back up first):

```bash
har -f testdata/full.har redact --in-place
```

Redact only query parameters, disabling the default set:

```bash
har -f testdata/full.har redact \
  --defaults=false \
  --query-param sig --query-param nonce \
  -o params-only.har
```

After redaction, validate to confirm the file is still spec-compliant:

```bash
har -f redacted.har validate --strict
```

::: warning redact is not encryption
`redact` is **structured scrubbing**, not encryption or hashing. Replaced values cannot be recovered, but the request/response structure stays intact so the recipient can still run performance or cache analysis. For highly sensitive material, run `redact`, then `validate`, then spot-check a few entries by hand.
:::

### How it works

The CLI calls `(*Har).Redact(opts)`, which returns a new `*Har` (the original is untouched). `opts` starts from `har.DefaultRedactOptions()`; the CLI merges `--header` and friends into the corresponding slices, writes `--replacement` into `Replacement`, and turns on `RedactIPs` when `--redact-ips` is set. Redaction walks each entry's headers, cookies, queryString, and postData.params, replacing any value whose name matches a target with the placeholder. IP anonymization runs a separate address-rewrite pass. With `--in-place` the result is written back to the input path via `ToJSON(true)`; otherwise it goes to the `-o` file (or stdout if none given).

## Summary

| Command | Purpose | Modifies files? |
|---------|---------|-----------------|
| `security` | Audit risks, assign a score | No — read-only |
| `redact` | Scrub sensitive values | Yes — new file, or `--in-place` |

A typical chain: `security` to find the risks and leaks, `redact` to scrub the values, then `validate` to re-check. See the full [Security Audit workflow](../workflows/security-audit.md).
