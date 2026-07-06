---
title: MCP Wrapper
---

# MCP Wrapper

MCP (Model Context Protocol) is the standard protocol that lets LLM clients invoke external tools. This page shows how to wrap HAR Skills' Go SDK into an MCP server that exposes the 24 CLI commands as MCP tools, so Claude Desktop / Claude Code can call them directly — no shell required.

::: warning Current status
**HAR Skills does not ship a built-in MCP server.** This page is an integration guide showing how to wrap one yourself with the SDK. If a third-party wrapper already exists, please drop a link in a repo issue.
:::

## What MCP is

MCP defines the standard handshake and call protocol between a client (the LLM host) and a server (the tool provider):

- **Tools**: functions the LLM can call, each with a name, description, and JSON-Schema parameters.
- **Resources**: readable data sources (not used in this scenario).
- **Transports**: stdio (local process) or HTTP/SSE (remote).

Once `har` is wrapped as an MCP server, Claude Desktop needs a single `command` line for the model to call `security_audit(file)`, `find_errors(file)` directly — instead of generating bash and parsing stdout.

## Wrapping approach

The project is a natural fit for MCP-ification:

1. **Zero-dep SDK, single process** — import the root package directly, no subprocess.
2. **24 commands with clear semantics** — map 1:1 to 24 tools, params mirroring CLI flags.
3. **JSON output is already structured** — a tool returns `[]byte` / `map` with no reshaping.

Architecture:

```
Claude Desktop ──stdio──▶ har-mcp server (Go)
                              │
                              ▼
                         har SDK root package
                              │
                              ▼
                         ParseHarFile / SecurityAudit / FindErrors …
```

## Minimal MCP server skeleton

Below is the tool-registration pattern using [`github.com/mark3labs/mcp-go`](https://github.com/mark3labs/mcp-go). This is not a fully runnable program — it shows the key skeleton of "register tool → parse args → call SDK → return result".

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

    // 1) har_info — overview
    s.AddTool(
        mcp.NewTool("har_info",
            mcp.WithDescription("Parse a HAR file and return overview stats (entry count, status distribution, domains, content types)"),
            mcp.WithString("file", mcp.Required(), mcp.Description("Path to the HAR file")),
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

    // 2) har_security_audit — security audit
    s.AddTool(
        mcp.NewTool("har_security_audit",
            mcp.WithDescription("Run a security audit on a HAR file; returns a 0-100 score and findings grouped by severity"),
            mcp.WithString("file", mcp.Required(), mcp.Description("Path to the HAR file")),
            mcp.WithString("severity", mcp.Description("Return only this level and above: info/low/medium/high; empty = all")),
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

    // 3) har_find_errors — surface error requests
    s.AddTool(
        mcp.NewTool("har_find_errors",
            mcp.WithDescription("Return URL, method, and status for every 4xx/5xx request"),
            mcp.WithString("file", mcp.Required(), mcp.Description("Path to the HAR file")),
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

    // 4) har_redact — redact sensitive data
    s.AddTool(
        mcp.NewTool("har_redact",
            mcp.WithDescription("Redact sensitive data (Authorization, cookies, tokens, IPs) and write a new file"),
            mcp.WithString("file", mcp.Required(), mcp.Description("Input HAR file path")),
            mcp.WithString("output", mcp.Required(), mcp.Description("Output HAR file path")),
            mcp.WithBoolean("redact_ips", mcp.Description("Whether to anonymize IP addresses")),
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
            // writing `output` omitted; return a summary
            return mcp.NewToolResultText(
                fmt.Sprintf(`{"output":%q,"entries":%d}`, output, len(redacted.Log.Entries)),
            ), nil
        },
    )

    // …the other 20 commands follow the same pattern:
    //   har_find_slow / har_cache / har_cookie / har_performance /
    //   har_waterfall / har_diff / har_merge / har_split / har_validate /
    //   har_transform / har_export / har_dedup / har_replay …

    server.ServeStdio(s)
}
```

::: tip Registration pattern
Every tool repeats the same template: `NewTool` with a description + `WithXxx` params → in the closure, `ParseHarFile` → call the matching SDK method → return JSON. You can extract a `registerHarTool(s, name, desc, params, handler)` helper to kill the boilerplate, but the expanded form above is easier to read.
:::

## Client configuration

After building, add one entry to Claude Desktop's `claude_desktop_config.json`:

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

Restart Claude Desktop and the model can call `har_info`, `har_security_audit`, etc. to analyze local HAR files in conversation — never touching a shell.

## Suitable scenarios

- **Claude Desktop / Claude Code direct connection**: the model answers "is this capture secure?" in natural language, auto-selecting a tool, passing the file path, and interpreting the JSON result.
- **No-shell environments**: for hosts whose policy forbids subprocesses, MCP over stdio is the only path.
- **Reuse the full SDK**: wrap 24 commands / 70+ methods once and benefit long-term — no per-command shell adapters.

## Comparison with other access methods

| Access method | Caller | Needs a process | Good for |
|---------------|--------|-----------------|----------|
| [AI Agent Skill](./skill.md) | Agent reads docs and drives the CLI | yes (CLI binary) | general purpose, zero wrapping |
| [CLI](./cli.md) | humans / scripts | yes | interactive, CI |
| [Go SDK](./sdk.md) | your Go program | no | embedding in an app |
| **MCP wrapper** | LLM client direct | yes (server process) | Claude Desktop, no-shell hosts |

## Next steps

- [Go SDK access](./sdk.md) — get familiar with SDK entry points before wrapping
- [AI Agent Skill access](./skill.md) — skip the server and let an Agent drive the CLI
- [Security audit workflow](../workflows/security-audit.md) — the real workflow the tools call internally
