# Tasks: Auto Update Notifier

- [ ] Task 1: Create `internal/updater/updater.go` with `UpdateCache` struct and file read/write logic.
- [ ] Task 2: Implement GitHub API release fetch logic inside `CheckUpdateAsync()` using a timeout.
- [ ] Task 3: Implement `PromptIfAvailable()` with blacklist checking and `isatty` checking for interactive terminal.
- [ ] Task 4: Hook `CheckUpdateAsync` and `PromptIfAvailable` into `cmd/1.main.go` via Cobra's `PersistentPreRun` and `PersistentPostRun`.
- [ ] Task 5: Add tests for `updater.go` covering cache reading/writing and version comparisons.
