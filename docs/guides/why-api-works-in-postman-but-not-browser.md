# Why an API works in Postman but not a browser

Browsers enforce CORS, mixed-content, cookie SameSite/Secure rules, forbidden headers, preflight requests, and origin-based credential policy. A Node or Postman request does not reproduce that security model.

Compare the actual browser HAR with the non-browser request: method, URL, Origin, preflight status, cookies, redirect chain, and CORS response headers. Generate Node Fetch only for server-side parity; do not call it a browser reproduction.

HAR cookies and tokens are sensitive. Sanitize locally before comparing or sharing.

