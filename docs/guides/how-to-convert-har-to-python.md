# How to convert HAR to Python

```bash
http-repro sanitize incident.har --request 3 --output request.json
http-repro generate request.json --language python --output generated
python generated/python/request.py
```

The generated `urllib` script sets method, headers, body, and timeout explicitly. Compare the echo receiver's method, path, query, content type, and body before targeting another environment.

Python proxy variables, CA stores, redirects, and automatic headers can differ from a browser. Preserve placeholders and avoid adding a token directly to source code or shell history.

