# Comparison to asdf / mise

UniRTM is highly inspired by `asdf` and `mise`, but it takes a fundamentally different approach to performance, security, and extensibility.

## Why not asdf?

`asdf` is a fantastic tool, but it is written entirely in Bash. This comes with massive performance penalties.

1. **Slow Shell Startup**: Every time you open a terminal, `asdf` executes hundreds of lines of bash to build its path manipulation, which can delay your shell prompt by 500ms or more.
2. **Shim Overhead**: `asdf` uses bash shims for every executable. When you type `node`, you aren't running Node directly; you are running a bash script that figures out which Node version to run, and *then* runs Node. This introduces severe latency, especially noticeable in tools that spawn many child processes (like `npm` or `cargo`).

**The UniRTM Solution**: UniRTM is written in compiled Go. It does not use slow bash shims. Instead, it dynamically injects the actual binary path directly into your `$PATH` when you change directories, resulting in zero overhead when running tools.

## Why not mise?

`mise` (formerly `rtx`) solved `asdf`'s performance issues by rewriting the core in Rust. However, UniRTM takes it a step further:

1. **Go Native vs Rust**: Go provides unmatched cross-compilation simplicity and concurrency models. Our native core plugins for Go, Python, and Node fetch concurrently at maximum line speed.
2. **Security Integrated**: UniRTM treats supply chain security as a first-class citizen. It natively integrates Trivy, Syft, and Gitleaks to actively scan downloaded artifacts, generating SBOMs and intercepting known vulnerabilities before they enter your development environment.
3. **GitHub Releases Native**: Unlike `mise` which heavily relies on custom plugins or `aqua`, UniRTM natively parses GitHub releases. You can install thousands of CLI tools just by referencing their GitHub repository: `unirtm use github:sharkdp/fd`.
