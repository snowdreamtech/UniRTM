# Research: Java Tools Provider

- **Decision:** Use `https://repo1.maven.org/maven2/` as the default base URL.
- **Rationale:** This is the authoritative Maven Central repository layout.
- **Alternatives considered:** Parsing `pom.xml` dependencies (rejected due to complexity; CLI tools are distributed as fat JARs natively).

- **Decision:** Shim execution via `java -jar`.
- **Rationale:** Ensures cross-platform execution without requiring the user to manually type the runtime command.
- **Alternatives considered:** Extracting classes and building a custom classpath. (Rejected: Java CLI tools use manifest Main-Class definitions).
