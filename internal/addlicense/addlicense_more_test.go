// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package addlicense

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAddlicense_CheckLicenseInFiles(t *testing.T) {
	tmpDir := t.TempDir()

	p := filepath.Join(tmpDir, "test.go")
	os.WriteFile(p, []byte("package main\n\nfunc main() {}\n"), 0644)

	opts := Options{
		License: "mit",
		Holder:  "test",
		Year:    "2026",
	}

	missing, err := CheckLicenseInFiles([]string{tmpDir}, opts)
	if err != nil {
		t.Errorf("expected no error running CheckLicenseInFiles")
	}
	if missing == 0 {
		t.Errorf("expected missing to be > 0 because license is missing")
	}

	count, err := AddLicenseToFiles([]string{tmpDir}, opts)
	if err != nil {
		t.Errorf("expected no error adding license")
	}
	if count == 0 {
		t.Errorf("expected count > 0")
	}

	missing, err = CheckLicenseInFiles([]string{tmpDir}, opts)
	if err != nil {
		t.Errorf("expected no error checking license because license is added")
	}
	if missing != 0 {
		t.Errorf("expected missing to be 0 after adding")
	}
}

func TestAddlicense_walk(t *testing.T) {
	tmpDir := t.TempDir()

	// nested
	sub := filepath.Join(tmpDir, "sub")
	os.Mkdir(sub, 0755)

	os.WriteFile(filepath.Join(sub, "test.go"), []byte("package sub\n"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "test.sh"), []byte("#!/bin/sh\n"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("some text\n"), 0644) // unhandled extension probably

	opts := Options{
		License: "mit",
		Holder:  "test",
		Year:    "2026",
	}

	count, _ := AddLicenseToFiles([]string{tmpDir}, opts)
	if count != 2 {
		t.Errorf("expected 2 files added, got %d", count)
	}
}

func TestAddlicense_IgnorePatterns(t *testing.T) {
	opts := Options{
		License:        "mit",
		Holder:         "test",
		Year:           "2026",
		IgnorePatterns: []string{"[invalidpattern"}, // invalid pattern to trigger doublestar error
	}
	_, err := CheckLicenseInFiles([]string{"."}, opts)
	if err == nil {
		t.Errorf("expected error for invalid ignore pattern")
	}

	opts.IgnorePatterns = []string{"**/*.go"} // Valid pattern
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "test.go"), []byte("package main\n"), 0644)

	count, _ := AddLicenseToFiles([]string{tmpDir}, opts)
	if count != 0 {
		t.Errorf("expected 0 because test.go is ignored, got %d", count)
	}
}
