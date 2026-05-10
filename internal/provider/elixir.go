// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"os"
	"path/filepath"

	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
)

// ElixirProvider implements the Provider interface for Elixir.
type ElixirProvider struct {
}

// NewElixirProvider creates a new Elixir provider.
func NewElixirProvider() *ElixirProvider {
	return &ElixirProvider{}
}

func (p *ElixirProvider) Name() string {
	return "elixir"
}

func (p *ElixirProvider) Install(ctx context.Context, installPath string, artifactPath string, version string) error {
	logger.Debug("Installing Elixir", map[string]interface{}{"version": version, "installPath": installPath, "artifactPath": artifactPath})

	if err := os.MkdirAll(installPath, 0755); err != nil {
		return err
	}

	return nil
}

func (p *ElixirProvider) PostInstall(ctx context.Context, installPath string, version string) error {
	return nil
}

func (p *ElixirProvider) GenerateShims(installPath string, version string) (map[string]string, error) {
	executables, err := p.ListExecutables(installPath, version)
	if err != nil {
		return nil, err
	}

	shims := make(map[string]string)
	for _, exe := range executables {
		name := filepath.Base(exe)
		shims[name] = exe
	}

	return shims, nil
}

func (p *ElixirProvider) DetectVersion(ctx context.Context, installPath string) (string, error) {
	return filepath.Base(installPath), nil
}

func (p *ElixirProvider) ListExecutables(installPath string, version string) ([]string, error) {
	var executables []string
	
	// Elixir binaries are usually in bin/ directory
	binPath := filepath.Join(installPath, "bin")
	if _, err := os.Stat(binPath); err == nil {
		err = filepath.Walk(binPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				executables = append(executables, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		// Fallback to walk entire installPath
		err := filepath.Walk(installPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && (info.Name() == "elixir" || info.Name() == "elixir.exe") {
				executables = append(executables, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return executables, nil
}

func (p *ElixirProvider) Uninstall(ctx context.Context, installPath string, version string) error {
	return nil
}
