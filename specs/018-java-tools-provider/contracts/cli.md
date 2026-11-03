# Interface Contracts: CLI

## Provider Input

`unirtm use [maven|gradle]:<groupId>/<artifactId>@<version>`
`unirtm use [maven|gradle]:<groupId>:<artifactId>@<version>`

## Environment Overrides

- `UNIRTM_MAVEN_MIRROR`: String. Overrides standard `https://repo1.maven.org/maven2/`.
- `UNIRTM_GRADLE_MIRROR`: String. Alias for the Maven mirror variable.
