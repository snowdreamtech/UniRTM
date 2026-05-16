// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/spf13/cobra"
)

var (
	editGlobal bool
)

func init() {
	editCmd.Flags().BoolVar(&editGlobal, "global", false, "edit the global config file")
	if rootCmd != nil {
		rootCmd.AddCommand(editCmd)
	}
}

// editCmd opens the config file in $EDITOR.
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open the config file in $EDITOR",
	Long: `Open the UniRTM config file in $EDITOR (or $VISUAL).

By default opens the project-level unirtm.toml. Use --global to edit
the global config at ~/.config/unirtm/unirtm.toml.

Falls back to vi if no editor is configured.

Examples:
  # Edit project config
  unirtm edit

  # Edit global config
  unirtm edit --global

  # Use a specific editor
  EDITOR=nano unirtm edit`,
	Args: cobra.NoArgs,
	RunE: runEdit,
}

func runEdit(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
	})

	cfgPath := resolveConfigFilePath(editGlobal)

	// Ensure the file exists before opening.
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		if err := os.WriteFile(cfgPath, []byte("# UniRTM configuration\n"), 0o644); err != nil {
			formatter.Error(fmt.Sprintf("Could not create %s: %v", cfgPath, err))
			return err
		}
		formatter.Info(fmt.Sprintf("Created new config file: %s", cfgPath), nil)
	}

	cfg, _ := config.Load()
	editor := env.Get("VISUAL")
	if editor == "" {
		editor = env.Get("EDITOR")
	}
	if editor == "" && cfg != nil && cfg.Settings.Editor != "" {
		editor = cfg.Settings.Editor
	}
	if editor == "" {
		editor = "vi"
	}

	c := exec.Command(editor, cfgPath)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
