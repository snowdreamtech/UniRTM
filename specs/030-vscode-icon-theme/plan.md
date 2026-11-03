# Implementation Plan: VS Code Icon Theme 100% Specification Alignment

**Branch**: `030-vscode-icon-theme` | **Date**: 2026-10-15 | **Spec**: [specs/030-vscode-icon-theme/spec.md](spec.md)

**Input**: Feature specification from `/specs/030-vscode-icon-theme/spec.md`

## Summary

Implement an extensible, declarative icon theme engine for UniRTM that 100% aligns with the official VS Code File Icon Theme specification (`icon-theme.json`). The system includes:
1. Canonical TypeScript types and JSON Schema matching VS Code's specification.
2. Built-in Seti UI default theme (`seti.woff` and `seti.json` extracted from official VS Code extension under MIT/CC-BY).
3. `IconThemeService` (Vue 3 composable `useIconTheme`) managing multi-tier resolution (`fileNames` -> `fileExtensions` -> `languageIds` -> `default file`) and dynamic CSS stylesheet injection.
4. Universal `<TechIcon>` component replacing all hardcoded emojis across the workbench (`RuntimesView`, `SidebarPanel`, `ToolsView`, `TitleBar`).
5. Complete third-party license documentation ensuring commercial safety.

---

## Technical Context

**Language/Version**: TypeScript 5.7, Vue 3.5, Vite 6.4

**Primary Dependencies**: Vue 3 (`vue`), Vue Router 4 (`vue-router`)

**Storage**: Local assets (`frontend/src/assets/themes/seti/`), User settings persisted via `/api/config` & `localStorage`

**Testing**: `vue-tsc --noEmit`, `vite build`, Go single binary build (`go build`)

**Target Platform**: Linux, macOS, Windows (Browser & Desktop WebView)

**Project Type**: Web application & embedded Go single-binary desktop workbench

**Performance Goals**:
- Icon resolution: $O(1)$ in $<0.01$ ms per item
- CSS injection: One-time `<style>` tag generation $(<5$ ms on app mount)
- Zero layout shift (CLS = 0)

**Constraints**:
- 100% offline-capable; zero runtime network requests to CDNs or external sites
- Commercial licensing compliance (retain MIT and CC-BY notices)
- Strict compliance with VS Code `icon-theme.json` specification

---

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **Principle I: Self-Contained & Modular**: The icon theme engine lives entirely under `frontend/src/services/icon-theme/` and `frontend/src/components/common/TechIcon.vue`.
- [x] **Principle II: Declarative & Extensible**: Driven purely by standard JSON manifests; zero hardcoding in UI views.
- [x] **Principle III: Verification & Quality**: Zero build errors, type safety via TypeScript, manual and visual regression testing across all views.
- [x] **Principle IV: Commercial & Legal Safety**: MIT / CC-BY notices documented; no copyleft risks.

---

## Project Structure

### Documentation (this feature)

```text
specs/030-vscode-icon-theme/
├── plan.md              # This implementation plan
├── research.md          # Phase 0 research & license analysis
├── data-model.md        # Phase 1 data model & contracts
├── quickstart.md        # Phase 1 usage guide
├── contracts/           # Phase 1 schema contracts
│   └── icon-theme-schema.json
└── tasks.md             # Phase 2 task breakdown (created by /speckit.tasks)
```

### Source Code Architecture

```text
frontend/
├── src/
│   ├── assets/
│   │   └── themes/
│   │       └── seti/
│   │           ├── seti.json        # Canonical Seti UI mapping table
│   │           └── seti.woff        # Vector font glyphs (140KB)
│   ├── types/
│   │   └── icon-theme.ts            # Canonical VS Code icon theme TS interfaces
│   ├── services/
│   │   └── iconThemeService.ts      # Theme loader, CSS dynamic injector, O(1) resolver
│   ├── composables/
│   │   └── useIconTheme.ts          # Reactive Vue composable for components
│   ├── components/
│   │   └── common/
│   │       └── TechIcon.vue         # Universal icon component (<TechIcon :id="id" />)
│   └── views/
│       ├── RuntimesView.vue         # Refactored: emojis replaced with <TechIcon>
│       ├── ToolsView.vue            # Refactored: emojis replaced with <TechIcon>
│       └── components/layout/
│           └── SidebarPanel.vue     # Refactored: emojis replaced with <TechIcon>
└── THIRD-PARTY-LICENSES.md          # Legal attribution for Seti UI and VS Code
```

**Structure Decision**: Web application frontend layer with local asset vendoring. Decoupled from Go backend; backend requires no modification because all theme assets compile directly into the single binary via Vite and Go `embed.FS`.

---

## Implementation Phases

### Phase 1: Engine & Asset Foundation
1. Create `frontend/src/types/icon-theme.ts` with 100% VS Code schema types.
2. Vendor `seti.woff` and `seti.json` under `frontend/src/assets/themes/seti/`.
3. Create `frontend/src/services/iconThemeService.ts` to parse manifests, inject `@font-face` / dynamic CSS rules, and build fast lookup maps.
4. Create `frontend/src/composables/useIconTheme.ts` exposing `resolveIcon(id)` and `activeTheme`.
5. Create `frontend/src/components/common/TechIcon.vue`.

### Phase 2: Workbench UI Migration
1. Update `SidebarPanel.vue`: Replace hardcoded emojis (`🔷`, `🟢`, `🐍`, `🦀`) with `<TechIcon :id="item.id" :size="16" />`.
2. Update `RuntimesView.vue`: Replace card emojis with `<TechIcon :id="runtime.id" :size="32" :colored="true" />`.
3. Update `ToolsView.vue`: Replace tool emojis with `<TechIcon :id="tool.id" :size="24" :colored="true" />`.
4. Update `DashboardView.vue` and `BottomStatusBar.vue` where applicable.

### Phase 3: Compliance & Validation
1. Add legal attribution to `THIRD-PARTY-LICENSES.md`.
2. Verify TypeScript type checking (`npm run build`).
3. Verify Go single-binary packaging (`go build`).
4. Validate visual fidelity in dark and light themes.
