---
title: 安装详解
titleTemplate: false
---

# 安装详解

HAR Skills 是一个单独的 `har` 二进制，无运行时依赖。下面给出三种安装方式、平台矩阵、验证、升级卸载以及环境变量配置。

## 方式一：预编译二进制（推荐）

最省事：从 GitHub Releases 下载对应平台的压缩包，解压后放入 `PATH`。

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
# 下载 har-skills_0.1.0_windows_x86_64.zip
# 解压得到 har.exe，加入 PATH
```

:::

::: tip macOS 首次运行
从网络下载的二进制在 macOS 上首次运行可能被 Gatekeeper 拦截。在"系统设置 → 隐私与安全性"里点"仍要打开"放行即可，或执行：

```bash
xattr -d com.apple.quarantine /usr/local/bin/har
```
:::

### 平台矩阵

预编译二进制覆盖以下平台：

| OS | 架构 |
| --- | --- |
| linux | x86_64 / arm64 / armv6 / armv7 / i386 |
| darwin (macOS) | x86_64 / arm64 |
| windows | x86_64 / i386 |
| freebsd | x86_64 / i386 |

下载文件名格式为 `har-skills_<version>_<os>_<arch>.tar.gz`（Windows 为 `.zip`）。到 [Releases 页面](https://github.com/cyberspacesec/har-skills/releases/latest) 按需选取。

## 方式二：go install

若已装 Go 1.19+，一行命令即可：

```bash
go install github.com/cyberspacesec/har-skills/cmd/har@latest
```

二进制会装到 `$GOPATH/bin`（或 `$GOBIN`），请确保该目录在 `PATH` 中：

```bash
# 确认 GOPATH/bin
go env GOPATH
export PATH="$(go env GOPATH)/bin:$PATH"
```

::: warning 版本钉制
生产环境建议把 `@latest` 换成具体版本号（如 `@v0.1.0`），避免引入未预期的新行为。
:::

## 方式三：源码构建

适合需要自定义构建标签、注入版本信息或交叉编译的场景。

```bash
git clone https://github.com/cyberspacesec/har-skills.git
cd har-skills
go build -o har ./cmd/har/
```

### 注入版本信息

用 `-ldflags` 把 git tag 注入二进制，`har --version` 即可显示：

```bash
go build -ldflags "-X github.com/cyberspacesec/har-skills/cmd/har/cmd.version=$(git describe --tags 2>/dev/null || echo dev)" -o har ./cmd/har/
```

### 交叉编译

设置 `GOOS` / `GOARCH` 即可产出其它平台二进制：

```bash
# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o har-darwin-arm64 ./cmd/har/

# Windows
GOOS=windows GOARCH=amd64 go build -o har.exe ./cmd/har/

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o har-linux-arm64 ./cmd/har/
```

## 验证安装

任意一条命令能跑通即说明安装成功：

```bash
har --version
```

或对一个真实 HAR 跑 `info`：

```bash
har -f testdata/example.har info
```

## 升级与卸载

### 升级

| 安装方式 | 升级命令 |
| --- | --- |
| 预编译二进制 | 重新下载新版本压缩包并覆盖 `har` |
| go install | `go install github.com/cyberspacesec/har-skills/cmd/har@latest` |
| 源码构建 | `git pull && go build -o har ./cmd/har/` |

### 卸载

删除二进制即可，HAR Skills 不在系统留下任何配置或数据目录：

```bash
rm -f $(which har)
# 或手动删除 /usr/local/bin/har、$GOPATH/bin/har 等
```

可选地清理环境变量（见下文）。

## 环境变量

CLI 通过 Viper 读取三个环境变量，等价于对应命令行参数，便于在 CI/容器里固定行为：

| 环境变量 | 等价参数 | 用途 |
| --- | --- | --- |
| `HAR_FILE` | `-f` / `--file` | 默认输入文件路径 |
| `HAR_FORMAT` | `--format` | 默认输出格式（text/json/csv/yaml） |
| `HAR_OUTPUT` | `-o` / `--output` | 默认输出文件路径 |

示例：

```bash
export HAR_FILE=capture.har
export HAR_FORMAT=json
har info          # 等价于 har -f capture.har info --format json
```

::: tip 优先级
命令行参数优先于环境变量；环境变量优先于默认值。显式传参会覆盖环境变量设定。
:::

此外还支持 `--config` 指定配置文件（默认 `$HOME/.har.yaml`），可把常用参数固化下来。

## 下一步

- 跑通第一条命令：[快速开始](./quick-start.md)
- 理解输入格式：[HAR 格式入门](./har-basics.md)
- 全部命令：[CLI 命令参考](./cli/global-flags.md)
