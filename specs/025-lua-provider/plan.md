# Implementation Plan: Native Lua & LuaRocks Provider

**Branch**: `025-lua-provider` | **Date**: 2026-06-12 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/025-lua-provider/spec.md`

## Summary

Implement a native Go provider (`internal/provider/lua.go`) in UniRTM that downloads pre-compiled Lua binaries from LuaBinaries (or compatible trusted GitHub mirror for missing architectures like arm64) and automatically bootstraps LuaRocks from source to provide a complete, clean, ASDF-free Lua environment.

## Technical Context

**Language/Version**: Go 1.21+

**Primary Dependencies**: `net/http` (downloads), `archive/zip` & `archive/tar` (extraction), `os/exec` (LuaRocks bootstrapping).

**Storage**: Local file system (UniRTM tool installation directory).

**Testing**: Standard `go test` for URL resolution and installation logic.

**Target Platform**: Linux (amd64, arm64), macOS (amd64, arm64), Windows (amd64).

**Project Type**: CLI tool provider within UniRTM.

**Constraints**: Must strictly bypass ASDF and install directly into the isolated UniRTM path.

## Constitution Check

*GATE: Passed*

No violations detected. Directly integrating Lua natively aligns with the principle of reducing third-party moving parts and providing secure, predictable environments.

## Project Structure

### Documentation (this feature)

```text
specs/025-lua-provider/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
└── tasks.md             # Phase 2 output (created in task generation)
```

### Source Code

```text
internal/
└── provider/
    ├── lua.go        # Main provider implementation for Lua & LuaRocks bootstrap
    ├── lua_test.go   # URL resolution and mock install tests
    └── registry.go   # Modify to register "lua" to NewLuaProvider()
```

**Structure Decision**: Added new provider file `lua.go` strictly inside the `internal/provider` package.
