// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/download"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/provider"
	"github.com/snowdreamtech/unirtm/internal/provider/native"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/snowdreamtech/unirtm/internal/service"
	"github.com/snowdreamtech/unirtm/internal/transaction"
	"github.com/spf13/cobra"
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

	// Create output formatter
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	// Load project configuration
	cfg, err := config.Load()
	if err != nil {
		formatter.Warning(fmt.Sprintf("Failed to load project config: %v", err))
	} else {
		// Apply [env] variables from config to current process
		cfg.ApplyEnvironment()
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
			backendName := tc.Backend
			toolName := name
			if backendName == "" {
				if idx := strings.Index(name, ":"); idx != -1 {
					backendName = name[:idx]
					toolName = name[idx+1:]
				} else if strings.Contains(name, "/") {
					backendName = "github"
				} else if native.IsNativeTool(toolName) {
					backendName = "native"
				} else {
					backendName = "asdf"
				}
			}

			// Intercept go: prefix and route to the internal go-pkg provider
			if backendName == "go" {
				backendName = "go-pkg"
			}
			toolsToInstall[name] = service.ToolSpec{
				Name:        toolName,
				Version:     tc.Version,
				BackendName: backendName,
			}
		}
	} else {
		// Install specific tool from arguments
		tool := args[0]
		version := "latest"

		// Parse "tool@version" syntax (like mise)
		if strings.Contains(tool, "@") {
			parts := strings.SplitN(tool, "@", 2)
			tool = parts[0]
			if parts[1] != "" {
				version = parts[1]
			}
		} else if len(args) == 2 {
			version = args[1]
		}

		backendName := getBackendName()
		if strings.Contains(tool, ":") {
			parts := strings.SplitN(tool, ":", 2)
			backendName = parts[0]
			tool = parts[1]

			// Intercept go: prefix and route to the internal go-pkg provider
			if backendName == "go" {
				backendName = "go-pkg"
			}
		}

		if backendName == "" {
			toolName := tool
			if strings.Contains(tool, "@") {
				toolName = strings.SplitN(tool, "@", 2)[0]
			}

			if strings.Contains(toolName, "/") {
				backendName = "github"
			} else if native.IsNativeTool(toolName) {
				backendName = "native"
			} else {
				backendName = "asdf"
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

		toolsToInstall = map[string]service.ToolSpec{
			tool: {
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

	// Extract and sort tool names to ensure deterministic installation order
	toolNames := make([]string, 0, len(toolsToInstall))
	for name := range toolsToInstall {
		toolNames = append(toolNames, name)
	}
	sort.Strings(toolNames)

	// Perform installations
	startTime := time.Now()
	for _, toolName := range toolNames {
		spec := toolsToInstall[toolName]
		tool := toolName
		if spec.Name != "" {
			tool = spec.Name
		}
		
		formatter.Info(fmt.Sprintf("Processing %s@%s...", toolName, spec.Version))
		
		err = installManager.Install(ctx, tool, spec.Version, spec.BackendName)
		if err != nil {
			// Check if already installed
			if strings.Contains(err.Error(), "already installed") {
				formatter.Info(fmt.Sprintf("%s@%s is already installed", toolName, spec.Version))
				continue
			}
			
			pterm.Error.Printf("Installation failed for %s: %v\n", toolName, err)
			formatter.Error(fmt.Sprintf("Installation failed for %s: %s", toolName, err.Error()))
			return fmt.Errorf("install %s: %w", toolName, err)
		}
		pterm.Success.Printf("Successfully installed %s@%s\n", toolName, spec.Version)
	}
	
	duration := time.Since(startTime)
	if len(toolsToInstall) > 1 {
		pterm.Success.Printf("All tools processed (took %s)\n", duration.Round(time.Millisecond).String())
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

