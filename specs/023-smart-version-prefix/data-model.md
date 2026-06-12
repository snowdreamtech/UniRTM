# Phase 1: Data Model & Contracts

## Data Model

This feature doesn't introduce any new persisted data models, database schemas, or complex structs. It operates purely via string transformations.

### Function Signature

```go
// NormalizeVersionPrefix ensures the version string conforms to the backend's 'v' prefix requirements.
// If requiresVPrefix is true, it prepends 'v' if missing (and the string starts with a digit).
// If requiresVPrefix is false, it strips 'v'/'V' if present (and followed by a digit).
func NormalizeVersionPrefix(version string, requiresVPrefix bool) string
```

### State Transitions

N/A
