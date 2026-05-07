package service

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
)

// DependencyResolver manages tool dependency resolution and installation ordering.
// It parses dependency declarations, builds dependency graphs, detects circular
// dependencies, and determines the correct installation order using topological sort.
type DependencyResolver struct {
	// providerRegistry provides access to provider metadata for dependency information
	providerRegistry ProviderRegistry
	// backendRegistry provides access to backend version resolution
	backendRegistry BackendRegistry
	// versionManager handles version constraint resolution
	versionManager *VersionManager
}

// ProviderRegistry defines the interface for accessing provider metadata.
type ProviderRegistry interface {
	// GetProvider returns the provider for a given tool name
	GetProvider(tool string) (ProviderMetadata, error)
}

// BackendRegistry defines the interface for accessing backend information.
type BackendRegistry interface {
	// ResolveVersion resolves a version specification to a concrete version
	ResolveVersion(ctx context.Context, tool string, versionSpec string) (string, error)
}

// ProviderMetadata contains metadata about a tool provider including dependencies.
type ProviderMetadata struct {
	Name         string
	Dependencies []Dependency
}

// Dependency represents a tool dependency with version constraints.
type Dependency struct {
	Tool              string // Tool name
	VersionConstraint string // Version constraint (e.g., ">=1.20.0", "^3.11", "latest")
}

// DependencyGraph represents the dependency relationships between tools.
type DependencyGraph struct {
	// nodes maps tool names to their dependencies
	nodes map[string][]string
	// versions maps tool names to their resolved versions
	versions map[string]string
}

// InstallationOrder represents the ordered list of tools to install.
type InstallationOrder struct {
	// Tools is the ordered list of tool names
	Tools []string
	// Versions maps tool names to their resolved versions
	Versions map[string]string
}

// DependencyConflict represents a version conflict between dependencies.
type DependencyConflict struct {
	Tool       string
	Requesters []ConflictingRequester
}

// ConflictingRequester represents a tool that requires a specific version.
type ConflictingRequester struct {
	Tool             string
	RequestedVersion string
}

// NewDependencyResolver creates a new DependencyResolver.
func NewDependencyResolver(
	providerRegistry ProviderRegistry,
	backendRegistry BackendRegistry,
	versionManager *VersionManager,
) *DependencyResolver {
	return &DependencyResolver{
		providerRegistry: providerRegistry,
		backendRegistry:  backendRegistry,
		versionManager:   versionManager,
	}
}

// ParseDependencies parses dependency declarations from provider metadata for the given tools.
// It returns a dependency graph containing all tools and their dependencies.
//
// Validates: Requirement 16.1
func (r *DependencyResolver) ParseDependencies(ctx context.Context, tools []string) (*DependencyGraph, error) {
	logger.Debug("Parsing dependencies", map[string]interface{}{
		"tools": tools,
	})

	graph := &DependencyGraph{
		nodes:    make(map[string][]string),
		versions: make(map[string]string),
	}

	// Use a queue to process tools and their dependencies
	queue := make([]string, len(tools))
	copy(queue, tools)
	visited := make(map[string]bool)

	for len(queue) > 0 {
		tool := queue[0]
		queue = queue[1:]

		if visited[tool] {
			continue
		}
		visited[tool] = true

		// Get provider metadata for this tool
		metadata, err := r.providerRegistry.GetProvider(tool)
		if err != nil {
			return nil, fmt.Errorf("get provider metadata for %s: %w", tool, err)
		}

		// Initialize node in graph
		if _, exists := graph.nodes[tool]; !exists {
			graph.nodes[tool] = []string{}
		}

		// Add dependencies to graph and queue
		for _, dep := range metadata.Dependencies {
			graph.nodes[tool] = append(graph.nodes[tool], dep.Tool)

			// Add dependency to queue for processing
			if !visited[dep.Tool] {
				queue = append(queue, dep.Tool)
			}
		}
	}

	logger.Info("Dependency graph parsed", map[string]interface{}{
		"tool_count":       len(graph.nodes),
		"dependency_count": countEdges(graph),
	})

	return graph, nil
}

// DetectCircularDependencies checks for circular dependencies in the graph.
// Returns an error if a circular dependency is detected.
//
// Validates: Requirement 16.3
func (r *DependencyResolver) DetectCircularDependencies(graph *DependencyGraph) error {
	logger.Debug("Detecting circular dependencies", map[string]interface{}{
		"tool_count": len(graph.nodes),
	})

	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := []string{}

	var hasCycle func(tool string) bool
	hasCycle = func(tool string) bool {
		visited[tool] = true
		recStack[tool] = true
		path = append(path, tool)

		for _, dep := range graph.nodes[tool] {
			if !visited[dep] {
				if hasCycle(dep) {
					return true
				}
			} else if recStack[dep] {
				// Found a cycle - build the cycle path
				cycleStart := -1
				for i, t := range path {
					if t == dep {
						cycleStart = i
						break
					}
				}
				_ = cycleStart // Will be used for detailed cycle reporting in future
				return true
			}
		}

		recStack[tool] = false
		path = path[:len(path)-1]
		return false
	}

	for tool := range graph.nodes {
		if !visited[tool] {
			if hasCycle(tool) {
				// Build cycle description
				cycleDesc := strings.Join(path, " -> ")
				logger.Error("Circular dependency detected", map[string]interface{}{
					"cycle": cycleDesc,
				})
				return fmt.Errorf("circular dependency detected: %s", cycleDesc)
			}
		}
	}

	logger.Info("No circular dependencies detected", nil)
	return nil
}

// TopologicalSort performs a topological sort on the dependency graph to determine
// the correct installation order. Tools with no dependencies come first, followed
// by tools that depend on them.
//
// Validates: Requirement 16.4
func (r *DependencyResolver) TopologicalSort(graph *DependencyGraph) (*InstallationOrder, error) {
	logger.Debug("Performing topological sort", map[string]interface{}{
		"tool_count": len(graph.nodes),
	})

	// First check for circular dependencies
	if err := r.DetectCircularDependencies(graph); err != nil {
		return nil, err
	}

	// Calculate in-degree for each node (number of dependencies it has)
	inDegree := make(map[string]int)
	for tool, deps := range graph.nodes {
		inDegree[tool] = len(deps)
	}

	// Build reverse graph: for each dependency, track which tools depend on it
	reverseDeps := make(map[string][]string)
	for tool := range graph.nodes {
		reverseDeps[tool] = []string{}
	}
	for tool, deps := range graph.nodes {
		for _, dep := range deps {
			reverseDeps[dep] = append(reverseDeps[dep], tool)
		}
	}

	// Initialize queue with nodes that have no dependencies (in-degree = 0)
	queue := []string{}
	for tool, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, tool)
		}
	}

	// Sort queue for deterministic output
	sort.Strings(queue)

	// Process nodes in topological order
	result := []string{}
	for len(queue) > 0 {
		// Pop from queue
		tool := queue[0]
		queue = queue[1:]
		result = append(result, tool)

		// For each tool that depends on this tool, reduce its in-degree
		for _, dependent := range reverseDeps[tool] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
				// Keep queue sorted for deterministic output
				sort.Strings(queue)
			}
		}
	}

	// If we haven't processed all nodes, there's a cycle (shouldn't happen after cycle check)
	if len(result) != len(graph.nodes) {
		return nil, fmt.Errorf("topological sort failed: graph contains a cycle")
	}

	order := &InstallationOrder{
		Tools:    result,
		Versions: graph.versions,
	}

	logger.Info("Topological sort complete", map[string]interface{}{
		"order": result,
	})

	return order, nil
}

// ResolveVersionConstraints resolves version constraints for all tools in the dependency graph.
// It checks for conflicts where multiple tools depend on different versions of the same tool.
//
// Validates: Requirements 16.6, 16.7
func (r *DependencyResolver) ResolveVersionConstraints(ctx context.Context, graph *DependencyGraph, requestedVersions map[string]string) error {
	logger.Debug("Resolving version constraints", map[string]interface{}{
		"tool_count": len(graph.nodes),
	})

	// Track version requirements for each tool
	requirements := make(map[string][]ConflictingRequester)

	// Add explicitly requested versions
	for tool, version := range requestedVersions {
		requirements[tool] = append(requirements[tool], ConflictingRequester{
			Tool:             "user",
			RequestedVersion: version,
		})
	}

	// Collect version requirements from dependencies
	for tool := range graph.nodes {
		metadata, err := r.providerRegistry.GetProvider(tool)
		if err != nil {
			return fmt.Errorf("get provider metadata for %s: %w", tool, err)
		}

		for _, dep := range metadata.Dependencies {
			requirements[dep.Tool] = append(requirements[dep.Tool], ConflictingRequester{
				Tool:             tool,
				RequestedVersion: dep.VersionConstraint,
			})
		}
	}

	// Check for conflicts and resolve versions
	conflicts := []DependencyConflict{}
	for tool, reqs := range requirements {
		if len(reqs) == 0 {
			continue
		}

		// If all requirements are the same, no conflict
		if allSame(reqs) {
			version := reqs[0].RequestedVersion
			// Resolve version specification to concrete version
			resolved, err := r.backendRegistry.ResolveVersion(ctx, tool, version)
			if err != nil {
				return fmt.Errorf("resolve version for %s: %w", tool, err)
			}
			graph.versions[tool] = resolved
			continue
		}

		// Check if requirements are compatible
		compatible, resolvedVersion, err := r.checkCompatibility(ctx, tool, reqs)
		if err != nil {
			return fmt.Errorf("check compatibility for %s: %w", tool, err)
		}

		if compatible {
			graph.versions[tool] = resolvedVersion
		} else {
			conflicts = append(conflicts, DependencyConflict{
				Tool:       tool,
				Requesters: reqs,
			})
		}
	}

	// Report conflicts if any
	if len(conflicts) > 0 {
		return r.formatConflictError(conflicts)
	}

	logger.Info("Version constraints resolved", map[string]interface{}{
		"resolved_count": len(graph.versions),
	})

	return nil
}

// ResolveDependencies is the main entry point that combines all dependency resolution steps.
// It parses dependencies, detects circular dependencies, resolves version constraints,
// and returns the installation order.
//
// Validates: Requirements 16.1, 16.2, 16.3, 16.4, 16.5, 16.6, 16.7
func (r *DependencyResolver) ResolveDependencies(ctx context.Context, tools []string, requestedVersions map[string]string) (*InstallationOrder, error) {
	logger.Info("Resolving dependencies", map[string]interface{}{
		"tools": tools,
	})

	// Parse dependency graph
	graph, err := r.ParseDependencies(ctx, tools)
	if err != nil {
		return nil, fmt.Errorf("parse dependencies: %w", err)
	}

	// Detect circular dependencies
	if err := r.DetectCircularDependencies(graph); err != nil {
		return nil, err
	}

	// Resolve version constraints
	if err := r.ResolveVersionConstraints(ctx, graph, requestedVersions); err != nil {
		return nil, err
	}

	// Perform topological sort to get installation order
	order, err := r.TopologicalSort(graph)
	if err != nil {
		return nil, fmt.Errorf("topological sort: %w", err)
	}

	logger.Info("Dependencies resolved successfully", map[string]interface{}{
		"tool_count":    len(order.Tools),
		"install_order": order.Tools,
	})

	return order, nil
}

// Helper functions

// countEdges counts the total number of edges in the graph
func countEdges(graph *DependencyGraph) int {
	count := 0
	for _, deps := range graph.nodes {
		count += len(deps)
	}
	return count
}

// allSame checks if all requesters have the same version requirement
func allSame(reqs []ConflictingRequester) bool {
	if len(reqs) <= 1 {
		return true
	}
	first := reqs[0].RequestedVersion
	for _, req := range reqs[1:] {
		if req.RequestedVersion != first {
			return false
		}
	}
	return true
}

// checkCompatibility checks if version requirements are compatible and returns a resolved version
func (r *DependencyResolver) checkCompatibility(ctx context.Context, tool string, reqs []ConflictingRequester) (bool, string, error) {
	// For now, we use a simple strategy: if all requirements can be satisfied by a single version,
	// they are compatible. This is a simplified implementation - a full implementation would
	// need to parse and compare version constraints (e.g., ">=1.20.0" and "^1.20" are compatible).

	// Try to resolve each requirement and see if they all resolve to the same version
	versions := make(map[string]bool)
	var lastResolved string

	for _, req := range reqs {
		resolved, err := r.backendRegistry.ResolveVersion(ctx, tool, req.RequestedVersion)
		if err != nil {
			return false, "", fmt.Errorf("resolve version %s for %s: %w", req.RequestedVersion, tool, err)
		}
		versions[resolved] = true
		lastResolved = resolved
	}

	// If all requirements resolve to the same version, they are compatible
	if len(versions) == 1 {
		return true, lastResolved, nil
	}

	// Multiple different versions required - conflict
	return false, "", nil
}

// formatConflictError formats a user-friendly error message for version conflicts
func (r *DependencyResolver) formatConflictError(conflicts []DependencyConflict) error {
	var sb strings.Builder
	sb.WriteString("version conflicts detected:\n")

	for _, conflict := range conflicts {
		sb.WriteString(fmt.Sprintf("\nTool: %s\n", conflict.Tool))
		sb.WriteString("  Conflicting requirements:\n")
		for _, req := range conflict.Requesters {
			sb.WriteString(fmt.Sprintf("    - %s requires version %s\n", req.Tool, req.RequestedVersion))
		}
		sb.WriteString("  Suggested resolutions:\n")
		sb.WriteString("    1. Explicitly specify a version that satisfies all requirements\n")
		sb.WriteString("    2. Update tools to use compatible version constraints\n")
		sb.WriteString("    3. Install conflicting tools in separate environments\n")
	}

	return fmt.Errorf("%s", sb.String())
}
