---
title: Installation
titleTemplate: false
---

# Installation

HAR Skills ships as a single `har` binary with no runtime dependencies. Below are three install methods, the platform matrix, verification, upgrade/uninstall, and environment variables.

## Method 1: Pre-built binary (recommended)

Simplest path: download the archive for your platform from GitHub Releases, extract, and put it on your `PATH`.

::: code-group

```bash [Linux x86_64]
curl -sL https://github.com/cyberspacesec/har-skills/releases/latest/download/har-skills_0.1.0_linux_x86_64.tar.gz | tar xz
sudo mv har /usr/local/bin/
```

```bash [macOS Apple Silicon]
curl -sL https://github.com/cyberspacesec/har-skills/releases/latest/download/har-skills_0.1.0_darwin_arm64.tar.gz | tar xz
sudo mv har /usr/local/bin/
```

```powershell [Windows]
# Download har-skills_0.1.0_windows_x86_64.zip
# Extract har.exe and add it to PATH
```

:::

::: tip First run on macOS
A binary downloaded from the web may be blocked by Gatekeeper on first run. Open "System Settings → Privacy & Security" and click "Open Anyway", or run:

```bash
xattr -d com.apple.quarantine /usr/local/bin/har
```
:::

### Platform matrix

Pre-built binaries cover these platforms:

| OS | Architectures |
| --- | --- |
| linux | x86_64 / arm64 / armv6 / armv7 / i386 |
| darwin (macOS) | x86_64 / arm64 |
| windows | x86_64 / i386 |
| freebsd | x86_64 / i386 |

Download filenames follow `har-skills_<version>_<os>_<arch>.tar.gz` (`.zip` for Windows). Pick yours on the [Releases page](https://github.com/cyberspacesec/har-skills/releases/latest).

## Method 2: go install

If you have Go 1.19+, one line does it:

```bash
go install github.com/cyberspacesec/har-skills/cmd/har@latest
```

The binary lands in `$GOPATH/bin` (or `$GOBIN`) — make sure that's on your `PATH`:

```bash
# Confirm GOPATH/bin
go env GOPATH
export PATH="$(go env GOPATH)/bin:$PATH"
```

::: warning Pin the version
For production, replace `@latest` with a concrete version (e.g. `@v0.1.0`) to avoid unexpected behavioral changes.
:::

## Method 3: Build from source

Useful when you need custom build tags, version injection, or cross-compilation.

```bash
git clone https://github.com/cyberspacesec/har-skills.git
cd har-skills
go build -o har ./cmd/har/
```

### Inject version info

Use `-ldflags` to stamp the git tag into the binary so `har --version` shows it:

```bash
go build -ldflags "-X github.com/cyberspacesec/har-skills/cmd/har/cmd.version=$(git describe --tags 2>/dev/null || echo dev)" -o har ./cmd/har/
```

### Cross-compilation

Set `GOOS` / `GOARCH` to produce binaries for other platforms:

```bash
# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o har-darwin-arm64 ./cmd/har/

# Windows
GOOS=windows GOARCH=amd64 go build -o har.exe ./cmd/har/

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o har-linux-arm64 ./cmd/har/
```

## Verify the install

Any command that runs confirms success:

```bash
har --version
```

Or run `info` on a real HAR:

```bash
har -f testdata/example.har info
```

## Upgrade and uninstall

### Upgrade

| Install method | Upgrade command |
| --- | --- |
| Pre-built binary | Download the new archive and overwrite `har` |
| go install | `go install github.com/cyberspacesec/har-skills/cmd/har@latest` |
| Source build | `git pull && go build -o har ./cmd/har/` |

### Uninstall

Just delete the binary — HAR Skills leaves no config or data directories on the system:

```bash
rm -f $(which har)
# or manually remove /usr/local/bin/har, $GOPATH/bin/har, etc.
```

Optionally clear the environment variables (see below).

## Environment variables

The CLI reads three environment variables via Viper, each mirroring a CLI flag — handy for pinning behavior in CI/containers:

| Variable | Equivalent flag | Purpose |
| --- | --- | --- |
| `HAR_FILE` | `-f` / `--file` | Default input file path |
| `HAR_FORMAT` | `--format` | Default output format (text/json/csv/yaml) |
| `HAR_OUTPUT` | `-o` / `--output` | Default output file path |

Example:

```bash
export HAR_FILE=capture.har
export HAR_FORMAT=json
har info          # equivalent to: har -f capture.har info --format json
```

::: tip Precedence
CLI flags take precedence over environment variables; environment variables take precedence over defaults. An explicit flag overrides the env var.
:::

A config file is also supported via `--config` (default `$HOME/.har.yaml`) to persist common parameters.

## Next steps

- Run your first command: [Quick Start](./quick-start.md)
- Understand the input format: [HAR Format Primer](./har-basics.md)
- All commands: [CLI Reference](./cli/global-flags.md)
