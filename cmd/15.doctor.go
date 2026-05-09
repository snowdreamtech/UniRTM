// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"text/tabwriter"
	"time"

	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/spf13/cobra"
)

// init registers the doctor command to the root command.
func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(doctorCmd)
	}
}

// doctorCmd runs a comprehensive system health check.
var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system health and diagnose issues",
	Long: `Check UniRTM system health and diagnose potential issues.

The doctor command runs a series of checks:
  • Database accessibility and integrity
  • Cache directory writability
  • Data directory writability
  • Shims directory
  • Installed tools presence
  • Configuration file validity
  • Network connectivity to backends
  • PATH environment check

Examples:
  unirtm doctor
  unirtm doctor --json`,
	Args: cobra.NoArgs,
	RunE: runDoctor,
}

// checkResult represents the result of a single health check.
type checkResult struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // ok, warning, error
	Message string `json:"message"`
}

// runDoctor executes the doctor command.
//
// Validates: Requirements 24.1, 24.2, 24.3, 24.4, 24.5, 24.6, 24.7, 23.2
func runDoctor(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	ctx := context.Background()

	var checks []checkResult
	hasErrors := false

	addCheck := func(name, status, msg string) {
		checks = append(checks, checkResult{Name: name, Status: status, Message: msg})
		if status == "error" {
			hasErrors = true
		}
	}

	if !quiet && !jsonOutput {
		fmt.Printf("UniRTM Doctor — System Health Check\n")
		fmt.Printf("OS: %s/%s\n\n", runtime.GOOS, runtime.GOARCH)
	}

	// ── Check 1: Database ─────────────────────────────────────────────────────
	dbPath := env.GetDatabasePath()
	db, err := database.Open(ctx, database.Config{
		Path:    dbPath,
		WALMode: true,
	})
	if err != nil {
		addCheck("database", "error", fmt.Sprintf("Cannot open database at %s: %s", dbPath, err.Error()))
	} else {
		if pingErr := db.Ping(ctx); pingErr != nil {
			addCheck("database", "error", fmt.Sprintf("Database ping failed: %s", pingErr.Error()))
		} else {
			addCheck("database", "ok", fmt.Sprintf("Database accessible (%s)", dbPath))
		}
		db.Close()
	}

	// ── Check 2: Cache directory ──────────────────────────────────────────────
	cacheDir := env.GetCacheDir()
	testFile := filepath.Join(cacheDir, ".write_test")
	if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
		addCheck("cache_dir", "error", fmt.Sprintf("Cache directory not writable: %s — %s", cacheDir, err.Error()))
	} else {
		os.Remove(testFile)
		addCheck("cache_dir", "ok", fmt.Sprintf("Cache directory writable (%s)", cacheDir))
	}

	// ── Check 3: Data directory ───────────────────────────────────────────────
	dataDir := env.GetDataDir()
	testFile = filepath.Join(dataDir, ".write_test")
	if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
		addCheck("data_dir", "error", fmt.Sprintf("Data directory not writable: %s — %s", dataDir, err.Error()))
	} else {
		os.Remove(testFile)
		addCheck("data_dir", "ok", fmt.Sprintf("Data directory writable (%s)", dataDir))
	}

	// ── Check 4: Shims directory ──────────────────────────────────────────────
	shimsDir := env.GetShimsDir()
	if info, err := os.Stat(shimsDir); err != nil {
		addCheck("shims_dir", "warning",
			fmt.Sprintf("Shims directory not found (%s) — will be created on first install", shimsDir))
	} else if !info.IsDir() {
		addCheck("shims_dir", "error", fmt.Sprintf("Shims path is not a directory: %s", shimsDir))
	} else {
		shims, _ := os.ReadDir(shimsDir)
		addCheck("shims_dir", "ok", fmt.Sprintf("Shims directory OK (%d shims)", len(shims)))
	}

	// ── Check 5: Installed tools ──────────────────────────────────────────────
	if db2, err := database.Open(ctx, database.Config{Path: dbPath, WALMode: true}); err == nil {
		if installRepo, err := sqlite.NewInstallationRepository(db2.Conn()); err == nil {
			installations, err := installRepo.List(ctx)
			if err != nil {
				addCheck("installed_tools", "warning", fmt.Sprintf("Cannot list installations: %s", err.Error()))
			} else if len(installations) == 0 {
				addCheck("installed_tools", "ok", "No tools installed yet")
			} else {
				missingCount := 0
				for _, inst := range installations {
					if _, statErr := os.Stat(inst.InstallPath); os.IsNotExist(statErr) {
						missingCount++
					}
				}
				if missingCount > 0 {
					addCheck("installed_tools", "warning",
						fmt.Sprintf("%d/%d tools have missing install directories", missingCount, len(installations)))
				} else {
					addCheck("installed_tools", "ok",
						fmt.Sprintf("All %d installed tools are present", len(installations)))
				}
			}
		}
		db2.Close()
	}

	// ── Check 6: Configuration ────────────────────────────────────────────────
	cm := config.NewConfigManager()
	cfg, loadErr := cm.LoadHierarchy(ctx)
	if loadErr != nil {
		addCheck("config", "warning", "No configuration files found (optional — using defaults)")
	} else {
		if valErr := cm.Validate(ctx, cfg); valErr != nil {
			addCheck("config", "error", fmt.Sprintf("Configuration invalid: %s", valErr.Error()))
		} else {
			addCheck("config", "ok", fmt.Sprintf("Configuration valid (%d tools configured)", len(cfg.Tools)))
		}
	}

	// ── Check 7: Network connectivity ────────────────────────────────────────
	networkTargets := []struct {
		name string
		url  string
	}{
		{"github_api", "https://api.github.com"},
		{"aqua_registry", "https://aquaproj.github.io"},
	}

	httpClient := &http.Client{Timeout: 5 * time.Second}
	for _, nt := range networkTargets {
		resp, netErr := httpClient.Get(nt.url)
		if netErr != nil {
			addCheck("network_"+nt.name, "warning",
				fmt.Sprintf("Cannot reach %s: %s (offline mode may apply)", nt.url, netErr.Error()))
		} else {
			resp.Body.Close()
			if resp.StatusCode >= 500 {
				addCheck("network_"+nt.name, "warning",
					fmt.Sprintf("%s returned HTTP %d", nt.url, resp.StatusCode))
			} else {
				addCheck("network_"+nt.name, "ok",
					fmt.Sprintf("%s reachable (HTTP %d)", nt.url, resp.StatusCode))
			}
		}
	}

	// ── Check 8: Shims in PATH ───────────────────────────────────────────────
	pathEnv := os.Getenv("PATH")
	shimsInPath := false
	for _, p := range filepath.SplitList(pathEnv) {
		if p == shimsDir {
			shimsInPath = true
			break
		}
	}
	if !shimsInPath {
		addCheck("path_shims", "warning",
			fmt.Sprintf("Shims dir not in PATH — run: eval \"$(unirtm activate)\""))
	} else {
		addCheck("path_shims", "ok", "Shims directory is in PATH")
	}

	// ── Check 9: Optional — go binary (verbose only) ──────────────────────────
	if verbose {
		if goPath, err := exec.LookPath("go"); err != nil {
			addCheck("go_runtime", "warning", "Go runtime not found in PATH")
		} else {
			addCheck("go_runtime", "ok", fmt.Sprintf("Go found at %s", goPath))
		}
	}

	// ── Render output ────────────────────────────────────────────────────────
	if jsonOutput {
		overallStatus := "ok"
		errorCount := countByStatus(checks, "error")
		warningCount := countByStatus(checks, "warning")
		if errorCount > 0 {
			overallStatus = "error"
		} else if warningCount > 0 {
			overallStatus = "warning"
		}
		formatter.Success("Health check complete", map[string]interface{}{
			"overall_status": overallStatus,
			"checks":         checks,
			"check_count":    len(checks),
			"error_count":    errorCount,
			"warning_count":  warningCount,
		})
	} else {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		for _, c := range checks {
			icon := "✓"
			if c.Status == "error" {
				icon = "✗"
			} else if c.Status == "warning" {
				icon = "⚠"
			}
			fmt.Fprintf(w, "%s  %-20s\t%s\n", icon, c.Name, c.Message)
		}
		w.Flush()

		fmt.Println()
		errorCount := countByStatus(checks, "error")
		warningCount := countByStatus(checks, "warning")
		if errorCount == 0 && warningCount == 0 {
			fmt.Println("✓ All checks passed — UniRTM is healthy.")
		} else {
			fmt.Printf("Summary: %d error(s), %d warning(s).\n", errorCount, warningCount)
		}
	}

	if hasErrors {
		return fmt.Errorf("health check found %d error(s)", countByStatus(checks, "error"))
	}
	return nil
}

// countByStatus counts check results with the given status.
func countByStatus(checks []checkResult, status string) int {
	n := 0
	for _, c := range checks {
		if c.Status == status {
			n++
		}
	}
	return n
}
