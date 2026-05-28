// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"os"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/spf13/cobra"

	"github.com/snowdreamtech/unirtm/internal/cli/output"
)

// untrustCmd represents the untrust command
var untrustCmd = &cobra.Command{
	Use:   "untrust [path]",
	Short: "Revoke trust from a configuration file",
	Long: `Revokes trust from a previously trusted configuration file.
Once untrusted, the file's environment variables and configuration will no longer be automatically loaded.`,
	Run: func(cmd *cobra.Command, args []string) {
		path := resolveConfigFilePath(false)
		if len(args) > 0 {
			path = args[0]
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			absPath = path
		}

		trustManager := config.NewTrustManager()
		if err := trustManager.Untrust(absPath); err != nil {
			output.Errorf("Failed to untrust configuration file: %v", err)
			os.Exit(1)
		}

		output.Successf("Untrusted configuration file: %s", pterm.LightGreen(absPath))
	},
}

func init() {
	rootCmd.AddCommand(untrustCmd)
}
