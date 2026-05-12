// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config_test

import (
	"fmt"

	"github.com/snowdreamtech/unirtm/internal/config"
)

// ExampleConfig demonstrates basic usage of the Config structure.
func ExampleConfig() {
	cfg := config.Config{
		Tools: map[string]config.ToolConfig{
			"node": {
				Version: "20.0.0",
				Backend: "github",
			},
			"python": {
				Version: "3.11.0",
			},
		},
		Env: map[string]interface{}{
			"NODE_ENV": "development",
		},
		Settings: config.Settings{
			CacheDir:    "/tmp/unirtm/cache",
			DataDir:     "/tmp/unirtm/data",
			CacheTTL:    3600,
			Concurrency: 4,
		},
		Tasks: map[string]config.Task{
			"build": {
				Description: "Build the project",
				Run:         "go build -o bin/app",
			},
			"test": {
				Description: "Run tests",
				Run:         "go test ./...",
				Depends:     []string{"build"},
			},
		},
	}

	if err := cfg.Validate(); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		return
	}

	fmt.Println("Configuration is valid")
	// Output: Configuration is valid
}

// ExampleToolConfig demonstrates tool configuration with version specification.
func ExampleToolConfig() {
	tool := config.ToolConfig{
		Version:  "1.20.0",
		Backend:  "github",
		Provider: "go",
	}

	if err := tool.Validate(); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		return
	}

	fmt.Println("Tool configuration is valid")
	// Output: Tool configuration is valid
}

// ExampleSettings demonstrates settings configuration.
func ExampleSettings() {
	settings := config.Settings{
		CacheDir:    "/var/cache/unirtm",
		DataDir:     "/var/lib/unirtm",
		CacheTTL:    86400, // 24 hours
		Concurrency: 8,
	}

	if err := settings.Validate(); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		return
	}

	fmt.Println("Settings are valid")
	// Output: Settings are valid
}

// ExampleTask demonstrates task configuration with dependencies.
func ExampleTask() {
	task := config.Task{
		Description: "Deploy to production",
		Run:         "./deploy.sh production",
		Env: map[string]interface{}{
			"ENVIRONMENT": "production",
		},
		Depends: []string{"test", "build"},
	}

	if err := task.Validate(); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		return
	}

	fmt.Println("Task configuration is valid")
	// Output: Task configuration is valid
}

// ExampleConfig_Validate_error demonstrates validation error handling.
func ExampleConfig_Validate_error() {
	cfg := config.Config{
		Tools: map[string]config.ToolConfig{
			"node": {
				Version: "", // Invalid: empty version
			},
		},
		Settings: config.Settings{
			CacheTTL: -100, // Invalid: negative value
		},
	}

	if err := cfg.Validate(); err != nil {
		fmt.Println("Validation failed as expected")
	}
	// Output: Validation failed as expected
}
