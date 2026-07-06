---
title: Contributing Guide
---

# Contributing Guide

Thank you for your interest in HAR Skills! We welcome contributions of every kind: bug reports, feature requests, documentation improvements, new examples, and code. This guide will help you participate smoothly.

## Reporting Issues

Before opening a new issue, please search the [existing issues](https://github.com/cyberspacesec/har-skills/issues) to avoid duplicates. If nothing matches, open a new one and include as many of the following elements as possible so maintainers can act quickly:

- **Version**: the `har` CLI version (`har --version` or `git describe --tags`) and your Go version (`go version`).
- **Platform**: OS and architecture (e.g. `linux/amd64`, `darwin/arm64`). If you use a pre-built binary, note that it came from Releases.
- **Reproduction steps**: a minimal command sequence, ideally reproducible against a real HAR file in `testdata/`.
- **Expected behavior**: what you believe the correct result should be.
- **Actual behavior**: what you observed, with the full error output or logs.
- **HAR file**: if feasible, attach a minimized HAR snippet. If it contains sensitive data, run `har redact` on it first.

A high-quality bug report looks roughly like this:

```markdown
**Version**: har 0.1.0 / go1.22.1
**Platform**: darwin/arm64 (pre-built binary from Releases)
**Reproduction steps**:
1. har -f testdata/example.har timing --summary
2. ...
**Expected**: aggregated DNS / connect / wait statistics
**Actual**: panic: runtime error: index out of range
**Full output**: <paste logs>
```

## Development Workflow

The standard GitHub collaboration flow:

1. **Fork** the repository to your GitHub account.
2. **Clone** your fork locally: `git clone https://github.com/<you>/har-skills.git`.
3. **Create a branch**: `git checkout -b feature/amazing-feature` (use `fix/...` for bug fixes, `docs/...` for documentation).
4. **Commit your changes**: follow the code standards below; use clear imperative commit messages (e.g. `Add cache-hit filter to find command`).
5. **Push the branch**: `git push origin feature/amazing-feature`.
6. **Open a Pull Request** targeting `main`; reference the relevant issue in the description (e.g. `Closes #42`).

For commit messages, follow the repo's existing [Conventional Commits](https://www.conventionalcommits.org/) style (`feat:` / `fix:` / `docs:` / `refactor:` / `test:`).

## Code Standards

### General Rules

- Format all code with `gofmt` (or `goimports`); CI enforces this.
- Follow Go's official [Effective Go](https://go.dev/doc/effective_go) and the existing project style: return errors explicitly, use named return values for documentation, and every exported symbol must have a comment.
- Every exported function, type, and constant must have a doc comment that begins with the symbol's name (readable by `go doc`).
- Behavioral changes to a public API must be reflected in the docs (see "Documentation Requirements" below).
- Do not introduce new runtime dependencies. The SDK root package stays zero-dependency at runtime; `testify` is for tests only.

### Naming Conventions

- Type names and exported function names use `MixedCaps`; no underscores (except in `*_test.go` files).
- Receiver names should be short and consistent: `h` for `*Har`, `fr` for `*FilterResult`, `eb` for `*EntryBuilder`.
- Constructors begin with `New` (`NewHar`, `NewHarError`, `NewStreamingParserFromFile`).
- Functional options begin with `With` (`WithFilterMethod`, `WithMemoryOptimized`).
- Boolean predicates begin with `Is` / `Has` (`IsFileSystemError`, `HasPartialErrors`).

### Clone Semantics and Mutable Methods

When contributing new methods, honor the project's existing clone semantics:

- Transformation methods on `*Har` return a **new instance**; the receiver is left unchanged (`Redact`, `RemoveHeaders`, `RewriteURL`, etc.).
- Only methods whose names end with `InPlace` mutate in place.
- Builders (`HarBuilder` / `EntryBuilder`) are chainable and produce an immutable result via `Build()`.

### Error Handling

- Public APIs return `error`; never panic.
- Parsing-related errors use the enhanced error system (`*HarError` in `errors.go`), carrying an `ErrorCode` and `Field` / `Metadata`.
- Do not swallow errors: either return them, or wrap with `fmt.Errorf("...: %w", err)` and return.
- Output aimed at AI agents must be readable — short sentences, not raw JSON-deserialization dumps.

## Testing

Code contributions must come with tests, and existing tests must not regress. Run the tests:

```bash
# Full test suite (CLI and examples included)
go test ./...

# Root package with race detection (recommended for contributors)
go test -race ./.

# Tests for a single file
go test -run TestFilter ./.

# Coverage
go test -coverprofile=coverage.out ./.
go tool cover -html=coverage.out
```

Requirements:

- New public APIs must have unit tests covering the happy path and at least one edge case.
- Validate parsing, filtering, and transformation against the real HAR files in `testdata/`.
- A bug-fix PR must include a regression test that fails before the fix and passes after.

## Documentation Requirements

HAR Skills is an AI-agent-oriented skill project, so documentation is as important as code. When you change public behavior, update these in lockstep:

- **README.md**: installation, command quick reference, SDK quick reference.
- **CLAUDE.md**: the project-level progressive-disclosure document consumed directly by AI agents.
- **website/**: this VitePress site. The Chinese (`zh/`) and English (`en/`) pages for a given topic must be updated together.
- **Code comments**: godoc comments for public APIs.

If your PR is documentation-only, prefix the title with `docs:`.

> Tip: Chinese and English pages must be updated in pairs and stay semantically consistent. If you are fluent in only one language, add the `needs-translation` label in the PR description and a maintainer will help complete the other language.

## Changes and Releases

- **Versioning**: the project follows Semantic Versioning (SemVer), with tags like `v0.1.0`. Releases are automated by maintainers via GitHub Actions + GoReleaser (see `.goreleaser.yaml`).
- **Changelog**: breaking changes must include a `BREAKING CHANGE:` paragraph in the PR description so maintainers can roll them into Release Notes.
- **Adding a CLI command**: add a command file under `cmd/har/cmd/`, register it in `root.go`, and update both the command reference in `CLAUDE.md` and the [CLI Reference](/en/cli/global-flags) page on this site.
- **Adding an SDK module**: add a `.go` file in the root package with a matching `_test.go`, and add a row to the cheat sheet in the [Architecture Overview](/en/contributing/architecture).

## Previewing the Docs Locally

The docs site is built on VitePress 1.x. To preview locally:

```bash
cd website
npm install      # install dependencies the first time
npm run docs:dev # start the dev server, default http://localhost:5173
```

Build the production bundle and preview it:

```bash
npm run docs:build   # output to website/.vitepress/dist
npm run docs:preview # preview the built output
```

To edit the sidebar navigation, modify `website/.vitepress/config.ts` (`ZH_SIDEBAR` and `EN_SIDEBAR`). Chinese and English pages must be committed in pairs.

Common local-preview pitfalls:

- **Port in use**: `npm run docs:dev` defaults to port 5173. Use `npx vitepress dev --port 5174` to pick another port.
- **Missing dependency**: if `vitepress` cannot be found, make sure you ran `npm install` inside the `website/` directory.
- **New page missing from sidebar**: after adding a page, you must add an entry to the corresponding `SIDEBAR` array in `config.ts`, or it will not appear in navigation.
- **Chinese/English drift**: the build does not error when one language version is missing; verify pairs manually.

## License

This project is released under the [MIT License](https://github.com/cyberspacesec/har-skills/blob/main/LICENSE). By contributing code, you agree that your contributions will be licensed under the same MIT License.

## Code of Conduct

We expect all contributors to treat each other with respect and maintain a professional, friendly environment. Unacceptable behavior includes but is not limited to: insulting comments, personal attacks, bullying, or harassment. Maintainers may close issues, PRs, or comments that violate this code.

## Getting Help

If you have any questions:

- Ask in the comments of a relevant issue.
- Open a new `question`-type issue to request help.
- Check this docs site and `CLAUDE.md` first — most usage is covered by examples.

## Pre-PR Self-Check

Before opening a Pull Request, walk through this checklist:

- [ ] Code formatted with `gofmt` / `goimports`.
- [ ] New/changed exported symbols have godoc comments.
- [ ] New public APIs come with unit tests covering at least one edge case.
- [ ] `go test ./...` passes in full; root package `go test -race ./.` reports no races.
- [ ] Bug-fix PRs include a regression test (fails before, passes after).
- [ ] Chinese/English doc pages (`website/zh/` and `website/en/`) updated as a pair.
- [ ] `README.md` and `CLAUDE.md` updated if public behavior changed.
- [ ] No new runtime dependencies in the SDK root package; any new CLI-side dependency must be justified in the PR description.
- [ ] Commit messages use Conventional Commits prefixes.
- [ ] PR description references the relevant issue (e.g. `Closes #42`).

## Common Pitfalls

- **Forgetting `ToStandard()`**: `ParseFile` returns a `HARProvider`, not a `*Har` — treating it as `*Har` will not compile. Convert explicitly when you need the full API.
- **Mixing struct-style and functional-style options**: `Filter(FilterOptions{...})` and `FilterWith(WithFilter*...)` are equivalent but their arguments cannot be mixed within one call. Don't pass `FilterOptions` fields and `WithFilter*` options together.
- **Breaking clone semantics**: mutating the receiver inside a `*Har` transformation method breaks chained safety. Always clone-then-modify, except for `InPlace`-suffixed methods.
- **Tests depending on specific testdata**: tests should rely on HAR files committed under `testdata/`, not on external networks or freshly generated files.

## Adding an Example

Every example under `examples/` is a standalone `package main`. Steps to add one:

1. Create a new subdirectory under `examples/` (e.g. `examples/audit/`) and add a `main.go`.
2. The top of `main.go` declares `package main`; import the SDK as `har "github.com/cyberspacesec/har-skills"`.
3. Prefer real HAR files from `testdata/` as default input, avoiding any dependency on external networks.
4. Provide command-line flags (via the standard library `flag`) so users can supply a file path and output format.
5. Register the new example in the [Examples](/en/examples/) page on this site and in the `examples/` directory layout.
6. Ensure the example runs directly via `go run main.go` with no new runtime dependencies.

Each example should focus on a single scenario and stay within single-file readability. If a scenario is complex, split it into multiple examples rather than piling everything into one `main.go`.

Thank you for contributing!
