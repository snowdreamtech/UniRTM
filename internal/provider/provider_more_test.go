// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestAllProviders_Properties(t *testing.T) {
	providers := []Provider{
		NewDenoProvider(),
		NewGoPkgProvider(),
		NewPypiProvider(),
		NewVfoxProvider(),
		NewNodeProvider(),
		NewGemProvider(),
		NewDotnetProvider(),
		NewZigProvider(),
		NewSpmProvider(),
		NewRustProvider(),
		NewContainerProvider("docker"),
		NewUbiProvider(),
		NewElixirProvider(),
		NewGolangProvider(),
		NewSwiftProvider(),
		NewPythonProvider(),
		NewCondaProvider(),
		NewCargoProvider(),
		NewNativeProvider(),
		NewGenericProvider(),
		NewRubyProvider(NewNativeProvider()),
		NewNpmProvider(),
		NewJavaProvider(),
		NewErlangProvider(),
		NewAsdfProvider(),
		NewBunProvider(),
		NewFlutterProvider(),
	}
	ctx := context.Background()

	tmpBinDir := t.TempDir()

	dummies := map[string]string{
		"vfox":   `#!/bin/sh\nexit 0\n`,
		"ubi":    `#!/bin/sh\nexit 0\n`,
		"swift":  `#!/bin/sh\nif [ "$1" = "build" ]; then mkdir -p .build/release; touch .build/release/mockbin; chmod +x .build/release/mockbin; fi\nexit 0\n`,
		"rustup": `#!/bin/sh\nexit 0\n`,
		"cargo":  `#!/bin/sh\nexit 0\n`,
		"rustc":  `#!/bin/sh\necho "rustc 1.70.0 (90c541806 2023-05-31)"\nexit 0\n`,
		"ruby":   `#!/bin/sh\necho "ruby 3.2.2 (2023-03-30 revision e51014f4c1) [x86_64-linux]"\nexit 0\n`,
		"gem":    `#!/bin/sh\nexit 0\n`,
		"deno":   `#!/bin/sh\nexit 0\n`,
		"dotnet": `#!/bin/sh\nexit 0\n`,
		"npm":    `#!/bin/sh\nexit 0\n`,
		"pip":    `#!/bin/sh\nexit 0\n`,
		"python": `#!/bin/sh\necho "Python 3.10.12"\nexit 0\n`,
		"python3": `#!/bin/sh\necho "Python 3.10.12"\nexit 0\n`,
		"conda":  `#!/bin/sh\necho "conda 23.5.0"\nexit 0\n`,
		"docker": `#!/bin/sh\nexit 0\n`,
		"podman": `#!/bin/sh\nexit 0\n`,
		"elixir": `#!/bin/sh\nexit 0\n`,
		"go":     `#!/bin/sh\necho "go version go1.21.0 darwin/arm64"\nexit 0\n`,
		"java":   `#!/bin/sh\nexit 0\n`,
		"erl":    `#!/bin/sh\nexit 0\n`,
		"asdf":   `#!/bin/sh\nexit 0\n`,
		"bun":    `#!/bin/sh\nexit 0\n`,
		"flutter": `#!/bin/sh\nexit 0\n`,
		"zig":    `#!/bin/sh\nexit 0\n`,
		"git":    `#!/bin/sh\nexit 0\n`,
	}

	for cmd, script := range dummies {
		name := cmd
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		path := filepath.Join(tmpBinDir, name)
		if runtime.GOOS == "windows" {
			_ = os.WriteFile(path, []byte(""), 0755)
		} else {
			_ = os.WriteFile(path, []byte(script), 0755)
		}
	}

	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", tmpBinDir+string(os.PathListSeparator)+oldPath)

	for _, p := range providers {
		installPath := t.TempDir()
		binDir := filepath.Join(installPath, "bin")
		os.MkdirAll(binDir, 0755)

		for cmd, script := range dummies {
			name := cmd
			if runtime.GOOS == "windows" {
				name += ".exe"
			}
			path := filepath.Join(binDir, name)
			if runtime.GOOS == "windows" {
				_ = os.WriteFile(path, []byte(""), 0755)
			} else {
				_ = os.WriteFile(path, []byte(script), 0755)
			}
		}

		cargoBin := filepath.Join(installPath, "cargo", "bin")
		os.MkdirAll(cargoBin, 0755)
		if runtime.GOOS == "windows" {
			_ = os.WriteFile(filepath.Join(cargoBin, "dummy.exe"), []byte(""), 0755)
		} else {
			_ = os.WriteFile(filepath.Join(cargoBin, "dummy"), []byte("#!/bin/sh\nexit 0\n"), 0755)
		}

		_ = p.Name()
		_ = p.Install(ctx, "tool", installPath, "artifactPath", "1.0.0")
		_ = p.PostInstall(ctx, "tool", installPath, "1.0.0")
		_, _ = p.GenerateShims("tool", installPath, "1.0.0")
		_, _ = p.DetectVersion(ctx, "tool", installPath)
		_, _ = p.ListExecutables("tool", installPath, "1.0.0")
		_, _ = p.GetBinPaths("tool", installPath, "1.0.0")
		_, _ = p.GetEnvVars("tool", installPath, "1.0.0")
		_ = p.Uninstall(ctx, "tool", installPath, "1.0.0")
	}
}

func TestRegistry_Coverage(t *testing.T) {
	r := NewRegistry()
	p := NewGenericProvider()
	r.Register("test", p)
	_ = r.Has("test")
	_ = r.Get("test")
	_, _ = r.GetExact("test")
	r.List()
	r.Unregister("test")

	Register("test", p)
	_ = Has("test")
	_ = Get("test")
	_, _ = GetExact("test")
	List()
	Unregister("test")
}
