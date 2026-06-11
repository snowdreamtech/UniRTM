---

description: "Task list for UniRTM Hook Execution Wrapper implementation"
---

# Tasks: UniRTM Hook Execution Wrapper

**Input**: Design documents from `/specs/022-hook-exec/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, quickstart.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 Create `cmd/63.hook-exec.go` file with basic cobra command scaffold

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T002 Register `hook-exec` subcommand under the `hookCmd` in `cmd/63.hook-exec.go`

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - Simplified Pre-commit Configuration (Priority: P1) 🎯 MVP

**Goal**: Use `unirtm hook-exec` as a wrapper without complex shell scripts, identifying base command and file args.

**Independent Test**: Running `unirtm hook-exec echo --flag dummy.txt` works natively.

### Implementation for User Story 1

- [x] T003 [US1] Implement right-to-left file argument splitting logic via `os.Lstat` in `cmd/63.hook-exec.go`
- [x] T004 [US1] Route non-chunked execution directly to `runExec(cmd, args)` in `cmd/63.hook-exec.go`
- [x] T005 [P] [US1] Create unit tests for argument splitting logic in `cmd/63.hook-exec_test.go`

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently on Unix systems.

---

## Phase 4: User Story 2 - Automatic Argument Chunking on Windows (Priority: P1)

**Goal**: Automatically chunk large file lists when command length exceeds 7000 characters on Windows.

**Independent Test**: Running a massive argument list on Windows gets batched properly instead of failing with OS error.

### Implementation for User Story 2

- [x] T006 [US2] Calculate total command line length accurately in `cmd/63.hook-exec.go`
- [x] T007 [US2] Implement batching loop for files when `totalLen >= 7000` on Windows in `cmd/63.hook-exec.go`
- [x] T008 [US2] Execute each batch sequentially by calling `runExec` inside the loop in `cmd/63.hook-exec.go`

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [x] T009 Validate `quickstart.md` test scenarios manually

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
- **Polish (Final Phase)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P1)**: Extends User Story 1 (adds chunking) - Should be implemented sequentially.

### Parallel Opportunities

- Unit tests for US1 can be built in parallel.
