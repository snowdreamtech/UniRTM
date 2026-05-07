// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service_test

import (
	"context"
	"fmt"
	"log"

	"github.com/snowdreamtech/unirtm/internal/service"
)

// This example demonstrates basic tool index management operations
func ExampleIndexManager_basic() {
	// Note: This example uses mock implementations for demonstration
	// In production, use actual repository and backend implementations

	ctx := context.Background()

	// Create index manager (using mocks for example)
	indexManager, err := service.NewIndexManager(
		nil, // indexRepo - use actual implementation
		nil, // auditRepo - use actual implementation
		nil, // backends - register actual backends
		service.IndexManagerConfig{},
	)
	if err != nil {
		log.Fatal(err)
	}

	// Add a tool to the index
	err = indexManager.UpsertTool(ctx,
		"node",
		"Node.js JavaScript runtime",
		"https://nodejs.org",
		"MIT",
		"github",
		&service.ToolMetadata{
			AvailableVersions: []string{"20.0.0", "18.0.0", "16.0.0"},
			Tags:              []string{"runtime", "javascript", "nodejs"},
			Stars:             95000,
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	// Retrieve the tool
	entry, err := indexManager.GetTool(ctx, "node")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Tool: %s\n", entry.Tool)
	fmt.Printf("Description: %s\n", entry.Description)
	fmt.Printf("Backend: %s\n", entry.Backend)
}

// This example demonstrates searching for tools
func ExampleIndexManager_search() {
	ctx := context.Background()

	// Create index manager (using mocks for example)
	indexManager, err := service.NewIndexManager(
		nil, // indexRepo
		nil, // auditRepo
		nil, // backends
		service.IndexManagerConfig{},
	)
	if err != nil {
		log.Fatal(err)
	}

	// Search for tools
	results, err := indexManager.SearchTools(ctx, service.SearchOptions{
		Query: "javascript",
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, tool := range results {
		fmt.Printf("%s: %s\n", tool.Tool, tool.Description)
	}
}

// This example demonstrates filtering tools by backend
func ExampleIndexManager_filterByBackend() {
	ctx := context.Background()

	// Create index manager
	indexManager, err := service.NewIndexManager(
		nil, // indexRepo
		nil, // auditRepo
		nil, // backends
		service.IndexManagerConfig{},
	)
	if err != nil {
		log.Fatal(err)
	}

	// Filter tools by backend
	githubTools, err := indexManager.FilterByBackend(ctx, "github")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d tools from GitHub backend\n", len(githubTools))
}

// This example demonstrates stale index detection
func ExampleIndexManager_staleDetection() {
	ctx := context.Background()

	// Create index manager with 7-day stale timeout
	indexManager, err := service.NewIndexManager(
		nil, // indexRepo
		nil, // auditRepo
		nil, // backends
		service.IndexManagerConfig{
			StaleTimeout: 7 * 24 * 60 * 60 * 1000000000, // 7 days in nanoseconds
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	// Check if index is stale
	isStale, err := indexManager.IsStale(ctx)
	if err != nil {
		log.Fatal(err)
	}

	if isStale {
		fmt.Println("Index is stale, updating...")
		err = indexManager.UpdateFromAllBackends(ctx)
		if err != nil {
			log.Printf("Failed to update index: %v", err)
		}
	}
}

// This example demonstrates offline operation
func ExampleIndexManager_offline() {
	ctx := context.Background()

	// Create index manager
	indexManager, err := service.NewIndexManager(
		nil, // indexRepo
		nil, // auditRepo
		nil, // backends
		service.IndexManagerConfig{},
	)
	if err != nil {
		log.Fatal(err)
	}

	// Check if offline operation is possible
	capable, err := indexManager.IsOfflineCapable(ctx)
	if err != nil {
		log.Fatal(err)
	}

	if !capable {
		fmt.Println("No cached index available. Please run 'unirtm index update' when online.")
		return
	}

	// Search offline
	results, err := indexManager.SearchTools(ctx, service.SearchOptions{
		Query: "python",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d tools matching 'python'\n", len(results))
}

// This example demonstrates pagination
func ExampleIndexManager_pagination() {
	ctx := context.Background()

	// Create index manager
	indexManager, err := service.NewIndexManager(
		nil, // indexRepo
		nil, // auditRepo
		nil, // backends
		service.IndexManagerConfig{},
	)
	if err != nil {
		log.Fatal(err)
	}

	// Search with pagination
	pageSize := 10
	page := 0

	for {
		results, err := indexManager.SearchTools(ctx, service.SearchOptions{
			Query:  "runtime",
			Limit:  pageSize,
			Offset: page * pageSize,
		})
		if err != nil {
			log.Fatal(err)
		}

		if len(results) == 0 {
			break
		}

		fmt.Printf("Page %d: %d results\n", page+1, len(results))
		for _, tool := range results {
			fmt.Printf("  - %s\n", tool.Tool)
		}

		page++
	}
}

// This example demonstrates getting tool metadata
func ExampleIndexManager_metadata() {
	ctx := context.Background()

	// Create index manager
	indexManager, err := service.NewIndexManager(
		nil, // indexRepo
		nil, // auditRepo
		nil, // backends
		service.IndexManagerConfig{},
	)
	if err != nil {
		log.Fatal(err)
	}

	// Get tool metadata
	metadata, err := indexManager.GetToolMetadata(ctx, "node")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Available versions: %v\n", metadata.AvailableVersions)
	fmt.Printf("Tags: %v\n", metadata.Tags)
	fmt.Printf("Stars: %d\n", metadata.Stars)
}

// This example demonstrates backend management
func ExampleIndexManager_backendManagement() {
	// Create index manager
	indexManager, err := service.NewIndexManager(
		nil, // indexRepo
		nil, // auditRepo
		nil, // backends
		service.IndexManagerConfig{},
	)
	if err != nil {
		log.Fatal(err)
	}

	// Register a backend
	// githubBackend := backend.NewGitHubBackend(...)
	// indexManager.RegisterBackend("github", githubBackend)

	// List registered backends
	backends := indexManager.ListBackends()
	fmt.Printf("Registered backends: %v\n", backends)

	// Unregister a backend
	indexManager.UnregisterBackend("github")
}
