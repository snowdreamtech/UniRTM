# 特性规范：基于 PTY、WebSocket 与 xterm.js 的工业级集成终端 (Spec)

## 1. 背景与目标
UniRTM 目前的集成终端采用 HTTP POST 单次执行模式，存在以下体验与架构短板：
1. **缺少真实交互与状态**：非交互式子进程，不支持交互式输入（如 vim, top, sudo 密码交互, git commit 编辑等）；
2. **缺少原生 Prompt ($PS1)**：无 `user@hostname:path$` 原生 Shell 提示符和动态目录跟踪；
3. **彩色与终端控制码缺失**：未解析 ANSI 转义序列，命令输出沦为黑白纯文本；
4. **快捷键与操作受限**：不支持 Tab 补全、`Ctrl+C` 中断、`Ctrl+R` 历史搜索等常用控制键。

**核心目标**：全面对齐 VS Code 架构，基于 **操作系统 PTY + WebSocket 长连接 + xterm.js 前端引擎**，打造全功能、真交互、高对比度、支持丰富主题与色彩的工业级 Web 终端。

---

## 2. 架构设计

### 2.1 整体架构图
```
[ Browser / UniRTM Web UI ]
       │  ▲
       │  │ (Keystrokes / Terminal Resize / Binary Stream)
       ▼  │
[ xterm.js + fit-addon ]
       │  ▲
       │  │ WebSocket (/api/terminal/ws?id=xxx&shell=xxx)
       ▼  │
[ Go Backend (unirtm ui) / Vite Dev Middleware ]
       │  ▲
       │  │ io.Copy (PTY Master Fd)
       ▼  │
[ creack/pty (POSIX PTY / Windows ConPTY) ]
       │  ▲
       ▼  │
[ Interactive Login Shell (bash/zsh/sh/pwsh with UNIRTM PATH injected) ]
```

---

## 3. 详细技术规范

### 3.1 环境变量与 PATH 注入 (PATH & Environment)
- 启动任何 PTY 会话时，子进程环境必须注入：
  - `PATH`: 必须前置 `~/.unirtm/shims:~/.unirtm/bin:$PATH`；
  - `TERM`: 设置为 `xterm-256color`，告知所有 CLI 工具启用 256 色/真彩色；
  - `COLORTERM`: 设置为 `truecolor`；
  - `UNIRTM_TERMINAL`: 设置为 `1`；
  - `LANG`: 默认继承宿主或 `C.UTF-8`，保证中文字符与符号正常显示。

### 3.2 WebSocket 协议规范 (`/api/terminal/ws`)
- **连接 URL**：`/api/terminal/ws?id={sessionOrPaneId}&shell={shellPath}&cwd={workdir}`
- **消息通道**：
  - **文本/二进制数据（stdin/stdout）**：
    - 前端 -> 后端：用户敲击键盘产生的数据（UTF-8 字符序列，包括回车、退格、方向键、控制键）；
    - 后端 -> 前端：Shell 及子进程产生的实时输出（包含 ANSI 转义码、彩色、光标移动指令）。
  - **控制信令（JSON 格式）**：
    - 尺寸调整：`{ "type": "resize", "cols": 120, "rows": 30 }`
    - 关闭退出：`{ "type": "kill" }`

### 3.3 后端 PTY 引擎 (Go Backend)
- 引入工业级 PTY 库：`github.com/creack/pty`；
- 使用标准 `golang.org/x/net/websocket` 处理长连接握手；
- 维护活跃 PTY 会话注册表（`sessionID -> *os.File, *exec.Cmd`）；
- 当 WebSocket 断开或收到关闭信号时，自动向进程组发送 SIGHUP/SIGKILL，回收 PTY 资源与子进程，杜绝僵尸进程。

### 3.4 前端 xterm.js 深度集成
- 引入 `@xterm/xterm` 与 `@xterm/addon-fit`；
- 封装独立、高复用的 Vue 终端组件；
- 适配已有的单终端、竖直拆分（Vertical Split）、横向拆分（Horizontal Split）多窗格会话；
- 监听窗格尺寸变化（ResizeObserver / 抽屉拉伸），自动调用 `fitAddon.fit()` 并向后端同步 `cols` 与 `rows`。

### 3.5 主题与色彩适配 (Theme & Aesthetics)
- **字体族**：优先使用 `'JetBrains Mono', 'Fira Code', 'Cascadia Code', Menlo, Monaco, Consolas, monospace`；
- **字号与行高**：字号 13px / 14px，行高 1.35，清晰饱满，杜绝发灰模糊；
- **深浅色主题动态切换**：
  - 深色模式：背景 `#18181b`，前景色 `#f1f5f9`，光标 `#38bdf8`，高对比度 16 色 ANSI 调色板；
  - 浅色模式：背景 `#f8fafc` / `#ffffff`，前景色 `#0f172a`（纯正深石板黑），光标 `#0284c7`，清晰醒目的深色彩色 ANSI 调色板。
