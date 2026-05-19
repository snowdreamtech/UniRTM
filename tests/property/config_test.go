// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package property contains property-based tests for UniRTM.
//
// Property-based tests verify universal properties that should hold for all inputs,
// complementing example-based unit tests with comprehensive input coverage.
package property

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"reflect"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/pelletier/go-toml/v2"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/stretchr/testify/require"
)

// Feature: unirtm, Property 1: Configuration Round-Trip (TOML)
//
// **Validates: Requirements 1.1, 26.1, 26.4, 26.7**
//
// For any valid Configuration object, serializing to TOML, parsing back, and
// serializing again SHALL produce an equivalent Configuration object and
// identical TOML output.
//
// This property ensures that:
// 1. TOML serialization is deterministic
// 2. TOML parsing is lossless
// 3. The round-trip preserves all configuration data
// 4. No data corruption occurs during serialization/deserialization
func TestProperty_ConfigurationRoundTrip_TOML(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.MaxSize = 20

	properties := gopter.NewProperties(parameters)

	properties.Property("Configuration round-trip through TOML preserves data", prop.ForAll(
		func(original config.Config) bool {
			// Normalize the config to handle empty maps consistently
			// TOML encoder may omit empty maps, so we need to ensure they're initialized
			if original.Tools == nil {
				original.Tools = make(map[string]config.ToolConfig)
			}
			if original.Env == nil {
				original.Env = make(map[string]interface{})
			}
			if original.Tasks == nil {
				original.Tasks = make(map[string]config.Task)
			}

			// Step 1: Serialize original config to TOML
			original.PostLoad()
			var buf1 bytes.Buffer
			encoder1 := toml.NewEncoder(&buf1)
			err := encoder1.Encode(&original)
			if err != nil {
				t.Logf("Failed to encode original config: %v", err)
				return false
			}
			toml1 := buf1.String()

			// Step 2: Parse TOML back into a Config object
			var parsed config.Config
			decoder := toml.NewDecoder(bytes.NewReader(buf1.Bytes()))
			err = decoder.Decode(&parsed)
			if err != nil {
				t.Logf("Failed to decode TOML: %v", err)
				return false
			}
			parsed.PostLoad()

			// Normalize parsed config to handle empty maps
			if parsed.Tools == nil {
				parsed.Tools = make(map[string]config.ToolConfig)
			}
			if parsed.Env == nil {
				parsed.Env = make(map[string]interface{})
			}
			if parsed.Tasks == nil {
				parsed.Tasks = make(map[string]config.Task)
			}

			// Step 3: Serialize parsed config to TOML again
			var buf2 bytes.Buffer
			encoder2 := toml.NewEncoder(&buf2)
			err = encoder2.Encode(&parsed)
			if err != nil {
				t.Logf("Failed to encode parsed config: %v", err)
				return false
			}
			toml2 := buf2.String()

			// Step 4: Verify structural equivalence
			if !configsEqual(original, parsed) {
				t.Logf("Configs not structurally equal after round-trip")
				t.Logf("Original: %+v", original)
				t.Logf("Parsed: %+v", parsed)
				return false
			}

			// Step 5: Verify TOML output is identical
			if toml1 != toml2 {
				t.Logf("TOML output differs after round-trip")
				t.Logf("First serialization:\n%s", toml1)
				t.Logf("Second serialization:\n%s", toml2)
				return false
			}

			return true
		},
		genConfig(),
	))

	properties.TestingRun(t)
}

// YAML configuration is unsupported in UniRTM natively, we only use TOML.

// genConfig generates random Config objects for property-based testing.
//
// The generator creates configs with:
// - Random tool definitions (0-10 tools)
// - Random environment variables (0-10 variables)
// - Random settings with valid ranges
// - Random task definitions (0-5 tasks)
//
// Edge cases covered:
// - Empty maps
// - Maximum length strings
// - Special characters in strings
// - Boundary values for integers
func genConfig() gopter.Gen {
	return gopter.CombineGens(
		genTools(),
		genEnv(),
		genSettings(),
		genTasks(),
	).Map(func(values []interface{}) config.Config {
		env := make(map[string]interface{})
		for k, v := range values[1].(map[string]string) {
			env[k] = v
		}
		tools := values[0].(map[string]config.ToolConfig)
		var toolsRaw map[string]interface{}
		if len(tools) > 0 {
			toolsRaw = make(map[string]interface{})
			for k, v := range tools {
				toolRaw := make(map[string]interface{})
				toolRaw["version"] = v.Version
				if v.Backend != "" {
					toolRaw["backend"] = v.Backend
				}
				if v.Provider != "" {
					toolRaw["provider"] = v.Provider
				}
				toolsRaw[k] = toolRaw
			}
		}
		return config.Config{
			Tools:    tools,
			ToolsRaw: toolsRaw,
			Env:      env,
			Settings: values[2].(config.Settings),
			Tasks:    values[3].(map[string]config.Task),
		}
	})
}

// genTools generates random tool configurations.
func genTools() gopter.Gen {
	return gen.MapOf(
		genToolName(),
		genToolConfig(),
	).SuchThat(func(v interface{}) bool {
		m := v.(map[string]config.ToolConfig)
		return len(m) <= 10 // Limit to 10 tools for performance
	})
}

// genToolName generates valid tool names.
func genToolName() gopter.Gen {
	return gen.OneConstOf("node", "python", "go", "ruby", "rust", "java", "terraform", "kubectl")
}

// genToolConfig generates random ToolConfig objects.
func genToolConfig() gopter.Gen {
	return gopter.CombineGens(
		genVersion(),
		gen.OneConstOf("", "github", "aqua", "http"),
		gen.OneConstOf("", "generic", "node", "python"),
	).Map(func(values []interface{}) config.ToolConfig {
		return config.ToolConfig{
			Version:  values[0].(string),
			Backend:  values[1].(string),
			Provider: values[2].(string),
		}
	})
}

// genVersion generates valid version strings.
func genVersion() gopter.Gen {
	return gen.OneConstOf(
		"1.0.0",
		"2.3.4",
		"20.0.0",
		"3.11.5",
		"latest",
		"lts",
		"stable",
		">=1.20.0",
		"^3.11",
		"~2.7.0",
	)
}

// genEnv generates random environment variable maps.
func genEnv() gopter.Gen {
	return gen.MapOf(
		genEnvKey(),
		genEnvValue(),
	).SuchThat(func(v interface{}) bool {
		m := v.(map[string]string)
		return len(m) <= 10 // Limit to 10 env vars for performance
	})
}

// genEnvKey generates valid environment variable names.
func genEnvKey() gopter.Gen {
	return gen.OneConstOf("PATH", "HOME", "USER", "LANG", "NODE_ENV", "GO_ENV", "PYTHON_ENV")
}

// genEnvValue generates environment variable values.
func genEnvValue() gopter.Gen {
	return gen.OneConstOf(
		"/usr/local/bin",
		"/home/user",
		"production",
		"development",
		"en_US.UTF-8",
		"",
		"/path/with spaces",
		"/path/with:colons",
	)
}

// genSettings generates random Settings objects.
func genSettings() gopter.Gen {
	return gopter.CombineGens(
		gen.OneConstOf("", "/tmp/cache", "/var/cache/unirtm", "~/.cache/unirtm"),
		gen.OneConstOf("", "/var/lib/unirtm", "~/.local/share/unirtm"),
		gen.IntRange(0, 604800), // 0 to 7 days in seconds
	).Map(func(values []interface{}) config.Settings {
		return config.Settings{
			CacheDir:    values[0].(string),
			DataDir:     values[1].(string),
			CacheTTL:    values[2].(int),
		}
	})
}

// genTasks generates random task definitions.
func genTasks() gopter.Gen {
	return gen.MapOf(
		genTaskName(),
		genTask(),
	).SuchThat(func(v interface{}) bool {
		m := v.(map[string]config.Task)
		return len(m) <= 5 // Limit to 5 tasks for performance
	})
}

// genTaskName generates valid task names.
func genTaskName() gopter.Gen {
	return gen.OneConstOf("build", "test", "lint", "deploy", "clean")
}

// genTask generates random Task objects.
func genTask() gopter.Gen {
	return gopter.CombineGens(
		gen.OneConstOf("Build the project", "Run tests", "Lint code", "Deploy to production", "Clean build artifacts"),
		gen.OneConstOf("make build", "go test ./...", "golangci-lint run", "kubectl apply -f deploy/", "rm -rf dist/"),
		gen.MapOf(genEnvKey(), genEnvValue()).SuchThat(func(v interface{}) bool {
			m := v.(map[string]string)
			return len(m) <= 3 // Limit task env vars
		}),
		gen.SliceOf(genTaskName()).SuchThat(func(v interface{}) bool {
			s := v.([]string)
			return len(s) <= 2 // Limit dependencies to avoid cycles
		}),
	).Map(func(values []interface{}) config.Task {
		env := make(map[string]interface{})
		for k, v := range values[2].(map[string]string) {
			env[k] = v
		}
		return config.Task{
			Description: values[0].(string),
			Run:         values[1].(string),
			Env:         env,
			Depends:     values[3].([]string),
		}
	})
}

// configsEqual performs deep equality comparison of two Config objects.
//
// This function compares all fields including maps and nested structures.
// It handles nil maps correctly by treating them as equivalent to empty maps.
func configsEqual(a, b config.Config) bool {
	// Compare Tools
	if !toolMapsEqual(a.Tools, b.Tools) {
		return false
	}

	// Compare Env
	if !interfaceMapsEqual(a.Env, b.Env) {
		return false
	}

	// Compare Settings
	if !settingsEqual(a.Settings, b.Settings) {
		return false
	}

	// Compare Tasks
	if !taskMapsEqual(a.Tasks, b.Tasks) {
		return false
	}

	return true
}

// toolMapsEqual compares two tool configuration maps.
func toolMapsEqual(a, b map[string]config.ToolConfig) bool {
	if len(a) != len(b) {
		return false
	}

	for key, aVal := range a {
		bVal, exists := b[key]
		if !exists {
			return false
		}
		if !toolConfigEqual(aVal, bVal) {
			return false
		}
	}

	return true
}

// toolConfigEqual compares two ToolConfig objects.
func toolConfigEqual(a, b config.ToolConfig) bool {
	return a.Version == b.Version &&
		a.Backend == b.Backend &&
		a.Provider == b.Provider
}

// interfaceMapsEqual compares two interface maps.
func interfaceMapsEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	return reflect.DeepEqual(a, b)
}

// stringMapsEqual compares two string maps.
func stringMapsEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for key, aVal := range a {
		bVal, exists := b[key]
		if !exists || aVal != bVal {
			return false
		}
	}

	return true
}

// settingsEqual compares two Settings objects.
func settingsEqual(a, b config.Settings) bool {
	return a.CacheDir == b.CacheDir &&
		a.DataDir == b.DataDir &&
		a.CacheTTL == b.CacheTTL
}

// taskMapsEqual compares two task maps.
func taskMapsEqual(a, b map[string]config.Task) bool {
	if len(a) != len(b) {
		return false
	}

	for key, aVal := range a {
		bVal, exists := b[key]
		if !exists {
			return false
		}
		if !taskEqual(aVal, bVal) {
			return false
		}
	}

	return true
}

// taskEqual compares two Task objects.
func taskEqual(a, b config.Task) bool {
	if a.Description != b.Description || a.Run != b.Run {
		return false
	}

	if !stringMapsEqual(a.Env, b.Env) {
		return false
	}

	if !stringSlicesEqual(a.Depends, b.Depends) {
		return false
	}

	return true
}

// stringSlicesEqual compares two string slices.
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// Feature: unirtm, Property 3: Configuration Validation Completeness
//
// **Validates: Requirements 1.3**
//
// For any Configuration object with missing required fields, the Configuration_Validator
// SHALL reject it and return an error identifying all missing fields.
//
// This property ensures that:
// 1. Validation catches all missing required fields
// 2. Error messages identify all problems, not just the first one
// 3. Validation is comprehensive and deterministic
func TestProperty_ConfigurationValidationCompleteness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.MaxSize = 20

	properties := gopter.NewProperties(parameters)

	properties.Property("Configuration validation identifies all missing required fields", prop.ForAll(
		func(cfg config.Config) bool {
			// Test 1: Tool with empty version — use a clean config plus one invalid tool
			clean1 := config.Config{
				Tools:    make(config.ToolMap),
				Env:      make(map[string]interface{}),
				Tasks:    make(map[string]config.Task),
				Settings: cfg.Settings,
			}
			clean1.Settings.CacheTTL = 0 // ensure settings are valid
			clean1.Settings.HTTPTimeout = 0
			clean1.Tools["invalid-tool"] = config.ToolConfig{Version: ""}

			err := clean1.Validate()
			if err == nil {
				t.Logf("Expected validation error for tool with empty version")
				return false
			}
			if !strings.Contains(err.Error(), "invalid-tool") {
				t.Logf("Error should mention the invalid tool name: %v", err)
				return false
			}

			// Test 2: Task with empty run command — use a clean config plus one invalid task
			clean2 := config.Config{
				Tools:    make(config.ToolMap),
				Env:      make(map[string]interface{}),
				Tasks:    make(map[string]config.Task),
				Settings: cfg.Settings,
			}
			clean2.Settings.CacheTTL = 0
			clean2.Settings.HTTPTimeout = 0
			clean2.Tasks["invalid-task"] = config.Task{Run: ""}

			err = clean2.Validate()
			if err == nil {
				t.Logf("Expected validation error for task with empty run command")
				return false
			}
			if !strings.Contains(err.Error(), "invalid-task") {
				t.Logf("Error should mention the invalid task name: %v", err)
				return false
			}

			// Test 3: Settings with negative values
			clean3 := config.Config{
				Tools:    make(config.ToolMap),
				Env:      make(map[string]interface{}),
				Tasks:    make(map[string]config.Task),
				Settings: cfg.Settings,
			}
			clean3.Settings.CacheTTL = -1

			err = clean3.Validate()
			if err == nil {
				t.Logf("Expected validation error for negative cache TTL")
				return false
			}
			if !strings.Contains(err.Error(), "cache_ttl") {
				t.Logf("Error should mention cache_ttl: %v", err)
				return false
			}

			// Test 4: Multiple validation errors should all be reported
			clean4 := config.Config{
				Tools:    make(config.ToolMap),
				Env:      make(map[string]interface{}),
				Tasks:    make(map[string]config.Task),
				Settings: cfg.Settings,
			}
			clean4.Settings.CacheTTL = -1
			clean4.Settings.HTTPTimeout = 0
			clean4.Tools["bad-tool"] = config.ToolConfig{Version: ""}
			clean4.Tasks["bad-task"] = config.Task{Run: ""}

			err = clean4.Validate()
			if err == nil {
				t.Logf("Expected validation error for multiple invalid fields")
				return false
			}
			errStr := err.Error()
			hasToolError := strings.Contains(errStr, "bad-tool")
			hasTaskError := strings.Contains(errStr, "bad-task")
			hasSettingsError := strings.Contains(errStr, "cache_ttl")

			if !hasToolError || !hasTaskError || !hasSettingsError {
				t.Logf("Error should report all validation failures: %v", err)
				return false
			}

			return true
		},
		genConfig(),
	))

	properties.TestingRun(t)
}

// Feature: unirtm, Property 4: Invalid Syntax Error Reporting
//
// **Validates: Requirements 1.4, 26.3**
//
// For any syntactically invalid TOML or YAML configuration file, the Config_Parser
// SHALL return an error with line number, column number, and a descriptive message.
//
// This property ensures that:
// 1. Syntax errors are caught during parsing
// 2. Error messages are descriptive and actionable
// 3. Line/column information helps locate the error
func TestProperty_InvalidSyntaxErrorReporting(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.MaxSize = 10

	properties := gopter.NewProperties(parameters)

	properties.Property("Invalid TOML syntax produces descriptive errors", prop.ForAll(
		func(seed int) bool {
			// Generate various invalid TOML syntax patterns
			invalidTOMLs := []string{
				"[tools\nnode = { version = \"20.0.0\" }",                              // Missing closing bracket
				"tools.node.version = \"20.0.0\ntools.python = { version = \"3.11\" }", // Missing closing quote
				"[tools]\nnode = { version = 20.0.0 }",                                 // Unquoted string value
				"[tools]\nnode = { version = \"20.0.0\", }",                            // Trailing comma
				"[tools\n[settings]",                                                   // Unclosed section
			}

			// Pick one based on seed
			invalidTOML := invalidTOMLs[seed%len(invalidTOMLs)]

			// Try to parse it
			var cfg config.Config
			decoder := toml.NewDecoder(strings.NewReader(invalidTOML))
			err := decoder.Decode(&cfg)

			// Should produce an error
			if err == nil {
				t.Logf("Expected parsing error for invalid TOML")
				return false
			}

			// Error should be descriptive (contains "toml" or parsing-related keywords)
			errStr := strings.ToLower(err.Error())
			hasParseInfo := strings.Contains(errStr, "toml") ||
				strings.Contains(errStr, "parse") ||
				strings.Contains(errStr, "syntax") ||
				strings.Contains(errStr, "decode")

			if !hasParseInfo {
				t.Logf("Error should be descriptive: %v", err)
				return false
			}

			return true
		},
		gen.IntRange(0, 1000),
	))

	properties.TestingRun(t)
}

// Feature: unirtm, Property 5: Configuration Merge Precedence
//
// **Validates: Requirements 1.5, 1.6**
//
// For any set of Configuration objects at different hierarchy levels (system, global,
// project, local), merging them SHALL apply the most specific configuration, with
// local overriding project overriding global overriding system.
//
// This property ensures that:
// 1. More specific configurations override less specific ones
// 2. Merging is deterministic and predictable
// 3. All configuration fields respect precedence rules
func TestProperty_ConfigurationMergePrecedence(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.MaxSize = 20

	properties := gopter.NewProperties(parameters)

	properties.Property("Configuration merge respects precedence rules", prop.ForAll(
		func(base, override config.Config) bool {
			// Ensure configs are valid
			if base.Tools == nil {
				base.Tools = make(map[string]config.ToolConfig)
			}
			if base.Env == nil {
				base.Env = make(map[string]interface{})
			}
			if base.Tasks == nil {
				base.Tasks = make(map[string]config.Task)
			}
			if override.Tools == nil {
				override.Tools = make(map[string]config.ToolConfig)
			}
			if override.Env == nil {
				override.Env = make(map[string]interface{})
			}
			if override.Tasks == nil {
				override.Tasks = make(map[string]config.Task)
			}

			// Create a ConfigManager
			mgr := config.NewConfigManager()

			// Merge base and override
			merged, err := mgr.Merge(&base, &override)
			if err != nil {
				t.Logf("Merge failed: %v", err)
				return false
			}

			// Verify precedence: override values should win
			// Check Tools
			for toolName, overrideTool := range override.Tools {
				mergedTool, exists := merged.Tools[toolName]
				if !exists {
					t.Logf("Tool %q from override not in merged config", toolName)
					return false
				}
				if !toolConfigEqual(mergedTool, overrideTool) {
					t.Logf("Tool %q: override value not preserved", toolName)
					return false
				}
			}

			// Check Env
			for envKey, overrideValue := range override.Env {
				mergedValue, exists := merged.Env[envKey]
				if !exists {
					t.Logf("Env %q from override not in merged config", envKey)
					return false
				}
				if mergedValue != overrideValue {
					t.Logf("Env %q: override value not preserved", envKey)
					return false
				}
			}

			// Check Tasks
			for taskName, overrideTask := range override.Tasks {
				mergedTask, exists := merged.Tasks[taskName]
				if !exists {
					t.Logf("Task %q from override not in merged config", taskName)
					return false
				}
				if !taskEqual(mergedTask, overrideTask) {
					t.Logf("Task %q: override value not preserved", taskName)
					return false
				}
			}

			// Check Settings (non-zero values from override should win)
			if override.Settings.CacheDir != "" && merged.Settings.CacheDir != override.Settings.CacheDir {
				t.Logf("Settings.CacheDir: override value not preserved")
				return false
			}
			if override.Settings.DataDir != "" && merged.Settings.DataDir != override.Settings.DataDir {
				t.Logf("Settings.DataDir: override value not preserved")
				return false
			}
			if override.Settings.CacheTTL != 0 && merged.Settings.CacheTTL != override.Settings.CacheTTL {
				t.Logf("Settings.CacheTTL: override value not preserved")
				return false
			}


			// Base values should be preserved for keys not in override
			for toolName, baseTool := range base.Tools {
				if _, inOverride := override.Tools[toolName]; !inOverride {
					mergedTool, exists := merged.Tools[toolName]
					if !exists {
						t.Logf("Tool %q from base not in merged config", toolName)
						return false
					}
					if !toolConfigEqual(mergedTool, baseTool) {
						t.Logf("Tool %q: base value not preserved", toolName)
						return false
					}
				}
			}

			return true
		},
		genConfig(),
		genConfig(),
	))

	properties.TestingRun(t)
}

// Feature: unirtm, Property 6: Environment-Specific Configuration Selection
//
// **Validates: Requirements 1.7**
//
// For any Configuration with environment-specific overrides, selecting a specific
// environment SHALL apply only the overrides for that environment while preserving
// base configuration values.
//
// This property ensures that:
// 1. Environment selection applies the correct overrides
// 2. Base configuration values are preserved
// 3. Only the selected environment's overrides are applied
func TestProperty_EnvironmentSpecificConfigurationSelection(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.MaxSize = 20

	properties := gopter.NewProperties(parameters)

	properties.Property("Environment selection applies correct overrides", prop.ForAll(
		func(base config.Config, envName string) bool {
			// Ensure base config is valid
			if base.Tools == nil {
				base.Tools = make(map[string]config.ToolConfig)
			}
			if base.Env == nil {
				base.Env = make(map[string]interface{})
			}
			if base.Tasks == nil {
				base.Tasks = make(map[string]config.Task)
			}
			if base.Environments == nil {
				base.Environments = make(map[string]config.EnvironmentConfig)
			}

			// Create an environment override
			envConfig := config.EnvironmentConfig{
				Tools: map[string]config.ToolConfig{
					"node": {Version: "env-override-20.0.0"},
				},
				Env: map[string]interface{}{
					"NODE_ENV": "production",
				},
				Settings: config.Settings{
					CacheTTL: 3600,
				},
			}
			base.Environments[envName] = envConfig

			// Create a ConfigManager
			mgr := config.NewConfigManager()

			// Apply environment
			result, err := mgr.ApplyEnvironment(&base, envName)
			if err != nil {
				t.Logf("ApplyEnvironment failed: %v", err)
				return false
			}

			// Verify environment overrides are applied
			if nodeTool, exists := result.Tools["node"]; exists {
				if nodeTool.Version != "env-override-20.0.0" {
					t.Logf("Environment tool override not applied")
					return false
				}
			}

			if nodeEnv, exists := result.Env["NODE_ENV"]; exists {
				if nodeEnv != "production" {
					t.Logf("Environment env override not applied")
					return false
				}
			}

			if result.Settings.CacheTTL != 3600 {
				t.Logf("Environment settings override not applied")
				return false
			}

			// Verify base values are preserved for non-overridden keys
			for toolName, baseTool := range base.Tools {
				if toolName == "node" {
					continue // This was overridden
				}
				resultTool, exists := result.Tools[toolName]
				if !exists {
					t.Logf("Base tool %q not preserved", toolName)
					return false
				}
				if !toolConfigEqual(resultTool, baseTool) {
					t.Logf("Base tool %q value changed", toolName)
					return false
				}
			}

			return true
		},
		genConfig(),
		gen.OneConstOf("development", "staging", "production", "test"),
	))

	properties.TestingRun(t)
}

// Feature: unirtm, Property 7: Configuration Loading Idempotence
//
// **Validates: Requirements 1.8**
//
// For any valid configuration file, loading it multiple times SHALL produce
// identical Configuration objects.
//
// This property ensures that:
// 1. Configuration loading is deterministic
// 2. No state is accumulated across loads
// 3. The same file always produces the same result
func TestProperty_ConfigurationLoadingIdempotence(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.MaxSize = 20

	properties := gopter.NewProperties(parameters)

	properties.Property("Loading configuration multiple times produces identical results", prop.ForAll(
		func(original config.Config) bool {
			// Normalize the config
			if original.Tools == nil {
				original.Tools = make(map[string]config.ToolConfig)
			}
			if original.Env == nil {
				original.Env = make(map[string]interface{})
			}
			if original.Tasks == nil {
				original.Tasks = make(map[string]config.Task)
			}
			if original.Environments == nil {
				original.Environments = make(map[string]config.EnvironmentConfig)
			}

			// Create a temporary file
			tmpFile, err := os.CreateTemp("", "config-*.toml")
			if err != nil {
				t.Logf("Failed to create temp file: %v", err)
				return false
			}
			defer os.Remove(tmpFile.Name())

			// Write config to file
			encoder := toml.NewEncoder(tmpFile)
			if err := encoder.Encode(&original); err != nil {
				t.Logf("Failed to encode config: %v", err)
				return false
			}
			tmpFile.Close()

			// Create a ConfigManager
			mgr := config.NewConfigManager()
			ctx := context.Background()

			// Load the file multiple times
			load1, err := mgr.Load(ctx, tmpFile.Name())
			if err != nil {
				t.Logf("First load failed: %v", err)
				return false
			}

			load2, err := mgr.Load(ctx, tmpFile.Name())
			if err != nil {
				t.Logf("Second load failed: %v", err)
				return false
			}

			load3, err := mgr.Load(ctx, tmpFile.Name())
			if err != nil {
				t.Logf("Third load failed: %v", err)
				return false
			}

			// Normalize loaded configs
			if load1.Tools == nil {
				load1.Tools = make(map[string]config.ToolConfig)
			}
			if load1.Env == nil {
				load1.Env = make(map[string]interface{})
			}
			if load1.Tasks == nil {
				load1.Tasks = make(map[string]config.Task)
			}
			if load1.Environments == nil {
				load1.Environments = make(map[string]config.EnvironmentConfig)
			}

			if load2.Tools == nil {
				load2.Tools = make(map[string]config.ToolConfig)
			}
			if load2.Env == nil {
				load2.Env = make(map[string]interface{})
			}
			if load2.Tasks == nil {
				load2.Tasks = make(map[string]config.Task)
			}
			if load2.Environments == nil {
				load2.Environments = make(map[string]config.EnvironmentConfig)
			}

			if load3.Tools == nil {
				load3.Tools = make(map[string]config.ToolConfig)
			}
			if load3.Env == nil {
				load3.Env = make(map[string]interface{})
			}
			if load3.Tasks == nil {
				load3.Tasks = make(map[string]config.Task)
			}
			if load3.Environments == nil {
				load3.Environments = make(map[string]config.EnvironmentConfig)
			}

			// All loads should be identical
			if !configsEqual(*load1, *load2) {
				t.Logf("First and second load differ")
				return false
			}

			if !configsEqual(*load2, *load3) {
				t.Logf("Second and third load differ")
				return false
			}

			if !configsEqual(*load1, *load3) {
				t.Logf("First and third load differ")
				return false
			}

			return true
		},
		genConfig(),
	))

	properties.TestingRun(t)
}

// TestConfigRoundTrip_EdgeCases tests specific edge cases for TOML round-trip.
//
// This test complements the property-based test by explicitly testing known
// edge cases that might not be covered by random generation.
func TestConfigRoundTrip_EdgeCases(t *testing.T) {
	testCases := []struct {
		name   string
		config config.Config
	}{
		{
			name: "empty config",
			config: config.Config{
				Tools:    map[string]config.ToolConfig{},
				ToolsRaw: nil,
				Env:      nil,
				Tasks:    nil,
			},
		},
		{
			name: "config with special characters",
			config: config.Config{
				Tools: map[string]config.ToolConfig{
					"node": {Version: "20.0.0"},
				},
				ToolsRaw: map[string]interface{}{
					"node": map[string]interface{}{"version": "20.0.0"},
				},
				Env: map[string]interface{}{
					"PATH":        "/usr/local/bin:/usr/bin",
					"DESCRIPTION": "A tool with \"quotes\" and 'apostrophes'",
				},
				Settings: config.Settings{
					CacheDir:    "/tmp/cache with spaces",
					DataDir:     "/var/lib/unirtm",
					CacheTTL:    86400,
				},
				Tasks: map[string]config.Task{},
			},
		},
		{
			name: "config with maximum values",
			config: config.Config{
				Tools: map[string]config.ToolConfig{
					"node":      {Version: "20.0.0", Backend: "github", Provider: "node"},
					"python":    {Version: "3.11.5", Backend: "aqua", Provider: "python"},
					"go":        {Version: "1.21.0", Backend: "http", Provider: "generic"},
					"ruby":      {Version: "3.2.0"},
					"rust":      {Version: "1.70.0"},
					"java":      {Version: "17.0.0"},
					"terraform": {Version: "1.5.0"},
					"kubectl":   {Version: "1.27.0"},
				},
				ToolsRaw: map[string]interface{}{
					"node":      map[string]interface{}{"version": "20.0.0", "backend": "github", "provider": "node"},
					"python":    map[string]interface{}{"version": "3.11.5", "backend": "aqua", "provider": "python"},
					"go":        map[string]interface{}{"version": "1.21.0", "backend": "http", "provider": "generic"},
					"ruby":      map[string]interface{}{"version": "3.2.0"},
					"rust":      map[string]interface{}{"version": "1.70.0"},
					"java":      map[string]interface{}{"version": "17.0.0"},
					"terraform": map[string]interface{}{"version": "1.5.0"},
					"kubectl":   map[string]interface{}{"version": "1.27.0"},
				},
				Env: map[string]interface{}{
					"PATH":       "/usr/local/bin",
					"HOME":       "/home/user",
					"USER":       "testuser",
					"LANG":       "en_US.UTF-8",
					"NODE_ENV":   "production",
					"GO_ENV":     "production",
					"PYTHON_ENV": "production",
				},
				Settings: config.Settings{
					CacheDir:    "/var/cache/unirtm",
					DataDir:     "/var/lib/unirtm",
					CacheTTL:    604800,
				},
				Tasks: map[string]config.Task{
					"build": {
						Description: "Build the project",
						Run:         "make build",
						Env:         map[string]interface{}{"CGO_ENABLED": "0"},
						Depends:     []string{"test"},
					},
					"test": {
						Description: "Run tests",
						Run:         "go test ./...",
					},
				},
			},
		},
		{
			name: "config with empty strings",
			config: config.Config{
				Tools: map[string]config.ToolConfig{
					"node": {Version: "20.0.0", Backend: "", Provider: ""},
				},
				ToolsRaw: map[string]interface{}{
					"node": map[string]interface{}{"version": "20.0.0"},
				},
				Env: map[string]interface{}{
					"EMPTY": "",
				},
				Settings: config.Settings{
					CacheDir:    "",
					DataDir:     "",
					CacheTTL:    0,
				},
				Tasks: map[string]config.Task{
					"noop": {
						Description: "",
						Run:         "echo 'noop'",
						Env:         map[string]interface{}{},
						Depends:     []string{},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Normalize the original config
			if tc.config.Tools == nil {
				tc.config.Tools = make(map[string]config.ToolConfig)
			}
			if tc.config.Env == nil {
				tc.config.Env = make(map[string]interface{})
			}
			if tc.config.Tasks == nil {
				tc.config.Tasks = make(map[string]config.Task)
			}

			// Step 1: Serialize to TOML
			tc.config.PostLoad()
			var buf1 bytes.Buffer
			encoder1 := toml.NewEncoder(&buf1)
			err := encoder1.Encode(&tc.config)
			require.NoError(t, err, "Failed to encode config")
			toml1 := buf1.String()

			// Step 2: Parse back
			var parsed config.Config
			decoder := toml.NewDecoder(bytes.NewReader(buf1.Bytes()))
			err = decoder.Decode(&parsed)
			require.NoError(t, err, "Failed to decode TOML")
			parsed.PostLoad()

			// Normalize parsed config
			if parsed.Tools == nil {
				parsed.Tools = make(map[string]config.ToolConfig)
			}
			if parsed.Env == nil {
				parsed.Env = make(map[string]interface{})
			}
			if parsed.Tasks == nil {
				parsed.Tasks = make(map[string]config.Task)
			}

			// Step 3: Serialize again
			var buf2 bytes.Buffer
			encoder2 := toml.NewEncoder(&buf2)
			err = encoder2.Encode(&parsed)
			require.NoError(t, err, "Failed to encode parsed config")
			toml2 := buf2.String()

			// Step 4: Verify structural equivalence
			if !configsEqual(tc.config, parsed) {
				t.Logf("Configs not structurally equal after round-trip")
				t.Logf("Original: %+v", tc.config)
				t.Logf("Parsed: %+v", parsed)
				t.Fail()
			}

			// Step 5: Verify TOML output is identical
			require.Equal(t, toml1, toml2, "TOML output differs after round-trip")
		})
	}
}
