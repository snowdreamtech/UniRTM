// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/provider"
	"github.com/snowdreamtech/unirtm/internal/repository"
	"github.com/stretchr/testify/require"
)

type mockInstallRepo struct {
	repository.InstallationRepository
	installations []*repository.Installation
	err           error
}

func (m *mockInstallRepo) List(ctx context.Context) ([]*repository.Installation, error) {
	return m.installations, m.err
}

type mockResolveProvider struct {
	provider.Provider
	executables []string
	envVars     map[string]string
}

func (m *mockResolveProvider) ListExecutables(tool string, installPath string, version string) ([]string, error) {
	return m.executables, nil
}

func (m *mockResolveProvider) GetEnvVars(tool string, installPath string, version string) (map[string]string, error) {
	return m.envVars, nil
}

func TestResolveExecutable_Success(t *testing.T) {
	br := backend.NewRegistry()
	pr := provider.NewRegistry()

	tempDir := t.TempDir()
	exePath := filepath.Join(tempDir, "testbin")
	// create dummy file and make it executable
	f, err := os.Create(exePath)
	require.NoError(t, err)
	f.Close()
	os.Chmod(exePath, 0755)

	pr.Register("mock", &mockResolveProvider{
		executables: []string{exePath},
		envVars:     map[string]string{"TEST_ENV": "1"},
	})

	repo := &mockInstallRepo{
		installations: []*repository.Installation{
			{
				Tool:        "testtool",
				Version:     "1.0.0",
				Backend:     "mock",
				InstallPath: tempDir,
			},
		},
	}

	im := NewInstallationManager(br, pr, nil, repo, nil, nil)

	ctx := context.Background()

	resolvedPath, envVars, err := im.ResolveExecutable(ctx, "testbin", backend.Platform{})
	require.NoError(t, err)
	require.Equal(t, exePath, resolvedPath)
	require.Equal(t, "1", envVars["TEST_ENV"])
}

func TestResolveExecutable_NotFound(t *testing.T) {
	br := backend.NewRegistry()
	pr := provider.NewRegistry()

	pr.Register("mock", &mockResolveProvider{
		executables: []string{"nonexistent"},
	})

	repo := &mockInstallRepo{
		installations: []*repository.Installation{
			{
				Tool:        "testtool",
				Version:     "1.0.0",
				Backend:     "mock",
				InstallPath: "/tmp",
			},
		},
	}

	im := NewInstallationManager(br, pr, nil, repo, nil, nil)

	ctx := context.Background()
	_, _, err := im.ResolveExecutable(ctx, "testbin", backend.Platform{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "executable testbin not found")
}

func TestResolveExecutable_PrefixMatch(t *testing.T) {
	br := backend.NewRegistry()
	pr := provider.NewRegistry()

	tempDir := t.TempDir()
	exePath := filepath.Join(tempDir, "testbin-1.0")
	// create dummy file and make it executable
	f, err := os.Create(exePath)
	require.NoError(t, err)
	f.Close()
	os.Chmod(exePath, 0755)

	pr.Register("mock", &mockResolveProvider{
		executables: []string{exePath},
	})

	repo := &mockInstallRepo{
		installations: []*repository.Installation{
			{
				Tool:        "testtool",
				Version:     "1.0.0",
				Backend:     "mock",
				InstallPath: tempDir,
			},
		},
	}

	im := NewInstallationManager(br, pr, nil, repo, nil, nil)

	ctx := context.Background()
	resolvedPath, _, err := im.ResolveExecutable(ctx, "testbin", backend.Platform{})
	require.NoError(t, err)
	require.Equal(t, exePath, resolvedPath)
}
