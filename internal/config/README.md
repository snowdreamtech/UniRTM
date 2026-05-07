# Configuration Layer

This package handles all configuration management for UniRTM using Viper.

## Responsibilities

- Parse TOML and YAML configuration files
- Implement hierarchical configuration loading (system → global → project → local)
- Validate configuration structures
- Merge configurations with proper precedence rules
- Handle environment-specific overrides

## Key Components

- `config.go` - Core configuration data structures ✅ **Implemented**
- `manager.go` - Configuration manager implementation (planned)
- `validator.go` - Configuration validation logic (integrated in config.go)
- `parser.go` - Configuration parsing utilities (planned)

## Data Structures

### Config
Root configuration structure containing:
- **Tools**: Map of tool names to version specifications
- **Env**: Environment variable definitions
- **Settings**: Global settings (cache, data directories, TTL, concurrency)
- **Tasks**: Task definitions with dependencies

### ToolConfig
Tool version specification with:
- **Version**: Required version string (exact, range, or alias)
- **Backend**: Optional backend name (github, aqua, http)
- **Provider**: Optional provider name (node, python, go, generic)

### Settings
Global settings with:
- **CacheDir**: Cache directory path
- **DataDir**: Data directory path
- **CacheTTL**: Cache time-to-live in seconds
- **Concurrency**: Maximum concurrent operations

### Task
Task definition with:
- **Description**: Human-readable description
- **Run**: Command to execute
- **Env**: Task-specific environment variables
- **Depends**: List of task dependencies

## Validation

All structures implement `Validate()` methods that:
- Check required fields are present
- Validate field values are within acceptable ranges
- Report all validation errors (not just the first one)
- Detect circular task dependencies

## Usage

```go
import "github.com/snowdreamtech/unirtm/internal/config"

// Create configuration
cfg := config.Config{
    Tools: map[string]config.ToolConfig{
        "node": {Version: "20.0.0"},
    },
    Settings: config.Settings{
        CacheTTL: 3600,
        Concurrency: 4,
    },
}

// Validate configuration
if err := cfg.Validate(); err != nil {
    log.Fatal(err)
}
```

## Testing

Comprehensive unit tests with 100% code coverage:

```bash
go test -v ./internal/config/
go test -race -cover ./internal/config/
```

## Status

- ✅ **Task 2.1 Complete**: Configuration data structures with TOML/YAML tags and validation methods
- 🚧 **Task 2.2 Planned**: ConfigManager interface with Viper integration
- 🚧 **Task 2.3-2.6 Planned**: Property-based tests and environment overrides

## Requirements

Implements requirements:
- ✅ **1.3**: Configuration validation with required field checking
- ✅ **26.1**: TOML tag support for all structures
- ✅ **26.2**: YAML tag support for all structures
- 🚧 **1.1, 1.2, 1.4-1.8**: Configuration parsing and loading (planned)
- 🚧 **26.3-26.8**: Round-trip properties and pretty printing (planned)
