---
title: AI Agent Skill 接入
---

# AI Agent Skill 接入

HAR Skills 项目从第一天起就按 **AI Agent Skill** 的形态设计：项目根目录的 `CLAUDE.md` 是一份渐进式披露文档（progressive disclosure），AI Agent 读取后即可理解全部能力并调用。本文说明如何让一个 LLM Agent（如 Claude Code）快速接入。

## 什么是 Skill 接入

Skill 接入 = 把一份「给 LLM 看的说明书」交给 Agent，让它自主完成 HAR 分析任务。

- **载体**：仓库根目录的 [`CLAUDE.md`](https://github.com/cyberspacesec/har-skills/blob/main/CLAUDE.md)，约 400 行，覆盖 CLI 全部 24 个命令与 SDK 70+ 方法。
- **结构**：渐进式披露——从 Quick Start 到 5 级命令参考，再到 SDK 速查与工作流，Agent 按需向下钻取，不必一次读完。
- **结果**：Agent 读完即知道「有哪些命令、每条命令的 flag、输出格式、典型组合」，无需反复试错。

::: tip 为什么不直接给文档站链接
文档站面向人类阅读，含大量排版与导航。`CLAUDE.md` 是纯文本、结构化、专门为 token 预算优化的版本，Agent 一次加载即可获得完整能力地图。两者内容同源，但 Skill 文档更紧凑。
:::

## 工作流

```
Agent 读取 CLAUDE.md
        │
        ▼
安装 har 二进制（go install 或下载 release）
        │
        ▼
用 CLI 命令完成任务（info / find / security / redact …）
        │
        ▼
输出 JSON 便于 Agent 解析与后续推理
```

每一步都可由 Agent 自主完成：读文档是上下文加载，安装是一次 shell 调用，分析与输出是 CLI 命令组合，最后把 JSON 结果纳入下一轮推理。

### 分步解析

1. **读 `CLAUDE.md`**：Agent 把这份渐进式文档加载进上下文，获得「24 命令 + 70+ SDK 方法 + 4 个工作流」的能力地图。文档结构刻意按使用频率排序，Agent 能快速跳到所需章节。
2. **装二进制**：Agent 选 `go install`（需要 Go 工具链）或下载预编译 release（零依赖）。后者只需 `curl + tar` 两步，适合纯运行时环境。
3. **跑命令**：Agent 根据用户目标选命令——安全审计走 `security`，性能走 `performance`，对比走 `diff`。每条命令都接受 `-f` 指定文件、`--format json` 出结构化结果。
4. **解析 JSON**：Agent 把命令的 JSON 输出纳入下一轮推理，决定是否追加命令（如 `redact` 脱敏后再 `validate`）。无需人写 jq，模型自己读 JSON。

### 能力地图速览

`CLAUDE.md` 给 Agent 提供的能力大致分四类，Agent 按任务对号入座：

| 任务类别 | 关键命令 | 典型输出 |
|----------|----------|----------|
| 概览与定位 | `info` `list` `find` | 条目数、状态码分布、错误/慢请求清单 |
| 安全与脱敏 | `security` `redact` `cookie` | 安全评分、findings、脱敏后的新 HAR |
| 性能与缓存 | `performance` `timing` `cache` `waterfall` | 评分等级、关键路径、可缓存性 |
| 转换与导出 | `transform` `export` `dedup` `replay` | curl/wget/Postman、去重文件、回放结果 |

## 一键引导提示词（One-Prompt Agent Bootstrap）

把下面这段提示词直接发给 Agent，它即可装好二进制并开始分析 HAR：

```text
你是一个 HAR 文件分析助手，使用 har-skills 工具完成任务。

第一步，安装 CLI（二选一）：
  go install github.com/cyberspacesec/har-skills/cmd/har@latest
或从 https://github.com/cyberspacesec/har-skills/releases/latest 下载对应平台的 tar.gz，解压后把 har 放进 PATH。

第二步，阅读项目根目录的 CLAUDE.md（或在线版本），掌握 24 个 CLI 命令、全局 flag
（-f/--file、--format、-o）与典型工作流。遇到 HAR 分析任务时优先用 CLI，
并通过 --format json 拿到机器可读结果。

第三步，分析用户提供的 HAR 文件：
  har -f <file> info --format json          # 总览
  har -f <file> find --errors --format json # 抓错误
  har -f <file> security --format json      # 安全审计
  har -f <file> performance --format json   # 性能打分

约定：
- 需要分享结果前先脱敏：har -f <file> redact --redact-ips -o redacted.har
- 大文件先拆分：har -f <file> split --by domain -o by-domain
- 始终用 JSON 输出以便你解析；面向人类时再用 text/table。

收到 HAR 文件后，先跑 info，再根据用户目标选择后续命令。
```

::: tip 可裁剪
如果你的 Agent 已经在项目仓库里工作，直接让它读本地 `CLAUDE.md` 即可，无需在线版本。提示词中的命令清单可按场景删减（例如只做安全审计时保留 security + redact）。
:::

## 为什么适合 Agent

| 特性 | 对 Agent 的价值 |
|------|----------------|
| CLI 输出 `--format json` | 机器可读，Agent 可直接解析并纳入推理 |
| 命令组合表达力强 | `find --slow 1000` → `extract` → `export curl` 一条管道完成多步 |
| Skill 文档本身面向 LLM | 一次加载即获得完整能力地图，省 token、省往返 |
| 24 命令覆盖全生命周期 | 解析、过滤、安全、性能、转换、导出、回放，闭环 |
| 零运行时依赖（SDK 侧） | 二进制单文件，部署无摩擦 |

## 适合的任务

- **安全审计**：`security` 出报告 → `redact` 脱敏 → `validate --strict` 校验，适合「检查这份抓包有没有泄密」类需求。
- **性能分析**：`performance` 打分 → `find --slow` 找慢请求 → `waterfall --critical-path` 看关键路径，适合「为什么这页慢」类需求。
- **API 行为对比**：`diff a.har b.har --include-body` 对比两份抓包，适合回归测试与接口迁移。
- **数据脱敏后分享**：`dedup --remove` 去重 → `redact --redact-ips` 脱敏 → `split --by domain` 拆分，适合把抓包安全地交给第三方。

## Skill 文档 vs 传统 README

很多项目有 README，但不是所有 README 都适合 Agent。HAR Skills 的 `CLAUDE.md` 专门针对 LLM 做了优化：

| 维度 | 传统 README | `CLAUDE.md`（Skill 文档） |
|------|-------------|---------------------------|
| 读者 | 人类开发者 | LLM Agent |
| 结构 | 营销 + 安装 + 简介 | 渐进式能力披露，按使用频率排序 |
| 命令描述 | 挑几个常用 | 全量 24 命令，含 flag 与输出格式 |
| 输出格式 | 多为文本示例 | 明确标注 `--format json` 可机器读 |
| 工作流 | 偶有 | 4 个端到端工作流，可直接照搬 |

正因如此，Agent 读完 `CLAUDE.md` 即可自主完成任务，而非反复试错或追问用户「这个工具能做什么」。

## 下一步

- [CLI 接入](./cli.md) —— 24 个命令的分级速查
- [Go SDK 接入](./sdk.md) —— 嵌入到 Go 程序
- [MCP 封装](./mcp.md) —— 让 Claude Desktop 直接调用
- [安全审计工作流](../workflows/security-audit.md) —— 端到端示例
