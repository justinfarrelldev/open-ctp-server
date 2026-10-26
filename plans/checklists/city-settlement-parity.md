# City settlement parity checklist

- [x] Read existing settlement investigation and released-source settlement/site checks.
- [x] Scope site repair and document plan in `plans/city-settlement-parity.md`.
- [x] RED: engine site, legacy/configured pin, atomicity and existing safeguard regressions (`/tmp/ctp-settlement-red.log`).
- [x] RED: advertised-action execution, viewer privacy and AI site regressions; observed the advertised foreign-site rejection, missing influence/adjacency checks and conflicting AI policy.
- [x] GREEN: shared original-compatible site rule and authoritative enforcement; focused legacy/configured-pin, replay/restart and guard regressions pass.
- [x] GREEN: viewer-safe action filtering and matching AI checks; advertised foreign-site payload executes and privacy regressions pass.
- [x] Correct old adjacent-founding unit fixtures and three-city continuity harness. Added a RED regression for distinct-but-adjacent sites, legal visible routes/transit and phase-correct movement refresh closeout. Use a deterministic non-tied seed for the bounded terminal-opener test; do not claim old campaign goldens still apply.
- [x] REFACTOR: reuse the existing tile index during action construction; review shared policy, the released army/event/player/unit/create-city/insert-city chain, bounded adjacency, privacy, legacy compatibility and deterministic replay. Rejection precedes all unit/city mutations.
- [x] Refresh production freeze manifest (102 entries) and investigation documentation (`docs/city-settlement-parity.md`); retain the original report as historical evidence.
- [x] Focused settlement engine/visibility/AI regressions and continuity harness regressions pass after refactoring.
- [x] `GOMAXPROCS=4 go test -timeout=16m -p=2 ./... -count=1` passes, including freeze and fullgame package tests (`/tmp/ctp-settlement-final-tests.log`).
- [x] `go test -race -timeout=10m -p=2 ./internal/gameplay/engine ./internal/gameplay/visibility ./internal/gameplay/ai ./internal/gameplay/service -count=1` passes.
- [x] `go vet ./...`, `go build ./...`, gofmt and `git diff --check` pass.
- [x] Both new engine site functions have 100% statement coverage. Whole-engine package coverage is 86.5%; the pre-existing 90% engine coverage gate remains unmet. No claim that the complete coverage task passes.
- [x] Record limits: opt-in historical catalog goldens not regenerated, no live game97 acceptance, no deployment, no client edits and no live/database writes.
- [x] Commit atomic repair with unchanged pre-commit hook; report verification and remaining limits without claiming deployment or live acceptance.
