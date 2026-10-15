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

### 4. Frontend UI Architecture (`MarketplaceView.vue` & `SettingsView.vue`)
- **Dedicated Marketplace (`frontend/src/views/MarketplaceView.vue`)**:
  - Two decoupled channels/tabs:
    1. **Themes & Icons (Open VSX)**:
       - Full search bar with quick recommendations (Material, Dracula, One Dark, etc.).
       - Grid of theme cards (Icon, DisplayName, Namespace, Version, Downloads, Rating, Description).
       - One-click "安装" / "已安装" / "应用" actions.
    2. **Plugins & Extensions (UniRTM Ecosystem)**:
       - Native plugins catalog (UniAgent, Security Suite, tool providers).
       - Isolated namespaces explanation and guidelines.
- **Slimmed Settings (`frontend/src/views/SettingsView.vue`)**:
  - Focused solely on preferences:
    - Color theme & File icon theme dropdown selectors.
    - Local custom installed theme list & manual `.vsix` file upload.
    - Navigation prompt: `🎨 获取更多主题与图标 ↗` linking directly to `/marketplace?tab=themes`.

### 5. Extension Detail Architecture (US5)
- **Backend Service Layer (`internal/theme/marketplace.go`)**:
  - `GetOpenVSXExtensionDetail(namespace, name string) (*MarketplaceExtensionDetail, error)`
  - Queries `https://open-vsx.org/api/{namespace}/{name}`.
  - Fetches extension metadata, concurrent fetch of README.md, and parses `package.json` contributions (`contributes.themes` and `contributes.iconThemes`).
- **REST API Endpoint (`cmd/64.ui.go`)**:
  - `GET /api/themes/marketplace/detail?namespace=<ns>&name=<name>`.
- **Frontend Composable & Router (`useThemeMarketplace.ts`, `router/index.ts`)**:
  - Route support: `/marketplace/:namespace?/:name?` allows deep-linking directly to extension details.
  - Composable method: `fetchExtensionDetail(namespace: string, name: string): Promise<MarketplaceExtensionDetail>`.
  - Zero-dependency secure markdown renderer: transforms headings, images, code blocks, lists, and tables.
  - Detailed view UI:
    - Breadcrumb navigation (`← 返回市场列表`).
    - Hero header (Icon, Name, Publisher badge, Version, Stats, Minimalist Install action).
    - Tabs: Overview/README (with preview screenshots), Contributed Sub-themes (with light/dark badges & 1-click Apply), Changelog.
    - Sidebar: Repository link, Open VSX link, Issue tracker, License, Tags, Direct VSIX download.
