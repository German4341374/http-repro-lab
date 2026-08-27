# How to convert cURL to Go

Treat the command as quoted data:

```bash
http-repro import --curl 'curl -H "Accept: application/json" https://example.invalid/health' --output request.json
http-repro generate request.json --language go --output generated
go run generated/go/main.go
```

The Go output uses context cancellation, an explicit timeout, error handling, and a redirect policy. Unsupported flags or shell syntax fail closed.

Verify URL encoding and repeated headers at the echo server. Do not paste a real Authorization value into a public issue or generated source.

