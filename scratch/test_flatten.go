package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func flattenDirectory(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	var visibleEntries []os.DirEntry
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, ".") && name != "__MACOSX" {
			visibleEntries = append(visibleEntries, entry)
		}
	}

	fmt.Printf("Visible entries in %s: %d\n", dir, len(visibleEntries))
	for _, e := range visibleEntries {
		fmt.Printf("  - %s (dir: %v)\n", e.Name(), e.IsDir())
	}

	if len(visibleEntries) == 1 && visibleEntries[0].IsDir() {
		subDirName := visibleEntries[0].Name()
		fmt.Printf("ℹ flattening redundant directory: %s\n", subDirName)
		subDir := filepath.Join(dir, subDirName)
		subEntries, err := os.ReadDir(subDir)
		if err != nil {
			return err
		}

		for _, entry := range subEntries {
			oldPath := filepath.Join(subDir, entry.Name())
			newPath := filepath.Join(dir, entry.Name())
			fmt.Printf("Moving %s -> %s\n", oldPath, newPath)
			if err := os.Rename(oldPath, newPath); err != nil {
				return err
			}
		}

		fmt.Printf("Removing %s\n", subDir)
		if err := os.Remove(subDir); err != nil {
			return err
		}

		// Recursive call to handle nested single directories
		return flattenDirectory(dir)
	}

	return nil
}

func main() {
	tmpDir, _ := os.MkdirTemp("", "flatten-test")
	defer os.RemoveAll(tmpDir)

	// Case 1: nested gradle
	// tmpDir/gradle-8.14.5/bin/...
	targetDir := filepath.Join(tmpDir, "8.14.5")
	os.MkdirAll(filepath.Join(targetDir, "gradle-8.14.5", "bin"), 0755)
	os.WriteFile(filepath.Join(targetDir, "gradle-8.14.5", "bin", "gradle"), []byte("echo gradle"), 0755)

	// Add __MACOSX to test filtering
	os.MkdirAll(filepath.Join(targetDir, "__MACOSX"), 0755)

	fmt.Println("--- Test Case 1 ---")
	err := flattenDirectory(targetDir)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Verify
	if _, err := os.Stat(filepath.Join(targetDir, "bin", "gradle")); err == nil {
		fmt.Println("✅ Success: gradle moved to 8.14.5/bin/gradle")
	} else {
		fmt.Println("❌ Failure: gradle not found in 8.14.5/bin/gradle")
	}
}
