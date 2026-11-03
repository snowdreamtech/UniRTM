# Implementation Plan: df command

**Branch**: `007-df-cmd` | **Date**: 2026-06-04 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/007-df-cmd/spec.md`

## Summary

Add a `df` command to `unirtm` to display the capacity and size statistics of various folders within the `unirtm` data directory. It will support a `-h` (`--human-readable`) flag to print sizes in powers of 1024, and will format the output nicely using the `pterm` library.

## Technical Context

**Language/Version**: Go (version defined in project)
**Primary Dependencies**: `github.com/pterm/pterm`, `github.com/spf13/cobra`
**Storage**: N/A (reads from filesystem)
**Testing**: Go standard testing (`testing` package)
**Target Platform**: Linux, macOS, Windows
**Project Type**: single (CLI application)
**Performance Goals**: Fast filesystem traversal (< 1s for typical data directories)
**Constraints**: Must be cross-platform (handling Windows paths/permissions vs POSIX)
**Scale/Scope**: Local CLI command execution

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Design**: Kept simple, integrates seamlessly into the existing `unirtm` command structure (using `cobra`).
- **Dependencies**: Reusing `pterm` (which is likely already used or a standard choice for CLI UI) and standard libraries.
- **Testing**: Requires unit testing for the directory size calculation logic.

## Project Structure

### Documentation (this feature)

```text
specs/007-df-cmd/
├── plan.md              # This file
├── spec.md              # Feature specification
└── tasks.md             # Tasks definition
```

### Source Code (repository root)

```text
cmd/
└── df.go                # Cobra command definition for 'df'

internal/
└── utils/
    └── disk.go          # Logic for calculating directory sizes

tests/
└── df_test.go           # Unit tests for df calculation logic
```

**Structure Decision**: Added a new command file `df.go` to the `cmd/` directory. Added core size calculation logic to `internal/utils/` to keep the command file clean and allow for unit testing.
