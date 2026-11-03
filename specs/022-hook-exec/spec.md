# Feature Specification: UniRTM Hook Execution Wrapper

**Feature Branch**: `022-hook-exec`

**Created**: 2026-06-12

**Status**: Draft

**Input**: User description: "新增专用的 unirtm hook-exec 命令（最推荐的终极方案）既然 unirtm 旨在统一接管项目的工具链，不如我们在 unirtm 内部提供一个专门为 Hook 定制的执行包装器。做法：在 UniRTM 源码中添加一个新的子命令，比如 unirtm hook-exec <tool> <args>..."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Simplified Pre-commit Configuration (Priority: P1)

As a developer configuring pre-commit hooks, I want to use a clean `unirtm hook-exec <tool> <args>` entry in `.pre-commit-config.yaml` without needing complex shell wrappers like `bash -c` and `xargs`, so that my configuration is clean, readable, and less error-prone.

**Why this priority**: It is the primary purpose of this feature, improving developer experience and preventing misconfiguration.

**Independent Test**: Can be fully tested by configuring a hook with `unirtm hook-exec` and verifying it runs the target tool correctly.

**Acceptance Scenarios**:

1. **Given** a `.pre-commit-config.yaml` using `unirtm hook-exec prettier --write`, **When** the hook is executed with a small number of files, **Then** the tool is invoked exactly once with all files.

---

### User Story 2 - Automatic Argument Chunking on Windows (Priority: P1)

As a Windows user running pre-commit on a large repository, I want `unirtm hook-exec` to automatically chunk large lists of files so that I do not encounter the "Command line is too long" error from `cmd.exe` limits (8191 characters).

**Why this priority**: This solves the critical cross-platform compatibility bug that currently breaks Windows environments for Node.js-based tools.

**Independent Test**: Can be fully tested by passing over 8000 characters of arguments to `unirtm hook-exec` on Windows and verifying multiple sub-processes are spawned successfully without errors.

**Acceptance Scenarios**:

1. **Given** a Windows environment, **When** `unirtm hook-exec` is called with file arguments totaling > 7000 characters, **Then** it automatically chunks the files and runs the underlying tool multiple times, returning success only if all chunks succeed.
2. **Given** a Linux/macOS environment, **When** `unirtm hook-exec` is called with file arguments totaling > 7000 characters, **Then** it executes the underlying tool in a single batch (or respects native OS limits without strict 7000 char splitting).

### Edge Cases

- What happens when one of the chunked executions fails? The command must exit with a non-zero status code and ideally not execute the remaining chunks (fail-fast).
- How does the system handle commands that do not take files as arguments? (Pre-commit generally passes files at the end, which aligns perfectly with chunking).
- What happens if the `os.Args` limit is reached without any file arguments (i.e. just configuration flags)?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a new command `unirtm hook-exec <tool> [args...]`.
- **FR-002**: System MUST detect the host operating system.
- **FR-003**: System MUST calculate the length of the command line arguments.
- **FR-004**: On Windows, if the total command line length approaches 7000 characters, System MUST chunk trailing file arguments into multiple batches and execute the underlying `<tool>` sequentially for each batch.
- **FR-005**: System MUST aggregate the exit codes of chunked executions, failing immediately if any batch fails.
- **FR-006**: System MUST pass standard IO streams (stdout/stderr) from the tool directly to the console.

### Key Entities

- **Hook Command Execution**: The wrapper that intercepts the command, analyzes arguments, and spawns one or more child processes.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Pre-commit hook configuration drops `bash -c` and `xargs` completely for all file-processing hooks.
- **SC-002**: Passing 10,000 files to `unirtm hook-exec prettier` on Windows successfully processes all files without "Command line is too long" error.
- **SC-003**: Zero noticeable performance overhead compared to raw execution for small file sets.

## Assumptions

- Pre-commit will always append file paths at the very end of the command arguments.
- Tools invoked via `hook-exec` are capable of being invoked multiple times sequentially with disjoint sets of files (idempotent and additive).
- 7000 characters is a safe threshold to avoid the 8191 limit of `cmd.exe` including the base command and environmental overhead.
