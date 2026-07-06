---
title: 全局参数
titleTemplate: false
---

# 全局参数

所有 `har` 子命令共享一组 persistent flag，由根命令统一注册、经 Viper 绑定。本文集中说明这些全局 flag 的取值、配置覆盖优先级、stdin 管道用法，以及它们背后的内部加载与输出架构。文末附 24 个命令的一览表，便于从分级视角快速定位。

所有示例都可在仓库根目录直接运行，使用 `testdata/example.har` 或 `testdata/full.har`。

## 全局 persistent flags

| Flag | 简写 | 默认值 | 含义 |
|------|------|--------|------|
| `--file` | `-f` | 空 | HAR 文件路径，`-` 表示从 stdin 读取 |
| `--format` | 无 | `text` | 输出格式：`text` / `json` / `csv` / `yaml` |
| `--output` | `-o` | 空 | 输出文件路径（为空则写 stdout） |
| `--no-header` | 无 | `false` | 隐藏 `text`/`csv` 输出的表头 |
| `--config` | 无 | 空 | 配置文件路径，默认 `$HOME/.har.yaml` |

`--file` 是绝大多数命令的输入源；但 `diff`、`merge` 这类多文件命令走位置参数、不走 `-f`（见各自页面）。`--format` 控制序列化形态：`text` 走各命令自带的格式化函数（多为表格或分节文本），`json` 走 `MarshalIndent`，`csv` 走专用 csv 函数（缺省回退到 JSON），`yaml` 优先调用数据的 `ToYAML()`，否则回退到缩进 JSON。

## Viper 集成

CLI 用 `spf13/viper` 把 flag、环境变量、配置文件三者打通，遵循 Viper 的覆盖优先级（高 → 低）：

1. **命令行 flag**（最高）
2. **环境变量**
3. **配置文件**
4. **flag 默认值**（最低）

注册时调 `viper.BindPFlag` 把 `file` / `format` / `output` / `no-header` 绑到对应 flag；初始化时 `viper.SetEnvPrefix("HAR")` + `viper.AutomaticEnv()` 让下列环境变量生效：

| 环境变量 | 对应 flag |
|----------|-----------|
| `HAR_FILE` | `--file` |
| `HAR_FORMAT` | `--format` |
| `HAR_OUTPUT` | `--output` |

::: tip 配置文件查找顺序
未显式传 `--config` 时，Viper 依次在 `$HOME` 与当前目录查找名为 `.har.yaml` 的配置文件，命中即读入。可在 `~/.har.yaml` 里写长期偏好，例如 `format: json`，每次跑命令默认输出 JSON。
:::

## stdin 支持

`--file` 留空且 stdin 有数据时，加载器会自动从 stdin 读取——便于把 `curl`、`cat`、管道上游工具的输出直接喂给 `har`：

```bash
cat capture.har | har info
```

显式用 `-` 也能强制走 stdin，常用于 stdin 看起来像终端（被误判）的边缘场景：

```bash
har -f - info < capture.har
```

stdin 路径走 `har.ParseHar(data)`（已读入内存的字节），文件路径走 `har.ParseHarFileAuto(path)`（自动检测 gzip）。两者最终都得到 `*har.Har`。

## 内部架构

每个子命令的执行路径高度一致，可拆成「加载 → 调 SDK → 分发输出」三步：

```
internal.LoadHar(cmd, args)          → *har.Har   （gzip 自动检测 / stdin 回退）
        │
        ▼
SDK 调用（Statistics / FilterWith / SecurityAudit / ...）
        │
        ▼
internal.WriteOutput(cmd, data, textFunc, csvFunc)   → stdout / 文件
```

**加载层** `internal.LoadHar`：读 `--file`；为空或 `-` 时走 stdin；为空且 stdin 无数据时直接报错退出。`diff`/`merge` 等多文件命令改用 `internal.LoadHarFromArg(arg)`，对每个位置参数独立加载并各自报错。

**输出层** `internal.WriteOutput`：按 `--format` 选择序列化分支，最终落到 `WriteToFileOrStdout(path, data)`。两条关键的 stdout/stderr 约定：

- **数据走 stdout**——便于接入管道（`| jq`、`| grep`、`> file`）。
- **进度与提示走 stderr**——例如 `已写入 1234 字节到 report.json`，避免污染数据流。

因此 `har -f t.har info --format json | jq '.totalRequests'` 能干净地取到字段值，而 `已写入...` 这类提示不会混进管道。

`--no-header` 由 `internal.NoHeader(cmd)` 读取，仅在 `text`/`csv` 表格输出里生效（抑制 `INDEX METHOD STATUS ...` 这类表头行）；JSON/YAML 不受影响。

## 命令一览表

24 个命令按分级（Level 1–5）编排，分级越高能力越深：

### 基础操作（Level 1）

| 命令 | 用途 | 分级 |
|------|------|------|
| `info` | 显示 HAR 概要统计 | 基础操作 |
| `list` | 列出条目 | 基础操作 |
| `find` | 多维搜索条目 | 基础操作 |
| `headers` | 显示请求/响应头部 | 基础操作 |
| `timing` | 计时分解 | 基础操作 |
| `extract` | 提取响应内容 | 基础操作 |

### 文件操作（Level 2）

| 命令 | 用途 | 分级 |
|------|------|------|
| `diff` | 对比两个 HAR | 文件操作 |
| `merge` | 合并多 HAR | 文件操作 |
| `split` | 拆分 HAR | 文件操作 |
| `validate` | 验证 HAR 规范 | 文件操作 |

### 安全隐私（Level 3）

| 命令 | 用途 | 分级 |
|------|------|------|
| `security` | 安全审计 | 安全隐私 |
| `redact` | 脱敏 | 安全隐私 |

### 深度分析（Level 4）

| 命令 | 用途 | 分级 |
|------|------|------|
| `performance` | 性能评分 | 深度分析 |
| `cookie` | Cookie 审计 | 深度分析 |
| `cache` | 缓存分析 | 深度分析 |
| `index` | 索引查询 | 深度分析 |
| `domains` | 域名统计 | 深度分析 |
| `content` | 内容类型 | 深度分析 |
| `connections` | 连接复用 | 深度分析 |
| `waterfall` | 瀑布流 | 深度分析 |

### 转换导出（Level 5）

| 命令 | 用途 | 分级 |
|------|------|------|
| `transform` | 转换请求 | 转换导出 |
| `export` | 导出格式 | 转换导出 |
| `dedup` | 去重 | 转换导出 |
| `replay` | 重放请求 | 转换导出 |

各命令详解按分级组织：基础操作见 [基础操作](./basic.md)，文件操作见 [文件操作](./files.md)，安全隐私见 [安全与隐私](./security.md)，深度分析见 [深度分析](./analysis.md)，转换导出见 [转换与导出](./transform.md)。
