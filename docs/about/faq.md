# Frequently Asked Questions (FAQ)

## What is the difference between UniRTM, mise, and asdf?

**asdf** is the pioneer in this field, but it is written in pure Bash, has extremely poor Windows support, and introduces significant performance overhead due to its `shims` mechanism.

**mise** is a fast rewrite of asdf that completely removes `shims` and provides a massive performance boost by dynamically altering the `$PATH`.

**UniRTM** pays homage to the architectural brilliance of **mise** and similarly abandons the `shims` mechanism in favor of lightning-fast shell prompt hooks. However, UniRTM's core advantage lies in the fact that it is **more than just a traditional runtime manager**. In the era of AI-assisted programming, UniRTM considers AI agent contexts (for Cursor, Windsurf, Copilot, etc.) as an indispensable part of the toolchain. It features native, built-in support for [SpecKit workflows](../guide/ide-integration.md).

## Why does a toolchain manager need to manage AI workflows (SpecKit)?

In the past, development environment consistency simply meant "having the same Node.js or Python version." Today, development teams face a new challenge: "How do we ensure that everyone's AI assistants are following the exact same project conventions, prompt contexts, and workflows?"

UniRTM believes that **AI Agent rules are fundamentally part of the modern developer toolchain**. Therefore, alongside traditional SDKs, we innovatively introduced the unified management of `.agent/rules`, creating a true "Single Source of Truth" for AI coding.

## Does UniRTM work on Windows?

**Absolutely.**
Unlike asdf, UniRTM is written in Go and treats Windows as an absolute first-class citizen. We have extensively optimized path resolution, environment variable handling, and native terminal support (PowerShell / CMD). You can seamlessly manage cross-platform tools on Windows without relying on WSL (unless you want to).

## Can I use my existing `.tool-versions` (asdf) or `mise.toml` files?

**Yes.**
To ensure a painless migration, UniRTM natively parses asdf's `.tool-versions` files and mise's `mise.toml` configurations. If you are working in a legacy project that already uses asdf or mise, simply installing UniRTM is enough to take over the environment smoothly.

## How does UniRTM dynamically modify environment variables?

UniRTM does not use slow `shims`. When you run `eval "$(unirtm activate bash/zsh)"` in your `~/.bashrc` or `~/.zshrc`, UniRTM injects a highly lightweight hook function into your Shell (such as `precmd` in zsh or `PROMPT_COMMAND` in bash).

Every time you change directories or hit enter, this hook evaluates the current directory structure at nanosecond speeds. If it detects a `.unirtm.toml`, it dynamically rewrites `$PATH` and other environment variables directly within the current Shell context. This means the `node` or `python` you execute is the actual binary on disk, with zero intermediate layer overhead.

## How do I set global default versions?

You can directly modify the global configuration file at `~/.config/unirtm/config.toml`, or use the command line for a quick setup:

```bash
unirtm use --global node@20
```

## How do I uninstall UniRTM?

UniRTM is entirely self-contained and leaves no persistent registries or deep system dependencies. Uninstalling takes just two steps:

1. Open your shell profile (e.g., `~/.zshrc` or `~/.bashrc`) and remove the `eval "$(unirtm activate ...)"` and `PATH` export lines.
2. Delete the data directories and the binary:

   ```bash
   rm -rf ~/.local/share/unirtm
   rm -rf ~/.config/unirtm
   rm -f ~/.local/bin/unirtm
   ```
