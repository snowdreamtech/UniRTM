# Phase 1: Quickstart & Validation

## Quickstart

This feature is an internal library mechanism rather than a user-facing CLI command. However, it can be tested directly using Go's testing framework to ensure normalization works flawlessly.

### Validation Scenario

1. **Verify Unit Tests**:
   Ensure you are in the repository root and execute the Go tests targeting the backend module.

   ```bash
   go test ./internal/backend -v -run TestNormalizeVersionPrefix
   ```

2. **Expected Outcome**:
   The test suite should output `PASS`. Internally, it validates:
   - `1.2.3` + requiresVPrefix=true -> `v1.2.3`
   - `v1.2.3` + requiresVPrefix=false -> `1.2.3`
   - `latest` + requiresVPrefix=true -> `latest` (Unchanged)
   - `beta-1.0` + requiresVPrefix=false -> `beta-1.0` (Unchanged)
