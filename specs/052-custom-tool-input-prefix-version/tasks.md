# 任务拆分与执行跟踪清单：自定义工具前缀选择与固定版本号录入 (Tasks)

## 任务执行进度概览
- **当前阶段**：Phase 0 - 方案设计与基线建立 (已完成)
- **总体进度**：2 / 12 完成 (16.7%)

---

## Phase 0: 方案设计与基线建立 (Docs Baseline)
- [x] Task 0.1: 创建 `specs/052-custom-tool-input-prefix-version/` 规范、方案与任务文档
- [x] Task 0.2: 检出特性分支 `052-custom-tool-input-prefix-version` 并提交文档基线

---

## Phase 1: 规范解析与格式化工具函数 (Tool Spec Parser & Tests)
- [ ] Task 1.1: 创建 `frontend/src/utils/toolSpecParser.ts`（支持前缀列表、全格式字符串粘贴智能解析、Spec 拼装）
- [ ] Task 1.2: 编写 `frontend/src/utils/toolSpecParser.test.ts` 完整单元测试（覆盖多前缀、作用域包、版本号提取）
- [ ] Task 1.3: 阶段性 Git 提交：`feat(tools): add tool spec parser and test suite`

---

## Phase 2: 复合分段输入栏 UI 渲染与确定性版本逻辑 (Segmented UI & Deterministic Add)
- [ ] Task 2.1: 在 `ToolsView.vue` 中重构自定义工具区域为复合三段输入栏（前缀下拉 + 工具名 + 固定版本号）
- [ ] Task 2.2: 实现智能粘贴监听（粘贴完整 Spec 自动拆解到三段字段）
- [ ] Task 2.3: 废弃 `latest` 逻辑，要求必须输入固定确切版本号
- [ ] Task 2.4: 增加实时配置预览（展示 `.unirtm.toml` 语法条目）
- [ ] Task 2.5: 阶段性 Git 提交：`feat(tools): implement segmented triple-field input for custom tool with explicit version`

---

## Phase 3: 53 种国际化本地化与全量回归 (i18n & Regression)
- [ ] Task 3.1: 在 53 种语言包中增加前缀说明与版本号占位文案（`tools.custom.prefix_*`, `tools.custom.version_placeholder`, `tools.custom.preview_hint`）
- [ ] Task 3.2: 运行全量单元测试与类型检查确保 0 报错
- [ ] Task 3.3: 阶段性 Git 提交：`feat(i18n): complete 53 locales and finalize custom tool input UX`
