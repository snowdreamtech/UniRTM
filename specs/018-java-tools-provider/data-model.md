# Data Model

The Maven provider manages artifacts using the following inferred data schema parsed from the user's URI:

- `Prefix`: `maven:` or `gradle:`
- `GroupId`: Dot-separated group identifier (e.g., `org.openapitools`). Mapped internally to slash-separated paths (`org/openapitools`) for HTTP downloads.
- `ArtifactId`: The tool name (e.g., `openapi-generator-cli`).
- `Version`: The target release version (e.g., `7.6.0`).

Internal State Transition:

1. `unirtm use maven:group/artifact@version`
2. Generate download path: `$UNIRTM_MAVEN_MIRROR/group_path/artifact/version/artifact-version.jar`
3. Download to `$UNIRTM_CACHE_DIR/maven/artifact@version.jar`
4. Generate shim `$UNIRTM_INSTALLS_DIR/bin/artifact`
