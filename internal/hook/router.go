// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package hook

import (
	"context"
	"fmt"
	"sort"
)

var runners []HookRunner

func getPriority(name string) int {
	switch name {
	case "lefthook":
		return 100
	case "husky":
		return 90
	case "pre-commit":
		return 80
	case "unirtm":
		return 10
	default:
		return 0
	}
}

// RegisterRunner adds a hook engine to the router.
// This is typically called in init() functions of specific runners.
func RegisterRunner(r HookRunner) {
	runners = append(runners, r)
}

// Run iterates through registered runners, detecting the appropriate engine,
// and delegates the execution of the hook to it.
func Run(ctx context.Context, dir string, hookName string, stage string, args []string) error {
	// Sort runners by priority descending to ensure deterministic execution
	sort.SliceStable(runners, func(i, j int) bool {
		return getPriority(runners[i].Name()) > getPriority(runners[j].Name())
	})

	for _, r := range runners {
		if r.Detect(dir) {
			fmt.Printf("🔧 UniRTM Hook: Delegating to %s\n", r.Name())
			return r.Run(ctx, hookName, stage, args)
		}
	}

	// If no hook runner is detected, we simply succeed silently.
	// This prevents git commits from blocking just because a repo has no hook config.
	return nil
}
