# Phase 0 Research: VS Code Theme Package Management System

**Branch**: `031-vscode-theme-manager` | **Date**: 2026-10-15

## Research Questions & Decisions

### 1. Structure of VS Code Theme Packages (.vsix)

**Question**: How does a VS Code `.vsix` file package themes, and how can Go extract it safely without Node.js?

**Findings**:
1. A `.vsix` file is an Open Packaging Conventions (OPC) ZIP archive.
2. Inside the archive, files are placed under the `extension/` directory:
   - `extension/package.json` contains metadata and contributions:
     ```json
     {
       "name": "theme-one-dark-pro",
       "displayName": "One Dark Pro",
       "contributes": {
         "themes": [
           {
             "label": "One Dark Pro",
             "uiTheme": "vs-dark",
             "path": "./themes/OneDark-Pro.json"
           }
         ],
         "iconThemes": [
           {
             "id": "vscode-icons",
             "label": "VSCode Icons",
             "path": "./icons/theme.json"
           }
         ]
       }
     }
     ```
3. Safe Go implementation:
   - Use standard library `archive/zip` and `io.Copy`.
   - Protect against Zip Slip vulnerabilities (ensure extracted file paths do not escape target directory with `filepath.Clean` and `strings.HasPrefix`).
   - Extract into user data directory:
     `~/.local/share/unirtm/themes/colors/<package-name>/` or `~/.local/share/unirtm/themes/icons/<package-name>/`.

---

### 2. VS Code Color Token Mapping to UniRTM Workbench

**Question**: How can VS Code color tokens dynamically map to UniRTM CSS variables?

**Findings**:
VS Code workbench themes define standard semantic tokens in `colors`:
- `activityBar.background` -> `--bg-color` (Workbench background)
- `sideBar.background` -> `--card-bg` (Panels and Cards)
- `editor.background` -> `--canvas-bg` (Main workspace canvas)
- `statusBar.background` -> `--statusbar-bg`
- `sideBar.border` / `activityBar.border` -> `--card-border`
- `foreground` / `editor.foreground` -> `--text-main`
- `descriptionForeground` -> `--text-muted`
- `activityBarBadge.background` -> `--accent-cyan`

A simple frontend bridge function iterates over the theme's `colors` object and updates CSS variables in `document.documentElement.style`.

---

### 3. API Contract & Lifecycle

- `GET /api/themes`: Lists all available themes (both built-in and user-installed).
- `POST /api/themes/install`: Multipart upload of `.vsix` or `.zip`.
- `DELETE /api/themes?type=(color|icon)&id=(id)`: Uninstalls user theme.
- `GET /api/themes/assets/*`: Serves unpacked font and SVG assets.
- `POST /api/themes/apply`: Saves user choice to `unirtm.toml`.
