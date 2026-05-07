# UniRTM User Documentation

## Installation

### Quick Install (Linux/macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/snowdreamtech/unirtm/main/install.sh | sh
```

### Manual Install

Download the latest binary from [GitHub Releases](https://github.com/snowdreamtech/unirtm/releases)
and place it in your `$PATH`.

---

## Quick Start

```bash
# Initialize shell integration (add to ~/.bashrc or ~/.zshrc)
eval "$(unirtm env bash)"

# Install a tool
unirtm install node 20.0.0

# Activate a tool version
unirtm activate node 20.0.0

# List installed tools
unirtm list

# Search available tools
unirtm search python

# Check system health
unirtm doctor
```

---

## Configuration

UniRTM uses TOML configuration files. Configuration is merged from multiple levels:

| Level | Location | Priority |
|-------|----------|----------|
| System | `/etc/unirtm/config.toml` | Lowest |
| Global | `~/.config/unirtm/config.toml` | Low |
| Project | `./unirtm.toml` | High |
| Local | `./unirtm.local.toml` | Highest |

### Example `unirtm.toml`

```toml
[settings]
cache_ttl = 3600          # Cache TTL in seconds
concurrency = 4           # Parallel install limit

[tools.node]
version = "20.0.0"
backend = "github"        # github | aqua | http

[tools.python]
version = "3.11.0"

[tools.go]
version = "1.21.0"

[environments.production]
[environments.production.tools.node]
version = "18.0.0"        # Pin to LTS in production
```

---

## CLI Command Reference

### `unirtm install <tool> <version>`

Install a specific version of a tool.

```bash
unirtm install node 20.0.0
unirtm install python latest
unirtm install go 1.21.0 --backend github
unirtm install node 20.0.0 --dry-run   # Preview without installing
```

### `unirtm uninstall <tool> <version>`

Remove an installed tool version.

```bash
unirtm uninstall node 18.0.0
unirtm uninstall node 18.0.0 --dry-run
```

### `unirtm list [tool]`

List all installed tools, optionally filter by tool name.

```bash
unirtm list
unirtm list node
unirtm list --json    # Machine-readable output
```

### `unirtm activate <tool> <version>`

Activate a tool version in the current shell session.

```bash
unirtm activate node 20.0.0
unirtm activate node 20.0.0 --shell bash
```

### `unirtm deactivate [tool]`

Deactivate the current tool environment.

```bash
unirtm deactivate
unirtm deactivate node
```

### `unirtm search <query>`

Search the tool index.

```bash
unirtm search python
unirtm search --backend aqua linter
unirtm search node --json
```

### `unirtm update [tool] [version]`

Update tools to newer versions.

```bash
unirtm update              # Update all tools
unirtm update node         # Update node to latest
unirtm update node 20.1.0  # Update to specific version
unirtm update --dry-run    # Preview updates
```

### `unirtm cache`

Manage the local artifact cache.

```bash
unirtm cache list          # List cached artifacts
unirtm cache stats         # Show cache statistics
unirtm cache clear         # Clear all cache
unirtm cache clear node    # Clear cache for specific tool
unirtm cache purge         # Remove expired entries
```

### `unirtm config`

Manage configuration values.

```bash
unirtm config validate              # Validate current config
unirtm config show                  # Display merged config
unirtm config get settings.cache_ttl
unirtm config set settings.concurrency 8
```

### `unirtm doctor`

Run system diagnostics and report issues.

```bash
unirtm doctor
unirtm doctor --json
```

### `unirtm migrate [source]`

Import configuration from mise or asdf.

```bash
unirtm migrate .mise.toml              # Migrate from mise
unirtm migrate .tool-versions          # Migrate from asdf
unirtm migrate .mise.toml --dry-run    # Preview migration
unirtm migrate .mise.toml -o unirtm.toml  # Specify output file
```

### `unirtm env <shell>`

Print shell integration script.

```bash
eval "$(unirtm env bash)"
eval "$(unirtm env zsh)"
unirtm env fish | source
unirtm env powershell | Invoke-Expression
```

### `unirtm completion <shell>`

Generate shell completion scripts.

```bash
unirtm completion bash >> ~/.bashrc
unirtm completion zsh >> ~/.zshrc
unirtm completion fish > ~/.config/fish/completions/unirtm.fish
```

---

## Global Flags

| Flag | Description |
|------|-------------|
| `--verbose` | Enable verbose output |
| `--quiet` | Suppress non-essential output |
| `--json` | Machine-readable JSON output |
| `--dry-run` | Preview actions without side effects |
| `--config <path>` | Use specific config file |

---

## Troubleshooting

### Tool not found after installation

Ensure shell integration is active:

```bash
eval "$(unirtm env bash)"  # or your shell
```

### Checksum verification failed

The downloaded archive did not match the expected checksum. Try:

```bash
unirtm cache clear <tool>
unirtm install <tool> <version>
```

### Network errors

Check connectivity and proxy settings:

```bash
export HTTPS_PROXY=http://proxy:8080
unirtm install <tool> <version>
```

For offline usage, UniRTM will use cached artifacts automatically when
network is unavailable.

### Run diagnostics

```bash
unirtm doctor
```
