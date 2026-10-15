# Feature Specification: VS Code Compatible Theme and Icon Management System

**Feature Branch**: `031-vscode-theme-manager`

**Created**: 2026-10-15

**Status**: Draft

**Input**: User description: "1. 我们需要完全兼容vscode的主题和图标集合，包括应用，安装，删除。 2. 我们应该不需要vscode那一套扩展吧？ 3. 我们需要在本地用户目录定义主题和图标的安装目录"

## Summary

Implement a full lifecycle management system for VS Code theme packages (both color themes and icon themes) in UniRTM:
1. **100% Declarative Compatibility**: Directly supports VS Code `.vsix` packages and theme definitions via `contributes.themes` and `contributes.iconThemes`.
2. **Zero Extension Host Overhead**: Pure static asset parsing and extraction; operates without Node.js or the heavyweight VS Code extension runtime.
3. **Dedicated User Directory**: Standardized local installation storage under `~/.local/share/unirtm/themes/` (or OS equivalent), partitioned into `icons/` and `colors/`.
4. **Full Lifecycle API & UI**: Install, list, apply/hot-switch, and delete themes via Go backend REST APIs and Vue 3 frontend Settings workbench.

---

## User Scenarios & Testing

### User Story 1 - Local User Directory & Theme Discovery (Priority: P1)

As a developer, I want UniRTM to automatically scan my local user directory (`~/.local/share/unirtm/themes/`) on startup, so that any theme or icon packages placed there are immediately discovered and available in the UI.

**Why this priority**: Foundational infrastructure for all theme operations.

**Independent Test**:
- Place a valid VS Code theme folder inside `~/.local/share/unirtm/themes/colors/` or `icons/`.
- Query `GET /api/themes`: The API returns the theme with its id, name, type, and source path.

**Acceptance Scenarios**:
1. **Given** the local directory `~/.local/share/unirtm/themes/`, **When** the backend initializes, **Then** it creates `icons/` and `colors/` subdirectories if they do not exist.
2. **Given** installed theme folders with `package.json`, **When** calling `GET /api/themes`, **Then** the backend parses `contributes.themes` and `contributes.iconThemes` and returns structured metadata.

---

### User Story 2 - Install & Delete Themes via .vsix or ZIP (Priority: P1)

As a developer, I want to install a downloaded VS Code theme (`.vsix` or `.zip`) through the UI or API and be able to uninstall it when no longer needed, without running any Node.js processes.

**Why this priority**: Core self-service theme acquisition without manual file unzipping.

**Independent Test**:
- Upload a `.vsix` theme package to `POST /api/themes/install`.
- Verify the package is extracted to `~/.local/share/unirtm/themes/<type>/<id>`.
- Call `DELETE /api/themes/<type>/<id>`: The directory is cleanly removed.

**Acceptance Scenarios**:
1. **Given** a standard `.vsix` package, **When** uploaded to `/api/themes/install`, **Then** the Go backend extracts the `extension/` payload into the proper themes directory and registers the theme.
2. **Given** an invalid zip or package lacking theme contributions, **When** uploaded, **Then** the backend rejects it with an informative HTTP 400 error.
3. **Given** an installed theme, **When** deleted via `DELETE /api/themes/<type>/<id>`, **Then** the filesystem directory is deleted and the list is updated.

---

### User Story 3 - Hot-Switch & Apply Themes in Settings (Priority: P1)

As a developer, I want to browse installed color and icon themes in the Settings view and switch between them instantly, with my choice saved to `unirtm.toml`.

**Why this priority**: Direct user-facing value and seamless customization experience.

**Independent Test**:
- Open Settings -> Appearance.
- Select a different color theme or icon theme from the dropdown: UI colors and icons update immediately without page reload.
- Refresh the page: The chosen theme remains active.

**Acceptance Scenarios**:
1. **Given** active themes in Settings, **When** the user selects a color theme, **Then** the Color Token Bridge maps VS Code tokens to CSS variables and updates the document root.
2. **Given** active themes in Settings, **When** the user selects an icon theme, **Then** `IconThemeService.switchTheme()` activates the manifest and rerenders all icons.
3. **Given** a theme selection change, **When** applied, **Then** `settings.theme` or `settings.icon_theme` is persisted to `~/.config/unirtm/unirtm.toml`.

---

## Technical Constraints & Requirements

1. **Storage Paths**:
   - Linux: `~/.local/share/unirtm/themes/`
   - macOS: `~/Library/Application Support/unirtm/themes/`
   - Windows: `%LOCALAPPDATA%\unirtm\themes\`
2. **Pure Static Extraction**: Go standard library `archive/zip` for `.vsix` unpacking.
3. **Token Mapping**: Direct mapping of standard VS Code color tokens:
   - `activityBar.background` -> `--bg-color` / ActivityBar bg
   - `sideBar.background` -> `--card-bg` / Sidebar bg
   - `editor.background` -> Canvas bg
   - `statusBar.background` -> StatusBar bg
   - `activityBarBadge.background` -> Accent color
4. **Resilience**: If an active theme is deleted, system immediately falls back to built-in default (Seti for icons, Dark/Light for colors).
