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
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/spf13/cobra"
)

func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(doctorCmd)
	}
}

// doctorCmd runs a comprehensive system health check.
var doctorCmd = &cobra.Command{
	Use:     "doctor",
	Aliases: []string{"dr"},
	Short:   "Check system health and diagnose issues",
	Long: `Check UniRTM system health and diagnose potential issues.

This command aligns with 'mise doctor' to ensure your environment is
correctly configured for UniRTM, while providing enhanced visual diagnostics.`,
	Args: cobra.NoArgs,
	RunE: runDoctor,
}

func runDoctor(cmd *cobra.Command, args []string) error {
	pterm.DefaultHeader.WithFullWidth().
		WithBackgroundStyle(pterm.NewStyle(pterm.BgLightBlue)).
		WithTextStyle(pterm.NewStyle(pterm.FgBlack)).
		Println("UniRTM System Diagnostics (Doctor)")

	ctx := context.Background()
	cfg, _ := config.LoadFull()

	// 1. Core Status
	pterm.DefaultSection.Println("🚀 Core Status")
	isActivated := env.Get("ACTIVE") != ""
	shimsDir := env.GetShimsDir()
	pathItems := filepath.SplitList(env.Get("PATH"))
	shimsOnPath := false
	for _, p := range pathItems {
		if strings.EqualFold(p, shimsDir) {
			shimsOnPath = true
			break
		}
	}

	statusItems := []pterm.BulletListItem{
		{Level: 0, Text: fmt.Sprintf("%-15s: %s (%s)", "version", pterm.LightCyan(env.GitTag), env.CommitHash)},
		{Level: 0, Text: fmt.Sprintf("%-15s: %s", "activated", formatBoolStatus(isActivated))},
		{Level: 0, Text: fmt.Sprintf("%-15s: %s", "shims_on_path", formatBoolStatus(shimsOnPath))},
	}
	pterm.DefaultBulletList.WithItems(statusItems).Render()

	// 2. Build Information
	pterm.DefaultSection.Println("🛠️  Build Information")
	pterm.DefaultBulletList.WithItems([]pterm.BulletListItem{
		{Level: 0, Text: fmt.Sprintf("%-12s: %s/%s", "Target", runtime.GOOS, runtime.GOARCH)},
		{Level: 0, Text: fmt.Sprintf("%-12s: %s", "Go Version", runtime.Version())},
		{Level: 0, Text: fmt.Sprintf("%-12s: %s", "Built", env.BuildTime)},
	}).Render()

	// 3. Shell Information
	pterm.DefaultSection.Println("🐚 Shell Information")
	shellPath := env.Get("SHELL")
	shellVer := ""
	if out, err := exec.Command(shellPath, "--version").Output(); err == nil {
		shellVer = strings.TrimSpace(string(out))
	}

	pterm.DefaultBulletList.WithItems([]pterm.BulletListItem{
		{Level: 0, Text: fmt.Sprintf("%-12s: %s", "Shell", pterm.LightCyan(shellPath))},
		{Level: 0, Text: fmt.Sprintf("%-12s: %s", "Version", shellVer)},
	}).Render()

	// 4. Directories
	pterm.DefaultSection.Println("📁 Directories")
	dirData := [][]string{
		{"cache", env.GetCacheDir()},
		{"config", env.GetConfigDir()},
		{"data", env.GetDataDir()},
		{"shims", env.GetShimsDir()},
		{"state", filepath.Join(env.GetDataDir(), "state")},
	}
	for i := range dirData {
		dirData[i][0] = pterm.Bold.Sprint(dirData[i][0])
		dirData[i][1] = pterm.FgGray.Sprint(dirData[i][1])
	}
	pterm.DefaultTable.WithData(dirData).Render()

	// 5. Config Files
	pterm.DefaultSection.Println("📝 Configuration Files")
	if cfg != nil {
		cwd, _ := os.Getwd()
		configs := []string{
			filepath.Join(cwd, ".unirtm.toml"),
			filepath.Join(cwd, "unirtm.toml"),
			filepath.Join(cwd, ".mise.toml"),
			filepath.Join(cwd, "mise.toml"),
			env.GetGlobalConfigPath(),
		}
		found := false
		for _, c := range configs {
			if _, err := os.Stat(c); err == nil {
				pterm.Success.Printf("Loaded: %s\n", pterm.FgGray.Sprint(c))
				found = true
			}
		}
		if !found {
			pterm.Info.Println("No configuration files found (using defaults).")
		}
	}

	// 6. Backends
	pterm.DefaultSection.Println("🔌 Available Backends")
	backends := backend.List()
	pterm.DefaultBulletList.WithItems(stringToBulletItems(backends)).Render()

	// 7. Toolset (Active Tools)
	pterm.DefaultSection.Println("🧰 Active Toolset")
	if cfg != nil && len(cfg.Tools) > 0 {
		var toolData [][]string
		toolData = append(toolData, []string{"Tool", "Version", "Backend/Provider"})
		for name, t := range cfg.Tools {
			source := t.Backend
			if source == "" {
				source = t.Provider
			}
			if source == "" {
				source = "default"
			}
			toolData = append(toolData, []string{
				pterm.LightBlue(name),
				pterm.LightCyan(t.Version),
				pterm.FgGray.Sprint(source),
			})
		}
		pterm.DefaultTable.WithHasHeader().WithData(toolData).Render()
	} else {
		pterm.Info.Println("No tools defined in active configuration.")
	}

	// 8. PATH Breakdown
	pterm.DefaultSection.Println("🛤️  PATH Breakdown")
	for _, p := range pathItems {
		if strings.EqualFold(p, shimsDir) {
			pterm.Success.Printf("%s %s\n", p, pterm.LightMagenta("(UniRTM Shims)"))
		} else if strings.Contains(p, "unirtm") || strings.Contains(p, ".local/share/unirtm") {
			pterm.Info.Printf("%s %s\n", p, pterm.LightCyan("(UniRTM Managed)"))
		} else {
			pterm.FgGray.Println(p)
		}
	}

	// 9. Environment Variables (Hierarchy)
	pterm.DefaultSection.Println("🔑 Environment Variables (Hierarchy Check)")
	var envData [][]string
	envData = append(envData, []string{"Key", "Source", "Value"})
	vars := []string{"DATA_DIR", "CONFIG_DIR", "CACHE_DIR", "EXPERIMENTAL", "DEBUG", "GITHUB_TOKEN"}
	for _, v := range vars {
		source := "Raw"
		if os.Getenv("UNIRTM_"+v) != "" {
			source = "UNIRTM_"
		} else if os.Getenv("MISE_"+v) != "" {
			source = "MISE_"
		}
		
		val := env.Get(v)
		if strings.Contains(v, "TOKEN") && val != "" {
			val = "******** [REDACTED]"
		}
		if len(val) > 40 {
			val = val[:37] + "..."
		}

		envData = append(envData, []string{pterm.LightBlue(v), pterm.FgGray.Sprint(source), pterm.LightCyan(val)})
	}
	pterm.DefaultTable.WithHasHeader().WithData(envData).Render()

	// 10. Database & Network (Quick Checks)
	pterm.DefaultSection.Println("🌐 Health Checks")
	
	// DB Check
	dbPath := env.GetDatabasePath()
	db, err := database.Open(ctx, database.Config{Path: dbPath, WALMode: true})
	if err != nil {
		pterm.Error.Printf("Database: %v\n", err)
	} else {
		db.Close()
		pterm.Success.Printf("Database: %s (Open successful)\n", pterm.FgGray.Sprint(dbPath))
	}

	// Network Check
	client := &http.Client{Timeout: 5 * time.Second}
	if resp, err := client.Get("https://api.github.com"); err == nil {
		pterm.Success.Printf("Network: GitHub API connected (HTTP %d)\n", resp.StatusCode)
		resp.Body.Close()
	} else {
		pterm.Error.Printf("Network: GitHub API unreachable (%v)\n", err)
	}

	fmt.Println()
	if !isActivated || !shimsOnPath {
		pterm.Warning.Println("UniRTM is not fully integrated into your shell.")
		pterm.Info.Printf("Run %s to see activation instructions.\n", pterm.LightMagenta("unirtm help activate"))
	} else {
		pterm.DefaultBox.WithTitle(pterm.LightGreen("Diagnostics Complete")).Println("Your UniRTM environment is perfectly configured!")
	}

	return nil
}

func formatBoolStatus(b bool) string {
	if b {
		return pterm.LightGreen("yes")
	}
	return pterm.LightRed("no")
}

func stringToBulletItems(ss []string) []pterm.BulletListItem {
	var items []pterm.BulletListItem
	for _, s := range ss {
		items = append(items, pterm.BulletListItem{Level: 0, Text: s})
	}
	return items
}
