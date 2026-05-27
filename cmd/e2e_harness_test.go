package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

type E2EHarness struct {
	t      *testing.T
	TmpDir string
}

func NewE2EHarness(t *testing.T) *E2EHarness {
	tmpDir := t.TempDir()
	return &E2EHarness{
		t:      t,
		TmpDir: tmpDir,
	}
}

// Run executes a command in the isolated environment and returns stdout/stderr.
func (h *E2EHarness) Run(args ...string) (stdout string, stderr string, err error) {
	h.t.Helper()

	// Isolate Environment
	h.t.Setenv("UNIRTM_HOME", h.TmpDir)
	h.t.Setenv("XDG_DATA_HOME", h.TmpDir)
	h.t.Setenv("XDG_CONFIG_HOME", h.TmpDir)
	h.t.Setenv("HOME", h.TmpDir)

	// Intercept os.Stdout and os.Stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	// Reset global state
	cwd = ""
	configPath = ""
	verbose = false
	quiet = false
	jsonOutput = false
	dryRun = false
	yes = true
	locked = false
	silent = false

	rootCmd.SetArgs(args)
	err = rootCmd.Execute()

	// Restore Stdout/Stderr
	wOut.Close()
	wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var bufOut bytes.Buffer
	var bufErr bytes.Buffer
	io.Copy(&bufOut, rOut)
	io.Copy(&bufErr, rErr)

	return bufOut.String(), bufErr.String(), err
}

func (h *E2EHarness) SetupMockTool(tool, version string) {
	h.t.Helper()
	
	toolPath := filepath.Join(h.TmpDir, "installs", tool, version)
	err := os.MkdirAll(toolPath, 0755)
	if err != nil {
		h.t.Fatalf("Failed to create mock tool dir: %v", err)
	}

	// Create a dummy binary
	binDir := filepath.Join(toolPath, "bin")
	os.MkdirAll(binDir, 0755)
	dummyBin := filepath.Join(binDir, tool)
	os.WriteFile(dummyBin, []byte("#!/bin/sh\necho mock"), 0755)
}
