# Project status

Status as of 2026-08-27. This file is the source of truth when a roadmap item and an implementation claim appear to conflict.

## Implemented

- Go CLI commands: `analyze`, `har summary`, `import --curl`, `sanitize`, `reproduce`, `compare`/`diff`, `generate`, `pack`, `validate`, and `version`.
- HAR request normalization; a safe subset of cURL flags; canonical RequestSpec v1.
- Rule-based secret and strict PII detection, stable replacement, and sanitized exports.
- Policy-gated HTTP execution with destination resolution checks, write-method approval, redirect revalidation, size limits, cancellation, explicit timeouts, basic timings, TLS metadata, and response hashing.
- Evidence-backed HAR findings and response comparison for statuses, selected headers, TLS, size, and JSON types.
- Eight generated clients, a deterministic mock API, offline report, and repro ZIP with SHA-256 checksums.
- Fastify in-memory analysis sessions, pagination, finding filters, one terminal SSE event, health/readiness/metrics, uniform errors, and security headers.
- Deterministic Python JSON minimizer with execution budgets.
- Reference Java, C#, and PHP runners with tests; PostgreSQL schema and retention migration.
- CI definitions for language checks, migrations, generated clients, Docker, security, SBOM, and release binaries.
- Reproducible 10k/50k/100k HAR parse and offline-report benchmarks with measured local results.

## In progress

- Connecting the Fastify control plane to PostgreSQL repositories and full job cancellation.
- Rich interactive request-detail and side-by-side comparison views. The shipped viewer is a read-only offline report.
- Broader HAR timing, redirect, cookie-transition, authentication, rate-limit, and request-family findings.
- Iterative/streaming HAR decoding to reduce the measured linear allocation cost.

## Planned

- Raw HTTP, Postman v2.x, and OpenAPI operation import.
- Network-backed query/header/multipart minimization through the Go safety boundary.
- JWT metadata viewer, repeated latency samples, richer DNS/TLS comparison, and browser Playwright flow.
- Authenticated remote mode, complete API endpoint set, OpenTelemetry hooks, Helm chart, optional Terraform example, and documentation site deployment.
- Signed repro manifests and external importer/generator plugin discovery.

## Blocked verification

- The local Windows host used for the 2026-08-27 verification did not have WSL installed, so Docker Desktop's engine remained stopped. `docker compose config` succeeded; image builds and Compose health are delegated to the Docker GitHub Actions workflow until a WSL2-capable local engine is available.
- No `v1.0.0` release is created. The complete Definition of Done, including Docker and clean-checkout validation, has not yet been met.
