# Security architecture

Inputs, headers, bodies, and responses are untrusted. The CLI never invokes a shell to parse cURL. Secret detection runs before export. Reproduction requires an explicit write flag for mutating methods, rejects userinfo in URLs, applies response-size and timeout budgets, and checks every redirect target.

Local services bind to `127.0.0.1`. Remote mode is a separate policy boundary and requires authentication plus allowlists. Sanitization reduces disclosure risk but is not proof that a document contains no sensitive information; a human privacy review remains required before sharing.

