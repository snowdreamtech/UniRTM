<div align="center">

<h1 align="center">
    UniRTM
</h1>

<p>
  <a href="https://github.com/snowdreamtech/UniRTM/blob/main/LICENSE"><img alt="GitHub" src="https://img.shields.io/github/license/snowdreamtech/UniRTM?style=for-the-badge&color=6B7F4E"></a>
  <a href="https://github.com/snowdreamtech/UniRTM/actions/workflows/pages.yml"><img alt="GitHub Workflow Status" src="https://img.shields.io/github/actions/workflow/status/snowdreamtech/UniRTM/pages.yml?style=for-the-badge&color=C5975B"></a>
</p>

<p><b>Dev tools, env vars, and tasks in one CLI with built-in security.</b><br><i>Inspired by and paying tribute to <a href="https://github.com/jdx/mise">mise</a>.</i></p>

<p align="center">
  <a href="https://unirtm.snowdream.tech/guide/getting-started.html">Getting Started</a> •
  <a href="https://unirtm.snowdream.tech">Documentation</a> •
  <a href="https://unirtm.snowdream.tech/dev-tools/overview.html">Dev Tools</a> •
  <a href="https://unirtm.snowdream.tech/environments/overview.html">Environments</a> •
  <a href="https://unirtm.snowdream.tech/tasks/overview.html">Tasks</a>
</p>

<hr />

</div>

## What is it?

`UniRTM` (Universal Runtime Manager) prepares your development environment before each command runs. It keeps project tools, environment variables, and tasks in one `.unirtm.toml` file so new shells, checkouts, and CI jobs all start from the exact same setup.

While taking heavy inspiration from the brilliant tool `mise` (dev tools, env vars, and tasks in one CLI), UniRTM introduces several distinct architectural choices:
- **Pure Go Engine**: Extreme parallel downloading capabilities leveraging goroutines.
- **Zero Shims**: UniRTM strictly avoids bash shims. It directly prepends the absolute paths of installed tools to your `$PATH`, ensuring 100% execution speed and transparency for IDEs.
- **Native Security**: Built-in integration with Trivy and Syft to generate SBOMs and scan for vulnerabilities whenever you install a tool.
- **Absolute Locking**: Generates a `unirtm.lock` file that pins the exact checksums and versions of your downloaded tools for reproducible environments.

### 🛠️ Dev Tools
Install and switch between dev tools like node, python, go, and more. 
```bash
unirtm use node@20
unirtm use python@3.11
```

### 🌍 Environments
Load environment variables per project directory, including values from `.env` files and secure secret managers like SOPS.
```toml
[env]
NODE_ENV = 'development'
```

### ⚡ Tasks
Define and run tasks for building, testing, linting, and deploying projects.
```toml
[tasks.build]
description = 'Build the project'
run = 'go build'
```
```bash
unirtm run build
```

## Quickstart

### 1. Install UniRTM
```bash
curl -sL https://raw.githubusercontent.com/snowdreamtech/UniRTM/main/install.sh | bash
```
Or via Go:
```bash
go install github.com/snowdreamtech/UniRTM@latest
```

### 2. Hook UniRTM into your shell
For `zsh`:
```bash
echo 'eval "$(unirtm env)"' >> ~/.zshrc
```
*(Run `unirtm help env` for other shells).*

### 3. Use tools in your project
```bash
cd my-project
unirtm use node@22
node -v # v22.x.x
```

## Documentation
For complete documentation, including advanced configuration, CI/CD integration, and plugin development, visit the [official website](https://unirtm.snowdream.tech).

## License
MIT License. Copyright (c) 2026-present SnowdreamTech Inc.
