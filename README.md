# HAR Skills

[![Go Reference](https://pkg.go.dev/badge/github.com/cyberspacesec/har-skills.svg)](https://pkg.go.dev/github.com/cyberspacesec/har-skills)
[![Go Report Card](https://goreportcard.com/badge/github.com/cyberspacesec/har-skills)](https://goreportcard.com/report/github.com/cyberspacesec/har-skills)
[![License](https://img.shields.io/github/license/cyberspacesec/har-skills)](https://github.com/cyberspacesec/har-skills/blob/main/LICENSE)
[![Release](https://img.shields.io/github/v/release/cyberspacesec/har-skills)](https://github.com/cyberspacesec/har-skills/releases/latest)
[![CI](https://github.com/cyberspacesec/har-skills/actions/workflows/release.yml/badge.svg)](https://github.com/cyberspacesec/har-skills/actions)

**HAR Skills** is an **AI-native** Go SDK and CLI for HAR (HTTP Archive) files. It wraps the complete HAR lifecycle — parsing, analysis, security audit, performance scoring, data redaction, request transformation, diff, merge/split, export — into **23 CLI commands** and **70+ SDK methods**, with progressive-disclosure documentation designed for direct AI agent consumption.

🤖 **AI Agents**: Read [CLAUDE.md](./CLAUDE.md) for the full progressive-disclosure Skill document.

---

## Features

- **23 CLI Commands**: info, list, find, headers, timing, extract, diff, merge, split, validate, redact, transform, export, security, cookie, cache, performance, waterfall, dedup, replay, index, domains, content, connections
- **70+ SDK Methods**: Full HAR lifecycle coverage
- **Multiple Parse Strategies**: Standard, memory-optimized, lazy-loading, streaming
- **Security Audit**: Header checks, cookie safety, mixed content, CORS, info leakage
- **Performance Scoring**: Lighthouse-style 6-dimension scoring (A/B/C/D grades)
- **Data Redaction**: Auto-strip passwords, tokens, API keys, IP addresses
- **Multi-format Export**: cURL, Wget, Python requests, Postman Collection, XML, YAML
- **Progressive Disclosure**: 5-level Skill docs consumable by AI agents

## Installation

### Pre-built Binaries (Recommended)

Download from [GitHub Releases](https://github.com/cyberspacesec/har-skills/releases/latest):

| Platform | Arch | File |
|----------|------|------|
| **Linux** | x86_64 | `har-skills_*_linux_x86_64.tar.gz` |
| **Linux** | arm64 | `har-skills_*_linux_arm64.tar.gz` |
| **Linux** | armv6 | `har-skills_*_linux_armv6.tar.gz` |
| **Linux** | armv7 | `har-skills_*_linux_armv7.tar.gz` |
| **Linux** | i386 | `har-skills_*_linux_i386.tar.gz` |
| **macOS** | Intel | `har-skills_*_darwin_x86_64.tar.gz` |
| **macOS** | Apple Silicon | `har-skills_*_darwin_arm64.tar.gz` |
| **Windows** | x86_64 | `har-skills_*_windows_x86_64.zip` |
| **Windows** | i386 | `har-skills_*_windows_i386.zip` |
| **FreeBSD** | x86_64 | `har-skills_*_freebsd_x86_64.tar.gz` |
| **FreeBSD** | i386 | `har-skills_*_freebsd_i386.tar.gz` |

```bash
# Linux/macOS example
curl -sL https://github.com/cyberspacesec/har-skills/releases/latest/download/har-skills_0.1.0_linux_x86_64.tar.gz | tar xz
sudo mv har /usr/local/bin/

# Verify
har --version
```

### Build from Source

```bash
# Clone
git clone https://github.com/cyberspacesec/har-skills.git
cd har-skills

# Build
go build -o har ./cmd/har/

# Install globally
go install ./cmd/har/

# Or with version info
go build -ldflags "-X github.com/cyberspacesec/har-skills/cmd/har/cmd.version=$(git describe --tags)" -o har ./cmd/har/
```

### Go Install

```bash
go install github.com/cyberspacesec/har-skills/cmd/har@latest
```

### Go Module (SDK Only)

```bash
go get github.com/cyberspacesec/har-skills
```

## Quick Start

### CLI

```bash
# Overview
har -f capture.har info

# List entries
har -f capture.har list --limit 20

# Search entries
har -f capture.har find "api/users"
har -f capture.har find --errors          # All error requests
har -f capture.har find --slow 1000       # Slower than 1s

# Security audit
har -f capture.har security

# Performance scoring
har -f capture.har performance

# Data redaction
har -f capture.har redact -o clean.har

# Export to cURL
har -f capture.har export curl

# Diff two HAR files
har diff v1.har v2.har

# All commands
har --help
```

### SDK

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

## AI Integration

HAR Skills is designed as an **AI-native** library. AI agents can use it in two ways:

### Option 1: Download Pre-built Binary

```
Download the latest release binary for the current platform from:
https://github.com/cyberspacesec/har-skills/releases/latest

Then use the CLI commands documented in CLAUDE.md to analyze HAR files.
```

### Option 2: Build from Source

```
1. git clone https://github.com/cyberspacesec/har-skills.git
2. cd har-skills && go build -o har ./cmd/har/
3. Use the har binary as documented in CLAUDE.md
```

### Option 3: Use as Go SDK

```go
import har "github.com/cyberspacesec/har-skills"
```

### One-Click AI Agent Prompt

Copy this prompt to give an AI agent full HAR analysis capability:

```
You have access to the HAR Skills CLI tool for analyzing HAR (HTTP Archive) files.

Install: go install github.com/cyberspacesec/har-skills/cmd/har@latest
Or download: https://github.com/cyberspacesec/har-skills/releases/latest

Commands: har -f <file> <command>
  info          — File overview
  list          — List entries
  find <query>  — Search entries (supports 20+ filter flags)
  security      — Security audit
  performance   — Performance scoring
  export <fmt>  — Export (curl/wget/python/postman/xml/yaml/json/csv/markdown/html)
  redact        — Redact sensitive data
  diff <f1> <f2> — Compare files
  merge <f1> <f2> — Merge files
  validate      — Validate HAR spec
  replay        — Replay HTTP requests
  index         — Build & query entry index
  domains       — Per-domain statistics
  content       — Content type analysis
  connections   — Connection reuse analysis
  --help        — All commands & flags

Skill docs: https://github.com/cyberspacesec/har-skills/blob/main/CLAUDE.md
```

## Command Reference

| Command | Description | Command | Description |
|---------|-------------|---------|-------------|
| `info` | File overview | `validate` | HAR spec validation |
| `list` | List entries | `redact` | Redact sensitive data |
| `find` | Search entries | `transform` | Transform requests |
| `headers` | View headers | `export` | Multi-format export |
| `timing` | Timing analysis | `security` | Security audit |
| `extract` | Extract content | `cookie` | Cookie analysis |
| `diff` | Compare files | `cache` | Cache analysis |
| `merge` | Merge files | `performance` | Performance scoring |
| `split` | Split files | `waterfall` | Waterfall view |
| `index` | Build & query index | `dedup` | Remove duplicates |
| `domains` | Domain statistics | `replay` | HTTP replay |
| `content` | Content analysis | `connections` | Connection reuse |

## Project Structure

- `pkg/har/` — SDK core (40 modules, 741 tests)
- `cmd/har/` — CLI commands (20 Cobra commands)
- `CLAUDE.md` — AI Agent Skill progressive-disclosure docs
- `examples/` — Example code
- `doc/` — Detailed documentation

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

[MIT License](LICENSE)

---

## 简体中文

**HAR Skills** 是一个 **AI 原生** 的 Go SDK 和命令行工具，用于 HAR（HTTP Archive）文件分析。它将 HAR 文件的解析、分析、安全审计、性能评分、数据脱敏、请求转换、差异比较、合并拆分、导出等全部能力封装为 **23 个 CLI 命令** 和 **70+ SDK 方法**，并附带渐进式披露文档，可直接作为 AI Agent 的 Skill 使用。

🤖 **AI Agent 接入**：阅读 [CLAUDE.md](./CLAUDE.md) 获取渐进式披露的完整 Skill 文档。

### 安装

```bash
# 从 Release 下载（推荐）
# https://github.com/cyberspacesec/har-skills/releases/latest

# 从源码编译
git clone https://github.com/cyberspacesec/har-skills.git
cd har-skills && go build -o har ./cmd/har/

# Go Install
go install github.com/cyberspacesec/har-skills/cmd/har@latest

# Go Module（仅 SDK）
go get github.com/cyberspacesec/har-skills
```

### CLI 使用

```bash
har -f capture.har info          # 概要
har -f capture.har security      # 安全审计
har -f capture.har performance   # 性能评分
har -f capture.har redact -o clean.har  # 数据脱敏
har -f capture.har export curl   # 导出 cURL
har diff v1.har v2.har           # 比较
```

### SDK 使用

```go
import har "github.com/cyberspacesec/har-skills"

h, _ := har.ParseHarFile("capture.har")
stats := h.Statistics()       // 统计
report := h.SecurityAudit()   // 安全审计
perf := h.PerformanceScore()  // 性能评分
```

### AI 一键接入

复制以下提示词给 AI Agent，即可获得完整的 HAR 分析能力：

```
你可以使用 HAR Skills CLI 工具来分析 HAR 文件。

安装方式：go install github.com/cyberspacesec/har-skills/cmd/har@latest
下载地址：https://github.com/cyberspacesec/har-skills/releases/latest

使用：har -f <文件> <命令>
  info       — 文件概要
  security   — 安全审计
  performance — 性能评分
  export     — 多格式导出
  redact     — 数据脱敏
  diff       — 文件比较
  --help     — 查看所有命令

Skill 文档：https://github.com/cyberspacesec/har-skills/blob/main/CLAUDE.md
```
