# The `.unirtm.toml` File

The `.unirtm.toml` file is the heart of any UniRTM-managed project. It is where you declare your development tools, environment variables, and project tasks.

This file acts as a single source of truth for your project's development environment. It completely replaces `.nvmrc`, `.python-version`, `Makefile`, and `.env` loaders.

## Example Configuration

```toml
[env]
# Set standard environment variables
NODE_ENV = "development"
PORT = "3000"
# Variables can execute shell commands
GIT_HASH = { run = "git rev-parse --short HEAD" }

[tools]
# Pin exact versions
node = "20.11.1"
go = "1.22.0"
# Or use semantic versioning prefixes
python = "3.11"
# Use specific backends (e.g. GitHub releases)
"github:aquasecurity/trivy" = "0.49.0"

[tasks.build]
description = "Build the application"
run = "go build -o bin/app ./cmd/main.go"

[tasks.test]
description = "Run tests with coverage"
depends = ["build"]
run = "go test -cover ./..."

[settings]
# Override global settings just for this project
legacy_version_file = true
```

## Section Breakdown

### `[env]`

Defines environment variables that are injected into your shell when you `cd` into the directory.

- UniRTM supports static strings (`KEY = "value"`).
- It also supports dynamic execution (`KEY = { run = "command" }`).

### `[tools]`

Lists the dependencies required to work on this project.
UniRTM supports standard shortnames (e.g., `node`, `go`) mapped to our internal plugins. It also natively supports pulling raw binaries from GitHub using the `github:org/repo` syntax.

### `[tasks.*]`

Defines your task runner configuration.
Tasks can have dependencies (`depends`), aliases, and specific environment overrides. See [Task Runner Overview](../tasks/overview.md) for a deep dive.

### `[settings]`

Allows you to override global UniRTM settings defined in `~/.config/unirtm/config.toml` at a per-project level.

## Loading Order

UniRTM looks for `.unirtm.toml` files traversing up the directory tree. This means you can have a global `~/.unirtm.toml` for your personal tools, and a local `~/projects/my-app/.unirtm.toml` that overrides the global ones specifically for that project.
