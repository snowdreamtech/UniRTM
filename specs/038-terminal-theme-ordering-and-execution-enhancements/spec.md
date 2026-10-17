# 特性规范：集成终端视觉主题适配、默认排序、弹层遮挡修复与 Shell 执行引擎 (Spec)

## 1. 背景与目标
根据用户在真实浅色模式与复杂命令运行场景下的体验反馈，对集成终端模块进行全面体验重构与核心能力补全：
1. **全模式主题适配 (Dark & Light Themes)**：解决浅色模式下终端背景、窗格 Banner 与右侧 Sidebar 出现大面积深黑灰色块的问题，提供原生精致的深浅主题调色板；
2. **系统默认终端置顶排序**：在任何终端选择列表与新建菜单中，系统默认终端（`is_default: true`）必须强制置顶排在最首位；
3. **下拉菜单遮挡修复**：新建终端下拉列表过长时会被抽屉底栏裁剪，需调整弹出方向为向上展开（Popup upwards）并增加滚动条；
4. **侧边栏按钮交互与布局优化**：增大右侧终端侧边栏加号按钮（加大、呼吸边距），将关闭按钮升级为更大、更易点击的垃圾桶图标（Trash Icon）并保证充足的右间距；
5. **真正的 Shell 命令行执行引擎与环境变量注入**：解决用户在终端输入普通 Linux 命令（如 `ls`, `git`, `pwd`, `unirtm version`）被错误转发至 tasks API 的问题，在后端实现真正的 `/api/terminal/exec` 接口，自动注入系统 PATH 与 UniRTM shims，原生执行 Shell 命令行。

## 2. 详细需求与改动范围

### 2.1 主题配色适配 (Themes Palette)
- 适配 CSS 变量：
  - 深色模式（Dark）：保持高品质极客黑底色，窗格边缘以细线分割，Banner 与侧边栏具有清晰的半透明层次；
  - 浅色模式（Light）：背景切换为纯白或浅底色，Banner 使用柔和的灰色（`#f1f5f9`），侧边栏使用浅灰底（`#f8fafc`），边框统一使用 `var(--card-border)`，提示文字使用中性深灰，彻底消除突兀的脏灰黑斑块。
- 窗格激活状态：使用 `var(--accent-cyan)` 或系统强调色聚焦边框，失焦窗格柔和融入背景。

### 2.2 终端检测结果置顶排序 (Terminal Ordering)
- 后端 Go `detectInstalledTerminals()`：在返回终端列表前，对 `is_default: true` 的终端进行优先置顶排序；
- 前端 `useTerminal.ts` 与 `useTerminalSessions.ts`：保证在任何下拉与选项中，默认终端始终位于列表第 1 位。

### 2.3 下拉菜单弹出定位 (Dropdown Positioning)
- `.term-new-dropdown` 改为 `bottom: calc(100% + 6px); top: auto;`；
- 限制最大高度 `max-height: 260px; overflow-y: auto;`；
- 增加阴影与圆角，确保在任何视窗高度下均能向上完整展现。

### 2.4 侧边栏按钮形态与间距 (Sidebar Ergonomics)
- 加宽侧边栏至 185px~195px；
- 加号按钮设为带 hover 背景的独立按纽（尺寸约 22x22px，距右侧 8px）；
- 标签项关闭按纽替换为大尺寸垃圾桶图标（Trash Icon，14px），鼠标悬停显示删除反馈，距右侧 6px。

### 2.5 后端 Shell 命令执行接口与前端对接 (Terminal Command Execution Engine)
- 后端新增 `POST /api/terminal/exec`：
  - 参数：`command` (命令字符串), `shell` (执行的 shell 路径), `cwd` (执行目录);
  - 环境变量：完整继承宿主 `os.Environ()`，并在 PATH 最前注入 UniRTM 安装目录（`~/.unirtm/shims`, `~/.unirtm/bin` 等）；
  - 执行方式：以真实 Shell（如 `/bin/bash -c "cmd"`）运行并收集标准输出/标准错误；
- 前端开发模式 `vite.config.ts`：同步支持 `/api/terminal/exec` Node.js child_process 本地模拟；
- `BottomDrawer.vue`：终端命令行输入改调 `/api/terminal/exec`，实现真正的全功能 Linux 交互。
