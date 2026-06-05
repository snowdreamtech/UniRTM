# Musl-Aware Binary Selection

## Background

UniRTM handles downloading pre-built binaries from various sources (GitHub Releases, Node.js distributions, etc.). Currently, the download scoring mechanism (`internal/backend/common.go:CalculateAssetScore`) unconditionally deducts 10 points from any asset filename containing `musl`.

While this was historically a sound heuristic trade-off to ensure standard Linux (glibc) environments did not accidentally pick up `musl` binaries due to identical base scores, it causes Alpine Linux (which *uses* `musl`) to erroneously prefer `glibc` binaries, leading to fatal execution errors. Furthermore, the lockfile's platform key generation (`internal/lockfile/platform.go`) hardcodes `musl=false`, breaking precise lockfile pinning on Alpine.

## Requirements

With the recent introduction of `sysinfo.IsMusl()`, UniRTM can confidently detect the runtime libc environment. The system should now use `IsMusl()` to become environment-aware when selecting binaries.

1.  **Platform Definition:** Add a `Musl` flag to the `backend.Platform` struct so that backends are aware of the runtime libc.
2.  **Scoring Logic Upgrade:**
    *   If running on Musl (`Platform.Musl == true`), assets containing `musl` should receive a positive bonus (`+50`), and missing `musl` should incur a minor penalty (`-10`), guaranteeing the `musl` package wins.
    *   If running on glibc (`Platform.Musl == false`), assets containing `musl` should be heavily penalized (`-50`), preserving the original safe behavior.
3.  **Lockfile Pinning:** `CurrentPlatformKey()` must dynamically inject `sysinfo.IsMusl()` so that lockfiles correctly record `linux-amd64-musl`.
4.  **Provider Specific Adjustments:**
    *   **Node.js:** If `MISE_NODE_FLAVOR` is not explicitly set, auto-detect using `sysinfo.IsMusl()`.
    *   **Rust:** Inject `-musl` targets dynamically if `sysinfo.IsMusl()` is true.
