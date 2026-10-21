# 任务拆分与执行跟踪清单：Quality Guard 一级入口升级与文件扩展名管线矩阵 (Tasks 054)

## 任务执行进度概览
- **当前阶段**：Phase 1 准备启动
- **总体进度**：0 / 12 完成 (0%)

---

## Phase 1: 升级 ActivityBar 与路由，确立 Quality Guard 一级入口
- [x] Task 1.1: 在 `frontend/src/router/index.ts` 中注册 `/guard` 路由，指向 `views/GuardView.vue`
- [x] Task 1.2: 在 `frontend/src/components/layout/ActivityBar.vue` 中增加专属一级盾牌图标导航项与徽章
- [x] Task 1.3: 在 `frontend/src/components/layout/SidebarPanel.vue` 中配置 Guard 侧边栏面板容器
- [x] Task 1.4: 在 `frontend/src/views/SourceControlView.vue` 中加入指向一级 Guard 视图的轻量引导条
- [x] Task 1.5: 阶段性 Git 提交：`feat(ui): elevate quality guard to first-class navigation entry`

---

## Phase 2: 后端 CRUD API 与 TOML 兜底持久化
- [x] Task 2.1: 在 `cmd/64.ui_guard.go` 中实现 `POST /api/guard/slots`（写入/更新 `.unirtm.toml`）与 `DELETE /api/guard/slots`（删除配置）
- [x] Task 2.2: 在 `frontend/vite.config.ts` 的 Vite dev mock 中支持 `POST` 和 `DELETE /api/guard/slots`
- [x] Task 2.3: 在 `frontend/src/composables/useGuard.ts` 中封装 `saveSlotConfig`、`deleteSlotConfig` 与 `fetchRawTomlConfig` / `saveRawTomlConfig`
- [x] Task 2.4: 编写 Go 后端 slots 配置持久化测试与 Vite mock 接口验证
- [x] Task 2.5: 阶段性 Git 提交：`feat(guard): provide restful crud api and toml persistence for tool slots`

---

## Phase 3: 构建按扩展名聚合矩阵表格与可视化 CRUD 模态框
- [ ] Task 3.1: 在 `frontend/src/views/GuardView.vue` 中构建扩展名管线矩阵表格（Extension Pipeline Table），直观展示各文件后缀对应的 Linter, Formatter, Fixer
- [ ] Task 3.2: 编写 `frontend/src/components/guard/ExtensionSlotModal.vue` 支持从 Tools 工具箱中下拉选取、文件后缀 Tag 输入、多字段定制
- [ ] Task 3.3: 编写 `frontend/src/components/guard/TomlConfigDrawer.vue` 实现 `.unirtm.toml` 兜底查看与直接编辑通道
- [ ] Task 3.4: 完善前端多语言（53 种语言扩展名表格与 CRUD 词条补齐）
- [ ] Task 3.5: 阶段性 Git 提交：`feat(ui): implement extension pipeline matrix table and visual crud modal`

---

## Phase 4: 全流程联调、自动化测试与规范检查
- [ ] Task 4.1: 运行前端自动化测试套件（`npm test`）确保所有组件与测试用例 100% 通过
- [ ] Task 4.2: 运行后端单元测试（`go test -v ./...`）确保持久化逻辑安全稳定
- [ ] Task 4.3: 运行 `unirtm check --all` 确保全仓 0 错误 0 警告
- [ ] Task 4.4: 阶段性 Git 提交：`test(guard): add comprehensive tests for first-class guard hub and extension matrix`
