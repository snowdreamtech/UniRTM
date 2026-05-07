# Version Manager

## Overview

The Version Manager is a service layer component that handles version constraint parsing and resolution for UniRTM. It enforces explicit version requirements and delegates to backends for version resolution.

## Architecture

The Version Manager sits between the CLI/configuration layer and the backend system:

```
CLI/Config → Version Manager → Backend System
                ↓
         Version Parser
```

## Key Features

### 1. Version Constraint Parsing

The Version Manager parses version constraints using the Version Parser (task 11.2):

- **Exact versions**: `1.20.0`, `v1.20.0`
- **Range constraints**: `>=1.20.0`, `>1.0.0`, `<=2.0.0`, `<3.0.0`, `=1.2.3`
- **Caret ranges**: `^1.20.0` (compatible with)
- **Tilde ranges**: `~2.7.0` (approximately equivalent to)
- **Aliases**: `latest`, `lts`, `stable`

### 2. Version Resolution

The Version Manager resolves version specifications to concrete versions by delegating to backends:

- **Exact versions**: Directly fetches download info from the backend
- **Aliases and ranges**: Delegates to backend's `ResolveVersion` method

### 3. Explicit Version Requirement Enforcement

The Version Manager enforces that all version specifications are explicit:

- Empty version strings are rejected with a clear error message
- Users must specify an exact version, range, or alias
- No implicit defaults or silent fallbacks

## API

### NewVersionManager

```go
func NewVersionManager(backends map[string]backend.Backend) *VersionManager
```

Creates a new Version Manager with the given backends.

### ParseVersionConstraint

```go
func (vm *VersionManager) ParseVersionConstraint(versionStr string) (*Version, error)
```

Parses a version constraint string into a Version object. Enforces explicit version requirements.

**Returns an error if:**
- The version string is empty (explicit version required)
- The version string is invalid

### ResolveVersion

```go
func (vm *VersionManager) ResolveVersion(
    ctx context.Context,
    backendName, tool, versionSpec string,
    platform backend.Platform,
) (*backend.VersionInfo, error)
```

Resolves a version specification to a concrete version using the specified backend.

**Parameters:**
- `ctx`: Context for cancellation and timeouts
- `backendName`: Name of the backend to use (e.g., "github", "aqua", "http")
- `tool`: Name of the tool to resolve the version for
- `versionSpec`: Version specification (exact version, range, or alias)
- `platform`: Target platform for the resolution

**Returns:**
- The resolved VersionInfo with concrete version and download information
- An error if resolution fails

**Resolution behavior:**
- **Exact versions** (e.g., "1.20.0"): Delegates to `backend.GetDownloadInfo`
- **Aliases** (e.g., "latest", "lts", "stable"): Delegates to `backend.ResolveVersion`
- **Ranges** (e.g., ">=1.20.0", "^3.11", "~2.7.0"): Delegates to `backend.ResolveVersion`

### ValidateVersionConstraint

```go
func (vm *VersionManager) ValidateVersionConstraint(versionStr string) error
```

Validates a version constraint string without resolving it. Useful for configuration validation.

### ListAvailableVersions

```go
func (vm *VersionManager) ListAvailableVersions(
    ctx context.Context,
    backendName, tool string,
    platform backend.Platform,
) ([]backend.VersionInfo, error)
```

Lists all available versions for a tool from the specified backend. Returns versions in descending order (newest first).

### SupportsChecksum

```go
func (vm *VersionManager) SupportsChecksum(backendName string) (bool, error)
```

Checks if the specified backend supports checksum verification.

### SupportsGPG

```go
func (vm *VersionManager) SupportsGPG(backendName string) (bool, error)
```

Checks if the specified backend supports GPG signature verification.

## Usage Examples

### Basic Version Resolution

```go
// Create Version Manager with backends
backends := map[string]backend.Backend{
    "github": githubBackend,
    "aqua":   aquaBackend,
}
vm := NewVersionManager(backends)

// Resolve exact version
ctx := context.Background()
platform := backend.CurrentPlatform()
versionInfo, err := vm.ResolveVersion(ctx, "github", "node", "20.0.0", platform)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Resolved to: %s\n", versionInfo.Version)
fmt.Printf("Download URL: %s\n", versionInfo.DownloadURL)
```

### Resolving Aliases

```go
// Resolve "latest" alias
versionInfo, err := vm.ResolveVersion(ctx, "github", "node", "latest", platform)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Latest version: %s\n", versionInfo.Version)
```

### Resolving Ranges

```go
// Resolve caret range
versionInfo, err := vm.ResolveVersion(ctx, "github", "node", "^20.0.0", platform)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Resolved ^20.0.0 to: %s\n", versionInfo.Version)

// Resolve comparison range
versionInfo, err = vm.ResolveVersion(ctx, "github", "python", ">=3.11.0", platform)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Resolved >=3.11.0 to: %s\n", versionInfo.Version)
```

### Configuration Validation

```go
// Validate version constraints in configuration
toolVersions := map[string]string{
    "node":   "20.0.0",
    "python": "^3.11",
    "go":     "latest",
}

for tool, version := range toolVersions {
    if err := vm.ValidateVersionConstraint(version); err != nil {
        log.Printf("Invalid version for %s: %v\n", tool, err)
    }
}
```

### Listing Available Versions

```go
// List all available versions for a tool
versions, err := vm.ListAvailableVersions(ctx, "github", "node", platform)
if err != nil {
    log.Fatal(err)
}

fmt.Println("Available Node.js versions:")
for _, v := range versions {
    fmt.Printf("  - %s\n", v.Version)
}
```

## Error Handling

The Version Manager provides clear, actionable error messages:

### Empty Version Specification

```
explicit version specification required for tool 'node': must specify an exact version (e.g., 1.20.0), range (e.g., >=1.20.0, ^3.11, ~2.7.0), or alias (latest, lts, stable)
```

### Invalid Version Format

```
resolve version for tool 'node': parse version constraint: invalid version string 'invalid': must be a valid semver (e.g., 1.2.3), range (e.g., >=1.2.0, ^1.2.3, ~1.2.0), or alias (latest, lts, stable)
```

### Backend Not Found

```
backend 'nonexistent' not found for tool 'node'
```

### Backend Resolution Error

```
resolve version 'latest' for tool 'node': backend error: API rate limit exceeded
```

## Integration with Other Components

### Configuration System

The Version Manager is used by the Configuration Manager to validate version specifications in configuration files:

```go
// In ConfigManager.Validate()
for tool, config := range cfg.Tools {
    if err := versionManager.ValidateVersionConstraint(config.Version); err != nil {
        return fmt.Errorf("invalid version for tool %s: %w", tool, err)
    }
}
```

### Installation Manager

The Installation Manager uses the Version Manager to resolve versions before installation:

```go
// In InstallationManager.Install()
versionInfo, err := versionManager.ResolveVersion(ctx, backendName, tool, versionSpec, platform)
if err != nil {
    return fmt.Errorf("resolve version: %w", err)
}

// Proceed with installation using versionInfo
```

### CLI Commands

CLI commands use the Version Manager to resolve user-provided version specifications:

```go
// In `unirtm install node@latest`
versionInfo, err := versionManager.ResolveVersion(ctx, "github", "node", "latest", platform)
if err != nil {
    return err
}

fmt.Printf("Installing Node.js %s...\n", versionInfo.Version)
```

## Design Principles

### 1. Explicit Over Implicit

The Version Manager enforces explicit version specifications. There are no implicit defaults or silent fallbacks. This ensures:

- Users always know what version they're getting
- No surprises from automatic version selection
- Clear audit trail of version choices

### 2. Separation of Concerns

The Version Manager focuses solely on version constraint parsing and resolution. It delegates:

- **Version parsing**: To the Version Parser (task 11.2)
- **Version resolution**: To the Backend system
- **Installation**: To the Installation Manager

### 3. Backend Agnostic

The Version Manager works with any backend that implements the Backend interface. It doesn't know or care about:

- How backends fetch version lists
- How backends resolve aliases
- Where backends download artifacts from

This allows new backends to be added without modifying the Version Manager.

### 4. Error Context

All errors include sufficient context for debugging:

- Tool name
- Version specification
- Backend name
- Underlying error cause

## Testing

The Version Manager has comprehensive unit tests covering:

- Version constraint parsing (exact, ranges, aliases)
- Version resolution (exact, aliases, ranges)
- Explicit version requirement enforcement
- Backend integration
- Error handling
- Edge cases

Run tests with:

```bash
go test -v ./internal/service -run TestVersionManager
```

## Requirements Satisfied

This implementation satisfies the following requirements:

- **Requirement 9.6**: Version constraint parsing (semver, ranges, aliases)
- **Requirement 9.7**: Version resolution (latest, lts, stable)
- **Requirement 9.8**: Explicit version requirement enforcement
- **Requirement 8.4**: No implicit version defaults

## Future Enhancements

Potential future enhancements:

1. **Version constraint validation against available versions**: Check if a range constraint matches any available versions before attempting resolution
2. **Version caching**: Cache resolved versions to avoid repeated backend calls
3. **Version pinning**: Support for lockfiles that pin resolved versions
4. **Version suggestions**: Suggest valid versions when an invalid version is provided
5. **Batch resolution**: Resolve multiple version specifications in parallel

## Related Components

- **Version Parser** (task 11.2): Parses version strings into structured Version objects
- **Backend System** (tasks 7.1-7.6): Provides version lists and resolution
- **Installation Manager** (task 10.1): Uses resolved versions for installation
- **Configuration Manager** (task 2): Validates version specifications in config files
