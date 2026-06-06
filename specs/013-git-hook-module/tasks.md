# Tasks: Git Hook Module

**Input**: Design documents from `specs/013-git-hook-module/`
**Prerequisites**: plan.md

## Phase 1: Core Interface & Setup

**Purpose**: Core infrastructure for the hook routing engine.

- [x] T001 [US1] Create `HookRunner` interface and error types in `internal/hook/runner.go`
- [x] T002 [US1] Create bridge script template and `InstallHook` helper in `internal/hook/install.go`
- [x] T003 [US1] Create `Router` loop logic in `internal/hook/router.go`

## Phase 2: Native Runner Implementation

**Purpose**: Provide fallback default hook execution using `.unirtm.toml`.

- [x] T004 [US2] Implement `NativeRunner` logic to read `[hooks]` from config in `internal/hook/native.go`

## Phase 3: External Extensions (MVP)

**Purpose**: Support the most common hook frameworks.

- [x] T005 [US3] Implement `PreCommitRunner` (detects `.pre-commit-config.yaml`) in `internal/hook/precommit.go`
- [x] T006 [US3] Implement `HuskyRunner` (detects `.husky/`) in `internal/hook/husky.go`
- [x] T007 [US3] Implement `LefthookRunner` (detects `lefthook.yml`) in `internal/hook/lefthook.go`

## Phase 4: CLI Commands

**Purpose**: Expose the functionality to the user.

- [x] T008 [US4] Add `unirtm hook` root command and `install`, `run` subcommands in `cmd/hook.go`
- [x] T009 [US4] Wire up the `cmd/hook.go` into `cmd/root.go`

## Dependencies & Execution Order

- Phase 1 must be completed first.
- Phase 2 and 3 can be done in parallel once Phase 1 is done.
- Phase 4 depends on all previous phases.
