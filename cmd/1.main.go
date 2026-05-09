// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package cmd contains all the command-line interface definitions and implementations
// for the unirtm application. It uses cobra library to handle command-line parsing,
// flag management, and command execution.
package cmd

import (
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
		PersistentPreRun: setupLogging,
		SilenceUsage:     true,
		SilenceErrors:    true,
	}

	buildFlags()
}

// setupLogging configures the global logging level based on command line flags.
// This function is executed before any command runs, ensuring proper log setup.
//
// The logging level priority is as follows:
// - Quiet flag (highest priority): Disables all logging
// - Verbose flag: Sets debug level logging
// - Default: Sets standard Info level logging
//
// Parameters:
//   - cmd: The Cobra command that is being executed
//   - args: Command line arguments passed to the command
func setupLogging(cmd *cobra.Command, args []string) {
	// Set log level from CLI flags
	// The log level priority is: Quiet > Verbose > Info (default)
	if quiet {
		zerolog.SetGlobalLevel(zerolog.Disabled)
	} else if verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
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
	// Validates: Requirement 8.7
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "show what would happen without making changes")

	// Store the config path in the env package for access by other components
	env.Config = configPath
}
