// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
)

func TestAllProviders_Coverage(t *testing.T) {
	nativeProvider := NewNativeProvider()
	providers := []Provider{
		NewPythonProvider(),
		NewNodeProvider(),
		NewNpmProvider(),
		NewPypiProvider(),
		NewRubyProvider(nativeProvider),
		NewRustProvider(),
		NewSpmProvider(),
		NewSwiftProvider(),
		NewUbiProvider(),
		NewVfoxProvider(),
		NewZigProvider(),
		NewGolangProvider(),
		NewGoPkgProvider(),
		NewJavaProvider(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	testPath := "test_path"
	version := "1.0.0"

	oses := []string{"linux", "windows", "darwin"}

	for _, p := range providers {
		for _, osName := range oses {
			t.Run(p.Name()+"_"+osName, func(t *testing.T) {
				oldOS := env.RuntimeGOOS
				env.RuntimeGOOS = osName
				defer func() { env.RuntimeGOOS = oldOS }()

				// Ensure env path wrapper is hit
				_ = p.Name()

				p.DetectVersion(ctx, p.Name(), "/invalid/path")
				p.Install(ctx, p.Name(), "invalid_path", "invalid_artifact", version)
				p.PostInstall(ctx, p.Name(), testPath, version)
				p.GenerateShims(p.Name(), testPath, version)
				p.ListExecutables(p.Name(), testPath, version)
				p.GetBinPaths(p.Name(), testPath, version)
				p.GetEnvVars(p.Name(), testPath, version)
				p.Uninstall(ctx, p.Name(), testPath, version)

				if py, ok := p.(*PythonProvider); ok {
					py.SkipAtomicRename()
					py.getRealPythonPath(testPath)
				}
				if pypi, ok := p.(*PypiProvider); ok {
					pypi.SkipAtomicRename()
				}

				// Trigger MkdirAll / path errors using a file path
				dummyFile := filepath.Join(t.TempDir(), "dummy")
				_ = os.WriteFile(dummyFile, []byte("test"), 0644)
				_ = p.Install(ctx, p.Name(), dummyFile, "invalid", version)
				_, _ = p.ListExecutables(p.Name(), dummyFile, version)
				_, _ = p.GetBinPaths(p.Name(), dummyFile, version)
			})
		}
	}
}
