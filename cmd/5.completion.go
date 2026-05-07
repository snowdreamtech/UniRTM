// Package cmd contains all the command-line interface definitions and implementations
// for the unirtm application. This file implements shell completion generation commands.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// init registers the completion command and its subcommands to the root command.
// This function is automatically called when the package is imported.
func init() {
	rootCmd.AddCommand(completionCmd)
}

// completionCmd represents the completion command which generates shell completion scripts.
// Shell completion allows users to use tab completion for commands, flags, and arguments
// in their shell, improving the user experience and reducing typing errors.
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate shell completion script for UniRTM.

To load completions:

Bash:

  $ source <(unirtm completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ unirtm completion bash > /etc/bash_completion.d/unirtm
  # macOS:
  $ unirtm completion bash > $(brew --prefix)/etc/bash_completion.d/unirtm

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ unirtm completion zsh > "${fpath[1]}/_unirtm"

  # You will need to start a new shell for this setup to take effect.

Fish:

  $ unirtm completion fish | source

  # To load completions for each session, execute once:
  $ unirtm completion fish > ~/.config/fish/completions/unirtm.fish

PowerShell:

  PS> unirtm completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> unirtm completion powershell > unirtm.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run:                   runCompletion,
}

// runCompletion generates the shell completion script for the specified shell.
// It delegates to Cobra's built-in completion generation functions.
//
// Parameters:
//   - cmd: The cobra command that triggered this function
//   - args: The arguments passed to the command (shell type)
func runCompletion(cmd *cobra.Command, args []string) {
	var err error
	switch args[0] {
	case "bash":
		err = cmd.Root().GenBashCompletion(os.Stdout)
	case "zsh":
		err = cmd.Root().GenZshCompletion(os.Stdout)
	case "fish":
		err = cmd.Root().GenFishCompletion(os.Stdout, true)
	case "powershell":
		err = cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
	}
	if err != nil {
		cmd.PrintErrf("Error generating %s completion: %v\n", args[0], err)
		os.Exit(1)
	}
}
