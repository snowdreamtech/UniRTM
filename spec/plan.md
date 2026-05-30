# Cross-Platform Path Handling Refactoring Plan

## 1. Background
In a cross-platform environment, environment variable paths (`PATH`) have vastly different formatting and separation rules depending on the execution context:
1. **OS API Context (Go `exec`, `os.Setenv`)**: Requires native path syntax and native `os.PathListSeparator` (`;` on Windows, `:` on Linux/macOS).
2. **POSIX Shell Context (Bash, Zsh)**: Strictly requires `:` as a list separator, even on Windows (e.g., Git Bash / MSYS2). Furthermore, injecting Windows paths with backslashes (`\`) into shell scripts often triggers escaping bugs (e.g., in `sed` or `case` matching).
3. **PowerShell Context**: Uses `os.PathListSeparator` natively, but requires special array splitting via `-split`.

Currently, this logic is scattered across `cmd/3.env.go`, `cmd/10.deactivate.go`, `cmd/23.exec.go`, `cmd/25.run.go`, `internal/service/activation.go`, and `internal/service/auto_activation.go` with inline `if runtime.GOOS == "windows"` checks.

## 2. Proposed Changes

### A. Introduce `pkg/envpath` package
Create a new utility package `internal/pkg/envpath` to abstract all formatting logic based on the context:

```go
package envpath

// JoinForOS APIs natively (e.g., os.Setenv, exec.Command)
func JoinForOS(paths []string) string

// JoinForPosix processes paths for bash/zsh scripts. (forces `:` and forward slashes)
func JoinForPosix(paths []string) string

// FormatDirForPosix formats a single directory (e.g., shimsDir) for bash/zsh.
func FormatDirForPosix(dir string) string

// JoinForPowerShell processes paths for pwsh/powershell scripts.
func JoinForPowerShell(paths []string) string
```

### B. Refactor Usage Sites
Refactor the scattered logic to depend purely on this semantic context instead of `runtime.GOOS` logic.

1. **Native OS execution paths**:
   - `cmd/25.run.go`: Use `envpath.JoinForOS` instead of manual splitting and joining with `os.PathListSeparator`.
   - `cmd/23.exec.go`: Use `envpath.JoinForOS` for setting environment variables.
2. **Script Generation (POSIX)**:
   - `internal/service/activation.go`: Replace inline loop with `envpath.JoinForPosix(config.InjectedPaths)` and `envpath.FormatDirForPosix(config.ShimsDir)`.
   - `internal/service/auto_activation.go`: Replace inline loop with `envpath.FormatDirForPosix` when cleaning paths.
   - `cmd/10.deactivate.go`: Use `envpath.FormatDirForPosix` for POSIX deactivation script.
   - `cmd/3.env.go`: Use `envpath.JoinForPosix` for POSIX shell `export PATH`.

## 3. Verification Plan
1. Ensure all `go test ./...` unit tests pass seamlessly.
2. Run GitHub Actions CI to confirm `Pre-flight Integrity Check` succeeds, particularly on `windows-latest`.
3. Assert that no manual `filepath.ToSlash` or `strings.Join(..., string(os.PathListSeparator))` logic exists directly inside `service/` or `cmd/` for PATH manipulation anymore.
