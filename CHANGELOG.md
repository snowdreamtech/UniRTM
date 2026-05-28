<!-- DO NOT EDIT MANUALLY - This file is managed by automated professional tools -->
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0](https://github.com/snowdreamtech/UniRTM/compare/v0.1.0...v0.2.0) (2026-05-28)


### Features

* add container backend and provider ([cda8e1c](https://github.com/snowdreamtech/UniRTM/commit/cda8e1c83a9d69ac20dc3027cc995cdd719bd6cb))
* add support for global proxy settings in config file ([b478b5f](https://github.com/snowdreamtech/UniRTM/commit/b478b5faa3ef0ae845b560959b0179d6fb321994))
* add UNIRTM_ACTIVE env var to track shell activation status ([b5eb733](https://github.com/snowdreamtech/UniRTM/commit/b5eb7331121c8849a4a565f7324df4c0c19fce62))
* **backend:** add gitlab slsa provenance verification and fix backend tests ([20cac21](https://github.com/snowdreamtech/UniRTM/commit/20cac21da379b6b36158593255a28300caed2c26))
* **backend:** implement dependency declaration for tool backends ([8a38a06](https://github.com/snowdreamtech/UniRTM/commit/8a38a06d1a65c159a55c62b84769611117b09bce))
* **backend:** implement smart HTTP/2 downgrade for provenance verification ([a1797bf](https://github.com/snowdreamtech/UniRTM/commit/a1797bff381aef82a1612bcbd9d9df0e6cd43b72))
* **bootstrap:** implement unirtm self-bootstrapping and fix exec arg parsing ([b10c49c](https://github.com/snowdreamtech/UniRTM/commit/b10c49ca19e84c399281684dbc5d74296f7fe308))
* **ci:** integrate npm publishing into GoReleaser workflow ([f6218f5](https://github.com/snowdreamtech/UniRTM/commit/f6218f53f07f51460d73784f5d492dddf3e07309))
* **cli:** add -i/--interactive upgrade checklist to outdated command ([890498a](https://github.com/snowdreamtech/UniRTM/commit/890498a2bb65ee0f87b1b36a0c2afd3e7194d8f2))
* **cli:** add concurrent reshim with dead-shim cleanup and freed disk space reporting in prune ([3114fac](https://github.com/snowdreamtech/UniRTM/commit/3114fac03ef33eebacdf9351f67d85cd5edc67e6))
* **cli:** add progressive loading spinner to search command ([4fae96d](https://github.com/snowdreamtech/UniRTM/commit/4fae96dc5d4f889d63442d3d0a177618dd57d347))
* **cli:** align alias command with mise and add hierarchical UI ([a5a8043](https://github.com/snowdreamtech/UniRTM/commit/a5a80431e0b85a8527f25578d1af26f24c689f64))
* **cli:** align cache command with mise and add stats panel UI ([618396a](https://github.com/snowdreamtech/UniRTM/commit/618396a7f43cdea98d44217c9f3c01d8341a6d27))
* **cli:** align cache command with mise and upgrade UI with pterm ([132727e](https://github.com/snowdreamtech/UniRTM/commit/132727ec445fc3d50c1ebbc5f2ee0302518bca33))
* **cli:** align config command with mise and enhance UI with mirror grouping ([bdeb8be](https://github.com/snowdreamtech/UniRTM/commit/bdeb8be156ba58ffe3e394ca802da57a369d78f7))
* **cli:** align current command with mise and enhance UI ([6674647](https://github.com/snowdreamtech/UniRTM/commit/6674647c39b9d934a52d9637ff7397c75850b32a))
* **cli:** align doctor command with mise and upgrade to dashboard UI ([171ba4b](https://github.com/snowdreamtech/UniRTM/commit/171ba4bb45381b0c523cb95d57ed2f50846686bb))
* **cli:** align where command with mise ([ba089e9](https://github.com/snowdreamtech/UniRTM/commit/ba089e992346283bdae04a85a31311b6b3bc8974))
* **cli:** allow 'unirtm run' to list tasks when invoked without arguments ([af260e4](https://github.com/snowdreamtech/UniRTM/commit/af260e465f8673d61de63e12712d4e5a86663dd0))
* **cli:** display active configuration files in 'config show' ([3196eb1](https://github.com/snowdreamtech/UniRTM/commit/3196eb151b11b17a0a4277c3bc236eff79271112))
* **cli:** enhance activate, deactivate, and bin-paths for full mise parity ([8cdc5ba](https://github.com/snowdreamtech/UniRTM/commit/8cdc5bae2a2c946f146d4a575f8c7f7d5be5e5c0))
* **cli:** enhance alias command with hierarchical view, smart resolution, and mise parity ([cd6c5b2](https://github.com/snowdreamtech/UniRTM/commit/cd6c5b279fcd2cf534155ecfa8326735f3451f5e))
* **cli:** enhance doctor network diagnostics with proxy detection and better error hints ([0882417](https://github.com/snowdreamtech/UniRTM/commit/0882417513ce760ffcd254ff753e9db6a496d632))
* **cli:** enhance generate command with gitlab-ci and dockerfile templates ([a280d0d](https://github.com/snowdreamtech/UniRTM/commit/a280d0d589568b443157f88281dbe89930693c64))
* **cli:** enhance link command with path validation and automatic shim generation ([d647835](https://github.com/snowdreamtech/UniRTM/commit/d647835ecdd3b2732e3b0ccf0406db2dd05fd78c))
* **cli:** enhance list command with disk size column and --current filter ([d140ea7](https://github.com/snowdreamtech/UniRTM/commit/d140ea70a569152cfe3122206be26db532ef74d3))
* **cli:** enhance ls-remote with --limit flag and installed version checkmarks ([bb20f52](https://github.com/snowdreamtech/UniRTM/commit/bb20f52802545e74ea41e10fef14d74b7fb94ba8))
* **cli:** enhance registry, search UI and migrate with visual side-by-side diffs ([fb3d6e9](https://github.com/snowdreamtech/UniRTM/commit/fb3d6e967a3016328db512417b239e90b5936619))
* **cli:** enhance tool and tool-stub commands to parity with mise ([7a9c6cb](https://github.com/snowdreamtech/UniRTM/commit/7a9c6cb37f7e1b0062ba8b5b49299bb686a0f7ab))
* **cli:** enhance where command with smart binary-to-tool resolution ([8ab6316](https://github.com/snowdreamtech/UniRTM/commit/8ab63167f64f5af3848aa346ff48afc00a8e068d))
* **cli:** implement and refine test-tool command to execute and verify installed tools ([f489a90](https://github.com/snowdreamtech/UniRTM/commit/f489a90a1d802520ac59ec771083745a94f99937))
* **cli:** implement cross-platform tool autodetection and symlinking for sync command ([3441195](https://github.com/snowdreamtech/UniRTM/commit/34411958b80f2bd115be065c8f64879becabf434))
* **cli:** implement parallel automatic dependency preparation check and installation ([771495f](https://github.com/snowdreamtech/UniRTM/commit/771495f1e52a4f0758e954137447e90ceefb1621))
* **cli:** list all effective config sources for tool command ([2106b72](https://github.com/snowdreamtech/UniRTM/commit/2106b72a0e5b6c0d16a75cdd46def9f2aacce438))
* **cli:** overhaul fmt command with premium UI and recursive processing ([03e0e4a](https://github.com/snowdreamtech/UniRTM/commit/03e0e4a52ffaa7e1e40a61ba8ecbe8f667ea3307))
* **cli:** upgrade completion command with pterm spinners and improved feedback ([c36af16](https://github.com/snowdreamtech/UniRTM/commit/c36af16f24f03ebcd5d7bf1acad35bacc6338320))
* **cli:** upgrade implode command with premium visuals and deep cleanup ([32691db](https://github.com/snowdreamtech/UniRTM/commit/32691db5703fae4e1a1defac68659c638ae0b514))
* **cmd:** add helpful tip when using system default editor in edit command ([32db92b](https://github.com/snowdreamtech/UniRTM/commit/32db92bfbd16275bc21dc0a1af2375f6c0a3b653))
* **cmd:** add trust status column to trust list output ([dff46e9](https://github.com/snowdreamtech/UniRTM/commit/dff46e9dbd42eedd6bec46004c1629efdf060b48))
* **cmd:** dynamically detect executable path for doctor fix suggestions ([8297ba9](https://github.com/snowdreamtech/UniRTM/commit/8297ba9e43e9f93034ff43d6eecc229c4f382601))
* **cmd:** enable auto-installation of tools in exec command ([45845ea](https://github.com/snowdreamtech/UniRTM/commit/45845eab86d9b68b53348e74c210a516b6e4cd85))
* **cmd:** enhance 'env' command with interactive dashboard and rich visualization ([7abf183](https://github.com/snowdreamtech/UniRTM/commit/7abf183cd6475eeffb5c3ee3656b4e7fd2d634a3))
* **cmd:** make trust auto-target local config and add --list flag ([13fb293](https://github.com/snowdreamtech/UniRTM/commit/13fb2933badb74d92e6c77aabb1de5bac87c9686))
* **cmd:** overhaul 'exec' command with tool@version syntax, zero-overhead process replacement, and unified branding ([8406131](https://github.com/snowdreamtech/UniRTM/commit/8406131894bc1a1351a4eccf82e0d8c70543f1b4))
* **cmd:** populate InjectedPaths in activate command for Env mode ([3141aa0](https://github.com/snowdreamtech/UniRTM/commit/3141aa0bb09404e69f884bb7d666916297d23da7))
* **cmd:** refactor self-update with install source detection, safe download, and post-upgrade verification ([f523063](https://github.com/snowdreamtech/UniRTM/commit/f523063de7eb5e9ab15e47cbf5c99c53bf76a9d5))
* **cmd:** show trusted files table after trusting a config file ([6640aac](https://github.com/snowdreamtech/UniRTM/commit/6640aaccdbefdea2a588413d86080917b10e32fb))
* **cmd:** ultimate upgrade for doctor command with settings audit and smart fix suggestions ([4464f98](https://github.com/snowdreamtech/UniRTM/commit/4464f98a069c9ab26f371ae1e3a5140187030c6b))
* **cmd:** ultimate upgrade for edit command with interactive selection and toml validation ([35b0cfb](https://github.com/snowdreamtech/UniRTM/commit/35b0cfbf059a97b6131dd3d7d86992793192a4cd))
* **cmd:** unify editor discovery with source traceability for edit and doctor commands ([b09edd7](https://github.com/snowdreamtech/UniRTM/commit/b09edd7f51cade7a7ae86d1ab651d2ccc29316ff))
* complete Phase 4 lifecycle and task runner enhancements ([c1399c8](https://github.com/snowdreamtech/UniRTM/commit/c1399c81539d63a3e4eb35e0633593d8d28954ee))
* complete triple-prefix environment variable support and settings mapping ([abb4027](https://github.com/snowdreamtech/UniRTM/commit/abb4027442815a773e80c216319efe9d150e51a6))
* **config:** implement global environment variable synchronization and exhaustive prefix support串 ([81c9695](https://github.com/snowdreamtech/UniRTM/commit/81c96952a42c793c9bdfa7e774ef3960d4cde98d))
* **config:** implement hierarchical configuration loading and merging ([c470bfe](https://github.com/snowdreamtech/UniRTM/commit/c470bfebdc5ccc404a778dc40ae6da1db0d6d0f8))
* **config:** migrate to pongo2 and implement Jinja2 syntax bridging串,Cwd: ([cc009a1](https://github.com/snowdreamtech/UniRTM/commit/cc009a1721e925883c5ef85091704428cea36662))
* **config:** support UNIRTM_ and MISE_ prefixes for all environment variables ([4be7f1b](https://github.com/snowdreamtech/UniRTM/commit/4be7f1bf5f880e23c406f099571d61d5946dd718))
* **config:** update Merge logic to support HttpProxy and HttpsProxy settings ([22fce41](https://github.com/snowdreamtech/UniRTM/commit/22fce419062fe399fc3dddcaec16ce198ada7f69))
* **core:** achieve full parity with mise in tool extraction and binary resolution ([d6f8674](https://github.com/snowdreamtech/UniRTM/commit/d6f86749affa3dd9c614034e89a6883e2a7e1701))
* **dependabot:** enrich directory ignore list with more ecosystems ([62edae9](https://github.com/snowdreamtech/UniRTM/commit/62edae9ce389d242a1f4047de1a6eb42435321dc))
* **dependabot:** migrate directory scan to git ls-files for accurate ecosystem detection ([a847214](https://github.com/snowdreamtech/UniRTM/commit/a847214e55828a7e0ade3c522c7dfb54479b3e38))
* **doctor:** enhance network connectivity diagnostics with proxy routing visibility ([7eea0cf](https://github.com/snowdreamtech/UniRTM/commit/7eea0cff0de95d333c4892949391e6e9ba44a3af))
* **download:** implement multi-threaded turbo download engine ([e09a41f](https://github.com/snowdreamtech/UniRTM/commit/e09a41ff59b78fcd4e5b0c3ce82e8456f81b8e4f))
* **download:** implement smart HTTP/2 to HTTP/1.1 protocol fallback ([c9ff2ce](https://github.com/snowdreamtech/UniRTM/commit/c9ff2cedfe355cbd99c9417d60ee7bf391c01be5))
* **download:** implement smart proxy bypass for domestic mirror domains ([0ef3868](https://github.com/snowdreamtech/UniRTM/commit/0ef386875d058096804ccc2cdcfe46f6d4805d93))
* **download:** implement smart thread scheduling based on file size ([29cbbcc](https://github.com/snowdreamtech/UniRTM/commit/29cbbcc7e19426772bbabcc324f522a80ee332c9))
* **env:** align native tool storage with mise by removing native- prefix ([6b60307](https://github.com/snowdreamtech/UniRTM/commit/6b603077e1252d1ba7f4a24d0f854cc8dd5d84ed))
* **env:** align unset and set with path and env config support ([e89442b](https://github.com/snowdreamtech/UniRTM/commit/e89442bd83bc0f1aa10cb591ee83a1205bd3b1e6))
* **env:** standardize tool installation directory structure to match mise scheme B ([2309438](https://github.com/snowdreamtech/UniRTM/commit/2309438f3ae3acae9c81243a093ce18de7e57446))
* **exec:** align with mise exec and exceed its core capabilities ([6bd495e](https://github.com/snowdreamtech/UniRTM/commit/6bd495ebc580260a1d525f9fba7c46a862094fda))
* **generate:** add dependabot generator in go natively ([c64d6d8](https://github.com/snowdreamtech/UniRTM/commit/c64d6d89da996feaa3db36864d1a597e7079a941))
* **go:** support GOSUMDB, GONOSUMDB and GOPRIVATE variables in go backend and provider ([c31afc9](https://github.com/snowdreamtech/UniRTM/commit/c31afc94379e8d988cd9384bcecba84db5aa32ff))
* implement triple-prefix environment variable overrides for all settings ([176e388](https://github.com/snowdreamtech/UniRTM/commit/176e388278c94aff2e974f92f3be837b5cfcfcc5))
* **index:** implement local tool index seeding, update logic, and status/clear commands ([6098906](https://github.com/snowdreamtech/UniRTM/commit/6098906342bcfa97412cc25e8cc929cb5e34fa3f))
* initial release for version 0.1.0 ([b25a79d](https://github.com/snowdreamtech/UniRTM/commit/b25a79ddc8f9a88396f46eb7653b55de4bd935a4))
* **install:** add cross-platform installation scripts for Linux/macOS/Windows ([6b03e8b](https://github.com/snowdreamtech/UniRTM/commit/6b03e8babf3361d2df9c090a7dbf054a8386e701))
* **install:** add dynamic concurrent multi-spinner UX dashboard for parallel downloads ([b2fb021](https://github.com/snowdreamtech/UniRTM/commit/b2fb02184a24fcff7731def262dffca47594d86f))
* **install:** add throttled download progress logging for large-batch/non-multi installations ([aeb35c7](https://github.com/snowdreamtech/UniRTM/commit/aeb35c7107daf78173c8ef62a795491776365519))
* **installation:** add escape hatch to skip provenance verification via env ([ff611b9](https://github.com/snowdreamtech/UniRTM/commit/ff611b96ecb3fab4e2606d4256e59ca29ff959f6))
* **install:** implement high-performance concurrent parallel installation and smart config fallback ([7ddec1a](https://github.com/snowdreamtech/UniRTM/commit/7ddec1af1f7f2e8216037348ee5fd36aa02be764))
* **install:** show latest versions first in interactive version selection menu ([fe3bc48](https://github.com/snowdreamtech/UniRTM/commit/fe3bc48174288f1453a23e03cf8b6d23d619e73a))
* **install:** sort version select list programmatically descending with semver and alphabetical fallback ([b8f9917](https://github.com/snowdreamtech/UniRTM/commit/b8f99172e216e9d22214a3745782922a86dcbc0c))
* **install:** support multi-tool installation and force flag with premium visual plan dashboard ([b55e9c7](https://github.com/snowdreamtech/UniRTM/commit/b55e9c78564d5d2110c4e28a73ce5515da03a2f2))
* **license:** add --skip flag and -s shorthand for --spdx ([93515c1](https://github.com/snowdreamtech/UniRTM/commit/93515c15d96491eb18947df1313274f93a18fd2e))
* **license:** integrate addlicense as built-in unirtm command ([361d709](https://github.com/snowdreamtech/UniRTM/commit/361d70982d9e65703194fffbcf26dcff77dec058))
* **npm:** add npm platform packaging structure for multi-platform distribution ([ee8615f](https://github.com/snowdreamtech/UniRTM/commit/ee8615f9186d118fbeaef68830b6b369c9a9ad55))
* **provenance:** implement resilient offline verification for GitHub attestations ([1e0e0b0](https://github.com/snowdreamtech/UniRTM/commit/1e0e0b08c706c1e588a775b49546563c6cb0f006))
* **provenance:** implement robust, proxy-aware SLSA verification with custom TUF fetcher串,Cwd: ([e42915d](https://github.com/snowdreamtech/UniRTM/commit/e42915df4803d9ed2e11a2b86d6dd9af924af3e9))
* **provenance:** support dynamic transparency log verification and global release builders ([eab3b4b](https://github.com/snowdreamtech/UniRTM/commit/eab3b4bb24a1a43c6fe77689a0eaab34131a2a60))
* **provider:** add gems.ruby-china.com to no_proxy list ([4fc8aa9](https://github.com/snowdreamtech/UniRTM/commit/4fc8aa9f18e4fa8b63be1b1539ebfa97189bd195))
* **provider:** add maven.aliyun.com and cleanup obsolete mirrors in no_proxy ([29e18fa](https://github.com/snowdreamtech/UniRTM/commit/29e18facb9e700ddcfd196b7d6306c504e9f273f))
* **provider:** add more university mirrors to no_proxy list ([4662e83](https://github.com/snowdreamtech/UniRTM/commit/4662e83f914686bc6dbd772dac19f221961438a6))
* **provider:** add sn0wdr1am.com to no_proxy list ([8d9eaaf](https://github.com/snowdreamtech/UniRTM/commit/8d9eaaf2a7f31343a81ce1986843194e1a7cbd18))
* **provider:** add standard User-Agent and Accept headers for API calls串,Cwd: ([c08b79d](https://github.com/snowdreamtech/UniRTM/commit/c08b79d042e9b04d1b6608298c201d00ab920000))
* **provider:** align storage directory structure with mise ([5555ad1](https://github.com/snowdreamtech/UniRTM/commit/5555ad1273965f19f5c3f59cdcdddaa647be06da))
* **provider:** enhance native providers with mirror and proxy support串,Cwd: ([4eae30d](https://github.com/snowdreamtech/UniRTM/commit/4eae30ddad151a56405fe9d76a182f24f149921a))
* **provider:** expand default no_proxy list with common mirror subdomains ([1e009a3](https://github.com/snowdreamtech/UniRTM/commit/1e009a362bbcbe70be4b51cdc09b8c34cc0057e2))
* **provider:** implement exponential backoff retry for GitHub API串,Cwd: ([cc6bb7c](https://github.com/snowdreamtech/UniRTM/commit/cc6bb7c377fb2221c39fba5ecb85dd32743837b2))
* **provider:** implement intelligent proxy bypass with config and dynamic mirror detection ([6e803b3](https://github.com/snowdreamtech/UniRTM/commit/6e803b3cf14f39eb414f59aff43a36025a18bab5))
* **provider:** implement quiet progress mode for all tool providers ([0c2737f](https://github.com/snowdreamtech/UniRTM/commit/0c2737fc01801d99b5d8bb864acef0924e5673ca))
* **provider:** implement zero-dependency archive extraction with smart format sniffing ([eebcea9](https://github.com/snowdreamtech/UniRTM/commit/eebcea9268531646bc87934adb94e793dcee24f3))
* **provider:** remove deprecated taobao mirrors from no_proxy list ([75bf9c3](https://github.com/snowdreamtech/UniRTM/commit/75bf9c34972a3c05c0fbf9f012cab248f7eeb63b))
* **provider:** restore taobao mirrors for legacy support ([eddff10](https://github.com/snowdreamtech/UniRTM/commit/eddff10b29663fb564a85f71fed4d5231645afc1))
* **provider:** standardize proxy handling and fix pipx installation failures ([53f8a74](https://github.com/snowdreamtech/UniRTM/commit/53f8a74246f79b22012a39383b530dd22fa46626))
* **provider:** support .zst/.xz extraction and improve which smart matching ([28f6732](https://github.com/snowdreamtech/UniRTM/commit/28f6732c152ad36f8b6ed4f31d423aa1d5c69e36))
* **provider:** switch to root domains in no_proxy to cover all mirror subdomains (fixes cdn.npmmirror.com error) ([a37a361](https://github.com/snowdreamtech/UniRTM/commit/a37a36135dc75fb3c0b5af419bbe5c9b570425f0))
* **security:** implement minimum_release_age supply chain protection ([d267a07](https://github.com/snowdreamtech/UniRTM/commit/d267a0740aa6f7a34112afe62b64554c7022cb1a))
* **security:** set minimum_release_age default to 7d ([c0cdcf5](https://github.com/snowdreamtech/UniRTM/commit/c0cdcf5f99bfbbf221405580d7808f557a4f56a0))
* **service:** centralize tool specification parsing and backend detection ([93d61a1](https://github.com/snowdreamtech/UniRTM/commit/93d61a1f59643868306d9df829ce0bceead2b20e))
* **service:** implement dependency-aware installation and automatic path migration ([c0f2357](https://github.com/snowdreamtech/UniRTM/commit/c0f235730b3c5718bcdc541dba967d389ed434fc))
* **service:** implement idempotent and cross-platform PATH activation logic ([e98fe6a](https://github.com/snowdreamtech/UniRTM/commit/e98fe6a3a16b7cc603107570a71ce8c6b8bc2dc8))
* **service:** support skipping checksum and restore env logic串,Cwd: ([52929bd](https://github.com/snowdreamtech/UniRTM/commit/52929bdeb4b822028fdc7edacaf629dbb7a87d2e))
* **shim:** add internal tool dispatcher to unirtm binary (Stage 1) ([1bfeab2](https://github.com/snowdreamtech/UniRTM/commit/1bfeab2892603c2162f381bc2f2ee74eeea0b01e))
* **shim:** refactor Unix shim generation to use symlinks (Stage 2) ([6b8374c](https://github.com/snowdreamtech/UniRTM/commit/6b8374c25e40a772073e180393b9cf5943a0a98e))
* **shim:** simplify backend loop by removing redundant native prefix check ([5e959b6](https://github.com/snowdreamtech/UniRTM/commit/5e959b6507e333c875b1a6715d30fc4e84c60899))
* support UNIRTM_ prefix for proxy environment variables ([cc7507b](https://github.com/snowdreamtech/UniRTM/commit/cc7507b43abf8947ddc01d5331e5c985c36bc401))
* **task:** add flexible output controls (flag, env, CI-detect) and show spinner output on success/failure ([3808c63](https://github.com/snowdreamtech/UniRTM/commit/3808c63ea4db4cb7921041e5dd91df3c0fd79aa3))
* **tasks:** add --fix flag to unirtm run ([7544bec](https://github.com/snowdreamtech/UniRTM/commit/7544bec82ec2754393a819b9a686c9c172752c9a))
* **task:** support cross-platform Windows shell delegation in native runner ([2145981](https://github.com/snowdreamtech/UniRTM/commit/214598144e9158168db785527288353417fbba98))
* **trust:** enhance trust/untrust commands with detailed listings and sha-256 validation ([39db719](https://github.com/snowdreamtech/UniRTM/commit/39db71924fb33d5c83b0acb7c2920590230d23ab))
* **ui:** implement Style C (unified concise feedback) for installation workflow串,Cwd: ([135c576](https://github.com/snowdreamtech/UniRTM/commit/135c5769092169142ff855f0fa11b15775817059))
* **uninstall:** support multi-tool package specs uninstallation with clean and professional terminal logs ([6fac68a](https://github.com/snowdreamtech/UniRTM/commit/6fac68abddae2b9f9864665a52c73a73e7bcd184))
* **uninstall:** support optional version and @-syntax with database resolution ([a53bd4d](https://github.com/snowdreamtech/UniRTM/commit/a53bd4d314ac9c73b3944e29cb89250146dba9dd))
* **unuse:** add global, path, and env flags to unuse command ([9f54234](https://github.com/snowdreamtech/UniRTM/commit/9f54234f6d2031fe81a7affa7b1a5b5ed50313e1))
* **update:** enhance update command with interactive loaders and high fidelity visual feedback ([0ed1fcc](https://github.com/snowdreamtech/UniRTM/commit/0ed1fcc05aca11e6be068ac2d765bd1824d37b7e))
* **use:** add --pin support for concrete version resolution ([05753d5](https://github.com/snowdreamtech/UniRTM/commit/05753d5c5ca93aa02ce1b98c481649c4e6178e25))
* **use:** align use command flags and features with mise, adding --force and --env support with comprehensive tests ([814a6d0](https://github.com/snowdreamtech/UniRTM/commit/814a6d040e2b305eef1c7beb7d23cb4d5a1e8520))
* **use:** automatically download and install tool version if missing during use ([84512f3](https://github.com/snowdreamtech/UniRTM/commit/84512f33585fe69d9055bc71cc6d4b62e91fc57f))
* **watch:** enhance watch command with process tree hot-reloading and UI ([ae7142f](https://github.com/snowdreamtech/UniRTM/commit/ae7142fd831bb8a77b97cc5090f82c540805891c))


### Bug Fixes

* **backend:** add go-pkg backend for go package installation and include pip-audit/govulncheck ([3023fa7](https://github.com/snowdreamtech/UniRTM/commit/3023fa79728bb500ddfb2c876af8bd4348b4e435))
* **backend:** disable HTTP/2 globally during provenance verification if HTTP2=0 ([6aa8182](https://github.com/snowdreamtech/UniRTM/commit/6aa81824c446c1abd34c01a19307b3797cc51fc3))
* **backend:** isolate http transport to prevent global DefaultTransport side-effects ([9d724f9](https://github.com/snowdreamtech/UniRTM/commit/9d724f94a07d04c6ac3882b881a9cfb78db601c6))
* **backend:** properly propagate checksum fetching errors ([093b2bc](https://github.com/snowdreamtech/UniRTM/commit/093b2bc9fce7305c5dd363015a85c7c313efe093))
* **ci:** add sparse checkout bootstrap before local harden-runner actions ([4fe39a9](https://github.com/snowdreamtech/UniRTM/commit/4fe39a9465c1445cb17f0ee1f39114ca80b9de41))
* **ci:** explicitly specify npm official registry for all publish commands ([c99f2ac](https://github.com/snowdreamtech/UniRTM/commit/c99f2acc5f309a29f7eac0f3dc9eb30a168f6740))
* **ci:** inline step-security/harden-runner in all 14 workflows ([450d136](https://github.com/snowdreamtech/UniRTM/commit/450d136f5d35bb402c7005bcb2644960881e6b9d))
* **ci:** move npm build step from goreleaser after_hooks to GitHub Actions ([ffcb2a9](https://github.com/snowdreamtech/UniRTM/commit/ffcb2a919a651cc4dfe205c1345f8dbd77ec5ed5))
* **ci:** remove emoji from sh run blocks to prevent UTF-8 parsing errors ([56a0e23](https://github.com/snowdreamtech/UniRTM/commit/56a0e23bb7ecedfa79ad5ea2326669676e7410dc))
* **ci:** remove harden-runner composite action wrappers ([9ae20a3](https://github.com/snowdreamtech/UniRTM/commit/9ae20a379964f1388dfb526de80efad71590fbb9))
* **ci:** use dist/metadata.json version instead of raw git tag ([b851ff5](https://github.com/snowdreamtech/UniRTM/commit/b851ff553288fd304577d6706b7699b1721dc72c))
* **ci:** use node instead of jq to parse dist/metadata.json ([b889491](https://github.com/snowdreamtech/UniRTM/commit/b8894914437fd36a58299cb3bc4b8aa99d2dc722))
* **ci:** use single quotes for echo to prevent history expansion in zsh ([cbc9e55](https://github.com/snowdreamtech/UniRTM/commit/cbc9e55d4a336dbdd6efe25ba172f48b4118e2d2))
* **cli:** correctly detect active versions for tool command by matching config request with installed versions instead of parsing symlink ([dde4db5](https://github.com/snowdreamtech/UniRTM/commit/dde4db5cf49ad4666f29893bf98750c5cd8958eb))
* **cli:** do not suggest alternatives if command fails with non-zero exit code ([ae936d6](https://github.com/snowdreamtech/UniRTM/commit/ae936d60a4ecaaedf1db953430b55a98eaf0b4c9))
* **cli:** ensure implode command deletes entire data directory ([0e57510](https://github.com/snowdreamtech/UniRTM/commit/0e57510c4fd3d6af3981bab7935b959fb2a49c71))
* **cli:** handle 'v' prefix tolerance in current command version detection ([05900ec](https://github.com/snowdreamtech/UniRTM/commit/05900ece015152178d1d7e05e1c01c6adab5a72b))
* **cli:** implement robust case-insensitive 'v/V' prefix handling for version detection ([538ed5a](https://github.com/snowdreamtech/UniRTM/commit/538ed5ac95112a16a6462a96bf262328ea5849f5))
* **cli:** implement robust normalized matching for date-based (CalVer) versions ([23096c4](https://github.com/snowdreamtech/UniRTM/commit/23096c46e23163339726649577680faf9fb6099e))
* **cli:** truncate long environment variables in doctor command to preserve layout ([9d9d4b6](https://github.com/snowdreamtech/UniRTM/commit/9d9d4b62c3e4843103f9d10711c753e6c129731c))
* **cmd:** dynamically resolve config file name in tasks message ([8927d3f](https://github.com/snowdreamtech/UniRTM/commit/8927d3f37481941b5e5e6f4a6fd267aea95e6205))
* **cmd:** fix tool installation path detection by normalizing tool names into slugs ([dd04623](https://github.com/snowdreamtech/UniRTM/commit/dd04623a669348f758865f84a026f38873839a16))
* **cmd:** fix tool version normalization in doctor command to handle 'v' prefixes ([7117f12](https://github.com/snowdreamtech/UniRTM/commit/7117f126c0516f8e97b475454151dff55a31a145))
* **cmd:** prioritize exact matches in which command and restrict tool fallback ([8937275](https://github.com/snowdreamtech/UniRTM/commit/8937275a4a14a96998fddd1213742a2ea4b2c4b6))
* **cmd:** remove unused term import in exec command ([ba2f6aa](https://github.com/snowdreamtech/UniRTM/commit/ba2f6aafa6ccdd8d62a6767467f8f44664f8ddb4))
* **cmd:** restore missing cache calculation logic in cache command ([880c1a5](https://github.com/snowdreamtech/UniRTM/commit/880c1a51c059b90994a67dba879ea41824e5ece7))
* **cmd:** strengthen detectInstallMethod with nix/snap/asdf/macports/pkgx support and safer Go path detection ([94c8af7](https://github.com/snowdreamtech/UniRTM/commit/94c8af72b2f4de63c2746450d938651a59f471a9))
* **cmd:** warn user when tasks are ignored due to modified config ([9c442b1](https://github.com/snowdreamtech/UniRTM/commit/9c442b1736d45790583ebe950695af397c71f0e8))
* **common.sh:** make versions.sh and bootstrap.sh sources optional ([41a78fc](https://github.com/snowdreamtech/UniRTM/commit/41a78fc27fd62d905fc1fdbc90f04ff02759451e))
* **config:** filter out non-existent files from trusted list ([1918b00](https://github.com/snowdreamtech/UniRTM/commit/1918b0047d9af9648a260b405a44b2d8459afb04))
* **config:** implement DurationOrInt custom unmarshaling and resolve tests ([6b8496d](https://github.com/snowdreamtech/UniRTM/commit/6b8496df90b2c176e0cff544ba96b241921bc57b))
* **core:** remove duplicate icons and newlines in CLI output ([705eba1](https://github.com/snowdreamtech/UniRTM/commit/705eba184f2aa3612019a153c3fb1b8e11f51bab))
* **database:** resolve SQLITE_BUSY errors during highly concurrent installations ([8a22331](https://github.com/snowdreamtech/UniRTM/commit/8a223316e32ab183f574b3be42d9d16f793d7eb6))
* **dependabot:** expand WalkDir skip list to prevent spurious ecosystem entries ([8509137](https://github.com/snowdreamtech/UniRTM/commit/850913778da9b30a1c88d53488fc62bce0701745))
* **dependabot:** skip .devcontainer dir in WalkDir to prevent spurious docker entry ([4b627ac](https://github.com/snowdreamtech/UniRTM/commit/4b627acf587fb5991e8f5ab57fd052a4f9e1a1ff))
* **download:** enforce strict HTTP/1.1 ALPN negotiation to bypass CDN HTTP/2 bugs ([4019b5c](https://github.com/snowdreamtech/UniRTM/commit/4019b5c335cce0648be0617b642a7e47b6a27fe1))
* **download:** enhance industrial-grade robustness for extreme network conditions ([e093dbe](https://github.com/snowdreamtech/UniRTM/commit/e093dbe527a5578eb2d56eef965fb4677ab6a13f))
* **download:** fix defer in loop causing persistent protocol mismatch ([2b8d2b7](https://github.com/snowdreamtech/UniRTM/commit/2b8d2b700d2faa62f5d8d98b0fc1804360e6d7ca))
* **download:** fix progress freeze and timeout for large files ([1767768](https://github.com/snowdreamtech/UniRTM/commit/176776875de5656ba8f0923b03625becfff1a9cb))
* **download:** fortify concurrent download engine with industrial-grade safeguards ([f36b23a](https://github.com/snowdreamtech/UniRTM/commit/f36b23a78010f71607fbf3b69c0b2a189a3cf121))
* **download:** implement Equal Jitter exponential backoff for concurrent threads ([a8c141b](https://github.com/snowdreamtech/UniRTM/commit/a8c141b4d2d7ea3b29c5ddd98075976828d389d8))
* **download:** implement robust physical protocol isolation for fallback ([727689b](https://github.com/snowdreamtech/UniRTM/commit/727689b09213c284c0e63920426a7994282bca12))
* **download:** prevent connection pool pointer leak during concurrent installations ([260e12c](https://github.com/snowdreamtech/UniRTM/commit/260e12c1520a34bced96a291bcb369519ef3295d))
* **download:** prevent redundant fallback on context cancellation ([a91d378](https://github.com/snowdreamtech/UniRTM/commit/a91d378253f0e112a8f61e9eb1be5eff29773838))
* **download:** strictly enforce chunk boundaries and fix premature EOF bug ([6d3fe10](https://github.com/snowdreamtech/UniRTM/commit/6d3fe1042ea7792846fc6e36f31619b9ff66fd93))
* **env:** align macOS cache path with mise (standard XDG/macOS) ([0f55d4c](https://github.com/snowdreamtech/UniRTM/commit/0f55d4cb1b90ab13a8b942c4663feae989134d17))
* **env:** bypass prefix lookup for PATH in env.Get to prevent task execution failures ([28470f9](https://github.com/snowdreamtech/UniRTM/commit/28470f95d77bf9e53a89ab08168aa90312bc2574))
* **env:** inject provider specific environment variables into active shell ([1aa8172](https://github.com/snowdreamtech/UniRTM/commit/1aa8172c0c3c709d9f0c9c4c714aaa69c8a8d5f4))
* **exec:** correctly parse tool spec to resolve env vars ([3cd87ba](https://github.com/snowdreamtech/UniRTM/commit/3cd87ba2f1a6ad1efae5e6afe6ff146564025526))
* **exec:** reject archive/non-executable files in ResolveExecutable ([61e036d](https://github.com/snowdreamtech/UniRTM/commit/61e036d9d77e3a94be5a7fd8a42d8c90d18fea73))
* **fmt:** eliminate DATA RACE in TestFmtCmd_* under go test -race ([4c77d80](https://github.com/snowdreamtech/UniRTM/commit/4c77d8074f5acac2b6a543e61e1fdca2540b931f))
* **http:** add ALL_PROXY fallback support with proper NO_PROXY enforcement ([0ae743c](https://github.com/snowdreamtech/UniRTM/commit/0ae743c12691df8e3a6face38ea73e39ce2cc703))
* **http:** honor NO_PROXY in custom proxy resolver via httpproxy.Config ([537ff93](https://github.com/snowdreamtech/UniRTM/commit/537ff933190164d0b28530e72481cfb45554ce2e))
* **install:** cap concurrent multispinner UI to avoid terminal scrolling glitches ([f9fa73a](https://github.com/snowdreamtech/UniRTM/commit/f9fa73a08384218d7d92a8ac2855c2a93d5b0429))
* **install:** implement custom thread-safe spinner manager to eliminate progress duplicates ([bd3619a](https://github.com/snowdreamtech/UniRTM/commit/bd3619a339e08a716e71f701f72ad1bdb8e044ff))
* **install:** prevent progress bar race condition in concurrent downloads ([d039334](https://github.com/snowdreamtech/UniRTM/commit/d03933440b06482cf9f5d96feaec2208a8e07cb8))
* **install:** suppress extraction flattening and hook execution logs during concurrent installation ([e2b4944](https://github.com/snowdreamtech/UniRTM/commit/e2b494426ae546fb8928ca4481ffc830b03f9bfc))
* **install:** suppress individual interactive progress UI during concurrent installation ([ade6f0b](https://github.com/snowdreamtech/UniRTM/commit/ade6f0b39a31ec2e1afc9eeeebe4e80d2b60a6c7))
* **install:** treat already installed status as success in concurrent manager to avoid skipping dependents ([561afd0](https://github.com/snowdreamtech/UniRTM/commit/561afd015dafc2570d8a7422af0a68566df3038a))
* **install:** use full tool key for lockfile validation ([46e7f6f](https://github.com/snowdreamtech/UniRTM/commit/46e7f6f62c070c8fd09354ebe510b4ba2d2e30bd))
* **license:** resolve flag shorthand conflicts with root persistent flags ([4c77d80](https://github.com/snowdreamtech/UniRTM/commit/4c77d8074f5acac2b6a543e61e1fdca2540b931f))
* **license:** use -H as shorthand for --holder flag ([92c0b59](https://github.com/snowdreamtech/UniRTM/commit/92c0b598c9c337b8954e8af4e892da712da44b9e))
* **license:** use -Y as shorthand for --year flag ([8ec0916](https://github.com/snowdreamtech/UniRTM/commit/8ec0916242547045dc466a11d7b55da46308daf1))
* **lint:** resolve pre-commit missing executable errors and checkmake issues ([ef2818a](https://github.com/snowdreamtech/UniRTM/commit/ef2818afcd4fcfa4c32f5c2a047e5525398f38dd))
* **lint:** resolve pre-commit shellcheck and markdownlint issues ([e403b43](https://github.com/snowdreamtech/UniRTM/commit/e403b43f97c0bd5b3f3a49b5dfad0b0ea49cfc17))
* **lint:** resolve yamllint setup version and fix shellcheck recursion issues ([2fed416](https://github.com/snowdreamtech/UniRTM/commit/2fed416030f4712a7de34803bd5c37ad5487ba46))
* **local-verification:** remove || true from security audit tasks so vulnerabilities fail the build locally ([a755027](https://github.com/snowdreamtech/UniRTM/commit/a755027bea0bbcc1ae76c26913ce3e5380a73c5a))
* **local-verification:** remove remaining || true to enforce strict audit failure ([c4ca553](https://github.com/snowdreamtech/UniRTM/commit/c4ca5538aec8f6df2acf9fd6e948fdb5ef718e80))
* **lockfile:** CheckStrict should not require URL for package manager backends ([da3c05a](https://github.com/snowdreamtech/UniRTM/commit/da3c05a2951d5ad1be6759aa5ee0d2e641309f94))
* **node:** resolve 404 download error and enhance template compatibility ([6cbce86](https://github.com/snowdreamtech/UniRTM/commit/6cbce8602be8975bcde9ea9d6048f8be141dc624))
* **npm:** correct repository.url format to git+https in all package.json.tpl files ([825c08c](https://github.com/snowdreamtech/UniRTM/commit/825c08c1ae69fdb24d9df1478757293151b5a02d))
* **npm:** export and concatenate NODE_PATH for npm plugins ([eb44f9a](https://github.com/snowdreamtech/UniRTM/commit/eb44f9a6cf137e3f98cd255ce9cb22f6a3613fd0))
* **npm:** remove bin field from platform-specific sub-packages to resolve global install collision ([2065068](https://github.com/snowdreamtech/UniRTM/commit/2065068e685377859976f3f992edad2d0c94c9d6))
* prevent silent installation of corrupted archives ([50ac2cd](https://github.com/snowdreamtech/UniRTM/commit/50ac2cd277bf6b58bb3961f680c4b0bf53831c55))
* **provenance:** remove broken GITHUB_PROXY hijacking of standard proxy env vars ([c0fc36b](https://github.com/snowdreamtech/UniRTM/commit/c0fc36b834731c869dfc2086644ca0828206d578))
* **provider/generic:** relativize symlinks and prevent overwriting binaries ([8584b96](https://github.com/snowdreamtech/UniRTM/commit/8584b96472645766317ff9e39cb76315f339946c))
* **provider:** auto-rename primary executable to standard tool name ([440c100](https://github.com/snowdreamtech/UniRTM/commit/440c100965a665364c827fc536f7c03ef1ae3616))
* **provider:** fix tool name resolution and cleanup duplicate code ([cf00c61](https://github.com/snowdreamtech/UniRTM/commit/cf00c614238e6de901b11753c3daa8757f8ec9f7))
* **provider:** implement virtualenv path rewriting in pypi PostInstall to resolve relocatability issues ([eb2bfe9](https://github.com/snowdreamtech/UniRTM/commit/eb2bfe9ab841f02323bafed2a7c6ab8407e1e4f0))
* **provider:** prefer UniRTM-managed conda installations for conda packages ([055ddd4](https://github.com/snowdreamtech/UniRTM/commit/055ddd41f73074673f58c002a04fef230246790f))
* **provider:** prefer UniRTM-managed dotnet installations for dotnet tools ([e3c07cd](https://github.com/snowdreamtech/UniRTM/commit/e3c07cdef7432ab703a25ba2d1e22c3dabb944ce))
* **provider:** prefer UniRTM-managed go installations for go packages ([7b74b04](https://github.com/snowdreamtech/UniRTM/commit/7b74b047af292a006e04246099031e059c30365b))
* **provider:** prefer UniRTM-managed node/npm installations for npm tools ([3a054d0](https://github.com/snowdreamtech/UniRTM/commit/3a054d02b79f6756e5d786880c19758cfec70626))
* **provider:** prefer UniRTM-managed python installations for PyPI tools ([5c7f163](https://github.com/snowdreamtech/UniRTM/commit/5c7f1635be377f4e312a7444f30fd9c04368406f))
* **provider:** prefer UniRTM-managed ruby/gem installations for ruby gems ([b946e75](https://github.com/snowdreamtech/UniRTM/commit/b946e75e569945c97be5d1ae22df722d2ae9297c))
* **provider:** prefer UniRTM-managed rust/cargo installations for cargo crates ([2cd79a8](https://github.com/snowdreamtech/UniRTM/commit/2cd79a882b4f6b0344f852c601ba0bdc809790d3))
* **provider:** prefer UniRTM-managed swift installations for SPM packages ([fdc79cf](https://github.com/snowdreamtech/UniRTM/commit/fdc79cffed51e87d649dfc3c85c54acadc71fdda))
* **provider:** remove official rust domain from no_proxy defaults ([984efaf](https://github.com/snowdreamtech/UniRTM/commit/984efaf30577d1c1f239f45e9967ba7c8be890e3))
* **provider:** remove unused env import after refactoring ([5ed3c61](https://github.com/snowdreamtech/UniRTM/commit/5ed3c61fc85ad44de5e26ca6fd2ecd717a47c739))
* **provider:** replace hardcoded extension list with heuristic ([74e7070](https://github.com/snowdreamtech/UniRTM/commit/74e7070426c97eec46d0f73d3fa52e3fbad2e93a))
* **python:** resolve 504 error and version mapping issues ([94e6851](https://github.com/snowdreamtech/UniRTM/commit/94e6851e28b0a92ea36c21641b3cb600d13895a3))
* resolve activation PATH issues and improve which command ([9e7fb50](https://github.com/snowdreamtech/UniRTM/commit/9e7fb50717d5ed7f5532cbdc57581fad90a43983))
* resolve checkmake duplicate log issue by improving version normalization in installation manager ([9b5503f](https://github.com/snowdreamtech/UniRTM/commit/9b5503f5adb971ce33203f06e1fa85d17a3b34b5))
* resolve deadlocks and timeouts in backend operations ([091a737](https://github.com/snowdreamtech/UniRTM/commit/091a737654e9cc79ec7454a5cb7b773138eab9bf))
* resolve editorconfig-checker false positives for Go and testdata ([7f7797d](https://github.com/snowdreamtech/UniRTM/commit/7f7797dffe96a9519a4fb8c6b27b97990a299e59))
* resolve trivy gomod vulnerabilities and add trivy to unirtm tools ([2d0d4d0](https://github.com/snowdreamtech/UniRTM/commit/2d0d4d06b7e9c87852496fe0762ab78340504d76))
* **service:** clean up all sub-commands and helper binary shims during uninstallation ([00e84ef](https://github.com/snowdreamtech/UniRTM/commit/00e84ef7ed7b61674701f935e1d92e63b30bbe4e))
* **service:** clean up empty parent directories and dead shim files during uninstallation ([1c3835a](https://github.com/snowdreamtech/UniRTM/commit/1c3835afc146580d7333f29b5f9bb6320fe03576))
* **service:** ignore hidden and OS system files like .DS_Store during parent directory cleanup ([c4b010b](https://github.com/snowdreamtech/UniRTM/commit/c4b010be630d7d5b252e3735754f586986b26132))
* **service:** implement progress update throttling to prevent terminal freezing ([0cac080](https://github.com/snowdreamtech/UniRTM/commit/0cac080e82d70a5f0b1f81beb5f430354b8ab696))
* **service:** refine shell config injection/removal patterns and enhance PowerShell compatibility ([1d27ca4](https://github.com/snowdreamtech/UniRTM/commit/1d27ca41be66d5e804437aa2241231a7ca5d81be))
* **service:** resolve 'already exists' error during re-installation by implementing Upsert ([d333422](https://github.com/snowdreamtech/UniRTM/commit/d333422c20ffdfda5e5b58f2716ec79190a6146b))
* **service:** unify installation detection with robust version variant support (v/V prefix) ([fe5734b](https://github.com/snowdreamtech/UniRTM/commit/fe5734b509faed50c69c0edf8f94a07e007bbaf1))
* **service:** use individual sed commands to ensure cross-platform compatibility and avoid regex grouping errors ([eb4e9c6](https://github.com/snowdreamtech/UniRTM/commit/eb4e9c6aa6ccd5a0fc67114a31071cb8482cc077))
* **task:** capture and display stdout/stderr on task failure ([2c34875](https://github.com/snowdreamtech/UniRTM/commit/2c348759911ae42c657b05311adf6c3114b800b3))
* **tasks:** remove os.Expand from native runner ([92a89a5](https://github.com/snowdreamtech/UniRTM/commit/92a89a5736fc7f58268aa37d8a74d1a1e930772e))
* **ui:** ensure final progress bar frame renders completely before stopping ([c965f4e](https://github.com/snowdreamtech/UniRTM/commit/c965f4e4a322fb5d000e594600e54895d5fbaa2e))
* **ui:** guarantee deterministic rendering of final 100% progress frame ([7ea8778](https://github.com/snowdreamtech/UniRTM/commit/7ea8778a26bc775f2c6d71433018bc2036b9bc19))
* **ui:** prevent progress bar from freezing at 100% when download resets ([90d5994](https://github.com/snowdreamtech/UniRTM/commit/90d5994e4fbb4d9dd6ca76e1840f4eb4604757d6))
* **uninstall:** resolve CLI error double printing by returning clean formatted errors to central cmd wrapper ([3e6cfd0](https://github.com/snowdreamtech/UniRTM/commit/3e6cfd0a7b07987bbe2e81545f9643df85eed339))
* **update:** allow --preview to run without specifying a tool or --all ([011af7a](https://github.com/snowdreamtech/UniRTM/commit/011af7a5c54e56a27dfa78eb4be8f9ca62693115))
* **which:** auto-delete stale DB records when install path missing on disk ([8d347a9](https://github.com/snowdreamtech/UniRTM/commit/8d347a9b4635de0813be8ab35c2c673b76ed43b0))
* **which:** resolve tools by executable name when tool-name lookup fails ([6aabb72](https://github.com/snowdreamtech/UniRTM/commit/6aabb72f44a446a53f7426bfcdc1ab82055b73ea))
* windows npm provider and node shims paths ([1eab3a8](https://github.com/snowdreamtech/UniRTM/commit/1eab3a8f5b4db0bc6bf7a01eba50eeed1c25fa10))


### Performance Improvements

* **download:** increase network timeouts for better proxy compatibility ([e792bdc](https://github.com/snowdreamtech/UniRTM/commit/e792bdced5bea19faad01190e22a0007d18bd961))
* **http:** resolve proxy config once at transport init, not per request ([bdc3839](https://github.com/snowdreamtech/UniRTM/commit/bdc38397fc41ae7f81bc3f20e03e18755b8eeff0))
* **service:** optimize installation flow with local-first check and enhance GitHub API resilience串,Cwd: ([9ddfcef](https://github.com/snowdreamtech/UniRTM/commit/9ddfcef5e33502b181b19119bffdc9fae6c680c3))

## [0.1.0](https://github.com/snowdreamtech/UniRTM/compare/v0.0.10...v0.1.0) (2026-05-25)


### Features

* initial release for version 0.1.0 ([b25a79d](https://github.com/snowdreamtech/UniRTM/commit/b25a79ddc8f9a88396f46eb7653b55de4bd935a4))

## [0.13.1](https://github.com/snowdreamtech/template/compare/v0.13.0...v0.13.1) (2026-04-29)


### Bug Fixes

* **ci:** disable shellcheck integration in actionlint to prevent Windows hang ([1d711d8](https://github.com/snowdreamtech/template/commit/1d711d8c1bc4c7482651f1abe228c970ba6f024c))

## [0.13.0](https://github.com/snowdreamtech/template/compare/v0.12.0...v0.13.0) (2026-04-28)


### Features

* **ci:** enable auto-merge for ALL dependabot updates including major ([9c8bae2](https://github.com/snowdreamtech/template/commit/9c8bae233e0e8834a611370dc897ac67db64057a))
* **deps:** consolidate all dependabot updates into single PR ([e753902](https://github.com/snowdreamtech/template/commit/e753902201ce2f059747de31b59681442bb149b0))


### Bug Fixes

* **deps:** correct devcontainers directory to root ([9d83806](https://github.com/snowdreamtech/template/commit/9d838067e3fb82e84c1107f0036cc4a7b7855f56))
* **deps:** exclude .devcontainer from docker ecosystem detection ([67e96c3](https://github.com/snowdreamtech/template/commit/67e96c3c1b94c72811d81af229a017f927fd3de4))

## [0.12.0](https://github.com/snowdreamtech/template/compare/v0.11.0...v0.12.0) (2026-04-27)


### Features

* **security:** add CDN and download mirror endpoints to harden runner ([3ff8a5d](https://github.com/snowdreamtech/template/commit/3ff8a5d3d073a57bb4681c619e94aa797983ef23))
* **security:** add DockerHub endpoints to harden runner ([48b6416](https://github.com/snowdreamtech/template/commit/48b6416ced65a68194a50d4ffb406c7e7257a37f))

## [0.11.0](https://github.com/snowdreamtech/template/compare/v0.10.0...v0.11.0) (2026-04-24)


### Features

* **ci:** add GitHub API rate limit info to summary reports ([8c4293a](https://github.com/snowdreamtech/template/commit/8c4293afbd82f74278af2db4ecb77abf6a9b1aab))
* **deps:** optimize dependabot configuration and grouping strategy ([900f3af](https://github.com/snowdreamtech/template/commit/900f3aff782a709643b81f9721a45da7a0d10a87))


### Bug Fixes

* **ci:** replace head -n -1 with sed for macOS compatibility ([67aaa95](https://github.com/snowdreamtech/template/commit/67aaa95afe68cdfcade622a3478fe3b0b000ccca))
* **deps:** correct indentation in test script for editorconfig compliance ([0a44088](https://github.com/snowdreamtech/template/commit/0a44088beafca6ea70111f98062e1889d21c069f))

## [0.10.0](https://github.com/snowdreamtech/template/compare/v0.9.2...v0.10.0) (2026-04-22)


### Features

* add rocket icon to performance workflow name ([e6a6146](https://github.com/snowdreamtech/template/commit/e6a61461a6315e84e993d6edfc618ef61f27003e))

## [0.9.2](https://github.com/snowdreamtech/template/compare/v0.9.1...v0.9.2) (2026-04-20)


### Bug Fixes

* **lint:** ensure pre-commit runs exactly 2 times, not 3 ([4085676](https://github.com/snowdreamtech/template/commit/40856766a964a95216a171fd5760bde2b91240c5))
* **release:** set separate-pull-requests to false for unified PR ([07054f5](https://github.com/snowdreamtech/template/commit/07054f5d69e832f976b46e1648d39779042cfb15))

## [0.9.1](https://github.com/snowdreamtech/template/compare/v0.9.0...v0.9.1) (2026-04-20)


### Bug Fixes

* **lint:** ignore CHANGELOG.md in all directories for markdownlint ([ebe338c](https://github.com/snowdreamtech/template/commit/ebe338c66126b82624c63e45b5c8a7c6bf77fcc1))

## [0.9.0](https://github.com/snowdreamtech/template/compare/v0.8.2...v0.9.0) (2026-04-19)


### Features

* **lint:** add two-pass auto-fix mechanism to make lint ([9fba0f1](https://github.com/snowdreamtech/template/commit/9fba0f1436429e3e3d161358ab1dbc8d299b7fe8))


### Bug Fixes

* **lint:** restore color output in Pass 1 auto-fix stage ([fe3d523](https://github.com/snowdreamtech/template/commit/fe3d52320f604686e0f5dededbd07da463a0f176))

## [0.8.2](https://github.com/snowdreamtech/template/compare/v0.8.1...v0.8.2) (2026-04-19)


### Bug Fixes

* **ci:** add GITHUB_TOKEN to all make commands for API rate limit prevention ([2060bca](https://github.com/snowdreamtech/template/commit/2060bca9cef56a23645775c519e8054fdf1b0261))
* **ci:** skip DCO check for snowdream user in release-please PRs ([dde607b](https://github.com/snowdreamtech/template/commit/dde607bc3c474bf8967811bb9902c672f998717b))
* **ci:** use GITHUB_TOKEN for release-please to prevent tag creation issues ([73266aa](https://github.com/snowdreamtech/template/commit/73266aa2d12b86d1cce4ba785d48141dbe817a44))


### Reverts

* **ci:** restore WORKFLOW_SECRET fallback for release-please ([4a1956b](https://github.com/snowdreamtech/template/commit/4a1956b4c14b7961932fd9c1851d5849e149107d))

## [0.8.1](https://github.com/snowdreamtech/template/compare/v0.8.0...v0.8.1) (2026-04-19)


### Bug Fixes

* **ci:** resolve Docker Buildx container restart race condition ([01a1c2d](https://github.com/snowdreamtech/template/commit/01a1c2d500ea125ed63e0091fd5a6a136fa58544))

## [0.8.0](https://github.com/snowdreamtech/template/compare/v0.7.1...v0.8.0) (2026-04-18)


### Features

* **release:** add separate-pull-requests for future manifest mode ([a543dde](https://github.com/snowdreamtech/template/commit/a543dded3032285624672de89e68d37130970bad))

## [0.7.1](https://github.com/snowdreamtech/template/compare/v0.7.0...v0.7.1) (2026-04-18)


### ⚠ BREAKING CHANGES

* **security:** mise now downloads binaries ONLY from GitHub Releases
* mise does NOT automatically detect musl

### Features

* add comprehensive performance testing and documentation infrastructure ([c1d8996](https://github.com/snowdreamtech/template/commit/c1d8996e4f44ce7222cf589812325b700e7e5f96))
* add missing performance and documentation scripts ([fe65a58](https://github.com/snowdreamtech/template/commit/fe65a58f25c3214e93d156a03079c51d169d3240))
* add VER_MISE to versions.sh for centralized version management ([b805d26](https://github.com/snowdreamtech/template/commit/b805d26395103dbea4c34e14b5c99521537dabbd))
* **ci:** add atomic tool verification for shfmt and editorconfig-checker ([dd255e5](https://github.com/snowdreamtech/template/commit/dd255e5956a3a9b96c15cc832fddc3667a3effb2))
* **ci:** add atomic verification for config/doc linting tools ([5d8de8b](https://github.com/snowdreamtech/template/commit/5d8de8bc179790eac4518f73f03d5f6e8d5c5af6))
* **ci:** add atomic verification for hadolint and dockerfile-utils ([9e5a824](https://github.com/snowdreamtech/template/commit/9e5a824b22e1b31bcf08c9e7ae265ec836077bb0))
* **ci:** add atomic verification for Node.js tools ([4a6c20f](https://github.com/snowdreamtech/template/commit/4a6c20f4756f2605971c12a7a12feabc1d56bad4))
* **ci:** add centralized Harden Runner endpoints configuration ([5389a6e](https://github.com/snowdreamtech/template/commit/5389a6eb811a6b1ecfeefd6a0a3512e9a7d8b078))
* **ci:** enable release-please for all configured branches ([bc94753](https://github.com/snowdreamtech/template/commit/bc94753270413815bb340e27fd6154f41e37d771))
* **ci:** harden stateless toolchain and remediate security audit findings ([ebab039](https://github.com/snowdreamtech/template/commit/ebab039b1f58b3a782bec8a71d23596b1c59f2ce))
* **ci:** implement universal binary-first tool verification ([5bedab1](https://github.com/snowdreamtech/template/commit/5bedab16aba4891d0395ba2e0462148810c8b2ac))
* **ci:** improve release-please configuration and remove fixed version ([e1fcd64](https://github.com/snowdreamtech/template/commit/e1fcd648d91c306f6f385eca7600b3d6d2fefc5e))
* **ci:** skip DCO check for bot commits ([b841cee](https://github.com/snowdreamtech/template/commit/b841cee3308b970f945f92692d807cf353a7b379))
* **ci:** sync Harden Runner endpoints from centralized config ([48b4221](https://github.com/snowdreamtech/template/commit/48b42213ee409890c4dd9a64f230c251dbac7b77))
* **common:** enhance mise backend dependency checks and add MISE_GITHUB_ENTERPRISE_TOKEN forwarding ([27e0f1e](https://github.com/snowdreamtech/template/commit/27e0f1e7dfd3758bfa4bdc8013410a4bdbec386e))
* **dev:** add POSIX-compatible CI simulation script ([a587a39](https://github.com/snowdreamtech/template/commit/a587a390faa640adce2d448b773222cbb373c41d))
* **devcontainer:** add comprehensive SSH and GPG permission configuration ([843180b](https://github.com/snowdreamtech/template/commit/843180b5722d20b1704e482d40b1c4a80f890fdc))
* **devcontainer:** add Docker availability check in init script ([7ca1fe4](https://github.com/snowdreamtech/template/commit/7ca1fe4c37ada1893e94e9ac28ec578952c1d99e))
* **devcontainer:** add support for local git config file ([2fb5e43](https://github.com/snowdreamtech/template/commit/2fb5e435b5038617746aff9837dec98865cd4d5c))
* **devcontainer:** enable GPG signing support ([4d88e12](https://github.com/snowdreamtech/template/commit/4d88e125ddc84966237d382d897c03b1cf73da06))
* **docker:** add docker-compose file detection for hadolint and dockerfile-utils ([836880d](https://github.com/snowdreamtech/template/commit/836880de8c86ea0e548ed9edabe589d38357903f))
* enhance PATH management for dynamically installed tools ([f307921](https://github.com/snowdreamtech/template/commit/f30792120c89c79cf2c7216b3ca5cfbe9d22dd31))
* enhanced PATH management for dynamic tool installation ([43c8ba6](https://github.com/snowdreamtech/template/commit/43c8ba6cd306cbe387593481dc6d7c28357557db))
* **hooks:** strengthen pre-commit with security & deep linting audits ([1427366](https://github.com/snowdreamtech/template/commit/14273660790f2c9f3806528114d9faec53c1f26b))
* **make:** add sync-harden-runner target for workflow endpoint management ([d24708d](https://github.com/snowdreamtech/template/commit/d24708dd86bb09946f3337dfd0747ca5fbcd51bc))
* **mise:** add yamllint to core tools in mise.toml ([406bbb3](https://github.com/snowdreamtech/template/commit/406bbb32f3a9f47bdfee72e2422b273a6cab252b))
* optimize mise shell activation with full path and official method ([77b21a6](https://github.com/snowdreamtech/template/commit/77b21a6257b5b219bfb5dc4057d28ffaeb2f0c82))
* optimize mise shell activation with full path support ([e71c33f](https://github.com/snowdreamtech/template/commit/e71c33f88abb2a2fd11561f7067b5c1e479f2b43))
* **scripts:** add cleanup script for duplicate mise activation lines ([b7020c6](https://github.com/snowdreamtech/template/commit/b7020c6ca19e167c71b776f1380c5f1dc8c20c3c))
* **scripts:** add missing performance and documentation scripts ([f8d9c61](https://github.com/snowdreamtech/template/commit/f8d9c613cfa599e503bed8190e13da66cbf928d7))
* **scripts:** add tool to remove Windows wrapper scripts ([5bfd3ef](https://github.com/snowdreamtech/template/commit/5bfd3ef8ec9241cb09eca577d58616d1c29e2987))
* **scripts:** implement Node.js JSON parser with CommonJS ([e0ffb06](https://github.com/snowdreamtech/template/commit/e0ffb06a5e8634f05d171a440aa515c881180d30))
* **security:** add sigstore wildcard endpoint to harden runner ([2e26b98](https://github.com/snowdreamtech/template/commit/2e26b98d181cc8e4c739febde72e90fcfdee99da))
* **security:** implement three-layer defense against Aqua Registry ([ef13665](https://github.com/snowdreamtech/template/commit/ef136651dc6d7d358cf4de2748a0a10a841730bf))
* **security:** refactor audit orchestration for automatic activation in local dev ([f085c71](https://github.com/snowdreamtech/template/commit/f085c711f6e3b5093b1c8ab9efff2b31e32725d1))
* use mise.jdx.dev install script with version support ([043b346](https://github.com/snowdreamtech/template/commit/043b3460ae67ae946018155cab31396770741561))
* use VER_MISE from versions.sh in common.sh ([d4c8160](https://github.com/snowdreamtech/template/commit/d4c81601aa888932a2b0b01ef3f9208f348b5a4e))


### Bug Fixes

* add pattern matching for binaries with version/platform suffixes ([9ce22d9](https://github.com/snowdreamtech/template/commit/9ce22d9e618960d62c4711bea1c9130d5d82f3df))
* add unified PATH management and CI persistence to run_mise ([19c6431](https://github.com/snowdreamtech/template/commit/19c64317dc58d31ddeb16ef4f7657795fe49ef35))
* align log_summary to use CI_STEP_SUMMARY for consistent reporting ([601ac65](https://github.com/snowdreamtech/template/commit/601ac65d390330fa51a7e2c025b2a56ac3834188))
* **base:** add version specifications to mise install calls ([71aa2fe](https://github.com/snowdreamtech/template/commit/71aa2fe0746f8fdfc8178de4971abaf002509108))
* **bootstrap:** prevent duplicate mise activation lines in shell RC files ([0f343f1](https://github.com/snowdreamtech/template/commit/0f343f1fb933646753e52afead32f0ef3d887099))
* **bootstrap:** prevent shell-specific syntax errors in POSIX sh ([cc6066d](https://github.com/snowdreamtech/template/commit/cc6066d1004c52e9ecb60892df1a2f10645a414f))
* **bootstrap:** use dynamic mise paths instead of hardcoded values ([062f35a](https://github.com/snowdreamtech/template/commit/062f35ac1900d580735f094e86f55befba2d3400))
* **cd:** add debug output and improve PATH setup for Windows ([892d860](https://github.com/snowdreamtech/template/commit/892d8608c81b7c2c27532a7cbf02b0d06792057a))
* **cd:** complete debug output alignment with CI workflow ([b52cd4f](https://github.com/snowdreamtech/template/commit/b52cd4f93ce813f0a2bea3d135ff81bfc761f4e6))
* **cd:** ensure mise shims are on PATH for all verification steps ([95687fe](https://github.com/snowdreamtech/template/commit/95687fe0f88e57b638e5d273be8b5bbd528ed760))
* centralize hardcoded provider values across shell scripts ([2a170e7](https://github.com/snowdreamtech/template/commit/2a170e7626c456e116c59e08eff7e49e324fa551))
* **check-env:** strip leading 'v' from versions for accurate comparison ([83456b3](https://github.com/snowdreamtech/template/commit/83456b3f8624aaf6cc2b3feee48c398102d16ed8))
* **check-env:** sync GITHUB_PATH to current shell in CI ([8fdde4e](https://github.com/snowdreamtech/template/commit/8fdde4e89bd198a7c5f8d78a396c27bae3abeae6))
* **check-env:** use full mise keys for version detection ([a5b1063](https://github.com/snowdreamtech/template/commit/a5b10635d40cde00554db96ec953025aeae7f5a6))
* **check-env:** version mismatch should warn not fail ([e98d7b3](https://github.com/snowdreamtech/template/commit/e98d7b3d973c69847f5f817fef3a374ab821a14c))
* **ci:** add aggressive cache refresh after uninstall in install_tool_safe ([433899c](https://github.com/snowdreamtech/template/commit/433899cd8435f737a6193f15e3189a987260e5ef))
* **ci:** add DCO signoff to dependabot-sync commits ([cfe7546](https://github.com/snowdreamtech/template/commit/cfe75468ce43c5b61c7ac8e5915eb767b0032815))
* **ci:** add DCO signoff to release-please commits ([8fd8b67](https://github.com/snowdreamtech/template/commit/8fd8b6729aa68027072f68acec9735397eea54cc))
* **ci:** add mise exec fallback for pnpm audit on Windows ([f1e93b7](https://github.com/snowdreamtech/template/commit/f1e93b73541e2a1f5c9c42c81d89e97cbf91c57d))
* **ci:** add mise.lock and versions.sh to mise cache key in pages workflow ([9d7b9d7](https://github.com/snowdreamtech/template/commit/9d7b9d75d68645afaa798a577758ea55fa94cb03))
* **ci:** add mise.lock to mise cache key in cd workflow ([d43cdf9](https://github.com/snowdreamtech/template/commit/d43cdf9ab8d2c22af779601acfbab8fe42b035be))
* **ci:** add mise.lock to mise cache key in dependabot-sync workflow ([070adf0](https://github.com/snowdreamtech/template/commit/070adf0610122d8ed3e098d72838714190ec8871))
* **ci:** add mise.lock to mise cache key in label-sync workflow ([fccb9f0](https://github.com/snowdreamtech/template/commit/fccb9f0669b118ecaeac1e88a771534a13fc8455))
* **ci:** add mise.lock to mise cache keys in ci workflow (3 jobs) ([ac8ecf6](https://github.com/snowdreamtech/template/commit/ac8ecf6adb81781556c969916c2e0a62b99f035c))
* **ci:** add missing endpoints for trivy and sigstore ([115a19d](https://github.com/snowdreamtech/template/commit/115a19de02c7cd1d274ecb0191bd035a02d273ef))
* **ci:** add node-audit special handling in CI fallback ([d82624c](https://github.com/snowdreamtech/template/commit/d82624cda69bbf48e21ef2693d5c5477c278463e))
* **ci:** add OSV-scanner and Zizmor to Tier 1 tools ([7bf9561](https://github.com/snowdreamtech/template/commit/7bf9561f2459bd68ff422d61e3e8967df5a9d65f))
* **ci:** comprehensive tool installation and execution fix ([bac4a9c](https://github.com/snowdreamtech/template/commit/bac4a9ce77e15170d3a8f3e1db970b00917e7abc))
* **ci:** correct action SHA and version tag mismatches in cd.yml ([f4d90ec](https://github.com/snowdreamtech/template/commit/f4d90ec5d3570189d0f8c5562f4eea59f70a899f))
* **ci:** detect docker-compose files in dependabot generator ([530ca75](https://github.com/snowdreamtech/template/commit/530ca7534583ea16bb876c0e8bfbf209d958a42a))
* **ci:** disable paranoid mode in CI for lockfile sync ([1b282ce](https://github.com/snowdreamtech/template/commit/1b282ce6fa3d1ecca5c18de89a4962ec7e50226b))
* **ci:** enable release-please for dev branch and fix DCO signoff ([ab8f065](https://github.com/snowdreamtech/template/commit/ab8f065dce45cc794d78aa075fc8c066eeec2bf5))
* **ci:** ensure PATH persistence after setup completes on Windows ([460dc8d](https://github.com/snowdreamtech/template/commit/460dc8d84067a89081762401269565a815ceab00))
* **ci:** ensure shfmt and handle trivy absence in CI ([fae3f46](https://github.com/snowdreamtech/template/commit/fae3f46d2c253d3c2883814d24f33169df3325b2))
* **ci:** exclude unreleased version compare links from lychee checks ([6165757](https://github.com/snowdreamtech/template/commit/6165757ebf73b2d1b7abf5c738d66f6402a06f14))
* **ci:** explicitly specify checkmake mise tool spec to avoid aqua registry lookup ([acefebb](https://github.com/snowdreamtech/template/commit/acefebb98efc3563a9011f2ea05e7cc7274f5aef))
* **ci:** extract version from TOML table syntax in install_tool_safe ([9740407](https://github.com/snowdreamtech/template/commit/9740407dd5b970e1bcecc26d58ce756f2fd36027))
* **ci:** fix PSScriptAnalyzer and yamllint configuration ([166826a](https://github.com/snowdreamtech/template/commit/166826afd22a16c5fe7ccf131ae28d6f46d48525))
* **ci:** fix yamllint truthy error in performance workflow ([7e7dfd0](https://github.com/snowdreamtech/template/commit/7e7dfd052065a8b361212d0a1611a60bc951b06b))
* **ci:** handle mise shims in verify_binary_exists ([bc0dfd9](https://github.com/snowdreamtech/template/commit/bc0dfd945090a3b63cf27ee90b9073547ff76ec5))
* **ci:** harden GITHUB_PATH persistence and Windows mise detection ([6c555db](https://github.com/snowdreamtech/template/commit/6c555dbefb0842a70ab06769935edf1df33f58a7))
* **ci:** immediately update PATH in current shell after writing to GITHUB_PATH ([6fd7c76](https://github.com/snowdreamtech/template/commit/6fd7c764c16a49fd51f180c4310ad2bf08db7401))
* **ci:** improve binary resolution fallback for GitHub tools ([53ae080](https://github.com/snowdreamtech/template/commit/53ae0802fb5eb5d618f71be3f9f0fe6d9d89232a))
* **ci:** improve post-install binary name resolution ([46ab8b0](https://github.com/snowdreamtech/template/commit/46ab8b00c8b5f65bdd1e2171964650ef46359e75))
* **ci:** make link checker non-blocking to avoid false failures ([99c2ec9](https://github.com/snowdreamtech/template/commit/99c2ec99befd980c996953bd53471ded5f6d404a))
* **ci:** make lint.sh fail when pre-commit hooks fail ([78ca50a](https://github.com/snowdreamtech/template/commit/78ca50ac3a1e670d5c11b7a0292585976384cbfd))
* **ci:** pass GITHUB_TOKEN to lychee for link checking ([dfb9fd1](https://github.com/snowdreamtech/template/commit/dfb9fd185e4f1bbe4987b4e991db85a4a03639e8))
* **ci:** pin GitHub Actions to commit SHA in performance workflow ([13623d9](https://github.com/snowdreamtech/template/commit/13623d9cf5ae7a6677a8645a8d568bb39a2752a2))
* **ci:** quote 'on' keyword in performance workflow ([4710acd](https://github.com/snowdreamtech/template/commit/4710acd942f39592bfe55f4ceb73af7ec47153c7))
* **ci:** remove --template flag from gh label list command ([27b29ea](https://github.com/snowdreamtech/template/commit/27b29ea1cabb48a35398324473ea3e8133f07ed3))
* **ci:** remove emoji from log messages to fix Windows printf errors ([f14a5f0](https://github.com/snowdreamtech/template/commit/f14a5f00870474761daa7a54a231a64cc70d668b))
* **ci:** remove MISE_OFFLINE=true from all workflow files ([4be4079](https://github.com/snowdreamtech/template/commit/4be407962334edee9c465e296d7e5d18bfc7d2ea))
* **ci:** remove SKIP_MODULES and stabilize test assertions ([241275f](https://github.com/snowdreamtech/template/commit/241275fa1ff08cb994b8c5dabdfb1a9b0ae1a538))
* **ci:** remove unsupported signoff parameter and add release-as 0.5.0 ([d01b2df](https://github.com/snowdreamtech/template/commit/d01b2dffa875d815316b5d41922b58ce9d574879))
* **ci:** remove Windows path conversion for GitHub Actions PATH persistence ([4af9416](https://github.com/snowdreamtech/template/commit/4af9416b6ea4fba25606ef2bace671895d56e921))
* **ci:** resolve actual binary names for platform-specific tools ([e6856c9](https://github.com/snowdreamtech/template/commit/e6856c9a622d12a915744b44f1542106c87034ad))
* **ci:** resolve npm-pnpm-audit hook failure on Windows ([095f3a5](https://github.com/snowdreamtech/template/commit/095f3a553e352c426e39aa727b37cd3f610b3408))
* **ci:** restore actions/cache version hash in ci workflow ([713881a](https://github.com/snowdreamtech/template/commit/713881ae0bd8929ce69c02ae89482cd26f643e99))
* **ci:** restore actions/cache version hash in dependabot-sync workflow ([e506978](https://github.com/snowdreamtech/template/commit/e5069784c2269abf3965875306a42997d358a4f5))
* **ci:** restore actions/cache version hash in pages workflow ([80a625d](https://github.com/snowdreamtech/template/commit/80a625dcbbe95259fbaae26d59cbb61fc7e2539e))
* **ci:** restrict release-please to main branch only ([0db8a91](https://github.com/snowdreamtech/template/commit/0db8a910c4bbba1426944ccfcca648a593bb097e))
* **ci:** restrict release-please to main branch only ([b97162a](https://github.com/snowdreamtech/template/commit/b97162aa0092875bb8af1820cb4db6c06d17665b))
* **ci:** revert goreleaser to use --version flag ([564d899](https://github.com/snowdreamtech/template/commit/564d899198c6f0c19872642d6b3e2968d54536ab))
* **ci:** specify bin names for shfmt and editorconfig-checker in mise.toml ([f249809](https://github.com/snowdreamtech/template/commit/f2498091f79349cfe84fe160ad78477ad28a22b5))
* **ci:** support binaries installed in root directory ([72631b4](https://github.com/snowdreamtech/template/commit/72631b4d0f2c3cd53ea50b5e42a64585a92e98fc))
* **ci:** suppress zizmor findings for performance.yml and skip flaky tests ([0616771](https://github.com/snowdreamtech/template/commit/06167717e490f73f34e49386e4a776907caa554c))
* **ci:** temporarily disable npm-pnpm-audit hook due to API compatibility ([6ed58d9](https://github.com/snowdreamtech/template/commit/6ed58d92ebd50ff746dfe5f1d630f59cc7eb4441))
* **ci:** update actions/cache version hash in ci workflow ([36520ec](https://github.com/snowdreamtech/template/commit/36520ec4b721ba79ccf6b410573d02e416a53f29))
* **ci:** update node-audit to use new npm bulk advisory endpoint ([526ca79](https://github.com/snowdreamtech/template/commit/526ca79e307d38240e54505bda3b06d75f1e8b9a))
* **ci:** update Phase 1 tools to use 4-parameter verify_tool_atomic ([5af97ed](https://github.com/snowdreamtech/template/commit/5af97ed9daa59c3f2a900f25ff22dd2f7bf634bb))
* **ci:** use correct version command for goreleaser ([b271332](https://github.com/snowdreamtech/template/commit/b271332e9b02f2f34513943f86ce8297f3c31dc4))
* **ci:** use GitHub binary providers for cross-platform security tools ([8d150f6](https://github.com/snowdreamtech/template/commit/8d150f6554d8f6a752f7c2351e62f637a1f52dc3))
* **ci:** use GITHUB_TOKEN for release-please to enable DCO signoff ([51051ba](https://github.com/snowdreamtech/template/commit/51051ba4a2e7fdca3a9ed927e09fdec94b20f77d))
* **ci:** use mise exec for shim smoke tests ([12306a9](https://github.com/snowdreamtech/template/commit/12306a9fbf8a8f4192002424aa03898223628077))
* **ci:** use npm instead of pnpm for audit to avoid API compatibility issues ([641c9f7](https://github.com/snowdreamtech/template/commit/641c9f744927c5aa15369cd56bcb7e31bb44bf3d))
* **ci:** use platform-specific binary name for editorconfig-checker ([2b38f2a](https://github.com/snowdreamtech/template/commit/2b38f2a2fbb5445b5b5118a64efa34bd56495023))
* **ci:** use valid log level for lychee verbose option ([171e721](https://github.com/snowdreamtech/template/commit/171e7218eb1e6cf00645c27f7d9d8187e5c9bf47))
* **ci:** use WORKFLOW_SECRET for release-please to enable chain-triggering ([9339b9c](https://github.com/snowdreamtech/template/commit/9339b9c343147d687e69ab362653b7df10accb82))
* **common:** correct find -maxdepth position in has_lang_files ([9866fe6](https://github.com/snowdreamtech/template/commit/9866fe68a40c64f47a3fc2d1c7abf94dd5d10022))
* **config:** remove deprecated ExperimentalScannerConfig from osv-scanner ([ce385db](https://github.com/snowdreamtech/template/commit/ce385db7ca970de93405cb85988e73e32f8f8c90))
* correct Alpine Linux Node.js installation behavior ([bbd85fa](https://github.com/snowdreamtech/template/commit/bbd85fa7048b058ee6a2d13ccf27debfac4912e7))
* **cpp:** add version specification to mise install call ([b6fbdaf](https://github.com/snowdreamtech/template/commit/b6fbdaf1f4866b77b7c1a8376e82f7d8dd5c7cd9))
* **dco:** use PR author instead of branch name to prevent bypass ([eb55adf](https://github.com/snowdreamtech/template/commit/eb55adf2ccc05f6a0bea0a1f9c37c5fb56f08e02))
* **deps:** correct toolchain provider mappings and dependabot schema ([76d515d](https://github.com/snowdreamtech/template/commit/76d515d4c3b0abe40ed6cc849b1f3ed2077887bb))
* **deps:** replace asdf pipx backend with native mise plugin to resolve Windows CI failures ([05aba02](https://github.com/snowdreamtech/template/commit/05aba026dd7348bb2fb17848076a68b15411e33f))
* **devcontainer:** add command availability checks before usage ([a0a4ce0](https://github.com/snowdreamtech/template/commit/a0a4ce0d1bed246c722270f920e79f2046a1bfe7))
* **devcontainer:** add command to keep container running ([5484045](https://github.com/snowdreamtech/template/commit/54840456dfb76e040655d49cf28b45e9b64ead1d))
* **devcontainer:** add workspace volume mount and fix configuration ([9ff708b](https://github.com/snowdreamtech/template/commit/9ff708be2aa74af62eae40012f163229bb688296))
* **devcontainer:** add YAML document start marker for yamllint ([44ccf81](https://github.com/snowdreamtech/template/commit/44ccf8108ad6ea15b92069f15bbc1ce852723ec0))
* **devcontainer:** correct updateContentCommand to use make install ([774bba2](https://github.com/snowdreamtech/template/commit/774bba245b5025973492b2cd8fb10826eae7af12))
* **devcontainer:** handle missing host directories gracefully ([9b79301](https://github.com/snowdreamtech/template/commit/9b7930189a20c8c44983a739d80ec05d8b5987f5))
* **devcontainer:** make git config cross-platform compatible ([600052a](https://github.com/snowdreamtech/template/commit/600052ae9d9067de738a875009df7e8371099625))
* **devcontainer:** use dynamic workspace folder variable ([73421d6](https://github.com/snowdreamtech/template/commit/73421d6f6724ae63613d30566103b55191d43c87))
* **devcontainer:** use find instead of ls for directory permissions ([ffeaee8](https://github.com/snowdreamtech/template/commit/ffeaee833e3a9f4ad7a88d45b906ad12ca52b62d))
* **devcontainer:** use global git config instead of local ([934bb1c](https://github.com/snowdreamtech/template/commit/934bb1c3bec475b571daab4168cdd233d10b6816))
* **docker:** add version specifications to mise install calls ([72149e4](https://github.com/snowdreamtech/template/commit/72149e4b14f49d2ad67543af54f1d654b5efcbb8))
* **docs:** ensure make docs build works correctly ([6a0df9c](https://github.com/snowdreamtech/template/commit/6a0df9ce6acf22d2d8606dda0444c48d369967bd))
* **docs:** fix broken relative links in documentation ([c13692c](https://github.com/snowdreamtech/template/commit/c13692c6ecf4fca79bf117fcda7076abef1acba1))
* **docs:** regenerate pnpm lockfile to match overrides ([b553918](https://github.com/snowdreamtech/template/commit/b553918d8c7d81bc4a7f740a756c590a06d48e2a))
* **docs:** resolve MD028 blockquote lint in snowdreamtech.init workflow ([4e9ecfa](https://github.com/snowdreamtech/template/commit/4e9ecfabb7e9bb7a2af0b3e0c73c56cb31f05262))
* **docs:** resolve vitepress from docs/node_modules/.bin ([13dc8f9](https://github.com/snowdreamtech/template/commit/13dc8f9deb1daf7c3ae60190a8b497dc4fbf0249))
* **docs:** update TPGi CCA link to GitHub releases page ([b061ca1](https://github.com/snowdreamtech/template/commit/b061ca1150e4dd23b47fd88b3f9992d0f97c1df8))
* **env:** improve mise tool resolution and register zizmor ([120db7e](https://github.com/snowdreamtech/template/commit/120db7ec0ee94004b7b06337793ac5a8c9a93b0a))
* **go:** add version specifications to mise install calls ([4b6964f](https://github.com/snowdreamtech/template/commit/4b6964fbb2e43ea3674d433058d41f6f0db80335))
* **helm:** add version specification to mise install call ([4b5f452](https://github.com/snowdreamtech/template/commit/4b5f452e8003e64639a89447d96abd16abf5ce2c))
* improve POSIX compatibility in bin-resolver.sh ([ebab76e](https://github.com/snowdreamtech/template/commit/ebab76e69041ace9d2bcbeef107044763a27a991))
* improve stat command cross-platform compatibility ([031ffce](https://github.com/snowdreamtech/template/commit/031ffceb774c785a104eb93e715e4966786dafd4))
* **install:** add atomic verification to all Phase 3 low priority tools ([631318f](https://github.com/snowdreamtech/template/commit/631318fd13a28676f6015eca24c3759a86d2c4f2))
* **install:** add atomic verification to clang-format, google-java-format, stylua, ktlint ([0b92a0d](https://github.com/snowdreamtech/template/commit/0b92a0d1e40ca2b00690a204e2c077526e6a2886))
* **install:** add atomic verification to Phase 2 remaining tools ([bb0ae96](https://github.com/snowdreamtech/template/commit/bb0ae96ddb968b00495196547fc5795f3444d8e3))
* **java:** add version specification to mise install call ([8f359a1](https://github.com/snowdreamtech/template/commit/8f359a1fef21822d9dfe733cd4a45e8bc0eaa28e))
* **lint:** add aggressive reinstall logic in lint-wrapper ([40ab04d](https://github.com/snowdreamtech/template/commit/40ab04d550adb207e6fd954afdae6fe48da7f6a3))
* **lint:** add auto-install fallback for missing tools in CI ([b293923](https://github.com/snowdreamtech/template/commit/b2939234dc31427706c716bc261d29e0635e0174))
* **lint:** add license headers to test files and fix markdown formatting ([02263cc](https://github.com/snowdreamtech/template/commit/02263ccdd8163b427643af8fce0103232b01b82c))
* **lint:** add reshim and direct execution fallback ([21be93b](https://github.com/snowdreamtech/template/commit/21be93bb2f143b16f2476a7bb5f46bb838a09961))
* **lint:** add tool spec mapping for shfmt, taplo, and editorconfig-checker ([86c47ac](https://github.com/snowdreamtech/template/commit/86c47aceaedce07dda24802869a9f024a19616df))
* **lint:** add zizmor to optional security tools list ([13b06fd](https://github.com/snowdreamtech/template/commit/13b06fdf2720ea8b493dff1301a889c461a1f589))
* **lint:** correct editorconfig-checker binary name to 'ec' ([c56679c](https://github.com/snowdreamtech/template/commit/c56679c2515de0754e593a7f295623a0e075577c))
* **lint:** exclude all markdown files from prettier formatting ([4fee420](https://github.com/snowdreamtech/template/commit/4fee420ee9f672867de2c0f0c59cf00baf8d8ad3))
* **lint:** exclude test files from shellcheck-posix and markdown from prettier ([0dfd5b2](https://github.com/snowdreamtech/template/commit/0dfd5b2eb22337c03b4c06a7879369328922e7fa))
* **lint:** fix yamllint, markdownlint, and editorconfig issues ([b7925a2](https://github.com/snowdreamtech/template/commit/b7925a2f9661b860f50134e5de1a028f172e971a))
* **lint:** prevent prettier/markdownlint emphasis format conflict ([0c38f99](https://github.com/snowdreamtech/template/commit/0c38f994d6f4592e35e02e9dbb85244722f1388c))
* **lint:** resolve spurious warning by reporting pre-commit version ([188a325](https://github.com/snowdreamtech/template/commit/188a3251cb2f76d053bbb5e9ccef5846eae4d269))
* **lint:** use mise exec as fallback when resolve_bin fails in CI ([8e52231](https://github.com/snowdreamtech/template/commit/8e5223113a198757c33ede338ebd6fc55fac84b0))
* **linux:** use -perm /111 instead of +111 for better compatibility ([512e69b](https://github.com/snowdreamtech/template/commit/512e69b504a1810c4fa126fab5b6079bfbba125d))
* **macos:** avoid -perm flag for BSD find compatibility ([79b98b0](https://github.com/snowdreamtech/template/commit/79b98b071476afe24a80e9876c7e2509022c1529))
* **macos:** use -perm +111 instead of -executable for BSD find compatibility ([911c1b4](https://github.com/snowdreamtech/template/commit/911c1b47e643bcb1d95e6400a7e6abdbfeea82fe))
* **mise:** add additional paranoid mode disable flags ([7e83141](https://github.com/snowdreamtech/template/commit/7e83141b005bb8e0e6dc930f35eec7436c1bceb1))
* **mise:** align pnpm version to 10.30.3 to fix detection failure ([c01671d](https://github.com/snowdreamtech/template/commit/c01671df9125246be64d26cb4129450684067ef6))
* **mise:** clear cache before sync-lock to prevent stale provenance verification ([027874d](https://github.com/snowdreamtech/template/commit/027874db7c3ce6b823576d2bdb3da0bf868f5209))
* **mise:** disable attestation checks in sync-lock to prevent false positives ([3f09708](https://github.com/snowdreamtech/template/commit/3f09708fefcf850b11f0ed2f7c29525f2c6c2b0d))
* **mise:** downgrade node to 22.14.0 for macOS 12 compatibility ([c5495b4](https://github.com/snowdreamtech/template/commit/c5495b491a7796d25bd2ff26fbb8cbd9fdc43aff))
* **mise:** downgrade Python from 3.14.3 to 3.13.12 ([71ae8fd](https://github.com/snowdreamtech/template/commit/71ae8fd1f1750ed0440b18be1515610063cee1e8))
* **mise:** enable GitHub proxy locally and fix checkmake version ([a0e65aa](https://github.com/snowdreamtech/template/commit/a0e65aaaa862ddc657beeac74fd598c509c54c76))
* **mise:** handle platform-specific editorconfig-checker binary names ([5ebb3ff](https://github.com/snowdreamtech/template/commit/5ebb3ff76b07f9eb31d61755fd064746512dbd69))
* **mise:** increase HTTP timeout to 300s for slow GitHub downloads ([123aa0c](https://github.com/snowdreamtech/template/commit/123aa0cd8b8f07ccc48e59758f887e778e76b316))
* **mise:** make lint ([097f32b](https://github.com/snowdreamtech/template/commit/097f32bdffbaf3ccc78eaa367ecb3e40a2935939))
* **mise:** permanently disable paranoid mode for lockfile sync ([49dbb04](https://github.com/snowdreamtech/template/commit/49dbb044f65c4c5203edd751218f3d90315214f6))
* **mise:** remove provenance fields from lockfile to prevent attestation errors ([f7790b5](https://github.com/snowdreamtech/template/commit/f7790b5803b47822c1191f28c155763a3ce43844))
* **mise:** skip Go checksum verification for Aliyun mirror ([8821f81](https://github.com/snowdreamtech/template/commit/8821f814b7bb37d56d8e45e4163cc3292624fafd))
* **mise:** synchronize multi-platform lockfile with pnpm 10.30.3 ([fc2113b](https://github.com/snowdreamtech/template/commit/fc2113b5c63b845ca4a8631cd50fde9db5e59d15))
* **mise:** update yamllint version to match versions.sh ([d9bcf61](https://github.com/snowdreamtech/template/commit/d9bcf61f7092f248186d41ceb66fcd98d3015a34))
* **mise:** workaround for GitHub attestation verification failures ([2911054](https://github.com/snowdreamtech/template/commit/29110547a86ebe51593eac637c0b4dc4c14977bc))
* **node:** add corepack resiliency and npm fallback ([d72055b](https://github.com/snowdreamtech/template/commit/d72055b2e0c3eeb97bb9103b413188fe2f21f0be))
* **node:** add version specifications to mise install calls ([afd9c04](https://github.com/snowdreamtech/template/commit/afd9c0456fea9d3a427dfe8f70913d7e745b92fb))
* **node:** auto-install bash in Alpine for npm compatibility ([0042fa2](https://github.com/snowdreamtech/template/commit/0042fa2abb6ae389277660b3d0656f54ea669643))
* **node:** configure mise for musl binaries in Alpine environments ([232bce2](https://github.com/snowdreamtech/template/commit/232bce29a4401ebe42fd529c448a47d404562abe))
* **node:** explicitly set MISE_NODE_* env vars for Alpine ([25d506d](https://github.com/snowdreamtech/template/commit/25d506d6d952fa479dca79e427d10905d37fc9b6))
* **node:** export ALPINE_VERSION for mise.toml template evaluation ([62d9344](https://github.com/snowdreamtech/template/commit/62d934444f440eec678ac2227c723ab7f1e46e8d))
* **node:** resolve install hang by skipping redundant corepack activation ([36b3d44](https://github.com/snowdreamtech/template/commit/36b3d448f4ad914621752f4504560e6d4d110975))
* **node:** use boolean false for node.compile setting ([97d5bd5](https://github.com/snowdreamtech/template/commit/97d5bd57d4dc1ffc52a777524c5aabf92a9a061d))
* **openapi:** add version specification to mise install call ([81e52b5](https://github.com/snowdreamtech/template/commit/81e52b5bcd5fa9f7a82036aabbb2fd24727783a0))
* **pre-commit:** migrate deprecated stage name ([8392cbf](https://github.com/snowdreamtech/template/commit/8392cbfe73aee175bb7732255a7f992c75078184))
* **pre-commit:** skip osv-scanner in local pre-commit hooks ([b4b82ac](https://github.com/snowdreamtech/template/commit/b4b82ac7078d197c8a4dfac85a2748cdee5f61e4))
* **protobuf:** add version specification to mise install call ([1f7938e](https://github.com/snowdreamtech/template/commit/1f7938e2cbaf3661a0387bd476ed73ce6e7d4e15))
* **python:** add version specifications to mise install calls ([b808c7e](https://github.com/snowdreamtech/template/commit/b808c7e16cafa6fdbf6c50ef8be5a8639401c9e4))
* **python:** install build dependencies for Alpine/musl environments ([3f53883](https://github.com/snowdreamtech/template/commit/3f53883e79ed5b9d74ca2369052b0014a3d4f0fc))
* re-enable refresh_mise_cache with timeout protection ([efe2b3e](https://github.com/snowdreamtech/template/commit/efe2b3e98fbf4650b3375bfb11cb0355b97e1081))
* **registry:** propagate mise install failures in CI ([582eeef](https://github.com/snowdreamtech/template/commit/582eeef44ae4d368ae59bc2d5582b49fb974064f))
* **rego:** add version specification to mise install call ([60eb8ee](https://github.com/snowdreamtech/template/commit/60eb8eecfb6c4f1b38d2ce41500c29774ee16e8c))
* replace %b with %s in printf to avoid Windows path issues ([29ba55b](https://github.com/snowdreamtech/template/commit/29ba55b82400d7ec41074c2951cbbd614eab3cb5))
* resolve devcontainer startup with docker-compose configuration ([fa788e4](https://github.com/snowdreamtech/template/commit/fa788e4f3c94e767f465eea9f17acb4ab8f4a01e))
* resolve make verify hanging and add Kiro config ([505e140](https://github.com/snowdreamtech/template/commit/505e140402edadd12c635a1b0d27f3785e0ea717))
* restore mise metadata cache initialization to resolve JIT tool detection ([9b15c4c](https://github.com/snowdreamtech/template/commit/9b15c4cd469e9af229ef73e9f6272705d176349f))
* **runner:** add version specifications to mise install calls ([bf3d5df](https://github.com/snowdreamtech/template/commit/bf3d5df78889a6e27e41f961ba38a239a5e0a8bd))
* **scripts:** add recursion lock cleanup advice to concurrency warning ([ce27ada](https://github.com/snowdreamtech/template/commit/ce27adafa7b711ba6d9ad993e4e95ef534d289b8))
* **scripts:** ensure summary table displays by default ([a0d1b69](https://github.com/snowdreamtech/template/commit/a0d1b69300d151b2f448a2f1375e0f81e97cd6b2))
* **scripts:** fix JSON output formatting issues ([7b30a64](https://github.com/snowdreamtech/template/commit/7b30a640d262c9d47bd75956f92d5d3819cd4ce8))
* **scripts:** harden gen-full-manifest with absolute paths and fix perl syntax ([41199c6](https://github.com/snowdreamtech/template/commit/41199c6493cc0ef4252975645600fe16db423296))
* **scripts:** improve mise path detection in utility scripts ([3c071d8](https://github.com/snowdreamtech/template/commit/3c071d8d4f546cae906e55092dfe9b54566e7d9f))
* **scripts:** improve path robustness and update verifier in release.sh ([000ec98](https://github.com/snowdreamtech/template/commit/000ec98593f2ef52842c6a37de0355fccfc80159))
* **scripts:** install PyYAML via pip instead of pipx in sync-harden-runner ([16285a8](https://github.com/snowdreamtech/template/commit/16285a8ec6f27dada410103a3118482e38b82875))
* **scripts:** prevent recursion lock accumulation across multiple runs ([6d41678](https://github.com/snowdreamtech/template/commit/6d4167845870c19af4f2ed39dec28f32e2096c9b))
* **scripts:** prevent sync-harden-runner from deleting workflow content ([32a4228](https://github.com/snowdreamtech/template/commit/32a4228e6e456a62cf761590491c5c56b6b24b8b))
* **scripts:** resolve infinite recursion and macOS compatibility in setup ([dbc5f9b](https://github.com/snowdreamtech/template/commit/dbc5f9b3bf417c7107641faacb51b4d5a637cfa1))
* **scripts:** resolve PSScriptAnalyzer warnings in sync-labels.ps1 ([5d72d18](https://github.com/snowdreamtech/template/commit/5d72d18de1681078f8299e92ba9dc3d62e90a290))
* **scripts:** resolve silent hang during mise tool resolution ([dec1f65](https://github.com/snowdreamtech/template/commit/dec1f65ace1e5238222771d1ef9d3861c7ea6675))
* **scripts:** standardize mise toolchain providers and fix pnpm setup ([5f8f030](https://github.com/snowdreamtech/template/commit/5f8f0309d42ac6158d865631c2c554bcc18a27d4))
* **scripts:** suppress redundant terminal execution summaries in local environment ([83776fd](https://github.com/snowdreamtech/template/commit/83776fd54d4fe6fe35c6a822c8907d2b7ed8a9d3))
* **scripts:** use absolute path prefix for gen-full-manifest in sync-lock.sh ([bb8401d](https://github.com/snowdreamtech/template/commit/bb8401df026f6389530b895a97950822431efa4a))
* **security:** add binary availability check after version match ([bec5038](https://github.com/snowdreamtech/template/commit/bec5038d8f52f187185c5c2647b38442d395ec5a))
* **security:** add dependabot cooldown config for zizmor compliance ([7394768](https://github.com/snowdreamtech/template/commit/739476828d7812d31505a3dea3d443ffec2d374f))
* **security:** add docker.yml to zizmor ignore rules ([66e35cf](https://github.com/snowdreamtech/template/commit/66e35cf2e63c97fc8b1ecbad2a13e283a16031c7))
* **security:** add strict validation for all CI security tools ([721f8ef](https://github.com/snowdreamtech/template/commit/721f8efbea838617d18cdd62408930752f627e73))
* **security:** add strict zizmor validation in CI ([b6fd596](https://github.com/snowdreamtech/template/commit/b6fd59623c1286a6db9b3b206414ece7dd91a030))
* **security:** add template-injection rule to zizmor config ([92396c1](https://github.com/snowdreamtech/template/commit/92396c15000fed537d3f4057f41dddd3bd7fe69b))
* **security:** add version specifications to mise install calls ([f58a226](https://github.com/snowdreamtech/template/commit/f58a22612304e482bf58f9c64c2614a5574ecf47))
* **security:** check setup_registry return values for critical tools ([cf91e3a](https://github.com/snowdreamtech/template/commit/cf91e3a2ec2274f8335fb6e6863a008d15ceb137))
* **security:** fix OSV-scanner and Zizmor installation in CI ([683639e](https://github.com/snowdreamtech/template/commit/683639e1424d9762128cde26e3e7fad716cbcdef))
* **setup:** add provider paths for lean, nim, racket, vala, and aptos ([c968143](https://github.com/snowdreamtech/template/commit/c968143d1199231c469ff9ad6919419b8c12aefe))
* **setup:** downgrade python to 3.12.3 and disable github proxy to prevent hangs ([4ef9873](https://github.com/snowdreamtech/template/commit/4ef98730f0f0c3f918b07657cdb92d45318cd8df))
* **setup:** ensure gitleaks installs in CI without .git directory ([9fb833d](https://github.com/snowdreamtech/template/commit/9fb833df3c5cbd2af280ebee4c3211ab16a2d858))
* **setup:** ensure yaml and toml modules are always installed locally ([7ca7026](https://github.com/snowdreamtech/template/commit/7ca7026761c0a58a939935db83e37922a872a243))
* **setup:** force reinstall non-executable tools in CI ([dee60a7](https://github.com/snowdreamtech/template/commit/dee60a74f8d22948bc55fca92a6d464eb131764d))
* **setup:** handle mise bin tools verification ([5c386fa](https://github.com/snowdreamtech/template/commit/5c386faae9715eeada52124c8315b71b7ad871de))
* **setup:** handle non-zero exit codes in smoke tests ([3fe6c05](https://github.com/snowdreamtech/template/commit/3fe6c054d4609b2263399b5df10d10f280919c2c))
* **setup:** improve network timeout reliability for large binary downloads ([432f092](https://github.com/snowdreamtech/template/commit/432f09206b8e2cad2a0aa148233c4975547a02b0))
* **setup:** normalize version strings in is_version_match to handle v prefix ([e5efefd](https://github.com/snowdreamtech/template/commit/e5efefdf81ca728f3b75703007cd475c6a2fa0f1))
* **setup:** use full provider path for checkmake installation ([1e40c13](https://github.com/snowdreamtech/template/commit/1e40c1347e7ac53de9419198323d3a7ab69d8435))
* **setup:** use full provider path for gitleaks installation ([7577489](https://github.com/snowdreamtech/template/commit/757748915a041537148f71085f9e377731a00dfa))
* **setup:** use mise which for binary verification in CI ([ded29c1](https://github.com/snowdreamtech/template/commit/ded29c17cdd851f46808a02ad086264f1038eb71))
* **setup:** verify tool executability in CI, not just mise registration ([5404189](https://github.com/snowdreamtech/template/commit/5404189d1f65c4064fdd773764c93e004af550dc))
* **shell:** add error handling for shell tool installation in CI ([967b8bf](https://github.com/snowdreamtech/template/commit/967b8bf69800365810dc023fd44db8d1ef3b9479))
* **shell:** add version specifications to mise install calls ([b576a7f](https://github.com/snowdreamtech/template/commit/b576a7f3c21726a3efeb6b001c1ec05332900aac))
* **shell:** remove duplicate log_summary line causing syntax error ([d8f75d9](https://github.com/snowdreamtech/template/commit/d8f75d9078c2a05b41a7a1762c1fc093030c5f22))
* **shell:** restore missing functions and add aggressive cache refresh ([f89f7f0](https://github.com/snowdreamtech/template/commit/f89f7f0913ecb6e165f8072d6b88f985b672dbcb))
* skip executable permission check on Windows (POSIX compliant) ([e365e6a](https://github.com/snowdreamtech/template/commit/e365e6aa58b67f08d058cc0a5853c8991e2e71ef))
* **sql:** add version specification to mise install call ([cde5de5](https://github.com/snowdreamtech/template/commit/cde5de54c3a705463cc0c01cf94e0da170ebed06))
* sync GITHUB_PATH to current shell for same-step tool availability ([e51048d](https://github.com/snowdreamtech/template/commit/e51048d1fdfeb320ad9922a4efd162d2da4c2821))
* **terraform:** add version specification to mise install call ([b99fa5a](https://github.com/snowdreamtech/template/commit/b99fa5a5be116a53300620261fcc5471b6aa092f))
* **test:** enforce mock path priority in check-env.bats ([babd8df](https://github.com/snowdreamtech/template/commit/babd8df6dfd551800ce3b1d4599e36a5af96b7c5))
* **testing:** add version specification to mise install call ([e4f4820](https://github.com/snowdreamtech/template/commit/e4f48202b510007a926776884945bac35c64b742))
* **toolchain:** extend ALF to handle full mise install with go: tools ([fec200c](https://github.com/snowdreamtech/template/commit/fec200c1b6f746507c2e7318eb88b1a88f159afb))
* **toolchain:** lazy-load mise cache and harden BATS detection ([0e2fb32](https://github.com/snowdreamtech/template/commit/0e2fb32dceae867bbdf55c4232796c0a818652a6))
* **toolchain:** migrate addlicense to github backend to support locked mode ([cf8ea8b](https://github.com/snowdreamtech/template/commit/cf8ea8b9611c39822f440ce6ebc4bcdbd18ac80d))
* **toolchain:** optimize summary logging and resolve BATS test failures ([f9639cb](https://github.com/snowdreamtech/template/commit/f9639cb97d22b9babc8c441c976ab1b802a11b55))
* **tools:** complete taplo provider migration to npm ([cdeaf76](https://github.com/snowdreamtech/template/commit/cdeaf762d62dfde4a14d90136cf298fe4c8e125d))
* **tools:** switch taplo from GitHub source to npm precompiled binary ([18de234](https://github.com/snowdreamtech/template/commit/18de2342744562645b849d24edbd94e6524259d1))
* update VER_MISE to 2026.4.0 to avoid RelativeUrlWithoutBase bug ([f93b0b3](https://github.com/snowdreamtech/template/commit/f93b0b39fd293b7e4e9eb4b97169f80eb38aa6eb))
* **versions:** update checkmake version to v0.3.2 ([6d1edd8](https://github.com/snowdreamtech/template/commit/6d1edd8a6299af947daf989f9badafe6efcf8f3e))
* **windows:** check root directory in verify_tool_atomic fallback ([9e2ace6](https://github.com/snowdreamtech/template/commit/9e2ace6e8cf960c25f159e53ec5b3a2d9cce3904))
* **windows:** convert Unix paths to Windows format for GITHUB_PATH ([de515d6](https://github.com/snowdreamtech/template/commit/de515d6aa91c1392fea55529a2c40550499a721c))
* **windows:** improve mise path detection for Git Bash environments ([d5a5337](https://github.com/snowdreamtech/template/commit/d5a5337ff038f35428034a2969842f99851c6297))
* **windows:** replace log_info with echo for mise path logging ([0bd8ee1](https://github.com/snowdreamtech/template/commit/0bd8ee1a951a0491649188430bb3c99be3067a93))
* **windows:** replace log_info with echo in PATH persistence ([7710710](https://github.com/snowdreamtech/template/commit/77107107bd15904198cef7146fdfb923192c160d))
* **windows:** skip command check for binaries without .exe extension ([a5ed6ca](https://github.com/snowdreamtech/template/commit/a5ed6caebb29aa8555d704cb1d8dabf4ea384111))
* **windows:** skip executable permission checks in lint-wrapper ([035ae54](https://github.com/snowdreamtech/template/commit/035ae54f20404a7c66a1a07459efb1d77678ab3f))
* **windows:** split executable check to avoid test evaluation on Windows ([cc76373](https://github.com/snowdreamtech/template/commit/cc763732a359bd52681a19b277854ea6bbe826e2))
* **windows:** use Unix-style paths for GITHUB_PATH in bash shell ([4eacf60](https://github.com/snowdreamtech/template/commit/4eacf6088d780d4896a533f3fe6acc864e65a56f))
* **windows:** write both Unix and Windows path formats to GITHUB_PATH ([62a0acf](https://github.com/snowdreamtech/template/commit/62a0acf12d04b0bd49d0e53d0bc45d2b261426c0))
* **workflows:** convert Windows LOCALAPPDATA path for Git Bash ([ba9192a](https://github.com/snowdreamtech/template/commit/ba9192a8013a8956d9d925227bd48b88f3703fcd))
* **zizmor:** ensure zizmor installs in CI regardless of workflow files presence ([9f6f620](https://github.com/snowdreamtech/template/commit/9f6f620ba69bc70304dbd7873ae9db2b5944ba76))
* **zizmor:** force reinstall in CI to guard against stale cache ([70597c2](https://github.com/snowdreamtech/template/commit/70597c227f45c9e7c0423768fc8d74323260f05b))
* **zizmor:** update repository from woodruffw to zizmorcore ([316a7cc](https://github.com/snowdreamtech/template/commit/316a7ccf6e3903022301ab3e3f11a4b866ed9484))


### Performance Improvements

* **infra:** implement lazy loading and unified refresh for mise cache ([f9e7aec](https://github.com/snowdreamtech/template/commit/f9e7aec490fb596525d362a79ef59726fc57bc90))
* **mise:** increase HTTP timeout to 300s for improved network reliability ([3335482](https://github.com/snowdreamtech/template/commit/3335482f99f98272149a30e831de066c11db87d2))
* **setup:** skip goreleaser installation in local dev to prevent setup hangs ([c95f801](https://github.com/snowdreamtech/template/commit/c95f80149c7440797995b872a27e3813079038bd))


### Documentation

* add comprehensive token usage audit report ([9983421](https://github.com/snowdreamtech/template/commit/9983421492acc679c2e27a2b01cdf3a6909c7ab9))
* **security:** add comprehensive asdf supply chain risk analysis ([515eea1](https://github.com/snowdreamtech/template/commit/515eea10cedc50f05f0cecd3f421129e1ce59cd4))
* **security:** add mise supply chain security analysis and mitigation ([c0c72f8](https://github.com/snowdreamtech/template/commit/c0c72f889b691ca3ed847e3b11dd3576c1586653))


### Miscellaneous Chores

* **deps:** bump mise to v2026.4.14 ([2bf29e3](https://github.com/snowdreamtech/template/commit/2bf29e3e3f8385913f9367de8a1170a151b5bd82))
* **main:** release 0.0.4 ([50f54df](https://github.com/snowdreamtech/template/commit/50f54df242ad2f1b3eb9af763359692728e38200))
* release 0.6.1 ([f6fc042](https://github.com/snowdreamtech/template/commit/f6fc042cad7d1c4991a20657655bc4b6b339d0d9))
* release 0.7.1 ([5535492](https://github.com/snowdreamtech/template/commit/5535492160f3525dff06ff9f0c6d78147467bed3))
* **release:** v0.4.0 - Fix Dependabot docker-compose detection ([e91f7d8](https://github.com/snowdreamtech/template/commit/e91f7d882f3c7b23260f4da02f0e5e53d6399968))
* remove temporary token audit report ([50a73c7](https://github.com/snowdreamtech/template/commit/50a73c7f88dbcf56332220a5941f6ca00f994bcd))


### Code Refactoring

* **ci:** standardize token usage across all workflows ([54b8445](https://github.com/snowdreamtech/template/commit/54b84456eb13b77b79cbbea32510405ec6f616b7))


### Continuous Integration

* **dependabot:** add path monitoring for mise toolchain config ([ebc7227](https://github.com/snowdreamtech/template/commit/ebc72273a85cb32f3e66b84305c726e47ce55b63))
* **deps:** remove MISE_SKIP_CHECKSUM workaround ([0a80645](https://github.com/snowdreamtech/template/commit/0a8064500cc0ecef8851af4aea31023f5809512e))

## [0.7.0](https://github.com/snowdreamtech/template/compare/v0.6.1...v0.7.0) (2026-04-18)


### Features

* **ci:** add centralized Harden Runner endpoints configuration ([5389a6e](https://github.com/snowdreamtech/template/commit/5389a6eb811a6b1ecfeefd6a0a3512e9a7d8b078))
* **ci:** sync Harden Runner endpoints from centralized config ([48b4221](https://github.com/snowdreamtech/template/commit/48b42213ee409890c4dd9a64f230c251dbac7b77))
* **docker:** add docker-compose file detection for hadolint and dockerfile-utils ([836880d](https://github.com/snowdreamtech/template/commit/836880de8c86ea0e548ed9edabe589d38357903f))
* **make:** add sync-harden-runner target for workflow endpoint management ([d24708d](https://github.com/snowdreamtech/template/commit/d24708dd86bb09946f3337dfd0747ca5fbcd51bc))
* **security:** add sigstore wildcard endpoint to harden runner ([2e26b98](https://github.com/snowdreamtech/template/commit/2e26b98d181cc8e4c739febde72e90fcfdee99da))


### Bug Fixes

* **ci:** add missing endpoints for trivy and sigstore ([115a19d](https://github.com/snowdreamtech/template/commit/115a19de02c7cd1d274ecb0191bd035a02d273ef))
* **ci:** restrict release-please to main branch only ([0db8a91](https://github.com/snowdreamtech/template/commit/0db8a910c4bbba1426944ccfcca648a593bb097e))
* **scripts:** ensure summary table displays by default ([a0d1b69](https://github.com/snowdreamtech/template/commit/a0d1b69300d151b2f448a2f1375e0f81e97cd6b2))
* **scripts:** install PyYAML via pip instead of pipx in sync-harden-runner ([16285a8](https://github.com/snowdreamtech/template/commit/16285a8ec6f27dada410103a3118482e38b82875))
* **scripts:** prevent sync-harden-runner from deleting workflow content ([32a4228](https://github.com/snowdreamtech/template/commit/32a4228e6e456a62cf761590491c5c56b6b24b8b))
* **security:** add docker.yml to zizmor ignore rules ([66e35cf](https://github.com/snowdreamtech/template/commit/66e35cf2e63c97fc8b1ecbad2a13e283a16031c7))
* **security:** add template-injection rule to zizmor config ([92396c1](https://github.com/snowdreamtech/template/commit/92396c15000fed537d3f4057f41dddd3bd7fe69b))

## [0.7.0](https://github.com/snowdreamtech/template/compare/v0.6.1...v0.7.0) (2026-04-18)


### Features

* **ci:** add centralized Harden Runner endpoints configuration ([5389a6e](https://github.com/snowdreamtech/template/commit/5389a6eb811a6b1ecfeefd6a0a3512e9a7d8b078))
* **ci:** sync Harden Runner endpoints from centralized config ([48b4221](https://github.com/snowdreamtech/template/commit/48b42213ee409890c4dd9a64f230c251dbac7b77))
* **docker:** add docker-compose file detection for hadolint and dockerfile-utils ([836880d](https://github.com/snowdreamtech/template/commit/836880de8c86ea0e548ed9edabe589d38357903f))
* **make:** add sync-harden-runner target for workflow endpoint management ([d24708d](https://github.com/snowdreamtech/template/commit/d24708dd86bb09946f3337dfd0747ca5fbcd51bc))


### Bug Fixes

* **ci:** add missing endpoints for trivy and sigstore ([115a19d](https://github.com/snowdreamtech/template/commit/115a19de02c7cd1d274ecb0191bd035a02d273ef))
* **ci:** restrict release-please to main branch only ([0db8a91](https://github.com/snowdreamtech/template/commit/0db8a910c4bbba1426944ccfcca648a593bb097e))
* **scripts:** ensure summary table displays by default ([a0d1b69](https://github.com/snowdreamtech/template/commit/a0d1b69300d151b2f448a2f1375e0f81e97cd6b2))
* **scripts:** install PyYAML via pip instead of pipx in sync-harden-runner ([16285a8](https://github.com/snowdreamtech/template/commit/16285a8ec6f27dada410103a3118482e38b82875))
* **scripts:** prevent sync-harden-runner from deleting workflow content ([32a4228](https://github.com/snowdreamtech/template/commit/32a4228e6e456a62cf761590491c5c56b6b24b8b))

## [0.6.1](https://github.com/snowdreamtech/template/compare/v0.6.0...v0.6.1) (2026-04-17)


### Bug Fixes

* **mise:** disable attestation checks in sync-lock to prevent false positives ([3f09708](https://github.com/snowdreamtech/template/commit/3f09708fefcf850b11f0ed2f7c29525f2c6c2b0d))
* **mise:** remove provenance fields from lockfile to prevent attestation errors ([f7790b5](https://github.com/snowdreamtech/template/commit/f7790b5803b47822c1191f28c155763a3ce43844))


### Miscellaneous Chores

* release 0.6.1 ([f6fc042](https://github.com/snowdreamtech/template/commit/f6fc042cad7d1c4991a20657655bc4b6b339d0d9))

## [0.6.0](https://github.com/snowdreamtech/template/compare/v0.5.0...v0.6.0) (2026-04-16)


### Features

* **scripts:** add tool to remove Windows wrapper scripts ([5bfd3ef](https://github.com/snowdreamtech/template/commit/5bfd3ef8ec9241cb09eca577d58616d1c29e2987))

## [0.5.0](https://github.com/snowdreamtech/template/compare/v0.4.0...v0.5.0) (2026-04-16)


### Features

* **ci:** improve release-please configuration and remove fixed version ([e1fcd64](https://github.com/snowdreamtech/template/commit/e1fcd648d91c306f6f385eca7600b3d6d2fefc5e))


### Bug Fixes

* **ci:** detect docker-compose files in dependabot generator ([530ca75](https://github.com/snowdreamtech/template/commit/530ca7534583ea16bb876c0e8bfbf209d958a42a))
* **ci:** explicitly specify checkmake mise tool spec to avoid aqua registry lookup ([acefebb](https://github.com/snowdreamtech/template/commit/acefebb98efc3563a9011f2ea05e7cc7274f5aef))
* **ci:** remove unsupported signoff parameter and add release-as 0.5.0 ([d01b2df](https://github.com/snowdreamtech/template/commit/d01b2dffa875d815316b5d41922b58ce9d574879))
* **mise:** workaround for GitHub attestation verification failures ([2911054](https://github.com/snowdreamtech/template/commit/29110547a86ebe51593eac637c0b4dc4c14977bc))


### Documentation

* **security:** add comprehensive asdf supply chain risk analysis ([515eea1](https://github.com/snowdreamtech/template/commit/515eea10cedc50f05f0cecd3f421129e1ce59cd4))
* **security:** add mise supply chain security analysis and mitigation ([c0c72f8](https://github.com/snowdreamtech/template/commit/c0c72f889b691ca3ed847e3b11dd3576c1586653))


### Miscellaneous Chores

* **deps:** bump mise to v2026.4.14 ([2bf29e3](https://github.com/snowdreamtech/template/commit/2bf29e3e3f8385913f9367de8a1170a151b5bd82))
* **release:** v0.4.0 - Fix Dependabot docker-compose detection ([e91f7d8](https://github.com/snowdreamtech/template/commit/e91f7d882f3c7b23260f4da02f0e5e53d6399968))


### Continuous Integration

* **dependabot:** add path monitoring for mise toolchain config ([ebc7227](https://github.com/snowdreamtech/template/commit/ebc72273a85cb32f3e66b84305c726e47ce55b63))
* **deps:** remove MISE_SKIP_CHECKSUM workaround ([0a80645](https://github.com/snowdreamtech/template/commit/0a8064500cc0ecef8851af4aea31023f5809512e))

## [0.4.0](https://github.com/snowdreamtech/template/compare/v0.4.0...v0.4.0) (2026-04-16)


### Bug Fixes

* **ci:** detect docker-compose files in dependabot generator ([530ca75](https://github.com/snowdreamtech/template/commit/530ca7534583ea16bb876c0e8bfbf209d958a42a))
* **ci:** explicitly specify checkmake mise tool spec to avoid aqua registry lookup ([acefebb](https://github.com/snowdreamtech/template/commit/acefebb98efc3563a9011f2ea05e7cc7274f5aef))
* **ci:** remove unsupported signoff parameter and add release-as 0.5.0 ([d01b2df](https://github.com/snowdreamtech/template/commit/d01b2dffa875d815316b5d41922b58ce9d574879))
* **mise:** workaround for GitHub attestation verification failures ([2911054](https://github.com/snowdreamtech/template/commit/29110547a86ebe51593eac637c0b4dc4c14977bc))


### Documentation

* **security:** add comprehensive asdf supply chain risk analysis ([515eea1](https://github.com/snowdreamtech/template/commit/515eea10cedc50f05f0cecd3f421129e1ce59cd4))


### Miscellaneous Chores

* **deps:** bump mise to v2026.4.14 ([2bf29e3](https://github.com/snowdreamtech/template/commit/2bf29e3e3f8385913f9367de8a1170a151b5bd82))
* **release:** v0.4.0 - Fix Dependabot docker-compose detection ([e91f7d8](https://github.com/snowdreamtech/template/commit/e91f7d882f3c7b23260f4da02f0e5e53d6399968))


### Continuous Integration

* **dependabot:** add path monitoring for mise toolchain config ([ebc7227](https://github.com/snowdreamtech/template/commit/ebc72273a85cb32f3e66b84305c726e47ce55b63))
* **deps:** remove MISE_SKIP_CHECKSUM workaround ([0a80645](https://github.com/snowdreamtech/template/commit/0a8064500cc0ecef8851af4aea31023f5809512e))

## [0.4.0](https://github.com/snowdreamtech/template/compare/v0.4.0...v0.4.0) (2026-04-16)


### Bug Fixes

* **mise:** workaround for GitHub attestation verification failures ([2911054](https://github.com/snowdreamtech/template/commit/29110547a86ebe51593eac637c0b4dc4c14977bc))


### Documentation

* **security:** add comprehensive asdf supply chain risk analysis ([515eea1](https://github.com/snowdreamtech/template/commit/515eea10cedc50f05f0cecd3f421129e1ce59cd4))

## [0.4.0](https://github.com/snowdreamtech/template/compare/v0.3.0...v0.4.0) (2026-04-15)


### Bug Fixes

* **ci:** temporarily disable npm-pnpm-audit hook due to API compatibility ([6ed58d9](https://github.com/snowdreamtech/template/commit/6ed58d92ebd50ff746dfe5f1d630f59cc7eb4441))
* **ci:** use npm instead of pnpm for audit to avoid API compatibility issues ([641c9f7](https://github.com/snowdreamtech/template/commit/641c9f744927c5aa15369cd56bcb7e31bb44bf3d))

## [0.3.0](https://github.com/snowdreamtech/template/compare/v0.2.0...v0.3.0) (2026-04-15)


### ⚠ BREAKING CHANGES

* **security:** mise now downloads binaries ONLY from GitHub Releases

### Features

* **dev:** add POSIX-compatible CI simulation script ([a587a39](https://github.com/snowdreamtech/template/commit/a587a390faa640adce2d448b773222cbb373c41d))
* **security:** implement three-layer defense against Aqua Registry ([ef13665](https://github.com/snowdreamtech/template/commit/ef136651dc6d7d358cf4de2748a0a10a841730bf))


### Bug Fixes

* **ci:** update node-audit to use new npm bulk advisory endpoint ([526ca79](https://github.com/snowdreamtech/template/commit/526ca79e307d38240e54505bda3b06d75f1e8b9a))
* **docs:** ensure make docs build works correctly ([6a0df9c](https://github.com/snowdreamtech/template/commit/6a0df9ce6acf22d2d8606dda0444c48d369967bd))

## [0.2.0](https://github.com/snowdreamtech/template/compare/v0.1.0...v0.2.0) (2026-04-14)


### Features

* **devcontainer:** add Docker availability check in init script ([7ca1fe4](https://github.com/snowdreamtech/template/commit/7ca1fe4c37ada1893e94e9ac28ec578952c1d99e))
* **mise:** add yamllint to core tools in mise.toml ([406bbb3](https://github.com/snowdreamtech/template/commit/406bbb32f3a9f47bdfee72e2422b273a6cab252b))


### Bug Fixes

* **devcontainer:** add command availability checks before usage ([a0a4ce0](https://github.com/snowdreamtech/template/commit/a0a4ce0d1bed246c722270f920e79f2046a1bfe7))
* **devcontainer:** add command to keep container running ([5484045](https://github.com/snowdreamtech/template/commit/54840456dfb76e040655d49cf28b45e9b64ead1d))
* **devcontainer:** add workspace volume mount and fix configuration ([9ff708b](https://github.com/snowdreamtech/template/commit/9ff708be2aa74af62eae40012f163229bb688296))
* **devcontainer:** add YAML document start marker for yamllint ([44ccf81](https://github.com/snowdreamtech/template/commit/44ccf8108ad6ea15b92069f15bbc1ce852723ec0))
* **devcontainer:** correct updateContentCommand to use make install ([774bba2](https://github.com/snowdreamtech/template/commit/774bba245b5025973492b2cd8fb10826eae7af12))
* **devcontainer:** handle missing host directories gracefully ([9b79301](https://github.com/snowdreamtech/template/commit/9b7930189a20c8c44983a739d80ec05d8b5987f5))
* **devcontainer:** use global git config instead of local ([934bb1c](https://github.com/snowdreamtech/template/commit/934bb1c3bec475b571daab4168cdd233d10b6816))
* **mise:** update yamllint version to match versions.sh ([d9bcf61](https://github.com/snowdreamtech/template/commit/d9bcf61f7092f248186d41ceb66fcd98d3015a34))
* resolve devcontainer startup with docker-compose configuration ([fa788e4](https://github.com/snowdreamtech/template/commit/fa788e4f3c94e767f465eea9f17acb4ab8f4a01e))
* **setup:** ensure yaml and toml modules are always installed locally ([7ca7026](https://github.com/snowdreamtech/template/commit/7ca7026761c0a58a939935db83e37922a872a243))
* **tools:** complete taplo provider migration to npm ([cdeaf76](https://github.com/snowdreamtech/template/commit/cdeaf762d62dfde4a14d90136cf298fe4c8e125d))
* **tools:** switch taplo from GitHub source to npm precompiled binary ([18de234](https://github.com/snowdreamtech/template/commit/18de2342744562645b849d24edbd94e6524259d1))

## [0.1.0](https://github.com/snowdreamtech/template/compare/v0.0.4...v0.1.0) (2026-04-14)


### Features

* **devcontainer:** add comprehensive SSH and GPG permission configuration ([843180b](https://github.com/snowdreamtech/template/commit/843180b5722d20b1704e482d40b1c4a80f890fdc))
* **devcontainer:** add support for local git config file ([2fb5e43](https://github.com/snowdreamtech/template/commit/2fb5e435b5038617746aff9837dec98865cd4d5c))
* **devcontainer:** enable GPG signing support ([4d88e12](https://github.com/snowdreamtech/template/commit/4d88e125ddc84966237d382d897c03b1cf73da06))


### Bug Fixes

* **ci:** pin GitHub Actions to commit SHA in performance workflow ([13623d9](https://github.com/snowdreamtech/template/commit/13623d9cf5ae7a6677a8645a8d568bb39a2752a2))
* **devcontainer:** make git config cross-platform compatible ([600052a](https://github.com/snowdreamtech/template/commit/600052ae9d9067de738a875009df7e8371099625))
* **devcontainer:** use dynamic workspace folder variable ([73421d6](https://github.com/snowdreamtech/template/commit/73421d6f6724ae63613d30566103b55191d43c87))
* **devcontainer:** use find instead of ls for directory permissions ([ffeaee8](https://github.com/snowdreamtech/template/commit/ffeaee833e3a9f4ad7a88d45b906ad12ca52b62d))
* **setup:** normalize version strings in is_version_match to handle v prefix ([e5efefd](https://github.com/snowdreamtech/template/commit/e5efefdf81ca728f3b75703007cd475c6a2fa0f1))

## [0.0.4](https://github.com/snowdreamtech/template/compare/v0.0.3...v0.0.4) (2026-04-12)


### Bug Fixes

* **python:** install build dependencies for Alpine/musl environments ([3f53883](https://github.com/snowdreamtech/template/commit/3f53883e79ed5b9d74ca2369052b0014a3d4f0fc))


### Miscellaneous Chores

* **main:** release 0.0.4 ([50f54df](https://github.com/snowdreamtech/template/commit/50f54df242ad2f1b3eb9af763359692728e38200))

## [0.0.4](https://github.com/snowdreamtech/template/compare/v0.0.3...v0.0.4) (2026-04-12)


### Bug Fixes

* **python:** install build dependencies for Alpine/musl environments ([3f53883](https://github.com/snowdreamtech/template/commit/3f53883e79ed5b9d74ca2369052b0014a3d4f0fc))

## [0.0.4](https://github.com/snowdreamtech/template/compare/v0.0.3...v0.0.4) (2026-04-12)


### Bug Fixes

* **python:** install build dependencies for Alpine/musl environments ([3f53883](https://github.com/snowdreamtech/template/commit/3f53883e79ed5b9d74ca2369052b0014a3d4f0fc))

## [0.0.3](https://github.com/snowdreamtech/template/compare/v0.0.2...v0.0.3) (2026-04-12)


### Bug Fixes

* **ci:** exclude unreleased version compare links from lychee checks ([6165757](https://github.com/snowdreamtech/template/commit/6165757ebf73b2d1b7abf5c738d66f6402a06f14))
* **ci:** make link checker non-blocking to avoid false failures ([99c2ec9](https://github.com/snowdreamtech/template/commit/99c2ec99befd980c996953bd53471ded5f6d404a))
* **ci:** pass GITHUB_TOKEN to lychee for link checking ([dfb9fd1](https://github.com/snowdreamtech/template/commit/dfb9fd185e4f1bbe4987b4e991db85a4a03639e8))
* **ci:** use valid log level for lychee verbose option ([171e721](https://github.com/snowdreamtech/template/commit/171e7218eb1e6cf00645c27f7d9d8187e5c9bf47))
* **docs:** update TPGi CCA link to GitHub releases page ([b061ca1](https://github.com/snowdreamtech/template/commit/b061ca1150e4dd23b47fd88b3f9992d0f97c1df8))

## [0.0.2](https://github.com/snowdreamtech/template/compare/v0.0.1...v0.0.2) (2026-04-12)


### Bug Fixes

* **node:** auto-install bash in Alpine for npm compatibility ([0042fa2](https://github.com/snowdreamtech/template/commit/0042fa2abb6ae389277660b3d0656f54ea669643))
* **node:** configure mise for musl binaries in Alpine environments ([232bce2](https://github.com/snowdreamtech/template/commit/232bce29a4401ebe42fd529c448a47d404562abe))
* **node:** explicitly set MISE_NODE_* env vars for Alpine ([25d506d](https://github.com/snowdreamtech/template/commit/25d506d6d952fa479dca79e427d10905d37fc9b6))
* **node:** export ALPINE_VERSION for mise.toml template evaluation ([62d9344](https://github.com/snowdreamtech/template/commit/62d934444f440eec678ac2227c723ab7f1e46e8d))
* **node:** use boolean false for node.compile setting ([97d5bd5](https://github.com/snowdreamtech/template/commit/97d5bd57d4dc1ffc52a777524c5aabf92a9a061d))

## 0.0.1 (2026-04-11)


### ⚠ BREAKING CHANGES

* mise does NOT automatically detect musl

### Features

* add comprehensive performance testing and documentation infrastructure ([c1d8996](https://github.com/snowdreamtech/template/commit/c1d8996e4f44ce7222cf589812325b700e7e5f96))
* add missing performance and documentation scripts ([fe65a58](https://github.com/snowdreamtech/template/commit/fe65a58f25c3214e93d156a03079c51d169d3240))
* add VER_MISE to versions.sh for centralized version management ([b805d26](https://github.com/snowdreamtech/template/commit/b805d26395103dbea4c34e14b5c99521537dabbd))
* **audit:** enforce mandatory version verification for Trivy ([e5bed3c](https://github.com/snowdreamtech/template/commit/e5bed3c75a8baf19f895b511dcd41329292ca0c2))
* **ci:** add atomic tool verification for shfmt and editorconfig-checker ([dd255e5](https://github.com/snowdreamtech/template/commit/dd255e5956a3a9b96c15cc832fddc3667a3effb2))
* **ci:** add atomic verification for config/doc linting tools ([5d8de8b](https://github.com/snowdreamtech/template/commit/5d8de8bc179790eac4518f73f03d5f6e8d5c5af6))
* **ci:** add atomic verification for hadolint and dockerfile-utils ([9e5a824](https://github.com/snowdreamtech/template/commit/9e5a824b22e1b31bcf08c9e7ae265ec836077bb0))
* **ci:** add atomic verification for Node.js tools ([4a6c20f](https://github.com/snowdreamtech/template/commit/4a6c20f4756f2605971c12a7a12feabc1d56bad4))
* **ci:** enable release-please for all configured branches ([bc94753](https://github.com/snowdreamtech/template/commit/bc94753270413815bb340e27fd6154f41e37d771))
* **ci:** harden cd.yml triggers, pathing, and link checking ([f33b0ec](https://github.com/snowdreamtech/template/commit/f33b0ec2e5da1f8b26ec808b24eafbf04c8d10c6))
* **ci:** harden stateless toolchain and remediate security audit findings ([ebab039](https://github.com/snowdreamtech/template/commit/ebab039b1f58b3a782bec8a71d23596b1c59f2ce))
* **ci:** implement universal binary-first tool verification ([5bedab1](https://github.com/snowdreamtech/template/commit/5bedab16aba4891d0395ba2e0462148810c8b2ac))
* **ci:** restore dev parity, fix windows path, and use relative doc links ([e348693](https://github.com/snowdreamtech/template/commit/e34869356ceb8e9b721cbe83f02e152ff07559f8))
* **ci:** skip DCO check for bot commits ([b841cee](https://github.com/snowdreamtech/template/commit/b841cee3308b970f945f92692d807cf353a7b379))
* **common:** enhance mise backend dependency checks and add MISE_GITHUB_ENTERPRISE_TOKEN forwarding ([27e0f1e](https://github.com/snowdreamtech/template/commit/27e0f1e7dfd3758bfa4bdc8013410a4bdbec386e))
* **devcontainer:** harden supply chain and eliminate insecure downloads ([6e44d86](https://github.com/snowdreamtech/template/commit/6e44d869573c265abc0f43f356f2a61e58a656e3))
* enhance PATH management for dynamically installed tools ([f307921](https://github.com/snowdreamtech/template/commit/f30792120c89c79cf2c7216b3ca5cfbe9d22dd31))
* enhanced PATH management for dynamic tool installation ([43c8ba6](https://github.com/snowdreamtech/template/commit/43c8ba6cd306cbe387593481dc6d7c28357557db))
* **hooks:** strengthen pre-commit with security & deep linting audits ([1427366](https://github.com/snowdreamtech/template/commit/14273660790f2c9f3806528114d9faec53c1f26b))
* **makefile:** add update-tools target and link to update lifecycle ([85e0381](https://github.com/snowdreamtech/template/commit/85e0381f64eb3992b2e5e47a3535b7bbe654a0da))
* **mise:** add Windows wrappers for manifest aggregator ([a851c75](https://github.com/snowdreamtech/template/commit/a851c75ffe0e11718ef34f2e81b3c2e51c689ade))
* **mise:** expand lock ritual to explicitly support alpine (musl) and debian (glibc) ([5be8bd6](https://github.com/snowdreamtech/template/commit/5be8bd674d31214765ebcd99a5445472a7859dff))
* **mise:** harden sync-lock with multi-platform support and exhaustive tool listing ([ba97b8d](https://github.com/snowdreamtech/template/commit/ba97b8de349ac803c330727e6008079b29a6612e))
* **mise:** implement ALF for go providers and revert osv-scanner for network compatibility ([546faff](https://github.com/snowdreamtech/template/commit/546faffef385d80fa6a290bb68652b7d16ed1528))
* **mise:** implement tiered tool management via lock ritual ([4c69d00](https://github.com/snowdreamtech/template/commit/4c69d00ec01fb9b98c786a4244c3956c7ca1a093))
* **mise:** implement universal tiered strategy with manifest aggregator ([73da439](https://github.com/snowdreamtech/template/commit/73da439bfea0462cf7e9369e1a0be1bafd81eae8))
* **mise:** standardize sync-lock with cmd-ps1-shell delegation pattern ([fe64c53](https://github.com/snowdreamtech/template/commit/fe64c53bbcdfcc596551b353675a413f9fb3618d))
* optimize mise shell activation with full path and official method ([77b21a6](https://github.com/snowdreamtech/template/commit/77b21a6257b5b219bfb5dc4057d28ffaeb2f0c82))
* optimize mise shell activation with full path support ([e71c33f](https://github.com/snowdreamtech/template/commit/e71c33f88abb2a2fd11561f7067b5c1e479f2b43))
* **scripts:** add cleanup script for duplicate mise activation lines ([b7020c6](https://github.com/snowdreamtech/template/commit/b7020c6ca19e167c71b776f1380c5f1dc8c20c3c))
* **scripts:** add missing performance and documentation scripts ([f8d9c61](https://github.com/snowdreamtech/template/commit/f8d9c613cfa599e503bed8190e13da66cbf928d7))
* **scripts:** exhaustive global safety hardening for all shell scripts ([86c199e](https://github.com/snowdreamtech/template/commit/86c199e77ddc95a6190fbf6e758ec640d32bc6a3))
* **scripts:** exhaustive semantic safety audit and local variable hardening ([76dc9a7](https://github.com/snowdreamtech/template/commit/76dc9a7be77ecc2cdcb7076af591da2ce5b5ae92))
* **scripts:** final project-wide audit and reinforcement of set -eu standard ([cf99842](https://github.com/snowdreamtech/template/commit/cf99842599544788770df3c55ceccd17a1e7165f))
* **scripts:** global recursive shell safety upgrade to set -eu ([d2b90e3](https://github.com/snowdreamtech/template/commit/d2b90e3232dd774d629a98dda941307e3681ea07))
* **scripts:** implement Node.js JSON parser with CommonJS ([e0ffb06](https://github.com/snowdreamtech/template/commit/e0ffb06a5e8634f05d171a440aa515c881180d30))
* **scripts:** unify all shell scripts to use set -eu for POSIX safety ([fda9d45](https://github.com/snowdreamtech/template/commit/fda9d45628b92a5a654f16ff354d37493ae65e8b))
* **security:** globally lock GitHub Actions to latest verified SHAs ([193eb0c](https://github.com/snowdreamtech/template/commit/193eb0cd997c8d19bb0d86c7f811cf99f332cde2))
* **security:** implement artifact signing with cosign and SBOM auditing ([df86f8f](https://github.com/snowdreamtech/template/commit/df86f8f092631da18eb93148674f623b59f32b99))
* **security:** implement binary artifact audit to prevent poisoning ([73cd7b3](https://github.com/snowdreamtech/template/commit/73cd7b399cd3ffd65c0134ea08f13ac587b89501))
* **security:** implement OpenSSF Scorecard for continuous security health monitoring ([8c5015a](https://github.com/snowdreamtech/template/commit/8c5015a67693d941cb3f893b555b515c533acd18))
* **security:** implement platinum-standard CI/CD supply chain hardening ([d101fe4](https://github.com/snowdreamtech/template/commit/d101fe4988cc532d5fa4efa1ae8b883da7314b0e))
* **security:** implement PR dependency review shield ([16fd84e](https://github.com/snowdreamtech/template/commit/16fd84e627d0f8bbaed92522d07bf345e3aa6dc7))
* **security:** integrate CycloneDX SBOM generation into audit script ([5e9598a](https://github.com/snowdreamtech/template/commit/5e9598a387f2f909d32605903bb1730e7e731096))
* **security:** refactor audit orchestration for automatic activation in local dev ([f085c71](https://github.com/snowdreamtech/template/commit/f085c711f6e3b5093b1c8ab9efff2b31e32725d1))
* use mise.jdx.dev install script with version support ([043b346](https://github.com/snowdreamtech/template/commit/043b3460ae67ae946018155cab31396770741561))
* use VER_MISE from versions.sh in common.sh ([d4c8160](https://github.com/snowdreamtech/template/commit/d4c81601aa888932a2b0b01ef3f9208f348b5a4e))


### Bug Fixes

* add pattern matching for binaries with version/platform suffixes ([9ce22d9](https://github.com/snowdreamtech/template/commit/9ce22d9e618960d62c4711bea1c9130d5d82f3df))
* add unified PATH management and CI persistence to run_mise ([19c6431](https://github.com/snowdreamtech/template/commit/19c64317dc58d31ddeb16ef4f7657795fe49ef35))
* align log_summary to use CI_STEP_SUMMARY for consistent reporting ([601ac65](https://github.com/snowdreamtech/template/commit/601ac65d390330fa51a7e2c025b2a56ac3834188))
* **audit:** handle missing GITHUB_TOKEN and make zizmor findings non-fatal ([28c7cc6](https://github.com/snowdreamtech/template/commit/28c7cc618a6dd109e6f27ff8fb07117aba96f409))
* **base:** add version specifications to mise install calls ([71aa2fe](https://github.com/snowdreamtech/template/commit/71aa2fe0746f8fdfc8178de4971abaf002509108))
* **bootstrap:** prevent duplicate mise activation lines in shell RC files ([0f343f1](https://github.com/snowdreamtech/template/commit/0f343f1fb933646753e52afead32f0ef3d887099))
* **bootstrap:** prevent shell-specific syntax errors in POSIX sh ([cc6066d](https://github.com/snowdreamtech/template/commit/cc6066d1004c52e9ecb60892df1a2f10645a414f))
* **bootstrap:** use dynamic mise paths instead of hardcoded values ([062f35a](https://github.com/snowdreamtech/template/commit/062f35ac1900d580735f094e86f55befba2d3400))
* **cd:** add debug output and improve PATH setup for Windows ([892d860](https://github.com/snowdreamtech/template/commit/892d8608c81b7c2c27532a7cbf02b0d06792057a))
* **cd:** align environment and checkout strategy with CI ([cf21294](https://github.com/snowdreamtech/template/commit/cf21294c59649198f7d37ceaea8a3cef60d7b4b9))
* **cd:** complete debug output alignment with CI workflow ([b52cd4f](https://github.com/snowdreamtech/template/commit/b52cd4f93ce813f0a2bea3d135ff81bfc761f4e6))
* **cd:** elevate release-please token for automation triggering ([64d63ab](https://github.com/snowdreamtech/template/commit/64d63ab6a2a2d9840009bb141d9015a6d69ed62d))
* **cd:** ensure mise shims are on PATH for all verification steps ([95687fe](https://github.com/snowdreamtech/template/commit/95687fe0f88e57b638e5d273be8b5bbd528ed760))
* **cd:** optimize release-please triggers for main branch ([1bc2e26](https://github.com/snowdreamtech/template/commit/1bc2e26d77a5e189a202f124057b3db0af2f2928))
* **cd:** split monolithic verify into explicit lint, test, audit ([049ba15](https://github.com/snowdreamtech/template/commit/049ba150dc0efdc3aaeca1a382beefe60fc8006c))
* **cd:** synchronize commit convention check from CI ([43f3752](https://github.com/snowdreamtech/template/commit/43f3752fba68ad3a3f90eb27eeebc1bb9c939b8d))
* **cd:** synchronize security audit depth with CI ([9757128](https://github.com/snowdreamtech/template/commit/9757128fe122be2acb0bed34f2e39d5892778b9b))
* centralize hardcoded provider values across shell scripts ([2a170e7](https://github.com/snowdreamtech/template/commit/2a170e7626c456e116c59e08eff7e49e324fa551))
* **check-env:** strip leading 'v' from versions for accurate comparison ([83456b3](https://github.com/snowdreamtech/template/commit/83456b3f8624aaf6cc2b3feee48c398102d16ed8))
* **check-env:** sync GITHUB_PATH to current shell in CI ([8fdde4e](https://github.com/snowdreamtech/template/commit/8fdde4e89bd198a7c5f8d78a396c27bae3abeae6))
* **check-env:** use full mise keys for version detection ([a5b1063](https://github.com/snowdreamtech/template/commit/a5b10635d40cde00554db96ec953025aeae7f5a6))
* **check-env:** version mismatch should warn not fail ([e98d7b3](https://github.com/snowdreamtech/template/commit/e98d7b3d973c69847f5f817fef3a374ab821a14c))
* **ci:** add aggressive cache refresh after uninstall in install_tool_safe ([433899c](https://github.com/snowdreamtech/template/commit/433899cd8435f737a6193f15e3189a987260e5ef))
* **ci:** add DCO signoff to dependabot-sync commits ([cfe7546](https://github.com/snowdreamtech/template/commit/cfe75468ce43c5b61c7ac8e5915eb767b0032815))
* **ci:** add DCO signoff to release-please commits ([8fd8b67](https://github.com/snowdreamtech/template/commit/8fd8b6729aa68027072f68acec9735397eea54cc))
* **ci:** add mise exec fallback for pnpm audit on Windows ([f1e93b7](https://github.com/snowdreamtech/template/commit/f1e93b73541e2a1f5c9c42c81d89e97cbf91c57d))
* **ci:** add mise setup to label-sync.yml and clean up version comments ([70397a3](https://github.com/snowdreamtech/template/commit/70397a3318f889ed66bc97b9c8aed29827f287cb))
* **ci:** add mise.lock and versions.sh to mise cache key in pages workflow ([9d7b9d7](https://github.com/snowdreamtech/template/commit/9d7b9d75d68645afaa798a577758ea55fa94cb03))
* **ci:** add mise.lock to mise cache key in cd workflow ([d43cdf9](https://github.com/snowdreamtech/template/commit/d43cdf9ab8d2c22af779601acfbab8fe42b035be))
* **ci:** add mise.lock to mise cache key in dependabot-sync workflow ([070adf0](https://github.com/snowdreamtech/template/commit/070adf0610122d8ed3e098d72838714190ec8871))
* **ci:** add mise.lock to mise cache key in label-sync workflow ([fccb9f0](https://github.com/snowdreamtech/template/commit/fccb9f0669b118ecaeac1e88a771534a13fc8455))
* **ci:** add mise.lock to mise cache keys in ci workflow (3 jobs) ([ac8ecf6](https://github.com/snowdreamtech/template/commit/ac8ecf6adb81781556c969916c2e0a62b99f035c))
* **ci:** add node-audit special handling in CI fallback ([d82624c](https://github.com/snowdreamtech/template/commit/d82624cda69bbf48e21ef2693d5c5477c278463e))
* **ci:** add OSV-scanner and Zizmor to Tier 1 tools ([7bf9561](https://github.com/snowdreamtech/template/commit/7bf9561f2459bd68ff422d61e3e8967df5a9d65f))
* **ci:** comprehensive tool installation and execution fix ([bac4a9c](https://github.com/snowdreamtech/template/commit/bac4a9ce77e15170d3a8f3e1db970b00917e7abc))
* **ci:** correct action SHA and version tag mismatches in cd.yml ([f4d90ec](https://github.com/snowdreamtech/template/commit/f4d90ec5d3570189d0f8c5562f4eea59f70a899f))
* **ci:** disable paranoid mode in CI for lockfile sync ([1b282ce](https://github.com/snowdreamtech/template/commit/1b282ce6fa3d1ecca5c18de89a4962ec7e50226b))
* **ci:** elevate GITHUB_TOKEN for rate-limit resilience ([7c2123f](https://github.com/snowdreamtech/template/commit/7c2123fe8f1dbced5332f7a16cdbb84c973b6c29))
* **ci:** enable release-please for dev branch and fix DCO signoff ([ab8f065](https://github.com/snowdreamtech/template/commit/ab8f065dce45cc794d78aa075fc8c066eeec2bf5))
* **ci:** ensure CI_STEP_SUMMARY is always defined and sourced in gen-dependabot ([e19a7ec](https://github.com/snowdreamtech/template/commit/e19a7ec5fa7c2c61f9cdccb0dae4ddc56afcbfc4))
* **ci:** ensure PATH persistence after setup completes on Windows ([460dc8d](https://github.com/snowdreamtech/template/commit/460dc8d84067a89081762401269565a815ceab00))
* **ci:** ensure project root is detected before initializing CI summary path ([341b3f8](https://github.com/snowdreamtech/template/commit/341b3f87ffbae0aee1d8bd56f294ed76a478b07c))
* **ci:** ensure shfmt and handle trivy absence in CI ([fae3f46](https://github.com/snowdreamtech/template/commit/fae3f46d2c253d3c2883814d24f33169df3325b2))
* **ci:** exhaustive project root detection to satisfy both smoke and bats tests ([efd9b07](https://github.com/snowdreamtech/template/commit/efd9b07d36a5234b3f00f8e3e49f01062225dc36))
* **ci:** extract version from TOML table syntax in install_tool_safe ([9740407](https://github.com/snowdreamtech/template/commit/9740407dd5b970e1bcecc26d58ce756f2fd36027))
* **ci:** fix PSScriptAnalyzer and yamllint configuration ([166826a](https://github.com/snowdreamtech/template/commit/166826afd22a16c5fe7ccf131ae28d6f46d48525))
* **ci:** fix yamllint truthy error in performance workflow ([7e7dfd0](https://github.com/snowdreamtech/template/commit/7e7dfd052065a8b361212d0a1611a60bc951b06b))
* **ci:** handle mise shims in verify_binary_exists ([bc0dfd9](https://github.com/snowdreamtech/template/commit/bc0dfd945090a3b63cf27ee90b9073547ff76ec5))
* **ci:** harden CI summary path with guaranteed fallback to avoid directory errors ([3d2c227](https://github.com/snowdreamtech/template/commit/3d2c2274990383a14b44ee96b96d34bb6f25757b))
* **ci:** harden GITHUB_PATH persistence and Windows mise detection ([6c555db](https://github.com/snowdreamtech/template/commit/6c555dbefb0842a70ab06769935edf1df33f58a7))
* **ci:** immediately update PATH in current shell after writing to GITHUB_PATH ([6fd7c76](https://github.com/snowdreamtech/template/commit/6fd7c764c16a49fd51f180c4310ad2bf08db7401))
* **ci:** implement robust sentinel-based project root detection ([7be109f](https://github.com/snowdreamtech/template/commit/7be109f9376961c2e4521ddd7b4997bcdaf29c2e))
* **ci:** improve binary resolution fallback for GitHub tools ([53ae080](https://github.com/snowdreamtech/template/commit/53ae0802fb5eb5d618f71be3f9f0fe6d9d89232a))
* **ci:** improve post-install binary name resolution ([46ab8b0](https://github.com/snowdreamtech/template/commit/46ab8b00c8b5f65bdd1e2171964650ef46359e75))
* **ci:** make lint.sh fail when pre-commit hooks fail ([78ca50a](https://github.com/snowdreamtech/template/commit/78ca50ac3a1e670d5c11b7a0292585976384cbfd))
* **ci:** optimize pull_request triggers for main branch ([5f5a5ca](https://github.com/snowdreamtech/template/commit/5f5a5ca742320b45f37aac09c4803c8f1d9e4fad))
* **ci:** pin CodeQL action to verified underlying commit SHA ([11cb79b](https://github.com/snowdreamtech/template/commit/11cb79b81eae47e66917c33adbf41e36bfe26983))
* **ci:** quote 'on' keyword in performance workflow ([4710acd](https://github.com/snowdreamtech/template/commit/4710acd942f39592bfe55f4ceb73af7ec47153c7))
* **ci:** remove --template flag from gh label list command ([27b29ea](https://github.com/snowdreamtech/template/commit/27b29ea1cabb48a35398324473ea3e8133f07ed3))
* **ci:** remove emoji from log messages to fix Windows printf errors ([f14a5f0](https://github.com/snowdreamtech/template/commit/f14a5f00870474761daa7a54a231a64cc70d668b))
* **ci:** remove MISE_OFFLINE=true from all workflow files ([4be4079](https://github.com/snowdreamtech/template/commit/4be407962334edee9c465e296d7e5d18bfc7d2ea))
* **ci:** remove SKIP_MODULES and stabilize test assertions ([241275f](https://github.com/snowdreamtech/template/commit/241275fa1ff08cb994b8c5dabdfb1a9b0ae1a538))
* **ci:** remove trailing space in goreleaser.yml tag pattern ([7cd8dd9](https://github.com/snowdreamtech/template/commit/7cd8dd9fbb69a81fb03ae0467538859b7effa4d8))
* **ci:** remove Windows path conversion for GitHub Actions PATH persistence ([4af9416](https://github.com/snowdreamtech/template/commit/4af9416b6ea4fba25606ef2bace671895d56e921))
* **ci:** resolve actual binary names for platform-specific tools ([e6856c9](https://github.com/snowdreamtech/template/commit/e6856c9a622d12a915744b44f1542106c87034ad))
* **ci:** resolve npm-pnpm-audit hook failure on Windows ([095f3a5](https://github.com/snowdreamtech/template/commit/095f3a553e352c426e39aa727b37cd3f610b3408))
* **ci:** restore actions/cache version hash in ci workflow ([713881a](https://github.com/snowdreamtech/template/commit/713881ae0bd8929ce69c02ae89482cd26f643e99))
* **ci:** restore actions/cache version hash in dependabot-sync workflow ([e506978](https://github.com/snowdreamtech/template/commit/e5069784c2269abf3965875306a42997d358a4f5))
* **ci:** restore actions/cache version hash in pages workflow ([80a625d](https://github.com/snowdreamtech/template/commit/80a625dcbbe95259fbaae26d59cbb61fc7e2539e))
* **ci:** restrict release-please to main branch only ([b97162a](https://github.com/snowdreamtech/template/commit/b97162aa0092875bb8af1820cb4db6c06d17665b))
* **ci:** revert goreleaser to use --version flag ([564d899](https://github.com/snowdreamtech/template/commit/564d899198c6f0c19872642d6b3e2968d54536ab))
* **ci:** specify bin names for shfmt and editorconfig-checker in mise.toml ([f249809](https://github.com/snowdreamtech/template/commit/f2498091f79349cfe84fe160ad78477ad28a22b5))
* **ci:** support binaries installed in root directory ([72631b4](https://github.com/snowdreamtech/template/commit/72631b4d0f2c3cd53ea50b5e42a64585a92e98fc))
* **ci:** suppress zizmor findings for performance.yml and skip flaky tests ([0616771](https://github.com/snowdreamtech/template/commit/06167717e490f73f34e49386e4a776907caa554c))
* **ci:** update actions/cache version hash in ci workflow ([36520ec](https://github.com/snowdreamtech/template/commit/36520ec4b721ba79ccf6b410573d02e416a53f29))
* **ci:** update Phase 1 tools to use 4-parameter verify_tool_atomic ([5af97ed](https://github.com/snowdreamtech/template/commit/5af97ed9daa59c3f2a900f25ff22dd2f7bf634bb))
* **ci:** use correct version command for goreleaser ([b271332](https://github.com/snowdreamtech/template/commit/b271332e9b02f2f34513943f86ce8297f3c31dc4))
* **ci:** use GitHub binary providers for cross-platform security tools ([8d150f6](https://github.com/snowdreamtech/template/commit/8d150f6554d8f6a752f7c2351e62f637a1f52dc3))
* **ci:** use GITHUB_TOKEN for release-please to enable DCO signoff ([51051ba](https://github.com/snowdreamtech/template/commit/51051ba4a2e7fdca3a9ed927e09fdec94b20f77d))
* **ci:** use mise exec for shim smoke tests ([12306a9](https://github.com/snowdreamtech/template/commit/12306a9fbf8a8f4192002424aa03898223628077))
* **ci:** use platform-specific binary name for editorconfig-checker ([2b38f2a](https://github.com/snowdreamtech/template/commit/2b38f2a2fbb5445b5b5118a64efa34bd56495023))
* **codeql:** optimize triggers for baseline shrinkage ([9eeca84](https://github.com/snowdreamtech/template/commit/9eeca8477fce10c5c4f8942ac7c60eb5e70efbcf))
* **codeql:** pin action to verified underlying commit SHA ([422a8c6](https://github.com/snowdreamtech/template/commit/422a8c6e9e1a3b5a095be9cd814d42c5fb33ee37))
* **common:** correct find -maxdepth position in has_lang_files ([9866fe6](https://github.com/snowdreamtech/template/commit/9866fe68a40c64f47a3fc2d1c7abf94dd5d10022))
* **config:** remove deprecated ExperimentalScannerConfig from osv-scanner ([ce385db](https://github.com/snowdreamtech/template/commit/ce385db7ca970de93405cb85988e73e32f8f8c90))
* correct Alpine Linux Node.js installation behavior ([bbd85fa](https://github.com/snowdreamtech/template/commit/bbd85fa7048b058ee6a2d13ccf27debfac4912e7))
* **cpp:** add version specification to mise install call ([b6fbdaf](https://github.com/snowdreamtech/template/commit/b6fbdaf1f4866b77b7c1a8376e82f7d8dd5c7cd9))
* **dco:** use PR author instead of branch name to prevent bypass ([eb55adf](https://github.com/snowdreamtech/template/commit/eb55adf2ccc05f6a0bea0a1f9c37c5fb56f08e02))
* **deps:** correct toolchain provider mappings and dependabot schema ([76d515d](https://github.com/snowdreamtech/template/commit/76d515d4c3b0abe40ed6cc849b1f3ed2077887bb))
* **deps:** remove invalid dependabot property and sync mise.lock ([fc00f83](https://github.com/snowdreamtech/template/commit/fc00f837c81a6e2e101144ed9d9181521903447d))
* **deps:** replace asdf pipx backend with native mise plugin to resolve Windows CI failures ([05aba02](https://github.com/snowdreamtech/template/commit/05aba026dd7348bb2fb17848076a68b15411e33f))
* **docker:** add version specifications to mise install calls ([72149e4](https://github.com/snowdreamtech/template/commit/72149e4b14f49d2ad67543af54f1d654b5efcbb8))
* **docs:** fix broken relative links in documentation ([c13692c](https://github.com/snowdreamtech/template/commit/c13692c6ecf4fca79bf117fcda7076abef1acba1))
* **docs:** regenerate pnpm lockfile to match overrides ([b553918](https://github.com/snowdreamtech/template/commit/b553918d8c7d81bc4a7f740a756c590a06d48e2a))
* **docs:** resolve MD028 blockquote lint in snowdreamtech.init workflow ([4e9ecfa](https://github.com/snowdreamtech/template/commit/4e9ecfabb7e9bb7a2af0b3e0c73c56cb31f05262))
* **docs:** resolve vitepress from docs/node_modules/.bin ([13dc8f9](https://github.com/snowdreamtech/template/commit/13dc8f9deb1daf7c3ae60190a8b497dc4fbf0249))
* **env:** improve mise tool resolution and register zizmor ([120db7e](https://github.com/snowdreamtech/template/commit/120db7ec0ee94004b7b06337793ac5a8c9a93b0a))
* **go:** add version specifications to mise install calls ([4b6964f](https://github.com/snowdreamtech/template/commit/4b6964fbb2e43ea3674d433058d41f6f0db80335))
* **goreleaser:** add MIPS architecture variants to tag triggers ([03922c4](https://github.com/snowdreamtech/template/commit/03922c4fe57d0e62e69c359c979d8f388af525be))
* **goreleaser:** expand architecture-specific tag triggers ([2d72240](https://github.com/snowdreamtech/template/commit/2d72240330308c1b7196f7ec8f86d25c93cb4125))
* **goreleaser:** expand tag compatibility for monorepos and release-prefixes ([e62a42d](https://github.com/snowdreamtech/template/commit/e62a42d866b898ac086b2588d1814390fb55c522))
* **goreleaser:** implement Absolute Architecture Coverage tag matrix ([4bcd30a](https://github.com/snowdreamtech/template/commit/4bcd30a1834db9886527ada27ab3abb5126c756f))
* **goreleaser:** implement Engineering Excellence tag matrix ([715d0db](https://github.com/snowdreamtech/template/commit/715d0db59bb93b9bcacdc435b94e07833ad94ef3))
* **goreleaser:** implement Grand Slam Tag Matrix ([1a1815f](https://github.com/snowdreamtech/template/commit/1a1815f26b277a5b8900edeaab7c936855e043bb))
* **goreleaser:** implement Universal Tag Matrix ([b9555be](https://github.com/snowdreamtech/template/commit/b9555beca2ce5ac8984695d8ddcac5edbc771bf1))
* **goreleaser:** implement Universe-Level Tag Matrix ([266e7a7](https://github.com/snowdreamtech/template/commit/266e7a7a45d6bc8523f5cf18988854c3c385c03e))
* **goreleaser:** support v, V, and numeric tag formats ([cd99ffa](https://github.com/snowdreamtech/template/commit/cd99ffacfdf6f6143c32a4f5c36f21f69cd980a1))
* **goreleaser:** ultimate expansion of tag triggers ([aa1c346](https://github.com/snowdreamtech/template/commit/aa1c3463b41ceb918fef18e134598d367e67ad0b))
* **goreleaser:** unify scoped tag triggers with top-level patterns ([b0286a7](https://github.com/snowdreamtech/template/commit/b0286a7a88ff117f3708f8802e49b3a75dfdd5e4))
* **helm:** add version specification to mise install call ([4b5f452](https://github.com/snowdreamtech/template/commit/4b5f452e8003e64639a89447d96abd16abf5ce2c))
* improve POSIX compatibility in bin-resolver.sh ([ebab76e](https://github.com/snowdreamtech/template/commit/ebab76e69041ace9d2bcbeef107044763a27a991))
* improve stat command cross-platform compatibility ([031ffce](https://github.com/snowdreamtech/template/commit/031ffceb774c785a104eb93e715e4966786dafd4))
* **install:** add atomic verification to all Phase 3 low priority tools ([631318f](https://github.com/snowdreamtech/template/commit/631318fd13a28676f6015eca24c3759a86d2c4f2))
* **install:** add atomic verification to clang-format, google-java-format, stylua, ktlint ([0b92a0d](https://github.com/snowdreamtech/template/commit/0b92a0d1e40ca2b00690a204e2c077526e6a2886))
* **install:** add atomic verification to Phase 2 remaining tools ([bb0ae96](https://github.com/snowdreamtech/template/commit/bb0ae96ddb968b00495196547fc5795f3444d8e3))
* **java:** add version specification to mise install call ([8f359a1](https://github.com/snowdreamtech/template/commit/8f359a1fef21822d9dfe733cd4a45e8bc0eaa28e))
* **lint:** add aggressive reinstall logic in lint-wrapper ([40ab04d](https://github.com/snowdreamtech/template/commit/40ab04d550adb207e6fd954afdae6fe48da7f6a3))
* **lint:** add auto-install fallback for missing tools in CI ([b293923](https://github.com/snowdreamtech/template/commit/b2939234dc31427706c716bc261d29e0635e0174))
* **lint:** add license headers to test files and fix markdown formatting ([02263cc](https://github.com/snowdreamtech/template/commit/02263ccdd8163b427643af8fce0103232b01b82c))
* **lint:** add reshim and direct execution fallback ([21be93b](https://github.com/snowdreamtech/template/commit/21be93bb2f143b16f2476a7bb5f46bb838a09961))
* **lint:** add tool spec mapping for shfmt, taplo, and editorconfig-checker ([86c47ac](https://github.com/snowdreamtech/template/commit/86c47aceaedce07dda24802869a9f024a19616df))
* **lint:** add zizmor to optional security tools list ([13b06fd](https://github.com/snowdreamtech/template/commit/13b06fdf2720ea8b493dff1301a889c461a1f589))
* **lint:** correct editorconfig-checker binary name to 'ec' ([c56679c](https://github.com/snowdreamtech/template/commit/c56679c2515de0754e593a7f295623a0e075577c))
* **lint:** exclude all markdown files from prettier formatting ([4fee420](https://github.com/snowdreamtech/template/commit/4fee420ee9f672867de2c0f0c59cf00baf8d8ad3))
* **lint:** exclude test files from shellcheck-posix and markdown from prettier ([0dfd5b2](https://github.com/snowdreamtech/template/commit/0dfd5b2eb22337c03b4c06a7879369328922e7fa))
* **lint:** fix yamllint, markdownlint, and editorconfig issues ([b7925a2](https://github.com/snowdreamtech/template/commit/b7925a2f9661b860f50134e5de1a028f172e971a))
* **lint:** prevent prettier/markdownlint emphasis format conflict ([0c38f99](https://github.com/snowdreamtech/template/commit/0c38f994d6f4592e35e02e9dbb85244722f1388c))
* **lint:** resolve spurious warning by reporting pre-commit version ([188a325](https://github.com/snowdreamtech/template/commit/188a3251cb2f76d053bbb5e9ccef5846eae4d269))
* **lint:** use mise exec as fallback when resolve_bin fails in CI ([8e52231](https://github.com/snowdreamtech/template/commit/8e5223113a198757c33ede338ebd6fc55fac84b0))
* **linux:** use -perm /111 instead of +111 for better compatibility ([512e69b](https://github.com/snowdreamtech/template/commit/512e69b504a1810c4fa126fab5b6079bfbba125d))
* **macos:** avoid -perm flag for BSD find compatibility ([79b98b0](https://github.com/snowdreamtech/template/commit/79b98b071476afe24a80e9876c7e2509022c1529))
* **macos:** use -perm +111 instead of -executable for BSD find compatibility ([911c1b4](https://github.com/snowdreamtech/template/commit/911c1b47e643bcb1d95e6400a7e6abdbfeea82fe))
* **mise:** add additional paranoid mode disable flags ([7e83141](https://github.com/snowdreamtech/template/commit/7e83141b005bb8e0e6dc930f35eec7436c1bceb1))
* **mise:** align pnpm version to 10.30.3 to fix detection failure ([c01671d](https://github.com/snowdreamtech/template/commit/c01671df9125246be64d26cb4129450684067ef6))
* **mise:** clear cache before sync-lock to prevent stale provenance verification ([027874d](https://github.com/snowdreamtech/template/commit/027874db7c3ce6b823576d2bdb3da0bf868f5209))
* **mise:** downgrade node to 22.14.0 for macOS 12 compatibility ([c5495b4](https://github.com/snowdreamtech/template/commit/c5495b491a7796d25bd2ff26fbb8cbd9fdc43aff))
* **mise:** downgrade Python from 3.14.3 to 3.13.12 ([71ae8fd](https://github.com/snowdreamtech/template/commit/71ae8fd1f1750ed0440b18be1515610063cee1e8))
* **mise:** enable GitHub proxy locally and fix checkmake version ([a0e65aa](https://github.com/snowdreamtech/template/commit/a0e65aaaa862ddc657beeac74fd598c509c54c76))
* **mise:** handle platform-specific editorconfig-checker binary names ([5ebb3ff](https://github.com/snowdreamtech/template/commit/5ebb3ff76b07f9eb31d61755fd064746512dbd69))
* **mise:** increase HTTP timeout to 300s for slow GitHub downloads ([123aa0c](https://github.com/snowdreamtech/template/commit/123aa0cd8b8f07ccc48e59758f887e778e76b316))
* **mise:** make lint ([097f32b](https://github.com/snowdreamtech/template/commit/097f32bdffbaf3ccc78eaa367ecb3e40a2935939))
* **mise:** permanently disable paranoid mode for lockfile sync ([49dbb04](https://github.com/snowdreamtech/template/commit/49dbb044f65c4c5203edd751218f3d90315214f6))
* **mise:** restore tier 1 tools and implement surgical lock ritual ([c404e7a](https://github.com/snowdreamtech/template/commit/c404e7a89ace7023f8df914ebfe7fffaa3fe2b29))
* **mise:** skip Go checksum verification for Aliyun mirror ([8821f81](https://github.com/snowdreamtech/template/commit/8821f814b7bb37d56d8e45e4163cc3292624fafd))
* **mise:** synchronize multi-platform lockfile with pnpm 10.30.3 ([fc2113b](https://github.com/snowdreamtech/template/commit/fc2113b5c63b845ca4a8631cd50fde9db5e59d15))
* **nightly:** add fetch-depth: 0 for comprehensive security audit ([8ef1a9b](https://github.com/snowdreamtech/template/commit/8ef1a9bb0390b4c5f0442219c9a7d71b38557c48))
* **nightly:** elevate tokens for security reporting ([55ddb37](https://github.com/snowdreamtech/template/commit/55ddb370027110804ea3924c88e9ea6d0fc63200))
* **node:** add corepack resiliency and npm fallback ([d72055b](https://github.com/snowdreamtech/template/commit/d72055b2e0c3eeb97bb9103b413188fe2f21f0be))
* **node:** add version specifications to mise install calls ([afd9c04](https://github.com/snowdreamtech/template/commit/afd9c0456fea9d3a427dfe8f70913d7e745b92fb))
* **node:** resolve install hang by skipping redundant corepack activation ([36b3d44](https://github.com/snowdreamtech/template/commit/36b3d448f4ad914621752f4504560e6d4d110975))
* **openapi:** add version specification to mise install call ([81e52b5](https://github.com/snowdreamtech/template/commit/81e52b5bcd5fa9f7a82036aabbb2fd24727783a0))
* **pre-commit:** migrate deprecated stage name ([8392cbf](https://github.com/snowdreamtech/template/commit/8392cbfe73aee175bb7732255a7f992c75078184))
* **pre-commit:** skip osv-scanner in local pre-commit hooks ([b4b82ac](https://github.com/snowdreamtech/template/commit/b4b82ac7078d197c8a4dfac85a2748cdee5f61e4))
* **protobuf:** add version specification to mise install call ([1f7938e](https://github.com/snowdreamtech/template/commit/1f7938e2cbaf3661a0387bd476ed73ce6e7d4e15))
* **python:** add version specifications to mise install calls ([b808c7e](https://github.com/snowdreamtech/template/commit/b808c7e16cafa6fdbf6c50ef8be5a8639401c9e4))
* re-enable refresh_mise_cache with timeout protection ([efe2b3e](https://github.com/snowdreamtech/template/commit/efe2b3e98fbf4650b3375bfb11cb0355b97e1081))
* **registry:** propagate mise install failures in CI ([582eeef](https://github.com/snowdreamtech/template/commit/582eeef44ae4d368ae59bc2d5582b49fb974064f))
* **rego:** add version specification to mise install call ([60eb8ee](https://github.com/snowdreamtech/template/commit/60eb8eecfb6c4f1b38d2ce41500c29774ee16e8c))
* replace %b with %s in printf to avoid Windows path issues ([29ba55b](https://github.com/snowdreamtech/template/commit/29ba55b82400d7ec41074c2951cbbd614eab3cb5))
* resolve make verify hanging and add Kiro config ([505e140](https://github.com/snowdreamtech/template/commit/505e140402edadd12c635a1b0d27f3785e0ea717))
* restore mise metadata cache initialization to resolve JIT tool detection ([9b15c4c](https://github.com/snowdreamtech/template/commit/9b15c4cd469e9af229ef73e9f6272705d176349f))
* **runner:** add version specifications to mise install calls ([bf3d5df](https://github.com/snowdreamtech/template/commit/bf3d5df78889a6e27e41f961ba38a239a5e0a8bd))
* **scorecard:** pin action to verified underlying commit SHA ([5f6a357](https://github.com/snowdreamtech/template/commit/5f6a357398a746b4558c8f3f82e006e8cab78372))
* **scripts:** add recursion lock cleanup advice to concurrency warning ([ce27ada](https://github.com/snowdreamtech/template/commit/ce27adafa7b711ba6d9ad993e4e95ef534d289b8))
* **scripts:** centralize GitHub token normalization ([50429be](https://github.com/snowdreamtech/template/commit/50429be4d71f62734f5a1bbceaf4d473ce008885))
* **scripts:** comprehensive set -u safety sweep for all shell scripts ([4c5ab38](https://github.com/snowdreamtech/template/commit/4c5ab38308c9cdded00c1ffa32fc6d90ee182ff6))
* **scripts:** fix JSON output formatting issues ([7b30a64](https://github.com/snowdreamtech/template/commit/7b30a640d262c9d47bd75956f92d5d3819cd4ce8))
* **scripts:** harden gen-full-manifest with absolute paths and fix perl syntax ([41199c6](https://github.com/snowdreamtech/template/commit/41199c6493cc0ef4252975645600fe16db423296))
* **scripts:** improve mise path detection in utility scripts ([3c071d8](https://github.com/snowdreamtech/template/commit/3c071d8d4f546cae906e55092dfe9b54566e7d9f))
* **scripts:** improve path robustness and update verifier in release.sh ([000ec98](https://github.com/snowdreamtech/template/commit/000ec98593f2ef52842c6a37de0355fccfc80159))
* **scripts:** integrate update-tools.sh into main update sequence ([c86fa15](https://github.com/snowdreamtech/template/commit/c86fa1561803cbbe7d0e96ec110db4d761c60225))
* **scripts:** making common.sh set -u safe ([20ce016](https://github.com/snowdreamtech/template/commit/20ce01692a416d8453fb2b3876bf5edf90aa0166))
* **scripts:** prevent recursion lock accumulation across multiple runs ([6d41678](https://github.com/snowdreamtech/template/commit/6d4167845870c19af4f2ed39dec28f32e2096c9b))
* **scripts:** resolve infinite recursion and macOS compatibility in setup ([dbc5f9b](https://github.com/snowdreamtech/template/commit/dbc5f9b3bf417c7107641faacb51b4d5a637cfa1))
* **scripts:** resolve nested unbound variable errors in common.sh ([f4de98a](https://github.com/snowdreamtech/template/commit/f4de98a4e679bc5c3c9023590b1539087e6b939c))
* **scripts:** resolve PSScriptAnalyzer warnings in sync-labels.ps1 ([5d72d18](https://github.com/snowdreamtech/template/commit/5d72d18de1681078f8299e92ba9dc3d62e90a290))
* **scripts:** resolve silent hang during mise tool resolution ([dec1f65](https://github.com/snowdreamtech/template/commit/dec1f65ace1e5238222771d1ef9d3861c7ea6675))
* **scripts:** standardize mise toolchain providers and fix pnpm setup ([5f8f030](https://github.com/snowdreamtech/template/commit/5f8f0309d42ac6158d865631c2c554bcc18a27d4))
* **scripts:** suppress redundant terminal execution summaries in local environment ([83776fd](https://github.com/snowdreamtech/template/commit/83776fd54d4fe6fe35c6a822c8907d2b7ed8a9d3))
* **scripts:** use absolute path prefix for gen-full-manifest in sync-lock.sh ([bb8401d](https://github.com/snowdreamtech/template/commit/bb8401df026f6389530b895a97950822431efa4a))
* **security:** add binary availability check after version match ([bec5038](https://github.com/snowdreamtech/template/commit/bec5038d8f52f187185c5c2647b38442d395ec5a))
* **security:** add dependabot cooldown config for zizmor compliance ([7394768](https://github.com/snowdreamtech/template/commit/739476828d7812d31505a3dea3d443ffec2d374f))
* **security:** add strict validation for all CI security tools ([721f8ef](https://github.com/snowdreamtech/template/commit/721f8efbea838617d18cdd62408930752f627e73))
* **security:** add strict zizmor validation in CI ([b6fd596](https://github.com/snowdreamtech/template/commit/b6fd59623c1286a6db9b3b206414ece7dd91a030))
* **security:** add version specifications to mise install calls ([f58a226](https://github.com/snowdreamtech/template/commit/f58a22612304e482bf58f9c64c2614a5574ecf47))
* **security:** check setup_registry return values for critical tools ([cf91e3a](https://github.com/snowdreamtech/template/commit/cf91e3a2ec2274f8335fb6e6863a008d15ceb137))
* **security:** complete universal egress sync and bootstrap hardening ([1f7aea5](https://github.com/snowdreamtech/template/commit/1f7aea5b6dfd7f998adaaabf171b0b45711cdeb0))
* **security:** extend mise.lock with Windows-x64 checksums for full cross-platform security ([8da506e](https://github.com/snowdreamtech/template/commit/8da506e9b56970578f9dd48e389b4899f7406cea))
* **security:** finalize global SHA-1 pinning for all 16 workflows ([2a14204](https://github.com/snowdreamtech/template/commit/2a1420406c5d80dcd14860dc7becc4919c50d8a7))
* **security:** fix OSV-scanner and Zizmor installation in CI ([683639e](https://github.com/snowdreamtech/template/commit/683639e1424d9762128cde26e3e7fad716cbcdef))
* **security:** formalize mise toolchain integrity standards ([fcc6026](https://github.com/snowdreamtech/template/commit/fcc602653556e39266b0b74cf9b7d45b0f4fb361))
* **security:** globally harmonize platinum-standard egress whitelists ([38f583c](https://github.com/snowdreamtech/template/commit/38f583c7e021db2df48647e54afd5b519efaecf9))
* **security:** implement fail-safe offline fallback for zizmor and fix binary scan logic ([35a0cd7](https://github.com/snowdreamtech/template/commit/35a0cd739563045365f2d6af1e5c4849b63510d5))
* **security:** implement MISE_LOCKED to prevent toolchain poisoning ([3aca5fa](https://github.com/snowdreamtech/template/commit/3aca5fa835300131470118b4e2b8c9249b9dfb11))
* **security:** implement mise.lock for cryptographic toolchain integrity ([14515fe](https://github.com/snowdreamtech/template/commit/14515fee125ed87c41b67b81a8a47af2986960a5))
* **security:** implement robust stealth binary detection in audit engine ([1c0f20d](https://github.com/snowdreamtech/template/commit/1c0f20ded821c5b44075cbd531af4dec1393b0a0))
* **security:** pin mise-action to immutable SHA-1 hash ([239ad6c](https://github.com/snowdreamtech/template/commit/239ad6c08655aa7bf8890a7f5a42fb84b38b3545))
* **security:** reach 100% universal platinum-standard egress sync ([1917d82](https://github.com/snowdreamtech/template/commit/1917d82dd69a730c948cba2b10430c14b2d5edb5))
* **security:** remediate zizmor finding and fix audit script ([b847843](https://github.com/snowdreamtech/template/commit/b8478438bc9ad2799a9071543476c7c55a12ea26))
* **security:** resolve literal newline artifact in scorecard.yml ([4b6a3af](https://github.com/snowdreamtech/template/commit/4b6a3af9738a3479c72e6b91c5471c3588f4c574))
* **security:** resolve Scorecard imposter commit error ([88e8bed](https://github.com/snowdreamtech/template/commit/88e8beded6f6fafccb39f7c5c209ad8b25a82630))
* **security:** resolve Scorecard Pinned-Dependencies and Token-Permissions warnings ([d5238a6](https://github.com/snowdreamtech/template/commit/d5238a6fd2a3a706358e323533ea08fd92839831))
* **security:** resolve syntax error in audit engine ([c222674](https://github.com/snowdreamtech/template/commit/c22267445a5a677fab2c0ee77e67af4db5057210))
* **security:** robustify zizmor token handling and binary audit logic ([2a327aa](https://github.com/snowdreamtech/template/commit/2a327aacbcd43be4c018cfc034c84a3c1810cbb5))
* **security:** universally synchronize egress whitelists with perfect formatting ([6ce96d8](https://github.com/snowdreamtech/template/commit/6ce96d8b94053b72030bddf9751035452042fff2))
* **security:** universally synchronize language repo whitelists ([e4b39ed](https://github.com/snowdreamtech/template/commit/e4b39edcc61e9853d63b1b104a9df613da441afc))
* **security:** widen egress whitelists for multi-distro linux support ([99dba63](https://github.com/snowdreamtech/template/commit/99dba63bfcb665451eba4bf694fc178c53fb2733))
* **setup:** add provider paths for lean, nim, racket, vala, and aptos ([c968143](https://github.com/snowdreamtech/template/commit/c968143d1199231c469ff9ad6919419b8c12aefe))
* **setup:** downgrade python to 3.12.3 and disable github proxy to prevent hangs ([4ef9873](https://github.com/snowdreamtech/template/commit/4ef98730f0f0c3f918b07657cdb92d45318cd8df))
* **setup:** ensure gitleaks installs in CI without .git directory ([9fb833d](https://github.com/snowdreamtech/template/commit/9fb833df3c5cbd2af280ebee4c3211ab16a2d858))
* **setup:** force reinstall non-executable tools in CI ([dee60a7](https://github.com/snowdreamtech/template/commit/dee60a74f8d22948bc55fca92a6d464eb131764d))
* **setup:** handle mise bin tools verification ([5c386fa](https://github.com/snowdreamtech/template/commit/5c386faae9715eeada52124c8315b71b7ad871de))
* **setup:** handle non-zero exit codes in smoke tests ([3fe6c05](https://github.com/snowdreamtech/template/commit/3fe6c054d4609b2263399b5df10d10f280919c2c))
* **setup:** improve network timeout reliability for large binary downloads ([432f092](https://github.com/snowdreamtech/template/commit/432f09206b8e2cad2a0aa148233c4975547a02b0))
* **setup:** resolve mise installation failure on Windows and unify pipx setup ([2ec9f9b](https://github.com/snowdreamtech/template/commit/2ec9f9bed0803737d8b194226acddc0b788569d7))
* **setup:** use full provider path for checkmake installation ([1e40c13](https://github.com/snowdreamtech/template/commit/1e40c1347e7ac53de9419198323d3a7ab69d8435))
* **setup:** use full provider path for gitleaks installation ([7577489](https://github.com/snowdreamtech/template/commit/757748915a041537148f71085f9e377731a00dfa))
* **setup:** use mise which for binary verification in CI ([ded29c1](https://github.com/snowdreamtech/template/commit/ded29c17cdd851f46808a02ad086264f1038eb71))
* **setup:** verify tool executability in CI, not just mise registration ([5404189](https://github.com/snowdreamtech/template/commit/5404189d1f65c4064fdd773764c93e004af550dc))
* **shell:** add error handling for shell tool installation in CI ([967b8bf](https://github.com/snowdreamtech/template/commit/967b8bf69800365810dc023fd44db8d1ef3b9479))
* **shell:** add version specifications to mise install calls ([b576a7f](https://github.com/snowdreamtech/template/commit/b576a7f3c21726a3efeb6b001c1ec05332900aac))
* **shell:** remove duplicate log_summary line causing syntax error ([d8f75d9](https://github.com/snowdreamtech/template/commit/d8f75d9078c2a05b41a7a1762c1fc093030c5f22))
* **shell:** restore missing functions and add aggressive cache refresh ([f89f7f0](https://github.com/snowdreamtech/template/commit/f89f7f0913ecb6e165f8072d6b88f985b672dbcb))
* skip executable permission check on Windows (POSIX compliant) ([e365e6a](https://github.com/snowdreamtech/template/commit/e365e6aa58b67f08d058cc0a5853c8991e2e71ef))
* **sql:** add version specification to mise install call ([cde5de5](https://github.com/snowdreamtech/template/commit/cde5de54c3a705463cc0c01cf94e0da170ebed06))
* sync GITHUB_PATH to current shell for same-step tool availability ([e51048d](https://github.com/snowdreamtech/template/commit/e51048d1fdfeb320ad9922a4efd162d2da4c2821))
* **terraform:** add version specification to mise install call ([b99fa5a](https://github.com/snowdreamtech/template/commit/b99fa5a5be116a53300620261fcc5471b6aa092f))
* **test:** enforce mock path priority in check-env.bats ([babd8df](https://github.com/snowdreamtech/template/commit/babd8df6dfd551800ce3b1d4599e36a5af96b7c5))
* **testing:** add version specification to mise install call ([e4f4820](https://github.com/snowdreamtech/template/commit/e4f48202b510007a926776884945bac35c64b742))
* **toolchain:** extend ALF to handle full mise install with go: tools ([fec200c](https://github.com/snowdreamtech/template/commit/fec200c1b6f746507c2e7318eb88b1a88f159afb))
* **toolchain:** lazy-load mise cache and harden BATS detection ([0e2fb32](https://github.com/snowdreamtech/template/commit/0e2fb32dceae867bbdf55c4232796c0a818652a6))
* **toolchain:** migrate addlicense to github backend to support locked mode ([cf8ea8b](https://github.com/snowdreamtech/template/commit/cf8ea8b9611c39822f440ce6ebc4bcdbd18ac80d))
* **toolchain:** optimize summary logging and resolve BATS test failures ([f9639cb](https://github.com/snowdreamtech/template/commit/f9639cb97d22b9babc8c441c976ab1b802a11b55))
* update VER_MISE to 2026.4.0 to avoid RelativeUrlWithoutBase bug ([f93b0b3](https://github.com/snowdreamtech/template/commit/f93b0b39fd293b7e4e9eb4b97169f80eb38aa6eb))
* **versions:** update checkmake version to v0.3.2 ([6d1edd8](https://github.com/snowdreamtech/template/commit/6d1edd8a6299af947daf989f9badafe6efcf8f3e))
* **windows:** check root directory in verify_tool_atomic fallback ([9e2ace6](https://github.com/snowdreamtech/template/commit/9e2ace6e8cf960c25f159e53ec5b3a2d9cce3904))
* **windows:** convert Unix paths to Windows format for GITHUB_PATH ([de515d6](https://github.com/snowdreamtech/template/commit/de515d6aa91c1392fea55529a2c40550499a721c))
* **windows:** improve mise path detection for Git Bash environments ([d5a5337](https://github.com/snowdreamtech/template/commit/d5a5337ff038f35428034a2969842f99851c6297))
* **windows:** replace log_info with echo for mise path logging ([0bd8ee1](https://github.com/snowdreamtech/template/commit/0bd8ee1a951a0491649188430bb3c99be3067a93))
* **windows:** replace log_info with echo in PATH persistence ([7710710](https://github.com/snowdreamtech/template/commit/77107107bd15904198cef7146fdfb923192c160d))
* **windows:** skip command check for binaries without .exe extension ([a5ed6ca](https://github.com/snowdreamtech/template/commit/a5ed6caebb29aa8555d704cb1d8dabf4ea384111))
* **windows:** skip executable permission checks in lint-wrapper ([035ae54](https://github.com/snowdreamtech/template/commit/035ae54f20404a7c66a1a07459efb1d77678ab3f))
* **windows:** split executable check to avoid test evaluation on Windows ([cc76373](https://github.com/snowdreamtech/template/commit/cc763732a359bd52681a19b277854ea6bbe826e2))
* **windows:** use Unix-style paths for GITHUB_PATH in bash shell ([4eacf60](https://github.com/snowdreamtech/template/commit/4eacf6088d780d4896a533f3fe6acc864e65a56f))
* **windows:** write both Unix and Windows path formats to GITHUB_PATH ([62a0acf](https://github.com/snowdreamtech/template/commit/62a0acf12d04b0bd49d0e53d0bc45d2b261426c0))
* **workflows:** comprehensive alignment with security and Makefile standards ([ee4114b](https://github.com/snowdreamtech/template/commit/ee4114b4d2b55edda9ab13b7dcb5c487cb2c5d15))
* **workflows:** convert Windows LOCALAPPDATA path for Git Bash ([ba9192a](https://github.com/snowdreamtech/template/commit/ba9192a8013a8956d9d925227bd48b88f3703fcd))
* **workflows:** standardize checkout naming and token injection ([f834d6c](https://github.com/snowdreamtech/template/commit/f834d6cac630efa7265f107c09bcf9657a667fdb))
* **zizmor:** ensure zizmor installs in CI regardless of workflow files presence ([9f6f620](https://github.com/snowdreamtech/template/commit/9f6f620ba69bc70304dbd7873ae9db2b5944ba76))
* **zizmor:** force reinstall in CI to guard against stale cache ([70597c2](https://github.com/snowdreamtech/template/commit/70597c227f45c9e7c0423768fc8d74323260f05b))
* **zizmor:** update repository from woodruffw to zizmorcore ([316a7cc](https://github.com/snowdreamtech/template/commit/316a7ccf6e3903022301ab3e3f11a4b866ed9484))


### Performance Improvements

* **infra:** implement lazy loading and unified refresh for mise cache ([f9e7aec](https://github.com/snowdreamtech/template/commit/f9e7aec490fb596525d362a79ef59726fc57bc90))
* **mise:** increase HTTP timeout to 300s for improved network reliability ([3335482](https://github.com/snowdreamtech/template/commit/3335482f99f98272149a30e831de066c11db87d2))
* **setup:** skip goreleaser installation in local dev to prevent setup hangs ([c95f801](https://github.com/snowdreamtech/template/commit/c95f80149c7440797995b872a27e3813079038bd))


### Documentation

* add comprehensive token usage audit report ([9983421](https://github.com/snowdreamtech/template/commit/9983421492acc679c2e27a2b01cdf3a6909c7ab9))


### Miscellaneous Chores

* remove temporary token audit report ([50a73c7](https://github.com/snowdreamtech/template/commit/50a73c7f88dbcf56332220a5941f6ca00f994bcd))


### Code Refactoring

* **ci:** standardize token usage across all workflows ([54b8445](https://github.com/snowdreamtech/template/commit/54b84456eb13b77b79cbbea32510405ec6f616b7))

## [Unreleased]

### Added

- Initial project template with AI IDE ecosystem support (50+ AI IDEs)
- Core rule system in `.agent/rules/` as Single Source of Truth
- SpecKit workflow suite for full feature lifecycle management
- DevContainer with Docker Compose for reproducible development environments
- Comprehensive CI/CD pipeline (lint, security scan, CodeQL, GoReleaser, stale)
- Pre-commit hooks with 50+ quality gate checks
- VS Code productivity configurations (tasks and launch profiles)
- Project hydration script (`scripts/init-project.sh`)
- GitHub community health files (SECURITY, CONTRIBUTING, CODE_OF_CONDUCT)
- Multi-IDE prompt/command shortcuts for all major AI coding assistants

### Changed

- N/A

### Deprecated

- N/A

### Removed

- N/A

### Fixed

- N/A

### Security

- N/A

[Unreleased]: https://github.com/snowdreamtech/template/commits/main/
