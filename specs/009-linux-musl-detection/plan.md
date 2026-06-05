# Implementation Plan: Linux libc Detection (musl vs glibc)

## Background
Go relies on `runtime.GOOS` for platform detection. On Linux, this defaults to `linux` without distinguishing between `glibc` (standard) and `musl` (Alpine Linux). This leads to runtime linking errors when deploying standard binaries on Alpine.

## Goal
Implement a runtime detection mechanism to distinguish between `musl` and `glibc` so that `UniRTM` can correctly handle or download specific musl-compiled artifacts.

## Proposed Changes

### `internal/sysinfo/env.go`
*   Add a new function `IsMusl() bool` or `GetLibc() string`.
*   **Logic**:
    1. Check for the existence of `/etc/alpine-release`. If it exists, return `true` (musl).
    2. Check common musl dynamic linker paths (e.g., `/lib/ld-musl-x86_64.so.1`, `/lib/ld-musl-aarch64.so.1`).
    3. (Optional) Run `ldd --version` and parse the output for the string "musl".

### `internal/sysinfo/env_test.go`
*   Add unit tests for `IsMusl()` mocking or validating the filesystem states.

### Musl Glibc Fallback Warning (Enhancement)
*   **Goal**: Warn users explicitly when downloading a `glibc` compiled binary on an Alpine (`musl`) system to prevent cryptic `not found` errors at runtime.
*   **`internal/backend/common.go`**:
    *   Add an `IsGlibcFallback bool` field to the `CommonAsset` struct.
    *   In `FindBestAsset`, if `platform.Musl` is true and the selected asset name does not contain "musl", set `IsGlibcFallback = true`.
*   **`internal/backend/github.go` (and other Generic Handlers)**:
    *   Map `IsGlibcFallback` to `VersionInfo.Metadata["IsGlibcFallback"] = "true"`.
*   **`internal/service/installation.go` (or wherever `backend.GetDownloadInfo` is handled)**:
    *   Check `Metadata["IsGlibcFallback"]`.
    *   If true, print a high-visibility compatibility warning to the user before or after installation, suggesting they run `apk add gcompat`.

## Verification
- Run unit tests: `go test ./internal/sysinfo -v`
- Run unit tests: `go test ./internal/backend -v`
- Compile and run manually on a local Linux container (e.g., using `docker run -rm -v $PWD:/app alpine:latest`).
