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
	Name                string `json:"name"`
	SupportsChecksum    bool   `json:"supports_checksum"`
	SupportsGPG         bool   `json:"supports_gpg"`
	SupportsAttestation bool   `json:"supports_attestation"`
	AttestationType     string `json:"attestation_type"`
	IsRecommended       bool   `json:"is_recommended"`
	IsScriptless        bool   `json:"is_scriptless"`
	Reach               string `json:"reach"`
	IsStable            bool   `json:"is_stable"`
	SupportsOffline     bool   `json:"supports_offline"`
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
			Name:                name,
			SupportsChecksum:    b.SupportsChecksum(),
			SupportsGPG:         b.SupportsGPG(),
			SupportsAttestation: b.AttestationType() != "",
			AttestationType:     b.AttestationType(),
			IsRecommended:       b.IsRecommended(),
			IsScriptless:        b.IsScriptless(),
			Reach:               b.GetReach(),
			IsStable:            b.IsStable(),
			SupportsOffline:     b.SupportsOffline(),
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

	// Classic Red/Green Table
	tableData := pterm.TableData{
		{"BACKEND", "RECOMMENDED", "CHECKSUM", "GPG", "VERIFY", "SCRIPTLESS", "REACH", "STABILITY", "OFFLINE"},
	}
	for _, e := range entries {
		recommended := pterm.FgRed.Sprint("✗")
		if e.IsRecommended {
			recommended = pterm.FgGreen.Sprint("✓")
		}

		checksum := pterm.FgRed.Sprint("✗")
		if e.SupportsChecksum {
			checksum = pterm.FgGreen.Sprint("✓")
		}
		gpg := pterm.FgRed.Sprint("✗")
		if e.SupportsGPG {
			gpg = pterm.FgGreen.Sprint("✓")
		}
		verify := pterm.FgRed.Sprint("✗")
		if e.SupportsAttestation {
			verify = pterm.FgGreen.Sprint("✓")
			if e.AttestationType != "" {
				verify += fmt.Sprintf(" (%s)", e.AttestationType)
			}
		}

		scriptless := pterm.FgRed.Sprint("✗")
		if e.IsScriptless {
			scriptless = pterm.FgGreen.Sprint("✓")
		}
		reach := pterm.FgGray.Sprint(e.Reach)
		switch e.Reach {
		case "Huge":
			reach = pterm.FgMagenta.Sprint(e.Reach)
		case "Large":
			reach = pterm.FgBlue.Sprint(e.Reach)
		case "Medium":
			reach = pterm.FgYellow.Sprint(e.Reach)
		case "Small":
			reach = pterm.FgGray.Sprint(e.Reach)
		}

		stability := pterm.FgGreen.Sprint("✓")
		if !e.IsStable {
			stability = pterm.FgYellow.Sprint("?")
		}
		offline := pterm.FgRed.Sprint("✗")
		if e.SupportsOffline {
			offline = pterm.FgGreen.Sprint("✓")
		}
		
		tableData = append(tableData, []string{
			pterm.FgCyan.Sprint(e.Name),
			recommended,
			checksum,
			gpg,
			verify,
			scriptless,
			reach,
			stability,
			offline,
		})
	}

	fmt.Println()
	pterm.DefaultTable.
		WithHasHeader(true).
		WithSeparator("   ").
		WithHeaderStyle(pterm.NewStyle(pterm.FgCyan, pterm.Bold)).
		WithData(tableData).
		Render()

	// Add Legend/Glossary at the bottom
	fmt.Println()
	pterm.DefaultSection.WithLevel(2).Println("Legend & Column Meanings")
	
	legendData := pterm.TableData{
		{pterm.FgCyan.Sprint("RECOMMENDED"), "Official certification by UniRTM team for security and reliability."},
		{pterm.FgCyan.Sprint("CHECKSUM"), "Integrity verification via hash files (sha256/sha512)."},
		{pterm.FgCyan.Sprint("GPG"), "Identity verification via digital signatures (.asc/.sig)."},
		{pterm.FgCyan.Sprint("VERIFY"), "Advanced supply chain endorsement (e.g., SLSA, GitHub Attestation)."},
		{pterm.FgCyan.Sprint("SCRIPTLESS"), "Security depth: '✓' means installation is declarative (no arbitrary script exec)."},
		{pterm.FgCyan.Sprint("REACH"), "Utility coverage: From 'Small' (core tools) to 'Huge' (global ecosystems)."},
		{pterm.FgCyan.Sprint("STABILITY"), "Reliability: '✓' means high stability; '?' indicates potential for bit-rot/broken links."},
		{pterm.FgCyan.Sprint("OFFLINE"), "Enterprise capability: Supports private mirrors or offline installation."},
	}

	pterm.DefaultTable.
		WithSeparator("  :  ").
		WithData(legendData).
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

	recommended := "no"
	if b.IsRecommended() {
		recommended = "yes"
	}

	scriptless := "no"
	if b.IsScriptless() {
		scriptless = "yes"
	}

	stability := "high"
	if !b.IsStable() {
		stability = "medium (bit-rot potential)"
	}
	offline := "no"
	if b.SupportsOffline() {
		offline = "yes"
	}

	fmt.Println()
	pterm.DefaultSection.Printf("Backend: %s", pterm.FgCyan.Sprint(b.Name()))
	pterm.DefaultTable.
		WithSeparator("   ").
		WithData(pterm.TableData{
			{"Recommended", recommended},
			{"Scriptless (No code exec during install)", scriptless},
			{"Reach / Coverage", b.GetReach()},
			{"Stability / Bit-rot resistance", stability},
			{"Offline / Mirror support", offline},
			{"Checksum verification", checksum},
			{"GPG signature", gpg},
		}).Render()
	return nil
}
