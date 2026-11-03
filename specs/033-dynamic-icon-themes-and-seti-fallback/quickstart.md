# Quickstart & Verification Guide: 033-dynamic-icon-themes-and-seti-fallback

## 1. Prerequisites
- Node.js $\ge 18$
- Go $\ge 1.22$
- Repo root: `/home/ansible/workspace/snowdreamtech/UniRTM`

## 2. Type Check & Frontend Build
Run TypeScript type-check and Vite production build to ensure clean compilation:

```bash
cd frontend
npm run build
```

## 3. Web UI Functional Verification

### Scenario A: Universal Icon Coverage & No Native Emojis
1. Start the UniRTM web development server or run `unirtm ui`:
   ```bash
   cd frontend && npm run dev
   ```
2. Navigate to `http://localhost:5173`.
3. Inspect `ActivityBar`:
   - Dashboard, Runtimes, Tools, Tasks, Marketplace, Doctor, Settings, About, Terminal icons must all render via `<TechIcon>` without any OS Emojis.
4. Inspect `SidebarPanel`, `DashboardView`, `DoctorView`, `TasksView`, `SettingsView`:
   - All section headers, status badges, and action buttons render through theme icons.

### Scenario B: Dynamic Theme Switch
1. Go to **Settings** -> **Themes** -> **Icon Themes** (or Marketplace).
2. Switch from **Seti UI** to a third-party icon theme (e.g. Material Icon Theme).
3. Confirm that:
   - All views immediately reflect the new theme's icons.
   - Transition is flicker-free, instant, and causes no console errors.

### Scenario C: Fallback to Seti UI on Missing Icons
1. In a theme with limited icon coverage (or by requesting an obscure file/runtime name like `unirtm.lock`, `zig`, or `workbench:doctor`):
   - The icon resolver falls back to `defaultSetiManifest` and displays the Seti UI icon.
2. If completely unknown (e.g. `nonexistent-runtime-xyz`), it displays Seti UI's default file icon (`_default` / `\uE023`) in neutral gray (`#d4d7d6`).
