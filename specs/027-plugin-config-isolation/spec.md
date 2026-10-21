# Feature Specification: Plugin & Extension Configuration Namespace Isolation

**Feature Branch**: `027-plugin-config-isolation`
**Created**: 2026-10-13
**Status**: Draft
**Input**: "要保证后续第三方插件（Plugins）和扩展（Extensions）能够安全扩展配置，且绝不互相覆盖、不污染核心、不发生键名冲突，建立一套严格的配置命名空间体系与写入准则（涵盖核心保留区、插件隔离区、轻量行内与外置 Sidecar 分片、写权限沙箱与并发写锁、Plugin Manifest Schema 强类型校验）。"

---

## 1. Background & Motivation

As UniRTM evolves to support a third-party plugin and extension marketplace, multiple independent extensions will need to persist project-level and machine-level configurations. Without a standardized namespacing and isolation architecture:
1. Third-party plugins might write arbitrary root keys into `.unirtm.toml`, clashing with future official core features or conflicting with other plugins.
2. Uncoordinated concurrent writes from different plugins, background daemons, or UI panels could clobber each other's configuration data or strip user comments.
3. Heavyweight plugins could cause `.unirtm.toml` to bloat uncontrollably, polluting Git diffs.
4. Unchecked plugin configurations could inject unvalidated types or tamper with core execution settings (`[tools]`, `[format]`, `[hooks]`).

Inspired by best practices from VS Code (`contributes.configuration`), Rust Cargo (`[package.metadata.*]`), and Kubernetes CRDs, this specification establishes an industrial-grade configuration partitioning and validation framework.

---

## 2. User Scenarios & Acceptance Criteria *(mandatory)*

### User Story 1 - Safe Plugin Configuration Declaration & Isolation (Priority: P1) 🎯 MVP

As a developer using multiple third-party UniRTM plugins (e.g. `snowdreamtech/docker` and `community/deploy`), I want their configurations in `.unirtm.toml` to be strictly isolated under dedicated namespaces, so that one plugin cannot overwrite, read, or tamper with another plugin's configuration.

**Why this priority**: Core prerequisite for launching the plugin ecosystem safely.
**Independent Test**: Configure two plugins in `.unirtm.toml`, verify through `unirtm config get` and programmatic API that each plugin's settings are completely segregated under `[plugins."<id>".config]`.

**Acceptance Scenarios**:
1. **Given** a `.unirtm.toml` with `[plugins."snowdreamtech/docker".config]` and `[plugins."community/deploy".config]`, **When** UniRTM loads the configuration, **Then** both plugin configurations are parsed into their respective isolated structures without key collision.
2. **Given** a plugin or user attempting to execute `unirtm config set foo.bar val` where `foo` is neither a core reserved table nor an installed plugin ID, **When** the command executes, **Then** UniRTM rejects the write with a clear error requiring the plugin namespace format.

---

### User Story 2 - Sidecar (External Split) Configuration for Heavyweight Plugins (Priority: P1)

As a team lead managing a project with complex plugin setups, I want heavy plugin configurations to optionally live in `.unirtm/plugins/<plugin-id>.toml` rather than cluttering the root `.unirtm.toml`, so that the main project config stays lightweight and Git history remains clean.

**Why this priority**: Prevents `.unirtm.toml` bloat and avoids merge conflicts on team repositories.
**Independent Test**: Create `.unirtm/plugins/snowdreamtech-docker.toml`, run `unirtm config get plugins."snowdreamtech/docker".config`, and verify UniRTM automatically merges it into the active runtime configuration view.

**Acceptance Scenarios**:
1. **Given** a split config file at `.unirtm/plugins/my-plugin.toml`, **When** `unirtm` commands run, **Then** the engine automatically merges its contents into `cfg.Plugins["my-plugin"].Config`.
2. **Given** a global split config file at `~/.config/unirtm/plugins/my-plugin.toml`, **When** running with `--global`, **Then** the engine reads and writes to the user's isolated sidecar file.

---

### User Story 3 - Write-Time Sandbox & Concurrency Protection (Priority: P2)

As a plugin developer or CI runner, I want configuration writes to be atomic, thread-safe, comment-preserving, and sandboxed, so that concurrent operations never corrupt `.unirtm.toml` or erase human-written comments.

**Why this priority**: Guarantees data integrity and prevents race conditions.
**Independent Test**: Run parallel `unirtm config set` commands targeting different plugin keys and verify that no keys are lost and file syntax remains valid.

**Acceptance Scenarios**:
1. **Given** concurrent write requests to `.unirtm.toml`, **When** the write executes, **Then** an OS file lock (`.unirtm.toml.lock`) serializes the writes, preventing corruption.
2. **Given** an existing `.unirtm.toml` with user comments, **When** updating a plugin setting, **Then** surrounding comments and unrelated keys remain intact.
3. **Given** plugin A trying to write to `plugins."pluginB".config.key`, **When** running under plugin A's identity, **Then** write-time validation rejects the attempt as unauthorized.

---

### User Story 4 - Manifest Schema Validation (Priority: P2)

As a developer configuring a plugin, I want immediate feedback if I enter an invalid configuration value (e.g. string instead of boolean), so that configuration mistakes are caught before runtime execution.

**Why this priority**: Eliminates silent configuration bugs and improves developer ergonomics.
**Independent Test**: Set an invalid type against a plugin that provides a schema in its manifest, and verify the CLI warns or errors immediately.

---

## 3. Requirements *(mandatory)*

### Functional Requirements

- **FR-001 (Core Reserved Namespaces)**: UniRTM MUST explicitly reserve the following root-level tables:
  `[tools]`, `[tasks]`, `[format]`, `[lint]`, `[hooks]`, `[env]`, `[settings]`, `[plugins]`.
  No third-party plugin or non-core command may write new root-level tables outside this list.
- **FR-002 (Plugin Isolation Namespace)**: Third-party inline plugin configurations MUST be located exclusively under `[plugins."<plugin-id>".config]`. The `<plugin-id>` MUST follow standard vendor/name syntax (e.g. `vendor/name` or reverse-domain `io.unirtm.name`).
- **FR-003 (Plugin Declaration List)**: The root table `[plugins]` MUST declare enabled plugins and their version constraints:
  ```toml
  [plugins]
  "snowdreamtech/docker" = "1.0.0"
  ```
- **FR-004 (Sidecar Discovery & Loading)**: In addition to inline configurations, UniRTM MUST automatically discover and merge sidecar files from:
  - Project level: `.unirtm/plugins/*.toml`
  - User global level: `~/.config/unirtm/plugins/*.toml`
- **FR-005 (Sidecar Precedence)**: If a setting is defined in both a sidecar file and inline in `.unirtm.toml`, the inline `.unirtm.toml` setting MUST take precedence (Inline overrides Sidecar).
- **FR-006 (Write-Time Namespace Validation)**: `unirtm config set` MUST enforce that any key beginning with `plugins.` strictly specifies an installed or declared plugin ID, and that unreserved root keys are rejected.
- **FR-007 (AST-Preserving Updates)**: When updating `.unirtm.toml` or sidecar TOML files, the serializer MUST preserve existing comments, formatting, and unknown third-party keys.
- **FR-008 (File Locking)**: Configuration mutations MUST acquire an exclusive file lock (`.unirtm.toml.lock` / `config.toml.lock`) before modifying and releasing upon completion, with a configurable acquisition timeout (default: 5s).
- **FR-009 (Plugin Manifest Schema Validation)**: When a plugin manifest specifies a `contributes.configuration` schema with types and constraints, `unirtm config set` and `unirtm check` MUST validate values against the schema.
- **FR-010 (CLI Inspection Support)**: `unirtm config` and `unirtm config --global` MUST display a dedicated "Plugins & Extensions" section clearly distinguishing between active plugins, inline configs, and loaded sidecar files.

---

## 4. Key Entities & Schema

```toml
# ==============================================================================
# Complete UniRTM Configuration Structure (.unirtm.toml)
# ==============================================================================

# Core Reserved Tables
[tools]
go = "1.23.0"
node = "20.18.0"

[settings]
cache_ttl = "168h"

[format]
indent = "space"
indent_size = 2

# Plugin Declaration Table
[plugins]
"snowdreamtech/docker" = "1.0.0"
"community/deploy" = "0.2.1"

# Plugin Isolated Configuration Table
[plugins."snowdreamtech/docker".config]
compose_file = "docker-compose.yml"
auto_prune = true
registry = "harbor.internal"

[plugins."community/deploy".config]
cluster = "k8s-staging"
timeout = "45s"
```

---

## 5. Success Criteria *(mandatory)*

- **SC-001**: 100% of third-party plugin configuration keys are isolated under `[plugins."<id>".config]` or sidecars; zero plugin keys leak into the root configuration table.
- **SC-002**: Attempting to write a root key that is not in the reserved core list (e.g. `unirtm config set myplugin.setting value`) is automatically rejected with an actionable error.
- **SC-003**: Sidecar configurations in `.unirtm/plugins/<id>.toml` are discovered and parsed seamlessly in `< 5ms` during CLI startup.
- **SC-004**: Concurrent writes from 5 parallel processes across different plugin keys complete with zero data corruption or lost keys.
- **SC-005**: All existing comments in `.unirtm.toml` are preserved after executing `unirtm config set` operations.
