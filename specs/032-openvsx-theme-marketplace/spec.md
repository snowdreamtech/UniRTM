# Feature Specification: 032-openvsx-theme-marketplace

## Overview
Connect UniRTM to [Open VSX Registry](https://open-vsx.org) to provide an integrated theme and icon marketplace. Users can search, browse, view details, install with a single click, delete, and enable/apply VS Code compatible color themes and file icon themes directly inside the UniRTM Workbench, strictly filtering out general extensions.

---

## User Stories

### US1: Search & Filter Only Themes & Icons from Open VSX
- **As a** developer using UniRTM,
- **I want to** search and browse themes and icon themes from Open VSX,
- **So that** I only see appearance-related themes (color and icon) and never see irrelevant heavy extensions (language servers, debuggers, code runners).

**Acceptance Criteria**:
- Open VSX API calls strictly query `category=Themes`.
- Search query `q` allows finding specific themes (e.g. "Dracula", "One Dark", "Material", "Catppuccin").
- Returned items provide: `id`, `name`, `displayName`, `version`, `publisher/namespace`, `description`, `downloadCount`, `rating`, `iconUrl`, `downloadUrl`.

### US2: One-Click Dynamic Online Installation & Deletion
- **As a** developer,
- **I want to** click "Install" on any theme in the marketplace list,
- **So that** UniRTM automatically streams the `.vsix` package in the backend, unpacks it into the standard user theme directory (`~/.local/share/unirtm/themes/{colors,icons}`), and makes it immediately available without manual download or file upload.

**Acceptance Criteria**:
- Backend endpoint `POST /api/themes/marketplace/install` downloads `.vsix` via Go HTTP client (respecting network proxy/mirrors).
- Security: enforces Zip Slip protection and rejects any package lacking `contributes.themes` or `contributes.iconThemes`.
- Already-installed packages show an "已安装 (Installed)" badge with an "卸载 (Uninstall)" button.

### US3: Seamless Enablement & Suite-Level Application
- **As a** developer,
- **I want to** click "Apply" or pick the installed theme from the theme dropdown,
- **So that** both light and dark modes adapt to the selected theme family and icons update in real time.

### US4: Dedicated Marketplace View with Channel-Level Decoupling
- **As a** developer,
- **I want to** browse themes and plugins from a top-level dedicated "Marketplace" view rather than buried inside Settings,
- **So that** I have an expansive, immersive discovery experience, and themes (from Open VSX) and native plugins (from UniRTM ecosystem) remain cleanly separated across dedicated channels.

**Acceptance Criteria**:
- Marketplace top-level page `/marketplace` features two decoupled tabs/channels:
  1. `🎨 Themes & Icons (Open VSX)`: full search, chips, grid, installation and activation.
  2. `🧩 Plugins & Extensions (UniRTM Ecosystem)`: native tool providers, AI agents, and workflows.
- Settings page `/settings` is slimmed down: focuses purely on preferences and configurations, with a quick link to `Marketplace` for discovering new themes.

---

## Non-Functional Requirements
- **100% Static Extraction**: No Node.js runtime, no Extension Host process.
- **Proxy Aware**: All downloads flow through UniRTM's configured network proxy/mirrors if enabled.
- **Offline Resilient**: Builtin themes (Dark, Light, Seti UI) remain 100% functional without internet connectivity.
