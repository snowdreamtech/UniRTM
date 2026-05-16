package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/spf13/cobra"
)

var (
	implodeYes    bool
	implodeConfig bool
)

func init() {
	implodeCmd.Flags().BoolVarP(&implodeYes, "yes", "y", false, "skip confirmation prompt")
	implodeCmd.Flags().BoolVar(&implodeConfig, "config", false, "also remove configuration directory (~/.config/unirtm)")
	
	if rootCmd != nil {
		rootCmd.AddCommand(implodeCmd)
	}
}

// implodeCmd removes all UniRTM data, shims, cache, and database.
var implodeCmd = &cobra.Command{
	Use:   "implode",
	Short: "Completely remove all UniRTM data and tool installations",
	Long: `Completely remove all UniRTM data and tool installations.

This command will internal-combust and erase:
  • All tool installations (binaries)
  • All shims and wrapper scripts
  • All download caches and temporary files
  • The central SQLite database
  • All external plugins
  • (Optional) Your configuration directory (~/.config/unirtm)

WARNING: This action is permanent and IRREVERSIBLE.`,
	Args: cobra.NoArgs,
	RunE: runImplode,
}

func runImplode(cmd *cobra.Command, args []string) error {
	// 1. Visual Banner
	pterm.DefaultCenter.Println(pterm.Red("!!! DANGER ZONE !!!"))
	pterm.DefaultBigText.WithLetters(
		pterm.NewLettersFromStringWithStyle("IM", pterm.NewStyle(pterm.FgRed)),
		pterm.NewLettersFromStringWithStyle("PLODE", pterm.NewStyle(pterm.FgWhite)),
	).Render()

	configDir := env.GetConfigDir()

	targets := []struct {
		name string
		path string
	}{
		{"Tool Installations", env.GetInstallsDir()},
		{"Shim Wrappers", env.GetShimsDir()},
		{"Download Cache", env.GetDownloadsDir()},
		{"Tool Database", env.GetDatabasePath()},
		{"External Plugins", env.GetPluginsDir()},
	}

	if implodeConfig {
		targets = append(targets, struct {
			name string
			path string
		}{"Configuration Files", configDir})
	}

	// 2. Confirmation
	if !implodeYes {
		pterm.Warning.Prefix = pterm.Prefix{Text: "WARNING", Style: pterm.NewStyle(pterm.BgRed, pterm.FgWhite)}
		pterm.Warning.Println("This will permanently destroy ALL UniRTM data and tools.")
		fmt.Printf("\nSelected Targets:\n")
		for _, t := range targets {
			pterm.BulletListPrinter{}.WithItems([]pterm.BulletListItem{
				{Level: 0, Text: fmt.Sprintf("%s (%s)", pterm.Bold.Sprint(t.name), t.path), Bullet: "•", BulletStyle: pterm.NewStyle(pterm.FgRed)},
			}).Render()
		}

		fmt.Print("\n" + pterm.LightRed("Type 'yes' to proceed with self-destruction: "))

		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "yes" {
			pterm.Info.Println("Implode aborted. You live to fight another day.")
			return nil
		}

		// 3. Countdown
		fmt.Println()
		for i := 3; i > 0; i-- {
			pterm.Info.Printf("Initiating sequence in %d...\r", i)
			time.Sleep(1 * time.Second)
		}
		fmt.Println()
	}

	// 4. Execution
	pterm.Println(pterm.LightRed("Self-destruct sequence active..."))
	fmt.Println()

	multi := pterm.DefaultMultiPrinter
	for _, t := range targets {
		spinner, _ := pterm.DefaultSpinner.WithWriter(multi.NewWriter()).Start("Destroying " + t.name + "...")
		
		if _, err := os.Stat(t.path); os.IsNotExist(err) {
			spinner.Info("Skipped (Already gone)")
			continue
		}

		if err := os.RemoveAll(t.path); err != nil {
			spinner.Fail(fmt.Sprintf("Failed to remove %s: %v", t.name, err))
		} else {
			spinner.Success("Erased: " + t.name)
		}
	}
	multi.Start()

	// 5. Final Message
	fmt.Println()
	pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgRed)).WithTextStyle(pterm.NewStyle(pterm.FgWhite)).Println("IMPLODE COMPLETE")
	fmt.Println()
	
	pterm.Info.Println("To complete the cleanup, you may want to:")
	pterm.BulletListPrinter{}.WithItems([]pterm.BulletListItem{
		{Level: 0, Text: "Remove the 'unirtm' binary from your PATH."},
		{Level: 0, Text: "Remove UniRTM activation code from your .zshrc/.bashrc if present."},
		{Level: 0, Text: "Say goodbye to your tools. They are gone forever."},
	}).Render()

	return nil
}
