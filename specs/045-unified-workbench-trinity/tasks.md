# 任务清单：文件管理器、编辑器与源代码管理三位一体协同系统 (Tasks)

## 阶段 1：统一工作区总线与 Diff 标签页模型 (Core Composable & Model)
- [x] 1.1 在 `frontend/src/composables/useEditorTabs.ts` 中增强 `EditorTab` 模型，支持 `isDiff`、`diffOriginalPath`、`diffOriginalContent`、`diffModifiedContent` 等属性
- [x] 1.2 在 `useEditorTabs.ts` 中实现 `openDiffTab(fileDiff: DiffTabOptions)` 与关闭/定位逻辑
- [x] 1.3 创建 `frontend/src/composables/useWorkspaceBus.ts` 轻量响应式协同总线（支持 `revealInExplorer`、`openInEditor`、`activeFileChanged` 等）
- [x] 1.4 编写 `frontend/src/composables/useWorkspaceBus.test.ts` 与 `useEditorTabs.test.ts` 扩展测试，确保用例 100% 通过
- [x] 1.5 **执行提交 1**：`feat(workbench): add workspace bus and diff tab support to useEditorTabs`


## 阶段 2：编辑器 Tab 视觉增强与 Diff 标签页渲染 (Editor Tab & Diff View)
- [ ] 2.1 在 `frontend/src/views/ExplorerView.vue` 与 `EditorTabs.vue` 中支持 Diff 模式的 Monaco 差异编辑器渲染（复用现有 diff 引擎）
- [ ] 2.2 在 Tab 栏中实现 Git 状态色彩（已修改黄色 M、新增绿色 U 等）与未保存圆点 `●` 的并存样式
- [ ] 2.3 验证多 Tab 之间普通代码视图与 Diff 视图的平滑切换与关闭逻辑
- [ ] 2.4 **执行提交 2**：`feat(editor): render diff tab in main editor and enhance tab git status decorators`

## 阶段 3：文件管理器 ⟷ SCM 双向定位联动 (Bidirectional Reveal & Navigation)
- [ ] 3.1 在 `frontend/src/views/ExplorerView.vue` 中监听 `workspaceBus.onRevealFile` 事件，实现递归自动展开父级目录并滚动平滑聚焦
- [ ] 3.2 在 `frontend/src/views/SourceControlView.vue` 变更列表项中增加「在资源管理器中揭示」按钮
- [ ] 3.3 在 SCM 列表点击文件或点击对比按钮时，通过总线在主编辑器中无缝打开 Diff Tab 并切换导航
- [ ] 3.4 **执行提交 3**：`feat(explorer): implement auto-reveal and highlight from source control`

## 阶段 4：文件管理器右键 Git 上下文动作融合 (Context Menu Cross-Pollination)
- [ ] 4.1 在 `frontend/src/components/explorer/FileContextMenu.vue` 中动态根据当前选中项的 Git 状态注入 Git 操作组（暂存/取消暂存/放弃更改/对比差异）
- [ ] 4.2 联动 `useGit`，操作完成后触发响应式状态总线通知，使 Explorer 树徽标与 SCM 变更列表即刻同步刷新
- [ ] 4.3 编写测试验证文件上下文菜单的 Git 行为派发
- [ ] 4.4 **执行提交 4**：`feat(explorer): integrate git actions into file context menu`

## 阶段 5：全 53 国多语言支持与端到端验证 (i18n & Verification)
- [ ] 5.1 更新 `frontend/src/i18n/types.ts`、`zh-CN.ts` 与 `en-US.ts`，增加 Reveal、Diff Tab、Git 菜单相关词条
- [ ] 5.2 自动同步词条到所有 53 种语言文件，确保 `i18n.test.ts` 0 报错
- [ ] 5.3 运行全量 `npm run test` 与 `npm run build`，确保 100% 成功
- [ ] 5.4 **执行提交 5**：`feat(i18n): sync multi-language keys and finalize workbench trinity verification`
