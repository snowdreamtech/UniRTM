// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// This example demonstrates how to use the UniRTM multi-backend and provider
// registry to dynamically route tool installation requests across different ecosystems.
//
// Usage: go run main.go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/provider"
)

func main() {
	// 1. Initialize Registries
	backendRegistry := backend.NewRegistry()
	providerRegistry := provider.NewRegistry()

	// 2. Define target ecosystems (Backends) and packages
	toolsToTest := []struct {
		Backend string
		Tool    string
		Version string
	}{
		{"npm", "typescript", "5.0.0"},
		{"pypi", "black", "23.3.0"},
		{"cargo", "ripgrep", "13.0.0"},
		{"asdf", "nodejs", "20.0.0"},
	}

	ctx := context.Background()
	platform := backend.Platform{OS: "linux", Arch: "amd64"}

	// 3. Demonstrate resolution for each tool
	for _, target := range toolsToTest {
		fmt.Printf("\n--- Ecosystem: %s | Package: %s@%s ---\n", target.Backend, target.Tool, target.Version)

		// Get the appropriate backend
		b, err := backendRegistry.Get(target.Backend)
		if err != nil {
			log.Printf("Failed to get backend %s: %v", target.Backend, err)
			continue
		}
		fmt.Printf("[Backend] Successfully retrieved backend: %s\n", b.Name())

		// Resolve version (This tests the API parsing logic: registry.npmjs.org / pypi.org / crates.io)
		// We use ResolveVersion to mimic looking up the download info.
		// Note: we just verify the object is returned, we ignore networking errors
		// so this example runs cleanly even if offline.
		versionInfo, err := b.ResolveVersion(ctx, target.Tool, target.Version, platform)
		if err != nil {
			fmt.Printf("[Backend] Network resolution error (expected if offline): %v\n", err)
		} else {
			fmt.Printf("[Backend] Resolved Version: %s\n", versionInfo.Version)
		}

		// Get the appropriate provider
		// Notice how we use GetWithBackend so that 'npm:typescript' maps strictly to the 'npm' provider,
		// rather than trying to look for a specific 'typescript' Go-based provider.
		p := providerRegistry.GetWithBackend(target.Tool, target.Backend)
		fmt.Printf("[Provider] Selected Provider: %s\n", p.Name())
		fmt.Printf("[Provider] The provider will execute '%s' specific sandbox installation hooks.\n", p.Name())
	}

	fmt.Println("\nAll ecosystem providers and backends loaded successfully!")
}
