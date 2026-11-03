# Comparison to asdf

UniRTM takes a fundamentally different approach to performance, security, and extensibility compared to traditional version managers.

## Why not asdf?

`asdf` is a fantastic tool, but it is written entirely in Bash. This comes with massive performance penalties.

1. **Slow Shell Startup**: Every time you open a terminal, `asdf` executes hundreds of lines of bash to build its path manipulation, which can delay your shell prompt by 500ms or more.
2. **Shim Overhead**: `asdf` uses bash shims for every executable. When you type `node`, you aren't running Node directly; you are running a bash script that figures out which Node version to run, and *then* runs Node. This introduces severe latency, especially noticeable in tools that spawn many child processes (like `npm` or `cargo`).

**The UniRTM Solution**: UniRTM is written in compiled Go. It does not use slow bash shims. Instead, it dynamically injects the actual binary path directly into your `$PATH` when you change directories, resulting in zero overhead when running tools.
