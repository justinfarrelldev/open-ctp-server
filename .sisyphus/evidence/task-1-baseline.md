# Task 1 baseline ledger

- Recorded: 2026-08-10
- Branch: `game-server-world`
- Initial HEAD: `ae95c53 docs(world): mark base-game generation checklist complete`
- Initial staged diff: none.

## Initial dirty status

`git status --short` was recorded before any Task 1 commit. The complete resulting path matrix is in `task-1-atomic-commits.txt`.

## Commands and results

| Command | Exit | Result |
|---|---:|---|
| `go build ./...` | 0 | Green baseline. |
| `env -u GAMEPLAY_POSTGRES_DSN -u SUPABASE_DB_URL -u SUPABASE_PROJECT_URL -u SUPABASE_API_KEY -u SUPABASE_SERVICE_ROLE GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 8m go test -timeout=7m -p=2 ./internal/e2e/fullgame -run '^TestPublicExecutorCompletesMinimalScenarioThroughProductionRuntime$' -count=1` | 1 | Row 1 exact red below. |
| `env -u GAMEPLAY_POSTGRES_DSN -u SUPABASE_DB_URL -u SUPABASE_PROJECT_URL -u SUPABASE_API_KEY -u SUPABASE_SERVICE_ROLE GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m go test -timeout=8m -p=2 ./internal/e2e/fullgame -run '^TestCatalogHasExactlySixtyFixedCoveredCases$' -count=1` | 1 | Catalog exact red below. |

## Expected red signatures

```text
--- FAIL: TestPublicExecutorCompletesMinimalScenarioThroughProductionRuntime
    fullgame_test.go:59: capture city: no visible legal frontier
```

```text
--- FAIL: TestCatalogHasExactlySixtyFixedCoveredCases
    fullgame_test.go:13: scenario fullgame-01-sequential-bloodlust_conquest has non-observed rules fixture hash
```
