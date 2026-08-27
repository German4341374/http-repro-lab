# How to debug HTTP 500

An HTTP 500 is server-error evidence, not a root cause. Preserve a sanitized request, response content type, request ID, time, and target environment.

```bash
http-repro reproduce request.json --target https://staging.example.invalid --output response.json
```

Correlate the request ID with application/proxy logs and compare deployed configuration and schema. Minimize only against a safe test environment with an explicit request budget.

Error bodies and logs often contain personal data or connection details; redact before attaching them to a ticket.

