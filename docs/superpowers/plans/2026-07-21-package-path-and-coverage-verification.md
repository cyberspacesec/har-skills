# 包路径一致性核查与覆盖率复核 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: `superpowers:subagent-driven-development`
> Steps use checkbox (`- [ ]`) syntax.

**Goal:** 核查全仓库 Go 包路径与 GitHub 仓库 `cyberspacesec/har-skills` 的一致性，并复核单元测试覆盖率是否仍为 100%，形成交付验收结论。

**Architecture:** 需求两项均属"核查性"而非"开发性"任务——无新功能需实现，仅需系统性调研全仓库 `package` 声明、`import` 路径、`go.mod` module 声明、goreleaser ldflags 注入路径，并复跑覆盖率测试（含 race 模式）确认 100% 稳健。数据流：go.mod module 声明 → 各子包 package 声明 → import 引用 → goreleaser ldflags `-X` 注入路径 → 实际构建产物版本输出，全链路核对同一 module 路径 `github.com/cyberspacesec/har-skills`。

**Tech Stack:** Go 1.24（go.mod），GoReleaser v2 schema，goreleaser-action `~> v2`，cobra/viper CLI，golangci-lint，GitHub Actions

**Risks:**
- 历史项目曾用旧模块路径 `cyberspacesec/go-har`，rename 后可能有残留 import → 缓解：全仓库 grep 旧路径
- ldflags 注入路径若与实际 package 路径不匹配会导致版本信息静默丢失 → 缓解：本地构建模拟 ldflags 验证 `--version` 输出
- testdata 在测试运行时会被重新生成时间戳，造成工作区脏 → 缓解：核查后 `git checkout` 还原

---

### Task 1: 调研全仓库包路径与 GitHub 仓库路径一致性

**Depends on:** None
**Files:**（核查性任务，无文件修改）
- Read: `go.mod`
- Read: `cmd/har/main.go`、`cmd/har/cmd/root.go`、`cmd/har/internal/*.go`
- Read: `.goreleaser.yaml`
- Read: `CLAUDE.md`

- [x] **Step 1: 核查 go.mod module 声明与 GitHub 远程仓库路径**

Run: `head -1 go.mod && git remote -v`
Expected:
  - Output contains: `module github.com/cyberspacesec/har-skills`
  - Output contains: `git@github.com:cyberspacesec/har-skills.git`
  - 两者 owner/name 段完全一致

- [x] **Step 2: 核查 cmd/har 子包 package 声明与 import 路径**

Run: `find cmd -name "*.go" | while read f; do echo "$f: $(grep -m1 '^package ' $f)"; done`
Expected:
  - `cmd/har/main.go: package main`，import `github.com/cyberspacesec/har-skills/cmd/har/cmd`
  - `cmd/har/cmd/*.go: package cmd`
  - `cmd/har/internal/*.go: package internal`
  - 所有 import 均以 `github.com/cyberspacesec/har-skills` 为根

- [x] **Step 3: 全仓库扫描旧模块路径 `cyberspacesec/go-har` 残留**

Run: `grep -rn "cyberspacesec/go-har" --include="*.go" --include="*.yaml" --include="*.yml" --include="*.md" .`
Expected:
  - 无任何输出（残留为零）

- [x] **Step 4: 核查 goreleaser ldflags 注入路径与实际包路径对齐**

Run: `grep -n "X github.com" .goreleaser.yaml`
Expected:
  - `-X github.com/cyberspacesec/har-skills/cmd/har/cmd.version={{.Version}}`
  - `-X github.com/cyberspacesec/har-skills/cmd/har/cmd.commit={{.Commit}}`
  - `-X github.com/cyberspacesec/har-skills/cmd/har/cmd.date={{.Date}}`
  - 与 `cmd/har/cmd/root.go` 中 `var version/commit/date` 声明的包路径一致

- [x] **Step 5: 本地构建模拟 ldflags 验证版本注入**

Run: `go build -ldflags "-X github.com/cyberspacesec/har-skills/cmd/har/cmd.version=v0.1.2-test -X github.com/cyberspacesec/har-skills/cmd/har/cmd.commit=abc123 -X github.com/cyberspacesec/har-skills/cmd/har/cmd.date=2026-07-21" -o /tmp/har-ldflag-test ./cmd/har/ && /tmp/har-ldflag-test --version`
Expected:
  - Exit code: 0
  - Output contains: `HAR Skills v0.1.2-test`、`commit: abc123`、`date: 2026-07-21`
  - 三个变量均注入成功（注意：ldflags 字符串必须整体用引号包裹，否则 shell 会拆分多个 `-X`）

---

### Task 2: 复核单元测试覆盖率 100%

**Depends on:** None
**Files:**（核查性任务，无文件修改）
- Read: 根包所有 `*_cov_test.go`、`coverage_final_test.go`

- [x] **Step 1: 常规模式覆盖率**

Run: `go test -coverprofile=/tmp/c.out -count=1 . 2>&1 | tail -1`
Expected:
  - Output contains: `coverage: 100.0% of statements`

- [x] **Step 2: race 模式覆盖率稳健性**

Run: `go test . -race -cover -count=1 -timeout 120s 2>&1 | tail -1`
Expected:
  - Exit code: 0
  - Output contains: `coverage: 100.0% of statements`
  - 无 race 警告

- [x] **Step 3: 函数级覆盖率无未达 100% 的函数**

Run: `go tool cover -func=/tmp/c.out | awk '$3 != "100.0%" && $1 != "total"'`
Expected:
  - 无任何输出（所有函数均 100%）

- [x] **Step 4: 还原测试副作用产生的 testdata 时间戳改动**

Run: `git checkout testdata/full.har testdata/invalid_date.har testdata/large.har`
Expected:
  - Exit code: 0
  - `git status --short` 无 testdata 改动残留

---

### Task 3: 交付验收结论

**Depends on:** Task 1, Task 2
**Files:**
- Create: 本 Plan 文档（已记录核查证据）

- [x] **Step 1: 汇总核查结论**

核查结果（2026-07-21）：

| 需求 | 状态 | 证据 |
|------|------|------|
| 包路径与 GitHub 仓库一致 | ✅ 已一致 | `module github.com/cyberspacesec/har-skills` = GitHub `cyberspacesec/har-skills`；全仓库零旧路径 `go-har` 残留；ldflags 注入路径与实际包路径对齐；本地构建 `--version` 三变量注入成功 |
| 单元测试覆盖率 100% | ✅ 已达 100% | 常规模式 `coverage: 100.0%`；race 模式 `coverage: 100.0%`；函数级无未达 100% 的函数 |

- [x] **Step 2: 提交 Plan 文档**

Run: `git add docs/superpowers/plans/2026-07-21-package-path-and-coverage-verification.md && git commit -m "docs(plans): verify package path consistency and 100% coverage"`
