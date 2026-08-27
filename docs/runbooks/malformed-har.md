# Runbook: malformed HAR

1. Preserve and hash the original; do not edit it in place.
2. Validate that the file is UTF-8 JSON and contains `log.entries`.
3. Check whether browser export was interrupted or the capture was truncated by a ticketing system.
4. Export a fresh synthetic capture from the source browser.
5. If the failure remains, reduce the structure without retaining credentials and open an unsupported-HAR issue.

Do not paste a production HAR into an online formatter. Local JSON validation avoids another disclosure boundary.

