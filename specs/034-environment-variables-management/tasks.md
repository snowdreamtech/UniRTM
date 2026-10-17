# Tasks Breakdown: Environment Variables Management (034-environment-variables-management)

## Phase 1: Specification & Backend API (Go & Vite Dev Server)
- [x] **1.1** Go 后端实现 `/api/env` API 处理器（`handleGetEnvVars`, `handleUpsertEnvVar`, `handleDeleteEnvVar`）并挂载至 `cmd/64.ui.go`。
- [x] **1.2** Vite 开发服务器在 `frontend/vite.config.ts` 中实现 `/api/env` 对应端点（GET / POST / PUT / DELETE），支持本地文件直接持久化。
- [x] **1.3** 单元测试验证：针对 `internal/config/edit.go` 或 API 路由编写测试确保读写无误。

## Phase 2: Frontend Route & Navigation Integration
- [x] **2.1** 在 `frontend/src/router/index.ts` 中注册 `/environments` 路由。
- [x] **2.2** 在 `frontend/src/components/layout/ActivityBar.vue` 与 `SidebarPanel.vue` 中添加「环境变量」入口并绑定 `codicon-variable-group` 图标。
- [x] **2.3** 在各语言包（`src/i18n/`）中补充 `environments` 与 `sidebar.environments` 的国际化翻译项（全 53 语言已同步）。

## Phase 3: EnvironmentsView 布局与核心 CURD
- [x] **3.1** 创建 `frontend/src/views/EnvironmentsView.vue`，结构对齐 `TasksView.vue` 与 `ToolsView.vue`。
- [x] **3.2** 实现双作用域（`scope=workspace` 与 `scope=user`）同步切换、搜索过滤及空状态展示。
- [x] **3.3** 实现新建/编辑弹窗 `frontend/src/components/EnvironmentModal.vue`（自动遵循全局当前作用域，无需冗余选择）。
- [x] **3.4** 实现删除确认与删除交互。

## Phase 4: 安全脱密展示与小眼睛交互
- [x] **4.1** 实现敏感词智能判定算法（匹配 TOKEN, KEY, SECRET, PASS, AUTH, PRIVATE 等）。
- [x] **4.2** 敏感字段默认掩码显示（`••••••••••••••••`），普通变量直接明文免去繁琐操作。
- [x] **4.3** 实现小眼睛图标切换按钮（原生 `codicon-eye` / `codicon-eye-closed`），仅针对敏感密钥切换明文与密文。
- [x] **4.4** 实现一键复制明文至剪贴板功能，并弹出 Toast 提示。

## Phase 5: 验证、回归测试与提交
- [x] **5.1** 运行 `npm run test` 确保所有前端单元测试通过（55 项测试 100% 通过）。
- [x] **5.2** 运行 `npm run build` 确保 TypeScript 检查与打包构建无误。
- [x] **5.3** 运行 Go 测试（`go test ./internal/config/...`）确保配置读写稳定。
- [x] **5.4** 使用 GPG 密钥完成规范化提交并总结汇报。
