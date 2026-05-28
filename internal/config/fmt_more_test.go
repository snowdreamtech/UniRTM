// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"os"
	"testing"
)

func TestFormatFile_Errors(t *testing.T) {
	_, err := FormatFile("nonexistent.toml", false)
	if err == nil {
		t.Errorf("expected error formatting nonexistent file")
	}

	f, _ := os.CreateTemp("", "bad-*.toml")
	f.WriteString(`[invalid toml`)
	f.Close()
	defer os.Remove(f.Name())

	_, err = FormatFile(f.Name(), false)
	if err != nil {
		t.Errorf("expected format toml to not fail")
	}
}

func TestGetSectionOrder(t *testing.T) {
	order := getSectionOrder("tools")
	if order != 8 {
		t.Errorf("expected tools to be 8")
	}
	order = getSectionOrder("aliases")
	if order != 9 {
		t.Errorf("expected aliases to be 9")
	}
	order = getSectionOrder("env")
	if order != 4 {
		t.Errorf("expected env to be 4")
	}
	order = getSectionOrder("tasks")
	if order != 10 {
		t.Errorf("expected tasks to be 10")
	}
	order = getSectionOrder("unknown")
	if order != 9 {
		t.Errorf("expected unknown to be 9")
	}
}
