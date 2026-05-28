---
description: "Task list template for feature implementation"
---

# Tasks: Built-in Cross-Platform Shell (mvdan-shell)

**Input**: Design documents from `/specs/003-mvdan-shell/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 Fetch and import dependencies: `mvdan.cc/sh/v3/interp` and `mvdan.cc/sh/v3/syntax` in `go.mod`
- [x] T002 Update `go mod tidy`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

- [x] T003 Remove system shell fallbacks (`cmd.exe /c` and `sh -c`) in `internal/task/native.go`
- [x] T004 Build shell environment parsing via `interp.Env(expand.ListEnviron(fullEnv...))`

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Seamless POSIX Execution on Windows (Priority: P1) 🎯 MVP

**Goal**: Windows users can run tasks written in POSIX syntax smoothly.

**Independent Test**: Can be tested by running tasks with `if` statements.

### Implementation for User Story 1

- [x] T005 [P] [US1] Parse `script` string with `syntax.NewParser().Parse()` in `internal/task/native.go`
- [x] T006 [P] [US1] Execute AST via `interp.New().Run()` in `internal/task/native.go`
- [x] T007 [US1] Handle context timeouts and errors properly and update assertions in `internal/task/native_test.go`
- [x] T008 [US1] Revert Python polyfill scripts back to POSIX (e.g. `audit:govulncheck`) in `.unirtm.toml`

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Consistent Output Streaming (Priority: P2)

**Goal**: Standard output and error output preserve styling and prefix interleaving natively.

**Independent Test**: Running long running outputs displays text simultaneously with correct prefix writing without buffering.

### Implementation for User Story 2

- [x] T009 [US2] Bind STDIN, STDOUT, STDERR using `interp.StdIO` mapped to existing `buf` and `prefixWriter` in `internal/task/native.go`

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [x] T010 Run unit tests via `go test ./internal/task/...`
- [x] T011 Verify Windows CI execution using `unirtm run audit:govulncheck` locally

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
- **Polish (Final Phase)**: Depends on all desired user stories being complete
