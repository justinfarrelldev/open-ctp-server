# Decisions


## DSN-free production start/world repair — 2026-07-19

Diagnosis: the first  branch in  returned success without mutating the named game. The real production start handler therefore correctly saw the default setup state and returned HTTP 409. The driver now persists setup fields, game lifecycle status/started timestamp, revisioned initial gameplay state, and start receipt/event/outbox rows transactionally; it leaves status changes to the production status UPDATE. Regression coverage asserts the setup state directly, valid handler start/world success, repeat-start HTTP 409, and world read from a new handle/runtime over the same named store.

Verification: ok  	github.com/justinfarrelldev/open-ctp-server/internal/e2e/onboarding	0.137s
ok  	github.com/justinfarrelldev/open-ctp-server/internal/e2e/memorysql	0.002s exited 0; telemetry  remained nil in the public route test.  exited 0.

## DSN-free production start/world repair — 2026-07-19

Diagnosis: the first `UPDATE games SET style = ...` branch in `memorysql.exec` returned success without mutating the named game. The real production start handler therefore correctly saw the default setup state and returned HTTP 409. The driver now persists setup fields, game lifecycle status/started timestamp, revisioned initial gameplay state, and start receipt/event/outbox rows transactionally; it leaves status changes to the production status UPDATE. Regression coverage asserts the setup state directly, valid handler start/world success, repeat-start HTTP 409, and world read from a new handle/runtime over the same named store.

Verification: `env -u GAMEPLAY_POSTGRES_DSN -u SUPABASE_DB_URL -u SUPABASE_PROJECT_URL -u SUPABASE_API_KEY -u SUPABASE_SERVICE_ROLE GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m go test -timeout=8m -p=2 ./internal/e2e/onboarding ./internal/e2e/memorysql -count=1` exited 0; telemetry `Unknown` remained nil in the public route test. `GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m go build ./...` exited 0.

## Task 17 Row 2 navigation/lifecycle decision — 2026-08-10

Kept the executor unchanged as an AI scheduler/observer. The production planner only receives `visibility.AIInput` plus its own accepted movement history, uses stale memory solely to rank currently legal visible movement, and emits embassy commands only for current visible cities. Terminality now travels on the existing transactional command-result/receipt contract, so the production driver—not the runner—ends scheduling immediately after the natural victory command.

## Task 3 setup contract placement — 2026-08-11

Placed the reusable setup contract in tracked `internal/e2e/setup` rather than the untracked inherited `internal/e2e/fullgame` tree. This keeps production packages free of E2E imports, lets onboarding and future fullgame orchestration share one validator, and preserves a dependency-complete clean-parent commit boundary without staging unrelated Task 17 files.

## Task 4 diagnostic and puppet boundary — 2026-08-11T11:47:48Z

Diagnostics return purpose-built copied values rather than `engine.State`, raw JSON, pointers, or mutable database records. Puppet inputs are only serialized public projection/history bytes plus a clearly named `TestHarnessPolicyLevel`; selection calls unchanged production `ai.Plan`, while mutation uses an authenticated `HumanHTTPClient` posting the ordinary public command envelope. Production AI remains isolated on `AIGateway.Submit` -> `ExecuteAICommand`, and `ObserveAICommand` remains internal instrumentation.

## Task 4 public history contract — 2026-08-11T12:07:29Z

The puppet public-history byte contract is a monotonic array of actor-visible move steps (`command_type`, `destination_tile_id`, `sequence`). It maps to `ai.ExplorationStep` and feeds unchanged production `ai.PlanWithHistory`; unknown fields, non-move entries, empty destinations, and non-increasing sequences fail before HTTP submission. This keeps history visibility-safe while matching production planning semantics.

## Task 5 scope decision — 2026-08-11T12:19:00Z

Retain only the evidence-backed runner guard and direct tests. Do not suppress the empty-result frontier error, alter combat composition/vector, change seed or budgets, or generalize the executor to recover from the deterministic attacker loss. Because those constraints prevent Row 1 completion, record a blocker and do not commit the untracked fullgame patch as accepted Task 5 work.

## Task 5 final scope decision correction — 2026-08-11T12:35:23Z

Supersede the blocker conclusion, not the constraints. Initial army composition now consumes all sorted, eligible, same-owner, co-located, unassigned units visible in the public bootstrap projection; exploration recovery consumes only projected terrain and the runner's public trail. Empty capture results still use the original projection fallback, and AI remains on the production gateway. Freeze the state/event literals only after the test itself proves equality across two checkpoint-0 clean runs and one checkpoint-2 restart run.

## Task 6 final contract decision — 2026-08-11T13:39:00Z

Keep acceptance fixtures catalog-independent and policy-explicit. Preserve Task 5's phase-bounded command/turn semantics rather than applying a later shared cross-phase counter that changes accepted behavior. Remove unsupported decorative rule ID/profile labels; bind the remaining rules hash to `ServerOwnedSyntheticRuleset(victory).ContentHash`. Treat active-turn restart cursor equality as monotonic and require the later terminal event to advance. Validate AI provenance from production observer submissions, never from runner-authored commands.

## Task 6 typed restart and observer correction — 2026-08-11

Acceptance restart configuration is an explicit AcceptanceRestartPlan: disabled for clean runs, after-durable-round for conquest, and one-based after-completed-turn for diplomacy. Policy validation rejects incompatible restart points and invalid turn indexes. AI provenance requires exact observer transcript length equality with the accepted AI command count, rejecting both missing and extra submissions. Active-turn restart evidence comes from TestPublicRuntimeResolvesSimultaneousQueueAfterRestart (Last-Event-ID 9 resumes exactly event 10), not Row 2.

## Task 6 post-review count independence — 2026-08-11

The AI observer is evidence, not the source of the expected command count. Each drive now compares its observer delta with the production driver result.Commands value before accumulating that independent count, including terminal diplomacy. The conquest runner no longer retains an integer checkpoint field; its restart callback branches directly on the validated typed restart point.

## Task 6 observable restart boundaries — 2026-08-11

Acceptance restart evidence is now emitted as a typed AcceptanceRestartEvent only after rebuilding the runtime. The clean/clean/restart tests assert one conquest durable-round event and one diplomacy completed-turn-12 event, so deterministic artifact equality cannot pass if either declared rebuild is removed or shifted.

## Task 7 defect/source classification — 2026-08-11T14:26:00Z

No CTP2-specific production discrepancy was found, so no pinned-source inference or production edit was made. The PostgreSQL lock-race result was classified as a stale optional test fixture because its ad-hoc schema omits the migration-backed `event_id`; Task 7 scope forbids changing source for that harness-only defect. Coverage and formatting debt are reported as blockers rather than weakened, auto-formatted, or repaired out of scope. Task 8 lock generation and Task 9 catalog authoring remain untouched.

## Task 7 post-fix scope decision — 2026-08-11T15:09:30Z

Raise only the committed HEAD aggregate coverage policy using asserted behavior/error/persistence/restart tests; do not enforce or edit the inherited dirty engine threshold. Repair only the test-private lock-race DDL to match accepted `event_id` storage, and apply behavior-preserving gofmt to every reported path. Preserve Row 1/Row 2 fixtures, budgets, restart plans, terminal logic, production behavior, and the untracked fullgame boundary.

## Task 8 — 2026-08-11
The manifest is intentionally independent of Git history and mutable environment state. Campaign work is allowlisted outside the freeze package; hash regeneration is not implemented, and any production-path change is an explicit freeze-reopen blocker. The guard documents that review is still required against malicious self-editing.

## Task 8 correction — 2026-08-11
The reviewed allowlist is a canonical ordered contract in code and JSON. CI runs the freeze package on every pull request through the existing coverage workflow, while all pre-existing coverage policy lines remain untouched.

## Task 8 retry 2 — 2026-08-11
The freeze CI invocation uses both an outer GNU timeout and an inner Go timeout with the repository resource cap. No coverage threshold or existing workflow command was changed.

## Task 9 catalog scope decision — 2026-08-29T02:15:27Z

The final catalog validator was split conceptually into `ValidateCatalogSchema` and `Validate`. This keeps Task 9 honest: schema/family assembly can be green while `Validate(Catalog)` still fails closed on unfinished immutable artifacts. Resignation was kept conservative because the pinned source evidence supports conquest/elimination and disconnect-to-robot continuity, but not a source-compatible resignation-defeat terminal. Row 1 was aligned to the existing accepted public executor fixture semantics and hashes, while the remaining rows were moved into the required files without inventing reviewed winners or hashes.

## Task 9 retry final decision — 2026-08-29T02:52:00Z

Task 9 retry is being returned as blocked, not green. I kept the useful schema split, corrected Row 1 identity drift (stable ID plus separate historical acceptance name), added a focused public blocker test, and refused to mark Rows 03–12 complete after timed public probes and AI-path failures. This preserves evidence and the freeze boundary without fabricating catalog artifacts.

## Task 9 AI open fallback decision — 2026-08-29T06:57:11Z

Kept the fix inside the two locked production files. The visibility layer became the sole authority for whether planning closeout is `open`, `queued`, `closed`, or omitted, and the AI planner consumes only that projected wire signal when choosing its score-100 fallback. This preserves sequential, long-turn, and hybrid-conflict behavior because omitted status still maps to the existing `pass` fallback, while simultaneous/hybrid open planning now closes through the ordinary `round-complete` gateway path with stable receipts and restart behavior.

[2026-08-29T10:25:45Z] Gaia persistence decision
- Stored a single absolute gaia_completion_turn on Nation with zero as inactive so countdown cancellation, reacquisition, and canonical restart remain deterministic without introducing a broader research tree.


## Task 9 Row 04 acceptance decision — 2026-08-29T10:25:45Z

Row 04 is now pinned as immutable accepted catalog data with observed artifacts rather than provisional placeholders. The catalog test asserts the stable ID, hybrid/tropical/six-slot/scientist script envelope, completed-turn-12 restart plan, and exact observed hashes so future drift fails at the narrow row-specific guard instead of a generic validator.

## Task 9 Row 05 acceptance — 2026-08-29T10:46:00Z

Bounded failure before correction:
- `GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m env TASK9_ROW=5 go test -timeout=8m -p=2 ./internal/e2e/fullgame -run TestObserveTask9Row -count=1`
- Exit 1 with `mixed-controller command budget exhausted in mode sequential at turn 12 revision 242 after 240 accepted commands (budget 240); actor=account-2`.
- Recent transcript proved ordinary autonomous AI moves/passes plus only one embassy path, so the old eight-player continuity fixture could not satisfy reciprocal embassies inside the unchanged 240-command/48-turn/2-minute envelope.

Correction rationale:
- Row 05 was a catalog mismatch, not a production defect. The row declared `diplomatic_world_union` but used the continuity/leave fixture family (`HumanPolicyContinuity` + `scripts/resignation-continuity.json`) that exists to exercise public continuity behavior, not diplomatic reachability.
- Accepted Row 02 already proved the truthful public diplomacy harness. Row 03 and Row 04 already preserve distinct continuity and larger-roster coverage, so Row 05 was reduced to the smallest honest sequential diplomatic fixture that still adds distinct mode coverage: sequential + 3 slots (human, ai, ai) + `scripts/city-economy-wonder.json` + `HumanPolicyDiplomacy`.

RED/GREEN evidence:
- Added `TestCatalogRowFiveDeclaresSequentialDiplomacyReachability`.
- RED command: `GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 5m go test -timeout=4m -p=2 ./internal/e2e/fullgame -run TestCatalogRowFiveDeclaresSequentialDiplomacyReachability -count=1`
- RED failure under old declaration: `catalog row 5 world drift` showing the old huge/arid/eight-slot/continuity fixture.
- GREEN command: same command exited 0 after the Row 05 declaration changed to the sequential three-player diplomacy fixture.

Accepted Row 05 public artifacts after equality gate:
- Winner: `account-1`
- Reason: `diplomatic_world_union`
- State SHA-256: `5b4deeb4f0a5106ed1351a6656f7e6db2fdca1de1571772ae78be3572dd861e8`
- Event SHA-256: `e3509ea099a3c7bf025683a4378539e40fe4769f615c5d31f8978cf597a44cf9`
- AI commands: `125`
- Restart event: `{Point:after-completed-turn CompletedTurn:12}`

Equality/verification commands:
- `GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m env TASK9_ROW=5 go test -timeout=8m -p=2 ./internal/e2e/fullgame -run TestObserveTask9Row -count=1`
- `GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m env TASK9_ROW=5 go test -v -timeout=8m -p=2 ./internal/e2e/fullgame -run TestObserveTask9Row -count=1`
- `GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m go test -timeout=8m -p=2 ./internal/e2e/fullgame -count=1`
- `GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 15m go test -race -timeout=13m -p=2 ./internal/gameplay/... ./internal/runtime ./internal/e2e/fullgame -count=1`
- `GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m go test -timeout=8m -p=2 ./internal/game ./internal/gameplay/... ./internal/world ./internal/runtime ./internal/e2e/memorysql ./internal/e2e/onboarding ./internal/e2e/fullgame -count=1`
- `GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m go build ./...`
- `GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 5m go vet ./internal/e2e/fullgame`
- `GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 5m go test -timeout=4m -p=2 ./internal/e2e/freeze -run '^TestEngineLock$' -count=1`
- `git diff --check`

## Task 9 Row 06 acceptance decision — 2026-08-29T07:50:58Z

Kept the fix entirely in unlocked E2E harness code. The real defect was not score timing but long-turn continuity staging: the leave actor's public transport/revision was not preserved cleanly across scheduler cycles, and the first continuity spread cycle was allowed to fall through into generic human planning. The minimal truthful fix was therefore continuity+long_turn only: retain the leave actor's `HTTPExplorationTransport`, carry its base URL through restart, reuse its revision across scheduler cycles, and stop the cycle immediately after accepted preparation moves.

Only after the full-executor RED turned green was Row 06 promoted from the old two-settler `scripts/resignation-continuity.json` fixture to the truthful three-settler `scripts/score-resignation-continuity.json` fixture. The world (`duel`/`cold`/wetness 90), stable ID, long-turn mode, two slots, default budgets, and restart-at-completed-turn-12 plan stayed unchanged.

Row 06 is now pinned with observed immutable artifacts only because the exact observer proved `clean-1 == clean-2 == restart-12`: winner `account-1`, state `5292b679fa07ec631929e4f4eb36805d89abe2ba00c2ccb2dbe0430c616477b5`, events `b8e31801117572d3cdb441fc16eb3d4eb42d409a953fe74761637573a5017282`, AI commands `232`, terminal reason `score_2300`. The generic provisional-artifact fail-closed mutation moved from Row 06 to Row 07 so final catalog validation remains strict.
- Accepted Row 07 uses  with 12 public round closes before the existing science sequence so restart-12 is crossed truthfully under unchanged budgets.
- Accepted Row 07 uses scripts/row07-delayed-city-economy-wonder.json with 12 public round closes before the existing science sequence so restart-12 is crossed truthfully under unchanged budgets.

- 2026-08-29: Abandoned speculative Row 08-only harness forks (delayed row08 script and script-only human mode) after they failed verification or produced AI no-progress. Reverted those experiments and preserved only defects demonstrated by focused RED tests.
