#!/usr/bin/env python3
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

"""Synchronize Harden Runner endpoints across workflow files.

This script reads the centralized endpoint configuration from
.github/harden-runner-endpoints.yml and updates all workflow files
with the appropriate endpoint profiles.
"""

from pathlib import Path
from typing import Dict, List

import yaml

CONFIG_FILE = Path(".github/harden-runner-endpoints.yml")
WORKFLOWS_DIR = Path(".github/workflows")


def load_config() -> Dict:
    """Load endpoint configuration from YAML file."""
    with open(CONFIG_FILE, "r", encoding="utf-8") as f:
        return yaml.safe_load(f)


def update_workflow_endpoints(filepath: Path, endpoints: List[str]) -> bool:
    """Update allowed-endpoints in a workflow file."""
    with open(filepath, "r", encoding="utf-8") as f:
        content = f.read()

    original = content
    lines = content.split("\n")
    result = []

    i = 0
    while i < len(lines):
        line = lines[i]

        # Detect allowed-endpoints section
        if "allowed-endpoints:" in line and ">" in line:
            # Multi-line format: allowed-endpoints: >
            result.append(line)
            i += 1

            # Detect indentation from next line
            if i < len(lines) and lines[i].strip():
                base_indent = len(lines[i]) - len(lines[i].lstrip())
            else:
                base_indent = 12  # Default indentation

            # Add new endpoints
            indent = " " * base_indent
            for endpoint in endpoints:
                result.append(f"{indent}{endpoint}")

            # Skip old endpoints (only lines with same or greater indentation that are endpoints)
            while i < len(lines):
                current_line = lines[i]
                if not current_line.strip():
                    # Empty line - keep it and continue
                    result.append(current_line)
                    i += 1
                    continue

                current_indent = len(current_line) - len(current_line.lstrip())

                # If indentation is less than base, we've reached the next section
                if current_indent < base_indent:
                    break

                # If indentation equals base and line looks like an endpoint, skip it
                if current_indent == base_indent and (
                    ":" in current_line
                    and not current_line.strip().startswith("-")
                    and not current_line.strip().endswith(":")
                ):
                    i += 1
                    continue

                # Otherwise, we've reached the next section
                break

            continue

        result.append(line)
        i += 1

    new_content = "\n".join(result)

    if new_content != original:
        with open(filepath, "w", encoding="utf-8") as f:
            f.write(new_content)
        return True
    return False


def main():
    """Main synchronization logic."""
    print("📋 Synchronizing Harden Runner endpoints...")
    print()

    # Load configuration
    try:
        config = load_config()
    except FileNotFoundError:
        print(f"❌ Error: Configuration file {CONFIG_FILE} not found")
        return 1
    except yaml.YAMLError as e:
        print(f"❌ Error parsing YAML: {e}")
        return 1

    workflow_profiles = config.get("workflow_profiles", {})
    updated_count = 0

    # Process each workflow
    for workflow_name, profile_name in workflow_profiles.items():
        workflow_path = WORKFLOWS_DIR / workflow_name

        if not workflow_path.exists():
            print(f"⚠️  Warning: {workflow_name} not found")
            continue

        # Get endpoints for this profile
        endpoints = config.get(profile_name, [])
        if not endpoints:
            print(f"⚠️  Warning: Profile '{profile_name}' not found for {workflow_name}")
            continue

        # Update workflow
        if update_workflow_endpoints(workflow_path, endpoints):
            print(f"✓ Updated: {workflow_name} (profile: {profile_name}, {len(endpoints)} endpoints)")
            updated_count += 1
        else:
            print(f"- Skipped: {workflow_name} (no changes)")

    print()
    print(f"📊 Summary: {updated_count} files updated")
    print()
    print("💡 Tip: Review changes with 'git diff .github/workflows/'")

    return 0


if __name__ == "__main__":
    exit(main())
