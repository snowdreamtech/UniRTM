# 任务拆分与执行跟踪清单：UniRTM 工程质量守门员与质检 Harness 重构 (Tasks 053)

## 任务执行进度概览
- **当前阶段**：Phase 0 (方案设计与基线建立)
- **总体进度**：2 / 18 完成 (11%)

---

## Phase 0: 方案设计与基线建立 (Docs Baseline)
- [x] Task 0.1: 创建 `specs/053-refactor-quality-guard-harness/` 规范、方案与任务文档
- [x] Task 0.2: 检出特性分支 `053-refactor-quality-guard-harness` 并提交文档基线

---

## Phase 1: Go 原生核心引擎与内置文本卫生/安全规则 (Core Engine & Builtin Rules)
- [x] Task 1.1: 创建 `internal/guard/types.go` 定义 Category (6大领域)、Phase (三阶流水线)、Severity、Issue、Rule 接口与配置结构
- [x] Task 1.2: 创建 `internal/guard/walker.go` 实现极速文件遍历引擎（支持 staged/changed/all 模式，按 extensions 过滤，遇到 node_modules/vendor/dist 等自动 SkipDir）
- [x] Task 1.3: 实现核心纯 Go 内置规则（`internal/guard/builtin/`）：
  - `whitespace.go` (去除行尾空格，支持 fix)
  - `eof.go` (单换行结尾，支持 fix)
  - `utf8bom.go` (去除 BOM，支持 fix)
  - `merge_conflict.go` (检测冲突标记)
  - `private_key.go` (检测私钥泄漏)
  - `env_safety.go` (检测 .env 生产凭据)
  - `large_files.go` (检测大文件拦截)
  - `case_conflict.go` (检测大小写文件名冲突)
  - `lockfiles.go` (检测 lockfile 共存与冲突)
- [x] Task 1.4: 编写 `internal/guard/builtin_test.go` 与 `walker_test.go` 单元测试，确保 100% 覆盖与极致性能
- [x] Task 1.5: 阶段性 Git 提交：`feat(guard): implement core engine and Go-native builtin rules`

---

## Phase 2: 工具插槽 (Slots)、三阶执行管道 (Pipeline) 与 CLI 命令 (`unirtm check/fix`)
- [x] Task 2.1: 实现 `internal/guard/slots.go`（支持插槽定义如 `web.formatter`, `web.linter`, `go.linter`，并支持自定义工具动态替换）
- [x] Task 2.2: 实现 `internal/guard/pipeline.go`（严格按照 Fixer -> Formatter -> Linter 顺序执行调度，支持自动探测 `.golangci.yml`, `.prettierrc` 等原生配置文件）
- [x] Task 2.3: 实现 CLI 命令 `cmd/66.check.go` (`unirtm check [--staged|--all]`) 与 `unirtm fix`
- [x] Task 2.4: 编写 CLI 命令行单元测试与 E2E 测试
- [x] Task 2.5: 阶段性 Git 提交：`feat(guard): implement pluggable tool slots, 3-phase pipeline, and check/fix CLI`

---

## Phase 3: 双 Profile 隔离 (Pre-commit vs CI) 与 Git Hook 桥接优化
- [x] Task 3.1: 在 `.unirtm.toml` 中支持 `[guard.profiles.pre-commit]`（轻量极速、包含 gitleaks protect --staged）与 `[guard.profiles.ci]`（深度全量审计）
- [x] Task 3.2: 升级 `cmd/62.hook.go`，将 `unirtm hook run pre-commit` 无缝桥接到 Guard pre-commit profile
- [x] Task 3.3: 编写 profile 调度测试与 Hook 桥接测试
- [x] Task 3.4: 阶段性 Git 提交：`feat(guard): support pre-commit and CI dual profiles with gitleaks integration`

---

## Phase 4: Web UI 守门中心、SCM 侧边栏联动与可视化配置抽屉
- [ ] Task 4.1: 在 Go 后端 `cmd/64.ui_guard.go` 中实现 `/api/guard/status`, `/api/guard/run`, `/api/guard/rules` REST API
- [ ] Task 4.2: 在 `frontend/src/components/scm/SourceControlSidebar.vue` 提交区嵌入守门状态指示胶囊与一键修复按钮
- [ ] Task 4.3: 在 `frontend/src/views/SourceControlView.vue`（或对应质量中心面板）中实现 6 大分类卡片、智能生态展示与规则开关抽屉
- [ ] Task 4.4: 支持点击违规文件行号与 Monaco Editor 联动跳转定位
- [ ] Task 4.5: 阶段性 Git 提交：`feat(ui): add quality guard hub, SCM sidebar status capsule, and one-click fix in Web UI`

---

## Phase 5: 全量测试回归、53 语言包本地化与总结归档
- [ ] Task 5.1: 在 53 种语言包中补充质量守门相关文案（`guard.*`）
- [ ] Task 5.2: 运行 Go 全量测试（`go test -v ./...`）与前端测试（`npm run test`, `vue-tsc`），确保 0 报错
- [ ] Task 5.3: 阶段性 Git 提交：`chore(i18n): complete guard locales and finalize test harness refactoring`
