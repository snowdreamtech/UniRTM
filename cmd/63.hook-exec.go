// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(hookExecCmd)
	}
}

var hookExecCmd = &cobra.Command{
	Use:   "hook-exec <tool> [args...]",
	Short: "Execution wrapper specifically designed for git hooks",
	Long: `A transparent execution wrapper for git hooks (like pre-commit).
It automatically chunks file arguments on Windows to bypass the 8191 character
command-line length limit in cmd.exe. On Unix-like systems, or for short commands,
it executes the command in a single pass.`,
	Args: cobra.MinimumNArgs(1),
	// DisableFlagParsing to pass all flags verbatim to the underlying tool.
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Identify base command and appended files.
		// Pre-commit appends file arguments at the very end. We scan backwards
		// until we find an argument that doesn't exist as a file or directory.
		splitIdx := len(args)
		for i := len(args) - 1; i >= 0; i-- {
			if _, err := os.Lstat(args[i]); err != nil {
				// Not a valid file/directory path, so this and everything before it
				// must be the base command/flags.
				break
			}
			splitIdx = i
		}

		// Ensure at least the tool itself is kept in the baseArgs.
		if splitIdx == 0 && len(args) > 0 {
			splitIdx = 1
		}

		baseArgs := args[:splitIdx]
		fileArgs := args[splitIdx:]

		// Calculate total command length estimate
		totalLen := 0
		for _, arg := range args {
			totalLen += len(arg) + 1
		}

		// On Unix or if total length is safely under the limit, run in a single pass.
		// 7000 is chosen as a conservative safe limit below Windows' 8191.
		if runtime.GOOS != "windows" || len(fileArgs) == 0 || totalLen < 7000 {
			return runExec(cmd, args)
		}

		// Windows chunking logic for long commands
		baseLen := 0
		for _, arg := range baseArgs {
			baseLen += len(arg) + 1
		}

		var currentChunk []string
		currentLen := 0

		for _, file := range fileArgs {
			fileLen := len(file) + 1

			if baseLen+currentLen+fileLen >= 7000 {
				if len(currentChunk) > 0 {
					chunkArgs := append(append([]string(nil), baseArgs...), currentChunk...)
					if err := runExec(cmd, chunkArgs); err != nil {
						// runExec on Windows will os.Exit on error, but just in case
						return err
					}
					currentChunk = nil
					currentLen = 0
				}
			}
			currentChunk = append(currentChunk, file)
			currentLen += fileLen
		}

		// Execute remaining
		if len(currentChunk) > 0 {
			chunkArgs := append(append([]string(nil), baseArgs...), currentChunk...)
			if err := runExec(cmd, chunkArgs); err != nil {
				return err
			}
		}

		return nil
	},
}
