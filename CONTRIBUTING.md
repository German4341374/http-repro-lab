# Contributing

Thank you for improving HTTP Repro Lab. Start with a synthetic fixture and an issue that describes the observable behavior. Never submit a production capture or credential.

## Development

1. Fork and create a focused branch.
2. Run `make setup` on Linux/WSL2, or install the tools listed in the README on Windows.
3. Add unit and adversarial tests before changing an importer, sanitizer, redirect policy, or generator.
4. Run `make verify` where Docker is available.
5. Use a Conventional Commit and explain security/semantic trade-offs in the pull request.

Generated client changes must demonstrate semantic parity for method, path, query, headers, body, and content type against the echo server. New input formats must normalize into RequestSpec rather than bypassing sanitization.

## Pull requests

Keep changes reviewable, update documentation and project status, do not weaken a safety gate for convenience, and do not add telemetry or network dependencies to the core CLI.

