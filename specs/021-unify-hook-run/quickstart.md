# Quickstart: Validating `unirtm hook run` Execution

This guide demonstrates how to validate the argument refactoring.

## Prerequisites

- A Go development environment.
- UniRTM repository (`/Users/snowdream/Workspace/snowdreamtech/UniRTM`).

## Build and Run Validation

### 1. Build the binary

```bash
go build -o unirtm main.go
```

### 2. Validate Argument Boundary Rejections

Test that 0 arguments fail:

```bash
./unirtm hook run
# Expected: Error: requires exactly 1 positional arg(s)
```

### 3. Validate Engine Mapping (Example: pre-commit)

Assuming `pre-commit` is installed and detected:

#### 1 Argument (hookname)

```bash
./unirtm hook run shellcheck
# Expected: Runs `pre-commit run shellcheck -- `
```

#### Hookname + Stage flag

```bash
./unirtm hook run shellcheck --stage pre-commit
# Expected: Runs `pre-commit run shellcheck --hook-stage pre-commit -- `
```

#### Hookname (all) + Stage flag + Trailing Args

```bash
./unirtm hook run all --stage commit-msg .git/COMMIT_EDITMSG
# Expected: Runs `pre-commit run --hook-stage commit-msg -- .git/COMMIT_EDITMSG`
```
