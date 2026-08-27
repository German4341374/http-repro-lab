# How to create a minimal API reproduction

Begin with a sanitized RequestSpec and a deterministic predicate such as `status=500`. Remove optional headers, query parameters, JSON fields, and array items one stable step at a time, keeping a change only when the predicate remains true.

Use a strict request/time budget and a test environment. The Python minimizer currently exposes the pure JSON algorithm; network execution remains behind the Go safety policy.

A minimal reproduction is easier to review and share, but it can change timing or authorization behavior. Never minimize a mutating production request without explicit authority and idempotency protection.

