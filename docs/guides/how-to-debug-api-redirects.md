# How to debug API redirects

Follow the chain entry by entry: status, method change, host, scheme, Location, Authorization, and cookies. A 302 can turn POST into GET; a cross-host hop must not inherit sensitive credentials.

HTTP Repro Lab revalidates each redirect destination and strips Authorization/Cookie on host change. Reproduce first with automatic redirects disabled so the intermediate evidence remains visible.

Redirect targets can become SSRF pivots. Never allow a public URL to redirect silently to loopback, a private service, or cloud metadata.

