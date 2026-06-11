// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/snowdreamtech/unirtm/internal/hook"
	"github.com/spf13/cobra"
)

var hookAll bool
var hookStage string

var hookCmd = &cobra.Command{
	Use:   "hook",
	Short: "Manage and execute git hooks",
	Long: `Manage git hooks using UniRTM's multi-engine runner.
This allows you to seamlessly integrate pre-commit, husky, lefthook, or native UniRTM hooks.`,
}

var hookInstallCmd = &cobra.Command{
	Use:   "install [hookName]",
	Short: "Install a git hook bridge script",
	Long: `Injects the UniRTM bridge script into .git/hooks/<hookName> to enable routing.
Use the -a/--all flag to install the bridge script to all non-sample hooks in .git/hooks.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pwd, err := os.Getwd()
		if err != nil {
			return err
		}

		if hookAll {
			updated, err := hook.InstallAllBridgeScripts(cmd.Context(), pwd)
			if err != nil {
				return fmt.Errorf("failed to install hooks: %w", err)
			}
			if len(updated) == 0 {
				fmt.Println("No active hooks found to update.")
			} else {
				fmt.Printf("✨ Successfully installed UniRTM bridge script for hooks: %v\n", updated)
			}
			return nil
		}

		if len(args) == 0 {
			return fmt.Errorf("requires at least 1 arg(s), only received 0")
		}

		hookName := args[0]
		if err := hook.ValidateHookName(hookName); err != nil {
			return err
		}

		err = hook.InstallBridgeScript(cmd.Context(), pwd, hookName)
		if err != nil {
			return fmt.Errorf("failed to install hook %s: %w", hookName, err)
		}

		fmt.Printf("✨ Successfully installed UniRTM bridge script for hook: %s\n", hookName)
		return nil
	},
}

var hookRunCmd = &cobra.Command{
	Use:   "run [hookName] [--stage stage] [args...]",
	Short: "Run a specific git hook",
	Long:  `Executes the specified git hook by routing it to the detected runner engine.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		hookName := args[0]
		if hookName != "all" {
			if err := hook.ValidateHookName(hookName); err != nil {
				return err
			}
		}
		hookArgs := args[1:]

		pwd, err := os.Getwd()
		if err != nil {
			return err
		}

		return hook.Run(context.Background(), pwd, hookName, hookStage, hookArgs)
	},
}

func init() {
	rootCmd.AddCommand(hookCmd)
	hookInstallCmd.Flags().BoolVarP(&hookAll, "all", "a", false, "Install bridge scripts to all non-sample hooks in .git/hooks")
	hookRunCmd.Flags().StringVar(&hookStage, "stage", "", "Git lifecycle stage (e.g. pre-commit)")
	hookRunCmd.Flags().SetInterspersed(false)
	hookCmd.AddCommand(hookInstallCmd)
	hookCmd.AddCommand(hookRunCmd)
}
