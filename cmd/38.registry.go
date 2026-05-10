// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/provider"
	"github.com/spf13/cobra"
)

var (
	registrySearch string
)

func init() {
	registryCmd.Flags().StringVarP(&registrySearch, "search", "s", "", "filter tools by name")
	if rootCmd != nil {
		rootCmd.AddCommand(registryCmd)
	}
}

// registryCmd lists all tools available in the UniRTM registry.
var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "List available tools in the UniRTM registry",
	Long: `List available tools in the UniRTM registry.

Shows all tools that can be installed via UniRTM, including which backend
and native provider handles each tool.

Examples:
  # List all available tools
  unirtm registry

  # Filter by name
  unirtm registry --search go

  # JSON output
  unirtm registry --json`,
	Args: cobra.NoArgs,
	RunE: runRegistry,
}

type registryEntry struct {
	Tool     string `json:"tool"`
	Backend  string `json:"backend"`
	Provider string `json:"provider"`
}

func runRegistry(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	backendRegistry := backend.NewRegistry()
	providerRegistry := provider.NewRegistry()

	// Combine: native providers + backend-registered names.
	toolSet := make(map[string]registryEntry)

	// Native providers (first-class support).
	for _, name := range providerRegistry.List() {
		toolSet[name] = registryEntry{
			Tool:     name,
			Backend:  "native",
			Provider: name,
		}
	}

	// Backends list (generic tools via github/aqua/http).
	for _, bName := range backendRegistry.List() {
		// Each backend is itself a source; register it as a backend entry.
		if _, exists := toolSet[bName]; !exists {
			toolSet[bName] = registryEntry{
				Tool:    bName,
				Backend: bName,
			}
		}
	}

	// Build sorted list.
	entries := make([]registryEntry, 0, len(toolSet))
	for _, e := range toolSet {
		entries = append(entries, e)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Tool < entries[j].Tool
	})

	// Apply search filter.
	if registrySearch != "" {
		q := strings.ToLower(registrySearch)
		filtered := entries[:0]
		for _, e := range entries {
			if strings.Contains(strings.ToLower(e.Tool), q) {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
	}

	if len(entries) == 0 {
		formatter.Info("No tools found matching your query.", nil)
		return nil
	}

	if jsonOutput {
		formatter.Success("Registry", map[string]interface{}{
			"count": len(entries),
			"tools": entries,
		})
		return nil
	}

	tableData := pterm.TableData{
		{"TOOL", "BACKEND", "PROVIDER"},
	}
	for _, e := range entries {
		providerStr := e.Provider
		if providerStr == "" {
			providerStr = "─"
		}
		backendStr := e.Backend
		if e.Backend == "native" {
			backendStr = pterm.FgGreen.Sprint("native")
		} else {
			backendStr = pterm.FgMagenta.Sprint(e.Backend)
		}
		tableData = append(tableData, []string{
			pterm.FgCyan.Sprint(e.Tool),
			backendStr,
			providerStr,
		})
	}

	fmt.Printf("\n%d tools available\n\n", len(entries))
	pterm.DefaultTable.
		WithHasHeader(true).
		WithSeparator("   ").
		WithHeaderStyle(pterm.NewStyle(pterm.FgCyan, pterm.Bold)).
		WithData(tableData).
		Render()
	return nil
}
