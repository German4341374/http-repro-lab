# HTTP TLS debugging

Compare TLS version, cipher, certificate subject, issuer, SAN, validity, chain length, and hostname validation. Confirm SNI and system time before changing trust settings.

Differences can come from load balancers, proxy interception, client trust stores, and IPv4/IPv6 routing. A certificate near expiry is evidence for renewal work, not proof of the current HTTP failure.

Never export client private keys or solve diagnostics by disabling verification in generated code.

