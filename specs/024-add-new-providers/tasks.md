# Tasks: Add New Backend Providers

**Input**: Design documents from `specs/024-add-new-providers/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)

## Path Conventions

- All source files are located in `internal/provider/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- No shared infrastructure setup required. The `Provider` interface and package manager abstraction exist.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

- No foundational prerequisites needed. We are simply adding to the existing provider architecture.

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel.

---

## Phase 3: User Story 1 - Install PHP Tools via Composer (Priority: P1) 🎯 MVP

**Goal**: Support installing PHP ecosystem tools globally via Composer without modifying the system `~/.composer` path.

**Independent Test**: Can be fully tested by configuring `composer:phpstan` in `.unirtm.toml` and verifying its bin is available.

### Implementation for User Story 1

- [ ] T001 [P] [US1] Create `ComposerProvider` implementation in `internal/provider/composer.go`
- [ ] T002 [P] [US1] Create unit tests for `ComposerProvider` in `internal/provider/composer_test.go`
- [ ] T003 [US1] Register `composer` -> `NewComposerProvider()` in `internal/provider/registry.go`

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Install Dart Tools via Pub (Priority: P2)

**Goal**: Support installing Dart ecosystem tools globally via Pub without modifying the system `~/.pub-cache`.

**Independent Test**: Can be fully tested by configuring `pub:fvm` in `.unirtm.toml` and verifying its bin is available.

### Implementation for User Story 2

- [ ] T004 [P] [US2] Create `PubProvider` implementation in `internal/provider/pub.go`
- [ ] T005 [P] [US2] Create unit tests for `PubProvider` in `internal/provider/pub_test.go`
- [ ] T006 [US2] Register `pub` -> `NewPubProvider()` in `internal/provider/registry.go`

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - Install Lua and Haskell Tools (Priority: P3)

**Goal**: Support installing Lua and Haskell ecosystem tools globally via LuaRocks and Cabal in isolated directories.

**Independent Test**: Can be fully tested by configuring `luarocks:luacheck` and `cabal:shellcheck`.

### Implementation for User Story 3

- [ ] T007 [P] [US3] Create `LuaRocksProvider` implementation in `internal/provider/luarocks.go`
- [ ] T008 [P] [US3] Create unit tests for `LuaRocksProvider` in `internal/provider/luarocks_test.go`
- [ ] T009 [P] [US3] Create `CabalProvider` implementation in `internal/provider/cabal.go`
- [ ] T010 [P] [US3] Create unit tests for `CabalProvider` in `internal/provider/cabal_test.go`
- [ ] T011 [US3] Register `luarocks` and `cabal` in `internal/provider/registry.go`

**Checkpoint**: All user stories should now be independently functional

---

## Phase N: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T012 Validate end-to-end scenarios listed in `specs/024-add-new-providers/quickstart.md`
- [ ] T013 Verify that unsupported platforms for specific package managers (e.g., if Cabal isn't installed) correctly return graceful failures instead of panics.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup / Foundational (Phases 1 & 2)**: N/A.
- **User Stories (Phase 3+)**: Can proceed in parallel.
- **Polish (Final Phase)**: Depends on all desired user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: Independent.
- **User Story 2 (P2)**: Independent.
- **User Story 3 (P3)**: Independent.

### Parallel Opportunities

- All provider `.go` and `_test.go` files can be written completely in parallel (T001, T002, T004, T005, T007, T008, T009, T010).
- Registry registrations can be batched together at the end.

---

## Implementation Strategy

### Incremental Delivery

1. Implement `ComposerProvider` + tests. Register it. Deploy/Demo (MVP!)
2. Implement `PubProvider` + tests. Register it. Deploy/Demo
3. Implement `LuaRocksProvider` & `CabalProvider` + tests. Register them. Deploy/Demo
