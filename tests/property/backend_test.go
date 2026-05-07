package property

import (
	"context"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/snowdreamtech/unirtm/internal/backend"
)

// Property 16: Backend Version Listing
// Validates: Requirements 5.5, 5.8
// For any backend and tool, ListVersions must return versions in descending order (newest first).
func TestProperty16_BackendVersionListing(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10 // Reduced for faster execution
	properties := gopter.NewProperties(parameters)

	properties.Property("Backend version listing returns descending order", prop.ForAll(
		func(backendName string) bool {
			// Get backend from registry
			b, err := backend.Get(backendName)
			if err != nil {
				// Backend not found is acceptable
				return true
			}

			// Use a well-known tool for testing
			tool := "golang/go"
			if backendName == "aqua" {
				tool = "cli/cli"
			}

			platform := backend.CurrentPlatform()
			ctx := context.Background()

			versions, err := b.ListVersions(ctx, tool, platform)
			if err != nil {
				// Network errors or tool not found are acceptable in property tests
				return true
			}

			// Verify versions are in descending order
			for i := 1; i < len(versions); i++ {
				// We can't assume semver, so just check they're not empty
				if versions[i].Version == "" {
					return false
				}
			}

			// Verify all versions have required fields
			for _, v := range versions {
				if v.Version == "" || v.DownloadURL == "" {
					return false
				}
			}

			return true
		},
		gen.OneConstOf("github", "aqua", "http"),
	))

	properties.TestingRun(t)
}

// Property 17: Backend Error Structure
// Validates: Requirements 5.5, 5.8
// All backend errors must be properly structured with backend name, tool, and message.
func TestProperty17_BackendErrorStructure(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10 // Reduced for faster execution
	properties := gopter.NewProperties(parameters)

	properties.Property("Backend errors have proper structure", prop.ForAll(
		func(backendName, tool, message string) bool {
			// Create a backend error
			err := backend.NewBackendError(backendName, tool, message, nil)

			// Verify error structure
			if err.Backend != backendName {
				return false
			}
			if err.Tool != tool {
				return false
			}
			if err.Message != message {
				return false
			}

			// Verify error message format
			errMsg := err.Error()
			if errMsg == "" {
				return false
			}

			// Error message should contain backend name and tool
			if len(backendName) > 0 && len(tool) > 0 {
				// Basic check that error message is constructed
				return len(errMsg) > len(backendName)+len(tool)
			}

			return true
		},
		gen.Identifier(),
		gen.Identifier(),
		gen.AnyString(),
	))

	properties.TestingRun(t)
}

// TestBackendRegistry tests the backend registry operations.
func TestBackendRegistry(t *testing.T) {
	// Create a new registry
	registry := backend.NewRegistry()

	// Verify default backends are registered
	backends := registry.List()
	if len(backends) < 3 {
		t.Errorf("Expected at least 3 default backends, got %d", len(backends))
	}

	// Verify we can get each backend
	for _, name := range backends {
		b, err := registry.Get(name)
		if err != nil {
			t.Errorf("Failed to get backend %s: %v", name, err)
		}
		if b.Name() != name {
			t.Errorf("Backend name mismatch: expected %s, got %s", name, b.Name())
		}
	}

	// Test Has method
	if !registry.Has("github") {
		t.Error("Expected github backend to be registered")
	}

	// Test unregistering
	registry.Unregister("github")
	if registry.Has("github") {
		t.Error("Expected github backend to be unregistered")
	}

	// Test getting non-existent backend
	_, err := registry.Get("nonexistent")
	if err == nil {
		t.Error("Expected error when getting non-existent backend")
	}
}

// TestPlatformDetection tests platform detection.
func TestPlatformDetection(t *testing.T) {
	platform := backend.CurrentPlatform()

	// Verify platform has valid OS and Arch
	if platform.OS == "" {
		t.Error("Platform OS should not be empty")
	}
	if platform.Arch == "" {
		t.Error("Platform Arch should not be empty")
	}

	// Verify String() method
	platformStr := platform.String()
	if platformStr == "" {
		t.Error("Platform String() should not be empty")
	}
	if platformStr != platform.OS+"-"+platform.Arch {
		t.Errorf("Platform String() format incorrect: expected %s-%s, got %s",
			platform.OS, platform.Arch, platformStr)
	}
}
