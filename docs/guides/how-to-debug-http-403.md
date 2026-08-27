# How to debug HTTP 403

HTTP 403 usually indicates that a request was understood but not permitted. Confirm the authenticated principal, resource, method, scope/role, tenant, and network policy.

Compare a sanitized request across environments, especially host, path, Authorization presence, and gateway headers. Do not infer that credentials are valid merely because the status changed from 401 to 403.

Use synthetic identities in shared evidence. Access-control details and customer identifiers can be sensitive even when tokens are removed.

