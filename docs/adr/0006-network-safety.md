# ADR 0006: Resolve and authorize every network destination

Status: Accepted

Initial targets and redirects pass the same scheme, port, hostname, and resolved-address policy. Server mode blocks loopback, private, link-local, multicast, and cloud metadata ranges unless a narrowly scoped local-development exception is configured.

