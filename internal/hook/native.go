package hook

import (
	"context"
	"os"
	"os/exec"
	"strings"

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
	// Bridge script logic is shared across runners, so InstallHook is called from CLI
	return nil
}

func (n NativeRunner) Run(ctx context.Context, hookName string, args []string) error {
	cfg, err := config.LoadHierarchy(".")
	if err != nil {
		return err
	}

	hookCmd, ok := cfg.Hooks[hookName]
	if !ok || hookCmd == "" {
		return nil // No hook defined for this event, silent success
	}

	cmdStr := hookCmd
	if len(args) > 0 {
		cmdStr += ` "$@"`
	}

	shArgs := []string{"-c", cmdStr, "--"}
	shArgs = append(shArgs, args...)

	cmd := exec.CommandContext(ctx, "sh", shArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
