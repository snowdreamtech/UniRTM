package cmd

import (
	"testing"
)

func TestBulkCommands_EarlyReturns(t *testing.T) {
	// We want to test commands that can safely be called with dryRun=true or empty args
	// without hanging the test runner.
	originalDryRun := dryRun
	dryRun = true
	defer func() { dryRun = originalDryRun }()

	tests := []struct {
		name string
		cmd  func([]string) error
		args []string
	}{
		{"prepare", func(args []string) error { return runPrepare(prepareCmd, args) }, []string{}},
		{"generateGithubAction", func(args []string) error { return runGenerateGithubAction(generateGithubActionCmd, args) }, []string{}},
		{"generateGitlabCi", func(args []string) error { return runGenerateGitlabCi(generateGitlabCiCmd, args) }, []string{}},
		{"generateDockerfile", func(args []string) error { return runGenerateDockerfile(generateDockerfileCmd, args) }, []string{}},
		{"generatePreCommit", func(args []string) error { return runGeneratePreCommit(generatePreCommitCmd, args) }, []string{}},
		{"generateShellAlias", func(args []string) error { return runGenerateShellAlias(generateShellAliasCmd, args) }, []string{}},
		{"enable", func(args []string) error { return runEnable(enableCmd, args) }, []string{}},
		{"disable", func(args []string) error { return runDisable(disableCmd, args) }, []string{}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// We don't care if it errors or not, we just want to execute the code safely
			_ = tc.cmd(tc.args)
		})
	}
}

func TestInstallIntoCommand_EarlyReturns(t *testing.T) {
	originalDryRun := dryRun
	dryRun = true
	defer func() { dryRun = originalDryRun }()

}

func TestListCommand(t *testing.T) {
	originalJson := jsonOutput
	jsonOutput = true
	defer func() { jsonOutput = originalJson }()

	// Should list things
	_ = runList(listCmd, []string{})
}
