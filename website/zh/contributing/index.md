---
title: 贡献指南
---

# 贡献指南

感谢您对 HAR Skills 项目的关注！我们欢迎所有形式的贡献：错误报告、功能请求、文档改进、示例补充和代码贡献。本指南将帮助您顺畅地参与项目开发。

## 报告问题

在提交新 issue 之前，请先搜索 [现有 issue](https://github.com/cyberspacesec/har-skills/issues) 是否已经覆盖了你的问题，避免重复。如果没有，请新建一个 issue，并尽量包含以下要素，以便维护者快速定位：

- **版本信息**：`har` CLI 版本（`har --version` 或 `git describe --tags`）、Go 版本（`go version`）。
- **平台信息**：操作系统与架构（如 `linux/amd64`、`darwin/arm64`），如使用预编译二进制请标注下载来源。
- **复现步骤**：最小可复现的命令序列，最好能基于 `testdata/` 中的真实 HAR 文件复现。
- **期望行为**：你认为正确的结果应该是什么。
- **实际行为**：你观察到的结果，附上完整的错误输出或日志。
- **HAR 文件**：如可行，附上最小化的 HAR 片段；若含敏感信息请先执行 `har redact` 脱敏。

一个高质量的 bug 报告模板大致如下：

```markdown
**版本**：har 0.1.0 / go1.22.1
**平台**：darwin/arm64（预编译二进制，Releases 下载）
**复现步骤**：
1. har -f testdata/example.har timing --summary
2. ...
**期望**：输出聚合的 DNS / connect / wait 统计
**实际**：panic: runtime error: index out of range
**完整输出**：<粘贴日志>
```

## 开发流程

标准的 GitHub 协作流程：

1. **Fork** 仓库到你的 GitHub 账号。
2. **克隆** 你的 fork 到本地：`git clone https://github.com/<you>/har-skills.git`。
3. **新建分支**：`git checkout -b feature/amazing-feature`（修复类用 `fix/...`，文档类用 `docs/...`）。
4. **提交更改**：遵循下文代码规范，提交信息用清晰的祈使句（如 `Add cache-hit filter to find command`）。
5. **推送分支**：`git push origin feature/amazing-feature`。
6. **创建 Pull Request**：目标分支为 `main`，在 PR 描述中引用相关 issue（如 `Closes #42`）。

提交信息建议参考仓库已有的 [Conventional Commits](https://www.conventionalcommits.org/) 风格（`feat:` / `fix:` / `docs:` / `refactor:` / `test:`）。

## 代码规范

### 通用规则

- 使用 `gofmt`（或 `goimports`）格式化全部代码，CI 会检查。
- 遵循 Go 官方 [Effective Go](https://go.dev/doc/effective_go) 与项目既有风格：错误显式返回、命名返回值用于文档化、导出符号必须有注释。
- 所有导出的函数、类型、常量须有以符号名开头的文档注释（`go doc` 可读）。
- 公共 API 的行为变更须同步更新文档（见下文「文档要求」）。
- 不要引入新的运行时依赖；SDK 根包保持零运行时依赖，`testify` 仅用于测试。

### 命名约定

- 类型名、导出函数名使用 `MixedCaps`，不使用下划线（除非是测试文件 `*_test.go`）。
- 接收者名应简短且一致：`*Har` 用 `h`，`*FilterResult` 用 `fr`，`*EntryBuilder` 用 `eb`。
- 构造函数以 `New` 开头（`NewHar`、`NewHarError`、`NewStreamingParserFromFile`）。
- 函数式选项以 `With` 开头（`WithFilterMethod`、`WithMemoryOptimized`）。
- 返回布尔判定的方法以 `Is` / `Has` 开头（`IsFileSystemError`、`HasPartialErrors`）。

### 克隆语义与可变方法

贡献新方法时，请遵守项目既有的克隆语义约定：

- `*Har` 上的变换方法返回**新实例**，原始对象不变（`Redact`、`RemoveHeaders`、`RewriteURL` 等）。
- 仅当方法名以 `InPlace` 结尾时才就地修改。
- 构建器（`HarBuilder` / `EntryBuilder`）可链式调用，最终通过 `Build()` 产出不可变结果。

### 错误处理

- 公共 API 返回 `error`，不要 panic。
- 解析相关错误使用增强错误体系（`errors.go` 中的 `*HarError`），附 `ErrorCode` 与 `Field` / `Metadata`。
- 不要吞掉错误：要么返回，要么用 `fmt.Errorf("...: %w", err)` 包裹后返回。
- 面向 AI Agent 的输出须可读，错误信息用简短句子，避免裸抛 JSON 反序列化原始报错。

## 测试

代码贡献必须配套测试，且现有测试不能回退。运行测试：

```bash
# 全量测试（含 CLI 与 examples）
go test ./...

# 根包测试带竞态检测（推荐贡献者在本地跑一遍）
go test -race ./.

# 单个文件的测试
go test -run TestFilter ./.

# 覆盖率
go test -coverprofile=coverage.out ./.
go tool cover -html=coverage.out
```

要求：

- 新增公共 API 必须有覆盖其正常路径与至少一个边缘情况的单元测试。
- 针对 `testdata/` 中的真实 HAR 文件验证解析、过滤、转换等行为。
- 修复 bug 的 PR 须附一个能在修复前失败、修复后通过的回归测试。

## 文档要求

HAR Skills 是一个面向 AI Agent 的技能型项目，文档与代码同等重要。改动公共行为时，请同步更新：

- **README.md**：安装方式、命令速查、SDK 速查。
- **CLAUDE.md**：项目级的渐进式披露文档，AI Agent 直接消费。
- **website/**：本 VitePress 文档站，中英双语对应页（`zh/` 与 `en/`）必须同步修改。
- **代码注释**：公共 API 的 godoc 注释。

如果只改文档，PR 标题用 `docs:` 前缀即可。

> 提示：中英文页面需成对修改且语义一致。若你只熟悉一种语言，可在 PR 描述中标注 `needs-translation`，由维护者协助补全另一语言版本。

## 变更与发布

- **版本号**：项目遵循语义化版本（SemVer），标签形如 `v0.1.0`。发布由维护者通过 GitHub Actions + GoReleaser 自动化（见 `.goreleaser.yaml`）。
- **变更日志**：重大变更须在 PR 描述中写明 `BREAKING CHANGE:` 段落，便于维护者汇总到 Release Notes。
- **新增 CLI 命令**：在 `cmd/har/cmd/` 新增命令文件，注册到 `root.go`，并同步更新 `CLAUDE.md` 命令参考与本站 [CLI 命令参考](/zh/cli/global-flags) 页。
- **新增 SDK 模块**：在根包新增 `.go` 文件，确保有对应 `_test.go`，并在 [架构总览](/zh/contributing/architecture) 的速查表中补一行。

## 文档站本地预览

文档站基于 VitePress 1.x，本地预览步骤：

```bash
cd website
npm install      # 首次需要安装依赖
npm run docs:dev # 启动开发服务器，默认 http://localhost:5173
```

构建生产版本与本地预览构建产物：

```bash
npm run docs:build   # 输出到 website/.vitepress/dist
npm run docs:preview # 预览构建产物
```

修改侧边栏导航请编辑 `website/.vitepress/config.ts`（`ZH_SIDEBAR` 与 `EN_SIDEBAR`）。中英双语页面需成对提交。

常见本地预览问题：

- **端口被占用**：`npm run docs:dev` 默认占用 5173 端口，可用 `npx vitepress dev --port 5174` 指定其他端口。
- **依赖缺失**：若提示找不到 `vitepress`，确认在 `website/` 目录下执行过 `npm install`。
- **侧边栏不显示新页面**：新增页面后须在 `config.ts` 对应 `SIDEBAR` 数组中添加条目，否则不会出现在导航中。
- **中英文不同步**：构建不会因缺失对应语言版本报错，请人工核对成对提交。

## 许可证

本项目基于 [MIT License](https://github.com/cyberspacesec/har-skills/blob/main/LICENSE) 发布。通过贡献代码，你同意你的贡献将在同一 MIT 许可证下发布。

## 行为准则

我们期望所有贡献者尊重彼此，保持专业与友好的交流环境。不可接受的行为包括但不限于：侮辱性评论、人身攻击、欺凌或骚扰。维护者有权关闭违反本准则的 issue、PR 或评论。

## 获取帮助

如果有任何疑问：

- 在相关 issue 中留言提问。
- 新建一个 `question` 类型的 issue 请求帮助。
- 优先查阅本文档站与 `CLAUDE.md`，绝大多数用法均有示例。

## PR 提交前自检清单

在提交 Pull Request 之前，请逐项核对：

- [ ] 代码已用 `gofmt` / `goimports` 格式化。
- [ ] 新增 / 修改的导出符号有 godoc 注释。
- [ ] 新增公共 API 配套单元测试，且覆盖至少一个边缘情况。
- [ ] `go test ./...` 全量通过；根包 `go test -race ./.` 无竞态。
- [ ] 修复类 PR 附回归测试（修复前失败、修复后通过）。
- [ ] 中英双语文档页（`website/zh/` 与 `website/en/`）成对更新。
- [ ] 如改动公共行为，`README.md` 与 `CLAUDE.md` 同步更新。
- [ ] 未引入新的运行时依赖（SDK 根包）；CLI 侧新依赖须在 PR 描述中说明理由。
- [ ] 提交信息使用 Conventional Commits 前缀。
- [ ] PR 描述引用了相关 issue（如 `Closes #42`）。

## 常见陷阱

- **忘记 `ToStandard()`**：`ParseFile` 返回的是 `HARProvider`，直接当作 `*Har` 用会编译失败。需要完整 API 时显式转换。
- **混淆结构体式与函数式选项**：`Filter(FilterOptions{...})` 与 `FilterWith(WithFilter*...)` 等价但不可混用参数；同一调用里别把 `FilterOptions` 字段和 `WithFilter*` 选项交叉传。
- **克隆语义被打破**：在 `*Har` 变换方法里就地修改 receiver 会破坏链式安全。除 `InPlace` 后缀方法外，一律克隆后修改。
- **测试依赖特定 testdata**：测试应基于 `testdata/` 中已提交的 HAR 文件，不要依赖外部网络或临时生成的文件。

## 新增示例

`examples/` 下的每个示例都是独立的 `package main`。新增示例的步骤：

1. 在 `examples/` 下新建子目录（如 `examples/audit/`），添加 `main.go`。
2. `main.go` 顶部 `package main`，`import` SDK 为 `har "github.com/cyberspacesec/har-skills"`。
3. 优先使用 `testdata/` 中的真实 HAR 文件作为默认输入，避免依赖外部网络。
4. 提供命令行参数（用标准库 `flag`）让用户指定文件路径与输出格式。
5. 在本站 [示例代码集](/zh/examples/) 页面与 `examples/` 的目录结构中登记新示例。
6. 确保示例可直接 `go run main.go` 运行，不引入新的运行时依赖。

示例应聚焦单一场景，代码量控制在单文件可读范围内；若场景复杂，拆分为多个示例而非堆叠在一个 `main.go` 中。

感谢你的贡献！
