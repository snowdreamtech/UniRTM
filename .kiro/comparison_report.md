# Mise vs UniRTM Configuration Comparison

This report provides a granular comparison of configuration capabilities between `mise` and `UniRTM`. Based on the latest source code analysis from both repositories, here is the current parity status.

## 1. Environment Directives (`[env]`)

Mise uses a sophisticated directive system for environment manipulation. UniRTM has recently been upgraded to support the most critical of these.

| Feature | Mise Syntax | UniRTM Support | Status | Notes |
| :--- | :--- | :---: | :---: | :--- |
| **Simple KV** | `KEY = "VAL"` | **Yes** | ✅ | Standard key-value injection. |
| **Shell Expansion** | `KEY = "$HOME"` | **Yes** | ✅ | Supports `$VAR` and `${VAR}` in all env values. |
| **Task Timeout** | `timeout = 30` | **Yes** | ✅ | Supports global `task_timeout` and per-task `timeout`. |
| **Task Output** | `output = "prefix"` | **Yes** | ✅ | Supports `plain` and `prefix` output styles. |
| **Tera Templates** | `KEY = "{{ env.HOME }}"` | **Yes** | ✅ | UniRTM uses `pongo2` (Jinja2-like), highly compatible with Tera. |
| **Dotenv Files** | `_.file = ".env"` | **Yes** | ✅ | Implemented in `internal/config/loader.go`. |
| **Path Prepending** | `_.path = "bin"` | **Yes** | ✅ | Supports both string and list formats in UniRTM; supports `$VAR`. |
| **Script Sourcing** | `_.source = "src.sh"` | **Yes** | ✅ | Implemented for POSIX, Fish, and PowerShell; supports `$VAR`. |
| **Required Vars** | `required = true` | **Yes** | ✅ | Fails if var is missing; supports custom help text. |
| **Secret Redacting** | `redact = true` | **Yes** | ✅ | Values are replaced with `[REDACTED]` in shell output. |
| **Unset Variable** | `KEY = { rm = true }` | **Yes** | ✅ | Removes the variable from the resolved environment. |
| **Python Venv** | `_.python_venv = ".venv"` | **Yes** | ✅ | Automatically activates venv and sets `VIRTUAL_ENV`. |
| **Modules/Vfox** | `_.module = "..."` | No | ❌ | UniRTM has its own provider system instead of vfox modules. |
| **Age Encryption** | `_.age = "..."` | No | ❌ | Experimental in Mise; not planned for UniRTM. |

## 2. Global Settings (`[settings]`)

Settings control the behavior of the tool itself. UniRTM targets the most frequently used settings for developer experience.

| Feature | Mise Key | UniRTM Support | Status | Notes |
| :--- | :--- | :---: | :---: | :--- |
| **GitHub Proxy** | `github_proxy` | **Yes** | ✅ | Fully aligned. |
| **GitHub Token** | `github_token` | **Yes** | ✅ | Fully aligned. |
| **Concurrency** | `jobs` | **Yes** | ✅ | UniRTM uses `concurrency` (mapped from `jobs` during migration). |
| **HTTP Timeout** | `http_timeout` | **Yes** | ✅ | Recently added to UniRTM. |
| **Experimental** | `experimental` | **Yes** | ✅ | Recently added to UniRTM. |
| **Lockfile** | `lockfile` | **Yes** | ✅ | Both support opt-in tool version locking. |
| **Strict Lock** | `locked` | **Yes** | ✅ | Useful for CI environments. |
| **Cache Dir** | `cache_dir` | **Yes** | ✅ | Supported via `settings.cache_dir`. |
| **Data Dir** | `data_dir` | **Yes** | ✅ | Supported via `settings.data_dir`. |
| **Always Keep DL** | `always_keep_download` | **Yes** | ✅ | Now supported; preserves artifacts in downloads directory. |
| **Auto Install** | `auto_install` | **Yes** | ✅ | Fully aligned; triggers on `run` and `exec`. |
| **Asdf Compat** | `asdf_compat` | No | ❌ | UniRTM uses its own modern logic exclusively. |
| **Color Control** | `color` | **Yes** | ✅ | Fully aligned; supports auto, always, and never. |

## 3. Advanced Comparison (Unique Capabilities)

### Mise Strengths
- **Shell-style expansion**: Supports `$FOO` expansion within the TOML itself after template rendering.
- **Environment Caching**: [Experimental] Caches computed environments to disk for ultra-fast nested calls.
- **Ceiling Paths**: Allows stopping config discovery at specific directory levels.

### UniRTM Alignment Level
UniRTM has achieved **~90% functional parity** for the average developer's daily workflow. The missing features are primarily edge cases (Age encryption), legacy compatibility (asdf_compat), or advanced plugin architectures (vfox modules).

---
*Report generated based on Mise source analysis (`settings.toml` & `src/config/env_directive/mod.rs`) and UniRTM current state.*
