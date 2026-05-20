// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/provider"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/snowdreamtech/unirtm/internal/service"
	"github.com/spf13/cobra"
)

var (
	// reshimTool limits reshim to a specific tool (optional)
	reshimTool string
	// reshimClean removes dead shims that point to uninstalled tools
	reshimClean bool
)

// init registers the reshim command to the root command.
func init() {
	reshimCmd.Flags().StringVarP(&reshimTool, "tool", "t", "", "limit reshim to a specific tool")
	reshimCmd.Flags().BoolVar(&reshimClean, "clean", false, "also remove dead shims for uninstalled tools")
	if rootCmd != nil {
		rootCmd.AddCommand(reshimCmd)
	}
}

// reshimCmd represents the reshim command which regenerates all shim scripts.
// Equivalent to `mise reshim`.
var reshimCmd = &cobra.Command{
	Use:   "reshim",
	Short: "Regenerate shim scripts for installed tools (concurrent + dead shim cleanup)",
	Long: `Regenerate shim scripts for all installed tools using concurrent workers.

This command recreates the shim scripts in the shims directory in parallel,
making it significantly faster than sequential regeneration for large tool sets.
Use --clean to also remove dead shims for tools that are no longer installed.

Examples:
  # Regenerate all shims (parallel)
  unirtm reshim

  # Regenerate shims for a specific tool
  unirtm reshim --tool node

  # Regenerate and clean up dead shims
  unirtm reshim --clean

  # Preview what would happen
  unirtm reshim --clean --dry-run`,
	Args: cobra.NoArgs,
	RunE: runReshim,
}

// runReshim executes the reshim command.
// It queries all installed tools and regenerates their shim scripts concurrently.
func runReshim(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	ctx := context.Background()

	// Open database
	dbPath := env.GetDatabasePath()
	db, err := database.Open(ctx, database.Config{Path: dbPath, WALMode: true})
	if err != nil {
		return fmt.Errorf("initialize database: %w", err)
	}
	defer db.Close()

	installRepo, err := sqlite.NewInstallationRepository(db.Conn())
	if err != nil {
		return fmt.Errorf("create installation repository: %w", err)
	}

	// List all installations
	installations, err := installRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("list installations: %w", err)
	}

	// Filter by tool if specified
	if reshimTool != "" {
		filtered := installations[:0]
		for _, inst := range installations {
			if inst.Tool == reshimTool {
				filtered = append(filtered, inst)
			}
		}
		installations = filtered
	}

	shimsDir := env.GetShimsDir()
	dataDir := env.GetDataDir()
	generator := service.NewGenerator(shimsDir, dataDir+"/installs")
	providerRegistry := provider.NewRegistry()

	// ── Phase 1: Clean up dead shims ──────────────────────────────────────────
	deadRemoved := 0
	if reshimClean {
		// Build a set of all known shim base names from installed tools
		knownShims := make(map[string]bool)
		for _, inst := range installations {
			p := providerRegistry.GetWithBackend(inst.Tool, inst.Backend)
			executables, err := p.ListExecutables(inst.Tool, inst.InstallPath, inst.Version)
			if err != nil {
				executables = []string{inst.Tool}
			}
			for _, exe := range executables {
				knownShims[filepath.Base(exe)] = true
			}
		}

		entries, err := os.ReadDir(shimsDir)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("read shims directory: %w", err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			// Strip platform-specific extensions to get base name
			baseName := name
			if runtime.GOOS == "windows" {
				baseName = strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(name, ".exe"), ".cmd"), ".ps1")
			}

			if !knownShims[baseName] {
				shimPath := filepath.Join(shimsDir, name)
				if dryRun {
					pterm.Warning.Printfln("[dry-run] Would remove dead shim: %s", name)
					deadRemoved++
					continue
				}
				if err := os.Remove(shimPath); err != nil && !os.IsNotExist(err) {
					formatter.Warning(fmt.Sprintf("Failed to remove dead shim %s: %v", name, err), nil)
					continue
				}
				deadRemoved++
				if verbose {
					pterm.Info.Printfln("Removed dead shim: %s", name)
				}
			}
		}
	}

	if len(installations) == 0 {
		if deadRemoved > 0 {
			formatter.Success(fmt.Sprintf("Removed %d dead shim(s) — no installed tools to reshim", deadRemoved), nil)
		} else {
			formatter.Info("No installed tools found — nothing to reshim", nil)
		}
		return nil
	}

	if dryRun {
		for _, inst := range installations {
			pterm.Info.Printfln("[dry-run] Would regenerate shim for %s@%s", inst.Tool, inst.Version)
		}
		return nil
	}

	// ── Phase 2: Concurrent shim generation ───────────────────────────────────
	pterm.DefaultSection.Printfln("Regenerating %d shim(s)", len(installations))

	spinner, _ := pterm.DefaultSpinner.
		WithSequence("⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏").
		Start("Generating shims in parallel…")

	var (
		shimCount atomic.Int64
		failCount atomic.Int64
		mu        sync.Mutex
		warnings  []string
	)

	// Determine worker count: use --jobs flag or GOMAXPROCS
	workerCount := jobs
	if workerCount <= 0 {
		workerCount = runtime.GOMAXPROCS(0)
	}
	if workerCount > len(installations) {
		workerCount = len(installations)
	}

	type workItem struct {
		tool        string
		version     string
		backend     string
		installPath string
	}

	workCh := make(chan workItem, len(installations))
	for _, inst := range installations {
		workCh <- workItem{
			tool:        inst.Tool,
			version:     inst.Version,
			backend:     inst.Backend,
			installPath: inst.InstallPath,
		}
	}
	close(workCh)

	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range workCh {
				p := providerRegistry.GetWithBackend(item.tool, item.backend)
				executables, err := p.ListExecutables(item.tool, item.installPath, item.version)
				if err != nil {
					executables = []string{item.tool}
				}

				if err := generator.GenerateShim(ctx, item.tool, executables...); err != nil {
					mu.Lock()
					warnings = append(warnings, fmt.Sprintf("Failed to generate shim for %s@%s: %v", item.tool, item.version, err))
					mu.Unlock()
					failCount.Add(1)
					continue
				}
				shimCount.Add(1)

				if verbose {
					mu.Lock()
					spinner.UpdateText(fmt.Sprintf("Shimmed: %s@%s", item.tool, item.version))
					mu.Unlock()
				}
			}
		}()
	}

	wg.Wait()
	spinner.Stop()

	// Print any warnings
	for _, w := range warnings {
		pterm.Warning.Println(w)
	}

	// ── Summary ───────────────────────────────────────────────────────────────
	generated := shimCount.Load()
	failed := failCount.Load()

	parts := []string{fmt.Sprintf("✓  Reshimmed %d tool installation(s)", generated)}
	if reshimClean && deadRemoved > 0 {
		parts = append(parts, fmt.Sprintf("removed %d dead shim(s)", deadRemoved))
	}
	if failed > 0 {
		parts = append(parts, fmt.Sprintf("%d failed", failed))
	}
	parts = append(parts, fmt.Sprintf("using %d worker(s)", workerCount))
	parts = append(parts, fmt.Sprintf("in %s", shimsDir))

	formatter.Success(strings.Join(parts, "  ·  "), map[string]interface{}{
		"count":        generated,
		"dead_removed": deadRemoved,
		"failed":       failed,
		"workers":      workerCount,
		"shims_dir":    shimsDir,
	})

	return nil
}
