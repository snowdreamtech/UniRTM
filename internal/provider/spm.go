// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
	"github.com/snowdreamtech/unirtm/internal/pkg/version"
)

// SpmProvider implements the Provider interface for Swift Package Manager.
type SpmProvider struct {
}

// NewSpmProvider creates a new spm provider.
func NewSpmProvider() *SpmProvider {
	return &SpmProvider{}
}

func (p *SpmProvider) Name() string {
	return "spm"
}

func (p *SpmProvider) Install(ctx context.Context, tool string, installPath string, artifactPath string, version string) error {
	// Usually tool is a URL for SPM
	if err := os.MkdirAll(installPath, 0755); err != nil {
	}

	swiftCmd, err := p.findSwift()
	if err != nil {
		return NewProviderError(p.Name(), tool, version, "swift is required to install spm packages but was not found", err)
	}
	gitCmd, err := exec.LookPath("git")
	if err != nil {
		return NewProviderError(p.Name(), tool, version, "git is required to install spm packages but was not found in PATH", err)
	}

	logger.Debug("Installing via SPM", map[string]interface{}{"tool": tool, "version": version, "installDir": installPath})

	// We clone the repository into a temp dir, checkout version, and run swift build -c release
	tmpDir, err := os.MkdirTemp("", "unirtm-spm-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	repoURL := fmt.Sprintf("https://github.com/%s.git", tool) // Simplified assumption for SPM

	cloneArgs := []string{"clone", repoURL, tmpDir}
	if version != "latest" && version != "" {
		cloneArgs = append(cloneArgs, "--branch", version, "--depth", "1")
	}

	gitClone := exec.CommandContext(ctx, gitCmd, cloneArgs...)
	gitClone.Stdout = os.Stdout
	gitClone.Stderr = os.Stderr
	if err := gitClone.Run(); err != nil {
		return NewProviderError(p.Name(), tool, version, "git clone failed", err)
	}

	buildCmd := exec.CommandContext(ctx, swiftCmd, "build", "-c", "release")
	buildCmd.Dir = tmpDir
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		return NewProviderError(p.Name(), tool, version, "swift build failed", err)
	}

	// Copy binary from .build/release/ to installPath/bin
	binDir := filepath.Join(installPath, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}

	releaseDir := filepath.Join(tmpDir, ".build", "release")
	entries, err := os.ReadDir(releaseDir)
	if err != nil {
		return NewProviderError(p.Name(), tool, version, "read release dir", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			info, err := entry.Info()
			if err == nil && info.Mode()&0111 != 0 && filepath.Ext(entry.Name()) == "" {
				src := filepath.Join(releaseDir, entry.Name())
				dst := filepath.Join(binDir, entry.Name())
				if err := p.copyFile(src, dst); err != nil {
					return NewProviderError(p.Name(), tool, version, "copy binary", err)
				}
				os.Chmod(dst, 0755)
			}
		}
	}

	return nil
}

func (p *SpmProvider) copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func (p *SpmProvider) PostInstall(ctx context.Context, tool string, installPath string, version string) error {
	return nil
}

func (p *SpmProvider) GenerateShims(tool string, installPath string, version string) (map[string]string, error) {
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

func (p *SpmProvider) DetectVersion(ctx context.Context, tool string, installPath string) (string, error) {
	return filepath.Base(installPath), nil
}

func (p *SpmProvider) ListExecutables(tool string, installPath string, version string) ([]string, error) {
	binDir := filepath.Join(installPath, "bin")

	entries, err := os.ReadDir(binDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var executables []string
	for _, entry := range entries {
		if !entry.IsDir() {
			info, err := entry.Info()
			if err == nil {
				if info.Mode()&0111 != 0 {
					executables = append(executables, filepath.Join(binDir, entry.Name()))
				}
			}
		}
	}

	return executables, nil
}

// GetBinPaths returns the absolute path to the bin directory.
func (p *SpmProvider) GetBinPaths(tool string, installPath string, version string) ([]string, error) {
	return []string{filepath.Join(installPath, "bin")}, nil
}

// GetEnvVars returns no special environment variables.
func (p *SpmProvider) GetEnvVars(tool string, installPath string, version string) (map[string]string, error) {
	return make(map[string]string), nil
}

func (p *SpmProvider) Uninstall(ctx context.Context, tool string, installPath string, version string) error {
	return nil
}

func (p *SpmProvider) findSwift() (string, error) {
	// 1. Try to find a UniRTM-managed Swift installation first
	swiftInstallsDir := filepath.Join(env.GetInstallsDir(), "swift")
	if entries, err := os.ReadDir(swiftInstallsDir); err == nil {
		var bestVer string
		var bestPath string
		for _, entry := range entries {
			if entry.IsDir() {
				verDir := filepath.Join(swiftInstallsDir, entry.Name())
				candidates := []string{
					filepath.Join(verDir, "bin", "swift"),
					filepath.Join(verDir, "swift"),
					filepath.Join(verDir, "swift.exe"),
				}
				for _, cand := range candidates {
					if info, err := os.Stat(cand); err == nil && !info.IsDir() {
						if bestVer == "" || version.CompareVersions(entry.Name(), bestVer) > 0 {
							bestVer = entry.Name()
							bestPath = cand
						}
						break
					}
				}
			}
		}
		if bestPath != "" {
			return bestPath, nil
		}
	}

	// 2. Fallback to system PATH
	return exec.LookPath("swift")
}
