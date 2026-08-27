# How to debug HTTP 429

Capture status 429, `Retry-After`, and provider-specific limit headers. Check whether the limit applies per token, IP, tenant, route, or time window.

Do not hammer the endpoint to confirm a rate limit. Use the deterministic `/rate-limit` lab or a controlled low request count. Honor `Retry-After` and retry only methods whose idempotency semantics are understood.

Rate-limit headers can reveal account identifiers or quotas; sanitize them before sharing.

