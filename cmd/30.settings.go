// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(settingsCmd)
	}
}

// settingsCmd is a shorthand command that provides compatibility with 'mise settings'.
// It delegates to the appropriate config subcommands based on the number of arguments.
var settingsCmd = &cobra.Command{
	Use:   "settings [key] [value]",
	Short: "Manage UniRTM settings (compatibility alias for config)",
	Long: `Manage UniRTM configuration settings.
This command is provided for compatibility with 'mise settings' and acts as a smart wrapper around 'unirtm config'.

Behaviors:
  0 args: unirtm settings               -> unirtm config show
  1 arg:  unirtm settings <key>         -> unirtm config get <key>
  2 args: unirtm settings <key> <value> -> unirtm config set <key> <value>

Examples:
  unirtm settings
  unirtm settings settings.cache_ttl
  unirtm settings settings.cache_ttl 48h`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return runConfigShow(cmd, args)
		} else if len(args) == 1 {
			return runConfigGet(cmd, args)
		} else if len(args) == 2 {
			return runConfigSet(cmd, args)
		}
		return cmd.Help()
	},
}
