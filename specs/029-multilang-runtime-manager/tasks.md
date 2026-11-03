# Feature Tasks: Unified Multi-Language Runtime Manager

**Feature**: Unified Multi-Language Runtime Manager (Replacing NVM, GVM, and Pyenv)
**Branch**: `dev`
**Spec**: [spec.md](./spec.md)
**Plan**: [plan.md](./plan.md)

---

## Phase 1: Legacy Version File Compatibility Engine

**Purpose**: Enable UniRTM to transparently detect and adopt `.nvmrc`, `.node-version`, `.go-version`, `.python-version`, and `.tool-versions` across directories without manual migration.

- [x] T001 [Config] Create `internal/config/legacy.go` to parse `.nvmrc`, `.node-version`, `.go-version`, `.python-version`, and `.tool-versions`.
- [x] T002 [Config] Integrate legacy detection into `internal/config/loader.go` in `LoadHierarchy`, preserving config priority and recording legacy file origins.
- [x] T003 [Tests] Write unit tests in `internal/config/legacy_test.go` verifying multi-level directory traversal and fallback precedence.

---

## Phase 2: Smart Semver & Alias Resolution (LTS, Latest, Prefix)

**Purpose**: Allow developers to install and use runtimes using intuitive expressions like `node@lts`, `node@20`, `go@latest`, `python@3.12`.

- [x] T004 [Service] Enhance version alias and prefix resolver in `internal/service/` supporting `lts`, `latest`, `lts/<codename>`, and major/minor prefix matching for Node, Go, and Python.
- [x] T005 [Provider] Ensure native providers for Node, Go, and Python verify pre-built distributions and checksums across Linux, macOS, and Windows.
- [x] T006 [Tests] Add unit tests for alias and prefix resolution in `internal/service/installation_resolve_test.go`.

---

## Phase 3: CLI Subcommand Suite Polish (`install`, `use -g`, `list`, `current`, `ls-remote`)

**Purpose**: Elevate CLI ergonomics so developers experience a complete superset of NVM, GVM, and Pyenv commands.

- [x] T007 [CLI] Enhance `unirtm list` / `unirtm ls` in `cmd/8.list.go` to display installed runtimes, active version marker `*`, and the origin source (e.g., `(set by .nvmrc)`, `(set by .unirtm.toml)`, `(set by global)`).
- [x] T008 [CLI] Polish `unirtm current` in `cmd/57.current.go` to report all active runtimes in the current working directory with origin indicators.
- [x] T009 [CLI] Enhance `unirtm use -g` / `--global` in `cmd/17.use.go` to update global default runtimes in `~/.config/unirtm/unirtm.toml` with immediate shim refresh.
- [x] T010 [CLI] Verify `unirtm ls-remote` in `cmd/52.ls-remote.go` supporting `--lts` and sorted releases for Node, Go, and Python.
- [x] T011 [Tests] Write and run CLI unit tests covering list, current, and use commands.

---

## Phase 4: Web Backend REST API Endpoints

**Purpose**: Expose comprehensive REST endpoints in `cmd/64.ui.go` for full Web UI runtime management.

- [x] T012 [API] Implement `GET /api/runtimes/status` returning all supported runtimes, installed versions, active version, origin source, and detected legacy files.
- [x] T013 [API] Implement `GET /api/runtimes/remote` to query available remote versions with LTS flags.
- [x] T014 [API] Implement `POST /api/runtimes/install`, `POST /api/runtimes/uninstall`, `POST /api/runtimes/use`, and `POST /api/runtimes/migrate-legacy`.
- [x] T015 [Tests] Write API test cases in `cmd/64.ui_runtimes_test.go`.

---

## Phase 5: Web UI Frontend Runtimes View & Native Localization

**Purpose**: Provide an interactive, modern, and natively localized Web workbench for runtime governance.

- [x] T016 [i18n] Add complete `runtimes.*` dictionary keys to `frontend/src/i18n/types.ts`, `zh-CN.ts`, `zh-TW.ts`, and `en-US.ts`.
- [x] T017 [UI] Create `frontend/src/views/RuntimesView.vue` with runtime cards (Node.js, Go, Python, Rust), active version selectors, legacy file alert banner, and remote install modal.
- [x] T018 [UI] Register the "Runtimes" navigation tab and route in `frontend/src/router/` and `frontend/src/App.vue`.
- [x] T019 [Build] Verify frontend production bundle compilation with `npm run build` in `frontend/`.

---

## Phase 6: End-to-End Verification & Commit Milestones

**Purpose**: Verify cross-platform functionality, test suites, and deliver clean, structured git commits.

- [x] T020 [Verify] Test legacy takeover in sample directories with `.nvmrc` and `.go-version`.
- [x] T021 [Verify] Execute full Go test suite (`go test ./...`) ensuring zero regressions.
- [x] T022 [Commit] Execute step-by-step git commits following Conventional Commits format.
