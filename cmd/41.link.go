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
	"github.com/snowdreamtech/unirtm/internal/provider"
	"github.com/snowdreamtech/unirtm/internal/repository"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/snowdreamtech/unirtm/internal/service"
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
	Aliases: []string{"ln"},
	Args:    cobra.ExactArgs(3),
	RunE:    runLink,
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

	// 1. Validate the path is a directory (most tool installations are directories, e.g. /usr/local/go or /usr/local/bin)
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to inspect path %s: %v", absPath, err))
		return err
	}
	if !fileInfo.IsDir() {
		formatter.Warning(fmt.Sprintf("Path %s is a file, not a directory. In UniRTM, linked tools should point to their installation root directory.", absPath))
	}

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

	// 2. Automatically generate shims for the newly linked tool
	providerRegistry := provider.NewRegistry()
	p := providerRegistry.GetWithBackend(tool, linkBackend)
	executables, err := p.ListExecutables(tool, absPath, version)
	if err != nil || len(executables) == 0 {
		// Fallback to using the tool name itself as the executable
		executables = []string{tool}
	}

	shimsDir := env.GetShimsDir()
	dataDir := env.GetDataDir()
	generator := service.NewGenerator(shimsDir, dataDir+"/installs")

	if err := generator.GenerateShim(ctx, tool, executables...); err != nil {
		formatter.Warning(fmt.Sprintf("Failed to generate shims for linked tool %s: %v", tool, err))
	} else {
		formatter.Success(fmt.Sprintf("Automatically generated shims for %s: %v", tool, executables), nil)
	}

	return nil
}
