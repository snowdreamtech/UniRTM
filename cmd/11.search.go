// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/snowdreamtech/unirtm/internal/service"
	"github.com/spf13/cobra"
)

var (
	// searchBackend filters search results by backend type
	searchBackend string
	// searchLimit limits the number of search results
	searchLimit int
	// searchInstall enables interactive "install on match" prompt
	searchInstall bool
)

// init registers the search command to the root command.
func init() {
	searchCmd.Flags().StringVarP(&searchBackend, "backend", "b", "", "filter by backend type (github, aqua, http)")
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "l", 50, "maximum number of results to display")
	searchCmd.Flags().BoolVarP(&searchInstall, "install", "i", false, "prompt to install a matching tool after search")

	if rootCmd != nil {
		rootCmd.AddCommand(searchCmd)
	}
}

// searchCmd represents the search command which searches the tool index.
var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for development tools and optionally install on match",
	Long: `Search for available development tools in the tool index.

The search command queries the local tool index by name or description.
Run 'unirtm index update' to refresh the index from remote backends.

Use --install (-i) to get an interactive prompt to install a tool right
after seeing the search results — no need to run a separate install command.

Examples:
  # Search for Node.js
  unirtm search node

  # Search and prompt to install immediately
  unirtm search go --install

  # Filter by backend
  unirtm search python --backend github

  # JSON output
  unirtm search go --json

  # Limit results
  unirtm search tool --limit 10`,
	Args: cobra.ExactArgs(1),
	RunE: runSearch,
}

// runSearch executes the search command.
//
// Validates: Requirements 11.4, 11.5, 23.2
func runSearch(cmd *cobra.Command, args []string) error {
	query := args[0]

	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	ctx := context.Background()

	dbPath := env.GetDatabasePath()
	db, err := database.Open(ctx, database.Config{
		Path:    dbPath,
		WALMode: true,
	})
	if err != nil {
		formatter.Error("Failed to initialize database", map[string]interface{}{
			"error": err.Error(),
			"path":  dbPath,
		})
		return fmt.Errorf("initialize database: %w", err)
	}
	defer db.Close()

	indexRepo, err := sqlite.NewIndexRepository(db.Conn())
	if err != nil {
		formatter.Error("Failed to create index repository", map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("create index repository: %w", err)
	}

	auditRepo, _ := sqlite.NewAuditRepository(db.Conn())

	// NewIndexManager accepts map[string]backend.Backend (nil = empty)
	indexManager, err := service.NewIndexManager(indexRepo, auditRepo, nil, service.IndexManagerConfig{})
	if err != nil {
		formatter.Error("Failed to create index manager", map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("create index manager: %w", err)
	}

	// Check if index is stale and notify user
	if !quiet {
		isStale, msg, staleErr := indexManager.PromptForUpdate(ctx)
		if staleErr == nil && isStale && msg != "" {
			formatter.Info("⚠  "+msg, nil)
		}
	}

	// Perform search
	var spinner *pterm.SpinnerPrinter
	if !jsonOutput && !quiet {
		spinner, _ = pterm.DefaultSpinner.Start("Searching tool index for " + pterm.FgCyan.Sprint(query) + "...")
	}

	results, err := indexManager.SearchTools(ctx, service.SearchOptions{
		Query:   query,
		Backend: searchBackend,
		Limit:   searchLimit,
	})

	if spinner != nil {
		if err != nil {
			spinner.Fail("Search failed")
		} else {
			spinner.Success("Search complete")
		}
	}

	if err != nil {
		formatter.Error("Search failed", map[string]interface{}{
			"query": query,
			"error": err.Error(),
		})
		return fmt.Errorf("search tools: %w", err)
	}

	if len(results) == 0 {
		if jsonOutput {
			fmt.Println("[]")
		} else {
			formatter.Info(fmt.Sprintf("No tools found matching %q", query), nil)
			pterm.Info.Println("Tip: Run 'unirtm index update' to refresh the tool index.")
		}
		return nil
	}

	if jsonOutput {
		type jsonResult struct {
			Tool        string `json:"tool"`
			Description string `json:"description"`
			Homepage    string `json:"homepage"`
			License     string `json:"license"`
			Backend     string `json:"backend"`
		}
		jsonResults := make([]jsonResult, 0, len(results))
		for _, entry := range results {
			jsonResults = append(jsonResults, jsonResult{
				Tool:        entry.Tool,
				Description: entry.Description,
				Homepage:    entry.Homepage,
				License:     entry.License,
				Backend:     entry.Backend,
			})
		}
		formatter.Success(fmt.Sprintf("Found %d tools", len(jsonResults)), map[string]interface{}{
			"count": len(jsonResults),
			"tools": jsonResults,
		})
		return nil
	}

	// ── Rich pterm table output ──────────────────────────────────────────────
	pterm.DefaultSection.Printfln("Search results for %q  (%d found)", query, len(results))

	tableData := pterm.TableData{
		{"TOOL", "BACKEND", "LICENSE", "DESCRIPTION"},
	}
	for _, entry := range results {
		desc := entry.Description
		if len(desc) > 60 {
			desc = desc[:57] + "…"
		}
		if desc == "" {
			desc = pterm.FgGray.Sprint("─")
		}

		license := entry.License
		if license == "" {
			license = pterm.FgGray.Sprint("─")
		}

		backendStr := entry.Backend
		switch strings.ToLower(entry.Backend) {
		case "github":
			backendStr = pterm.FgYellow.Sprint(entry.Backend)
		case "aqua":
			backendStr = pterm.FgCyan.Sprint(entry.Backend)
		case "native":
			backendStr = pterm.FgGreen.Sprint(entry.Backend)
		default:
			backendStr = pterm.FgMagenta.Sprint(entry.Backend)
		}

		tableData = append(tableData, []string{
			pterm.FgCyan.Sprint(entry.Tool),
			backendStr,
			license,
			desc,
		})
	}

	_ = pterm.DefaultTable.
		WithHasHeader(true).
		WithSeparator("  ").
		WithHeaderStyle(pterm.NewStyle(pterm.FgLightCyan, pterm.Bold)).
		WithData(tableData).
		Render()

	// Show homepage hint for first result if available
	if len(results) > 0 && results[0].Homepage != "" {
		fmt.Printf("\n  🔗 %s → %s\n", pterm.FgCyan.Sprint(results[0].Tool), pterm.FgBlue.Sprint(results[0].Homepage))
	}

	// ── Interactive install prompt ───────────────────────────────────────────
	if searchInstall && !jsonOutput {
		fmt.Println()
		// Build tool selection options
		toolNames := make([]string, 0, len(results))
		for _, r := range results {
			toolNames = append(toolNames, r.Tool)
		}

		selectedTool, err := pterm.DefaultInteractiveSelect.
			WithOptions(toolNames).
			WithDefaultText("Select a tool to install (ESC to skip)").
			Show()
		if err != nil {
			// User pressed ESC or cancelled — not an error
			pterm.Info.Println("Installation skipped.")
			return nil
		}

		// Prompt for version
		version, err := pterm.DefaultInteractiveTextInput.
			WithDefaultText("latest").
			Show(fmt.Sprintf("Version for %s", pterm.FgCyan.Sprint(selectedTool)))
		if err != nil || strings.TrimSpace(version) == "" {
			version = "latest"
		}
		version = strings.TrimSpace(version)

		pterm.Info.Printfln("Installing %s@%s…", pterm.FgCyan.Sprint(selectedTool), version)

		// Delegate to the install command logic
		installArgs := []string{fmt.Sprintf("%s@%s", selectedTool, version)}
		return runInstall(cmd, installArgs)
	}

	fmt.Println()
	pterm.Info.Printfln("Use 'unirtm install <tool>@<version>' or run 'unirtm search %s --install' for quick install.", query)
	return nil
}
