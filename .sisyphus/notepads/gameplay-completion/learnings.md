# Learnings

## Task 17 — 2026-07-19

The existing Task 17 scaffolding is reusable: `internal/gameplay/testkit` supplies deterministic IDs, clocks, RNG, deadlines, and failure evidence; `internal/e2e/onboarding` supplies public signup/lobby/game HTTP and runtime restart; `internal/e2e/fullgame` supplies projection-only scripts, HTTP/SSE commands, terminal parsing, and canonical audit hashes. It currently validates only catalog shape/value presence, not immutable artifact completeness or all required executions.

## Correction and prerequisite repair — 2026-07-19

`advanceRound` already restores non-transport movement from `runtime.Starter.Movement`; a clean local public conquest reaches natural terminal with the pinned case-1 hashes. The actual minimal AI issue was its fallback `noop`: it could not invoke lifecycle evaluation. RED changed `TestPlannerUsesOnlyProjectionAndStableCandidateTuples` to require `CommandRoundComplete`; GREEN changed only `ai.Plan`'s fallback candidate. Both the bounded AI unit test and local `TestPublicRuntimeDrivesAIThroughOrdinaryGateway` passed using `GAMEPLAY_POSTGRES_DSN` after a local reset.


## E2E Supabase override — 2026-07-19

User override applied to `internal/e2e/onboarding`: the SQL/Supabase live test routes were replaced with a deterministic in-memory HTTP runtime test double. `TestMemoryPublicRuntimeNeverUsesExternalConnector` uses a fake connector and asserts its call count remains zero after the public signup/lobby/game/world journey; `TestMemoryPublicRuntimeRestartRebuildsHandlerFromPersistedState` rebuilds a handler from shared persisted fake state after restart. GREEN command (with all Supabase variables unset): `GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m go test -timeout=8m -p=2 ./internal/e2e/onboarding -run '^TestMemoryPublicRuntime' -count=1` exited 0. A rescan of `internal/e2e` found no `sqlx.Connect`, `database/sql`, DSN environment variables, or Supabase references.

## Factory-backed memory SQL seam — 2026-07-19

`internal/runtime.New` is the shared production composition factory. It preserves default production rate limiters, while onboarding injects high-capacity limiter instances through `Config.NewLimiter`. The test-only named `memorysql` driver retains state across reopened handles and drives real account/session/lobby/game/start/world handler SQL. `TestPublicRuntimeCreatesLobbyWithNamedMemoryStore` performs two signup calls, lobby creation/join, game creation/setup/country selection/start, projected-world read, repeated-start conflict, and restart read using the same named store. Verified with all Supabase environment variables unset: `GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m go test -timeout=8m -p=2 ./internal/e2e/onboarding -count=1` exited 0. Static E2E scan found no DSN, Supabase, PostgreSQL open, or test `HandleFunc` route registrations.

## Production-handler stub boundary — 2026-07-19

The rejected hand-written E2E router was removed. `internal/e2e/onboarding/production_runtime_test.go` now constructs the existing `NewPublicRuntime` around `sqlmock` wrapped in `sqlx.NewDb`: it exercises the unchanged production `/account/create_account` handler through `httptest`, verifies a session is issued, and verifies every expected storage operation. This reaches no PostgreSQL/Supabase service and has no DSN configuration. Command: `env -u GAMEPLAY_POSTGRES_DSN -u SUPABASE_DB_URL -u SUPABASE_PROJECT_URL -u SUPABASE_API_KEY -u SUPABASE_SERVICE_ROLE GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m go test -timeout=8m -p=2 ./internal/e2e/onboarding -count=1` exited 0.

## MemorySQL public setup milestone — 2026-07-19

Added direct ordered lobby_members driver coverage (stable insertion order plus empty sql.ErrNoRows) and extended the named-memory production handler smoke route through two-user lobby/game setup/country selection. The DSN-unset bounded onboarding command now reaches the real world-bootstrap guard; no external DB connector was used.

## DSN-free production start/world repair — 2026-07-19

Diagnosis: the first  branch in  returned success without mutating the named game. The real production start handler therefore correctly saw the default setup state and returned HTTP 409. The driver now persists setup fields, game lifecycle status/started timestamp, revisioned initial gameplay state, and start receipt/event/outbox rows transactionally; it leaves status changes to the production status UPDATE. Regression coverage asserts the setup state directly, valid handler start/world success, repeat-start HTTP 409, and world read from a new handle/runtime over the same named store.

Verification: ok  	github.com/justinfarrelldev/open-ctp-server/internal/e2e/onboarding	0.137s
ok  	github.com/justinfarrelldev/open-ctp-server/internal/e2e/memorysql	0.002s exited 0; telemetry  remained nil in the public route test.  exited 0.

## DSN-free production start/world repair — 2026-07-19

Diagnosis: the first `UPDATE games SET style = ...` branch in `memorysql.exec` returned success without mutating the named game. The real production start handler therefore correctly saw the default setup state and returned HTTP 409. The driver now persists setup fields, game lifecycle status/started timestamp, revisioned initial gameplay state, and start receipt/event/outbox rows transactionally; it leaves status changes to the production status UPDATE. Regression coverage asserts the setup state directly, valid handler start/world success, repeat-start HTTP 409, and world read from a new handle/runtime over the same named store.

Verification: `env -u GAMEPLAY_POSTGRES_DSN -u SUPABASE_DB_URL -u SUPABASE_PROJECT_URL -u SUPABASE_API_KEY -u SUPABASE_SERVICE_ROLE GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m go test -timeout=8m -p=2 ./internal/e2e/onboarding ./internal/e2e/memorysql -count=1` exited 0; telemetry `Unknown` remained nil in the public route test. `GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m go build ./...` exited 0.

## Task 17 — memorysql gameplay persistence — 2026-07-19

`memorysql` now persists and reads the production gameplay aggregate join (including immutable rules pin), receipts, ordered events, outbox delivery state, SSE cursor/resync/terminal state, and fullgame audit state/events. Transaction rollback restores aggregate and persisted record mutations; reopening the same named store retains committed data. `TestPublicRuntimeExecutesGameplayCommandAndResumesSSEFromNamedMemoryStore` starts a real game through `internal/runtime.New`, submits HTTP `pass`, and receives sequence 2 through `Last-Event-ID: 1` SSE. DSN-unset verification: `GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m go test -timeout=8m -p=2 ./internal/e2e/memorysql ./internal/e2e/onboarding -count=1` exited 0.


## Task 17 public deterministic runner — 2026-08-10

`PublicExecutor` now completes the fixed sequential military scenario entirely through `internal/runtime.New` production handlers over a named `memorysql` store: public signup/lobby/setup/start, SSE bootstrap/resume, restart, visible-only role resolution/exploration, real deterministic AI `army-create` plus `round-complete` commands, natural `bloodlust_conquest` terminal SSE, and DB-backed state/event audits. The two clean store executions match frozen artifacts: state `e7bd74d9fcf55f321d2acf453c566c61c441d53c1f264752fcd7b9662635088a`, events `0e3af0f37e06b445f9e6c0ad67b9fa8cf21231b9ba50df2381af6a88ced83034`. DSN-unset focused tests and `go build ./...` passed.


## Persisted turn-policy state — 2026-08-10

Canonical `engine.State` now carries a typed `TurnPolicyState` (queued commands, early-end markers, deadline, conflict components, stage/index, recalculation count). Canonical clone/JSON normalization deep-copies it and sorts unordered early-end/queue/component-member values; validation rejects unknown queued actors/targets, invalid deadlines, duplicate early ends/component members, invalid counters, and components referencing unknown players. `State.AsTurnPolicyState` rebuilds the existing `turns.State` resolver input exclusively from persisted canonical data, including active player state and deep-copying all slices. Verification: bounded `go test ./internal/gameplay/engine`, `go test ./internal/gameplay/service`, and DSN-unset `go build ./...` passed.


## Replayable policy queues and hybrid wars — 2026-08-10

`TurnPolicyCommand` now persists the command ID, idempotency key, canonical JSON payload, stable receipt sequence, and policy ordering metadata. `TurnPolicyState` now persists war edges and `AsTurnPolicyState` deep-copies queue payloads, receipts, and wars into `turns.State`. Canonical validation rejects invalid/missing command identity or payload, unknown targets, and invalid war endpoints; canonical normalization copies payload bytes and orders war pairs. Engine policy tests and the complete engine package pass under bounded commands.


## Policy replay integration — 2026-08-10

`engine.Apply` now separates preparation, validated mechanics dispatch, and finalization. Simultaneous/hybrid actions persist a canonical `engine.Command` envelope (including idempotency key) in policy state; resolution decodes it and reuses validated mechanics dispatch rather than calling direct state mutators. Queue receipt sequence derives from the accepted event sequence, and the resolver’s stable tuple remains authoritative. The service authorizes every eligible planner in simultaneous/hybrid planning while keeping the existing single aggregate transaction, receipt, event, outbox, and post-commit behavior.


## Long-turn deadline worker — 2026-08-10

Long-turn deadline authority remains in canonical `policy.deadline`; `store.FindDueLongTurnDeadlines` queries that JSON state without adding a second deadline column. `service.DeadlineWorker` takes its clock from injected engine sources, derives a stable receipt key from game/player/deadline, and calls ordinary `ExecuteCommand` with `pass`; the receipt and aggregate revision guards make retries, restarts, and manual-pass races no-ops. Runtime wires its post-commit dispatcher callback, and the DSN-free public runtime exposes a test trigger. Public HTTP/SSE coverage advances `FakeClock` without sleeps, observes exactly event ID 2 after two worker passes, and keeps `memorysql` free of unknown queries. Changed-package tests and `go build ./...` passed. The repository-wide test command still has unrelated Task 17 fixture/game-route failures.


## Simultaneous public replay resolution — 2026-08-10

The public `round-complete` route now invokes the existing `CommandRoundComplete` policy path. A two-player DSN-free production test queues real `army-create` commands through HTTP, persists both early-end markers, and verifies that the second early-end performs one revision-5 resolution. `resolvePlanning` finishes policy state before the engine replays queued commands, so replay must temporarily restore `Phase.ResolutionLocked`; `activeActor` recognizes this locked simultaneous-only internal replay context while ordinary live commands remain guarded. The test restarts the public runtime and resumes SSE at cursor 4, receiving only event ID 5 with no unknown memory SQL. Affected engine/turns/http/service/memorysql/onboarding tests and `go build ./...` passed.


## Three-player starter ownership — 2026-08-10

The DSN-free production runtime preserves one owned starter for each of three canonical players through start and restart. The apparent `account-2` loss occurred only after the test intentionally passed the human turn and invoked the real AI driver: its declared `found-city` command correctly consumed that starter into an `account-2` city. The regression now asserts starter ownership before and after restart, then separately verifies the legal post-pass AI city action.


## Task 17 Row 2 projection/cursor repair — 2026-08-10

`HTTPExplorationTransport` now retains a monotonic SSE `Cursor` alongside its revision and projection. `Bootstrap` sends the current cursor plus `X-Gameplay-Resync: current`; the production SSE handler returns a current projected resync frame for that request. The transport rejects a resync whose cursor or revision regresses, or whose equal revision has different bytes. `CurrentProjection` is used before each human action, after every AI drive, on restart, and for failure diagnostics. Focused transport RED/GREEN tests prove a stale resync cannot replace a newer projection and a following one-step move sends the refreshed `expected_revision`.

The human scheduler still submits exactly one projected path tile. Its public-frontier sweep now reaches visible `ai-city-settler-2-1` and `ai-city-settler-3-1`, and the public viewer reported both `account-2` and `account-3` in `embassy_player_ids`; no canonical state was read or hidden coordinates selected.


## Row 2 AI frontier experiment — 2026-08-10

A projection-only AI frontier planner based on current/stale visibility and one-step `visibility.Plan` paths was added with a RED deterministic non-bounce fixture. It preserves the existing embassy-first candidate priority. In the live Row 2 transcript, `account-2` and `account-3` both autonomously discovered `fullgame-city-human-1` and submitted ordinary production-gateway `establish-embassy` commands; `account-3` also autonomously established an embassy with `account-2`.

## Task 17 AI exploration persistence — 2026-08-10

Added `gameplay_ai_exploration_history`, keyed by `(game_id, actor_player_id, sequence)`, with only accepted AI move type and destination step. `PersistCommandInTx` writes it after aggregate/receipt validation in the same transaction; memorysql supports rollback/restart parity. `AIGateway` resolves `(game_id, account_id)` to the gameplay-player primary key before reading scoped history, so production account IDs are never assumed to equal game-player IDs. `Driver` passes only current `visibility.Projection` bytes to tactical planning and applies history solely to exclude recorded destination steps when a different currently legal visible route exists.

Focused DSN-free verification passed: `GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m go test -timeout=8m -p=2 ./internal/gameplay/ai ./internal/gameplay/store ./internal/gameplay/service ./internal/e2e/memorysql -count=1`; `go build ./...` also passed.

## Row 1 victory rules selection — 2026-08-10

A declared scenario victory is now passed through the public `/game/update_rules` configuration endpoint as `victory:<id>`. Production start parses selected evaluator tokens, creates a hash-verified canonical pin with `enabled_victory_evaluator_ids`, and runtime evaluation therefore retains only the selected evaluator(s); unselected score/diplomatic/gaia evaluators remain available when explicitly configured. `memorysql` persists the production `rules_json` update transactionally across restart. Victory selectors are excluded only from the world-generator seed, retaining the fixed map while the configured selector remains in the persisted setup and canonical gameplay pin. Verification passed: `go test ./internal/gameplay/rules ./internal/gameplay/engine ./internal/gameplay/service ./internal/e2e/memorysql ./internal/e2e/onboarding -count=1`, `go test ./internal/world -count=1`, `go test ./internal/gameplay/events -count=1`, Row 2 `-count=2`, and `go build ./...` under the required bounded commands.

## Task 17 Row 2 production navigation and terminal signal — 2026-08-10

RED: the exact DSN-free public runtime duplicate campaign exhausted all 48 turns. Its transcript showed route candidates ranked a distant visible endpoint while submitting only `path[:1]`, letting the planner repeatedly use already-recorded first hops. A later run proved that a terminal created by an accepted AI pass was persisted but not reported by `Driver.Drive`, causing the next human HTTP command to receive 409.

GREEN: `ai` now ranks the actual current visible hop it submits (including novelty, frontier, stale-memory pursuit coverage, and embassy approaches); stale cities remain ineligible for embassy payloads. Its deterministic actor-derived exploration lane diversifies post-contact coverage without player/coordinate/action hard-coding. `service.Result` now persists `Terminal` in the normal receipt result, `AIGateway` passes it through, and `Driver` returns `ai.ErrTerminal` immediately after the accepted terminal command.

Verification: `GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m go test -timeout=8m -p=2 ./internal/gameplay/ai ./internal/gameplay/service -count=1` passed. The exact DSN-free `TestPublicExecutorCompletesRowTwoThroughProductionRuntime` passed twice as two independent invocations; each test invocation itself runs two clean namespaces and compares the complete `Artifacts` values, including terminal SSE outcome, audit state/event SHA-256 hashes, and the full autonomous AI transcript. `GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m go build ./...` passed.

## Row 1 city-directed projected frontier attempt — 2026-08-10

RED/GREEN focused coverage was added in `TestCityDirectedFrontierUsesReachableVisibleStep`: when the current foreign city tile lacks planning due to an unavailable full route, the human selector ranks only current visible tiles by Manhattan distance to that current city coordinate, then accepts only a non-empty `visibility.Plan` result and submits its first step. The focused test passed after implementation. This remains human E2E routing only; it does not inspect canonical state or direct AI.


## Eight-player public campaign setup — 2026-08-10

`TestPublicCampaignSetupEightPlayersOneHumanRestAI` exercises the production HTTP setup path over a named `memorysql` store: eight signups, one lobby owner plus seven joins, public game configuration with two settlers, `campaignCountriesFor(8)` selection, exactly seven post-`ReadGame` `AddAI` conversions, start, and a new runtime over a reopened named store. The public `/game/get_game` DTO intentionally does not select or serialize `controller_kind`; the test therefore verifies the one-human/seven-AI composition from the successful public conversion sequence while asserting the restarted public roster order, countries, account identities, started state, settlers, and zero unknown memory SQL on both stores.

Actual command/result: `env -u GAMEPLAY_POSTGRES_DSN -u SUPABASE_DB_URL -u SUPABASE_PROJECT_URL -u SUPABASE_API_KEY -u SUPABASE_SERVICE_ROLE GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m go test -timeout=8m -p=2 ./internal/e2e/fullgame -run '^TestPublicCampaignSetupEightPlayersOneHumanRestAI$' -count=1` exited 0: `ok github.com/justinfarrelldev/open-ctp-server/internal/e2e/fullgame 0.220s`.


## Eight-player controller assertion correction — 2026-08-10

`TestPublicCampaignSetupEightPlayersOneHumanRestAI` now asserts actual `gameDTO.Players[index].Controller` values (`human` for the owner, `ai` for indices 1–7), includes index/account/country and actual/expected controller in mismatches, and counts only returned controller values before and after named-store restart. The test no longer uses a local conversion flag as proof of controller persistence.


## Game-player controller readback repair — 2026-08-10

`internal/game/helpers.go` now includes `controller_kind` in the ordered `gamePlayersSelectByGameID` projection, which maps directly to `GamePlayer.Controller`. The matching named-memory `game_players` SELECT now returns the same ninth column and stored controller (defaulting an absent legacy value to `human`). The existing eight-player production HTTP regression verifies actual `human`/`ai` values and counts before and after reopening the named store, without local controller inference.

Actual command/result: `env -u GAMEPLAY_POSTGRES_DSN -u SUPABASE_DB_URL -u SUPABASE_PROJECT_URL -u SUPABASE_API_KEY -u SUPABASE_SERVICE_ROLE GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m go test -timeout=8m -p=2 ./internal/game ./internal/e2e/memorysql ./internal/e2e/fullgame -run '^TestPublicCampaignSetupEightPlayersOneHumanRestAI$' -count=1` exited 0: `ok github.com/justinfarrelldev/open-ctp-server/internal/game 0.003s [no tests to run]`; `ok github.com/justinfarrelldev/open-ctp-server/internal/e2e/memorysql 0.004s [no tests to run]`; `ok github.com/justinfarrelldev/open-ctp-server/internal/e2e/fullgame 0.216s`.


## Eight-player all-human public setup — 2026-08-10

Added `TestPublicCampaignSetupEightPlayersAllHumanScripted` beside the one-human/rest-AI regression. It runs eight public signups, owner lobby plus seven joins, production game setup with two settlers, `campaignCountriesFor(8)` country selection, start, and a runtime rebuilt over the same named memory store. It makes zero `AddAI` calls and asserts ordered account/country roster, eight actual `human` controllers, started setup state, and nil unknown-SQL telemetry before and after restart.

Actual command/result: `env -u GAMEPLAY_POSTGRES_DSN -u SUPABASE_DB_URL -u SUPABASE_PROJECT_URL -u SUPABASE_API_KEY -u SUPABASE_SERVICE_ROLE GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m go test -timeout=8m -p=2 ./internal/e2e/fullgame -run '^(TestPublicCampaignSetupEightPlayersOneHumanRestAI|TestPublicCampaignSetupEightPlayersAllHumanScripted)$' -count=1` exited 0: `ok github.com/justinfarrelldev/open-ctp-server/internal/e2e/fullgame 0.417s`.


## Three-player half-human public setup — 2026-08-10

Added `TestPublicCampaignSetupThreePlayersHalfHumanHalfAI`, which defines the odd composition boundary as `humanCount = (countries+1)/2` and verifies it yields two humans for `campaignCountriesFor(3)`. The real public flow signs up three actors, creates the lobby/game, configures/selects countries, converts only index 2 through one public `AddAI` call, starts, and reads the reopened named-memory runtime. Both reads assert ordered account/country roster, actual `human,human,ai` controllers, and nil unknown SQL telemetry.

Actual command/result: `env -u GAMEPLAY_POSTGRES_DSN -u SUPABASE_DB_URL -u SUPABASE_PROJECT_URL -u SUPABASE_API_KEY -u SUPABASE_SERVICE_ROLE GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m go test -timeout=8m -p=2 ./internal/e2e/fullgame -run '^TestPublicCampaignSetupThreePlayersHalfHumanHalfAI$' -count=1` exited 0: `ok github.com/justinfarrelldev/open-ctp-server/internal/e2e/fullgame 0.148s`.


## Deterministic catalog controller assignments — 2026-08-10

Added private `controllerAssignmentsFor(composition, count)` beside `campaignCountriesFor`. It accepts only the supported roster counts (2, 3, 4, 6, 8) and `one-human-rest-ai`, `half-human-half-ai`, or `all-human-scripted`; it returns newly allocated stable human/AI index partitions. Direct table tests cover every composition/count pair, the odd three-player split (`[0,1]` human and `[2]` AI), full partition/no-overlap validation, fresh-slice mutation isolation, and invalid count/composition rejection. No public setup flow uses the helper yet.

Actual command/result: `env -u GAMEPLAY_POSTGRES_DSN -u SUPABASE_DB_URL -u SUPABASE_PROJECT_URL -u SUPABASE_API_KEY -u SUPABASE_SERVICE_ROLE GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m go test -timeout=8m -p=2 ./internal/e2e/fullgame -run '^TestControllerAssignmentsFor' -count=1` exited 0: `ok github.com/justinfarrelldev/open-ctp-server/internal/e2e/fullgame 0.004s`.


## Generalized public campaign setup — 2026-08-10

`setupCampaignGame` now owns the production public setup sequence for any supported roster/composition: it obtains countries through `campaignCountriesFor`, controller index partitions through `controllerAssignmentsFor`, signs up and joins the ordered roster, configures/selects, resolves public account IDs, calls `AddAI` only for assigned AI indexes, starts, and returns typed ordered player records plus owner/game ID. The legacy two-player and diplomatic wrappers delegate using their existing one-human/rest-AI effective behavior, preserving Row 1/Row 2 composition behavior. `TestPublicCampaignSetupThreePlayersHalfHumanHalfAI` now exercises the helper with its true half-human assignment and verifies typed result/actual `human,human,ai` public readback before and after named-store restart.

Actual focused command/result: `env -u GAMEPLAY_POSTGRES_DSN -u SUPABASE_DB_URL -u SUPABASE_PROJECT_URL -u SUPABASE_API_KEY -u SUPABASE_SERVICE_ROLE GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m go test -timeout=8m -p=2 ./internal/e2e/fullgame -run '^(TestControllerAssignmentsFor|TestPublicCampaignSetupThreePlayersHalfHumanHalfAI)$' -count=1` exited 0: `ok github.com/justinfarrelldev/open-ctp-server/internal/e2e/fullgame 0.128s`.

## Task 1 preservation — 2026-08-10

Task 1 preserved six clean-parent units: lobby response (`19a4327`), world seed normalization (`59e091b`), deterministic rules (`56f9cf0`), canonical engine (`4f9461b`), rules import CLI (`dabe0ae`), and visible AI/event projection (`d4b7521`). Every unit passed its bounded focused test, `go build ./...`, and `git diff --check` in a fresh applied-patch worktree.

## Task 2 bridge and migration baseline — 2026-08-10

RED exposed stale eight-column SQL mocks and an empty-roster panic in `loadStartControllers`. GREEN centralized ordered/account player SQL contracts with persisted `controller_kind`, added empty-roster fail-closed handling, and kept named memorysql query behavior aligned. Fresh and legacy schema verification now runs through `supabase/tests/task2/run_isolated_postgres.sh`, which asks Docker for a dynamic loopback port and removes its container/anonymous volume in an EXIT trap. Focused bridge/memorysql, isolated migrations, clean-worktree builds passed. Commits: `be6fe00`, `2194180`.

## Task 2 runtime route-contract repair — 2026-08-10

Replaced the brittle `main.go` string scan with `internal/runtime` httptest coverage against `runtime.New(...).Handler`. The contract proves a deliberately incomplete mux is rejected, requires lobby chat plus world/space routes, and rejects `/game/chat` and `/world/chat` as chat surfaces. The mandated game/memorysql command, build, and isolated fresh/legacy migrations passed. Follow-up commit: `e9b22cb`.

## Task 3 source evidence — 2026-08-10

Pinned original CTP2 source was inspected without copying records: `gstypes.h` retains a 32-player engine capacity; `ns_gamesetup.h` defines eight network setup slots and a separate four-human UI cap; `gameinit.cpp:2432-2454` independently applies each hotseat slot's human/robot setting and civilization selection. The civilization data schema stores a civilization index alongside country/name fields, and the default civilization data is record-based. The replacement setup must therefore model ordered slots independently from a human-count policy and derive selectable IDs from validated imported metadata.

## Task 3 slot setup verification — 2026-08-10

`SetupSlot{Slot, CountryRecordID, ControllerKind}` now validates ordered one-based slots, 2–8 cardinality, unique imported selectable IDs, and only `human|ai`. `Civilisation.ctp2synthetic` importer support requires boolean `Playable`, `Special`, and `Barbarian` metadata, then exposes only playable non-special/non-barbarian IDs. Tests cover all 510 controller vectors for n=2..8 and DSN-unset public setup/start/restart for every count including 5 and 7. Focused rules/fullgame command and Row 2 preservation command passed; `go build ./...` passed; scan of the affected paths found no retail marker or former hard-coded country identifiers.

## Task 3 standalone setup extraction — 2026-08-10

Moved the Task 3-only roster and public setup tests from the untracked fullgame tree into the self-contained `internal/e2e/onboarding/task3_slot_{contract,public}_test.go` unit. It calls tracked `runtime.New` with named `memorysql` and a high-capacity test limiter, not an E2E router, and imports synthetic selectable metadata through the committed rules importer. A detached worktree at `f72a4e5` accepted only the staged two-file patch: DSN-unset focused onboarding tests, `go build ./...`, and `git diff --check` passed without the untracked fullgame tree.

## Task 3 reusable explicit slot contract — 2026-08-11

`internal/e2e/setup` now owns the reusable one-based `Slot{Slot, CountryRecordID, ControllerKind}` contract and validates it against importer-backed playable civilization metadata. Onboarding public lifecycle tests consume that contract through `runtime.New`, authenticated HTTP, named `memorysql`, start, reopen, and ordered country/controller/account readback. Direct tests cover all 510 vectors, invalid slot/count/country/controller cases including special/barbarian records, and ten imported playable synthetic IDs; public tests persist every one across restart. Clean-parent focused tests, `go build ./...`, and diff checks passed for commits `c989bbb`, `b95509e`, and `a5747a7`.

## Task 4 test-only boundaries — 2026-08-11T11:47:48Z

Moved deterministic clocks, IDs, RNG, policy checks, barriers, and failure evidence from `internal/gameplay/testkit` to tracked `internal/e2e/testkit`; all six workspace consumers now use the E2E path. Added detached SQL-backed diagnostic values, public-byte-only puppet planning through production `ai.Plan`, authenticated human HTTP submission, hidden-canary checks, and production import/route/topic/config guards. Focused tests and `go build ./...` passed both in the working tree and once per atomic clean-parent patch. Commits: `504c49d`, `ffa0a52`, `a562569`, `0dd5455`.

## Task 4 review correction — 2026-08-11T12:07:29Z

Post-implementation review found that public puppet history was initially validated but not consumed and that the wrong-actor test accepted any error status. Commit `212e070` now decodes a strict serialized public exploration-step schema, invokes production `ai.PlanWithHistory`, and proves a visited tile changes selection to a fresh visible route. Public HTTP integration now asserts exact 422 for an out-of-turn member, 403 for a nonmember, and revision advancement only for the authorized actor. Goal and code-quality re-reviews passed.

## Task 5 public capture acknowledgement — 2026-08-11T12:19:00Z

A focused runner decision now treats non-empty public `CombatTransition.Result.CapturedCityID` as authoritative success and never uses the ID for lookup; nil/empty results invoke the existing projection fallback unchanged. Direct tests, Row 2, combat HTTP refresh, projected city-frontier selection, engine/service/visibility/AI packages, build, and diff checks pass.

## Task 5 Row 1 deterministic completion correction — 2026-08-11T12:35:23Z

The earlier single-attacker diagnosis was incomplete: the public bootstrap projection exposed additional eligible, co-located, unassigned human starter units, but initial `army-create` selected only one. Building the army from the sorted public unit set, then using only projected known land plus the runner's public movement trail for stable exploration/backtracking, allows ordinary public combat to reach natural `bloodlust_conquest`. Two clean executions and one checkpoint-2 runtime restart produced identical complete `Artifacts`: state `6fbb8ce4a34d95058afa995f3b9bad15e21af0e97aabd1b7023e63f513a8dcb6`, events `104c65dc8ea937e023b0b6ba8918d5ef75a00f769c9264df09f9227ad4d3b169`. Row 2, the focused matrix, race matrix, and build pass.

## Task 6 executor refactor — 2026-08-11T13:06:00Z

`PublicExecutor.Execute` now consumes an explicit `AcceptanceConfig` and delegates to a configured human harness policy; it no longer infers behavior from scenario ID or dimensions. Standalone Row 1/Row 2 fixtures live outside catalog validation, and Row 2 passes two clean runs plus one Last-Event-ID restart run. Contract/failure tests, focused non-blocked suite, race, build, static provenance scans, and diff checks pass.

## Task 6 correction and final verification — 2026-08-11T13:39:00Z

The Row 1 regression came from a later recreation of `runner.go` that changed accepted phase-bounded budgets into one shared cross-phase turn counter. Restoring separate configured exploration, capture, and progression phase budgets recovered the exact Task 5 three-run behavior and hashes without changing any budget value. Acceptance rules now retain only operational fields: victory/script plus a SHA-256 hash verified against the production synthetic rules pin. AI observer entries carry validated game/actor/idempotency provenance; restart cursors cannot regress and terminal SSE must advance. Focused, race, build, formatting, LSP, and diff gates pass.

## Task 7 production acceptance matrix — 2026-08-11T14:26:00Z

Task 2 fresh/legacy migrations passed on dynamic port 32768 and cleaned their container/volume. Bridge, memorysql, counts 2–8 and all 510 controller vectors, imported playable civilizations, public setup/start/restart, Task 4 boundaries, all production gameplay/world/runtime packages, Row 1 and Row 2 independently/together, explicit restart/replay/deadline/SSE/authorization, both race matrices, broad classification, build, vet, protected-data, import, and diff gates passed. Row 1 retained state `6fbb8ce4a34d95058afa995f3b9bad15e21af0e97aabd1b7023e63f513a8dcb6` and events `104c65dc8ea937e023b0b6ba8918d5ef75a00f769c9264df09f9227ad4d3b169`; Row 2 complete artifacts remained equal across clean and completed-turn-12 restart runs. The unfiltered suite had only the Task 9 `non-observed rules fixture hash` failure. A disposable worktree hidden-canary perturbation failed the intended public-projection predicate and was removed.

## Task 7 post-fix acceptance PASS — 2026-08-11T15:09:30Z

Meaningful memorysql production-runtime/persistence/error/deadline tests, events broker/SSE boundary tests, puppet HTTP tests, world fixture tests, and a visible embassy-approach test raised HEAD-policy aggregate coverage from 76.484% to 80.4% with only the Task 9 catalog validator excluded. The stale PostgreSQL lock-race private schema now includes migration/store-backed `event_id text`; its dynamic-port Postgres test passed and cleaned its container/volume. Gofmt is empty across tracked plus inherited fullgame/onboarding files. Final focused, Row 1/Row 2 independent+together, restart/replay, migration, non/full race, build, vet, boundary/protected, LSP, and broad-only-Task9 classification gates passed.

## Task 8 — 2026-08-11
The freeze guard enumerates fully locked runtime/gameplay roots plus the selected world/game/migration/main paths, hashes regular files directly, and parses Go imports with go/parser. Temporary copied roots are sufficient to test post-freeze mutation without relying on Git refs or touching inherited dirty work.

## Task 8 correction — 2026-08-11
The campaign write allowlist must be validated as an exact ordered sequence, not merely checked for presence. Length, order, duplicate, and value tampering now produce indexed diagnostics, and the PR coverage workflow invokes the focused freeze test.

## Task 8 retry 2 — 2026-08-11
Static allowlist paths must be mandatory during enumeration; otherwise deleting both a file and its manifest row can evade count comparison. Lstat prevents symlink substitution, and path.Clean plus explicit backslash rejection keeps canonical paths portable.

## Task 9 catalog schema split — 2026-08-29T02:15:27Z

`internal/e2e/fullgame/catalog.go` now owns the nested catalog schema and shared acceptance-facing definitions (world/rules fixture, restart plan, budgets, policy kind, terminal oracle, observed artifacts). The 60 stable row declarations are split across `catalog_rows_01_12.go`, `catalog_rows_13_24.go`, `catalog_rows_25_36.go`, `catalog_rows_37_48.go`, and `catalog_rows_49_60.go`. `ValidateCatalogSchema` now enforces stable IDs/seeds, explicit ordered slots, rules-pin hashes, policy-specific restart semantics, public-only script targeting, required value coverage, and mode/victory coverage without pretending unfinished rows are accepted. Row 1 now carries the pinned accepted immutable hashes in catalog form; Rows 02–60 remain honest provisional rows with no partial fake artifacts.

## Task 9 retry blocker proof — 2026-08-29T02:52:00Z

Row 1 remained reproducible only when its historical acceptance namespace `fullgame-minimal-public-executor` was preserved separately from the stable catalog ID `fullgame-01-sequential-bloodlust_conquest`; changing the namespace changed deterministic artifacts. Row 2 also reproduced cleanly with immutable hashes: state `e12b2a683037003ab4ad6b5c61867c7c408df43a37d5c45594714452aaea2204`, events `ccc11fb47c241e516693700c5d0454d29ee6f70710a9a8567bdbcd35c6af01fc`.


## Task 9 pending takeover completion — 2026-08-29T03:36:04Z

`PendingTakeoverWorker` is now restart-safe authoritative background work: `store.PendingTakeoversQuery` discovers only canonical human players with `takeover_pending=true` in started games, the worker submits the ordinary `CommandTakeoverAI` transaction with stable `takeover:<game>:<player>:<revision>` keys, and policy-boundary/duplicate/stale conflicts are intentionally harmless retries. `runtime.New` wires the worker beside the dispatcher, `main.go` runs it in production, and DSN-free public continuity coverage now proves `leave` stays public while the later automatic controller change reaches both SSE and persisted `game_players.controller_kind` without exposing `/takeover-ai`.


## Task 9 TDD evidence correction — 2026-08-29T03:49:27Z

A genuine behavioral RED for autonomous takeover can be reconstructed safely from the current dirty GREEN workspace by copying it to a temp tree and withholding only the autonomous completion behavior there while keeping the continuity tests compiling. Tightening `TestPublicLeaveTransitionsToAutonomousAIWithoutPublicTakeoverCall` and the restart companion to assert `controller after worker = "human", want ai autonomous takeover after public leave` makes the RED failure name the actual missing behavior directly, before the later SSE assertion.


Task 9 continuation after parent timeout — 2026-08-29T04:22:30Z
- Parent polling expired during the row-shard investigation phase, specifically while iterating the `TASK9_ROW=3` observation path after a `driveAIPlayers` continuity fix.
- Preserved valid partial work: Row 01/02 artifacts, separate Row 01 acceptance namespace, continuity worker tests, continuity harness policy, starting-settler support, and fail-closed provisional rows.
- Current shortest repros after the latest continuity reshaping:
  * `env -u GAMEPLAY_POSTGRES_DSN -u SUPABASE_DB_URL -u SUPABASE_PROJECT_URL -u SUPABASE_API_KEY -u SUPABASE_SERVICE_ROLE TASK9_ROW=3 GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 12m go test -timeout=10m -p=2 ./internal/e2e/fullgame -run '^TestObserveTask9Row$' -count=1 -v` -> exit 1, safe diagnostic `context deadline exceeded` after the public leave/takeover handoff.
  * `env -u GAMEPLAY_POSTGRES_DSN -u SUPABASE_DB_URL -u SUPABASE_PROJECT_URL -u SUPABASE_API_KEY -u SUPABASE_SERVICE_ROLE TASK9_ROW=6 GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 12m go test -timeout=10m -p=2 ./internal/e2e/fullgame -run '^TestObserveTask9Row$' -count=1 -v` -> exit 1, same handoff stall.
- Updated hypothesis: the remaining blocker is no longer pending takeover itself, but deterministic AI-phase continuation for post-leave sequential/long-turn continuity rows under the frozen public contract. No unobserved row hashes were written.


Task 9 design correction pass — 2026-08-29T04:35:00Z
- Restored truthful Rows 01–12 ID/mode pairings and representative composition diversity in `catalog_rows_01_12.go`; the stable IDs no longer lie about turn mode.
- Restored required policy coverage in the row envelope: one-human/rest-AI, all-AI observe-only, alternating/test-controlled humans, and designated continuity rows.
- Corrected continuity harness design in `executor.go`: only the designated leave actor is converted through `RunPendingTakeovers`; surviving humans remain available for projection-derived public actions where present.
- Removed temporary `zz_task9_observe_test.go` before final reporting.
- Focused fail-closed verification remains green: `go test ./internal/e2e/fullgame -run '^(TestCatalogSchemaAssemblesExactlySixtyStableCases|TestCatalogRowOneMatchesAcceptedFixture|TestCatalogRows01_12CoverTaskNineRosterAndPolicyEnvelope|TestCatalogRejectsProvisionalScenarios|TestValidate.*|TestPublicLeave.*)$' -count=1`, `go test ./internal/e2e/freeze -run '^TestEngineLock$' -count=1`, `git diff --check`, and LSP diagnostics all pass.
- Rows 03–12 remain unobserved after the earlier bounded continuity probes; no hashes were written for them.

## Task 9 mixed-controller scheduler — 2026-08-29T01:30:45Z

Focused TDD status in this session: RED cases now exist for survivor repetition, AI-only fail-close with humans present, and mode-specific closeout. GREEN refactor replaced the continuity one-shot handoff with , added script-only leave handling for the designated quitter, re-read public controllers after takeover, removed blanket one-settler overrides from catalog Rows 03–11, and made row coverage require one-human/rest-AI, all-AI observe-only, alternating mixed, scripted-human, and continuity-with-survivor cases. Fast focused checks are green for the catalog family contract, survivor repetition, and AI-only fail-close.

## Task 9 mixed-controller scheduler — 2026-08-29T01:30:45Z

Focused TDD status in this session: RED cases now exist for survivor repetition, AI-only fail-close with humans present, and mode-specific closeout. GREEN refactor replaced the continuity one-shot handoff with runMixedControllerCampaign, added script-only leave handling for the designated quitter, re-read public controllers after takeover, removed blanket one-settler overrides from catalog Rows 03–11, and made row coverage require one-human/rest-AI, all-AI observe-only, alternating mixed, scripted-human, and continuity-with-survivor cases. Fast focused checks are green for the catalog family contract, survivor repetition, and AI-only fail-close.

## Task 9 command-budget fail-close — 2026-08-29T01:43:30Z

The remaining masking defect was that runMixedControllerCampaign only advanced its outer budget when the public projection turn changed. The harness now counts accepted human and AI commands globally, so simultaneous/hybrid planning can fail quickly in one unchanged turn with mode/turn/revision/actor plus recent public AI transcript diagnostics. This converted the old deadline-only blocker into a finite proof that the current production AI repeatedly accepts found-city in planning closeout before any pass or round-complete.


## Task 9 AI planning closeout — 2026-08-29T05:54:37Z

`visibility.Projection` now exposes only actor-owned persisted planning status via additive `planning_status,omitempty`: omitted for open/sequential/long-turn and hybrid conflict, `queued` when the viewer owns a persisted queued planning command, and `closed` when the viewer owns an early-end marker (which takes precedence over queue). `ai.Driver.Drive` now inspects the persisted projection before tactical planning, submits ordinary `round-complete` through the normal gateway when queued, and returns zero commands at the same revision when already closed. Restart-safe unit and observe-only fullgame coverage passed, and the freeze manifest changed only the locked `internal/gameplay/ai/ai.go` and `internal/gameplay/visibility/visibility.go` hashes.

## Task 9 Row 03 blocker pass — 2026-08-29T02:34:00Z

Bounded public row observation is now available in  via . The mixed-controller harness in  now refreshes each human projection before follow-up planning, honors public  by submitting ordinary  when queued, batches human+AI work in simultaneous/hybrid planning, and emits richer AI no-progress diagnostics. Focused guards remained green after restoring the catalog to its truthful provisional Row 03 shape.

## Task 9 Row 03 blocker pass — 2026-08-29T02:34:00Z

Bounded public row observation is now available in `internal/e2e/fullgame/task9_rows_test.go` via `TASK9_ROW=1..12`. The mixed-controller harness in `internal/e2e/fullgame/executor.go` now refreshes each human projection before follow-up planning, honors public `planning_status` by submitting ordinary `round-complete` when queued, batches human+AI work in simultaneous/hybrid planning, and emits richer AI no-progress diagnostics. Focused guards remained green after restoring the catalog to its truthful provisional Row 03 shape.

## Task 9 leave closeout — 2026-08-29T06:42:00Z

Simultaneous and hybrid `leave` must remain a replayable planning action, but the engine can still close that actor atomically by applying an immediate policy `early-end` right after queueing the same leave command inside `applyPolicyCommand`. This preserves stable replay order, keeps `TakeoverPending` false until resolution replays the queued leave, and lets the last eligible survivor closeout unlock the boundary so `PendingTakeoverWorker` can submit ordinary `takeover-ai` without any departed-client input.

A second boundary nuance surfaced during GREEN: in planning modes, `CommandTakeoverAI` must bypass policy routing once `ResolutionLocked` is already false, otherwise the worker only emits another `command.queued`. The lifecycle lock itself remains the authority that forbids takeover before the boundary.

## Task 9 AI open fallback — 2026-08-29T06:57:11Z

`visibility.planningStatus` now distinguishes actor-owned open planning from omitted-status contexts: active viewers in simultaneous/hybrid planning stage see `planning_status:"open"` only when they have neither a queued command nor an early-end marker, while `queued` and `closed` still take precedence and terminal, sequential, long-turn, hybrid-conflict, and unknown-viewer contexts omit the field. `ai.PlanWithHistory` now keeps the score-100 fallback on `pass` for omitted-status projections and switches that fallback to ordinary `round-complete` only for explicit `planning_status:"open"`, so `Driver.Drive` can submit accepted planning closeout through the normal AI gateway without any E2E shortcut. Focused gameplay tests, freeze guard, build, and gameplay race all passed after updating only the locked `ai.go` and `visibility.go` hashes.

## Task 9 Row 06 long-turn continuity transport seam — 2026-08-29T07:50:58Z

The real long-turn Row 06 blocker was not score evaluation. The full-executor RED showed the opener failing in `executeContinuityAcceptance -> runMixedControllerCampaign -> driveScriptedHuman`, while direct owner-bootstrap controls on the same duel/cold/two-slot fixture still showed three owned settlers. The honest second RED was transport-stage specific: `TestLongTurnThreeCityContinuitySchedulerCachedTransportSurvivesSecondCycle` failed because the first long-turn continuity cycle fell through after accepted preparation moves and consumed opener state before the three-city script could run on the next scheduler cycle.

The minimal unlocked fix is continuity+long_turn only:
- pre-bootstrap and retain the leave actor's public `HTTPExplorationTransport` in scheduler state,
- carry that transport's `BaseURL` across restart,
- reuse that stable transport/revision across scheduler cycles instead of rebuilding from `Bootstrap(ctx, 0)`, and
- stop the cycle immediately after accepted preparation moves, refreshing the same public projection before returning.

With that seam fixed, the exact Row 06 observer passed two clean runs plus the completed-turn-12 restart with identical immutable artifacts: winner `account-1`, state `5292b679fa07ec631929e4f4eb36805d89abe2ba00c2ccb2dbe0430c616477b5`, events `b8e31801117572d3cdb441fc16eb3d4eb42d409a953fe74761637573a5017282`, AI commands `232`, terminal reason `score_2300`. Row 03 remained unchanged at winner `account-1`, state `b6e048ab5587ff79e424eeeaff4006a5d37379172663cff729be697a26e206a3`, events `9d072ef27cd301f8fd494e6a263d1d35049b1949efe36c2e3815f4ea5b3e0cb7`.
- Row 07 can satisfy restart-at-completed-turn-12 without scheduler delay state by prefixing explicit public  commands in a checked-in scripted-human science path, then reusing the existing city/improvement/wonder sequence.
- Row 07 can satisfy restart-at-completed-turn-12 without scheduler delay state by prefixing explicit public round-complete commands in a checked-in scripted-human science path, then reusing the existing city/improvement/wonder sequence.

- 2026-08-29: Row 08 baseline reproduced with TASK9_ROW=8 before edits: clean-1 failed at public observer with production AI drive for account-4 returning invalid round-complete because army composition was invalid. After fixing the shared scripted-human army id, the remaining clean-1 blocker persisted as account-4 round-complete invalid: unit cannot found a city, and later narrowed to AI no-progress under one speculative row-specific script-only experiment.
- 2026-08-29: Proved and fixed a checked-in script resolver defect: scripts/military-zoc-combat.json reused a global literal army_id across rebound scripted humans. Added a focused RED that account-2 incorrectly resolved fullgame-army-human-1, then updated the resolver/script so account-1 stays fullgame-army-human-1 and account-2 resolves fullgame-army-human-2.
- 2026-08-29: Proved and fixed two planner defects in internal/gameplay/ai: no-city planning leaked create-army for settler units and leaked found-city for non-settlers. Both now have focused RED tests and are filtered from candidate generation.

- 2026-08-29: Removed the final Row 08 debug residue from `internal/e2e/fullgame/executor.go` by deleting `humanDriveDebug` allocation/append and restoring the direct `return aiCommands, err` path in both `human_loop` branches. Verified with `go test ./internal/e2e/fullgame -run '^TestCatalogRowOneMatchesAcceptedFixture$'`, `go test ./internal/e2e/fullgame`, and a repo-wide grep showing no `humanDriveDebug|human_loop` matches.
