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
	registryCmd.Flags().StringVarP(&registrySearch, "search", "s", "", "filter tools by name or description")
	if rootCmd != nil {
		rootCmd.AddCommand(registryCmd)
	}
}

// wellKnownTools provides descriptions and homepage URLs for commonly used tools.
var wellKnownTools = map[string]struct {
	Description string
	Homepage    string
}{
	"node":              {"JavaScript runtime built on Chrome's V8 engine", "https://nodejs.org"},
	"go":                {"Open source programming language for scalable systems", "https://go.dev"},
	"python":            {"High-level general-purpose programming language", "https://python.org"},
	"ruby":              {"Dynamic, reflective, object-oriented language", "https://ruby-lang.org"},
	"java":              {"Platform-independent, compiled, OOP language", "https://java.com"},
	"rust":              {"Systems language focused on safety and performance", "https://rust-lang.org"},
	"deno":              {"Modern JavaScript/TypeScript runtime (Rust-based)", "https://deno.land"},
	"bun":               {"All-in-one JavaScript runtime & toolkit", "https://bun.sh"},
	"pnpm":              {"Fast, disk space-efficient package manager for Node.js", "https://pnpm.io"},
	"yarn":              {"Fast, reliable, and secure npm alternative", "https://yarnpkg.com"},
	"terraform":         {"Infrastructure as Code tool by HashiCorp", "https://terraform.io"},
	"kubectl":           {"Kubernetes CLI for cluster management", "https://kubernetes.io/docs/reference/kubectl"},
	"helm":              {"Package manager for Kubernetes", "https://helm.sh"},
	"pre-commit":        {"Framework for managing multi-language Git hooks", "https://pre-commit.com"},
	"ruff":              {"Extremely fast Python linter, written in Rust", "https://astral.sh/ruff"},
	"uv":                {"Python package and project manager, written in Rust", "https://astral.sh/uv"},
	"shellcheck":        {"Static analysis tool for shell scripts", "https://shellcheck.net"},
	"hadolint":          {"Dockerfile linter & shell checker", "https://github.com/hadolint/hadolint"},
	"actionlint":        {"Static checker for GitHub Actions workflow files", "https://github.com/rhysd/actionlint"},
	"gitleaks":          {"Detect hardcoded secrets in Git repos", "https://github.com/gitleaks/gitleaks"},
	"prettier":          {"Opinionated code formatter for JS/TS/CSS/HTML", "https://prettier.io"},
	"eslint":            {"Find and fix problems in JavaScript/TypeScript code", "https://eslint.org"},
	"stylelint":         {"Mighty CSS linter to avoid errors & enforce conventions", "https://stylelint.io"},
	"gh":                {"GitHub CLI - work with GitHub from the terminal", "https://cli.github.com"},
	"jq":                {"Lightweight command-line JSON processor", "https://jqlang.github.io/jq"},
	"syft":              {"Generate software bill of materials (SBOM)", "https://github.com/anchore/syft"},
	"osv-scanner":       {"Find vulnerabilities using the OSV database", "https://github.com/google/osv-scanner"},
	"yamllint":          {"Linter for YAML files", "https://github.com/adrienverge/yamllint"},
	"markdownlint-cli2": {"Fast, flexible, configuration-based markdown linter", "https://github.com/DavidAnson/markdownlint-cli2"},
	"commitizen":        {"Release management tool — conventional commits helper", "https://commitizen-tools.github.io"},
	"bats":              {"Bash Automated Testing System", "https://bats-core.readthedocs.io"},
}

// registryCmd lists all tools available in the UniRTM registry.
var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "List available tools in the UniRTM registry with descriptions and homepages",
	Long: `List available tools in the UniRTM registry.

Shows all tools that can be installed via UniRTM, including the backend source,
native provider, human-readable description, and homepage URL for well-known tools.

Examples:
  # List all available tools
  unirtm registry

  # Filter by name or description
  unirtm registry --search go

  # JSON output
  unirtm registry --json`,
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	RunE:    runRegistry,
}

type registryEntry struct {
	Tool        string `json:"tool"`
	Backend     string `json:"backend"`
	Provider    string `json:"provider"`
	Description string `json:"description,omitempty"`
	Homepage    string `json:"homepage,omitempty"`
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
		entry := registryEntry{
			Tool:     name,
			Backend:  "native",
			Provider: name,
		}
		if meta, ok := wellKnownTools[name]; ok {
			entry.Description = meta.Description
			entry.Homepage = meta.Homepage
		}
		toolSet[name] = entry
	}

	// Backends list (generic tools via github/aqua/http).
	for _, bName := range backendRegistry.List() {
		if _, exists := toolSet[bName]; !exists {
			entry := registryEntry{
				Tool:    bName,
				Backend: bName,
			}
			if meta, ok := wellKnownTools[bName]; ok {
				entry.Description = meta.Description
				entry.Homepage = meta.Homepage
			}
			toolSet[bName] = entry
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

	// Apply search filter (name or description).
	if registrySearch != "" {
		q := strings.ToLower(registrySearch)
		filtered := entries[:0]
		for _, e := range entries {
			if strings.Contains(strings.ToLower(e.Tool), q) ||
				strings.Contains(strings.ToLower(e.Description), q) {
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

	// Rich pterm table with descriptions and homepage links
	pterm.DefaultSection.Printfln("%d tools available in the UniRTM registry", len(entries))

	tableData := pterm.TableData{
		{"TOOL", "BACKEND", "DESCRIPTION", "HOMEPAGE"},
	}
	for _, e := range entries {
		backendStr := e.Backend
		switch e.Backend {
		case "native":
			backendStr = pterm.FgGreen.Sprint("native")
		case "github":
			backendStr = pterm.FgYellow.Sprint("github")
		case "aqua":
			backendStr = pterm.FgCyan.Sprint("aqua")
		default:
			backendStr = pterm.FgMagenta.Sprint(e.Backend)
		}

		desc := e.Description
		if len(desc) > 55 {
			desc = desc[:52] + "…"
		}
		if desc == "" {
			desc = pterm.FgGray.Sprint("─")
		}

		homepage := e.Homepage
		if homepage == "" {
			homepage = pterm.FgGray.Sprint("─")
		} else {
			homepage = pterm.FgBlue.Sprint(homepage)
		}

		tableData = append(tableData, []string{
			pterm.FgCyan.Sprint(e.Tool),
			backendStr,
			desc,
			homepage,
		})
	}

	_ = pterm.DefaultTable.
		WithHasHeader(true).
		WithSeparator("  ").
		WithHeaderStyle(pterm.NewStyle(pterm.FgLightCyan, pterm.Bold)).
		WithData(tableData).
		Render()

	fmt.Println()
	output.Infof("Use 'unirtm install <tool>@<version>' to install any tool above.")
	return nil
}
