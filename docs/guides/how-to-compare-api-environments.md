# How to compare API environments

```bash
http-repro compare request.json --target-a https://staging.example.invalid --target-b https://api.example.invalid --output comparison.json
```

Review status, functional/security/cache headers, JSON types, response size, TLS metadata, redirects, and timings. Report an observed difference, evidence, possible interpretation, and suggested verification in that order.

Dynamic fields such as request IDs can create noise. Never send production credentials to staging merely to make a comparison symmetrical.

