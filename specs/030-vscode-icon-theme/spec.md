# Feature Specification: VS Code Icon Theme 100% Specification Alignment

**Feature Branch**: `030-vscode-icon-theme`

**Created**: 2026-10-15

**Status**: Draft

**Input**: User description: "如果从商业或者协议上看没什么问题的话，通过speckit来制定计划，在规范上 100% 对齐 VS Code 的标准结构。"

## Summary

Build a fully extensible, declarative icon theme engine for UniRTM that is 100% compatible with the official VS Code File Icon Theme specification (`icon-theme.json`). This architecture decouples icon resolution from UI code, permanently replaces brittle ad-hoc emojis/hardcoded SVGs, incorporates the industry-standard Seti UI icon theme as the built-in default, and supports loading third-party or custom icon themes with zero core code changes.

---

## User Scenarios & Testing

### User Story 1 - Consistent & Accurate Language/Tool Icons Across Workbench (Priority: P1)

As a developer using UniRTM, I want all programming languages, runtimes, and CLI toolchains in the UI (ActivityBar, Sidebar, Canvas cards, and Status indicators) to display authentic, recognizable developer icons matching their official visual identity, so that the workbench feels professional and trustworthy.

**Why this priority**: Directly solves the user's primary pain point (distorted emojis like 🔷 for Go and 🟢 for Node.js, and manual hardcoding).

**Independent Test**:
- Open the Runtimes view: Go displays the official Seti/Go icon with `#00ADD8` color, Node.js displays the official Node.js hexagon with `#5FA04E`, Python displays the official Python logo, Rust displays Ferris/Rust gear.
- Open the Sidebar: All active items render their corresponding font glyphs sharply at 16px.

**Acceptance Scenarios**:
1. **Given** a runtime tool with ID `go` or `golang`, **When** rendered in the UI, **Then** it automatically resolves via `languageIds["go"]` or `languageIds["golang"]` to the official Go icon definition.
2. **Given** a runtime tool with ID `node` or `nodejs`, **When** rendered in the UI, **Then** it automatically resolves to the official Node icon definition.
3. **Given** an unknown or custom tool (e.g. `my-internal-tool`), **When** rendered in the UI, **Then** it gracefully falls back to the theme's `default` file/tool icon without console errors or layout breakage.

---

### User Story 2 - Zero Hardcoding & Unified `<TechIcon>` Component (Priority: P1)

As a frontend developer working on UniRTM, I want a single drop-in component `<TechIcon :id="toolId" :size="size" />` that handles icon resolution, size scaling, and color theming dynamically via an underlying `IconThemeService`, so that I never have to write a `switch-case` or edit frontend code when new tools are added.

**Why this priority**: Core architectural reusability. Eliminates technical debt and prevents regressions.

**Independent Test**:
- Search all `.vue` files: no `switch(id)` returning emojis, no duplicated icon mappings in `RuntimesView`, `SidebarPanel`, or `ToolsView`.
- Rendering `<TechIcon id="unirtm.lock" />` resolves via `fileNames["unirtm.lock"]` to the lockfile icon.

**Acceptance Scenarios**:
1. **Given** any component requiring an icon, **When** using `<TechIcon id="rust" :size="20" />`, **Then** it outputs a semantic `<i class="tech-icon ...">` backed by the dynamic icon theme stylesheet.
2. **Given** a file name or extension (e.g., `Cargo.toml` or `py`), **When** passed to `<TechIcon id="Cargo.toml" />`, **Then** the engine checks `fileNames`, then `fileExtensions`, then `languageIds` in standard VS Code order.

---

### User Story 3 - 100% VS Code Icon Theme Schema Compatibility & Extensibility (Priority: P2)

As an enterprise developer or advanced user, I want UniRTM's icon theme system to strictly adhere to VS Code's `icon-theme.json` schema, so that third-party icon packs (such as Material Icon Theme or custom enterprise icon packs) can be imported or customized simply by placing a JSON definition.

**Why this priority**: Unlocks open ecosystem compatibility and prevents vendor lock-in.

**Independent Test**:
- Load the official VS Code `theme-seti/icons/seti.json` directly into the engine; it parses and registers all 2,000+ file/language definitions without syntax or mapping errors.

**Acceptance Scenarios**:
1. **Given** a valid VS Code `icon-theme.json` file with `fonts`, `iconDefinitions`, `languageIds`, `fileExtensions`, and `fileNames`, **When** loaded by `IconThemeService`, **Then** all glyphs, colors, and mappings are registered and rendered identically to VS Code.
2. **Given** an invalid or corrupted theme JSON, **When** loaded, **Then** the service logs a clear warning and falls back to the safe built-in default theme.

---

### User Story 4 - Commercial & Licensing Compliance (Priority: P1)

As an open-source or commercial distributor of UniRTM, I want to ensure that all fonts, glyphs, and code borrowed or referenced from VS Code and Seti UI fully comply with MIT and CC-BY-SA licenses, with all legal notices preserved.

**Why this priority**: Essential prerequisite explicitly requested by the user.

**Acceptance Scenarios**:
1. **Given** the repository root, **When** checking `THIRD-PARTY-LICENSES.md` or `LICENSE`, **Then** the MIT License for VS Code extension code and CC-BY / MIT for Seti/Codicons font assets are properly attributed.

---

## Technical Constraints & Requirements

1. **Schema Compliance**: Must support the full VS Code File Icon Theme schema:
   - `fonts`: Font family, font format (`woff2`/`woff`/`ttf`), font weight, style.
   - `iconDefinitions`: Keyed icon definitions with `fontCharacter`, `fontColor`, and optional `iconPath`.
   - `languageIds`: Map from language identifier to definition key.
   - `fileExtensions`: Map from file extension to definition key.
   - `fileNames`: Map from exact filename to definition key.
   - `file`: Default fallback file icon.
2. **Offline & Self-Contained**:
   - Zero runtime network calls to external CDNs.
   - Standard `seti.woff` (~140KB) and `seti.json` embedded directly into frontend assets.
3. **Performance**:
   - Dynamic stylesheet generated and injected once upon theme activation.
   - O(1) resolution via pre-computed JavaScript `Map` or hash table.
4. **Theme Interoperability**:
   - Works seamlessly in both Dark and Light themes (supporting `fontColor` inheritance or dark/light overrides).
