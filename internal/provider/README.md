# Provider System

This package implements tool-specific installation logic and shim generation.

## Responsibilities

- Define the Provider interface for tool-specific operations
- Implement generic provider for standard tools
- Implement Node.js provider with npm/npx support
- Implement Python provider with pip/virtualenv support
- Implement Go provider with GOPATH management
- Generate platform-specific shim scripts
- Handle post-installation hooks
- Detect installed tool versions

## Key Components

- `provider.go` - Provider interface definition
- `generic.go` - Generic provider implementation
- `node.go` - Node.js provider implementation
- `python.go` - Python provider implementation
- `go.go` - Go provider implementation
- `registry.go` - Provider registry and discovery
- `shim.go` - Shim generation utilities

## Usage

```go
import "github.com/snowdreamtech/unirtm/internal/provider"

// Register providers
registry := provider.NewRegistry()
registry.Register("node", provider.NewNodeProvider())
registry.Register("python", provider.NewPythonProvider())

// Use a provider
provider := registry.Get("node")
err := provider.Install(ctx, "/path/to/install", "20.0.0")
if err != nil {
    log.Fatal(err)
}
```

## Requirements

Implements requirements: 6.1-6.7, 14.1-14.7
