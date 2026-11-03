# Implementation Plan: Fix release-please VERSION file sync

## 1. Changes

* cate the `"VERSION"` string inside the `extra-files` array.
* Replace `"VERSION"` with an object `{"type": "xml", "path": "VERSION"}` to instruct `release-please` to use the XML parser, which supports generic `x-release-please` annotations.

## 2. Verification

* Verify that `.release-please-config.json` is a valid JSON.
* Run the project's formatting tools to ensure JSON formatting is consistent.
