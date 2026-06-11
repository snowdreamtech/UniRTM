# Feature Specification: AI IDE Integration Architecture

**Feature Branch**: `020-ai-ide-integration`

**Created**: 2026-06-11

**Status**: Draft

**Input**: User description: "总结上面的讨论和方案，总结出计划"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Unified Execution via Workflows (Priority: P1)

As a developer using any AI IDE (Cursor, Windsurf, Roo, etc.), I want the AI to understand and correctly execute Spec Kit workflows via a single unified directory (`.agent/workflows/`), so that I get a consistent experience across all platforms.

**Why this priority**: Without this, the AI will fail to recognize or execute the Spec Kit workflows effectively depending on the IDE used.

**Independent Test**: Can be tested by invoking a slash command (e.g., `/speckit.plan`) in two different AI IDEs (e.g., Cursor and Windsurf) and verifying they both successfully locate the workflow definitions and invoke the underlying `.specify/scripts/` logic.

**Acceptance Scenarios**:

1. **Given** a multi-IDE project workspace with `.agent/workflows` populated, **When** the developer instructs the AI to "plan this feature using Spec Kit", **Then** the AI successfully reads the SOP from the workflow file and executes the underlying scripts in `.specify/`.

---

### User Story 2 - IDE Routing and Context Provisioning (Priority: P1)

As a developer switching IDEs, I want my IDE-specific config files (`.cursorrules`, `.windsurfrules`, `.clinerules`) to act purely as pointers to `.agent/rules/`, so that I don't have to duplicate instructions.

**Why this priority**: Eliminates context fragmentation and ensures the Single Source of Truth (SSOT) is always respected.

**Independent Test**: Can be tested by deleting all logic in `.cursorrules` except a pointer, asking the AI a question about a core project rule, and verifying the AI correctly follows the pointer to read `.agent/rules/00-index.md`.

**Acceptance Scenarios**:

1. **Given** an empty `.cursorrules` containing only a redirect to `.agent/rules/`, **When** the AI initializes, **Then** it autonomously navigates to `.agent/rules/` to fetch the core system instructions.

---

### User Story 3 - Cleanup of Redundant Directories (Priority: P2)

As a repository maintainer, I want to eliminate redundant command directories (`.agents/commands/`), moving all custom workflows exclusively into `.agent/workflows/`, so that the repository is clean and the AI's context window isn't polluted by ambiguous paths.

**Why this priority**: Reduces confusion and potential hallucinations caused by similarly named folders (`.agent` vs `.agents`).

**Independent Test**: Verify that `.agents/` no longer exists, and all functionality operates perfectly from `.agent/workflows/`.

**Acceptance Scenarios**:

1. **Given** the repository has both `.agent` and `.agents` directories, **When** the migration is completed, **Then** `.agents` is completely removed and all commands reside in `.agent/workflows`.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST consolidate all AI-facing workflow instructions (Slash Commands) into the `.agent/workflows/` directory.
- **FR-002**: System MUST retain `.specify/` strictly for execution scripts, configuration (`init-options.json`, `extensions.yml`), templates, and memories, preventing AI from being confused by excessive unguided context.
- **FR-003**: System MUST update all IDE-specific root files (e.g., `.cursorrules`, `.windsurfrules`, `.cline/mcp.json`) to serve solely as redirect pointers to `.agent/rules/00-index.md` and `.agent/workflows/`.
- **FR-004**: System MUST remove the `.agents` folder completely to prevent naming collisions and AI hallucination.
- **FR-005**: All workflow definitions MUST be written as intelligent Standard Operating Procedures (SOPs) that guide the AI to invoke the correct `.specify/scripts/` logic.

### Key Entities 

- **Unified Routing Layer**: The collection of `.cursorrules`, `AGENTS.md`, and other IDE-specific files acting as pointers.
- **AI Interface Layer**: The `.agent/` directory, exposing `rules/` and `workflows/`.
- **Core Engine**: The `.specify/` directory, containing the actual bash execution scripts and markdown templates.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The directory `.agents/` is fully removed from the codebase.
- **SC-002**: 100% of IDE-specific rules files (e.g., `.cursorrules`, `.windsurfrules`, `.clinerules`) contain no more than 10 lines of text and solely redirect to the `.agent/` directory.
- **SC-003**: The AI successfully executes at least 3 distinct Spec Kit workflows (`speckit.plan`, `speckit.specify`, `speckit.tasks`) guided exclusively by `.agent/workflows/` without needing to interpret the bash scripts directly.

## Assumptions

- We assume the user has the ability to register the `.agent/workflows` directory as custom commands in their specific AI IDE (e.g., via Roo Code's custom modes or standard context).
- We assume standard markdown parsing and bash execution capabilities are available in the AI IDE.
