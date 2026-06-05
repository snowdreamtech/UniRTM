# Implementation Plan: Auto Update Notifier

**Branch**: `008-auto-update-prompt` | **Date**: 2026-06-05 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/008-auto-update-prompt/spec.md`

## Summary
Introduce an automated, non-intrusive update prompt by implementing a background updater in `internal/updater` and hooking it into Cobra's `PersistentPreRun` (for async check) and `PersistentPostRun` (for prompt).

## Technical Context
**Language/Version**: Go 1.21+
**Primary Dependencies**: `github.com/spf13/cobra`, `github.com/mattn/go-isatty` (indirect via pterm)
**Storage**: `update-cache.json` in user's data directory.
**Testing**: Go unit tests (for version parsing and updater logic).
**Target Platform**: Linux, macOS, Windows (interactive terminal).

## Constitution Check
- Must not break declarative config.
- Must ensure atomic operation: Background goroutine avoids blocking CLI execution.
- Silent mode (`--silent` or `--quiet`) must suppress the update prompt.

## Project Structure

### Documentation (this feature)
```text
specs/008-auto-update-prompt/
├── spec.md
├── plan.md
└── tasks.md
```

### Source Code
```text
internal/
├── updater/
│   ├── updater.go          # Core logic for caching and prompting
│   └── updater_test.go     # Tests
cmd/
├── 1.main.go               # Cobra hooks (PersistentPreRun/PostRun)
```

## Technical Design

### `internal/updater/updater.go`
1. **State Management**:
   ```go
   type UpdateCache struct {
       LatestVersion string    `json:"latest_version"`
       LastChecked   time.Time `json:"last_checked"`
       LastPrompted  time.Time `json:"last_prompted"`
   }
   ```
2. **`CheckUpdateAsync(currentVersion string)`**:
   - Non-blocking goroutine to check latest version via GitHub API if 24 hours have passed since `LastChecked`.
   - Uses HTTP client with a strict 3-second timeout.
3. **`PromptIfAvailable(currentVersion string, cmdName string)`**:
   - Returns early if command is in the blacklist (`env`, `completion`, `version`, `self-update`).
   - Returns early if `isatty.IsTerminal(os.Stderr.Fd())` is false.
   - If `LatestVersion > currentVersion` and 24 hours since `LastPrompted`:
     - Print `WARN` to stderr.
     - Update `LastPrompted` in cache.

### `cmd/1.main.go`
1. In `setupGlobalOptions` (or `PersistentPreRun`), call `updater.CheckUpdateAsync(env.GitTag)`.
2. Add a `PersistentPostRun` to `rootCmd` that calls `updater.PromptIfAvailable(env.GitTag, cmd.Name())`.
