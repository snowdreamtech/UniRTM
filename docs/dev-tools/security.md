# Supply Chain Security

In today's thriving open-source ecosystem, Software Supply Chain Security has become paramount. Incidents of malicious mirror tampering and publishing compromised package versions (poisoning attacks) are occurring with alarming frequency.

Instead of merely relying on post-incident external scanners, UniRTM integrates a robust, proactive defense system **directly into the core architecture of its downloading and environment mounting engine**.

## 1. Enforced Checksum & Lockfile Verification

When you install a tool via `unirtm install`, UniRTM never blindly trusts the network channel:

- **Real-time SHA-256 Matching**: Before writing anything to disk, the system calculates the `sha256` checksum of the downloaded file in memory.
- **Lockfile Enforcement**: This calculated hash is strictly matched against the tamper-proof fingerprint recorded in your `unirtm.lock` file. If there is even the slightest discrepancy, the installation process is **immediately aborted** with a critical warning. This completely mitigates the risk of Man-in-the-Middle (MITM) attacks and malicious binary replacement.

## 2. Mandatory GPG Signature Verification

Many official tools (such as Node.js and Go) provide GPG (GNU Privacy Guard) signatures for their release packages.
UniRTM has a built-in GPG verification mechanism:

- It automatically fetches the issuer's public keys and performs cryptographic integrity checks against the `.sig` or `.asc` files during package resolution.
- This ensures the codebase was authentically released by the official team, and not by a third-party imposter.

## 3. GitHub Provenance & Attestation

For open-source projects hosted on GitHub, UniRTM deeply integrates with GitHub's native **Attestation** and Provenance features:

- It verifies the build provenance signed by the official GitHub Actions environment.
- This guarantees that the downloaded toolchain was built within a trusted CI/CD pipeline, preventing malicious maintainers from planting backdoors during local builds.

## 4. Strict Adherence to Official Channels

Unlike many tools that default to various uncontrollable third-party mirrors for the sake of speed, UniRTM **defaults to extracting data exclusively from the official sources** of languages and tools (e.g., the official Node.js dist page, the native Go distribution gateway).
Unless explicitly overridden by the user in the configuration, we never sacrifice security for so-called "acceleration" by exposing your terminal environment to untrusted third-party mirrors.

## 5. Minimum Update Cooling-off Period

Supply chain poisoning often exploits the developer's desire to "stay on the latest version," expecting a massive number of users to automatically update to a maliciously patched release within hours.
To counter this, UniRTM introduces a **Minimum Update Cooling-off Period**:

- When resolving `latest` tags (instead of pinned versions), the system defaults to delaying the adoption of newly released versions that are less than 24-48 hours old.
- This golden time window provides the open-source security community enough time to discover and yank polluted packages, preventing your environment from becoming a zero-day victim of poisoning attacks.
