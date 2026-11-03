# Feature Specification: Environment Variables Management (034-environment-variables-management)

## 1. Summary
UniRTM 需要为工作区（Workspace Scope）与用户全局区（User Scope）提供系统化、直观且安全的环境变量（Environment Variables）管理能力。
界面布局与交互体系对齐现有的 `TasksView` (`/tasks?scope=workspace`) 与 `ToolsView` (`/tools?scope=workspace`)，URL 路径统一规范为 `/environments`（支持 `?scope=workspace` 与 `?scope=user`）。

核心能力包括：
1. **工作区与用户区双作用域隔离管理**：无缝读取并持久化写入 `.unirtm.toml`（工作区）与全局配置 `~/.config/unirtm/unirtm.toml`（用户区）的 `[env]` 区块。
2. **完整的 CRUD 操作**：支持环境变量的添加（Create）、展示过滤与搜索（Read）、编辑修改（Update）、删除（Delete）。
3. **严格的安全脱密机制**：
   - 包含敏感标识（如 `TOKEN`、`KEY`、`SECRET`、`PASS`、`AUTH`、`CREDENTIAL`、`PRIVATE` 等）的字段值**默认进行脱密遮罩显示**（如 `••••••••`）。
   - 提供「小眼睛」显隐切换交互按钮，用户主动点击后才展示明文。
   - 提供安全一键复制功能，安全复制明文到系统剪贴板。

---

## 2. User Scenarios & Acceptance Criteria

### User Scenario 1: 查看环境变量列表与作用域切换 (Read)
- **Given** 用户访问 `/environments`。
- **When** 切换顶部作用域开关（工作区 / 用户全局）。
- **Then** 页面正确加载对应作用域下的 `[env]` 变量列表，显示变量名、脱密后的值、所属作用域标签，并支持关键词快速搜索过滤。

### User Scenario 2: 安全脱密与明文切换 (Security & Privacy)
- **Given** 环境变量名为 `GITHUB_TOKEN` 或 `DATABASE_PASSWORD` 等敏感键。
- **When** 变量渲染在列表中时。
- **Then** 默认处于脱密遮盖状态（`••••••••`）；点击小眼睛图标可单项切换为明文；再次点击重新脱密；支持一键复制明文。

### User Scenario 3: 新建环境变量 (Create)
- **Given** 用户点击右上角「新建变量」按钮。
- **When** 在弹窗中输入 Key（自动标准化格式）、Value，并选择保存至工作区或用户区，点击保存。
- **Then** 后端将变量写入对应配置文件的 `[env]` 区块中，列表即时更新并高亮提示。

### User Scenario 4: 编辑与重命名环境变量 (Update)
- **Given** 用户点击某一环境变量的「编辑」按钮。
- **When** 修改 Key 或 Value（若修改 Key 则删除旧 Key 并创建新 Key），点击保存。
- **Then** 配置文件正确同步更新，原有的脱密逻辑保持生效。

### User Scenario 5: 删除环境变量 (Delete)
- **Given** 用户点击某一环境变量的「删除」按钮。
- **When** 确认删除。
- **Then** 后端从配置文件的 `[env]` 区块中移除该键值对，配置文件规范化保存，列表即时移除。

---

## 3. Requirements & Constraints
- **URL**: `/environments` 支持 `scope=workspace` 与 `scope=user` 查询参数。
- **UI/UX**: 采用 UniRTM Glassmorphism 设计规范，侧边栏（SidebarPanel）新增「环境变量」导航项，适配浅色与深色主题。
- **Consistency**: 遵循 POSIX 与 Go 后端 `internal/config/edit.go` 中既有的 `UpsertEnvVar` 和 `UnsetEnvVar` 规范。
