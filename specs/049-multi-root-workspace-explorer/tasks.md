# 任务拆分与执行跟踪清单：多根工作区与文件浏览器管理系统 (Tasks)

## 任务执行进度概览
- **当前阶段**：Phase 0 - 方案设计与基线建立
- **总体进度**：0 / 15 完成 (0%)

---

## Phase 0: 方案设计与基线建立 (Docs Baseline)
- [ ] Task 0.1: 创建 `specs/049-multi-root-workspace-explorer/` 规范、方案与任务文档
- [ ] Task 0.2: 检出特性分支 `049-multi-root-workspace-explorer` 并提交文档基线

---

## Phase 1: 多根数据模型与持久化 API (Data Model & Backend API)
- [ ] Task 1.1: 扩展 `useFileTree` 支持多根数据模型 `workspaceRoots: Ref<FSNode[]>` 及判定函数 `isWorkspaceRoot(node)`
- [ ] Task 1.2: 在 Go 后端 (`cmd/`) 与前端开发服务器 (`vite.config.ts`) 实现 `/api/workspace/roots` 读取、添加与移除接口
- [ ] Task 1.3: 编写多根工作区管理核心单元测试
- [ ] Task 1.4: 阶段性 Git 提交：`feat(workspace): implement multi-root workspace data model and API`

---

## Phase 2: 右键菜单扩展与添加/删除交互 (Context Menu & Modals)
- [ ] Task 2.1: 在 `FileContextMenu.vue` 中集成“将文件夹添加到工作区...”与“将文件夹从工作区删除”菜单项
- [ ] Task 2.2: 实现添加文件夹到工作区弹窗 `AddFolderModal.vue`（支持自定义路径输入与快捷子文件夹选择）
- [ ] Task 2.3: 实现从工作区移除文件夹的二次确认与防误删提示
- [ ] Task 2.4: 阶段性 Git 提交：`feat(explorer): add workspace folder actions to context menu and modal`

---

## Phase 3: 资源管理器多根视图渲染与适配 (Multi-root Tree Rendering)
- [ ] Task 3.1: 改造 `ExplorerSidebar.vue` 顶栏与树容器，支持自适应单根/多根视图呈现（`工作区 (N 个文件夹)`）
- [ ] Task 3.2: 改造各根节点的独立折叠控制、新建文件/文件夹与快捷刷新
- [ ] Task 3.3: 阶段性 Git 提交：`feat(explorer): render multi-root workspace tree with independent folder sections`

---

## Phase 4: 53 种国际化本地化与单测覆盖 (i18n & Localization)
- [ ] Task 4.1: 在 53 种语言包中同步新增工作区右键操作文案（`explorer.add_folder_to_workspace`, `explorer.remove_folder_from_workspace`, 等）
- [ ] Task 4.2: 运行国际化完整性单测确保 53 种语言完全一致
- [ ] Task 4.3: 阶段性 Git 提交：`feat(i18n): complete 53 locales for multi-root workspace actions`

---

## Phase 5: 全量回归与生产封箱 (Regression & Release)
- [ ] Task 5.1: 运行全量单元测试（`npm run test`）全部绿灯通过
- [ ] Task 5.2: 运行类型检查与生产构建（`npm run build`）确保 0 警告 0 报错
- [ ] Task 5.3: 阶段性 Git 提交：`feat(workspace): finalize multi-root workspace explorer support`
