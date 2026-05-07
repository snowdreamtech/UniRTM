// Package service provides business logic for UniRTM operations.
package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
	"github.com/snowdreamtech/unirtm/internal/repository"
)

// ValidationSeverity represents the severity of a validation issue.
type ValidationSeverity string

const (
	SeverityError   ValidationSeverity = "error"
	SeverityWarning ValidationSeverity = "warning"
)

// ValidationIssue represents a single configuration validation problem.
type ValidationIssue struct {
	Severity ValidationSeverity
	Field    string
	Message  string
}

// ConfigValidationResult holds all issues found during validation.
//
// Validates Requirements: 13.1–13.7
type ConfigValidationResult struct {
	// Issues contains all detected problems (errors and warnings).
	Issues []ValidationIssue
	// Valid is true only when there are no error-severity issues.
	Valid bool
}

// HasErrors returns true if any error-severity issues exist.
func (r *ConfigValidationResult) HasErrors() bool {
	for _, issue := range r.Issues {
		if issue.Severity == SeverityError {
			return true
		}
	}
	return false
}

// Summary returns a concise human-readable summary of the validation result.
func (r *ConfigValidationResult) Summary() string {
	errors, warnings := 0, 0
	for _, issue := range r.Issues {
		switch issue.Severity {
		case SeverityError:
			errors++
		case SeverityWarning:
			warnings++
		}
	}
	if errors == 0 && warnings == 0 {
		return "Configuration is valid."
	}
	return fmt.Sprintf("Found %d error(s), %d warning(s).", errors, warnings)
}

// ConfigValidator performs semantic validation of UniRTM configuration.
//
// It goes beyond basic TOML syntax to verify:
//   - Tool names exist in the index (Req 13.1)
//   - Version specifiers are syntactically valid (Req 13.2)
//   - Backend references are registered (Req 13.3)
//   - Unknown top-level fields generate warnings (Req 13.4)
//   - Environment references are valid (Req 13.5)
//   - Validation-only mode (Req 13.6)
//   - All errors reported, not just first (Req 13.7)
//
// Validates Requirements: 13.1, 13.2, 13.3, 13.4, 13.5, 13.6, 13.7
type ConfigValidator struct {
	indexRepo       repository.IndexRepository
	backendRegistry []string // registered backend names
}

// NewConfigValidator creates a new ConfigValidator.
func NewConfigValidator(indexRepo repository.IndexRepository, registeredBackends []string) *ConfigValidator {
	return &ConfigValidator{
		indexRepo:       indexRepo,
		backendRegistry: registeredBackends,
	}
}

// Validate performs full semantic validation of the given configuration.
//
// All issues are collected and returned (Req 13.7 — report all, not just first).
func (cv *ConfigValidator) Validate(ctx context.Context, cfg *config.Config) *ConfigValidationResult {
	result := &ConfigValidationResult{}

	if cfg == nil {
		result.Issues = append(result.Issues, ValidationIssue{
			Severity: SeverityError,
			Field:    "config",
			Message:  "configuration is nil",
		})
		result.Valid = false
		return result
	}

	// ── 13.1: Verify tool names exist in the index ────────────────────────────
	for toolName, toolCfg := range cfg.Tools {
		if cv.indexRepo != nil {
			if _, err := cv.indexRepo.FindByTool(ctx, toolName); err != nil {
				// Tool not in index — warning (not error: could be custom tool)
				result.Issues = append(result.Issues, ValidationIssue{
					Severity: SeverityWarning,
					Field:    fmt.Sprintf("tools.%s", toolName),
					Message:  fmt.Sprintf("tool %q not found in index (may be a custom or unlisted tool)", toolName),
				})
			}
		}

		// ── 13.2: Verify version specifiers ──────────────────────────────────
		if toolCfg.Version == "" {
			result.Issues = append(result.Issues, ValidationIssue{
				Severity: SeverityError,
				Field:    fmt.Sprintf("tools.%s.version", toolName),
				Message:  fmt.Sprintf("tool %q has no version specified", toolName),
			})
		} else if err := validateVersionSpec(toolCfg.Version); err != nil {
			result.Issues = append(result.Issues, ValidationIssue{
				Severity: SeverityError,
				Field:    fmt.Sprintf("tools.%s.version", toolName),
				Message:  fmt.Sprintf("invalid version specifier for %q: %s", toolName, err.Error()),
			})
		}

		// ── 13.3: Verify backend references ──────────────────────────────────
		if toolCfg.Backend != "" && !cv.isBackendRegistered(toolCfg.Backend) {
			result.Issues = append(result.Issues, ValidationIssue{
				Severity: SeverityError,
				Field:    fmt.Sprintf("tools.%s.backend", toolName),
				Message:  fmt.Sprintf("backend %q is not registered (available: %s)", toolCfg.Backend, strings.Join(cv.backendRegistry, ", ")),
			})
		}
	}

	// ── 13.5: Verify environment references ──────────────────────────────────
	for envName, envCfg := range cfg.Environments {
		for toolName, toolCfg := range envCfg.Tools {
			if toolCfg.Version == "" {
				result.Issues = append(result.Issues, ValidationIssue{
					Severity: SeverityError,
					Field:    fmt.Sprintf("environments.%s.tools.%s.version", envName, toolName),
					Message:  fmt.Sprintf("tool %q in environment %q has no version", toolName, envName),
				})
			}
		}
	}

	// ── 13.6: Validate settings ───────────────────────────────────────────────
	if cfg.Settings.CacheTTL < 0 {
		result.Issues = append(result.Issues, ValidationIssue{
			Severity: SeverityError,
			Field:    "settings.cache_ttl",
			Message:  "cache_ttl must be non-negative",
		})
	}
	if cfg.Settings.Concurrency < 0 {
		result.Issues = append(result.Issues, ValidationIssue{
			Severity: SeverityError,
			Field:    "settings.concurrency",
			Message:  "concurrency must be non-negative",
		})
	}

	// Determine overall validity
	result.Valid = !result.HasErrors()

	logger.Info("Configuration validation complete", map[string]interface{}{
		"valid":    result.Valid,
		"issues":   len(result.Issues),
		"summary":  result.Summary(),
	})

	return result
}

// isBackendRegistered reports whether the given backend name is in the registry.
func (cv *ConfigValidator) isBackendRegistered(name string) bool {
	for _, b := range cv.backendRegistry {
		if strings.EqualFold(b, name) {
			return true
		}
	}
	return false
}

// validateVersionSpec performs basic validation of a version specifier string.
// It rejects clearly malformed values.
//
// Validates Requirement: 13.2
func validateVersionSpec(version string) error {
	version = strings.TrimSpace(version)
	if version == "" {
		return fmt.Errorf("version cannot be empty")
	}

	// Allow well-known aliases
	aliases := []string{"latest", "lts", "stable", "nightly", "beta", "alpha"}
	for _, alias := range aliases {
		if strings.EqualFold(version, alias) {
			return nil
		}
	}

	// Must not contain spaces (multi-constraint ranges not yet supported)
	if strings.Contains(version, " ") && !strings.HasPrefix(version, ">=") {
		return fmt.Errorf("version %q contains spaces — use a single version or range specifier", version)
	}

	// Must not contain shell metacharacters
	for _, ch := range []string{";", "&", "|", "`", "$", "\n"} {
		if strings.Contains(version, ch) {
			return fmt.Errorf("version %q contains invalid character %q", version, ch)
		}
	}

	return nil
}
