# Implementation Plan: AI IDE Integration Architecture

**Branch**: `020-ai-ide-integration` | **Date**: 2026-06-11 | **Spec**: [spec.md](file:///Users/snowdream/Workspace/snowdreamtech/UniRTM/specs/020-ai-ide-integration/spec.md)

**Input**: Feature specification from `specs/020-ai-ide-integration/spec.md`

## Summary

Consolidate all Spec Kit AI IDE instructions into a unified `.agent/workflows/` directory and configure IDE-specific rule files (`.cursorrules`, `.windsurfrules`, `.clinerules`, etc.) to act as pointers to `.agent/rules/`. The redundant `.agents/commands` directory will be deleted.

## Technical Context

**Language/Version**: Markdown, Bash

**Primary Dependencies**: Various AI IDEs (Cursor, Windsurf, Roo/Cline, Gemini)

**Storage**: Local files (git)

**Testing**: Manual slash command triggers

**Target Platform**: Any AI IDE that reads workspace contexts

**Project Type**: Developer Tooling / Agentic Architecture

**Performance Goals**: N/A

**Constraints**: Must strictly preserve `.specify/` as the underlying execution engine.

**Scale/Scope**: Impacts repository root rule files and AI workflows.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Principle Check**: Aligns with keeping systems clean, reducing complexity (deleting `.agents`), and maintaining clear boundaries (IDE Interface vs Execution Engine). No violations detected.

## Project Structure

### Documentation (this feature)

```text
specs/020-ai-ide-integration/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
.agent/
├── rules/
│   └── 00-index.md
└── workflows/
    ├── speckit.plan.md
    ├── speckit.specify.md
    └── ... (migrated from .agents/commands)

# IDE Specific (Redirect Pointers)
.cursorrules
.windsurfrules
.clinerules

.specify/
└── scripts/
    └── ... (unchanged)
```

**Structure Decision**: A unified proxy structure using `.agent/workflows/` as the primary AI interface, while treating IDE-specific config files as thin pointers. The redundant `.agents/` folder is completely removed.

### Phase 2: Universal Compiler (Auto-Sync)
To solve the maintenance burden of proxy files and support 50+ AI IDEs natively, we introduce a `compile-ide-adapters.sh` script.

#### [NEW] .specify/scripts/bash/compile-ide-adapters.sh
- **Purpose**: Scans `.specify/commands/*.md` and generates corresponding adapter files.
- **Logic**:
  1. For every `cmd.md` in `.specify/commands/`, creates an `.agent/workflows/cmd.md` proxy file.
  2. Detects `.cursor/rules/`: Creates `.cursor/rules/speckit_cmd.mdc` using Cursor's MDC format.
  3. Detects `.cline/`, `.roo/`, `.windsurf/`, `.github/` and auto-compiles instructions if necessary (Phase 2 targets `.agent/workflows` as primary, `.cursor/rules` as secondary native target).

#### [MODIFY] .specify/scripts/bash/check-prerequisites.sh
- **Change**: Invoke `compile-ide-adapters.sh` silently to ensure workflows are always up-to-date before any Spec Kit command executes.
