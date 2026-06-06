package hook

import "fmt"

// validHookNames is the exhaustive set of Git-defined hook names.
// Only these names are permitted to prevent path traversal attacks.
var validHookNames = map[string]struct{}{
	// Commit-workflow hooks
	"pre-commit":        {},
	"prepare-commit-msg": {},
	"commit-msg":        {},
	"post-commit":       {},
	// Email-workflow hooks
	"applypatch-msg":    {},
	"pre-applypatch":    {},
	"post-applypatch":   {},
	// Other client hooks
	"pre-rebase":        {},
	"pre-push":          {},
	"post-checkout":     {},
	"post-merge":        {},
	"post-rewrite":      {},
	"pre-auto-gc":       {},
	"fsmonitor-watchman": {},
	// Server-side hooks
	"pre-receive":       {},
	"update":            {},
	"post-receive":      {},
	"post-update":       {},
	"push-to-checkout":  {},
	"pre-receive-hook":  {},
}

// ValidateHookName returns an error if hookName is not a known Git hook.
// This prevents path traversal attacks when hookName is used to construct file paths.
func ValidateHookName(hookName string) error {
	if _, ok := validHookNames[hookName]; !ok {
		return fmt.Errorf("invalid hook name %q: must be a valid Git hook name (e.g. pre-commit, commit-msg)", hookName)
	}
	return nil
}
