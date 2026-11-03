// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCalculateDirectorySize(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create some files with known sizes
	file1 := filepath.Join(tempDir, "file1.txt")
	err := os.WriteFile(file1, []byte("hello"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	subDir := filepath.Join(tempDir, "subdir")
	err = os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	file2 := filepath.Join(subDir, "file2.txt")
	err = os.WriteFile(file2, []byte("world!"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Calculate size: "hello" is 5 bytes, "world!" is 6 bytes. Total = 11 bytes.
	size, err := CalculateDirectorySize(tempDir)
	if err != nil {
		t.Fatalf("CalculateDirectorySize failed: %v", err)
	}

	if size != 11 {
		t.Errorf("Expected size 11, got %d", size)
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1023, "1023 B"},
		{1024, "1.0 K"},
		{1536, "1.5 K"},
		{1048576, "1.0 M"},
		{1048576 * 2.5, "2.5 M"},
		{1073741824, "1.0 G"},
	}

	for _, test := range tests {
		result := FormatBytes(test.bytes)
		if result != test.expected {
			t.Errorf("FormatBytes(%d): expected %s, got %s", test.bytes, test.expected, result)
		}
	}
}
