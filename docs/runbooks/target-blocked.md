# Runbook: target blocked

`TARGET_BLOCKED` means the destination failed scheme, host, port, or resolved-address policy.

1. Record the exact error without adding credentials.
2. Resolve the hostname independently and inspect every IPv4/IPv6 address.
3. Confirm the target is the intended environment and not metadata, loopback, link-local, or a private service reached through an unexpected DNS answer.
4. For the bundled local lab only, repeat with `--allow-private`.
5. In server mode, change a narrow host/port allowlist rather than globally allowing private networks.

Do not bypass the gate merely because the hostname looks public; redirect and DNS answers are the security boundary.

