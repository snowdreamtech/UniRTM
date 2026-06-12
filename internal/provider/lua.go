// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
)

// LuaProvider implements the Provider interface for Lua via LuaBinaries.
type LuaProvider struct{}

// NewLuaProvider creates a new LuaProvider instance.
func NewLuaProvider() Provider {
	return &LuaProvider{}
}

// Name returns the provider name.
func (p *LuaProvider) Name() string {
	return "lua"
}

// EnsureInstalled is removed as it's not in the new Provider interface

// Install installs the tool.
func (p *LuaProvider) Install(ctx context.Context, tool string, installPath string, artifactPath string, version string) error {
	// 1. Download and Install Lua
	if runtime.GOOS == "windows" {
		url, err := resolveLuaBinariesURL(version, runtime.GOARCH)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(installPath, 0755); err != nil {
			return fmt.Errorf("failed to create install directory: %w", err)
		}
		if err := downloadAndExtract(ctx, url, installPath); err != nil {
			return err
		}
	} else {
		if err := downloadAndCompileSource(ctx, version, installPath); err != nil {
			return err
		}
	}

	// 2. Bootstrap LuaRocks natively
	// We use the same installPath so luarocks is inside the lua directory
	if err := bootstrapLuaRocks(ctx, installPath, version); err != nil {
		return fmt.Errorf("failed to bootstrap LuaRocks: %w", err)
	}

	return nil
}

// PostInstall performs any post-installation steps.
func (p *LuaProvider) PostInstall(ctx context.Context, tool string, installPath string, version string) error {
	return nil
}

// GenerateShims generates shim scripts for the tool's executables.
func (p *LuaProvider) GenerateShims(tool string, installPath string, version string) (map[string]string, error) {
	return nil, nil // let generic shim generator handle it
}

// DetectVersion detects the version of an installed tool.
func (p *LuaProvider) DetectVersion(ctx context.Context, tool string, installPath string) (string, error) {
	return "", nil // default generic detection
}

// Uninstall uninstalls the tool.
func (p *LuaProvider) Uninstall(ctx context.Context, tool string, installPath string, version string) error {
	return os.RemoveAll(installPath)
}

// GetBinPaths returns the binary paths for the tool.
func (p *LuaProvider) GetBinPaths(tool string, installPath string, version string) ([]string, error) {
	// For LuaBinaries, binaries are in the root of the extracted archive.
	// For LuaRocks, binaries are in the 'bin' subdirectory (or root depending on install.bat).
	return []string{installPath, filepath.Join(installPath, "bin")}, nil
}

// ListExecutables lists the executables provided by the tool.
func (p *LuaProvider) ListExecutables(tool string, installPath string, version string) ([]string, error) {
	return []string{"lua", "luac", "luarocks", "luarocks-admin"}, nil
}

// GetEnvVars returns a map of environment variables.
func (p *LuaProvider) GetEnvVars(tool string, installPath string, version string) (map[string]string, error) {
	return map[string]string{}, nil
}

// resolveLuaBinariesURL resolves the download URL for LuaBinaries (Windows only)
func resolveLuaBinariesURL(version, goarch string) (string, error) {
	var osSuffix string
	if goarch == "amd64" {
		osSuffix = "Win64"
	} else {
		osSuffix = "Win32"
	}

	return fmt.Sprintf("https://sourceforge.net/projects/luabinaries/files/%s/Tools%%20Executables/lua-%s_%s_bin.zip/download", version, version, osSuffix), nil
}

// downloadAndCompileSource downloads Lua source from lua.org and compiles it
func downloadAndCompileSource(ctx context.Context, version, installPath string) error {
	tmpDir, err := os.MkdirTemp("", "lua-src-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	url := fmt.Sprintf("https://www.lua.org/ftp/lua-%s.tar.gz", version)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download lua source: HTTP %d", resp.StatusCode)
	}

	if err := extractTarGz(resp.Body, tmpDir); err != nil {
		return err
	}

	srcDir := filepath.Join(tmpDir, fmt.Sprintf("lua-%s", version))

	target := "linux"
	if runtime.GOOS == "darwin" {
		target = "macosx"
	}

	makeCmd := exec.CommandContext(ctx, "make", target, "install", "INSTALL_TOP="+installPath)
	makeCmd.Dir = srcDir
	out, err := makeCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("make install failed: %v\nOutput: %s", err, string(out))
	}

	return nil
}

// downloadAndExtract downloads and extracts the archive
func downloadAndExtract(ctx context.Context, url, dest string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download lua: HTTP %d", resp.StatusCode)
	}

	if strings.HasSuffix(url, ".zip") || strings.Contains(url, "zip") {
		return extractZip(resp.Body, dest)
	}
	return extractTarGz(resp.Body, dest)
}

func extractZip(r io.Reader, dest string) error {
	tmpFile, err := os.CreateTemp("", "lua-*.zip")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, r); err != nil {
		return err
	}

	archive, err := zip.OpenReader(tmpFile.Name())
	if err != nil {
		return err
	}
	defer archive.Close()

	for _, f := range archive.File {
		fpath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", fpath) // zip slip mitigation
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

func extractTarGz(r io.Reader, dest string) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		fpath := filepath.Join(dest, header.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", fpath) // zip slip mitigation
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(fpath, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
				return err
			}
			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}
	return nil
}

// bootstrapLuaRocks downloads and builds luarocks
func bootstrapLuaRocks(ctx context.Context, luaInstallPath, luaVersion string) error {
	luarocksVer := "3.11.1" // Latest stable

	tmpDir, err := os.MkdirTemp("", "luarocks-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	githubProxy := env.Get("GITHUB_PROXY")
	if githubProxy != "" && !strings.HasSuffix(githubProxy, "/") {
		githubProxy += "/"
	}
	url := fmt.Sprintf("%shttps://github.com/luarocks/luarocks/archive/refs/tags/v%s.tar.gz", githubProxy, luarocksVer)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download luarocks: HTTP %d", resp.StatusCode)
	}

	if err := extractTarGz(resp.Body, tmpDir); err != nil {
		return err
	}

	srcDir := filepath.Join(tmpDir, fmt.Sprintf("luarocks-%s", luarocksVer))

	if runtime.GOOS == "windows" {
		cmd := exec.CommandContext(ctx, "install.bat", "/F", "/MW", "/LUA", luaInstallPath, "/P", luaInstallPath, "/Q")
		cmd.Dir = srcDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("luarocks install.bat failed: %v\nOutput: %s", err, string(out))
		}
	} else {
		cmd := exec.CommandContext(ctx, "./configure", "--prefix="+luaInstallPath, "--with-lua="+luaInstallPath)
		cmd.Dir = srcDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("luarocks configure failed: %v\nOutput: %s", err, string(out))
		}

		makeBuildCmd := exec.CommandContext(ctx, "make", "build")
		makeBuildCmd.Dir = srcDir
		out, err = makeBuildCmd.CombinedOutput()
		if err != nil {
			// fallback to just 'make' if 'make build' fails
			makeBuildCmd = exec.CommandContext(ctx, "make")
			makeBuildCmd.Dir = srcDir
			out, err = makeBuildCmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("luarocks make build failed: %v\nOutput: %s", err, string(out))
			}
		}

		makeCmd := exec.CommandContext(ctx, "make", "install")
		makeCmd.Dir = srcDir
		out, err = makeCmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("luarocks make install failed: %v\nOutput: %s", err, string(out))
		}
	}

	return nil
}
