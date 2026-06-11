# Phase 0: Outline & Research

## Findings

There are no major "NEEDS CLARIFICATION" points since the strategy was fully established in the previous discussion.

### IDE Configuration Loading

- **Decision**: Retain IDE-specific configuration files (`.cursorrules`, `.windsurfrules`, `.clinerules`) but empty their logic.
- **Rationale**: IDEs hardcode their entry points and will not auto-discover `.agent/rules/00-index.md` without a pointer. Removing them would break the AI's ability to find the context.
- **Alternatives considered**: Instructing users to manually type instructions on every prompt. This was rejected because it introduces significant friction.

### Command Directory Unification

- **Decision**: Migrate all workflows from `.agents/commands` to `.agent/workflows`.
- **Rationale**: Having `.agent` and `.agents` coexisting causes ambiguity and potential hallucination for the LLM. `.agent/workflows` is the established standard for defining AI slash commands and SOPs.
- **Alternatives considered**: Keeping both or putting everything in `.specify/`. Rejected because `.specify/` contains complex bash scripts that consume too much context window.

### Proxy Workflow Pattern

- **Decision**: Use `.agent/workflows/` files as Markdown SOPs (Standard Operating Procedures) that instruct the AI to execute scripts in `.specify/scripts/`.
- **Rationale**: Prevents the AI from needing to analyze complex bash code. It just reads the English instructions and runs the shell command safely.
