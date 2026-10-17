# 任务分解与执行清单 (Tasks)

- [ ] **Task 1: 后端依赖引入与 PTY WebSocket 会话管理器实现 (Backend PTY & WebSocket)**
  - 在 Go 项目中引入 `github.com/creack/pty`；
  - 创建 `cmd/64.ui_pty.go`，实现 PTY 会话管理器：
    - 支持按 sessionID / paneID 注册与管理 PTY 子进程；
    - 启动 Shell（如 `/bin/bash`、`/bin/zsh` 等）时注入关键环境变量（前置 `~/.unirtm/shims:~/.unirtm/bin` 到 PATH，注入 `TERM=xterm-256color`、`COLORTERM=truecolor` 等）；
    - 实现 WebSocket 连接处理器 `/api/terminal/ws`，支持全双工数据流（`io.Copy`）与 JSON 控制包（`resize`、`kill`）；
    - 支持优雅清理与连接中断进程回收；
  - 在 `frontend/vite.config.ts` 中同步提供开发环境的 WebSocket PTY 处理，确保在 `npm run dev` 模式下直接可用；
  - *提交命名*：`feat(terminal): implement backend PTY session manager and websocket streaming handler`

- [ ] **Task 2: 前端 @xterm/xterm 引入与终端组件封装 (Frontend xterm.js Pane Component)**
  - 在 `frontend` 中安装 `@xterm/xterm` 与 `@xterm/addon-fit`；
  - 创建 `frontend/src/components/terminal/XTerminalPane.vue`：
    - 挂载 xterm 容器，初始化 `Terminal` 实例并加载 `FitAddon`；
    - 连接 `/api/terminal/ws` WebSocket 并处理数据上下行；
    - 接入 ResizeObserver 自动自适应尺寸并发送 `resize` 控制帧；
    - 支持断开重连交互与状态提示；
  - *提交命名*：`feat(terminal): introduce xterm.js with fit addon and websocket terminal pane component`

- [ ] **Task 3: 终端主题系统（深浅色 16 色 ANSI 调色板）精细化适配 (Theme & Palette Alignment)**
  - 封装终端深浅色主题配置映射函数：
    - 适配浅色模式：白色/浅灰背景，纯正深石板色高对比度前景色（`#0f172a`），清晰饱和的 ANSI 16 色调色板；
    - 适配深色模式：现代暗色背景（`#18181b` / `#0f172a`），明亮代码前景色（`#f1f5f9`），鲜明 ANSI 16 色；
    - 精选等宽字体族（`JetBrains Mono`, `Fira Code`, `Consolas`, monospace）与行高排版；
  - 响应全局主题切换（Dark / Light / Follow System）即时热更新终端配色；
  - *提交命名*：`style(terminal): craft high-contrast 16-color ANSI theme and dynamic theme switching for xterm`

- [ ] **Task 4: BottomDrawer 抽屉多窗格整合与多终端生命周期联动 (Drawer Multi-Pane Integration)**
  - 在 `frontend/src/components/layout/BottomDrawer.vue` 中：
    - 将旧版模拟命令行 DOM 替换为真实的 `XTerminalPane`；
    - 保持多终端标签列表、竖直拆分（1 分为 2）、横向拆分（1 分为 2）正常工作；
    - 切换标签时保持终端会话不中断，销毁窗格时向后端发送销毁信令；
  - *提交命名*：`feat(drawer): integrate xterm pane with multi-tab and split-pane session lifecycle`

- [ ] **Task 5: 全量测试、自动化回归与构建打包 (Validation & Tests)**
  - 运行 Go 终端单元测试；
  - 运行前端 Vitest 单元测试；
  - 运行 `npm run build`（`vue-tsc` 与生产打包）；
  - 验证终端在浅色与深色模式下的真实渲染效果、Prompt 颜色、`unirtm version` 彩色 Banner 与常用命令交互；
  - *提交命名*：`chore(terminal): complete end-to-end validation for PTY websocket terminal integration`
