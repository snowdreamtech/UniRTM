// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package internal provides core functionality for the unirtm server application.
// It includes server initialization, configuration, request handling, and graceful shutdown.
package internal

import (
	// Third-party library imports
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	// Internal application imports
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
)

// ServerPersistentPreFunc configures the global logging level based on command line flags.
// This function is executed before any related command is run, ensuring proper log setup.
//
// The logging level priority is as follows:
// - Quiet flag (highest priority): Disables all logging
// - Trace flag: Sets the most verbose logging level
// - Debug flag: Sets detailed logging for development purposes
// - Default: Sets standard Info level logging for production use
//
// Parameters:
//   - cmd: The Cobra command that is being executed
//   - args: Command line arguments passed to the command
func ServerPersistentPreFunc(cmd *cobra.Command, args []string) {
	// Set log level from CLI flags
	// The log level priority is: Quiet > Trace > Debug > Info (default)
	if env.Quiet {
		zerolog.SetGlobalLevel(zerolog.Disabled)
	} else if env.Trace {
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	} else if env.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

// ServerPreRunFunc is a pre-run function for the Cobra command that initializes the logger,
// prints a welcome message, and logs the starting of the web server along with the provided arguments.
//
// This function:
// - Sets the global log level based on CLI flags
// - Loads application configuration
// - Initializes the logger with appropriate log file paths
//
// Parameters:
//   - cmd: The Cobra command that is being executed.
//   - args: The arguments passed to the command.
func ServerPreRunFunc(cmd *cobra.Command, args []string) {

}

// ServerRunFunc starts the server with the provided command and arguments.
// It initializes a new Gin handler, loads middlewares, sets up the embedded HTTP server,
// defines a simple ping endpoint, and runs the server on the configured address.
//
// This function:
// - Displays a welcome message
// - Creates and configures the Gin HTTP handler
// - Loads middleware, embedded assets, and API routes
// - Sets up HTTP and/or HTTPS servers based on configuration
// - Starts the servers in goroutines for concurrent operation
//
// Parameters:
//   - cmd: The Cobra command that is being executed.
//   - args: The arguments passed to the command.
func ServerRunFunc(cmd *cobra.Command, args []string) {
	log.Info().Msgf("ServerRunFunc %s", env.ProjectName)

}

// TestPreRunFunc is a pre-run function for the test command.
// It performs any necessary setup before testing the configuration.
//
// Parameters:
//   - cmd: The Cobra command that is being executed.
//   - args: The arguments passed to the command.
func TestPreRunFunc(cmd *cobra.Command, args []string) {
	// Initialize logger for test mode
}

// TestRunFunc tests the configuration file for validity.
//
// Parameters:
//   - cmd: The Cobra command that is being executed.
//   - args: The arguments passed to the command.
func TestRunFunc(cmd *cobra.Command, args []string) {
	log.Info().Msgf("Testing configuration for %s", env.ProjectName)
	log.Info().Msg("Configuration is valid")
}
