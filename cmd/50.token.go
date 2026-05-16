// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"os"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/spf13/cobra"
)

func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(tokenCmd)
	}
}

var tokenCmd = &cobra.Command{
	Use:   "token [provider]",
	Short: "Show tokens from environment variables",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runToken,
}

type tokenEntry struct {
	Provider string `json:"provider"`
	EnvVar   string `json:"env_var"`
}

type tokenOutput struct {
	Provider string `json:"provider"`
	EnvVar   string `json:"env_var"`
	Token    string `json:"token"`
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
		filterProvider = args[0]
	}

	results := []tokenOutput{}
	for _, t := range knownTokens {
		if filterProvider != "" && t.Provider != filterProvider {
			continue
		}
		val := env.Get(t.EnvVar)
		masked := "(not set)"
		set := false
		if val != "" {
			masked = maskToken(val)
			set = true
		}
		results = append(results, tokenOutput{
			Provider: t.Provider,
			EnvVar:   t.EnvVar,
			Token:    masked,
			Set:      set,
		})
	}

	if getOutputFormat() == "json" {
		formatter.Data(results)
		return nil
	}

	pterm.DefaultSection.Println("Provider Tokens")
	tableData := pterm.TableData{
		{"Provider", "Env Var", "Token Status"},
	}

	for _, r := range results {
		status := pterm.LightRed("not set")
		if r.Set {
			status = pterm.LightGreen(r.Token)
		}
		tableData = append(tableData, []string{
			pterm.LightBlue(r.Provider),
			pterm.FgGray.Sprint(r.EnvVar),
			status,
		})
	}

	return pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
}

func maskToken(token string) string {
	if len(token) <= 8 {
		return "********"
	}
	return token[:4] + "...." + token[len(token)-4:]
}
