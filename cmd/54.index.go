// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/snowdreamtech/unirtm/internal/service"
	"github.com/spf13/cobra"
)

func init() {
	indexCmd.AddCommand(indexUpdateCmd)
	indexCmd.AddCommand(indexStatusCmd)
	indexCmd.AddCommand(indexClearCmd)
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

// indexStatusCmd shows the status of the local tool index.
var indexStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of the local tool index",
	Long:  `Show status of the local tool index, including health, size, last update time, and tools count.`,
	Args:  cobra.NoArgs,
	RunE:  runIndexStatus,
}

// indexClearCmd clears the local tool index cache.
var indexClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear the local tool index cache",
	Long:  `Clear the local tool index cache by deleting all entries from the index database.`,
	Args:  cobra.NoArgs,
	RunE:  runIndexClear,
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

	// Initialize the IndexManager with all registered backends
	indexManager, err := service.NewIndexManager(indexRepo, auditRepo, backend.Backends(), service.IndexManagerConfig{})
	if err != nil {
		return err
	}

	formatter.Info("Refreshing tool index from registered backends...", nil)

	startTime := time.Now()
	if err := indexManager.UpdateFromAllBackends(ctx); err != nil {
		formatter.Warning(fmt.Sprintf("Index update completed with issues: %v", err))
		return nil
	}
	duration := time.Since(startTime)

	entries, _ := indexRepo.List(ctx)

	formatter.Success(fmt.Sprintf("Tool index refreshed successfully in %s.", duration.Round(time.Millisecond)), nil)
	fmt.Printf("\nTotal tools indexed and available offline: %d\n", len(entries))
	fmt.Println("You can now search for tools using: unirtm search <query>")

	return nil
}

func runIndexStatus(cmd *cobra.Command, args []string) error {
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

	indexManager, err := service.NewIndexManager(indexRepo, auditRepo, backend.Backends(), service.IndexManagerConfig{})
	if err != nil {
		return err
	}

	entries, err := indexRepo.List(ctx)
	if err != nil {
		formatter.Error("Failed to list index entries", map[string]interface{}{"error": err.Error()})
		return err
	}

	isStale, _ := indexManager.IsStale(ctx)

	// Get file size
	var fileSize int64
	if info, err := os.Stat(dbPath); err == nil {
		fileSize = info.Size()
	}

	// Calculate backend counts and last updated time
	backendCounts := make(map[string]int)
	var lastUpdated time.Time
	for _, entry := range entries {
		backendCounts[entry.Backend]++
		if entry.UpdatedAt.After(lastUpdated) {
			lastUpdated = entry.UpdatedAt
		}
	}

	fmt.Println("\x1b[1;36m=== UniRTM Local Tool Index Status ===\x1b[0m")
	fmt.Printf("Database Path:  %s\n", dbPath)
	fmt.Printf("Database Size:  %.2f KB\n", float64(fileSize)/1024.0)

	statusStr := "\x1b[1;32mHealthy (Up-to-date)\x1b[0m"
	if isStale {
		statusStr = "\x1b[1;33mStale (Requires Update)\x1b[0m"
	}
	if len(entries) == 0 {
		statusStr = "\x1b[1;31mEmpty (Requires Initialization)\x1b[0m"
	}
	fmt.Printf("Index Health:   %s\n", statusStr)
	fmt.Printf("Total Tools:    %d\n", len(entries))

	if !lastUpdated.IsZero() {
		fmt.Printf("Last Updated:   %s (%s ago)\n", lastUpdated.Format(time.RFC1123), time.Since(lastUpdated).Round(time.Second))
	} else {
		fmt.Printf("Last Updated:   Never\n")
	}

	if len(backendCounts) > 0 {
		fmt.Println("\n\x1b[1;34mTools by Backend:\x1b[0m")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
		fmt.Fprintln(w, "  BACKEND\tCOUNT")
		fmt.Fprintln(w, "  -------\t-----")
		for b, count := range backendCounts {
			fmt.Fprintf(w, "  %s\t%d\n", b, count)
		}
		w.Flush()
	}

	fmt.Println()
	return nil
}

func runIndexClear(cmd *cobra.Command, args []string) error {
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

	formatter.Info("Clearing local tool index cache...", nil)

	_, err = db.Conn().ExecContext(ctx, "DELETE FROM tool_index")
	if err != nil {
		formatter.Error("Failed to clear index", map[string]interface{}{"error": err.Error()})
		return err
	}

	formatter.Success("Tool index cache cleared successfully.", nil)
	return nil
}
