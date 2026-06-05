<div align="center">

<img src="./docs/public/logo.svg" width="200" alt="UniRTM Logo" />

<h1 align="center">
    UniRTM
</h1>

<p>
  <a href="https://github.com/snowdreamtech/UniRTM/blob/main/LICENSE"><img alt="GitHub" src="https://img.shields.io/github/license/snowdreamtech/UniRTM?style=for-the-badge&color=6B7F4E"></a>
  <a href="https://github.com/snowdreamtech/UniRTM/actions/workflows/pages.yml"><img alt="GitHub Workflow Status" src="https://img.shields.io/github/actions/workflow/status/snowdreamtech/UniRTM/pages.yml?style=for-the-badge&color=C5975B"></a>
  <a href="https://github.com/snowdreamtech/UniRTM/actions/workflows/ci.yml"><img alt="Continuous Integration" src="https://github.com/snowdreamtech/UniRTM/actions/workflows/ci.yml/badge.svg"></a>
  <a href="https://github.com/snowdreamtech/UniRTM/actions/workflows/cd.yml"><img alt="Continuous Delivery" src="https://github.com/snowdreamtech/UniRTM/actions/workflows/cd.yml/badge.svg"></a>
</p>

<p><b>集开发工具、环境变量和任务管理于一体的 CLI（内置安全扫描）。</b><br><i>灵感来源于出色的 <a href="https://github.com/jdx/mise">mise</a> 项目，在此向其致敬。</i></p>

<p align="center">
  <a href="https://snowdreamtech.github.io/UniRTM/zh/guide/getting-started.html">快速开始</a> •
  <a href="https://snowdreamtech.github.io/UniRTM/zh/">官方文档</a> •
  <a href="https://snowdreamtech.github.io/UniRTM/zh/dev-tools/overview.html">开发工具</a> •
  <a href="https://snowdreamtech.github.io/UniRTM/zh/environments/overview.html">环境变量</a> •
  <a href="https://snowdreamtech.github.io/UniRTM/zh/tasks/overview.html">任务系统</a>
</p>

<hr />

</div>

> [!TIP]
> UniRTM 强大的任务编排能力让你能够极其轻松地将 Trivy、Syft 等安全工具无缝集成到你的日常构建和部署流程中。

## 介绍

`UniRTM` (Universal Runtime Manager) 能够在每次执行命令前自动准备好你的开发环境。它将项目所需的工具版本、环境变量和常用任务统一集中在 `.unirtm.toml` 文件中进行管理，确保每次打开新终端、切换分支或运行 CI 任务时，环境配置都绝对一致。

- 安装并在诸如 node, python, go 等各类 **开发工具** 之间无缝切换。
- 基于不同目录加载隔离的 **环境变量**，支持读取 `.env` 以及 SOPS 加密数据。
- 编写并运行项目的构建、测试、代码检查及部署 **任务**。

虽然在概念上我们深受伟大的 `mise` 启发（在此致敬），但 UniRTM 在架构上引入了几个独特的选择：

- **纯 Go 引擎**: 底层完全由 Go 语言编写，利用 goroutines 实现了极致的并发下载能力。
- **轻量级 Go 垫片 (Lightweight Shims)**: 彻底抛弃了缓慢的 Bash 垫片。所有的命令均由高性能的单一 Go 引擎拦截并极速路由，既避免了环境变量（PATH）无限制爆炸，又保证了极致的执行性能。
- **任务无缝集成安全工具**: 虽然保持了引擎本身的极致精简，但你能在项目中利用 UniRTM 的任务系统，丝滑地串联 Trivy、Syft 等安全漏洞扫描工作流。
- **强制版本锁定**: 自动生成 `unirtm.lock` 锁文件，不仅锁定版本号，还精确锁定所有下载包的校验和，从而保障团队环境的绝对可复现。

## 支持的操作系统 (Supported Platforms)

*完全支持 macOS (Apple Silicon / Intel)、Linux (glibc & musl/Alpine) 以及 Windows 平台。*

## 演示 (Demo)

以下演示展示了如何使用 `UniRTM` 全局安装指定版本的 `go`。
请注意安装的速度以及内置的安全漏洞扫描功能！

[![demo](https://raw.githubusercontent.com/snowdreamtech/UniRTM/main/docs/tapes/demo.gif)](https://github.com/snowdreamtech/UniRTM/blob/main/docs/tapes/demo.mp4)

## 快速入门

### 安装 UniRTM

你可以通过多种方式安装 UniRTM，详见 [快速开始](https://snowdreamtech.github.io/UniRTM/zh/guide/getting-started.html)。

```sh-session
$ curl -sL https://raw.githubusercontent.com/snowdreamtech/UniRTM/main/install.sh | bash
$ ~/.local/bin/unirtm --version
UniRTM v0.1.0 macos-arm64 (2026-05-28)
```

将 UniRTM 挂载到你的 Shell 中（请选择与你对应的 Shell）：

```sh-session
# 假设 unirtm 被安装在默认的 ~/.local/bin/unirtm 下
echo 'eval "$(~/.local/bin/unirtm env)"' >> ~/.bashrc
echo 'eval "$(~/.local/bin/unirtm env)"' >> ~/.zshrc
echo '~/.local/bin/unirtm env | source' >> ~/.config/fish/config.fish
```

### 单次执行指定版本的工具

```sh-session
$ unirtm exec node@20 -- node -v
unirtm node@20.x.x ✓ installed
v20.x.x
```

### 全局安装开发工具

```sh-session
$ unirtm use --global node@22 go@1.22
$ node -v
v22.x.x
$ go version
go version go1.22.x macos/arm64
```

查看 [开发工具指南](https://snowdreamtech.github.io/UniRTM/zh/dev-tools/) 获取更多示例。

### 管理环境变量

```toml
# .unirtm.toml
[env]
SOME_VAR = "foo"
```

```sh-session
$ unirtm set SOME_VAR=bar
$ echo $SOME_VAR
bar
```

此外，`UniRTM` 同样可以自动读取本地的 [`.env` 文件](https://snowdreamtech.github.io/UniRTM/zh/environments/#env-directives)。

### 运行任务

```toml
# .unirtm.toml
[tasks.build]
description = "编译项目"
run = "echo building..."
```

```sh-session
$ unirtm run build
building...
```

查看 [任务系统指南](https://snowdreamtech.github.io/UniRTM/zh/tasks/) 获取更多高级用法。

### UniRTM 综合实战配置

下面是一个综合的 `.unirtm.toml` 示例，展示了如何在一个项目中同时管理开发工具、环境变量，并使用内置的安全扫描执行高级的部署任务编排：

```toml
# .unirtm.toml
[tools]
terraform = "1"
aws-cli = "2"
node = "20"

[env]
TF_WORKSPACE = "development"
AWS_REGION = "us-west-2"
NODE_ENV = "production"

[tasks.plan]
description = "运行带有工作区配置的 terraform plan"
run = """
terraform init
terraform workspace select $TF_WORKSPACE
terraform plan
"""

[tasks.validate]
description = "验证 AWS 凭据与 terraform 配置"
run = """
aws sts get-caller-identity
terraform validate
"""

[tasks.audit]
description = "执行深度的底层安全扫描"
run = """
trivy fs --format cyclonedx --output sbom.json .
gitleaks detect --source . --no-banner
"""

[tasks.deploy]
description = "在验证和安全扫描后正式部署基础设施"
depends = ["validate", "audit", "plan"]
run = "terraform apply -auto-approve"
```

你可以这样运行：

```sh-session
unirtm install # 安装所需要的所有开发工具
unirtm run deploy # 在部署前，将自动按顺序并行执行校验和安全审查任务
```

## 官方文档

完整的架构解析与高级配置指南请前往官网：[snowdreamtech.github.io/UniRTM](https://snowdreamtech.github.io/UniRTM/zh/)

### 多维度深度对比 (Architecture & Environments)

为了满足现代企业级和高并发容器化环境的严苛要求，`UniRTM` 在底层架构上做出了极具针对性的设计抉择。以下是我们在多个关键维度上与优秀的生态前辈 (`mise`, `asdf`) 的深度对比：

#### 1. 核心架构与执行路径

| 对比维度 | `asdf` (Bash) | `mise` (Rust) | `UniRTM` (Go) | 核心优势与意义 |
| :--- | :--- | :--- | :--- | :--- |
| **命令执行路径** | Bash 垫片 | Rust 垫片 / PATH | **轻量级 Go 垫片** | 使用 Go 编译的单一入口代理，彻底解决 Bash 垫片的性能低下问题，且完美解决跨平台兼容性。 |
| **并发下载模型** | 无 | 操作系统线程 | **原生 Goroutines** | 依托 Go 语言极其轻量的协程机制，在海量工具链下载更新时实现极致的并发吞吐量。 |
| **配置层级** | `.tool-versions` | `mise.toml` | **`.unirtm.toml`** | 使用标准化的 TOML 文件统一管理项目的工具、环境和任务。 |

#### 2. 环境变量与上下文管理

| 功能特性 | `asdf` | `mise` | `UniRTM` | 详细说明 |
| :--- | :--- | :--- | :--- | :--- |
| **统一管理范畴** | 仅开发工具 | 工具+环境+任务 | **工具+环境+任务** | 根据当前所在目录，自动且无缝地切换整体上下文环境。 |
| **`.env` 文件解析** | ❌ | 原生支持 | **原生支持** | 告别额外的环境变量加载器，内置完美解析传统 `.env` 文件。 |
| **密钥加密管理** | ❌ | 插件集成体系 | **内置 SOPS 原生支持** | 将基于 SOPS 的加密环境变量作为一等公民，提供开箱即用的敏感信息保护。 |

#### 3. 跨平台兼容与系统韧性

| 功能特性 | `asdf` | `mise` | `UniRTM` | 详细说明 |
| :--- | :--- | :--- | :--- | :--- |
| **Windows 原生支持** | 依赖 WSL/MSYS | 支持 | **自底向上原生设计** | 从设计之初就兼顾 Windows 与 Cygdrive，提供极其丝滑的原生跨平台体验。 |
| **Alpine / Musl 兼容** | 部分支持 | 支持 | **硬核底层兼容** | 完美无缝运行在没有任何 `glibc` 依赖的极简 Alpine 容器环境之中。 |
| **版本与校验锁定** | 仅版本号 | `mise.lock` (支持) | **`unirtm.lock` (默认)** | 锁定版本的同时严格校验所有底层文件的 Hash，保障团队环境的一致性。 |

#### 4. 生态亲和度与极简主义

| 功能特性 | `asdf` | `mise` | `UniRTM` | 详细说明 |
| :--- | :--- | :--- | :--- | :--- |
| **混合路径解析** | ❌ | 部分支持 | **深度支持 (Cygdrive)** | 对 Windows Git Bash / MSYS2 的原生路径转换进行了专项优化。 |
| **外部依赖要求** | 依赖 Bash 生态 | 极少 | **绝对零依赖** | 核心插件直接编译入独立二进制，丢进任意极简系统开箱即用。 |
| **垫片开销** | 常规 (Bash 脚本) | 已优化 (Rust 二进制) | **极速 (Go 二进制)** | 所有工具软链回 unirtm 引擎进行毫秒级极速路由，彻底杜绝 PATH 爆炸问题。 |
| **DevOps 集成** | 自定义脚本 | 良好 | **原生级无缝集成** | Go 语言血统，天生匹配云原生基础设施，极其适合企业内部自研平台的二次集成。 |


## 社区支持与 Issues

如果你遇到任何 bug 或有新功能提议，欢迎在 GitHub 提交：

- [Discussions (讨论区)](https://github.com/snowdreamtech/UniRTM/discussions)
- [Issues (问题追踪)](https://github.com/snowdreamtech/UniRTM/issues)

## 致谢鸣谢

<p>
  本项目在架构设计和开发者体验上深受伟大的 <a href="https://github.com/jdx/mise">mise</a> 的启发。
</p>

## 开源协议

本项目基于 MIT 许可证开源。版权所有 (c) 2026-至今 SnowdreamTech Inc。
