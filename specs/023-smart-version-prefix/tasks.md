# Tasks: Smart Version Prefix Normalization

**Input**: Design documents from `/specs/023-smart-version-prefix/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 Initialize unit testing module for the common backend utility (if it doesn't already exist).

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

- [x] T002 Ensure `internal/backend/` package is properly set up to host generic utilities independently of specific backend providers.

---

## Phase 3: User Story 1 - Graceful Version Resolution for Git Tags (Priority: P1) 🎯 MVP

**Goal**: Users shouldn't have to remember whether a specific tool or repository uses `v` prefixes for their versions. The tool should be smart enough to adapt dynamically.

**Independent Test**: Execute the unit tests `go test ./internal/backend -v -run TestNormalizeVersionPrefix`

### Tests for User Story 1

- [x] T003 [US1] Create unit tests for `NormalizeVersionPrefix` in `internal/backend/common_test.go` checking all edge cases (v to no-v, no-v to v, unaltered alias).

### Implementation for User Story 1

- [x] T004 [US1] Implement `NormalizeVersionPrefix(version string, requiresVPrefix bool) string` in `internal/backend/common.go`.
- [x] T005 [US1] Refactor `GithubHandler` (and any other backends relying on git tags) in `internal/backend/` to utilize `NormalizeVersionPrefix`.

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase N: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [x] T006 Run quickstart.md validation locally to verify output.
- [x] T007 Clean up any unused code related to manual version slicing.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Can start immediately.
- **Foundational (Phase 2)**: Quick checks for directory availability.
- **User Stories (Phase 3+)**: Can begin immediately after Setup/Foundational.
- **Polish (Final Phase)**: Depends on US1 being complete.

### User Story Dependencies

- **User Story 1 (P1)**: No dependencies on other stories.

### Parallel Opportunities

- Tests (T003) and implementation skeleton (T004) can be created in parallel, but test execution will fail until T004 is completed.
