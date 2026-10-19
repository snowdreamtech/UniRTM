# 架构设计与实施计划：文件管理器、编辑器与源代码管理三位一体 (Plan)

## 1. 架构设计与状态总线 (Architecture)

### 1.1 核心理念：以「统一工作区状态总线」为粘合剂
三者之间的协同不是强耦合，而是基于轻量响应式事件与响应式 Store：
- **`useEditorTabs.ts` 升级**：扩充 Tab 模型，增加 `isDiffTab`、`diffType`、`originalPath`、`modifiedPath`，使普通编辑 Tab 与差异对比 Tab 完全统一；
- **`useWorkspaceBus.ts` 响应式总线**：
  - `revealInExplorer(path: string)`：通知文件树展开父目录、滚动定位并高亮；
  - `openInEditor(path: string, options?: { diff?: boolean; staged?: boolean })`：统一切换到编辑器打开或比对；
  - `gitStatusChanged`：广播文件系统或暂存状态更新，联动刷新 Explorer 树与 Tab 样式；
- **Tab 状态融合渲染器**：在 `EditorTabs.vue` 与 `ExplorerView.vue` 中统一引入 Git 状态（M, U, D）与 Dirty 状态的视觉优先级；
- **Explorer 右键菜单增强**：在 `FileContextMenu.vue` 中注入 Git Action 组，直接调用 `useGit` 的 `stageFile`、`unstageFile`、`discardFile`。

---

## 2. 详细组件改造矩阵

| 组件/文件 | 改造目标与职责 |
| :--- | :--- |
| `frontend/src/composables/useEditorTabs.ts` | 增加 Diff 标签页支持（`openDiffTab`、`diffInfo`）；提供根据 Git 变动状态自动装饰 Tab 样式的计算属性 |
| `frontend/src/composables/useWorkspaceBus.ts` | 创建轻量工作区协同事件总线，处理 `revealFile`、`openDiff`、`gitActionTriggered` 等事件 |
| `frontend/src/components/explorer/FileTreeNode.vue` 与 `FileContextMenu.vue` | 增强右键上下文菜单，注入 Git 操作组（暂存/放弃/差异比较）；强化目录与文件节点的 Git 状态徽标 |
| `frontend/src/views/ExplorerView.vue` | 监听 `revealFile` 事件，实现递归自动展开父级目录与平滑滚动动画；支持在主编辑区渲染 Diff 标签页（使用现有的 Monaco Diff 引擎） |
| `frontend/src/views/SourceControlView.vue` | 增加“在资源管理器中定位”按钮；点击变更直接代理调用 `useWorkspaceBus.openDiff` 在主编辑器中打开，保持双向连通 |
| `frontend/src/i18n/` | 同步新增上下文菜单与 Tab 提示的多语言文案（涵盖全部 53 种语言包） |

---

## 3. 分阶段实施策略 (Phased Implementation)

1. **阶段 1：统一工作区总线与 Diff 标签页模型 (Core Composable & Model)**
   - 扩充 `useEditorTabs.ts`，定义 `diffTab` 结构与 `openDiffTab` 方法；
   - 创建 `useWorkspaceBus.ts`，编写单元测试验证事件发布、订阅与 Tab 状态计算。
   - **分步提交 1**：`feat(workbench): add workspace bus and diff tab support to useEditorTabs`

2. **阶段 2：编辑器 Tab 视觉增强与 Diff 视图无缝渲染 (Editor Tab & Diff View)**
   - 在 `ExplorerView.vue` 与 `EditorTabs.vue` 中支持 Diff 模式的 Monaco 差异编辑器挂载；
   - 实现 Tab 标签的 Git 状态色（黄色 M、绿色 U）与未保存 `●` 视觉共存。
   - **分步提交 2**：`feat(editor): render diff tab in main editor and enhance tab git status decorators`

3. **阶段 3：文件管理器 ⟷ SCM 双向定位联动 (Bidirectional Reveal & Navigation)**
   - 实现 `revealInExplorer` 逻辑：自动深度递归展开父目录，计算滚动位置并添加平滑高亮动画；
   - 在 `SourceControlView.vue` 中集成“在文件管理器中定位”按钮与快捷方式。
   - **分步提交 3**：`feat(explorer): implement auto-reveal and highlight from source control`

4. **阶段 4：文件管理器右键 Git 上下文动作跨界融合 (Context Menu Cross-Pollination)**
   - 在 `FileContextMenu.vue` 中根据文件/目录的 Git 状态，动态挂载“暂存 (Stage)”、“取消暂存 (Unstage)”、“放弃更改 (Discard)”、“对比工作区修改 (Compare)”；
   - 联动 `useGit`，操作后自动触发 Explorer 与 SCM 的协同刷新。
   - **分步提交 4**：`feat(explorer): integrate git actions into file context menu`

5. **阶段 5：全 53 国多语言支持与质量验证 (i18n & Verification)**
   - 扩展多语言字典，通过脚本同步到所有 53 种语言包；
   - 跑通 Vitest 全量单元测试（包含 i18n 完整性校验）和生产环境构建 `npm run build`。
   - **分步提交 5**：`feat(i18n): sync multi-language keys and finalize workbench trinity verification`
