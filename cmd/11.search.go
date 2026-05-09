// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

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
)

// init registers the search command to the root command.
func init() {
	searchCmd.Flags().StringVarP(&searchBackend, "backend", "b", "", "filter by backend type (github, aqua, http)")
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "l", 50, "maximum number of results to display")

	if rootCmd != nil {
		rootCmd.AddCommand(searchCmd)
	}
}

// searchCmd represents the search command which searches the tool index.
var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for available development tools",
	Long: `Search for available development tools in the tool index.

The search command queries the local tool index by name or description.
Run 'unirtm index update' to refresh the index from remote backends.

Examples:
  # Search for Node.js
  unirtm search node

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
	results, err := indexManager.SearchTools(ctx, service.SearchOptions{
		Query:   query,
		Backend: searchBackend,
		Limit:   searchLimit,
	})
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
			fmt.Println("Tip: Run 'unirtm index update' to refresh the tool index.")
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

	fmt.Printf("Search results for %q (%d found):\n\n", query, len(results))
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TOOL\tBACKEND\tDESCRIPTION")
	fmt.Fprintln(w, "----\t-------\t-----------")
	for _, entry := range results {
		desc := entry.Description
		if len(desc) > 60 {
			desc = desc[:57] + "..."
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", entry.Tool, entry.Backend, desc)
	}
	w.Flush()
	return nil
}
