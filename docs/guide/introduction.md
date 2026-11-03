# Introduction

Welcome to the **UniRTM (Universal Runtime Manager)** documentation.

UniRTM is an enterprise-grade, foundational template and tool manager inspired by `mise` and `asdf`. It allows you to unify your development environment, tasks, and environment variables into a single, blazing fast Go executable.

## Why UniRTM?

In modern software development, projects require various runtimes (Node.js, Go, Python), CLI tools (linters, formatters), and environment variables. Traditionally, developers use an assortment of tools like `nvm`, `pyenv`, `direnv`, and `make`.

UniRTM replaces them all:

- **100% Native Architecture**: Built entirely in Go. Zero dependencies on external plugins, bash scripts, or system-level tools (like `pipx` or `make`), delivering maximum performance.
- **Absolute Zero-Pollution**: Manage environments without needing invasive shell hooks (`.zshrc`/`.bashrc` manipulations) or directory-switching aliases.
- **Built-in MCP Server**: Out-of-the-box support for the Model Context Protocol (MCP), allowing AI agents to natively manage your environments.
- **Polyglot Tool Manager**: Installs and manages versions for multiple languages natively.
- **Environment Management**: Declarative loading of variables and secrets via `.unirtm.toml`.
- **Task Runner**: A simple cross-platform task runner built-in.

## Getting Started

To get started, follow the installation instructions in the repository README, or run the universal installation script if you are on macOS or Linux:

```bash
# Example installation script
curl -sL https://raw.githubusercontent.com/snowdreamtech/UniRTM/main/install.sh | bash
```

Once installed, you can use `unirtm` to install tools, run tasks, and manage your environment.
