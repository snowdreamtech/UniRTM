// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package service provides business logic for UniRTM operations.
package service

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
)

// MigrationSource represents the source format of a migration.
type MigrationSource string

const (
	// MigrationSourceMiseToml represents a .mise.toml file.
	MigrationSourceMiseToml MigrationSource = "mise.toml"
	// MigrationSourceToolVersions represents an asdf/mise .tool-versions file.
	MigrationSourceToolVersions MigrationSource = ".tool-versions"
)

// MigrationTool represents a single tool discovered during migration.
type MigrationTool struct {
	Name    string
	Version string
	Source  MigrationSource
}

// MigrationReport summarizes the result of a migration operation.
type MigrationReport struct {
	Source            string
	Tools             []MigrationTool
	OutputFile        string
	UnsupportedFields []string
	Errors            []string
	DryRun            bool
	GeneratedAt       time.Time
}

// MigrationManager converts mise/asdf configuration to UniRTM format.
//
// Validates Requirements: 21.1, 21.2, 21.3, 21.4, 21.5, 21.6, 21.7
type MigrationManager struct{}

// NewMigrationManager creates a new MigrationManager.
func NewMigrationManager() *MigrationManager {
	return &MigrationManager{}
}

// MigrateFile converts a single mise/asdf configuration file to UniRTM format.
//
// It auto-detects the source format from the filename.
//
// Validates Requirements: 21.1, 21.2, 21.3, 21.6
func (mm *MigrationManager) MigrateFile(ctx context.Context, sourcePath string, outputPath string, dryRun bool) (*MigrationReport, error) {
	report := &MigrationReport{
		Source:      sourcePath,
		OutputFile:  outputPath,
		DryRun:      dryRun,
		GeneratedAt: time.Now(),
	}

	// Detect source format
	base := filepath.Base(sourcePath)
	var tools []MigrationTool
	var unsupported []string
	var err error

	switch {
	case base == ".tool-versions" || strings.HasSuffix(base, ".tool-versions"):
		tools, err = mm.parseToolVersions(sourcePath)
		report.Source = string(MigrationSourceToolVersions)
	default: // mise.toml or .mise.toml
		tools, unsupported, err = mm.parseMiseToml(sourcePath)
		report.Source = string(MigrationSourceMiseToml)
		report.UnsupportedFields = unsupported
	}

	if err != nil {
		report.Errors = append(report.Errors, err.Error())
		return report, fmt.Errorf("parse source file: %w", err)
	}

	report.Tools = tools

	if len(tools) == 0 {
		report.Errors = append(report.Errors, "no tools found in source file")
		return report, nil
	}

	logger.Info("Parsed migration source", map[string]interface{}{
		"source":     sourcePath,
		"tool_count": len(tools),
	})

	// Generate UniRTM TOML output
	content := mm.generateUnirtmToml(tools)

	if dryRun {
		logger.Info("[dry-run] Would write UniRTM config", map[string]interface{}{
			"output":  outputPath,
			"content": content,
		})
		return report, nil
	}

	// Write output file
	if outputPath == "" {
		outputPath = "unirtm.toml"
		report.OutputFile = outputPath
	}

	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		report.Errors = append(report.Errors, fmt.Sprintf("write output: %s", err.Error()))
		return report, fmt.Errorf("write output file: %w", err)
	}

	logger.Info("Migration complete", map[string]interface{}{
		"output":     outputPath,
		"tool_count": len(tools),
	})

	return report, nil
}

// MigrateDirectory scans a directory for mise/asdf config files and migrates them.
//
// Validates Requirement: 21.4
func (mm *MigrationManager) MigrateDirectory(ctx context.Context, dir string, dryRun bool) ([]*MigrationReport, error) {
	candidates := []string{
		filepath.Join(dir, ".mise.toml"),
		filepath.Join(dir, "mise.toml"),
		filepath.Join(dir, ".tool-versions"),
	}

	var reports []*MigrationReport
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			continue
		}
		outputPath := filepath.Join(dir, "unirtm.toml")
		report, err := mm.MigrateFile(ctx, candidate, outputPath, dryRun)
		if err != nil {
			report.Errors = append(report.Errors, err.Error())
		}
		reports = append(reports, report)
		break // Stop after first found file
	}

	if len(reports) == 0 {
		return nil, fmt.Errorf("no mise or .tool-versions files found in %s", dir)
	}

	return reports, nil
}

// parseToolVersions parses an asdf/mise .tool-versions file.
//
// Format: one tool per line — "<name> <version>"
// Validates Requirement: 21.2
func (mm *MigrationManager) parseToolVersions(path string) ([]MigrationTool, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open .tool-versions: %w", err)
	}
	defer f.Close()

	var tools []MigrationTool
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		tools = append(tools, MigrationTool{
			Name:    parts[0],
			Version: parts[1],
			Source:  MigrationSourceToolVersions,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan .tool-versions: %w", err)
	}
	return tools, nil
}

// parseMiseToml parses a .mise.toml file and extracts tool versions.
//
// This is a minimal TOML parser focused on the [tools] section.
// For full TOML parsing, a proper TOML library would be used.
// Validates Requirement: 21.1
func (mm *MigrationManager) parseMiseToml(path string) ([]MigrationTool, []string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("open mise.toml: %w", err)
	}
	defer f.Close()

	var tools []MigrationTool
	var unsupported []string
	inToolsSection := false

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Section header
		if strings.HasPrefix(line, "[") {
			inToolsSection = line == "[tools]"
			// Track unsupported sections
			if !inToolsSection && line != "[settings]" && !strings.HasPrefix(line, "[tools.") {
				sectionName := strings.Trim(line, "[]")
				if sectionName != "env" && sectionName != "tasks" {
					unsupported = append(unsupported, fmt.Sprintf("section %q may require manual migration", sectionName))
				}
			}
			continue
		}

		if !inToolsSection {
			continue
		}

		// Parse key = "version" or key = ["version", ...]
		eqIdx := strings.Index(line, "=")
		if eqIdx < 0 {
			continue
		}

		toolName := strings.TrimSpace(line[:eqIdx])
		rawVersion := strings.TrimSpace(line[eqIdx+1:])

		// Handle array syntax: ["1.0.0", "2.0.0"] — take first
		if strings.HasPrefix(rawVersion, "[") {
			rawVersion = strings.Trim(rawVersion, "[]")
			if idx := strings.Index(rawVersion, ","); idx >= 0 {
				rawVersion = rawVersion[:idx]
			}
		}

		// Strip surrounding quotes
		version := strings.Trim(rawVersion, `"'`)

		if toolName != "" && version != "" {
			tools = append(tools, MigrationTool{
				Name:    toolName,
				Version: version,
				Source:  MigrationSourceMiseToml,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, unsupported, fmt.Errorf("scan mise.toml: %w", err)
	}

	return tools, unsupported, nil
}

// generateUnirtmToml generates a UniRTM TOML configuration from a list of tools.
//
// Validates Requirements: 21.3, 21.5
func (mm *MigrationManager) generateUnirtmToml(tools []MigrationTool) string {
	var sb strings.Builder

	sb.WriteString("# UniRTM configuration\n")
	sb.WriteString("# Generated by 'unirtm migrate' — review before use\n\n")

	sb.WriteString("[tools]\n")
	for _, t := range tools {
		sb.WriteString(fmt.Sprintf("  [tools.%s]\n", t.Name))
		sb.WriteString(fmt.Sprintf("    version = \"%s\"\n", t.Version))
	}

	return sb.String()
}

// FormatReport returns a human-readable migration report string.
//
// Validates Requirement: 21.7 (generate migration report)
func (mm *MigrationManager) FormatReport(report *MigrationReport) string {
	var sb strings.Builder

	sb.WriteString("Migration Report\n")
	sb.WriteString(strings.Repeat("─", 40) + "\n")
	sb.WriteString(fmt.Sprintf("Source:  %s\n", report.Source))
	sb.WriteString(fmt.Sprintf("Output:  %s\n", report.OutputFile))
	if report.DryRun {
		sb.WriteString("Mode:    dry-run (no files written)\n")
	}
	sb.WriteString(fmt.Sprintf("Tools:   %d\n\n", len(report.Tools)))

	if len(report.Tools) > 0 {
		sb.WriteString("Migrated tools:\n")
		for _, t := range report.Tools {
			sb.WriteString(fmt.Sprintf("  • %s@%s\n", t.Name, t.Version))
		}
		sb.WriteString("\n")
	}

	if len(report.UnsupportedFields) > 0 {
		sb.WriteString("Warnings (manual review needed):\n")
		for _, u := range report.UnsupportedFields {
			sb.WriteString(fmt.Sprintf("  ⚠  %s\n", u))
		}
		sb.WriteString("\n")
	}

	if len(report.Errors) > 0 {
		sb.WriteString("Errors:\n")
		for _, e := range report.Errors {
			sb.WriteString(fmt.Sprintf("  ✗  %s\n", e))
		}
	}

	return sb.String()
}
