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
	ShimPath     string                 `json:"shim_path,omitempty"`
	InstallDir   string                 `json:"install_dir,omitempty"`
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
	configSources := findToolConfigSources(ctx, toolName)
	configSourceStr := ""
	if len(configSources) > 0 {
		configSourceStr = strings.Join(configSources, "\n")
	}
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

	// Active versions are those that are requested AND installed in the current context.
	// Since UniRTM shims are symlinks to the unirtm binary (not specific versions),
	// we compute active versions logically rather than inspecting shim symlink targets.
	activeVersions := detectActiveVersions(requestedVersions, versions)

	shimsDir := env.GetShimsDir()
	installsDir := env.GetInstallsDir()

	shimPath := detectShimPath(shimsDir, toolName)
	if installDir == "" {
		installDir = filepath.Join(installsDir, toolName)
	}

	info := toolInfo{
		Tool:         toolName,
		Backend:      detectedBackend,
		Installed:    versions,
		Requested:    requestedVersions,
		Active:       activeVersions,
		ConfigSource: configSourceStr,
		ToolOptions:  toolOpts,
		Security:     []string{},
		ShimPath:     shimPath,
		InstallDir:   installDir,
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
	pterm.DefaultSection.Printf("Tool: %s\n", pterm.FgCyan.Sprint(info.Tool))

	rows := pterm.TableData{
		{"Backend", pterm.FgMagenta.Sprint(info.Backend)},
		{"Install dir", pterm.FgLightCyan.Sprint(info.InstallDir)},
		{"Shim", pterm.FgLightCyan.Sprint(info.ShimPath)},
	}

	if info.Description != "" {
		rows = append(rows, []string{"Description", info.Description})
	}

	rows = append(rows, []string{"Installed", formatInstalledWithActive(info.Installed, info.Active)})

	if len(info.Active) > 0 {
		rows = append(rows, []string{"Active", pterm.FgGreen.Sprint(strings.Join(info.Active, " "))})
	} else {
		rows = append(rows, []string{"Active", pterm.FgDefault.Sprint("(none)")})
	}

	if len(info.Requested) > 0 {
		rows = append(rows, []string{"Requested", pterm.FgLightBlue.Sprint(strings.Join(info.Requested, " "))})
	}

	if info.ConfigSource != "" {
		rows = append(rows, []string{"Config Source", pterm.FgLightCyan.Sprint(info.ConfigSource)})
	}

	if len(info.ToolOptions) == 0 {
		rows = append(rows, []string{"Tool Options", pterm.FgDefault.Sprint("(none)")})
	} else {
		var optsStr []string
		for k, v := range info.ToolOptions {
			optsStr = append(optsStr, pterm.FgCyan.Sprintf("%s=", k)+fmt.Sprintf("%v", v))
		}
		rows = append(rows, []string{"Tool Options", strings.Join(optsStr, ", ")})
	}

	if len(info.Security) == 0 {
		rows = append(rows, []string{"Security", pterm.FgDefault.Sprint("(none)")})
	} else {
		rows = append(rows, []string{"Security", pterm.FgLightGreen.Sprint(strings.Join(info.Security, ", "))})
	}

	pterm.DefaultTable.
		WithSeparator("   ").
		WithData(rows).
		Render()

	// Verify backend knows about this tool.
	if b, err := backendRegistry.Get(info.Backend); err == nil {
		platform := backend.CurrentPlatform()
		if latest, err := b.ResolveVersion(ctx, toolName, "latest", platform); err == nil && latest != nil {
			fmt.Printf("\n%s Latest available: %s\n",
				pterm.FgDefault.Sprint("→"),
				pterm.FgGreen.Sprint(latest.Version),
			)
		}
	}

	return nil
}

func outputJSON(data interface{}) {
	b, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(b))
}

func formatInstalledWithActive(installed, active []string) string {
	if len(installed) == 0 {
		return pterm.FgYellow.Sprint("(none)")
	}

	activeMap := make(map[string]bool)
	for _, a := range active {
		activeMap[a] = true
	}

	var res []string
	for _, v := range installed {
		if activeMap[v] {
			res = append(res, pterm.FgGreen.Sprint(v+" ✓ active"))
		} else {
			res = append(res, pterm.FgYellow.Sprint(v))
		}
	}

	return strings.Join(res, " ")
}

// detectActiveVersions computes the active versions by checking which of the requested versions
// are currently installed. It supports exact matching and prefix matching (e.g. req "20" matches "20.12.0").
func detectActiveVersions(requestedVersions []string, installedVersions []string) []string {
	var active []string
	seen := make(map[string]bool)

	for _, req := range requestedVersions {
		// Find the best match among installed versions
		for _, v := range installedVersions {
			if v == req || strings.HasPrefix(v, req) || strings.HasPrefix(v, "v"+req) {
				if !seen[v] {
					active = append(active, v)
					seen[v] = true
				}
				break
			}
		}
	}
	return active
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

// findToolConfigSources returns a list of paths to all config files that define the given tool.
func findToolConfigSources(ctx context.Context, toolName string) []string {
	var sources []string
	cfgMgr := config.NewConfigManager()

	// 1. System
	systemPaths := []string{
		"/etc/unirtm/config.toml",
		"/etc/unirtm/config.yaml",
		"/etc/unirtm/config.yml",
	}
	for _, p := range systemPaths {
		if cfg, err := cfgMgr.Load(ctx, p); err == nil && cfg != nil && cfg.ToolsRaw != nil {
			if _, ok := cfg.ToolsRaw[toolName]; ok {
				sources = append(sources, p)
			}
		}
	}

	// 2. Global
	globalPaths := []string{
		filepath.Join(env.GetConfigDir(), "config.toml"),
		filepath.Join(env.GetConfigDir(), "config.yaml"),
		filepath.Join(env.GetConfigDir(), "config.yml"),
	}
	for _, p := range globalPaths {
		if cfg, err := cfgMgr.Load(ctx, p); err == nil && cfg != nil && cfg.ToolsRaw != nil {
			if _, ok := cfg.ToolsRaw[toolName]; ok {
				sources = append(sources, p)
			}
		}
	}

	// 3. Project and Local (traverse from cwd up to root)
	cwd, err := os.Getwd()
	if err == nil {
		curr := cwd
		for {
			files := []string{
				filepath.Join(curr, ".mise.yml"),
				filepath.Join(curr, ".mise.yaml"),
				filepath.Join(curr, ".mise.toml"),
				filepath.Join(curr, "unirtm.yml"),
				filepath.Join(curr, "unirtm.yaml"),
				filepath.Join(curr, "unirtm.toml"),
				filepath.Join(curr, ".unirtm.yml"),
				filepath.Join(curr, ".unirtm.yaml"),
				filepath.Join(curr, ".unirtm.toml"),
				filepath.Join(curr, ".mise.local.yml"),
				filepath.Join(curr, ".mise.local.yaml"),
				filepath.Join(curr, ".mise.local.toml"),
				filepath.Join(curr, "unirtm.local.yml"),
				filepath.Join(curr, "unirtm.local.yaml"),
				filepath.Join(curr, "unirtm.local.toml"),
				filepath.Join(curr, ".unirtm.local.yml"),
				filepath.Join(curr, ".unirtm.local.yaml"),
				filepath.Join(curr, ".unirtm.local.toml"),
			}
			var dirSources []string
			for _, p := range files {
				if cfg, err := cfgMgr.Load(ctx, p); err == nil && cfg != nil && cfg.ToolsRaw != nil {
					if _, ok := cfg.ToolsRaw[toolName]; ok {
						dirSources = append(dirSources, p)
					}
				}
			}
			// Prepend to list since closer to cwd = higher precedence, but usually printed top-down or bottom-up?
			// Let's just append them. The order of resolution in LoadHierarchy goes system -> global -> project -> local,
			// where local overrides project. We'll just collect them in traversal order (local -> global)
			// and then reverse them or just return them. Appending is fine.
			sources = append(sources, dirSources...)

			parent := filepath.Dir(curr)
			if parent == curr {
				break
			}
			curr = parent
		}
	}

	return sources
}
