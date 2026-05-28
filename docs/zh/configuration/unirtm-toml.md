# `.unirtm.toml` 详解

`.unirtm.toml` 文件是任何由 UniRTM 管理的项目的核心。在这里，你可以声明开发工具、环境变量和项目任务。

这个文件作为你项目开发环境的“唯一事实来源（Single Source of Truth）”。它可以完全取代 `.nvmrc`、`.python-version`、`Makefile` 和类似 `direnv` 的 `.env` 加载器。

## 配置示例

```toml
[env]
# 设置标准的环境变量
NODE_ENV = "development"
PORT = "3000"
# 环境变量也可以动态执行 Shell 命令
GIT_HASH = { run = "git rev-parse --short HEAD" }

[tools]
# 锁定精确版本
node = "20.11.1"
go = "1.22.0"
# 或者使用语义化前缀
python = "3.11"
# 使用特定后端（比如直接拉取 GitHub Releases）
"github:aquasecurity/trivy" = "0.49.0"

[tasks.build]
description = "编译应用程序"
run = "go build -o bin/app ./cmd/main.go"

[tasks.test]
description = "运行带有覆盖率的测试"
depends = ["build"]
run = "go test -cover ./..."

[settings]
# 仅针对此项目覆盖全局设置
legacy_version_file = true
```

## 模块解析

### `[env]`

定义当你 `cd` 进入该目录时，自动注入到你的 Shell 中的环境变量。

- 支持静态字符串 (`KEY = "value"`)。
- 支持动态执行命令取值 (`KEY = { run = "command" }`)。

### `[tools]`

列出在该项目上工作所需的工具依赖。
UniRTM 支持映射到内部插件的标准简写（如 `node`、`go`）。它还原生支持使用 `github:org/repo` 语法直接从 GitHub 提取原生二进制文件。

### `[tasks.*]`

定义任务运行器的配置。
任务可以配置依赖项 (`depends`)、别名以及特定的环境变量覆盖。请参阅 [任务运行器概览](../tasks/overview.md) 了解更多细节。

### `[settings]`

允许你在项目级别覆盖定义在 `~/.config/unirtm/config.toml` 中的全局 UniRTM 设置。

## 加载顺序与层级

UniRTM 会向上遍历目录树查找 `.unirtm.toml` 文件。这意味着你可以有一个全局的 `~/.unirtm.toml` 用于管理你的个人通用工具，同时在 `~/projects/my-app/.unirtm.toml` 中定义一个局部的配置，用于覆盖该项目的特定设置。
