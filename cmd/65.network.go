// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/network"
	"github.com/spf13/cobra"
)

var (
	networkGlobal bool
	networkJSON   bool
)

func init() {
	networkCmd.PersistentFlags().BoolVarP(&networkGlobal, "global", "g", false, "Apply to global user configuration")
	networkCmd.PersistentFlags().BoolVar(&networkJSON, "json", false, "Output results in JSON format")

	// Profile subcommands
	networkProfileCmd.AddCommand(networkProfileListCmd)
	networkProfileCmd.AddCommand(networkProfileUseCmd)
	networkProfileCmd.AddCommand(networkProfileCreateCmd)
	networkProfileCmd.AddCommand(networkProfileSaveAsCmd)
	networkProfileCmd.AddCommand(networkProfileResetCmd)
	networkProfileCmd.AddCommand(networkProfileDeleteCmd)

	// Registry subcommands
	networkRegistryCmd.AddCommand(networkRegistryListCmd)
	networkRegistryCmd.AddCommand(networkRegistrySetCmd)
	networkRegistryCmd.AddCommand(networkRegistryAddCmd)
	networkRegistryCmd.AddCommand(networkRegistryDeleteCmd)
	networkRegistryCmd.AddCommand(networkRegistryResetCmd)
	networkRegistryCmd.AddCommand(networkRegistryResetAllCmd)

	// Proxy subcommands
	networkProxyCmd.AddCommand(networkProxyStatusCmd)
	networkProxyCmd.AddCommand(networkProxyOnCmd)
	networkProxyCmd.AddCommand(networkProxyOffCmd)
	networkProxySetCmd.Flags().StringVar(&proxyFlagHTTP, "http", "", "HTTP proxy server address")
	networkProxySetCmd.Flags().StringVar(&proxyFlagHTTPS, "https", "", "HTTPS proxy server address")
	networkProxySetCmd.Flags().StringVar(&proxyFlagSOCKS5, "socks5", "", "SOCKS5 proxy server address")
	networkProxySetCmd.Flags().StringVar(&proxyFlagNoProxy, "no-proxy", "", "Direct connection bypass list (comma-separated)")
	networkProxySetCmd.Flags().BoolVar(&proxyFlagOff, "off", false, "Save configuration and disable proxy")
	networkProxyCmd.AddCommand(networkProxySetCmd)

	// Add subcommands to networkCmd
	networkCmd.AddCommand(networkStatusCmd)
	networkCmd.AddCommand(networkProfileCmd)
	networkCmd.AddCommand(networkRegistryCmd)
	networkCmd.AddCommand(networkProxyCmd)
	networkCmd.AddCommand(networkTestCmd)

	if rootCmd != nil {
		rootCmd.AddCommand(networkCmd)
	}
}

func getNetworkManager() (*network.Manager, string, error) {
	globalPath := config.GetGlobalConfigPath()
	localPath := ".unirtm.toml"

	if networkGlobal {
		mgr := network.NewManager(nil)
		_ = mgr.LoadFromPath(globalPath)
		return mgr, globalPath, nil
	}

	mgr := network.NewManager(nil)
	hasLocal, _ := mgr.LoadCascading(localPath, globalPath)
	targetPath := localPath
	if _, err := os.Stat(localPath); err != nil && !hasLocal {
		targetPath = globalPath
	}
	return mgr, targetPath, nil
}

func resolveNetworkConfigFile(global bool) string {
	if global {
		return config.GetGlobalConfigPath()
	}
	localPath := ".unirtm.toml"
	if _, err := os.Stat(localPath); err == nil {
		return localPath
	}
	return config.GetGlobalConfigPath()
}

// networkCmd is the parent command for network and registry management.
var networkCmd = &cobra.Command{
	Use:     "network",
	Aliases: []string{"net"},
	Short:   "Manage multi-language network environments, mirror matrices, and proxy configurations",
	Long: `Manage mirror sources and local proxies for UniRTM multi-language ecosystems (Node.js, Go, Python, Rust, Ruby, PHP, Java, etc.).

Solves network dependency fetching and environment readiness before compilation and build. Supports one-click profile switching, single-language overrides, and real-time latency probing.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runNetworkStatus(cmd, args)
	},
}

// networkStatusCmd displays the active network status.
var networkStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "View active network profile, proxy status, and all ecosystem mirror registries",
	RunE:  runNetworkStatus,
}

func runNetworkStatus(cmd *cobra.Command, args []string) error {
	mgr, path, _ := getNetworkManager()
	status := mgr.GetStatus()

	if networkJSON {
		data := map[string]interface{}{
			"status":      status,
			"recommended": network.RecommendedMirrors(),
			"file_path":   path,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}

	pterm.DefaultHeader.WithFullWidth().Println("UniRTM Network & Registry Configuration Status")
	pterm.Info.Printf("Config file path: %s\n", path)
	pterm.Info.Printf("Active profile: %s (%s)\n", status.ActiveProfile.Name, status.ActiveProfileID)

	// Proxy state
	if status.Proxy.Enabled {
		pterm.Success.Printf("Network proxy: Enabled (HTTP: %s, SOCKS5: %s)\n", status.Proxy.HTTP, status.Proxy.SOCKS5)
		if status.Proxy.NoProxy != "" {
			pterm.Println(pterm.LightCyan(fmt.Sprintf("  Direct bypass list (NO_PROXY): %s", status.Proxy.NoProxy)))
		}
	} else {
		pterm.Warning.Println("Network proxy: Disabled (all traffic uses direct/mirror endpoints)")
	}

	pterm.Println()
	pterm.DefaultSection.Println("Ecosystem Registry Status")

	tableData := pterm.TableData{
		{"Ecosystem", "Active Registry URL", "Source"},
	}

	// Sort ecosystems with standard baseline first, then others alphabetically
	var orderedEcos []string
	seenEco := make(map[string]bool)
	for _, eco := range network.KnownEcosystems {
		if _, ok := status.Effective[eco]; ok {
			orderedEcos = append(orderedEcos, eco)
			seenEco[eco] = true
		}
	}
	var extraEcos []string
	for eco := range status.Effective {
		if !seenEco[eco] {
			extraEcos = append(extraEcos, eco)
		}
	}
	sort.Strings(extraEcos)
	orderedEcos = append(orderedEcos, extraEcos...)

	for _, eco := range orderedEcos {
		eff := status.Effective[eco]
		sourceBadge := "[Preset]"
		if eff.IsOverridden {
			sourceBadge = pterm.LightYellow("[Custom]")
		} else if eff.Source == "default" {
			sourceBadge = pterm.Gray("[Official Default]")
		} else {
			sourceBadge = pterm.LightGreen("[Preset]")
		}
		tableData = append(tableData, []string{strings.ToUpper(eco), eff.URL, sourceBadge})
	}

	_ = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
	return nil
}

// networkProfileCmd manages network profiles.
var networkProfileCmd = &cobra.Command{
	Use:     "profile",
	Aliases: []string{"prof"},
	Short:   "Manage network profiles",
}

var networkProfileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available network profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, _, _ := getNetworkManager()
		status := mgr.GetStatus()

		tableData := pterm.TableData{
			{"Active", "Profile ID", "Name", "Type", "Description"},
		}

		var profileIDs []string
		for id := range status.Profiles {
			profileIDs = append(profileIDs, id)
		}
		sort.Slice(profileIDs, func(i, j int) bool {
			idA, idB := profileIDs[i], profileIDs[j]
			if idA == "official" {
				return true
			}
			if idB == "official" {
				return false
			}
			if idA == "cn" {
				return true
			}
			if idB == "cn" {
				return false
			}
			return idA < idB
		})

		for _, id := range profileIDs {
			p := status.Profiles[id]
			activeMark := " "
			if id == status.ActiveProfileID {
				activeMark = pterm.Green("✔")
			}
			typeStr := "Custom"
			if p.Builtin {
				typeStr = pterm.LightBlue("Builtin")
			}
			tableData = append(tableData, []string{activeMark, id, p.Name, typeStr, p.Description})
		}

		return pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
	},
}

var networkProfileUseCmd = &cobra.Command{
	Use:   "use <profile_id>",
	Short: "Switch the active network profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, path, _ := getNetworkManager()
		id := args[0]
		if err := mgr.UseProfile(id); err != nil {
			return err
		}
		if err := mgr.SaveToPath(path); err != nil {
			return err
		}
		pterm.Success.Printf("Switched active network profile to: %s\n", id)
		return nil
	},
}

var networkProfileCreateCmd = &cobra.Command{
	Use:   "create <profile_id>",
	Short: "Create a new custom network profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, path, _ := getNetworkManager()
		id := args[0]
		p := network.NetworkProfile{
			ID:          id,
			Name:        id,
			Description: "User custom profile",
		}
		if err := mgr.CreateProfile(p); err != nil {
			return err
		}
		if err := mgr.SaveToPath(path); err != nil {
			return err
		}
		pterm.Success.Printf("Created custom network profile: %s\n", id)
		return nil
	},
}

var networkProfileSaveAsCmd = &cobra.Command{
	Use:   "save-as <new_profile_id>",
	Short: "Save current network and registry state as a new profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, path, _ := getNetworkManager()
		newID := args[0]
		if err := mgr.SaveAsProfile(newID, newID, "Saved based on current custom state"); err != nil {
			return err
		}
		if err := mgr.SaveToPath(path); err != nil {
			return err
		}
		pterm.Success.Printf("Saved current state and activated new profile: %s\n", newID)
		return nil
	},
}

var networkProfileResetCmd = &cobra.Command{
	Use:   "reset <profile_id>",
	Short: "Reset builtin network profile to factory defaults",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, path, _ := getNetworkManager()
		id := args[0]
		if err := mgr.ResetProfile(id); err != nil {
			return err
		}
		if err := mgr.SaveToPath(path); err != nil {
			return err
		}
		pterm.Success.Printf("Reset profile %s to factory default registry configuration\n", id)
		return nil
	},
}

var networkProfileDeleteCmd = &cobra.Command{
	Use:   "delete <profile_id>",
	Short: "Delete a specified custom network profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, path, _ := getNetworkManager()
		id := args[0]
		if err := mgr.DeleteProfile(id); err != nil {
			return err
		}
		if err := mgr.SaveToPath(path); err != nil {
			return err
		}
		pterm.Success.Printf("Deleted custom profile: %s\n", id)
		return nil
	},
}

// networkRegistryCmd manages individual ecosystem registry overrides.
var networkRegistryCmd = &cobra.Command{
	Use:     "registry",
	Aliases: []string{"reg"},
	Short:   "Manage single ecosystem registry overrides",
}

var networkRegistryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recommended public mirrors",
	RunE: func(cmd *cobra.Command, args []string) error {
		rec := network.RecommendedMirrors()
		for _, eco := range network.KnownEcosystems {
			list, ok := rec[eco]
			if !ok {
				continue
			}
			pterm.DefaultSection.Printf("%s Recommended Mirrors", strings.ToUpper(eco))
			for _, opt := range list {
				pterm.Println(fmt.Sprintf("  • %-28s %s", opt.Name, pterm.LightCyan(opt.URL)))
			}
			pterm.Println()
		}
		return nil
	},
}

var networkRegistrySetCmd = &cobra.Command{
	Use:   "set <ecosystem> <url>",
	Short: "Override registry for a single ecosystem (higher priority than Profile)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, path, _ := getNetworkManager()
		eco := args[0]
		url := args[1]
		if err := mgr.SetOverride(eco, url); err != nil {
			return err
		}
		if err := mgr.SaveToPath(path); err != nil {
			return err
		}
		pterm.Success.Printf("Set single override registry for %s: %s\n", strings.ToUpper(eco), url)
		return nil
	},
}

var networkRegistryAddCmd = &cobra.Command{
	Use:   "add <ecosystem> <url>",
	Short: "Add or restore ecosystem mirror registry",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, path, _ := getNetworkManager()
		eco := args[0]
		url := args[1]
		if err := mgr.AddEcosystem(eco, url); err != nil {
			return err
		}
		if err := mgr.SaveToPath(path); err != nil {
			return err
		}
		pterm.Success.Printf("Added registry for ecosystem %s: %s\n", strings.ToUpper(eco), url)
		return nil
	},
}

var networkRegistryDeleteCmd = &cobra.Command{
	Use:     "delete <ecosystem>",
	Aliases: []string{"del", "remove", "rm"},
	Short:   "Remove or hide an ecosystem from the mirror matrix",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, path, _ := getNetworkManager()
		eco := args[0]
		if err := mgr.DeleteEcosystem(eco); err != nil {
			return err
		}
		if err := mgr.SaveToPath(path); err != nil {
			return err
		}
		pterm.Success.Printf("Removed ecosystem from mirror matrix: %s\n", strings.ToUpper(eco))
		return nil
	},
}

var networkRegistryResetCmd = &cobra.Command{
	Use:   "reset <ecosystem>",
	Short: "Clear single ecosystem override to follow current profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, path, _ := getNetworkManager()
		eco := args[0]
		if err := mgr.ResetOverride(eco); err != nil {
			return err
		}
		if err := mgr.SaveToPath(path); err != nil {
			return err
		}
		pterm.Success.Printf("Cleared single override for %s, now following active profile\n", strings.ToUpper(eco))
		return nil
	},
}

var networkRegistryResetAllCmd = &cobra.Command{
	Use:     "reset-all",
	Aliases: []string{"restore-defaults", "reset-matrix"},
	Short:   "Reset and restore all 13 standard ecosystem default matrices",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, path, _ := getNetworkManager()
		if err := mgr.ResetDefaultEcosystems(); err != nil {
			return err
		}
		if err := mgr.SaveToPath(path); err != nil {
			return err
		}
		pterm.Success.Println("Restored all 13 standard ecosystem default matrices and builtin presets")
		return nil
	},
}

// networkProxyCmd manages proxy settings.
var networkProxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Configure or toggle global network proxy",
}

var networkProxyStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "View current network proxy configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, _, _ := getNetworkManager()
		p := mgr.Config().Proxy
		if p.Enabled {
			pterm.Success.Println("Network proxy status: Enabled")
			pterm.Println(fmt.Sprintf("  HTTP Proxy:   %s", p.HTTP))
			pterm.Println(fmt.Sprintf("  HTTPS Proxy:  %s", p.HTTPS))
			pterm.Println(fmt.Sprintf("  SOCKS5 Proxy: %s", p.SOCKS5))
			pterm.Println(fmt.Sprintf("  Bypass List:  %s", p.NoProxy))
		} else {
			pterm.Info.Println("Network proxy status: Disabled (Direct Mode)")
			if p.HTTP != "" {
				pterm.Println(fmt.Sprintf("  Configured HTTP Proxy (Inactive): %s", p.HTTP))
			}
		}
		return nil
	},
}

var networkProxyOnCmd = &cobra.Command{
	Use:   "on",
	Short: "Enable network proxy",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, path, _ := getNetworkManager()
		p := mgr.Config().Proxy
		if p.HTTP == "" && p.SOCKS5 == "" {
			return fmt.Errorf("Proxy address not configured yet, please run: unirtm network proxy set http://127.0.0.1:7890")
		}
		p.Enabled = true
		_ = mgr.SetProxy(p)
		if err := mgr.SaveToPath(path); err != nil {
			return err
		}
		pterm.Success.Println("Network proxy enabled")
		return nil
	},
}

var networkProxyOffCmd = &cobra.Command{
	Use:   "off",
	Short: "Disable network proxy",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, path, _ := getNetworkManager()
		p := mgr.Config().Proxy
		p.Enabled = false
		_ = mgr.SetProxy(p)
		if err := mgr.SaveToPath(path); err != nil {
			return err
		}
		pterm.Success.Println("Network proxy disabled")
		return nil
	},
}

var (
	proxyFlagHTTP    string
	proxyFlagHTTPS   string
	proxyFlagSOCKS5  string
	proxyFlagNoProxy string
	proxyFlagOff     bool
)

var networkProxySetCmd = &cobra.Command{
	Use:   "set [proxy_url]",
	Short: "Set proxy server address (e.g. http://127.0.0.1:7890 or socks5://127.0.0.1:10808)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, path, _ := getNetworkManager()
		p := mgr.Config().Proxy

		if len(args) > 0 {
			rawURL := args[0]
			if strings.HasPrefix(rawURL, "socks5://") || strings.HasPrefix(rawURL, "socks5h://") {
				p.SOCKS5 = rawURL
			} else {
				p.HTTP = rawURL
				p.HTTPS = rawURL
			}
		}

		if cmd.Flags().Changed("http") {
			p.HTTP = proxyFlagHTTP
		}
		if cmd.Flags().Changed("https") {
			p.HTTPS = proxyFlagHTTPS
		}
		if cmd.Flags().Changed("socks5") {
			p.SOCKS5 = proxyFlagSOCKS5
		}
		if cmd.Flags().Changed("no-proxy") {
			p.NoProxy = proxyFlagNoProxy
		}

		if proxyFlagOff {
			p.Enabled = false
		} else if len(args) > 0 || cmd.Flags().Changed("http") || cmd.Flags().Changed("https") || cmd.Flags().Changed("socks5") {
			p.Enabled = true
		}

		if err := mgr.SetProxy(p); err != nil {
			return err
		}
		if err := mgr.SaveToPath(path); err != nil {
			return err
		}
		if p.Enabled {
			pterm.Success.Println("Network proxy configuration updated and enabled")
		} else {
			pterm.Success.Println("Network proxy configuration updated (kept disabled)")
		}
		return nil
	},
}

// networkTestCmd executes latency test against probe targets.
var networkTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test round-trip latency to official registries and public mirrors",
	RunE: func(cmd *cobra.Command, args []string) error {
		spinner, _ := pterm.DefaultSpinner.Start("Concurrently probing network latency to official registries and public mirrors...")
		results := network.ProbeAll(network.DefaultProbeTargets(), 2500*time.Millisecond)
		spinner.Success("Network latency probing complete")

		tableData := pterm.TableData{
			{"Ecosystem", "Mirror Node Name", "Test Endpoint", "Latency", "Rating"},
		}

		for _, r := range results {
			rttStr := fmt.Sprintf("%d ms", r.RTTMs)
			rating := ""
			switch r.Status {
			case "fast":
				rating = pterm.Green("Fast (<100ms)")
			case "normal":
				rating = pterm.LightCyan("Good (<300ms)")
			case "slow":
				rating = pterm.Yellow("Slow (>300ms)")
			default:
				rttStr = "TIMEOUT"
				rating = pterm.Red("Timeout / Unreachable")
			}
			tableData = append(tableData, []string{strings.ToUpper(r.Ecosystem), r.Name, r.Endpoint, rttStr, rating})
		}

		_ = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
		return nil
	},
}
