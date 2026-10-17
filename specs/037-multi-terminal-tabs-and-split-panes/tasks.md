# 任务分解与执行清单 (Tasks)

- [x] **Task 1: 前端多终端会话与拆分状态管理 (Composable & Tests)**
  - 实现 `frontend/src/composables/useTerminalSessions.ts`：
    - 定义 `TerminalPane`、`TerminalSession`、`SplitDirection` 类型；
    - 支持新建终端会话 `createSession`（以系统默认或指定 Shell）；
    - 支持竖直拆分 `splitActiveSession('vertical')` 与横向拆分 `splitActiveSession('horizontal')`；
    - 支持关闭窗格 `closeActivePane` / 销毁会话 `closeSession`；
    - 支持独立输入、命令历史与输出记录；
  - 编写 `useTerminalSessions.test.ts` 单元测试验证核心逻辑；
  - *提交命名*：`feat(terminal): implement useTerminalSessions composable with multi-tab and split support`

- [x] **Task 2: 国际化多语言补齐 (i18n Locales)**
  - 在 `en-US.ts` 与 `zh-CN.ts` 中补充多终端与分屏操作相关词条：
    - `terminal.new_terminal`: 新建终端 / New Terminal
    - `terminal.split_vertical`: 竖直拆分终端 (一分为二) / Split Terminal Vertically
    - `terminal.split_horizontal`: 横向拆分终端 (一分为二) / Split Terminal Horizontally
    - `terminal.kill_terminal`: 关闭当前终端 / Kill Terminal
    - `terminal.tabs_title`: 终端标签列表 / Terminal Tabs
    - `terminal.split_badge`: 拆分 / Split
  - 同步更新全站 53 个语言文件并运行 `i18n.test.ts` 验证；
  - *提交命名*：`feat(i18n): add multi-terminal and split pane translation keys for 53 locales`

- [x] **Task 3: 底部抽屉工具栏与 VSCode 操作按钮集成 (Toolbar Actions)**
  - 在 `frontend/src/components/layout/BottomDrawer.vue` 顶部操作区中：
    - 集成 `+` 新建终端按钮与下拉 Shell 选单；
    - 集成竖直拆分按钮 `[ | ]` 与横向拆分按钮 `[ — ]`；
    - 集成垃圾桶关闭按钮（Kill Terminal）；
    - 联动 `useTerminalSessions` 状态；
  - *提交命名*：`feat(drawer): add vscode-like terminal toolbar actions for new, split, and kill`

- [x] **Task 4: VSCode 风格侧边终端列表与分屏视窗渲染 (Terminal Sidebar & Split Panes)**
  - 在终端视窗内实现两栏布局：
    - 左侧主区域：根据 `splitDirection`（vertical/horizontal/none）灵活排版窗格（flex-row 或 flex-col）；
    - 每个窗格支持独立输入行、命令历史、欢迎信息以及激活边框高亮（focus border）；
    - 右侧侧边栏：VSCode 经典 Terminal Tabs 列表，展示所有终端会话卡片、图标与悬浮关闭图标；
  - 完善深色与浅色主题适配；
  - *提交命名*：`feat(terminal): render vscode-style split panes and terminal tabs sidebar`

- [ ] **Task 5: 全量回归与构建验证 (Validation & Polish)**
  - 运行全量 Vitest 单元测试，确保测试套件全部通过；
  - 运行 `vue-tsc` 与生产打包编译，零警告零报错；
  - *提交命名*：`chore(terminal): complete multi-terminal split verification and tests`
