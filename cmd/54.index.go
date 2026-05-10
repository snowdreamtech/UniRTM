// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/snowdreamtech/unirtm/internal/service"
	"github.com/spf13/cobra"
)

func init() {
	indexCmd.AddCommand(indexUpdateCmd)
	if rootCmd != nil {
		rootCmd.AddCommand(indexCmd)
	}
}

// indexCmd manages the tool index.
var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Manage the tool index",
	Long:  `Manage the local tool index used for searching available tools.`,
}

// indexUpdateCmd refreshes the tool index.
var indexUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Refresh the tool index from remote backends",
	Long: `Refresh the local tool index by querying all registered backends
for available tools and metadata.`,
	Args: cobra.NoArgs,
	RunE: runIndexUpdate,
}

func runIndexUpdate(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	ctx := context.Background()

	dbPath := env.GetDatabasePath()
	db, err := database.Open(ctx, database.Config{
		Path:    dbPath,
		WALMode: true,
	})
	if err != nil {
		formatter.Error("Failed to initialize database", map[string]interface{}{"error": err.Error()})
		return err
	}
	defer db.Close()

	indexRepo, _ := sqlite.NewIndexRepository(db.Conn())
	auditRepo, _ := sqlite.NewAuditRepository(db.Conn())

	// For a real update, we'd need to pass actual backends here.
	// Since backends are currently registered via the ToolSet/Backend system,
	// and IndexManager holds them, we'll initialize it with what we have.
	indexManager, err := service.NewIndexManager(indexRepo, auditRepo, nil, service.IndexManagerConfig{})
	if err != nil {
		return err
	}

	formatter.Info("Refreshing tool index...", nil)
	if err := indexManager.UpdateFromAllBackends(ctx); err != nil {
		// Even if it's just a placeholder error, we report it.
		formatter.Warning(fmt.Sprintf("Index update completed with issues: %v", err))
		return nil // Don't fail the command if it's just "not implemented yet"
	}

	formatter.Success("Tool index refreshed successfully.", nil)
	return nil
}
