// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"fmt"
	"os"
	"sort"

	"github.com/pelletier/go-toml/v2"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/spf13/cobra"
)

var (
	fmtCheck bool
)

func init() {
	fmtCmd.Flags().BoolVar(&fmtCheck, "check", false, "check if config is formatted without modifying it (exit 1 if not)")
	if rootCmd != nil {
		rootCmd.AddCommand(fmtCmd)
	}
}

// fmtCmd formats unirtm.toml with canonical key ordering and indentation.
var fmtCmd = &cobra.Command{
	Use:   "fmt",
	Short: "Format unirtm.toml with canonical key ordering",
	Long: `Format unirtm.toml with canonical key ordering and indentation.

Reads the project config file, normalises section ordering, and writes it
back in-place. Use --check in CI to verify formatting without modifying the file.

Section order: [env] → [tools] → [tasks] → [settings] → rest

Examples:
  # Format the config file in-place
  unirtm fmt

  # CI mode: exit 1 if file is not already formatted
  unirtm fmt --check`,
	Args: cobra.NoArgs,
	RunE: runFmt,
}

// canonicalSectionOrder defines the preferred top-level key order in unirtm.toml.
var canonicalSectionOrder = []string{
	"env",
	"tools",
	"tasks",
	"settings",
}

func runFmt(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	cfgPath := resolveConfigFilePath(false)

	original, err := os.ReadFile(cfgPath)
	if os.IsNotExist(err) {
		formatter.Warning(fmt.Sprintf("Config file %q not found, nothing to format.", cfgPath))
		return nil
	}
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to read %s: %v", cfgPath, err))
		return err
	}

	// Parse into generic map.
	var m map[string]interface{}
	if err := toml.Unmarshal(original, &m); err != nil {
		formatter.Error(fmt.Sprintf("Failed to parse %s: %v", cfgPath, err))
		return err
	}
	if m == nil {
		m = make(map[string]interface{})
	}

	// Re-encode in canonical order.
	formatted, err := formatTOML(m)
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to format TOML: %v", err))
		return err
	}

	if bytes.Equal(original, formatted) {
		formatter.Info(fmt.Sprintf("%s is already formatted ✓", cfgPath), nil)
		return nil
	}

	if fmtCheck {
		formatter.Error(fmt.Sprintf("%s is not formatted. Run 'unirtm fmt' to fix.", cfgPath))
		return fmt.Errorf("%s is not formatted", cfgPath)
	}

	if err := os.WriteFile(cfgPath, formatted, 0o644); err != nil {
		formatter.Error(fmt.Sprintf("Failed to write %s: %v", cfgPath, err))
		return err
	}

	formatter.Success(fmt.Sprintf("Formatted %s", cfgPath), nil)
	return nil
}

// formatTOML encodes a map into TOML bytes with canonical section ordering.
func formatTOML(m map[string]interface{}) ([]byte, error) {
	// Build ordered output: canonical sections first, then remaining keys alphabetically.
	ordered := make([]string, 0, len(m))
	inCanonical := make(map[string]bool)
	for _, k := range canonicalSectionOrder {
		if _, ok := m[k]; ok {
			ordered = append(ordered, k)
			inCanonical[k] = true
		}
	}
	rest := make([]string, 0)
	for k := range m {
		if !inCanonical[k] {
			rest = append(rest, k)
		}
	}
	sort.Strings(rest)
	ordered = append(ordered, rest...)

	// Write section by section.
	var buf bytes.Buffer
	for _, k := range ordered {
		single := map[string]interface{}{k: m[k]}
		enc := toml.NewEncoder(&buf)
		if err := enc.Encode(single); err != nil {
			return nil, err
		}
		buf.WriteByte('\n')
	}
	return buf.Bytes(), nil
}
