// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/spf13/cobra"
)

var (
	// useGlobal writes the tool version to the global config (~/.config/unirtm/config.toml)
	useGlobal bool
	// usePath specifies the directory to write the config file into
	usePath string
)

// init registers the use command to the root command.
func init() {
	useCmd.Flags().BoolVarP(&useGlobal, "global", "g", false, "write to global config (~/.config/unirtm/config.toml)")
	useCmd.Flags().StringVarP(&usePath, "path", "p", "", "directory to write config file into (default: current directory)")

	if rootCmd != nil {
		rootCmd.AddCommand(useCmd)
	}
}

// useCmd represents the use command which sets a tool version in the config file.
// This is equivalent to `mise use` — it writes the tool@version declaration into
// the nearest unirtm.toml or .unirtm.toml config file.
var useCmd = &cobra.Command{
	Use:   "use <tool>@<version>",
	Short: "Set a tool version in the config file",
	Long: `Set a tool version in the configuration file.

The use command writes a tool version declaration into the nearest
unirtm.toml or .unirtm.toml configuration file. If neither file exists
it creates unirtm.toml in the target directory.

Examples:
  # Set node version in current directory's config
  unirtm use node@20.0.0

  # Set globally (writes to ~/.config/unirtm/config.toml)
  unirtm use node@20.0.0 --global

  # Set in a specific directory
  unirtm use node@20.0.0 --path /path/to/project

  # Set multiple tools
  unirtm use node@20.0.0 python@3.11.0`,
	Aliases: []string{"u"},
	Args:    cobra.MinimumNArgs(1),
	RunE:    runUse,
}

// runUse executes the use command.
// It parses tool@version arguments and writes them to the config file.
func runUse(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	// Parse tool@version pairs
	type toolVersion struct {
		tool    string
		version string
	}
	pairs := make([]toolVersion, 0, len(args))
	for _, arg := range args {
		tool := arg
		version := ""

		if strings.Contains(arg, "@") {
			parts := strings.SplitN(arg, "@", 2)
			tool = parts[0]
			version = parts[1]
		}

		if version == "" {
			if !jsonOutput && pterm.IsTerminal(os.Stdin) {
				cfg, _ := config.Load()
				im := getInstallationManager(cmd.Context(), cfg)
				selected, err := im.SelectVersionInteractive(cmd.Context(), tool)
				if err == nil {
					version = selected
				} else {
					return fmt.Errorf("interactive selection failed for %s: %w", tool, err)
				}
			} else {
				return fmt.Errorf("invalid format %q: expected <tool>@<version> (e.g. node@20.0.0)", arg)
			}
		}
		pairs = append(pairs, toolVersion{tool: tool, version: version})
	}

	// Resolve target directory
	targetDir := usePath
	if useGlobal {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("get home directory: %w", err)
		}
		targetDir = filepath.Join(homeDir, ".config", "unirtm")
	}
	if targetDir == "" {
		var err error
		targetDir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}
	}

	// Find or create the config file in targetDir
	configFile := findOrCreateConfigFile(targetDir)

	if dryRun {
		for _, p := range pairs {
			formatter.Info(fmt.Sprintf("[dry-run] Would write %s = %q to %s", p.tool, p.version, configFile), nil)
		}
		return nil
	}

	// Ensure directory exists
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("create config directory %s: %w", targetDir, err)
	}

	// Read existing content or create empty
	content, err := readFileOrEmpty(configFile)
	if err != nil {
		return fmt.Errorf("read config file: %w", err)
	}

	// Apply each tool@version into the [tools] section
	for _, p := range pairs {
		content = upsertToolVersion(content, p.tool, p.version)
	}

	// Write back
	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("write config file %s: %w", configFile, err)
	}

	for _, p := range pairs {
		formatter.Success(fmt.Sprintf("Set %s = %q in %s", p.tool, p.version, configFile), map[string]interface{}{
			"tool":    p.tool,
			"version": p.version,
			"file":    configFile,
		})
	}

	return nil
}

// findOrCreateConfigFile finds an existing config file in the directory,
// or returns the path to a new unirtm.toml to be created.
func findOrCreateConfigFile(dir string) string {
	for _, name := range []string{".unirtm.toml", "unirtm.toml"} {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return filepath.Join(dir, "unirtm.toml")
}

// readFileOrEmpty reads a file content or returns an empty string if the file doesn't exist.
func readFileOrEmpty(path string) (string, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// upsertToolVersion adds or updates a tool version entry in the TOML [tools] section.
// It handles three cases:
//  1. [tools] section exists with this tool key → update the version
//  2. [tools] section exists but tool key is missing → append entry
//  3. No [tools] section → append the whole section
func upsertToolVersion(content, tool, version string) string {
	lines := strings.Split(content, "\n")
	newEntry := fmt.Sprintf("%s = %q", tool, version)

	// Look for [tools] section
	inTools := false
	toolsStart := -1
	toolsEnd := -1
	toolLineIdx := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "[tools]" {
			inTools = true
			toolsStart = i
			continue
		}
		if inTools {
			if strings.HasPrefix(trimmed, "[") && trimmed != "[tools]" {
				// New section started
				toolsEnd = i
				inTools = false
				break
			}
			if strings.HasPrefix(trimmed, tool+"=") || strings.HasPrefix(trimmed, tool+" =") {
				toolLineIdx = i
			}
		}
	}
	if inTools {
		toolsEnd = len(lines)
	}

	if toolsStart == -1 {
		// No [tools] section — append it
		if content != "" && !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += "\n[tools]\n" + newEntry + "\n"
		return content
	}

	if toolLineIdx != -1 {
		// Update existing line
		lines[toolLineIdx] = newEntry
		return strings.Join(lines, "\n")
	}

	// Insert before toolsEnd
	insertAt := toolsEnd
	if inTools || toolsEnd == len(lines) {
		insertAt = toolsEnd
	}
	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:insertAt]...)
	newLines = append(newLines, newEntry)
	newLines = append(newLines, lines[insertAt:]...)
	return strings.Join(newLines, "\n")
}
