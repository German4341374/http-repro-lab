# How to reproduce API errors

Start with a sanitized, read-only request and a controlled target:

```bash
http-repro reproduce request.json --target https://staging.example.invalid --output response.json
```

For POST/PUT/PATCH/DELETE, review side effects and add `--allow-write`; for the bundled loopback lab also add `--allow-private`. Record status, selected headers, timings, redirect chain, and request ID.

One successful or failed request is a diagnostic sample. Avoid automatic retries for non-idempotent operations and never reuse captured credentials without explicit authorization.

