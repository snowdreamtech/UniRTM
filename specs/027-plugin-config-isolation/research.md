# Research & Precedents: Plugin Configuration Isolation

**Feature**: `027-plugin-config-isolation`

---

## 1. Industry Precedents

### VS Code (`contributes.configuration`)
* **Mechanism**: VS Code extensions declare settings under their extension ID namespace (e.g. `prettier.tabWidth`, `eslint.enable`).
* **Protection**: VS Code rejects extensions trying to contribute to core namespaces like `editor.*` or `workbench.*` without special system permissions.
* **Takeaway for UniRTM**: Enforce a mandatory `<plugin-id>` prefix for any non-core setting.

### Cargo (`[package.metadata.*]`)
* **Mechanism**: Cargo reserves tables like `[dependencies]`, `[features]`, `[profile]`. Third-party tools are strictly confined to `[package.metadata.<tool_name>]` (e.g. `[package.metadata.wasm-pack]`).
* **Takeaway for UniRTM**: Formally define `ReservedCoreTables` so that plugins cannot create new top-level tables in `.unirtm.toml`.

### Kubernetes CRDs (`spec` Namespacing)
* **Mechanism**: Custom Resource Definitions isolate vendor configurations under group-version-kind namespaces to prevent collisions across multiple operators.

---

## 2. File Locking & Concurrency
* **Cross-Platform File Locks**:
  * On Unix: `syscall.Flock(fd, syscall.LOCK_EX|syscall.LOCK_NB)`
  * On Windows: `windows.LockFileEx`
  * Using a dedicated `.unirtm.toml.lock` companion file prevents corrupting the primary file while ensuring mutual exclusion across concurrent CLI or plugin processes.
