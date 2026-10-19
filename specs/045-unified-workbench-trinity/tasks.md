# 任务清单：文件管理器、编辑器与源代码管理三位一体协同系统 (Tasks)

## 阶段 1：统一工作区总线与 Diff 标签页模型 (Core Composable & Model) [已完成]
- [x] 1.1 在 `frontend/src/composables/useEditorTabs.ts` 中增强 `EditorTab` 模型，支持 `isDiff`、`diffOriginalPath`、`diffOriginalContent`、`diffModifiedContent` 等属性
- [x] 1.2 在 `useEditorTabs.ts` 中实现 `openDiffTab(fileDiff: DiffTabOptions)` 与关闭/定位逻辑
- [x] 1.3 创建 `frontend/src/composables/useWorkspaceBus.ts` 轻量响应式协同总线（支持 `revealInExplorer`、`openInEditor`、`activeFileChanged` 等）
- [x] 1.4 编写 `frontend/src/composables/useWorkspaceBus.test.ts` 与 `useEditorTabs.test.ts` 扩展测试，确保用例 100% 通过
- [x] 1.5 **执行提交 1**：`feat(workbench): add workspace bus and diff tab support to useEditorTabs`


## 阶段 2：编辑器 Tab 视觉增强与 Diff 标签页渲染 (Editor Tab & Diff View) [已完成]
- [x] 2.1 在 `frontend/src/views/ExplorerView.vue` 与 `EditorTabs.vue` 中支持 Diff 模式的 Monaco 差异编辑器渲染（复用现有 diff 引擎）
- [x] 2.2 在 Tab 栏中实现 Git 状态色彩（已修改黄色 M、新增绿色 U 等）与未保存圆点 `●` 的并存样式
- [x] 2.3 验证多 Tab 之间普通代码视图与 Diff 视图的平滑切换与关闭逻辑
- [x] 2.4 **执行提交 2**：`feat(editor): render diff tab in main editor and enhance tab git status decorators`


## 阶段 3：文件管理器 ⟷ SCM 双向定位联动 (Bidirectional Reveal & Navigation) [已完成]
- [x] 3.1 在 `frontend/src/views/ExplorerView.vue` 中监听 `workspaceBus.onRevealFile` 事件，实现递归自动展开父级目录并滚动平滑聚焦
- [x] 3.2 在 `frontend/src/views/SourceControlView.vue` 变更列表项中增加「在资源管理器中揭示」按钮
- [x] 3.3 在 SCM 列表点击文件或点击对比按钮时，通过总线在主编辑器中无缝打开 Diff Tab 并切换导航
- [x] 3.4 **执行提交 3**：`feat(explorer): implement auto-reveal and highlight from source control`


## 阶段 4：文件管理器右键 Git 上下文动作融合 (Context Menu Cross-Pollination) [已完成]
- [x] 4.1 在 `frontend/src/components/explorer/FileContextMenu.vue` 中动态根据当前选中项的 Git 状态注入 Git 操作组（暂存/取消暂存/放弃更改/对比差异）
- [x] 4.2 联动 `useGit`，操作完成后触发响应式状态总线通知，使 Explorer 树徽标与 SCM 变更列表即刻同步刷新
- [x] 4.3 编写测试验证文件上下文菜单的 Git 行为派发
- [x] 4.4 **执行提交 4**：`feat(explorer): integrate git actions into file context menu`


## 阶段 5：全 53 国多语言支持与端到端验证 (i18n & Verification) [已完成]
- [x] 5.1 更新 `frontend/src/i18n/types.ts`、`zh-CN.ts` 与 `en-US.ts`，增加 Reveal、Diff Tab、Git 菜单相关词条
- [x] 5.2 自动同步词条到所有 53 种语言文件，确保 `i18n.test.ts` 0 报错
- [x] 5.3 运行全量 `npm run test` 与 `npm run build`，确保 100% 成功
- [x] 5.4 **执行提交 5**：`feat(i18n): sync multi-language keys and finalize workbench trinity verification`


## 阶段 6：文件系统与生命周期状态深度同步 (Lifecycle Sync) [已完成]
- [x] 6.1 在 `useWorkspaceBus.ts` 中引入 `fileRenamed` 与 `fileDeleted` 事件广播
- [x] 6.2 在 `useEditorTabs.ts` 中实现重命名与删除侦听，自动平滑迁移已打开 Tab 的路径与原始内容引用
- [x] 6.3 在 `useGit.ts` 中增加与编辑器未保存脏状态（Dirty）的安全保护与状态双向同步
- [x] 6.4 **执行提交 6**：`feat(workbench): synchronize editor dirty state with git stage, discard and checkout operations` & `feat(workbench): propagate file rename and delete events to seamlessly migrate open editor tabs`


## 阶段 7：编辑器实时 Git Gutter 差异指示与行内 Diff 交互 (Git Gutter & Inline Diff) [已完成]
- [x] 7.1 实现 `frontend/src/utils/diffGutter.ts`，利用高效行级 Diff 算法计算新增、修改与删除行区间
- [x] 7.2 在 `MonacoEditor.vue` 中动态挂载行号槽增量指示条（新增绿色、修改蓝色、删除红色小三角形）
- [x] 7.3 点击行号槽指示条弹出 `inline-diff-peek` 气泡微窗，并支持一键「还原此块 (Revert Chunk)」
- [x] 7.4 **执行提交 7**：`feat(editor): render real-time git diff gutter indicators for added, modified, and deleted lines` & `feat(editor): support inline diff peek and single block revert by clicking gutter indicators`


## 阶段 8：文件树 Git 聚焦排查模式与 Gitignore 识别 (Git Focus Mode) [已完成]
- [x] 8.1 在 `ExplorerSidebar.vue` 顶部操作栏引入「Git 聚焦模式」切换开关
- [x] 8.2 在 `useFileTree.ts` 中实现树节点修剪算法，激活聚焦时仅保留有 Git 修改文件的目录祖先分支
- [x] 8.3 在 `FileTreeNode.vue` 与 Vite Dev Server 中动态对齐 `.gitignore` 规则匹配文件的灰度透明样式
- [x] 8.4 **执行提交 8**：`feat(explorer): add Git Focus mode to filter and prune file tree to changed branches only`


## 阶段 9：底部状态栏 Git 联动与多维穿透导航 (Status Bar Synergy) [已完成]
- [x] 9.1 在 `ExplorerView.vue` 底部状态栏展示当前活动 Git 分支名与工作区修改文件计数统计
- [x] 9.2 点击分支名自动呼出 SCM 面板并聚焦当前文件的 Git 状态
- [x] 9.3 **执行提交 9**：`feat(workbench): connect editor status bar with active branch and file Git status navigation`


## 阶段 10：Monaco 编辑器右键菜单主题与本地化对齐 (Context Menu Refinement) [已完成]
- [x] 10.1 解决 Shadow DOM 渲染下的上下文菜单样式隔离与对比度问题
- [x] 10.2 注入并持久化多语言翻译字典，使右键菜单与 53 国语言系统保持一致
- [x] 10.3 **执行提交 10**：`style(editor): ensure outline and border none on context menu action items`


## 阶段 11：行内代码块一键暂存与部分提交 (Stage Chunk / Partial Staging) [待执行]
- [ ] 11.1 在 `frontend/src/composables/useGit.ts` 中扩充暂存局部代码块支持（`stageChunk(filePath: string, chunk: DiffLineChange, currentFileContent: string)`）
- [ ] 11.2 在 `frontend/src/components/editor/MonacoEditor.vue` 的 `inline-diff-peek` 操作区增加「暂存此块 (Stage Chunk)」按钮及响应逻辑
- [ ] 11.3 暂存完成后，触发 `workspaceBus.emitGitActionTriggered`，自动刷新 Git 状态与行号槽指示条
- [ ] 11.4 编写单测验证局部代码块暂存与行号槽响应
- [ ] 11.5 **执行提交 11**：`feat(editor): support staging single diff chunk from inline peek widget`


## 阶段 12：文件树折叠目录变更聚合徽标 (Folder Git Bubble Badges) [待执行]
- [ ] 12.1 在 `frontend/src/composables/useFileTree.ts` 中增加递归计算折叠目录内部子孙变动文件总数的响应式方法
- [ ] 12.2 在 `frontend/src/components/explorer/FileTreeNode.vue` 目录节点行尾增加高对比度数字气泡徽标（Bubble Count Badge）
- [ ] 12.3 折叠时显示气泡，展开时自适应隐藏或淡化，避免视觉杂乱
- [ ] 12.4 编写单测验证目录聚合徽标计算正确性
- [ ] 12.5 **执行提交 12**：`feat(explorer): display aggregated git change badges on collapsed directory nodes`


## 阶段 13：全量 53 国多语言同步与端到端回归验证 (Final Verification) [待执行]
- [ ] 13.1 在 `frontend/src/i18n/` 中补齐“暂存此块”等多语言词条，并同步更新至 53 种语言
- [ ] 13.2 运行前端全量单元测试（`npm run test -- --run`）与打包构建检查（`npm run build`），确保 0 警告 0 报错
- [ ] 13.3 更新 `tasks.md` 标记全部阶段任务为已完成 `[x]`
- [ ] 13.4 **执行提交 13**：`feat(i18n): sync multi-language keys and finalize phase 11-12 validation`


