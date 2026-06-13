# CLAUDE.md — HAR Skills: AI Agent Skill for HAR File Analysis

> This file provides progressive disclosure documentation for AI agents to use the HAR Skills SDK and CLI tool. It covers everything from basic usage to advanced analysis workflows. This repository is designed to be consumed as an AI Agent Skill.

## Project Overview

**HAR Skills** is an AI-oriented Go SDK and CLI tool for HAR (HTTP Archive) files. It provides:
- **SDK** (`pkg/har/`): 40 Go modules with 70+ methods for HAR parsing, analysis, transformation, and export
- **CLI** (`cmd/har/`): 23 Cobra-based commands exposing all SDK capabilities via terminal
- **Skill Docs**: Progressive disclosure documentation (this file) for AI agent consumption
- **Install**: `go install github.com/cyberspacesec/har-skills/cmd/har@latest`

## Quick Start (CLI)

### Installation

**Option 1: Download Pre-built Binary (Recommended)**

Download the latest release for your platform from:
https://github.com/cyberspacesec/har-skills/releases/latest

```bash
# Linux x86_64 example
curl -sL https://github.com/cyberspacesec/har-skills/releases/latest/download/har-skills_0.1.0_linux_x86_64.tar.gz | tar xz
sudo mv har /usr/local/bin/

# macOS Apple Silicon example
curl -sL https://github.com/cyberspacesec/har-skills/releases/latest/download/har-skills_0.1.0_darwin_arm64.tar.gz | tar xz
sudo mv har /usr/local/bin/

# Windows: download .zip, extract har.exe, add to PATH
```

Available platforms: linux (x86_64/arm64/armv6/armv7/i386), darwin (x86_64/arm64), windows (x86_64/i386), freebsd (x86_64/i386)

**Option 2: Go Install**

```bash
go install github.com/cyberspacesec/har-skills/cmd/har@latest
```

**Option 3: Build from Source**

```bash
git clone https://github.com/cyberspacesec/har-skills.git
cd har-skills
go build -o har ./cmd/har/

# With version info
go build -ldflags "-X github.com/cyberspacesec/har-skills/cmd/har/cmd.version=$(git describe --tags 2>/dev/null || echo dev)" -o har ./cmd/har/

# Cross-compile for other platforms
GOOS=darwin GOARCH=arm64 go build -o har-darwin-arm64 ./cmd/har/
GOOS=windows GOARCH=amd64 go build -o har.exe ./cmd/har/
```

### Basic Usage

```bash
har -f <har-file> <command> [flags]

# Read from stdin
cat capture.har | har info

# Output formats
har -f capture.har info --format json     # JSON
har -f capture.har list --format csv      # CSV
har -f capture.har list --format yaml     # YAML (default: text)

# Write to file
har -f capture.har info -o report.json
```

## Global Flags

All commands inherit these flags:

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--file` | `-f` | | HAR file path (use `-` for stdin) |
| `--format` | | `text` | Output format: text, json, csv, yaml |
| `--output` | `-o` | | Output file path |
| `--no-header` | | `false` | Suppress table headers in text/csv |
| `--config` | | | Config file path (default `$HOME/.har.yaml`) |

Environment variables: `HAR_FILE`, `HAR_FORMAT`, `HAR_OUTPUT` (via Viper).

---

## Command Reference (Progressive Disclosure)

### Level 1: Basic Operations

These commands cover the most common HAR analysis tasks.

#### `info` — HAR File Summary

Show overview: version, creator, entry count, transfer size, timing percentiles, status codes, methods, domains, content types.

```bash
har -f capture.har info                    # Text summary
har -f capture.har info --format json      # Full statistics as JSON
```

#### `list` — List Entries

List all requests with filtering, sorting, and limiting.

```bash
har -f capture.har list                    # All entries
har -f capture.har list --limit 20         # Top 20
har -f capture.har list --sort size        # Sort by response size
har -f capture.har list --method GET       # Only GET requests
har -f capture.har list --status 200       # Only 200 responses
har -f capture.har list --domain api.example.com
```

**Flags**: `--limit/-n`, `--sort` (time/size/url/status), `--order` (asc/desc), `--method`, `--status`, `--domain`

#### `find` — Search Entries

Search by URL pattern, status code, headers, cookies, time range, server IP, and more.

```bash
har -f capture.har find "api/users"        # URL substring match
har -f capture.har find "^/api/v2" --regex # Regex match
har -f capture.har find --errors            # All 4xx/5xx
har -f capture.har find --redirects         # All 3xx
har -f capture.har find --slow 1000         # Slower than 1s
har -f capture.har find --slowest 10        # Top 10 slowest
har -f capture.har find --fastest 5         # Top 5 fastest
har -f capture.har find --largest 10        # Top 10 by response size
har -f capture.har find --status-min 400 --status-max 599
har -f capture.har find --domain api.example.com --content-type "application/json"
har -f capture.har find --header "Authorization"              # Has request header
har -f capture.har find --response-header "Content-Type:application/json"  # Response header
har -f capture.har find --cookie "session_id"                 # Has cookie
har -f capture.har find --start-time "2024-01-01T00:00:00Z" --end-time "2024-12-31T23:59:59Z"
har -f capture.har find --server-ip "10.0.0.1"               # By server IP
har -f capture.har find --connection "ABC123"                  # By connection ID
har -f capture.har find --cache-hits                           # Entries with cache hits
```

**Flags**: `--regex`, `--method`, `--status-code`, `--status-min`, `--status-max`, `--content-type`, `--domain`, `--header` (stringSlice, request headers), `--response-header` (stringSlice, response headers), `--cookie`, `--start-time`, `--end-time`, `--server-ip`, `--connection`, `--cache-hits`, `--resource-type`, `--errors`, `--redirects`, `--slow`, `--slowest`, `--fastest`, `--largest`, `--limit/-n`

#### `headers` — Show Headers

Display request and response headers for matching entries.

```bash
har -f capture.har headers "example.com"    # All headers
har -f capture.har headers --request        # Only request headers
har -f capture.har headers --response --name content-type
har -f capture.har headers --limit 5        # Show 5 entries
```

**Flags**: `--request`, `--response`, `--name`, `--limit/-n`

#### `timing` — Timing Breakdown

Analyze DNS, connect, SSL, send, wait, receive phases per request.

```bash
har -f capture.har timing                  # Per-entry timing
har -f capture.har timing --summary        # Aggregate summary
har -f capture.har timing --sort wait      # Sort by wait time
har -f capture.har timing --filter "api"   # Filter by URL
har -f capture.har timing --limit 10       # Top 10
```

**Flags**: `--filter`, `--sort` (time/wait/dns/connect), `--limit/-n`, `--summary`

#### `extract` — Extract Response Content

Extract and decode response bodies from HAR entries.

```bash
har -f capture.har extract "data.json"     # Extract matching content
har -f capture.har extract --index 0       # Extract entry at index 0
har -f capture.har extract --all          # Extract all matching
har -f capture.har extract --index 3 -o response.json  # Save to file
```

**Flags**: `--index`, `--decode` (default true), `--all`

---

### Level 2: File Operations

Commands for comparing, merging, splitting, and validating HAR files.

#### `diff` — Compare Two HAR Files

Find added, removed, and modified requests between two HAR files.

```bash
har diff capture1.har capture2.har               # Basic diff
har diff a.har b.har --include-body              # Compare response bodies
har diff a.har b.har --compare-by-url            # Match by URL instead of index
har diff a.har b.har --ignore-headers Cookie,Date  # Ignore specific headers
```

**Flags**: `--ignore-headers` (stringSlice), `--ignore-timings`, `--ignore-dates`, `--include-body`, `--compare-by-url`

#### `merge` — Merge HAR Files

Combine multiple HAR files into one.

```bash
har merge part1.har part2.har part3.har           # Merge 3 files
har merge *.har --deduplicate -o merged.har       # Merge & deduplicate
har merge a.har b.har --sort-by-time=false        # Don't sort by time
```

**Flags**: `--sort-by-time` (default true), `--deduplicate`

#### `split` — Split HAR Files

Break a large HAR file into smaller files by various criteria.

```bash
har -f capture.har split --by domain -o by-domain  # Split by domain
har -f capture.har split --by page                   # Split by page reference
har -f capture.har split --by time --interval 30m    # Split every 30 minutes
har -f capture.har split --by size --max-entries 50  # Split every 50 entries
har -f capture.har split --by status                 # Split by status code range
har -f capture.har split --by method                 # Split by HTTP method
```

**Flags**: `--by` (page/domain/time/size/status/method), `--interval` (duration), `--max-entries` (int), `-o` (output prefix)

#### `validate` — Validate HAR Spec Compliance

Check HAR file against the specification.

```bash
har -f capture.har validate                        # Standard validation
har -f capture.har validate --strict               # Strict validation
har -f capture.har validate --strict --timings-tolerance 5  # Custom tolerance
```

**Flags**: `--strict`, `--timings-tolerance` (float64, default 10)

---

### Level 3: Security & Privacy

Commands for security auditing and data sanitization.

#### `security` — Security Audit

Comprehensive security analysis checking headers, cookies, mixed content, CORS, and information disclosure.

```bash
har -f capture.har security                        # Full audit
har -f capture.har security --severity high        # Only HIGH findings
har -f capture.har security --format json          # JSON report
har -f capture.har security --check-headers --check-cors  # Selective checks
```

**Output**: Security score (0-100), findings grouped by severity (HIGH/MEDIUM/LOW/INFO).

**Flags**: `--check-headers`, `--check-cookies`, `--check-mixed-content`, `--check-sensitive-data`, `--check-cors`, `--check-info-disclosure`, `--severity` (all/info/low/medium/high)

#### `redact` — Redact Sensitive Data

Remove passwords, tokens, API keys, and other secrets from HAR files.

```bash
har -f capture.har redact -o redacted.har          # Default redaction
har -f capture.har redact --header X-Custom-Key    # Add custom header to redact
har -f capture.har redact --cookie session         # Add custom cookie to redact
har -f capture.har redact --redact-ips             # Anonymize IP addresses
har -f capture.har redact --replacement "***"       # Custom replacement text
har -f capture.har redact --in-place               # Modify file in place
```

**Default redaction targets**: Authorization, Proxy-Authorization, X-Api-Key, X-Auth-Token, cookies named session/token/auth/password, query params password/token/api_key/secret/access_token, POST fields password/secret/token.

**Flags**: `--defaults` (default true), `--header`, `--cookie`, `--query-param`, `--post-field`, `--replacement`, `--redact-ips`, `--in-place`

---

### Level 4: Deep Analysis

Advanced analysis commands for performance, caching, cookies, and waterfall visualization.

#### `performance` — Performance Scoring

Lighthouse-style performance scoring with grade (A/B/C/D) and recommendations.

```bash
har -f capture.har performance                    # Score + recommendations
har -f capture.har performance --format json      # JSON report
```

**Categories**: TTFB (20%), Total Load Time (20%), Request Count (15%), Transfer Size (15%), Cache Efficiency (15%), Compression (15%)

#### `cookie` — Cookie Analysis

Audit cookie security attributes and track cookie evolution across requests.

```bash
har -f capture.har cookie                          # Cookie security audit
har -f capture.har cookie --evolution             # Track cookie changes over time
har -f capture.har cookie --name "session_id"     # Filter by cookie name
har -f capture.har cookie --severity medium       # Only MEDIUM+ findings
```

**Flags**: `--audit` (default true), `--evolution`, `--name`, `--severity`

#### `cache` — Cache Analysis

Analyze Cache-Control, ETag, Last-Modified, Vary headers and assess cacheability.

```bash
har -f capture.har cache                          # Full cache analysis
har -f capture.har cache --non-cacheable         # Only non-cacheable entries
har -f capture.har cache --url "https://api.example.com"
```

**Flags**: `--non-cacheable`, `--url`

#### `index` — Build & Query Entry Index

Build an in-memory index for fast lookups by URL, method, status, domain, MIME type, or URL pattern.

```bash
har -f capture.har index --stats                  # Index statistics
har -f capture.har index --url "https://api.example.com/users"  # Lookup by exact URL
har -f capture.har index --method POST             # Lookup by method
har -f capture.har index --status 404              # Lookup by status code
har -f capture.har index --domain "api.example.com" # Lookup by domain
har -f capture.har index --mime "application/json"  # Lookup by MIME type
har -f capture.har index --pattern "api/v[0-9]+"    # Lookup by URL regex
```

**Flags**: `--stats`, `--url`, `--method`, `--status`, `--domain`, `--mime`, `--pattern`

#### `domains` — Per-Domain Statistics

Show detailed statistics broken down by domain.

```bash
har -f capture.har domains                        # All domains sorted by request count
har -f capture.har domains --sort time            # Sort by total time
har -f capture.har domains --sort size --limit 10 # Top 10 domains by size
har -f capture.har domains --sort errors          # Sort by error count
```

**Flags**: `--sort` (count/time/size/errors), `--limit/-n`

#### `content` — Content Type Analysis

Show content type breakdown with sizes, compression, and MIME category distribution.

```bash
har -f capture.har content                        # Content summary by category
har -f capture.har content --by-mime              # Detailed MIME type breakdown
har -f capture.har content --format json          # JSON output
```

**Flags**: `--by-mime`

#### `connections` — Connection Reuse Analysis

Show which entries share connections (connection pooling analysis).

```bash
har -f capture.har connections                    # Connection reuse table
har -f capture.har connections --format json      # JSON output
```

#### `waterfall` — Waterfall & Timeline

Visualize request timing as a waterfall, analyze critical path, concurrency, and SLA compliance.

```bash
har -f capture.har waterfall                      # ASCII waterfall
har -f capture.har waterfall --critical-path      # Critical rendering path
har -f capture.har waterfall --concurrency         # Concurrency timeline
har -f capture.har waterfall --sla "API:/api/:2000" "Static:/static/:500"
har -f capture.har waterfall --page-timings        # Page timing metrics
```

**Flags**: `--critical-path`, `--concurrency`, `--sla` (stringSlice: name:urlPattern:maxDurationMs), `--page-timings`

---

### Level 5: Transformation & Export

Commands for modifying and exporting HAR data.

#### `transform` — Transform Requests

Rewrite URLs, add/remove headers, change schemes, modify query parameters.

```bash
har -f staging.har transform --rewrite-url "http://localhost->https://api.example.com" -o prod.har
har -f capture.har transform --remove-header Authorization,Cookie
har -f capture.har transform --add-header "X-Env:production" --add-header-target request
har -f capture.har transform --change-scheme "http->https"
har -f capture.har transform --remove-query-param "_"
```

**Flags**: `--rewrite-url` (format: from->to), `--remove-header`, `--add-header` (format: name:value), `--add-header-target` (request/response/both), `--change-scheme` (format: from->to), `--remove-query-param`, `--in-place`

#### `export` — Export to Other Formats

Convert HAR data to various formats for replay, analysis, or documentation.

```bash
har -f capture.har export curl                    # Generate curl commands
har -f capture.har export wget                    # Generate wget commands
har -f capture.har export python                  # Generate Python requests code
har -f capture.har export postman -o collection.json
har -f capture.har export xml -o capture.xml
har -f capture.har export yaml -o capture.yaml
har -f capture.har export json --index 0          # Single entry as JSON
har -f capture.har export jsonl -o entries.jsonl  # JSON Lines (one per entry)
har -f capture.har export csv -o data.csv         # CSV table
har -f capture.har export markdown -o report.md   # Markdown table
har -f capture.har export html -o report.html     # HTML table
har -f capture.har export text                    # Plain text table
```

**Positional arg**: format (curl/wget/python/postman/xml/yaml/json/jsonl/csv/markdown/html/text)
**Flags**: `--index`, `--filter`

#### `dedup` — Find/Remove Duplicates

Identify or remove duplicate/near-duplicate requests using three strategies.

```bash
har -f capture.har dedup                          # Find duplicates (pattern strategy)
har -f capture.har dedup --strategy exact         # Exact URL matching
har -f capture.har dedup --strategy content-hash  # Content hash matching
har -f capture.har dedup --remove -o cleaned.har  # Remove duplicates
har -f capture.har dedup --ignore-param "timestamp" --ignore-param "_"
har -f capture.har dedup --compare-headers --compare-body
```

**Flags**: `--strategy` (exact/pattern/content-hash), `--ignore-param`, `--compare-headers`, `--compare-body`, `--remove`

#### `replay` — Replay HTTP Requests

Re-execute recorded HTTP requests with configurable options.

```bash
har -f capture.har replay --dry-run               # Preview (no real requests)
har -f capture.har replay --timeout 10s           # 10s timeout
har -f capture.har replay --skip-ssl             # Skip SSL verification
har -f capture.har replay --index 0              # Replay single entry
har -f capture.har replay --filter "api"          # Replay only API requests
har -f capture.har replay --header "Authorization:Bearer token"
har -f capture.har replay --save-har results.har # Save replay results as HAR
```

**Flags**: `--dry-run`, `--timeout`, `--no-follow-redirects`, `--max-redirects`, `--skip-ssl`, `--header`, `--index`, `--filter`, `--save-har`

---

## SDK Quick Reference (Go API)

### Import

```go
import har "github.com/cyberspacesec/har-skills"
```

### Parse

```go
// From file
h, err := har.ParseHarFile("capture.har")

// From file with auto-detect (gzip, etc.)
h, err := har.ParseHarFileAuto("capture.har")

// From bytes
h, err := har.ParseHar(data)

// From io.Reader
h, err := har.ParseHarFromReader(reader)

// With options
provider, err := har.Parse("capture.har", har.OptFast)
```

### Analyze

```go
// Statistics
stats := h.Statistics()
summary := h.Summary()         // string
timing := h.TimingStatistics()
domains := h.DomainSummary()
statusCodes := h.StatusCodeDistribution()
methods := h.MethodDistribution()
contentTypes := h.ContentTypeDistribution()
slowest := h.SlowestRequests(10)
fastest := h.FastestRequests(10)
largest := h.LargestResponses(10)

// Security
report := h.SecurityAudit()    // *SecurityReport
score := report.Score           // 0-100
highFindings := report.FindBySeverity("high")

// Cookie
cookieReport := h.CookieAudit()         // *CookieAuditReport
evolution := h.CookieEvolution()        // map[string][]CookieEvolutionEntry

// Cache
cacheReport := h.CacheAnalysis()        // *CacheReport

// Performance
perfReport := h.PerformanceScore()      // *PerformanceReport
grade := perfReport.Grade()             // "A", "B", "C", "D"
```

### Filter & Search

```go
// Filter with options
result := h.FilterWith(
    har.WithFilterURL("api/users"),
    har.WithFilterMethod("GET"),
    har.WithFilterStatusCode(200),
)
entries := result.GetAll()            // []Entries
result = result.SortByDurationDesc().Limit(10)

// Direct find methods
errors := h.FindErrors()              // *FilterResult
redirects := h.FindRedirects()        // *FilterResult
slow := h.FindSlowRequests(1000)      // *FilterResult (ms)
byDomain := h.FindByDomain("api.example.com")
byURL := h.FindByURL("pattern", true) // true = regex
byRange := h.FindByStatusCodeRange(400, 599)
```

### Transform & Redact

```go
// Redact sensitive data
opts := har.DefaultRedactOptions()
redacted := h.Redact(opts)            // returns new *Har

// Transform URLs
transformed := h.RewriteURL("http://localhost", "https://prod.example.com")

// Remove headers
cleaned := h.RemoveHeaders([]string{"Authorization", "Cookie"})

// Add headers
withHeaders := h.AddHeaders(map[string]string{"X-Env": "prod"}, "both")

// Custom transform rules
rules := []har.TransformRule{
    {Type: har.TransformURLRewrite, Pattern: "http://", Replacement: "https://"},
    {Type: har.TransformHeaderRemove, HeaderName: "X-Debug"},
}
result := h.Transform(rules)
```

### Export

```go
curl := h.ToCurl()                     // string
wget := h.ToWget()                     // string
python := h.ToPythonRequests()         // string
postman, _ := h.ToPostmanCollection() // []byte
xml, _ := h.ToXML()                   // string
yaml, _ := h.ToYAML()                 // string
json, _ := h.ToJSON(true)             // []byte (indent=true)
```

### Diff & Merge

```go
// Diff
diffResult := har.Diff(har1, har2, har.DefaultDiffOptions())
report := diffResult.Report("text")   // string

// Merge
merged := har.Merge(har1, har2, har3)
```

### Split

```go
byPage := h.SplitByPage()             // map[string]*Har
byDomain := h.SplitByDomain()        // map[string]*Har
byTime := h.SplitByTimeRange(time.Hour) // []*Har
bySize := h.SplitBySize(100)          // []*Har
byStatus := h.SplitByStatusCode()    // map[string]*Har
byMethod := h.SplitByMethod()        // map[string]*Har
```

### Validate

```go
err := har.ValidateHarFile(h)          // error if invalid
err = har.ValidateStrict(h)           // stricter checks
timingErrs := har.ValidateTimingsConsistency(h, 10.0) // tolerance in ms
```

---

## Common Workflows

### Workflow: Security Audit & Remediation

```bash
# 1. Run security audit
har -f capture.har security --format json -o security-report.json

# 2. Redact sensitive data before sharing
har -f capture.har redact --redact-ips -o redacted.har

# 3. Validate the redacted file
har -f redacted.har validate --strict
```

### Workflow: Performance Optimization

```bash
# 1. Get performance score
har -f capture.har performance

# 2. Find slow requests
har -f capture.har find --slow 1000 --format json

# 3. Check cache efficiency
har -f capture.har cache --non-cacheable

# 4. Analyze waterfall for critical path
har -f capture.har waterfall --critical-path --page-timings
```

### Workflow: API Migration Testing

```bash
# 1. Capture from staging
har -f staging.har info

# 2. Transform URLs to production
har -f staging.har transform --rewrite-url "http://staging->https://prod" -o prod-ready.har

# 3. Dry-run replay to test
har -f prod-ready.har replay --dry-run

# 4. Actually replay with timeout
har -f prod-ready.har replay --timeout 10s --format json -o replay-results.json

# 5. Diff original vs replay
# (requires capturing replay results into a HAR first)
```

### Workflow: Data Cleaning & Sharing

```bash
# 1. Remove duplicates
har -f raw.har dedup --remove -o deduped.har

# 2. Redact sensitive data
har -f deduped.har redact --redact-ips -o clean.har

# 3. Split by domain for per-team analysis
har -f clean.har split --by domain -o per-domain

# 4. Export as Postman collection
har -f clean.har export postman -o collection.json
```

---

## Architecture Notes

### Package Structure
- **Root package** (`har.go`): Type aliases re-exporting everything from `pkg/har`
- **SDK** (`pkg/har/`): 40 modules, 741 tests, all functionality
- **CLI** (`cmd/har/`): 20 Cobra commands, installable binary
- **Internal helpers** (`cmd/har/internal/`): Shared loader and output formatter

### CLI Install Targets
```bash
# Install globally
go install github.com/cyberspacesec/har-skills/cmd/har@latest

# Build from source
go build -o har ./cmd/har/

# Cross-compile
GOOS=darwin GOARCH=arm64 go build -o har-darwin-arm64 ./cmd/har/
GOOS=windows GOARCH=amd64 go build -o har.exe ./cmd/har/
```

### Key SDK Patterns
- All `*Har` methods return new `*Har` instances (clone + modify) unless named `InPlace`
- `FilterResult` supports chaining: `h.FilterWith(...).SortByDurationDesc().Limit(10)`
- `HARProvider` interface allows different parsing strategies (standard, optimized, lazy, streaming)
- All `Parse*` functions return `HARProvider`; use `.ToStandard()` to get `*Har` for full API

### Go Version & Dependencies
- Go 1.19+
- CLI: `spf13/cobra` v1.8.0, `spf13/viper` v1.18.2
- SDK: Zero external runtime dependencies (testify for tests only)
