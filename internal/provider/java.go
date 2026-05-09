// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// JavaProvider implements the Provider interface for Java.
type JavaProvider struct {
	generic *GenericProvider
}

// NewJavaProvider creates a new Java provider.
func NewJavaProvider() *JavaProvider {
	return &JavaProvider{
		generic: NewGenericProvider(),
	}
}

// Name returns the provider identifier.
func (j *JavaProvider) Name() string {
	return "java"
}

// Install performs Java-specific installation.
func (j *JavaProvider) Install(ctx context.Context, installPath string, artifactPath string, version string) error {
	return j.generic.Install(ctx, installPath, artifactPath, version)
}

// PostInstall performs post-installation steps.
func (j *JavaProvider) PostInstall(ctx context.Context, installPath string, version string) error {
	return nil
}

// getJavaHome returns the JAVA_HOME directory inside the install path.
// On macOS, JDK tarballs often extract to Contents/Home.
func (j *JavaProvider) getJavaHome(installPath string) string {
	macOSPath := filepath.Join(installPath, "Contents", "Home")
	if _, err := os.Stat(filepath.Join(macOSPath, "bin")); err == nil {
		return macOSPath
	}
	return installPath
}

// GenerateShims generates shims for java executables.
func (j *JavaProvider) GenerateShims(installPath string, version string) (map[string]string, error) {
	shims := make(map[string]string)
	javaHome := j.getJavaHome(installPath)

	executables := []string{"java", "javac", "jar", "javadoc"}
	for _, exe := range executables {
		exePath := filepath.Join(javaHome, "bin", exe)
		if runtime.GOOS == "windows" {
			exePath += ".exe"
		}

		shimContent := j.generateJavaShim(exe, exePath, javaHome, version)
		shims[exe] = shimContent
	}

	return shims, nil
}

// DetectVersion detects Java version.
func (j *JavaProvider) DetectVersion(ctx context.Context, installPath string) (string, error) {
	javaHome := j.getJavaHome(installPath)
	javaPath := filepath.Join(javaHome, "bin", "java")
	if runtime.GOOS == "windows" {
		javaPath += ".exe"
	}

	cmd := exec.CommandContext(ctx, javaPath, "-version")
	// java -version prints to stderr, not stdout
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", NewProviderError("java", "java", "", "failed to detect version", err)
	}

	// Output format is usually: java version "17.0.8" 2023-07-18 LTS ...
	// or openjdk version "17.0.8" ...
	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")
	if len(lines) > 0 {
		firstLine := lines[0]
		parts := strings.Split(firstLine, "\"")
		if len(parts) >= 3 {
			return parts[1], nil
		}
	}

	return "", NewProviderError("java", "java", "", "failed to parse version", nil)
}

// ListExecutables returns Java executables.
func (j *JavaProvider) ListExecutables(installPath string, version string) ([]string, error) {
	executables := []string{"java", "javac", "jar", "javadoc"}
	if runtime.GOOS == "windows" {
		for i := range executables {
			executables[i] += ".exe"
		}
	}
	return executables, nil
}

// Uninstall performs Java-specific cleanup.
func (j *JavaProvider) Uninstall(ctx context.Context, installPath string, version string) error {
	return nil
}

// generateJavaShim generates a Java-specific shim.
func (j *JavaProvider) generateJavaShim(name, exePath, javaHome, version string) string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf(`@echo off
REM UniRTM shim for %s (version %s)
set "JAVA_HOME=%s"
"%s" %%*
`, name, version, javaHome, exePath)
	}

	return fmt.Sprintf(`#!/bin/sh
# UniRTM shim for %s (version %s)
export JAVA_HOME="%s"
exec "%s" "$@"
`, name, version, javaHome, exePath)
}
