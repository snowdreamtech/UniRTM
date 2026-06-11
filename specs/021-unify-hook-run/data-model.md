# Data Model

This project is a CLI tool and relies entirely on runtime arguments and interface definitions.

## Core Interface

```go
type HookRunner interface {
 // Detect returns true if the engine is present in the repository
 Detect(ctx context.Context, pwd string) (bool, error)
 // InstallBridgeScript installs the bridge script
 InstallBridgeScript(ctx context.Context, pwd string, hookName string) error
 // InstallAllBridgeScripts installs bridge scripts to all non-sample hooks
 InstallAllBridgeScripts(ctx context.Context, pwd string) ([]string, error)
 // Run executes the specific hook or stage
 Run(ctx context.Context, pwd string, hookName string, stage string, args []string) error
}
```

The input context from the user executing `unirtm hook run`.

- **`HookName`** (string): The single positional argument (`all`, `shellcheck`, etc.).
- **`Stage`** (string): Provided via `--stage` flag (e.g., `pre-commit`). Defaults to empty string if not provided.
- **`Args`** ([]string): Any trailing arguments to be passed down.
