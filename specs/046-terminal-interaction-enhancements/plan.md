# 架构设计与实施计划：集成终端深度交互与工作区协同增强系统 (Plan)

## 1. 架构总览与设计目标

为了让集成终端具备与专业现代 IDE（如 VS Code）一致的高效交互体验，本架构在现有 `XTerminalPane.vue` 和 `useTerminalTabs.ts` 基础上，构建一套松耦合、高响应、模块化的交互增强套件。

```mermaid
graph TD
    subgraph Integrated Terminal
        XP[XTerminalPane.vue]
        TL[useTerminalLinks.ts]
        SW[TerminalSearchWidget.vue]
        TCM[TerminalContextMenu.vue]
        AS[@xterm/addon-search]
        AWL[@xterm/addon-web-links]
    end

    subgraph Workspace Coordination
        WB[useWorkspaceBus.ts]
    end

    subgraph Workbench Ecosystem
        ME[MonacoEditor.vue<br/>Shift+Enter Dispatcher]
        FT[FileTreeNode.vue / useFileTree.ts<br/>Open in Integrated Terminal]
        ED[EditorTabStore / Workspace]
    end

    XP --> TL
    XP --> SW
    XP --> TCM
    TL --> AS
    TL --> AWL

    TL -- "openInEditor({ path, line, col })" --> WB
    WB --> ED

    ME -- "sendToTerminal({ text })" --> WB
    FT -- "revealInTerminal({ path })" --> WB
    WB --> XP
```

---

## 2. 关键技术选型与模块拆解

### 2.1 依赖项扩展
- **`@xterm/addon-search`**：官方高性能搜索插件，支持在 xterm 离屏与主缓冲区中快速执行正则、大小写与全字匹配，并支持匹配高亮及双向导航。
- **`@xterm/addon-web-links`**：官方 Web 链接识别与拦截分发。

### 2.2 核心模块设计

#### 模块 A：`useTerminalLinks.ts` (超链接与文件堆栈导航器)
- **职责**：
  - 注册 `ILinkProvider` 扫描终端视口字符流；
  - 识别规则覆盖：
    - Web URL：`https?://[^\s/$.?#].[^\s]*`
    - 文件路径与行列：`(?:\/|[a-zA-Z0-9_\-\.\/]+)\.[a-zA-Z0-9]+(?::\d+)?(?::\d+)?`
    - 常见堆栈模式：`at ... (path:line:col)`、`--> path:line:col`、`File "path", line X`
  - 点击交互：
    - 若为 Web URL：调用 `window.open(url, '_blank')`；
    - 若为文件路径：提取路径、行号与列号，调用 `workspaceBus.emit('openInEditor', { filePath, line, column })`，若文件树中为相对路径，自动基于当前工作区目录规范化。

#### 模块 B：`TerminalSearchWidget.vue` (终端内置浮动搜索条)
- **职责**：
  - 作为浮动挂件渲染在 `XTerminalPane.vue` 的右上角，UI 风格严格与 VS Code 浮动搜索条对齐；
  - 绑定快捷键 `Ctrl+F` (或 `Cmd+F`) 聚焦搜索框；
  - 核心状态：
    - `searchText`: 搜索关键字
    - `caseSensitive`: 是否区分大小写
    - `wholeWord`: 是否全字匹配
    - `regex`: 是否启用正则表达式
  - 与 `@xterm/addon-search` 的 `SearchAddon` 实例绑定，支持 `findNext()` 与 `findPrevious()`；
  - 支持 `Enter` 向下查找、`Shift+Enter` 向上查找、`Esc` 退出并恢复终端焦点。

#### 模块 C：`TerminalContextMenu.vue` (右键上下文菜单)
- **职责**：
  - 拦截终端容器的 `contextmenu` 事件；
  - 提供核心操作：
    - 复制 (`copy`)：如果有选区文字，调用 `navigator.clipboard.writeText`；
    - 粘贴 (`paste`)：读取系统剪贴板并向 PTY 发送输入数据；
    - 全选 (`selectAll`)：调用 `terminal.selectAll()`；
    - 清屏 (`clear`)：调用 `terminal.clear()`；
    - 查找 (`find`)：唤出搜索框；
    - 拆分 (`split`)：向终端管理总线分发拆分命令；
  - 支持键盘 `Esc` 或任意外部点击自动收起。

#### 模块 D：`useWorkspaceBus` 事件总线通信升级
- 增加通道：
  - `sendToTerminal`: 接收 `{ text: string, addNewline?: boolean }`，由当前活动的终端实例接收并写入 PTY；
  - `openInIntegratedTerminal`: 接收 `{ directoryPath: string }`，向活动终端发送 `cd "<directoryPath>"\n`，并展开底部面板切到终端 Tab。

#### 模块 E：Monaco 编辑器与文件管理器联动
- **MonacoEditor.vue**：
  - 注册快捷键 `Shift+Enter` 与命令 `editor.action.sendSelectionToTerminal`；
  - 获取选中文字（若无选中则获取光标所在当前行），通过 `workspaceBus` 发送至终端，光标自动下移一行；
  - 右键上下文菜单新增「在活动终端中运行代码 (Run in Active Terminal)」。
- **FileTreeNode.vue**：
  - 当节点类型为 `directory` 时，右键菜单中新增「在集成终端中打开 (Open in Integrated Terminal)」。

---

## 3. 分阶段实施策略 (Phased Implementation Strategy)

为了保证开发的高内聚、低耦合与稳健交付，整体实施划分为 4 个连续递进的 Phase，并在每个 Phase 结束前完成针对性测试与独立 Git 提交：

- **Phase 0：Spec Kit 规划与文档初始化 (Docs Baseline)**
  - 创建规范、计划与任务清单；
  - 初始化 Git 分支并提交基线。

- **Phase 1：依赖升级、超链接识别与堆栈穿透跳转 (Hyperlinks & Navigation)**
  - 引入 `@xterm/addon-search` 和 `@xterm/addon-web-links`；
  - 实现 `useTerminalLinks.ts`，支持 Web URL 打开及文件堆栈跳转主编辑器；
  - 编写单测验证正则解析与跳转事件。

- **Phase 2：终端浮动搜索工具条 (Terminal In-Panel Search)**
  - 实现 `TerminalSearchWidget.vue` 浮动组件；
  - 对接 `SearchAddon` 并适配快捷键体系；
  - 编写单测验证搜索状态切换与查找触发。

- **Phase 3：终端右键上下文菜单与编辑器/文件树双向互通 (Context Menu & Bidirectional Flow)**
  - 实现 `TerminalContextMenu.vue`；
  - 编辑器支持 `Shift+Enter` 直发终端及右键菜单项；
  - 文件管理器目录右键提供「在集成终端中打开」；
  - 完善单元测试。

- **Phase 4：53 种语言国际化字典适配与全量集成验证 (i18n & Production Release)**
  - 同步所有新增文案至 53 种多语言文件；
  - 运行全量单元测试与类型检查；
  - 运行生产环境打包并打齐全部任务钩选。
