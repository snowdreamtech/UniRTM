// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/download"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/provider"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/snowdreamtech/unirtm/internal/service"
	"github.com/snowdreamtech/unirtm/internal/transaction"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	// installBackend specifies the backend to use for installation
	installBackend string
)

// init registers the install command to the root command.
func init() {
	// Register command flags
	installCmd.Flags().StringVarP(&installBackend, "backend", "b", "", "backend to use for installation (default: auto-detect)")

	// Add command to root - this will be called after rootCmd is initialized
	if rootCmd != nil {
		rootCmd.AddCommand(installCmd)
	}
}

// RegisterInstallCommand registers the install command to the root command.
// This function should be called after rootCmd is initialized.
func RegisterInstallCommand() {
	if rootCmd != nil {
		rootCmd.AddCommand(installCmd)
	}
}

// installCmd represents the install command which installs a specific version of a tool.
var installCmd = &cobra.Command{
	Use:   "install <tool> <version>",
	Short: "Install a specific version of a development tool",
	Long: `Install a specific version of a development tool.

The install command downloads and installs the specified version of a tool
using the appropriate backend. It validates the installation, records it in
the database, and generates shim scripts.

Examples:
  # Install Node.js version 20.0.0
  unirtm install node 20.0.0

  # Install Python version 3.11.5 using a specific backend
  unirtm install python 3.11.5 --backend github

  # Install with JSON output
  unirtm install go 1.21.0 --json`,
	Aliases: []string{"i"},
	Args:    cobra.MaximumNArgs(2),
	RunE:    runInstall,
}

// runInstall executes the install command.
// It validates input, delegates to the Installation Manager, and displays progress and results.
//
// Validates: Requirements 9.1, 9.2, 9.3, 23.2
func runInstall(cmd *cobra.Command, args []string) error {
	// Initialize dependencies
	ctx := context.Background()

	// Load project configuration
	cfg, err := config.Load()
	if err != nil {
		// Log warning and continue
	}

	// Create output formatter
	formatter := getFormatter(cfg)

	if cfg != nil {
		// Apply [env] variables from config to current process
		cfg.ApplyEnvironment()
	}

	// Get installation manager early for parsing and selection
	im, err := getInstallationManager(ctx, cfg)
	if err != nil {
		formatter.Error("Failed to initialize installation manager", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("get installation manager: %w", err)
	}

	var toolsToInstall map[string]service.ToolSpec
	if len(args) == 0 {
		// Use tools from config if no arguments provided
		if cfg == nil || len(cfg.Tools) == 0 {
			formatter.Warning("No tools found in project configuration.")
			return nil
		}
		toolsToInstall = make(map[string]service.ToolSpec, len(cfg.Tools))
		for name, tc := range cfg.Tools {
			backendName, toolName, version, _ := im.ParseToolSpec(name)
			if tc.Backend != "" {
				backendName = tc.Backend
			}
			if tc.Version != "" {
				version = tc.Version
			}

			toolsToInstall[name] = service.ToolSpec{
				Name:        toolName,
				Version:     version,
				BackendName: backendName,
			}
		}
	} else {
		// Use centralized ToolSpec parsing
		backendName, tool, version, explicitVersion := im.ParseToolSpec(args[0])

		// Override with explicit version if provided as second argument
		if len(args) == 2 {
			version = args[1]
			explicitVersion = true
		}

		// If version is not explicit, try to resolve from configuration file first
		if !explicitVersion && cfg != nil {
			if tc, ok := cfg.Tools[tool]; ok && tc.Version != "" {
				version = tc.Version
				explicitVersion = true
				if tc.Backend != "" {
					backendName = tc.Backend
				}
			}
		}

		if tool == "" {
			formatter.Error("Tool name cannot be empty")
			return fmt.Errorf("tool name is required")
		}
		if version == "" {
			formatter.Error("Version cannot be empty")
			return fmt.Errorf("version is required")
		}

		if !explicitVersion && !jsonOutput {
			// Check if stdin is a terminal
			if stat, _ := os.Stdin.Stat(); (stat.Mode() & os.ModeCharDevice) != 0 {
				selected, err := im.SelectVersionInteractive(ctx, tool, backendName)
				if err == nil {
					version = selected
				} else {
					pterm.Warning.Printf("Interactive selection failed: %v, falling back to latest\n", err)
				}
			}
		}

		toolsToInstall = map[string]service.ToolSpec{
			tool: {
				Name:        tool,
				Version:     version,
				BackendName: backendName,
			},
		}
	}

	// Display start message
	if len(args) > 0 {
		tool := args[0] // original arg for display
		formatter.Info(fmt.Sprintf("Installing %s", tool), map[string]interface{}{
			"args": args,
		})
	} else {
		formatter.Info("Installing all tools from configuration", map[string]interface{}{
			"count": len(toolsToInstall),
		})
	}

	// Dry-run: show intent without side effects
	if dryRun {
		formatter.Info("[dry-run] Would install tools — no changes made", map[string]interface{}{
			"tools":   toolsToInstall,
			"dry_run": true,
		})
		return nil
	}

	// Create backend registry
	backendRegistry := backend.NewRegistry()

	// Create provider registry
	providerRegistry := provider.NewRegistry()

	// Create download manager
	downloadManager := download.NewManager()
	downloadManager.Register("https", download.NewHTTPDownloader())
	downloadManager.Register("http", download.NewHTTPDownloader())

	// Create database connection
	dbPath := env.GetDatabasePath()
	db, err := database.Open(ctx, database.Config{
		Path:    dbPath,
		WALMode: true,
	})
	if err != nil {
		formatter.Error("Failed to initialize database", map[string]interface{}{
			"error": err.Error(),
			"path":  dbPath,
		})
		return fmt.Errorf("initialize database: %w", err)
	}
	defer db.Close()

	// Create repositories
	installRepo, err := sqlite.NewInstallationRepository(db.Conn())
	if err != nil {
		formatter.Error("Failed to create installation repository", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("create installation repository: %w", err)
	}

	// Create transaction manager
	txManager := transaction.NewSQLiteTransactionManager(db.Conn())

	// Create lock service if lockfile exists
	var lockSvc *service.LockService
	lockPath := env.GetLockFilePath()
	if _, err := os.Stat(lockPath); err == nil {
		lockSvc, _ = service.NewLockService(service.LockServiceOptions{
			LockfilePath: lockPath,
		})
		if lockSvc != nil {
			lockSvc.SetBackendRegistry(backendRegistry)
		}
	}

	// Create installation manager with optional lock support
	var settings *config.Settings
	if cfg != nil {
		settings = &cfg.Settings
	}
	installManager := service.NewInstallationManagerWithLock(
		backendRegistry,
		providerRegistry,
		downloadManager,
		installRepo,
		txManager,
		lockSvc,
		settings,
	)

	formatter.Info(fmt.Sprintf("Processing %d tool(s)...", len(toolsToInstall)))

	// Sort tools by dependency to ensure runtimes are installed first
	sortedTools := installManager.SortToolsFromSpecs(toolsToInstall)

	// Pre-filter already installed tools to prevent them from corrupting the MultiPrinter UI
	var activeTools []service.ToolToInstall
	for _, t := range sortedTools {
		isInstalled, _ := installManager.IsInstalled(ctx, t.ToolName, t.Version, t.BackendName)
		if isInstalled {
			if !jsonOutput {
				pterm.FgGreen.Printf("✓ %s@%s (already installed)\n", t.OriginalName, t.Version)
			}
		} else {
			activeTools = append(activeTools, t)
		}
	}
	sortedTools = activeTools

	// Decide concurrency
	concurrencyLimit := jobs
	if concurrencyLimit <= 0 {
		if cfg != nil && cfg.Settings.Concurrency > 0 {
			concurrencyLimit = cfg.Settings.Concurrency
		} else {
			concurrencyLimit = runtime.NumCPU()
		}
	}

	// We run concurrently if there are multiple tools and concurrency limit > 1
	runConcurrent := len(sortedTools) > 1 && concurrencyLimit > 1

	if runConcurrent {
		formatter.Info(fmt.Sprintf("Installing %d tool(s) concurrently (max %d parallel jobs)...", len(sortedTools), concurrencyLimit))

		// 1. Build tool list for ConcurrentManager
		var requests []service.ToolInstallRequest
		
		// Create a map of toolName -> ToolSpec to check for dependencies within the list
		specsMap := make(map[string]service.ToolSpec)
		for _, spec := range toolsToInstall {
			specsMap[spec.Name] = spec
		}

		for _, t := range sortedTools {
			// Find dependencies for this tool
			var dependsOn []string
			if b, err := backendRegistry.Get(t.BackendName); err == nil {
				for _, dep := range b.Dependencies() {
					// If the dependency is also scheduled to be installed, record it
					if _, exists := specsMap[dep]; exists {
						dependsOn = append(dependsOn, dep)
					}
				}
			}

			requests = append(requests, service.ToolInstallRequest{
				Tool:      t.ToolName,
				Version:   t.Version,
				Backend:   t.BackendName,
				DependsOn: dependsOn,
			})
		}

		// 2. Create ConcurrentManager
		var (
			spinners         = make(map[string]*pterm.SpinnerPrinter)
			spinnersMu       sync.Mutex
			multi            *pterm.MultiPrinter
			useMulti         = pterm.PrintColor && !pterm.RawOutput && !jsonOutput && term.IsTerminal(int(os.Stdout.Fd()))
			bufferedMessages []string
			bufferedMu       sync.Mutex
		)

		if useMulti {
			multi = &pterm.DefaultMultiPrinter
			go func() {
				_, _ = multi.Start()
			}()
			defer func() {
				time.Sleep(50 * time.Millisecond)
				_, _ = multi.Stop()
				bufferedMu.Lock()
				for _, msg := range bufferedMessages {
					pterm.Println(msg)
				}
				bufferedMu.Unlock()
			}()
		}

		progressFn := func(tool, version, status string) {
			if jsonOutput {
				return
			}
			switch status {
			case "starting":
				if useMulti {
					spinnersMu.Lock()
					spinner, _ := pterm.DefaultSpinner.
						WithWriter(multi.NewWriter()).
						Start(fmt.Sprintf("Installing %s@%s...", tool, version))
					spinner.SuccessPrinter = pterm.Success.WithWriter(spinner.Writer)
					spinner.FailPrinter = pterm.Error.WithWriter(spinner.Writer)
					spinners[tool] = spinner
					spinnersMu.Unlock()
				} else {
					pterm.Info.Printf("Starting installation of %s@%s...\n", tool, version)
				}
			case "done":
				if useMulti {
					spinnersMu.Lock()
					spinner, exists := spinners[tool]
					spinnersMu.Unlock()
					if exists && spinner != nil {
						spinner.Success(fmt.Sprintf("Successfully installed %s@%s", tool, version))
					}
				} else {
					pterm.FgGreen.Printf("✓ Successfully installed %s@%s\n", tool, version)
				}
			default:
				if strings.HasPrefix(status, "failed:") {
					errMsg := strings.TrimPrefix(status, "failed: ")
					if errMsg == service.ErrAlreadyInstalled.Error() || strings.Contains(errMsg, "already installed") {
						if useMulti {
							spinnersMu.Lock()
							spinner, exists := spinners[tool]
							spinnersMu.Unlock()
							if exists && spinner != nil {
								spinner.Success(fmt.Sprintf("%s@%s (already installed)", tool, version))
							} else {
								bufferedMu.Lock()
								bufferedMessages = append(bufferedMessages, pterm.FgGreen.Sprintf("✓ %s@%s (already installed)", tool, version))
								bufferedMu.Unlock()
							}
						} else {
							pterm.FgGreen.Printf("✓ %s@%s (already installed)\n", tool, version)
						}
					} else {
						if useMulti {
							spinnersMu.Lock()
							spinner, exists := spinners[tool]
							spinnersMu.Unlock()
							if exists && spinner != nil {
								spinner.Fail(fmt.Sprintf("Failed to install %s@%s: %s", tool, version, errMsg))
							} else {
								bufferedMu.Lock()
								bufferedMessages = append(bufferedMessages, pterm.Error.Sprintf("Failed to install %s@%s: %s", tool, version, errMsg))
								bufferedMu.Unlock()
							}
						} else {
							pterm.Error.Printf("Failed to install %s@%s: %s\n", tool, version, errMsg)
						}
					}
				} else {
					if useMulti {
						spinnersMu.Lock()
						spinner, exists := spinners[tool]
						spinnersMu.Unlock()
						if exists && spinner != nil {
							spinner.UpdateText(fmt.Sprintf("%s@%s: %s", tool, version, status))
						}
					} else {
						pterm.Info.Printf("%s@%s: %s\n", tool, version, status)
					}
				}
			}
		}

		concurrentConfig := service.ConcurrentManagerConfig{
			MaxConcurrency: concurrencyLimit,
			ProgressFn:     progressFn,
		}
		cm := service.NewConcurrentManager(installManager, concurrentConfig)

		// Set up ProgressReporter for live concurrent download updates
		if useMulti {
			reporter := func(toolName string, downloaded, total int64) {
				spinnersMu.Lock()
				spinner, exists := spinners[toolName]
				spinnersMu.Unlock()
				if exists && spinner != nil {
					if total > 0 {
						percent := (downloaded * 100) / total
						spinner.UpdateText(fmt.Sprintf("Downloading %s: %s/%s (%d%%)", 
							toolName, 
							humanize.Bytes(uint64(downloaded)), 
							humanize.Bytes(uint64(total)), 
							percent))
					} else {
						spinner.UpdateText(fmt.Sprintf("Downloading %s: %s", 
							toolName, 
							humanize.Bytes(uint64(downloaded))))
					}
				}
			}
			ctx = context.WithValue(ctx, service.ContextKeyProgressReporter, service.ProgressReporter(reporter))
		}

		// 3. Execute installation
		startTime := time.Now()
		results, err := cm.InstallAll(ctx, requests)
		if err != nil {
			formatter.Error("Concurrent installation failed", map[string]interface{}{
				"error": err.Error(),
			})
			return fmt.Errorf("concurrent install: %w", err)
		}

		// Check for failures
		var failedTools []string
		for _, r := range results {
			if !r.Success {
				// "already installed" is not a failure
				if r.Error != service.ErrAlreadyInstalled.Error() && !strings.Contains(r.Error, "already installed") {
					failedTools = append(failedTools, fmt.Sprintf("%s@%s (%s)", r.Tool, r.Version, r.Error))
				}
			}
		}

		duration := time.Since(startTime)
		if len(failedTools) > 0 {
			pterm.Error.Printf("Installation finished with errors (took %s):\n", duration.Round(time.Millisecond).String())
			for _, f := range failedTools {
				pterm.FgRed.Printf("  • %s\n", f)
			}
			return fmt.Errorf("some installations failed: %s", strings.Join(failedTools, ", "))
		}

		if !jsonOutput {
			pterm.FgGreen.Printf("✓ All tools processed successfully (took %s)\n", duration.Round(time.Millisecond).String())
		} else {
			// If JSON output, render the results JSON
			outputData, _ := json.MarshalIndent(results, "", "  ")
			fmt.Println(string(outputData))
		}
	} else {
		// Sequential installation (better for single tool interactive output)
		startTime := time.Now()
		for _, t := range sortedTools {
			toolName := t.OriginalName
			tool := t.ToolName
			version := t.Version
			backendName := t.BackendName

			err = installManager.Install(ctx, tool, version, backendName)
			if err != nil {
				// Check if already installed
				if err == service.ErrAlreadyInstalled || strings.Contains(err.Error(), "already installed") {
					pterm.FgGreen.Printf("✓ %s@%s (already installed)\n", toolName, version)
					continue
				}

				pterm.Error.Printf("Installation failed for %s: %v\n", toolName, err)
				return fmt.Errorf("install %s: %w", toolName, err)
			}
			pterm.FgGreen.Printf("✓ Successfully installed %s@%s\n", toolName, version)
		}

		duration := time.Since(startTime)
		if len(toolsToInstall) > 1 {
			pterm.FgGreen.Printf("✓ All tools processed (took %s)\n", duration.Round(time.Millisecond).String())
		}
	}

	return nil
}

// getOutputFormat returns the output format based on the jsonOutput flag.
func getOutputFormat() output.OutputFormat {
	if jsonOutput {
		return output.FormatJSON
	}
	return output.FormatHuman
}

func getBackendName() string {
	if installBackend != "" {
		return installBackend
	}
	return "" // Default to empty to allow auto-detection
}
