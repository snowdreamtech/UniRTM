// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package cmd contains all the command line interface definitions for the unirtm application.
// It handles command registration, flag setup, and dispatching to appropriate handlers.
package cmd

import (
	"fmt"
	"runtime"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/spf13/cobra"
)

// init registers the version command to the root command.
// This function is automatically called when the package is imported.
func init() {
	rootCmd.AddCommand(versionCmd)
}

// versionCmd represents the version command which displays the application version information.
// This command is useful for users to check which version of the software they are running,
// which is essential for reporting bugs and ensuring compatibility.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of " + env.ProjectName,
	Long:  "Display version information including build details, commit hash, and build time.",
	Run:   runVersion,
}

// runVersion prints the version information of the application.
// It includes the OS and architecture, build version, copyright details,
// license information, author details, and build time.
//
// Parameters:
//   - cmd: The cobra command that triggered this function.
//   - args: The arguments passed to the command.
func runVersion(cmd *cobra.Command, args []string) {
	// OSArch represents the current operating system and architecture in the format "GOOS/GOARCH"
	osArch := runtime.GOOS + "/" + runtime.GOARCH

	// Print version information
	fmt.Printf("%s version %s-%s %s\n", env.ProjectName, env.GitTag, env.CommitHash, osArch)
	fmt.Printf("%s\n", env.COPYRIGHT)
	fmt.Printf("License: %s\n", env.LICENSE)
	fmt.Printf("\n")
	fmt.Printf("Written by %s\n", env.Author)
	fmt.Printf("Built at %s\n", env.BuildTime)
}
