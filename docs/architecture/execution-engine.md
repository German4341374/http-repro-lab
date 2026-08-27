# Execution engine

The Go engine accepts only a validated RequestSpec. It checks the method before DNS, resolves the destination, blocks non-public ranges by default, applies an absolute context timeout, caps captured bytes, and re-runs destination policy for every redirect. Authorization and cookies are removed on a host-changing redirect.

`httptrace` records available DNS, connect, TLS, write, and first-byte timestamps. Connection reuse can omit phases, so empty timing fields mean “not observed,” not zero network cost. Response metadata includes a content hash, declared/observed size, truncation flag, resolved addresses, selected text body, headers, redirects, and peer TLS metadata.

Local fixtures require `--allow-private`; mutating methods require `--allow-write`. These independent gates prevent a convenience flag from silently expanding both network and method authority.

