---
description: "Task list template for feature implementation"
---

# Tasks: Native Lua & LuaRocks Provider

**Input**: Design documents from `/specs/025-lua-provider/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [ ] T001 Register "lua" provider in `internal/provider/registry.go`
- [ ] T002 Create stub provider implementation in `internal/provider/lua.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T003 Implement OS/Arch resolution mapping helper in `internal/provider/lua.go` (maps to `_Linux54_64_bin.tar.gz`, `_Win64_bin.zip`, etc., or dyne/luabinaries mirror for arm64)

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - Install Pure Lua Environment (Priority: P1) 🎯 MVP

**Goal**: Install Lua directly from the official LuaBinaries source without relying on ASDF.

**Independent Test**: Can be fully tested by configuring `lua = "5.4.2"` in `.unirtm.toml` and verifying `lua -v`.

### Implementation for User Story 1

- [ ] T004 [US1] Implement `Install` method logic to fetch binary archive using `net/http` in `internal/provider/lua.go`
- [ ] T005 [US1] Implement archive extraction logic (`zip`/`tar.gz`) into the target installation directory in `internal/provider/lua.go`
- [ ] T006 [US1] Implement `GetBinPaths` and `ListExecutables` to correctly identify `lua` and `luac` executables in the installed directory in `internal/provider/lua.go`
- [ ] T007 [US1] Add unit test for URL resolution and platform mapping in `internal/provider/lua_test.go`

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently (pure Lua works).

---

## Phase 4: User Story 2 - Automatic LuaRocks Bootstrapping (Priority: P1)

**Goal**: LuaRocks automatically bootstrapped and installed alongside Lua to manage packages natively.

**Independent Test**: Can be fully tested by configuring `"luarocks:busted" = "2.0.0"` in `.unirtm.toml` and verifying package installation.

### Implementation for User Story 2

- [ ] T008 [US2] Implement `bootstrapLuaRocks` helper function in `internal/provider/lua.go` to download `luarocks` tarball.
- [ ] T009 [US2] Implement logic in `bootstrapLuaRocks` to extract tarball and execute `./configure` + `make` (Unix) or `install.bat` (Windows) inside `internal/provider/lua.go`
- [ ] T010 [US2] Hook `bootstrapLuaRocks` into the end of the `Install` method lifecycle in `internal/provider/lua.go`
- [ ] T011 [US2] Update `ListExecutables` to also scan for `luarocks` and `luarocks-admin` in `internal/provider/lua.go`

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently. Lua and LuaRocks are natively available.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T012 Security hardening: Ensure downloaded tarball extraction is safe from zip slip vulnerabilities in `internal/provider/lua.go`
- [ ] T013 Add robust error handling and debug logging for the `make`/`install.bat` execution phases in `internal/provider/lua.go`
- [ ] T014 Run validation scenarios from `specs/025-lua-provider/quickstart.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - US1 and US2 are sequential because US2 (LuaRocks bootstrap) requires US1 (Lua engine) to be installed.
- **Polish (Final Phase)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2)
- **User Story 2 (P1)**: Can start after User Story 1 (relies on the engine being present to bootstrap LuaRocks)

### Parallel Opportunities

- Due to the sequential nature of installing the engine first and bootstrapping the package manager second, parallelization is limited mostly to writing tests (T007) while implementing extraction (T005).

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 & 2
2. Complete Phase 3 (US1)
3. **STOP and VALIDATE**: Test Lua engine installation independently
4. Demo: `unirtm exec lua -- lua -v`

### Incremental Delivery

1. Deliver Lua engine natively first.
2. Add LuaRocks bootstrap logic onto the end of the install phase.
3. Validate complete package management integration.
