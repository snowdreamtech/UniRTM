// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/spf13/cobra"
)

func init() {
	backendsCmd.AddCommand(backendsListCmd)
	backendsCmd.AddCommand(backendsInfoCmd)
	if rootCmd != nil {
		rootCmd.AddCommand(backendsCmd)
	}
}

// backendsCmd is the root of the backends sub-command group.
var backendsCmd = &cobra.Command{
	Use:     "backends",
	Short:   "Manage and list UniRTM backends",
	Aliases: []string{"b"},
	Long: `Manage and list UniRTM backends.

Backends define where tools are downloaded from (GitHub Releases, Aqua
registry, HTTP URLs, etc.).

Sub-commands:
  ls    List all registered backends
  info  Show details about a specific backend

Examples:
  unirtm backends
  unirtm backends ls
  unirtm backends info github
  unirtm backends --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default: run the list sub-command.
		return runBackendsList(cmd, args)
	},
}

// backendsListCmd lists all registered backends.
var backendsListCmd = &cobra.Command{
	Use:     "ls",
	Short:   "List all registered backends",
	Aliases: []string{"list", "b"},
	Args:    cobra.NoArgs,
	RunE:    runBackendsList,
}

// backendsInfoCmd shows details about a specific backend.
var backendsInfoCmd = &cobra.Command{
	Use:   "info <backend>",
	Short: "Show details about a specific backend",
	Args:  cobra.ExactArgs(1),
	RunE:  runBackendsInfo,
}

type backendEntry struct {
	Name           string `json:"name"`
	SupportsChecksum bool  `json:"supports_checksum"`
	SupportsGPG    bool   `json:"supports_gpg"`
}

func runBackendsList(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	registry := backend.NewRegistry()
	names := registry.List()
	sort.Strings(names)

	entries := make([]backendEntry, 0, len(names))
	for _, name := range names {
		b, err := registry.Get(name)
		if err != nil {
			continue
		}
		entries = append(entries, backendEntry{
			Name:             name,
			SupportsChecksum: b.SupportsChecksum(),
			SupportsGPG:      b.SupportsGPG(),
		})
	}

	if len(entries) == 0 {
		formatter.Info("No backends registered.", nil)
		return nil
	}

	if jsonOutput {
		formatter.Success("Backends", map[string]interface{}{
			"count":    len(entries),
			"backends": entries,
		})
		return nil
	}

	tableData := pterm.TableData{
		{"BACKEND", "CHECKSUM", "GPG"},
	}
	for _, e := range entries {
		checksum := pterm.FgRed.Sprint("✗")
		if e.SupportsChecksum {
			checksum = pterm.FgGreen.Sprint("✓")
		}
		gpg := pterm.FgRed.Sprint("✗")
		if e.SupportsGPG {
			gpg = pterm.FgGreen.Sprint("✓")
		}
		tableData = append(tableData, []string{
			pterm.FgCyan.Sprint(e.Name),
			checksum,
			gpg,
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

func runBackendsInfo(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	name := args[0]
	registry := backend.NewRegistry()
	b, err := registry.Get(name)
	if err != nil {
		formatter.Error(fmt.Sprintf("Backend %q not found. Run 'unirtm backends' to see available backends.", name))
		return err
	}

	checksum := "no"
	if b.SupportsChecksum() {
		checksum = "yes"
	}
	gpg := "no"
	if b.SupportsGPG() {
		gpg = "yes"
	}

	if jsonOutput {
		formatter.Success(fmt.Sprintf("Backend: %s", name), map[string]interface{}{
			"name":              b.Name(),
			"supports_checksum": b.SupportsChecksum(),
			"supports_gpg":     b.SupportsGPG(),
		})
		return nil
	}

	fmt.Println()
	pterm.DefaultSection.Printf("Backend: %s", pterm.FgCyan.Sprint(b.Name()))
	pterm.DefaultTable.
		WithSeparator("   ").
		WithData(pterm.TableData{
			{"Checksum verification", checksum},
			{"GPG signature", gpg},
		}).Render()
	return nil
}
