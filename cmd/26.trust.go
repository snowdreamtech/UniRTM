// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"

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
If no path is provided, it lists all currently trusted configuration files.`,
	Run: func(cmd *cobra.Command, args []string) {
		trustManager := config.NewTrustManager()

		if len(args) == 0 {
			// List all trusted files
			trusted, err := trustManager.List()
			if err != nil {
				pterm.Error.Printfln("Failed to list trusted files: %v", err)
				os.Exit(1)
			}
			if len(trusted) == 0 {
				pterm.Info.Println("No trusted configuration files found.")
				return
			}
			pterm.DefaultSection.Println("Trusted Configuration Files")
			tableData := pterm.TableData{
				{"Configuration File Path", "SHA-256 Content Hash", "Status"},
			}
			for p, h := range trusted {
				status := trustManager.TrustStatus(p)
				statusStr := ""
				switch status {
				case config.TrustStatusTrusted:
					statusStr = pterm.FgGreen.Sprint("Trusted")
				case config.TrustStatusModified:
					statusStr = pterm.FgRed.Sprint("Modified")
				case config.TrustStatusUntrusted:
					statusStr = pterm.FgYellow.Sprint("Untrusted")
				}

				hashStr := h
				if hashStr == "" {
					hashStr = pterm.FgYellow.Sprint("Legacy / No Hash")
				} else {
					if len(hashStr) > 16 {
						hashStr = hashStr[:16] + "..."
					}
					hashStr = pterm.FgGray.Sprint(hashStr)
				}
				tableData = append(tableData, []string{pterm.FgCyan.Sprint(p), hashStr, statusStr})
			}
			pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
			return
		}

		path := args[0]
		absPath, err := filepath.Abs(path)
		if err != nil {
			absPath = path
		}

		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			pterm.Error.Printfln("Configuration file not found: %s", absPath)
			os.Exit(1)
		}

		if err := trustManager.Trust(absPath); err != nil {
			pterm.Error.Printfln("Failed to trust configuration file: %v", err)
			os.Exit(1)
		}

		// Calculate SHA-256 hash to display
		hash := ""
		if data, err := os.ReadFile(absPath); err == nil {
			h := sha256.Sum256(data)
			hash = hex.EncodeToString(h[:])
		}
		if len(hash) > 16 {
			hash = hash[:16] + "..."
		}
		pterm.Success.Printfln("Trusted configuration file: %s (hash: %s)", pterm.LightGreen(absPath), pterm.FgGray.Sprint(hash))
	},
}

func init() {
	rootCmd.AddCommand(trustCmd)
}
