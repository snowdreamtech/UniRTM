// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock implementations for testing

type mockDependencyProviderRegistry struct {
	providers map[string]ProviderMetadata
}

func (m *mockDependencyProviderRegistry) GetProvider(tool string) (ProviderMetadata, error) {
	if provider, exists := m.providers[tool]; exists {
		return provider, nil
	}
	return ProviderMetadata{}, fmt.Errorf("provider not found for tool: %s", tool)
}

type mockDependencyBackendRegistry struct {
	versions map[string]map[string]string // tool -> versionSpec -> resolvedVersion
}

func (m *mockDependencyBackendRegistry) ResolveVersion(ctx context.Context, tool string, versionSpec string) (string, error) {
	if toolVersions, exists := m.versions[tool]; exists {
		if resolved, exists := toolVersions[versionSpec]; exists {
			return resolved, nil
		}
	}
	return "", fmt.Errorf("version not found for tool %s: %s", tool, versionSpec)
}

// Test ParseDependencies

func TestDependencyResolver_ParseDependencies(t *testing.T) {
	ctx := context.Background()

	t.Run("parse simple dependency graph", func(t *testing.T) {
		providerRegistry := &mockDependencyProviderRegistry{
			providers: map[string]ProviderMetadata{
				"node": {
					Name:         "node",
					Dependencies: []Dependency{},
				},
				"python": {
					Name: "python",
					Dependencies: []Dependency{
						{Tool: "openssl", VersionConstraint: ">=1.1.1"},
					},
				},
				"openssl": {
					Name:         "openssl",
					Dependencies: []Dependency{},
				},
			},
		}

		resolver := NewDependencyResolver(providerRegistry, nil, nil)
		graph, err := resolver.ParseDependencies(ctx, []string{"node", "python"})

		require.NoError(t, err)
		assert.NotNil(t, graph)
		assert.Len(t, graph.nodes, 3) // node, python, openssl
		assert.Empty(t, graph.nodes["node"])
		assert.Equal(t, []string{"openssl"}, graph.nodes["python"])
		assert.Empty(t, graph.nodes["openssl"])
	})

	t.Run("parse transitive dependencies", func(t *testing.T) {
		providerRegistry := &mockDependencyProviderRegistry{
			providers: map[string]ProviderMetadata{
				"app": {
					Name: "app",
					Dependencies: []Dependency{
						{Tool: "lib1", VersionConstraint: "1.0.0"},
					},
				},
				"lib1": {
					Name: "lib1",
					Dependencies: []Dependency{
						{Tool: "lib2", VersionConstraint: "2.0.0"},
					},
				},
				"lib2": {
					Name:         "lib2",
					Dependencies: []Dependency{},
				},
			},
		}

		resolver := NewDependencyResolver(providerRegistry, nil, nil)
		graph, err := resolver.ParseDependencies(ctx, []string{"app"})

		require.NoError(t, err)
		assert.Len(t, graph.nodes, 3) // app, lib1, lib2
		assert.Equal(t, []string{"lib1"}, graph.nodes["app"])
		assert.Equal(t, []string{"lib2"}, graph.nodes["lib1"])
		assert.Empty(t, graph.nodes["lib2"])
	})

	t.Run("parse diamond dependency", func(t *testing.T) {
		providerRegistry := &mockDependencyProviderRegistry{
			providers: map[string]ProviderMetadata{
				"app": {
					Name: "app",
					Dependencies: []Dependency{
						{Tool: "lib1", VersionConstraint: "1.0.0"},
						{Tool: "lib2", VersionConstraint: "2.0.0"},
					},
				},
				"lib1": {
					Name: "lib1",
					Dependencies: []Dependency{
						{Tool: "common", VersionConstraint: "1.0.0"},
					},
				},
				"lib2": {
					Name: "lib2",
					Dependencies: []Dependency{
						{Tool: "common", VersionConstraint: "1.0.0"},
					},
				},
				"common": {
					Name:         "common",
					Dependencies: []Dependency{},
				},
			},
		}

		resolver := NewDependencyResolver(providerRegistry, nil, nil)
		graph, err := resolver.ParseDependencies(ctx, []string{"app"})

		require.NoError(t, err)
		assert.Len(t, graph.nodes, 4) // app, lib1, lib2, common
		assert.ElementsMatch(t, []string{"lib1", "lib2"}, graph.nodes["app"])
		assert.Equal(t, []string{"common"}, graph.nodes["lib1"])
		assert.Equal(t, []string{"common"}, graph.nodes["lib2"])
		assert.Empty(t, graph.nodes["common"])
	})

	t.Run("error on missing provider", func(t *testing.T) {
		providerRegistry := &mockDependencyProviderRegistry{
			providers: map[string]ProviderMetadata{
				"node": {
					Name: "node",
					Dependencies: []Dependency{
						{Tool: "missing", VersionConstraint: "1.0.0"},
					},
				},
			},
		}

		resolver := NewDependencyResolver(providerRegistry, nil, nil)
		_, err := resolver.ParseDependencies(ctx, []string{"node"})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "provider not found")
	})
}

// Test DetectCircularDependencies

func TestDependencyResolver_DetectCircularDependencies(t *testing.T) {
	resolver := NewDependencyResolver(nil, nil, nil)

	t.Run("no circular dependencies", func(t *testing.T) {
		graph := &DependencyGraph{
			nodes: map[string][]string{
				"app":    {"lib1", "lib2"},
				"lib1":   {"common"},
				"lib2":   {"common"},
				"common": {},
			},
			versions: make(map[string]string),
		}

		err := resolver.DetectCircularDependencies(graph)
		assert.NoError(t, err)
	})

	t.Run("detect simple circular dependency", func(t *testing.T) {
		graph := &DependencyGraph{
			nodes: map[string][]string{
				"a": {"b"},
				"b": {"a"},
			},
			versions: make(map[string]string),
		}

		err := resolver.DetectCircularDependencies(graph)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "circular dependency detected")
	})

	t.Run("detect self-dependency", func(t *testing.T) {
		graph := &DependencyGraph{
			nodes: map[string][]string{
				"a": {"a"},
			},
			versions: make(map[string]string),
		}

		err := resolver.DetectCircularDependencies(graph)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "circular dependency detected")
	})

	t.Run("detect complex circular dependency", func(t *testing.T) {
		graph := &DependencyGraph{
			nodes: map[string][]string{
				"a": {"b"},
				"b": {"c"},
				"c": {"d"},
				"d": {"b"}, // cycle: b -> c -> d -> b
			},
			versions: make(map[string]string),
		}

		err := resolver.DetectCircularDependencies(graph)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "circular dependency detected")
	})
}

// Test TopologicalSort

func TestDependencyResolver_TopologicalSort(t *testing.T) {
	resolver := NewDependencyResolver(nil, nil, nil)

	t.Run("sort simple graph", func(t *testing.T) {
		graph := &DependencyGraph{
			nodes: map[string][]string{
				"app": {"lib"},
				"lib": {},
			},
			versions: map[string]string{
				"app": "1.0.0",
				"lib": "2.0.0",
			},
		}

		order, err := resolver.TopologicalSort(graph)
		require.NoError(t, err)
		assert.Equal(t, []string{"lib", "app"}, order.Tools)
		assert.Equal(t, "1.0.0", order.Versions["app"])
		assert.Equal(t, "2.0.0", order.Versions["lib"])
	})

	t.Run("sort diamond dependency", func(t *testing.T) {
		graph := &DependencyGraph{
			nodes: map[string][]string{
				"app":    {"lib1", "lib2"},
				"lib1":   {"common"},
				"lib2":   {"common"},
				"common": {},
			},
			versions: make(map[string]string),
		}

		order, err := resolver.TopologicalSort(graph)
		require.NoError(t, err)
		assert.Len(t, order.Tools, 4)

		// common must come before lib1 and lib2
		commonIdx := indexOf(order.Tools, "common")
		lib1Idx := indexOf(order.Tools, "lib1")
		lib2Idx := indexOf(order.Tools, "lib2")
		appIdx := indexOf(order.Tools, "app")

		assert.True(t, commonIdx < lib1Idx)
		assert.True(t, commonIdx < lib2Idx)
		assert.True(t, lib1Idx < appIdx)
		assert.True(t, lib2Idx < appIdx)
	})

	t.Run("sort complex graph", func(t *testing.T) {
		graph := &DependencyGraph{
			nodes: map[string][]string{
				"app":  {"web", "db"},
				"web":  {"http", "json"},
				"db":   {"sql"},
				"http": {},
				"json": {},
				"sql":  {},
			},
			versions: make(map[string]string),
		}

		order, err := resolver.TopologicalSort(graph)
		require.NoError(t, err)
		assert.Len(t, order.Tools, 6)

		// Verify dependencies come before dependents
		appIdx := indexOf(order.Tools, "app")
		webIdx := indexOf(order.Tools, "web")
		dbIdx := indexOf(order.Tools, "db")
		httpIdx := indexOf(order.Tools, "http")
		jsonIdx := indexOf(order.Tools, "json")
		sqlIdx := indexOf(order.Tools, "sql")

		assert.True(t, webIdx < appIdx)
		assert.True(t, dbIdx < appIdx)
		assert.True(t, httpIdx < webIdx)
		assert.True(t, jsonIdx < webIdx)
		assert.True(t, sqlIdx < dbIdx)
	})

	t.Run("error on circular dependency", func(t *testing.T) {
		graph := &DependencyGraph{
			nodes: map[string][]string{
				"a": {"b"},
				"b": {"a"},
			},
			versions: make(map[string]string),
		}

		_, err := resolver.TopologicalSort(graph)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "circular dependency detected")
	})
}

// Test ResolveVersionConstraints

func TestDependencyResolver_ResolveVersionConstraints(t *testing.T) {
	ctx := context.Background()

	t.Run("resolve compatible versions", func(t *testing.T) {
		providerRegistry := &mockDependencyProviderRegistry{
			providers: map[string]ProviderMetadata{
				"app": {
					Name: "app",
					Dependencies: []Dependency{
						{Tool: "lib", VersionConstraint: "1.0.0"},
					},
				},
				"lib": {
					Name:         "lib",
					Dependencies: []Dependency{},
				},
			},
		}

		backendRegistry := &mockDependencyBackendRegistry{
			versions: map[string]map[string]string{
				"app": {
					"2.0.0": "2.0.0",
				},
				"lib": {
					"1.0.0": "1.0.0",
				},
			},
		}

		resolver := NewDependencyResolver(providerRegistry, backendRegistry, nil)

		graph := &DependencyGraph{
			nodes: map[string][]string{
				"app": {"lib"},
				"lib": {},
			},
			versions: make(map[string]string),
		}

		requestedVersions := map[string]string{
			"app": "2.0.0",
		}

		err := resolver.ResolveVersionConstraints(ctx, graph, requestedVersions)
		require.NoError(t, err)
		assert.Equal(t, "1.0.0", graph.versions["lib"])
		assert.Equal(t, "2.0.0", graph.versions["app"])
	})

	t.Run("resolve same version from multiple requesters", func(t *testing.T) {
		providerRegistry := &mockDependencyProviderRegistry{
			providers: map[string]ProviderMetadata{
				"app1": {
					Name: "app1",
					Dependencies: []Dependency{
						{Tool: "common", VersionConstraint: "1.0.0"},
					},
				},
				"app2": {
					Name: "app2",
					Dependencies: []Dependency{
						{Tool: "common", VersionConstraint: "1.0.0"},
					},
				},
				"common": {
					Name:         "common",
					Dependencies: []Dependency{},
				},
			},
		}

		backendRegistry := &mockDependencyBackendRegistry{
			versions: map[string]map[string]string{
				"app1": {
					"1.0.0": "1.0.0",
				},
				"app2": {
					"2.0.0": "2.0.0",
				},
				"common": {
					"1.0.0": "1.0.0",
				},
			},
		}

		resolver := NewDependencyResolver(providerRegistry, backendRegistry, nil)

		graph := &DependencyGraph{
			nodes: map[string][]string{
				"app1":   {"common"},
				"app2":   {"common"},
				"common": {},
			},
			versions: make(map[string]string),
		}

		requestedVersions := map[string]string{
			"app1": "1.0.0",
			"app2": "2.0.0",
		}

		err := resolver.ResolveVersionConstraints(ctx, graph, requestedVersions)
		require.NoError(t, err)
		assert.Equal(t, "1.0.0", graph.versions["common"])
	})

	t.Run("detect version conflict", func(t *testing.T) {
		providerRegistry := &mockDependencyProviderRegistry{
			providers: map[string]ProviderMetadata{
				"app1": {
					Name: "app1",
					Dependencies: []Dependency{
						{Tool: "common", VersionConstraint: "1.0.0"},
					},
				},
				"app2": {
					Name: "app2",
					Dependencies: []Dependency{
						{Tool: "common", VersionConstraint: "2.0.0"},
					},
				},
				"common": {
					Name:         "common",
					Dependencies: []Dependency{},
				},
			},
		}

		backendRegistry := &mockDependencyBackendRegistry{
			versions: map[string]map[string]string{
				"app1": {
					"1.0.0": "1.0.0",
				},
				"app2": {
					"1.0.0": "1.0.0",
					"2.0.0": "2.0.0",
				},
				"common": {
					"1.0.0": "1.0.0",
					"2.0.0": "2.0.0",
				},
			},
		}

		resolver := NewDependencyResolver(providerRegistry, backendRegistry, nil)

		graph := &DependencyGraph{
			nodes: map[string][]string{
				"app1":   {"common"},
				"app2":   {"common"},
				"common": {},
			},
			versions: make(map[string]string),
		}

		requestedVersions := map[string]string{
			"app1": "1.0.0",
			"app2": "2.0.0",
		}

		err := resolver.ResolveVersionConstraints(ctx, graph, requestedVersions)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "version conflicts detected")
		assert.Contains(t, err.Error(), "common")
	})
}

// Test ResolveDependencies (integration)

func TestDependencyResolver_ResolveDependencies(t *testing.T) {
	ctx := context.Background()

	t.Run("resolve complete dependency tree", func(t *testing.T) {
		providerRegistry := &mockDependencyProviderRegistry{
			providers: map[string]ProviderMetadata{
				"app": {
					Name: "app",
					Dependencies: []Dependency{
						{Tool: "lib1", VersionConstraint: "1.0.0"},
						{Tool: "lib2", VersionConstraint: "2.0.0"},
					},
				},
				"lib1": {
					Name: "lib1",
					Dependencies: []Dependency{
						{Tool: "common", VersionConstraint: "1.0.0"},
					},
				},
				"lib2": {
					Name: "lib2",
					Dependencies: []Dependency{
						{Tool: "common", VersionConstraint: "1.0.0"},
					},
				},
				"common": {
					Name:         "common",
					Dependencies: []Dependency{},
				},
			},
		}

		backendRegistry := &mockDependencyBackendRegistry{
			versions: map[string]map[string]string{
				"app":    {"latest": "3.0.0"},
				"lib1":   {"1.0.0": "1.0.0"},
				"lib2":   {"2.0.0": "2.0.0"},
				"common": {"1.0.0": "1.0.0"},
			},
		}

		resolver := NewDependencyResolver(providerRegistry, backendRegistry, nil)

		requestedVersions := map[string]string{
			"app": "latest",
		}

		order, err := resolver.ResolveDependencies(ctx, []string{"app"}, requestedVersions)
		require.NoError(t, err)
		assert.NotNil(t, order)
		assert.Len(t, order.Tools, 4)

		// Verify installation order
		commonIdx := indexOf(order.Tools, "common")
		lib1Idx := indexOf(order.Tools, "lib1")
		lib2Idx := indexOf(order.Tools, "lib2")
		appIdx := indexOf(order.Tools, "app")

		assert.True(t, commonIdx < lib1Idx)
		assert.True(t, commonIdx < lib2Idx)
		assert.True(t, lib1Idx < appIdx)
		assert.True(t, lib2Idx < appIdx)

		// Verify versions
		assert.Equal(t, "3.0.0", order.Versions["app"])
		assert.Equal(t, "1.0.0", order.Versions["lib1"])
		assert.Equal(t, "2.0.0", order.Versions["lib2"])
		assert.Equal(t, "1.0.0", order.Versions["common"])
	})

	t.Run("error on circular dependency", func(t *testing.T) {
		providerRegistry := &mockDependencyProviderRegistry{
			providers: map[string]ProviderMetadata{
				"a": {
					Name: "a",
					Dependencies: []Dependency{
						{Tool: "b", VersionConstraint: "1.0.0"},
					},
				},
				"b": {
					Name: "b",
					Dependencies: []Dependency{
						{Tool: "a", VersionConstraint: "1.0.0"},
					},
				},
			},
		}

		backendRegistry := &mockDependencyBackendRegistry{
			versions: map[string]map[string]string{
				"a": {"1.0.0": "1.0.0"},
				"b": {"1.0.0": "1.0.0"},
			},
		}

		resolver := NewDependencyResolver(providerRegistry, backendRegistry, nil)

		_, err := resolver.ResolveDependencies(ctx, []string{"a"}, map[string]string{"a": "1.0.0"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "circular dependency detected")
	})

	t.Run("error on version conflict", func(t *testing.T) {
		providerRegistry := &mockDependencyProviderRegistry{
			providers: map[string]ProviderMetadata{
				"app1": {
					Name: "app1",
					Dependencies: []Dependency{
						{Tool: "lib", VersionConstraint: "1.0.0"},
					},
				},
				"app2": {
					Name: "app2",
					Dependencies: []Dependency{
						{Tool: "lib", VersionConstraint: "2.0.0"},
					},
				},
				"lib": {
					Name:         "lib",
					Dependencies: []Dependency{},
				},
			},
		}

		backendRegistry := &mockDependencyBackendRegistry{
			versions: map[string]map[string]string{
				"app1": {"1.0.0": "1.0.0"},
				"app2": {"1.0.0": "1.0.0"},
				"lib":  {"1.0.0": "1.0.0", "2.0.0": "2.0.0"},
			},
		}

		resolver := NewDependencyResolver(providerRegistry, backendRegistry, nil)

		requestedVersions := map[string]string{
			"app1": "1.0.0",
			"app2": "1.0.0",
		}

		_, err := resolver.ResolveDependencies(ctx, []string{"app1", "app2"}, requestedVersions)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "version conflicts detected")
	})
}

// Helper function

func indexOf(slice []string, value string) int {
	for i, v := range slice {
		if v == value {
			return i
		}
	}
	return -1
}
