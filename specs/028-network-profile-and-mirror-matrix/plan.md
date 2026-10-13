# Technical Plan: Network Profiles, Domestic Mirror Matrix & Cascading Overrides

## 1. 架构目标与设计原则

1. **绝对中立与安全**：系统绝不内置任何第三方代理服务器或 GitHub 中继；代理默认关闭、地址留空。
2. **三层解析模型**：`Overrides (单项覆盖) > Active Profile (激活档案) > Official Baseline (官方兜底)`。
3. **沙箱式环境变量注入**：UniRTM 在调度 shims 或运行任务时通过子进程注入环境，不污染、不篡改宿主环境与永久配置。
4. **两端完全对齐与母语化**：CLI 与 Web 界面拥有 1:1 的完整功能，全部 UI 与 CLI 输出全面支持 i18n 母语化（中文、英文等）。

---

## 2. 后端核心模块规划 (Go Backend)

### 2.1 配置解析器与工厂 (`internal/network/`)
- `internal/network/types.go`：定义配置结构体与 DTO。
- `internal/network/defaults.go`：硬编码原厂出厂标准模板（`official` 与 `cn`），用于初始化与 `ResetToDefault`。
- `internal/network/manager.go`：
  - `GetStatus(cfg *config.Config) *ActiveNetworkStatus`：计算解析出全生态最终生效的源与来源标记。
  - `SetProfile(profileID string) error`：切换激活 Profile。
  - `CreateProfile(p NetworkProfile) error`：创建新自定义 Profile。
  - `SaveAsProfile(newID, name string) error`：把当前有效状态打包另存为新 Profile。
  - `ResetProfile(profileID string) error`：重置内置 Profile 为出厂初始值。
  - `DeleteProfile(profileID string) error`：安全删除自定义 Profile，自动回退激活。
  - `SetOverride(ecosystem, url string) error`：设置单项语言覆盖。
  - `ResetOverride(ecosystem string) error`：清除单项语言覆盖。
  - `SetProxy(p ProxyConfig) error`：保存代理开关与地址。
- `internal/network/probe.go`：
  - 并发 HTTP HEAD/GET 探针，并发测试各大源 RTT 往返耗时（带 2.5s 超时）。

### 2.2 运行期沙箱环境注入 (`internal/runner/` / `cmd/run.go`)
- 在构建子进程 `exec.Cmd.Env` 时，自动合并：
  - `HTTP_PROXY`, `HTTPS_PROXY`, `ALL_PROXY`, `NO_PROXY`（若 `proxy.enabled == true`）
  - `NPM_CONFIG_REGISTRY`
  - `GOPROXY`
  - `PIP_INDEX_URL`
  - `CARGO_REGISTRIES_CRATES_IO_INDEX`
  - `NODEJS_ORG_MIRROR`

### 2.3 CLI 命令组 (`cmd/network.go`)
- `unirtm network status`
- `unirtm network profile [list | use <id> | create <id> | save-as <id> | reset <id> | delete <id>]`
- `unirtm network registry [set <lang> <url> | reset <lang> | list]`
- `unirtm network proxy [on | off | set <url>]`
- `unirtm network test`

### 2.4 Web 后端 REST API (`cmd/ui.go`)
- `GET  /api/network/status`
- `POST /api/network/profile/use`
- `POST /api/network/profile/save`
- `POST /api/network/profile/reset`
- `POST /api/network/profile/delete`
- `POST /api/network/registry/override`
- `POST /api/network/registry/reset`
- `POST /api/network/proxy`
- `GET  /api/network/ping`

---

## 3. 前端界面与国际化规划 (Vue 3 Frontend)

### 3.1 页面重构与视觉设计
1. **SettingsView.vue（设置视图）**：
   - 增加独立一级导航或优化网络标签栏：
     - **“网络代理 (Proxy)”板块**：
       - 主开关（默认关）
       - 代理地址输入框（默认空，由用户输入本地端口，明确说明不提供第三方代理）
       - `NO_PROXY` 默认安全白名单
     - **“镜像源管理 (Mirrors & Registries)”板块**：
       - 顶部卡片栏：`🚀 国内镜像加速`、`🌐 官方直连`、自定义 Profile 卡片，带 `[+ 新建档案]` 和 `[↺ 恢复出厂设置]`。
       - 各语言列表（Node.js, Go, Python, Rust, Ruby）：
         - 知名公共大厂/高校镜像下拉快捷菜单
         - 输入框与实时状态徽章（`[来自预设]` / `[单项覆盖]`）
         - `[↺ 恢复跟随]` 按钮
         - `[⚡ 测速]` 按钮与毫秒指示灯（绿色 <100ms，黄色 <300ms，红色超时）
2. **DashboardView.vue（仪表盘视图）**：
   - 网络态势指示卡片：清晰显示当前 Profile、代理开关状态、一键全面测速。

### 3.2 国际化 (i18n) 全覆盖
在 `types.ts`、`zh-CN.ts`、`zh-TW.ts`、`en-US.ts` 等中全面扩展：
- `network.proxy_title`: "网络代理" / "Network Proxy"
- `network.proxy_desc`: "全局网络代理配置（支持本地或企业网关，默认不内置任何公网代理）"
- `network.mirrors_title`: "镜像源矩阵" / "Registry Mirrors"
- `network.profile_cn`: "国内镜像加速" / "China Fast Mirrors"
- `network.profile_official`: "官方直连" / "Official Direct"
- `network.badge_overridden`: "单项覆盖" / "Overridden"
- `network.badge_inherited`: "跟随预设" / "Inherited"
- `network.reset_factory`: "恢复出厂默认值" / "Reset to Factory Defaults"
- `network.restore_inherited`: "恢复跟随预设" / "Follow Profile"
- `network.save_as_profile`: "另存为新档案" / "Save as New Profile"
- `network.speed_test`: "实时测速" / "Speed Test"
- 严禁任何形式的中英文混合括号文本。
