// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"context"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/repository"
)

type mockIndexRepo struct{}

func (m *mockIndexRepo) Upsert(ctx context.Context, entry *repository.IndexEntry) error {
	return nil
}

func (m *mockIndexRepo) FindByTool(ctx context.Context, toolName string) (*repository.IndexEntry, error) {
	if toolName == "known" {
		return &repository.IndexEntry{Tool: "known"}, nil
	}
	return nil, context.DeadlineExceeded
}

func (m *mockIndexRepo) List(ctx context.Context) ([]*repository.IndexEntry, error) {
	return nil, nil
}

func (m *mockIndexRepo) Search(ctx context.Context, query string) ([]*repository.IndexEntry, error) {
	return nil, nil
}

func (m *mockIndexRepo) Delete(ctx context.Context, tool string) error {
	return nil
}

func TestConfigValidator_Validate(t *testing.T) {
	validator := NewConfigValidator(&mockIndexRepo{}, []string{"native", "asdf"})

	// 1. Nil config
	res := validator.Validate(context.Background(), nil)
	if res.Valid {
		t.Error("expected invalid for nil config")
	}
	if !res.HasErrors() {
		t.Error("expected errors for nil config")
	}
	if summary := res.Summary(); summary != "Found 1 error(s), 0 warning(s)." {
		t.Errorf("expected 1 error summary, got %q", summary)
	}

	// 2. Empty valid config
	cfg := &config.Config{}
	res = validator.Validate(context.Background(), cfg)
	if !res.Valid {
		t.Error("expected valid for empty config")
	}
	if res.Summary() != "Configuration is valid." {
		t.Errorf("expected valid summary, got %q", res.Summary())
	}

	// 3. Tool validation
	cfg = &config.Config{
		Tools: map[string]config.ToolConfig{
			"known": {
				Version: "latest",
				Backend: "native",
			},
			"unknown": {
				Version: "1.0.0",
				Backend: "asdf",
			},
			"bad_version": {
				Version: "1.0 ; rm -rf /",
				Backend: "native",
			},
			"no_version": {
				Version: "",
			},
			"bad_backend": {
				Version: "1.0",
				Backend: "nonexistent",
			},
		},
		Settings: config.Settings{
			CacheTTL: -1,
			Jobs:     -1,
		},
	}
	res = validator.Validate(context.Background(), cfg)
	if res.Valid {
		t.Error("expected invalid config")
	}

	// Should have multiple errors and warnings
	errCount, warnCount := 0, 0
	for _, issue := range res.Issues {
		if issue.Severity == SeverityError {
			errCount++
		} else if issue.Severity == SeverityWarning {
			warnCount++
		}
	}
	// Errors: bad_version(1), no_version(1), bad_backend(1), cache_ttl(1), jobs(1) -> 5
	if warnCount != 4 {
		t.Errorf("expected 4 warnings, got %d", warnCount)
	}
	if errCount != 5 {
		t.Errorf("expected 5 errors, got %d", errCount)
	}
}

func TestValidateVersionSpec(t *testing.T) {
	tests := []struct {
		version string
		wantErr bool
	}{
		{"", true},
		{"latest", false},
		{"LTS", false},
		{"1.0.0", false},
		{">= 1.0.0", false},
		{"1.0.0 ", false},
		{"1.0.0 2.0.0", true},
		{"1.0.0; ls", true},
		{"$VAR", true},
		{"`ls`", true},
		{"1.0.0|grep", true},
		{"1.0.0&echo", true},
	}

	for _, tt := range tests {
		err := validateVersionSpec(tt.version)
		if (err != nil) != tt.wantErr {
			t.Errorf("validateVersionSpec(%q) error = %v, wantErr %v", tt.version, err, tt.wantErr)
		}
	}
}
