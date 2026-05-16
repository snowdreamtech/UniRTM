// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package cmd contains all the command-line interface definitions and implementations
// for the unirtm application. It uses cobra library to handle command-line parsing,
// flag management, and command execution.
package cmd

import (
	"log"
	"os"

	"github.com/pterm/pterm"
	"github.com/rs/zerolog"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/spf13/cobra"
)

var (
	// configPath specifies the path to the configuration file
	configPath string

	// verbose enables verbose output (debug logging)
	verbose bool

	// quiet enables quiet mode (minimal output)
	quiet bool

	// jsonOutput enables JSON output format
	jsonOutput bool

	// dryRun enables dry-run mode — shows what would happen without side effects
	dryRun bool

	// cwd specifies the current working directory for the application
	cwd string

	// envName specifies the environment name for loading environment-specific configs
	envName string

	// jobs specifies the number of parallel jobs to run
	jobs int

	// yes indicates whether to automatically answer yes to all confirmation prompts
	yes bool

	// locked indicates whether to require lockfile URLs to be present during installation
	locked bool

	// silent indicates whether to suppress all output and non-error messages
	silent bool
)

// init initializes the root command for the UniRTM CLI application.
// It sets up the command structure, global flags, and persistent pre-run hooks
// for logging configuration.
func init() {
	rootCmd = &cobra.Command{
		Use:   "unirtm",
		Short: "Universal Runtime Manager - Manage development tool versions",
		Long: `UniRTM (Universal Runtime Manager) is a development environment management tool
that manages multiple development tool versions, provides declarative configuration
management, supports multiple backends and providers, and offers comprehensive
audit and logging capabilities.`,
		PersistentPreRun: setupGlobalOptions,
		SilenceUsage:     true,
		SilenceErrors:    true,
		Args:             cobra.ArbitraryArgs,
		SuggestionsMinimumDistance: 2,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				// If the first argument is not a known command, treat it as a task name.
				// We delegate to the 'run' command implementation.
				return runTaskCommand(cmd, args)
			}
			return cmd.Help()
		},
	}

	buildFlags()
}

// setupGlobalOptions configures the global logging level and handles global flags like --cd.
// This function is executed before any command runs.
func setupGlobalOptions(cmd *cobra.Command, args []string) {
	// Handle --cd (change directory)
	if cwd != "" {
		if err := os.Chdir(cwd); err != nil {
			log.Fatalf("failed to change directory to %s: %v", cwd, err)
		}
	}

	// Set log level from CLI flags
	// The log level priority is: Silent > Quiet > Verbose > Info (default)
	if silent {
		zerolog.SetGlobalLevel(zerolog.Disabled)
		pterm.DisableColor()
		pterm.DisableStyling()
	} else if quiet {
		zerolog.SetGlobalLevel(zerolog.Disabled)
		pterm.DisableColor()
	} else if verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Sync settings to env package
	env.Cwd = cwd
	env.EnvName = envName
	env.Jobs = jobs
	env.Yes = yes
	env.Locked = locked
	env.Silent = silent
	env.Quiet = quiet
	env.Config = configPath

	// Synchronize prefixed environment variables to native ones
	syncEnv()
}

// syncEnv synchronizes prefixed environment variables (UNIRTM_ and MISE_) to native ones
// if the native ones are not already set. This ensures compatibility with third-party
// libraries and tools that only check standard environment variables.
func syncEnv() {
	vars := []string{
		"HTTP_PROXY",
		"HTTPS_PROXY",
		"NO_PROXY",
		"GITHUB_TOKEN",
		"GITHUB_PROXY",
		"GITHUB_API_TOKEN",
		"NO_COLOR",
		"TERM",
		"EDITOR",
		"VISUAL",
		"SHELL",
	}

	for _, v := range vars {
		if os.Getenv(v) == "" {
			if val := env.Get(v); val != "" {
				os.Setenv(v, val)
			}
		}
	}
}

// buildFlags configures all the global command-line flags available in the application.
// These persistent flags are available to all commands and subcommands.
// Each flag is bound to a specific variable that will be set when the flag is used.
func buildFlags() {
	// Config flag: Specifies the path to the configuration file
	// Default: .unirtm.toml or unirtm.toml in the current directory
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "config file path (default: .unirtm.toml or unirtm.toml)")

	// Verbose flag: Enables verbose output with debug logging
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output (debug logging)")

	// Quiet flag: Enables quiet mode with minimal output
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "enable quiet mode (minimal output)")

	// JSON flag: Enables JSON output format for scripting
	rootCmd.PersistentFlags().BoolVarP(&jsonOutput, "json", "j", false, "enable JSON output format")

	// Dry-run flag: Shows what would happen without making any changes
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "show what would happen without making changes")

	// CD flag: Change directory before running command
	rootCmd.PersistentFlags().StringVarP(&cwd, "cd", "C", "", "change directory before running command")

	// Env flag: Set the environment for loading configuration
	rootCmd.PersistentFlags().StringVarP(&envName, "env", "E", "", "set the environment for loading configuration")

	// Jobs flag: How many jobs to run in parallel
	rootCmd.PersistentFlags().IntVar(&jobs, "jobs", 8, "how many jobs to run in parallel")

	// Yes flag: Answer yes to all confirmation prompts
	rootCmd.PersistentFlags().BoolVarP(&yes, "yes", "y", false, "answer yes to all confirmation prompts")

	// Locked flag: Require lockfile URLs to be present during installation
	rootCmd.PersistentFlags().BoolVar(&locked, "locked", false, "require lockfile URLs to be present during installation")

	// Silent flag: Suppress all task output and non-error messages
	rootCmd.PersistentFlags().BoolVar(&silent, "silent", false, "suppress all task output and non-error messages")

	// No-config, no-env, no-hooks flags (placeholders for now to match mise surface)
	rootCmd.PersistentFlags().Bool("no-config", false, "do not load any config files")
	rootCmd.PersistentFlags().Bool("no-env", false, "do not load environment variables from config files")
	rootCmd.PersistentFlags().Bool("no-hooks", false, "do not execute hooks from config files")
}
