# Tasks: Musl-Aware Binary Selection

- [x] 1. Update `Platform` struct in `internal/backend/backend.go`.
- [x] 2. Update `CalculateAssetScore` in `internal/backend/common.go` to use context-aware musl scoring.
- [x] 3. Update unit tests in `internal/backend/common_test.go` to verify the new scoring logic.
- [x] 4. Commit: "feat(backend): add context-aware musl asset scoring"
- [x] 5. Update `CurrentPlatformKey` in `internal/lockfile/platform.go` to use `sysinfo.IsMusl()`.
- [x] 6. Commit: "fix(lockfile): use IsMusl for CurrentPlatformKey"
- [x] 7. Update Node.js provider in `internal/provider/native/nodejs.go` to fallback to `sysinfo.IsMusl()`.
- [x] 8. Update Rust provider in `internal/provider/native/rust.go` to inject `-musl` targets based on `sysinfo.IsMusl()`.
- [x] 9. Commit: "feat(provider): enable auto-detection for musl environments in nodejs and rust"
- [x] 10. Run `go test ./...` and `golangci-lint run` (or `make lint`) to verify codebase integrity.
