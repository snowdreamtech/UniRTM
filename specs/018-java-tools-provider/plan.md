# Implementation Plan: Java Tools Provider (Maven/Gradle)

## 1. Goal Description

Implement a unified package provider for the Java ecosystem in UniRTM that natively supports `maven:` and `gradle:` artifact resolution, downloads, and intelligent cross-platform shimming.

## 2. Technical Context

- **Provider Framework**: The new provider will integrate into `internal/provider/`.
- **URL Resolution**: The logic will parse the format `maven:groupId/artifactId@version` (e.g., `maven:org.flywaydb/flyway-commandline@10.0.0`) and resolve it to Maven Central's standard layout.
- **Cross-Platform Shim**:
  - Unix: `exec java -jar "$JAR_PATH" "$@"`
  - Windows: `@echo off\r\njava -jar "%JAR_PATH%" %*`
- **Mirror Environment Variables**: We will read `UNIRTM_MAVEN_MIRROR` and `UNIRTM_GRADLE_MIRROR` to override the default Maven Central domain.

## 3. Proposed Changes

### `internal/provider/`

#### [NEW] [maven.go](file:///Users/snowdream/Workspace/snowdreamtech/UniRTM/internal/provider/maven.go)

- Implement the `Provider` interface.
- Add registry hooks for `maven` and `gradle`.
- **Download logic**: Constructs URL `https://repo1.maven.org/maven2/<group_path>/<artifact>/<version>/<artifact>-<version>.jar`.
- **Shim generation logic**: Automatically builds a custom shell/cmd wrapper referencing the `.jar` path.

#### [NEW] [maven_test.go](file:///Users/snowdream/Workspace/snowdreamtech/UniRTM/internal/provider/maven_test.go)

- Table-driven tests validating URL construction across default and mirrored URLs.
- Cross-platform shim generation string matching.

#### [MODIFY] [registry.go](file:///Users/snowdream/Workspace/snowdreamtech/UniRTM/internal/provider/registry.go)

- Add `maven.go` provider initialization to the `DefaultRegistry`.

## 4. Verification Plan

- **Automated Tests**: Execute `go test ./internal/provider/ -run TestMaven` across Linux, macOS, and Windows.
- **Manual Verification**: Run `unirtm use maven:com.puppycrawl.tools/checkstyle` and assert `checkstyle --version` outputs the expected version string.

## 5. User Review Required
>
> [!IMPORTANT]
> The default mirror used will be `https://repo1.maven.org/maven2/`.
> Is there any specific edge case regarding SNAPSHOT versions you want to handle? Currently, this plan focuses on Release versions since SNAPSHOTs have a more complex metadata resolution requirement. For now, we will assume strict semantic release versions.
