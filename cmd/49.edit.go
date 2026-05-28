// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/spf13/cobra"
)

var (
	editGlobal bool
	editFile   string
)

func init() {
	editCmd.Flags().BoolVarP(&editGlobal, "global", "g", false, "Edit the global config file")
	editCmd.Flags().StringVarP(&editFile, "file", "f", "", "Edit a specific config file")
	if rootCmd != nil {
		rootCmd.AddCommand(editCmd)
	}
}

// editCmd opens the config file in $EDITOR.
var editCmd = &cobra.Command{
	Use:   "edit [FILE]",
	Short: "Open the config file in $EDITOR",
	Long: `Open a UniRTM config file in your preferred editor.

If no file is specified, it intelligently discovers all relevant config files
in the current hierarchy and lets you choose interactively.

Priority for finding an editor:
1.  --global / --file flag or argument
2.  UNIRTM_EDITOR environment variable
3.  VISUAL environment variable
4.  EDITOR environment variable
5.  UniRTM settings.editor configuration
6.  Standard system defaults (vim, nano, notepad, etc.)

Examples:
  # Interactively choose which config to edit
  unirtm edit

  # Edit the global config (~/.config/unirtm/unirtm.toml)
  unirtm edit --global

  # Edit a specific file
  unirtm edit .unirtm.toml`,
	Args: cobra.MaximumNArgs(1),
	RunE: runEdit,
}

func runEdit(cmd *cobra.Command, args []string) error {
	// 3. Find editor
	cfg, _ := config.Load()
	editor, source := getBestEditorWithSource(cfg)

	pterm.Info.Printf("Opening configuration editor (using %s via %s)...\n", pterm.Bold.Sprint(editor), source)

	// Add a tip if we are using a fallback/system default
	if source == "system default" || source == "fallback" {
		fmt.Printf("%s Set $EDITOR or run 'unirtm settings set editor <editor>' to change your preference.\n\n", pterm.FgGray.Sprint("Tip:"))
	}

	targetFile := ""

	// 1. Determine which file to edit
	if len(args) > 0 {
		targetFile = args[0]
	} else if editFile != "" {
		targetFile = editFile
	} else if editGlobal {
		targetFile = env.GetGlobalConfigPath()
	} else {
		// Discover files interactively
		candidates := discoverConfigFiles()
		if len(candidates) == 0 {
			pterm.Info.Println("No configuration files found. Creating a new one in the current directory.")
			targetFile = "unirtm.toml"
		} else if len(candidates) == 1 {
			targetFile = candidates[0].Path
		} else {
			// Show interactive selection
			var options []string
			pathMap := make(map[string]string)
			for _, c := range candidates {
				opt := fmt.Sprintf("%-15s %s", pterm.Bold.Sprint(c.Name), pterm.FgGray.Sprint(c.Path))
				options = append(options, opt)
				pathMap[opt] = c.Path
			}

			selected, _ := pterm.DefaultInteractiveSelect.
				WithDefaultText("Select configuration file to edit").
				WithOptions(options).
				Show()

			targetFile = pathMap[selected]
		}
	}

	// 2. Ensure file exists
	if _, err := os.Stat(targetFile); os.IsNotExist(err) {
		pterm.Info.Printf("Creating new config file: %s\n", targetFile)
		if err := os.MkdirAll(filepath.Dir(targetFile), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
		if err := os.WriteFile(targetFile, []byte("# UniRTM Configuration\n\n[tools]\n# tool = \"version\"\n"), 0644); err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
	}

	// 4. Edit Loop (with validation)
	for {
		c := exec.Command(editor, targetFile)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		if err := c.Run(); err != nil {
			return fmt.Errorf("editor failed: %w", err)
		}

		// Validate TOML
		data, err := os.ReadFile(targetFile)
		if err != nil {
			return fmt.Errorf("failed to read file after edit: %w", err)
		}

		var m map[string]interface{}
		if err := toml.Unmarshal(data, &m); err != nil {
			pterm.Error.Printf("Invalid TOML syntax: %v\n", err)
			confirm, _ := pterm.DefaultInteractiveConfirm.
				WithDefaultText("Do you want to re-edit to fix the error?").
				WithDefaultValue(true).
				Show()

			if confirm {
				continue
			} else {
				pterm.Warning.Println("Changes saved with syntax errors. They may fail to load.")
				break
			}
		}

		pterm.FgGreen.Printf("Configuration saved and validated: %s\n", targetFile)
		break
	}

	return nil
}

type configCandidate struct {
	Name string
	Path string
}

func discoverConfigFiles() []configCandidate {
	var candidates []configCandidate

	// Local candidates
	cwd, _ := os.Getwd()
	localFiles := []string{".unirtm.toml", "unirtm.toml", ".mise.toml", "mise.toml"}
	for _, f := range localFiles {
		p := filepath.Join(cwd, f)
		if _, err := os.Stat(p); err == nil {
			candidates = append(candidates, configCandidate{Name: f, Path: p})
		}
	}

	// Global candidate
	globalPath := env.GetGlobalConfigPath()
	if _, err := os.Stat(globalPath); err == nil {
		candidates = append(candidates, configCandidate{Name: "global", Path: globalPath})
	}

	return candidates
}
