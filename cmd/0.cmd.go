// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"go.uber.org/automaxprocs/maxprocs"
)

var (
	rootCmd *cobra.Command
)

func init() {
	// Disable automaxprocs log
	// https://github.com/uber-go/automaxprocs/issues/19#issuecomment-557382150
	nopLog := func(string, ...interface{}) {}
	maxprocs.Set(maxprocs.Logger(nopLog))
}

// Execute executes the root command.
func Execute() {
	// Handle asdf alias/symlink
	if filepath.Base(os.Args[0]) == "asdf" {
		if len(os.Args) > 1 {
			command := os.Args[1]
			switch command {
			case "reshim", "update-nodebuild", "update-ruby-build":
				// Silently succeed for these commands common in plugins
				os.Exit(0)
			default:
				// For other asdf commands, we could eventually map them to unirtm commands
				fmt.Printf("unirtm (as asdf) - intercepted %s command\n", command)
				os.Exit(0)
			}
		}
		os.Exit(0)
	}

	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
