# Task 3 standalone slot setup extraction

RED: the untracked `internal/e2e/fullgame` setup harness could not form an atomic clean-checkout commit because `git ls-files internal/e2e/fullgame` had no baseline paths.

GREEN: moved the Task 3-only setup tests into `internal/e2e/onboarding/task3_slot_contract_test.go` and `internal/e2e/onboarding/task3_slot_public_test.go`. The unit imports synthetic playable civilization metadata, validates all 510 controller vectors for slot counts 2 through 8, and exercises public signup/lobby/game setup, AI conversion, start, ordered readback, and restart over named `memorysql` via tracked `runtime.New`.

Clean applied-patch verification at detached `f72a4e5`:

```sh
env -u GAMEPLAY_POSTGRES_DSN -u SUPABASE_DB_URL -u SUPABASE_PROJECT_URL -u SUPABASE_API_KEY -u SUPABASE_SERVICE_ROLE GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m go test -timeout=8m -p=2 ./internal/e2e/onboarding -run '^(TestOrderedSetupSlots|TestPublicOrderedSetupStartsAndRestartsForEveryCount)$' -count=1
# ok github.com/justinfarrelldev/open-ctp-server/internal/e2e/onboarding

env -u GAMEPLAY_POSTGRES_DSN -u SUPABASE_DB_URL -u SUPABASE_PROJECT_URL -u SUPABASE_API_KEY -u SUPABASE_SERVICE_ROLE GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m go build ./...
# exit 0

GIT_MASTER=1 git diff --check
# clean
```

REFACTOR/COMMIT: `3dca4e1 test(onboarding): cover ordered public setup slots`.
