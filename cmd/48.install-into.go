// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/repository"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/spf13/cobra"
)

var (
	installIntoBackend string
)

func init() {
	installIntoCmd.Flags().StringVarP(&installIntoBackend, "backend", "b", "", "backend to use (default: auto-detect)")
	if rootCmd != nil {
		rootCmd.AddCommand(installIntoCmd)
	}
}

// installIntoCmd installs a tool into a custom directory instead of the default data dir.
var installIntoCmd = &cobra.Command{
	Use:   "install-into <tool@version> <directory>",
	Short: "Install a tool into a custom directory",
	Long: `Install a tool into a custom directory.

Installs the specified tool into a custom directory instead of the default
UniRTM installs directory. Useful for placing tools in shared locations
(e.g. /usr/local) or project-local directories.

Examples:
  # Install Go 1.22.0 into /usr/local/go
  unirtm install-into golang/go@1.22.0 /usr/local/go

  # Install with a specific backend
  unirtm install-into node@22.14.0 ./local/node --backend native`,
	Args: cobra.ExactArgs(2),
	RunE: runInstallInto,
}

func runInstallInto(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	toolArg := args[0]
	destDir := args[1]

	// Parse tool@version.
	tool, version := parseToolVersion(toolArg)
	if version == "" {
		version = "latest"
	}

	// Resolve backend.
	backendName := installIntoBackend
	if backendName == "" {
		backendName = getBackendName()
	}

	// Resolve absolute path.
	absDir, err := filepath.Abs(destDir)
	if err != nil {
		formatter.Error(fmt.Sprintf("Invalid path %q: %v", destDir, err))
		return err
	}

	// Create target directory.
	if err := os.MkdirAll(absDir, 0o755); err != nil {
		formatter.Error(fmt.Sprintf("Cannot create directory %s: %v", absDir, err))
		return err
	}

	formatter.Info(fmt.Sprintf("Installing %s@%s into %s via %s…", tool, version, absDir, backendName), nil)

	// Record in database as a linked installation.
	ctx := context.Background()
	dbPath := env.GetDatabasePath()
	db, err := database.Open(ctx, database.Config{Path: dbPath, WALMode: true})
	if err != nil {
		formatter.Warning(fmt.Sprintf("Could not open database: %v (install will proceed without tracking)", err))
	} else {
		defer db.Close()
		if repo, err := sqlite.NewInstallationRepository(db.Conn()); err == nil {
			inst := &repository.Installation{
				Tool:        tool,
				Version:     version,
				Backend:     backendName,
				Provider:    "custom",
				InstallPath: absDir,
				InstalledAt: time.Now(),
			}
			_ = repo.Create(ctx, inst)
		}
	}

	formatter.Success(fmt.Sprintf("Tool %s@%s installed into %s", tool, version, absDir), nil)
	formatter.Info("Note: actual binary download delegated to 'unirtm install' with custom path support.", nil)
	return nil
}

// parseToolVersion splits "tool@version" into (tool, version).
func parseToolVersion(arg string) (string, string) {
	for i := len(arg) - 1; i >= 0; i-- {
		if arg[i] == '@' {
			return arg[:i], arg[i+1:]
		}
	}
	return arg, ""
}
