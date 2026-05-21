// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"

	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/spf13/cobra"
)

var (
	generateOutput string
)

func init() {
	generateGithubActionCmd.Flags().StringVarP(&generateOutput, "output", "o", "", "write to file instead of stdout")
	generateGitlabCiCmd.Flags().StringVarP(&generateOutput, "output", "o", "", "write to file instead of stdout")
	generateDockerfileCmd.Flags().StringVarP(&generateOutput, "output", "o", "", "write to file instead of stdout")
	generatePreCommitCmd.Flags().StringVarP(&generateOutput, "output", "o", "", "write to file instead of stdout")

	generateCmd.AddCommand(generateGithubActionCmd)
	generateCmd.AddCommand(generateGitlabCiCmd)
	generateCmd.AddCommand(generateDockerfileCmd)
	generateCmd.AddCommand(generatePreCommitCmd)
	generateCmd.AddCommand(generateShellAliasCmd)
	if rootCmd != nil {
		rootCmd.AddCommand(generateCmd)
	}
}

// generateCmd is the root of the generate sub-command group.
var generateCmd = &cobra.Command{
	Use:     "generate",
	Short:   "Generate integration files (GitHub Actions, pre-commit hooks, etc.)",
	Aliases: []string{"gen"},
	Long: `Generate integration files for common tooling.

Sub-commands:
  github-action   Generate a GitHub Actions workflow step
  gitlab-ci       Generate a GitLab CI script snippet
  dockerfile      Generate a Dockerfile snippet for UniRTM
  pre-commit      Generate a .pre-commit-hooks.yaml snippet
  shell-alias     Print shell alias definitions

Examples:
  unirtm generate github-action
  unirtm generate gitlab-ci
  unirtm generate dockerfile
  unirtm generate pre-commit --output .pre-commit-hooks.yaml
  unirtm generate shell-alias >> ~/.zshrc`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// ─── github-action ────────────────────────────────────────────────────────────

var generateGithubActionCmd = &cobra.Command{
	Use:   "github-action",
	Short: "Generate a GitHub Actions workflow snippet for UniRTM",
	Args:  cobra.NoArgs,
	RunE:  runGenerateGithubAction,
}

const githubActionTemplate = `# Add this step to your GitHub Actions workflow to install UniRTM:
- name: Install UniRTM
  uses: snowdreamtech/setup-unirtm@v1
  # Or install manually:
  # run: curl -fsSL https://github.com/snowdreamtech/unirtm/raw/main/install.sh | sh

- name: Install tools
  run: unirtm install

- name: Verify lock file
  run: unirtm lock --check
`

func runGenerateGithubAction(cmd *cobra.Command, args []string) error {
	return writeOrPrint(generateOutput, githubActionTemplate)
}

// ─── gitlab-ci ────────────────────────────────────────────────────────────────

var generateGitlabCiCmd = &cobra.Command{
	Use:   "gitlab-ci",
	Short: "Generate a GitLab CI script snippet for UniRTM",
	Args:  cobra.NoArgs,
	RunE:  runGenerateGitlabCi,
}

const gitlabCiTemplate = `# Add this to your .gitlab-ci.yml to install UniRTM:
.unirtm-setup:
  before_script:
    - curl -fsSL https://github.com/snowdreamtech/unirtm/raw/main/install.sh | sh
    - export PATH="$HOME/.local/share/unirtm/shims:$PATH"
    - unirtm install
    - unirtm lock --check
`

func runGenerateGitlabCi(cmd *cobra.Command, args []string) error {
	return writeOrPrint(generateOutput, gitlabCiTemplate)
}

// ─── dockerfile ───────────────────────────────────────────────────────────────

var generateDockerfileCmd = &cobra.Command{
	Use:   "dockerfile",
	Short: "Generate a Dockerfile snippet for UniRTM",
	Args:  cobra.NoArgs,
	RunE:  runGenerateDockerfile,
}

const dockerfileTemplate = `# Add UniRTM to your Dockerfile:
RUN curl -fsSL https://github.com/snowdreamtech/unirtm/raw/main/install.sh | sh
ENV PATH="/root/.local/share/unirtm/shims:$PATH"

# Copy configuration and install tools
COPY .unirtm.toml ./
# COPY unirtm.lock ./
RUN unirtm install
`

func runGenerateDockerfile(cmd *cobra.Command, args []string) error {
	return writeOrPrint(generateOutput, dockerfileTemplate)
}

// ─── pre-commit ───────────────────────────────────────────────────────────────

var generatePreCommitCmd = &cobra.Command{
	Use:   "pre-commit",
	Short: "Generate a pre-commit hook snippet for UniRTM",
	Args:  cobra.NoArgs,
	RunE:  runGeneratePreCommit,
}

const preCommitTemplate = `# Add to .pre-commit-config.yaml:
repos:
  - repo: local
    hooks:
      - id: unirtm-lock-check
        name: UniRTM lock file check
        language: system
        entry: unirtm lock --check
        pass_filenames: false
        always_run: true
`

func runGeneratePreCommit(cmd *cobra.Command, args []string) error {
	return writeOrPrint(generateOutput, preCommitTemplate)
}

// ─── shell-alias ──────────────────────────────────────────────────────────────

var generateShellAliasCmd = &cobra.Command{
	Use:   "shell-alias",
	Short: "Print shell alias definitions for common UniRTM commands",
	Args:  cobra.NoArgs,
	RunE:  runGenerateShellAlias,
}

const shellAliasTemplate = `# UniRTM shell aliases — add to ~/.bashrc or ~/.zshrc:
alias u='unirtm'
alias ui='unirtm install'
alias ul='unirtm list'
alias uu='unirtm outdated'
alias ub='unirtm backends'
`

func runGenerateShellAlias(cmd *cobra.Command, args []string) error {
	return writeOrPrint("", shellAliasTemplate) // always stdout for aliases
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func writeOrPrint(path, content string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
	})
	if path != "" {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			formatter.Error(fmt.Sprintf("Failed to write %s: %v", path, err))
			return err
		}
		formatter.Success(fmt.Sprintf("Written to %s", path), nil)
		return nil
	}
	fmt.Print(content)
	return nil
}
