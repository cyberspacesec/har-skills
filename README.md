# HAR Skills

[![Go Reference](https://pkg.go.dev/badge/github.com/cyberspacesec/har-skills.svg)](https://pkg.go.dev/github.com/cyberspacesec/har-skills)
[![Go Report Card](https://goreportcard.com/badge/github.com/cyberspacesec/har-skills)](https://goreportcard.com/report/github.com/cyberspacesec/har-skills)
[![License](https://img.shields.io/github/license/cyberspacesec/har-skills)](https://github.com/cyberspacesec/har-skills/blob/main/LICENSE)
[![Release](https://img.shields.io/github/v/release/cyberspacesec/har-skills)](https://github.com/cyberspacesec/har-skills/releases/latest)

HAR Skills 是一个面向 AI Agent 的 HAR (HTTP Archive) 全能力 SDK 和命令行工具，用 Go 语言实现。它将 HAR 文件的解析、分析、安全审计、性能评分、数据脱敏、请求转换、差异比较、合并拆分、导出等全部能力封装为 **20 个 CLI 命令** 和 **70+ SDK 方法**，并附带渐进式披露文档，可直接作为 AI Agent 的 Skill 使用。

🤖 **AI Agent 使用**：阅读 [CLAUDE.md](./CLAUDE.md) 获取渐进式披露的完整 Skill 文档。

## 特性

- **20 个 CLI 命令**：info, list, find, headers, timing, extract, diff, merge, split, validate, redact, transform, export, security, cookie, cache, performance, waterfall, dedup, replay
- **70+ SDK 方法**：覆盖 HAR 文件全生命周期操作
- **多种解析策略**：标准、内存优化、懒加载、流式处理
- **安全审计**：头部检查、Cookie安全、混合内容、CORS、信息泄露
- **性能评分**：Lighthouse 风格的 6 维度评分（A/B/C/D 等级）
- **数据脱敏**：自动清除密码、令牌、API 密钥、IP 地址
- **多格式导出**：cURL、Wget、Python requests、Postman、XML、YAML
- **渐进式披露**：5 层级 Skill 文档，AI Agent 可直接消费

## 安装

### 预编译二进制（推荐）

从 [GitHub Releases](https://github.com/cyberspacesec/har-skills/releases/latest) 下载对应平台的二进制文件：

| 平台 | 架构 | 下载 |
|------|------|------|
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
# Linux/macOS 示例
curl -sL https://github.com/cyberspacesec/har-skills/releases/latest/download/har-skills_0.1.0_linux_x86_64.tar.gz | tar xz
sudo mv har /usr/local/bin/

# 验证
har --version
```

### Go Install

```bash
go install github.com/cyberspacesec/har-skills/cmd/har@latest
```

### Go Module

```bash
go get github.com/cyberspacesec/har-skills
```

## 快速开始

### CLI 使用

```bash
# 查看概要
har -f capture.har info

# 列出请求
har -f capture.har list --limit 20

# 搜索请求
har -f capture.har find "api/users"
har -f capture.har find --errors          # 所有错误请求
har -f capture.har find --slow 1000       # 慢于1秒的请求

# 安全审计
har -f capture.har security

# 性能评分
har -f capture.har performance

# 数据脱敏
har -f capture.har redact -o clean.har

# 导出为 cURL 命令
har -f capture.har export curl

# 比较两个 HAR 文件
har diff v1.har v2.har

# 查看所有命令
har --help
```

### SDK 使用

```go
package main

import (
    "fmt"
    "log"

    har "github.com/cyberspacesec/har-skills"
)

func main() {
    // 解析 HAR 文件
    h, err := har.ParseHarFile("capture.har")
    if err != nil {
        log.Fatal(err)
    }

    // 统计信息
    stats := h.Statistics()
    fmt.Printf("请求数: %d, 平均时间: %.1fms\n", stats.TotalRequests, stats.AvgTime)

    // 安全审计
    report := h.SecurityAudit()
    fmt.Printf("安全评分: %d/100\n", report.Score)

    // 性能评分
    perf := h.PerformanceScore()
    fmt.Printf("性能等级: %s (%.1f/100)\n", perf.Grade(), perf.Score)

    // 数据脱敏
    redacted := h.Redact(har.DefaultRedactOptions())
    _ = redacted // 安全的 HAR 数据
}
```

## 命令一览

| 命令 | 用途 | 命令 | 用途 |
|------|------|------|------|
| `info` | 文件概要 | `validate` | 规范验证 |
| `list` | 列出条目 | `redact` | 数据脱敏 |
| `find` | 搜索条目 | `transform` | 请求转换 |
| `headers` | 查看头部 | `export` | 格式导出 |
| `timing` | 计时分析 | `security` | 安全审计 |
| `extract` | 提取内容 | `cookie` | Cookie分析 |
| `diff` | 文件比较 | `cache` | 缓存分析 |
| `merge` | 文件合并 | `performance` | 性能评分 |
| `split` | 文件拆分 | `waterfall` | 瀑布流 |
| | | `dedup` | 去重 |
| | | `replay` | HTTP重放 |

## 项目结构

- `pkg/har/` — SDK 核心代码（40 模块，741 测试）
- `cmd/har/` — CLI 命令（20 个 Cobra 命令）
- `CLAUDE.md` — AI Agent Skill 渐进式披露文档
- `examples/` — 示例代码
- `doc/` — 详细文档

## 贡献

欢迎贡献！请查看 [贡献指南](CONTRIBUTING.md) 了解如何参与项目开发。

## 许可证

本项目使用 [MIT 许可证](LICENSE)。
