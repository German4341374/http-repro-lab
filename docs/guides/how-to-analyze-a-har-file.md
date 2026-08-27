# How to analyze a HAR file safely

A HAR records browser HTTP traffic and often embeds headers, cookies, and bodies. Keep it local and begin with a summary:

```bash
http-repro har summary incident.har
http-repro analyze incident.har --strict --output report
```

Review failures, slow entries, domains, and sensitive-value detections. Open the offline report and connect every hypothesis to an entry, status, header, or timing value. A 401 is authentication evidence, not proof of an expired token.

Pitfalls include incomplete exports, cached responses, service-worker traffic, and connection reuse that hides timing phases. Never publish the original HAR; sanitize and manually inspect a repro pack first.

