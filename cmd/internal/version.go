// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package internal provides core functionality for the unirtm command line tool.
// This package includes version information handling and other internal utilities
// that support the main application commands.
package internal

import (
	"fmt"
	"github.com/common-nighthawk/go-figure"
	"github.com/rs/zerolog"
	"runtime"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
	"github.com/spf13/cobra"
)

// welcome prints a stylized "unirtm" title in green color to the default Gin writer.
// It uses the "figure" package to create an ASCII art title with the "larry3d" font.
// The welcome banner is only displayed when logging is enabled.
func welcome() {
	if zerolog.GlobalLevel() != zerolog.Disabled {
		title := figure.NewColorFigure("UNIRTM", "larry3d", "green", true)
		title.Print()
	}
}

// VersionPreRunFunc initializes the environment before displaying version information.
// This function is called before the main version command runs to set up any
// necessary prerequisites such as logging configuration.
//
// Parameters:
//   - cmd: The cobra command that triggered this function.
//   - args: The arguments passed to the command.
func VersionPreRunFunc(cmd *cobra.Command, args []string) {
	//init logger with default configuration
	logger.InitLogger("", "")
}

// VersionRunFunc prints the version information of the application.
// It includes the OS and architecture, build version, copyright details,
// license information, author details, and build time.
//
// Parameters:
//   - cmd: The cobra command that triggered this function.
//   - args: The arguments passed to the command.
func VersionRunFunc(cmd *cobra.Command, args []string) {
	welcome()

	// OSArch represents the current operating system and architecture in the format "GOOS/GOARCH"
	OSArch := runtime.GOOS + "/" + runtime.GOARCH

	// BuildVersion formats the complete version string including project name, git tag, commit hash, and OS/arch
	BuildVersion := fmt.Sprintf("%s version %s-%s %s\n", env.ProjectName, env.GitTag, env.CommitHash, OSArch)

	// CopyrightDetail contains the copyright information of the application
	CopyrightDetail := fmt.Sprintf("%s\n", env.COPYRIGHT)

	// LicenseDetail contains the license information of the application
	LicenseDetail := fmt.Sprintf("License: %s\n", env.LICENSE)

	// AuthorDetail contains information about the application's author
	AuthorDetail := fmt.Sprintf("Written by %s", env.Author)

	// BuildDetail contains the timestamp when the application was built
	BuildDetail := fmt.Sprintf("Built at %s", env.BuildTime)

	// Create a string builder to efficiently concatenate multiple strings
	var builder strings.Builder

	// Add a blank line for better readability
	builder.WriteString("\n")

	// Add version information to the output
	builder.WriteString(BuildVersion)

	// Add copyright information to the output
	builder.WriteString(CopyrightDetail)

	// Add license information to the output
	builder.WriteString(LicenseDetail)

	// Add a blank line for better readability
	builder.WriteString("\n")

	// Add author information to the output
	builder.WriteString(AuthorDetail)

	// Add a blank line for better readability
	builder.WriteString("\n")

	// Add build time information to the output
	builder.WriteString(BuildDetail)

	// Print the complete version information to the console
	fmt.Println(builder.String())
}
