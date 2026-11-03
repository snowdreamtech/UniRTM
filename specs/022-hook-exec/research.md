# Research: UniRTM Hook Execution Wrapper

## Decision: Chunking Argument Parsing Algorithm

- **Decision**: Scan arguments from right-to-left using `os.Lstat` to dynamically partition file arguments from base arguments.
- **Rationale**: `pre-commit` appends file arguments to the end of the entry command. Since file paths provided by git hooks exist on disk (as they are tracked/modified files), checking their existence allows reliable splitting without needing to build complex parsers or configure specific delimiters.
- **Alternatives considered**:
  - Using a hardcoded `--` delimiter: Not supported natively by standard pre-commit without complex bash wrappers (which defeats the purpose of this feature).
  - Whitelisting tool flags: Unmaintainable due to the sheer number of possible hooks and flags.

## Decision: Command Execution Engine reuse

- **Decision**: Reuse `runExec(cmd, args)` (which handles standard `unirtm exec` logic) directly for each chunk.
- **Rationale**: `runExec` correctly parses tool dependencies, ensures tools are installed, correctly modifies `PATH`, and sets tool-specific environment variables. Since it encapsulates `execWindows` which delegates exit codes to `os.Exit` internally upon failure, it automatically provides the required "fail-fast" behavior when multiple chunks are evaluated sequentially.
- **Alternatives considered**:
  - Re-implementing raw `os.Exec` calls inside `hook-exec`: Duplicates environment injection and dependency resolution code.

## Decision: Safe Command Length Threshold

- **Decision**: 7000 characters.
- **Rationale**: The hard limit in Windows `cmd.exe` is 8191 characters. Setting the threshold at 7000 provides a comfortable buffer of ~1191 characters for environmental overrides, base command overhead, and quote escaping added by Go's `os/exec` under the hood.
