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
	// installForce forces reinstallation of tools even if already installed
	installForce   bool
)

// init registers the install command to the root command.
func init() {
	// Register command flags
	installCmd.Flags().StringVarP(&installBackend, "backend", "b", "", "backend to use for installation (default: auto-detect)")
	installCmd.Flags().BoolVarP(&installForce, "force", "f", false, "force installation even if already installed")

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
	Use:   "install [tool[@version]...]",
	Short: "Install development tools and package specifications",
	Long: `Install development tools and package specifications.

The install command downloads and installs one or more specified versions of tools
using the appropriate backends. It validates the installation, records it in
the database, and generates shim scripts.

Examples:
  # Install Node.js version 20.0.0
  unirtm install node 20.0.0

  # Install Python version 3.11.5 using a specific backend
  unirtm install python 3.11.5 --backend github

  # Install multiple tools concurrently with package syntax
  unirtm install node@20.11.0 go@1.22.1 python@3.12.0

  # Force reinstall a tool even if already installed
  unirtm install node@20.11.0 --force

  # Install with JSON output
  unirtm install go 1.21.0 --json`,
	Aliases: []string{"i"},
	Args:    cobra.ArbitraryArgs,
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
		toolsToInstall = make(map[string]service.ToolSpec)

		// Determine if we should treat it as a single tool + version (backward compatibility)
		// e.g. "unirtm install node 20.0.0"
		isLegacySingleTool := false
		if len(args) == 2 {
			firstArg := args[0]
			secondArg := args[1]

			if !strings.Contains(firstArg, "@") && !strings.Contains(firstArg, ":") &&
				!strings.Contains(secondArg, "@") {
				isLegacySingleTool = true
			}
		}

		if isLegacySingleTool {
			backendName, tool, _, _ := im.ParseToolSpec(args[0])
			version := args[1]

			if tool == "" {
				formatter.Error("Tool name cannot be empty")
				return fmt.Errorf("tool name is required")
			}

			if version == "" {
				formatter.Error("Version cannot be empty")
				return fmt.Errorf("version is required")
			}

			if installBackend != "" {
				backendName = installBackend
			}

			toolsToInstall[tool] = service.ToolSpec{
				Name:        tool,
				Version:     version,
				BackendName: backendName,
			}
		} else {
			// Parse each argument as a separate tool spec
			for _, arg := range args {
				backendName, tool, version, explicitVersion := im.ParseToolSpec(arg)

				if installBackend != "" {
					backendName = installBackend
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
					version = "latest"
				}

				if !explicitVersion && !jsonOutput {
					// Check if stdin is a terminal
					if stat, _ := os.Stdin.Stat(); (stat.Mode() & os.ModeCharDevice) != 0 {
						selected, err := im.SelectVersionInteractive(ctx, tool, backendName)
						if err == nil {
							version = selected
						} else {
							pterm.Warning.Printf("Interactive selection failed for %s: %v, falling back to latest\n", tool, err)
						}
					}
				}

				key := tool
				if backendName != "" && backendName != "asdf" {
					key = backendName + ":" + tool
				}
				toolsToInstall[key] = service.ToolSpec{
					Name:        tool,
					Version:     version,
					BackendName: backendName,
				}
			}
		}
	}

	// Display start message
	if len(args) > 0 {
		formatter.Info(fmt.Sprintf("Installing %d tool(s)", len(toolsToInstall)), map[string]interface{}{
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

	// Sort tools by dependency to ensure runtimes are installed first
	sortedTools := installManager.SortToolsFromSpecs(toolsToInstall)

	// Pre-filter already installed tools to prevent them from corrupting the MultiPrinter UI
	var activeTools []service.ToolToInstall
	for _, t := range sortedTools {
		isInstalled := false
		if !installForce {
			isInstalled, _ = installManager.IsInstalled(ctx, t.ToolName, t.Version, t.BackendName)
		}

		if isInstalled {
			if !jsonOutput {
				pterm.Success.Printf("✓ %s@%s (already installed, use --force to reinstall)\n", t.ToolName, t.Version)
			}
		} else {
			if installForce {
				alreadyOnDisk, _ := installManager.IsInstalled(ctx, t.ToolName, t.Version, t.BackendName)
				if alreadyOnDisk {
					if !jsonOutput {
						pterm.Warning.Printf("⚠️  [force] Uninstalling existing %s@%s before reinstallation...\n", t.ToolName, t.Version)
					}
					_ = installManager.Uninstall(ctx, t.ToolName, t.Version)
				}
			}
			activeTools = append(activeTools, t)
		}
	}
	sortedTools = activeTools

	// Decide concurrency
	concurrencyLimit := jobs
	if !cmd.Flags().Changed("jobs") && cmd.Root() != nil && !cmd.Root().PersistentFlags().Changed("jobs") {
		if cfg != nil && cfg.Settings.Jobs > 0 {
			concurrencyLimit = cfg.Settings.Jobs
		}
	}
	if concurrencyLimit <= 0 {
		concurrencyLimit = runtime.NumCPU()
	}

	// We run concurrently if there are multiple tools and concurrency limit > 1
	runConcurrent := len(sortedTools) > 1 && concurrencyLimit > 1

	if runConcurrent {
		if !jsonOutput {
			pterm.Info.Printf("Installing %d tool(s) concurrently (max %d parallel jobs)...\n", len(sortedTools), concurrencyLimit)
		}

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
			spinnerMgr *concurrentSpinnerManager
			useMulti   = pterm.PrintColor && !pterm.RawOutput && !jsonOutput && term.IsTerminal(int(os.Stdout.Fd())) && len(requests) <= 10
		)

		if useMulti {
			spinnerMgr = newConcurrentSpinnerManager()
			spinnerMgr.Start()
			defer func() {
				if spinnerMgr != nil {
					spinnerMgr.Stop()
				}
			}()
		}

		progressFn := func(tool, version, status string) {
			if jsonOutput {
				return
			}
			if useMulti {
				switch status {
				case "starting":
					spinnerMgr.Add(tool, version)
				case "done":
					spinnerMgr.Complete(tool, version, "done")
				default:
					if strings.HasPrefix(status, "failed:") {
						spinnerMgr.Complete(tool, version, status)
					} else {
						spinnerMgr.Update(tool, status)
					}
				}
			} else {
				switch status {
				case "starting":
					pterm.Info.Printf("Starting installation of %s@%s...\n", tool, version)
				case "done":
					pterm.FgGreen.Printf("✓ Successfully installed %s@%s\n", tool, version)
				default:
					if strings.HasPrefix(status, "failed:") {
						errMsg := strings.TrimPrefix(status, "failed: ")
						if errMsg == service.ErrAlreadyInstalled.Error() || strings.Contains(errMsg, "already installed") {
							pterm.FgGreen.Printf("✓ %s@%s (already installed)\n", tool, version)
						} else {
							pterm.Error.Printf("Failed to install %s@%s: %s\n", tool, version, errMsg)
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
				if total > 0 {
					percent := (downloaded * 100) / total
					spinnerMgr.Update(toolName, fmt.Sprintf("Downloading %s/%s (%d%%)",
						humanize.Bytes(uint64(downloaded)),
						humanize.Bytes(uint64(total)),
						percent))
				} else {
					spinnerMgr.Update(toolName, fmt.Sprintf("Downloading %s",
						humanize.Bytes(uint64(downloaded))))
				}
			}
			ctx = context.WithValue(ctx, service.ContextKeyProgressReporter, service.ProgressReporter(reporter))
		} else {
			// Throttled progress reporter to prevent terminal flooding in non-multi/large-batch mode
			var (
				lastPercentMu sync.Mutex
				lastPercent   = make(map[string]int64)
			)
			reporter := func(toolName string, downloaded, total int64) {
				if total <= 0 {
					// Throttled by size: report every 10MB
					const tenMB = 10 * 1024 * 1024
					lastPercentMu.Lock()
					prevSize, exists := lastPercent[toolName]
					if !exists || downloaded-prevSize >= tenMB {
						lastPercent[toolName] = downloaded
						lastPercentMu.Unlock()
						pterm.Info.Printf("Downloading %s: %s\n",
							toolName,
							humanize.Bytes(uint64(downloaded)))
					} else {
						lastPercentMu.Unlock()
					}
					return
				}

				percent := (downloaded * 100) / total

				// Only report progress at 25%, 50%, 75%, and 90%+ intervals to be clean
				var shouldReport bool
				lastPercentMu.Lock()
				prev, exists := lastPercent[toolName]
				if !exists {
					shouldReport = true
					lastPercent[toolName] = percent
				} else if percent >= 100 && prev < 100 {
					shouldReport = true
					lastPercent[toolName] = 100
				} else {
					// Report if crossed a 25% threshold
					for _, threshold := range []int64{25, 50, 75, 90} {
						if percent >= threshold && prev < threshold {
							shouldReport = true
							lastPercent[toolName] = percent
							break
						}
					}
				}
				lastPercentMu.Unlock()

				if shouldReport {
					pterm.Info.Printf("Downloading %s: %s/%s (%d%%)\n",
						toolName,
						humanize.Bytes(uint64(downloaded)),
						humanize.Bytes(uint64(total)),
						percent)
				}
			}
			ctx = context.WithValue(ctx, service.ContextKeyProgressReporter, service.ProgressReporter(reporter))
		}

		// 3. Execute installation
		startTime := time.Now()
		results, err := cm.InstallAll(ctx, requests)

		if useMulti && spinnerMgr != nil {
			spinnerMgr.Stop()
		}

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
			pterm.Success.Printf("✓ All tools processed successfully (took %s)\n", duration.Round(time.Millisecond).String())
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
					pterm.Success.Printf("✓ %s@%s (already installed)\n", toolName, version)
					continue
				}

				pterm.Error.Printf("Installation failed for %s: %v\n", toolName, err)
				return fmt.Errorf("install %s: %w", toolName, err)
			}
			pterm.Success.Printf("✓ Successfully installed %s@%s\n", toolName, version)
		}

		duration := time.Since(startTime)
		if len(toolsToInstall) > 1 && !jsonOutput {
			pterm.Success.Printf("✓ All tools processed (took %s)\n", duration.Round(time.Millisecond).String())
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

type concurrentSpinnerFrame struct {
	tool    string
	version string
	status  string
}

type concurrentSpinnerManager struct {
	mu          sync.Mutex
	active      []*concurrentSpinnerFrame
	activeMap   map[string]*concurrentSpinnerFrame
	ticker      *time.Ticker
	done        chan struct{}
	frames      []string
	frameCount  int
	lastHeight  int
}

func newConcurrentSpinnerManager() *concurrentSpinnerManager {
	return &concurrentSpinnerManager{
		activeMap: make(map[string]*concurrentSpinnerFrame),
		done:      make(chan struct{}),
		frames:    []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	}
}

func (m *concurrentSpinnerManager) Start() {
	m.ticker = time.NewTicker(100 * time.Millisecond)
	go func() {
		for {
			select {
			case <-m.ticker.C:
				m.render()
			case <-m.done:
				return
			}
		}
	}()
}

func (m *concurrentSpinnerManager) Add(tool, version string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	frame := &concurrentSpinnerFrame{
		tool:    tool,
		version: version,
		status:  "starting",
	}
	m.active = append(m.active, frame)
	m.activeMap[tool] = frame
}

func (m *concurrentSpinnerManager) Update(tool, status string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if frame, exists := m.activeMap[tool]; exists {
		frame.status = status
	}
}

func (m *concurrentSpinnerManager) Complete(tool, version, status string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clear previous output first
	m.renderClear()

	// Print static status message based on status
	switch status {
	case "done":
		pterm.Success.Printf("Successfully installed %s@%s\n", tool, version)
	default:
		if strings.HasPrefix(status, "failed:") {
			errMsg := strings.TrimPrefix(status, "failed: ")
			if errMsg == service.ErrAlreadyInstalled.Error() || strings.Contains(errMsg, "already installed") {
				pterm.FgGreen.Printf("✓ %s@%s (already installed)\n", tool, version)
			} else {
				pterm.Error.Printf("Failed to install %s@%s: %s\n", tool, version, errMsg)
			}
		} else {
			pterm.Info.Printf("%s@%s: %s\n", tool, version, status)
		}
	}

	// Remove from active list
	var newActive []*concurrentSpinnerFrame
	for _, f := range m.active {
		if f.tool != tool {
			newActive = append(newActive, f)
		}
	}
	m.active = newActive
	delete(m.activeMap, tool)

	// Re-render remaining active spinners immediately
	m.renderWrite()
}

func (m *concurrentSpinnerManager) Stop() {
	if m.ticker != nil {
		m.ticker.Stop()
		// Safe channel closing check
		select {
		case <-m.done:
		default:
			close(m.done)
		}
	}
	m.mu.Lock()
	m.renderClear()
	m.mu.Unlock()
}

func (m *concurrentSpinnerManager) render() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.renderClear()
	m.renderWrite()
}

func (m *concurrentSpinnerManager) renderClear() {
	if m.lastHeight > 0 {
		for i := 0; i < m.lastHeight; i++ {
			fmt.Print("\033[A\033[K")
		}
		m.lastHeight = 0
	}
}

func (m *concurrentSpinnerManager) renderWrite() {
	if len(m.active) == 0 {
		return
	}

	spinnerChar := m.frames[m.frameCount%len(m.frames)]
	m.frameCount++

	for _, f := range m.active {
		pterm.Printf("%s  %s: %s\n", pterm.FgCyan.Sprint(spinnerChar), pterm.FgLightBlue.Sprint(f.tool), f.status)
		m.lastHeight++
	}
}
