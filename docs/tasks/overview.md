# Task Runner Overview

UniRTM includes a built-in task runner designed to replace standard toolchain wrappers like `make`, `npm scripts`, or `just`. 

By integrating the task runner directly into the tool manager, UniRTM guarantees that tasks execute within the exact environment and toolchain defined by your `.unirtm.toml`, solving the "it works on my machine" problem permanently.

## Why use UniRTM's Task Runner?

- **Zero dependencies**: No need to install `make` on Windows or ensure `npm` is present if you are running a Python project.
- **Polyglot Execution**: Effortlessly mix Go, Python, Bash, and Node tasks within the same project.
- **Dependency Graph**: Tasks can declare dependencies. UniRTM will ensure dependencies run in topological order before executing your target task.
- **Inherited Context**: Tasks automatically inherit the environment variables from the `[env]` section and the `$PATH` overrides from the `[tools]` section of your `.unirtm.toml`.

## Quick Start

Tasks are defined under the `[tasks]` key in your `.unirtm.toml`.

```toml
[tasks.lint]
description = "Lint the codebase"
run = "golangci-lint run ./..."

[tasks.test]
description = "Run unit tests"
run = "go test ./..."
depends = ["lint"]
```

To run a task, use the `unirtm run` command:

```bash
$ unirtm run test
→ running task "lint"
✓ lint completed
→ running task "test"
✓ test completed
```

## Running Multiple Tasks

You can run multiple tasks sequentially by providing multiple arguments:

```bash
$ unirtm run lint test build
```

If a task fails, the entire execution chain halts immediately with a non-zero exit code.

## Exploring Project Tasks

New developers joining a project don't need to read complex `Makefiles` to figure out what commands are available. They can simply run:

```bash
$ unirtm run
Available tasks:
  build      Compile the production binary
  lint       Lint the codebase
  test       Run unit tests
  dev        Start the local dev server
```

For more advanced workflows, see [TOML Tasks](./toml-tasks.md) and [File Tasks](./file-tasks.md).
