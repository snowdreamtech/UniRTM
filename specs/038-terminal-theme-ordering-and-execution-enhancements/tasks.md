# 任务分解与执行清单 (Tasks)

- [x] **Task 1: 后端排序优化与 Shell 执行接口 (Backend Engine & Sorting)**
  - 在 `cmd/64.ui.go` 中：
    - 在 `detectInstalledTerminals()` 中对终端列表进行优先级排序，使 `is_default: true` 始终居于第一位；
    - 实现 `POST /api/terminal/exec` 接口，接收 `{ command, shell, cwd }`，自动注入系统环境与 UniRTM PATH，使用真实 Shell 执行并返回 output 和 exit_code；
    - 在 `cmd/64.ui_test.go` 中添加终端命令执行测试；
  - 在 `frontend/vite.config.ts` 中实现本地开发服务器对 `/api/terminal/exec` 的模拟支持；
  - *提交命名*：`feat(api): implement terminal command execution engine and default terminal top-sorting`

- [ ] **Task 2: 前端对接 Shell 执行引擎与排序保障 (Frontend Execution & Ordering)**
  - 在 `frontend/src/composables/useTerminal.ts` 中添加排序保障（系统默认排第一）；
  - 在 `frontend/src/components/layout/BottomDrawer.vue` 中：
    - 将 `handlePaneRunCommand` 改为请求 `/api/terminal/exec`，实现真正的 Linux 命令执行与输出回显；
  - *提交命名*：`feat(terminal): connect real shell execution engine and ensure default terminal sorting`

- [ ] **Task 3: 弹层向上展开与侧边栏加号/垃圾桶图标优化 (Popups & Sidebar Ergonomics)**
  - 修改 `.term-new-dropdown` 为向上弹出并带有滚动与阴影，防止被抽屉底部遮挡；
  - 加宽侧边栏至 185px，增大加号按钮并留出呼吸间距；
  - 将标签项关闭按钮替换为更清晰的大号垃圾桶图标（Trash Icon），加大点击区域与右边距；
  - *提交命名*：`feat(drawer): fix dropdown clipping with upward popup and improve sidebar ergonomics`

- [ ] **Task 4: 深色模式与浅色模式终端全套主题精细化配色 (Dark/Light Palette)**
  - 为 `BottomDrawer.vue` 补充浅色主题 `[data-theme="light"]` 专属变量：
    - 纯白终端背景、淡灰 Banner 与侧边栏背景，细致的边框；
    - 浅色模式下文字使用高对比度深色，命令提示符与高亮青色加深可读性；
    - 深色模式下保持低饱和度深色层次；
  - *提交命名*：`style(terminal): polish comprehensive dark and light theme styles for terminal`

- [ ] **Task 5: 全量回归与构建验证 (Validation & Tests)**
  - 运行全量 Vitest 单元测试；
  - 运行 Go 后端测试；
  - 运行 `vue-tsc` 与生产打包验证；
  - *提交命名*：`chore(terminal): complete validation and testing for theme and execution enhancements`
