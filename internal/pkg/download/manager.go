// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package download

import (
	"fmt"
	"sync"
)

// Manager manages multiple downloader implementations and provides
// a registry for custom downloaders.
type Manager struct {
	mu          sync.RWMutex
	downloaders map[string]Downloader
}

// NewManager creates a new download manager with the default HTTP downloader registered.
func NewManager() *Manager {
	m := &Manager{
		downloaders: make(map[string]Downloader),
	}

	// Register default HTTP downloader
	m.Register("http", NewHTTPDownloader())
	m.Register("https", NewHTTPDownloader())

	return m
}

// Register registers a custom downloader for a specific scheme (e.g., "http", "https", "s3", "ftp").
// If a downloader with the same scheme already exists, it will be replaced.
func (m *Manager) Register(scheme string, downloader Downloader) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.downloaders[scheme] = downloader
}

// Unregister removes a downloader for a specific scheme.
func (m *Manager) Unregister(scheme string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.downloaders, scheme)
}

// Get retrieves a downloader for a specific scheme.
// Returns an error if no downloader is registered for the scheme.
func (m *Manager) Get(scheme string) (Downloader, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	downloader, ok := m.downloaders[scheme]
	if !ok {
		return nil, fmt.Errorf("no downloader registered for scheme: %s", scheme)
	}

	return downloader, nil
}

// ListSchemes returns a list of all registered schemes.
func (m *Manager) ListSchemes() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	schemes := make([]string, 0, len(m.downloaders))
	for scheme := range m.downloaders {
		schemes = append(schemes, scheme)
	}

	return schemes
}

// HasScheme checks if a downloader is registered for a specific scheme.
func (m *Manager) HasScheme(scheme string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, ok := m.downloaders[scheme]
	return ok
}

// DefaultManager is the global default download manager instance.
var DefaultManager = NewManager()

// Register registers a downloader in the default manager.
func Register(scheme string, downloader Downloader) {
	DefaultManager.Register(scheme, downloader)
}

// Unregister removes a downloader from the default manager.
func Unregister(scheme string) {
	DefaultManager.Unregister(scheme)
}

// Get retrieves a downloader from the default manager.
func Get(scheme string) (Downloader, error) {
	return DefaultManager.Get(scheme)
}

// ListSchemes returns all registered schemes from the default manager.
func ListSchemes() []string {
	return DefaultManager.ListSchemes()
}

// HasScheme checks if a scheme is registered in the default manager.
func HasScheme(scheme string) bool {
	return DefaultManager.HasScheme(scheme)
}
