# Feature Specification: df command

**Feature Branch**: `007-df-cmd`
**Created**: 2026-06-04
**Status**: Draft

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Show human readable disk usage (Priority: P1)

## Requirements

1. Add a `df` command to the CLI (`cmd/df.go`).
2. Command should calculate the size of directories within the `unirtm` data directory.
3. The calculation should summarize usage (similar to standard `df -h`).
4. Output should use `pterm` (tables, progress or just styled text) for friendly presentation.
5. The sizes should be printed in powers of 1024 (e.g., 1023M, 1.2G).
