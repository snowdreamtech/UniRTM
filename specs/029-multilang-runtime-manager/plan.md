# Technical Plan: Unified Multi-Language Runtime Manager

**Feature**: Unified Multi-Language Runtime Manager (Replacing NVM, GVM, and Pyenv)  
**Branch**: `dev`  
**Spec**: [spec.md](./spec.md)  
**Tasks**: [tasks.md](./tasks.md)  

---

## 1. Architectural Goals & Design Principles

1. **Native Single-Binary Governance**: Provide zero-dependency runtime management for Node.js, Go, Python, Rust, etc., across Linux (glibc/musl), macOS (Intel/ARM), and Windows (x86_64/ARM64).
2. **Zero-Configuration Legacy Interoperability**: Transparently ingest legacy version files (`.nvmrc`, `.node-version`, `.go-version`, `.python-version`, `.tool-versions`) during hierarchical config lookup without requiring upfront migrations.
3. **Cross-Platform Shim Architecture**:
   - Unix: Direct symlinks pointing to UniRTM binary with fast trampoline delegation.
   - Windows: Scoop-style hardlinks / executable trampolines avoiding administrator privileges and symlink permission restrictions.
4. **Toolchain & Environment Boundary Protection**:
   - Prevent GOROOT cross-contamination across different Go versions.
   - Segregate global npm packages (`npm-global`) per Node version to eliminate version drift.
   - Provide pre-compiled standalone Python binaries avoiding host C/C++ compilation overhead.
5. **1:1 CLI and Web Parity**: Every command-line runtime operation (`install`, `use -g`, `use`, `list`, `ls-remote`, `shell`) is fully accessible and reactive in the embedded Web UI.

---

## 2. Go Backend Architecture

### 2.1 Legacy Version File Discovery Engine (`internal/config/`)
- Extend `internal/config/loader.go` with a legacy file reader:
  - `ReadLegacyVersionFile(dir string) (map[string]string, error)`
  - Files supported:
    - `.nvmrc` -> `node`
    - `.node-version` -> `node`
    - `.go-version` -> `go`
    - `.python-version` -> `python`
    - `.tool-versions` -> multi-tool map
  - Integrate into `LoadHierarchy(dir)`:
    - Priority: `.unirtm.toml` > `unirtm.toml` > `.mise.toml` > `mise.toml` > Legacy version files > Global config.
  - Expose `FoundLegacyFiles []LegacyFileInfo` in `Config` metadata so the CLI and Web UI can inform the user.

### 2.2 Smart Semver & Alias Resolution (`internal/backend/` & `internal/service/`)
- Support semantic alias keywords:
  - `node`: `lts`, `latest`, `current`, `lts/iron`, `lts/hydrogen`, `20`, `22`
  - `go`: `latest`, `1.23`, `1.22`
  - `python`: `latest`, `3.12`, `3.11`
- Implement `ResolveVersionAlias(ctx context.Context, tool, alias string) (string, error)`:
  - Fast-path cache of remote release indexes.
  - Prefix matching: if given `20`, find highest version matching `20.*.*`.
  - Channel matching: if given `lts`, query remote metadata for the designated active LTS release.

### 2.3 Runtime Providers & Pre-compiled Binaries (`internal/provider/native/`)
- Ensure Node.js:
  - Upstream official index: `https://nodejs.org/dist/index.json`.
  - Mirror acceleration: respects active network profile (`npmmirror`, `ustc`, etc.).
  - Windows: handles `.zip` extraction and `npm.cmd` / `npx.cmd` shim generation.
- Ensure Go:
  - Upstream index & mirror: `https://go.dev/dl/?mode=json` or domestic mirror.
  - Windows: `.zip` distribution extraction.
  - Clean `GOROOT` and `GOBIN` isolation.
- Ensure Python:
  - Standalone binaries (`python-build-standalone` / official installers).
  - Pre-compiled for Linux (glibc & musl), macOS, and Windows.
  - Zero local C toolchain compilation dependency.

### 2.4 Multi-Scope Version Management (`cmd/`)
- `unirtm install <tool>[@<version>]`:
  - Downloads, verifies SHA256 checksums, unzips to `~/.unirtm/installs/<tool>/<version>`, and registers shims.
- `unirtm use [-g|--global] <tool>[@<version>]`:
  - Without `-g`: records into local `.unirtm.toml`.
  - With `-g`: records into global config `~/.config/unirtm/unirtm.toml` and updates default active shims.
- `unirtm list` / `unirtm ls`:
  - Lists all installed versions per runtime.
  - Highlights active version with `*` and displays configuration source (`project: .unirtm.toml`, `legacy: .nvmrc`, `global: ~/.config/unirtm/unirtm.toml`).
- `unirtm current`:
  - Concise view of active runtimes in the current working directory.
- `unirtm ls-remote <tool> [--lts] [--all]`:
  - Lists remote available versions from official/mirror indexes.
- `unirtm shell <tool>@<version>`:
  - Generates subshell or prints environment variables for session-only activation.

### 2.5 Web REST API Endpoints (`cmd/64.ui.go`)
- `GET  /api/runtimes/status`: Returns all known runtimes (Node, Go, Python, etc.) with installed versions, active version, source, and any detected legacy files.
- `GET  /api/runtimes/remote?tool=<name>&lts=<bool>`: Returns remote versions list.
- `POST /api/runtimes/install`: Initiates background installation with live progress reporting.
- `POST /api/runtimes/uninstall`: Removes an installed version cleanly.
- `POST /api/runtimes/use`: Switches version globally or for the project.
- `POST /api/runtimes/migrate-legacy`: Migrates detected `.nvmrc` / `.node-version` into `.unirtm.toml`.

---

## 3. Web UI Frontend Architecture (Vue 3 + Vite)

### 3.1 Views & Components
1. **RuntimesView.vue** (Dedicated Runtimes Management):
   - **Header & Metrics**: Total runtimes governed, total disk space consumed, legacy files status banner.
   - **Runtime Cards Grid**:
     - Dedicated cards for Node.js, Go, Python, Rust, etc.
     - Active project version badge + Global default version badge.
     - Installed versions pills with active indicator.
     - Quick "Switch" dropdown to change active version instantly.
     - "Install New Version" button launching the Remote Browser Modal.
   - **Legacy File Auto-Detection Banner**:
     - Highlights if the project relies on `.nvmrc`, `.go-version`, etc.
     - One-click "Import to .unirtm.toml" button with confirmation modal.
2. **RemoteVersionModal.vue**:
   - Searchable and filterable list of remote versions (with LTS tags).
   - One-click install with reactive progress spinner.
3. **DashboardView.vue**:
   - Integrates a compact "Active Runtimes" summary widget showing Node, Go, and Python versions in real-time.

### 3.2 Full i18n Localization
- Expand `frontend/src/i18n/` with comprehensive `runtimes.*` dictionaries across:
  - `en-US.ts` (English)
  - `zh-CN.ts` (Simplified Chinese)
  - `zh-TW.ts` (Traditional Chinese)
- Zero mixed-language text or awkward parenthetical combinations.

---

## 4. Verification & Testing Strategy

1. **Unit Tests**:
   - Legacy version file discovery across multiple folder hierarchies (`loader_legacy_test.go`).
   - Semver prefix and LTS alias resolution (`alias_test.go`).
   - Windows/Unix shim creation and trampoline execution (`shim_test.go`).
2. **CLI Integration Tests**:
   - `unirtm install`, `use`, `use -g`, `list`, `current`, `ls-remote` command tests.
3. **Web API & Frontend Build Verification**:
   - API handler tests for `/api/runtimes/*`.
   - `npm run build` frontend asset bundle generation with zero TypeScript/Vue compiler errors.
4. **Dogfooding & End-to-End**:
   - Test in directories with `.nvmrc`, `.go-version`, and `.python-version`.
   - Verify `node`, `go`, and `python` invocations delegate directly to the correct isolated binary.
