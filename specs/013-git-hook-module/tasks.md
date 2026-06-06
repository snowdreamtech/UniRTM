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

## Phase 5: Testing and Bugfixes

**Purpose**: Fix critical edge-cases found in review and ensure robust test coverage.

- [x] T010 [US5] Fix shell injection vulnerability in `internal/hook/native.go`
- [x] T011 [US5] Fix `.git/hooks` path resolution in `internal/hook/install.go` using `git rev-parse`
- [x] T012 [US5] Define explicit priority order in `internal/hook/router.go`
- [x] T013 [US5] Add unit tests for `install.go`
- [x] T014 [US5] Add unit tests for `native.go`
- [x] T015 [US5] Add unit tests for `router.go`

## Phase 6: Security Hardening

**Purpose**: Fix security vulnerabilities found in code review.

- [x] T016 [US6] Add `ValidateHookName()` whitelist in `internal/hook/validate.go`
- [x] T017 [US6] Call `ValidateHookName()` in `cmd/62.hook.go` entry points
- [x] T018 [US6] Fix Bridge Script nested-quote issue in `internal/hook/install.go`
- [x] T019 [US6] Fix `plan.md` doc: `source` → `.`, fix bridge script example

## Phase 7: Cross-Platform Compatibility

**Purpose**: Ensure Windows / Linux / macOS / POSIX shell compatibility.

- [x] T020 [US7] Create `internal/hook/shell.go` with cross-platform `GetShell()` helper
- [x] T021 [US7] Refactor `NativeRunner.Run` to use `GetShell()` instead of hardcoded `"sh"`
- [x] T022 [US7] Refactor `NativeRunner.Run` to accept explicit `dir` param (remove os.Chdir dependency)

## Phase 8: Test Suite Improvements

**Purpose**: Fix thread-safety issues and improve coverage.

- [x] T023 [US8] Rewrite `install_test.go`: verify file content and `0755` permissions
- [x] T024 [US8] Rewrite `native_test.go`: pass `dir` explicitly, remove `os.Chdir`
- [x] T025 [US8] Add `validate_test.go`: whitelist allow/deny test cases
- [x] T026 [US8] Add `shell_test.go`: verify `GetShell()` per platform
- [x] T027 [US8] Add `husky_test.go`: verify path traversal rejection
