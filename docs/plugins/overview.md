# Plugins Overview

While UniRTM contains native support (via core Go plugins) for popular languages like Node, Python, and Go, it cannot natively support every tool in existence.

To support an infinite ecosystem of development tools, UniRTM is **100% backward compatible with the `asdf` plugin ecosystem** and natively supports extracting raw binaries from **GitHub releases**.

## How Plugins Work

When you request a tool that is not a core plugin (e.g., `kubectl`, `terraform`), UniRTM delegates the installation logic to a backend plugin.

### The GitHub Backend (Recommended)

If the tool you need publishes raw binaries to GitHub Releases, you don't even need a plugin. UniRTM can figure out how to download and extract it natively based on your OS and Architecture.

```bash
# This downloads the raw binary from github.com/cli/cli
unirtm use github:cli/cli@latest
```

### The ASDF Backend

If the tool requires complex compilation (like Erlang) or specific build flags, you can leverage the thousands of existing `asdf` plugins.

```bash
unirtm plugin install erlang https://github.com/asdf-vm/asdf-erlang.git
unirtm use erlang@26.2.2
```

Behind the scenes, UniRTM executes the shell scripts provided by the `asdf` plugin (`bin/download`, `bin/install`, `bin/list-all`), but manages the resulting binaries using its own blazing-fast Go execution engine.

## Creating a Plugin

If a tool doesn't exist on GitHub and lacks an `asdf` plugin, you can easily create your own. See [Backend Development](./backend-development.md) for a guide on writing custom plugins.
