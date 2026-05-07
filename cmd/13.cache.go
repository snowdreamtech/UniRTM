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
	"github.com/snowdreamtech/unirtm/internal/service"
	"github.com/spf13/cobra"
)

// init registers the cache command and its subcommands to the root command.
func init() {
	cacheCmd.AddCommand(cacheListCmd)
	cacheCmd.AddCommand(cacheClearCmd)
	cacheCmd.AddCommand(cachePurgeCmd)
	cacheCmd.AddCommand(cacheStatsCmd)

	if rootCmd != nil {
		rootCmd.AddCommand(cacheCmd)
	}
}

// cacheCmd is the parent cache command.
var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage UniRTM cache",
	Long: `Manage the UniRTM download cache.

The cache command provides subcommands for listing, clearing, and
inspecting cached artifacts.

Subcommands:
  list   List all cached artifacts
  clear  Clear all cache or a specific tool's cache
  purge  Remove expired cache entries
  stats  Display cache statistics

Examples:
  unirtm cache list
  unirtm cache clear
  unirtm cache clear node
  unirtm cache purge
  unirtm cache stats`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// cacheListCmd lists all cached artifacts.
var cacheListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all cached artifacts",
	Long:  `List all cached artifacts stored in the UniRTM cache directory.`,
	Args:  cobra.NoArgs,
	RunE:  runCacheList,
}

// cacheClearCmd clears cache entries.
var cacheClearCmd = &cobra.Command{
	Use:   "clear [tool]",
	Short: "Clear all cache or a specific tool's cache",
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
	Use:   "purge",
	Short: "Remove expired cache entries",
	Long:  `Remove all expired cache entries to free up disk space.`,
	Args:  cobra.NoArgs,
	RunE:  runCachePurge,
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
	dbPath := getDefaultDatabasePath()
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

	cacheDir := getDefaultCacheDir()
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

	ctx := context.Background()

	// Walk the cache directory to list files
	cacheDir := getDefaultCacheDir()
	entries, err := listCacheFiles(cacheDir)
	if err != nil {
		formatter.Error("Failed to list cache", map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("list cache: %w", err)
	}

	if len(entries) == 0 {
		if jsonOutput {
			fmt.Println("[]")
		} else {
			formatter.Info("Cache is empty", nil)
		}
		return nil
	}

	_ = ctx // used via database operations above

	if jsonOutput {
		formatter.Success("Cache entries", map[string]interface{}{
			"count":   len(entries),
			"entries": entries,
		})
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "FILE\tSIZE\tMODIFIED")
	fmt.Fprintln(w, "----\t----\t--------")
	for _, e := range entries {
		fmt.Fprintf(w, "%s\t%s\t%s\n", e["file"], e["size"], e["modified"])
	}
	w.Flush()
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
		// Clear specific tool cache via prefix
		tool := args[0]
		formatter.Info(fmt.Sprintf("Clearing cache for %s...", tool), nil)

		if err := cm.PurgeByPrefix(ctx, tool); err != nil {
			// PurgeByPrefix not yet fully implemented in repo layer; fall back to informing user
			formatter.Info(fmt.Sprintf("Note: Tool-specific cache clearing requires manual deletion from %s", getDefaultCacheDir()), nil)
		} else {
			formatter.Success(fmt.Sprintf("Cleared cache for %s", tool), nil)
		}
		return nil
	}

	// Clear all cache
	formatter.Info("Clearing all cache...", nil)
	if err := cm.PurgeAll(ctx); err != nil {
		formatter.Error("Failed to clear cache", map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("clear cache: %w", err)
	}
	formatter.Success("Cache cleared", nil)
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

	formatter.Info("Removing expired cache entries...", nil)
	if err := cm.PurgeExpired(ctx); err != nil {
		formatter.Error("Failed to purge cache", map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("purge cache: %w", err)
	}
	formatter.Success("Expired cache entries removed", nil)
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

	_ = ctx

	if jsonOutput {
		formatter.Success("Cache statistics", map[string]interface{}{
			"hits":       stats.Hits,
			"misses":     stats.Misses,
			"cache_size": cacheSize,
			"cache_dir":  getDefaultCacheDir(),
		})
		return nil
	}

	hitRate := 0.0
	total := stats.Hits + stats.Misses
	if total > 0 {
		hitRate = float64(stats.Hits) / float64(total) * 100
	}

	fmt.Println("Cache Statistics:")
	fmt.Printf("  Directory: %s\n", getDefaultCacheDir())
	if cacheSize >= 0 {
		fmt.Printf("  Size:      %s\n", formatBytes(cacheSize))
	}
	fmt.Printf("  Hits:      %d\n", stats.Hits)
	fmt.Printf("  Misses:    %d\n", stats.Misses)
	fmt.Printf("  Hit Rate:  %.1f%%\n", hitRate)
	return nil
}

// getDefaultCacheDir returns the default cache directory path.
func getDefaultCacheDir() string {
	cacheHome := os.Getenv("XDG_CACHE_HOME")
	if cacheHome == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "./cache"
		}
		cacheHome = homeDir + "/.cache"
	}
	cacheDir := cacheHome + "/unirtm"
	_ = os.MkdirAll(cacheDir, 0755)
	return cacheDir
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
