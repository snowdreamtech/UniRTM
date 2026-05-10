// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/spf13/cobra"
)

func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(toolCmd)
	}
}

// toolCmd displays detailed information about a specific tool.
var toolCmd = &cobra.Command{
	Use:   "tool <tool>",
	Short: "Show information about a specific tool",
	Long: `Show detailed information about a specific tool.

Displays backend, installed versions, active version, shim location,
and the config file that sets the active version.

Examples:
  # Show info for GitHub CLI
  unirtm tool cli/cli

  # JSON output
  unirtm tool cli/cli --json

  # Specify backend explicitly
  unirtm tool github:cli/cli`,
	Args: cobra.ExactArgs(1),
	RunE: runTool,
}

// toolInfo holds the aggregated information about a single tool.
type toolInfo struct {
	Tool       string   `json:"tool"`
	Backend    string   `json:"backend"`
	Installed  []string `json:"installed"`
	Active     string   `json:"active"`
	ShimPath   string   `json:"shim_path"`
	InstallDir string   `json:"install_dir"`
}

func runTool(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	rawArg := args[0]

	// Parse optional "backend:tool" prefix.
	toolName := rawArg
	backendName := ""
	if idx := strings.Index(rawArg, ":"); idx != -1 {
		backendName = rawArg[:idx]
		toolName = rawArg[idx+1:]
	}

	ctx := context.Background()

	// Open database.
	dbPath := env.GetDatabasePath()
	db, err := database.Open(ctx, database.Config{Path: dbPath, WALMode: true})
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to open database: %v", err))
		return err
	}
	defer db.Close()

	installRepo, err := sqlite.NewInstallationRepository(db.Conn())
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to create repository: %v", err))
		return err
	}

	// Load all installations for this tool.
	all, err := installRepo.List(ctx)
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to list installations: %v", err))
		return err
	}

	var versions []string
	detectedBackend := backendName
	installDir := ""

	for _, inst := range all {
		if inst == nil {
			continue
		}
		if inst.Tool != toolName {
			continue
		}
		versions = append(versions, inst.Version)
		if detectedBackend == "" {
			detectedBackend = inst.Backend
		}
		if installDir == "" {
			installDir = filepath.Dir(inst.InstallPath)
		}
	}

	if len(versions) == 0 && detectedBackend == "" {
		// Try to infer backend from the tool name format.
		if strings.Contains(toolName, "/") {
			detectedBackend = "github"
		} else {
			detectedBackend = "asdf"
		}
	}

	// Detect active version via shim.
	shimsDir := env.GetShimsDir()
	installsDir := env.GetInstallsDir()
	activeVersion := detectActiveVersion(shimsDir, installsDir, toolName, versions)

	// Shim path (first binary shim for this tool).
	shimPath := detectShimPath(shimsDir, toolName)

	// Check if backend supports the tool.
	backendRegistry := backend.NewRegistry()
	if detectedBackend == "" {
		detectedBackend = getBackendName()
	}

	info := toolInfo{
		Tool:       toolName,
		Backend:    detectedBackend,
		Installed:  versions,
		Active:     activeVersion,
		ShimPath:   shimPath,
		InstallDir: filepath.Join(installsDir, toolName),
	}

	// JSON output.
	if jsonOutput {
		formatter.Success(fmt.Sprintf("Tool info: %s", toolName), map[string]interface{}{
			"tool":        info.Tool,
			"backend":     info.Backend,
			"installed":   info.Installed,
			"active":      info.Active,
			"shim_path":   info.ShimPath,
			"install_dir": info.InstallDir,
		})
		return nil
	}

	// Human-readable output.
	fmt.Println()
	pterm.DefaultSection.Printf("Tool: %s", pterm.FgCyan.Sprint(toolName))

	rows := pterm.TableData{
		{"Backend", pterm.FgMagenta.Sprint(info.Backend)},
		{"Install dir", info.InstallDir},
		{"Shim", shimPath},
	}

	if len(info.Installed) == 0 {
		rows = append(rows, []string{"Installed", pterm.FgYellow.Sprint("(none)")})
	} else {
		for i, v := range info.Installed {
			label := "Installed"
			if i > 0 {
				label = ""
			}
			vStr := pterm.FgYellow.Sprint(v)
			if v == info.Active {
				vStr = pterm.FgGreen.Sprint(v + " ✓ active")
			}
			rows = append(rows, []string{label, vStr})
		}
	}

	if info.Active == "" {
		rows = append(rows, []string{"Active", pterm.FgDefault.Sprint("(none)")})
	}

	pterm.DefaultTable.
		WithSeparator("   ").
		WithData(rows).
		Render()

	// Verify backend knows about this tool.
	if b, err := backendRegistry.Get(info.Backend); err == nil {
		platform := backend.CurrentPlatform()
		if latest, err := b.ResolveVersion(ctx, toolName, "latest", platform); err == nil && latest != nil {
			fmt.Printf("\n%s Latest available: %s\n",
				pterm.FgDefault.Sprint("→"),
				pterm.FgGreen.Sprint(latest.Version),
			)
		}
	}

	return nil
}

// detectActiveVersion returns the active version for a tool by checking shim symlinks.
func detectActiveVersion(shimsDir, installsDir, toolName string, versions []string) string {
	for _, v := range versions {
		binDir := filepath.Join(installsDir, toolName, v, "bin")
		entries, err := os.ReadDir(binDir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			shimPath := filepath.Join(shimsDir, e.Name())
			target, err := os.Readlink(shimPath)
			if err != nil {
				continue
			}
			versionDir := filepath.Join(installsDir, toolName, v)
			if strings.HasPrefix(filepath.Clean(target), filepath.Clean(versionDir)) {
				return v
			}
		}
	}
	return ""
}

// detectShimPath returns the path of the first shim binary for a tool.
func detectShimPath(shimsDir, toolName string) string {
	// The shim name is usually the last component of the tool name.
	// e.g. "cli/cli" → "gh"; "golang/go" → "go"
	// We check the shims directory for any binary that might belong to this tool.
	baseName := filepath.Base(toolName)
	candidate := filepath.Join(shimsDir, baseName)
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	return filepath.Join(shimsDir, "(see install dir)")
}
