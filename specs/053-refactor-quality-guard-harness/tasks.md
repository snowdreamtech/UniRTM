# 任务拆分与执行跟踪清单：UniRTM 工程质量守门员与质检 Harness 重构 (Tasks 053)

## 任务执行进度概览
- **当前阶段**：全部完成 (100%)
- **总体进度**：18 / 18 完成 (100%)

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
  - `forbidden_files.go` (检测临时与残留垃圾文件)
  - `symlinks.go` (检测死链与绝对路径软链)
  - `shebang.go` (可执行脚本 shebang 校验与脚本执行权限自愈修复)
  - `svg_safety.go` (检测 SVG 脚本注入与 XSS 安全漏洞)
  - `json.go` (JSON / JSONC 语法格式有效性校验)
  - `toml.go` (TOML 配置文件语法有效性校验)
  - `mixed_line_endings.go` (跨平台混合换行符检测与自动自愈修复)
  - `rule_xml.go` (XML/XSD/XSL/PLIST 闭合良好性与语法有效性校验)
  - `rule_cloud_secrets.go` (离线云厂商及平台凭据/API Key 泄漏防范)
  - `rule_commit_msg.go` (Conventional Commits 规范与提交信息格式校验)
  - `rule_yaml.go` (纯 Go 多文档 YAML / YML 配置文件语法有效性校验)
  - `rule_debug_statements.go` (防呆拦截遗留在源码中的断点与调试语句)
  - `rule_binary_artifacts.go` (拦截意外提交的编译二进制、静态库与压缩包产物)
  - `rule_empty_files.go` (检测无意义的空代码文件残留，白名单放行占位文件)
  - `rule_script_crlf.go` (检测 Shell 脚本 CRLF 换行符并在 Phase 1 自动修复为 LF，避免 bad interpreter 错误)
  - `rule_markdown.go` (检测 Markdown 标题层级跳跃、未闭合代码块与空链接/占位链接)
  - `rule_ip_addresses.go` (检测源码和配置文件中硬编码的私有和公网 IPv4 地址)
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
- [x] Task 4.1: 在 Go 后端 `cmd/64.ui_guard.go` 中实现 `/api/guard/status`, `/api/guard/run`, `/api/guard/rules` REST API
- [x] Task 4.2: 在 `frontend/src/components/scm/SourceControlSidebar.vue` 提交区嵌入守门状态指示胶囊与一键修复按钮
- [x] Task 4.3: 在 `frontend/src/views/SourceControlView.vue` 中实现质量守门概览大盘与诊断展开
- [x] Task 4.4: 编写 Web UI 端状态交互测试与后端 Guard API 单元测试
- [x] Task 4.5: 阶段性 Git 提交：`feat(ui): add quality guard hub, SCM sidebar status capsule, and one-click fix in Web UI`

---

## Phase 5: 全量测试回归、53 语言包本地化与总结归档
- [x] Task 5.1: 在 53 种语言包中补充质量守门相关文案（`guard.*`）
- [x] Task 5.2: 运行 Go 全量测试（`go test -v ./...`）与前端测试（`npm run test`, `vue-tsc`），确保 0 报错
- [x] Task 5.3: 阶段性 Git 提交：`chore(i18n): complete guard locales and finalize test harness refactoring`
