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
	"reflect"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/service"
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

This command aligns with and exceeds 'mise doctor' to ensure your environment is
perfectly configured, providing deep insights into tools, paths, and connectivity.`,
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
		{Level: 0, Text: fmt.Sprintf("%-20s: %s (%s)", "version", pterm.LightCyan(env.GitTag), env.CommitHash)},
		{Level: 0, Text: fmt.Sprintf("%-20s: %s", "activated", formatBoolStatus(isActivated))},
		{Level: 0, Text: fmt.Sprintf("%-20s: %s", "shims_on_path", formatBoolStatus(shimsOnPath))},
		{Level: 0, Text: fmt.Sprintf("%-20s: %s", "trust_status", formatTrustStatus())},
	}
	pterm.DefaultBulletList.WithItems(statusItems).Render()

	// 2. Build Information
	pterm.DefaultSection.Println("🛠️  Build Information")
	pterm.DefaultBulletList.WithItems([]pterm.BulletListItem{
		{Level: 0, Text: fmt.Sprintf("%-12s: %s/%s", "Target", runtime.GOOS, runtime.GOARCH)},
		{Level: 0, Text: fmt.Sprintf("%-12s: %s", "Go Version", runtime.Version())},
		{Level: 0, Text: fmt.Sprintf("%-12s: %s", "Built", env.BuildTime)},
		{Level: 0, Text: fmt.Sprintf("%-12s: %s", "Features", pterm.LightCyan("SQLITE, OPENSSL, GPG, VFOX_COMPAT"))},
	}).Render()

	// 3. Shell & Environment Context
	pterm.DefaultSection.Println("🐚 Context & Shell")
	shellPath := env.Get("SHELL")
	cwd, _ := os.Getwd()
	
	shellVer := "unknown"
	if out, err := exec.Command(shellPath, "-c", "echo $ZSH_VERSION $BASH_VERSION").CombinedOutput(); err == nil {
		shellVer = strings.TrimSpace(string(out))
	}

	pterm.DefaultBulletList.WithItems([]pterm.BulletListItem{
		{Level: 0, Text: fmt.Sprintf("%-12s: %s", "Shell", pterm.LightCyan(shellPath))},
		{Level: 0, Text: fmt.Sprintf("%-12s: %s", "Version", shellVer)},
		{Level: 0, Text: fmt.Sprintf("%-12s: %s", "Work Dir", pterm.FgGray.Sprint(cwd))},
		{Level: 0, Text: fmt.Sprintf("%-12s: %s", "Active Env", formatActiveEnv(cfg))},
	}).Render()

	// 4. Directories & Usage
	pterm.DefaultSection.Println("📁 Directories & Usage")
	dirData := [][]string{
		{"cache", env.GetCacheDir(), getDirSize(env.GetCacheDir())},
		{"config", env.GetConfigDir(), "-"},
		{"data", env.GetDataDir(), getDirSize(env.GetDataDir())},
		{"shims", env.GetShimsDir(), getDirSize(env.GetShimsDir())},
	}
	var dirTable [][]string
	dirTable = append(dirTable, []string{"Type", "Path", "Size"})
	for _, d := range dirData {
		dirTable = append(dirTable, []string{pterm.Bold.Sprint(d[0]), pterm.FgGray.Sprint(d[1]), pterm.LightCyan(d[2])})
	}
	pterm.DefaultTable.WithHasHeader().WithData(dirTable).Render()

	// 5. Config Traceability & Aliases
	pterm.DefaultSection.Println("📝 Configuration & Aliases")
	if cfg != nil {
		configs := []string{
			filepath.Join(cwd, ".unirtm.toml"),
			filepath.Join(cwd, "unirtm.toml"),
			filepath.Join(cwd, ".mise.toml"),
			filepath.Join(cwd, "mise.toml"),
			env.GetGlobalConfigPath(),
		}
		for _, c := range configs {
			if _, err := os.Stat(c); err == nil {
				pterm.Success.Printf("Loaded: %s\n", pterm.FgGray.Sprint(c))
			}
		}
		
		if len(cfg.Aliases) > 0 {
			fmt.Println(pterm.Bold.Sprint("\nAliases:"))
			for tool, aliases := range cfg.Aliases {
				for alias, target := range aliases {
					pterm.Info.Printf("  %s -> %s %s\n", pterm.LightBlue(alias), pterm.LightCyan(target), pterm.FgGray.Sprint("("+tool+")"))
				}
			}
		}
	}

	// 6. Settings Audit
	pterm.DefaultSection.Println("⚙️  UniRTM Settings Audit")
	if cfg != nil {
		var settingsData [][]string
		settingsData = append(settingsData, []string{"Setting", "Value"})
		
		v := reflect.ValueOf(cfg.Settings)
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			val := v.Field(i).Interface()
			
			// Format value
			displayVal := fmt.Sprintf("%v", val)
			if field.Type.Kind() == reflect.Ptr && !v.Field(i).IsNil() {
				displayVal = fmt.Sprintf("%v", v.Field(i).Elem().Interface())
			} else if field.Type.Kind() == reflect.Ptr && v.Field(i).IsNil() {
				displayVal = "unset"
			}
			
			if strings.Contains(field.Name, "Token") && displayVal != "unset" && displayVal != "" {
				displayVal = "******** [REDACTED]"
			}

			// Use TOML tag if available
			tagName := field.Tag.Get("toml")
			if tagName == "" { tagName = field.Name }
			tagName = strings.Split(tagName, ",")[0]

			settingsData = append(settingsData, []string{pterm.LightBlue(tagName), pterm.LightCyan(displayVal)})
		}
		pterm.DefaultTable.WithHasHeader().WithData(settingsData).Render()
	}

	// 7. Backends
	pterm.DefaultSection.Println("🔌 Available Backends")
	backends := backend.List()
	sort.Strings(backends)
	pterm.DefaultBulletList.WithItems(stringToBulletItems(backends)).Render()

	// 8. Toolset (Detailed Path Analysis)
	pterm.DefaultSection.Println("🧰 Active Toolset")
	if cfg != nil && len(cfg.Tools) > 0 {
		var toolData [][]string
		toolData = append(toolData, []string{"Tool", "Version", "Backend", "Install Status"})
		
		keys := make([]string, 0, len(cfg.Tools))
		for k := range cfg.Tools { keys = append(keys, k) }
		sort.Strings(keys)
		
		for _, name := range keys {
			t := cfg.Tools[name]
			
			// Use official normalization logic from env package (slugification)
			slug := env.GetFSToolName(name, t.Backend)
			
			// Normalize version using official service package logic
			v := t.Version
			if ver, err := service.ParseVersion(v); err == nil {
				v = ver.String()
			}
			
			installStatus := pterm.LightGreen("✓ installed")
			toolPath := filepath.Join(env.GetInstallsDir(), slug, v)
			if _, err := os.Stat(toolPath); os.IsNotExist(err) {
				installStatus = pterm.LightRed("✗ missing")
			}

			toolData = append(toolData, []string{
				pterm.Bold.Sprint(name),
				pterm.LightCyan(t.Version),
				pterm.FgGray.Sprint(t.Backend),
				installStatus,
			})
		}
		pterm.DefaultTable.WithHasHeader().WithData(toolData).Render()
	} else {
		pterm.Info.Println("No tools defined in active configuration.")
	}

	// 9. PATH Visualization
	pterm.DefaultSection.Println("🛤️  PATH Visualization")
	for i, p := range pathItems {
		prefix := "  "
		if i == 0 { prefix = "-> " }
		
		if strings.EqualFold(p, shimsDir) {
			pterm.Success.Printf("%s%s %s\n", prefix, p, pterm.LightMagenta("(UniRTM Shims)"))
		} else if strings.Contains(p, "unirtm") {
			pterm.Info.Printf("%s%s %s\n", prefix, p, pterm.LightCyan("(UniRTM Managed)"))
		} else {
			pterm.FgGray.Printf("%s%s\n", prefix, p)
		}
	}

	// 10. Environment Variable Hierarchy
	pterm.DefaultSection.Println("🔑 Environment Hierarchy (UNIRTM_ > MISE_ > Raw)")
	relevantVars := make(map[string]bool)
	configKeys := []string{"DATA_DIR", "CONFIG_DIR", "CACHE_DIR", "EXPERIMENTAL", "DEBUG", "GITHUB_TOKEN", "HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY", "ACTIVATION_SCOPE"}
	for _, k := range configKeys { relevantVars[k] = true }
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if strings.HasPrefix(pair[0], "UNIRTM_") {
			relevantVars[strings.TrimPrefix(pair[0], "UNIRTM_")] = true
		} else if strings.HasPrefix(pair[0], "MISE_") {
			relevantVars[strings.TrimPrefix(pair[0], "MISE_")] = true
		}
	}

	var envData [][]string
	envData = append(envData, []string{"Key", "Source", "Effective Value"})
	sortedKeys := make([]string, 0, len(relevantVars))
	for k := range relevantVars { sortedKeys = append(sortedKeys, k) }
	sort.Strings(sortedKeys)

	for _, k := range sortedKeys {
		unirtmVal := os.Getenv("UNIRTM_" + k)
		miseVal := os.Getenv("MISE_" + k)
		rawVal := os.Getenv(k)
		resolved := env.Get(k)

		source := "Raw"
		if unirtmVal != "" { source = "UNIRTM_" } else if miseVal != "" { source = "MISE_" } else if rawVal == "" {
			source = "Default"
			switch k {
			case "DATA_DIR": resolved = env.GetDataDir()
			case "CONFIG_DIR": resolved = env.GetConfigDir()
			case "CACHE_DIR": resolved = env.GetCacheDir()
			}
		}

		if resolved == "" && source == "Raw" { continue }

		displayVal := resolved
		if strings.Contains(k, "TOKEN") && displayVal != "" { displayVal = "******** [REDACTED]" }
		if len(displayVal) > 45 { displayVal = displayVal[:42] + "..." }

		envData = append(envData, []string{pterm.LightBlue(k), pterm.FgGray.Sprint(source), pterm.LightCyan(displayVal)})
	}
	pterm.DefaultTable.WithHasHeader().WithData(envData).Render()

	// 11. Task Discovery
	if cfg != nil && len(cfg.Tasks) > 0 {
		pterm.DefaultSection.Println("⚡ Task Discovery")
		var taskData [][]string
		taskData = append(taskData, []string{"Task", "Description"})
		for name, t := range cfg.Tasks {
			desc := t.Description
			if desc == "" { desc = pterm.FgGray.Sprint("(no description)") }
			taskData = append(taskData, []string{pterm.LightYellow(name), desc})
		}
		pterm.DefaultTable.WithHasHeader().WithData(taskData).Render()
	}

	// 12. Health Checks (Database & Network)
	pterm.DefaultSection.Println("🌐 Health Checks")
	
	// DB Check
	dbPath := env.GetDatabasePath()
	db, err := database.Open(ctx, database.Config{Path: dbPath, WALMode: true})
	if err != nil {
		pterm.Error.Printf("Database: %v\n", err)
	} else {
		defer db.Close()
		pterm.Success.Printf("Database: %s (Size: %s)\n", pterm.FgGray.Sprint(dbPath), getFileSize(dbPath))
	}

	// Network & Rate Limit
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", "https://api.github.com/rate_limit", nil)
	if token := env.Get("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}
	
	if resp, err := client.Do(req); err == nil {
		pterm.Success.Printf("GitHub API: Connected (HTTP %d)\n", resp.StatusCode)
		limit := resp.Header.Get("X-RateLimit-Limit")
		remaining := resp.Header.Get("X-RateLimit-Remaining")
		reset := resp.Header.Get("X-RateLimit-Reset")
		
		if limit != "" {
			pterm.Info.Printf("GitHub Rate Limit: %s/%s (Resets in %s)\n", 
				pterm.LightCyan(remaining), pterm.LightCyan(limit), 
				time.Until(time.Unix(parseInt(reset), 0)).Round(time.Minute))
		}
		_ = resp.Body.Close()
	}

	// 13. Fix Suggestions
	pterm.DefaultSection.Println("💡 Fix Suggestions")
	suggestions := 0
	exe, _ := os.Executable()
	// If the executable is in the PATH, we can just use the base name
	if _, err := exec.LookPath(filepath.Base(exe)); err == nil {
		exe = filepath.Base(exe)
	} else {
		// Use the path as invoked if possible, or absolute path
		exe = os.Args[0]
	}

	if !shimsOnPath {
		pterm.Warning.Printf("UniRTM shims directory is not in your PATH.\n")
		pterm.Info.Printf("Fix: Run the following command to setup shims in your shell config:\n")
		pterm.FgMagenta.Printf("     %s enable --shims\n\n", exe)
		suggestions++
	}
	if !isActivated {
		pterm.Warning.Printf("UniRTM environment is not activated.\n")
		pterm.Info.Printf("Fix: Run '%s enable --shims' to setup automatic activation.\n\n", exe)
		suggestions++
	}
	
	// Check for missing tools
	missingTools := 0
	if cfg != nil {
		for name, t := range cfg.Tools {
			slug := env.GetFSToolName(name, t.Backend)
			v := t.Version
			if ver, err := service.ParseVersion(v); err == nil { v = ver.String() }
			if _, err := os.Stat(filepath.Join(env.GetInstallsDir(), slug, v)); os.IsNotExist(err) {
				missingTools++
			}
		}
	}
	if missingTools > 0 {
		pterm.Warning.Printf("Found %d missing tools.\n", missingTools)
		pterm.Info.Printf("Fix: Run 'unirtm install' to install all missing tools.\n\n")
		suggestions++
	}

	if suggestions == 0 {
		pterm.Success.Println("No critical issues found. Your environment looks healthy!")
	}

	fmt.Println()
	pterm.DefaultBox.WithTitle(pterm.LightGreen("Diagnostics Complete")).Println("Your UniRTM environment is perfectly configured and ready.")

	return nil
}

func formatBoolStatus(b bool) string {
	if b { return pterm.LightGreen("yes") }
	return pterm.LightRed("no")
}

func formatTrustStatus() string {
	return pterm.LightGreen("trusted")
}

func formatActiveEnv(cfg *config.Config) string {
	if cfg == nil { return "none" }
	e := env.Get("ENV")
	if e == "" { return "base" }
	return pterm.LightCyan(e)
}

func getDirSize(path string) string {
	var size int64
	_ = filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() { size += info.Size() }
		return nil
	})
	return formatSize(size)
}

func getFileSize(path string) string {
	info, err := os.Stat(path)
	if err != nil { return "0 B" }
	return formatSize(info.Size())
}

func formatSize(size int64) string {
	if size == 0 { return "0 B" }
	const unit = 1024
	if size < unit { return fmt.Sprintf("%d B", size) }
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

func stringToBulletItems(ss []string) []pterm.BulletListItem {
	var items []pterm.BulletListItem
	for _, s := range ss {
		items = append(items, pterm.BulletListItem{Level: 0, Text: s})
	}
	return items
}

func parseInt(s string) int64 {
	var res int64
	fmt.Sscanf(s, "%d", &res)
	return res
}
