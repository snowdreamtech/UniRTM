# Running Tasks

Once your tasks are configured via `.unirtm.toml` or the `.unirtm/tasks` directory, you can unleash their power using the `unirtm run` command.

## Basic Usage

```bash
# Run a single task
unirtm run build

# Run multiple tasks (sequentially or in parallel, depending on the DAG)
unirtm run lint test build
```

If you execute `unirtm run` without any arguments, it smartly parses all available tasks and presents an interactive list alongside their short descriptions (similar to a dynamic `make help`).

## Parallel Execution & Dependency Resolution (DAG)

UniRTM features an internal concurrent Directed Acyclic Graph (DAG) engine. If you run a task `C` that defines prerequisites `depends_on = ["A", "B"]`:

1. UniRTM recursively checks if A and B have their own dependencies.
2. UniRTM executes A and B **simultaneously in parallel**.
3. Task C will only begin execution once both A and B return successfully with an exit code of `0`.
4. If either A or B fails, the entire execution chain is immediately aborted with an error.

This architecture allows you to squeeze maximum performance out of your multi-core CPU!

## Terminal UI & Pterm Integration

UniRTM utilizes a high-performance terminal UI library (Pterm) to manage task output.

- **Stream Isolation**: During parallel execution, the output of different tasks is smartly isolated and prefixed (e.g., `[build] ...`, `[test] ...`) to prevent log interleaving and race conditions on stdout.
- **Dynamic Status Bar**: A dynamic progress indicator is displayed at the bottom, providing real-time visibility into which tasks are currently Running and which are Done.
- **Plaintext Degradation**: When a CI environment (like GitHub Actions) is detected, UniRTM automatically disables the dynamic UI and colors, degrading gracefully to a plaintext streaming output to ensure log file readability.

## Advanced Run Flags

### 1. Passing Arguments

If you need to pass additional command-line arguments to the underlying script, use `--` after the task name:

```bash
# The --fix argument will be passed directly to the lint script
unirtm run lint -- --fix
```

### 2. Dry-run Mode

When debugging a complex task graph, you might only want to verify the execution order without actually running dangerous commands. You can utilize the dry-run mode:

```bash
unirtm run deploy --dry-run
```

This prints the sequentially ordered execution plan alongside the actual shell commands that would be run.

### 3. Concurrency Limits

By default, UniRTM determines the maximum number of parallel threads based on your machine's CPU core count. You can forcefully override this behavior:

```bash
unirtm run build --jobs 2
```
