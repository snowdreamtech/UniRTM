# Introduction

Welcome to the **UniRTM (Universal Runtime Manager)** documentation.

UniRTM is an enterprise-grade, foundational template and tool manager inspired by `mise` and `asdf`. It allows you to unify your development environment, tasks, and environment variables into a single, blazing fast Go executable.

## Why UniRTM?

In modern software development, projects require various runtimes (Node.js, Go, Python), CLI tools (linters, formatters), and environment variables. Traditionally, developers use an assortment of tools like `nvm`, `pyenv`, `direnv`, and `make`.

UniRTM replaces them all:

- **Polyglot Tool Manager**: Installs and manages versions for multiple languages and tools automatically.
- **Environment Management**: Automatically loads environment variables when entering a directory.
- **Task Runner**: A simple cross-platform task runner built-in.
- **Fast**: Written in Go for maximum performance.

## Getting Started

To get started, follow the installation instructions in the repository README, or run the universal installation script if you are on macOS or Linux:

```bash
# Example installation script
curl -sL https://raw.githubusercontent.com/snowdreamtech/UniRTM/main/install.sh | bash
```

Once installed, you can use `unirtm` to install tools, run tasks, and manage your environment.
