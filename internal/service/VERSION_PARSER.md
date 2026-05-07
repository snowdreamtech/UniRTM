# Version Parser and Formatter

## Overview

The version parser and formatter module provides comprehensive support for parsing and formatting version specifications in UniRTM. It handles semantic versions (semver), version ranges, and version aliases.

## Features

### Supported Version Types

1. **Exact Versions (Semver)**
   - Basic semver: `1.2.3`, `0.0.1`
   - With optional `v` prefix: `v1.2.3`
   - With prerelease: `1.2.3-alpha.1`, `2.0.0-rc.1`
   - With build metadata: `1.2.3+build.123`
   - With both: `1.2.3-beta.2+build.456`

2. **Version Ranges**
   - Comparison operators: `>=1.20.0`, `>2.0.0`, `<=3.0.0`, `<4.0.0`, `=1.2.3`
   - Caret ranges (compatible with): `^3.11.0`, `^1.2.3`
   - Tilde ranges (approximately equivalent): `~2.7.0`, `~1.2`, `~1`

3. **Version Aliases**
   - `latest` - Latest version
   - `lts` - Latest LTS (Long Term Support) version
   - `stable` - Latest stable version
   - Case-insensitive: `LATEST`, `Latest`, `latest` all work

## API Reference

### Types

#### `Version`

Represents a version specification.

```go
type Version struct {
    Type VersionType

    // For VersionTypeExact
    Exact *SemVer

    // For VersionTypeRange
    RangeOp  RangeOperator
    RangeVer *SemVer

    // For VersionTypeAlias
    Alias VersionAlias
}
```

#### `SemVer`

Represents a semantic version.

```go
type SemVer struct {
    Major      int
    Minor      int
    Patch      int
    Prerelease string // e.g., "alpha.1", "beta.2", "rc.1"
    Build      string // e.g., "20130313144700"
}
```

### Functions

#### `ParseVersion(versionStr string) (*Version, error)`

Parses a version string into a Version object.

**Parameters:**

- `versionStr` - The version string to parse

**Returns:**

- `*Version` - The parsed version object
- `error` - Error if parsing fails

**Examples:**

```go
// Exact version
v, err := ParseVersion("1.20.0")
// v.Type == VersionTypeExact
// v.Exact.Major == 1, v.Exact.Minor == 20, v.Exact.Patch == 0

// Version range
v, err := ParseVersion(">=1.20.0")
// v.Type == VersionTypeRange
// v.RangeOp == RangeOperatorGTE

// Caret range
v, err := ParseVersion("^3.11.0")
// v.Type == VersionTypeRange
// v.RangeOp == RangeOperatorCaret

// Alias
v, err := ParseVersion("latest")
// v.Type == VersionTypeAlias
// v.Alias == VersionAliasLatest
```

#### `FormatVersion(v *Version) (string, error)`

Formats a Version object back into a version string.

**Parameters:**

- `v` - The version object to format

**Returns:**

- `string` - The formatted version string
- `error` - Error if formatting fails

**Examples:**

```go
v := &Version{
    Type: VersionTypeExact,
    Exact: &SemVer{Major: 1, Minor: 20, Patch: 0},
}
str, err := FormatVersion(v)
// str == "1.20.0"

v := &Version{
    Type:    VersionTypeRange,
    RangeOp: RangeOperatorGTE,
    RangeVer: &SemVer{Major: 1, Minor: 20, Patch: 0},
}
str, err := FormatVersion(v)
// str == ">=1.20.0"
```

### Methods

#### `(*Version) String() string`

Returns the string representation of a Version.

```go
v, _ := ParseVersion("^3.11.0")
fmt.Println(v.String()) // Output: ^3.11.0
```

#### `(*Version) Equal(other *Version) bool`

Checks if two Version objects are equivalent.

```go
v1, _ := ParseVersion("1.2.3")
v2, _ := ParseVersion("v1.2.3")
v1.Equal(v2) // true
```

#### `(*SemVer) Equal(other *SemVer) bool`

Checks if two SemVer objects are equivalent.

```go
s1 := &SemVer{Major: 1, Minor: 2, Patch: 3}
s2 := &SemVer{Major: 1, Minor: 2, Patch: 3}
s1.Equal(s2) // true
```

## Round-Trip Property

The parser and formatter support round-trip conversion, meaning:

```
ParseVersion(str) -> FormatVersion() -> ParseVersion() == original
```

This property is validated by the `TestVersionRoundTrip` test.

**Example:**

```go
original := "^3.11.0"
parsed, _ := ParseVersion(original)
formatted, _ := FormatVersion(parsed)
reparsed, _ := ParseVersion(formatted)

// parsed.Equal(reparsed) == true
// formatted == "^3.11.0"
```

## Error Handling

The parser returns descriptive errors for invalid input:

```go
_, err := ParseVersion("")
// Error: "version string cannot be empty"

_, err := ParseVersion("1.2.x")
// Error: "invalid version string '1.2.x': must be a valid semver..."

_, err := ParseVersion(">=1.2.x")
// Error: "invalid range version: invalid semver format: 1.2.x"
```

## Implementation Details

### Regular Expressions

The parser uses several regular expressions to match different version formats:

- `semverRegex` - Matches semantic versions with optional prerelease and build metadata
- `rangeRegex` - Matches version ranges with comparison operators
- `caretRegex` - Matches caret ranges
- `tildeRegex` - Matches tilde ranges

### Parsing Order

The parser checks version strings in the following order:

1. Empty string check
2. Alias check (latest, lts, stable)
3. Caret range check (^)
4. Tilde range check (~)
5. Comparison range check (>=, >, <=, <, =)
6. Exact semver check

This order ensures that more specific patterns are matched before more general ones.

## Testing

The module includes comprehensive unit tests covering:

- Semver parsing with various formats
- Version parsing for all supported types
- Version formatting for all supported types
- Round-trip property validation
- Equality checks for Version and SemVer objects
- Error cases and edge cases

Run tests with:

```bash
go test -v ./internal/service/version_test.go ./internal/service/version.go
```

## Requirements Validation

This implementation validates the following requirements:

- **Requirement 27.1**: Semver version parsing
- **Requirement 27.2**: Version range parsing (>=, ^, ~)
- **Requirement 27.3**: Alias parsing (latest, lts, stable)
- **Requirement 27.4**: Invalid version error reporting
- **Requirement 27.5**: Version formatting
- **Requirement 27.6**: Round-trip property support

## Design Properties

This implementation supports the following design properties:

- **Property 26**: Version Specifier Round-Trip
  - For any valid Version object, formatting to string, parsing back, and formatting again produces an equivalent Version object and identical string output.

## Usage Examples

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/snowdreamtech/unirtm/internal/service"
)

func main() {
    // Parse a version
    v, err := service.ParseVersion("^3.11.0")
    if err != nil {
        panic(err)
    }

    // Check version type
    if v.Type == service.VersionTypeRange {
        fmt.Printf("Range operator: %s\n", v.RangeOp)
        fmt.Printf("Version: %d.%d.%d\n",
            v.RangeVer.Major,
            v.RangeVer.Minor,
            v.RangeVer.Patch)
    }

    // Format back to string
    str, err := service.FormatVersion(v)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Formatted: %s\n", str)
}
```

### Version Comparison

```go
// Parse two versions
v1, _ := service.ParseVersion("1.2.3")
v2, _ := service.ParseVersion("v1.2.3")

// Check equality
if v1.Equal(v2) {
    fmt.Println("Versions are equal")
}
```

### Handling Different Version Types

```go
v, _ := service.ParseVersion(versionStr)

switch v.Type {
case service.VersionTypeExact:
    fmt.Printf("Exact version: %s\n", v.Exact)
case service.VersionTypeRange:
    fmt.Printf("Range: %s %s\n", v.RangeOp, v.RangeVer)
case service.VersionTypeAlias:
    fmt.Printf("Alias: %s\n", v.Alias)
}
```

## Future Enhancements

Potential future enhancements include:

1. Version comparison and ordering (e.g., `v1.Compare(v2)`)
2. Version range matching (e.g., `v.Matches(range)`)
3. Support for additional version formats (date-based, custom)
4. Version constraint resolution
5. Integration with backend version resolution

## References

- [Semantic Versioning 2.0.0](https://semver.org/)
- [npm semver ranges](https://docs.npmjs.com/cli/v6/using-npm/semver)
- UniRTM Requirements Document (Requirement 27)
- UniRTM Design Document (Property 26)
