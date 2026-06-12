# Data Model: Native Lua Provider

## 1. LuaProvider (Struct)

Implements the `Provider` interface for Lua.

**Fields**:
None (stateless struct, heavily relies on contextual paths like `installPath`).

**Key Methods**:

- `Name() string`: Returns `"lua"`
- `Install(...) error`: Handles HTTP GET to the resolved binary URL, archive extraction, and invoking the LuaRocks bootstrap function.
- `resolveURL(version, os, arch string) string`: Internal helper to map version/platform to SourceForge or GitHub release URL.
- `bootstrapLuaRocks(ctx, installPath, luaVersion string) error`: Internal helper that downloads `luarocks` tarball, extracts it, and executes `./configure` & `make install` or `install.bat`.
- `ListExecutables(...)`: Scans the installation directory for `lua` and `luac` (and `.exe` for Windows), as well as `luarocks` and `luarocks-admin`.

## 2. Platform Map

A conceptual internal map to resolve archive suffixes based on exact versions or heuristic defaults:

- `linux/amd64` -> `_Linux54_64_bin.tar.gz`
- `darwin/amd64` -> `_MacOS1015_64_bin.tar.gz`
- `windows/amd64` -> `_Win64_bin.zip`
- `darwin/arm64` -> GitHub Release URL `.../macos-arm64/...`
