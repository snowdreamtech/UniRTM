# Implementation Tasks: Path Handling Refactoring

- [ ] 1. **Create `internal/pkg/envpath` package**
  - Implement `JoinForOS` to handle native OS execution (e.g. `cmd.Env`).
  - Implement `JoinForPosix` to handle Bash/Zsh POSIX scripting path assembly (uses `:` and forward slashes on Windows).
  - Implement `FormatDirForPosix` for single directory string adjustments.
  - Implement unit tests for `envpath` to ensure absolute correctness.

- [ ] 2. **Refactor Native OS Execution commands (`cmd` package)**
  - Update `cmd/25.run.go` to use `envpath.JoinForOS` where paths are natively injected.
  - Update `cmd/23.exec.go` to use `envpath.JoinForOS` and abstract away custom `deduplicatePathString` logic using `envpath` if possible, or just call `JoinForOS` instead of manual `os.PathListSeparator` joins.

- [ ] 3. **Refactor Shell Script Generation logic (`internal/service` and `cmd` package)**
  - Update `internal/service/activation.go` (generatePosixScript) to invoke `envpath.JoinForPosix` and `envpath.FormatDirForPosix`.
  - Update `internal/service/auto_activation.go` (generatePosixDeactivation) to invoke `envpath.FormatDirForPosix`.
  - Update `cmd/10.deactivate.go` (generatePosixDeactivationScript) to invoke `envpath.FormatDirForPosix`.
  - Update `cmd/3.env.go` (cmd.PrintEnv) to invoke `envpath.JoinForPosix` for posix shells, removing the inline Windows logic.

- [ ] 4. **Run Verification**
  - Run `go test ./...` across the entire codebase to ensure existing tests pass.
  - Perform Git commit atomically `feat: introduce envpath for centralized cross-platform path handling`.
