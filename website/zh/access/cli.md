---
title: CLI 接入
---

# CLI 接入

`har` 是一个单文件命令行工具，把 HAR Skills SDK 的全部能力暴露给终端。零运行时依赖、JSON 优先、可管道组合，适合脚本、CI 与交互式排查。

## CLI 概览

- **24 个 Cobra 子命令**，覆盖 HAR 分析全生命周期。
- **统一入参**：`-f/--file` 指定 HAR 文件，`-` 表示从 stdin 读取。
- **统一出参**：`--format` 选 `text`（默认）/`json`/`csv`/`yaml`，`-o` 写文件。
- **环境变量**：`HAR_FILE`、`HAR_FORMAT`、`HAR_OUTPUT`（经 Viper 读取）。

安装：

```bash
go install github.com/cyberspacesec/har-skills/cmd/har@latest
# 或从 https://github.com/cyberspacesec/har-skills/releases/latest 下载预编译二进制
```

全局参数的完整说明见 [全局参数](../cli/global-flags.md)。

## 命令分级

`har` 把 24 个命令分成 5 级，从最常用到最高级：

| 级别 | 主题 | 命令 | 详见 |
|------|------|------|------|
| Level 1 | 基础操作 | `info` `list` `find` `headers` `timing` `extract` | [基础操作](../cli/basic.md) |
| Level 2 | 文件操作 | `diff` `merge` `split` `validate` | [文件操作](../cli/files.md) |
| Level 3 | 安全与隐私 | `security` `redact` | [安全与隐私](../cli/security.md) |
| Level 4 | 深度分析 | `performance` `cookie` `cache` `index` `domains` `content` `connections` `waterfall` | [深度分析](../cli/analysis.md) |
| Level 5 | 转换与导出 | `transform` `export` `dedup` `replay` | [转换与导出](../cli/transform.md) |

::: tip 分级不是难度门槛
分级只反映「使用频率与概念纵深」。Level 1 的 `find --errors` 一样能解决大问题；Level 5 的 `replay` 也能写进一行脚本。按任务挑命令即可。
:::

## 典型管道用法

CLI 的 `--format json` 输出天然适合 `jq` 二次加工：

```bash
# 找出所有 4xx/5xx，提取 URL 与状态码
har -f capture.har find --errors --format json \
  | jq '.entries[] | {url: .request.url, status: .response.status}'

# 按域统计请求数，降序取前 10
har -f capture.har domains --format json \
  | jq 'to_entries | sort_by(-.value.count) | .[0:10]'

# 安全审计只看 HIGH
har -f capture.har security --format json \
  | jq '.findings[] | select(.severity=="high")'
```

一条管道完成「过滤 → 投影 → 排序」，无需写中间文件。

## stdin 管道

`-f -` 或省略 `-f`（配合 stdin）即可从管道读 HAR：

```bash
cat capture.har | har info
curl -sL https://example.com/capture.har | har -f - find --slow 1000
zcat capture.har.gz | har -f - info --format json
```

::: tip 自动解压
`ParseHarFileAuto` 会按扩展名与 gzip magic bytes 自动判断压缩格式，CLI 读取时同样透明支持 `.har.gz`。
:::

## 与脚本集成

### bash 循环批处理

```bash
#!/usr/bin/env bash
set -euo pipefail

for f in captures/*.har; do
  echo "=== $f ==="
  har -f "$f" security --severity high --format json \
    | jq '.score, (.findings | length)'
done
```

### cron 定时审计

```bash
# 每天凌晨审计当天抓包，分数低于 60 就告警
0 2 * * * har -f /data/$(date +\%F).har security --format json \
  | jq -e '.score >= 60' >/dev/null \
  || /usr/local/bin/notify-slack "HAR 安全分数告警：$(date +\%F)"
```

### CI 集成

```bash
# 回归测试：对比 staging 与 prod 抓包，差异超阈值则失败
har diff staging.har prod.har --compare-by-url --ignore-timings -o diff.txt
test ! -s diff.txt || { echo "API 行为有差异"; cat diff.txt; exit 1; }
```

## 输出格式速查

| `--format` | 适用 | 特点 |
|------------|------|------|
| `text` | 人看、终端 | 默认；表格带表头，`--no-header` 可关 |
| `json` | Agent / jq / 程序 | 完整结构，适合管道二次加工 |
| `csv` | Excel / 表格工具 | 适合 `list`、`timing` 等表格型命令 |
| `yaml` | 配置 / 评审 | 可读性好，适合 `info`、`security` |

## 下一步

- [全局参数](../cli/global-flags.md) —— 所有命令共享的 flag 详解
- [AI Agent Skill 接入](./skill.md) —— 让 Agent 直接驱动 CLI
- [Go SDK 接入](./sdk.md) —— 需要嵌入程序时改用 SDK
- [MCP 封装](./mcp.md) —— 把 CLI 包装成 MCP tools
