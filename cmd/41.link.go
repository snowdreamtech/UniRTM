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
	linkBackend string
)

func init() {
	linkCmd.Flags().StringVarP(&linkBackend, "backend", "b", "custom", "backend name to record in the database")
	if rootCmd != nil {
		rootCmd.AddCommand(linkCmd)
	}
}

// linkCmd registers an externally installed tool path into UniRTM's management.
var linkCmd = &cobra.Command{
	Use:   "link <tool> <version> <path>",
	Short: "Register an externally installed tool into UniRTM",
	Long: `Register an externally installed tool into UniRTM management.

Links an existing binary or directory into UniRTM's database so it
appears in 'unirtm list' and can be used with 'unirtm use'.
No files are moved or copied.

Examples:
  # Register a locally compiled Go binary
  unirtm link golang/go 1.22.0 /usr/local/go

  # Register with a specific backend label
  unirtm link node 22.14.0 ~/.nvm/versions/node/v22.14.0 --backend nvm`,
	Args: cobra.ExactArgs(3),
	RunE: runLink,
}

func runLink(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	tool := args[0]
	version := args[1]
	installPath := args[2]

	// Resolve to absolute path.
	absPath, err := filepath.Abs(installPath)
	if err != nil {
		formatter.Error(fmt.Sprintf("Invalid path %q: %v", installPath, err))
		return err
	}

	// Verify path exists.
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		formatter.Error(fmt.Sprintf("Path does not exist: %s", absPath))
		return fmt.Errorf("path does not exist: %s", absPath)
	}

	ctx := context.Background()

	dbPath := env.GetDatabasePath()
	db, err := database.Open(ctx, database.Config{Path: dbPath, WALMode: true})
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to open database: %v", err))
		return err
	}
	defer db.Close()

	repo, err := sqlite.NewInstallationRepository(db.Conn())
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to create repository: %v", err))
		return err
	}

	inst := &repository.Installation{
		Tool:        tool,
		Version:     version,
		Backend:     linkBackend,
		Provider:    "custom",
		InstallPath: absPath,
		InstalledAt: time.Now(),
	}

	if err := repo.Create(ctx, inst); err != nil {
		formatter.Error(fmt.Sprintf("Failed to register tool: %v", err))
		return err
	}

	formatter.Success(fmt.Sprintf("Linked %s@%s → %s", tool, version, absPath), nil)
	return nil
}
