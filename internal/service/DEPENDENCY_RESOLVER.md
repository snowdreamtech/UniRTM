# Dependency Resolver

The `DependencyResolver` manages tool dependency resolution and installation ordering for UniRTM. It parses dependency declarations, builds dependency graphs, detects circular dependencies, and determines the correct installation order using topological sort.

## Overview

The Dependency Resolver is a critical component of the Service Layer that ensures tools are installed in the correct order, respecting their dependencies. It integrates with the Provider Registry for dependency metadata and the Backend Registry for version resolution.

## Features

- **Dependency Graph Parsing**: Parses tool dependency declarations from provider metadata
- **Circular Dependency Detection**: Detects and reports circular dependencies with detailed cycle paths
- **Topological Sorting**: Determines the correct installation order using Kahn's algorithm
- **Version Constraint Resolution**: Resolves version constraints and detects conflicts
- **Automatic Dependency Installation**: Ensures dependencies are installed before dependent tools
- **Conflict Detection and Reporting**: Provides detailed conflict reports with suggested resolutions

## Architecture

```
DependencyResolver
├── ParseDependencies()       # Build dependency graph
├── DetectCircularDependencies()  # Check for cycles
├── TopologicalSort()         # Determine installation order
├── ResolveVersionConstraints()   # Resolve versions and detect conflicts
└── ResolveDependencies()     # Main entry point (combines all steps)
```

## Usage Example

```go
// Initialize dependencies
providerRegistry := NewProviderRegistry()
backendRegistry := NewBackendRegistry()
versionManager := NewVersionManager()

// Create dependency resolver
resolver := service.NewDependencyResolver(
    providerRegistry,
    backendRegistry,
    versionManager,
)

// Resolve dependencies for tools
tools := []string{"node", "python"}
requestedVersions := map[string]string{
    "node":   "20.0.0",
    "python": "3.11.0",
}

order, err := resolver.ResolveDependencies(ctx, tools, requestedVersions)
if err != nil {
    return err
}

// Install tools in the correct order
for _, tool := range order.Tools {
    version := order.Versions[tool]
    fmt.Printf("Installing %s@%s\n", tool, version)
    // Install tool...
}
```

## Dependency Graph Structure

The dependency graph is represented as an adjacency list where:

- `graph.nodes[tool]` contains the list of tools that `tool` depends on
- `graph.versions[tool]` contains the resolved version for `tool`

Example:

```go
graph := &DependencyGraph{
    nodes: map[string][]string{
        "app":    {"lib1", "lib2"},  // app depends on lib1 and lib2
        "lib1":   {"common"},         // lib1 depends on common
        "lib2":   {"common"},         // lib2 depends on common
        "common": {},                 // common has no dependencies
    },
    versions: map[string]string{
        "app":    "1.0.0",
        "lib1":   "2.0.0",
        "lib2":   "3.0.0",
        "common": "1.0.0",
    },
}
```

## Topological Sort Algorithm

The resolver uses Kahn's algorithm for topological sorting:

1. Calculate in-degree for each node (number of dependencies)
2. Initialize queue with nodes that have no dependencies (in-degree = 0)
3. Process nodes from queue:
   - Add node to result
   - For each tool that depends on this node, reduce its in-degree
   - If in-degree becomes 0, add to queue
4. If all nodes are processed, return result; otherwise, there's a cycle

This ensures that dependencies are always installed before the tools that depend on them.

## Circular Dependency Detection

The resolver uses depth-first search (DFS) with a recursion stack to detect cycles:

1. Maintain visited set and recursion stack
2. For each unvisited node, perform DFS
3. If we encounter a node in the recursion stack, a cycle is detected
4. Report the cycle path for debugging

Example cycle detection:

```
Circular dependency detected: a -> b -> c -> d -> b
```

## Version Constraint Resolution

The resolver resolves version constraints and detects conflicts:

1. Collect version requirements from:
   - Explicitly requested versions (user input)
   - Dependency declarations (provider metadata)
2. For each tool with multiple requirements:
   - If all requirements are identical, use that version
   - If requirements are compatible, resolve to a common version
   - If requirements conflict, report the conflict
3. Resolve version specifications to concrete versions using the backend

### Conflict Detection

When multiple tools depend on different versions of the same tool, the resolver detects the conflict and provides a detailed report:

```
version conflicts detected:

Tool: common
  Conflicting requirements:
    - app1 requires version 1.0.0
    - app2 requires version 2.0.0
  Suggested resolutions:
    1. Explicitly specify a version that satisfies all requirements
    2. Update tools to use compatible version constraints
    3. Install conflicting tools in separate environments
```

## Error Handling

The resolver provides detailed error messages for common issues:

- **Missing Provider**: "provider not found for tool: <tool>"
- **Circular Dependency**: "circular dependency detected: <cycle path>"
- **Version Conflict**: "version conflicts detected: <detailed conflict report>"
- **Version Not Found**: "version not found for tool <tool>: <version>"

## Integration Points

### Provider Registry

The resolver queries the Provider Registry for tool metadata:

```go
type ProviderRegistry interface {
    GetProvider(tool string) (ProviderMetadata, error)
}

type ProviderMetadata struct {
    Name         string
    Dependencies []Dependency
}

type Dependency struct {
    Tool              string
    VersionConstraint string
}
```

### Backend Registry

The resolver uses the Backend Registry for version resolution:

```go
type BackendRegistry interface {
    ResolveVersion(ctx context.Context, tool string, versionSpec string) (string, error)
}
```

### Version Manager

The resolver integrates with the Version Manager for advanced version constraint resolution (future enhancement).

## Requirements Validation

The Dependency Resolver validates the following requirements:

- **Requirement 16.1**: Parse tool dependency declarations from provider metadata
- **Requirement 16.2**: Build a dependency graph for all requested tools
- **Requirement 16.3**: Detect circular dependencies and report errors
- **Requirement 16.4**: Determine the correct installation order (topological sort)
- **Requirement 16.5**: Install dependencies before dependent tools
- **Requirement 16.6**: Support version constraints for dependencies
- **Requirement 16.7**: Detect and report conflicts when multiple tools depend on different versions

## Testing

The resolver includes comprehensive unit tests covering:

- **Dependency Parsing**: Simple, transitive, and diamond dependencies
- **Circular Detection**: Simple cycles, self-dependencies, complex cycles
- **Topological Sort**: Simple graphs, diamond dependencies, complex graphs
- **Version Resolution**: Compatible versions, same version from multiple requesters, conflicts
- **Integration**: Complete dependency tree resolution, error handling

Run tests:

```bash
go test -v ./internal/service -run TestDependencyResolver
```

## Future Enhancements

1. **Advanced Version Constraint Resolution**: Support for semver ranges (e.g., ">=1.20.0", "^3.11")
2. **Dependency Caching**: Cache resolved dependency graphs for performance
3. **Parallel Installation**: Install independent tools in parallel
4. **Dependency Visualization**: Generate visual dependency graphs
5. **Conflict Resolution Strategies**: Automatic conflict resolution based on policies
6. **Optional Dependencies**: Support for optional dependencies that don't block installation

## References

- [Design Document](../../README.md)
- [Requirements Document](../../README.md)
- [Service Layer README](./README.md)
- [Provider System](../provider/provider.go)
