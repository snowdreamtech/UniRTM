# Backend System

This package implements pluggable backend systems for tool installation from multiple sources.

## Responsibilities

- Define the Backend interface for tool sources
- Implement GitHub Releases backend
- Implement Aqua registry backend
- Implement HTTP download backend
- Manage backend registry and discovery
- Handle version listing and resolution
- Coordinate artifact downloads

## Key Components

- `backend.go` - Backend interface definition
- `github.go` - GitHub Releases backend implementation
- `aqua.go` - Aqua registry backend implementation
- `http.go` - HTTP download backend implementation
- `registry.go` - Backend registry and discovery

## Usage

```go
import "github.com/snowdreamtech/unirtm/internal/backend"

// Register backends
registry := backend.NewRegistry()
registry.Register("github", backend.NewGitHubBackend())
registry.Register("aqua", backend.NewAquaBackend())

// Use a backend
backend := registry.Get("github")
versions, err := backend.ListVersions(ctx, "nodejs/node")
if err != nil {
    log.Fatal(err)
}
```

## Requirements

Implements requirements: 5.1-5.8
