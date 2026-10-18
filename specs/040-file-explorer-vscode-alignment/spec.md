# 特性规范：对标 VS Code 的工业级资源管理器 (File Explorer) (Spec)

## 1. 背景与目标

UniRTM 正在从通用运行时与工具管理器逐步演进为轻量级、AI 原生的现代化 Web 开发工作台（Web Studio）。
一个专业工作台的基石是**文件与目录的浏览、组织与管理能力**。

**核心目标**：
在 Activity Bar 的 `dashboard` 与 `tools` 之间增加「资源管理器 (Explorer)」导航图标，并在界面与功能上**全面对标 Visual Studio Code 的资源管理器**：
1. **VS Code 视觉语言与交互对标**：包含可折叠根目录面板、标准四键操作工具栏（新建文件、新建文件夹、刷新、全部折叠）、层级嵌套缩进线；
2. **原生 `@vscode/seti-ui` 图标引擎集成**：深度利用项目中已内置合规的 Seti UI 图标体系，为每个文件根据扩展名及文件名（`.go`, `.ts`, `.vue`, `.toml`, `.json`, `Dockerfile`, `.gitignore` 等）赋予精准的文件类型彩色图标；
3. **完整的文件与目录 CRUD 操作**：
   - 树中原地内联新建（输入名称回车确认，Esc 取消）
   - 原地重命名（F2 / 菜单触发）
   - 删除（带安全确认对话框）
   - 复制绝对路径与相对路径
   - 在集成终端中快速打开选中目录
4. **智能过滤与忽略规则**：遵循 `.gitignore` 规范置灰显示忽略项，支持隐藏文件切换；
5. **双端无缝 API 支撑**：在 Go 后端（`cmd/64.ui.go`）与 Vite 前端开发中间件（`vite.config.ts`）中同时实现高性能本地文件系统 RESTful API，保证生产与开发环境 100% 一致与极速响应。

---

## 2. 界面布局与交互规范

### 2.1 导航栏位置（Activity Bar）
- **顺序**：位于 `Dashboard`（仪表盘）与 `Toolchains`（工具链）之间。
- **图标**：`workbench:explorer`（对应 VS Code `codicon-files`）。
- **路由**：`/explorer`，名称 `explorer`。

### 2.2 资源管理器侧边栏结构（Primary Sidebar & Main View）
1. **标题栏（Section Header）**：
   - 标题：“资源管理器 (EXPLORER)”
   - 顶部工具栏（悬浮显示 / 常驻工具栏）：
     - `codicon-new-file`：在当前选中目录或根目录新建文件
     - `codicon-new-folder`：在当前选中目录或根目录新建文件夹
     - `codicon-refresh`：刷新当前目录树
     - `codicon-collapse-all`：折叠所有已展开的目录
2. **工作区根节点（Root Section）**：
   - 标头显示当前工作区名称（如 `UNIRTM`），支持整体折叠/展开；
   - 支持空状态指引（若工作区为空，展示添加文件按钮）。
3. **树形节点交互规范（Tree Item）**：
   - **文件夹节点**：
     - 左侧折叠箭头（展开为 `codicon-chevron-down`，折叠为 `codicon-chevron-right`）；
     - 文件夹图标（折叠为 `codicon-folder`，展开为 `codicon-folder-opened`）；
     - 点击展开/收起目录，支持递归异步按需展开或首屏预加载深度。
   - **文件节点**：
     - 左侧显示对应语言/类型的 Seti 图标（如 Gopher 图标对应 `.go`、Vue 图标对应 `.vue`、TS 蓝标对应 `.ts`、齿轮对应 `.toml` 等）；
     - 文件名支持高亮搜索过滤；
     - 被 `.gitignore` 忽略的文件颜色变暗（置灰）。
4. **右键上下文菜单（Context Menu）**：
   - 在任意节点右键呼出 VS Code 风格的上下文菜单：
     - 📄 新建文件 (`New File`)
     - 📁 新建文件夹 (`New Folder`)
     - ✏️ 重命名 (`Rename` /快捷键 `F2`)
     - 🗑️ 删除 (`Delete` /快捷键 `Delete`)
     - 📋 复制相对路径 (`Copy Relative Path`)
     - 📋 复制绝对路径 (`Copy Path`)
     - 💻 在终端中打开 (`Open in Integrated Terminal`)

---

## 3. 后端 API 规范

在生产环境 Go 服务（`cmd/64.ui.go`）与开发环境 Vite 中间件（`vite.config.ts`）提供如下端点：

### 3.1 获取文件树
- **端点**：`GET /api/fs/tree?path={relativePath}&depth={number}&showHidden={bool}`
- **响应**：
```json
{
  "root": "/home/user/workspace/project",
  "name": "project",
  "entries": [
    {
      "name": "src",
      "path": "src",
      "is_dir": true,
      "size": 4096,
      "mod_time": "2026-10-18T08:00:00Z",
      "is_ignored": false,
      "children": [...]
    },
    {
      "name": "main.go",
      "path": "main.go",
      "is_dir": false,
      "size": 1024,
      "mod_time": "2026-10-18T08:05:00Z",
      "is_ignored": false
    }
  ]
}
```

### 3.2 创建文件或目录
- **端点**：`POST /api/fs/create`
- **载荷**：`{ "path": "src/utils/new_file.ts", "is_dir": false }`
- **安全约束**：严格防路径穿透（Path Traversal 防御，禁止逃逸出当前工作区）。

### 3.3 重命名
- **端点**：`POST /api/fs/rename`
- **载荷**：`{ "old_path": "old.txt", "new_path": "new.txt" }`

### 3.4 删除
- **端点**：`POST /api/fs/delete`
- **载荷**：`{ "path": "to_delete.txt", "recursive": true }`

### 3.5 读取与写入（为后续编辑器及快速预览保留）
- `GET /api/fs/read?path={path}`
- `POST /api/fs/write`：`{ "path": "path", "content": "..." }`

---

## 4. 非功能性指标
1. **极速响应**：千级文件目录下加载与展开耗时不超过 50ms；
2. **纯本地离线**：不依赖任何外网资源，所有图标与样式均为本地打包；
3. **防穿透安全**：所有路径计算强制基于 `filepath.Clean` 与 `strings.HasPrefix` 校验工作区根目录。
