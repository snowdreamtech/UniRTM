// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"fmt"
	"sync"
)

// Registry manages backend implementations and provides discovery by name.
type Registry struct {
	mu       sync.RWMutex
	backends map[string]Backend
}

// NewRegistry creates a new backend registry with default backends registered.
func NewRegistry() *Registry {
	r := &Registry{
		backends: make(map[string]Backend),
	}

	// Register default backends
	r.Register(NewGitHubBackend())
	r.Register(NewAquaBackend())
	r.Register(NewHTTPBackend())
	r.Register(NewAsdfBackend())
	r.Register(NewNpmBackend())
	r.Register(NewPypiBackend())
	r.Register(NewCargoBackend())
	r.Register(NewGemBackend())
	r.Register(NewDotnetBackend())
	r.Register(NewCondaBackend())
	r.Register(NewGitlabBackend())
	r.Register(NewForgejoBackend())
	r.Register(NewVfoxBackend())
	r.Register(NewSpmBackend())
	r.Register(NewS3Backend())

	return r
}

// Register registers a backend in the registry.
// If a backend with the same name already exists, it will be replaced.
func (r *Registry) Register(backend Backend) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.backends[backend.Name()] = backend
}

// Unregister removes a backend from the registry.
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.backends, name)
}

// Get retrieves a backend by name.
// Returns an error if the backend is not found.
func (r *Registry) Get(name string) (Backend, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	backend, ok := r.backends[name]
	if !ok {
		return nil, fmt.Errorf("backend not found: %s", name)
	}

	return backend, nil
}

// List returns a list of all registered backend names.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.backends))
	for name := range r.backends {
		names = append(names, name)
	}

	return names
}

// Has checks if a backend is registered.
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.backends[name]
	return ok
}

// DefaultRegistry is the global default backend registry.
var DefaultRegistry = NewRegistry()

// Register registers a backend in the default registry.
func Register(backend Backend) {
	DefaultRegistry.Register(backend)
}

// Unregister removes a backend from the default registry.
func Unregister(name string) {
	DefaultRegistry.Unregister(name)
}

// Get retrieves a backend from the default registry.
func Get(name string) (Backend, error) {
	return DefaultRegistry.Get(name)
}

// List returns all registered backend names from the default registry.
func List() []string {
	return DefaultRegistry.List()
}

// Has checks if a backend is registered in the default registry.
func Has(name string) bool {
	return DefaultRegistry.Has(name)
}
