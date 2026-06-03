# Cross-Platform Path Handling Refactoring Plan

## Goal
抽象出 `pkg/envpath` 工具包，集中处理跨平台下的路径格式和列表分隔符，解决之前散落在代码各处的环境路径处理（特别是 Windows 与 POSIX shell 之间）导致的兼容性 bug。

## User Review Required
> [!IMPORTANT]
> 请确认以下决策：
> - 抽象出 `JoinForOS`、`JoinForPosix`、`FormatDirForPosix` 及 `JoinForPowerShell` 接口行为是否覆盖了所有目标场景？

## Open Questions
> [!WARNING]
> - 旧版本中注入 `cmd/23.exec.go` 的路径是否存在针对特定用户场景的硬编码？（重构后应当统一）。

## Proposed Changes
### New Package
#### [NEW] `internal/pkg/envpath/envpath.go`
- 实现根据上下文动态转换的路径组装和格式化函数。

### Refactored Call Sites
#### [MODIFY] `cmd/25.run.go`
- 将原生操作的 `strings.Join` 替换为 `envpath.JoinForOS`。
#### [MODIFY] `cmd/23.exec.go`
- 替换路径组装。
#### [MODIFY] `internal/service/activation.go`
- 使用 `envpath.JoinForPosix` 处理 bash/zsh 的脚本路径替换。
#### [MODIFY] `cmd/3.env.go`
- 提取 shell 输出环境变量的行为。

## Verification Plan
### Automated Tests
- 为 `pkg/envpath` 编写完整的单元测试覆盖率。
- 执行 `go test ./...` 确保重构不打破现有逻辑。

### Manual Verification
- 跨平台（Windows + Linux）运行 GitHub Actions 的 Pre-flight Integrity Check。
