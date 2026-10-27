// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/server"
	"github.com/spf13/cobra"
)

var (
	uiPort    int
	uiHost    string
	uiNoOpen  bool
	uiBrowser string
)

func init() {
	uiCmd.Flags().IntVarP(&uiPort, "port", "p", 5173, "port to listen on")
	uiCmd.Flags().StringVarP(&uiHost, "host", "H", "127.0.0.1", "host address to bind")
	uiCmd.Flags().BoolVar(&uiNoOpen, "no-open", false, "do not open browser automatically")
	uiCmd.Flags().StringVarP(&uiBrowser, "browser", "b", "", "browser to open (e.g. chrome, firefox, edge, or executable path)")

	if rootCmd != nil {
		rootCmd.AddCommand(uiCmd)
	}

	server.CheckSelfUpdateFn = func() (*server.SelfUpdateResult, error) {
		res, err := CheckSelfUpdate()
		if err != nil {
			return nil, err
		}
		return &server.SelfUpdateResult{
			CurrentVersion: res.CurrentVersion,
			LatestVersion:  res.LatestVersion,
			HasUpdate:      res.HasUpdate,
			ReleaseName:    res.ReleaseName,
			ReleaseNotes:   res.ReleaseNotes,
			InstallMethod:  res.InstallMethod,
			UpgradeHint:    res.UpgradeHint,
			IsUnsupported:  res.IsUnsupported,
		}, nil
	}
	server.ExecuteSelfUpdateFn = func(version string) (string, error) {
		return ExecuteSelfUpdate(version)
	}
}

// uiCmd starts the embedded interactive graphical workbench.
var uiCmd = &cobra.Command{
	Use:     "ui [path]",
	Aliases: []string{"dashboard", "web"},
	Short:   "Launch the interactive graphical workbench",
	Long: `Launch the interactive graphical workbench in your default browser.

Starts an embedded local web server serving the UniRTM visual workspace to
manage tools, run tasks, inspect doctor health, and manage extensions.

Examples:
  # Start UI and open browser for current directory
  unirtm ui

  # Open a specific project directory
  unirtm ui ./my-project

  # Open a parent directory containing multiple git repositories (multi-root workspace)
  unirtm ui /path/to/workspace

  # Start UI on a custom port without opening browser
  unirtm ui --port 8080 --no-open

  # Use alias
  unirtm dashboard
  unirtm web`,
	Args: cobra.MaximumNArgs(1),
	RunE: runUI,
}

func runUI(cmd *cobra.Command, args []string) error {
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		targetDir := strings.TrimSpace(args[0])
		absDir, err := filepath.Abs(targetDir)
		if err != nil {
			return fmt.Errorf("failed to resolve directory path %q: %w", targetDir, err)
		}
		st, err := os.Stat(absDir)
		if err != nil {
			return fmt.Errorf("directory not found: %s", targetDir)
		}
		if !st.IsDir() {
			return fmt.Errorf("path is not a directory: %s", targetDir)
		}
		if err := os.Chdir(absDir); err != nil {
			return fmt.Errorf("failed to change working directory to %s: %w", absDir, err)
		}
	}

	targetBrowser := strings.TrimSpace(uiBrowser)
	if targetBrowser == "" {
		if envBrowser := os.Getenv("BROWSER"); envBrowser != "" {
			targetBrowser = strings.TrimSpace(envBrowser)
		} else if globalCfg, err := config.LoadGlobal(); err == nil && globalCfg != nil && globalCfg.Settings.Browser != "" {
			targetBrowser = strings.TrimSpace(globalCfg.Settings.Browser)
		}
	}

	if localCfg, err := loadConfig(cmd.Context()); err == nil && localCfg != nil && localCfg.Settings.Browser != "" {
		pterm.Warning.Println("Notice: 'settings.browser' in project configuration is ignored (browser preferences are user-global only).")
	}

	opts := server.Options{
		Host:    uiHost,
		Port:    uiPort,
		NoOpen:  uiNoOpen,
		Browser: targetBrowser,
	}

	return server.Run(cmd.Context(), opts)
}
