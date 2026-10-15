# Implementation Plan: VS Code Theme Package Management System

**Branch**: `031-vscode-theme-manager` | **Date**: 2026-10-15 | **Spec**: [specs/031-vscode-theme-manager/spec.md](spec.md)

**Input**: Feature specification from `/specs/031-vscode-theme-manager/spec.md`

## Summary

Implement a lightweight, standalone VS Code theme package manager in UniRTM:
1. **User Storage Directory**: Standardized location under `~/.local/share/unirtm/themes/{icons,colors}` (cross-platform with macOS and Windows equivalents).
2. **Backend Engine (Go)**:
   - Scans installed theme packages (parsing `package.json` for `contributes.themes` and `contributes.iconThemes`).
   - Unpacks `.vsix` (zip) archives safely without any external Node.js extension host runtime.
   - Serves REST APIs (`/api/themes`, `/api/themes/install`, `/api/themes/delete`, `/api/themes/apply`).
   - Serves static theme assets via `/api/themes/assets/*`.
3. **Frontend Workbench (Vue 3)**:
   - Extends `useTheme` and `useIconTheme` to support dynamic loading from `/api/themes`.
   - Adds Color Token Bridge (maps standard VS Code `colors` to CSS variables).
   - Adds Theme Management UI in `SettingsView.vue` with install/uninstall and theme dropdown pickers.

---

## Technical Context

**Language/Version**: Go 1.24, TypeScript 5.7, Vue 3.5

**Primary Dependencies**: Go standard library (`archive/zip`, `net/http`), Vue 3, Vue Router 4

**Storage**: Local user directory `~/.local/share/unirtm/themes/` and user config `~/.config/unirtm/unirtm.toml`

**Target Platform**: Linux, macOS, Windows

**Performance Goals**:
- Theme scanning: $<10$ ms for 50 installed themes
- VSIX extraction: $<200$ ms for typical 1MB theme package
- Theme hot-switch: instantaneous ($<16$ ms, within 1 frame)

---

## Constitution Check

- [x] **Principle I: Self-Contained**: Go package in `internal/theme/` handles all file discovery, unpacking, and parsing.
- [x] **Principle II: Zero Extension Overhead**: Pure static declarative asset parsing; NO Node.js or extension host runtime.
- [x] **Principle III: Verification**: Zero build errors, full REST API testing.

---

## Project Structure

```text
internal/
└── theme/
    ├── manager.go        # Theme discovery, package.json parsing, .vsix zip unpacker
    └── manager_test.go   # Unit tests for theme manager

cmd/
└── 64.ui.go             # Register /api/themes endpoints and asset router

frontend/
├── src/
│   ├── composables/
│   │   ├── useThemeManager.ts   # Composable for managing themes list, install, delete
│   │   └── useColorTheme.ts     # VS Code Color Token to CSS variables bridge
│   └── views/
│       └── SettingsView.vue     # Theme management UI & pickers
```
