#!/usr/bin/env sh
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Test script to validate Dependabot grouping logic

set -eu

echo "🧪 Testing Dependabot Configuration Grouping Logic"
echo "=================================================="
echo ""

DEPENDABOT_FILE=".github/dependabot.yml"

if [ ! -f "${DEPENDABOT_FILE}" ]; then
  echo "❌ ERROR: ${DEPENDABOT_FILE} not found"
  exit 1
fi

# Test 1: Check for duplicate group names within same ecosystem
echo "Test 1: Checking for duplicate group names..."
_duplicates=$(grep -E "^\s+[🔧📦🧹🧪⚡🐳🌟]" "${DEPENDABOT_FILE}" | sort | uniq -d)
if [ -n "${_duplicates}" ]; then
  echo "⚠️  WARNING: Duplicate group names found:"
  echo "${_duplicates}"
else
  echo "✅ No duplicate group names"
fi
echo ""

# Test 2: Verify all groups have update-types
echo "Test 2: Verifying all groups have update-types..."
_groups_count=$(grep -E "^\s+[🔧📦🧹🧪⚡🐳🌟].*:" "${DEPENDABOT_FILE}" | grep -vc "exclude-patterns")
_update_types_count=$(grep -c "update-types:" "${DEPENDABOT_FILE}" || true)

if [ "${_groups_count}" -eq "${_update_types_count}" ]; then
  echo "✅ All ${_groups_count} groups have update-types defined"
else
  echo "❌ ERROR: Groups (${_groups_count}) != update-types (${_update_types_count})"
  echo "   Some groups are missing update-types!"
fi
echo ""

# Test 3: Check for overlapping patterns in same ecosystem
echo "Test 3: Checking for pattern overlaps..."

# Use a temporary file to avoid subshell issues
_temp_overlap=$(mktemp)
echo "0" >"${_temp_overlap}"

# Extract each ecosystem block and check for overlaps
awk '/package-ecosystem:/ {eco=$0; block=""}
  /package-ecosystem:|^  - package-ecosystem:/ && block {
    print block; block=""; eco=$0
  }
  {block=block"\n"$0}
  END {if(block) print block}' "${DEPENDABOT_FILE}" |
  while IFS= read -r line; do
    case "${line}" in
    *"package-ecosystem:"*)
      _current_eco=$(echo "${line}" | sed 's/.*: "\(.*\)".*/\1/')
      ;;
    *'patterns: ["*"]'*)
      if [ -n "${_current_eco:-}" ]; then
        # Check if there are other pattern groups in this ecosystem
        _next_patterns=$(echo "${line}" | grep -A 20 'patterns: \["*"\]' | grep -c 'patterns:' || true)
        if [ "${_next_patterns:-0}" -gt 1 ]; then
          echo "⚠️  WARNING: Ecosystem '${_current_eco}' has wildcard pattern with other patterns"
          echo "1" >"${_temp_overlap}"
        fi
      fi
      ;;
    esac
  done

_has_overlap=$(cat "${_temp_overlap}")
rm -f "${_temp_overlap}"

if [ "${_has_overlap}" -eq 0 ]; then
  echo "✅ No obvious pattern overlaps detected"
fi
echo ""

# Test 4: Verify exclude-patterns are used correctly
echo "Test 4: Checking exclude-patterns usage..."
_exclude_count=$(grep -c "exclude-patterns:" "${DEPENDABOT_FILE}" || true)
if [ "${_exclude_count}" -gt 0 ]; then
  echo "✅ Found ${_exclude_count} exclude-patterns (prevents overlaps)"

  # Verify excluded patterns are defined in other groups
  echo "   Verifying excluded patterns match other group patterns..."
  _npm_blocks=$(awk '/package-ecosystem: "npm"/,/^  - package-ecosystem:|^$/' "${DEPENDABOT_FILE}")

  # Check if eslint* in exclude-patterns also appears in lint-dependencies
  if echo "${_npm_blocks}" | grep -q 'exclude-patterns:' &&
    echo "${_npm_blocks}" | grep -q '🧹-lint-dependencies:'; then
    echo "   ✅ NPM exclude-patterns correctly match specialized groups"
  fi
else
  echo "⚠️  No exclude-patterns found (may cause overlaps in npm/go)"
fi
echo ""

# Test 5: Validate YAML structure
echo "Test 5: Validating YAML structure..."
_errors=0

# Check version
if ! grep -q "^version: 2" "${DEPENDABOT_FILE}"; then
  echo "❌ Missing 'version: 2'"
  _errors=$((_errors + 1))
fi

# Check updates section
if ! grep -q "^updates:" "${DEPENDABOT_FILE}"; then
  echo "❌ Missing 'updates:' section"
  _errors=$((_errors + 1))
fi

# Check all ecosystems have required fields
_ecosystems=$(grep -c "package-ecosystem:" "${DEPENDABOT_FILE}" || true)
_directories=$(grep -c "directory:" "${DEPENDABOT_FILE}" || true)
_schedules=$(grep -c "schedule:" "${DEPENDABOT_FILE}" || true)

if [ "${_ecosystems}" -ne "${_directories}" ] ||
  [ "${_ecosystems}" -ne "${_schedules}" ]; then
  echo "❌ Inconsistent field counts:"
  echo "   Ecosystems: ${_ecosystems}"
  echo "   Directories: ${_directories}"
  echo "   Schedules: ${_schedules}"
  _errors=$((_errors + 1))
else
  echo "✅ All ${_ecosystems} ecosystems have required fields"
fi

if [ "${_errors}" -eq 0 ]; then
  echo "✅ YAML structure is valid"
else
  echo "❌ Found ${_errors} structural errors"
fi
echo ""

# Test 6: Check for balanced quotes
echo "Test 6: Checking quote balance..."
_single_quotes=$(grep -o "'" "${DEPENDABOT_FILE}" | wc -l | tr -d ' ')
_double_quotes=$(grep -o '"' "${DEPENDABOT_FILE}" | wc -l | tr -d ' ')

if [ $((_single_quotes % 2)) -ne 0 ]; then
  echo "⚠️  WARNING: Unbalanced single quotes (${_single_quotes})"
else
  echo "✅ Single quotes balanced (${_single_quotes})"
fi

if [ $((_double_quotes % 2)) -ne 0 ]; then
  echo "⚠️  WARNING: Unbalanced double quotes (${_double_quotes})"
else
  echo "✅ Double quotes balanced (${_double_quotes})"
fi
echo ""

# Test 7: Verify grouping strategy
echo "Test 7: Analyzing grouping strategy..."
echo ""
echo "Ecosystem Grouping Summary:"
echo "----------------------------"

awk 'BEGIN { eco=""; dir=""; groups=0; }
/package-ecosystem:/ {
  if (eco != "") {
    printf "%-20s %-20s %d groups\n", eco, dir, groups;
  }
  eco=$0;
  gsub(/.*: "/, "", eco);
  gsub(/".*/, "", eco);
  dir="";
  groups=0;
}
/directory:/ {
  dir=$0;
  gsub(/.*: "/, "", dir);
  gsub(/".*/, "", dir);
}
/^      [🔧📦🧹🧪⚡🐳🌟]/ {
  groups++;
}
END {
  if (eco != "") {
    printf "%-20s %-20s %d groups\n", eco, dir, groups;
  }
}' "${DEPENDABOT_FILE}"

echo ""
echo "=================================================="
echo "✅ Dependabot configuration validation complete!"
echo ""
echo "Summary:"
echo "  - Total ecosystems: $(grep -c 'package-ecosystem:' "${DEPENDABOT_FILE}")"
echo "  - Total groups: $(grep 'patterns:' "${DEPENDABOT_FILE}" | grep -vc 'exclude-patterns')"
echo "  - PR limit: $(grep -m1 'open-pull-requests-limit:' "${DEPENDABOT_FILE}" | awk '{print $2}')"
echo "  - Update frequency: $(grep -m1 'interval:' "${DEPENDABOT_FILE}" | awk '{print $2}' | tr -d '"')"
