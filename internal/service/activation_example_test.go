// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service_test

import (
	"github.com/snowdreamtech/unirtm/internal/provider"

	"context"
	"fmt"

	"github.com/snowdreamtech/unirtm/internal/service"
)

// ExampleActivationManager_GenerateGlobalActivation demonstrates generating
// a global activation script for bash.
func ExampleActivationManager_GenerateGlobalActivation() {
	manager := service.NewActivationManager("/usr/local/unirtm/shims", "/var/lib/unirtm", provider.NewRegistry())
	ctx := context.Background()

	toolVersions := map[string]string{
		"node":   "20.0.0",
		"python": "3.11.0",
		"go":     "1.21.0",
	}

	script, err := manager.GenerateGlobalActivation(ctx, service.ShellBash, toolVersions)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Shell: %s\n", script.Shell)
	fmt.Println("Script contains PATH modification:", containsString(script.Content, "export PATH="))
	fmt.Println("Script contains node version:", containsString(script.Content, "UNIRTM_NODE_VERSION"))
	fmt.Println("Script contains python version:", containsString(script.Content, "UNIRTM_PYTHON_VERSION"))
	fmt.Println("Script contains go version:", containsString(script.Content, "UNIRTM_GO_VERSION"))

	// Output:
	// Shell: bash
	// Script contains PATH modification: true
	// Script contains node version: true
	// Script contains python version: true
	// Script contains go version: true
}

// ExampleActivationManager_GenerateProjectActivation demonstrates generating
// a project-specific activation script.
func ExampleActivationManager_GenerateProjectActivation() {
	manager := service.NewActivationManager("/usr/local/unirtm/shims", "/var/lib/unirtm", provider.NewRegistry())
	ctx := context.Background()

	toolVersions := map[string]string{
		"node": "18.0.0",
	}

	envVars := map[string]string{
		"NODE_ENV": "development",
		"DEBUG":    "app:*",
	}

	script, err := manager.GenerateProjectActivation(
		ctx,
		service.ShellBash,
		"/home/user/myproject",
		toolVersions,
		envVars,
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Shell: %s\n", script.Shell)
	fmt.Println("Script contains project directory:", containsString(script.Content, "UNIRTM_PROJECT_DIR"))
	fmt.Println("Script contains NODE_ENV:", containsString(script.Content, "NODE_ENV"))
	fmt.Println("Script contains DEBUG:", containsString(script.Content, "DEBUG"))
	fmt.Println("Scope is project:", containsString(script.Content, "UNIRTM_ACTIVATION_SCOPE=\"project\""))

	// Output:
	// Shell: bash
	// Script contains project directory: true
	// Script contains NODE_ENV: true
	// Script contains DEBUG: true
	// Scope is project: true
}

// ExampleActivationManager_GenerateActivationScript_fish demonstrates generating
// an activation script for fish shell.
func ExampleActivationManager_GenerateActivationScript_fish() {
	manager := service.NewActivationManager("/usr/local/unirtm/shims", "/var/lib/unirtm", provider.NewRegistry())
	ctx := context.Background()

	config := service.ActivationConfig{
		Shell:    service.ShellFish,
		Scope:    service.ScopeGlobal,
		ShimsDir: "/usr/local/unirtm/shims",
		ToolVersions: map[string]string{
			"ruby": "3.2.0",
		},
	}

	script, err := manager.GenerateActivationScript(ctx, config)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Shell: %s\n", script.Shell)
	fmt.Println("Uses fish syntax:", containsString(script.Content, "set -gx"))
	fmt.Println("Contains ruby version:", containsString(script.Content, "UNIRTM_RUBY_VERSION"))

	// Output:
	// Shell: fish
	// Uses fish syntax: true
	// Contains ruby version: true
}

// ExampleActivationManager_GenerateActivationScript_powershell demonstrates
// generating an activation script for PowerShell.
func ExampleActivationManager_GenerateActivationScript_powershell() {
	manager := service.NewActivationManager("/usr/local/unirtm/shims", "/var/lib/unirtm", provider.NewRegistry())
	ctx := context.Background()

	config := service.ActivationConfig{
		Shell:    service.ShellPowerShell,
		Scope:    service.ScopeGlobal,
		ShimsDir: "/usr/local/unirtm/shims",
		ToolVersions: map[string]string{
			"dotnet": "8.0.0",
		},
		EnvVars: map[string]string{
			"DOTNET_CLI_TELEMETRY_OPTOUT": "1",
		},
	}

	script, err := manager.GenerateActivationScript(ctx, config)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Shell: %s\n", script.Shell)
	fmt.Println("Uses PowerShell syntax:", containsString(script.Content, "$env:"))
	fmt.Println("Contains dotnet version:", containsString(script.Content, "UNIRTM_DOTNET_VERSION"))
	fmt.Println("Contains telemetry opt-out:", containsString(script.Content, "DOTNET_CLI_TELEMETRY_OPTOUT"))

	// Output:
	// Shell: powershell
	// Uses PowerShell syntax: true
	// Contains dotnet version: true
	// Contains telemetry opt-out: true
}

// ExampleActivationManager_GenerateActivationScript_multipleTools demonstrates
// generating an activation script with multiple tools.
func ExampleActivationManager_GenerateActivationScript_multipleTools() {
	manager := service.NewActivationManager("/usr/local/unirtm/shims", "/var/lib/unirtm", provider.NewRegistry())
	ctx := context.Background()

	config := service.ActivationConfig{
		Shell:    service.ShellBash,
		Scope:    service.ScopeGlobal,
		ShimsDir: "/usr/local/unirtm/shims",
		ToolVersions: map[string]string{
			"node":   "20.0.0",
			"python": "3.11.0",
			"go":     "1.21.0",
			"ruby":   "3.2.0",
		},
	}

	script, err := manager.GenerateActivationScript(ctx, config)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Shell: %s\n", script.Shell)
	fmt.Printf("Tool count: %d\n", len(config.ToolVersions))
	fmt.Println("Contains all tools:",
		containsString(script.Content, "UNIRTM_NODE_VERSION") &&
			containsString(script.Content, "UNIRTM_PYTHON_VERSION") &&
			containsString(script.Content, "UNIRTM_GO_VERSION") &&
			containsString(script.Content, "UNIRTM_RUBY_VERSION"))

	// Output:
	// Shell: bash
	// Tool count: 4
	// Contains all tools: true
}

// ExampleDetectShell demonstrates detecting the current shell.
func ExampleDetectShell() {
	shell, err := service.DetectShell()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Detected shell type: %s\n", shell)
	fmt.Println("Shell is supported:",
		shell == service.ShellBash ||
			shell == service.ShellZsh ||
			shell == service.ShellFish ||
			shell == service.ShellPowerShell)

	// Output will vary by platform, but should always be supported
	// Example output on Linux:
	// Detected shell type: bash
	// Shell is supported: true
}

// Helper function for examples
func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || len(s) >= len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
				findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
