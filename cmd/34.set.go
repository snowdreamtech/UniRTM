// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
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

// loadRawTOML reads a TOML file into a generic map.
// Returns an empty map when the file does not yet exist.
func loadRawTOML(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return make(map[string]interface{}), nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var m map[string]interface{}
	if err := toml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if m == nil {
		m = make(map[string]interface{})
	}
	return m, nil
}

// saveRawTOML writes a generic map to a TOML file.
func saveRawTOML(path string, m map[string]interface{}) error {
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	if err := enc.Encode(m); err != nil {
		return fmt.Errorf("encode TOML: %w", err)
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

// envSection returns (or creates) the [env] sub-map.
func envSection(m map[string]interface{}) map[string]interface{} {
	if raw, ok := m["env"]; ok {
		if envMap, ok := raw.(map[string]interface{}); ok {
			return envMap
		}
	}
	envMap := make(map[string]interface{})
	m["env"] = envMap
	return envMap
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
	m, err := loadRawTOML(cfgPath)
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to load config: %v", err))
		return err
	}

	envMap := envSection(m)
	for k, v := range pairs {
		envMap[k] = v
	}

	if err := saveRawTOML(cfgPath, m); err != nil {
		formatter.Error(fmt.Sprintf("Failed to save config: %v", err))
		return err
	}

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
	m, err := loadRawTOML(cfgPath)
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to load config: %v", err))
		return err
	}

	envMap := envSection(m)
	for _, key := range args {
		if _, exists := envMap[key]; !exists {
			formatter.Warning(fmt.Sprintf("Key %q not found in [env] of %s", key, cfgPath))
			continue
		}
		delete(envMap, key)
		formatter.Success(fmt.Sprintf("Unset %s from %s", key, cfgPath), nil)
	}

	if len(envMap) == 0 {
		delete(m, "env")
	}

	if err := saveRawTOML(cfgPath, m); err != nil {
		formatter.Error(fmt.Sprintf("Failed to save config: %v", err))
		return err
	}
	return nil
}
