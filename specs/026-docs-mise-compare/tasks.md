---
description: "Task list for UniRTM Documentation Upgrade"
---

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

- [X] T001 Identify missing configuration for Mermaid.js in VitePress (if any) in `docs/.vitepress/config.mts`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

- [X] T002 Identify all CLI commands currently existing in UniRTM codebase to prepare for docs/cli/overview.md
- [ ] T003 Read `docs/guide/architecture.md` (and related architecture docs) to establish comparison baseline

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Deep Architecture Dive (Priority: P1) 🎯 MVP

**Goal**: As an evaluator, I want to read a detailed architecture deep dive.

**Independent Test**: Verify that the generated flowcharts and architecture descriptions render correctly in VitePress.

### Implementation for User Story 1

- [X] T004 [P] [US1] Write `docs/advanced/architecture.md` including code-level deep dive and Mermaid diagrams
- [X] T005 [P] [US1] Write `docs/zh/advanced/architecture.md` including code-level deep dive and Mermaid diagrams
- [X] T006 [P] [US1] Write `docs/advanced/cache-behavior.md` deep dive covering remote version caching, parallel downloads, and auto-pruning
- [X] T007 [P] [US1] Write `docs/zh/advanced/cache-behavior.md` deep dive covering remote version caching, parallel downloads, and auto-pruning

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Exhaustive CLI Index (Priority: P1)

**Goal**: As a user, I want to see an exhaustive list of all CLI commands.

**Independent Test**: Ensure every single command is detailed and listed on the overview page.

### Implementation for User Story 2

- [X] T008 [P] [US2] Rewrite `docs/cli/overview.md` with an exhaustive, collapsible command list
- [X] T009 [P] [US2] Rewrite `docs/zh/cli/overview.md` with an exhaustive, collapsible command list

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - Environments Management (Priority: P2)

**Goal**: As a user, I want to deeply understand the environment variable resolution.

**Independent Test**: Ensure the `.unirtm.toml` dynamic environment handling and zero shell-pollution is comprehensively explained.

### Implementation for User Story 3

- [X] T010 [P] [US3] Rewrite `docs/environments/overview.md` with deep details
- [X] T011 [P] [US3] Rewrite `docs/zh/environments/overview.md` with deep details

**Checkpoint**: All user stories should now be independently functional

---

## Phase 6: User Story 4 - Security Integrations (Priority: P3)

**Goal**: As a potential enterprise user, I want to see exactly how security tools are integrated.

**Independent Test**: Ensure Trivy, Syft, and Gitleaks are explicitly listed as external integrations, not exaggerated features.

### Implementation for User Story 4

- [X] T012 [P] [US4] Update `docs/guide/getting-started.md` to accurately reflect security tools as integrations
- [X] T013 [P] [US4] Update `docs/zh/guide/getting-started.md` to accurately reflect security tools as integrations

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [X] T014 Run validation by building documentation via `npm run docs:build` in `docs/` to verify links and formatting

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel
- **Polish (Final Phase)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies
- **User Story 2 (P1)**: Can start after Foundational (Phase 2) - No dependencies
- **User Story 3 (P2)**: Can start after Foundational (Phase 2) - No dependencies
- **User Story 4 (P3)**: Can start after Foundational (Phase 2) - No dependencies

### Parallel Opportunities

- All User Story implementations (T004 - T013) are marked [P] and can be run completely in parallel, as they edit distinct markdown files.

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Verify the architecture and caching docs locally.
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational
2. Add US1 → Test independently
3. Add US2 → Test independently
4. Add US3 → Test independently
5. Add US4 → Test independently
6. Run full build validation (Phase 7)
