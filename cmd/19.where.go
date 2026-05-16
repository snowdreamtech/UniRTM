// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/spf13/cobra"
)

// init registers the where command to the root command.
func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(whereCmd)
	}
}

// whereCmd represents the where command which prints the installation directory
// of a specific tool version. Equivalent to `mise where <tool> [version]`.
var whereCmd = &cobra.Command{
	Use:   "where <tool> [version]",
	Short: "Display the installation path of a tool",
	Long: `Display the installation path of a specific tool version.

The where command prints the directory where a tool version is installed.
If no version is specified, it returns the path of the latest installed version.

Examples:
  # Show installation path of node (latest installed)
  unirtm where node

  # Show installation path of a specific version
  unirtm where node 20.0.0`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runWhere,
}

// runWhere executes the where command.
// It queries the database for the installation path of the specified tool.
func runWhere(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	cfg, _ := config.LoadFull()
	im, err := getInstallationManager(ctx, cfg)
	if err != nil {
		return err
	}

	// 1. Parse input: could be "node", "node@20", or a binary like "gofmt"
	input := args[0]
	var backendName, toolName, version string
	
	if len(args) == 2 {
		toolName = args[0]
		version = args[1]
	} else {
		_, toolName, version, _ = im.ParseToolSpec(input)
	}

	// 2. Resolve the version and backend
	if version == "" || version == "latest" {
		if cfg != nil {
			if toolSpec, ok := cfg.Tools[toolName]; ok {
				version = toolSpec.Version
				backendName = toolSpec.Backend
			}
		}
	}

	if backendName == "" {
		backendName = im.AutoDetectBackend(toolName)
	}

	// 3. SMART RESOLVE: If not found as a tool, try as a binary
	fsName := env.GetFSToolName(toolName, backendName)
	installPath := filepath.Join(env.GetInstallsDir(), fsName, version)
	
	if _, err := os.Stat(installPath); err != nil {
		// Try resolving it as an executable first
		platform := backend.CurrentPlatform()
		binPath, _, err := im.ResolveExecutable(ctx, input, platform)
		if err == nil && binPath != "" {
			// Found it! Now extract the root directory.
			// Path is like .../installs/<tool>/<version>/bin/<exe>
			// We need the part up to <version>
			dir := filepath.Dir(binPath) // .../bin
			for {
				parent := filepath.Dir(dir)
				if parent == dir || parent == "." || parent == "/" {
					break
				}
				// Check if this parent's parent is the installs directory
				grandparent := filepath.Dir(parent)
				if grandparent == env.GetInstallsDir() {
					// We found .../installs/<tool>/<version>
					fmt.Println(parent)
					return nil
				}
				// Also check for nested installs (backend/tool/version)
				greatgrandparent := filepath.Dir(grandparent)
				if greatgrandparent == env.GetInstallsDir() {
					// We found .../installs/<backend>/<tool>/<version>
					fmt.Println(parent)
					return nil
				}
				dir = parent
			}
		}
		
		// If still not found, return the original error
		return fmt.Errorf("tool %s@%s is not installed", toolName, version)
	}

	fmt.Println(installPath)
	return nil
}
