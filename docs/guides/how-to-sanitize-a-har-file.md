# How to sanitize a HAR file

Use strict mode when a capture may contain personal data:

```bash
http-repro sanitize incident.har --request 42 --strict --output request.json
```

The tool replaces known credential locations and pseudonymizes supported email/phone patterns. Re-run sanitization to confirm stable output, then inspect headers, query, body, filenames, and hostnames manually.

Rule-based detection cannot identify every customer ID or secret hidden in free text. Do not treat a successful command as a guarantee. Rotate any credential that was already shared outside its intended boundary.

