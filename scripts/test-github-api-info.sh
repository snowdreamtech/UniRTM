#!/usr/bin/env sh
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Test script for GitHub API info functionality

set -eu

SCRIPT_DIR=$(cd "$(dirname "${0:-}")" && pwd)
. "${SCRIPT_DIR}/lib/github-api-info.sh"

echo "🧪 Testing GitHub API Info Functions"
echo "======================================"
echo ""

# Test 1: Check if GITHUB_TOKEN is available
echo "Test 1: Checking GITHUB_TOKEN availability..."
if [ -n "${GITHUB_TOKEN:-}" ]; then
  echo "✅ GITHUB_TOKEN is set"

  echo ""
  echo "Test 2: Fetching rate limit..."
  _response=$(get_github_rate_limit "GITHUB_TOKEN" "${GITHUB_TOKEN}")

  if [ -n "${_response}" ]; then
    echo "✅ Successfully fetched rate limit"
    echo ""
    echo "Raw response (first 200 chars):"
    echo "${_response}" | head -c 200
    echo "..."
    echo ""

    echo "Test 3: Parsing rate limit..."
    _parsed=$(parse_rate_limit "${_response}")
    echo "✅ Parsed info: ${_parsed}"
    echo ""

    _limit=$(echo "${_parsed}" | cut -d'|' -f1)
    _remaining=$(echo "${_parsed}" | cut -d'|' -f2)
    _used=$(echo "${_parsed}" | cut -d'|' -f3)
    _percentage=$(echo "${_parsed}" | cut -d'|' -f4)
    _reset_time=$(echo "${_parsed}" | cut -d'|' -f5)

    echo "Formatted output:"
    echo "  Limit: ${_limit}"
    echo "  Remaining: ${_remaining}"
    echo "  Used: ${_used}"
    echo "  Usage: ${_percentage}"
    echo "  Reset Time: ${_reset_time}"
  else
    echo "❌ Failed to fetch rate limit"
  fi
else
  echo "⚠️  GITHUB_TOKEN not set, skipping tests"
fi

echo ""
echo "Test 4: Generating summary table..."
export CI_STEP_SUMMARY="/tmp/test-summary-$$.md"
generate_github_api_summary

if [ -f "${CI_STEP_SUMMARY}" ]; then
  echo "✅ Summary generated successfully"
  echo ""
  echo "Generated summary:"
  echo "=================="
  cat "${CI_STEP_SUMMARY}"
  echo "=================="
  rm -f "${CI_STEP_SUMMARY}"
else
  echo "❌ Failed to generate summary"
fi

echo ""
echo "✨ Tests complete!"
