# Threat model

## Assets and boundaries

Assets include credentials in captures, request/response data, generated repro packs, local workspace files, and the machine's network reachability. Boundaries are import parsing, sanitization/export, network execution, static rendering, local API, and optional persistent server mode.

## Threats and controls

| Threat | Control | Residual risk |
|---|---|---|
| Malicious HAR or cURL | Size limits, JSON parsing, no shell, unsupported chaining rejection | Parser defects and expensive inputs require fuzzing and review |
| Credential disclosure | Typed detection, placeholders, strict mode, no telemetry, restrictive file permissions | Unknown secret formats require human review |
| SSRF and DNS rebinding | Scheme/host/port policy, resolution checks, blocked ranges, redirect revalidation | DNS can change between authorization and connection; server mode needs a pinned dialer before remote exposure |
| Dangerous replay | GET/HEAD default, explicit write flag, destructive-path warning, timeouts and budgets | A nominally safe method can still trigger side effects on a poorly designed API |
| Stored/reflected XSS | Offline viewer uses text DOM APIs, no CDN, CSP in server mode | Browser extensions and local file access remain outside this boundary |
| Resource exhaustion | Input/body/response/request/time budgets and bounded API body size | Current HAR parser materializes the document once |
| Cross-host credential forwarding | Authorization and cookies removed on host-changing redirect | Other custom credential headers need organization-specific policy |
| Remote API abuse | Remote mode is not shipped as complete; loopback is the default | Do not expose the current local API to an untrusted network |

## Security properties

- Imported commands are never passed to a shell.
- Sanitizing an already sanitized RequestSpec is idempotent.
- Every automatically followed redirect is authorized independently.
- Generated artifacts contain placeholders rather than detected credentials.
- Findings cite evidence and distinguish observation from interpretation.

