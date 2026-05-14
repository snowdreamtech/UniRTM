// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/provider"
	"github.com/snowdreamtech/unirtm/internal/service"
	"github.com/spf13/cobra"
)

var (
	hookEnvShell string
)

func init() {
	hookEnvCmd.Flags().StringVarP(&hookEnvShell, "shell", "s", "", "shell type (bash, zsh, fish, powershell)")
	if rootCmd != nil {
		rootCmd.AddCommand(hookEnvCmd)
	}
}

// hookEnvCmd handles environment changes on directory navigation.
var hookEnvCmd = &cobra.Command{
	Use:    "hook-env",
	Short:  "Internal command to handle environment changes on directory navigation",
	Hidden: true,
	RunE:   runHookEnv,
}

func runHookEnv(cmd *cobra.Command, args []string) error {
	// We use stderr for log output so stdout remains clean for eval
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stderr,
		Quiet:   true, // Usually quiet for hook
	})

	ctx := context.Background()

	// 1. Detect shell
	shell := hookEnvShell
	if shell == "" {
		detected, err := service.DetectShell()
		if err != nil {
			shell = "bash"
		} else {
			shell = string(detected)
		}
	}
	shellType := service.ShellType(shell)

	// 2. Reconstruct current environment state from env vars
	currentState := &service.EnvironmentState{
		ProjectDir:   os.Getenv("UNIRTM_PROJECT_DIR"),
		ToolVersions: make(map[string]string),
		EnvVars:      make(map[string]string),
	}

	// Extract tool versions from environment (UNIRTM_XXX_VERSION)
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) == 2 {
			key := pair[0]
			if strings.HasPrefix(key, "UNIRTM_") && strings.HasSuffix(key, "_VERSION") {
				toolName := strings.TrimPrefix(key, "UNIRTM_")
				toolName = strings.TrimSuffix(toolName, "_VERSION")
				currentState.ToolVersions[toolName] = pair[1]
			}
		}
	}

	// 3. Initialize AutoActivationManager
	shimsDir := env.GetShimsDir()
	dataDir := env.GetDataDir()
	registry := provider.NewRegistry()
	am := service.NewActivationManager(shimsDir, dataDir, registry)
	aam := service.NewAutoActivationManager(am)

	// 4. Handle directory change
	pwd, _ := os.Getwd()
	oldPwd := os.Getenv("UNIRTM_OLD_PWD")
	
	event := service.DirectoryChangeEvent{
		OldDir: oldPwd,
		NewDir: pwd,
		Shell:  shellType,
	}

	change, err := aam.HandleDirectoryChange(ctx, event, currentState)
	if err != nil {
		formatter.Error("Failed to handle directory change", map[string]interface{}{"error": err.Error()})
		return err
	}

	// 5. Output the script
	if change.Action != service.ActionNone {
		fmt.Print(change.Script)
		
		// Update project dir and old pwd in the shell
		switch shellType {
		case service.ShellFish:
			fmt.Printf("set -gx UNIRTM_PROJECT_DIR \"%s\"\n", change.NewState.ProjectDir)
			fmt.Printf("set -gx UNIRTM_OLD_PWD \"%s\"\n", pwd)
		case service.ShellPowerShell:
			fmt.Printf("$env:UNIRTM_PROJECT_DIR = \"%s\"\n", change.NewState.ProjectDir)
			fmt.Printf("$env:UNIRTM_OLD_PWD = \"%s\"\n", pwd)
		default: // POSIX
			fmt.Printf("export UNIRTM_PROJECT_DIR=\"%s\"\n", change.NewState.ProjectDir)
			fmt.Printf("export UNIRTM_OLD_PWD=\"%s\"\n", pwd)
		}
	}

	return nil
}
