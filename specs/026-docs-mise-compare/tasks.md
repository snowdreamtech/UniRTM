# Tasks: Documentation Upgrade: Aligning and Surpassing `mise`

**Input**: Design documents from `/specs/026-docs-mise-compare/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, quickstart.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 Verify local VitePress documentation environment can be run via `npm run docs:dev`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

- [x] T002 Update sidebar configuration in `docs/.vitepress/config.mts` to include `comparisons.md` under `guide` for both EN and ZH locales.
- [x] T003 Implement `navigator.language` auto-redirect logic via `<head>` script injection inside `docs/.vitepress/config.mts`.

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Evaluate UniRTM against existing tools (Priority: P1) 🎯 MVP

**Goal**: Provide a clear, side-by-side comparison with UniRTM against legacy and modern tools (nvm, gvm, pyenv/pipx, asdf, mise, direnv).

**Independent Test**: Can be tested by navigating to the new "Comparisons" section in the docs and verifying clear, structured contrast points for each tool, explicitly highlighting UniRTM's superiority over `mise`.

### Implementation for User Story 1

- [x] T004 [P] [US1] Create and implement `docs/guide/comparisons.md` with comparisons for nvm, gvm, pyenv, asdf, mise, direnv, ensuring UniRTM's superiority over `mise` is highlighted.
- [x] T005 [P] [US1] Create and implement `docs/zh/guide/comparisons.md` with translated comparisons based on `research.md`.

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Discover UniRTM's Architectural Superiority (Priority: P1)

**Goal**: Ensure technical evaluators reading the introduction understand how UniRTM's architecture surpasses competitors (native Go architecture, zero-pollution, built-in MCP).

**Independent Test**: Can be tested by reviewing the Introduction and Getting Started pages for highlighted superiority points, and verifying automatic redirect for Chinese browsers to `/zh/`.

### Implementation for User Story 2

- [x] T006 [P] [US2] Update `docs/guide/introduction.md` to highlight 100% Native Architecture, Zero-Pollution, and MCP Capabilities.
- [x] T007 [P] [US2] Update `docs/zh/guide/introduction.md` to highlight 100% Native Architecture, Zero-Pollution, and MCP Capabilities.
- [x] T008 [P] [US2] Update `docs/guide/getting-started.md` to emphasize simplified installation without external plugins.
- [x] T009 [P] [US2] Update `docs/zh/guide/getting-started.md` to emphasize simplified installation without external plugins.

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [x] T010 Run `npm run docs:build` locally to verify there are no broken links in the VitePress build.
- [x] T011 Perform visual review using `quickstart.md` validation scenarios.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **Polish (Final Phase)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories

### Parallel Opportunities

- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, all user stories can start in parallel (if team capacity allows)
- Translation updates can happen in parallel with English updates.

---

## Parallel Example: User Story 1

```bash
# Launch all files for User Story 1 together:
Task: "Create and implement docs/guide/comparisons.md"
Task: "Create and implement docs/zh/guide/comparisons.md"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Test User Story 1 independently
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 → Test independently → Deploy/Demo
4. Each story adds value without breaking previous stories
