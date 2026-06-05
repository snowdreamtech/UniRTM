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

<p><b>Dev tools, env vars, and tasks in one CLI with built-in security.</b><br><i>Inspired by and paying tribute to <a href="https://github.com/jdx/mise">mise</a>.</i></p>

<p align="center">
  <a href="https://snowdreamtech.github.io/UniRTM/guide/getting-started.html">Getting Started</a> •
  <a href="https://snowdreamtech.github.io/UniRTM">Documentation</a> •
  <a href="https://snowdreamtech.github.io/UniRTM/dev-tools/overview.html">Dev Tools</a> •
  <a href="https://snowdreamtech.github.io/UniRTM/environments/overview.html">Environments</a> •
  <a href="https://snowdreamtech.github.io/UniRTM/tasks/overview.html">Tasks</a>
</p>

<hr />

</div>

> [!TIP]
> UniRTM's powerful task runner empowers you to orchestrate deep security scans (like Trivy & Syft) perfectly alongside your environments.

## What is it?

`UniRTM` (Universal Runtime Manager) prepares your development environment before each command runs. It keeps project tools, environment variables, and tasks in one `.unirtm.toml` file so new shells, checkouts, and CI jobs all start from the exact same setup.

- Install and switch between **dev tools** like node, python, go, and more.
- Load **environment variables** per project directory, including values from `.env` files and secure secret managers like SOPS.
- Define and run **tasks** for building, testing, linting, and deploying projects.

While taking heavy inspiration from the brilliant tool `mise` (dev tools, env vars, and tasks in one CLI), UniRTM introduces several distinct architectural choices:

- **Pure Go Engine**: Extreme parallel downloading capabilities leveraging goroutines.
- **Zero Shims**: UniRTM strictly avoids bash shims. It directly prepends the absolute paths of installed tools to your `$PATH`, ensuring 100% execution speed and transparency for IDEs.
- **Unifying Security via Tasks**: While keeping the core engine minimal, UniRTM allows you to perfectly integrate external security scanners like Trivy or Syft into your reproducible task workflows.
- **Absolute Locking**: Generates a `unirtm.lock` file that pins the exact checksums and versions of your downloaded tools for reproducible environments.

## Supported Platforms

*Fully supported on macOS (Apple Silicon / Intel), Linux (glibc & musl/Alpine), and Windows.*

## Demo

The following demo shows how to use `UniRTM` to install a specific version of `go` globally.
Notice the speed and the built-in vulnerability scanning!

[![demo](https://raw.githubusercontent.com/snowdreamtech/UniRTM/main/docs/tapes/demo.gif)](https://github.com/snowdreamtech/UniRTM/blob/main/docs/tapes/demo.mp4)

## Quickstart

### Install UniRTM

See [Getting started](https://snowdreamtech.github.io/UniRTM/guide/getting-started.html) for more options.

```sh-session
$ curl -sL https://raw.githubusercontent.com/snowdreamtech/UniRTM/main/install.sh | bash
$ ~/.local/bin/unirtm --version
UniRTM v0.1.0 macos-arm64 (2026-05-28)
```

Hook UniRTM into your shell (pick the right one for your shell):

```sh-session
# note this assumes unirtm is located at ~/.local/bin/unirtm
echo 'eval "$(~/.local/bin/unirtm env)"' >> ~/.bashrc
echo 'eval "$(~/.local/bin/unirtm env)"' >> ~/.zshrc
echo '~/.local/bin/unirtm env | source' >> ~/.config/fish/config.fish
```

### Execute commands with specific tools

```sh-session
$ unirtm exec node@20 -- node -v
unirtm node@20.x.x ✓ installed
v20.x.x
```

### Install tools

```sh-session
$ unirtm use --global node@22 go@1.22
$ node -v
v22.x.x
$ go version
go version go1.22.x macos/arm64
```

See [dev tools](https://snowdreamtech.github.io/UniRTM/dev-tools/) for more examples.

### Manage environment variables

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

Note that `UniRTM` can also [load `.env` files](https://snowdreamtech.github.io/UniRTM/environments/#env-directives).

### Run tasks

```toml
# .unirtm.toml
[tasks.build]
description = "build the project"
run = "echo building..."
```

```sh-session
$ unirtm run build
building...
```

See [tasks](https://snowdreamtech.github.io/UniRTM/tasks/) for more information.

### Example UniRTM project

Here is a combined example to give you an idea of how you can use UniRTM to manage your a project's tools, environment, and tasks with security baked in.

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
description = "Run terraform plan with configured workspace"
run = """
terraform init
terraform workspace select $TF_WORKSPACE
terraform plan
"""

[tasks.validate]
description = "Validate AWS credentials and terraform config"
run = """
aws sts get-caller-identity
terraform validate
"""

[tasks.audit]
description = "Run deep security scans using Trivy and Gitleaks"
run = """
trivy fs --format cyclonedx --output sbom.json .
gitleaks detect --source . --no-banner
"""

[tasks.deploy]
description = "Deploy infrastructure after validation and audit"
depends = ["validate", "audit", "plan"]
run = "terraform apply -auto-approve"
```

Run it with:

```sh-session
unirtm install # install all required tools specified in .unirtm.toml
unirtm run deploy # automatically runs validation and audit dependencies first
```

## Full Documentation

See [snowdreamtech.github.io/UniRTM](https://snowdreamtech.github.io/UniRTM)

### Architecture & Environments Comparison

We believe in making deliberate architectural choices to support modern enterprise environments. Here is how `UniRTM` compares to ecosystem pioneers like `mise` and `asdf` across multiple dimensions:

#### 1. Core Architecture & Execution

| Dimension | `asdf` (Bash) | `mise` (Rust) | `UniRTM` (Go) | Why it matters |
| :--- | :--- | :--- | :--- | :--- |
| **Execution Path** | Shims | Shims (Default) / PATH | **Strictly Direct PATH** | UniRTM injects absolute paths directly into your `$PATH`, providing a minimal execution path and transparent IDE experience. |
| **Concurrency Model** | None | OS Threads | **Goroutines** | Go's ultra-lightweight goroutines provide extreme parallel throughput during massive toolchain downloads. |
| **Config Hierarchy** | `.tool-versions` | `mise.toml` | **`.unirtm.toml`** | Standardized, unified TOML configs across your entire project. |

#### 2. Environment & Context

| Feature | `asdf` | `mise` | `UniRTM` | Details |
| :--- | :--- | :--- | :--- | :--- |
| **Scope Management** | Tools only | Tools + Env + Tasks | **Tools + Env + Tasks** | Seamlessly manage environments contextually per directory. |
| **`.env` Parsing** | ❌ | Native | **Native** | Reads traditional `.env` files automatically without external loaders. |
| **Secrets Engine** | ❌ | Plugins / Integrations | **Native SOPS** | Treats secure secret management as a first-class citizen using native SOPS integration. |

#### 3. Cross-Platform & Resilience

| Feature | `asdf` | `mise` | `UniRTM` | Details |
| :--- | :--- | :--- | :--- | :--- |
| **Windows Support** | WSL/MSYS Dependent | Supported | **Native (Ground-up)** | Engineered natively for a flawless Windows and Cygdrive experience. |
| **Alpine/Musl** | Partial | Supported | **Hardcore Supported** | Runs flawlessly in minimal `musl`/Alpine containers without `glibc` overhead. |
| **Reproducibility** | Versions | `mise.lock` (Supported) | **`unirtm.lock` (Default)** | Strictly pins exact checksums to ensure reproducible team environments. |

#### 4. Ecosystem Affinity & Minimalism

| Feature | `asdf` | `mise` | `UniRTM` | Details |
| :--- | :--- | :--- | :--- | :--- |
| **Hybrid Path Resolution** | ❌ | Partial | **Deeply Supported** | Dedicated optimizations for translating paths seamlessly across Windows Git Bash / Cygdrive environments. |
| **External Dependencies** | Bash Ecosystem | Minimal | **Absolute Zero** | Core plugins are compiled directly into the binary. Drop it into any minimal OS and it runs out of the box. |
| **Shim Overhead** | Standard (Bash script) | Optimized (Rust binary) | **Completely Eliminated** | Pure PATH injection design executes the real binary directly, which is highly compatible with complex IDE debuggers. |
| **DevOps Integration** | Custom Scripts | Good | **Native Synergy** | Go-based architecture naturally aligns with cloud-native infrastructure, making custom integrations frictionless. |


## GitHub Issues & Discussions

For feature requests, bug reports, and community support:

- [Discussions](https://github.com/snowdreamtech/UniRTM/discussions)
- [Issues](https://github.com/snowdreamtech/UniRTM/issues)

## Special Thanks

<p>
  Inspired by the architecture and developer experience pioneered by <a href="https://github.com/jdx/mise">mise</a>.
</p>

## License

MIT License. Copyright (c) 2026-present SnowdreamTech Inc.
