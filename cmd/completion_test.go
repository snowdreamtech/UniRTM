// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCompletionCommand tests the completion command functionality
func TestCompletionCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		wantErr        bool
		wantContains   []string
		wantNotContain []string
	}{
		{
			name:         "bash completion",
			args:         []string{"completion", "bash"},
			wantErr:      false,
			wantContains: []string{"# bash completion for unirtm", "__unirtm_debug"},
		},
		{
			name:         "zsh completion",
			args:         []string{"completion", "zsh"},
			wantErr:      false,
			wantContains: []string{"#compdef unirtm", "_unirtm()"},
		},
		{
			name:         "fish completion",
			args:         []string{"completion", "fish"},
			wantErr:      false,
			wantContains: []string{"# fish completion for unirtm", "__unirtm_perform_completion"},
		},
		{
			name:         "powershell completion",
			args:         []string{"completion", "powershell"},
			wantErr:      false,
			wantContains: []string{"# powershell completion for unirtm", "__unirtm_debug"},
		},
		{
			name:    "invalid shell type",
			args:    []string{"completion", "invalid"},
			wantErr: true,
		},
		{
			name:    "no shell type provided",
			args:    []string{"completion"},
			wantErr: true,
		},
		{
			name:    "too many arguments",
			args:    []string{"completion", "bash", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture output
			buf := new(bytes.Buffer)

			// Create a temporary completion command with output redirected
			tempCompletionCmd := &cobra.Command{
				Use:                   "completion [bash|zsh|fish|powershell]",
				Short:                 "Generate shell completion script",
				DisableFlagsInUseLine: true,
				ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
				Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
				Run: func(cmd *cobra.Command, args []string) {
					var err error
					switch args[0] {
					case "bash":
						err = cmd.Root().GenBashCompletion(buf)
					case "zsh":
						err = cmd.Root().GenZshCompletion(buf)
					case "fish":
						err = cmd.Root().GenFishCompletion(buf, true)
					case "powershell":
						err = cmd.Root().GenPowerShellCompletionWithDesc(buf)
					}
					if err != nil {
						cmd.PrintErrf("Error generating %s completion: %v\n", args[0], err)
					}
				},
			}

			// Create a new root command for each test to avoid state pollution
			cmd := &cobra.Command{
				Use:   "unirtm",
				Short: "Universal Runtime Manager",
			}
			cmd.AddCommand(tempCompletionCmd)

			// Set output for error messages
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

			// Execute command
			err := cmd.Execute()

			// Check error expectation
			if tt.wantErr {
				require.Error(t, err, "Expected error but got none")
				return
			}

			require.NoError(t, err, "Unexpected error: %v", err)

			// Check output contains expected strings
			output := buf.String()
			for _, want := range tt.wantContains {
				assert.Contains(t, output, want, "Output should contain: %s", want)
			}

			// Check output does not contain unwanted strings
			for _, notWant := range tt.wantNotContain {
				assert.NotContains(t, output, notWant, "Output should not contain: %s", notWant)
			}

			// Verify output is not empty for valid completions
			if !tt.wantErr {
				assert.NotEmpty(t, output, "Completion output should not be empty")
				// Verify output has multiple lines (completion scripts are multi-line)
				lines := strings.Split(strings.TrimSpace(output), "\n")
				assert.Greater(t, len(lines), 10, "Completion script should have multiple lines")
			}
		})
	}
}

// TestCompletionCommandHelp tests the help output of the completion command
func TestCompletionCommandHelp(t *testing.T) {
	cmd := &cobra.Command{
		Use:   "unirtm",
		Short: "Universal Runtime Manager",
	}
	cmd.AddCommand(completionCmd)

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"completion", "--help"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()

	// Verify help output contains key information
	expectedStrings := []string{
		"Generate shell completion script",
		"bash",
		"zsh",
		"fish",
		"powershell",
		"source <(unirtm completion bash)",
		"unirtm completion zsh",
		"unirtm completion fish",
		"unirtm completion powershell",
	}

	for _, expected := range expectedStrings {
		assert.Contains(t, output, expected, "Help output should contain: %s", expected)
	}
}

// TestCompletionValidArgs tests that only valid shell types are accepted
func TestCompletionValidArgs(t *testing.T) {
	validArgs := []string{"bash", "fish", "powershell", "zsh"}

	// Sort both slices for comparison since order doesn't matter
	expected := make([]string, len(validArgs))
	copy(expected, validArgs)
	actual := make([]string, len(completionCmd.ValidArgs))
	copy(actual, completionCmd.ValidArgs)

	assert.ElementsMatch(t, expected, actual,
		"ValidArgs should contain all expected shell types")
}

// TestCompletionCommandStructure tests the command structure and metadata
func TestCompletionCommandStructure(t *testing.T) {
	assert.Equal(t, "completion [bash|zsh|fish|powershell]", completionCmd.Use,
		"Use field should match expected format")

	assert.NotEmpty(t, completionCmd.Short, "Short description should not be empty")
	assert.NotEmpty(t, completionCmd.Long, "Long description should not be empty")

	assert.True(t, completionCmd.DisableFlagsInUseLine,
		"DisableFlagsInUseLine should be true for cleaner usage output")

	assert.NotNil(t, completionCmd.Run, "Run function should be defined")
}
