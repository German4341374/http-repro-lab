# Verification record

Observed locally on 2026-08-27 on Windows with Go 1.26.7, Node.js 24, Python 3.14, .NET 8, Java 20, PHP 8.4, Maven 3.9.11, and Composer 2.10.2.

Passed commands:

- `go test ./cmd/... ./internal/...`
- `go vet ./cmd/... ./internal/...`
- API Prettier, ESLint, strict TypeScript check, 8 Vitest tests, build, and `npm audit` with zero known advisories
- Python Ruff, strict mypy, and 11 Pytest tests
- `.NET` Release test run: 2 passed
- Maven verify: 2 JUnit tests passed
- PHPUnit: 2 tests / 2 assertions; PHPStan level max passed; Composer reported no PHP advisories
- `docker compose config --quiet`
- `actionlint` and configured `yamllint`
- Gitleaks 8.28.0 scanned all local commits and reported no leaks
- Local HAR analysis, strict sanitization, POST reproduction, report generation, client generation, and repro pack generation
- cURL, JavaScript, TypeScript, Python, Go, Java, C#, and PHP generated clients executed against the local echo server and retained method/query semantics
- Two local environment variants produced the expected comparison exit code 1 plus content-type and JSON `$.id` type evidence
- Three one-operation benchmark samples for 10k/50k/100k HAR parsing and report generation; medians and caveats are recorded in `benchmarks/results.md`

Not passed locally:

- Docker image and Compose runtime checks because Docker Desktop had no usable engine while WSL was absent.
- PostgreSQL migrations against a local server; the integration workflow provisions PostgreSQL 17 for this check.
- Clean-checkout release verification; no release was created.
- The Go race detector requires a C compiler that was not available on this Windows host; the Linux reusable CI workflow runs the race-enabled suite.
