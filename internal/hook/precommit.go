// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package hook

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
)

type PreCommitRunner struct{}

func init() {
	RegisterRunner(PreCommitRunner{})
}

func (p PreCommitRunner) Name() string {
	return "pre-commit"
}

func (p PreCommitRunner) Detect(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".pre-commit-config.yaml"))
	return err == nil
}

func (p PreCommitRunner) Install(ctx context.Context, dir string) error {
	return nil
}

func (p PreCommitRunner) Run(ctx context.Context, hookName string, stage string, args []string) error {
	var cmdArgs []string
	if hookName == "all" {
		cmdArgs = []string{"exec", "--", "pre-commit", "run"}
		if stage != "" {
			cmdArgs = append(cmdArgs, "--hook-stage", stage)
		}
	} else {
		cmdArgs = []string{"exec", "--", "pre-commit", "run", hookName}
		if stage != "" {
			cmdArgs = append(cmdArgs, "--hook-stage", stage)
		}
	}
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.CommandContext(ctx, "unirtm", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
