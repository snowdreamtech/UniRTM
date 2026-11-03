# Research Findings: Native Lua & LuaRocks Provider

## 1. LuaBinaries Source and Architectures

**Decision**: Use `sourceforge.net/projects/luabinaries` for x64 architectures, and fallback/mirror from `dyne/luabinaries` or GitHub Releases for Apple Silicon (arm64) and Linux (arm64) if native builds are requested.
**Rationale**:
The original SourceForge LuaBinaries primarily caters to x86_64 and older x86 platforms. Modern Mac users (M1/M2/M3) require `arm64` native builds, which SourceForge lacks. The community-maintained `dyne/luabinaries` on GitHub builds the exact same static engine but natively compiles for modern architectures (`macos-arm64`, `linux-arm64`). By routing arm64 architectures to the GitHub mirror and x64 to SourceForge (or mapping them appropriately), we achieve 100% platform coverage without compiling from source.
**Alternatives considered**:

- Using Homebrew internally (violates UniRTM's isolated hermetic environment principle).
- Compiling from source (requires a C compiler installed on the host, which breaks the "ready-to-go" pre-compiled nature of providers).

## 2. LuaRocks Bootstrap Process

**Decision**: Download the source tarball of LuaRocks (e.g., `luarocks-3.11.1.tar.gz`), extract it into the `installPath`, and run `./configure --prefix="..." --with-lua="..."` followed by `make` and `make install`. Wait, `make` is required. If `make` is missing (Windows), use the `install.bat` provided by LuaRocks.
**Rationale**:
LuaRocks needs to be "bound" to the specific Lua interpreter we just downloaded. Bootstrapping from source allows LuaRocks to correctly configure its internal paths and Lua version headers.
*Wait! Is there a pure-Lua way to bootstrap?* LuaRocks actually ships with `install.bat` (Windows) and `./configure && make` (Unix). While `make` requires a host tool, it is generally available on Linux/macOS. For Windows, `install.bat` works natively. We will rely on these standard bootstrap tools.
**Alternatives considered**:

- Pre-compiling LuaRocks ourselves and hosting it (requires massive maintenance overhead).
- Calling ASDF (violates user requirement).
