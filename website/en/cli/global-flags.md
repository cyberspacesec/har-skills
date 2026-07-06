---
title: Global Flags
titleTemplate: false
---

# Global Flags

Every `har` subcommand shares a set of persistent flags registered by the root command and bound through Viper. This page covers their values, the configuration-override precedence, stdin piping, and the internal load/output architecture behind them. A full table of all 24 commands at the end gives a quick map by level.

Every example below runs from the repository root against `testdata/example.har` or `testdata/full.har`.

## Global persistent flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--file` | `-f` | empty | HAR file path; `-` reads from stdin |
| `--format` | none | `text` | Output format: `text` / `json` / `csv` / `yaml` |
| `--output` | `-o` | empty | Output file path (empty writes to stdout) |
| `--no-header` | none | `false` | Suppress the header row in `text`/`csv` output |
| `--config` | none | empty | Config file path; defaults to `$HOME/.har.yaml` |

`--file` is the input source for most commands; multi-file commands like `diff` and `merge` take positional args instead and bypass `-f` (see their own pages). `--format` selects the serializer: `text` calls each command's own formatter (usually a table or sectioned text), `json` uses `MarshalIndent`, `csv` calls a dedicated csv function (falling back to JSON), and `yaml` prefers the data's `ToYAML()` and otherwise falls back to indented JSON.

## Viper integration

The CLI wires flags, environment variables, and a config file together with `spf13/viper`, following Viper's precedence (high to low):

1. **Command-line flag** (highest)
2. **Environment variable**
3. **Config file**
4. **Flag default** (lowest)

Registration calls `viper.BindPFlag` to bind `file` / `format` / `output` / `no-header` to their flags; initialization sets `viper.SetEnvPrefix("HAR")` + `viper.AutomaticEnv()`, making these environment variables effective:

| Environment variable | Flag |
|----------------------|------|
| `HAR_FILE` | `--file` |
| `HAR_FORMAT` | `--format` |
| `HAR_OUTPUT` | `--output` |

::: tip Config file lookup order
When `--config` is not passed, Viper looks for a file named `.har.yaml` in `$HOME` and then the current directory, reading the first hit. Put long-running preferences in `~/.har.yaml`, e.g. `format: json`, to default every command to JSON output.
:::

## stdin support

When `--file` is empty and stdin has data, the loader automatically reads from stdin — handy for piping output from `curl`, `cat`, or upstream tools straight into `har`:

```bash
cat capture.har | har info
```

Passing `-` explicitly also forces stdin, useful for the edge case where stdin looks like a terminal and is misdetected:

```bash
har -f - info < capture.har
```

The stdin path calls `har.ParseHar(data)` (bytes already in memory); the file path calls `har.ParseHarFileAuto(path)` (auto-detects gzip). Both ultimately return a `*har.Har`.

## Internal architecture

Every subcommand follows the same execution path, split into "load → call SDK → dispatch output":

```
internal.LoadHar(cmd, args)          → *har.Har   (gzip auto-detect / stdin fallback)
        │
        ▼
SDK call (Statistics / FilterWith / SecurityAudit / ...)
        │
        ▼
internal.WriteOutput(cmd, data, textFunc, csvFunc)   → stdout / file
```

**Load layer** `internal.LoadHar`: reads `--file`; when empty or `-`, it goes to stdin; when empty and stdin has no data, it errors out. Multi-file commands (`diff`/`merge`) use `internal.LoadHarFromArg(arg)` instead, loading each positional arg independently and reporting errors per file.

**Output layer** `internal.WriteOutput`: picks a serialization branch by `--format` and ultimately lands in `WriteToFileOrStdout(path, data)`. Two stdout/stderr conventions matter:

- **Data goes to stdout** — so it can be piped (`| jq`, `| grep`, `> file`).
- **Progress and notices go to stderr** — e.g. `wrote 1234 bytes to report.json`, so it never pollutes the data stream.

That is why `har -f t.har info --format json | jq '.totalRequests'` cleanly returns the field value, while the `wrote ...` notice stays out of the pipe.

`--no-header` is read by `internal.NoHeader(cmd)` and only affects `text`/`csv` table output (suppressing the `INDEX METHOD STATUS ...` header row); JSON/YAML are unaffected.

## Command overview

The 24 commands are organized by level (Level 1–5); the higher the level, the deeper the capability:

### Basic operations (Level 1)

| Command | Purpose | Level |
|---------|---------|-------|
| `info` | HAR summary statistics | Basic |
| `list` | List entries | Basic |
| `find` | Multi-dimensional search | Basic |
| `headers` | Show request/response headers | Basic |
| `timing` | Timing breakdown | Basic |
| `extract` | Extract response content | Basic |

### File operations (Level 2)

| Command | Purpose | Level |
|---------|---------|-------|
| `diff` | Compare two HAR files | File |
| `merge` | Merge multiple HAR files | File |
| `split` | Split a HAR file | File |
| `validate` | Validate HAR spec compliance | File |

### Security & privacy (Level 3)

| Command | Purpose | Level |
|---------|---------|-------|
| `security` | Security audit | Security |
| `redact` | Redact sensitive data | Security |

### Deep analysis (Level 4)

| Command | Purpose | Level |
|---------|---------|-------|
| `performance` | Performance scoring | Deep analysis |
| `cookie` | Cookie audit | Deep analysis |
| `cache` | Cache analysis | Deep analysis |
| `index` | Index queries | Deep analysis |
| `domains` | Per-domain statistics | Deep analysis |
| `content` | Content-type analysis | Deep analysis |
| `connections` | Connection reuse | Deep analysis |
| `waterfall` | Waterfall & timeline | Deep analysis |

### Transform & export (Level 5)

| Command | Purpose | Level |
|---------|---------|-------|
| `transform` | Transform requests | Transform & export |
| `export` | Export to other formats | Transform & export |
| `dedup` | Find/remove duplicates | Transform & export |
| `replay` | Replay HTTP requests | Transform & export |

Per-command detail is organized by level: basic operations in [Basic Operations](./basic.md), file operations in [File Operations](./files.md), security & privacy in [Security & Privacy](./security.md), deep analysis in [Deep Analysis](./analysis.md), and transform & export in [Transform & Export](./transform.md).
