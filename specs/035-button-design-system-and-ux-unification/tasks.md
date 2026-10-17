# 任务分解与执行清单 (Tasks)

- [x] **Task 1: 全局按钮体系基建下沉 (CSS Foundation)**
  - 在 `frontend/src/styles/theme.css` 中引入 `.btn`, `.btn-primary`, `.btn-secondary`, `.btn-ghost`, `.btn-danger`, `.btn-icon`, `.btn-capsule` 等标准类；
  - 适配暗黑与明亮双主题；
  - 支持 `:active` 轻微物理按压反馈（`scale(0.96)`）。
  - *提交命名*：`style(theme): establish unified button design system classes in theme.css`

- [x] **Task 2: 环境变量模块操作统一与裸字符淘汰 (Environments)**
  - 重构 `frontend/src/views/EnvironmentsView.vue`：
    - 将 Profile 标签栏中的 `✕` 替换为标准 TechIcon；
    - 将卡片与列表中的 `↺` 恢复基准值按钮升级为 `.btn-icon.success` 并使用标准图标；
    - 将编辑、删除、显隐密码眼睛按钮、复制按钮全部对齐为标准 `.btn-icon` 体系；
  - 重构 `frontend/src/components/EnvironmentModal.vue`：
    - 将弹窗右上角的关闭按钮 `✕` 替换为 `<TechIcon id="workbench:close" />`；
    - 优化 `AA` 大写转换按钮与底部取消/确定按钮。
  - *提交命名*：`refactor(environments): unify action buttons and replace hardcoded raw characters`

- [x] **Task 3: 扩展市场操作列与卡片按钮治理 (Marketplace)**
  - 重构 `frontend/src/views/MarketplaceView.vue`：
    - 统一列表视图与卡片视图中的操作按钮（安装、应用、生效中）；
    - 统一度量衡为标准高度 `26px`，确保中英文切换时不发生高度跳动；
    - 统一已安装专栏的垃圾桶删除操作为标准 `.btn-icon.danger`。
  - *提交命名*：`refactor(marketplace): align action buttons and density with global design system`

- [x] **Task 4: 工具链与任务模块操作规范化 (Tools & Tasks)**
  - 检查并重构 `frontend/src/views/ToolsView.vue` 与 `frontend/src/views/TasksView.vue`；
  - 统一运行（Play）、升级、锁定、更多菜单的按钮规范，消除尺寸各异的内边距。
  - *提交命名*：`refactor(tools,tasks): standardize action buttons and run triggers`

- [ ] **Task 5: 设置与通用弹窗按钮规范化 (Settings & Modals)**
  - 重构 `frontend/src/views/SettingsView.vue` 与全局各类确认/编辑弹窗；
  - 统一保存、重置、别名删除、钩子执行等按钮规范。
  - *提交命名*：`refactor(settings): standardize form and modal buttons`

- [ ] **Task 6: 自动化测试回归与生产构建验证 (Validation)**
  - 运行 Vitest 55 项单元测试确保无前端逻辑回归；
  - 运行 `vue-tsc` 与 `vite build` 生产构建；
  - 检查各页面主题（Dark/Light）对比度与交互响应。
