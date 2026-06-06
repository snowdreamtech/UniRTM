package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/snowdreamtech/unirtm/internal/hook"
	"github.com/spf13/cobra"
)

var hookCmd = &cobra.Command{
	Use:   "hook",
	Short: "Manage and execute git hooks",
	Long: `Manage git hooks using UniRTM's multi-engine runner.
This allows you to seamlessly integrate pre-commit, husky, lefthook, or native UniRTM hooks.`,
}

var hookInstallCmd = &cobra.Command{
	Use:   "install [hookName]",
	Short: "Install a git hook bridge script",
	Long:  `Injects the UniRTM bridge script into .git/hooks/<hookName> to enable routing.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		hookName := args[0]
		if err := hook.ValidateHookName(hookName); err != nil {
			return err
		}
		pwd, err := os.Getwd()
		if err != nil {
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
	Use:   "run [hookName]",
	Short: "Run a specific git hook",
	Long:  `Executes the specified git hook by routing it to the detected runner engine.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		hookName := args[0]
		if err := hook.ValidateHookName(hookName); err != nil {
			return err
		}
		hookArgs := args[1:]
		
		pwd, err := os.Getwd()
		if err != nil {
			return err
		}

		return hook.Run(context.Background(), pwd, hookName, hookArgs)
	},
}

func init() {
	rootCmd.AddCommand(hookCmd)
	hookCmd.AddCommand(hookInstallCmd)
	hookCmd.AddCommand(hookRunCmd)
}
