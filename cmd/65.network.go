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
	networkCmd.PersistentFlags().BoolVarP(&networkGlobal, "global", "g", false, "作用于全局用户配置")
	networkCmd.PersistentFlags().BoolVar(&networkJSON, "json", false, "以 JSON 格式输出结果")

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

	// Proxy subcommands
	networkProxyCmd.AddCommand(networkProxyStatusCmd)
	networkProxyCmd.AddCommand(networkProxyOnCmd)
	networkProxyCmd.AddCommand(networkProxyOffCmd)
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
	path := resolveNetworkConfigFile(networkGlobal)
	mgr := network.NewManager(nil)
	_ = mgr.LoadFromPath(path)
	return mgr, path, nil
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
	Short:   "管理多语言编译前网络环境、镜像矩阵与代理配置",
	Long: `管理 UniRTM 多语言生态（Node.js, Go, Python, Rust, Ruby）的国内镜像矩阵与本地网络代理。

解决项目在编译构建前的网络拉取与环境就绪问题。支持 Profile 档案一键切换、单语言覆盖以及实时测速。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runNetworkStatus(cmd, args)
	},
}

// networkStatusCmd displays the active network status.
var networkStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看当前生效的网络档案、代理与全生态镜像源",
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

	pterm.DefaultHeader.WithFullWidth().Println("UniRTM 网络与镜像源配置状态")
	pterm.Info.Printf("配置文件路径: %s\n", path)
	pterm.Info.Printf("当前生效档案: %s (%s)\n", status.ActiveProfile.Name, status.ActiveProfileID)

	// Proxy state
	if status.Proxy.Enabled {
		pterm.Success.Printf("网络代理: 已启用 (HTTP: %s, SOCKS5: %s)\n", status.Proxy.HTTP, status.Proxy.SOCKS5)
		if status.Proxy.NoProxy != "" {
			pterm.Println(pterm.LightCyan(fmt.Sprintf("  直连白名单 (NO_PROXY): %s", status.Proxy.NoProxy)))
		}
	} else {
		pterm.Warning.Println("网络代理: 未启用 (所有流量使用官方直连/镜像源)")
	}

	pterm.Println()
	pterm.DefaultSection.Println("多语言生态镜像源生效状态")

	tableData := pterm.TableData{
		{"生态", "生效镜像源地址", "来源"},
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
		sourceBadge := "[继承自预设]"
		if eff.IsOverridden {
			sourceBadge = pterm.LightYellow("[单项自定义覆盖]")
		} else {
			sourceBadge = pterm.LightGreen("[跟随 " + status.ActiveProfileID + " 预设]")
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
	Short:   "管理网络预设档案 (Profile)",
}

var networkProfileListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有可用网络档案",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, _, _ := getNetworkManager()
		status := mgr.GetStatus()

		tableData := pterm.TableData{
			{"激活", "档案标识", "名称", "类型", "描述"},
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
			typeStr := "自定义"
			if p.Builtin {
				typeStr = pterm.LightBlue("内置出厂")
			}
			tableData = append(tableData, []string{activeMark, id, p.Name, typeStr, p.Description})
		}

		return pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
	},
}

var networkProfileUseCmd = &cobra.Command{
	Use:   "use <profile_id>",
	Short: "切换当前激活的网络档案",
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
		pterm.Success.Printf("已切换当前网络档案为: %s\n", id)
		return nil
	},
}

var networkProfileCreateCmd = &cobra.Command{
	Use:   "create <profile_id>",
	Short: "创建新的自定义网络档案",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, path, _ := getNetworkManager()
		id := args[0]
		p := network.NetworkProfile{
			ID:          id,
			Name:        id,
			Description: "用户自定义档案",
		}
		if err := mgr.CreateProfile(p); err != nil {
			return err
		}
		if err := mgr.SaveToPath(path); err != nil {
			return err
		}
		pterm.Success.Printf("已创建自定义网络档案: %s\n", id)
		return nil
	},
}

var networkProfileSaveAsCmd = &cobra.Command{
	Use:   "save-as <new_profile_id>",
	Short: "将当前生效的网络与镜像状态另存为新档案",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, path, _ := getNetworkManager()
		newID := args[0]
		if err := mgr.SaveAsProfile(newID, newID, "基于当前自定义状态保存"); err != nil {
			return err
		}
		if err := mgr.SaveToPath(path); err != nil {
			return err
		}
		pterm.Success.Printf("已将当前状态另存并激活新档案: %s\n", newID)
		return nil
	},
}

var networkProfileResetCmd = &cobra.Command{
	Use:   "reset <profile_id>",
	Short: "将内置网络档案恢复为出厂默认值",
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
		pterm.Success.Printf("已将档案 %s 恢复为原厂默认镜像配置\n", id)
		return nil
	},
}

var networkProfileDeleteCmd = &cobra.Command{
	Use:   "delete <profile_id>",
	Short: "删除指定的自定义网络档案",
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
		pterm.Success.Printf("已删除自定义档案: %s\n", id)
		return nil
	},
}

// networkRegistryCmd manages individual ecosystem registry overrides.
var networkRegistryCmd = &cobra.Command{
	Use:     "registry",
	Aliases: []string{"reg"},
	Short:   "管理单语言生态的独立镜像源覆盖 (Overrides)",
}

var networkRegistryListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出推荐公共镜像列表",
	RunE: func(cmd *cobra.Command, args []string) error {
		rec := network.RecommendedMirrors()
		for _, eco := range network.KnownEcosystems {
			list, ok := rec[eco]
			if !ok {
				continue
			}
			pterm.DefaultSection.Printf("%s 推荐镜像源", strings.ToUpper(eco))
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
	Short: "单独覆盖某语言生态的镜像源 (优先级高于 Profile)",
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
		pterm.Success.Printf("已设置 %s 单项覆盖镜像源: %s\n", strings.ToUpper(eco), url)
		return nil
	},
}

var networkRegistryAddCmd = &cobra.Command{
	Use:   "add <ecosystem> <url>",
	Short: "添加或恢复语言生态镜像源",
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
		pterm.Success.Printf("已添加语言生态 %s 镜像源: %s\n", strings.ToUpper(eco), url)
		return nil
	},
}

var networkRegistryDeleteCmd = &cobra.Command{
	Use:     "delete <ecosystem>",
	Aliases: []string{"del", "remove", "rm"},
	Short:   "从镜像源矩阵中删除或隐藏某语言生态",
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
		pterm.Success.Printf("已从镜像矩阵中移除语言生态: %s\n", strings.ToUpper(eco))
		return nil
	},
}

var networkRegistryResetCmd = &cobra.Command{
	Use:   "reset <ecosystem>",
	Short: "清除某语言生态的单项覆盖，恢复跟随当前 Profile",
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
		pterm.Success.Printf("已清除 %s 单项覆盖，恢复跟随当前预设档案\n", strings.ToUpper(eco))
		return nil
	},
}

// networkProxyCmd manages proxy settings.
var networkProxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "配置或切换全局网络代理",
}

var networkProxyStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看当前网络代理配置",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, _, _ := getNetworkManager()
		p := mgr.Config().Proxy
		if p.Enabled {
			pterm.Success.Println("网络代理状态: 已开启")
			pterm.Println(fmt.Sprintf("  HTTP 代理:   %s", p.HTTP))
			pterm.Println(fmt.Sprintf("  HTTPS 代理:  %s", p.HTTPS))
			pterm.Println(fmt.Sprintf("  SOCKS5 代理: %s", p.SOCKS5))
			pterm.Println(fmt.Sprintf("  直连白名单:  %s", p.NoProxy))
		} else {
			pterm.Info.Println("网络代理状态: 已关闭 (直连模式)")
			if p.HTTP != "" {
				pterm.Println(fmt.Sprintf("  已配置 HTTP 代理 (未激活): %s", p.HTTP))
			}
		}
		return nil
	},
}

var networkProxyOnCmd = &cobra.Command{
	Use:   "on",
	Short: "启用网络代理",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, path, _ := getNetworkManager()
		p := mgr.Config().Proxy
		if p.HTTP == "" && p.SOCKS5 == "" {
			return fmt.Errorf("代理地址尚未配置，请先运行: unirtm network proxy set http://127.0.0.1:7890")
		}
		p.Enabled = true
		_ = mgr.SetProxy(p)
		if err := mgr.SaveToPath(path); err != nil {
			return err
		}
		pterm.Success.Println("网络代理已启用")
		return nil
	},
}

var networkProxyOffCmd = &cobra.Command{
	Use:   "off",
	Short: "关闭网络代理",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, path, _ := getNetworkManager()
		p := mgr.Config().Proxy
		p.Enabled = false
		_ = mgr.SetProxy(p)
		if err := mgr.SaveToPath(path); err != nil {
			return err
		}
		pterm.Success.Println("网络代理已关闭")
		return nil
	},
}

var networkProxySetCmd = &cobra.Command{
	Use:   "set <proxy_url>",
	Short: "设置代理服务器地址 (例如 http://127.0.0.1:7890 或 socks5://127.0.0.1:10808)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, path, _ := getNetworkManager()
		rawURL := args[0]
		p := mgr.Config().Proxy
		p.Enabled = true

		if strings.HasPrefix(rawURL, "socks5://") || strings.HasPrefix(rawURL, "socks5h://") {
			p.SOCKS5 = rawURL
		} else {
			p.HTTP = rawURL
			p.HTTPS = rawURL
		}

		if err := mgr.SetProxy(p); err != nil {
			return err
		}
		if err := mgr.SaveToPath(path); err != nil {
			return err
		}
		pterm.Success.Printf("代理服务器已配置并启用: %s\n", rawURL)
		return nil
	},
}

// networkTestCmd executes latency test against probe targets.
var networkTestCmd = &cobra.Command{
	Use:   "test",
	Short: "测试各生态官方源与主流国内镜像源的往返延迟",
	RunE: func(cmd *cobra.Command, args []string) error {
		spinner, _ := pterm.DefaultSpinner.Start("正在并发探测各大官方源与国内镜像站网络延迟...")
		results := network.ProbeAll(network.DefaultProbeTargets(), 2500*time.Millisecond)
		spinner.Success("网络延迟探测完成")

		tableData := pterm.TableData{
			{"生态", "镜像节点名称", "测试端点", "往返延迟", "状态评级"},
		}

		for _, r := range results {
			rttStr := fmt.Sprintf("%d ms", r.RTTMs)
			rating := ""
			switch r.Status {
			case "fast":
				rating = pterm.Green("极速 (<100ms)")
			case "normal":
				rating = pterm.LightCyan("良好 (<300ms)")
			case "slow":
				rating = pterm.Yellow("缓慢 (>300ms)")
			default:
				rttStr = "TIMEOUT"
				rating = pterm.Red("超时/无法连通")
			}
			tableData = append(tableData, []string{strings.ToUpper(r.Ecosystem), r.Name, r.Endpoint, rttStr, rating})
		}

		_ = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
		return nil
	},
}
