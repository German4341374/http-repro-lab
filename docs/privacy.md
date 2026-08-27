# Privacy model

HAR, Postman, cURL, and raw HTTP artifacts can contain bearer tokens, cookies, passwords, private keys, customer identifiers, form values, and response bodies. Treat an unreviewed capture like a credential file.

HTTP Repro Lab runs locally by default. It has no telemetry, analytics, tracking pixel, cloud upload, external processing, or automatic credential reuse. The standard profile replaces common credential locations. Strict mode additionally pseudonymizes email and phone patterns while preserving repeated-value correlation.

Sanitization has limits. Organization-specific identifiers, secrets with unusual names, binary uploads, and values embedded in free text can evade rules. Before sharing:

1. Run strict sanitization.
2. Review the sensitive-values view and every body.
3. Exclude uploaded files unless explicitly required.
4. Search the pack for known hostnames, user names, and identifiers.
5. Rotate any credential that already left its intended boundary.

Never attach a production capture to a public GitHub issue. Reproduce the shape with synthetic values instead.

