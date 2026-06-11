// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/pkg/download"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
)

// MavenPkgProvider implements the Provider interface for Maven/Gradle packages.
type MavenPkgProvider struct {
}

// NewMavenPkgProvider creates a new Maven provider.
func NewMavenPkgProvider() *MavenPkgProvider {
	return &MavenPkgProvider{}
}

func (p *MavenPkgProvider) Name() string {
	return "maven-pkg"
}

func (p *MavenPkgProvider) Install(ctx context.Context, tool string, installPath string, artifactPath string, version string) error {
	// Ensure install path exists
	binDir := filepath.Join(installPath, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}

	// parse group and artifact
	parts := strings.Split(tool, "/")
	if len(parts) != 2 {
		return NewProviderError(p.Name(), tool, version, "tool name must be in the format groupId/artifactId (e.g., org.openapitools/openapi-generator-cli)", nil)
	}

	groupId := parts[0]
	artifactId := parts[1]

	groupPath := strings.ReplaceAll(groupId, ".", "/")

	// Determine mirror
	mirror := env.Get("UNIRTM_MAVEN_MIRROR")
	if mirror == "" {
		mirror = env.Get("UNIRTM_GRADLE_MIRROR")
	}
	if mirror == "" {
		mirror = "https://repo1.maven.org/maven2/"
	}
	if !strings.HasSuffix(mirror, "/") {
		mirror += "/"
	}

	jarName := fmt.Sprintf("%s-%s.jar", artifactId, version)
	url := fmt.Sprintf("%s%s/%s/%s/%s", mirror, groupPath, artifactId, version, jarName)

	targetJarPath := filepath.Join(installPath, jarName)

	logger.Debug("Downloading Maven artifact", map[string]interface{}{"url": url, "target": targetJarPath})

	dl, err := download.Get("https")
	if err != nil {
		return NewProviderError(p.Name(), tool, version, "failed to get downloader", err)
	}

	// Download the jar
	if err := dl.Download(ctx, url, targetJarPath, download.DefaultDownloadOptions()); err != nil {
		return NewProviderError(p.Name(), tool, version, "failed to download jar artifact", err)
	}

	// Generate shim wrapper logic
	// In UniRTM, GenerateShims returns the target to point to.
	// But since we need to run `java -jar ...`, we will create actual executable scripts in `bin/` here,
	// and then let GenerateShims return them so the core logic can link them if needed.

	isWindows := os.PathSeparator == '\\'
	shimContent := ""
	shimExt := ""

	if isWindows {
		shimContent = fmt.Sprintf(`@echo off
java -jar "%s" %%*
`, targetJarPath)
		shimExt = ".cmd"
	} else {
		shimContent = fmt.Sprintf(`#!/bin/sh
exec java -jar "%s" "$@"
`, targetJarPath)
	}

	shimPath := filepath.Join(binDir, artifactId+shimExt)
	if err := os.WriteFile(shimPath, []byte(shimContent), 0755); err != nil {
		return NewProviderError(p.Name(), tool, version, "failed to write shim", err)
	}

	return nil
}

// PostInstall performs post-installation steps.
func (p *MavenPkgProvider) PostInstall(ctx context.Context, tool string, installPath string, version string) error {
	p.checkAndWarnJava()
	return nil
}

// checkAndWarnJava warns the user if Java is not found on the system.
func (p *MavenPkgProvider) checkAndWarnJava() {
	if _, err := exec.LookPath("java"); err != nil {
		logger.Warn("⚠️ WARNING: 'java' was not found in PATH. This tool requires a Java runtime to execute. Please ensure Java is installed (e.g., 'unirtm use java').")
	}
}

func (p *MavenPkgProvider) GenerateShims(tool string, installPath string, version string) (map[string]string, error) {
	executables, err := p.ListExecutables(tool, installPath, version)
	if err != nil {
		return nil, err
	}

	shims := make(map[string]string)
	for _, exe := range executables {
		// Just map the executable name to its full path
		name := filepath.Base(exe)
		if strings.HasSuffix(strings.ToLower(name), ".cmd") {
			name = name[:len(name)-4]
		}
		shims[name] = exe
	}

	return shims, nil
}

func (p *MavenPkgProvider) DetectVersion(ctx context.Context, tool string, installPath string) (string, error) {
	return filepath.Base(installPath), nil
}

func (p *MavenPkgProvider) ListExecutables(tool string, installPath string, version string) ([]string, error) {
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
			executables = append(executables, filepath.Join(binDir, entry.Name()))
		}
	}

	return executables, nil
}

// GetBinPaths returns the absolute path to the bin directory.
func (p *MavenPkgProvider) GetBinPaths(tool string, installPath string, version string) ([]string, error) {
	return []string{filepath.Join(installPath, "bin")}, nil
}

// GetEnvVars returns no special environment variables.
func (p *MavenPkgProvider) GetEnvVars(tool string, installPath string, version string) (map[string]string, error) {
	return make(map[string]string), nil
}

func (p *MavenPkgProvider) Uninstall(ctx context.Context, tool string, installPath string, version string) error {
	return nil
}
