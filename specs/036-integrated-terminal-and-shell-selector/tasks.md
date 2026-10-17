# 任务分解与执行清单 (Tasks)

- [x] **Task 1: 后端终端扫描接口与配置持久化支持 (Backend & Schema)**
  - 在 `schemas/unirtm.schema.json` 中为 `settings` 补充 `terminal` 字段定义；
  - 在 `cmd/64.ui.go` 与 `frontend/vite.config.ts` 中实现 `GET /api/terminals` 接口：
    - 读取 `/etc/shells` 与查找常见 Shell（bash, zsh, fish, sh, pwsh, cmd, git-bash）；
    - 智能匹配当前系统环境变量 `$SHELL` 或 `COMSPEC` 判定 `is_default`；
    - 读取用户级配置 `settings.terminal` 判定 `current`；
  - 完善持久化写入支持（支持通过 `/api/config` 写入并保存）；
  - *提交命名*：`feat(api): implement terminal detection and persistence endpoints`

- [x] **Task 2: 前端终端管理 Composable 与国际化字典 (Frontend Composable & i18n)**
  - 创建 `frontend/src/composables/useTerminal.ts`：
    - 维护终端列表、当前终端、系统默认终端状态；
    - 封装 `fetchTerminals()` 与 `updateTerminal()` 方法；
  - 编写 `useTerminal.test.ts` 单元测试验证核心逻辑；
  - 在 `zh-CN.ts`、`en-US.ts` 等字典中补充终端相关的国际化词条（`drawer.tab_terminal`, `terminal.default`, `terminal.select_placeholder`, `terminal.rescan`, `settings.terminal_title` 等）；
  - *提交命名*：`feat(terminal): add useTerminal composable, tests and i18n locales`

- [x] **Task 3: 底栏面板集成「终端」Tab 与快速切换器 (BottomDrawer Integration)**
  - 在 `frontend/src/components/layout/BottomDrawer.vue` 中：
    - 增加第四个选项卡：`[ 终端 (Terminal) ]`（使用 `workbench:terminal` 图标）；
    - 在抽屉顶部右侧/操作区增加终端下拉切换器，展示已检测到的 Shell 列表及默认标记；
    - 抽屉内容区构建集成终端容器（包含终端欢迎信息、当前执行 Shell 信息、会话控制按扭）；
    - 支持清屏、全屏放大/还原与快捷打开设置；
  - 在 `frontend/src/App.vue` 中支持快捷键与任务启动时关联终端 Tab；
  - *提交命名*：`feat(drawer): integrate terminal tab with interactive shell switcher`

- [x] **Task 4: 设置中心「终端设置」卡片与可视化配置 (Settings Integration)**
  - 在 `frontend/src/views/SettingsView.vue` 中加入「集成终端设置」卡片：
    - 展示探测到的所有本地终端；
    - 提供路径、系统默认徽章与单选切换；
    - 点击后调用持久化保存并弹出 Toast 提示；
    - 支持手动点击“重新扫描终端”；
  - *提交命名*：`feat(settings): add terminal configuration section and persistence`

- [ ] **Task 5: 全面测试回归与构建验证 (Validation & Polish)**
  - 运行全量 Vitest 单元测试，确保无回归；
  - 运行 `vue-tsc` 与 Vite 生产构建；
  - 验证多语言切换与深浅主题切换下的终端视觉效果。
  - *提交命名*：`chore(terminal): complete end-to-end verification and test coverage`
