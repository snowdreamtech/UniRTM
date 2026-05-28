// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pelletier/go-toml/v2"
		"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/service"
	"github.com/spf13/cobra"

	"github.com/snowdreamtech/unirtm/internal/cli/output"
)

func init() {
	// Add command to root
	if rootCmd != nil {
		rootCmd.AddCommand(toolStubCmd)
	}
}

// toolStubCmd represents the tool-stub command which executes a tool stub.
var toolStubCmd = &cobra.Command{
	Use:   "tool-stub <file> [args...]",
	Short: "Execute a tool stub",
	Long: `Execute a tool stub.

Stubs are placeholders for tools that are not yet installed but are available
in the registry. When a stub is executed, UniRTM prompts to install the tool.`,
	Args:               cobra.MinimumNArgs(1),
	DisableFlagParsing: true,
	RunE:               runToolStub,
}

// toolStubConfig represents the TOML configuration inside a tool stub.
type toolStubConfig struct {
	Tool       string            `toml:"tool"`
	Version    string            `toml:"version"`
	Bin        string            `toml:"bin"`
	InstallEnv map[string]string `toml:"install_env"`
}

// extractTOMLFromBootstrap extracts TOML between # UNIRTM_TOOL_STUB: and # :UNIRTM_TOOL_STUB
// falling back to MISE_TOOL_STUB for compatibility.
func extractTOMLFromBootstrap(content string) string {
	startMarkers := []string{"# UNIRTM_TOOL_STUB:", "# MISE_TOOL_STUB:"}
	endMarkers := []string{"# :UNIRTM_TOOL_STUB", "# :MISE_TOOL_STUB"}

	for i, startMarker := range startMarkers {
		endMarker := endMarkers[i]
		startPos := strings.Index(content, startMarker)
		if startPos == -1 {
			continue
		}
		endPos := strings.Index(content, endMarker)
		if endPos == -1 || startPos >= endPos {
			continue
		}

		between := content[startPos+len(startMarker) : endPos]
		var lines []string
		for _, line := range strings.Split(between, "\n") {
			lines = append(lines, strings.TrimPrefix(line, "# "))
		}
		return strings.TrimSpace(strings.Join(lines, "\n"))
	}
	return content
}

// runToolStub executes the tool-stub command.
func runToolStub(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	file := args[0]
	commandArgs := args[1:]

	contentBytes, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read stub file: %w", err)
	}

	content := string(contentBytes)
	tomlContent := extractTOMLFromBootstrap(content)

	var stub toolStubConfig
	if err := toml.Unmarshal([]byte(tomlContent), &stub); err != nil {
		return fmt.Errorf("failed to parse stub TOML: %w", err)
	}

	// Determine tool name from tool field or derive from stub filename
	toolName := stub.Tool
	if toolName == "" {
		toolName = filepath.Base(file)
	}

	// Set bin to filename if not specified
	binName := stub.Bin
	if binName == "" {
		binName = filepath.Base(file)
	}

	version := stub.Version
	if version == "" {
		version = "latest"
	}

	cfg, _ := config.LoadFull()
	installManager, err := getInstallationManager(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to init installation manager: %w", err)
	}

	autoInstall := cfg != nil && (cfg.Settings.AutoInstall == nil || *cfg.Settings.AutoInstall)

	backendName := installManager.AutoDetectBackend(toolName)

	if autoInstall {
		spec := map[string]service.ToolSpec{
			toolName: {
				Name:        toolName,
				Version:     version,
				BackendName: backendName,
			},
		}
		if err := installManager.EnsureInstalledFromSpecs(ctx, spec); err != nil {
			output.Warningf("Failed to install context tool %s: %v", toolName, err)
		}
	}

	// Resolve the environment for the tool
	additionalEnv := make(map[string]string)
	toolEnv := installManager.ResolveToolEnvBySpec(toolName, version, backendName)
	if len(toolEnv) > 0 {
		mergeEnvMaps(additionalEnv, toolEnv)
	}

	// Add install env overrides
	for k, v := range stub.InstallEnv {
		additionalEnv[k] = v
	}

	// Add shims to PATH
	shimsDir := env.GetShimsDir()
	if existing := additionalEnv["PATH"]; existing != "" {
		additionalEnv["PATH"] = existing + string(os.PathListSeparator) + shimsDir
	} else {
		additionalEnv["PATH"] = shimsDir
	}

	applyEnvMap(additionalEnv)

	// Resolve binary
	var binary string
	if resolved, _, err := installManager.ResolveExecutable(ctx, binName, backend.CurrentPlatform()); err == nil {
		binary = resolved
	}

	if binary == "" {
		var err error
		binary, err = exec.LookPath(binName)
		if err != nil {
			return fmt.Errorf("command not found: %s (checked UniRTM tools and PATH)", binName)
		}
	}

	// Execute
	if runtime.GOOS != "windows" {
		allArgs := append([]string{binary}, commandArgs...)
		return execUnix(binary, allArgs, os.Environ())
	}

	return execWindows(binary, commandArgs)
}
