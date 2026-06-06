# Mvdan-shell Integration Implementation Plan

## Goal

>
## User Review Required
>
>
> [!IMPORTANT]
> 请确认以下决策：
>
> - 是否所有由 `unirtm` 执行的 hook 脚本和任务（预设为 bash/sh）均交由 `mvdan-shell` 执行器处理？
>
> - 报错信息的抛出层级（若 shell 解析出错，是中断安装还是跳过 hook？建议中断安装）。
>

## Open Questions
>

> [!WARNING]

>

> - 原有依赖本机 `bash` 执行的系统特性（如特殊的 GNU ext）可能会有解析差异，是否需要回退选项？

## Proposed Changes

#### [MODIFY] `internal/task/native.go`

- 将基于 `exec.Command("sh", "-c", ...)` 的执行方式替换为调用 `mvdan.cc/sh/v3/interp`。

- 引入 `mvdan.cc/sh/v3` 依赖库。

## Verification Plan

### Automated Tests

- 在 `task_test.go` 中，编写包含复杂环境变量注入及管道操作的 bash 脚本测试。

### Manual Verification

- 编写包含 `&&`, `||`, 和循环的测试配置文件，并使用 `unirtm run` 测试执行是否与本机行为一致。
