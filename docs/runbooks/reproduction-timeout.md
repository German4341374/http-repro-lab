# Runbook: reproduction timeout

1. Distinguish DNS, connect, TLS, first-byte, and download phases using available timings.
2. Confirm target policy did not block a redirect.
3. Test the same sanitized request once; avoid retrying POST without idempotency review.
4. Increase `timeoutMs` only within the configured policy and record the new value.
5. Correlate a request ID with server/proxy logs.

A timeout is an observation, not proof that the application is slow. Packet loss, proxy policy, DNS, TLS negotiation, and response size can produce the same symptom.

