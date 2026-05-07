# UniRTM Project Structure

This document describes the directory structure and organization of the UniRTM project.

## Overview

UniRTM follows the [Standard Go Project Layout](https://github.com/golang-standards/project-layout) conventions, inspired by the [tott project](https://github.com/evilmartians/tott) for code style and organization.

## Directory Structure

```
UniRTM/
├── cmd/                    # Command-line interface entry points
│   ├── internal/          # Internal CLI implementation
│   └── *.go               # Cobra command definitions
│
├── internal/              # Private application code
│   ├── config/           # Configuration management (Viper, TOML/YAML)
│   ├── service/          # Business logic layer
│   ├── backend/          # Backend system (GitHub, Aqua, HTTP)
│   ├── provider/         # Provider system (tool-specific logic)
│   ├── repository/       # Data access layer (SQLite)
│   └── pkg/              # Shared internal packages
│       ├── logger/       # Zerolog-based logging
│       └── env/          # Environment and build metadata
│
├── tests/                # Test files
│   ├── unit/            # Unit tests
│   ├── integration/     # Integration tests
│   └── fixtures/        # Test fixtures and data
│
├── docs/                 # Documentation
│   ├── development/     # Developer documentation
│   ├── guide/           # User guides
│   ├── reference/       # API reference
│   └── adr/             # Architecture Decision Records
│
├── scripts/              # Build and utility scripts
│   └── lib/             # Shared script libraries
│
├── .github/              # GitHub configuration
│   └── workflows/       # GitHub Actions CI/CD
│
├── .kiro/                # Kiro AI IDE configuration
│   └── specs/           # Feature specifications
│       └── unirtm/      # UniRTM spec
│
├── go.mod                # Go module definition
├── go.sum                # Go module checksums
├── Makefile              # Build automation
├── .goreleaser.yaml      # Release configuration
└── .golangci.yml         # Linter configuration
```

## Layer Responsibilities

### CLI Layer (`cmd/`)

- Command-line interface using Cobra framework
- Argument parsing and validation
- User interaction and progress reporting
- Delegates to service layer for business logic

### Configuration Layer (`internal/config/`)

- TOML/YAML parsing using Viper
- Hierarchical configuration loading (system → global → project → local)
- Environment-specific overrides
- Configuration validation and schema enforcement

### Service Layer (`internal/service/`)

- Core business logic
- Orchestrates operations across backends, providers, and database
- Implements atomic operations with transaction support
- Error handling and recovery logic

### Backend System (`internal/backend/`)

- Pluggable backend implementations (GitHub, Aqua, HTTP)
- Version listing and resolution
- Artifact download coordination
- Backend registry and discovery

### Provider System (`internal/provider/`)

- Tool-specific installation logic
- Post-installation hooks
- Shim generation
- Version detection from existing installations

### Data Layer (`internal/repository/`)

- SQLite database access
- Cache management
- Index management
- Audit log persistence

### Infrastructure Layer (`internal/pkg/`)

- Download implementations
- File system operations
- Platform-specific utilities
- Logging (zerolog)

## Key Technologies

- **Language**: Go 1.21+
- **CLI Framework**: Cobra
- **Configuration**: Viper (TOML/YAML)
- **Database**: SQLite with mattn/go-sqlite3
- **Logging**: zerolog (in `internal/pkg/logger`)
- **Testing**: testify, gopter (for property-based tests)
- **Build**: Makefile, goreleaser
- **CI/CD**: GitHub Actions
- **Linting**: golangci-lint

## Build System

### Makefile Targets

- `make setup` - Install system-level development tools
- `make install` - Install project-level dependencies
- `make build` - Build project artifacts
- `make test` - Run test suite
- `make lint` - Run linters
- `make verify` - Run full verification (lint + test + audit)
- `make clean` - Remove temporary and generated files

### Go Commands

```bash
# Build
go build ./...

# Test
go test ./...
go test -race ./...  # with race detector

# Lint
golangci-lint run ./...

# Format
gofmt -w .
goimports -w .
```

## Development Workflow

1. **Setup**: `make setup && make install`
2. **Develop**: Write code following Go conventions
3. **Format**: Auto-format on save (goimports)
4. **Lint**: `make lint` (runs pre-commit hooks)
5. **Test**: `make test`
6. **Build**: `make build`
7. **Commit**: `make commit` (Commitizen)
8. **Verify**: `make verify` (full check before PR)

## Code Style

UniRTM follows the code style conventions from the [tott project](https://github.com/evilmartians/tott):

- Use `goimports` for formatting (includes `gofmt`)
- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `golangci-lint` for comprehensive linting
- Keep functions small and focused
- Write clear, descriptive names
- Document all exported functions and types
- Handle all errors explicitly
- Use context.Context for cancellation and timeouts

## Testing Strategy

- **Unit Tests**: Test individual functions and methods
- **Integration Tests**: Test component interactions
- **Property-Based Tests**: Test universal properties with gopter
- **E2E Tests**: Test complete workflows

All tests should be in `*_test.go` files alongside the code they test.

## References

- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [tott project](https://github.com/evilmartians/tott)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
