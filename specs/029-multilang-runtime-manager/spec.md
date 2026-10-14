# Feature Specification: Unified Multi-Language Runtime Manager

**Feature Identifier**: `029-multilang-runtime-manager`  
**Created**: 2026-10-14  
**Status**: In Progress  
**Input**: "Replace NVM, GVM, and Pyenv across platforms (Linux, macOS, Windows). Establish UniRTM as the definitive multi-language runtime manager with comprehensive CLI and Web UI support, seamless legacy version file compatibility (.nvmrc, .node-version, .go-version, .python-version, .tool-versions), smart semver/alias resolution (lts, latest, fuzzy prefixes), and isolated global package toolchains."

---

## 1. Background & Vision

### 1.1 The Fragmented Status Quo
Developers currently endure a painful, fragmented landscape of language-specific version managers:
- **NVM (Node Version Manager)**: POSIX-only shell scripts. On Windows, developers are forced to use the unrelated `nvm-windows` project, frequently encountering symlink permission errors (requiring elevated administrator prompts) and corrupted global `npm` package paths.
- **GVM (Go Version Manager)**: Shell-based, completely unsupported on Windows, tightly coupled to compiling Go from source, and notorious for network timeouts in restricted or domestic network environments.
- **Pyenv (Python Version Manager)**: Incurs heavy compilation times and fails frequently when missing system C libraries (OpenSSL, zlib, libffi). On Windows, it delegates to an independent fork (`pyenv-win`).
- **Tool Sprawl**: Developers routinely configure 4 to 5 separate managers, cluttering shell rc files, slowing down shell startup times, and creating friction across heterogeneous engineering teams.

### 1.2 The UniRTM Supremacy Paradigm
UniRTM eliminates this fragmentation by providing a **single, zero-dependency Go binary** that natively governs Node.js, Go, Python, and other toolchains across Linux (glibc & Alpine musl), macOS (Intel & Apple Silicon), and Windows (x86_64 & arm64).

Key pillars:
1. **Drop-in Replacement**: Instant adoption without rewriting existing project configurations by transparently reading legacy version files (`.nvmrc`, `.node-version`, `.go-version`, `.python-version`, `.tool-versions`).
2. **True Cross-Platform Shims**: POSIX symlinks / shell proxies on Unix, and Scoop-style hardlinks / executable trampolines on Windows requiring **zero administrator privileges**.
3. **Smart Resolution & Pre-compiled Binaries**: Instant downloads of pre-compiled official/mirror distributions with fuzzy version resolution (`node@20`, `go@1.23`, `python@3.12`, `node@lts`, `latest`).
4. **1:1 CLI & Web UI Parity**: Manage, switch, inspect, and install runtimes interchangeably via modern command-line ergonomics or an embedded interactive Web UI workbench.

---

## 2. User Stories & Acceptance Criteria

### User Story 1 - Legacy Project Zero-Config Takeover (Priority: P1) 🎯 MVP

As a developer cloning existing repositories configured for NVM, GVM, or Pyenv, I want UniRTM to automatically detect and honor legacy version files without manual configuration so that my tools switch to the exact required version seamlessly.

**Acceptance Scenarios**:
1. **Given** a directory containing `.nvmrc` (e.g. `v20.11.0` or `20`) or `.node-version`, **When** the developer enters the directory and executes `node -v` or `unirtm current`, **Then** UniRTM resolves and activates Node 20.11.0.
2. **Given** a directory containing `.go-version` (e.g. `1.22.4`), **When** running `go version`, **Then** UniRTM resolves Go 1.22.4 without host `GOROOT` contamination.
3. **Given** a directory containing `.python-version` (e.g. `3.12.1`), **When** running `python --version`, **Then** UniRTM resolves Python 3.12.1 without invoking local C compilers.
4. **Given** a directory containing `.tool-versions`, **When** evaluated, **Then** all declared runtimes are resolved in accordance with hierarchical lookup rules.

---

### User Story 2 - Fluent Installation with Aliases and Fuzzy Ranges (Priority: P1)

As a developer, I want to install runtimes using intuitive version requests (`node@lts`, `node@latest`, `node@20`, `go@1.23`, `python@3.12`) so that I don't need to look up exact patch releases.

**Acceptance Scenarios**:
1. **Given** `unirtm install node@lts`, **When** executed, **Then** UniRTM queries remote metadata, determines the active LTS release, downloads the pre-built distribution, verifies integrity, and updates the local registry.
2. **Given** `unirtm install go@1.23`, **When** executed, **Then** UniRTM resolves the newest `1.23.x` release and sets up the toolchain.
3. **Given** `unirtm install python@3.12`, **When** executed, **Then** UniRTM pulls pre-built standalone binaries for the host platform without compiling from source.
4. **Given** `unirtm ls-remote <runtime>`, **When** run, **Then** available remote releases are listed with LTS and latest markers.

---

### User Story 3 - Multi-Scope Version Switching (Priority: P1)

As a developer, I want to switch runtime versions at the project scope, global user scope, or ephemeral shell session scope so that I can control version activation precisely.

**Acceptance Scenarios**:
1. **Project Scope**: `unirtm use node@20.18.0` updates the project configuration (`.unirtm.toml`).
2. **Global Default Scope**: `unirtm use -g node@20.18.0` (or `unirtm use --global go@1.23.0`) sets the default fallback version across the user's entire machine in `~/.config/unirtm/unirtm.toml`.
3. **Session Scope**: `unirtm shell node@18` activates Node 18 only within the current shell context without mutating any configuration files on disk.

---

### User Story 4 - Global Package & Tool Isolation (Priority: P1)

As a developer installing global packages (e.g. `npm install -g pnpm`, `go install ...`, `pip install --user ...`), I want my tools to remain isolated per runtime version without breaking or colliding when switching versions.

**Acceptance Scenarios**:
1. **Node.js**: Global npm packages installed under Node 20 remain safely segregated and executable without corrupting system directories or Node 22 paths.
2. **Go**: `GOROOT` points strictly to the active version's root, preventing build errors, while `GOBIN` is safely resolved in the active execution path.
3. **Python**: `python`, `pip`, and companion shims operate against the selected version cleanly.

---

### User Story 5 - Full Web UI Workbench Parity (Priority: P1)

As a developer using the UniRTM Web Workbench, I want a dedicated "Runtimes" (多语言运行时) section to browse, install, switch, and inspect versions visually with one click.

**Acceptance Scenarios**:
1. **Visual Runtime Grid**: Displays cards for Node.js, Go, Python, Rust, etc., showing installed versions, active project version, and global default.
2. **One-Click Switcher**: Allows toggling the active version for the project or globally via interactive dropdowns.
3. **Remote Version Browser Modal**: Shows searchable remote versions with "Install" buttons and download progress indicators.
4. **Legacy Detection Badge**: Displays an alert banner when a legacy `.nvmrc` / `.node-version` / `.go-version` is detected, with an optional "Convert to .unirtm.toml" button.

---

## 3. Functional Requirements

- **FR-001 (Legacy File Discovery)**: `internal/config` MUST automatically discover `.nvmrc`, `.node-version`, `.go-version`, `.python-version`, and `.tool-versions` during upward hierarchical traversal if primary UniRTM files do not specify the runtime.
- **FR-002 (Semver & Alias Resolution)**: The runtime resolver MUST support resolving `lts`, `latest`, major prefixes (e.g., `20`, `1.23`), and semver ranges into concrete versions.
- **FR-003 (Pre-built Binaries)**: Node.js, Go, and Python distributions MUST be acquired as verified pre-built binaries without requiring C/C++ build chains on host systems.
- **FR-004 (Windows Non-Admin Shims)**: Windows shimming MUST utilize Scoop-style hardlinks or executable wrappers that operate under standard user permissions without administrator elevation.
- **FR-005 (Global Version Configuration)**: `unirtm use --global <tool>@<version>` MUST persist settings to the global user configuration file and update global shims immediately.
- **FR-006 (Ephemeral Shell Switching)**: `unirtm shell <tool>@<version>` MUST spawn or print an environment payload configuring the requested runtime strictly for the active session.
- **FR-007 (List & Status Inspection)**: `unirtm list` / `unirtm ls` MUST display installed versions per tool, highlighting the currently active version and its configuration source (`*`).
- **FR-008 (Remote Listing)**: `unirtm ls-remote <tool>` MUST return sorted remote releases, supporting pagination and channel filters (`--lts`, `--all`).
- **FR-009 (Web REST API Endpoints)**: The HTTP server MUST expose:
  - `GET  /api/runtimes/status`
  - `GET  /api/runtimes/remote?tool=<name>`
  - `POST /api/runtimes/install`
  - `POST /api/runtimes/uninstall`
  - `POST /api/runtimes/use`
  - `POST /api/runtimes/migrate-legacy`
- **FR-010 (Web UI Workbench View)**: The frontend MUST provide a dedicated "Runtimes" view with multi-language cards, version selectors, remote install dialogs, and legacy version indicator badges.
- **FR-011 (Complete i18n Localization)**: All CLI messages and Web UI elements MUST be 100% natively localized in `en-US`, `zh-CN`, and `zh-TW` without awkward mixed-language strings.

---

## 4. Success Criteria

1. **Complete Parity**: A developer can uninstall `nvm`, `gvm`, and `pyenv` and perform all daily development operations solely with `unirtm`.
2. **Zero-Config Compatibility**: Standard open-source projects using `.nvmrc` or `.go-version` run immediately out of the box with `unirtm`.
3. **Cross-Platform Consistency**: The exact same commands work identically on Linux, macOS, and Windows.
4. **Performance**: Shim dispatch overhead is under 15ms; pre-built binaries install in under 10 seconds under standard broadband.
