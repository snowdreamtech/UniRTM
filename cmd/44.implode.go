// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/spf13/cobra"
)

var (
	implodeYes bool
)

func init() {
	implodeCmd.Flags().BoolVarP(&implodeYes, "yes", "y", false, "skip confirmation prompt")
	if rootCmd != nil {
		rootCmd.AddCommand(implodeCmd)
	}
}

// implodeCmd removes all UniRTM data, shims, cache, and database.
var implodeCmd = &cobra.Command{
	Use:   "implode",
	Short: "Remove all UniRTM data, shims, cache, and database",
	Long: `Remove all UniRTM data, shims, cache, and database.

This command deletes:
  • Installed tool files (installs directory)
  • Shims directory
  • Downloads cache
  • SQLite database
  • Plugins directory

The UniRTM binary itself is NOT removed. Config files in ~/.config/unirtm
are also kept.

Use --yes / -y to skip the confirmation prompt in scripts.

Examples:
  unirtm implode
  unirtm implode --yes`,
	Args: cobra.NoArgs,
	RunE: runImplode,
}

func runImplode(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	dataDir := env.GetDataDir()
	targets := []string{
		env.GetInstallsDir(),
		env.GetShimsDir(),
		env.GetDownloadsDir(),
		env.GetDatabasePath(),
		env.GetPluginsDir(),
	}

	if !implodeYes {
		fmt.Printf("\n⚠  This will delete ALL UniRTM data under: %s\n\n", dataDir)
		fmt.Println("  • installs/   (all tool binaries)")
		fmt.Println("  • shims/      (shell wrapper scripts)")
		fmt.Println("  • downloads/  (cached archives)")
		fmt.Println("  • unirtm.db   (installation database)")
		fmt.Println("  • plugins/    (backend plugins)")
		fmt.Print("\nType 'yes' to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "yes" {
			formatter.Info("Implode cancelled.", nil)
			return nil
		}
	}

	anyErr := false
	for _, path := range targets {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if verbose {
				formatter.Info(fmt.Sprintf("Skipping (not found): %s", path), nil)
			}
			continue
		}
		if err := os.RemoveAll(path); err != nil {
			formatter.Warning(fmt.Sprintf("Failed to remove %s: %v", path, err))
			anyErr = true
		} else {
			formatter.Success(fmt.Sprintf("Removed: %s", path), nil)
		}
	}

	if anyErr {
		formatter.Warning("Some directories could not be removed (see above).")
	} else {
		formatter.Success("UniRTM data removed. Run 'unirtm install' to start fresh.", nil)
	}
	return nil
}
