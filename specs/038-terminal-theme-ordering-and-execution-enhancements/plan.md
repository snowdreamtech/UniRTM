# 特性实施计划：终端主题美化、排序、弹层定位与命令执行引擎 (Plan)

## 1. 架构与改动点

### 1.1 后端与开发服务器
- `cmd/64.ui.go`:
  - 完善 `detectInstalledTerminals()` 排序逻辑，确保 `is_default: true` 的项排在索引 0；
  - 注册并实现 `POST /api/terminal/exec` 接口，支持指定 Shell 执行命令并返回输出；
  - 在 `cmd/64.ui_test.go` 中添加测试用例；
- `frontend/vite.config.ts`:
  - 在开发服务器中间件增加 `/api/terminal/exec`，通过 `child_process.exec` 执行命令。

### 1.2 前端状态与组件
- `frontend/src/composables/useTerminal.ts`:
  - 保证终端列表接收时，系统默认终端排序在第一位；
- `frontend/src/components/layout/BottomDrawer.vue`:
  - 将命令运行逻辑从 `/api/tasks/run` 替换为 `/api/terminal/exec`，实现真正的 Linux/Shell 命令执行与 PATH 解析；
  - 修改 `.term-new-dropdown` 的绝对定位为向上弹出（`bottom: calc(...)`）；
  - 优化右侧 Terminal Tabs 侧边栏宽度、加号大小与右间距、将关闭按钮升级为大号垃圾桶图标；
  - 重写深浅模式 CSS 变量适配：针对 `[data-theme="light"]` 提供专属的 `#ffffff` 终端背景、`#f1f5f9` Banner 和 `#f8fafc` 侧边栏底色。

## 2. 任务分解
- **Task 1**: 后端排序优化与 Shell 执行引擎接口实现 (`POST /api/terminal/exec`) 及单元测试；
- **Task 2**: 前端执行引擎对接与默认终端首位排序保证；
- **Task 3**: 弹层向上展开定位与侧边栏加号/垃圾桶图标与间距优化；
- **Task 4**: 深色模式与浅色模式终端全套主题精细化配色；
- **Task 5**: 全量测试与构建验证回归。
