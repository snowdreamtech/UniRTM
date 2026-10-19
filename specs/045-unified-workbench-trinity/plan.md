# 架构设计与实施计划：文件管理器、编辑器与源代码管理三位一体 (Plan)

## 1. 架构设计与状态总线 (Architecture)

### 1.1 核心理念：以「统一工作区状态总线」为粘合剂
三者之间的协同不是强耦合，而是基于轻量响应式事件与响应式 Store：
- **`useEditorTabs.ts` 升级**：扩充 Tab 模型，增加 `isDiffTab`、`diffType`、`originalPath`、`modifiedPath`、`originalContent`，使普通编辑 Tab 与差异对比 Tab 完全统一；
- **`useWorkspaceBus.ts` 响应式总线**：
  - `revealInExplorer(path: string)`：通知文件树展开父目录、滚动定位并高亮；
  - `openInEditor(path: string, options?: { diff?: boolean; staged?: boolean })`：统一切换到编辑器打开或比对；
  - `gitStatusChanged`：广播文件系统或暂存状态更新，联动刷新 Explorer 树与 Tab 样式；
  - `fileRenamed(oldPath, newPath)` / `fileDeleted(path)`：广播文件生命周期变动，自动同步迁移编辑器 Tab 路径与模型；
- **行级微观感知与差异操作引擎 (`diffGutter.ts`)**：
  - 基于当前内存缓冲与 Git 基准文本实时计算增量差异；
  - Monaco Editor 挂载行号槽指示条（Gutter Diff），支持点击展开 Inline Diff Peek 浮层；
  - 支持单块一键回滚 (`revertActiveDiffRange`) 与单块一键暂存 (`stageActiveDiffRange`)；
- **文件树智能排查与聚合引擎 (`useFileTree.ts`)**：
  - 接入 Git 聚焦过滤算法（Git Focus Mode），只保留有修改文件的祖先链；
  - 接入折叠目录子孙变更聚合算法（Folder Bubble Count），折叠父目录展示内部变更总数；
- **底栏与全局 SCM 联动**：
  - 编辑器状态栏展示当前 Git 分支与工作区修改统计，点击支持穿透定位。

---

## 2. 详细组件改造矩阵

| 组件/文件 | 改造目标与职责 |
| :--- | :--- |
| `frontend/src/composables/useEditorTabs.ts` | 增加 Diff 标签页支持（`openDiffTab`、`diffInfo`）；文件重命名/删除事件响应与 Tab 路径迁移；Tab 脏状态与 Git 生命周期同步 |
| `frontend/src/composables/useWorkspaceBus.ts` | 创建轻量工作区协同事件总线，处理 `revealFile`、`openDiff`、`gitActionTriggered`、`fileRenamed`、`fileDeleted` 等事件 |
| `frontend/src/composables/useFileTree.ts` | 实现 Git Focus 过滤裁剪算法；实现折叠目录的递归 Git 变动计数汇总逻辑 |
| `frontend/src/composables/useGit.ts` | 增加单块局部暂存支持与状态同步刷新；提供安全提交/丢弃检查 |
| `frontend/src/components/editor/MonacoEditor.vue` | 渲染实时 Git Gutter 增量条；提供 Inline Diff Peek 浮层，支持一键单块回退与一键单块暂存 |
| `frontend/src/components/explorer/FileTreeNode.vue` 与 `FileContextMenu.vue` | 增强右键上下文菜单，注入 Git 操作组（暂存/放弃/差异比较）；强化目录与文件节点的 Git 状态徽标与折叠气泡 |
| `frontend/src/views/ExplorerView.vue` | 监听 `revealFile` 事件，实现递归自动展开父级目录与平滑滚动动画；状态栏展示当前分支与变动统计，点击穿透至 SCM |
| `frontend/src/views/SourceControlView.vue` | 增加“在资源管理器中定位”按钮；点击变更直接代理调用 `useWorkspaceBus.openDiff` 在主编辑器中打开，保持双向连通 |
| `frontend/src/i18n/` | 同步新增上下文菜单、Tab 提示、行内 Peek 操作的多语言文案（涵盖全部 53 种语言包） |

---

## 3. 分阶段实施策略 (Phased Implementation)

1. **阶段 1：统一工作区总线与 Diff 标签页模型 (Core Composable & Model)** [已完成]
   - 扩充 `useEditorTabs.ts`，定义 `diffTab` 结构与 `openDiffTab` 方法；
   - 创建 `useWorkspaceBus.ts`，编写单元测试验证事件发布、订阅与 Tab 状态计算。
   - **分步提交 1**：`feat(workbench): add workspace bus and diff tab support to useEditorTabs`

2. **阶段 2：编辑器 Tab 视觉增强与 Diff 视图无缝渲染 (Editor Tab & Diff View)** [已完成]
   - 在 `ExplorerView.vue` 与 `EditorTabs.vue` 中支持 Diff 模式的 Monaco 差异编辑器挂载；
   - 实现 Tab 标签的 Git 状态色（黄色 M、绿色 U）与未保存 `●` 视觉共存。
   - **分步提交 2**：`feat(editor): render diff tab in main editor and enhance tab git status decorators`

3. **阶段 3：文件管理器 ⟷ SCM 双向定位联动 (Bidirectional Reveal & Navigation)** [已完成]
   - 实现 `revealInExplorer` 逻辑：自动深度递归展开父目录，计算滚动位置并添加平滑高亮动画；
   - 在 `SourceControlView.vue` 中集成“在文件管理器中定位”按钮与快捷方式。
   - **分步提交 3**：`feat(explorer): implement auto-reveal and highlight from source control`

4. **阶段 4：文件管理器右键 Git 上下文动作跨界融合 (Context Menu Cross-Pollination)** [已完成]
   - 在 `FileContextMenu.vue` 中根据文件/目录的 Git 状态，动态挂载“暂存 (Stage)”、“取消暂存 (Unstage)”、“放弃更改 (Discard)”、“对比工作区修改 (Compare)”；
   - 联动 `useGit`，操作后自动触发 Explorer 与 SCM 的协同刷新。
   - **分步提交 4**：`feat(explorer): integrate git actions into file context menu`

5. **阶段 5：全 53 国多语言支持与质量验证 (i18n & Verification)** [已完成]
   - 扩展多语言字典，通过脚本同步到所有 53 种语言包；跑通 Vitest 单元测试和构建。
   - **分步提交 5**：`feat(i18n): sync multi-language keys and finalize workbench trinity verification`

6. **阶段 6：文件系统与生命周期状态深度同步 (Lifecycle Sync)** [已完成]
   - 重命名和删除在总线广播并平滑迁移 Editor Tab；未保存状态与 Git 操作安全保护。
   - **分步提交 6**：`feat(workbench): synchronize editor dirty state with git operations` & `feat(workbench): propagate file rename and delete events`

7. **阶段 7：编辑器实时 Git Gutter 差异指示与行内 Diff 交互 (Git Gutter & Inline Diff)** [已完成]
   - Monaco 行号槽实时渲染绿/蓝/红增量条，支持点击展开 Inline Diff Peek 与单块回滚。
   - **分步提交 7**：`feat(editor): render real-time git diff gutter indicators` & `feat(editor): support inline diff peek and single block revert`

8. **阶段 8：文件树 Git 聚焦排查模式与 Gitignore 识别 (Git Focus Mode)** [已完成]
   - 文件树顶栏一键切换 Git 聚焦视图，仅展示变动链路；忽略文件精细化灰显。
   - **分步提交 8**：`feat(explorer): add Git Focus mode to filter and prune file tree to changed branches only`

9. **阶段 9：底部状态栏 Git 联动与多维穿透导航 (Status Bar Synergy)** [已完成]
   - 底部状态栏展示当前分支名称与工作区修改计数，点击直达 SCM 视图。
   - **分步提交 9**：`feat(workbench): connect editor status bar with active branch and file Git status navigation`

10. **阶段 10：Monaco 编辑器右键菜单主题与本地化修复 (Context Menu Refinement)** [已完成]
    - 修复 Shadow DOM 样式对比度并对齐 53 种语言本地化标签。
    - **分步提交 10**：`style(editor): ensure outline and border none on context menu action items`

11. **阶段 11：行内代码块一键暂存与部分提交 (Stage Chunk / Partial Staging)** [待执行]
    - 在 Inline Diff Peek 浮层中增加「暂存此块 (Stage Chunk)」动作；
    - 联动 `useGit` 完成差异代码块暂存并即刻刷新 Gutter 与 SCM 状态；编写单元测试。
    - **分步提交 11**：`feat(editor): support staging single diff chunk from inline peek widget`

12. **阶段 12：文件树折叠目录变更聚合徽标 (Folder Git Bubble Badges)** [待执行]
    - 在 `useFileTree` 中增加折叠父级目录的递归变动总数统计算法；
    - 在 `FileTreeNode.vue` 中渲染微型气泡徽标（Bubble Count）；编写单元测试。
    - **分步提交 12**：`feat(explorer): display aggregated git change badges on collapsed directory nodes`

13. **阶段 13：全量 53 国多语言同步与端到端回归验证 (Final Verification)** [待执行]
    - 补齐新增词条，跑通 `npm run test` 与 `npm run build`，更新任务清单。
    - **分步提交 13**：`feat(i18n): sync multi-language keys and finalize phase 11-12 validation`

