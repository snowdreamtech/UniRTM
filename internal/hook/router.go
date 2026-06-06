package hook

import (
	"context"
	"fmt"
)

var runners []HookRunner

// RegisterRunner adds a hook engine to the router.
// This is typically called in init() functions of specific runners.
func RegisterRunner(r HookRunner) {
	runners = append(runners, r)
}

// Run iterates through registered runners, detecting the appropriate engine,
// and delegates the execution of the hook to it.
func Run(ctx context.Context, dir string, hookName string, args []string) error {
	for _, r := range runners {
		if r.Detect(dir) {
			fmt.Printf("🔧 UniRTM Hook: Delegating to %s\n", r.Name())
			return r.Run(ctx, hookName, args)
		}
	}

	// If no hook runner is detected, we simply succeed silently.
	// This prevents git commits from blocking just because a repo has no hook config.
	return nil
}
