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
  <a href="https://unirtm.snowdream.tech/zh/guide/getting-started.html">快速开始</a> •
  <a href="https://unirtm.snowdream.tech/zh/">官方文档</a> •
  <a href="https://unirtm.snowdream.tech/zh/dev-tools/overview.html">开发工具</a> •
  <a href="https://unirtm.snowdream.tech/zh/environments/overview.html">环境变量</a> •
  <a href="https://unirtm.snowdream.tech/zh/tasks/overview.html">任务系统</a>
</p>

<hr />

</div>

> [!TIP]
> UniRTM 将 SBOM 自动生成与 Trivy/Syft 安全漏洞扫描直接无缝集成到了你的工具安装流程中！

## 介绍

`UniRTM` (Universal Runtime Manager) 能够在每次执行命令前自动准备好你的开发环境。它将项目所需的工具版本、环境变量和常用任务统一集中在 `.unirtm.toml` 文件中进行管理，确保每次打开新终端、切换分支或运行 CI 任务时，环境配置都绝对一致。

- 安装并在诸如 node, python, go 等各类 **开发工具** 之间无缝切换。
- 基于不同目录加载隔离的 **环境变量**，支持读取 `.env` 以及 SOPS 加密数据。
- 编写并运行项目的构建、测试、代码检查及部署 **任务**。

虽然在概念上我们深受伟大的 `mise` 启发（在此致敬），但 UniRTM 在架构上引入了几个独特的选择：

- **纯 Go 引擎**: 底层完全由 Go 语言编写，利用 goroutines 实现了极致的并发下载能力。
- **无 Shims 垫片**: UniRTM 严格避免使用 Bash 垫片（Shims）。它直接将安装工具的绝对路径前置插入到你的 `$PATH` 变量中，保证了 100% 的执行性能，并对各种 IDE 完全透明。
- **原生安全检测**: 底层原生集成了 Trivy 和 Syft。当你下载工具时，会自动生成 SBOM 并扫描已知的安全漏洞。
- **强制版本锁定**: 自动生成 `unirtm.lock` 锁文件，不仅锁定版本号，还精确锁定所有下载包的校验和，从而保障团队环境的绝对可复现。

### 多维度深度对比 (Architecture & Environments)

为了满足现代企业级和高并发容器化环境的严苛要求，`UniRTM` 在底层架构上做出了极具针对性的设计抉择。以下是我们在多个关键维度上与优秀的生态前辈 (`mise`, `asdf`) 的深度对比：

#### 1. 核心架构与执行路径

| 对比维度 | `asdf` (Bash) | `mise` (Rust) | `UniRTM` (Go) | 核心优势与意义 |
| :--- | :--- | :--- | :--- | :--- |
| **命令执行路径** | 垫片 (Shims) | 垫片(默认) / PATH | **绝对零垫片 (Strict PATH)** | 直接将绝对路径注入环境变量，保证了 **100% 无损的执行性能** 和绝对的 IDE 透明度。 |
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
| **Windows 原生支持** | 局限性大 | 支持 | **自底向上原生设计** | 从设计之初就兼顾 Windows 与 Cygdrive，提供极其丝滑的原生跨平台体验。 |
| **Alpine / Musl 兼容** | 部分支持 | 支持 | **硬核底层兼容** | 完美无缝运行在没有任何 `glibc` 依赖的极简 Alpine 容器环境之中。 |
| **版本与校验锁定** | 仅版本号 | `mise.lock` | **`unirtm.lock` 严格锁定** | 锁定版本的同时严格校验所有底层文件的 Hash，彻底杜绝“在我的电脑上能跑”的问题。 |

## 支持的操作系统 (Supported Platforms)

*完全支持 macOS (Apple Silicon / Intel)、Linux (glibc & musl/Alpine) 以及 Windows 平台。*

## 演示 (Demo)

以下演示展示了如何使用 `UniRTM` 全局安装指定版本的 `go`。
请注意安装的速度以及内置的安全漏洞扫描功能！

[![demo](https://raw.githubusercontent.com/snowdreamtech/UniRTM/main/docs/tapes/demo.gif)](https://github.com/snowdreamtech/UniRTM/blob/main/docs/tapes/demo.mp4)

## 快速入门

### 安装 UniRTM

你可以通过多种方式安装 UniRTM，详见 [快速开始](https://unirtm.snowdream.tech/zh/guide/getting-started.html)。

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

查看 [开发工具指南](https://unirtm.snowdream.tech/zh/dev-tools/) 获取更多示例。

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

此外，`UniRTM` 同样可以自动读取本地的 [`.env` 文件](https://unirtm.snowdream.tech/zh/environments/#env-directives)。

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

查看 [任务系统指南](https://unirtm.snowdream.tech/zh/tasks/) 获取更多高级用法。

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
unirtm install # 安装所需要的所有工具，并自动在底层生成安全 SBOM
unirtm run deploy # 在部署前，将自动按顺序并行执行校验和安全审查任务
```

## 官方文档

完整的架构解析与高级配置指南请前往官网：[unirtm.snowdream.tech](https://unirtm.snowdream.tech/zh/)

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
