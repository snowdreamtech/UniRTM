# Feature Specification: Native Lua & LuaRocks Provider

**Feature Branch**: `025-lua-provider`

**Created**: 2026-06-12

**Status**: Draft

**Input**: User description: "啃下原生路线：实现 LuaBinaries 下载 + LuaRocks 自动引导安装（体验最好，最符合您远离供应链投毒的期望，但实现有一定复杂度）。"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Install Pure Lua Environment (Priority: P1)

As a developer, I want to install Lua directly from the official LuaBinaries source so that I have a clean, fast, and trustworthy execution environment without relying on third-party version managers like ASDF.

**Why this priority**: Lua execution is the core foundation. Without it, scripts cannot run.

**Independent Test**: Can be fully tested by configuring `lua = "5.4.2"` in `.unirtm.toml` and verifying that `lua -v` executes correctly and reports the correct version.

**Acceptance Scenarios**:

1. **Given** a new project without Lua installed, **When** I run the installation command with Lua configured, **Then** Lua binaries are downloaded directly from LuaBinaries and placed in the toolchain path.
2. **Given** an installed Lua environment, **When** I execute a lua script, **Then** the script executes successfully using the downloaded native binary.

---

### User Story 2 - Automatic LuaRocks Bootstrapping (Priority: P1)

As a developer, I want LuaRocks to be automatically bootstrapped and installed alongside Lua so that I can immediately manage and install third-party Lua packages without additional manual setup.

**Why this priority**: Package management is essential for modern development. Without LuaRocks, users cannot utilize the existing `luarocks` integration in UniRTM.

**Independent Test**: Can be fully tested by configuring `"luarocks:busted" = "2.0.0"` in `.unirtm.toml` and verifying that the package installs successfully via the bootstrapped package manager.

**Acceptance Scenarios**:

1. **Given** a successful Lua engine installation, **When** the post-install process completes, **Then** LuaRocks is automatically downloaded, configured, and available in the environment path.
2. **Given** an environment with the bootstrapped LuaRocks, **When** I request a third-party package installation, **Then** LuaRocks successfully resolves and installs the requested package using the native Lua engine.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST download pre-compiled Lua archives natively from the official LuaBinaries source (SourceForge).
- **FR-002**: System MUST automatically resolve the correct LuaBinaries file suffix based on the host operating system and architecture.
- **FR-003**: System MUST automatically download the LuaRocks source or bootstrap payload during the Lua installation lifecycle.
- **FR-004**: System MUST execute the LuaRocks bootstrap installation using the newly downloaded native Lua engine.
- **FR-005**: System MUST ensure that both `lua` and `luarocks` executables are exposed in the environment path after installation.
- **FR-006**: System MUST NOT rely on ASDF or any external third-party version manager for Lua or LuaRocks installation.

### Key Entities

- **LuaBinaries Archive**: The compressed file (zip/tar.gz) retrieved from SourceForge containing the pre-compiled `lua` and `luac` executables.
- **LuaRocks Source Payload**: The source code or bootstrap script required to compile and configure the LuaRocks package manager locally.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can install a fully functional Lua + LuaRocks environment strictly via the native provider implementation without triggering any ASDF fallbacks.
- **SC-002**: Installation of third-party Lua packages (e.g., via `luarocks:busted`) succeeds using the natively bootstrapped package manager.
- **SC-003**: The download and installation pipeline successfully completes on macOS, Linux, and Windows.

## Assumptions

- Users have stable internet connectivity to access SourceForge and LuaRocks distribution servers.
- The host environment has the necessary fundamental build tools (if any are strictly required by the LuaRocks bootstrap script, though pure Lua environments minimize this).
- Existing LuaRocks configurations or global system installations will not interfere with the isolated toolchain environment.
