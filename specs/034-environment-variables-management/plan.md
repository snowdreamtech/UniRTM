# Technical Architecture & Plan: Environment Variables Management (034-environment-variables-management)

## 1. Architecture Overview

### 1.1 Backend Storage & Handlers
- **Storage Target**:
  - `workspace`: `./.unirtm.toml`（或当前工作目录下的 `.unirtm.toml`）中的 `[env]` 区块。
  - `user`: `~/.config/unirtm/unirtm.toml` 中的 `[env]` 区块。
- **Core Operations**:
  - 利用已有的 `internal/config/edit.go` 中的 `UpsertEnvVar(content, key, value)` 与 `UnsetEnvVar(content, key)`。
  - 在 `cmd/64.ui.go` 注册 `/api/env` 路由处理器：
    - `handleGetEnvVars`: 提取指定 scope 下的 `[env]` 映射，按名称字典序排序，并自动做敏感词模式匹配（`isSensitiveKey`）。
    - `handleUpsertEnvVar`: 支持创建与更新，若 key 变更则删除旧 key 并插入新 key，自动调用 `config.WriteTOMLFile`。
    - `handleDeleteEnvVar`: 删除指定 key 并持久化回写。
- **Frontend Dev Mock**:
  - 在 `frontend/vite.config.ts` 中注册对应的 `/api/env` 模拟接口，保证纯前端开发与测试模式（`npm run dev`）下数据实时读写本地配置文件。

### 1.2 Frontend Components & Views
- **View**: `frontend/src/views/EnvironmentsView.vue`
  - 模版对齐 `TasksView.vue` 与 `ToolsView.vue`：
    - 顶部 Header：标题「环境变量」、`ScopeSwitcher`（工作区 / 用户全局）、搜索框、快速新增按钮。
    - 列表卡片网格：`glass-card` 设计，包含 Key、脱密后的 Value、所属作用域 Badge、小眼睛切换按钮、复制按钮、编辑按钮、删除按钮。
    - 空状态提示与快速创建引导。
- **Modal Component**: `frontend/src/components/EnvironmentModal.vue`
  - 新建/编辑弹窗：
    - 键名输入框（支持自动大写处理）
    - 键值输入框（自带明文/密文切换）
    - 作用域选择（单选或者下拉框：工作区 vs 用户区）
    - 常用变量快捷预设提示（如 `PORT`, `NODE_ENV`, `HTTP_PROXY`, `DEBUG` 等）
- **Sidebar Integration**:
  - `frontend/src/components/layout/SidebarPanel.vue`: 在工作台主导航中添加「环境变量」入口，使用 `workbench:config` / `codicon-variable` 图标。
  - `frontend/src/router/index.ts`: 注册路由 `/environments`。
- **Security & Masking Logic**:
  - 敏感词正则匹配规则：`/(TOKEN|KEY|SECRET|PASSWORD|PASS|AUTH|PRIVATE|CREDENTIAL|CERT|SALT|HASH|SIGNATURE)/i`。
  - 默认 `isRevealed = false`，脱密显示为 `••••••••••••`。
  - 点击小眼睛图标切换 `isRevealed`。
  - 复制按钮直接将真实值写入剪贴板，并触发轻量 Toast 提示。
