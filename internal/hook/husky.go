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

func (h HuskyRunner) Run(ctx context.Context, hookName string, args []string) error {
	hookScript := filepath.Join(".husky", hookName)
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
