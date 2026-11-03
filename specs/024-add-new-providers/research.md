# Research: Provider Implementation

## Composer (PHP)

- **Decision**: Set `COMPOSER_HOME` environment variable to `<installPath>`.
- **Rationale**: Composer automatically creates `vendor/bin` under `COMPOSER_HOME`. This achieves perfect isolation without extra flags.
- **Alternatives**: Using `--working-dir`. Rejected because it creates a project rather than a global tool isolated environment.

## LuaRocks (Lua)

- **Decision**: Use `--tree <installPath>` flag.
- **Rationale**: LuaRocks natively supports isolated "trees". The executable will be in `<installPath>/bin`.

## Pub (Dart)

- **Decision**: Set `PUB_CACHE` environment variable to `<installPath>`.
- **Rationale**: Dart pub places global activations into `$PUB_CACHE/bin`.

## Cabal (Haskell)

- **Decision**: Use `--installdir <installPath>/bin` flag.
- **Rationale**: Cabal allows directly specifying where to place the output binary via this flag.
