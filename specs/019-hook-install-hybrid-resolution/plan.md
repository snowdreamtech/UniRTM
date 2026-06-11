# Implementation Plan: Hook Install Hybrid Resolution

**Branch**: `019-hook-install-hybrid-resolution` | **Date**: 2026-06-11 | **Spec**: [spec.md](file:///Users/snowdream/Workspace/snowdreamtech/UniRTM/specs/019-hook-install-hybrid-resolution/spec.md)

**Input**: Feature specification from `/specs/019-hook-install-hybrid-resolution/spec.md`

## Summary

Refactor `unirtm hook install` to act as a non-destructive environment injector. Instead of overwriting hooks or triggering duplicate `hook run` execution, the install command will surgically inject a block of bash code that finds and loads the `unirtm` environment dynamically via `git rev-parse --show-toplevel`. This allows the hook to execute correctly in headless AI/CI environments without breaking existing custom hooks.

## Technical Context

**Language/Version**: Go 1.21+

**Primary Dependencies**: `os`, `path/filepath`, `strings` (Standard Library)

**Storage**: Local Filesystem (`.git/hooks/`)

**Testing**: Go `testing` package

**Target Platform**: Unix-like environments (Linux, macOS) and Git Bash (Windows)

**Project Type**: CLI Tool

**Constraints**: Must accurately parse bash scripts to find Shebangs and replace blocks statelessly.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*
No violations. The feature operates entirely within the CLI boundaries and adheres to standard Git conventions without external side effects.

## Project Structure

### Documentation (this feature)

```text
specs/019-hook-install-hybrid-resolution/
├── plan.md              # This file
├── research.md          # Implementation analysis
├── data-model.md        # Regex and Data structures
├── quickstart.md        # Testing guide
└── contracts/           # No new contracts needed
```

### Source Code

```text
internal/
└── hook/
    ├── install.go       # The logic for writing/updating hook scripts
    └── install_test.go  # Unit tests for injection logic
```

**Structure Decision**: The logic will reside entirely in `internal/hook/install.go`. We will separate the string manipulation logic (block detection/injection) from the file IO logic for easier unit testing.
