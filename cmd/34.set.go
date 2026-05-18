// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/spf13/cobra"
)

var (
	setGlobal bool
)

func init() {
	setCmd.Flags().BoolVar(&setGlobal, "global", false, "write to global config (~/.config/unirtm/unirtm.toml)")
	unsetCmd.Flags().BoolVar(&setGlobal, "global", false, "write to global config (~/.config/unirtm/unirtm.toml)")
	if rootCmd != nil {
		rootCmd.AddCommand(setCmd)
		rootCmd.AddCommand(unsetCmd)
	}
}

// setCmd sets environment variables in unirtm.toml.
var setCmd = &cobra.Command{
	Use:   "set <KEY=value>...",
	Short: "Set environment variables in the config file",
	Long: `Set environment variables in the config file.

Variables are written to the [env] section of unirtm.toml in the current
directory (or --global for the global config).

Examples:
  # Set a single variable
  unirtm set NODE_ENV=production

  # Set multiple variables at once
  unirtm set FOO=bar BAZ=qux

  # Write to global config
  unirtm set --global GITHUB_TOKEN=ghp_xxxx`,
	Args: cobra.MinimumNArgs(1),
	RunE: runSet,
}

// unsetCmd removes environment variables from unirtm.toml.
var unsetCmd = &cobra.Command{
	Use:   "unset <KEY>...",
	Short: "Remove environment variables from the config file",
	Long: `Remove environment variables from the config file.

Deletes the specified key(s) from the [env] section of unirtm.toml.

Examples:
  # Remove a variable
  unirtm unset NODE_ENV

  # Remove from global config
  unirtm unset --global GITHUB_TOKEN`,
	Args: cobra.MinimumNArgs(1),
	RunE: runUnset,
}

// ─── helpers ──────────────────────────────────────────────────────────────────

// resolveConfigFilePath returns the config file to edit.
func resolveConfigFilePath(global bool) string {
	if global {
		return env.GetGlobalConfigPath()
	}
	if configPath != "" {
		return configPath
	}
	candidates := []string{"unirtm.toml", ".unirtm.toml", "unirtm.yaml", ".unirtm.yaml"}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return "unirtm.toml"
}

// ─── set ──────────────────────────────────────────────────────────────────────

func runSet(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	pairs := make(map[string]string, len(args))
	for _, arg := range args {
		idx := strings.Index(arg, "=")
		if idx <= 0 {
			formatter.Error(fmt.Sprintf("Invalid format %q — expected KEY=value", arg))
			return fmt.Errorf("invalid argument: %q", arg)
		}
		pairs[arg[:idx]] = arg[idx+1:]
	}

	cfgPath := resolveConfigFilePath(setGlobal)
	content, err := config.ReadFileOrEmpty(cfgPath)
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to load config: %v", err))
		return err
	}

	for k, v := range pairs {
		content = config.UpsertEnvVar(content, k, v)
	}

	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		formatter.Error(fmt.Sprintf("Failed to save config: %v", err))
		return err
	}

	// Apply canonical format and taplo formatting to ensure correct block order
	_, _ = config.FormatFile(cfgPath, false)

	for k, v := range pairs {
		formatter.Success(fmt.Sprintf("Set %s=%s in %s", k, v, cfgPath), nil)
	}
	return nil
}

// ─── unset ────────────────────────────────────────────────────────────────────

func runUnset(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	cfgPath := resolveConfigFilePath(setGlobal)
	content, err := config.ReadFileOrEmpty(cfgPath)
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to load config: %v", err))
		return err
	}

	anyRemoved := false
	for _, key := range args {
		var removed bool
		content, removed = config.UnsetEnvVar(content, key)
		if removed {
			anyRemoved = true
			formatter.Success(fmt.Sprintf("Unset %s from %s", key, cfgPath), nil)
		} else {
			formatter.Warning(fmt.Sprintf("Key %q not found in [env] of %s", key, cfgPath))
		}
	}

	if anyRemoved {
		if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
			formatter.Error(fmt.Sprintf("Failed to save config: %v", err))
			return err
		}

		// Apply canonical format and taplo formatting
		_, _ = config.FormatFile(cfgPath, false)
	}
	return nil
}
