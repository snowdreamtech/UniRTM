---
description: "Task list for hook install hybrid resolution implementation"
---

# Tasks: Hook Install Hybrid Resolution

**Input**: Design documents from `/specs/019-hook-install-hybrid-resolution/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, quickstart.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 Open `internal/hook/install.go` to review the current overwrite-based installation logic

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

- [x] T002 Implement regex/string constants for `BEGIN UNIRTM MANAGED BLOCK` and `END UNIRTM MANAGED BLOCK` in `internal/hook/install.go`
- [x] T003 Implement the environment injection payload constant in `internal/hook/install.go` according to `data-model.md`

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Non-Destructive Hook Installation (Priority: P1) 🎯 MVP

**Goal**: As a developer with existing custom Git hooks, I want unirtm hook install to safely co-exist with my existing scripts rather than overwriting them, so that I don't lose my manual configurations.

**Independent Test**: Can be tested by creating a custom .git/hooks/pre-commit file, running unirtm hook install, and verifying the original custom logic is still present and runs.

### Implementation for User Story 1

- [x] T004 [US1] Implement `readExistingHook` function to check for shebang lines in `internal/hook/install.go`
- [x] T005 [US1] Implement `injectOrUpdateBlock` function to replace the block if it exists, or insert it after the shebang if it doesn't in `internal/hook/install.go`
- [x] T006 [US1] Refactor `InstallBridgeScript` in `internal/hook/install.go` to use the new non-destructive injection logic instead of `os.WriteFile`

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Headless AI/CI Environment Execution (Priority: P1)

**Goal**: As an automated CI/CD runner or AI Agent operating in a sandboxed shell, I want the Git hooks to automatically find and load the UniRTM environment without relying on user profile scripts.

**Independent Test**: Can be tested by running git commit in a completely empty environment and verifying it still successfully triggers the UniRTM hooks.

### Implementation for User Story 2

- [x] T007 [US2] Update the injection payload to use `git rev-parse --show-toplevel` for local `unirtm` binary resolution in `internal/hook/install.go`
- [x] T008 [US2] Ensure the payload falls back to global path if local binary is missing in `internal/hook/install.go`

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - Batch Installation (Priority: P2)

**Goal**: Support `-a/--all` batch installation for all active hooks.

- [x] T011 [US3] Add `-a/--all` flag to `hookInstallCmd` in `cmd/62.hook.go`
- [x] T012 [US3] Implement `InstallAllBridgeScripts` function in `internal/hook/install.go` to iterate over `.git/hooks/` and apply `InstallBridgeScript` to non-sample files

---

## Phase N: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [x] T009 Write unit tests for injection logic in `internal/hook/install_test.go`
- [x] T010 Run `quickstart.md` validation tests
- [x] T013 [FR-005] Add explicit unit test in `internal/hook/install_test.go` to verify `hook run` is not injected
- [x] T014 [SC-002] Create an automated integration test script or add a test case to simulate headless `env -i` execution of the injected script

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
- **Polish (Final Phase)**: Depends on all desired user stories being complete
