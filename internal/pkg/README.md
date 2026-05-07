# Internal Packages

This directory contains shared internal packages used across UniRTM.

## Packages

### logger

Zerolog-based logging system with rotating file support.

**Features:**
- Structured logging with JSON output
- Multiple log levels (Trace, Debug, Info, Warn, Error, Fatal, Panic)
- Rotating file writers for error.log and unirtm.log
- Console output with color-coded levels
- Stack trace capture for errors

**Usage:**
```go
import "github.com/snowdreamtech/unirtm/internal/pkg/logger"

// Initialize logger
errorWriter, ginWriter := logger.InitLogger("error.log", "unirtm.log")

// Use logger
logger.Logger.Info().Str("tool", "node").Str("version", "20.0.0").Msg("installing tool")
logger.Logger.Error().Err(err).Msg("installation failed")
```

### env

Environment and build metadata management.

**Features:**
- Project metadata (name, version, author)
- Build information (time, git tag, commit hash)
- Runtime configuration flags (debug, trace, quiet)

**Usage:**
```go
import "github.com/snowdreamtech/unirtm/internal/pkg/env"

fmt.Printf("Project: %s\n", env.ProjectName)
fmt.Printf("Version: %s\n", env.GitTag)
```

## Requirements

- logger implements requirements: 7.1-7.8
- env provides build metadata for version command
