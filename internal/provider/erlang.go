// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"os"
	"path/filepath"

	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
)

// ErlangProvider implements the Provider interface for Erlang.
type ErlangProvider struct {
}

// NewErlangProvider creates a new Erlang provider.
func NewErlangProvider() *ErlangProvider {
	return &ErlangProvider{}
}

func (p *ErlangProvider) Name() string {
	return "erlang"
}

func (p *ErlangProvider) Install(ctx context.Context, tool string, installPath string, artifactPath string, version string) error {
	logger.Debug("Installing Erlang", map[string]interface{}{"version": version, "installPath": installPath, "artifactPath": artifactPath})

	if err := os.MkdirAll(installPath, 0755); err != nil {
		return err
	}

	return nil
}

func (p *ErlangProvider) PostInstall(ctx context.Context, tool string, installPath string, version string) error {
	return nil
}

func (p *ErlangProvider) GenerateShims(tool string, installPath string, version string) (map[string]string, error) {
	executables, err := p.ListExecutables(tool, installPath, version)
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

func (p *ErlangProvider) DetectVersion(ctx context.Context, tool string, installPath string) (string, error) {
	return filepath.Base(installPath), nil
}

func (p *ErlangProvider) ListExecutables(tool string, installPath string, version string) ([]string, error) {
	var executables []string
	
	// Erlang binaries are usually in bin/ directory
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
			if !info.IsDir() && (info.Name() == "erl" || info.Name() == "erl.exe") {
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

// GetBinPaths returns the absolute path to the bin directory.
func (p *ErlangProvider) GetBinPaths(tool string, installPath string, version string) ([]string, error) {
	return []string{filepath.Join(installPath, "bin")}, nil
}

// GetEnvVars returns no special environment variables.
func (p *ErlangProvider) GetEnvVars(tool string, installPath string, version string) (map[string]string, error) {
	return make(map[string]string), nil
}

func (p *ErlangProvider) Uninstall(ctx context.Context, tool string, installPath string, version string) error {
	return nil
}
