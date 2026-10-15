# Implementation Plan: 032-openvsx-theme-marketplace

## Architecture & Design

### 1. Backend Service Layer (`internal/theme/marketplace.go`)
- **OpenVSXClient**:
  - `SearchThemes(query string, size int, offset int) (*MarketplaceSearchResponse, error)`
  - URL: `https://open-vsx.org/api/-/search?category=Themes&query=<q>&size=<size>&offset=<offset>`
  - Uses `http.DefaultClient` or proxy-configured client from `internal/network`.
- **Online Installer**:
  - `InstallFromURL(downloadURL string) ([]ThemeItem, error)`
  - Fetches the `.vsix` stream directly and feeds it into `m.InstallVSIX(resp.Body)`.

### 2. REST API Endpoints (`cmd/64.ui.go`)
- `GET /api/themes/marketplace/search`: params `q`, `size`, `offset`.
- `POST /api/themes/marketplace/install`: json body `{ "download_url": "...", "package_id": "..." }`.

### 3. Frontend Composable & State (`frontend/src/composables/useThemeMarketplace.ts`)
- `searchQuery`: `ref("")`
- `searchResults`: `ref<MarketplaceExtension[]>`
- `isSearching`: `ref(false)`
- `installingMap`: `reactive<Record<string, boolean>>`
- `searchThemes(q?: string)`
- `installOnline(item: MarketplaceExtension)`

### 4. Frontend UI (`frontend/src/views/SettingsView.vue`)
- In "VS Code 主题扩展包管理":
  - Add search input for Open VSX marketplace with quick tags (热门/Popular, Dracula, Material, One Dark, etc.).
  - Grid/List of marketplace theme cards:
    - Icon / Logo
    - DisplayName, Publisher, Version
    - Download count, Rating
    - Button: "一键安装" (or "已安装" / "卸载")
    - Fast install and hot apply prompt.
