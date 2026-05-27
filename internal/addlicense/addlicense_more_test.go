package addlicense

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAddLicenseToFiles(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.go")
	content := []byte("package main\n")
	os.WriteFile(testFile, content, 0644)

	opts := Options{
		License: "MIT",
		Year:    "2026",
		Holder:  "TestHolder",
		Verbose: false,
	}

	// Test successful add
	_, err := AddLicenseToFiles([]string{testFile}, opts)
	if err != nil {
		t.Fatalf("AddLicenseToFiles failed: %v", err)
	}

	// Test CheckLicenseInFiles (should return true because it has license)
	// We expect 0 files missing license
	count, err := CheckLicenseInFiles([]string{testFile}, opts)
	if err != nil {
		t.Fatalf("CheckLicenseInFiles failed: %v", err)
	}
	if count > 0 {
		t.Errorf("expected CheckLicenseInFiles to find license")
	}

	// Create a file without license
	testFile2 := filepath.Join(dir, "test2.go")
	os.WriteFile(testFile2, []byte("package main\nfunc a(){}\n"), 0644)

	// Test CheckLicenseInFiles (should return missing count = 1)
	count, _ = CheckLicenseInFiles([]string{testFile2}, opts)
	if count != 1 {
		t.Errorf("expected CheckLicenseInFiles to report 1 file missing license")
	}

	// Test walking directory
	_, err = AddLicenseToFiles([]string{dir}, opts)
	if err != nil {
		t.Fatalf("AddLicenseToFiles dir failed: %v", err)
	}
	
	// Test fileHasLicense directly
	hasLic, err := fileHasLicense(testFile)
	if err != nil {
		t.Fatalf("fileHasLicense err: %v", err)
	}
	if !hasLic {
		t.Errorf("expected fileHasLicense to be true")
	}
}
