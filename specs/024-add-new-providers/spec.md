# Feature Specification: Add New Backend Providers

**Feature Branch**: `024-add-new-providers`

**Created**: 2026-06-12

**Status**: Draft

**Input**: User description: "你设计的 UniRTM 的架构非常优秀，它的 Provider 接口具有极强的抽象能力... 1. PHP 的 Composer 2. Lua 的 LuaRocks 3. Dart 的 Pub 4. Haskell 的 Cabal"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Install PHP Tools via Composer (Priority: P1)

Users with PHP projects need to install PHP-based developer tools (e.g., `phpstan`, `phpunit`) securely and reliably without modifying global system state or learning composer's global path rules.

**Why this priority**: PHP is an extremely popular ecosystem and Composer is its primary package manager. Expanding UniRTM to cover it adds massive value to web developers.

**Independent Test**: Can be fully tested by configuring `composer:phpstan` in `.unirtm.toml`, running `unirtm install`, and verifying the `phpstan` command is executable from the workspace.

**Acceptance Scenarios**:

1. **Given** a valid `.unirtm.toml` containing a Composer package, **When** the user runs the install command, **Then** UniRTM successfully delegates to Composer and installs the package in an isolated directory.
2. **Given** a successfully installed Composer package, **When** the user runs the tool's binary shim, **Then** it executes successfully without path resolution errors.

---

### User Story 2 - Install Dart Tools via Pub (Priority: P2)

Flutter and Dart developers need to manage tools (e.g., `fvm`, `melos`) predictably across their team without relying on the system-wide `.pub-cache`.

**Why this priority**: Dart is rapidly growing due to Flutter, and isolated tool management is highly requested.

**Independent Test**: Can be fully tested by configuring `pub:fvm` in `.unirtm.toml` and verifying installation isolation.

**Acceptance Scenarios**:

1. **Given** a request to install a Pub package, **When** UniRTM processes the installation, **Then** it isolates the installation using the appropriate cache isolation variables.

---

### User Story 3 - Install Lua and Haskell Tools (Priority: P3)

Developers working in Lua (e.g., Neovim config hackers, game devs) or Haskell need isolated toolchain binaries (e.g., `luacheck`, `shellcheck`).

**Why this priority**: While more niche than PHP, `shellcheck` (Haskell) and `luacheck` (Lua) are ubiquitous linters that currently require cumbersome system-level installations.

**Independent Test**: Can be fully tested by configuring `luarocks:luacheck` and `cabal:shellcheck`.

**Acceptance Scenarios**:

1. **Given** a LuaRocks or Cabal package configuration, **When** UniRTM installs it, **Then** the generated binary is correctly linked or shimmed for immediate use.

### Edge Cases

- What happens if the host system does not have the necessary language runtime installed (e.g., PHP, Dart, Lua)?
- How does the system handle concurrent installations of tools from the same provider?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST support interpreting packages prefixed with `composer:`, `luarocks:`, `pub:`, and `cabal:`.
- **FR-002**: System MUST install tools in isolated, provider-specific directories within the UniRTM cache, rather than the global user directory.
- **FR-003**: System MUST transparently set necessary environment variables (like `COMPOSER_HOME` or `PUB_CACHE`) when delegating to the underlying package managers.
- **FR-004**: System MUST locate and expose the generated executables for the newly installed tools so they are available in the user's PATH or via shims.
- **FR-005**: System MUST gracefully fail and notify the user if the underlying package manager executable (`composer`, `luarocks`, `dart`, `cabal`) is not found on the host system.

### Key Entities

- **Backend Provider**: A modular component responsible for translating UniRTM commands into ecosystem-specific package manager commands.
- **Package Configuration**: The string definition (e.g., `provider:tool@version`) provided by the user.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can successfully install at least one known tool from each of the 4 new ecosystems (Composer, LuaRocks, Pub, Cabal) using standard UniRTM configuration.
- **SC-002**: Installing these tools must not modify any files in the user's standard global cache directories (e.g., `~/.composer`, `~/.pub-cache`).
- **SC-003**: Tool installation times via UniRTM are within 10% of the time it takes to run the native package manager install command directly.

## Assumptions

- Users already have the necessary language runtimes (PHP, Lua, Dart, Haskell) and package managers installed on their system path. UniRTM is not responsible for bootstrapping the language runtimes themselves in this feature iteration.
- The default installation behavior of these package managers provides standard binary outputs (e.g., `.bat` or `.cmd` on Windows, standard executable on Unix) that do not require complex post-installation patching.
