# HAR Skills

[![Go Reference](https://pkg.go.dev/badge/github.com/cyberspacesec/har-skills.svg)](https://pkg.go.dev/github.com/cyberspacesec/har-skills)
[![Go Report Card](https://goreportcard.com/badge/github.com/cyberspacesec/har-skills)](https://goreportcard.com/report/github.com/cyberspacesec/har-skills)
[![License](https://img.shields.io/github/license/cyberspacesec/har-skills)](https://github.com/cyberspacesec/har-skills/blob/main/LICENSE)
[![Release](https://img.shields.io/github/v/release/cyberspacesec/har-skills)](https://github.com/cyberspacesec/har-skills/releases/latest)
[![CI](https://github.com/cyberspacesec/har-skills/actions/workflows/release.yml/badge.svg)](https://github.com/cyberspacesec/har-skills/actions)

**HAR Skills** is an **AI-native** library for HAR (HTTP Archive) file analysis. It wraps the complete HAR lifecycle — parsing, analysis, security audit, performance scoring, data redaction, request transformation, diff, merge/split, export — into **23 CLI commands** and **70+ SDK methods**.

## Access Methods

HAR Skills can be accessed in **4 ways** — Skills first:

| Method | Best For | Quick Start |
|--------|----------|-------------|
| 🤖 **Skills** | AI agents (Claude, GPT, etc.) | Read [CLAUDE.md](./CLAUDE.md) |
| 📦 **Go SDK** | Go applications | `go get github.com/cyberspacesec/har-skills` |
| 🖥️ **CLI** | Terminal / scripts | `go install github.com/cyberspacesec/har-skills/cmd/har@latest` |
| 🔌 **MCP** | MCP-compatible AI tools | *(coming soon)* |

### 1. 🤖 Skills (AI Agent)

HAR Skills ships with a **progressive-disclosure Skill document** ([CLAUDE.md](./CLAUDE.md)) designed for direct AI agent consumption. AI agents can:

- Download a pre-built binary and use CLI commands
- Clone source and compile
- Call SDK methods programmatically

**One-click Skill prompt** — copy and paste into any AI agent:

```
You have access to the HAR Skills tool for HAR (HTTP Archive) file analysis.

Install: go install github.com/cyberspacesec/har-skills/cmd/har@latest
Or download binary: https://github.com/cyberspacesec/har-skills/releases/latest
Or build from source: git clone https://github.com/cyberspacesec/har-skills.git && cd har-skills && go build -o har ./cmd/har/

Usage: har -f <file> <command>

Commands:
  info              File overview & statistics
  list              List entries with filters
  find <pattern>    Search entries (20+ filter flags)
  security          Security audit (headers, cookies, CORS, mixed content)
  performance       Performance scoring (A/B/C/D grade)
  export <format>   Export to curl/wget/python/postman/xml/yaml/json/csv/markdown/html/jsonl
  redact            Redact sensitive data (passwords, tokens, IPs)
  diff <f1> <f2>    Compare two HAR files
  merge <f1> <f2>   Merge HAR files
  split             Split HAR by domain/page/time/size/status/method
  validate          Validate HAR spec compliance
  replay            Replay HTTP requests
  index             Build & query entry index
  domains           Per-domain statistics
  content           Content type & size analysis
  connections       Connection reuse analysis
  cookie            Cookie security audit
  cache             Cache analysis
  waterfall         Waterfall timeline
  timing            Timing breakdown
  headers           View request/response headers
  extract           Extract response content
  dedup             Find/remove duplicates
  transform         Transform URLs, headers, schemes

Full docs: https://github.com/cyberspacesec/har-skills/blob/main/CLAUDE.md
```

### 2. 📦 Go SDK

```go
package main

import (
    "fmt"
    "log"

    har "github.com/cyberspacesec/har-skills"
)

func main() {
    // Parse HAR file
    h, err := har.ParseHarFile("capture.har")
    if err != nil {
        log.Fatal(err)
    }

    // Statistics
    stats := h.Statistics()
    fmt.Printf("Requests: %d, Avg time: %.1fms\n", stats.TotalRequests, stats.AvgTime)

    // Security audit
    report := h.SecurityAudit()
    fmt.Printf("Security score: %d/100\n", report.Score)

    // Performance scoring
    perf := h.PerformanceScore()
    fmt.Printf("Grade: %s (%.1f/100)\n", perf.Grade(), perf.Score)

    // Data redaction
    redacted := h.Redact(har.DefaultRedactOptions())
    _ = redacted // Safe HAR data
}
```

### 3. 🖥️ CLI

#### Installation

**Pre-built binary (Recommended)**

Download from [GitHub Releases](https://github.com/cyberspacesec/har-skills/releases/latest):

| Platform | Arch | File |
|----------|------|------|
| **Linux** | x86_64 | `har-skills_*_linux_x86_64.tar.gz` |
| **Linux** | arm64 | `har-skills_*_linux_arm64.tar.gz` |
| **Linux** | armv6/v7/i386 | `har-skills_*_linux_*.tar.gz` |
| **macOS** | Intel | `har-skills_*_darwin_x86_64.tar.gz` |
| **macOS** | Apple Silicon | `har-skills_*_darwin_arm64.tar.gz` |
| **Windows** | x86_64/i386 | `har-skills_*_windows_*.zip` |
| **FreeBSD** | x86_64/i386 | `har-skills_*_freebsd_*.tar.gz` |

```bash
# Linux x86_64 example
curl -sL https://github.com/cyberspacesec/har-skills/releases/latest/download/har-skills_0.1.0_linux_x86_64.tar.gz | tar xz
sudo mv har /usr/local/bin/
har --version
```

**Build from source**

```bash
git clone https://github.com/cyberspacesec/har-skills.git
cd har-skills
go build -ldflags "-X github.com/cyberspacesec/har-skills/cmd/har/cmd.version=$(git describe --tags 2>/dev/null || echo dev)" -o har ./cmd/har/
```

**Go Install**

```bash
go install github.com/cyberspacesec/har-skills/cmd/har@latest
```

#### CLI Usage

```bash
har -f capture.har info                              # Overview
har -f capture.har list --limit 20                   # List entries
har -f capture.har find "api/users"                  # Search
har -f capture.har find --errors                     # Error requests
har -f capture.har find --slow 1000                  # Slow requests
har -f capture.har find --response-header "X-Debug"  # By response header
har -f capture.har find --cookie "session_id"        # By cookie name
har -f capture.har security                          # Security audit
har -f capture.har performance                       # Performance score
har -f capture.har redact -o clean.har               # Redact sensitive data
har -f capture.har export curl                       # Export as cURL
har -f capture.har export csv -o data.csv            # Export as CSV
har diff v1.har v2.har                               # Compare files
har merge a.har b.har -o merged.har                  # Merge files
har --help                                           # All commands
```

### 4. 🔌 MCP

MCP (Model Context Protocol) integration is coming soon. It will allow MCP-compatible AI tools to use HAR Skills as a tool server.

## Features

- **23 CLI Commands**: Full HAR lifecycle coverage
- **70+ SDK Methods**: Parsing, analysis, transformation, export
- **Multiple Parse Strategies**: Standard, memory-optimized, lazy-loading, streaming
- **Security Audit**: Header checks, cookie safety, mixed content, CORS, info leakage
- **Performance Scoring**: Lighthouse-style 6-dimension scoring (A/B/C/D grades)
- **Data Redaction**: Auto-strip passwords, tokens, API keys, IP addresses
- **Multi-format Export**: cURL, Wget, Python, Postman, CSV, Markdown, HTML, JSON, YAML, XML, JSONL
- **Progressive Disclosure**: 5-level Skill docs in [CLAUDE.md](./CLAUDE.md)

## Command Reference

| Command | Description | Command | Description |
|---------|-------------|---------|-------------|
| `info` | File overview | `validate` | HAR spec validation |
| `list` | List entries | `redact` | Redact sensitive data |
| `find` | Search entries (20+ filters) | `transform` | Transform requests |
| `headers` | View headers | `export` | 12-format export |
| `timing` | Timing breakdown | `security` | Security audit |
| `extract` | Extract content | `cookie` | Cookie analysis |
| `diff` | Compare files | `cache` | Cache analysis |
| `merge` | Merge files | `performance` | Performance scoring |
| `split` | Split files | `waterfall` | Waterfall timeline |
| `index` | Build & query index | `dedup` | Remove duplicates |
| `domains` | Domain statistics | `replay` | HTTP replay |
| `content` | Content analysis | `connections` | Connection reuse |

## Project Structure

- `pkg/har/` — Go SDK core (40 modules, 741 tests)
- `cmd/har/` — CLI (23 Cobra commands)
- `CLAUDE.md` — AI Agent Skill progressive-disclosure document
- `examples/` — Example code

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

[MIT License](LICENSE)

---

## 简体中文

**HAR Skills** 是一个 **AI 原生** 的 HAR（HTTP Archive）文件分析库，支持 **4 种接入方式**：

| 接入方式 | 适用场景 | 快速开始 |
|----------|----------|----------|
| 🤖 **Skills** | AI Agent（Claude、GPT 等） | 阅读 [CLAUDE.md](./CLAUDE.md) |
| 📦 **Go SDK** | Go 应用程序 | `go get github.com/cyberspacesec/har-skills` |
| 🖥️ **CLI** | 终端 / 脚本 | `go install github.com/cyberspacesec/har-skills/cmd/har@latest` |
| 🔌 **MCP** | MCP 兼容的 AI 工具 | *（即将推出）* |

### 🤖 Skills 接入（AI Agent 一键复制）

将以下提示词复制给 AI Agent，即可获得完整的 HAR 分析能力：

```
你可以使用 HAR Skills 工具来分析 HAR（HTTP Archive）文件。

安装方式：go install github.com/cyberspacesec/har-skills/cmd/har@latest
下载地址：https://github.com/cyberspacesec/har-skills/releases/latest
源码编译：git clone https://github.com/cyberspacesec/har-skills.git && cd har-skills && go build -o har ./cmd/har/

使用：har -f <文件> <命令>

命令：
  info              文件概要和统计
  list              列出条目
  find <pattern>    搜索条目（支持 20+ 过滤参数）
  security          安全审计
  performance       性能评分
  export <format>   导出为 curl/wget/python/postman/xml/yaml/json/csv/markdown/html/jsonl
  redact            数据脱敏
  diff <f1> <f2>    比较两个 HAR 文件
  merge             合并 HAR 文件
  split             拆分 HAR 文件
  validate          验证 HAR 规范
  replay            重放 HTTP 请求
  index             构建索引并查询
  domains           按域名统计
  content           内容类型分析
  connections       连接复用分析
  --help            查看所有命令

完整文档：https://github.com/cyberspacesec/har-skills/blob/main/CLAUDE.md
```

### 📦 Go SDK

```go
import har "github.com/cyberspacesec/har-skills"

h, _ := har.ParseHarFile("capture.har")
stats := h.Statistics()       // 统计信息
report := h.SecurityAudit()   // 安全审计
perf := h.PerformanceScore()  // 性能评分
```

### 🖥️ CLI

```bash
har -f capture.har info          # 概要
har -f capture.har security      # 安全审计
har -f capture.har performance   # 性能评分
har -f capture.har redact -o clean.har  # 数据脱敏
har -f capture.har export curl   # 导出 cURL
har diff v1.har v2.har           # 比较
```
