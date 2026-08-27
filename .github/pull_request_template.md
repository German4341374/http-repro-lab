## Summary

Describe the observable problem and the focused change.

## Verification

- [ ] Tests cover success and malformed/adversarial input.
- [ ] Generated-client changes preserve method, path, query, headers, body, and content type.
- [ ] `make verify` passed, or the unavailable checks are listed with reasons.
- [ ] Documentation and `docs/project-status.md` match the implementation.

## Security and privacy

- [ ] No real credentials, captures, personal data, or private infrastructure details are included.
- [ ] Imported cURL remains data and is never executed through a shell.
- [ ] Network, redirect, write-method, timeout, and size policies are not weakened.

