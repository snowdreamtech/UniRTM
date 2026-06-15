# Feature Tasks: AI IDE Integration Architecture

**Feature**: AI IDE Integration Architecture
**Branch**: `020-ai-ide-integration`
**Spec**: [spec.md](./spec.md)
**Plan**: [plan.md](./plan.md)

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic directory structure for the unified agent context

- [x] T001 Create `.agent/workflows` directory if it does not exist
- [x] T002 Create `.agent/rules` directory if it does not exist

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T003 Ensure `.specify/` core engine scripts remain untouched and functional

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Unified Execution via Workflows (Priority: P1) 🎯 MVP

**Goal**: The AI should understand and correctly execute Spec Kit workflows via a single unified directory (`.agent/workflows/`)

**Independent Test**: Can be tested by invoking a slash command in the AI IDE after the files are moved.

### Implementation for User Story 1

- [x] T004 [P] [US1] Move `.agents/commands/speckit.plan.md` to `.agent/workflows/speckit.plan.md`
- [x] T005 [P] [US1] Move `.agents/commands/speckit.tasks.md` to `.agent/workflows/speckit.tasks.md`
- [x] T006 [P] [US1] Move `.agents/commands/speckit.specify.md` to `.agent/workflows/speckit.specify.md`
- [x] T007 [P] [US1] Move all remaining files from `.agents/commands/` to `.agent/workflows/`

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - IDE Routing and Context Provisioning (Priority: P1)

**Goal**: Update IDE-specific config files to act purely as pointers to `.agent/rules/` and `.agent/workflows/`

**Independent Test**: Can be tested by opening the project in Cursor/Windsurf/Roo and verifying the AI correctly reads `.agent/rules/00-index.md` on startup.

### Implementation for User Story 2

- [x] T008 [P] [US2] Update `.cursorrules` to redirect to `.agent/rules/00-index.md` and `.agent/workflows/`
- [x] T009 [P] [US2] Update `.windsurfrules` to redirect to `.agent/rules/00-index.md` and `.agent/workflows/`
- [x] T010 [P] [US2] Update `.clinerules` to redirect to `.agent/rules/00-index.md` and `.agent/workflows/`
- [x] T011 [P] [US2] Verify `AGENTS.md` and `GEMINI.md` reflect the new proxy structure

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - Cleanup of Redundant Directories (Priority: P2)

**Goal**: Eliminate redundant command directories (`.agents/`) so that the AI's context window isn't polluted

**Independent Test**: Verify that `.agents/` no longer exists

### Implementation for User Story 3

- [x] T012 [US3] Delete `.agents/` directory completely

**Checkpoint**: All user stories should now be independently functional

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [x] T013 Verify `/speckit.analyze` or other slash commands execute without directory resolution errors
- [x] T014 Run quickstart.md validation locally

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
- **User Story 2 (P1)**: Can start after Foundational (Phase 2) - Independent of US1
- **User Story 3 (P2)**: MUST start AFTER User Story 1 (US1), since US1 moves the files out of the folder before US3 deletes it.

### Parallel Opportunities

- All file moves in US1 (T004-T007) can be done simultaneously using a wild card.
- All IDE config updates in US2 (T008-T011) can be done in parallel.

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Test User Story 1 independently
5. Deploy/demo if ready

## Phase 7: Universal Compiler Implementation (Priority: P1)

- [x] T015 Create `.specify/scripts/bash/compile-ide-adapters.sh` to compile workflows into `.agent/workflows/`
- [x] T016 Add `.cursor/rules/*.mdc` generator logic inside the compiler script
- [x] T017 Hook `compile-ide-adapters.sh` into `check-prerequisites.sh` for auto-sync
- [x] T018 Test compiler by generating proxies for existing `.specify/commands/*.md`

## Phase 8: Universal Compiler V2 (Priority: P1)

- [x] T019 Implement orphan cleanup logic for `.agent/workflows/`
- [x] T019 Implement orphan cleanup logic for `.cursor/rules/speckit_*.mdc`
- [x] T020 Refactor file generation to be idempotent using `cmp -s`
- [x] T021 Add universal adapter injector for `.clinerules`, `.windsurfrules`, `.roo-rules`, `.traerules`, and `.github/copilot-instructions.md`
- [x] T022 Test V2 compiler
