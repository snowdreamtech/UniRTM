# Feature Tasks: Network Profiles, Domestic Mirror Matrix & Cascading Overrides

**Feature**: Network Profiles, Domestic Mirror Matrix & Cascading Overrides
**Branch**: `dev`
**Spec**: [spec.md](./spec.md)
**Plan**: [plan.md](./plan.md)
**Data Model**: [data-model.md](./data-model.md)

---

## Phase 1: Go Backend Data Structures & Network Engine

**Purpose**: Establish data models, defaults, resolution engine, lifecycle CRUD, and probe capabilities.

- [x] T001 [Setup] Create `internal/network/types.go` defining `NetworkConfig`, `NetworkProfile`, `ProxyConfig`, and `EffectiveRegistry`.
- [x] T002 [Setup] Create `internal/network/defaults.go` with factory built-ins (`official` and `cn` matrix) and vetted public mirror recommendations.
- [x] T003 [Core] Implement `internal/network/manager.go`:
  - Cascading resolution engine: `GetEffectiveStatus` applying `Overrides > ActiveProfile > Default`.
  - Profile lifecycle CRUD: create, edit, save-as (`SaveAs`), delete, and factory reset (`ResetToDefault`).
  - Per-ecosystem override support: `SetOverride` and `ResetOverride`.
  - Proxy management: validation of schemes/endpoints and atomic configuration storage.
- [x] T004 [Core] Implement `internal/network/probe.go`: concurrent network latency probe with a 2.5s timeout.
- [x] T005 [Tests] Write Go unit tests in `internal/network/manager_test.go` verifying cascading resolution, factory reset, proxy defaults, and file persistence.

---

## Phase 2: CLI Subcommand Implementation (`unirtm network`)

**Purpose**: Provide a full-featured CLI suite for terminal users and automation scripts.

- [x] T006 [CLI] Create `cmd/65.network.go` and register the `unirtm network` parent command.
- [x] T007 [CLI] Implement `unirtm network status` with tabular and `--json` machine-readable output.
- [x] T008 [CLI] Implement `unirtm network profile [list | use | create | save-as | reset | delete]` command group.
- [x] T009 [CLI] Implement `unirtm network registry [set | reset | list]` command group.
- [x] T010 [CLI] Implement `unirtm network proxy [on | off | set]` command group.
- [x] T011 [CLI] Implement `unirtm network test` for concurrent latency table probing in the terminal.
- [x] T012 [Core] Integrate environment sandbox injection via `ExportEnvironment` in `internal/network`.
- [x] T013 [Tests] Write CLI unit tests in `cmd/65.network_test.go`.

---

## Phase 3: Web Backend REST API

**Purpose**: Expose REST endpoints for Web UI interactions and real-time state management.

- [x] T014 [API] Register `/api/network/*` routes in `cmd/64.ui.go`.
- [x] T015 [API] Implement status retrieval, profile switching, saving, resetting, per-language overrides, proxy configuration, and ping endpoints.
- [x] T016 [Tests] Write API integration tests in `cmd/64.ui_network_test.go`.

---

## Phase 4: Frontend UI Components & i18n Localization

**Purpose**: Provide an intuitive, responsive, and beautifully localized GUI dashboard and settings view.

- [x] T017 [i18n] Add complete `network.*` key dictionary to `frontend/src/i18n/types.ts`, `zh-CN.ts`, `zh-TW.ts`, and `en-US.ts`.
- [x] T018 [UI] Upgrade Network settings section in `frontend/src/views/SettingsView.vue`:
  - Network Proxy subsection: master toggle, HTTP/SOCKS5 inputs, bypass list.
  - Profile Management: profile cards, theme-adaptive modal dialog, factory reset.
  - Ecosystem Registry Matrix table: recommended mirror dropdowns, custom inputs, override badges, restore buttons.
- [x] T019 [UI] Integrate network diagnostics card into `frontend/src/views/DashboardView.vue` with real-time probe latency.
- [x] T020 [Dev] Implement dev middleware in `frontend/vite.config.ts` for Vite dev server parity.
- [x] T021 [Build] Verify frontend production bundle with `npm run build` (`vue-tsc && vite build`).

---

## Phase 5: End-to-End Verification & Documentation

**Purpose**: Ensure parity, reliability, and security compliance.

- [x] T022 [Verify] Cross-validate CLI and Web two-way synchronization.
- [x] T023 [Verify] Verify pure native mother-tongue localization with zero bilingual parenthesis mixing.
- [x] T024 [Verify] Run full test suite with Go 1.27.2 and verify 100% test pass rate.
