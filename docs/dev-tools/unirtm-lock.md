# Lockfile (unirtm.lock)

In team collaborations and CI/CD pipelines, simply declaring `node = "20"` is far from enough. To achieve the ultimate goal of "it works on my machine, and it works on the server," UniRTM introduces a strict lockfile mechanism: `unirtm.lock`.

> [!IMPORTANT]
> Always commit the `unirtm.lock` file into your version control system (e.g., Git).

## Why do we need a Lockfile?

If you define the following in your `.unirtm.toml`:

```toml
[tools]
node = "20.x"
golang = "latest"
```

This configuration might resolve to `node 20.11.0` today, but could be `node 20.12.0` next month. If the underlying runtime changes silently, it can lead to fatal inconsistencies across different developers' environments.

The existence of `unirtm.lock` is to record the **absolute exact version** after resolution.

## Anatomy of unirtm.lock

`unirtm.lock` is generated in JSON format. It records not only the specific version but also cryptographic hashes for anti-tampering.

```json
{
  "version": "1.0",
  "tools": {
    "node": {
      "version": "20.11.0",
      "source": "https://nodejs.org/dist/v20.11.0/node-v20.11.0-darwin-arm64.tar.gz",
      "checksum": "sha256:d8b2...a7f",
      "resolved_at": "2024-01-01T12:00:00Z"
    },
    "golang": {
      "version": "1.22.0",
      "source": "https://go.dev/dl/go1.22.0.darwin-arm64.tar.gz",
      "checksum": "sha256:f4e...b1c",
      "resolved_at": "2024-01-01T12:00:05Z"
    }
  }
}
```

### Supply Chain Security & Tamper Resistance

When downloading and installing tools, UniRTM strictly calculates the `sha256` checksum of the downloaded archive and compares it against the `checksum` field in the `unirtm.lock`.
If a mismatch is detected, UniRTM immediately aborts the installation and throws a fatal error! This effectively prevents Man-in-the-Middle (MITM) attacks and malicious replacements on official mirrors, securing your supply chain.

## Workflow Best Practices

1. **Generating the Lockfile**: Whenever you run `unirtm install` or modify your `.unirtm.toml`, UniRTM automatically updates or creates the `unirtm.lock`.
2. **Upgrading Dependencies**: If you intentionally want to bump `node 20.x` to its latest patch, run `unirtm update node`. This resolves the newest version and overwrites the lockfile with the new hash.
3. **CI/CD Environments**: In your CI pipelines, use the appropriate frozen lockfile installation mechanism. If the current configuration diverges from the Lockfile, the command will fail and exit instead of silently updating, thereby ensuring absolute determinism for release builds.
