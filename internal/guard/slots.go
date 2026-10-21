// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard

import (
	"os"
	"path/filepath"
)

// ToolSlot defines a standard ecosystem slot for formatters, linters, or checkers
type ToolSlot struct {
	ID                 string   `json:"id"`
	Description        string   `json:"description"`
	Category           Category `json:"category"`
	Phase              Phase    `json:"phase"`
	DefaultToolName    string   `json:"default_tool_name"`
	DefaultCommand     string   `json:"default_command"`
	FixCommand         string   `json:"fix_command"`
	DefaultExtensions  []string `json:"default_extensions"`
	StandardConfigFiles []string `json:"standard_config_files"`
}

// ToolDefinition represents a bound or custom tool implementation
type ToolDefinition struct {
	Name        string   `toml:"name" json:"name"`
	Slot        string   `toml:"slot" json:"slot"`
	Category    Category `toml:"category" json:"category"`
	Phase       Phase    `toml:"phase" json:"phase"`
	Extensions  []string `toml:"extensions" json:"extensions"`
	Command     string   `toml:"command" json:"command"`
	FixCommand  string   `toml:"fix_command" json:"fix_command"`
	ConfigPath  string   `toml:"config" json:"config"`
	Enabled     bool     `toml:"enabled" json:"enabled"`
	Stage       string   `toml:"stage" json:"stage"` // "pre-commit", "ci", "all"
}

// DefaultToolSlots provides the standard slot catalog for UniRTM
var DefaultToolSlots = []ToolSlot{
	{
		ID:                 "go.formatter",
		Description:        "Go canonical code formatter",
		Category:           CategoryFormat,
		Phase:              PhaseFormatter,
		DefaultToolName:    "gofmt",
		DefaultCommand:     "gofmt -l {{files}}",
		FixCommand:         "gofmt -w {{files}}",
		DefaultExtensions:  []string{".go"},
	},
	{
		ID:                 "go.imports",
		Description:        "Go imports manager and organizer",
		Category:           CategoryHygiene,
		Phase:              PhaseFixer,
		DefaultToolName:    "goimports",
		DefaultCommand:     "goimports -l {{files}}",
		FixCommand:         "goimports -w {{files}}",
		DefaultExtensions:  []string{".go"},
	},
	{
		ID:                 "go.linter",
		Description:        "Go static analysis & quality linter",
		Category:           CategoryLint,
		Phase:              PhaseLinter,
		DefaultToolName:    "golangci-lint",
		DefaultCommand:     "golangci-lint run",
		FixCommand:         "golangci-lint run --fix",
		DefaultExtensions:  []string{".go"},
		StandardConfigFiles: []string{".golangci.yml", ".golangci.yaml", ".golangci.toml"},
	},
	{
		ID:                 "web.formatter",
		Description:        "Web multi-language formatter (Prettier / Biome)",
		Category:           CategoryFormat,
		Phase:              PhaseFormatter,
		DefaultToolName:    "prettier",
		DefaultCommand:     "prettier --check {{files}}",
		FixCommand:         "prettier --write {{files}}",
		DefaultExtensions:  []string{".js", ".jsx", ".ts", ".tsx", ".vue", ".json", ".css", ".md"},
		StandardConfigFiles: []string{".prettierrc", ".prettierrc.json", ".prettierrc.yaml", "prettier.config.js"},
	},
	{
		ID:                 "web.linter",
		Description:        "JavaScript / TypeScript / Vue linter (ESLint / Oxlint / Biome)",
		Category:           CategoryLint,
		Phase:              PhaseLinter,
		DefaultToolName:    "eslint",
		DefaultCommand:     "eslint {{files}}",
		FixCommand:         "eslint --fix {{files}}",
		DefaultExtensions:  []string{".js", ".jsx", ".ts", ".tsx", ".vue"},
		StandardConfigFiles: []string{"eslint.config.js", "eslint.config.mjs", ".eslintrc.json", ".eslintrc.js", ".eslintrc.yml"},
	},
	{
		ID:                 "python.suite",
		Description:        "Python fast formatter and linter (Ruff / Black / Flake8)",
		Category:           CategoryLint,
		Phase:              PhaseLinter,
		DefaultToolName:    "ruff",
		DefaultCommand:     "ruff check {{files}}",
		FixCommand:         "ruff check --fix {{files}}",
		DefaultExtensions:  []string{".py", ".pyi"},
		StandardConfigFiles: []string{"ruff.toml", ".ruff.toml", "pyproject.toml"},
	},
	{
		ID:                 "toml.formatter",
		Description:        "TOML file formatter",
		Category:           CategoryFormat,
		Phase:              PhaseFormatter,
		DefaultToolName:    "taplo",
		DefaultCommand:     "taplo format --check {{files}}",
		FixCommand:         "taplo format {{files}}",
		DefaultExtensions:  []string{".toml"},
		StandardConfigFiles: []string{".taplo.toml", "taplo.toml"},
	},
	{
		ID:                 "shell.formatter",
		Description:        "Shell script standard POSIX formatter",
		Category:           CategoryFormat,
		Phase:              PhaseFormatter,
		DefaultToolName:    "shfmt",
		DefaultCommand:     "shfmt -d -s -i 2 {{files}}",
		FixCommand:         "shfmt -w -s -i 2 {{files}}",
		DefaultExtensions:  []string{".sh", ".bash"},
	},
	{
		ID:                 "shell.linter",
		Description:        "Shell script static analyzer",
		Category:           CategoryLint,
		Phase:              PhaseLinter,
		DefaultToolName:    "shellcheck",
		DefaultCommand:     "shellcheck {{files}}",
		DefaultExtensions:  []string{".sh", ".bash"},
		StandardConfigFiles: []string{".shellcheckrc"},
	},
	{
		ID:                 "secret.gitleaks",
		Description:        "Secret and API key leak protector",
		Category:           CategorySecret,
		Phase:              PhaseLinter,
		DefaultToolName:    "gitleaks",
		DefaultCommand:     "gitleaks protect --staged --redact",
		StandardConfigFiles: []string{".gitleaks.toml"},
	},
}

// DetectStandardConfig searches for existing standard config files in workspace
func DetectStandardConfig(workspaceDir string, candidates []string) string {
	for _, c := range candidates {
		full := filepath.Join(workspaceDir, c)
		if fi, err := os.Stat(full); err == nil && !fi.IsDir() {
			return c
		}
	}
	return ""
}

// GetResolvedToolSlots returns the active tool slot catalog after applying user configuration
func GetResolvedToolSlots(cfg GuardConfig) []ToolSlot {
	slotsMap := make(map[string]ToolSlot)
	var slotOrder []string
	for _, slot := range DefaultToolSlots {
		slotsMap[slot.ID] = slot
		slotOrder = append(slotOrder, slot.ID)
	}

	for id, sCfg := range cfg.Slots {
		slotID := id
		if sCfg.ID != "" {
			slotID = sCfg.ID
		}
		existing, found := slotsMap[slotID]
		if !found {
			existing = ToolSlot{
				ID:       slotID,
				Category: sCfg.Category,
				Phase:    sCfg.Phase,
			}
			if existing.Category == "" {
				existing.Category = CategoryLint
			}
			if existing.Phase == "" {
				existing.Phase = PhaseLinter
			}
			slotOrder = append(slotOrder, slotID)
		}
		if sCfg.Tool != "" {
			existing.DefaultToolName = sCfg.Tool
		}
		if sCfg.Command != "" {
			existing.DefaultCommand = sCfg.Command
		}
		if sCfg.FixCommand != "" {
			existing.FixCommand = sCfg.FixCommand
		}
		if len(sCfg.Extensions) > 0 {
			existing.DefaultExtensions = sCfg.Extensions
		}
		if sCfg.Category != "" {
			existing.Category = sCfg.Category
		}
		if sCfg.Phase != "" {
			existing.Phase = sCfg.Phase
		}
		slotsMap[slotID] = existing
	}

	var resolved []ToolSlot
	for _, id := range slotOrder {
		if sCfg, ok := cfg.Slots[id]; ok && sCfg.Enabled != nil && !*sCfg.Enabled {
			continue
		}
		resolved = append(resolved, slotsMap[id])
	}
	return resolved
}

