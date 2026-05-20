// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/snowdreamtech/unirtm/internal/addlicense"
	"github.com/spf13/cobra"
)

// ---- flags shared between add and check ------------------------------------

var (
	licenseType    string
	licenseFile    string
	licenseHolder  string
	licenseYear    string
	licenseIgnore  []string
	licenseVerbose bool
	licenseSPDX    string
)

// ---- root license command --------------------------------------------------

func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(licenseCmd)
		licenseCmd.AddCommand(licenseAddCmd)
		licenseCmd.AddCommand(licenseCheckCmd)
	}
}

var licenseCmd = &cobra.Command{
	Use:   "license",
	Short: "Manage copyright license headers in source files",
	Long: `license ensures source code files contain copyright license headers.

Subcommands:
  add    Add missing license headers to source files
  check  Report source files that are missing a license header

Examples:
  unirtm license add   -f .github/license-header.txt cmd/ internal/ scripts/
  unirtm license check -f .github/license-header.txt cmd/ internal/ scripts/
  unirtm license add   -l MIT -c "SnowdreamTech" -y 2026 cmd/`,
}

// ---- addflags is a helper that registers common flags on a command ----------

func addLicenseFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&licenseType, "license", "l", "", "license type: MIT, Apache-2.0, MPL-2.0, bsd (mutually exclusive with --file)")
	cmd.Flags().StringVarP(&licenseFile, "file", "f", "", "path to a custom license header template file")
	cmd.Flags().StringVarP(&licenseHolder, "holder", "c", "", "copyright holder name (used with -l)")
	cmd.Flags().StringVarP(&licenseYear, "year", "y", fmt.Sprint(time.Now().Year()), "copyright year (default: current year)")
	cmd.Flags().StringArrayVar(&licenseIgnore, "ignore", nil, "file pattern to ignore (repeatable, supports ** glob)")
	cmd.Flags().BoolVarP(&licenseVerbose, "verbose", "v", false, "print each modified file")
	cmd.Flags().StringVar(&licenseSPDX, "spdx", "off", "SPDX identifier mode: off | on | only")
}

// ---- license add -----------------------------------------------------------

var licenseAddCmd = &cobra.Command{
	Use:   "add [flags] path [path ...]",
	Short: "Add missing license headers to source files (Safe Mode)",
	Long: `add scans the given paths recursively and prepends a license header
to every source file that does not already have one.

Files that already contain "copyright", "mozilla public", or
"spdx-license-identifier" in the first 1000 bytes are left untouched.
Auto-generated files (matching "Code generated … DO NOT EDIT") are also skipped.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts, err := buildOpts()
		if err != nil {
			return err
		}
		n, err := addlicense.AddLicenseToFiles(args, opts)
		if err != nil {
			return fmt.Errorf("license add: %w", err)
		}
		if n == 0 {
			fmt.Println("✓ All files already have license headers.")
		} else {
			fmt.Printf("✓ Added license headers to %d file(s).\n", n)
		}
		return nil
	},
}

func init() {
	addLicenseFlags(licenseAddCmd)
}

// ---- license check ---------------------------------------------------------

var licenseCheckCmd = &cobra.Command{
	Use:   "check [flags] path [path ...]",
	Short: "Check for missing license headers (exits non-zero if any found)",
	Long: `check scans the given paths recursively and reports every source file
that does not contain a license header.

Exits with code 1 if any files are missing headers (suitable for CI gates).`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts, err := buildOpts()
		if err != nil {
			return err
		}
		n, err := addlicense.CheckLicenseInFiles(args, opts)
		if err != nil && n == 0 {
			return fmt.Errorf("license check: %w", err)
		}
		if n > 0 {
			fmt.Fprintf(os.Stderr, "\n✗ %d file(s) missing license headers.\n", n)
			os.Exit(1)
		}
		fmt.Println("✓ All files have license headers.")
		return nil
	},
}

func init() {
	addLicenseFlags(licenseCheckCmd)
}

// ---- helpers ---------------------------------------------------------------

// buildOpts translates CLI flags into addlicense.Options.
func buildOpts() (addlicense.Options, error) {
	if licenseFile == "" && licenseType == "" {
		return addlicense.Options{}, fmt.Errorf("either --file (-f) or --license (-l) must be specified")
	}
	spdx := addlicense.SpdxOff
	switch licenseSPDX {
	case "on":
		spdx = addlicense.SpdxOn
	case "only":
		spdx = addlicense.SpdxOnly
	case "off", "":
		spdx = addlicense.SpdxOff
	default:
		return addlicense.Options{}, fmt.Errorf("invalid --spdx value %q: must be off, on, or only", licenseSPDX)
	}
	return addlicense.Options{
		License:        licenseType,
		TemplateFile:   licenseFile,
		Holder:         licenseHolder,
		Year:           licenseYear,
		SPDX:           spdx,
		IgnorePatterns: licenseIgnore,
		Verbose:        licenseVerbose,
	}, nil
}
