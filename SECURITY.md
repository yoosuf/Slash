# Security Policy

Slash is a plugin that runs in trust-gated hook contexts within coding tools. It reads source code from your workspace and may intercept tool output. Security and privacy are fundamental to the design.

## Core Guarantees

**Zero network by default.** Slash makes no outbound network calls. The CCR (content-addressed retrieval) cache stores your source code on disk under your existing file permissions (`$XDG_CACHE_HOME/slash` on Linux/macOS, `%LOCALAPPDATA%\slash` on Windows). Nothing leaves your machine without explicit opt-in.

**Reversible compression.** Every token reduction is reversible via the `retrieve(handle)` tool. If compression loses context the model needs, it can fetch the original. This design eliminates the risk of irreversible quality loss.

**Opt-in for any new features.** Cloud compression (learned models, team-shared state) is a *separate, separately-installed* package (`slash-cloud`) and is disabled by default. The core repo contains only local-first functionality. Cloud features require explicit opt-in via configuration and network permission prompts.

## Reporting Security Issues

**Do not open public issues for security vulnerabilities.** Instead, email security concerns to the maintainers (see MAINTAINERS.md or GitHub org settings).

Include:
- Description of the vulnerability.
- Steps to reproduce (if applicable).
- Affected versions.
- Suggested fix (optional).

Maintainers will acknowledge receipt within 48 hours and aim to release a fix within 2 weeks of confirmation.

## Supply-Chain Security

Slash is distributed as:

1. **Signed, checksummed binaries** — every release includes `.sha256` checksums and GPG signatures. Verify before use:
   ```bash
   sha256sum -c slash_1.0.0_linux_amd64.sha256
   gpg --verify slash_1.0.0_linux_amd64.sig
   ```

2. **Plugin bundles** — packaged through official marketplaces (Claude, Codex, Antigravity, Copilot). These enforce code review and permission disclosure.

3. **Source transparency** — all code is public. Builds are reproducible; we publish the Docker environment and build script so anyone can verify that released binaries match the source.

### Permissions

Slash requests the minimum necessary:
- **File read** (workspace files, `.gitignore` for secret patterns).
- **File write** (CCR cache under `$XDG_CACHE_HOME/slash`; never outside workspace).
- **Unix socket** (IPC to the daemon; localhost only).
- **Process spawning** (tree-sitter parsing, optional; sandboxed).

Hosts gate third-party hooks and MCP servers behind trust prompts. When you install Slash, your host will ask you to approve it. The approved list is visible in each tool's settings.

## Data Handling

### What is Cached

The CCR cache stores:
- File bodies for reads + their compression ratio + timestamp + original path.
- Bash output + stderr when compressed.
- Structured metadata (token counts, compression savings).

### What is NOT Cached

- Source code is never persisted beyond the session's CCR TTL (~24h) unless you edit the file again.
- Passwords, environment variable values, or secrets marked in `.gitignore`.
- Tool prompts, model outputs, or agent reasoning.
- Network requests or credentials.

### Cache Expiration

- **TTL**: 24 hours from creation (configurable).
- **LRU eviction**: cache size is capped (default 1 GB); oldest items are dropped first.
- **Manual purge**: `slash purge` wipes the entire cache.
- **Path invalidation**: if you edit a file, handles for that file are immediately invalidated.

### Respecting Secrets

Slash respects `.gitignore` patterns and common secret patterns (`.env`, `*_key`, `*_secret`, `*.pem`). These files are never cached. If you mark additional files as sensitive (e.g., via a `.slashignore` file), they will also be excluded.

## Integrity & Tampering

1. **Reproducible builds** — the Dockerfile and build script are committed to the repo. Anyone can run:
   ```bash
   docker build -f ci/Dockerfile . --tag slash:verify
   docker run slash:verify /bin/slash version
   ```
   and verify the binary matches the release checksums.

2. **Signed commits** — maintainers sign commits and tags with GPG. Verify:
   ```bash
   git verify-commit <commit-hash>
   git verify-tag v1.0.0
   ```

3. **No auto-update** — Slash does not auto-update or call home for new versions. Updates are manual via marketplace or `slash upgrade`.

## Known Risks & Mitigations

| Risk | Mitigation |
|---|---|
| **Malicious hook injection** | Hosts gate hook installation behind trust prompts and permission audits. Slash source is public and reviewable. |
| **Cache poisoning** (old source retrieved after edit) | Handles are path-keyed and invalidated on file write. The cache is append-only; no stale updates. |
| **Secret leakage via compression** | `.gitignore` + secret-file pattern matching; explicit `.slashignore` for custom rules. Manual audit of cache: `slash audit`. |
| **Performance DoS** (hot-path latency) | Daemon runs warm with a ~50ms timeout on hook calls; if latency exceeds budget, the hook fails open (returns input unchanged). |
| **Privilege escalation** (daemon runs as root) | The daemon runs as the user who installed it (same privileges as your editor). No privilege escalation. |

## Auditing

Slash provides tools to audit its behavior:

```bash
# View cache contents and size.
slash cache ls

# Audit compression savings by file.
slash audit

# Check which files are cached (none for .gitignored patterns).
slash cache check <path>

# Purge cache entirely.
slash purge

# Current session: token savings, latency.
slash stats

# Check version.
slash version
```

## Telemetry (Opt-In)

Slash collects no telemetry by default. If you explicitly enable it via `.slash/config.json`:
```json
{
  "telemetry": {
    "enabled": true,
    "events": ["compression_metrics"],
    "aggregate_only": true
  }
}
```

Telemetry is **aggregate-only** and never includes source code, file paths, or model outputs. Only:
- Token reduction ratios.
- Compression method (diff-only, skeleton, JSON, log).
- Tool type (read, bash, etc.).

You can disable telemetry at any time or delete the config.

## Compliance

Slash is designed with privacy-first principles and complies with:
- **GDPR** — no personal data collected or processed by default.
- **HIPAA** / **SOC 2** — source code and computation remain local; audit log available.
- **Corporate policy** — reviewable, air-gappable, no phone-home by design.

If you are using Slash in a regulated environment, review SECURITY.md and the source code with your security team. Open an issue if you have specific requirements.

## Version Support

| Version | Supported | End of Life |
|---|---|---|
| 1.x | Yes | TBD (at least 12 months from 2.0 release) |
| 0.x (pre-release) | Community best-effort | N/A |

Security updates for 1.x will be released as patch versions (1.0.1, 1.0.2, etc.). No breaking changes in patch releases.

## Acknowledgments

We thank the security community and early users who have helped identify and responsibly report issues. Fixes are credited in release notes with permission.
