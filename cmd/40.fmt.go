package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var (
	fmtCheck      bool
	fmtRecursive  bool
	fmtAllConfigs bool
)

func init() {
	fmtCmd.Flags().BoolVar(&fmtCheck, "check", false, "check if config is formatted without modifying it (exit 1 if not)")
	fmtCmd.Flags().BoolVarP(&fmtRecursive, "recursive", "r", false, "format all unirtm.toml files in subdirectories")
	fmtCmd.Flags().BoolVarP(&fmtAllConfigs, "all", "a", false, "format all supported config files (unirtm.toml, .tool-versions)")

	if rootCmd != nil {
		rootCmd.AddCommand(fmtCmd)
	}
}

// fmtCmd formats unirtm.toml with canonical key ordering and indentation.
var fmtCmd = &cobra.Command{
	Use:   "fmt",
	Short: "Format configuration files with canonical key ordering",
	Long: `Format configuration files with canonical key ordering and indentation.

Reads UniRTM configuration files, normalizes section ordering, and writes them
back in-place. Use --check in CI to verify formatting without modifying files.

Canonical Section Order:
  [env] -> [tools] -> [tasks] -> [settings] -> [plugins] -> [alias]

Examples:
  # Format the project config file
  unirtm fmt

  # Format all configs in current and subdirectories
  unirtm fmt -r

  # CI mode: exit 1 if files are not formatted
  unirtm fmt --check`,
	Args: cobra.NoArgs,
	RunE: runFmt,
}

// canonicalSectionOrder defines the preferred top-level key order in unirtm.toml.
var canonicalSectionOrder = []string{
	"env",
	"tools",
	"tasks",
	"settings",
	"plugins",
	"alias",
}

func runFmt(cmd *cobra.Command, args []string) error {
	pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgCyan)).WithTextStyle(pterm.NewStyle(pterm.FgBlack)).Println("UniRTM Formatter")
	fmt.Println()

	spinner, _ := pterm.DefaultSpinner.Start("Scanning configuration files...")

	var files []string
	if fmtRecursive {
		filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				// Skip hidden directories and common heavy folders
				name := info.Name()
				if (strings.HasPrefix(name, ".") && name != ".") || name == "node_modules" || name == "vendor" || name == "dist" {
					return filepath.SkipDir
				}
				return nil
			}

			name := info.Name()
			if name == "unirtm.toml" || name == ".unirtm.toml" || (fmtAllConfigs && name == ".tool-versions") {
				files = append(files, path)
			}
			return nil
		})
	} else {
		cfgPath := resolveConfigFilePath(false)
		if _, err := os.Stat(cfgPath); err == nil {
			files = append(files, cfgPath)
		}
		if fmtAllConfigs {
			if _, err := os.Stat(".tool-versions"); err == nil {
				files = append(files, ".tool-versions")
			}
		}
	}

	if len(files) == 0 {
		spinner.Warning("No configuration files found to format.")
		return nil
	}

	spinner.Success(fmt.Sprintf("Found %d file(s)", len(files)))
	fmt.Println()

	var modifiedCount, errorCount int
	for _, path := range files {
		isModified, err := formatFile(path)
		if err != nil {
			pterm.Error.Prefix = pterm.Prefix{Text: "FAILED", Style: pterm.NewStyle(pterm.BgRed, pterm.FgWhite)}
			pterm.Error.Printf("%s: %v\n", path, err)
			errorCount++
			continue
		}

		if isModified {
			if fmtCheck {
				pterm.Warning.Prefix = pterm.Prefix{Text: "CHECK", Style: pterm.NewStyle(pterm.BgYellow, pterm.FgBlack)}
				pterm.Warning.Printf("%s: Needs formatting\n", path)
				modifiedCount++
			} else {
				pterm.Success.Printf("%s: Formatted ✓\n", path)
				modifiedCount++
			}
		} else {
			pterm.Info.Printf("%s: Already formatted\n", path)
		}
	}

	fmt.Println()
	summary := pterm.DefaultTable.WithData(pterm.TableData{
		{"Metric", "Value"},
		{"Total Processed", fmt.Sprintf("%d", len(files))},
		{"Modified/Dirty", fmt.Sprintf("%d", modifiedCount)},
		{"Errors", fmt.Sprintf("%d", errorCount)},
	})
	summary.Render()

	if fmtCheck && modifiedCount > 0 {
		return fmt.Errorf("%d file(s) are not formatted", modifiedCount)
	}

	if errorCount > 0 {
		return fmt.Errorf("formatting completed with %d error(s)", errorCount)
	}

	return nil
}

func formatFile(path string) (bool, error) {
	original, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	var formatted []byte
	if strings.HasSuffix(path, ".toml") {
		var m map[string]interface{}
		if err := toml.Unmarshal(original, &m); err != nil {
			return false, fmt.Errorf("parse TOML: %w", err)
		}
		if m == nil {
			m = make(map[string]interface{})
		}
		formatted, err = formatTOML(m)
		if err != nil {
			return false, fmt.Errorf("format TOML: %w", err)
		}
	} else {
		// Just trim whitespace for other files for now
		formatted = []byte(strings.TrimSpace(string(original)) + "\n")
	}

	if bytes.Equal(original, formatted) {
		return false, nil
	}

	if !fmtCheck {
		if err := os.WriteFile(path, formatted, 0o644); err != nil {
			return false, err
		}
	}

	return true, nil
}

func formatTOML(m map[string]interface{}) ([]byte, error) {
	ordered := make([]string, 0, len(m))
	inCanonical := make(map[string]bool)
	for _, k := range canonicalSectionOrder {
		if _, ok := m[k]; ok {
			ordered = append(ordered, k)
			inCanonical[k] = true
		}
	}
	rest := make([]string, 0)
	for k := range m {
		if !inCanonical[k] {
			rest = append(rest, k)
		}
	}
	sort.Strings(rest)
	ordered = append(ordered, rest...)

	var buf bytes.Buffer
	for i, k := range ordered {
		single := map[string]interface{}{k: m[k]}
		enc := toml.NewEncoder(&buf)
		// Custom formatting: Ensure top-level indentation is nice
		if err := enc.Encode(single); err != nil {
			return nil, err
		}
		if i < len(ordered)-1 {
			buf.WriteByte('\n')
		}
	}
	return buf.Bytes(), nil
}
