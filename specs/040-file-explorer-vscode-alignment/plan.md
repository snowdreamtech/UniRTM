# 技术实施计划：资源管理器 (File Explorer) 对标 VS Code (Plan)

## 1. 架构设计与数据流

```
[ ActivityBar ] ──(Click 'explorer')──► [ Router: /explorer ]
                                              │
                    ┌─────────────────────────┴─────────────────────────┐
                    ▼                                                   ▼
         [ Sidebar: ExplorerPanel ]                          [ ExplorerView (Main Canvas) ]
          - Workspace Title                                   - Active File Info / Empty State
          - Tool Actions (New, Refresh, Collapse)             - Detailed File Metadata Pane
          - FileTree (Recursive, Virtual/Lazy)                - Quick Actions (Open in Terminal)
                    │
                    ▼ (API Calls)
      ┌─────────────────────────────┐
      │  /api/fs/tree               │
      │  /api/fs/create             │
      │  /api/fs/rename             │
      │  /api/fs/delete             │
      └──────────────┬──────────────┘
                     │
       ┌─────────────┴─────────────┐
       ▼                           ▼
[ Go Backend: cmd/64.ui.go ]  [ Vite Middleware: vite.config.ts ]
  - os.ReadDir & filepath       - fs.readdirSync & path
  - Workdir Boundary Check      - Workdir Boundary Check
  - gitignore check via helper  - gitignore check
```

---

## 2. 核心模块与技术实现

### 2.1 导航与路由接入
- 在 `ActivityBar.vue` 中，将 `explorer` 插入在 `dashboard` 与 `tools` 之间。
- 在 `router/index.ts` 注册 `/explorer` 路由，关联 `ExplorerView.vue`。
- 在 `App.vue` 的 `currentTabLabel`、面包屑、活动项处理中增加 `explorer` 支持。
- 在 `i18n` 语言包中注册 `nav.explorer` 与 `explorer.*` 系列国际化翻译词条。

### 2.2 树形组件设计与 Seti UI 图标集成
- 封装 `FileTree.vue` 与 `FileTreeNode.vue`：
  - **缩进线机制**：使用 CSS 变量与多重 `padding-left` 绘制层级垂直参考线；
  - **Seti UI 图标匹配引擎**：复用项目中已有的 `iconThemeService.ts` / `@vscode/seti-ui` 图标映射：
    - 根据文件扩展名（`.go`, `.ts`, `.vue`, `.toml` 等）或者特殊全名（`.gitignore`, `Dockerfile`, `package.json`, `README.md`）渲染对应的矢量/字体图标；
    - 文件夹采用 `codicon-folder` 与 `codicon-folder-opened`；
  - **状态管理**：选中项 `selectedPath`、展开集 `expandedSet`（基于 Set 管理快速判断）、内联正在编辑/重命名路径 `editingPath`。

### 2.3 操作能力与上下文交互
- **内联新建/重命名**：
  - 双击或按 F2 时，文件名原地切换为 `<input>` 输入框；
  - 自动选中不包含后缀名的主文件名，按 Enter 提交 API，按 Esc 撤销。
- **右键上下文菜单 (`ContextMenu.vue`)**：
  - 全局捕获 `contextmenu` 事件，自适应屏幕边界渲染原生 VS Code 风格的右键菜单；
  - 支持删除二次模态框确认。

### 2.4 后端文件系统服务（Go & Vite 双端）
- **根目录范围锁定**：所有针对 `/api/fs/*` 的操作，严格限定在当前运行工作区目录（`rootDir`）内部。任何通过 `../` 越界的尝试均返回 `403 Forbidden`。
- **安全删除与创建**：提供原子性与幂等操作保障。

---

## 3. 分阶段实施策略
1. **阶段 1：底座与导航接入**
   - 导航栏图标注入与路由挂载；
   - 国际化支持；
   - Go 与 Vite 后端 `/api/fs/*` 接口实现与防越界单元测试。
2. **阶段 2：VS Code 风格树形组件与 Seti UI 渲染**
   - 文件树递归渲染与展开/折叠交互；
   - Seti 文件图标无缝对齐；
   - 工具栏 4 大快捷功能（新建文件/文件夹、刷新、全部折叠）。
3. **阶段 3：文件操作交互与上下文菜单**
   - 原地内联新建与重命名；
   - 删除确认与右键菜单；
   - “在集成终端中打开”联动。
4. **阶段 4：端到端验证与测试加固**
   - Vitest 前端单元测试；
   - Go 后端测试；
   - 生产打包 `npm run build`。
