# Implementation Plan: Musl-Aware Binary Selection

## 1. Backend Core Update (`internal/backend/backend.go`)

- Update `CurrentPlatform()` to initialize `Musl: sysinfo.IsMusl()`.
- Update `(p Platform) String()` to append `-musl` when `p.Musl` is true.

## 2. Backend Scoring Update (`internal/backend/common.go` & `common_test.go`)

- In `CalculateAssetScore`, replace the hardcoded `-10` musl penalty with context-aware logic:
  - If `platform.Musl`: `+50` if asset has "musl", `-10` if not.
  - If `!platform.Musl`: `-50` if asset has "musl".
- Update tests in `common_test.go` to assert the new scoring logic works for both `musl: true` and `musl: false` scenarios.

## 3. Lockfile Fix (`internal/lockfile/platform.go`)

## 4. Provider Refinements

- **Node.js (`internal/provider/native/nodejs.go`)**: When checking `MISE_NODE_FLAVOR`, if it's empty, fallback to `sysinfo.IsMusl()` to automatically download `musl` flavor on Alpine.
- **Rust (`internal/provider/native/rust.go`)**: Detect `sysinfo.IsMusl()` and provide `-musl` compilation targets instead of only `gnu` for `linux`.

## 5. Verification

- Run `go test ./internal/backend/...` to ensure all tests (including new Musl tests) pass.
- Run `go test ./internal/lockfile/...` to verify lockfile integrity.
- Run `go test ./internal/provider/native/...` to verify provider tests still pass.
