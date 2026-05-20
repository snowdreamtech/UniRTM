// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/spf13/cobra"
)

func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(toolCmd)
	}

	toolCmd.Flags().BoolVar(&toolActive, "active", false, "Only show active versions")
	toolCmd.Flags().BoolVar(&toolBackend, "backend", false, "Only show backend field")
	toolCmd.Flags().BoolVar(&toolConfigSource, "config-source", false, "Only show config source")
	toolCmd.Flags().BoolVar(&toolDescription, "description", false, "Only show description field")
	toolCmd.Flags().BoolVar(&toolInstalled, "installed", false, "Only show installed versions")
	toolCmd.Flags().BoolVar(&toolRequested, "requested", false, "Only show requested versions")
	toolCmd.Flags().BoolVar(&toolToolOptions, "tool-options", false, "Only show tool options")
}

var (
	toolActive       bool
	toolBackend      bool
	toolConfigSource bool
	toolDescription  bool
	toolInstalled    bool
	toolRequested    bool
	toolToolOptions  bool
)

// toolCmd displays detailed information about a specific tool.
var toolCmd = &cobra.Command{
	Use:   "tool <tool>",
	Short: "Show information about a specific tool",
	Long: `Show detailed information about a specific tool.

Displays backend, installed versions, active version, shim location,
and the config file that sets the active version.

Examples:
  # Show info for GitHub CLI
  unirtm tool cli/cli

  # JSON output
  unirtm tool cli/cli --json

  # Specify backend explicitly
  unirtm tool github:cli/cli`,
	Args: cobra.ExactArgs(1),
	RunE: runTool,
}

// toolInfo holds the aggregated information about a single tool.
type toolInfo struct {
	Tool         string                 `json:"-"`
	Backend      string                 `json:"backend"`
	Description  string                 `json:"description,omitempty"`
	Installed    []string               `json:"installed_versions"`
	Requested    []string               `json:"requested_versions,omitempty"`
	Active       []string               `json:"active_versions,omitempty"`
	ConfigSource string                 `json:"config_source,omitempty"`
	ToolOptions  map[string]interface{} `json:"tool_options,omitempty"`
	Security     []string               `json:"security"`
}

func runTool(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	rawArg := args[0]

	// Parse optional "backend:tool" prefix.
	toolName := rawArg
	backendName := ""
	if idx := strings.Index(rawArg, ":"); idx != -1 {
		backendName = rawArg[:idx]
		toolName = rawArg[idx+1:]
	}

	ctx := context.Background()

	// Open database.
	dbPath := env.GetDatabasePath()
	db, err := database.Open(ctx, database.Config{Path: dbPath, WALMode: true})
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to open database: %v", err))
		return err
	}
	defer db.Close()

	installRepo, err := sqlite.NewInstallationRepository(db.Conn())
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to create repository: %v", err))
		return err
	}

	// Load all installations for this tool.
	all, err := installRepo.List(ctx)
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to list installations: %v", err))
		return err
	}

	var versions []string
	detectedBackend := backendName
	installDir := ""

	for _, inst := range all {
		if inst == nil {
			continue
		}
		if inst.Tool != toolName {
			continue
		}
		versions = append(versions, inst.Version)
		if detectedBackend == "" {
			detectedBackend = inst.Backend
		}
		if installDir == "" {
			installDir = filepath.Dir(inst.InstallPath)
		}
	}

	if len(versions) == 0 && detectedBackend == "" {
		// Try to infer backend from the tool name format.
		if strings.Contains(toolName, "/") {
			detectedBackend = "github"
		} else {
			detectedBackend = "asdf"
		}
	}

	// Check if backend supports the tool.
	backendRegistry := backend.NewRegistry()
	if detectedBackend == "" {
		detectedBackend = getBackendName()
	}

	// Load Config to get requested versions, tool options
	cfgMgr := config.NewConfigManager()
	cfg, _ := cfgMgr.LoadHierarchy(ctx)

	var requestedVersions []string
	var configSource = "Merged Hierarchy Config"
	var toolOpts map[string]interface{}
	
	if cfg != nil && cfg.Tools != nil {
		if tc, ok := cfg.Tools[toolName]; ok {
			if tc.Version != "" {
				requestedVersions = append(requestedVersions, tc.Version)
			}
			
			// Try to build tool options
			toolOpts = make(map[string]interface{})
			if tc.Backend != "" {
				toolOpts["backend"] = tc.Backend
			}
			if tc.Provider != "" {
				toolOpts["provider"] = tc.Provider
			}
			if len(tc.GPGKeys) > 0 {
				toolOpts["gpg_keys"] = tc.GPGKeys
			}
		}
	}

	// Active versions handling
	var activeVersions []string
	
	// Detect active version via shim.
	shimsDir := env.GetShimsDir()
	installsDir := env.GetInstallsDir()
	activeVersion := detectActiveVersion(shimsDir, installsDir, toolName, versions)
	
	if activeVersion != "" {
		activeVersions = append(activeVersions, activeVersion)
	}

	info := toolInfo{
		Tool:         toolName,
		Backend:      detectedBackend,
		Installed:    versions,
		Requested:    requestedVersions,
		Active:       activeVersions,
		ConfigSource: configSource,
		ToolOptions:  toolOpts,
		Security:     []string{},
	}

	// Get backend features
	if b, err := backendRegistry.Get(info.Backend); err == nil {
		if b.SupportsChecksum() {
			info.Security = append(info.Security, "checksum")
		}
		if b.SupportsGPG() {
			info.Security = append(info.Security, "gpg")
		}
		if att := b.AttestationType(); att != "" {
			info.Security = append(info.Security, strings.ToLower(att))
		}
	}

	// JSON Output Handling with Filters
	if jsonOutput {
		if toolBackend {
			outputJSON(info.Backend)
		} else if toolDescription {
			outputJSON(info.Description)
		} else if toolInstalled {
			outputJSON(info.Installed)
		} else if toolActive {
			outputJSON(info.Active)
		} else if toolRequested {
			outputJSON(info.Requested)
		} else if toolConfigSource {
			outputJSON(info.ConfigSource)
		} else if toolToolOptions {
			outputJSON(info.ToolOptions)
		} else {
			outputJSON(info)
		}
		return nil
	}

	// Human-readable filtered output
	if toolBackend {
		fmt.Println(info.Backend)
		return nil
	} else if toolDescription {
		if info.Description != "" {
			fmt.Println(info.Description)
		} else {
			fmt.Println("[none]")
		}
		return nil
	} else if toolInstalled {
		fmt.Println(formatInstalledWithActive(info.Installed, info.Active))
		return nil
	} else if toolActive {
		if len(info.Active) > 0 {
			fmt.Println(strings.Join(info.Active, " "))
		} else {
			fmt.Println("[none]")
		}
		return nil
	} else if toolRequested {
		if len(info.Requested) > 0 {
			fmt.Println(strings.Join(info.Requested, " "))
		} else {
			fmt.Println("[none]")
		}
		return nil
	} else if toolConfigSource {
		if info.ConfigSource != "" {
			fmt.Println(info.ConfigSource)
		} else {
			fmt.Println("[none]")
		}
		return nil
	} else if toolToolOptions {
		if len(info.ToolOptions) == 0 {
			fmt.Println("[none]")
		} else {
			for k, v := range info.ToolOptions {
				fmt.Printf("%s=%v\n", k, v)
			}
		}
		return nil
	}

	// Full Human-readable Table output
	fmt.Println()
	rows := pterm.TableData{
		{"Backend:", pterm.FgDefault.Sprint(info.Backend)},
	}

	if info.Description != "" {
		rows = append(rows, []string{"Description:", info.Description})
	}

	rows = append(rows, []string{"Installed Versions:", formatInstalledWithActive(info.Installed, info.Active)})

	if len(info.Active) > 0 {
		rows = append(rows, []string{"Active Version:", pterm.FgGreen.Sprint(strings.Join(info.Active, " "))})
	} else {
		rows = append(rows, []string{"Active Version:", "[none]"})
	}

	if len(info.Requested) > 0 {
		rows = append(rows, []string{"Requested Version:", strings.Join(info.Requested, " ")})
	}

	if info.ConfigSource != "" {
		rows = append(rows, []string{"Config Source:", info.ConfigSource})
	}

	if len(info.ToolOptions) == 0 {
		rows = append(rows, []string{"Tool Options:", "[none]"})
	} else {
		var optsStr []string
		for k, v := range info.ToolOptions {
			optsStr = append(optsStr, fmt.Sprintf("%s=%v", k, v))
		}
		rows = append(rows, []string{"Tool Options:", strings.Join(optsStr, ", ")})
	}

	if len(info.Security) == 0 {
		rows = append(rows, []string{"Security:", "[none]"})
	} else {
		rows = append(rows, []string{"Security:", strings.Join(info.Security, ", ")})
	}

	pterm.DefaultTable.
		WithSeparator("   ").
		WithData(rows).
		Render()

	return nil
}

func outputJSON(data interface{}) {
	b, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(b))
}

func formatInstalledWithActive(installed, active []string) string {
	if len(installed) == 0 {
		return "[none]"
	}
	
	activeMap := make(map[string]bool)
	for _, a := range active {
		activeMap[a] = true
	}
	
	var res []string
	for _, v := range installed {
		if activeMap[v] {
			res = append(res, pterm.FgGreen.Sprint(v))
		} else {
			res = append(res, v)
		}
	}
	
	return strings.Join(res, " ")
}

// detectActiveVersion returns the active version for a tool by checking shim symlinks.
func detectActiveVersion(shimsDir, installsDir, toolName string, versions []string) string {
	for _, v := range versions {
		binDir := filepath.Join(installsDir, toolName, v, "bin")
		entries, err := os.ReadDir(binDir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			shimPath := filepath.Join(shimsDir, e.Name())
			target, err := os.Readlink(shimPath)
			if err != nil {
				continue
			}
			versionDir := filepath.Join(installsDir, toolName, v)
			if strings.HasPrefix(filepath.Clean(target), filepath.Clean(versionDir)) {
				return v
			}
		}
	}
	return ""
}

// detectShimPath returns the path of the first shim binary for a tool.
func detectShimPath(shimsDir, toolName string) string {
	// The shim name is usually the last component of the tool name.
	// e.g. "cli/cli" → "gh"; "golang/go" → "go"
	// We check the shims directory for any binary that might belong to this tool.
	baseName := filepath.Base(toolName)
	candidate := filepath.Join(shimsDir, baseName)
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	return filepath.Join(shimsDir, "(see install dir)")
}
