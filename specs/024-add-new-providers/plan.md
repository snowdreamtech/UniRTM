# Implementation Plan: Add New Backend Providers

**Branch**: `024-add-new-providers` | **Date**: 2026-06-12 | **Spec**: [spec.md](file:///Users/snowdream/Workspace/snowdreamtech/UniRTM/specs/024-add-new-providers/spec.md)

**Input**: Feature specification from `specs/024-add-new-providers/spec.md`

## Summary

Expand UniRTM ecosystem support by implementing four new backend providers: Composer (PHP), LuaRocks (Lua), Pub (Dart), and Cabal (Haskell). The approach leverages the existing `Provider` interface, manipulating environment variables and installation flags to isolate tool installations securely within the UniRTM cache.

## Technical Context

**Language/Version**: Go 1.26+

**Primary Dependencies**: None (Standard Go library + underlying system package managers)

**Storage**: Local Filesystem (UniRTM isolated directory)

**Testing**: Go testing (`go test`)

**Target Platform**: Mac, Linux, Windows

**Project Type**: CLI Tool

**Constraints**: Must isolate completely from global system state. Must gracefully degrade if native package manager is not found.

**Scale/Scope**: 4 new providers

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

No Constitution violations detected.

## Project Structure

### Documentation (this feature)

```text
specs/024-add-new-providers/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
└── tasks.md
```

### Source Code (repository root)

```text
internal/
└── provider/
    ├── composer.go
    ├── composer_test.go
    ├── luarocks.go
    ├── luarocks_test.go
    ├── pub.go
    ├── pub_test.go
    ├── cabal.go
    ├── cabal_test.go
    └── registry.go
```

**Structure Decision**: Add independent `.go` and `_test.go` files for each provider under `internal/provider/` and register them in `registry.go`.
