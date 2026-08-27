# How to debug HTTP 401

HTTP 401 means the server requested authentication. Inspect `WWW-Authenticate`, whether Authorization exists after sanitization, redirect host changes, and JWT metadata without claiming signature validity.

```bash
http-repro analyze auth.har --strict --output report
```

Compare issuer, audience, scope, expiry, proxy stripping, and credential-forwarding policy between environments. A decoded JWT is untrusted data until a signature is verified with the correct key.

Never include the token in a report. Replace it with `${AUTH_TOKEN}` and rotate it if exposed.

