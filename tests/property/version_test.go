// Package property contains property-based tests for UniRTM.
//
// Property-based tests verify universal properties that should hold for all inputs,
// complementing example-based unit tests with comprehensive input coverage.
package property

import (
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/snowdreamtech/unirtm/internal/service"
)

// Feature: unirtm, Property 25: Explicit Version Requirement
//
// **Validates: Requirements 8.3, 8.4**
//
// For any tool request without an explicit version specification, the Version_Resolver
// SHALL return an error requiring the user to specify "latest", "lts", or a concrete version.
//
// This property ensures that:
// 1. No implicit version resolution occurs
// 2. Users must explicitly specify version requirements
// 3. Error messages guide users to valid version specifications
func TestProperty_ExplicitVersionRequirement(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.MaxSize = 20

	properties := gopter.NewProperties(parameters)

	properties.Property("Empty version string requires explicit specification", prop.ForAll(
		func() bool {
			// Empty version string should fail
			_, err := service.ParseVersion("")
			if err == nil {
				t.Logf("Expected error for empty version string")
				return false
			}

			// Error message should be descriptive
			errStr := err.Error()
			if !strings.Contains(errStr, "empty") && !strings.Contains(errStr, "cannot be empty") {
				t.Logf("Error should mention empty version: %v", err)
				return false
			}

			return true
		},
	))

	properties.Property("Whitespace-only version string requires explicit specification", prop.ForAll(
		func(spaces int) bool {
			// Generate whitespace-only string
			versionStr := strings.Repeat(" ", spaces%10+1)

			// Should fail to parse
			_, err := service.ParseVersion(versionStr)
			if err == nil {
				t.Logf("Expected error for whitespace-only version string")
				return false
			}

			return true
		},
		gen.IntRange(0, 100),
	))

	properties.TestingRun(t)
}

// Feature: unirtm, Property 26: Version Specifier Round-Trip
//
// **Validates: Requirements 27.1, 27.2, 27.3, 27.5, 27.6**
//
// For any valid Version object (semver, range, or alias), formatting it to a string
// and parsing it back SHALL produce an equivalent Version object.
//
// This property ensures that:
// 1. Version formatting is deterministic
// 2. Version parsing is lossless
// 3. Round-trip preserves all version data
// 4. No data corruption occurs during formatting/parsing
func TestProperty_VersionSpecifierRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.MaxSize = 20

	properties := gopter.NewProperties(parameters)

	properties.Property("Version round-trip preserves data", prop.ForAll(
		func(original *service.Version) bool {
			// Step 1: Format the original version to a string
			formatted, err := service.FormatVersion(original)
			if err != nil {
				t.Logf("Failed to format version: %v", err)
				return false
			}

			// Step 2: Parse the formatted string back to a Version object
			parsed, err := service.ParseVersion(formatted)
			if err != nil {
				t.Logf("Failed to parse formatted version '%s': %v", formatted, err)
				return false
			}

			// Step 3: Verify structural equivalence
			if !original.Equal(parsed) {
				t.Logf("Versions not equal after round-trip")
				t.Logf("Original: %+v", original)
				t.Logf("Formatted: %s", formatted)
				t.Logf("Parsed: %+v", parsed)
				return false
			}

			// Step 4: Format the parsed version again
			reformatted, err := service.FormatVersion(parsed)
			if err != nil {
				t.Logf("Failed to reformat version: %v", err)
				return false
			}

			// Step 5: Verify formatted strings are identical
			if formatted != reformatted {
				t.Logf("Formatted strings differ after round-trip")
				t.Logf("First: %s", formatted)
				t.Logf("Second: %s", reformatted)
				return false
			}

			return true
		},
		genVersionObject(),
	))

	properties.TestingRun(t)
}

// Feature: unirtm, Property 27: Invalid Version Error Reporting
//
// **Validates: Requirements 27.4**
//
// For any invalid version string, the Version_Parser SHALL return an error
// describing why the version string is invalid.
//
// This property ensures that:
// 1. Invalid version strings are rejected
// 2. Error messages are descriptive and actionable
// 3. Various types of invalid input are handled correctly
func TestProperty_InvalidVersionErrorReporting(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.MaxSize = 20

	properties := gopter.NewProperties(parameters)

	properties.Property("Invalid semver format produces descriptive error", prop.ForAll(
		func(seed int) bool {
			// Generate various invalid semver patterns
			invalidVersions := []string{
				"1.2",       // Missing patch
				"1.2.x",     // Non-numeric patch
				"1.x.3",     // Non-numeric minor
				"x.2.3",     // Non-numeric major
				"1.2.3.4",   // Too many components
				"1.2.3-",    // Empty prerelease
				"1.2.3+",    // Empty build
				"a.b.c",     // All non-numeric
				"1.2.3-@#$", // Invalid prerelease characters
				"1.2.3+@#$", // Invalid build characters
				"...",       // Only dots
				"1..3",      // Double dots
				".1.2.3",    // Leading dot
				"1.2.3.",    // Trailing dot
				"-1.2.3",    // Negative major
				"1.-2.3",    // Negative minor
				"1.2.-3",    // Negative patch
			}

			// Pick one based on seed
			invalidVersion := invalidVersions[seed%len(invalidVersions)]

			// Try to parse it
			_, err := service.ParseVersion(invalidVersion)

			// Should produce an error
			if err == nil {
				t.Logf("Expected parsing error for invalid version: %s", invalidVersion)
				return false
			}

			// Error should be descriptive
			errStr := strings.ToLower(err.Error())
			hasVersionInfo := strings.Contains(errStr, "version") ||
				strings.Contains(errStr, "semver") ||
				strings.Contains(errStr, "invalid") ||
				strings.Contains(errStr, "format")

			if !hasVersionInfo {
				t.Logf("Error should be descriptive for '%s': %v", invalidVersion, err)
				return false
			}

			return true
		},
		gen.IntRange(0, 1000),
	))

	properties.Property("Invalid range format produces descriptive error", prop.ForAll(
		func(seed int) bool {
			// Generate various invalid range patterns
			invalidRanges := []string{
				">=1.2",     // Missing patch in range
				">1.2.x",    // Non-numeric in range
				"<=",        // Missing version
				"^",         // Missing version
				"~",         // Missing version
				">>1.2.3",   // Invalid operator
				"<<1.2.3",   // Invalid operator
				"=<1.2.3",   // Wrong operator order
				"=>1.2.3",   // Wrong operator order
				">=1.2.3.4", // Too many components
				"^1.2.x",    // Non-numeric in caret range
				"~1.x.3",    // Non-numeric in tilde range
				">=a.b.c",   // All non-numeric in range
			}

			// Pick one based on seed
			invalidRange := invalidRanges[seed%len(invalidRanges)]

			// Try to parse it
			_, err := service.ParseVersion(invalidRange)

			// Should produce an error
			if err == nil {
				t.Logf("Expected parsing error for invalid range: %s", invalidRange)
				return false
			}

			// Error should be descriptive
			errStr := strings.ToLower(err.Error())
			hasRangeInfo := strings.Contains(errStr, "range") ||
				strings.Contains(errStr, "version") ||
				strings.Contains(errStr, "invalid") ||
				strings.Contains(errStr, "format")

			if !hasRangeInfo {
				t.Logf("Error should be descriptive for '%s': %v", invalidRange, err)
				return false
			}

			return true
		},
		gen.IntRange(0, 1000),
	))

	properties.Property("Invalid alias produces descriptive error", prop.ForAll(
		func(invalidAlias string) bool {
			// Skip valid aliases and valid semver patterns
			lowerAlias := strings.ToLower(invalidAlias)
			if lowerAlias == "latest" || lowerAlias == "lts" || lowerAlias == "stable" {
				return true // Skip valid aliases
			}

			// Skip if it looks like a valid semver or range
			if strings.Contains(invalidAlias, ".") ||
				strings.HasPrefix(invalidAlias, ">=") ||
				strings.HasPrefix(invalidAlias, ">") ||
				strings.HasPrefix(invalidAlias, "<=") ||
				strings.HasPrefix(invalidAlias, "<") ||
				strings.HasPrefix(invalidAlias, "=") ||
				strings.HasPrefix(invalidAlias, "^") ||
				strings.HasPrefix(invalidAlias, "~") ||
				strings.HasPrefix(invalidAlias, "v") {
				return true // Skip patterns that might be valid
			}

			// Try to parse it
			_, err := service.ParseVersion(invalidAlias)

			// Should produce an error
			if err == nil {
				t.Logf("Expected parsing error for invalid alias: %s", invalidAlias)
				return false
			}

			// Error should be descriptive
			errStr := strings.ToLower(err.Error())
			hasAliasInfo := strings.Contains(errStr, "alias") ||
				strings.Contains(errStr, "version") ||
				strings.Contains(errStr, "invalid") ||
				strings.Contains(errStr, "must be")

			if !hasAliasInfo {
				t.Logf("Error should be descriptive for '%s': %v", invalidAlias, err)
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(v interface{}) bool {
			s := v.(string)
			return len(s) > 0 && len(s) <= 20 // Reasonable length
		}),
	))

	properties.Property("Special characters in version produce descriptive error", prop.ForAll(
		func(seed int) bool {
			// Generate versions with special characters
			specialCharVersions := []string{
				"1.2.3@",
				"1.2.3#",
				"1.2.3$",
				"1.2.3%",
				"1.2.3&",
				"1.2.3*",
				"1.2.3(",
				"1.2.3)",
				"1.2.3[",
				"1.2.3]",
				"1.2.3{",
				"1.2.3}",
				"1.2.3|",
				"1.2.3\\",
				"1.2.3/",
				"1.2.3?",
				"1.2.3<",
				"1.2.3>",
				"1.2.3,",
				"1.2.3;",
				"1.2.3:",
				"1.2.3'",
				"1.2.3\"",
				"1.2.3`",
				"1.2.3~invalid",
				"1.2.3^invalid",
			}

			// Pick one based on seed
			specialCharVersion := specialCharVersions[seed%len(specialCharVersions)]

			// Try to parse it
			_, err := service.ParseVersion(specialCharVersion)

			// Should produce an error
			if err == nil {
				t.Logf("Expected parsing error for version with special chars: %s", specialCharVersion)
				return false
			}

			// Error should be descriptive
			errStr := err.Error()
			if len(errStr) == 0 {
				t.Logf("Error message should not be empty")
				return false
			}

			return true
		},
		gen.IntRange(0, 1000),
	))

	properties.TestingRun(t)
}

// genVersionObject generates random Version objects for property-based testing.
//
// The generator creates versions of all types:
// - Exact versions (semver with optional prerelease and build)
// - Range versions (>=, >, <=, <, =, ^, ~)
// - Alias versions (latest, lts, stable)
//
// Edge cases covered:
// - Zero versions (0.0.0)
// - Large version numbers
// - Prerelease versions
// - Build metadata
// - All range operators
// - All known aliases
func genVersionObject() gopter.Gen {
	return gen.OneGenOf(
		genExactVersion(),
		genRangeVersion(),
		genAliasVersion(),
	)
}

// genExactVersion generates random exact semver versions.
func genExactVersion() gopter.Gen {
	return gopter.CombineGens(
		genSemVer(),
	).Map(func(values []interface{}) *service.Version {
		return &service.Version{
			Type:  service.VersionTypeExact,
			Exact: values[0].(*service.SemVer),
		}
	})
}

// genRangeVersion generates random range versions.
func genRangeVersion() gopter.Gen {
	return gopter.CombineGens(
		genRangeOperator(),
		genSemVer(),
	).Map(func(values []interface{}) *service.Version {
		return &service.Version{
			Type:     service.VersionTypeRange,
			RangeOp:  values[0].(service.RangeOperator),
			RangeVer: values[1].(*service.SemVer),
		}
	})
}

// genAliasVersion generates random alias versions.
func genAliasVersion() gopter.Gen {
	return gen.OneConstOf(
		service.VersionAliasLatest,
		service.VersionAliasLTS,
		service.VersionAliasStable,
	).Map(func(alias service.VersionAlias) *service.Version {
		return &service.Version{
			Type:  service.VersionTypeAlias,
			Alias: alias,
		}
	})
}

// genSemVer generates random SemVer objects.
func genSemVer() gopter.Gen {
	return gopter.CombineGens(
		gen.IntRange(0, 100), // Major
		gen.IntRange(0, 100), // Minor
		gen.IntRange(0, 100), // Patch
		genPrerelease(),      // Prerelease
		genBuild(),           // Build
	).Map(func(values []interface{}) *service.SemVer {
		return &service.SemVer{
			Major:      values[0].(int),
			Minor:      values[1].(int),
			Patch:      values[2].(int),
			Prerelease: values[3].(string),
			Build:      values[4].(string),
		}
	})
}

// genPrerelease generates random prerelease strings.
func genPrerelease() gopter.Gen {
	return gen.OneConstOf(
		"",
		"alpha",
		"alpha.1",
		"beta",
		"beta.2",
		"rc.1",
		"rc.2",
		"preview1",
		"dev",
		"snapshot",
	)
}

// genBuild generates random build metadata strings.
func genBuild() gopter.Gen {
	return gen.OneConstOf(
		"",
		"build.123",
		"build.456",
		"20130313144700",
		"exp.sha.5114f85",
		"21AF26D3",
	)
}

// genRangeOperator generates random range operators.
func genRangeOperator() gopter.Gen {
	return gen.OneConstOf(
		service.RangeOperatorGTE,
		service.RangeOperatorGT,
		service.RangeOperatorLTE,
		service.RangeOperatorLT,
		service.RangeOperatorEQ,
		service.RangeOperatorCaret,
		service.RangeOperatorTilde,
	)
}
