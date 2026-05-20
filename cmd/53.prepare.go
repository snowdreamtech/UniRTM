// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/service"
	"github.com/spf13/cobra"
)

func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(prepareCmd)
	}
}

// prepareCmd represents the prepare command which ensures project dependencies are ready.
var prepareCmd = &cobra.Command{
	Use:   "prepare [tool]",
	Short: "Ensure project dependencies are ready by running applicable prepare steps",
	Long: `Ensure project dependencies are ready by running applicable prepare steps.

This command:
1. Loads the project configuration (e.g., unirtm.toml).
2. Parallel downloads and installs any missing tools configured for this project.
3. Detects other package managers (like npm/pnpm, Go, Python) and prints preparation checks.

Examples:
  # Prepare everything for the current project
  unirtm prepare

  # Prepare only a specific tool
  unirtm prepare node`,
	Aliases: []string{"prep"},
	Args:    cobra.MaximumNArgs(1),
	RunE:    runPrepare,
}

// runPrepare executes the prepare command.
func runPrepare(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cfg, _ := loadConfig(ctx)
	if cfg != nil {
		cfg.ApplyEnvironment()
	}

	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	cwd, err := os.Getwd()
	if err != nil {
		formatter.Error("Failed to get current working directory", map[string]interface{}{"error": err.Error()})
		return err
	}

	targetTool := ""
	if len(args) > 0 {
		targetTool = strings.ToLower(strings.TrimSpace(args[0]))
	}

	pterm.DefaultSection.Println("Project Preparation Check")


	// 1. Run local project structural checks (Node, Go, Python)
	detectProjectStructure(cwd, targetTool)

	// 2. Load installation manager and check UniRTM tools
	im, err := getInstallationManager(ctx, cfg)
	if err != nil {
		formatter.Error("Failed to initialize installation manager", map[string]interface{}{"error": err.Error()})
		return err
	}

	if cfg == nil || len(cfg.Tools) == 0 {
		pterm.Info.Println("No UniRTM tools configured in project manifest.")
		pterm.Println()
		pterm.Success.Println("Project preparation check complete!")
		return nil
	}

	// Filter and build requests for missing tools
	var requests []service.ToolInstallRequest
	var alreadyInstalled []string

	backendRegistry := backend.NewRegistry()

	for name, tc := range cfg.Tools {
		backendName, toolName, version, _ := im.ParseToolSpec(name)
		if tc.Backend != "" {
			backendName = tc.Backend
		}
		if tc.Version != "" {
			version = tc.Version
		}

		// Filter by target tool if specified in args
		if targetTool != "" && strings.ToLower(toolName) != targetTool {
			continue
		}

		isInstalled, _ := im.IsInstalled(ctx, toolName, version, backendName)
		if isInstalled {
			alreadyInstalled = append(alreadyInstalled, fmt.Sprintf("%s@%s", toolName, version))
			continue
		}

		// Find dependencies for this tool
		var dependsOn []string
		if b, err := backendRegistry.Get(backendName); err == nil {
			for _, dep := range b.Dependencies() {
				if _, exists := cfg.Tools[dep]; exists {
					dependsOn = append(dependsOn, dep)
				}
			}
		}

		requests = append(requests, service.ToolInstallRequest{
			Tool:      toolName,
			Version:   version,
			Backend:   backendName,
			DependsOn: dependsOn,
		})
	}

	// Report already installed tools
	if len(alreadyInstalled) > 0 {
		pterm.Info.Printf("Configured tools already installed: %s\n", strings.Join(alreadyInstalled, ", "))
	}

	if len(requests) == 0 {
		pterm.Println()
		pterm.Success.Println("All configured UniRTM tools are already installed and ready!")
		return nil
	}

	pterm.Println()
	pterm.Info.Printf("Found %d tool(s) to install/prepare...\n", len(requests))
	pterm.Println()

	// 3. Execute parallel automatic download and installation
	concurrencyLimit := 8
	if cfg != nil && cfg.Settings.Jobs > 0 {
		concurrencyLimit = cfg.Settings.Jobs
	}

	progressFn := func(tool, version, status string) {
		if jsonOutput {
			return
		}
		switch status {
		case "starting":
			pterm.Info.Printf("Preparing %s@%s...\n", tool, version)
		case "done":
			pterm.FgGreen.Printf("✓ Ready: %s@%s\n", tool, version)
		default:
			if strings.HasPrefix(status, "failed:") {
				errMsg := strings.TrimPrefix(status, "failed: ")
				if errMsg == service.ErrAlreadyInstalled.Error() || strings.Contains(errMsg, "already installed") {
					pterm.FgGreen.Printf("✓ Ready: %s@%s (already installed)\n", tool, version)
				} else {
					pterm.Error.Printf("Failed to prepare %s@%s: %s\n", tool, version, errMsg)
				}
			} else {
				pterm.Info.Printf("%s@%s: %s\n", tool, version, status)
			}
		}
	}

	concurrentConfig := service.ConcurrentManagerConfig{
		MaxConcurrency: concurrencyLimit,
		ProgressFn:     progressFn,
	}
	cmManager := service.NewConcurrentManager(im, concurrentConfig)

	results, err := cmManager.InstallAll(ctx, requests)
	if err != nil {
		pterm.Error.Printf("Preparation failed: %v\n", err)
		return err
	}

	pterm.DefaultSection.Println("Preparation Summary")


	allSuccess := true
	for _, r := range results {
		if !r.Success && r.Error != service.ErrAlreadyInstalled.Error() && !strings.Contains(r.Error, "already installed") {
			allSuccess = false
			pterm.Error.Printf("  %s@%s: %s\n", r.Tool, r.Version, r.Error)
		} else {
			pterm.Success.Printf("  %s@%s is ready\n", r.Tool, r.Version)
		}
	}

	pterm.Println()
	if allSuccess {
		pterm.Success.Println("Project preparation complete! All dependencies are ready.")
	} else {
		pterm.Warning.Println("Some project dependencies failed to prepare. Please review the errors above.")
	}

	return nil
}

func detectProjectStructure(cwd string, targetTool string) {
	found := false

	// Check for Node.js
	if targetTool == "" || targetTool == "node" || targetTool == "nodejs" {
		if _, err := os.Stat(filepath.Join(cwd, "package.json")); err == nil {
			found = true
			pterm.Info.Println("Detected Node.js project")
			if _, err := os.Stat(filepath.Join(cwd, "node_modules")); os.IsNotExist(err) {
				pterm.Warning.Println("  node_modules missing. Suggestion: run 'npm install' or 'pnpm install'")
			} else {
				pterm.FgGreen.Println("  ✅ node_modules present")
			}
		}
	}

	// Check for Go
	if targetTool == "" || targetTool == "go" || targetTool == "golang" {
		if _, err := os.Stat(filepath.Join(cwd, "go.mod")); err == nil {
			found = true
			pterm.Info.Println("Detected Go project")
			pterm.FgGreen.Println("  ✅ go.mod present")
		}
	}

	// Check for Python
	if targetTool == "" || targetTool == "python" {
		if _, err := os.Stat(filepath.Join(cwd, "requirements.txt")); err == nil {
			found = true
			pterm.Info.Println("Detected Python project (requirements.txt)")
		}
		if _, err := os.Stat(filepath.Join(cwd, "pyproject.toml")); err == nil {
			found = true
			pterm.Info.Println("Detected Python project (pyproject.toml)")
		}
	}

	if found {
		pterm.Println()
	}
}

// getFSToolName returns the folder name for a tool based on backend.
func getFSToolName(toolName, backendName string) string {
	return env.GetFSToolName(toolName, backendName)
}
