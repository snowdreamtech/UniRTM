# Hook Arrays Support Implementation Plan

## Goal
支持在配置文件中将 `pre_install`, `post_install`, 和 task `run` 的定义从单个字符串改为字符串数组，以便于顺序执行多个脚本并提高配置的可读性。

## User Review Required
> [!IMPORTANT]
> 请确认以下决策：
> - 是否向下兼容旧有的单字符串配置？（是，已要求兼容）。

## Open Questions
> [!WARNING]
> - 数组元素之间的执行环境状态（如 `cd` 或 `export` 变量）是否需要跨数组元素共享？（通常建议用 `&&` 连结或视作同一 session 执行）。

## Proposed Changes
### Configuration Parser
#### [MODIFY] [internal/config/config.go](file:///Users/snowdream/Workspace/snowdreamtech/UniRTM/internal/config/config.go)
- 添加 `StringArray` 类型解析器，支持从单字符串和字符串数组解析到 `[]string`。
- 修改 `PreInstall`, `PostInstall`, 和 `Task.Run` 字段为 `StringArray`。

### Execution Logic
#### [MODIFY] [internal/service/installation.go](file:///Users/snowdream/Workspace/snowdreamtech/UniRTM/internal/service/installation.go)
- 调用 `Script()` 方法组合数组元素执行。
#### [MODIFY] [internal/task/native.go](file:///Users/snowdream/Workspace/snowdreamtech/UniRTM/internal/task/native.go)
- 调用 `Script()` 方法处理 `run` 的数组指令。

## Verification Plan
### Automated Tests
- 更新 `config_test.go` 确保可以正确解析混合数组与单字符串。
- 更新 `installation_test.go` 验证执行顺序和兼容性。

### Manual Verification
- 使用带有数组形式 script 字段的配置文件执行安装流程。
