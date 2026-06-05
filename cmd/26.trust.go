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
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/spf13/cobra"

	"github.com/snowdreamtech/unirtm/internal/cli/output"
)

// trustCmd represents the trust command
var (
	trustList bool
	trustAll  bool
)

var trustCmd = &cobra.Command{
	Use:   "trust [path]",
	Short: "Mark a configuration file as trusted",
	Long: `Marks all project-related configuration files as trusted.
Trusted files are allowed to be automatically loaded and their environment variables applied.
If no path is provided, it automatically trusts all configuration files in the current directory.

Use --all (-a) to also trust the global configuration file (~/.config/unirtm/unirtm.toml).
Use --list (-l) to view the current project's trusted configuration files.
Use --list --all (-la) to view all globally trusted configuration files.`,
	Run: func(cmd *cobra.Command, args []string) {
		trustManager := config.NewTrustManager()

		if trustList {
			if trustAll {
				// -la: show all globally trusted files
				trusted, err := trustManager.List()
				if err != nil {
					output.Errorf("Failed to list trusted files: %v", err)
					os.Exit(1)
				}
				if len(trusted) == 0 {
					output.Info("No trusted configuration files found.")
					return
				}
				pterm.DefaultSection.Println("Trusted Configuration Files")
				printTrustTable(trustManager, trusted)
			} else {
				// -l: show trusted config files for the current project
				dir, err := os.Getwd()
				if err != nil {
					dir = "."
				}
				if len(args) > 0 {
					if absArg, err := filepath.Abs(args[0]); err == nil {
						dir = absArg
					} else {
						dir = args[0]
					}
				}
				projectPaths := findAllProjectConfigFiles(dir)
				trusted, err := trustManager.List()
				if err != nil {
					output.Errorf("Failed to list trusted files: %v", err)
					os.Exit(1)
				}
				current := map[string]string{}
				for _, p := range projectPaths {
					if h, ok := trusted[p]; ok {
						current[p] = h
					}
				}
				if len(current) == 0 {
					output.Infof("No trusted configuration files found for current project in: %s", dir)
					return
				}
				pterm.DefaultSection.Println("Trusted Configuration Files")
				printTrustTable(trustManager, current)
			}
			return
		}

		// Determine target directory
		dir, err := os.Getwd()
		if err != nil {
			dir = "."
		}
		if len(args) > 0 {
			if absArg, err := filepath.Abs(args[0]); err == nil {
				dir = absArg
			} else {
				dir = args[0]
			}
		}

		// Collect files to trust: always include all project config files
		paths := findAllProjectConfigFiles(dir)

		if trustAll {
			// -a: also include the global config file
			globalPath := env.GetGlobalConfigPath()
			if _, err := os.Stat(globalPath); err == nil {
				absGlobal, _ := filepath.Abs(globalPath)
				alreadyIncluded := false
				for _, p := range paths {
					if p == absGlobal {
						alreadyIncluded = true
						break
					}
				}
				if !alreadyIncluded {
					paths = append(paths, absGlobal)
				}
			}
		}

		if len(paths) == 0 {
			output.Info("No project configuration files found in the current directory.")
			return
		}

		trustedPaths := map[string]string{}
		for _, p := range paths {
			if err := trustManager.Trust(p); err != nil {
				output.Errorf("Failed to trust %s: %v", p, err)
				continue
			}
			hash := ""
			if data, err := os.ReadFile(p); err == nil {
				h := sha256.Sum256(data)
				hash = hex.EncodeToString(h[:])
			}
			if len(hash) > 16 {
				hash = hash[:16] + "..."
			}
			output.Successf("Trusted configuration file: %s (hash: %s)", pterm.LightGreen(p), pterm.FgGray.Sprint(hash))
			if trusted, err := trustManager.List(); err == nil {
				if h, ok := trusted[p]; ok {
					trustedPaths[p] = h
				}
			}
		}

		if len(trustedPaths) > 0 {
			pterm.DefaultSection.Println("Trusted Configuration Files")
			printTrustTable(trustManager, trustedPaths)
		}
	},
}

// findAllProjectConfigFiles returns all unirtm config files found in the given directory.
// Matches: .unirtm.toml, unirtm.toml, .unirtm.*.toml, unirtm.*.toml,
//
//	.unirtm.yaml, unirtm.yaml, .unirtm.*.yaml, unirtm.*.yaml
func findAllProjectConfigFiles(dir string) []string {
	staticNames := []string{
		".unirtm.toml", "unirtm.toml",
		".unirtm.yaml", "unirtm.yaml",
	}
	globPatterns := []string{
		".unirtm.*.toml", "unirtm.*.toml",
		".unirtm.*.yaml", "unirtm.*.yaml",
	}

	found := []string{}
	seen := map[string]bool{}

	for _, name := range staticNames {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err == nil {
			abs, _ := filepath.Abs(p)
			if !seen[abs] {
				found = append(found, abs)
				seen[abs] = true
			}
		}
	}

	for _, pattern := range globPatterns {
		matches, err := filepath.Glob(filepath.Join(dir, pattern))
		if err != nil {
			continue
		}
		for _, m := range matches {
			abs, _ := filepath.Abs(m)
			if !seen[abs] {
				found = append(found, abs)
				seen[abs] = true
			}
		}
	}

	return found
}

// printTrustTable renders a trust status table for the given path→hash map.
func printTrustTable(trustManager config.TrustManager, trusted map[string]string) {
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
}

func init() {
	trustCmd.Flags().BoolVarP(&trustList, "list", "l", false, "list trusted configuration files (current project by default)")
	trustCmd.Flags().BoolVarP(&trustAll, "all", "a", false, "also trust the global config file (~/.config/unirtm/unirtm.toml)")
	rootCmd.AddCommand(trustCmd)
}
