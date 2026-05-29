# Feature Specification: Built-in Cross-Platform Shell (mvdan-shell)

**Feature Branch**: `003-mvdan-shell`
**Created**: 2026-05-28
**Status**: Draft
**Input**: User description: "业界还有一个非常优雅的“终极跨平台 Shell”解决方案： 内置一个纯 Go 语言实现的跨平台 Shell 解释器（例如 mvdan.cc/sh） ，作为feature，在spec文件夹创建规范，并逐步实施"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Seamless POSIX Execution on Windows (Priority: P1)

作为一名 Windows 用户，我希望在没有任何 Bash 环境（未安装 Git Bash、WSL 等）的纯净系统下，直接运行 `unirtm run <task>`，并且任务中哪怕使用了 `if [ -f "go.mod" ]; then` 这种 POSIX 语法的脚本，也能被成功执行而不会报错崩溃。

**Why this priority**: 解决 `.unirtm.toml` 跨平台兼容性的核心痛点，抹平操作系统底层的执行环境差异，真正实现一次配置、全平台运行。

**Independent Test**: Can be fully tested by running a task containing POSIX if-statements and environment variable expansions (`$VAR`) on a Windows system (or via a unit test without invoking `/bin/sh`).

**Acceptance Scenarios**:

1. **Given** 任务 `run` 配置为一段纯 POSIX Shell 脚本，**When** 用户在不包含 `sh` 环境变量的 Windows CMD 下执行 `unirtm run`，**Then** 脚本正确解析执行，无任何语法错误抛出。
2. **Given** 脚本中使用了 `$UNIRTM_FIX` 等环境变量，**When** 执行脚本，**Then** 解释器能正确捕获并展开当前执行上下文中的环境变量。

---

### User Story 2 - Consistent Output Streaming (Priority: P2)

作为持续集成系统（CI）或终端用户，我希望通过内置 Shell 运行命令时，它的标准输出（stdout）和错误输出（stderr）能像原生执行一样，被实时无损地推流打印到控制台，不丢失 ANSI 高亮颜色，并完美支持前缀打印（Prefix Writer）。

**Why this priority**: 任务执行过程中的日志反馈至关重要，特别是 `pre-commit` 这类强依赖终端色彩和实时滚动的工具。

**Independent Test**: Can be tested by invoking a long-running sub-command that emits stdout incrementally and verifying the output appears in real-time.

**Acceptance Scenarios**:

1. **Given** 任务命令为 `echo "test" && sleep 1 && echo "done"`，**When** 运行任务，**Then** 先输出 `test`，停顿 1 秒后输出 `done`。
2. **Given** 环境变量 `UNIRTM_TASK_OUTPUT=interleaved` 被设置，**When** 并发运行任务，**Then** Shell 输出应被正确加上前缀并拦截打印。

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST 解析并执行遵循 POSIX Shell 标准的脚本内容，作为 `.unirtm.toml` 任务的默认执行引擎。
- **FR-002**: System MUST 引入 `mvdan.cc/sh/v3` 作为内置解析器与解释器，彻底替代基于 `os/exec` 的 `sh -c` 及 `cmd.exe /c` 回退机制。
- **FR-003**: System MUST 将当前进程的所有环境变量以及任务层定义的额外环境变量安全地注入到内置解释器的上下文中。
- **FR-004**: System MUST 将解释器的标准输入（Stdin）、标准输出（Stdout）和错误输出（Stderr）绑定到现有的流处理器上（兼容当前的 `prefixWriter`）。
- **FR-005**: System MUST 正确捕捉脚本执行完毕的退出码（Exit Code），如果非零，则需将失败状态向上传递以中断执行流。
- **FR-006**: System MUST 保持对 `go-task` 解析器（`go_task.go`）的兼容性不造成破坏性修改。

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 将此前因为跨平台问题替换为 Python 的 `.unirtm.toml` 任务（如 `audit:zizmor`、`audit:npm`）恢复为原版的 POSIX Shell 脚本语法（如 `if [ -f "..." ]`）后，Windows CI 能够 100% 成功通过。
- **SC-002**: Windows 下的 `unirtm run` 任务执行过程完全不再触发寻找 `cmd.exe` 作为 Fallback 解释器的逻辑。
- **SC-003**: 内部 `native.go` 引擎通过所有现有的测试套件（执行速度衰减在可接受范围 < 10%）。
