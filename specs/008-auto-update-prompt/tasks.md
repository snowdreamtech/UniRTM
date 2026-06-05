# Tasks: Auto Update Notifier

- [x] Task 1: Create `internal/updater/updater.go` with `UpdateCache` struct and file read/write logic.
- [x] Task 2: Implement GitHub API release fetch logic inside `CheckUpdateAsync()` using a timeout.
- [x] Task 3: Implement `PromptIfAvailable()` with blacklist checking and `isatty` checking for interactive terminal.
- [x] Task 4: Hook `CheckUpdateAsync` and `PromptIfAvailable` into `cmd/1.main.go` via Cobra's `PersistentPreRun` and `PersistentPostRun`.
- [x] Task 5: Add tests for `updater.go` covering cache reading/writing and version comparisons.
