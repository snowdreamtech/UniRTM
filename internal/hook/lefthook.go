// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package hook

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
)

type LefthookRunner struct{}

func init() {
	RegisterRunner(LefthookRunner{})
}

func (l LefthookRunner) Name() string {
	return "lefthook"
}

func (l LefthookRunner) Detect(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "lefthook.yml"))
	return err == nil
}

func (l LefthookRunner) Install(ctx context.Context, dir string) error {
	return nil
}

func (l LefthookRunner) Run(ctx context.Context, hookName string, stage string, args []string) error {
	var cmdArgs []string
	if hookName == "all" {
		if stage == "" {
			cmdArgs = []string{"exec", "--", "lefthook", "run"}
		} else {
			cmdArgs = []string{"exec", "--", "lefthook", "run", stage}
		}
	} else {
		if stage == "" {
			cmdArgs = []string{"exec", "--", "lefthook", "run", hookName}
		} else {
			cmdArgs = []string{"exec", "--", "lefthook", "run", stage, "--commands", hookName}
		}
	}
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.CommandContext(ctx, "unirtm", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
