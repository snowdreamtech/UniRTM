# Quickstart Guide: Plugin Configuration

**Feature**: `027-plugin-config-isolation`  

---

## 1. Setting Plugin Configuration via CLI

### Project-Level Inline Configuration

```bash
# Set a configuration value for plugin 'snowdreamtech/docker'
unirtm config set plugins."snowdreamtech/docker".config.compose_file "docker-compose.yml"

# Set a boolean flag
unirtm config set plugins."snowdreamtech/docker".config.auto_prune true

# Read back plugin configuration
unirtm config get plugins."snowdreamtech/docker".config.compose_file
```

---

## 2. Using Sidecar Configuration Files

For large or complex plugin configurations:

1. Create a sidecar file in `.unirtm/plugins/`:
   ```bash
   mkdir -p .unirtm/plugins
   ```
2. Name the file after the plugin identifier (e.g. `.unirtm/plugins/docker.toml`):
   ```toml
   compose_file = "docker-compose.prod.yml"
   auto_prune = true
   registry = "harbor.internal"
   ```
3. UniRTM automatically detects and merges this file into `plugins."docker".config`.

---

## 3. Viewing All Active Plugin Configurations

```bash
# View all merged configurations including plugins and sidecars
unirtm config
```
