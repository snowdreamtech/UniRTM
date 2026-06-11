// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
)

// GradlePkgProvider implements the Provider interface for Gradle packages.
// Since Gradle packages (fetched via coordinates) are typically hosted on Maven Central
// and share the same jar format and URL structure as Maven packages, this provider
// simply delegates the actual artifact resolution and shimming to the MavenPkgProvider,
// while preserving the gradle-pkg namespace.
type GradlePkgProvider struct {
	mavenProvider *MavenPkgProvider
}

// NewGradlePkgProvider creates a new Gradle packages provider.
func NewGradlePkgProvider() *GradlePkgProvider {
	return &GradlePkgProvider{
		mavenProvider: NewMavenPkgProvider(),
	}
}

// Name returns the provider identifier.
func (p *GradlePkgProvider) Name() string {
	return "gradle-pkg"
}

func (p *GradlePkgProvider) Install(ctx context.Context, tool string, installPath string, artifactPath string, version string) error {
	return p.mavenProvider.Install(ctx, tool, installPath, artifactPath, version)
}

func (p *GradlePkgProvider) PostInstall(ctx context.Context, tool string, installPath string, version string) error {
	return p.mavenProvider.PostInstall(ctx, tool, installPath, version)
}

func (p *GradlePkgProvider) GenerateShims(tool string, installPath string, version string) (map[string]string, error) {
	return p.mavenProvider.GenerateShims(tool, installPath, version)
}

func (p *GradlePkgProvider) DetectVersion(ctx context.Context, tool string, installPath string) (string, error) {
	return p.mavenProvider.DetectVersion(ctx, tool, installPath)
}

func (p *GradlePkgProvider) ListExecutables(tool string, installPath string, version string) ([]string, error) {
	return p.mavenProvider.ListExecutables(tool, installPath, version)
}

func (p *GradlePkgProvider) GetBinPaths(tool string, installPath string, version string) ([]string, error) {
	return p.mavenProvider.GetBinPaths(tool, installPath, version)
}

func (p *GradlePkgProvider) GetEnvVars(tool string, installPath string, version string) (map[string]string, error) {
	return p.mavenProvider.GetEnvVars(tool, installPath, version)
}

func (p *GradlePkgProvider) Uninstall(ctx context.Context, tool string, installPath string, version string) error {
	return p.mavenProvider.Uninstall(ctx, tool, installPath, version)
}
