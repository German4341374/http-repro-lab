# Runbook: TLS failure

1. Confirm system time and the exact hostname.
2. Inspect the certificate subject, SANs, issuer, validity, and chain with a local TLS tool.
3. Compare trust stores and proxy interception between the failing runtime and browser.
4. Reproduce with a synthetic GET and capture the error code; never disable verification in a shareable client.
5. If one environment differs, verify load-balancer certificate and SNI routing.

Do not commit client keys or private CA material. A temporary insecure flag would change the behavior under investigation and is intentionally not generated.

