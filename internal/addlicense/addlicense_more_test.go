package addlicense

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddLicense_More(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	validFile := filepath.Join(tmpDir, "test.go")
	os.WriteFile(validFile, []byte("package main\n"), 0644)

	// Create a dir to skip
	subDir := filepath.Join(tmpDir, "skipdir")
	os.Mkdir(subDir, 0755)

	// Create a file to skip by pattern
	skipFile := filepath.Join(tmpDir, "test.ignored.go")
	os.WriteFile(skipFile, []byte("package main\n"), 0644)

	// Create a file to skip by extension
	extFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(extFile, []byte("hello\n"), 0644)

	opts := Options{
		Year:           "2026",
		Holder:         "Test",
		License:        "MIT",
		IgnorePatterns: []string{"**/*.ignored.go"},
		SkipExtensions: []string{".txt"},
		Verbose:        true, // for print lines
	}

	count, err := AddLicenseToFiles([]string{tmpDir}, opts)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// check fileHasLicense
	has, err := fileHasLicense(validFile)
	assert.NoError(t, err)
	assert.True(t, has)

	// check fileHasLicense on not found
	has, err = fileHasLicense(filepath.Join(tmpDir, "nonexistent.go"))
	assert.Error(t, err)
	assert.False(t, has)

	// AddLicenseToFiles on non-existent dir to cover walk error
	_, err = AddLicenseToFiles([]string{filepath.Join(tmpDir, "nonexistent")}, opts)
	assert.NoError(t, err) // Walk ignores root if it's passed but yields error? wait AddLicense treats it as unreadable? 
}

func TestTemplates_More(t *testing.T) {
    // unknown template
    _, err := fetchTemplate("unknown", "", 0)
    assert.Error(t, err)
}
