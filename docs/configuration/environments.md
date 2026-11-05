# Configuration Environments

In modern software projects, you frequently need different configurations depending on whether you are running in local `development`, automated `ci`, `staging`, or `production`.

UniRTM adopts a clean **Multi-File Cascading Overlay** model (`.unirtm.<env>.toml`), keeping each configuration modular, focused, and collision-free.

## Architecture & Priority

UniRTM loads configurations hierarchically in the following precedence order (higher numbers override lower numbers):

1. **Global Configuration**: `~/.config/unirtm/unirtm.toml`
2. **Project Base Configuration**: `.unirtm.toml` (developer local tools & baseline tasks)
3. **Project Environment Overlay**: `.unirtm.<env>.toml` (e.g. `.unirtm.ci.toml`, `.unirtm.production.toml`)
4. **Local Private Override**: `.unirtm.local.toml` (machine-specific overrides, gitignored)

## Defining Environment Overlays

Instead of stuffing everything into a single bloated configuration file, create a dedicated `.unirtm.<env>.toml` file in the project root.

### Example: Base Configuration (`.unirtm.toml`)

Contains standard developer runtime tools and configurations:

```toml
[tools]
node = "20.18.0"
go = "1.23.0"

[env]
APP_ENV = "development"
PORT = 3000

[settings]
cache_ttl = 3600
```

### Example: CI Environment Overlay (`.unirtm.ci.toml`)

Contains CI-only scanners, security linters, and release binaries without polluting local developers' machines:

```toml
[tools]
"github:aquasecurity/trivy" = "0.76.1"
"github:google/osv-scanner" = "2.7.0"
"github:goreleaser/goreleaser" = "2.19.0"

[env]
APP_ENV = "ci"
CI = "true"
```

## Activating an Environment

An environment can be activated using any of the following methods:

1. **CLI Flag (`-E` or `--env`)**:
   ```bash
   unirtm -E ci run verify
   unirtm --env production run start
   ```

2. **Environment Variable (`UNIRTM_ENV`)**:
   ```bash
   export UNIRTM_ENV=ci
   unirtm run verify
   ```

## Complete Lockfile Generation (`unirtm lock`)

A major advantage of UniRTM's multi-file environment specification is full lockfile synchronization:

When you run `unirtm lock`, UniRTM automatically inspects `.unirtm.toml` and **all** `.unirtm.<env>.toml` overlays in your project. It generates a single, unified `unirtm.lock` file that locks every dependency across all environments, preventing version drift or accidental lockfile erasure between local development and CI runs.
