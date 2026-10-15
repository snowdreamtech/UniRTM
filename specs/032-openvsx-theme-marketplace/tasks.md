# Tasks: Open VSX Theme and Icon Marketplace Integration

**Input**: Design documents from `/specs/032-openvsx-theme-marketplace/`
**Prerequisites**: [plan.md](plan.md), [spec.md](spec.md), [data-model.md](data-model.md)

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel
- **[Story]**: User story reference

---

## Phase 1: Backend Client & Installation Engine

- [x] T001 [US1] Implement `internal/theme/marketplace.go`:
  - `SearchOpenVSXThemes(query string, size int, offset int)` with `category=Themes` query filter
  - `InstallFromURL(downloadURL string) ([]ThemeItem, error)` streaming directly to `m.InstallVSIX(resp.Body)`
- [x] T002 [P] [US1] Add unit tests in `internal/theme/marketplace_test.go` verifying search and online installation parsing.
- [x] T003 [US1/US2] Register REST endpoints in `cmd/64.ui.go`:
  - `GET /api/themes/marketplace/search`
  - `POST /api/themes/marketplace/install`

---

## Phase 2: Frontend Data Models & Composable

- [x] T004 [P] [US1] Add `MarketplaceExtension` and `MarketplaceSearchResponse` types in `frontend/src/types/theme-manager.ts`.
- [x] T005 [P] [US1/US2] Implement `frontend/src/composables/useThemeMarketplace.ts` with searching, pagination, and online installation methods.

---

## Phase 3: Workbench UI Integration

- [x] T006 [US1/US2/US3] Add Theme Marketplace browser into `frontend/src/views/SettingsView.vue`:
  - Online search bar with instant query and debounce
  - Filtered theme cards display (Logo, Title, Author, Downloads, Rating)
  - One-click "安装 (Install)" button with loading spinner
  - Installed status badge and "立即应用 (Apply)" shortcut
  - Direct integration with installed themes management

---

## Phase 4: Verification & Build Validation

- [x] T007 [P] Run `go test ./internal/theme/...` and build Go binary (`go build -o bin/unirtm main.go`).
- [x] T008 [P] Run `npm run build` in `frontend/` to verify zero TypeScript errors.
- [x] T009 Test end-to-end online search and installation against Open VSX.

---

## Phase 5: Information Architecture Migration (Marketplace View & Slimmed Settings)

- [x] T010 [US4] Rebuild `frontend/src/views/MarketplaceView.vue`:
  - Channels/Tabs: `🎨 主题与图标 (Open VSX)` and `🧩 功能插件 (UniRTM 生态)`
  - Full Open VSX search, chip filters, rich cards grid, install/apply actions
  - Route query param support (`?tab=themes` or `?tab=plugins`)
- [x] T011 [US4] Slim down `frontend/src/views/SettingsView.vue`:
  - Remove heavy Open VSX search grid from settings
  - Retain clean color/icon theme dropdown selectors and custom installed package list
  - Add quick link: `🎨 获取更多主题与图标 ↗` linking to `/marketplace?tab=themes`
- [x] T012 [P] [US4] Update sidebar/activity bar navigation and verify all flows with zero TypeScript/Vue errors.
