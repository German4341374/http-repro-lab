# Security policy

## Supported versions

The `main` branch receives security fixes during pre-1.0 development. Tagged releases will document their support window here.

## Reporting a vulnerability

Use GitHub private vulnerability reporting for this repository. Do not open a public issue containing an exploit, production HAR, cookie, token, host inventory, or personal data. Include a synthetic reproduction, affected commit, impact, and suggested mitigation when possible.

Expected response targets are acknowledgement within five business days and an initial assessment within ten. These are maintainer goals, not a paid support SLA.

## Security assumptions

The CLI is local-first. Remote exposure is not supported until token authentication, destination allowlists, a DNS-pinning dialer, and rate limiting are complete. Sanitization is defense in depth and does not replace manual privacy review.

