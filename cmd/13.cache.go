// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/snowdreamtech/unirtm/internal/service"
	"github.com/spf13/cobra"
)

// init registers the cache command and its subcommands to the root command.
func init() {
	cacheCmd.AddCommand(cacheListCmd)
	cacheCmd.AddCommand(cacheClearCmd)
	cacheCmd.AddCommand(cachePurgeCmd)
	cacheCmd.AddCommand(cacheStatsCmd)
	cacheCmd.AddCommand(cachePathCmd)

	if rootCmd != nil {
		rootCmd.AddCommand(cacheCmd)
	}
}

// cacheCmd is the parent cache command.
var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage UniRTM cache",
	Long: `Manage the UniRTM download cache.

If no subcommand is provided, it displays the path to the cache directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cacheDir := env.GetCacheDir()
		var fileCount int
		var totalSize int64
		_ = filepath.Walk(cacheDir, func(_ string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				fileCount++
				totalSize += info.Size()
			}
			return nil
		})
		sizeStr := fmt.Sprintf("%.2f MB", float64(totalSize)/(1024*1024))
		pterm.DefaultBox.WithTitle(pterm.LightCyan("Cache Information")).Println(
			fmt.Sprintf("%s: %s\n%s: %s\n%s: %s",
				pterm.Bold.Sprint("Path "), pterm.LightBlue(cacheDir),
				pterm.Bold.Sprint("Size "), pterm.LightGreen(sizeStr),
				pterm.Bold.Sprint("Files"), pterm.LightYellow(fileCount)),
		)
		return nil
	},
}

// cachePathCmd displays the cache directory path.
var cachePathCmd = &cobra.Command{
	Use:   "path",
	Short: "Display the cache directory path",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(env.GetCacheDir())
		return nil
	},
}

// cacheListCmd lists all cached artifacts.
var cacheListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all cached artifacts",
	Long:    `List all cached artifacts stored in the UniRTM cache directory.`,
	Args:    cobra.NoArgs,
	RunE:    runCacheList,
}

// cacheClearCmd clears cache entries.
var cacheClearCmd = &cobra.Command{
	Use:     "clear [tool]",
	Aliases: []string{"clean", "remove", "rm"},
	Short:   "Clear all cache or a specific tool's cache",
	Long: `Clear all cache or a specific tool's cached artifacts.

Examples:
  # Clear all cache
  unirtm cache clear

  # Clear cache for a specific tool
  unirtm cache clear node`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCacheClear,
}

// cachePurgeCmd removes expired cache entries.
var cachePurgeCmd = &cobra.Command{
	Use:     "prune",
	Aliases: []string{"purge"},
	Short:   "Remove expired cache entries",
	Long:    `Remove all expired cache entries to free up disk space.`,
	Args:    cobra.NoArgs,
	RunE:    runCachePurge,
}

// cacheStatsCmd displays cache statistics.
var cacheStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Display cache statistics",
	Long:  `Display cache statistics including size, hit rate, and entry count.`,
	Args:  cobra.NoArgs,
	RunE:  runCacheStats,
}

// newCacheManager creates a configured cache manager from the database.
func newCacheManager(ctx context.Context, formatter output.Formatter) (*service.CacheManager, *database.DB, error) {
	dbPath := env.GetDatabasePath()
	db, err := database.Open(ctx, database.Config{
		Path:    dbPath,
		WALMode: true,
	})
	if err != nil {
		if formatter != nil {
			formatter.Error("Failed to initialize database", map[string]interface{}{
				"error": err.Error(),
				"path":  dbPath,
			})
		}
		return nil, nil, fmt.Errorf("initialize database: %w", err)
	}

	cacheRepo, err := sqlite.NewCacheRepository(db.Conn())
	if err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("create cache repository: %w", err)
	}

	auditRepo, _ := sqlite.NewAuditRepository(db.Conn())

	cacheDir := env.GetCacheDir()
	cm, err := service.NewCacheManager(cacheRepo, auditRepo, service.CacheManagerConfig{
		CacheDir: cacheDir,
	})
	if err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("create cache manager: %w", err)
	}

	return cm, db, nil
}

// runCacheList lists all cached artifacts.
//
// Validates: Requirements 10.6, 23.2
func runCacheList(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	// Walk the cache directory to list files
	cacheDir := env.GetCacheDir()
	entries, err := listCacheFiles(cacheDir)
	if err != nil {
		formatter.Error("Failed to list cache", map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("list cache: %w", err)
	}

	if len(entries) == 0 {
		if jsonOutput {
			fmt.Println("[]")
		} else {
			pterm.Info.Println("Cache is empty")
		}
		return nil
	}

	if jsonOutput {
		formatter.Success("Cache entries", map[string]interface{}{
			"count":   len(entries),
			"entries": entries,
		})
		return nil
	}

	var data [][]string
	data = append(data, []string{"File", "Size", "Modified"})
	for _, e := range entries {
		data = append(data, []string{e["file"], e["size"], e["modified"]})
	}
	pterm.DefaultTable.WithHasHeader().WithData(data).Render()
	return nil
}

// runCacheClear clears all or tool-specific cache.
//
// Validates: Requirements 10.6, 23.2
func runCacheClear(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	ctx := context.Background()

	cm, db, err := newCacheManager(ctx, formatter)
	if err != nil {
		return err
	}
	defer db.Close()

	if len(args) == 1 {
		tool := args[0]
		spinner, _ := pterm.DefaultSpinner.Start(fmt.Sprintf("Clearing cache for %s...", tool))

		if err := cm.PurgeByPrefix(ctx, tool); err != nil {
			spinner.Warning(fmt.Sprintf("Tool-specific cache clearing requires manual deletion from %s", env.GetCacheDir()))
		} else {
			spinner.Success(fmt.Sprintf("Cleared cache for %s", tool))
		}
		return nil
	}

	spinner, _ := pterm.DefaultSpinner.Start("Clearing all cache...")
	if err := cm.PurgeAll(ctx); err != nil {
		spinner.Fail(fmt.Sprintf("Failed to clear cache: %v", err))
		return fmt.Errorf("clear cache: %w", err)
	}
	spinner.Success("Cache cleared")
	return nil
}

// runCachePurge removes expired cache entries.
//
// Validates: Requirements 10.6, 23.2
func runCachePurge(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	ctx := context.Background()

	cm, db, err := newCacheManager(ctx, formatter)
	if err != nil {
		return err
	}
	defer db.Close()

	spinner, _ := pterm.DefaultSpinner.Start("Removing expired cache entries...")
	if err := cm.PurgeExpired(ctx); err != nil {
		spinner.Fail(fmt.Sprintf("Failed to purge cache: %v", err))
		return fmt.Errorf("purge cache: %w", err)
	}
	spinner.Success("Expired cache entries removed")
	return nil
}

// runCacheStats displays cache statistics.
//
// Validates: Requirements 10.6, 23.2
func runCacheStats(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	ctx := context.Background()

	cm, db, err := newCacheManager(ctx, formatter)
	if err != nil {
		return err
	}
	defer db.Close()

	stats := cm.GetStats()
	cacheSize, err := cm.GetCacheSize()
	if err != nil {
		cacheSize = -1
	}

	if jsonOutput {
		formatter.Success("Cache statistics", map[string]interface{}{
			"hits":       stats.Hits,
			"misses":     stats.Misses,
			"cache_size": cacheSize,
			"cache_dir":  env.GetCacheDir(),
		})
		return nil
	}

	hitRate := 0.0
	total := stats.Hits + stats.Misses
	if total > 0 {
		hitRate = float64(stats.Hits) / float64(total) * 100
	}

	pterm.DefaultSection.Println("Cache Statistics")
	pterm.BulletListPrinter{
		Items: []pterm.BulletListItem{
			{Level: 0, Text: fmt.Sprintf("Directory: %s", env.GetCacheDir())},
			{Level: 0, Text: fmt.Sprintf("Size:      %s", formatBytes(cacheSize))},
			{Level: 0, Text: fmt.Sprintf("Hits:      %d", stats.Hits)},
			{Level: 0, Text: fmt.Sprintf("Misses:    %d", stats.Misses)},
			{Level: 0, Text: fmt.Sprintf("Hit Rate:  %.1f%%", hitRate)},
		},
	}.Render()
	return nil
}

// getDefaultCacheDir returns the default cache directory path.
func getDefaultCacheDir() string {
	return env.GetCacheDir()
}

// listCacheFiles walks the cache directory and returns a list of file info maps.
func listCacheFiles(dir string) ([]map[string]string, error) {
	var entries []map[string]string

	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return entries, nil
	}

	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read cache directory: %w", err)
	}

	for _, entry := range dirEntries {
		fi, err := entry.Info()
		if err != nil {
			continue
		}
		entries = append(entries, map[string]string{
			"file":     entry.Name(),
			"size":     formatBytes(fi.Size()),
			"modified": fi.ModTime().Format(time.RFC3339),
		})
	}
	return entries, nil
}

// formatBytes formats a byte count as a human-readable string.
func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
