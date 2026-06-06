# Cross-Platform Path Handling Refactoring Plan

## Goal
>
>
## User Review Required
>
>
> [!IMPORTANT]
> 请确认以下决策：
>
>
> - 抽象出 `JoinForOS`、`JoinForPosix`、`FormatDirForPosix` 及 `JoinForPowerShell` 接口行为是否覆盖了所有目标场景？
>

## Open Questions
>
>
> [!WARNING]
>
>
>
> - 旧版本中注入 `cmd/23.exec.go` 的路径是否存在针对特定用户场景的硬编码？（重构后应当统一）。

## Proposed Changes

### New Package

- 实现根据上下文动态转换的路径组装和格式化函数。

#### [MODIFY] `cmd/25.run.go`

#### [MODIFY] `cmd/23.exec.go`

#### [MODIFY] `internal/service/activation.go`

- 使用 `envpath.JoinForPosix` 处理 bash/zsh 的脚本路径替换。

- 提取 shell 输出环境变量的行为。

## Verification Plan

### Automated Tests

- 为 `pkg/envpath` 编写完整的单元测试覆盖率。
- 执行 `go test ./...` 确保重构不打破现有逻辑。

### Manual Verification

- 跨平台（Windows + Linux）运行 GitHub Actions 的 Pre-flight Integrity Check。
