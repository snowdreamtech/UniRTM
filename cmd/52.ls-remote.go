// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/provider/native"
	"github.com/spf13/cobra"
)

var (
	// lsRemoteBackend specifies the backend to use for listing versions
	lsRemoteBackend string
)

func init() {
	// Register command flags
	lsRemoteCmd.Flags().StringVarP(&lsRemoteBackend, "backend", "b", "", "backend to use for listing versions (default: auto-detect)")

	// Add command to root
	if rootCmd != nil {
		rootCmd.AddCommand(lsRemoteCmd)
	}
}

// lsRemoteCmd represents the ls-remote command which lists available versions for a tool.
var lsRemoteCmd = &cobra.Command{
	Use:     "ls-remote <tool> [version-prefix]",
	Aliases: []string{"lsr"},
	Short:   "List runtime versions available for install",
	Long: `List runtime versions available for install from the backend.

The results are fetched from the remote backend and may be cached locally.

Examples:
  # List all available versions of Node.js
  unirtm ls-remote node

  # List versions matching a prefix
  unirtm ls-remote node 20

  # Use a specific backend
  unirtm ls-remote typescript --backend npm

  # JSON output
  unirtm ls-remote node --json`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runLsRemote,
}

// runLsRemote executes the ls-remote command.
func runLsRemote(cmd *cobra.Command, args []string) error {
	// Load project configuration
	cfg, _ := loadConfig(context.Background())
	if cfg != nil {
		cfg.ApplyEnvironment()
	}

	formatter := getFormatter(cfg)

	ctx := context.Background()
	im, _ := getInstallationManager(ctx, cfg)

	backendName, tool, versionPrefix, explicit := im.ParseToolSpec(args[0])
	if lsRemoteBackend != "" {
		backendName = lsRemoteBackend
	}

	if len(args) == 2 {
		versionPrefix = args[1]
	} else if !explicit {
		versionPrefix = ""
	}

	ctx = context.Background()
	backendRegistry := backend.NewRegistry()

	b, err := backendRegistry.Get(backendName)
	if err != nil {
		formatter.Error(fmt.Sprintf("Backend %q not found: %v", backendName, err))
		return err
	}

	platform := backend.CurrentPlatform()
	versions, err := b.ListVersions(ctx, tool, platform)
	if err != nil {
		formatter.Error(fmt.Sprintf("Could not list versions for %s: %v", tool, err))
		return err
	}

	filteredVersions := versions
	if versionPrefix != "" {
		filteredVersions = nil
		for _, v := range versions {
			if strings.HasPrefix(v.Version, versionPrefix) {
				filteredVersions = append(filteredVersions, v)
			}
		}
	}

	if jsonOutput {
		formatter.Success(fmt.Sprintf("Available versions for %s", tool), map[string]interface{}{
			"tool":     tool,
			"versions": filteredVersions,
			"backend":  backendName,
		})
		return nil
	}

	if len(filteredVersions) == 0 {
		if versionPrefix != "" {
			formatter.Info(fmt.Sprintf("No versions found for %s matching prefix %q", tool, versionPrefix), nil)
		} else {
			formatter.Info(fmt.Sprintf("No versions found for %s", tool), nil)
		}
		return nil
	}

	for _, v := range filteredVersions {
		fmt.Println(v.Version)
	}

	return nil
}

// getLsRemoteBackendName returns the backend name to use.
func getLsRemoteBackendName(tool string) string {
	if lsRemoteBackend != "" {
		return lsRemoteBackend
	}

	// If the tool is a known native tool, use the native backend by default
	if native.IsNativeTool(tool) {
		return "native"
	}

	return "github" // Fallback to github
}
