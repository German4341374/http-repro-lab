# Runbook: generated client failure

1. Regenerate from the same sanitized RequestSpec and record generator/runtime versions.
2. Compile with warnings treated as errors where supported.
3. Compare method, URL encoding, query order, content type, headers, and exact body at the echo server.
4. Inspect proxy variables, CA stores, automatic compression, and redirect defaults.
5. File a generator issue using only synthetic values and the minimal RequestSpec.

Never “fix” parity by embedding a captured credential or disabling TLS verification.

