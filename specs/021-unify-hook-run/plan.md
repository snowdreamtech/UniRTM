# Implementation Plan: Unify `unirtm hook run` Arguments

**Branch**: `021-unify-hook-run` | **Date**: 2026-06-11 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/021-unify-hook-run/spec.md`

## Proposed Changes

### `cmd/62.hook.go`

- Add a new persistent flag: `hookCmd.Flags().StringVar(&hookStage, "stage", "", "Git lifecycle stage (e.g. pre-commit)")`.
- Update the Cobra validation hook: Use `cobra.MinimumNArgs(1)` instead of `RangeArgs(1, 2)`.
- Configure `hookRunCmd` to skip parsing flags after the positional arguments using `hookRunCmd.Flags().SetInterspersed(false)` (or just let trailing arguments flow naturally) so we can easily collect `args...`.
- In `RunE`, extract `hookname` from `args[0]`.
- Extract trailing arguments `trailingArgs = args[1:]`.
- Pass `hookname`, `hookStage`, and `trailingArgs` to `runner.Run(...)`.

### `internal/hook/runner.go`

- Update the `HookRunner` interface `Run` method signature:
  `Run(hookName string, stage string, args []string) error`

### Engine Implementations (`internal/hook/*.go`)

- Update `precommit.go`, `lefthook.go`, `husky.go`, `native.go`, `shell.go` to match the new `Run` signature.
- **Native/Shell**: If `hookName` != "all", execute `.git/hooks/<hookName> <args...>`. If "all", execute `.git/hooks/<stage> <args...>`.
- **Husky**: Husky operates fully monolithic. `hookName` doesn't apply cleanly unless it's the exact script name. If `stage` is provided, run `sh .husky/<stage> <args...>`.
- **Lefthook**: Lefthook maps `stage` to its command `run` arguments. If `hookName` != "all", run `lefthook run <stage> --commands <hookName> <args...>`. If `hookName` == "all", run `lefthook run <stage> <args...>`.
- **Pre-commit**: If `hookName` != "all", run `pre-commit run <hookName> --hook-stage <stage> -- <args...>`. If "all", run `pre-commit run --hook-stage <stage> -- <args...>`.

## Summary

The goal is to strictly enforce `unirtm hook run [hookname] [--stage stage] [args...]` and implement an intelligent abstraction layer that maps these arguments consistently across heterogeneous engine backends (`pre-commit`, `husky`, `lefthook`).

## Technical Context

**Language/Version**: Go 1.22+

**Primary Dependencies**:

- `github.com/spf13/cobra` for CLI parsing
- Internal packages `internal/hook`

**Storage**: N/A

**Testing**: Standard `testing` package in Go.

**Target Platform**: Cross-platform (Windows, macOS, Linux) CLI environment.

**Project Type**: CLI Tool / Git integration.

**Performance Goals**: Negligible latency during CLI parsing/routing.

**Constraints**: Must fallback gracefully for engines like Husky that do not natively support running single linting rules (Hooks).

**Scale/Scope**: Impacts `cmd/62.hook.go` and multiple `.go` engine implementations inside `internal/hook/`.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] Does not introduce new top-level project paradigms.
- [x] Keeps the core `unirtm` binary lightweight without external heavy dependencies.

## Project Structure

### Documentation (this feature)

```text
specs/021-unify-hook-run/
├── plan.md              # This file
├── research.md          # Strategy and mapping validation
├── data-model.md        # Command line arguments structure
├── quickstart.md        # Validation scenarios
├── contracts/           # CLI usage contract
└── tasks.md             # To be created by /speckit.tasks
```

### Source Code (repository root)

```text
cmd/
└── 62.hook.go

internal/hook/
├── runner.go
├── router.go
├── husky.go
├── lefthook.go
├── native.go
├── precommit.go
└── shell.go
```

**Structure Decision**: The project is a single CLI application. The existing `cmd/` and `internal/hook/` directory structure is strictly maintained.

## Complexity Tracking

No violations.
