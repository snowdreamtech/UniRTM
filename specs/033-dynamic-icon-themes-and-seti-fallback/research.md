# Research & Design Decisions: 033-dynamic-icon-themes-and-seti-fallback

## 1. Universal Workbench Icon Theme Integration & Semantic Mapping

### Context & Problem
Currently, UniRTM WEB uses `<TechIcon>` in a few views (`RuntimesView`, `SidebarPanel` runtime/tool items, and `BottomStatusBar` lock status). However, the major workbench UI surfaces (`ActivityBar`, `SidebarPanel` navigation/status items, `DashboardView` metrics cards, `DoctorView` checklist cards, `TasksView` task lists, `MarketplaceView`, and `SettingsView` tabs) still rely on hardcoded Unicode Emojis (e.g. 📊, 🌐, 🛠️, ⚡, 🧩, 🩺, ⚙️, ℹ️, 📜, 🚀, 💻, 📦, 🔌, ❌, ⚠️, ✅).
When users install and activate a custom icon theme (such as Material Icon Theme or VS Code Minimal Icons), these static emojis remain untouched, breaking visual cohesion.

### Decision
1. **Workbench Semantic Icon Dictionary (`workbenchSemanticIcons`)**:
   Establish a centralized semantic mapping dictionary in `frontend/src/services/icon-theme/workbenchIcons.ts`.
   This maps workbench intent IDs (e.g. `workbench:dashboard`, `workbench:runtimes`, `workbench:tools`, `workbench:tasks`, `workbench:marketplace`, `workbench:doctor`, `workbench:settings`, `workbench:about`, `workbench:terminal`, `workbench:extensions`, `workbench:env`, `workbench:success`, `workbench:warning`, `workbench:error`, etc.) to primary file names, extensions, or language IDs recognized by standard VS Code icon themes.
   - Example mappings:
     - `workbench:dashboard` -> `chart` / `stats` -> Seti fallback: `_default` / `json` (`\uE023` / `\uE01A`)
     - `workbench:runtimes` -> `globe` / `html` / `js`
     - `workbench:tools` -> `build` / `tool` / `makefile` (`\uE02D`)
     - `workbench:tasks` -> `task` / `sh` / `shell` / `bash` (`\uE034`)
     - `workbench:marketplace` -> `plugin` / `package.json` (`\uE031`)
     - `workbench:doctor` -> `check` / `test` / `spec`
     - `workbench:settings` -> `settings.json` / `config` (`\uE019`)
     - `workbench:terminal` / `workbench:logs` -> `console` / `log` / `sh` (`\uE034`)
     - `workbench:about` -> `info` / `readme.md` / `markdown` (`\uE02C`)
2. **Component Upgrades**:
   - Refactor `ActivityBar.vue`, `SidebarPanel.vue`, `DashboardView.vue`, `DoctorView.vue`, `TasksView.vue`, `MarketplaceView.vue`, and `SettingsView.vue` to replace all emoji icons with `<TechIcon :id="identifier" />`.
   - Update `TechIcon.vue` to support transparent sizing, clean SVG/font rendering, and optional CSS inheritance when `:colored="false"`.

### Alternatives Considered
- *Using third-party icon fonts like FontAwesome or Lucide alongside Icon Themes*: Rejected because it introduces dual icon systems, increases bundle size, and does not respond when the user switches icon theme packages from OpenVSX.
- *Hardcoding SVGs*: Rejected because user icon themes could not customize workbench icons.

---

## 2. Multi-Tier Seti UI Fallback Engine

### Context & Problem
Third-party icon themes downloaded from OpenVSX or custom user themes frequently have coverage gaps (e.g., missing specific language IDs like `zig`, `gleam`, `unirtm.lock`, or lacking definitions for custom workbench semantic keys). If a key is missing in the active theme, currently `IconThemeService` calls `getDefaultIcon()` which returns the default file icon of that active theme. If that active theme's default icon itself is broken, missing, or unsuitable, users see blank squares or broken layouts. Furthermore, when a third-party theme lacks a specific icon (like `dockerfile` or `go`), falling back directly to a generic blank file is suboptimal when Seti UI has a rich, high-quality icon for it.

### Decision
Implement a 2-tier resolution pipeline with Seti UI fallback:
1. **Tier 1 (Active Custom Theme)**:
   - Check `fileNames` -> `fileExtensions` -> `languageIds` in active manifest.
   - If resolved, return the custom theme's icon (SVG or Font).
2. **Tier 2 (Seti UI Precise Fallback)**:
   - If not resolved in active theme, check `fileNames` -> `fileExtensions` -> `languageIds` in `defaultSetiManifest`.
   - If resolved in Seti UI, return the Seti Font icon with its canonical font character and color.
3. **Tier 3 (Seti UI Default Fallback)**:
   - If still unresolved in both active theme and Seti UI definitions, fall back to Seti UI's default file icon `_default` (`\uE023`, `#d4d7d6`).

### Performance & Cache
- In-memory resolution cache (`Map<string, ResolvedIcon>`) in `IconThemeService` reset only on theme switch.
- Resolution latency guaranteed $<0.01$ ms per lookup.
