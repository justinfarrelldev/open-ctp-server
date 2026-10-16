# Gameplay Outbox PostgreSQL Protocol Fix Checklist

- [x] Trace the dispatcher and lib/pq protocol handling; confirm UPDATE runs while SELECT rows remain unread.
- [x] Record the focused fix plan and implementation checklist.
- [x] RED: Add and run failing read-error and real PostgreSQL regression tests. The isolated PostgreSQL test reproduced the exact `pq: unexpected Parse response 'D'` error.
- [x] GREEN: Consume and close pending rows before publishing/updating in the transaction; package and real PostgreSQL tests pass.
- [x] REFACTOR: Simplify row collection with `tx.SelectContext`, retaining ordering, locking, and retries. Restrict integration-test connections to `pg_temp`, including reconnects.
- [x] Verify scan/iteration errors and update failures roll back safely, and a failed delivery can be retried.
- [x] Refresh the dispatcher production freeze digest.
- [x] Run targeted package, PostgreSQL integration, race, freeze, and full-suite validation.
- [x] Complete final code review against the plan and record validation results.

## Validation and Review

- RED: Both partial-read failure cases failed before the fix; the PostgreSQL regression reproduced the reported protocol error.
- GREEN/REFACTOR: The events package passes with the real PostgreSQL test enabled, including the race detector. All dispatcher tests also pass for 20 consecutive race-enabled runs.
- Full `GOMAXPROCS=4 go test -timeout=16m -p=2 ./... -count=1` passed, including full-game E2E coverage. Freeze, boundary, and runtime checks passed independently; `git diff --check` passed.
- Reviewed the complete change: SELECT results are closed before any UPDATE; row locking, ordering, publish-before-mark behavior, rollback, and retries are retained. No new dependency or gameplay mechanic change requires CTP2 source verification.
- Integration data stays in a session-local temporary table. `search_path=pg_temp` applies to every test connection, so reconnects cannot touch application tables.
