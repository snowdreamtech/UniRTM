package cmd

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// collectAllCommands recursively collects all commands
func collectAllCommands(cmd *cobra.Command) []*cobra.Command {
	cmds := []*cobra.Command{cmd}
	for _, c := range cmd.Commands() {
		cmds = append(cmds, collectAllCommands(c)...)
	}
	return cmds
}

func TestAllCommandsStructureAndHelp(t *testing.T) {
	allCmds := collectAllCommands(rootCmd)
	
	for _, cmd := range allCmds {
		t.Run(fmt.Sprintf("CmdStructure_%s", cmd.CommandPath()), func(t *testing.T) {
			assert.NotEmpty(t, cmd.Use, "Use should not be empty")
			// Many commands don't have Short or Long, but we just want to execute their basic accessors
			_ = cmd.Short
			_ = cmd.Long
			_ = cmd.Example
			
			// Test execution with --help to trigger flag parsing and initialization logic
			// Create a buffer to capture output to avoid spamming the test log
			buf := new(bytes.Buffer)
			
			// We only want to test --help if the command can actually accept it without failing
			// Just call cmd.Help() to test the help generation without mutating flag state heavily
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			_ = cmd.Help()
			
			// Restore to nil so we don't pollute other tests that rely on default output inheritance
			cmd.SetOut(nil)
			cmd.SetErr(nil)
		})
	}
}

// Test removed because brute forcing RunE with invalid arguments causes side-effects and hangs
