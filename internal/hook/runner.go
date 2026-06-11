// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package hook

import "context"

// HookRunner defines the interface for all supported git-hook engines
type HookRunner interface {
	// Detect returns true if this engine's config exists in the workspace
	Detect(dir string) bool

	// Install injects the bridge script into .git/hooks/ or runs engine-specific setup
	Install(ctx context.Context, dir string) error

	// Run executes the specific hook (e.g., "shellcheck") or stage (e.g., "pre-commit")
	// args contains any trailing arguments passed by Git
	Run(ctx context.Context, hookName string, stage string, args []string) error

	// Name returns the identifier of the engine
	Name() string
}
