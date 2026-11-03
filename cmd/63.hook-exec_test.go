// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// Export a helper function to test the internal splitting logic of hook-exec
func splitArgsHelper(args []string) (baseArgs []string, fileArgs []string) {
	splitIdx := len(args)
	for i := len(args) - 1; i >= 0; i-- {
		if _, err := os.Lstat(args[i]); err != nil {
			break
		}
		splitIdx = i
	}

	if splitIdx == 0 && len(args) > 0 {
		splitIdx = 1
	}

	return args[:splitIdx], args[splitIdx:]
}

func TestSplitArgs(t *testing.T) {
	// Create some temporary files for testing
	tempDir := t.TempDir()
	file1 := filepath.Join(tempDir, "file1.txt")
	file2 := filepath.Join(tempDir, "file2.txt")

	if err := os.WriteFile(file1, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		args         []string
		wantBaseArgs []string
		wantFileArgs []string
	}{
		{
			name:         "typical hook execution",
			args:         []string{"prettier", "--write", file1, file2},
			wantBaseArgs: []string{"prettier", "--write"},
			wantFileArgs: []string{file1, file2},
		},
		{
			name:         "no files provided",
			args:         []string{"eslint", "--fix"},
			wantBaseArgs: []string{"eslint", "--fix"},
			wantFileArgs: []string{},
		},
		{
			name:         "command is only one file",
			args:         []string{file1},
			wantBaseArgs: []string{file1}, // Should keep at least one as base
			wantFileArgs: []string{},
		},
		{
			name:         "tool and one file",
			args:         []string{"cat", file1},
			wantBaseArgs: []string{"cat"},
			wantFileArgs: []string{file1},
		},
		{
			name:         "mixed non-existent files (stops at first non-existent from right)",
			args:         []string{"prettier", "non-existent.txt", file1, file2},
			wantBaseArgs: []string{"prettier", "non-existent.txt"},
			wantFileArgs: []string{file1, file2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBase, gotFile := splitArgsHelper(tt.args)
			if !reflect.DeepEqual(gotBase, tt.wantBaseArgs) {
				t.Errorf("splitArgs() gotBase = %v, want %v", gotBase, tt.wantBaseArgs)
			}
			if !reflect.DeepEqual(gotFile, tt.wantFileArgs) {
				t.Errorf("splitArgs() gotFile = %v, want %v", gotFile, tt.wantFileArgs)
			}
		})
	}
}
