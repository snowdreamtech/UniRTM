// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"os"
	"path/filepath"

	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
)

// FlutterProvider implements the Provider interface for Flutter.
type FlutterProvider struct {
}

// NewFlutterProvider creates a new Flutter provider.
func NewFlutterProvider() *FlutterProvider {
	return &FlutterProvider{}
}

func (p *FlutterProvider) Name() string {
	return "flutter"
}

func (p *FlutterProvider) Install(ctx context.Context, tool string, installPath string, artifactPath string, version string) error {
	logger.Debug("Installing Flutter", map[string]interface{}{"version": version, "installPath": installPath, "artifactPath": artifactPath})

	if err := os.MkdirAll(installPath, 0755); err != nil {
		return err
	}

	return nil
}

func (p *FlutterProvider) PostInstall(ctx context.Context, tool string, installPath string, version string) error {
	return nil
}

func (p *FlutterProvider) GenerateShims(tool string, installPath string, version string) (map[string]string, error) {
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

func (p *FlutterProvider) DetectVersion(ctx context.Context, tool string, installPath string) (string, error) {
	return filepath.Base(installPath), nil
}

func (p *FlutterProvider) ListExecutables(tool string, installPath string, version string) ([]string, error) {
	var executables []string
	
	// Flutter binary is usually in bin/
	// Dart is in bin/cache/dart-sdk/bin/
	
	binPath := filepath.Join(installPath, "bin")
	if _, err := os.Stat(binPath); err == nil {
		err = filepath.Walk(binPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && (info.Name() == "flutter" || info.Name() == "flutter.bat") {
				executables = append(executables, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	dartPath := filepath.Join(installPath, "bin", "cache", "dart-sdk", "bin")
	if _, err := os.Stat(dartPath); err == nil {
		err = filepath.Walk(dartPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && (info.Name() == "dart" || info.Name() == "dart.exe") {
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

// GetBinPaths returns the absolute paths to the bin directories.
func (p *FlutterProvider) GetBinPaths(tool string, installPath string, version string) ([]string, error) {
	return []string{
		filepath.Join(installPath, "bin"),
		filepath.Join(installPath, "bin", "cache", "dart-sdk", "bin"),
	}, nil
}

// GetEnvVars returns no special environment variables.
func (p *FlutterProvider) GetEnvVars(tool string, installPath string, version string) (map[string]string, error) {
	return make(map[string]string), nil
}

func (p *FlutterProvider) Uninstall(ctx context.Context, tool string, installPath string, version string) error {
	return nil
}
