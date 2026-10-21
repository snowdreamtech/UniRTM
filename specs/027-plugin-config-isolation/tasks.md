# Feature Tasks: Plugin & Extension Configuration Namespace Isolation

**Feature**: Plugin & Extension Configuration Namespace Isolation
**Branch**: `027-plugin-config-isolation`
**Spec**: [spec.md](./spec.md)
**Plan**: [plan.md](./plan.md)

---

## Phase 1: Setup & Data Modeling

**Purpose**: Establish data models, interfaces, and core reserved table definitions.

- [x] T001 [Setup] Define `ReservedCoreTables` map in `internal/config/reserved.go`
- [x] T002 [Setup] Add `Plugins map[string]PluginDeclaration` to `Config` struct in `internal/config/config.go`
- [x] T003 [Setup] Create `PluginDeclaration` struct with `Version` and `Config map[string]interface{}`
- [x] T004 [Setup] Write unit tests for reserved table check in `internal/config/reserved_test.go`

---

## Phase 2: Foundational Namespace Validation (Blocking Prerequisites)

**Purpose**: Enforce write-time and read-time namespace validation rules.

- [x] T005 [Foundational] Implement `ValidateKeyNamespace(key string) error` in `internal/config/reserved.go`
- [x] T006 [Foundational] Reject unreserved root-level keys in `ValidateKeyNamespace`
- [x] T007 [Foundational] Validate plugin ID formatting under `plugins."<plugin-id>"` namespace
- [x] T008 [Foundational] Add unit tests covering legal and illegal keys in `internal/config/reserved_test.go`

---

## Phase 3: User Story 1 - Inline Plugin Config & Namespace Isolation (Priority: P1) 🎯 MVP

**Goal**: Support reading and writing inline plugin configurations safely under `[plugins."<id>".config]`.

- [x] T009 [US1] Update `runConfigSet` in `cmd/14.config.go` to call `ValidateKeyNamespace`
- [x] T010 [US1] Support setting plugin configuration via `unirtm config set plugins."<id>".config.<key> <val>`
- [x] T011 [US1] Support retrieving plugin configuration via `unirtm config get plugins."<id>".config.<key>`
- [x] T012 [US1] Ensure `unirtm config` show command outputs plugin configuration under a dedicated section
- [x] T013 [US1] Verify that plugin A cannot overwrite or clobber plugin B's keys
- [x] T014 [US1] Add integration tests in `cmd/14.config_test.go` testing inline plugin config flows

---

## Phase 4: User Story 2 - Sidecar (External Split) Configuration (Priority: P1)

**Goal**: Allow complex plugin configs to live in `.unirtm/plugins/*.toml` and auto-merge at runtime.

- [x] T015 [US2] Implement sidecar directory scanner in `internal/config/sidecar.go` (`.unirtm/plugins/*.toml`)
- [x] T016 [US2] Implement global sidecar directory scanner (`~/.config/unirtm/plugins/*.toml`)
- [x] T017 [US2] Implement deep merge logic between sidecar files and `.unirtm.toml` in `internal/config/loader.go`
- [x] T018 [US2] Enforce precedence rule: inline `.unirtm.toml` overrides sidecar defaults
- [x] T019 [US2] Add unit tests for sidecar discovery, merging, and precedence in `internal/config/sidecar_test.go`

---

## Phase 5: User Story 3 - Write-Time Sandbox & Concurrency File Locking (Priority: P2)

**Goal**: Guarantee thread-safe, process-safe atomic writes and comment preservation.

- [x] T020 [US3] Implement cross-platform file locking (`.unirtm.toml.lock`) in `internal/config/lock.go`
- [x] T021 [US3] Integrate file locking into `runConfigSet` with automatic timeout handling
- [x] T022 [US3] Ensure comment preservation when updating TOML files via AST-aware serializer
- [x] T023 [US3] Add unit tests simulating concurrent configuration writers in `internal/config/lock_test.go`
- [x] T024 [US3] Add test ensuring comments and whitespace are preserved after config modifications

---

## Phase 6: User Story 4 - Manifest Schema Validation & Inspection (Priority: P2)

**Goal**: Validate plugin configuration values against plugin manifest schemas at write-time.

- [x] T025 [US4] Define `PluginManifest` and `ContributedConfiguration` structures in `internal/config/schema.go`
- [x] T026 [US4] Implement type checking (`string`, `int`, `bool`, `list`) against manifest schema
- [x] T027 [US4] Integrate schema validation check into `runConfigSet`
- [x] T028 [US4] Add unit tests validating schema compliance in `internal/config/schema_test.go`

---

## Phase 7: Polish & Documentation

**Purpose**: Finalize CLI experience, error messages, and documentation.

- [x] T029 [Docs] Update documentation comparing plugin configuration architecture to VS Code / Cargo
- [x] T030 [Docs] Provide plugin developer guide in `docs/plugins/configuration.md`
- [x] T031 [UI] Expose plugins and active sidecars in `unirtm ui` settings and status views
- [x] T032 [Verify] Run full test suite (`go test ./...`) to ensure 100% pass rate
