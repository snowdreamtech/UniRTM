// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package hook

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
)

type HuskyRunner struct{}

func init() {
	RegisterRunner(HuskyRunner{})
}

func (h HuskyRunner) Name() string {
	return "husky"
}

func (h HuskyRunner) Detect(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, ".husky"))
	return err == nil && info.IsDir()
}

func (h HuskyRunner) Install(ctx context.Context, dir string) error {
	return nil
}

func (h HuskyRunner) Run(ctx context.Context, hookName string, stage string, args []string) error {
	var targetScript string
	if hookName == "all" {
		if stage == "" {
			return nil
		}
		targetScript = stage
	} else {
		if stage != "" {
			targetScript = stage
		} else {
			targetScript = hookName
		}
	}

	hookScript := filepath.Join(".husky", targetScript)
	if _, err := os.Stat(hookScript); os.IsNotExist(err) {
		return nil // Hook not defined in Husky
	}

	cmdArgs := append([]string{"exec", "--", "sh", hookScript}, args...)
	cmd := exec.CommandContext(ctx, "unirtm", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
