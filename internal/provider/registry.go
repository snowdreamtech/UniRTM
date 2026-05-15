// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"fmt"
	"strings"
	"sync"
)

// Registry manages provider implementations and provides discovery by tool name.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider
	generic   Provider
}

// NewRegistry creates a new provider registry with default providers registered.
func NewRegistry() *Registry {
	nativeProvider := NewNativeProvider()
	r := &Registry{
		providers: make(map[string]Provider),
		generic:   NewGenericProvider(),
	}

	// Register default providers
	r.Register("native", nativeProvider)
	r.Register("node", NewNodeProvider())
	r.Register("nodejs", NewNodeProvider())
	r.Register("python", NewPythonProvider())
	r.Register("go", NewGolangProvider())
	r.Register("golang", NewGolangProvider())
	r.Register("go-pkg", NewGoPkgProvider())
	r.Register("java", NewJavaProvider())
	r.Register("jdk", NewJavaProvider())
	r.Register("jre", NewJavaProvider())
	r.Register("ruby", NewRubyProvider(nativeProvider))
	r.Register("rust", NewRustProvider())
	r.Register("asdf", NewAsdfProvider())
	r.Register("npm", NewNpmProvider())
	r.Register("pypi", NewPypiProvider())
	r.Register("cargo", NewCargoProvider())
	r.Register("ubi", NewUbiProvider())
	r.Register("gem", NewGemProvider())
	r.Register("dotnet", NewDotnetProvider())
	r.Register("conda", NewCondaProvider())
	r.Register("vfox", NewVfoxProvider())
	r.Register("spm", NewSpmProvider())
	r.Register("bun", NewBunProvider())
	r.Register("deno", NewDenoProvider())
	r.Register("elixir", NewElixirProvider())
	r.Register("erlang", NewErlangProvider())
	r.Register("swift", NewSwiftProvider())
	r.Register("zig", NewZigProvider())
	r.Register("php", r.providers["asdf"])
	r.Register("flutter", NewFlutterProvider())
	r.Register("pipx", NewPypiProvider())
	r.Register("terraform", NewGenericProvider())
	r.Register("opentofu", NewGenericProvider())
	r.Register("tofu", NewGenericProvider())
	r.Register("kubectl", NewGenericProvider())
	r.Register("maven", NewGenericProvider())
	r.Register("gradle", NewGenericProvider())
	
	return r
}

// Register registers a provider for a specific tool name.
// If a provider with the same name already exists, it will be replaced.
func (r *Registry) Register(toolName string, provider Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[strings.ToLower(toolName)] = provider
}

// Unregister removes a provider for a specific tool name.
func (r *Registry) Unregister(toolName string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.providers, strings.ToLower(toolName))
}

// Get retrieves a provider for a specific tool name.
// If no specific provider is found, returns the generic provider.
func (r *Registry) Get(toolName string) Provider {
	return r.GetWithBackend(toolName, "")
}

// GetWithBackend retrieves a provider for a specific tool name or backend name.
func (r *Registry) GetWithBackend(toolName string, backendName string) Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Try backend match (e.g., "asdf", "npm", "cargo")
	if backendName != "" {
		if provider, ok := r.providers[strings.ToLower(backendName)]; ok {
			return provider
		}
	}

	// Try exact match
	if provider, ok := r.providers[strings.ToLower(toolName)]; ok {
		return provider
	}

	// Try partial match (e.g., "node@18" -> "node")
	baseName := strings.Split(toolName, "@")[0]
	baseName = strings.Split(baseName, "/")[0]
	if provider, ok := r.providers[strings.ToLower(baseName)]; ok {
		return provider
	}

	// Try colon-prefix backend match (e.g., "pipx:clang-format" -> "pipx", "npm:prettier" -> "npm")
	if idx := strings.Index(toolName, ":"); idx != -1 {
		backend := strings.ToLower(toolName[:idx])
		if provider, ok := r.providers[backend]; ok {
			return provider
		}
	}

	// Fallback to generic provider
	return r.generic
}

// GetExact retrieves a provider for a specific tool name without fallback.
// Returns an error if no specific provider is found.
func (r *Registry) GetExact(toolName string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, ok := r.providers[strings.ToLower(toolName)]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", toolName)
	}

	return provider, nil
}

// List returns a list of all registered tool names.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}

	return names
}

// Has checks if a provider is registered for a specific tool name.
func (r *Registry) Has(toolName string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.providers[strings.ToLower(toolName)]
	return ok
}

// DefaultRegistry is the global default provider registry.
var DefaultRegistry = NewRegistry()

// Register registers a provider in the default registry.
func Register(toolName string, provider Provider) {
	DefaultRegistry.Register(toolName, provider)
}

// Unregister removes a provider from the default registry.
func Unregister(toolName string) {
	DefaultRegistry.Unregister(toolName)
}

// Get retrieves a provider from the default registry.
func Get(toolName string) Provider {
	return DefaultRegistry.Get(toolName)
}

// GetExact retrieves a provider from the default registry without fallback.
func GetExact(toolName string) (Provider, error) {
	return DefaultRegistry.GetExact(toolName)
}

// List returns all registered tool names from the default registry.
func List() []string {
	return DefaultRegistry.List()
}

// Has checks if a provider is registered in the default registry.
func Has(toolName string) bool {
	return DefaultRegistry.Has(toolName)
}
