// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"os"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/spf13/cobra"
)

// untrustCmd represents the untrust command
var untrustCmd = &cobra.Command{
	Use:   "untrust [path]",
	Short: "Revoke trust from a configuration file",
	Long: `Revokes trust from a previously trusted configuration file.
Once untrusted, the file's environment variables and configuration will no longer be automatically loaded.
If no path is provided, it defaults to ./unirtm.toml in the current directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		path := "./unirtm.toml"
		if len(args) > 0 {
			path = args[0]
		}

		trustManager := config.NewTrustManager()
		if err := trustManager.Untrust(path); err != nil {
			pterm.Error.Printfln("Failed to untrust configuration file: %v", err)
			os.Exit(1)
		}

		pterm.Success.Printfln("Untrusted configuration file: %s", path)
	},
}

func init() {
	rootCmd.AddCommand(untrustCmd)
}
