# Dev Tools Management Overview

UniRTM handles the installation and switching of development tools on your system automatically. Whether you are working with `node`, `python`, `go`, or raw binaries published to GitHub, UniRTM manages them seamlessly without relying on sluggish shell scripts.

## How it works

When you run `unirtm use node@20`, two things happen:

1. **Resolution & Download**: UniRTM resolves the version from its configured backend, downloads the binary to a global cache directory (usually `~/.local/share/unirtm/installs`), and makes it executable.
2. **Registration**: It writes the dependency to your local `.unirtm.toml`.

When you navigate to this directory in your terminal, UniRTM detects `.unirtm.toml` and injects its internal `bin` directory into your system `$PATH`. This guarantees that when you type `node`, you are executing the exact version specified for this project.

## Backends

A core difference between UniRTM and legacy tool managers is how it fetches software. UniRTM uses a pluggable backend architecture:

- **Core Plugins**: UniRTM ships with core plugins compiled directly into its Go binary. Tools like Node.js, Go, and Python have native support with maximum download speeds and checksum validation.
- **GitHub Backend**: You can install almost any tool directly from GitHub releases without needing a custom plugin.

  ```bash
  unirtm use github:cli/cli@latest
  ```

- **Cargo/NPM/Pip Backends**: UniRTM can proxy installations through other package managers, managing the resulting binary versions independently per project.

## Tool Resolution Order

UniRTM resolves tool versions by inspecting the following locations (in order):

1. `UNIRTM_NODE_VERSION` (Environment variables)
2. `.unirtm.toml` (Current directory)
3. `.nvmrc` or `.node-version` (Legacy files, if enabled)
4. `.unirtm.toml` (Parent directories, recursively up to root)
5. `~/.config/unirtm/config.toml` (Global config)

If a tool is declared but not installed locally, UniRTM will gracefully fallback or prompt you to run `unirtm install`.
