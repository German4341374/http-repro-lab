# How to convert HAR to cURL

Normalize and generate from a selected request:

```bash
http-repro sanitize incident.har --request 3 --output request.json
http-repro generate request.json --language curl --output generated
```

Inspect `generated/curl/request.sh`. It contains placeholders, an explicit timeout, and no shell expansion from the input capture. Run it only against an intended target after reviewing method and body.

Copying “cURL” directly from an untrusted ticket can execute shell syntax. HTTP Repro Lab parses supported cURL as data and rejects command substitution/chaining; it never passes imported text to a shell.

