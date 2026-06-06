package hook

import (
	"testing"
)

func TestValidateHookName_ValidNames(t *testing.T) {
	valid := []string{
		"pre-commit",
		"commit-msg",
		"prepare-commit-msg",
		"post-commit",
		"pre-push",
		"pre-rebase",
		"post-checkout",
		"post-merge",
		"post-rewrite",
		"pre-receive",
		"update",
		"post-receive",
		"post-update",
	}
	for _, name := range valid {
		t.Run(name, func(t *testing.T) {
			if err := ValidateHookName(name); err != nil {
				t.Errorf("ValidateHookName(%q) returned error for valid name: %v", name, err)
			}
		})
	}
}

func TestValidateHookName_InvalidNames(t *testing.T) {
	invalid := []string{
		"../etc/passwd",
		"../../evil",
		"",
		"notahook",
		"pre_commit",      // underscore instead of dash
		"PRE-COMMIT",      // uppercase
		"pre-commit\x00",  // null byte
		"pre-commit; rm -rf /", // shell injection attempt
		"/etc/passwd",
		".",
		"..",
	}
	for _, name := range invalid {
		t.Run(name, func(t *testing.T) {
			if err := ValidateHookName(name); err == nil {
				t.Errorf("ValidateHookName(%q) should have returned error for invalid name", name)
			}
		})
	}
}
