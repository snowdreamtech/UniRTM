# UniRTM Plugin Development Guide

UniRTM supports **custom backends** and **custom providers** via Go shared-object plugins
compiled with `-buildmode=plugin`. This document describes the stable API contract
(PluginAPIVersion `"1"`) and provides worked examples.

---

## Plugin Types

| Type | What it does | Required export |
|------|-------------|----------------|
| **Backend** | Resolves download URLs and checksums for a tool version | `NewBackend() backend.Backend` |
| **Provider** | Handles tool-specific install/uninstall/shim logic | `NewProvider() provider.Provider` |

---

## PluginMetadata Export

Every plugin **must** export a `PluginMetadata` variable:

```go
package main

import "github.com/snowdreamtech/unirtm/internal/service"

var Metadata = service.PluginMetadata{
    Name:       "my-custom-backend",
    Version:    "0.1.0",
    APIVersion: "1",           // Must match PluginAPIVersion constant
    Type:       service.PluginTypeBackend,
    Author:     "Your Name",
    Description: "A custom backend for <tool>",
}
```

---

## Implementing a Custom Backend

### Step 1 — Implement the `backend.Backend` interface

```go
package main

import (
    "context"
    "fmt"

    "github.com/snowdreamtech/unirtm/internal/backend"
)

// MyBackend implements backend.Backend.
type MyBackend struct{}

func (b *MyBackend) Name() string { return "my-backend" }

func (b *MyBackend) ListVersions(ctx context.Context, tool string) ([]string, error) {
    // Return available versions for the tool
    return []string{"1.0.0", "2.0.0"}, nil
}

func (b *MyBackend) ResolveVersion(ctx context.Context, tool, version string) (string, error) {
    if version == "latest" {
        return "2.0.0", nil
    }
    return version, nil
}

func (b *MyBackend) GetDownloadInfo(
    ctx context.Context,
    tool, version string,
    platform backend.Platform,
) (*backend.VersionInfo, error) {
    url := fmt.Sprintf("https://example.com/%s/%s/%s-%s.tar.gz",
        tool, version, platform.OS, platform.Arch)
    return &backend.VersionInfo{
        DownloadURL: url,
        Checksum:    "", // Provide SHA-256 checksum when available
    }, nil
}
```

### Step 2 — Export `NewBackend`

```go
// NewBackend is the plugin entry-point. UniRTM calls this to create the backend.
func NewBackend() backend.Backend {
    return &MyBackend{}
}
```

### Step 3 — Build the plugin

```bash
go build -buildmode=plugin -o ~/.config/unirtm/plugins/my-backend.so ./my-backend/
```

> **Important:** The plugin **must** be compiled with the same Go version and module
> dependencies as UniRTM. Use `go version` and `go list -m all` to verify compatibility.

---

## Implementing a Custom Provider

### Step 1 — Implement the `provider.Provider` interface

```go
package main

import (
    "context"
    "os"
    "path/filepath"

    "github.com/snowdreamtech/unirtm/internal/provider"
)

// MyProvider implements provider.Provider.
type MyProvider struct{}

func (p *MyProvider) Install(ctx context.Context, installPath, archivePath, version string) error {
    // Extract archive and place binaries in installPath
    return os.MkdirAll(filepath.Join(installPath, "bin"), 0755)
}

func (p *MyProvider) Uninstall(ctx context.Context, installPath string) error {
    return os.RemoveAll(installPath)
}

func (p *MyProvider) PostInstall(ctx context.Context, installPath, version string) error {
    // Run post-install hooks (e.g., npm install, pip install)
    return nil
}

func (p *MyProvider) GenerateShim(ctx context.Context, tool, installPath string) (string, error) {
    // Return shim script content for this tool
    return "#!/bin/sh\nexec " + filepath.Join(installPath, "bin", tool) + " \"$@\"\n", nil
}

func (p *MyProvider) DetectVersion(ctx context.Context, installPath string) (string, error) {
    // Read version from install directory
    return "1.0.0", nil
}

func (p *MyProvider) BinaryNames() []string {
    // Return all binary names this provider manages
    return []string{"my-tool"}
}
```

### Step 2 — Export `NewProvider` and `Metadata`

```go
var Metadata = service.PluginMetadata{
    Name:        "my-tool",    // Matches the tool name in unirtm.toml
    APIVersion:  "1",
    Type:        service.PluginTypeProvider,
    Description: "Provider for my-tool",
}

func NewProvider() provider.Provider { return &MyProvider{} }
```

### Step 3 — Build and install

```bash
go build -buildmode=plugin -o ~/.config/unirtm/plugins/my-tool.so ./my-tool-provider/
```

---

## Plugin Directory

Plugins are loaded from (in order):

1. `$UNIRTM_PLUGIN_DIR` (environment variable)
2. `~/.config/unirtm/plugins/`
3. `/etc/unirtm/plugins/` (system-wide)

---

## API Stability

The **PluginAPIVersion** constant in `internal/service/plugin.go` is `"1"`.
UniRTM will **refuse to load** plugins with a mismatched API version and log a warning.

Breaking changes to the API will increment this version. Plugins must be recompiled
when upgrading across major API versions.

---

## Error Isolation

If a plugin panics during loading or execution, UniRTM catches the panic and logs
an error. Other plugins and core functionality continue to work normally.

---

## Testing Your Plugin

```go
package main_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestMyBackend_ListVersions(t *testing.T) {
    b := NewBackend()
    versions, err := b.ListVersions(context.Background(), "my-tool")
    require.NoError(t, err)
    assert.NotEmpty(t, versions)
}
```

Run with standard `go test ./...`.
