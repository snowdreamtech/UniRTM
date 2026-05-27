// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── Command structure tests ────────────────────────────────────────────────

func TestExecStructure(t *testing.T) {
	assert.Contains(t, execCmd.Use, "exec", "execCmd.Use should contain 'exec'")
	assert.NotEmpty(t, execCmd.Short, "execCmd.Short should not be empty")
	assert.True(t, execCmd.Run != nil || execCmd.RunE != nil,
		"Run or RunE function should be set")
}

func TestExecAlias(t *testing.T) {
	aliases := execCmd.Aliases
	assert.Contains(t, aliases, "x", "exec command should have 'x' alias")
}

func TestExecFlags(t *testing.T) {
	cases := []struct {
		name     string
		flagName string
	}{
		{"exec-command flag", "exec-command"},
		{"raw flag", "raw"},
		{"fresh-env flag", "fresh-env"},
		{"no-deps flag", "no-deps"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := execCmd.Flags().Lookup(tc.flagName)
			require.NotNil(t, f, "flag --%s should be registered", tc.flagName)
		})
	}
}

// TestExecFlagNamesDoNotConflictWithGlobalConfig verifies that exec-specific
// flags do not accidentally shadow the global -c/--config flag.
func TestExecFlagNamesDoNotConflictWithGlobalConfig(t *testing.T) {
	// The global flag is registered on the root command as persistent.
	f := execCmd.Flags().Lookup("exec-command")
	require.NotNil(t, f)
	assert.Equal(t, "x", f.Shorthand, "-x shorthand should be registered for exec-command")

	// Global -c must still reach the root command.
	globalC := rootCmd.PersistentFlags().Lookup("config")
	require.NotNil(t, globalC, "global --config flag should still exist")
	assert.Equal(t, "c", globalC.Shorthand)
}

// ─── mergeEnvMaps tests ─────────────────────────────────────────────────────

func TestMergeEnvMapsBasic(t *testing.T) {
	dst := map[string]string{"A": "1"}
	src := map[string]string{"B": "2"}
	mergeEnvMaps(dst, src)
	assert.Equal(t, "1", dst["A"])
	assert.Equal(t, "2", dst["B"])
}

func TestMergeEnvMapsPathCombination(t *testing.T) {
	sep := string(os.PathListSeparator)
	dst := map[string]string{"PATH": "/usr/bin"}
	src := map[string]string{"PATH": "/tool/bin"}
	mergeEnvMaps(dst, src)
	// src PATH should be prepended to dst PATH.
	assert.Equal(t, "/tool/bin"+sep+"/usr/bin", dst["PATH"])
}

func TestMergeEnvMapsPathEmptyDst(t *testing.T) {
	dst := map[string]string{}
	src := map[string]string{"PATH": "/tool/bin"}
	mergeEnvMaps(dst, src)
	assert.Equal(t, "/tool/bin", dst["PATH"])
}

func TestMergeEnvMapsMultipleSources(t *testing.T) {
	sep := string(os.PathListSeparator)
	dst := map[string]string{}
	// Simulate two tools being merged in order.
	mergeEnvMaps(dst, map[string]string{"PATH": "/node/bin", "NODE_ENV": "prod"})
	mergeEnvMaps(dst, map[string]string{"PATH": "/python/bin", "PYTHONHOME": "/python"})

	// python PATH should be prepended after node PATH was set.
	assert.True(t, strings.HasPrefix(dst["PATH"], "/python/bin"+sep),
		"python PATH should lead")
	assert.Contains(t, dst["PATH"], "/node/bin")
	assert.Equal(t, "prod", dst["NODE_ENV"])
	assert.Equal(t, "/python", dst["PYTHONHOME"])
}

// ─── applyEnvMap tests ──────────────────────────────────────────────────────

func TestApplyEnvMapSetsVars(t *testing.T) {
	const key = "UNIRTM_TEST_EXEC_VAR"
	t.Cleanup(func() { os.Unsetenv(key) })

	applyEnvMap(map[string]string{key: "hello"})
	assert.Equal(t, "hello", os.Getenv(key))
}

func TestApplyEnvMapPrependsPath(t *testing.T) {
	const testDir = "/unirtm/test/bin"
	sep := string(os.PathListSeparator)

	// Snapshot and restore PATH.
	original := os.Getenv("PATH")
	t.Cleanup(func() { os.Setenv("PATH", original) })

	applyEnvMap(map[string]string{"PATH": testDir})

	result := os.Getenv("PATH")
	assert.True(t, strings.HasPrefix(result, testDir+sep) || result == testDir,
		"PATH should start with the injected directory, got: %s", result)
}

func TestApplyEnvMapSkipsEmptyValues(t *testing.T) {
	const key = "UNIRTM_TEST_EMPTY_KEY"
	os.Unsetenv(key)

	applyEnvMap(map[string]string{key: ""})
	_, set := os.LookupEnv(key)
	assert.False(t, set, "empty env vars should not be set")
}

// ─── Argument parsing behaviour tests ──────────────────────────────────────
//
// These tests exercise the argument-splitting logic inside runExec indirectly
// by validating that execCmd accepts at least 0 args (cobra.ArbitraryArgs /
// custom validation inside RunE).

func TestExecAcceptsArbitraryArgs(t *testing.T) {
	// execCmd should not reject unknown positional args at validation time.
	// The RunE handler is responsible for semantics; cobra should not block.
	assert.Nil(t, execCmd.Args, "execCmd.Args should be nil (arbitrary args)")
}

// ─── Dry-run integration test ────────────────────────────────────────────────

func TestExecDryRunDoesNotExecute(t *testing.T) {
	previous := dryRun
	dryRun = true
	t.Cleanup(func() { dryRun = previous })

	// Also disable auto-install to prevent hanging
	prevNoDeps := execNoDeps
	execNoDeps = true
	t.Cleanup(func() { execNoDeps = prevNoDeps })

	// execCmd with dry-run should return nil without actually running anything.
	// We pass a harmless command that would fail if actually executed.
	err := runExec(execCmd, []string{"--", "false"})
	assert.NoError(t, err, "dry-run should succeed without executing the command")
}

// ─── Command-not-found error test ────────────────────────────────────────────

func TestExecCommandNotFound(t *testing.T) {
	previous := dryRun
	dryRun = false
	t.Cleanup(func() { dryRun = previous })

	prevNoDeps := execNoDeps
	execNoDeps = true
	t.Cleanup(func() { execNoDeps = prevNoDeps })

	err := runExec(execCmd, []string{"--", "__unirtm_nonexistent_cmd_12345__"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "command not found",
		"should report command not found for unknown binary")
}

// ─── No-command error test ───────────────────────────────────────────────────

func TestExecNoCommandReturnsError(t *testing.T) {
	// Passing only "--" with no subsequent command should return an error.
	err := runExec(execCmd, []string{"--"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no command specified")
}
