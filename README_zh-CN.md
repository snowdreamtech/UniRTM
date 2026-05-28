<div align="center">

<h1 align="center">
    UniRTM
</h1>

<p>
  <a href="https://github.com/snowdreamtech/UniRTM/blob/main/LICENSE"><img alt="GitHub" src="https://img.shields.io/github/license/snowdreamtech/UniRTM?style=for-the-badge&color=6B7F4E"></a>
  <a href="https://github.com/snowdreamtech/UniRTM/actions/workflows/pages.yml"><img alt="GitHub Workflow Status" src="https://img.shields.io/github/actions/workflow/status/snowdreamtech/UniRTM/pages.yml?style=for-the-badge&color=C5975B"></a>
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

## 介绍

`UniRTM` (Universal Runtime Manager) 能够在每次执行命令前自动准备好你的开发环境。它将项目所需的工具版本、环境变量和常用任务统一集中在 `.unirtm.toml` 文件中进行管理，确保每次打开新终端、切换分支或运行 CI 任务时，环境配置都绝对一致。

虽然在概念上我们深受伟大的 `mise` 启发（在此致敬），但 UniRTM 在架构上引入了几个独特的选择：
- **纯 Go 引擎**: 底层完全由 Go 语言编写，利用 goroutines 实现了极致的并发下载能力。
- **无 Shims 垫片**: UniRTM 严格避免使用 Bash 垫片（Shims）。它直接将安装工具的绝对路径前置插入到你的 `$PATH` 变量中，保证了 100% 的执行性能，并对各种 IDE 完全透明。
- **原生安全检测**: 底层原生集成了 Trivy 和 Syft。当你下载工具时，会自动生成 SBOM 并扫描已知的安全漏洞。
- **强制版本锁定**: 自动生成 `unirtm.lock` 锁文件，不仅锁定版本号，还精确锁定所有下载包的校验和，从而保障团队环境的绝对可复现。

### 🛠️ 开发工具管理
一键安装并无缝切换 node, python, go 等各种开发工具的版本。
```bash
unirtm use node@20
unirtm use python@3.11
```

### 🌍 环境变量管理
基于不同目录加载独立的环境变量，支持从 `.env` 文件以及 SOPS 等加密密钥管理器中安全读取数据。
```toml
[env]
NODE_ENV = 'development'
```

### ⚡ 任务执行系统
编写并运行项目的构建、测试、代码检查及部署任务。
```toml
[tasks.build]
description = '编译项目'
run = 'go build'
```
```bash
unirtm run build
```

## 快速入门

### 1. 安装 UniRTM
```bash
curl -sL https://raw.githubusercontent.com/snowdreamtech/UniRTM/main/install.sh | bash
```
或者通过 Go 安装：
```bash
go install github.com/snowdreamtech/UniRTM@latest
```

### 2. 配置 Shell 钩子
以 `zsh` 为例：
```bash
echo 'eval "$(unirtm env)"' >> ~/.zshrc
```
*（对于其他 Shell 环境，请运行 `unirtm help env` 查看具体说明）。*

### 3. 在项目中使用工具
```bash
cd my-project
unirtm use node@22
node -v # v22.x.x
```

## 官方文档
关于高级配置选项、CI/CD 持续集成方案以及插件开发的完整指南，请访问 [官方文档网站](https://unirtm.snowdream.tech/zh/)。

## 开源协议
本项目基于 MIT 许可证开源。版权所有 (c) 2026-至今 SnowdreamTech Inc。
