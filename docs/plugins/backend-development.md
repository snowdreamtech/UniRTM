# Backend Development

UniRTM's core philosophy centers on **extreme performance and native cross-platform compatibility**. Traditional Bash-based tool managers (like asdf) are notoriously slow and often plagued with OS-specific compatibility issues during parsing, downloading, and compiling phases.

If you need a tool that isn't on the officially supported list, you can write your own **Backend Plugin** for UniRTM.

## Architecture Design

UniRTM plugins are not built on Bash scripts; they are **natively compiled modules**. This allows them to run at native speeds on Windows, macOS, and Linux, leveraging multi-threading capabilities for downloads and extractions.

Currently, the primary way to develop a backend plugin is by writing a **Go Interface**.

## The Core Interface (`ToolBackend`)

Every custom tool backend needs to implement the `ToolBackend` interface exposed by UniRTM:

```go
type ToolBackend interface {
    // The name of the plugin, e.g., "nodejs"
    Name() string

    // List all available versions (used for resolving `latest`)
    ListRemoteVersions() ([]string, error)

    // Resolve and populate download metadata, including URL and checksum
    ResolveVersion(version string) (*ResolvedTool, error)

    // Download the corresponding distribution package
    Download(resolved *ResolvedTool, destPath string) error

    // Extract and execute the installation logic
    Install(downloadedPath string, installDir string) error
}
```

### 1. `ListRemoteVersions`

Typically, you make HTTP requests to APIs (like GitHub Releases, Node.js Dist) here, fetch the release tags, and sort them using Semantic Versioning (SemVer).

### 2. `ResolveVersion`

Once the user specifies a version (e.g., `20.11.0`), you construct the exact download URL based on the user's OS (`GOOS`) and Architecture (`GOARCH`).
If the official provider offers checksums, you should fetch the `sha256` hash and return it here, as it will be used to populate the `unirtm.lock` file.

### 3. `Download`

UniRTM provides a built-in, high-performance `Downloader` module that supports resumable downloads and progress bars. In most cases, you simply delegate the work: `unirtm.DefaultDownloader.Download(url, destPath)`.

### 4. `Install`

Extract the downloaded archive. If it's a binary distribution, simply extract it to the target directory and set up `bin/` symlinks. If it's a source tarball, you would execute compilation commands like `make && make install` here.

## Registering Your Plugin

Once your backend is developed, you register it in the main entry point:

```go
package main

import (
    "github.com/snowdreamtech/UniRTM/core"
    "github.com/my-org/unirtm-plugin-rust"
)

func main() {
    core.RegisterBackend(rust.NewRustBackend())
    core.Execute()
}
```

> **Note:** In the future, we will support an isolated WebAssembly (WASM) plugin system. This will allow developers to write high-performance, sandboxed plugins using Rust, Zig, or TypeScript, without needing to recompile the main UniRTM binary. Stay tuned!
