# Feature Specification: Fix release-please VERSION file sync

## 1. Overview
The goal of this feature is to ensure that the `VERSION` file is correctly synchronized by `release-please` when a new release is generated.

## 2. Motivation
The user reported that the `VERSION` file is not updated during releases despite having the `x-release-please-start-version` and `x-release-please-end` generic annotations. This happens because `release-please` does not automatically parse files named `VERSION` without an extension unless a type is explicitly specified in the configuration.

## 3. Requirements
*   **Correct Synchronization:** The `VERSION` file must be parsed by `release-please` using its generic annotations.
*   **Backward Compatibility:** No other release mechanisms should be affected.

## 4. Proposed Solution
Update `.release-please-config.json` to change the `VERSION` entry in `extra-files` to an object that specifies `"type": "xml"`. The `xml` type is the recommended workaround in `release-please` for enabling generic annotations in plain text files without recognizable extensions.
