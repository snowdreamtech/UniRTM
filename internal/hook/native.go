// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package hook

import (
	"context"
	"os"
	"os/exec"

	"github.com/snowdreamtech/unirtm/internal/config"
)

type NativeRunner struct{}

func init() {
	RegisterRunner(NativeRunner{})
}

func (n NativeRunner) Name() string {
	return "unirtm"
}

func (n NativeRunner) Detect(dir string) bool {
	cfg, err := config.LoadHierarchy(dir)
	if err != nil {
		return false
	}
	return len(cfg.Hooks) > 0
}

func (n NativeRunner) Install(ctx context.Context, dir string) error {
	// Bridge script logic is shared across runners; InstallBridgeScript is called from CLI
	return nil
}

func (n NativeRunner) Run(ctx context.Context, hookName string, stage string, args []string) error {
	return n.RunInDir(ctx, ".", hookName, stage, args)
}

// RunInDir executes the hook defined in the config at the given directory.
// Separating dir from Run() eliminates the need for os.Chdir in tests,
// making the function safe for concurrent test execution.
func (n NativeRunner) RunInDir(ctx context.Context, dir string, hookName string, stage string, args []string) error {
	cfg, err := config.LoadHierarchy(dir)
	if err != nil {
		return err
	}

	var targetHook string
	if hookName == "all" {
		if stage == "" {
			return nil
		}
		targetHook = stage
	} else {
		if stage != "" {
			targetHook = hookName // Native maps actual command to hookName key
		} else {
			targetHook = hookName
		}
	}

	hookCmd, ok := cfg.Hooks[targetHook]
	if !ok || hookCmd == "" {
		return nil // No hook defined for this event — silent success
	}

	// Build sh arguments safely:
	//   sh -c '<command> "$@"' -- arg1 arg2 ...
	// The shell receives $@ as positional params, never interpolated into the script string.
	cmdStr := hookCmd
	if len(args) > 0 {
		cmdStr += ` "$@"`
	}

	shArgs := []string{"-c", cmdStr, "--"}
	shArgs = append(shArgs, args...)

	shell := GetShell()
	cmd := exec.CommandContext(ctx, shell, shArgs...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
