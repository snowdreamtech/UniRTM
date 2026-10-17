// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build windows

package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

type windowsPipeSession struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
}

func (s *windowsPipeSession) Read(p []byte) (int, error) {
	return s.stdout.Read(p)
}

func (s *windowsPipeSession) Write(p []byte) (int, error) {
	return s.stdin.Write(p)
}

func (s *windowsPipeSession) Close() error {
	if s.stdin != nil {
		_ = s.stdin.Close()
	}
	if s.stdout != nil {
		_ = s.stdout.Close()
	}
	if s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
		_ = s.cmd.Wait()
	}
	return nil
}

func (s *windowsPipeSession) Resize(cols, rows uint16) {
	// Fallback stub on Windows when ConPTY is not used
}

// startPTYSession creates a streaming pipe session on Windows.
func startPTYSession(shellPath, cwd string, cols, rows uint16) (PTYSession, error) {
	if shellPath == "" {
		if pwsh, err := exec.LookPath("powershell.exe"); err == nil {
			shellPath = pwsh
		} else if cmdPath, err := exec.LookPath("cmd.exe"); err == nil {
			shellPath = cmdPath
		} else {
			shellPath = "cmd.exe"
		}
	}

	cmd := exec.Command(shellPath)
	cmd.Env = buildTerminalEnv()

	if cwd != "" {
		if fi, err := os.Stat(cwd); err == nil && fi.IsDir() {
			cmd.Dir = cwd
		}
	}
	if cmd.Dir == "" {
		if wd, err := os.Getwd(); err == nil {
			cmd.Dir = wd
		}
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to open stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		return nil, fmt.Errorf("failed to open stdout pipe: %w", err)
	}
	cmd.Stderr = cmd.Stdout // merge stderr into stdout stream

	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		_ = stdout.Close()
		return nil, fmt.Errorf("failed to start windows shell: %w", err)
	}

	return &windowsPipeSession{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
	}, nil
}
