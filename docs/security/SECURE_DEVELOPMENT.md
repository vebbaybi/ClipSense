# Secure Development And CI Policy

Conceptually aligned with NIST SSDF preparation/protection/production/response,
OWASP ASVS/API/file-upload guidance and SAMM incremental improvement. Not a
certification or claim of comprehensive control coverage.

## PR Checklist

- Bound scope and ownership; describe threats and handling of private data.
- Add focused behavioral/security tests and update contracts/runbooks.
- No secrets, raw exception/payload logging or high-cardinality metric labels.
- Review new dependencies, version pins, image privileges and data retention.
- Attach exact SHA, commands, CI evidence and remaining acceptance gaps.
- Keep broad parent stories open when only a subset is accepted.

## Pipeline Policy

| Tool/control | PR/integration | Release |
|---|---|---|
| Go tests/vet/govulncheck | Fail on tests or reachable vulnerabilities | Required |
| Python unittest/compile | Fail on failures | Required |
| Node tests/Next build/TypeScript | Fail on failures | Required, not browser E2E |
| Semgrep pinned local Go/Python/JS/TS rules | ERROR findings and scanner errors block | Required; narrow rules are not comprehensive SAST |
| Gitleaks full Git history, redacted report | Any finding blocks pending exact false-positive review | Confirmed secrets require revocation/remediation |
| npm audit | JSON must be complete; high/critical debt reported | High/critical findings block absent reviewed exception |
| pip-audit | Actual worker package inventory, complete JSON; unresolved advisory debt reported | Findings require severity/reachability triage and resolution/exception |
| Trivy images/IaC | All three application images scanned; complete reports required | High/critical image and applicable misconfiguration findings require resolution/exception |
| CycloneDX image SBOM | Three nonempty SBOMs retained | Inventory is not signing or provenance |
| Compose/runtime/security/monitoring | Fail on regression or incomplete evidence | Required but insufficient for creator workflow/recovery acceptance |

Bootstrap inventory window ends **2026-10-06**. Until then dependency/image findings
can be inventoried on this foundation branch, not treated as resolved or approved.
`security-policy.py --release` blocks immediately; after that date high/critical npm/
container findings and unresolved Python advisories also block integration. Scanner
failure, skipped dependency or malformed evidence always blocks. This temporary window
is not risk acceptance: final work-unit decision must disclose unresolved debt.
Before merging follow-on work, owners should replace the window with finding-specific
triage. Never renew it silently or suppress all findings by package/ecosystem.

Exceptions require finding ID, affected version/path, reachability assessment, owner,
reason, compensating control, deadline and approval. No broad suppression exists.
Semgrep local rules cover code execution, shell use, raw HTML and disabled Go TLS;
taint/interprocedural coverage remains a pre-release improvement, not implied by green.
IaC/root/base-image findings are evidence requiring review, not silently ignored safety.

## Supply Chain And Enforcement

CI uses least-privilege contents:read, no production deploy credentials, and retains
artifacts for 14 days. Work-unit workflows pin Actions to immutable revisions. Tool/image
versions are explicit, but version tags are not cryptographic provenance. SBOM includes
base/application packages and available relationships. No image signing, attestations
or release deployment exist. Scan evidence must match the candidate commit/image.

At audit, main had no branch protection. Repository owner must require CI and source
security checks and review before merge; workflow YAML alone cannot enforce branch
policy. Do not claim protected merges until GitHub configuration is verified.

## Runtime And Release

Runtime must preserve Gates 1-3 and prove telemetry privacy/correlation/scraping.
Release remains blocked on remaining dependency triage, isolation/CORS acceptance,
processing reliability, creator E2E, recovery and production configuration. Local
loopback monitoring is not production-secure deployment. Alerts need an owned delivery
channel before production operation. Stop and report gaps rather than declare release.

References: https://csrc.nist.gov/projects/ssdf,
https://owasp.org/www-project-application-security-verification-standard/,
https://cheatsheetseries.owasp.org/cheatsheets/File_Upload_Cheat_Sheet.html,
https://owaspsamm.org/.
