// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
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

	// 1. Build & System Info
	pterm.DefaultSection.Println("🛠️  Build & System Information")
	osArch := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	pterm.DefaultBulletList.WithItems([]pterm.BulletListItem{
		{Level: 0, Text: fmt.Sprintf("%-12s: %s", "Project", pterm.LightCyan(env.ProjectName))},
		{Level: 0, Text: fmt.Sprintf("%-12s: %s", "Version", pterm.LightCyan(fmt.Sprintf("%s (%s)", env.GitTag, env.CommitHash)))},
		{Level: 0, Text: fmt.Sprintf("%-12s: %s", "Build Time", pterm.LightCyan(env.BuildTime))},
		{Level: 0, Text: fmt.Sprintf("%-12s: %s", "Platform", pterm.LightCyan(osArch))},
		{Level: 0, Text: fmt.Sprintf("%-12s: %s", "Go Version", pterm.LightCyan(runtime.Version()))},
	}).Render()

	// 2. Shell Info
	pterm.DefaultSection.Println("🐚 Shell Information")
	shellPath := env.Get("SHELL")
	shellName := filepath.Base(shellPath)
	isActivated := env.Get("ACTIVE") != ""
	
	statusStr := pterm.LightGreen("✓ activated")
	if !isActivated {
		statusStr = pterm.LightRed("✗ NOT activated")
	}

	pterm.DefaultBulletList.WithItems([]pterm.BulletListItem{
		{Level: 0, Text: fmt.Sprintf("%-12s: %s (%s)", "Shell", pterm.LightCyan(shellName), pterm.FgGray.Sprint(shellPath))},
		{Level: 0, Text: fmt.Sprintf("%-12s: %s", "Status", statusStr)},
	}).Render()

	// 3. Directory Health
	pterm.DefaultSection.Println("📁 Directory Health Check")
	dirChecks := []struct {
		name string
		path string
	}{
		{"Config", env.GetConfigDir()},
		{"Data", env.GetDataDir()},
		{"Cache", env.GetCacheDir()},
		{"Shims", env.GetShimsDir()},
		{"State", filepath.Join(env.GetDataDir(), "state")},
	}

	var dirTable [][]string
	dirTable = append(dirTable, []string{"Directory", "Path", "Status"})
	for _, dc := range dirChecks {
		status := pterm.LightGreen("✓ ok")
		if _, err := os.Stat(dc.path); os.IsNotExist(err) {
			status = pterm.LightYellow("⚠ missing")
		} else {
			testFile := filepath.Join(dc.path, ".doctor_write_test")
			if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
				status = pterm.LightRed("✗ unwritable")
			} else {
				_ = os.Remove(testFile)
			}
		}
		dirTable = append(dirTable, []string{pterm.Bold.Sprint(dc.name), pterm.FgGray.Sprint(dc.path), status})
	}
	pterm.DefaultTable.WithHasHeader().WithData(dirTable).Render()

	// 4. PATH & Shadowing Detection
	pterm.DefaultSection.Println("🛤️  PATH & Shadowing Detection")
	pathItems := filepath.SplitList(env.Get("PATH"))
	shimsDir := env.GetShimsDir()
	
	shimIdx := -1
	for i, p := range pathItems {
		if strings.EqualFold(p, shimsDir) {
			shimIdx = i
			break
		}
	}

	if shimIdx == -1 {
		pterm.Error.Println("UniRTM shims directory is NOT in your PATH.")
		pterm.DefaultBox.WithTitle("Fix Suggestion").Println(
			fmt.Sprintf("Add this to your shell profile (%s):\n%s", 
			pterm.LightBlue("~/.zshrc" /* or appropriate */),
			pterm.LightMagenta(fmt.Sprintf("eval \"$(unirtm activate %s)\"", shellName))),
		)
	} else if shimIdx > 0 {
		pterm.Warning.Println("UniRTM shims are NOT at the front of your PATH. Other tools may shadow UniRTM shims.")
		shadowed := pathItems[:shimIdx]
		pterm.Info.Printf("Preceding paths (shadowing UniRTM):\n  %s\n", pterm.FgGray.Sprint(strings.Join(shadowed, "\n  ")))
	} else {
		pterm.Success.Println("UniRTM shims are correctly placed at the front of your PATH.")
	}

	// 5. Environment Variables (Redacted & Hierarchical)
	pterm.DefaultSection.Println("🔑 Environment Variables (UniRTM / Mise / Raw)")
	var envTable [][]string
	envTable = append(envTable, []string{"Variable", "Source", "Resolved Value"})
	
	// Track some key variables we want to show
	keyVars := []string{"DATA_DIR", "CONFIG_DIR", "CACHE_DIR", "EXPERIMENTAL", "DEBUG", "PATH"}
	for _, kv := range keyVars {
		unirtmVal := os.Getenv("UNIRTM_" + kv)
		miseVal := os.Getenv("MISE_" + kv)
		resolved := env.Get(kv)

		source := "Native"
		if unirtmVal != "" {
			source = "UNIRTM_"
		} else if miseVal != "" {
			source = "MISE_"
		}

		displayVal := resolved
		if kv == "PATH" && len(displayVal) > 50 {
			displayVal = displayVal[:47] + "..."
		}
		
		envTable = append(envTable, []string{
			pterm.LightBlue(kv),
			pterm.FgGray.Sprint(source),
			pterm.LightCyan(displayVal),
		})
	}
	pterm.DefaultTable.WithHasHeader().WithData(envTable).Render()

	// 6. Active Configuration Files
	pterm.DefaultSection.Println("📝 Active Configuration Files")
	if cfg != nil {
		cwd, _ := os.Getwd()
		possible := []string{
			filepath.Join(cwd, ".unirtm.toml"),
			filepath.Join(cwd, "unirtm.toml"),
			filepath.Join(cwd, ".mise.toml"),
			filepath.Join(cwd, "mise.toml"),
			env.GetGlobalConfigPath(),
		}
		foundCfg := false
		for _, p := range possible {
			if _, err := os.Stat(p); err == nil {
				foundCfg = true
				pterm.Success.Printf("Loaded: %s\n", pterm.FgGray.Sprint(p))
			}
		}
		if !foundCfg {
			pterm.Info.Println("No project or global config files found (using defaults).")
		}
	}

	// 7. Settings (Effective)
	if cfg != nil {
		pterm.DefaultSection.Println("⚙️  Effective Settings")
		var setTable [][]string
		setTable = append(setTable, []string{"Setting", "Value"})
		setTable = append(setTable, []string{"Experimental", fmt.Sprintf("%v", cfg.Settings.Experimental)})
		
		autoInstall := "default"
		if cfg.Settings.AutoInstall != nil {
			autoInstall = fmt.Sprintf("%v", *cfg.Settings.AutoInstall)
		}
		setTable = append(setTable, []string{"Auto Install", autoInstall})
		setTable = append(setTable, []string{"Always Keep Download", fmt.Sprintf("%v", cfg.Settings.AlwaysKeepDownload)})
		setTable = append(setTable, []string{"HTTP Timeout", fmt.Sprintf("%ds", cfg.Settings.HTTPTimeout)})
		
		pterm.DefaultTable.WithHasHeader().WithData(setTable).Render()
	}

	// 8. Network Connectivity
	pterm.DefaultSection.Println("🌐 Network Connectivity")
	
	// Detailed Proxy Summary
	proxyReport := []string{}
	if v := env.Get("HTTP_PROXY"); v != "" { proxyReport = append(proxyReport, fmt.Sprintf("HTTP: %s", v)) }
	if v := env.Get("HTTPS_PROXY"); v != "" { proxyReport = append(proxyReport, fmt.Sprintf("HTTPS: %s", v)) }
	if v := env.Get("ALL_PROXY"); v != "" { proxyReport = append(proxyReport, fmt.Sprintf("ALL: %s", v)) }
	
	if len(proxyReport) > 0 {
		pterm.Info.Printf("Proxy Chain: %s\n", pterm.FgGray.Sprint(strings.Join(proxyReport, " | ")))
	} else {
		pterm.Info.Println("Direct Connection (No system proxy detected)")
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
	
	targets := []struct{ name, url string }{
		{"GitHub API", "https://api.github.com"},
		{"Aqua Registry", "https://aquaproj.github.io"},
	}

	for _, t := range targets {
		start := time.Now()
		resp, err := client.Get(t.url)
		duration := time.Since(start)

		if err != nil {
			pterm.Error.Printf("%-15s %s (%v)\n", t.name, "Unreachable", err)
		} else {
			pterm.Success.Printf("%-15s %s (HTTP %d, %s)\n", t.name, "Connected", resp.StatusCode, duration.Round(time.Millisecond).String())
			_ = resp.Body.Close()
		}
	}

	// 9. Database & Tool Integrity
	pterm.DefaultSection.Println("📊 Database & Tool Integrity")
	dbPath := env.GetDatabasePath()
	db, err := database.Open(ctx, database.Config{Path: dbPath, WALMode: true})
	if err != nil {
		pterm.Error.Printf("Database: %v\n", err)
	} else {
		defer db.Close()
		pterm.Success.Printf("Database: %s\n", pterm.FgGray.Sprint(dbPath))
		
		installRepo, _ := sqlite.NewInstallationRepository(db.Conn())
		installs, _ := installRepo.List(ctx)
		missing := 0
		for _, inst := range installs {
			if _, err := os.Stat(inst.InstallPath); os.IsNotExist(err) {
				missing++
			}
		}
		if missing > 0 {
			pterm.Warning.Printf("Tools: %d/%d installations have missing directories.\n", missing, len(installs))
		} else if len(installs) > 0 {
			pterm.Success.Printf("Tools: All %d installed tools are present on disk.\n", len(installs))
		} else {
			pterm.Info.Println("Tools: No tools installed yet.")
		}
	}

	fmt.Println()
	pterm.DefaultBox.WithTitle(pterm.LightGreen("Diagnostics Complete")).Println("Your UniRTM environment is healthy and ready for action!")
	return nil
}
