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

func (p *SpmProvider) Install(ctx context.Context, installPath string, artifactPath string, version string) error {
	// Usually tool is a URL for SPM
	// Extract the full tool name (including scope if present) from the install path.
	installsDir := env.GetInstallsDir()
	toolDir := filepath.Dir(installPath)
	tool, err := filepath.Rel(installsDir, toolDir)
	if err != nil {
		tool = filepath.Base(toolDir) // fallback
	}

	if err := os.MkdirAll(installPath, 0755); err != nil {
		return err
	}

	swiftCmd, err := exec.LookPath("swift")
	if err != nil {
		return NewProviderError(p.Name(), tool, version, "swift is required to install spm packages but was not found in PATH", err)
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

func (p *SpmProvider) PostInstall(ctx context.Context, installPath string, version string) error {
	return nil
}

func (p *SpmProvider) GenerateShims(installPath string, version string) (map[string]string, error) {
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

func (p *SpmProvider) DetectVersion(ctx context.Context, installPath string) (string, error) {
	return filepath.Base(installPath), nil
}

func (p *SpmProvider) ListExecutables(installPath string, version string) ([]string, error) {
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

func (p *SpmProvider) Uninstall(ctx context.Context, installPath string, version string) error {
	return nil
}
