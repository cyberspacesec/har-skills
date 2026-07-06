---
title: MCP 封装
---

# MCP 封装

MCP（Model Context Protocol）是让 LLM 客户端调用外部工具的标准协议。本文说明如何用 HAR Skills 的 Go SDK 包一层 MCP server，把 24 个 CLI 命令暴露为 MCP tools，让 Claude Desktop / Claude Code 直接调用，无需 shell。

::: warning 现状
**HAR Skills 项目本身未内置 MCP server。** 本页是接入指南，教你如何用 SDK 自行封装一个。社区若已有第三方封装，欢迎在仓库 issue 留下链接。
:::

## MCP 是什么

MCP 定义了「客户端（LLM 宿主）↔ server（工具提供方）」之间的标准握手与调用协议：

- **Tools**：可被 LLM 调用的函数，带名称、描述、JSON Schema 参数。
- **Resources**：可被读取的数据源（本场景一般不用）。
- **Transports**：stdio（本地进程）或 HTTP/SSE（远程）。

把 `har` 包成 MCP server 后，Claude Desktop 配置一行 `command` 即可让模型直接 `security_audit(file)`、`find_errors(file)`，而不必生成 bash 再解析 stdout。

## 封装思路

本项目天然适合 MCP 化：

1. **SDK 零依赖、单进程**——直接 import 根包，无需起子进程。
2. **24 个命令语义清晰**——一对一映射成 24 个 tool，参数对应 CLI flag。
3. **JSON 输出已是结构化**——tool 直接返回 `[]byte`/`map`，无需额外整形。

架构：

```
Claude Desktop ──stdio──▶ har-mcp server (Go)
                              │
                              ▼
                         har SDK 根包
                              │
                              ▼
                         ParseHarFile / SecurityAudit / FindErrors …
```

## 最小 MCP server 骨架

下面用 [`github.com/mark3labs/mcp-go`](https://github.com/mark3labs/mcp-go) 给出 tool 注册模式。这不是完整可运行程序，只展示「注册 tool → 解析参数 → 调 SDK → 返回结果」的关键骨架。

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/mark3labs/mcp-go/mcp"
    "github.com/mark3labs/mcp-go/server"

    har "github.com/cyberspacesec/har-skills"
)

func main() {
    s := server.NewMCPServer(
        "har-skills",
        "0.1.0",
        server.WithToolCapabilities(true),
    )

    // 1) har_info —— 总览
    s.AddTool(
        mcp.NewTool("har_info",
            mcp.WithDescription("解析 HAR 文件并返回总览统计（条目数、状态码分布、域名、内容类型等）"),
            mcp.WithString("file", mcp.Required(), mcp.Description("HAR 文件路径")),
        ),
        func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
            file, _ := req.Params.Arguments["file"].(string)
            h, err := har.ParseHarFile(file)
            if err != nil {
                return mcp.NewToolResultError(err.Error()), nil
            }
            stats := h.Statistics()
            b, _ := json.Marshal(stats)
            return mcp.NewToolResultText(string(b)), nil
        },
    )

    // 2) har_security_audit —— 安全审计
    s.AddTool(
        mcp.NewTool("har_security_audit",
            mcp.WithDescription("对 HAR 文件做安全审计，返回 0-100 分与按严重度分组的 findings"),
            mcp.WithString("file", mcp.Required(), mcp.Description("HAR 文件路径")),
            mcp.WithString("severity", mcp.Description("只返回该级别及以上：info/low/medium/high，留空返回全部")),
        ),
        func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
            file, _ := req.Params.Arguments["file"].(string)
            h, err := har.ParseHarFile(file)
            if err != nil {
                return mcp.NewToolResultError(err.Error()), nil
            }
            report := h.SecurityAudit()
            b, _ := json.Marshal(report)
            return mcp.NewToolResultText(string(b)), nil
        },
    )

    // 3) har_find_errors —— 抓错误请求
    s.AddTool(
        mcp.NewTool("har_find_errors",
            mcp.WithDescription("返回所有 4xx/5xx 请求的 URL、方法、状态码"),
            mcp.WithString("file", mcp.Required(), mcp.Description("HAR 文件路径")),
        ),
        func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
            file, _ := req.Params.Arguments["file"].(string)
            h, err := har.ParseHarFile(file)
            if err != nil {
                return mcp.NewToolResultError(err.Error()), nil
            }
            entries := h.FindErrors().GetAll()
            type row struct {
                Method string `json:"method"`
                URL    string `json:"url"`
                Status int    `json:"status"`
            }
            out := make([]row, 0, len(entries))
            for _, e := range entries {
                out = append(out, row{e.Request.Method, e.Request.URL, e.Response.Status})
            }
            b, _ := json.Marshal(out)
            return mcp.NewToolResultText(string(b)), nil
        },
    )

    // 4) har_redact —— 脱敏
    s.AddTool(
        mcp.NewTool("har_redact",
            mcp.WithDescription("脱敏敏感数据（Authorization、cookie、token、IP 等）并写出新文件"),
            mcp.WithString("file", mcp.Required(), mcp.Description("输入 HAR 文件路径")),
            mcp.WithString("output", mcp.Required(), mcp.Description("输出 HAR 文件路径")),
            mcp.WithBoolean("redact_ips", mcp.Description("是否匿名化 IP 地址")),
        ),
        func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
            file, _ := req.Params.Arguments["file"].(string)
            output, _ := req.Params.Arguments["output"].(string)
            h, err := har.ParseHarFile(file)
            if err != nil {
                return mcp.NewToolResultError(err.Error()), nil
            }
            opts := har.DefaultRedactOptions()
            redacted := h.Redact(opts)
            data, _ := json.MarshalIndent(redacted, "", "  ")
            // 写 output 略；返回摘要
            return mcp.NewToolResultText(
                fmt.Sprintf(`{"output":%q,"entries":%d}`, output, len(redacted.Log.Entries)),
            ), nil
            _ = data
        },
    )

    // …其余 20 个命令同理：har_find_slow / har_cache / har_cookie /
    //   har_performance / har_waterfall / har_diff / har_merge / har_split /
    //   har_validate / har_transform / har_export / har_dedup / har_replay …

    server.ServeStdio(s)
}
```

::: tip 注册模式
每个 tool 重复「`NewTool` 描述 + `WithXxx` 参数 → 闭包里 `ParseHarFile` → 调对应 SDK 方法 → JSON 返回」这一模板。可抽一个 `registerHarTool(s, name, desc, params, handler)` 辅助函数消除样板，但上面的展开形式更易读懂。
:::

## 客户端配置

编译后，在 Claude Desktop 的 `claude_desktop_config.json` 里加一行：

```json
{
  "mcpServers": {
    "har-skills": {
      "command": "/path/to/har-mcp",
      "env": {}
    }
  }
}
```

重启 Claude Desktop，模型即可在对话中调用 `har_info`、`har_security_audit` 等工具分析你本地的 HAR 文件——全程不碰 shell。

## 适合场景

- **Claude Desktop / Claude Code 直连**：模型自然语言提问「这份抓包安全吗」，自动选 tool、传文件路径、解读 JSON 结果。
- **无 shell 环境**：受限于策略不能起子进程的宿主，MCP over stdio 是唯一通路。
- **复用 SDK 全部能力**：24 命令 70+ 方法一次封装，长期受益，不必逐个写 shell 适配。

## 与其他接入方式对比

| 接入方式 | 谁调用 | 需要起进程 | 适合 |
|----------|--------|-----------|------|
| [AI Agent Skill](./skill.md) | Agent 自己读文档驱动 CLI | 是（CLI 二进制） | 通用、零封装 |
| [CLI](./cli.md) | 人 / 脚本 | 是 | 交互式、CI |
| [Go SDK](./sdk.md) | 你的 Go 程序 | 否 | 嵌入应用 |
| **MCP 封装** | LLM 客户端直连 | 是（server 进程） | Claude Desktop、无 shell 宿主 |

## 下一步

- [Go SDK 接入](./sdk.md) —— 封装前先熟悉 SDK 入口与方法
- [AI Agent Skill 接入](./skill.md) —— 不想写 server 时，让 Agent 直接驱动 CLI
- [安全审计工作流](../workflows/security-audit.md) —— tool 内部调用的真实工作流
