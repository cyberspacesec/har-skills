---
title: AI Agent Skill
---

# AI Agent Skill Access

HAR Skills is designed from day one as an **AI Agent Skill**: the `CLAUDE.md` at the repository root is a progressive-disclosure document that an AI Agent reads to understand every capability and invoke it. This page explains how to onboard an LLM Agent (e.g. Claude Code) quickly.

## What Skill access means

Skill access = hand an "LLM-readable manual" to an Agent so it can autonomously complete HAR analysis tasks.

- **The artifact**: [`CLAUDE.md`](https://github.com/cyberspacesec/har-skills/blob/main/CLAUDE.md) at the repo root, ~400 lines, covering all 24 CLI commands and 70+ SDK methods.
- **Structure**: progressive disclosure — from Quick Start to a 5-level command reference, then an SDK cheat sheet and workflows. The Agent drills down on demand instead of reading everything at once.
- **Outcome**: after reading, the Agent knows which commands exist, their flags, output formats, and typical combinations — no trial and error needed.

::: tip Why not just link the docs site
The docs site is for humans, full of layout and navigation. `CLAUDE.md` is plain text, structured, and token-budget-optimized — an Agent loads it once and gets the full capability map. The two share source content, but the Skill doc is denser.
:::

## Workflow

```
Agent reads CLAUDE.md
        │
        ▼
Install the har binary (go install or download release)
        │
        ▼
Run CLI commands to do the task (info / find / security / redact …)
        │
        ▼
Emit JSON so the Agent can parse and reason further
```

Each step is autonomous: reading the doc is context loading, install is one shell call, analysis is command composition, and the JSON result feeds the next reasoning turn.

### Step by step

1. **Read `CLAUDE.md`** — the Agent loads the progressive-disclosure doc into context and gets a capability map of "24 commands + 70+ SDK methods + 4 workflows." The doc is deliberately ordered by usage frequency, so the Agent can jump to the section it needs.
2. **Install the binary** — the Agent picks `go install` (needs a Go toolchain) or a prebuilt release (zero deps). The latter is just `curl + tar`, ideal for runtime-only environments.
3. **Run commands** — the Agent picks by goal: `security` for audit, `performance` for speed, `diff` for comparison. Every command takes `-f` for the file and `--format json` for structured output.
4. **Parse JSON** — the Agent folds the JSON output into the next reasoning turn and decides whether to chain another command (e.g. `redact` then `validate`). No hand-written `jq` — the model reads JSON natively.

### Capability map at a glance

`CLAUDE.md` gives the Agent four rough capability buckets; the Agent maps the task to the right one:

| Task family | Key commands | Typical output |
|-------------|--------------|----------------|
| Overview & locate | `info` `list` `find` | entry count, status distribution, error/slow-request lists |
| Security & redact | `security` `redact` `cookie` | security score, findings, a redacted new HAR |
| Performance & cache | `performance` `timing` `cache` `waterfall` | grade, critical path, cacheability |
| Transform & export | `transform` `export` `dedup` `replay` | curl/wget/Postman, deduped file, replay results |

## One-Prompt Agent Bootstrap

Send the prompt below to an Agent and it will install the binary and start analyzing HAR files:

```text
You are a HAR file analysis assistant using the har-skills tool.

Step 1 — install the CLI (pick one):
  go install github.com/cyberspacesec/har-skills/cmd/har@latest
or download the tar.gz for your platform from
  https://github.com/cyberspacesec/har-skills/releases/latest
extract it and put `har` on your PATH.

Step 2 — read CLAUDE.md at the repo root (or online) to learn the 24 CLI
commands, global flags (-f/--file, --format, -o) and typical workflows.
Prefer the CLI for HAR tasks and use --format json for machine-readable output.

Step 3 — analyze the user's HAR file:
  har -f <file> info --format json          # overview
  har -f <file> find --errors --format json # surface errors
  har -f <file> security --format json      # security audit
  har -f <file> performance --format json   # performance score

Conventions:
- Redact before sharing: har -f <file> redact --redact-ips -o redacted.har
- Split large files first: har -f <file> split --by domain -o by-domain
- Always output JSON for your own parsing; switch to text/table only for humans.

When given a HAR file, run info first, then choose follow-up commands by goal.
```

::: tip Trim it
If your Agent already works inside the repo, just point it at the local `CLAUDE.md` — no online copy needed. Prune the command list to your scenario (e.g. keep only security + redact for audit-only agents).
:::

## Why it fits Agents

| Property | Value to an Agent |
|----------|-------------------|
| CLI emits `--format json` | Machine-readable; Agent parses and reasons over it directly |
| Composable commands | `find --slow 1000` → `extract` → `export curl` is one pipeline |
| Skill doc is LLM-first | One load yields the full capability map — fewer tokens, fewer round-trips |
| 24 commands span the lifecycle | Parse, filter, security, performance, transform, export, replay — closed loop |
| Zero runtime deps (SDK side) | Single binary, frictionless deploy |

## Suitable tasks

- **Security audit**: `security` reports → `redact` scrubs → `validate --strict` checks. Fits "does this capture leak anything" requests.
- **Performance analysis**: `performance` scores → `find --slow` isolates slow requests → `waterfall --critical-path` shows the critical path. Fits "why is this page slow" requests.
- **API behavior diff**: `diff a.har b.har --include-body` compares two captures — useful for regression tests and API migration.
- **Redact-then-share**: `dedup --remove` → `redact --redact-ips` → `split --by domain`. Safely hand captures to a third party.

## Skill doc vs a traditional README

Many projects have a README, but not every README suits an Agent. HAR Skills' `CLAUDE.md` is tuned for LLMs:

| Dimension | Traditional README | `CLAUDE.md` (Skill doc) |
|-----------|--------------------|---------------------------|
| Reader | human developer | an LLM Agent |
| Structure | marketing + install + blurb | progressive capability disclosure, ordered by usage frequency |
| Command coverage | a few popular ones | all 24 commands with flags and output formats |
| Output format | mostly text examples | explicitly marks `--format json` as machine-readable |
| Workflows | occasional | 4 end-to-end workflows ready to copy |

That is why an Agent can finish a task autonomously after reading `CLAUDE.md` — instead of trial-and-error or asking the user "what can this tool do?"

## Next steps

- [CLI access](./cli.md) — the 24 commands by level
- [Go SDK access](./sdk.md) — embed in a Go program
- [MCP wrapper](./mcp.md) — let Claude Desktop call har directly
- [Security audit workflow](../workflows/security-audit.md) — end-to-end example
