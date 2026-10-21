// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard

import (
	"context"
)

// Category represents one of the 6 core functional pillars of Quality Guard
type Category string

const (
	CategoryHygiene  Category = "hygiene"
	CategoryFormat   Category = "format"
	CategoryLint     Category = "lint"
	CategorySecret   Category = "secret"
	CategorySecurity Category = "security"
	CategoryPolicy   Category = "policy"
)

// Phase represents the 3-phase execution order
type Phase string

const (
	PhaseFixer     Phase = "fixer"     // Phase 1: In-place text and AST auto-fixers
	PhaseFormatter Phase = "formatter" // Phase 2: In-place visual styling & layout
	PhaseLinter    Phase = "linter"    // Phase 3: Strict read-only verification
)

// Severity indicates the seriousness of a detected issue
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// Issue describes a detected violation or fix performed on a file
type Issue struct {
	FilePath string   `json:"file_path"`
	Line     int      `json:"line"`
	Column   int      `json:"column"`
	RuleID   string   `json:"rule_id"`
	Category Category `json:"category"`
	Phase    Phase    `json:"phase"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
	Fixable  bool     `json:"fixable"`
	Fixed    bool     `json:"fixed"`
}

// ExecuteOptions specifies runtime configuration for rule execution
type ExecuteOptions struct {
	WorkspaceDir   string
	Mode           string // "staged", "changed", "all"
	Fix            bool   // If true, rules that support fixing will mutate files in-place
	DryRun         bool   // If true, report what would be fixed without modifying disk
	MaxFileSize    int64
	OnRuleComplete func(report RuleReport)
}

// Rule defines the execution contract for any builtin or custom guard rule
type Rule interface {
	ID() string
	Category() Category
	Phase() Phase
	DefaultExtensions() []string
	Execute(ctx context.Context, files []string, opts ExecuteOptions) ([]Issue, error)
}

// RuleConfig provides user configuration for a specific rule from .unirtm.toml
type RuleConfig struct {
	Enabled      *bool    `toml:"enabled" json:"enabled"`
	Extensions   []string `toml:"extensions" json:"extensions"`
	IncludeDirs  []string `toml:"include_dirs" json:"include_dirs"`
	ExcludeDirs  []string `toml:"exclude_dirs" json:"exclude_dirs"`
	ExcludeFiles []string `toml:"exclude_files" json:"exclude_files"`
	MaxFileSize  int64    `toml:"max_file_size" json:"max_file_size"`
}

// ProfileConfig represents execution presets (e.g. pre-commit, ci)
type ProfileConfig struct {
	Mode      string   `toml:"mode" json:"mode"` // "staged", "changed", "all"
	Stages    []string `toml:"stages" json:"stages"`
	Offline   bool     `toml:"offline" json:"offline"`
	Timeout   string   `toml:"timeout" json:"timeout"`
	SkipTools []string `toml:"skip_tools" json:"skip_tools"`
	Tools     []string `toml:"tools" json:"tools"`
}

// SlotConfig allows user customization or overriding of tool slots from .unirtm.toml
type SlotConfig struct {
	ID         string   `toml:"id" json:"id"`
	Tool       string   `toml:"tool" json:"tool"`
	Command    string   `toml:"command" json:"command"`
	FixCommand string   `toml:"fix_command" json:"fix_command"`
	Extensions []string `toml:"extensions" json:"extensions"`
	Category   Category `toml:"category" json:"category"`
	Phase      Phase    `toml:"phase" json:"phase"`
	Enabled    *bool    `toml:"enabled" json:"enabled"`
}

// GuardConfig represents the top-level [guard] section in .unirtm.toml
type GuardConfig struct {
	Enabled      bool                     `toml:"enabled" json:"enabled"`
	Mode         string                   `toml:"mode" json:"mode"` // "changed" (default), "staged", "all"
	AutoDetect   bool                     `toml:"auto_detect" json:"auto_detect"`
	IncludeDirs  []string                 `toml:"include_dirs" json:"include_dirs"`
	ExcludeDirs  []string                 `toml:"exclude_dirs" json:"exclude_dirs"`
	ExcludeFiles []string                 `toml:"exclude_files" json:"exclude_files"`
	Slots        map[string]SlotConfig    `toml:"slots" json:"slots"`
	Rules        map[string]RuleConfig    `toml:"rules" json:"rules"`
	Profiles     map[string]ProfileConfig `toml:"profiles" json:"profiles"`
}
