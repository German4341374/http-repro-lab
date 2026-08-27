# Runbook: database unavailable

The CLI remains usable without PostgreSQL. Server-mode history does not.

1. Check `/health` and `/ready` separately.
2. Verify `pg_isready`, connection limits, disk space, and migration state.
3. Do not replay a partially committed migration; inspect the transaction result first.
4. Restore connectivity, then restart only the API process.
5. Reconcile sessions that were accepted in memory while persistence was unavailable once durable repositories are enabled.

Back up before destructive recovery. Credentials stay in environment variables or a secret store, never a runbook command.

