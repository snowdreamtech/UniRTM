// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
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
correctly configured for UniRTM.`,
	Args: cobra.NoArgs,
	RunE: runDoctor,
}

func runDoctor(cmd *cobra.Command, args []string) error {
	pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgLightBlue)).WithTextStyle(pterm.NewStyle(pterm.FgBlack)).Println("UniRTM System Diagnostics (Doctor)")

	ctx := context.Background()

	// 1. Build & System Info
	pterm.DefaultSection.Println("Build & System Information")
	osArch := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	pterm.DefaultBulletList.WithItems([]pterm.BulletListItem{
		{Level: 0, Text: fmt.Sprintf("Project: %s", pterm.LightCyan(env.ProjectName))},
		{Level: 0, Text: fmt.Sprintf("Version: %s", pterm.LightCyan(fmt.Sprintf("%s-%s", env.GitTag, env.CommitHash)))},
		{Level: 0, Text: fmt.Sprintf("Platform: %s", pterm.LightCyan(osArch))},
		{Level: 0, Text: fmt.Sprintf("Go Version: %s", pterm.LightCyan(runtime.Version()))},
	}).Render()

	// 2. Shell Info
	pterm.DefaultSection.Println("Shell Information")
	shellPath := os.Getenv("SHELL")
	shellName := filepath.Base(shellPath)
	isActivated := os.Getenv("UNIRTM_ACTIVE") != ""
	
	statusStr := pterm.LightGreen("✓ activated")
	if !isActivated {
		statusStr = pterm.LightRed("✗ NOT activated")
	}

	pterm.DefaultBulletList.WithItems([]pterm.BulletListItem{
		{Level: 0, Text: fmt.Sprintf("Shell: %s (%s)", pterm.LightCyan(shellName), shellPath)},
		{Level: 0, Text: fmt.Sprintf("Status: %s", statusStr)},
	}).Render()

	// 3. Directory Health
	pterm.DefaultSection.Println("Directory Health Check")
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
			// Check writability
			testFile := filepath.Join(dc.path, ".doctor_write_test")
			if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
				status = pterm.LightRed("✗ unwritable")
			} else {
				os.Remove(testFile)
			}
		}
		dirTable = append(dirTable, []string{pterm.Bold.Sprint(dc.name), pterm.FgGray.Sprint(dc.path), status})
	}
	pterm.DefaultTable.WithHasHeader().WithData(dirTable).Render()

	// 4. PATH & Shadowing Detection
	pterm.DefaultSection.Println("PATH & Shadowing Detection")
	pathItems := filepath.SplitList(os.Getenv("PATH"))
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
		pterm.DefaultBox.WithTitle("Action Required").Println(
			fmt.Sprintf("Add this to your shell config:\n%s", 
			pterm.LightMagenta(fmt.Sprintf("eval \"$(unirtm activate %s)\"", shellName))),
		)
	} else if shimIdx > 0 {
		pterm.Warning.Println("UniRTM shims are not at the front of your PATH.")
		shadowed := pathItems[:shimIdx]
		pterm.Info.Printf("Preceding paths:\n  %s\n", strings.Join(shadowed, "\n  "))
	} else {
		pterm.Success.Println("UniRTM shims are correctly placed at the front of your PATH.")
	}

	// 5. Environment Variables (Redacted)
	pterm.DefaultSection.Println("Environment Variables (UNIRTM_*)")
	var envTable [][]string
	envTable = append(envTable, []string{"Variable", "Value"})
	foundAny := false
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if strings.HasPrefix(pair[0], "UNIRTM_") {
			foundAny = true
			val := pair[1]
			// Redact sensitive tokens
			if strings.Contains(pair[0], "TOKEN") || strings.Contains(pair[0], "KEY") || strings.Contains(pair[0], "SECRET") {
				val = pterm.LightMagenta("******** [REDACTED]")
			} else if len(val) > 60 {
				// Surpass: Truncate long values like UNIRTM_PATH to prevent layout issues
				val = val[:57] + pterm.FgGray.Sprint("...")
			}
			envTable = append(envTable, []string{pterm.LightBlue(pair[0]), val})
		}
	}
	if foundAny {
		pterm.DefaultTable.WithHasHeader().WithData(envTable).Render()
	} else {
		pterm.Info.Println("No UNIRTM_ environment variables defined.")
	}

	// 6. Config Files
	pterm.DefaultSection.Println("Active Configuration Files")
	cfg, _ := config.LoadFull()
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
			pterm.Info.Println("No config files found (using defaults).")
		}
	}

	// 7. Network Connectivity
	pterm.DefaultSection.Println("Network Connectivity")
	
	// Detect Proxy (Check both upper and lower case)
	proxies := []string{"HTTP_PROXY", "http_proxy", "HTTPS_PROXY", "https_proxy", "ALL_PROXY", "all_proxy", "NO_PROXY", "no_proxy"}
	foundProxy := false
	for _, p := range proxies {
		if val := os.Getenv(p); val != "" {
			foundProxy = true
			pterm.Info.Printf("Proxy detected: %s=%s\n", pterm.LightBlue(p), pterm.FgGray.Sprint(val))
		}
	}
	
	// Check UniRTM specific GitHub Proxy
	if cfg != nil && cfg.Settings.GitHubProxy != "" {
		foundProxy = true
		pterm.Info.Printf("UniRTM GitHub Proxy: %s\n", pterm.FgGray.Sprint(cfg.Settings.GitHubProxy))
	}

	if foundProxy {
		fmt.Println()
	} else {
		pterm.Info.Println("No system proxy detected. UniRTM will attempt direct connections.")
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
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
		
		// Check which proxy will be used for this specific URL
		targetURL, _ := url.Parse(t.url)
		proxyURL, _ := http.ProxyFromEnvironment(&http.Request{URL: targetURL})
		proxyStr := "Direct"
		if proxyURL != nil {
			proxyStr = proxyURL.String()
		}

		resp, err := client.Get(t.url)
		duration := time.Since(start)

		if err != nil {
			errMsg := err.Error()
			if strings.Contains(errMsg, "connection reset") || strings.Contains(errMsg, "refused") {
				errMsg += pterm.LightYellow(" (Check your proxy/firewall settings or ensure the proxy is active)")
			}
			pterm.Error.Printf("%-15s %-25s %s (%s)\n", t.name, pterm.FgGray.Sprint(proxyStr), "Unreachable", errMsg)
		} else {
			pterm.Success.Printf("%-15s %-25s %s (HTTP %d, %s)\n", t.name, pterm.FgGray.Sprint(proxyStr), "OK", resp.StatusCode, duration.Round(time.Millisecond).String())
			resp.Body.Close()
		}
	}

	// 8. Database & Tools
	pterm.DefaultSection.Println("Database & Tool Integrity")
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
	pterm.Success.Println("Diagnostics complete. Your UniRTM is ready!")
	return nil
}
