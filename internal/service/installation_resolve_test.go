// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
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
	binPaths    []string
	envVars     map[string]string
}

func (m *mockResolveProvider) ListExecutables(tool string, installPath string, version string) ([]string, error) {
	return m.executables, nil
}

func (m *mockResolveProvider) GetBinPaths(tool string, installPath string, version string) ([]string, error) {
	if len(m.binPaths) > 0 {
		return m.binPaths, nil
	}
	return []string{installPath}, nil
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

	exeName := "testbin-1.0"
	if runtime.GOOS == "windows" {
		exeName += ".exe"
	}
	exePath := filepath.Join(tempDir, exeName)

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

func TestResolveExecutable_WindowsExtensionMatching(t *testing.T) {
	br := backend.NewRegistry()
	pr := provider.NewRegistry()

	tempDir := t.TempDir()
	exePath := filepath.Join(tempDir, "node.exe")

	f, err := os.Create(exePath)
	require.NoError(t, err)
	f.Close()
	os.Chmod(exePath, 0755)

	pr.Register("mock", &mockResolveProvider{
		executables: []string{exePath},
		envVars:     map[string]string{"NODE_ENV": "production"},
	})

	repo := &mockInstallRepo{
		installations: []*repository.Installation{
			{
				Tool:        "node",
				Version:     "20.0.0",
				Backend:     "mock",
				InstallPath: tempDir,
			},
		},
	}

	im := NewInstallationManager(br, pr, nil, repo, nil, nil)
	ctx := context.Background()

	// 1. Resolve with "node" when binary is "node.exe" on Windows platform
	resolved, _, err := im.ResolveExecutable(ctx, "node", backend.Platform{OS: "windows"})
	require.NoError(t, err)
	require.Equal(t, exePath, resolved)

	// 2. Resolve with "node.exe" when binary is "node.exe" on Windows platform
	resolved, _, err = im.ResolveExecutable(ctx, "node.exe", backend.Platform{OS: "windows"})
	require.NoError(t, err)
	require.Equal(t, exePath, resolved)

	// 3. Resolve with "NODE.EXE" (case-insensitive) on Windows platform
	resolved, _, err = im.ResolveExecutable(ctx, "NODE.EXE", backend.Platform{OS: "windows"})
	require.NoError(t, err)
	require.Equal(t, exePath, resolved)
}

func TestResolveExecutable_PythonAliasMatching(t *testing.T) {
	br := backend.NewRegistry()
	pr := provider.NewRegistry()

	tempDir := t.TempDir()
	exePath := filepath.Join(tempDir, "python.exe")

	f, err := os.Create(exePath)
	require.NoError(t, err)
	f.Close()
	os.Chmod(exePath, 0755)

	pr.Register("mock", &mockResolveProvider{
		executables: []string{exePath},
		envVars:     map[string]string{"PYTHONUTF8": "1"},
	})

	repo := &mockInstallRepo{
		installations: []*repository.Installation{
			{
				Tool:        "python",
				Version:     "3.14.7",
				Backend:     "mock",
				InstallPath: tempDir,
			},
		},
	}

	im := NewInstallationManager(br, pr, nil, repo, nil, nil)
	ctx := context.Background()

	// 1. On Windows: resolving "python3" when only "python.exe" is installed should succeed
	resolved, _, err := im.ResolveExecutable(ctx, "python3", backend.Platform{OS: "windows"})
	require.NoError(t, err)
	require.Equal(t, exePath, resolved)

	// 2. On Windows: resolving "python3.EXE" should succeed
	resolved, _, err = im.ResolveExecutable(ctx, "python3.EXE", backend.Platform{OS: "windows"})
	require.NoError(t, err)
	require.Equal(t, exePath, resolved)

	// 3. Resolving "python" directly should succeed
	resolved, _, err = im.ResolveExecutable(ctx, "python", backend.Platform{OS: "windows"})
	require.NoError(t, err)
	require.Equal(t, exePath, resolved)
}

func TestResolveExecutable_GeneralAliases(t *testing.T) {
	br := backend.NewRegistry()
	pr := provider.NewRegistry()

	tempDir := t.TempDir()
	nodePath := filepath.Join(tempDir, "node")
	makePath := filepath.Join(tempDir, "make")
	f1, _ := os.Create(nodePath)
	f1.Close()
	os.Chmod(nodePath, 0755)
	f2, _ := os.Create(makePath)
	f2.Close()
	os.Chmod(makePath, 0755)

	pr.Register("mock-node", &mockResolveProvider{
		executables: []string{nodePath},
	})
	pr.Register("mock-make", &mockResolveProvider{
		executables: []string{makePath},
	})

	repo := &mockInstallRepo{
		installations: []*repository.Installation{
			{Tool: "node", Version: "20.0.0", Backend: "mock-node", InstallPath: tempDir},
			{Tool: "make", Version: "4.4.1", Backend: "mock-make", InstallPath: tempDir},
		},
	}

	im := NewInstallationManager(br, pr, nil, repo, nil, nil)
	ctx := context.Background()

	// 1. Resolve "nodejs" when installed binary is "node"
	resolved, _, err := im.ResolveExecutable(ctx, "nodejs", backend.Platform{OS: "linux"})
	require.NoError(t, err)
	require.Equal(t, nodePath, resolved)

	// 2. Resolve "gmake" when installed binary is "make"
	resolved, _, err = im.ResolveExecutable(ctx, "gmake", backend.Platform{OS: "linux"})
	require.NoError(t, err)
	require.Equal(t, makePath, resolved)

	// 3. Resolve "mingw32-make" when installed binary is "make"
	resolved, _, err = im.ResolveExecutable(ctx, "mingw32-make", backend.Platform{OS: "windows"})
	require.NoError(t, err)
	require.Equal(t, makePath, resolved)
}

func TestResolveExecutable_BinPathsFallback(t *testing.T) {
	br := backend.NewRegistry()
	pr := provider.NewRegistry()

	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	require.NoError(t, os.MkdirAll(binDir, 0755))

	realJavaPath := filepath.Join(binDir, "java")
	f, err := os.Create(realJavaPath)
	require.NoError(t, err)
	f.Close()
	os.Chmod(realJavaPath, 0755)

	// Provider returns relative name "java" without subfolder, but specifies binPaths
	pr.Register("mock-java", &mockResolveProvider{
		executables: []string{"java"},
		binPaths:    []string{binDir},
	})

	repo := &mockInstallRepo{
		installations: []*repository.Installation{
			{Tool: "java", Version: "21.0.0", Backend: "mock-java", InstallPath: tempDir},
		},
	}

	im := NewInstallationManager(br, pr, nil, repo, nil, nil)
	ctx := context.Background()

	resolved, _, err := im.ResolveExecutable(ctx, "java", backend.Platform{OS: "linux"})
	require.NoError(t, err)
	require.Equal(t, realJavaPath, resolved)
}

func TestResolveExecutable_PrefixMatchWithDigits(t *testing.T) {
	br := backend.NewRegistry()
	pr := provider.NewRegistry()

	tempDir := t.TempDir()
	lua54Path := filepath.Join(tempDir, "lua54")
	f, err := os.Create(lua54Path)
	require.NoError(t, err)
	f.Close()
	os.Chmod(lua54Path, 0755)

	pr.Register("mock-lua", &mockResolveProvider{
		executables: []string{lua54Path},
	})

	repo := &mockInstallRepo{
		installations: []*repository.Installation{
			{Tool: "lua", Version: "5.4.6", Backend: "mock-lua", InstallPath: tempDir},
		},
	}

	im := NewInstallationManager(br, pr, nil, repo, nil, nil)
	ctx := context.Background()

	// Query "lua" matches "lua54" because '5' is a digit
	resolved, _, err := im.ResolveExecutable(ctx, "lua", backend.Platform{OS: "linux"})
	require.NoError(t, err)
	require.Equal(t, lua54Path, resolved)
}
