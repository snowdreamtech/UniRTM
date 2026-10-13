# Technical Plan: Network Profiles, Domestic Mirror Matrix & Cascading Overrides

**Feature**: Network Profiles, Domestic Mirror Matrix & Cascading Overrides  
**Branch**: `dev`  
**Spec**: [spec.md](./spec.md)  
**Tasks**: [tasks.md](./tasks.md)  
**Data Model**: [data-model.md](./data-model.md)  

---

## 1. Architectural Goals & Design Principles

1. **Strict Neutrality & Security**: The system never embeds any third-party public proxy servers or GitHub relay domains. Proxies default to disabled with empty URLs.
2. **Three-Tier Cascading Resolution**: Resolves effective registry endpoints using `Overrides (Granular User Override) > Active Profile (Selected Matrix) > Official Baseline (Upstream Default)`.
3. **Sandbox Environment Injection**: Injects proxy and mirror environment variables into executed child processes (shims and task runs) without mutating the host's global shell rc files (`~/.bashrc`, `/etc/environment`).
4. **1:1 CLI and Web Parity with Native i18n**: The CLI and Web UI share identical functional capabilities and support 100% native localization without awkward mixed-language text.

---

## 2. Go Backend Architecture (`internal/network/`)

### 2.1 Configuration Parser & Lifecycle Engine (`internal/network/`)
- `internal/network/types.go`: Core data types and DTOs.
- `internal/network/defaults.go`: Hardcoded factory blueprints (`official` and `cn`) used for initial seeding and factory resets (`ResetToDefault`). Includes vetted public mirror suggestions.
- `internal/network/manager.go`:
  - `GetStatus() ActiveNetworkStatus`: Computes effective registry URLs with source tracking.
  - `UseProfile(profileID string) error`: Switches the active profile.
  - `CreateProfile(p NetworkProfile) error`: Adds a new custom profile.
  - `SaveAsProfile(newID, name, description string) error`: Snapshots current effective state into a new profile.
  - `ResetProfile(profileID string) error`: Restores built-in profiles to factory defaults.
  - `DeleteProfile(profileID string) error`: Removes a custom profile with automatic fallback to `official`.
  - `SetOverride(ecosystem, url string) error`: Stores a single-ecosystem override.
  - `ResetOverride(ecosystem string) error`: Clears a single-ecosystem override.
  - `SetProxy(p ProxyConfig) error`: Validates and saves proxy settings.
  - `ExportEnvironment(target *[]string)`: Formats effective registries and proxies into environment variables.
- `internal/network/probe.go`:
  - Concurrent HTTP HEAD/GET probes measuring endpoint RTT with a 2.5s timeout.

### 2.2 CLI Subcommand Suite (`cmd/65.network.go`)
- `unirtm network status [--json]`: Displays active profile, proxy status, and effective registries.
- `unirtm network profile [list | use <id> | create <id> | save-as <id> | reset <id> | delete <id>]`: Profile lifecycle operations.
- `unirtm network registry [set <lang> <url> | reset <lang> | list]`: Granular ecosystem override management.
- `unirtm network proxy [status | on | off | set <url>]`: Proxy state controls.
- `unirtm network test`: Terminal-based concurrent latency probe table.

### 2.3 Web REST API Endpoints (`cmd/64.ui.go`)
- `GET  /api/network/status?scope=workspace|user`
- `POST /api/network/profile/use`
- `POST /api/network/profile/save`
- `POST /api/network/profile/reset`
- `POST /api/network/profile/delete`
- `POST /api/network/registry/override`
- `POST /api/network/registry/reset`
- `POST /api/network/proxy`
- `GET  /api/network/ping`

---

## 3. Frontend Architecture (Vue 3 + Vite)

### 3.1 Component Structure & Views
1. **SettingsView.vue**:
   - **Network Proxy Section**:
     - Master switch toggle (disabled by default).
     - Reactive input fields for HTTP, SOCKS5, and `NO_PROXY` bypass list.
     - Responsive help text on enterprise gateway usage.
   - **Mirror Matrix & Profile Management**:
     - Profile cards displaying status badges (`active`, `builtin`, `custom`).
     - Modal dialog for creating new profiles with live form validation.
     - Factory reset action for built-in profiles.
     - Ecosystem Registries Matrix table with preset dropdowns, custom inputs, override tags, restore buttons, and latency probe buttons.
2. **DashboardView.vue**:
   - Network diagnostics card reporting active profile, proxy status, and round-trip ping latency.

### 3.2 Full i18n Dictionary Expansion
- Complete coverage across `en-US.ts`, `zh-CN.ts`, and `zh-TW.ts`.
- Zero mixed-language parentheses.
- Strict consistency across button labels, badges, alerts, and placeholders.
