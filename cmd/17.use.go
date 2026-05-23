// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	// useGlobal writes the tool version to the global config (~/.config/unirtm/config.toml)
	useGlobal bool
	// usePath specifies the directory to write the config file into
	usePath string
	// useForce forces reinstallation of the tool version
	useForce bool
	// useEnv specifies an environment-specific config file (e.g. unirtm.<env>.toml)
	useEnv string
	// usePin resolves fuzzy/prefix versions to precise concrete versions in the config file
	usePin bool
)

// init registers the use command to the root command.
func init() {
	useCmd.Flags().BoolVarP(&useGlobal, "global", "g", false, "write to global config (~/.config/unirtm/config.toml)")
	useCmd.Flags().StringVarP(&usePath, "path", "p", "", "directory to write config file into (default: current directory)")
	useCmd.Flags().BoolVarP(&useForce, "force", "f", false, "force reinstall even if the tool is already installed")
	useCmd.Flags().StringVarP(&useEnv, "env", "e", "", "environment-specific config file (e.g. unirtm.<env>.toml)")
	useCmd.Flags().BoolVarP(&usePin, "pin", "P", false, "resolve prefix/alias/range to precise concrete version")

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

	// Initialize installation manager for interactive selection if needed
	cfg, _ := config.Load()
	ctx := context.Background()
	backendRegistry := backend.NewRegistry()
	im, _ := getInstallationManager(ctx, cfg)

	// Parse tool@version pairs
	type toolVersion struct {
		key     string // key in config (e.g. github:org/repo)
		version string
	}
	pairs := make([]toolVersion, 0, len(args))
	for _, arg := range args {
		backendName, toolName, version, explicit := im.ParseToolSpec(arg)

		// If no explicit version, try interactive selection
		if !explicit || version == "latest" {
			isTerminal := term.IsTerminal(int(os.Stdin.Fd()))
			if !jsonOutput && isTerminal && im != nil {
				selected, err := im.SelectVersionInteractive(cmd.Context(), toolName, backendName)
				if err == nil {
					version = selected
				} else {
					return fmt.Errorf("interactive selection failed for %s: %w", toolName, err)
				}
			} else if version == "" || version == "latest" {
				// Try to get latest from backend if possible
				b, err := backendRegistry.Get(backendName)
				if err == nil {
					platform := backend.CurrentPlatform()
					if info, err := b.ResolveVersion(cmd.Context(), toolName, "latest", platform); err == nil && info != nil {
						version = info.Version
					}
				}
			}
		} else if usePin {
			// Resolve fuzzy/range/prefix/alias to exact concrete version
			b, err := backendRegistry.Get(backendName)
			if err == nil {
				platform := backend.CurrentPlatform()
				if info, err := b.ResolveVersion(cmd.Context(), toolName, version, platform); err == nil && info != nil {
					version = info.Version
				}
			}
		}

		if version == "" || version == "latest" {
			return fmt.Errorf("invalid format %q: expected <tool>@<version> (e.g. node@20.0.0)", arg)
		}

		// Reconstruct key for config: backend:tool (if backend is not auto-detected or is explicit)
		configKey := toolName
		if backendName != "" && (strings.Contains(arg, ":") || backendName == "github") {
			// If it's github or explicit backend, use backend:tool as key
			// but for github, we often just use owner/repo which is auto-detected.
			// Actually, let's keep it simple: if it contains '/', it's github.
			if strings.Contains(toolName, "/") {
				configKey = "github:" + toolName
			} else if strings.Contains(arg, ":") {
				// Was explicit backend in input
				origBackend := strings.SplitN(arg, ":", 2)[0]
				configKey = origBackend + ":" + toolName
			}
		}

		pairs = append(pairs, toolVersion{key: configKey, version: version})
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
	configFile := findOrCreateConfigFile(targetDir, useEnv)

	if dryRun {
		for _, p := range pairs {
			formatter.Info(fmt.Sprintf("[dry-run] Would write %s = %q to %s", p.key, p.version, configFile), nil)
		}
		return nil
	}

	// Ensure directory exists
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("create config directory %s: %w", targetDir, err)
	}

	// Read existing content or create empty
	content, err := config.ReadFileOrEmpty(configFile)
	if err != nil {
		return fmt.Errorf("read config file: %w", err)
	}

	// Apply each tool@version into the [tools] section
	for _, p := range pairs {
		content = config.UpsertToolVersion(content, p.key, p.version)
	}

	// Write back
	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("write config file %s: %w", configFile, err)
	}

	// Apply canonical format and taplo formatting to ensure correct block order
	_, _ = config.FormatFile(configFile, false)

	for _, p := range pairs {
		formatter.Success(fmt.Sprintf("Set %s = %q in %s", p.key, p.version, configFile), map[string]interface{}{
			"tool":    p.key,
			"version": p.version,
			"file":    configFile,
		})
	}

	// Automatically install the tool versions if they are not already installed
	if im != nil {
		for _, p := range pairs {
			// Extract tool name and backend name from key
			toolName := p.key
			backendName := ""
			if strings.Contains(toolName, ":") {
				parts := strings.SplitN(toolName, ":", 2)
				backendName = parts[0]
				toolName = parts[1]
			}

			// If force is enabled, perform clean uninstallation first if it is installed
			if useForce {
				alreadyOnDisk, _ := im.IsInstalled(ctx, toolName, p.version, backendName)
				if alreadyOnDisk {
					formatter.Info(fmt.Sprintf("Tool %s@%s is already installed. [force] Uninstalling first...", toolName, p.version), nil)
					_ = im.Uninstall(ctx, toolName, p.version)
				}
			}

			isInstalled, _ := im.IsInstalled(ctx, toolName, p.version, backendName)
			if !isInstalled {
				formatter.Info(fmt.Sprintf("Tool %s@%s is not installed. Installing now...", toolName, p.version), nil)
				if err := im.Install(ctx, p.key, toolName, p.version, backendName); err != nil {
					return fmt.Errorf("failed to automatically install %s@%s: %w", toolName, p.version, err)
				}
			} else {
				formatter.Success(fmt.Sprintf("Tool %s@%s is already installed", toolName, p.version), nil)
			}
		}
	}

	return nil
}

// findOrCreateConfigFile finds an existing config file in the directory,
// or returns the path to a new unirtm.toml to be created, supporting env-specific name config.
func findOrCreateConfigFile(dir, envName string) string {
	var names []string
	if envName != "" {
		names = []string{fmt.Sprintf(".unirtm.%s.toml", envName), fmt.Sprintf("unirtm.%s.toml", envName)}
	} else {
		names = []string{".unirtm.toml", "unirtm.toml"}
	}
	for _, name := range names {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	if envName != "" {
		return filepath.Join(dir, fmt.Sprintf("unirtm.%s.toml", envName))
	}
	return filepath.Join(dir, "unirtm.toml")
}
