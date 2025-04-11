# Release Versioning

This project follows **Semantic Versioning** (`MAJOR.MINOR.PATCH`) for all public releases. This versioning policy is designed to communicate the impact of changes clearly to developers, node operators, and integrators building on or interacting with the protocol.

## 🔢 Version Format

Examples: `v1.0.0`, `v2.3.1`, `v0.9.5`

---

## 🧱 MAJOR Releases (`X.0.0`)

A new **MAJOR** version is released when **breaking changes** are introduced to any core protocol components, including:

- Changes that are **not backward-compatible** (e.g. consensus logic, block structure, cryptographic primitives)
- Upgrades requiring **hard forks** or **network migrations**
- Deprecation or removal of public APIs, CLI commands, or RPC endpoints
- Significant redesign of core data structures or system behavior

> These releases often require node operators, validators, and developers to take action to stay compatible.
>

---

## 🌱 MINOR Releases (`X.Y.0`)

A new **MINOR** version is released when **new features** or capabilities are added in a **backward-compatible** manner, including:

- Addition of new RPC endpoints or CLI flags
- New smart contract APIs or interfaces (non-breaking)
- Performance improvements that do not alter behavior
- Network upgrades that do **not** require forks or migrations
- Enhancements to developer tools or SDKs that preserve existing interfaces

> Minor releases are safe to adopt without breaking existing integrations, but may offer new capabilities.
>

---

## 🛠 PATCH Releases (`X.Y.Z`)

A new **PATCH** version is released for **backward-compatible bug fixes**, including:

- Fixes for security vulnerabilities or consensus edge cases
- Minor logic corrections
- Documentation updates, CLI usability tweaks
- Fixes for non-breaking SDK or toolchain bugs

> Patch releases are always safe and recommended to adopt for stability and correctness.
>

---

## 🧪 Testnet vs Mainnet Releases

We maintain a clear separation between **testnet** and **mainnet** releases to support stable production environments while enabling faster iteration and experimentation on testnets.

### 🔁 Testnet Releases

- Tagged with a **`testnet`** suffix or pre-release label (e.g. `v1.2.0-testnet`, `v1.3.0-beta.1`)
- May include:
    - Experimental features
    - Early versions of consensus or protocol upgrades
    - Non-finalized RPC changes or CLI commands
- Intended **for testing and validator experimentation** only — **not production-safe**
- Can be released **more frequently** than mainnet versions
- Changes may be **revised or reverted** before mainnet release

### 🟢 Mainnet Releases

- Tagged with **final version numbers only**, e.g. `v1.2.0`, `v2.0.0`
- Must be:
    - **Thoroughly tested and audited**
    - **Backward-compatible** within the same MAJOR version (unless otherwise stated)
- Require community coordination if involving consensus or network upgrades
- Documented with migration guides and upgrade instructions

### 🔒 Versioning Summary

| Release Type | Example Tag | Intended Use | Stability |
| --- | --- | --- | --- |
| Mainnet | `v1.2.0` | Production | Stable |
| Testnet | `v1.3.0-rc` | Experimentation | Unstable |

---

## 🔖 Tagging and Releases

All versions are tagged with Git in the format `vX.Y.Z` and include:

- Release notes (summarizing changes and migration notes if needed)
- Reference to merged PRs and relevant issues
- Upgrade instructions (when necessary)

---

Examples: `v1.0.0`, `v2.3.1`, `v0.9.5`

---

## 🧱 MAJOR Releases (`X.0.0`)

A new **MAJOR** version is released when **breaking changes** are introduced to any core protocol components, including:

- Changes that are **not backward-compatible** (e.g. consensus logic, block structure, cryptographic primitives)
- Upgrades requiring **hard forks** or **network migrations**
- Deprecation or removal of public APIs, CLI commands, or RPC endpoints
- Significant redesign of core data structures or system behavior

> These releases often require node operators, validators, and developers to take action to stay compatible.
>

---

## 🌱 MINOR Releases (`X.Y.0`)

A new **MINOR** version is released when **new features** or capabilities are added in a **backward-compatible** manner, including:

- Addition of new RPC endpoints or CLI flags
- New smart contract APIs or interfaces (non-breaking)
- Performance improvements that do not alter behavior
- Network upgrades that do **not** require forks or migrations
- Enhancements to developer tools or SDKs that preserve existing interfaces

> Minor releases are safe to adopt without breaking existing integrations, but may offer new capabilities.
>

---

## 🛠 PATCH Releases (`X.Y.Z`)

A new **PATCH** version is released for **backward-compatible bug fixes**, including:

- Fixes for security vulnerabilities or consensus edge cases
- Minor logic corrections
- Documentation updates, CLI usability tweaks
- Fixes for non-breaking SDK or toolchain bugs

> Patch releases are always safe and recommended to adopt for stability and correctness.
>

---

## 🧪 Testnet vs Mainnet Releases

We maintain a clear separation between **testnet** and **mainnet** releases to support stable production environments while enabling safe experimentation on testnets.

### 🔁 Testnet Releases (`rc`)

- Use the **`rc` (release candidate)** suffix (e.g. `v1.3.0-rc.1`, `v2.0.0-rc.2`)
- Released specifically for **testnet deployments**
- May include:
    - Experimental or unfinished protocol features
    - Pending upgrades intended for a future mainnet release
    - APIs or behaviors that are still under review
- **Not stable**, may differ from the final mainnet release
- Published as **GitHub pre-releases** to indicate they are **not for production use**

> Testnet releases allow developers and node operators to validate behavior before mainnet deployment.
>

### 🟢 Mainnet Releases

- Tagged with **final version numbers only**, e.g. `v1.2.0`, `v2.0.0`
- Must be:
    - **Thoroughly tested and audited**
    - **Backward-compatible** within the same MAJOR version (unless otherwise stated)
- Require community coordination if involving consensus or network upgrades
- Documented with migration guides and upgrade instructions

### 🔒 Versioning Summary

| Release Type | Example Tag | Intended Use | Stability |
| --- | --- | --- | --- |
| Mainnet | `v1.2.0` | Production | Stable |
| Testnet RC | `v1.3.0-rc.1` | Testnet Only | Unstable |

---

## 📌 Stability Notes

- **v0.x.y** releases are considered **pre-stable**. Breaking changes may occur even in minor version bumps until `v1.0.0` is released.
- Once `v1.0.0` is reached, the versioning becomes strictly semantic as described above.

---

## 🔖 Tagging and Releases

All versions are tagged with Git in the format `vX.Y.Z` and include:

- Release notes (summarizing changes and migration notes if needed)
- Reference to merged PRs and relevant issues
- Upgrade instructions (when necessary)

---
