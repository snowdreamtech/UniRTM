// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"os"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/spf13/cobra"
)

// trustCmd represents the trust command
var trustCmd = &cobra.Command{
	Use:   "trust [path]",
	Short: "Mark a configuration file as trusted",
	Long: `Marks a configuration file (like unirtm.toml) as trusted.
Trusted files are allowed to be automatically loaded and their environment variables applied.
If no path is provided, it defaults to ./unirtm.toml in the current directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		path := resolveConfigFilePath(false)
		if len(args) > 0 {
			path = args[0]
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			pterm.Error.Printfln("Configuration file not found: %s", path)
			os.Exit(1)
		}

		trustManager := config.NewTrustManager()
		if err := trustManager.Trust(path); err != nil {
			pterm.Error.Printfln("Failed to trust configuration file: %v", err)
			os.Exit(1)
		}

		pterm.FgGreen.Printfln("✅ Trusted configuration file: %s", path)
	},
}

func init() {
	rootCmd.AddCommand(trustCmd)
}
