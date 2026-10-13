# Plugin & Extension Configuration Guide

UniRTM provides a strict three-layer configuration namespace isolation model inspired by industry standards such as **VS Code Configuration Contributions** and **Cargo Metadata Namespaces**.

This architecture guarantees that third-party plugins and extensions can safely expand configuration without collisions, root table pollution, or cross-plugin overrides.

---

## 1. Architecture: Three-Layer Namespace Isolation

```
.unirtm.toml Root Structure
├── 1. Core Reserved Tables       -> [tools], [tasks], [format], [lint], [hooks], [env], [settings]
├── 2. Plugin Declarations        -> [plugins] (version constraints & enabled status)
└── 3. Isolated Plugin Spaces     -> [plugins."<plugin-id>".config] (exclusive entrypoint)
```

### Core Reserved Namespaces
The following top-level tables are reserved for UniRTM core features:
* `[tools]`: Tool version definitions.
* `[tasks]`: Automation task workflows.
* `[format]`: Code formatter rules.
* `[lint]`: Linter specifications.
* `[hooks]`: Lifecycle trigger hooks.
* `[env]`: Environment variables.
* `[settings]`: CLI runtime settings.
* `[plugins]`: Enabled plugin declarations.

Third-party plugins and CLI commands **cannot write new top-level tables** outside this reserved list.

---

## 2. Configuration Patterns

### Pattern A: Inline Configuration (`.unirtm.toml`)

For lightweight configuration, settings can be written directly inside `.unirtm.toml`:

```toml
[plugins]
"snowdreamtech/docker" = "1.0.0"
"community/deploy" = "0.2.1"

[plugins."snowdreamtech/docker".config]
compose_file = "docker-compose.yml"
port = 8080
auto_prune = true

[plugins."community/deploy".config]
cluster = "k8s-staging"
timeout = "45s"
```

### Pattern B: Sidecar Configuration (`.unirtm/plugins/*.toml`)

For complex plugins with dozens of configuration parameters, configuration can live in external sidecar files:

* **Project Sidecar**: `.unirtm/plugins/<plugin-id>.toml` (or nested `vendor/plugin.toml`)
* **Global Sidecar**: `~/.config/unirtm/plugins/<plugin-id>.toml`

Example `.unirtm/plugins/docker.toml`:

```toml
port = 8080
registry = "harbor.internal"
auto_prune = true
```

UniRTM automatically discovers and deep-merges sidecars at startup in `< 5ms`.

---

## 3. Precedence Hierarchy

When a configuration key is defined across multiple locations, UniRTM resolves conflicts according to the following precedence (highest to lowest):

1. **Inline Project Configuration** (`.unirtm.toml` under `[plugins."<id>".config]`) *(Wins conflicts)*
2. **Project Sidecar File** (`.unirtm/plugins/<id>.toml`)
3. **Global Sidecar File** (`~/.config/unirtm/plugins/<id>.toml`)

---

## 4. CLI Operations

### Writing Configuration
```bash
# Set inline plugin setting
unirtm config set plugins."snowdreamtech/docker".config.port 8080

# Set string setting
unirtm config set plugins."community/deploy".config.cluster staging
```

### Reading Configuration
```bash
# Query plugin configuration
unirtm config get plugins."snowdreamtech/docker".config.port

# Inspect all active plugins and configurations
unirtm config
```

---

## 5. Plugin Developer Guide: Manifest Schema Validation

Plugins can provide a manifest schema (`.manifest.json`) declaring types, descriptions, and allowed enums:

```json
{
  "name": "snowdreamtech/docker",
  "version": "1.0.0",
  "contributes": {
    "configuration": {
      "port": {
        "type": "integer",
        "description": "Port to bind container reverse proxy",
        "default": 8080
      },
      "driver": {
        "type": "string",
        "description": "Storage driver",
        "enum": ["overlay2", "btrfs"]
      }
    }
  }
}
```

When a manifest is present:
* `unirtm config set` strictly checks input types and enums before writing.
* Type mismatches or invalid enum values are immediately rejected with clear error messages.
