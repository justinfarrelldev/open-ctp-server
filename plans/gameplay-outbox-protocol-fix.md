# Gameplay Outbox PostgreSQL Protocol Fix

## Goal

Stop repeated `pq: unexpected Parse response 'D'` failures in the gameplay event dispatcher without changing event ordering, wakeup semantics, row locking, or retry safety.

## Scope and Execution

1. RED: Add dispatcher regression coverage for reading all pending rows before any wakeup/update, read failures, and real PostgreSQL delivery/restart behavior. Use the existing optional `GAMEPLAY_POSTGRES_DSN` integration-test convention and session-local temporary data only.
2. GREEN: Fully consume and close the locked outbox SELECT result before publishing wakeups or executing delivery UPDATEs on the same transaction connection.
3. REFACTOR: Simplify result collection using the existing sqlx APIs, preserving `FOR UPDATE SKIP LOCKED`, transaction lifetime, ordered delivery, and publish-before-mark semantics.
4. Verify rollback/no-publication on read failures and rollback/retry on update failures. Refresh the production freeze digest, run targeted, integration, race, and full-suite checks where feasible, and review the final diff.

## Review Boundaries

- No database reset, application-data mutation in tests, new dependencies, migration, worker timing change, or gameplay mechanic change.
- This is server-side PostgreSQL protocol handling; CTP2 gameplay source parity is not affected.
- Keep `plans/checklists/gameplay-outbox-protocol-fix.md` current and do not change or bypass commit hooks.
