// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/pkg/download"
	"github.com/snowdreamtech/unirtm/internal/provider"
	"github.com/snowdreamtech/unirtm/internal/repository"
)

type mockProviderForEnv struct {
	mockProvider
	MockGetEnvVars func(tool, installPath, version string) (map[string]string, error)
}

func (m *mockProviderForEnv) GetEnvVars(tool string, installPath string, version string) (map[string]string, error) {
	if m.MockGetEnvVars != nil {
		return m.MockGetEnvVars(tool, installPath, version)
	}
	return map[string]string{}, nil
}

func (m *mockProviderForEnv) GetBackendName() string {
	return m.name
}

func (m *mockProviderForEnv) AutoDetectBackend(toolName string) string {
	return m.name
}

// Ensure mockProviderForEnv implements provider.Provider
var _ provider.Provider = (*mockProviderForEnv)(nil)

func TestInstallationManager_Aliases(t *testing.T) {
	im := &InstallationManager{}

	aliases := map[string]map[string]string{
		"node": {
			"lts": "20.0.0",
		},
	}
	im.SetAliases(aliases)

	assert.Equal(t, "20.0.0", im.resolveAlias("node", "lts"))
	assert.Equal(t, "latest", im.resolveAlias("node", "latest"))
	assert.Equal(t, "lts", im.resolveAlias("go", "lts"))
}

func TestInstallationManager_ToolConfigs(t *testing.T) {
	im := &InstallationManager{}

	configs := map[string]config.ToolConfig{
		"node": {
			PostInstall: "npm install -g yarn",
		},
	}
	im.SetToolConfigs(configs)

	assert.Equal(t, configs, im.toolConfigs)
}

func TestInstallationManager_ExecuteHook(t *testing.T) {
	im := &InstallationManager{}

	// Test empty command
	err := im.executeHook(context.Background(), "", "node", "18.0.0")
	require.NoError(t, err)

	// Test a simple echo command that shouldn't fail
	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "out.txt")

	cmd := "echo success > " + outFile
	if os.PathSeparator == '\\' {
		cmd = "echo success > " + outFile
	}

	// Create context with quiet progress to test both branches
	ctx := context.WithValue(context.Background(), ContextKeyQuietProgress, true)

	err = im.executeHook(ctx, cmd, "node", "18.0.0")
	require.NoError(t, err)

	// Verify it actually ran
	content, err := os.ReadFile(outFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "success")

	// Test failing command
	err = im.executeHook(context.Background(), "exit 1", "node", "18.0.0")
	if os.PathSeparator != '\\' {
		require.Error(t, err)
	}
}

func TestInstallationManager_IsInstalled(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)

	installRepo := &mockInstallationRepo{
		installations: map[string]*repository.Installation{
			"node-18.0.0": {Tool: "node", Version: "18.0.0", Backend: "github"}, "github-node-18.0.0": {
				Tool:    "node",
				Version: "18.0.0",
				Backend: "github",
			},
		},
	}

	im := NewInstallationManager(
		backend.NewRegistry(),
		provider.NewRegistry(),
		download.NewManager(),
		installRepo,
		&mockTransactionManager{},
		&config.Settings{},
	)

	im.SetAliases(map[string]map[string]string{
		"node": {
			"lts": "18.0.0",
		},
	})

	// Create fake installation paths
	installPath18 := filepath.Join(tmpDir, "installs", "github-node", "18.0.0")
	os.MkdirAll(installPath18, 0755)

	// Test basic lookup
	t.Logf("Checking node 18.0.0")
	installed, inst := im.IsInstalled(context.Background(), "node", "18.0.0", "github")
	assert.True(t, installed)
	assert.NotNil(t, inst)

	// Test alias lookup
	installed, inst = im.IsInstalled(context.Background(), "node", "lts", "github")
	assert.True(t, installed)
	assert.NotNil(t, inst)

	// Test missing
	installed, inst = im.IsInstalled(context.Background(), "node", "19.0.0", "github")
	assert.False(t, installed)
	assert.Nil(t, inst)
}

func TestInstallationManager_Uninstall(t *testing.T) {
	installRepo := &mockInstallationRepo{
		installations: map[string]*repository.Installation{
			"node-18.0.0": {Tool: "node", Version: "18.0.0", Backend: "github"}, "github-node-18.0.0": {
				Tool:    "node",
				Version: "18.0.0",
				Backend: "github",
			},
		},
	}

	mockProv := &mockProviderForEnv{mockProvider: mockProvider{name: "node"}}
	providerRegistry := provider.NewRegistry()
	providerRegistry.Register("node", mockProv)

	txManager := &mockTransactionManager{
		tx: &mockTransaction{
			installationRepo: installRepo,
			auditRepo:        &mockAuditRepo{},
		},
	}

	im := NewInstallationManager(
		backend.NewRegistry(),
		providerRegistry,
		download.NewManager(),
		installRepo,
		txManager,
		&config.Settings{},
	)

	// First uninstall (success)
	err := im.Uninstall(context.Background(), "node", "18.0.0")
	require.NoError(t, err)

	// Second uninstall (not found error)
	err = im.Uninstall(context.Background(), "node", "18.0.0")
	require.ErrorIs(t, err, repository.ErrNotFound)
}

func TestInstallationManager_ParseToolSpec(t *testing.T) {
	im := &InstallationManager{}

	tests := []struct {
		spec         string
		wantBackend  string
		wantTool     string
		wantVersion  string
		wantExplicit bool
	}{
		{"node", "native", "node", "latest", false},
		{"node@18.0.0", "native", "node", "18.0.0", true},
		{"github:node@18.0.0", "github", "node", "18.0.0", true},
		{"go:golang/go", "go-pkg", "golang/go", "latest", false}, // Backend remapping
		{"@scope/package@1.0.0", "github", "@scope/package", "1.0.0", true},
		{"@scope/package", "github", "@scope/package", "latest", false},
	}

	for _, tt := range tests {
		b, tool, v, exp := im.ParseToolSpec(tt.spec)
		assert.Equal(t, tt.wantBackend, b, "backend mismatch for %s", tt.spec)
		assert.Equal(t, tt.wantTool, tool, "tool mismatch for %s", tt.spec)
		assert.Equal(t, tt.wantVersion, v, "version mismatch for %s", tt.spec)
		assert.Equal(t, tt.wantExplicit, exp, "explicit mismatch for %s", tt.spec)
	}
}

func TestInstallationManager_AutoDetectBackend(t *testing.T) {
	backendRegistry := backend.NewRegistry()
	b := &mockUpdateBackend{name: "github"}
	backendRegistry.Register(b)

	im := NewInstallationManager(
		backendRegistry,
		provider.NewRegistry(),
		download.NewManager(),
		&mockInstallationRepo{},
		&mockTransactionManager{},
		&config.Settings{},
	)

	im.SetToolConfigs(map[string]config.ToolConfig{
		"node": {
			Backend: "github",
		},
	})

	// Has config (wait, AutoDetectBackend DOES NOT use toolConfigs!)
	// Wait, AutoDetectBackend checks strings.Contains(toolName, "/") -> "github"
	assert.Equal(t, "github", im.AutoDetectBackend("foo/bar"))
	// Native tool
	assert.Equal(t, "native", im.AutoDetectBackend("node"))
	// Defaults to asdf
	assert.Equal(t, "asdf", im.AutoDetectBackend("unknown"))
}

func TestInstallationManager_RemoveEmptyDirs(t *testing.T) {
	im := &InstallationManager{}
	tmpDir := t.TempDir()

	deepDir := filepath.Join(tmpDir, "a", "b", "c")
	os.MkdirAll(deepDir, 0755)

	// Create a file in a/b to prevent it from being deleted
	os.WriteFile(filepath.Join(tmpDir, "a", "b", "file.txt"), []byte("test"), 0644)

	im.removeEmptyDirs(filepath.Join(deepDir, "dummy"), tmpDir)

	// c should be removed, a/b should remain
	// In Go 1.20+, RemoveAll returns nil on non-existent. But we use Remove, which errors.
	// Actually removeEmptyDirs swallows errors. We should check if deepDir is removed.
	_, err := os.Stat(deepDir)
	assert.True(t, os.IsNotExist(err))

	_, err = os.Stat(filepath.Join(tmpDir, "a", "b"))
	assert.NoError(t, err)
}

func TestInstallationManager_IsExecutableFile(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "exec")
	os.WriteFile(file1, []byte("test"), 0755)

	file2 := filepath.Join(tmpDir, "nonexec")
	os.WriteFile(file2, []byte("test"), 0644)

	dir1 := filepath.Join(tmpDir, "dir")
	os.MkdirAll(dir1, 0755)

	assert.True(t, isExecutableFile(file1))
	if runtime.GOOS != "windows" {
		assert.False(t, isExecutableFile(file2))
	}
	assert.False(t, isExecutableFile(dir1))
	assert.False(t, isExecutableFile(filepath.Join(tmpDir, "missing")))
}

func TestInstallationManager_ResolveToolEnvBySpec(t *testing.T) {
	im := &InstallationManager{
		providerRegistry: provider.NewRegistry(),
	}

	// Create mock provider
	mp := &mockProviderForEnv{
		mockProvider: mockProvider{name: "mock"},
		MockGetEnvVars: func(tool, installPath, version string) (map[string]string, error) {
			if tool == "node" {
				return map[string]string{"NODE_ENV": "production"}, nil
			}
			return nil, fmt.Errorf("no env")
		},
	}
	im.providerRegistry.Register("mock", mp)

	// Mock environment
	tmpDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")

	// 1. Not installed
	envMap := im.ResolveToolEnvBySpec("node", "18.0.0", "mock")
	assert.Empty(t, envMap)

	// 2. Installed, with bin dir and provider env
	installDir := filepath.Join(tmpDir, "installs", "mock-node", "18.0.0")
	binDir := filepath.Join(installDir, "bin")
	os.MkdirAll(binDir, 0755)

	envMap2 := im.ResolveToolEnvBySpec("node", "18.0.0", "mock")
	assert.Equal(t, "production", envMap2["NODE_ENV"])
	assert.Equal(t, "18.0.0", envMap2["UNIRTM_NODE_VERSION"])
	assert.Equal(t, binDir, envMap2["PATH"])

	// 3. Installed, no bin dir but install dir is used as PATH, provider fails to get env
	installDir3 := filepath.Join(tmpDir, "installs", "mock-python", "3.9.0")
	os.MkdirAll(installDir3, 0755)

	envMap3 := im.ResolveToolEnvBySpec("python", "3.9.0", "mock")
	assert.Empty(t, envMap3["NODE_ENV"]) // Provider failed
	assert.Equal(t, "3.9.0", envMap3["UNIRTM_PYTHON_VERSION"])
	assert.Equal(t, installDir3, envMap3["PATH"])

	// 4. Test PATH prepend
	mp2 := &mockProviderForEnv{
		mockProvider: mockProvider{name: "mock2"},
		MockGetEnvVars: func(tool, installPath, version string) (map[string]string, error) {
			return map[string]string{"PATH": "/existing/path"}, nil
		},
	}
	im.providerRegistry.Register("mock2", mp2)
	installDir4 := filepath.Join(tmpDir, "installs", "mock2-ruby", "3.0.0")
	os.MkdirAll(filepath.Join(installDir4, "bin"), 0755)

	envMap4 := im.ResolveToolEnvBySpec("ruby", "3.0.0", "mock2")
	expectedPath := filepath.Join(installDir4, "bin") + string(os.PathListSeparator) + "/existing/path"
	assert.Equal(t, expectedPath, envMap4["PATH"])
}

type mockProviderForExecutables struct {
	mockProviderForEnv
	MockListExecutables func(tool, installPath, version string) ([]string, error)
}

func (m *mockProviderForExecutables) ListExecutables(tool string, installPath string, version string) ([]string, error) {
	if m.MockListExecutables != nil {
		return m.MockListExecutables(tool, installPath, version)
	}
	return nil, nil
}

// Ensure mockProviderForExecutables implements provider.Provider
var _ provider.Provider = (*mockProviderForExecutables)(nil)

func TestInstallationManager_ResolveExecutable(t *testing.T) {
	tmpDir := t.TempDir()

	installRepo := &mockInstallationRepo{
		installations: map[string]*repository.Installation{
			"node-18.0.0": {Tool: "node", Version: "18.0.0", Backend: "github", InstallPath: filepath.Join(tmpDir, "node")},
			"python-3.9":  {Tool: "python", Version: "3.9", Backend: "github", InstallPath: filepath.Join(tmpDir, "python")},
		},
	}

	im := &InstallationManager{
		installRepo:      installRepo,
		providerRegistry: provider.NewRegistry(),
		toolConfigs:      make(map[string]config.ToolConfig),
		aliases:          make(map[string]map[string]string),
	}

	// Create executable files
	nodeDir := filepath.Join(tmpDir, "node", "bin")
	os.MkdirAll(nodeDir, 0755)
	nodeExec := filepath.Join(nodeDir, "node")
	os.WriteFile(nodeExec, []byte("exec"), 0755)

	pyDir := filepath.Join(tmpDir, "python", "bin")
	os.MkdirAll(pyDir, 0755)
	pyExec := filepath.Join(pyDir, "python-3.9")
	os.WriteFile(pyExec, []byte("exec"), 0755)

	// Mock provider
	mp := &mockProviderForExecutables{
		mockProviderForEnv: mockProviderForEnv{
			mockProvider: mockProvider{name: "github"},
			MockGetEnvVars: func(tool, installPath, version string) (map[string]string, error) {
				return map[string]string{"ENV_VAR": tool}, nil
			},
		},
		MockListExecutables: func(tool, installPath, version string) ([]string, error) {
			if tool == "node" {
				return []string{nodeExec}, nil
			} else if tool == "python" {
				return []string{pyExec}, nil
			}
			return nil, nil
		},
	}
	im.providerRegistry.Register("github", mp)

	// 1. Exact match
	exePath, envVars, err := im.ResolveExecutable(context.Background(), "node", backend.Platform{})
	require.NoError(t, err)
	assert.Equal(t, nodeExec, exePath)
	assert.Equal(t, "node", envVars["ENV_VAR"])

	// 2. Prefix match
	exePath, envVars, err = im.ResolveExecutable(context.Background(), "python", backend.Platform{})
	require.NoError(t, err)
	assert.Equal(t, pyExec, exePath)
	assert.Equal(t, "python", envVars["ENV_VAR"])

	// 3. Not found
	_, _, err = im.ResolveExecutable(context.Background(), "ruby", backend.Platform{})
	require.Error(t, err)

	// 4. Test toolConfigs filtering
	im.toolConfigs["node"] = config.ToolConfig{Version: "20.0.0"}
	_, _, err = im.ResolveExecutable(context.Background(), "node", backend.Platform{})
	require.Error(t, err) // Version 18.0.0 should be filtered out
}

func TestInstallationManager_EnsureInstalled(t *testing.T) {
	tempShims := t.TempDir()
	tempInstalls := t.TempDir()

	backendReg := backend.NewRegistry()
	backendReg.Register(&mockUpdateBackend{
		name: "native",
		versions: map[string]*backend.VersionInfo{
			"18.0.0": {Version: "18.0.0", DownloadURL: "https://example.com/node"},
			"3.9":    {Version: "3.9", DownloadURL: "https://example.com/dummy-tool"},
		},
	})
	dm1 := download.NewManager()
	dm1.Register("https", &mockDownloaderForInstall{})

	im := &InstallationManager{
		providerRegistry: provider.NewRegistry(),
		backendRegistry:  backendReg,
		downloadManager:  dm1,
		txManager:        &mockTransactionManager{tx: &mockTransaction{installationRepo: &mockInstallationRepo{}, auditRepo: &mockAuditRepo{}}},
		installRepo: &mockInstallationRepo{
			installations: make(map[string]*repository.Installation),
		},
		settings:      &config.Settings{},
		shimGenerator: NewGenerator(tempShims, tempInstalls),
	}

	im.SetToolConfigs(map[string]config.ToolConfig{
		"node": {Version: "18.0.0"},
	})

	// Ensure EnsureInstalled routes to EnsureInstalledFromSpecs and succeeds (or errors)
	err := im.EnsureInstalled(context.Background(), map[string]config.ToolConfig{
		"node": {Version: "18.0.0"},
	})
	// It should succeed since all mocks are properly set up
	require.NoError(t, err)
}

func TestInstallationManager_EnsureInstalledFromSpecs(t *testing.T) {
	installRepo := &mockInstallationRepo{
		installations: map[string]*repository.Installation{
			"node-18.0.0": {Tool: "node", Version: "18.0.0", Backend: "native"},
		},
	}

	backendReg := backend.NewRegistry()
	backendReg.Register(&mockUpdateBackend{
		name: "native",
		versions: map[string]*backend.VersionInfo{
			"18.0.0": {Version: "18.0.0", DownloadURL: "https://example.com/node"},
			"3.9":    {Version: "3.9", DownloadURL: "https://example.com/dummy-tool"},
		},
	})
	dm2 := download.NewManager()
	dm2.Register("https", &mockDownloaderForInstall{})

	tempShims := t.TempDir()
	tempInstalls := t.TempDir()

	im := &InstallationManager{
		providerRegistry: provider.NewRegistry(),
		backendRegistry:  backendReg,
		downloadManager:  dm2,
		txManager:        &mockTransactionManager{tx: &mockTransaction{installationRepo: installRepo, auditRepo: &mockAuditRepo{}}},
		installRepo:      installRepo,
		settings:         &config.Settings{},
		shimGenerator:    NewGenerator(tempShims, tempInstalls),
	}

	// Installed tool
	err := im.EnsureInstalledFromSpecs(context.Background(), map[string]ToolSpec{
		"node": {Name: "node", Version: "18.0.0", BackendName: "native", OriginalName: "node"},
	})
	require.NoError(t, err)

	// Not installed tool
	err = im.EnsureInstalledFromSpecs(context.Background(), map[string]ToolSpec{
		"dummy-tool": {Name: "dummy-tool", Version: "3.9", BackendName: "native", OriginalName: "dummy-tool"},
	})
	require.NoError(t, err) // Should succeed with mocks
}

// mockDownloaderForInstall implements download.Downloader
type mockDownloaderForInstall struct{}

func (m *mockDownloaderForInstall) Download(ctx context.Context, url, dest string, opts download.DownloadOptions) error {
	// Create a dummy file
	return os.WriteFile(dest, []byte("dummy content"), 0644)
}
func (m *mockDownloaderForInstall) VerifyChecksum(ctx context.Context, path, checksum string) error {
	return nil
}
