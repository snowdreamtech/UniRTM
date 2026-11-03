# Tasks: Java Tools Provider (Maven & Gradle)

## Implementation Strategy

- **MVP (User Story 1)**: Basic download and shim generation for standard Maven Central URLs.
- **Increment 1 (User Story 2)**: Add corporate mirror support via environment variables.

## Phase 1: Setup

**Goal**: Initialize the provider structure.
[x] T001 Create `internal/provider/maven.go` stub implementing the `Provider` interface.
[x] T002 Update `internal/provider/registry.go` to register both `maven` and `gradle` prefixes.

## Phase 2: User Story 1 - Maven Artifact Download & Shimming

**Story Goal**: Users can download Java tools and run them natively on Mac, Windows, and Linux via generated shims.
**Independent Test Criteria**: Running `unirtm use maven:com.puppycrawl.tools/checkstyle` successfully downloads the JAR and generates functional wrappers.
[x] T003 [US1] Implement `maven.go` URL resolution logic parsing `maven:group/artifact@version`.
[x] T004 [US1] Implement download logic in `maven.go` utilizing standard HTTP utilities.
[x] T005 [US1] Implement cross-platform shim generation (`.cmd` for Windows, shell for Unix) in `maven.go`.
[x] T006 [P] [US1] Write table-driven unit tests for URL parsing and shim logic in `internal/provider/maven_test.go`.

## Phase 3: User Story 2 - Environment Mirror Support

**Story Goal**: Users behind firewalls can use private Nexus/Artifactory mirrors via environment variables.
**Independent Test Criteria**: Setting `UNIRTM_MAVEN_MIRROR` changes the HTTP GET request domain.
[x] T007 [US2] Update `maven.go` to read `UNIRTM_MAVEN_MIRROR` and `UNIRTM_GRADLE_MIRROR` to override the base URL.
[x] T008 [P] [US2] Add mirror override unit tests to `internal/provider/maven_test.go`.

## Dependencies

- Phase 2 requires Phase 1.
- Phase 3 requires Phase 2 URL resolution logic.

## Parallel Execution Examples

- T006 and T008 can be written concurrently by a pair programmer while T007 is being developed.
