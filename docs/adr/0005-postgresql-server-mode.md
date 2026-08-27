# ADR 0005: PostgreSQL is optional server-mode persistence

Status: Accepted

The CLI works without a database. PostgreSQL stores sessions and evidence for the local control plane when durable history is required. Migrations are forward-only and state is never embedded in source control.

