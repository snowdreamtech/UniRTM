# Quickstart & Validation: UniRTM Hook Execution Wrapper

## Validation Scenarios

### Scenario 1: Short Command Execution

Verify that small commands pass through transparently.

```bash
./unirtm hook-exec echo "Hello World"
# Expected: "Hello World"
```

### Scenario 2: Right-to-Left File Splitting

Verify that the splitting logic accurately identifies files vs base command flags.

1. Create a dummy file: `touch dummy.txt`
2. Run: `./unirtm hook-exec echo --flag dummy.txt`
3. Expected: Should successfully echo `--flag dummy.txt`. The `os.Lstat` algorithm should correctly split `echo --flag` as base args, and `dummy.txt` as file args.

### Scenario 3: Chunking on Windows

Verify chunking behavior using a very long command.

1. Build `unirtm` for Windows or test natively on Windows.
2. Generate 1000 dummy files or dummy arguments representing files that exist on disk.
3. Pass them to `unirtm hook-exec`.
4. Add debug logs locally in `hook-exec.go` to print `baseArgs` and each `currentChunk` size before it hits the `7000` limit.
5. Verify that `echo` (or equivalent) is executed multiple times, batching the files correctly without throwing a `Command line is too long` OS error.

### Scenario 4: .pre-commit-config.yaml integration

1. Update `.pre-commit-config.yaml`:

```yaml
      entry: unirtm hook-exec prettier --write
```

1. Run `./unirtm hook run pre-commit --all-files` on Windows.
2. Verify that `prettier` completes successfully across the entire repository.
