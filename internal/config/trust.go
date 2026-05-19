// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
)

// TrustStatus represents the current trust state of a configuration file
type TrustStatus int

const (
	// TrustStatusUntrusted means the file has never been trusted
	TrustStatusUntrusted TrustStatus = iota
	// TrustStatusTrusted means the file is trusted and its hash matches
	TrustStatusTrusted
	// TrustStatusModified means the file was trusted previously but its contents have changed
	TrustStatusModified
)

// TrustManager handles the trust mechanism for local and project configurations.
// To prevent executing malicious scripts from automatically loaded configuration files
// in untrusted repositories, we maintain a list of trusted config absolute paths and their content hashes.
type TrustManager interface {
	// TrustStatus returns the current trust state of the configuration file.
	TrustStatus(path string) TrustStatus

	// Trust adds the specified configuration file to the trusted list with its current hash.
	Trust(path string) error

	// Untrust removes the specified configuration file from the trusted list.
	Untrust(path string) error

	// List returns all currently trusted configuration files.
	List() (map[string]string, error)
}

type fileTrustManager struct {
	trustFilePath string
	mu            sync.RWMutex
}

// NewTrustManager creates a new TrustManager that persists trusted paths
// to ~/.config/unirtm/trusted_configs.
func NewTrustManager() TrustManager {
	trustFilePath := filepath.Join(env.GetConfigDir(), "trusted_configs")
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

// loadTrustedPaths reads the trusted_configs file and returns a map of path -> hash.
func (m *fileTrustManager) loadTrustedPaths() (map[string]string, error) {
	if err := m.ensureTrustFileExists(); err != nil {
		return nil, err
	}

	f, err := os.Open(m.trustFilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	paths := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			paths[parts[0]] = parts[1]
		} else {
			// Legacy support or missing hash: store empty hash to force a "Modified" status
			paths[parts[0]] = ""
		}
	}

	return paths, scanner.Err()
}

// saveTrustedPaths writes the trusted paths map back to the file.
func (m *fileTrustManager) saveTrustedPaths(paths map[string]string) error {
	if err := m.ensureTrustFileExists(); err != nil {
		return err
	}

	f, err := os.OpenFile(m.trustFilePath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	for path, hash := range paths {
		if _, err := writer.WriteString(fmt.Sprintf("%s %s\n", path, hash)); err != nil {
			return err
		}
	}
	return writer.Flush()
}

// calculateHash computes the SHA256 hash of a file's contents.
func calculateHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func (m *fileTrustManager) TrustStatus(path string) TrustStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return TrustStatusUntrusted
	}

	paths, err := m.loadTrustedPaths()
	if err != nil {
		return TrustStatusUntrusted
	}

	storedHash, exists := paths[absPath]
	if !exists {
		return TrustStatusUntrusted
	}

	currentHash, err := calculateHash(absPath)
	if err != nil {
		// If we can't calculate the hash (e.g. file deleted), we treat it as untrusted
		return TrustStatusUntrusted
	}

	if storedHash == "" || storedHash != currentHash {
		return TrustStatusModified
	}

	return TrustStatusTrusted
}

func (m *fileTrustManager) Trust(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	hash, err := calculateHash(absPath)
	if err != nil {
		return fmt.Errorf("failed to calculate file hash: %w", err)
	}

	paths, err := m.loadTrustedPaths()
	if err != nil {
		return fmt.Errorf("failed to load trusted paths: %w", err)
	}

	if paths[absPath] == hash {
		// Already trusted and up to date
		return nil
	}

	paths[absPath] = hash
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

	if _, exists := paths[absPath]; !exists {
		// Already untrusted
		return nil
	}

	delete(paths, absPath)
	return m.saveTrustedPaths(paths)
}

func (m *fileTrustManager) List() (map[string]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.loadTrustedPaths()
}

