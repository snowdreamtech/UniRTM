# 特性实施计划：真实 PTY 会话、WebSocket 与 xterm.js 改造 (Plan)

## 1. 架构模块划分

### 1.1 后端服务层 (Go Backend & Vite Dev Middleware)
- **依赖库引入**：
  - Go：引入 `github.com/creack/pty`，使用 `golang.org/x/net/websocket` 处理长连接；
- **PTY 会话管理器 (`cmd/64.ui_pty.go`)**：
  - 会话生命周期创建、读取、写入、调整尺寸（Winsize）与销毁；
  - 启动 Shell 时注入完整环境变量（包括前置 `PATH=~/.unirtm/shims:~/.unirtm/bin:$PATH`、`TERM=xterm-256color` 等）；
  - 在 `cmd/64.ui.go` 注册 `/api/terminal/ws` WebSocket 处理器；
- **Vite 本地开发服务器支持 (`frontend/vite.config.ts`)**：
  - 增加对 `/api/terminal/ws` 的支持（使用 Node.js 原生 WebSocket 或 child_process/pty 桥接，或代理转发至 Go 后端）；

### 1.2 前端渲染与交互层 (`frontend`)
- **依赖库引入**：
  - 安装 `@xterm/xterm` 与 `@xterm/addon-fit`；
- **终端组件封装 (`frontend/src/components/terminal/XTerminalPane.vue`)**：
  - 挂载 xterm.js DOM 容器，初始化 `Terminal` 实例与 `FitAddon`；
  - 建立与 `/api/terminal/ws` 的 WebSocket 通信管道，实现按键上行和控制台输出下行；
  - 响应 ResizeObserver，实时自适应窗格大小并发送 resize 控制包；
  - 监听全局深色/浅色模式切换事件，热更新 xterm 主题样式（`theme.background`, `theme.foreground`, 16 种 ANSI 颜色）；
- **抽屉布局整合 (`frontend/src/components/layout/BottomDrawer.vue`)**：
  - 将旧版文本行与 input 替换为 `XTerminalPane`；
  - 保证多终端标签切换、横竖窗格拆分生命周期与 WebSocket 良好协同（切换标签不丢会话、关闭窗格通知后端释放 PTY）。

---

## 2. 关键路径与依赖
1. **Go 后端 PTY 库下载与验证**：确保在 Linux、macOS 以及 Windows 下编译无阻；
2. **前端 xterm 样式与字体加载**：保证等宽字体清晰无伪影，光标平滑闪烁；
3. **断网重连与异常熔断**：网络抖动或服务端重启时友好提示并提供重连按钮。

---

## 3. 任务分解
- **Task 1: 后端依赖引入与 PTY WebSocket 会话管理器实现**；
- **Task 2: 前端 @xterm/xterm 依赖安装与 XTerminalPane 组件开发**；
- **Task 3: 终端主题系统（深浅色 16 色 ANSI 调色板）精细化适配**；
- **Task 4: BottomDrawer 抽屉多窗格整合与多终端生命周期联动**；
- **Task 5: 全量测试、自动化回归与构建打包**。
