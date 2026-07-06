---
title: CLI Access
---

# CLI Access

`har` is a single-binary CLI that exposes the entire HAR Skills SDK to the terminal. Zero runtime dependencies, JSON-first, pipe-friendly — suitable for scripts, CI, and interactive triage.

## CLI overview

- **24 Cobra subcommands** spanning the full HAR analysis lifecycle.
- **Uniform input**: `-f/--file` points at the HAR file; `-` reads from stdin.
- **Uniform output**: `--format` selects `text` (default) / `json` / `csv` / `yaml`; `-o` writes to a file.
- **Env vars**: `HAR_FILE`, `HAR_FORMAT`, `HAR_OUTPUT` (read via Viper).

Install:

```bash
go install github.com/cyberspacesec/har-skills/cmd/har@latest
# or download a prebuilt binary from https://github.com/cyberspacesec/har-skills/releases/latest
```

See [Global Flags](../cli/global-flags.md) for the full flag reference.

## Command levels

`har` groups its 24 commands into 5 levels, from most-used to most-advanced:

| Level | Theme | Commands | See |
|-------|-------|----------|-----|
| Level 1 | Basic operations | `info` `list` `find` `headers` `timing` `extract` | [Basic Operations](../cli/basic.md) |
| Level 2 | File operations | `diff` `merge` `split` `validate` | [File Operations](../cli/files.md) |
| Level 3 | Security & privacy | `security` `redact` | [Security & Privacy](../cli/security.md) |
| Level 4 | Deep analysis | `performance` `cookie` `cache` `index` `domains` `content` `connections` `waterfall` | [Deep Analysis](../cli/analysis.md) |
| Level 5 | Transform & export | `transform` `export` `dedup` `replay` | [Transform & Export](../cli/transform.md) |

::: tip Levels aren't a difficulty gate
A level only reflects "usage frequency and conceptual depth." A Level 1 `find --errors` solves big problems too; a Level 5 `replay` fits in a one-liner. Pick the command that matches the task.
:::

## Typical pipeline usage

The CLI's `--format json` output is a natural fit for `jq` post-processing:

```bash
# Find all 4xx/5xx, project URL and status
har -f capture.har find --errors --format json \
  | jq '.entries[] | {url: .request.url, status: .response.status}'

# Requests per domain, top 10 by count
har -f capture.har domains --format json \
  | jq 'to_entries | sort_by(-.value.count) | .[0:10]'

# Security audit, HIGH only
har -f capture.har security --format json \
  | jq '.findings[] | select(.severity=="high")'
```

One pipeline does "filter → project → sort" with no intermediate files.

## stdin pipeline

`-f -` (or omitting `-f` with stdin) reads the HAR from a pipe:

```bash
cat capture.har | har info
curl -sL https://example.com/capture.har | har -f - find --slow 1000
zcat capture.har.gz | har -f - info --format json
```

::: tip Auto-decompress
`ParseHarFileAuto` detects gzip by extension and magic bytes; the CLI transparently supports `.har.gz` the same way.
:::

## Script integration

### bash batch loop

```bash
#!/usr/bin/env bash
set -euo pipefail

for f in captures/*.har; do
  echo "=== $f ==="
  har -f "$f" security --severity high --format json \
    | jq '.score, (.findings | length)'
done
```

### cron scheduled audit

```bash
# Audit today's capture at 02:00; alert if score < 60
0 2 * * * har -f /data/$(date +\%F).har security --format json \
  | jq -e '.score >= 60' >/dev/null \
  || /usr/local/bin/notify-slack "HAR security alert: $(date +\%F)"
```

### CI integration

```bash
# Regression: diff staging vs prod captures, fail if they diverge
har diff staging.har prod.har --compare-by-url --ignore-timings -o diff.txt
test ! -s diff.txt || { echo "API behavior diverged"; cat diff.txt; exit 1; }
```

## Output format cheat sheet

| `--format` | Good for | Notes |
|------------|-----------|-------|
| `text` | humans, terminals | default; tables have headers, `--no-header` suppresses them |
| `json` | Agents / jq / programs | full structure, ideal for pipe post-processing |
| `csv` | Excel / spreadsheets | fits table-style commands like `list`, `timing` |
| `yaml` | config / review | readable, good for `info`, `security` |

## Next steps

- [Global Flags](../cli/global-flags.md) — the flags every command shares
- [AI Agent Skill access](./skill.md) — let an Agent drive the CLI
- [Go SDK access](./sdk.md) — embed in a program when the SDK fits better
- [MCP wrapper](./mcp.md) — wrap the CLI as MCP tools
