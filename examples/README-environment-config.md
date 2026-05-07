# Environment-Specific Configuration

UniRTM supports environment-specific configuration overrides, allowing you to define different tool versions, environment variables, settings, and tasks for different environments (e.g., development, staging, production).

## Overview

Environment-specific configuration allows you to:

- Use different tool versions across environments
- Override environment variables per environment
- Adjust settings (cache TTL, concurrency) based on environment needs
- Define environment-specific tasks

## Configuration Structure

The configuration file supports an `environments` section where you can define overrides for specific environments:

```toml
# Base configuration
[tools]
node = { version = "20.0.0" }

[env]
NODE_ENV = "production"

[settings]
cache_ttl = 86400

# Environment-specific overrides
[environments.development]
[environments.development.tools]
node = { version = "18.0.0" }

[environments.development.env]
NODE_ENV = "development"
DEBUG = "true"

[environments.development.settings]
cache_ttl = 3600
```

## Supported Environments

You can define any environment name you need. Common examples include:

- `development` - Local development environment
- `staging` - Pre-production staging environment
- `production` - Production environment
- `test` - Testing environment
- `ci` - Continuous integration environment

## Override Behavior

When an environment is selected, the following merge rules apply:

1. **Tools**: Environment-specific tool versions completely override base tool versions (per tool name)
2. **Environment Variables**: Environment-specific variables are merged with base variables (environment overrides base)
3. **Tasks**: Environment-specific tasks completely override base tasks (per task name)
4. **Settings**: Non-zero environment settings override base settings

## Example Usage

### Basic Example

```toml
# Base configuration
[tools]
node = { version = "20.0.0" }
python = { version = "3.11.0" }

[env]
NODE_ENV = "production"

# Development environment
[environments.development]
[environments.development.tools]
node = { version = "18.0.0" }

[environments.development.env]
NODE_ENV = "development"
DEBUG = "true"
```

When loading with the `development` environment:

- Node.js version will be `18.0.0` (overridden)
- Python version will be `3.11.0` (preserved from base)
- `NODE_ENV` will be `development` (overridden)
- `DEBUG` will be `true` (added)

### Complete Example

See [environment-config.toml](./environment-config.toml) for a comprehensive example with multiple environments.

## API Usage

### Loading Configuration with Environment

```go
import (
    "context"
    "github.com/snowdreamtech/unirtm/internal/config"
)

func main() {
    ctx := context.Background()
    manager := config.NewConfigManager()

    // Load configuration with development environment
    cfg, err := manager.LoadWithEnvironment(ctx, "development")
    if err != nil {
        // Handle error
    }

    // Use the configuration
    // cfg.Tools["node"].Version will be the development override
}
```

### Applying Environment to Existing Configuration

```go
// Load base configuration
cfg, err := manager.LoadHierarchy(ctx)
if err != nil {
    // Handle error
}

// Apply environment-specific overrides
cfg, err = manager.ApplyEnvironment(cfg, "production")
if err != nil {
    // Handle error
}
```

## Validation

Environment-specific configurations are validated just like base configurations:

- Tool versions must be specified
- Settings values must be non-negative
- Task run commands must be present
- All validation rules apply to environment overrides

## Best Practices

1. **Define sensible defaults**: Use the base configuration for production-like defaults
2. **Override only what's needed**: Don't duplicate configuration in environments
3. **Use consistent naming**: Stick to standard environment names (development, staging, production)
4. **Document environment-specific behavior**: Add comments explaining why overrides are needed
5. **Test environment configurations**: Validate that each environment configuration works as expected

## Common Use Cases

### Different Tool Versions for Testing

```toml
[tools]
node = { version = "20.0.0" }

[environments.compatibility-test]
[environments.compatibility-test.tools]
node = { version = "18.0.0" }  # Test with older version
```

### Environment-Specific Database URLs

```toml
[env]
DATABASE_URL = "postgres://prod-db:5432/myapp"

[environments.development]
[environments.development.env]
DATABASE_URL = "postgres://localhost:5432/myapp_dev"

[environments.staging]
[environments.staging.env]
DATABASE_URL = "postgres://staging-db:5432/myapp"
```

### Performance Tuning per Environment

```toml
[settings]
concurrency = 16
cache_ttl = 86400

[environments.development]
[environments.development.settings]
concurrency = 4      # Lower concurrency for local development
cache_ttl = 3600     # Shorter cache for faster iteration

[environments.production]
[environments.production.settings]
concurrency = 32     # Higher concurrency for production load
cache_ttl = 172800   # Longer cache for better performance
```

### Environment-Specific Tasks

```toml
[tasks.deploy]
run = "kubectl apply -f k8s/production/"

[environments.development]
[environments.development.tasks.deploy]
run = "docker-compose up -d"

[environments.staging]
[environments.staging.tasks.deploy]
run = "kubectl apply -f k8s/staging/"
```

## Troubleshooting

### Environment Not Found

If you get an error like `environment "xyz" not found`, ensure:

1. The environment is defined in the `[environments.xyz]` section
2. The environment name matches exactly (case-sensitive)
3. The configuration file is being loaded correctly

### Overrides Not Applied

If overrides aren't being applied:

1. Verify you're using `LoadWithEnvironment()` or `ApplyEnvironment()`
2. Check that the environment name is correct
3. Ensure the override is in the correct section (e.g., `[environments.dev.tools]` not `[environments.tools]`)

### Validation Errors

If you get validation errors:

1. Check that all required fields are present in environment overrides
2. Verify that settings values are non-negative
3. Ensure task run commands are not empty
