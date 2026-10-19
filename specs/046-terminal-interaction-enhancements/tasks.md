# 任务拆分与执行跟踪清单：集成终端深度交互与工作区协同增强系统 (Tasks)

## 任务执行进度概览
- **当前阶段**：Phase 0 - 准备就绪
- **总体进度**：0 / 14 已完成 (0%)

---

## Phase 0: 方案设计与基线建立 (Docs Baseline)
- [x] Task 0.1: 创建 `specs/046-terminal-interaction-enhancements/` 规范、架构与任务文档
- [x] Task 0.2: 检出特性分支 `046-terminal-interaction-enhancements` 并提交文档基线

---

## Phase 1: 依赖升级、超链接识别与堆栈穿透跳转 (Hyperlinks & Navigation)
- [x] Task 1.1: 安装 `@xterm/addon-search` 与 `@xterm/addon-web-links`
- [x] Task 1.2: 实现 `frontend/src/composables/useTerminalLinks.ts`，支持 Web 链接与文件路径解析（`:line:col` 提取）
- [x] Task 1.3: 在 `frontend/src/components/terminal/XTerminalPane.vue` 中集成链接提供者，与 `useWorkspaceBus` 的 `openInEditor` 联动
- [x] Task 1.4: 编写 `frontend/src/composables/useTerminalLinks.test.ts` 单元测试并通过验证
- [x] Task 1.5: 阶段性 Git 提交：`feat(terminal): support file stack trace and web url click navigation`

---

## Phase 2: 终端浮动搜索工具条 (Terminal In-Panel Search)
- [x] Task 2.1: 创建 `frontend/src/components/terminal/TerminalSearchWidget.vue` 浮动搜索条组件
- [x] Task 2.2: 在 `XTerminalPane.vue` 中集成 `@xterm/addon-search` 的 `SearchAddon` 实例并与搜索组件绑定
- [x] Task 2.3: 支持快捷键 `Ctrl+F` 呼出/聚焦，`Enter` / `Shift+Enter` 查找与 `Esc` 退出
- [x] Task 2.4: 编写 `frontend/src/composables/useTerminalSearch.test.ts` 测试并通过验证
- [x] Task 2.5: 阶段性 Git 提交：`feat(terminal): add in-terminal search widget and shortcut navigation`

---

## Phase 3: 终端右键上下文菜单与编辑器/文件树双向互通 (Context Menu & Bidirectional Flow)
- [ ] Task 3.1: 创建 `frontend/src/components/terminal/TerminalContextMenu.vue` 右键菜单（复制/粘贴/清屏/全选/拆分/查找）
- [ ] Task 3.2: 升级 `useWorkspaceBus.ts`，增加 `sendToTerminal` 与 `openInIntegratedTerminal` 通道
- [ ] Task 3.3: 在 `MonacoEditor.vue` 中注册 `Shift+Enter` 快捷键及右键菜单「在活动终端中运行代码」
- [ ] Task 3.4: 在 `FileTreeNode.vue` 中为目录节点右键菜单新增「在集成终端中打开」
- [ ] Task 3.5: 编写相关单元测试并通过验证
- [ ] Task 3.6: 阶段性 Git 提交：`feat(terminal): implement context menu and bidirectional workbench flow`

---

## Phase 4: 53 种语言国际化字典适配与全量集成验证 (i18n & Production Release)
- [ ] Task 4.1: 在 `frontend/src/locales/` 中为所有 53 种语言同步新增终端交互相关词条
- [ ] Task 4.2: 运行全量单元测试（`npm run test`）确保所有测试用例绿灯
- [ ] Task 4.3: 运行生产环境打包（`npm run build`）确保 0 警告与 0 错误
- [ ] Task 4.4: 阶段性 Git 提交：`feat(terminal): complete i18n support and production release`
