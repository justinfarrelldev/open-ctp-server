# Issues

## Task 17 — 2026-07-19

Blocked after the required RED stage. `internal/e2e/fullgame/catalog.go` contains 60 rows but only row 1 is enabled and rows 2–60 have no rules fixture hash, natural terminal metadata, state hash, or event hash. The new RED test `TestCatalogRejectsProvisionalScenarios` demonstrates that `Validate(Catalog)` incorrectly accepts this incomplete manifest.

A compliant GREEN implementation cannot create those 59 artifacts yet: the existing public-conquest evidence shows legal two-player spawn separation (Manhattan distance >=3), visibility radius 2, starter movement 2, no movement refresh on round completion, and public planning that rejects hidden targets. Thus a natural combat terminal cannot be reached through the public HTTP/SSE gateway without forbidden hidden-state access, direct mutation, or randomized exploration. The production lifecycle/movement prerequisite must be corrected before immutable real-flow goldens can be observed.

RED command/result: `GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 18m go test -timeout=16m -p=2 ./internal/e2e/fullgame/... -run '^TestCatalogRejectsProvisionalScenarios$' -count=1` exited 1 because `Validate` accepted provisional scenarios. The pre-existing focused package command exited 0 but does not execute the 60 games.

Further contract inspection: all four natural evaluators (`bloodlust_conquest`, `diplomatic_world_union`, `score_2300`, `science_gaia_controller`) run only during `round-complete`. The production AI's deterministic planner always chooses the highest-scored `fortify` candidate before moves/noop and contains no policy that advances any victory condition. Thus switching victory declarations cannot yield compliant public, real-AI terminal artifacts; the missing lifecycle/movement and AI progression prerequisites must be implemented outside Task 17's permitted file scope.

## Correction — 2026-07-19

The prior assertion that movement does not refresh was contradicted by source and a clean local public run. `engine.advanceRound` resets non-transport unit movement to `runtime.Starter.Movement` when positive, and `CommandRoundComplete` invokes it. After `timeout --signal=TERM --kill-after=30s 10m supabase db reset --local`, the bounded local `TestPublicConquestRunnerOverLocalSupabase` passed in 3.26s through public signup/lobby/setup, HTTP command gateway, SSE resume, runtime restart, and natural `bloodlust_conquest`, producing the pinned case-1 hashes. Its immediate rerun failed only because the existing test reuses its fixed namespace without cleanup; that test-isolation defect, not movement reachability, explains the observed conflict.


## E2E Supabase override — 2026-07-19

RED evidence: `TestMemoryPublicRuntimeNeverUsesExternalConnector` was first added before `NewMemoryPublicRuntime`, and the bounded DSN-free command failed to build with `undefined: NewMemoryPublicRuntime`. The replacement is test-scoped and does not change production SQL constructors. The wider `./internal/e2e/onboarding ./internal/e2e/fullgame` command still fails only at the pre-existing Task 17 RED test `TestCatalogRejectsProvisionalScenarios`, because the incomplete 60-row catalog validator intentionally still accepts disabled rows; no database or Supabase call occurred.

## Rejected E2E memory-router correction — 2026-07-19

The hand-written `MemoryPublicRuntime` and wrapper tests were removed because they duplicated production routes and did not exercise the real handler graph. The production composition has no narrow storage seam: account/lobby/game/world handlers, `auth.SessionStore`, gameplay command storage, and SSE membership all take concrete `*sqlx.DB` and own direct transaction/query behavior. A compliant SQL-free public-flow runtime requires explicit storage interfaces and adapters across those production packages; it cannot be implemented by an E2E-only fake without bypassing handlers.

## Composition seam implementation — 2026-07-19

`internal/runtime.New` now owns the production handler graph and `main.go` delegates to it after opening PostgreSQL. `internal/e2e/onboarding.NewPublicRuntime` delegates to that same factory. The first memory-driver contract supports the real account-signup transaction and fails closed on all other SQL. Bounded GREEN result with all Supabase variables unset: `GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m go test -timeout=8m -p=2 ./internal/e2e/onboarding -run '^TestPublicRuntimeUsesProductionAccountHandlerWithNamedMemoryStore$' -count=1` exited 0. The broader lobby/game/world/gameplay/SSE/restart SQL contract and transactional clone/rollback semantics remain unimplemented and must not be treated as complete.

## Production-handler stub boundary follow-up — 2026-07-19

`./internal/e2e/...` with every Supabase variable unset reaches no external configuration path. Onboarding previously passed against `sqlmock`; it now uses `internal/e2e/memorysql` for the real signup handler. The current bounded command `GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m go test -timeout=8m -p=2 ./internal/e2e/... -count=1` still fails only at the inherited catalog RED test (`TestCatalogRejectsProvisionalScenarios`), not storage setup. A full public signup/lobby/game/SSE/restart execution against real handlers still needs a sizeable deterministic SQL driver or production repository refactor because the handler graph owns direct concrete SQL transactions across account/lobby/game/world/gameplay/event packages.

## MemorySQL next RED — 2026-07-19

The public start/world route is RED at the exact fail-closed statement SELECT game_id FROM world_snapshots WHERE game_id = $1. This is the next required world snapshot persistence contract; it has not been stubbed or redirected.

## Task 17 Row 2 AI autonomy follow-up — 2026-08-10

RED/GREEN evidence: the campaign source guard rejects executor `Driver.Pass`, gateway submissions, and constructed AI submissions; AI now replans through the production gateway within its action budget. `TestAIGatewayProjectionExcludesHiddenCanonicalCanaryBeforePlanning` passes with an actual `AIGateway` reading canonical state and proves `HIDDEN_CANARY_TILE`/`HIDDEN_CANARY_UNIT` are absent from the raw bytes passed to `ai.Plan`. Terminal SSE resync was repaired so a terminal snapshot contains the filtered projection and matching revision plus terminal metadata.

Remaining concrete blocker: `env -u GAMEPLAY_POSTGRES_DSN -u SUPABASE_DB_URL -u SUPABASE_PROJECT_URL -u SUPABASE_API_KEY -u SUPABASE_SERVICE_ROLE GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 8m go test -timeout=7m -p=2 ./internal/e2e/fullgame -run '^TestPublicExecutorCompletesRowTwoThroughProductionRuntime$' -count=1` still exits 1 with `diplomatic scenario did not reach a natural terminal`. `executeDiplomaticScenario` needs a repeated human-visible exploration/movement/embassy loop and terminal audit return; it currently runs one city script/AI cycle then returns that unconditional error. No synthetic terminal or scripted AI embassy was introduced.

## Task 17 Row 2 scheduler attempt — 2026-08-10

Implemented a persistent human-only projected scheduler with campaign-wide action accounting, per-unit visited tiles, fresh SSE projection after every move/embassy, and a no-immediate-bounce fallback. RED: `TestHumanDiplomacySchedulerAdvancesEachEligibleUnitWithoutBounce` initially failed due to missing scheduler and now passes. The exact DSN-free duplicate-run test still fails after 48 turns with no natural terminal; its safe final human projection contains only `fullgame-city-human-1` and scout `settler-1-2` at `tile-0-1`, so no AI city is visible and no legal embassy may be submitted. The production AI was kept on `Driver.Drive`; clipping its planner moves to one projected step eliminated its initial `movement obstructed` failure, and repeated `Driver.Drive` calls let each AI emit its own pass, but the terminal remains unreachable in the fixed budget.

A second deterministic scheduler ordering used only the current human projection to favor unvisited legal tiles farthest from human-owned visible cities (stable coordinate ties). It keeps `TestHumanDiplomacySchedulerAdvancesEachEligibleUnitWithoutBounce` green, but the exact duplicate-run still exhausted turn 48. Its final safe projection placed `settler-1-2` at `tile-10-2`, adjacent to visible AI `settler-2-2` at `tile-10-0`, while still exposing only the human city. This proves the remaining blocker is not a local bounce or invalid human route: no visible foreign city exists from which the public embassy command can be derived.


## Task 17 Row 2 terminal blocker after projection repair — 2026-08-10

Exact DSN-free command: `env -u GAMEPLAY_POSTGRES_DSN -u SUPABASE_DB_URL -u SUPABASE_PROJECT_URL -u SUPABASE_API_KEY -u SUPABASE_SERVICE_ROLE GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 8m go test -timeout=7m -p=2 ./internal/e2e/fullgame -run '^TestPublicExecutorCompletesRowTwoThroughProductionRuntime$' -count=1` exits 1 after 48 turns. The final safe viewer projection at revision 425 shows both foreign cities and `embassy_player_ids:["account-2","account-3"]` for the human.

The concrete remaining cause is `engine.unionWinner`: it requires *every* active nation to contain every other active nation in its `Embassies`, so the terminal additionally needs autonomous AI embassies with the human. The runner must not script those actions, read canonical state, weaken victory rules, alter vision, or extend `MaxTurns`. The AI transcript remains autonomous production-gateway output; it founded both AI cities and submitted ordinary moves/passes, but its viewer did not reach the human city within the fixed budget. The runner’s exhausted error contains both the transcript and safe projection as the first complete public diagnostic.


## Row 2 reciprocal AI-to-AI discovery limit — 2026-08-10

The production evaluator requires every nation to embassy every other nation. With the visible-only frontier planner, both AI players embassy the human and account-3 embassies account-2, but account-2 does not rediscover account-3 by turn 48. The planner has no persisted history of its successful move destinations, so any pure current-snapshot ordering can cycle at a map edge/restart. The reviewed compliant resolution is AI-owned receipt-backed exploration history, rebuilt after restart from AI command receipts; runner state, hidden state, scenario values, and victory semantics remain prohibited.


## Current verification checkpoint — 2026-08-10

`GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m go test -timeout=8m -p=2 ./internal/gameplay/ai -count=1` exited 0. `GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m go build ./...` exited 0. The exact DSN-free Row 2 duplicate command remains red because no durable actor-scoped accepted-AI-move history is yet persisted/reloaded after runtime restart.

## Task 17 Row 2 persistence follow-up — 2026-08-10

The exact DSN-free duplicate Row 2 command remains RED after the minimal scoped persistence implementation: `env -u GAMEPLAY_POSTGRES_DSN -u SUPABASE_DB_URL -u SUPABASE_PROJECT_URL -u SUPABASE_API_KEY -u SUPABASE_SERVICE_ROLE GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 8m go test -timeout=7m -p=2 ./internal/e2e/fullgame -run '^TestPublicExecutorCompletesRowTwoThroughProductionRuntime$' -count=1`. It exhausts the fixed 48-turn budget without natural `diplomatic_world_union`; the terminal transcript reaches AI2→human and AI3→AI2 but not AI2→AI3. Three projection-only planner variants were evaluated and reverted because they did not satisfy the fixed scenario: least-recent revisit fallback, home-latitude tie-breaking, and frontier-neighbor-count ranking. Failure output: `/home/justi/.local/share/opencode/tool-output/tool_fec92f505001fjiMTkUz1FMU2I`.

## Row 1 victory rules selection blocker — 2026-08-10

RED public Row 1 initially ended through the generic `score_2300` path/post-terminal 409. After selecting only `bloodlust_conquest` in the durable production rules pin and preserving the original generator seed, the exact DSN-free run reaches combat but fails `capture city: no visible foreign city capture route` within the unchanged 48-turn budget. The remaining obstacle is public projected routing/visibility after combat, not rules selection; changing map, turns, movement, combat, or adding a runner shortcut is prohibited by this task. The exact DSN-free Row 1 command with `-count=2` reproduced this same routing error in both runs; it did not terminate as `score_2300`.

## Row 1 city-directed projected frontier checkpoint — 2026-08-10

The new focused city-directed frontier regression passes, but the exact DSN-free Row 1 duplicate command remains RED: it reaches the fixed action budget with `ai-city-settler-3-1` currently projected at `tile-4-7` and no capture. The terminal safe projection has the city and its tile non-stale, while human units occupy separate visible components; no stale tile was submitted. Therefore duplicate Row 1 artifacts/hashes must not be frozen, and Row 2 duplicate verification has not been rerun from this checkpoint.


## Eight-player controller persistence assertion RED — 2026-08-10

Actual command/result: `env -u GAMEPLAY_POSTGRES_DSN -u SUPABASE_DB_URL -u SUPABASE_PROJECT_URL -u SUPABASE_API_KEY -u SUPABASE_SERVICE_ROLE GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m go test -timeout=8m -p=2 ./internal/e2e/fullgame -run '^TestPublicCampaignSetupEightPlayersOneHumanRestAI$' -count=1` exited 1. The corrected production DTO assertion failed at `started player 0 account 1 country "america" controller = "", want "human"`. `GamePlayer.Controller` is persisted as `controller_kind`, but the current public game-player read path returns an empty controller value; fixing that production read behavior is outside this test-only task.


## Eight-player controller persistence assertion resolved — 2026-08-10

The empty controller readback was resolved by adding `controller_kind` to the production ordered game-player projection and matching named-memory SELECT. The original focused DSN-free command now exits 0 with actual persisted `human`/`ai` assertions before and after restart.


## Eight-player all-human setup verification — 2026-08-10

The DSN-free all-human setup/restart regression passed alongside the existing one-human/rest-AI regression. A source grep found one `AddAI` call only in the pre-existing AI test body; the new all-human test has none.


## Three-player half-human setup verification — 2026-08-10

The DSN-free three-player half-human setup/restart regression passed. A source check found no direct SQL (`GetContext`, `QueryContext`, `ExecContext`) or direct controller mutation in `setup_public_test.go`.


## Deterministic controller assignment helper verification — 2026-08-10

Focused DSN-free helper tests passed. Scope grep found no handler, AI-command, or direct-SQL calls in the new controller-assignment unit test.


## Generalized setup compatibility checkpoint — 2026-08-10

The requested Row 1/Row 2 compatibility command was run after wrapper delegation. It exited 1 at the pre-existing Row 1 terminal-route failure `capture city: no visible legal frontier`, already documented in the Row 1 checkpoint; this task did not change runner, rules, map, action budget, or gameplay. The helper-focused and three-player setup/restart command exited 0.

## Task 1 preservation blockers — 2026-08-10

Clean worktree verification rejected gameplay service/store/HTTP/combat because service required uncommitted testkit, store required prohibited migrations, and HTTP required dirty generated docs. Testkit itself required the dirty Taskfile timeout policy. Runtime/onboarding/memorysql remain bridge-dependent on prohibited `internal/game`; all details are in `.sisyphus/evidence/task-1-blocked-candidate.txt`.

## Task 2 verification boundary — 2026-08-10

The required unfiltered `go test ./internal/game ./internal/e2e/memorysql -count=1` still fails only at pre-existing `TestActiveGameChatUsesLobby` and `TestWorldRouteSurfaceContract` active-game-chat route registration assertions. Task 2 did not change that route surface. The Task 2-focused bridge/start suite, named-memory driver tests, clean-worktree builds, and fresh/legacy isolated migration checks pass.

## Task 2 route-test rejection resolved — 2026-08-10

`TestActiveGameChatUsesLobby` and `TestWorldRouteSurfaceContract` had failed only because they searched uncommitted `main.go` after composition moved to `internal/runtime`. They now validate actual runtime HTTP behavior; the required `go test ./internal/game ./internal/e2e/memorysql -count=1` passes.

## Task 3 commit boundary — 2026-08-10

`git ls-files internal/e2e/fullgame` returned no paths even though the workspace contains a complete untracked E2E tree with unrelated Task 17 work. Committing only the Task 3 E2E files would leave an incomplete package in a clean checkout; staging the entire directory would violate the single-task boundary. The self-contained importer unit was committed as `f72a4e5`; the verified E2E slot changes remain unstaged pending restoration/commit of the inherited E2E baseline by the orchestrator.

## Task 3 commit boundary resolved — 2026-08-10

The setup-only contract and public lifecycle coverage were moved into a new standalone onboarding test package that imports only tracked production runtime/memorysql/rules foundations. The isolated clean applied-patch worktree passed DSN-unset focused setup tests, `go build ./...`, and `git diff --check`; it was committed atomically as `3dca4e1` without staging `internal/e2e/fullgame/**`.

## Task 5 fixed Row 1 contradiction — 2026-08-11T12:19:00Z

The permitted non-empty capture acknowledgement guard is green but cannot affect fixed Row 1: its public combat result is deterministically `defender_held`, attacker casualties 1, defender casualties 0, and empty `CapturedCityID`. Preserving the required empty-result fallback therefore reproduces `capture city: no visible legal frontier`. A removed experiment continuing ordinary rounds after that error still reached no natural terminal within unchanged budgets. No repeat/restart hashes can be accepted.

## Task 5 contradiction resolved — 2026-08-11T12:35:23Z

The preceding conclusion applied only to the incorrectly composed one-unit army. Public bootstrap data showed multiple eligible human units at the selected unit's tile; including that sorted public set repaired the actual fixed-scenario progression invariant without changing combat vectors or using hidden state. Projection-only stable land exploration/backtracking then reached natural terminality. The exact Row 1 three-run test now passes with identical clean/restart artifacts. The unfiltered broad fullgame command still fails only at `TestCatalogHasExactlySixtyFixedCoveredCases` for the separate non-observed Task 9 rules fixture hash.

## Task 6 Row 1 contradiction — 2026-08-11T13:06:00Z

The exact historical Row 1 fixture (`fullgame-minimal-public-executor`, seed `seedBase+3`, unchanged map/rules/budgets) reproducibly fails both checkpoint-0 and checkpoint-2 execution with `capture city: no visible foreign city capture route`. The bounded safe projection shows the two-unit public army but no visible city at turn 50. This is additional to the expected Task 9 catalog-hash red; no seed, budget, engine, hidden-state, or terminal assertion workaround was applied.

## Task 6 Row 1 contradiction resolved — 2026-08-11T13:39:00Z

The regression was not inherited and was not caused by the generic contract. `runner.go` had been recreated after Task 5 with an unaccepted shared cross-phase budget change; it exhausted the 48-turn counter at turn 50 before the accepted capture phase completed. Restoring phase-bounded counters makes the exact Task 5 semantic test and both Task 6 fixtures green. The only broad-suite red is again the separate Task 9 catalog rules hash.

## Task 7 pre-freeze blockers — 2026-08-11T14:26:00Z

The complete freeze gate is not green despite production behavior passing. Aggregate `./internal/...` coverage with only `TestCatalogHasExactlySixtyFixedCoveredCases` excluded is 76.7%, below the committed >80% policy; engine-only coverage is 88.8%, below the unrelated inherited dirty >=90% policy. Repository `gofmt -l` is nonempty for pre-existing tracked files plus six inherited untracked E2E files. The optional DSN `TestExecuteCommandPostgresLockRace` creates stale private DDL without `gameplay_events.event_id` and fails when actually enabled; the real migration contains the column and designated race gates pass, so no production change was made. Full details are in `.sisyphus/evidence/task-7-pre-freeze.txt`.

## Task 7 blockers cleared — 2026-08-11T15:09:30Z

The committed aggregate >80% blocker is cleared at 80.4%; the unrelated dirty engine >=90% rule remains untouched and is not a Task 7 gate. The optional lock-race test is no longer skipped-or-red when enabled, and repository Go formatting is clean. The inherited fullgame package remains intentionally untracked; final evidence binds 44 fullgame/onboarding acceptance inputs by SHA-256 `c4f125c3983cd178e1d5d8c596aed4f5639a6139f3cbb7bbf141daaa87a907e7` and preserves commit-feasibility context.

## Task 8 — 2026-08-11
No implementation blocker remains. The only operational caveat is inherent: a repository-local guard cannot detect a malicious actor who edits both production files and the manifest, so independent review remains necessary.

## Task 8 correction — 2026-08-11
Primary review gaps are resolved. No new blocker was found; the existing unrelated dirty workflow content remains preserved except for the single freeze-test step.

## Task 8 retry 2 — 2026-08-11
Retry-2 fail-closed and bounded-CI gaps are resolved. No remaining Task 8 issue was found.

## Task 9 remaining blocker — 2026-08-29T02:15:27Z

Task 9 now finalizes the catalog schema and split files, but it does not promote Rows 03–12 to immutable accepted artifacts. Doing so would require reviewed public execution evidence for those rows and, for resignation semantics in particular, a source-compatible public contract that does not currently exist in the provided evidence. The focused fullgame package, race, build, vet, diff, and freeze guard all pass, but `Validate(Catalog)` intentionally still fails on non-observed immutable artifacts after Row 1 because Rows 02–60 remain provisional/fail-closed.

## Task 9 retry freeze blocker — 2026-08-29T02:52:00Z

A real public leave/disconnect continuity row is blocked by the frozen production contract. `TestPublicLeaveHasNoAutonomousTakeoverPath` proves that after a successful public `leave`, the gameplay gateway has no public `takeover-ai` route (404) and the production AI driver cannot autonomously recover the slot. The locked symbols are `internal/gameplay/http/http.go` (`commands` map) and `internal/gameplay/ai/ai.go` (`PlanWithHistory` / `Driver`).

## Task 9 retry verification refresh — 2026-08-29T03:00:00Z

After landing the public continuity blocker test and the Row 1 identity fix, the complete `internal/e2e/fullgame` package, its race run, `go build ./...`, `go vet ./internal/e2e/fullgame`, the freeze guard, and `git diff --check` all pass. The blocked status is therefore due to the frozen public leave/takeover contract itself, not a broken harness tree.


## Task 9 pending takeover correction — 2026-08-29T03:36:04Z

RED showed the branch already contained a partial worker file (`internal/gameplay/service/takeover.go`) but no `FindPendingTakeovers` store contract, so the intended takeover path failed at compile time before any autonomous completion could occur. GREEN added only the missing persisted query/update/runtime seams plus exact `memorysql` contracts; the first rerun then exposed one narrow E2E assertion bug because projected SSE event data contains only `type` and `sequence`, not the raw lifecycle payload. After aligning the test to the production projection contract and updating the freeze manifest/count for the new locked `internal/gameplay/service/takeover.go` path, focused, race, build, vet, diff, and freeze verification all passed.


## Task 9 TDD evidence correction — 2026-08-29T03:49:27Z

Independent review correctly rejected the original RED because it was only `FindPendingTakeovers undefined`, proving implementation skeleton code already existed and behavior was never exercised. The corrected RED was reconstructed honestly in `/tmp/opencode/task9-tdd-reconstruction` from the current GREEN workspace with the takeover behavior withheld only in the temp copy; the focused public continuity command then compiled and failed with the expected behavioral predicate that the controller remained `human` after the worker cycle. The real workspace preserved the reviewed implementation and passed the same exact focused command plus the refactor verification suite before the evidence files were overwritten.


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

## Task 9 mixed-controller scheduler blocker — 2026-08-29T01:30:45Z

Current live blocker is still harness-side focused mode closeout. TestPublicExecutorModeSpecificCloseoutCompletesObserveOnlyPlanningModes remains red even after batching AI revision propagation and removing the extra per-player AI retry loop. Current executable proof: env -u GAMEPLAY_POSTGRES_DSN -u SUPABASE_DB_URL -u SUPABASE_PROJECT_URL -u SUPABASE_API_KEY -u SUPABASE_SERVICE_ROLE GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 8m go test -timeout=7m -p=2 ./internal/e2e/fullgame -run ^TestPublicExecutorModeSpecificCloseoutCompletesObserveOnlyPlanningModes$ -count=1 exits 1 with simultaneous and hybrid both timing out after 2m while production AI drive for account-3 context deadline exceeded (or the subsequent batch refresh timeout). Rows 03–12 were not attempted because the focused scheduler gate is not fully green yet.

## Task 9 finite stall proof — 2026-08-29T01:43:30Z

After the harness-only global command budget fix, the focused simultaneous/hybrid observe-only scheduler test no longer times out. It now fails quickly at the declared command budget with public AI transcript evidence showing repeated accepted found-city commands for the same actors and payloads in turn 1/revision 13, without reaching pass or round-complete. This is current production-stall proof, not a stale timeout artifact. Rows 03–12 remain unattempted until that focused planning-mode blocker is resolved.


## Task 9 AI planning closeout follow-up — 2026-08-29T05:54:37Z

The only post-GREEN integration mismatch was in the observe-only fullgame harness, not production gameplay: `validateAIObservation` and two caller loops had assumed every AI drive must submit at least one command. Once persisted planning closeout became legitimate, a fresh `Driver.Drive` could correctly return `0 commands / 0 observed submissions` for an already-closed actor. Updating that harness expectation restored the simultaneous/hybrid gate without introducing `Driver.Pass`, direct gateway submission, or synthetic AI commands.

## Task 9 Row 03 public runtime blocker — 2026-08-29T02:34:00Z

Exact bounded observation command used repeatedly: === RUN   TestObserveTask9Row
2026/08/29 02:35:07 Session created: &{5183600185528097376 1 2026-08-29 02:35:07.677822033 -0400 EDT m=+0.073791160 2026-08-29 14:35:07.677822073 -0400 EDT m=+43200.073791200}
2026/08/29 02:35:07 Session created: &{7150261912773349396 2 2026-08-29 02:35:07.726255946 -0400 EDT m=+0.122225073 2026-08-29 14:35:07.726255986 -0400 EDT m=+43200.122225113}
2026/08/29 02:35:07 Session created: &{1079462517016342741 3 2026-08-29 02:35:07.740019326 -0400 EDT m=+0.135988454 2026-08-29 14:35:07.740019377 -0400 EDT m=+43200.135988494}
- After Row 07 acceptance, negative validation that previously mutated row index 6 had to move to Row 08 because Row 07 is no longer provisional.
- After Row 07 acceptance, the negative validation case that previously mutated row index 6 had to move to Row 08 because Row 07 is no longer provisional.

- 2026-08-29: Row 08 is not yet accepted. After the army-id and AI-candidate fixes, TASK9_ROW=8 still fails on clean-1 through the public observer. With the normal scripted-human flow it fails at production AI round-complete for account-4 with unit cannot found a city; with a speculative second-human script-only branch it instead fails closed with production AI could not make progress at the current policy boundary.
