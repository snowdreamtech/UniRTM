// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/spf13/cobra"
)

func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(latestCmd)
	}
}

// latestCmd queries the latest available version for a tool.
var latestCmd = &cobra.Command{
	Use:   "latest <tool> [version-prefix]",
	Short: "Get the latest available version for a tool",
	Long: `Get the latest available version for a tool from its backend.

When a version prefix is provided, returns the latest patch release within
that prefix (e.g. "1.22" returns the latest 1.22.x release).

Examples:
  # Latest release of the GitHub CLI
  unirtm latest cli/cli

  # Latest 1.22.x release of Go
  unirtm latest golang/go 1.22

  # JSON output
  unirtm latest cli/cli --json

  # Use a specific backend
  unirtm latest typescript --backend npm`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runLatest,
}

var (
	latestBackend string
)

func init() {
	latestCmd.Flags().StringVarP(&latestBackend, "backend", "b", "", "backend to use (default: auto-detect)")
}

func runLatest(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	tool := args[0]
	versionPrefix := ""
	if len(args) == 2 {
		versionPrefix = args[1]
	}

	// Parse "backend:tool" syntax.
	backendName := latestBackend
	if strings.Contains(tool, ":") {
		parts := strings.SplitN(tool, ":", 2)
		backendName = parts[0]
		tool = parts[1]
	}
	if backendName == "" {
		backendName = getBackendName()
	}

	ctx := context.Background()
	backendRegistry := backend.NewRegistry()

	b, err := backendRegistry.Get(backendName)
	if err != nil {
		formatter.Error(fmt.Sprintf("Backend %q not found: %v", backendName, err))
		return err
	}

	platform := backend.CurrentPlatform()

	// Build version request: "latest" or prefix-qualified.
	versionReq := "latest"
	if versionPrefix != "" {
		versionReq = versionPrefix
	}

	info, err := b.ResolveVersion(ctx, tool, versionReq, platform)
	if err != nil {
		formatter.Error(fmt.Sprintf("Could not resolve latest version for %s: %v", tool, err))
		return err
	}
	if info == nil {
		formatter.Error(fmt.Sprintf("No version found for %s", tool))
		return fmt.Errorf("no version found for %s", tool)
	}

	if jsonOutput {
		formatter.Success(fmt.Sprintf("Latest version of %s", tool), map[string]interface{}{
			"tool":    tool,
			"version": info.Version,
			"backend": backendName,
		})
		return nil
	}

	// Plain output — just the version string (scriptable).
	fmt.Println(info.Version)
	return nil
}
