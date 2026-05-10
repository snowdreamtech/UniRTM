// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/spf13/cobra"
)

func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(tokenCmd)
	}
}

// tokenCmd shows configured provider tokens (masked).
var tokenCmd = &cobra.Command{
	Use:   "token [provider]",
	Short: "Show configured provider API tokens (masked)",
	Long: `Show configured provider API tokens (masked).

Displays tokens read from environment variables for each known provider.
Sensitive values are masked — only the last 4 characters are shown.

Known environment variables:
  GITHUB_TOKEN / GH_TOKEN   GitHub API token (github backend)
  GITLAB_TOKEN              GitLab token
  AQUA_GITHUB_TOKEN         Aqua registry GitHub token

Examples:
  # Show all tokens
  unirtm token

  # Show token for a specific provider
  unirtm token github`,
	Args: cobra.MaximumNArgs(1),
	RunE: runToken,
}

type tokenEntry struct {
	Provider string `json:"provider"`
	EnvVar   string `json:"env_var"`
	Masked   string `json:"value"`
	Set      bool   `json:"set"`
}

var knownTokens = []tokenEntry{
	{Provider: "github", EnvVar: "GITHUB_TOKEN"},
	{Provider: "github", EnvVar: "GH_TOKEN"},
	{Provider: "gitlab", EnvVar: "GITLAB_TOKEN"},
	{Provider: "aqua", EnvVar: "AQUA_GITHUB_TOKEN"},
	{Provider: "npm", EnvVar: "NPM_TOKEN"},
}

func runToken(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	filterProvider := ""
	if len(args) > 0 {
		filterProvider = strings.ToLower(args[0])
	}

	// Resolve values from environment.
	entries := make([]tokenEntry, 0, len(knownTokens))
	for _, t := range knownTokens {
		if filterProvider != "" && t.Provider != filterProvider {
			continue
		}
		val := os.Getenv(t.EnvVar)
		masked := "(not set)"
		set := false
		if val != "" {
			set = true
			if len(val) > 4 {
				masked = strings.Repeat("*", len(val)-4) + val[len(val)-4:]
			} else {
				masked = strings.Repeat("*", len(val))
			}
		}
		entries = append(entries, tokenEntry{
			Provider: t.Provider,
			EnvVar:   t.EnvVar,
			Masked:   masked,
			Set:      set,
		})
	}

	if len(entries) == 0 {
		formatter.Info("No token entries found for the specified provider.", nil)
		return nil
	}

	if jsonOutput {
		formatter.Success("Tokens", map[string]interface{}{
			"tokens": entries,
		})
		return nil
	}

	tableData := pterm.TableData{
		{"PROVIDER", "ENV VAR", "VALUE"},
	}
	for _, e := range entries {
		val := pterm.FgRed.Sprint(e.Masked)
		if e.Set {
			val = pterm.FgGreen.Sprint(e.Masked)
		}
		tableData = append(tableData, []string{
			pterm.FgCyan.Sprint(e.Provider),
			e.EnvVar,
			val,
		})
	}

	fmt.Println()
	pterm.DefaultTable.
		WithHasHeader(true).
		WithSeparator("   ").
		WithHeaderStyle(pterm.NewStyle(pterm.FgCyan, pterm.Bold)).
		WithData(tableData).
		Render()
	return nil
}
