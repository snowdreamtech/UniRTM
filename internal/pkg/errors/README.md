# Error Handling Package

This package provides a comprehensive error handling system for UniRTM, implementing custom error types, error classification, and error wrapping patterns as specified in Requirements 12.1 and 12.2.

## Features

- **Sentinel Errors**: Pre-defined error constants for common error conditions
- **Error Classification**: Categorize errors into user, system, and external errors
- **Error Wrapping**: Context-preserving error wrapping using `fmt.Errorf` with `%w`
- **Error Chain Support**: Full support for `errors.Is()` and `errors.As()`
- **Exit Code Mapping**: Automatic exit code determination based on error category

## Sentinel Errors

The package defines six sentinel errors that can be used with `errors.Is()`:

```go
var (
    ErrNotFound          = errors.New("not found")
    ErrAlreadyExists     = errors.New("already exists")
    ErrInvalidConfig     = errors.New("invalid configuration")
    ErrNetworkFailure    = errors.New("network failure")
    ErrChecksumMismatch  = errors.New("checksum mismatch")
    ErrTransactionFailed = errors.New("transaction failed")
)
```

### Usage Example

```go
user, err := repo.FindByID(ctx, id)
if err != nil {
    if errors.Is(err, errors.ErrNotFound) {
        return nil, echo.ErrNotFound
    }
    return nil, err
}
```

## Error Classification

Errors are classified into three categories, each with specific handling requirements:

### 1. User Errors (Exit Code: 1)

Invalid input, configuration errors, version not found.

**Characteristics:**

- Return descriptive error messages
- Suggest corrective actions
- Safe to display to end users

**Example:**

```go
if version == "" {
    return errors.NewUserError("version not specified", errors.ErrInvalidConfig)
}
```

### 2. System Errors (Exit Code: 2)

Disk full, permission denied, database corruption.

**Characteristics:**

- Log full error context
- Return generic user-safe messages
- Require system-level intervention

**Example:**

```go
if err := os.WriteFile(path, data, 0644); err != nil {
    return errors.NewSystemError("failed to write file", err)
}
```

### 3. External Errors (Exit Code: 3)

Network failures, backend API errors, download failures.

**Characteristics:**

- Implement retry logic
- Return wrapped errors with context
- May be transient

**Example:**

```go
resp, err := http.Get(url)
if err != nil {
    return errors.NewExternalError("failed to fetch release", errors.ErrNetworkFailure)
}
```

## Error Wrapping

Use the `Wrap()` function to add context to errors while preserving the error chain:

```go
user, err := repo.FindByID(ctx, id)
if err != nil {
    return nil, errors.Wrap(err, "find user %d", id)
}
```

This produces error messages like: `find user 123: not found`

### Multiple Wrapping Layers

Errors can be wrapped multiple times to build a context chain:

```go
// Layer 1: Repository
if err := db.Query(...); err != nil {
    return errors.Wrap(err, "query users table")
}

// Layer 2: Service
users, err := repo.FindAll(ctx)
if err != nil {
    return errors.Wrap(err, "fetch all users")
}

// Layer 3: Handler
users, err := service.GetUsers(ctx)
if err != nil {
    return errors.Wrap(err, "handle GET /users")
}
```

Result: `handle GET /users: fetch all users: query users table: connection refused`

## Error Checking

### Check Error Category

```go
if errors.IsUserError(err) {
    // Display error message to user
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    os.Exit(1)
}

if errors.IsSystemError(err) {
    // Log full context, show generic message to user
    logger.Error("system error", "error", err)
    fmt.Fprintf(os.Stderr, "A system error occurred. Please contact support.\n")
    os.Exit(2)
}

if errors.IsExternalError(err) {
    // Retry or show transient error message
    fmt.Fprintf(os.Stderr, "External service error: %v\n", err)
    os.Exit(3)
}
```

### Get Exit Code

```go
if err != nil {
    os.Exit(errors.ExitCode(err))
}
```

### Check Specific Errors

```go
if errors.Is(err, errors.ErrNotFound) {
    // Handle not found case
}

if errors.Is(err, errors.ErrChecksumMismatch) {
    // Handle checksum mismatch
}
```

## Best Practices

### 1. Always Wrap Errors with Context

❌ **Bad:**

```go
if err != nil {
    return err
}
```

✅ **Good:**

```go
if err != nil {
    return errors.Wrap(err, "install tool %s version %s", tool, version)
}
```

### 2. Use Appropriate Error Categories

❌ **Bad:**

```go
return errors.New("invalid version")
```

✅ **Good:**

```go
return errors.NewUserError("invalid version format", errors.ErrInvalidConfig)
```

### 3. Preserve Error Chains

❌ **Bad:**

```go
if err != nil {
    return fmt.Errorf("operation failed: %s", err.Error())
}
```

✅ **Good:**

```go
if err != nil {
    return errors.Wrap(err, "operation failed")
}
```

### 4. Check Errors with errors.Is()

❌ **Bad:**

```go
if err.Error() == "not found" {
    // ...
}
```

✅ **Good:**

```go
if errors.Is(err, errors.ErrNotFound) {
    // ...
}
```

### 5. Use Typed Errors for Structured Data

When you need to attach structured data to an error:

```go
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation: %s — %s", e.Field, e.Message)
}

// Usage
if email == "" {
    return errors.NewUserError("validation failed", &ValidationError{
        Field:   "email",
        Message: "email is required",
    })
}
```

## Integration with Logging

Errors should be logged with full context before being returned to users:

```go
if err != nil {
    logger.Error("operation failed",
        "error", err,
        "tool", tool,
        "version", version,
        "category", errors.GetCategory(err).String(),
    )
    return errors.Wrap(err, "install tool %s@%s", tool, version)
}
```

## Testing

The package includes comprehensive unit tests covering:

- Sentinel error definitions
- Error category classification
- Error wrapping and unwrapping
- Error chain preservation
- Exit code mapping
- Multiple wrapping layers

Run tests with:

```bash
go test ./internal/pkg/errors/...
```

## References

- **Requirements**: 12.1, 12.2
- **Design Document**: Error Handling section
- **Go Error Handling**: <https://go.dev/blog/go1.13-errors>
