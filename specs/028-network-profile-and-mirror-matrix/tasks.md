# Tasks: Network Profiles, Domestic Mirror Matrix & Cascading Overrides

## Phase 1: Go Backend Data Structures & Network Engine
- [x] **T-101**: 创建 `internal/network/types.go`，定义 `NetworkConfig`、`NetworkProfile`、`ProxyConfig`、`EffectiveRegistry`。
- [x] **T-102**: 创建 `internal/network/defaults.go`，配置原厂出厂内置模板（`official` 与 `cn` 矩阵）与各生态推荐公共镜像列表。
- [x] **T-103**: 实现 `internal/network/manager.go`：
  - 核心解析器：`GetEffectiveStatus` 实行 `Overrides > ActiveProfile > Default` 级联计算。
  - Profile 生命周期 CRUD：新建、编辑、另存为 (`SaveAs`)、删除、内置出厂重置 (`ResetToDefault`)。
  - 单项覆盖支持：`SetOverride` 与 `ResetOverride`。
  - 代理管理：保存代理配置，严格校验协议与格式。
- [x] **T-104**: 实现 `internal/network/probe.go`：并发网络延迟测速探针，带 2.5s 超时控制。
- [x] **T-105**: 编写 Go 单元测试 `internal/network/manager_test.go`，覆盖全部级联优先级、出厂重置、代理默认关闭断言。

## Phase 2: CLI Subcommand Implementation (`unirtm network`)
- [x] **T-201**: 创建 `cmd/65.network.go`，注册 `unirtm network` 父命令。
- [x] **T-202**: 实现 `unirtm network status`：纯母语打印当前生效模式、代理状态及各语言当前生效镜像与来源。
- [x] **T-203**: 实现 `unirtm network profile [list | use | create | save-as | reset | delete]` 命令组。
- [x] **T-204**: 实现 `unirtm network registry [set | reset | list]` 命令组。
- [x] **T-205**: 实现 `unirtm network proxy [on | off | set]` 命令组。
- [x] **T-206**: 实现 `unirtm network test`：终端表格格式化并发测速。
- [x] **T-207**: 运行环境沙箱接入：在 `internal/runner` / `internal/network` 中接入网络配置，将生效环境变量 (`ExportEnvironment`) 注入子进程。

## Phase 3: Web Backend REST API
- [x] **T-301**: 在 `cmd/64.ui.go` 中注册 `/api/network/*` 路由端点。
- [x] **T-302**: 实现状态获取、Profile 切换、新建、重置、单项覆盖、代理设置及测速 API。
- [x] **T-303**: 编写 API 测试 `cmd/64.ui_network_test.go`，验证前后端数据交互与错误响应。

## Phase 4: Frontend UI Components & i18n Localization
- [x] **T-401**: 在 `frontend/src/i18n/types.ts`、`zh-CN.ts`、`zh-TW.ts`、`en-US.ts` 中补充完整的网络、代理与镜像矩阵词典（纯母语无杂质）。
- [x] **T-402**: 在 `frontend/src/views/SettingsView.vue` 中重构网络设置面板：
  - “网络代理”模块：主开关、地址输入、安全提示。
  - “镜像源管理”模块：顶部 Profile 档位卡片、新建模态框、出厂还原按钮。
  - 各语言生态矩阵表格：下拉知名镜像推荐、自定义输入框、覆盖徽章、恢复跟随按钮、单项测速。
- [x] **T-403**: 在 `frontend/src/views/DashboardView.vue` 中集成网络态势卡片，支持实时状态展示与快速测速。
- [x] **T-404**: 前端编译与生产构建校验 (`npm run build`)。

## Phase 5: End-to-End Verification & Documentation
- [x] **T-501**: CLI 与 Web 双端集成联动测试（在 Web 修改能实时反映在 CLI，在 CLI 修改 Web 自动刷新）。
- [x] **T-502**: 检查全系统文案，验证 100% 纯正母语化，绝无中英文夹杂。
- [x] **T-503**: 执行 Go 单元与集成测试套件验证。
