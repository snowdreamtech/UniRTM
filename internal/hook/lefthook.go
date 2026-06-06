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

func (l LefthookRunner) Run(ctx context.Context, hookName string, args []string) error {
	cmdArgs := append([]string{"exec", "--", "lefthook", "run", hookName}, args...)
	cmd := exec.CommandContext(ctx, "unirtm", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
