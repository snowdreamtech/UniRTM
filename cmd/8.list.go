// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/spf13/cobra"
)

var (
	// listToolFilter filters list output by tool name
	listToolFilter string
)

// init registers the list command to the root command.
func init() {
	// Register command flags
	listCmd.Flags().StringVarP(&listToolFilter, "tool", "t", "", "filter by tool name")

	// Add command to root
	if rootCmd != nil {
		rootCmd.AddCommand(listCmd)
	}
}

// listCmd represents the list command which shows all installed tools.
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all installed development tools",
	Long: `List all installed development tools.

The list command shows all tools that have been installed with UniRTM,
including their version, backend, and installation path.

Examples:
  # List all installed tools
  unirtm list

  # Filter by tool name
  unirtm list --tool node

  # JSON output
  unirtm list --json`,
	Args: cobra.NoArgs,
	RunE: runList,
}

// runList executes the list command.
// It lists all installed tools with their versions, backends, and install paths.
//
// Validates: Requirements 23.2
func runList(cmd *cobra.Command, args []string) error {
	// Create output formatter
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	// Initialize dependencies
	ctx := context.Background()

	// Create database connection
	dbPath := getDefaultDatabasePath()
	db, err := database.Open(ctx, database.Config{
		Path:    dbPath,
		WALMode: true,
	})
	if err != nil {
		formatter.Error("Failed to initialize database", map[string]interface{}{
			"error": err.Error(),
			"path":  dbPath,
		})
		return fmt.Errorf("initialize database: %w", err)
	}
	defer db.Close()

	// Create installation repository
	installRepo, err := sqlite.NewInstallationRepository(db.Conn())
	if err != nil {
		formatter.Error("Failed to create installation repository", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("create installation repository: %w", err)
	}

	// List all installations
	installations, err := installRepo.List(ctx)
	if err != nil {
		formatter.Error("Failed to list installations", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("list installations: %w", err)
	}

	// Apply filter if specified
	if listToolFilter != "" {
		filtered := installations[:0]
		for _, inst := range installations {
			if inst.Tool == listToolFilter {
				filtered = append(filtered, inst)
			}
		}
		installations = filtered
	}

	if len(installations) == 0 {
		if jsonOutput {
			fmt.Println("[]")
		} else {
			formatter.Info("No tools installed", nil)
		}
		return nil
	}

	// JSON output
	if jsonOutput {
		type jsonInstallation struct {
			Tool        string    `json:"tool"`
			Version     string    `json:"version"`
			Backend     string    `json:"backend"`
			InstallPath string    `json:"install_path"`
			InstalledAt time.Time `json:"installed_at"`
		}

		results := make([]jsonInstallation, 0, len(installations))
		for _, inst := range installations {
			results = append(results, jsonInstallation{
				Tool:        inst.Tool,
				Version:     inst.Version,
				Backend:     inst.Backend,
				InstallPath: inst.InstallPath,
				InstalledAt: inst.InstalledAt,
			})
		}

		formatter.Success("Installed tools", map[string]interface{}{
			"count": len(results),
			"tools": results,
		})
		return nil
	}

	// Human-readable table output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TOOL\tVERSION\tBACKEND\tINSTALL PATH\tINSTALLED AT")
	fmt.Fprintln(w, "----\t-------\t-------\t------------\t------------")

	for _, inst := range installations {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			inst.Tool,
			inst.Version,
			inst.Backend,
			inst.InstallPath,
			inst.InstalledAt.Format("2006-01-02 15:04:05"),
		)
	}
	w.Flush()

	return nil
}
