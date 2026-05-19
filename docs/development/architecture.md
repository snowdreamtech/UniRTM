# UniRTM Architecture & Developer Guide

## Overview

UniRTM is a universal runtime manager written in Go 1.21+. It follows a strict
layered architecture to ensure testability, extensibility, and clear separation of concerns.

## Architecture Layers

```
┌─────────────────────────────────────────────┐
│              CLI Layer (cmd/)               │
│    Cobra commands, flags, output formatting │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│           Service Layer (internal/service/) │
│  InstallationManager, CacheManager, etc.    │
└──────┬──────────────────────────┬───────────┘
       │                          │
┌──────▼──────┐          ┌────────▼──────────┐
│  Backend    │          │    Provider       │
│  (GitHub,   │          │  (Node, Python,   │
│   Aqua, HTTP│          │   Go, Generic)    │
└──────┬──────┘          └────────┬──────────┘
       │                          │
┌──────▼──────────────────────────▼──────────┐
│        Repository Layer                     │
│   SQLite implementations (WAL mode)        │
└─────────────────────────────────────────────┘
```

## Directory Structure

```
UniRTM/
├── cmd/                    # Cobra CLI commands
│   ├── 1.main.go           # Root command, global flags
│   ├── 6.install.go        # install command
│   ├── 7.uninstall.go      # uninstall command
│   └── ...                 # Other commands
├── internal/
│   ├── backend/            # Backend implementations
│   │   ├── github.go       # GitHub Releases backend
│   │   ├── aqua.go         # Aqua registry backend
│   │   ├── http.go         # Direct HTTP backend
│   │   └── registry.go     # Backend registry
│   ├── config/             # Configuration management
│   │   ├── config.go       # Data structures
│   │   └── manager.go      # ConfigManager (go-toml/yaml.v3-based)
│   ├── database/           # SQLite database layer
│   │   ├── database.go     # Connection management
│   │   ├── migration.go    # Schema migration
│   │   └── schema.go       # Table definitions
│   ├── provider/           # Tool-specific install logic
│   │   ├── node.go         # Node.js provider
│   │   ├── python.go       # Python provider
│   │   ├── go.go           # Go provider
│   │   ├── generic.go      # Fallback provider
│   │   └── registry.go     # Provider registry
│   ├── repository/         # Repository interfaces + SQLite impls
│   │   ├── repository.go   # Interface definitions
│   │   └── sqlite/         # SQLite implementations
│   ├── service/            # Business logic layer
│   │   ├── installation.go # Install workflow
│   │   ├── activation.go   # Shell activation
│   │   ├── cache.go        # Cache management
│   │   ├── migration.go    # mise/asdf migration
│   │   ├── performance.go  # Performance monitoring
│   │   ├── plugin.go       # Plugin system
│   │   ├── recovery.go     # Recovery & cleanup
│   │   ├── concurrent.go   # Parallel installs
│   │   ├── offline.go      # Offline mode
│   │   ├── security.go     # Checksum verification
│   │   ├── shim.go         # Shim generation
│   │   └── config_validator.go  # Semantic validation
│   ├── pkg/
│   │   ├── download/       # HTTP downloader with retry
│   │   ├── errors/         # Custom error types
│   │   └── logger/         # Structured logging (zerolog)
│   └── transaction/        # Transaction manager
├── tests/
│   ├── property/           # Property-based tests (gopter)
│   ├── integration/        # Integration tests
│   └── bench/              # Performance benchmarks
└── docs/                   # Documentation
```

## Development Setup

```bash
# Clone the repository
git clone https://github.com/snowdreamtech/unirtm
cd unirtm

# Install dependencies
go mod download

# Run all tests
make test

# Build
go build -o unirtm ./cmd/internal/

# Run linter
make lint
```

## Testing Strategy

| Test Type | Location | Purpose |
|-----------|----------|---------|
| Unit | `internal/**/*_test.go` | Package-level correctness |
| Property | `tests/property/` | Universal invariants (gopter) |
| Integration | `tests/integration/` | Cross-layer workflows |
| Benchmark | `tests/bench/` | Performance regression detection |

```bash
# Unit + integration tests
go test ./...

# Property tests (may take longer)
go test ./tests/property/... -timeout 120s

# Benchmarks
go test ./tests/bench/... -bench=. -benchmem
```

## Key Design Principles

1. **Idempotency** — All operations can be safely retried
2. **Atomicity** — Install either fully succeeds or leaves no traces
3. **Layered Architecture** — CLI → Service → Backend/Provider → Repository
4. **Repository Pattern** — All persistence through interfaces (SQLite by default)
5. **Context Propagation** — All operations accept `context.Context` for cancellation
6. **Structured Logging** — All log output uses zerolog with consistent fields

## Adding a New Backend

1. Implement `backend.Backend` interface in `internal/backend/`
2. Register in `backend.NewRegistry()` or as a plugin
3. Add tests in `internal/backend/your_backend_test.go`

## Adding a New Provider

1. Implement `provider.Provider` interface in `internal/provider/`
2. Register in `provider.NewRegistry()` with tool name(s)
3. Add tests

## Adding a New CLI Command

1. Create `cmd/N.command.go` (N = next number)
2. Define Cobra command with `Use`, `Short`, `Long`, `RunE`
3. Add to `rootCmd` in `cmd/0.cmd.go`
4. Respect `--dry-run`, `--json`, `--verbose`, `--quiet` global flags

## Error Handling

Use custom error types from `internal/pkg/errors`:

```go
import pkgerrors "github.com/snowdreamtech/unirtm/internal/pkg/errors"

// Return typed errors
return pkgerrors.NewNotFoundError("tool", toolName)
return pkgerrors.NewAlreadyExistsError("tool", toolName, version)
return pkgerrors.NewNetworkError(err)
```
