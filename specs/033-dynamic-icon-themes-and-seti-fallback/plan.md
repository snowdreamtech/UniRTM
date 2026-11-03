# Implementation Plan: Dynamic Icon Themes & Seti UI Fallback

**Branch**: `033-dynamic-icon-themes-and-seti-fallback` | **Date**: 2026-10-15 | **Spec**: [specs/033-dynamic-icon-themes-and-seti-fallback/spec.md](spec.md)

**Input**: Feature specification from `/specs/033-dynamic-icon-themes-and-seti-fallback/spec.md`

## Summary

Build an end-to-end dynamic icon theme engine and comprehensive fallback system for UniRTM WEB:
1. **Universal Icon Theme Wiring**: Wire all workbench navigation bars, sidebars, dashboard widgets, tasks, doctor, settings, and status items to `<TechIcon>`, eliminating all hardcoded Unicode Emojis.
2. **Workbench Semantic Icon Dictionary**: Establish standard mappings from workbench UI concepts (`workbench:dashboard`, `workbench:settings`, etc.) to theme candidates, allowing full visual skinning across themes.
3. **Multi-Tier Seti UI Fallback Engine**: Upgrade `IconThemeService.resolveIcon` to implement strict multi-tier fallback (Active Theme $\rightarrow$ Seti UI Exact $\rightarrow$ Seti UI `_default`), ensuring zero broken icons or empty squares.
4. **LRU / Map Caching**: Guarantee sub-millisecond lookup and zero layout shift.

## Technical Context

**Language/Version**: TypeScript 5.7, Vue 3.5, Vite 6.4

**Primary Dependencies**: Vue 3 (`vue`), Vue Router 4 (`vue-router`), `@snowdreamtech/unirtm`

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

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **Principle I: Self-Contained & Modular**: The icon theme engine lives entirely under `frontend/src/services/icon-theme/` and `frontend/src/components/common/TechIcon.vue`.
- [x] **Principle II: Declarative & Extensible**: Driven purely by standard JSON manifests; zero hardcoding in UI views.
- [x] **Principle III: Verification & Quality**: Zero build errors, type safety via TypeScript, manual and visual regression testing across all views.
- [x] **Principle IV: Commercial & Legal Safety**: MIT / CC-BY notices documented; no copyleft risks.

## Project Structure

### Documentation (this feature)

```text
specs/033-dynamic-icon-themes-and-seti-fallback/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
└── contracts/           # Phase 1 output
    └── icon-theme-contract.md
```

### Source Code (repository root)

```text
frontend/
├── src/
│   ├── assets/themes/seti/
│   │   ├── seti.json
│   │   ├── seti.woff
│   │   └── seti.woff2
│   ├── components/
│   │   ├── common/
│   │   │   └── TechIcon.vue
│   │   └── layout/
│   │       ├── ActivityBar.vue
│   │       ├── SidebarPanel.vue
│   │       └── BottomStatusBar.vue
│   ├── composables/
│   │   ├── useIconTheme.ts
│   │   └── useThemeMarketplace.ts
│   ├── services/
│   │   └── icon-theme/
│   │       ├── iconThemeService.ts
│   │       └── workbenchIcons.ts
│   ├── types/
│   │   └── icon-theme.ts
│   └── views/
│       ├── DashboardView.vue
│       ├── DoctorView.vue
│       ├── MarketplaceView.vue
│       ├── RuntimesView.vue
│       ├── SettingsView.vue
│       ├── TasksView.vue
│       └── ToolsView.vue
```

**Structure Decision**: Web application structure (`frontend/src/`). All frontend changes adhere to existing modular patterns and composables.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| *None*    | N/A        | N/A                                 |
