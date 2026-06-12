package provider

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
)

// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// PubProvider implements the Provider interface for Dart Pub packages.
type PubProvider struct {
}

// NewPubProvider creates a new pub provider.
func NewPubProvider() *PubProvider {
	return &PubProvider{}
}

func (p *PubProvider) Name() string {
	return "pub"
}

func (p *PubProvider) Install(ctx context.Context, tool string, installPath string, artifactPath string, version string) error {
	if err := os.MkdirAll(installPath, 0755); err != nil {
		return err
	}

	dartCmd, err := p.findDart()
	if err != nil {
		return NewProviderError(p.Name(), tool, version, "failed to find native dart", err)
	}

	var cmdArgs []string
	baseCmd := filepath.Base(dartCmd)
	if baseCmd == "pub" || baseCmd == "pub.bat" || baseCmd == "pub.exe" {
		cmdArgs = []string{"global", "activate"}
	} else {
		cmdArgs = []string{"pub", "global", "activate"}
	}

	if version != "" && version != "latest" {
		cmdArgs = append(cmdArgs, tool, version)
	} else {
		cmdArgs = append(cmdArgs, tool)
	}

	logger.Debug("Installing pub package", map[string]interface{}{"pkg": tool, "version": version, "prefix": installPath})

	cmd := exec.CommandContext(ctx, dartCmd, cmdArgs...)
	if ctx != nil && ctx.Value("quietProgress") == true {
		cmd.Stdout = nil
		cmd.Stderr = nil
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	noProxyEnv := GetNoProxyEnv()
	var finalEnv []string
	for _, e := range noProxyEnv {
		if !strings.HasPrefix(e, "PUB_CACHE=") {
			finalEnv = append(finalEnv, e)
		}
	}
	finalEnv = append(finalEnv, fmt.Sprintf("PUB_CACHE=%s", installPath))
	cmd.Env = finalEnv

	if err := cmd.Run(); err != nil {
		return NewProviderError(p.Name(), tool, version, "pub install failed", err)
	}

	return nil
}

func (p *PubProvider) PostInstall(ctx context.Context, tool string, installPath string, version string) error {
	return nil
}

func (p *PubProvider) GenerateShims(tool string, installPath string, version string) (map[string]string, error) {
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

func (p *PubProvider) DetectVersion(ctx context.Context, tool string, installPath string) (string, error) {
	return filepath.Base(installPath), nil
}

func (p *PubProvider) ListExecutables(tool string, installPath string, version string) ([]string, error) {
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
				ext := strings.ToLower(filepath.Ext(entry.Name()))
				if info.Mode()&0111 != 0 || ext == ".bat" || ext == ".cmd" || ext == ".exe" {
					executables = append(executables, filepath.Join(binDir, entry.Name()))
				}
			}
		}
	}

	return executables, nil
}

func (p *PubProvider) GetBinPaths(tool string, installPath string, version string) ([]string, error) {
	binDir := filepath.Join(installPath, "bin")
	return []string{binDir}, nil
}

func (p *PubProvider) GetEnvVars(tool string, installPath string, version string) (map[string]string, error) {
	envVars := make(map[string]string)
	envVars["PUB_CACHE"] = installPath
	return envVars, nil
}

func (p *PubProvider) Uninstall(ctx context.Context, tool string, installPath string, version string) error {
	return nil
}

func (p *PubProvider) findDart() (string, error) {
	dartCmd := ""
	for _, backendPrefix := range []string{"dart-dart", "github-dart", "ubi-dart", "native-dart"} {
		dartInstallBase := filepath.Join(env.GetInstallsDir(), backendPrefix)
		if dirs, err := os.ReadDir(dartInstallBase); err == nil {
			for _, d := range dirs {
				if d.IsDir() {
					binPath := filepath.Join(dartInstallBase, d.Name(), "bin", "dart")
					if runtime.GOOS == "windows" {
						binPath += ".exe"
					}
					if _, err := os.Stat(binPath); err == nil {
						dartCmd = binPath
						break
					}
					rootPath := filepath.Join(dartInstallBase, d.Name(), "dart")
					if runtime.GOOS == "windows" {
						rootPath += ".exe"
					}
					if _, err := os.Stat(rootPath); err == nil {
						dartCmd = rootPath
						break
					}
					// Sometimes dart is nested in dart-sdk/bin
					sdkPath := filepath.Join(dartInstallBase, d.Name(), "dart-sdk", "bin", "dart")
					if runtime.GOOS == "windows" {
						sdkPath += ".exe"
					}
					if _, err := os.Stat(sdkPath); err == nil {
						dartCmd = sdkPath
						break
					}
				}
			}
		}
		if dartCmd != "" {
			break
		}
	}

	if dartCmd == "" {
		return "", fmt.Errorf("dart is required but was not found natively managed by UniRTM")
	}

	return dartCmd, nil
}
