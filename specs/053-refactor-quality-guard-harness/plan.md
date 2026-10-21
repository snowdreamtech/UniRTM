# 技术方案与架构设计：UniRTM 工程质量守卫与质检 Harness 重构 (Plan 053)

## 1. 总体架构设计

质量守卫系统（UniRTM Guard & Quality Harness）分为三大核心层次：
1. **Core Guard Engine (Go 核心库 `internal/guard`)**：负责模式判定、文件过滤、生态嗅探、三阶流水线调度及内置原生纯 Go 规则；
2. **CLI Command Facade (`cmd/66.guard.go` / `cmd/66.check.go` / `cmd/62.hook.go`)**：暴露极简的 `unirtm check`, `unirtm fix`, `unirtm fmt` 与 Git Hook 桥接命令；
3. **Web UI & API Layer (`cmd/64.ui_guard.go` + `frontend/src/components/scm/SourceControlSidebar.vue` + `frontend/src/views/SourceControlView.vue`)**：提供 HTTP 状态/执行接口，呈现守卫胶囊、一键修复按钮、规则管理抽屉与结构化问题定位。

```
+-----------------------------------------------------------------------------------+
|                                 UniRTM Web UI                                     |
|  [守卫胶囊 🟢/🟡/🔴]  [⚡ 一键全自动修复]  [⚙️ 规则开关与过滤抽屉]  [Monaco 诊断定位]      |
+-----------------------------------------------------------------------------------+
                                       |
                   HTTP REST APIs (/api/guard/*)
                                       |
+-----------------------------------------------------------------------------------+
|                             CLI Interface (cmd)                                   |
|       unirtm check [--staged|--all]  |  unirtm fix  |  unirtm hook run pre-commit |
+-----------------------------------------------------------------------------------+
                                       |
+-----------------------------------------------------------------------------------+
|                        Core Guard Engine (internal/guard)                         |
|                                                                                   |
|  1. Fast File Walker: Changed/Staged/All + Extensions + SkipDir (node_modules)   |
|  2. Ecosystem Sniffer: Auto-detect go.mod, package.json, pyproject.toml           |
|  3. Slot & Tool Registry: Pluggable tools (eslint, biome, golangci-lint, ruff...) |
|  4. Three-Phase Pipeline Orchestrator:                                            |
|       Phase 1: Fixer (BOM, Whitespace, EOF, Shebang, imports, ruff --fix)        |
|       Phase 2: Formatter (gofmt, prettier, taplo, shfmt)                         |
|       Phase 3: Linter & Auditor (Conflict, Secret/Gitleaks, SAST, CVE/Trivy)     |
+-----------------------------------------------------------------------------------+
                                       |
+-----------------------------------------------------------------------------------+
|                        Single Source of Truth (.unirtm.toml)                      |
|                  [guard] + [guard.profiles] + [guard.rules]                       |
+-----------------------------------------------------------------------------------+
```

---

## 2. 核心模块与数据结构设计

### 2.1 规则与工具抽象 (`internal/guard/types.go`)
```go
package guard

import "context"

type Category string
const (
    CategoryHygiene  Category = "hygiene"
    CategoryFormat   Category = "format"
    CategoryLint     Category = "lint"
    CategorySecret   Category = "secret"
    CategorySecurity Category = "security"
    CategoryPolicy   Category = "policy"
)

type Phase string
const (
    PhaseFixer     Phase = "fixer"
    PhaseFormatter Phase = "formatter"
    PhaseLinter    Phase = "linter"
)

type Severity string
const (
    SeverityError   Severity = "error"
    SeverityWarning Severity = "warning"
    SeverityInfo    Severity = "info"
)

type Issue struct {
    FilePath string   `json:"file_path"`
    Line     int      `json:"line"`
    Column   int      `json:"column"`
    RuleID   string   `json:"rule_id"`
    Category Category `json:"category"`
    Phase    Phase    `json:"phase"`
    Severity Severity `json:"severity"`
    Message  string   `json:"message"`
    Fixable  bool     `json:"fixable"`
}

type Rule interface {
    ID() string
    Category() Category
    Phase() Phase
    DefaultExtensions() []string
    Execute(ctx context.Context, files []string, opts ExecuteOptions) ([]Issue, error)
}
```

### 2.2 内置原生规则库 (`internal/guard/builtin/`)
迁移原 Python 脚本核心逻辑为纯 Go 实现：
- `whitespace.go`：去除行尾多余空白（支持自动 fix），跳过符号链接和大文件；
- `eof.go`：确保文件末尾有单一 `\n`（支持自动 fix）；
- `utf8bom.go`：检测并剥离 UTF-8 BOM 头（支持自动 fix）；
- `merge_conflict.go`：检测 `<<<<<<<`, `=======`, `>>>>>>>` 冲突标记；
- `private_key.go`：检测 `BEGIN (RSA|EC|DSA|OPENSSH) PRIVATE KEY`；
- `env_safety.go`：检测 `.env` 文件内的高熵密码与秘钥泄漏；
- `large_files.go`：检测增量中大于限制（默认 1MB）的大文件；
- `case_conflict.go`：比对文件路径大小写冲突；
- `lockfiles.go`：检测包管理器共存冲突与锁文件标记。

### 2.3 文件遍历引擎 (`internal/guard/walker.go`)
- 模式判定：
  - `staged`：执行 `git diff --cached --name-only --diff-filter=ACMR`；
  - `changed`：执行 `git status --porcelain` 汇聚 staged + unstaged 增量；
  - `all`：遍历工作区目录。
- 剪枝策略：
  - 遇到 `node_modules`, `vendor`, `dist`, `build`, `out`, `target`, `.git`, `.venv` 时，返回 `filepath.SkipDir`；
  - 后缀匹配采用 `map[string]bool` 瞬时检索。

### 2.4 插槽与工具调度器 (`internal/guard/runner.go`)
- 读取 `.unirtm.toml` 中 `[guard]` 配置，支持槽位平替；
- 按照 `Fixer -> Formatter -> Linter` 管道分阶段并发或串行运行；
- 外部工具驱动使用 `unirtm exec` 或直接执行二进制，自动探测 `.golangci.yml`, `.prettierrc` 等原生配置文件。

### 2.5 Web 交互与 API 响应 (`cmd/64.ui_guard.go`)
- `GET /api/guard/status`：返回当前分支变动文件检测概况及各分类违规计数；
- `POST /api/guard/run`：触发指定阶段（`fixer` / `formatter` / `linter` / `all`）运行，返回结构化 Issues 列表；
- `GET /api/guard/rules` & `PUT /api/guard/rules`：管理规则开关及目录/后缀过滤配置。

---

## 3. 分阶段实施演进计划

- **Phase 0: 规范制定、技术方案与基线构建 (Baseline)**
- **Phase 1: Go 原生核心引擎与内置文本卫生/安全规则 (Core Engine & Builtin Rules)**
- **Phase 2: 工具插槽 (Slots)、三阶执行管道 (Pipeline) 与 CLI 命令 (`unirtm check/fix`)**
- **Phase 3: 双 Profile 隔离 (Pre-commit vs CI) 与 Git Hook 桥接优化**
- **Phase 4: Web UI 守卫中心、SCM 侧边栏联动与可视化配置抽屉**
- **Phase 5: 全量测试回归、性能对比验证与渐进提交**
