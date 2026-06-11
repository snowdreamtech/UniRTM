# Feature Specification: Java Tools Provider (Maven & Gradle)

## 1. Feature Description

This feature introduces unified package management for the Java ecosystem (Maven and Gradle artifacts). It strictly separates the installation of the Java runtime (`java`) from the installation of Java-based tools/packages (`maven:` / `gradle:` providers), mirroring the existing philosophy for Go (`go` runtime vs `go:` package provider). The provider will intelligently download `.jar` artifacts and wrap them in cross-platform shims, enabling users to seamlessly run Java CLI tools directly from their terminals.

## 2. User Scenarios

* **Scenario A (Installing a tool):** A backend developer wants to install `openapi-generator-cli`. They execute `unirtm use maven:org.openapitools/openapi-generator-cli`. The system downloads the `.jar` and provides an `openapi-generator-cli` executable shim.
* **Scenario B (Corporate Network):** A developer is behind a corporate firewall and needs to fetch artifacts from an internal Nexus/Artifactory repository. They set a specific environment variable (e.g., `UNIRTM_MAVEN_MIRROR`), and the system resolves downloads against the corporate mirror instead of Maven Central.
* **Scenario C (Cross-Platform Execution):** A developer installs a `.jar` tool on a Windows machine. The system generates a valid `.cmd` shim. Another developer on macOS installs the same tool, and the system generates a POSIX shell shim. Both can seamlessly run the tool.

## 3. Functional Requirements

1. **Separation of Runtime and Package Management:**
   - The Java runtime must be managed independently from the packages.
   - Introduce `maven:` and `gradle:` prefixes for package installations (both ultimately map to standard Maven repository artifact resolution).
2. **Cross-Platform Compatibility:**
   - Logic code must be compatible with Windows, macOS, and Linux.
   - Unit tests must be designed to pass across all supported operating systems.
   - Shim generation must dynamically adapt to the host OS (POSIX shell scripts for Unix-like systems, `.cmd` batch scripts for Windows).
3. **Environment-Based Mirror Support:**
   - Support environment variables to override the default download registries for both Maven and Gradle.
   - If configured, artifact URLs must correctly map to the user-provided mirror domains.
4. **Smart Shims & JAR Handling:**
   - The provider must dynamically generate wrapper shims that invoke `java -jar <downloaded-artifact.jar>`.
   - The user must be able to run the tool by calling its base name without explicitly typing `java -jar`.

## 4. Success Criteria

* **Measurable:** 100% of the newly added unit tests for the Maven/Gradle provider pass successfully on CI pipelines running Linux, macOS, and Windows.
* **Measurable:** A user can successfully install and execute a known Java CLI tool (e.g., `checkstyle`) in less than 30 seconds (network dependent).
* **Verifiable:** Changing the mirror environment variable causes the system to attempt downloads from the specified URL rather than the default Maven Central URL.
* **Verifiable:** The generated shim on Windows is a `.cmd` file, and on Linux/Mac is a shell script without extension. Both properly pass command-line arguments to the underlying `.jar`.

## 5. Assumptions & Dependencies

* **Assumption:** The user has the `java` runtime installed (either via UniRTM or system-wide) and available in their `PATH` when executing the shim.
* **Assumption:** The artifacts requested by the user are packaged as executable ("fat") `.jar` files. Non-executable jars or artifacts lacking a `Main-Class` manifest entry are out of scope for CLI execution.
* **Dependency:** Maven Central acts as the default upstream source of truth for both `maven` and `gradle` dependencies.
