// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// TrustManager handles the trust mechanism for local and project configurations.
// To prevent executing malicious scripts from automatically loaded configuration files
// in untrusted repositories, we maintain a list of trusted config absolute paths.
type TrustManager interface {
	// IsTrusted returns true if the specified configuration file is trusted.
	IsTrusted(path string) bool

	// Trust adds the specified configuration file to the trusted list.
	Trust(path string) error

	// Untrust removes the specified configuration file from the trusted list.
	Untrust(path string) error
}

type fileTrustManager struct {
	trustFilePath string
	mu            sync.RWMutex
}

// NewTrustManager creates a new TrustManager that persists trusted paths
// to ~/.config/unirtm/trusted_configs.
func NewTrustManager() TrustManager {
	homeDir, _ := os.UserHomeDir()
	trustFilePath := filepath.Join(homeDir, ".config", "unirtm", "trusted_configs")
	return &fileTrustManager{
		trustFilePath: trustFilePath,
	}
}

// ensureTrustFileExists creates the directory and file if they do not exist.
func (m *fileTrustManager) ensureTrustFileExists() error {
	dir := filepath.Dir(m.trustFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(m.trustFilePath, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	f.Close()
	return nil
}

// loadTrustedPaths reads the trusted_configs file and returns a map for quick lookup.
func (m *fileTrustManager) loadTrustedPaths() (map[string]bool, error) {
	if err := m.ensureTrustFileExists(); err != nil {
		return nil, err
	}

	f, err := os.Open(m.trustFilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	paths := make(map[string]bool)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			paths[line] = true
		}
	}

	return paths, scanner.Err()
}

// saveTrustedPaths writes the trusted paths map back to the file.
func (m *fileTrustManager) saveTrustedPaths(paths map[string]bool) error {
	if err := m.ensureTrustFileExists(); err != nil {
		return err
	}

	f, err := os.OpenFile(m.trustFilePath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	for path := range paths {
		if _, err := writer.WriteString(path + "\n"); err != nil {
			return err
		}
	}
	return writer.Flush()
}

func (m *fileTrustManager) IsTrusted(path string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	paths, err := m.loadTrustedPaths()
	if err != nil {
		return false
	}

	return paths[absPath]
}

func (m *fileTrustManager) Trust(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	paths, err := m.loadTrustedPaths()
	if err != nil {
		return fmt.Errorf("failed to load trusted paths: %w", err)
	}

	if paths[absPath] {
		// Already trusted
		return nil
	}

	paths[absPath] = true
	return m.saveTrustedPaths(paths)
}

func (m *fileTrustManager) Untrust(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	paths, err := m.loadTrustedPaths()
	if err != nil {
		return fmt.Errorf("failed to load trusted paths: %w", err)
	}

	if !paths[absPath] {
		// Already untrusted
		return nil
	}

	delete(paths, absPath)
	return m.saveTrustedPaths(paths)
}
