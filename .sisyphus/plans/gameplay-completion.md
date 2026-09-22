# Complete and Freeze Gameplay, Then Prove Deterministic Full Games

## TL;DR
> **Summary**: Preserve dependency-complete green work atomically, finish and freeze the production gameplay execution path with strict TDD, then complete deterministic public-handler campaigns without changing frozen engine code.
> **Deliverables**:
> - Atomic preservation commits backed by clean-tree verification
> - Public CTP2 player counts 2–8, including 5 and 7, with arbitrary ordered human/AI slot assignments
> - Data-driven support for every playable base-game civilization imported from a user-owned installation
> - Test-only diagnostics and optional human-slot AI puppeteering with no production exposure
> - Green Row 1/Row 2 execution engine, immutable SHA-256 engine lock, fixed 60-case campaign catalog, disruption matrix, CI, and traceability
> **Effort**: XL
> **Parallel**: YES — 4 waves
> **Critical Path**: Task 1 → 2 → 3 → 5 → 6 → 7 → 8 → Tasks 9–13 → 14 → F1–F4

## Context
### Original Request
Complete Gameplay using TDD RED–GREEN–REFACTOR and deterministic beginning-to-end E2E games. Preserve verified progress first; support all real-game player counts (including 5 and 7), every base-game country, and every per-slot human/AI mixture; keep AI autonomous; provide test-only diagnostics; freeze the execution engine before campaign authoring.

### Interview Summary
- CTP2 is the compatibility baseline; CTP1 and space-layer gameplay remain excluded.
- Pinned behavioral source is `civctp2/civctp2@0668dd199c2519ff8ab7d2871e7c7a89e21c8e52`.
- Public CTP2 setup defines eight slots; supported product counts are **2–8 inclusive**. The 32-slot engine constant is not the public setup contract.
- Named compositions and the half-human rule are deleted. Controller assignment is an ordered per-slot `human|ai` vector. This server deliberately allows arbitrary mixtures, including more than CTP2 network setup’s original four-human cap.
- All playable base-game civilizations are supported through the existing user-owned-installation importer. No protected names/records/assets are copied into the repository; barbarian/special records are excluded by metadata/record classification.
- Human E2E actors submit deterministic legal public commands. Production AI selects its own commands from player-visible projections/history. Optional test-only puppets reuse AI decision logic to choose commands for a human actor but submit through that actor’s ordinary public client.
- After freeze, campaign work may change only E2E orchestration, synthetic fixtures, scripts/catalog data, evidence, CI, and documentation. A production-path defect stops campaign work and requires an explicit freeze-reopen task.
- Plan TODO text and order are immutable after generation. Executors may only change `[ ]` to `[x]`; they may not add, remove, reorder, split, merge, or rewrite TODOs.

### Verified Current Status
- Tasks 1–16 from the superseded plan exist and are mostly covered by focused tests, but are not accepted wholesale.
- `go build ./...` is green.
- Setup-only tests and the Row 2 public diplomatic run are green.
- Row 1 is red: `capture city: no visible legal frontier`.
- Catalog validation is red: Row 1 has a non-observed/missing rules fixture hash.
- `internal/game` focused tests and `internal/e2e/fullgame` are not clean preservation commits.
- Migration verification is environmentally blocked by a local port-54322 collision; this plan replaces it with an isolated ephemeral database.

### Metis and Oracle Review (gaps addressed)
- Preserve only changes that pass from a clean worktree containing exactly the staged patch.
- Use a baseline evidence ledger and known-failure signatures so existing reds cannot hide new regressions.
- Freeze selected production files, not all of `internal/game`; move testkit code out of the frozen production tree first.
- Enforce freeze through a complete path+SHA-256 manifest that rejects changed, added, or deleted production files without requiring a historical Git ref.
- Treat test puppeteer strength as harness policy, not as original per-player CTP2 difficulty; original difficulty is global (0–6).
- Define operational oracles for “all civilizations,” “full game,” “deterministic,” “autonomous AI,” and “freeze.”

## Work Objectives
### Core Objective
Deliver a restart-safe deterministic gameplay service whose production execution path supports every player count 2–8 and arbitrary per-slot controllers, then prove complete games through public HTTP/SSE using autonomous production AI while preventing campaign work from changing the frozen engine.

### Deliverables
- Reproducible baseline ledger and atomic preservation history
- Green setup/start/command/event/AI/world/terminal production path
- Full imported playable-civilization selection with protected-data guard
- Test-only detached diagnostics and puppet policy
- SHA-256 freeze manifest and post-freeze allowlist guard
- Exactly 60 fixed complete-game scenarios plus explicit lifecycle/fault suites
- Determinism hashes, restart evidence, CI gates, and README traceability

### Operational Definitions
- **All player counts**: every integer 2, 3, 4, 5, 6, 7, and 8 passes setup, start, persistence/restart, deterministic slot order, and at least one complete-game representative.
- **All controller assignments**: exhaustive setup tests cover all `2^n` ordered `human|ai` vectors for each player count `n` from 2–8, including all-human and all-AI; full campaigns cover all-human, all-AI, one-human/rest-AI, alternating, 2H/6AI, and 3H/2AI.
- **All base-game civilizations**: every playable civilization record parsed from a user-owned base-game installation is selectable and restart-persistent; barbarian/special, malformed, duplicate, and unknown records are rejected. CI proves the same contract with synthetic records, never retail data.
- **Autonomous AI**: production AI receives only its serialized public projection/history and chooses its own ordinary command through `service.AIGateway`/`ExecuteAICommand`; scripts never prescribe an AI objective or command.
- **Full game**: setup and start succeed, play reaches a terminal reason emitted by the production engine (victory, elimination, resignation/leave, or production rules-defined draw), terminal commands are rejected, and terminal SSE/state/event records agree. A harness command/turn/time cap is always a test failure, never a successful completion.
- **Deterministic**: two clean runs and one restart-at-checkpoint run have identical ordered command schedule, generated IDs, canonical event log, terminal result, state hash, and event hash after explicitly excluded audit timestamps.
- **Frozen engine**: every declared production file exists exactly once with the SHA-256 in `internal/e2e/freeze/engine-lock.json`; the lock test also rejects undeclared additions beneath locked roots.

### Must Have
- Strict test-first RED → minimal GREEN → behavior-preserving REFACTOR evidence for every production correction
- Fixed seeds, fake clocks, deterministic IDs/RNG, sorted iteration, stable tie-breakers, canonical serialization
- Public HTTP/SSE command flow for campaign actions; direct DB access is audit/diagnostic only
- No production package imports `internal/e2e/**`
- Inner Go timeout plus outer GNU `timeout --signal=TERM --kill-after=30s`, `GOMAXPROCS=4`, and `-p=2` for every test command
- Atomic commits with implementation and direct tests together
- Original-source citation before implementing CTP2-specific behavior

### Must NOT Have
- Supabase/network database calls from E2E; use named `memorysql` or an isolated ephemeral migration database
- Production debug/puppet endpoint, SSE topic, command, environment switch, or test branch
- Hidden-state input to production AI, human scripts, or puppet policy
- Hard-coded eight-country roster, named composition labels, half-human rule, or count set `{2,3,4,6,8}`
- Protected retail data, translated strings, civilization records, or generated snapshots committed
- Skipped/disabled/provisional catalog rows; weakened assertions; automatic golden updates; seed changes used to conceal defects
- Campaign modifications beneath the frozen production boundary

## Source and TDD Protocols
### Original Source Review Protocol
For every task marked **Source review required**, the executor must inspect the pinned source before code changes and save `.sisyphus/evidence/task-{N}-source.md` containing commit, exact paths/symbols, observed behavior, and implementation consequence. Source evidence informs behavior; protected source prose/data is not copied.

### RED–GREEN–REFACTOR Protocol
1. **RED**: add/change the narrowest test; run it and capture the intended failure predicate.
2. **GREEN**: make the minimum implementation change; rerun the RED test plus focused regressions.
3. **REFACTOR**: clean structure without changing behavior; rerun focused tests and deterministic hashes.
Evidence: `.sisyphus/evidence/task-{N}-{red|green|refactor}.txt`. A pre-existing unrelated failure is not valid RED evidence.

### Immutable TODO Protocol
After this plan is accepted, no executor or subagent may alter TODO text/order or append TODOs. Only checkbox state may change. Discovery of missing work is reported as a blocker; it does not mutate this plan.

## Verification Strategy
> ZERO HUMAN INTERVENTION DURING VERIFICATION — agents execute every test, QA action, evidence capture, and F1–F4 review. After F1–F4 finish and results are presented, explicit user approval is required only as an administrative completion gate; it is not a verification action or prerequisite.
- Tests: Go `testing`, `httptest`, named `memorysql`, HTTP/SSE collectors, and an isolated ephemeral PostgreSQL container only for migrations.
- Focused gameplay command: `GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m go test -timeout=8m -p=2 ./internal/game ./internal/gameplay/... ./internal/world ./internal/runtime ./internal/e2e/memorysql ./internal/e2e/onboarding ./internal/e2e/fullgame -count=1`.
- Full command: `GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 18m go test -timeout=16m -p=2 ./... -count=1`.
- Gameplay race command: `GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 15m go test -race -timeout=13m -p=2 ./internal/gameplay/... ./internal/runtime ./internal/e2e/fullgame -count=1`.
- Every command records command, UTC timestamp, environment prerequisites, exit code, and output under `.sisyphus/evidence/`.
- No Playwright is required: this scope has no browser UI.

## Execution Strategy
### Parallel Execution Waves
Wave 1: Tasks 1–4 — preserve green work, restore bridge/migrations, generalize roster/civilizations, isolate test infrastructure
Wave 2: Tasks 5–8 — fix Row 1, generalize executor, accept/freeze engine, freeze enforcement
Wave 3A: Task 9 — establish split catalog schema/files and complete Rows 01–12 outside frozen paths
Wave 3B: Tasks 10–13 — complete disjoint row-family files and disruption suites in parallel
Wave 4: Task 14 — aggregate catalog/CI/docs and repository-wide verification

### Dependency Matrix
| Task | Depends On | Blocks |
|---|---|---|
| 1 | — | 2–14 |
| 2 | 1 | 3–14 |
| 3 | 2 | 5–14 |
| 4 | 1 | 5–14 |
| 5 | 2,3,4 | 6–14 |
| 6 | 3,4,5 | 7–14 |
| 7 | 2–6 | 8–14 |
| 8 | 7 | 9–14 |
| 9 | 8 | 10–14 |
| 10 | 9 | 14 |
| 11 | 9 | 14 |
| 12 | 9 | 14 |
| 13 | 9 | 14 |
| 14 | 9–13 | F1–F4 |

### Agent Dispatch Summary
- Wave 1: 4 tasks — `deep` × 4
- Wave 2: 4 tasks — `deep`, `deep`, `unspecified-high`, `quick`
- Wave 3A: 1 task — `deep`
- Wave 3B: 4 tasks — `deep` × 4
- Wave 4: 1 task — `unspecified-high`

## TODOs

- [x] 1. Reproduce the baseline and atomically preserve dependency-complete green work

  **What to do**:
  - Create `.sisyphus/evidence/task-1-baseline.md` containing current branch/HEAD, `git status --short`, exact green and known-red commands, exit codes, and the two expected red signatures. Do not change source to obtain this ledger.
  - Build a mutually exclusive file matrix: `commit-now`, `blocked-red`, `unrelated-user-work`. No file may appear in two buckets.
  - Stage and commit only these candidate units, one unit per atomic commit, and only when a temporary clean worktree made from `HEAD` plus that unit’s staged patch builds and passes its focused command:
    1. account/lobby/world support with direct tests;
    2. generated docs/CI/Taskfile changes only if regeneration is reproducible and they are gameplay-related;
    3. rules importer with synthetic fixtures and direct tests;
    4. gameplay packages plus `memorysql`/onboarding/runtime only if the clean-worktree patch is dependency-complete without blocked `internal/game` or `fullgame` changes.
  - If a candidate clean-worktree verification fails, leave it uncommitted in `blocked-red`; do not broaden the commit. Preserve all unrelated work untouched.

  **Must NOT do**: commit `internal/game/**`, `internal/e2e/fullgame/**`, gameplay migrations, plan/evidence files, or any candidate that depends on unstaged files; stash/reset/discard user work; make one catch-all commit.

  **Recommended Agent Profile**:
  - Category: `deep` — requires dependency and dirty-tree analysis.
  - Skills: [`git-master`] — mandatory atomic staging, style detection, and clean-patch verification.
  - Omitted: [`playwright`] — no browser surface.

  **Parallelization**: Can Parallel: NO | Wave 1 | Blocks: 2–14 | Blocked By: none

  **References**:
  - Current plan status: `.sisyphus/notepads/gameplay-completion/{learnings,issues,decisions,problems}.md`
  - Candidate world files: `internal/account/create_account.go`, `internal/lobby/create_lobby.go`, `internal/world/{generator,snapshot,store}.go`
  - Importer: `cmd/import-rules/`, `internal/gameplay/rules/`
  - Engine candidates: `internal/gameplay/`, `internal/e2e/{memorysql,onboarding}/`, `internal/runtime/runtime.go`
  - Blocked areas: `internal/game/`, `internal/e2e/fullgame/`, `supabase/migrations/20260719*.sql`, `supabase/migrations/20260810160000_gameplay_ai_exploration_history_task17.sql`

  **Acceptance Criteria**:
  - [ ] Baseline ledger records `go build ./...` exit 0 and the exact expected Row 1/catalog failure predicates.
  - [ ] Every created commit has a disjoint file list, implementation+tests together, clean-worktree focused test exit 0, clean-worktree build exit 0, and `git diff --check` exit 0.
  - [ ] `git status --short` after commits retains every blocked/unrelated change and has no accidentally staged files.

  **QA Scenarios**:
  ```
  Scenario: Preserve a dependency-complete candidate
    Tool: Bash with git-master
    Steps: Export candidate staged patch; apply to a temporary clean worktree at pre-commit HEAD; run its focused tests and go build; compare staged paths to allowlist; commit only after exit 0.
    Expected: Commit contains exactly the allowlisted implementation/tests and can be built from its parent without any dirty-tree dependency.
    Evidence: .sisyphus/evidence/task-1-atomic-commits.txt

  Scenario: Candidate relies on blocked dirty files
    Tool: Bash with git-master
    Steps: Apply only that candidate patch to a clean worktree and run focused tests/build.
    Expected: Nonzero verification prevents commit; files remain unchanged and are recorded only in blocked-red.
    Evidence: .sisyphus/evidence/task-1-blocked-candidate.txt
  ```

  **Commit**: YES | Messages: repository-style, one per accepted unit | Files: only clean-worktree-verified candidate matrices

- [x] 2. Restore the production bridge and migration baseline under TDD

  **What to do**:
  - Capture current `internal/game` failures as baseline, then add narrow RED tests for query/scan/controller/rules/world-start contract mismatches.
  - Make the setup/start/read bridge internally consistent: persisted ordered players expose country and `controller_kind`; start loads exactly the selected rules/world/controller state; wrong/missing rows fail closed.
  - Run gameplay migrations against an isolated ephemeral PostgreSQL container on a dynamically allocated host port; assert fresh install and legacy upgrade paths, then tear it down.
  - Commit bridge implementation+tests and migration+tests as separate atomic commits after full focused gates pass.

  **Must NOT do**: point tests at Supabase or a developer database; hard-code port 54322; weaken SQL expectations; mix fullgame runner changes into bridge commits.

  **Recommended Agent Profile**:
  - Category: `deep` — cross-package SQL/startup repair.
  - Skills: [`git-master`] — atomic bridge and migration commits.
  - Omitted: [`playwright`] — HTTP/SQL tests suffice.

  **Parallelization**: Can Parallel: NO | Wave 1 | Blocks: 3,5–14 | Blocked By: 1

  **References**:
  - `internal/game/{create_game,endpoints,game_handler,helpers,models,player_endpoints,start_gameplay,world_contract_stubs}.go`
  - `internal/game/*_test.go`
  - `internal/e2e/memorysql/memorysql.go`
  - `supabase/migrations/20260719120000_gameplay_persistence_task2.sql`
  - `supabase/migrations/20260719130000_gameplay_event_ids_task17.sql`
  - `supabase/migrations/20260719140000_game_setup_dimensions_task17.sql`
  - `supabase/migrations/20260810160000_gameplay_ai_exploration_history_task17.sql`
  - `supabase/tests/task2/`

  **Acceptance Criteria**:
  - [ ] RED/GREEN/REFACTOR evidence exists for each repaired bridge contract.
  - [ ] Bounded `go test ./internal/game ./internal/e2e/memorysql -count=1` exits 0.
  - [ ] Fresh and legacy migration tests exit 0 in an ephemeral container and leave no running container/network/volume.
  - [ ] `go build ./...` exits 0 from the committed tree.

  **QA Scenarios**:
  ```
  Scenario: Persist and restart ordered controllers
    Tool: Go testing + memorysql
    Steps: Create mixed-controller players, start, rebuild runtime, read game/player state.
    Expected: Slot order, country, controller_kind, rules pin, and world seed are identical after restart.
    Evidence: .sisyphus/evidence/task-2-bridge-green.txt

  Scenario: Malformed or missing persisted controller
    Tool: Go testing + ephemeral PostgreSQL
    Steps: Exercise invalid controller_kind and legacy-upgrade fixtures.
    Expected: Schema/start rejects invalid state with stable error; legacy valid rows upgrade successfully; no external DB contacted.
    Evidence: .sisyphus/evidence/task-2-migration-failure.txt
  ```

  **Commit**: YES | Messages: `fix(game): restore gameplay setup and start contracts`; `fix(db): validate gameplay persistence migrations` | Files: bridge implementation/tests; migrations/tests

- [x] 3. Replace count, country, and composition restrictions with explicit slot setup

  **Source review required**: inspect pinned `gstypes.h`, `ns_gamesetup.h`, `gameinit.cpp:2432-2454`, and civilization DB parser/schema; record consequences without copying protected data.

  **What to do**:
  - Introduce one ordered E2E/setup contract: `{slot, countryRecordID, controllerKind}` with controller `human|ai`; delete composition labels and half-human derivation.
  - Accept every player count 2–8 and reject 1, 9, duplicate/missing slots, duplicate countries, and unsupported controllers.
  - Replace hard-coded `campaignCountries` with records supplied by the imported ruleset’s selectable-playable civilization metadata. Exclude barbarian/special records; preserve production’s current generic country string only where compatibility requires it, but make campaign setup resolve imported IDs.
  - Add exhaustive/property setup tests for every controller bit vector at counts 2–8 and every imported playable synthetic civilization. Add public start/restart tests for each count, explicitly 5 and 7.
  - Move/rename existing roster/setup files when splitting responsibilities; do not rewrite working infrastructure from scratch.

  **Must NOT do**: retain `Composition`, `all-human-scripted`, `half-human`, `{2,3,4,6,8}`, eight hard-coded countries, original four-human cap, or committed retail civilization content.

  **Recommended Agent Profile**:
  - Category: `deep` — contract migration across setup/catalog/tests.
  - Skills: [`git-master`] — file moves and atomic commits.
  - Omitted: [`playwright`] — direct HTTP tests.

  **Parallelization**: Can Parallel: NO | Wave 1 | Blocks: 5–14 | Blocked By: 2

  **References**:
  - `internal/e2e/fullgame/{catalog,validate,executor,setup_roster_test,setup_public_test}.go`
  - `internal/e2e/onboarding/{onboarding.go,onboarding_test.go}`
  - `internal/game/{player_endpoints,start_gameplay}.go`
  - `internal/gameplay/rules/{rules,runtime}.go`
  - External: `civctp2/civctp2@0668dd.../ctp2_code/ui/netshell/ns_gamesetup.h:39-46`
  - External: `civctp2/civctp2@0668dd.../ctp2_code/gs/utility/gameinit.cpp:2432-2454`

  **Acceptance Criteria**:
  - [ ] No production/E2E symbol or fixture contains named composition/half-human logic or fixed count set.
  - [ ] Exhaustive setup tests cover all controller vectors for counts 2–8 and complete within bounded timeout.
  - [ ] Public setup/start/restart tests pass for every count 2–8 with deterministic slot and turn order.
  - [ ] Synthetic import tests retain every playable record and reject special/barbarian/malformed/duplicate/unknown records; absent source produces actionable error and no repository mutation.

  **QA Scenarios**:
  ```
  Scenario: Seven-player arbitrary assignment
    Tool: Go httptest + memorysql
    Steps: Configure slots [human,ai,human,ai,ai,human,ai] with seven distinct imported synthetic playable records; start and restart.
    Expected: Seven ordered slots and controllers persist; turn order/world placement are identical across restart.
    Evidence: .sisyphus/evidence/task-3-seven-player.txt

  Scenario: Invalid setup and protected-data boundary
    Tool: Go testing + repository scan
    Steps: Test counts 1/9, duplicate slot/country, special/barbarian record, unknown record, and missing installation; scan tracked files for source fixture signatures.
    Expected: Every invalid setup fails closed; importer creates no tracked retail output; scan exits 0.
    Evidence: .sisyphus/evidence/task-3-invalid-roster.txt
  ```

  **Commit**: YES | Message: `refactor(e2e): model explicit gameplay player slots` | Files: onboarding/fullgame setup contracts, synthetic importer metadata, direct tests

- [x] 4. Isolate test-only diagnostics and human-slot puppeteering

  **What to do**:
  - Move `internal/gameplay/testkit` to `internal/e2e/testkit` with history-preserving filesystem moves and update test imports.
  - Add `internal/e2e/diagnostic` as a detached, read-only DB inspector for game/player/country/city/unit/controller/phase/revision/terminal/event-cursor data. Returned snapshots must not alias mutable production state.
  - Add `internal/e2e/puppet`: given a human actor’s public projection/history and a harness policy level, reuse AI planning logic to choose a legal command, then submit through that human actor’s normal authenticated HTTP client. Label policy levels as test harness policy, not CTP2 per-player difficulty.
  - Add import/route/topic scans proving no production package imports or registers diagnostics/puppet code. Inject a hidden-state canary and prove it never appears in AI/puppet input.

  **Must NOT do**: use a nonstandard `//go:build test` tag; add production routes/topics/commands/config switches; permit diagnostic data to influence action selection; let puppet bypass public authorization/command validation.

  **Recommended Agent Profile**:
  - Category: `deep` — security boundary plus reusable testing architecture.
  - Skills: [`git-master`] — preserve moves and pair tests.
  - Omitted: [`playwright`] — no UI.

  **Parallelization**: Can Parallel: YES | Wave 1 | Blocks: 5–14 | Blocked By: 1

  **References**:
  - `internal/gameplay/testkit/`
  - `internal/e2e/fullgame/audit.go` — existing direct DB audit pattern
  - `internal/gameplay/ai/ai.go`
  - `internal/gameplay/service/{ai,service}.go`
  - `internal/e2e/onboarding/http_client.go`
  - `internal/runtime/runtime.go` — must remain free of test registrations

  **Acceptance Criteria**:
  - [ ] `internal/gameplay/testkit` no longer exists and all moved tests remain green.
  - [ ] Diagnostic snapshots expose required canonical fields only inside tests and cannot mutate persisted state.
  - [ ] Puppet-selected commands traverse the human HTTP path; production AI still traverses `ExecuteAICommand`.
  - [ ] Static import and runtime route/topic enumeration prove no production exposure; hidden-state canary is absent from serialized AI and puppet inputs.

  **QA Scenarios**:
  ```
  Scenario: Diagnose and puppet through approved boundaries
    Tool: Go httptest + memorysql
    Steps: Read detached diagnostic snapshot; derive one puppet command from the same actor’s public projection; submit via authenticated human HTTP client.
    Expected: Diagnostic sees canonical state; command succeeds only if public/legal; production state changes solely through normal command gateway.
    Evidence: .sisyphus/evidence/task-4-test-boundaries.txt

  Scenario: Attempt production or hidden-state access
    Tool: Go tests + static scan
    Steps: Enumerate runtime routes/SSE topics/import graph; inject secret canary into unseen enemy state; serialize AI/puppet input.
    Expected: No diagnostic/puppet production surface/import; canary absent; unauthorized actor command rejected.
    Evidence: .sisyphus/evidence/task-4-boundary-failure.txt
  ```

  **Commit**: YES | Message: `test(e2e): isolate gameplay diagnostics and puppet policies` | Files: moved E2E testkit, diagnostic/puppet packages, tests

- [x] 5. Repair the Row 1 public conquest execution path with strict TDD

  **Source review required**: inspect pinned CTP2 movement, visibility, turn-reset, and city-capture paths relevant to the observed failure; cite symbols and distinguish source behavior from this server’s command model.

  **What to do**:
  - Preserve the current Row 1 failure as RED: fixed seed, `capture city: no visible legal frontier`, public projections only.
  - Identify and fix the production invariant causing starters to have no legal visible progression (spawn visibility, movement reset, turn transition, or capture legality). Correct the narrowest production layer; do not route around it in the runner.
  - Prove ordinary production AI commands continue to use the AI gateway and that human conquest commands derive only from public projection/history.
  - Restore observed rules/state/event hashes only after behavior is green twice from clean DBs and once across the defined restart checkpoint. Never auto-write goldens.

  **Must NOT do**: change seed to evade failure; reveal hidden map/city state to planner; script AI to create units/embassies/targets; increase budgets/timeouts as the fix; weaken terminal or hash assertions.

  **Recommended Agent Profile**:
  - Category: `deep` — production execution debugging with deterministic proof.
  - Skills: [`git-master`] — atomic production fix and tests.
  - Omitted: [`playwright`] — HTTP/SSE harness suffices.

  **Parallelization**: Can Parallel: NO | Wave 2 | Blocks: 6–14 | Blocked By: 2,3,4

  **References**:
  - `internal/e2e/fullgame/{runner,exploration,exploration_http,movement,conquest,fullgame_test}.go`
  - `internal/gameplay/{engine,service,visibility,ai}/`
  - `internal/world/{generator,recalculation}.go`
  - `.sisyphus/evidence/task-17-public-conquest-blocker.txt`
  - External: pinned CTP2 movement/turn/visibility/capture symbols recorded in Task 5 source evidence

  **Acceptance Criteria**:
  - [ ] Row 1 RED reproduces the exact legal-frontier failure before production changes.
  - [ ] Minimal production fix makes Row 1 reach a natural production terminal result through HTTP/SSE.
  - [ ] Two clean runs plus restart checkpoint produce identical schedule, IDs, state hash, event hash, and terminal result.
  - [ ] Row 2, focused engine/service/visibility/AI tests, race tests, and build remain green after refactor.

  **QA Scenarios**:
  ```
  Scenario: Deterministic public conquest
    Tool: Go test + HTTP/SSE collector + diagnostic audit
    Steps: Run Row 1 twice clean and once restarting after first durable round boundary; collect public transcript and audit hashes.
    Expected: Natural conquest terminal event; all three canonical schedules/results/hashes identical; no hidden-state reads in planner trace.
    Evidence: .sisyphus/evidence/task-5-row1-green.txt

  Scenario: Illegal hidden shortcut or exhausted movement
    Tool: Go testing
    Steps: Construct unseen enemy city and exhausted starter cases; request next public action and wrong-player/duplicate/terminal commands.
    Expected: Planner cannot target unseen city; exhausted move fails stably until production turn reset; wrong/duplicate/terminal commands follow contract.
    Evidence: .sisyphus/evidence/task-5-row1-edge.txt
  ```

  **Commit**: YES | Message: `fix(gameplay): restore public conquest progression` | Files: narrow production invariant, direct tests, Row 1 regression

- [x] 6. Generalize the public executor and make Row 1/Row 2 engine acceptance independent of catalog authoring

  **What to do**:
  - Refactor `PublicExecutor.Execute` from two scenario-ID special cases into an explicit engine-acceptance runner contract driven by slot setup, fixed seed, turn mode, rules fixture, human policy, and autonomous AI.
  - Keep catalog artifact validation separate: engine acceptance uses two pinned acceptance fixtures (Row 1 conquest and Row 2 diplomacy) and may not be blocked by unfinished 60-row campaign metadata.
  - Verify restart/resume with `Last-Event-ID`, terminal SSE/state agreement, AI autonomy traces, and bounded command/turn budgets.
  - Replace any broad file rewrites with moves/extractions from existing runner/scheduler/transport files.

  **Must NOT do**: put scenario-ID switches back; let runner submit AI commands; use diagnostics to choose actions; call external DB; classify harness scripts as production controller modes.

  **Recommended Agent Profile**:
  - Category: `deep` — executor architecture and black-box acceptance.
  - Skills: [`git-master`] — history-preserving file splits and atomic commit.
  - Omitted: [`playwright`] — no UI.

  **Parallelization**: Can Parallel: NO | Wave 2 | Blocks: 7–14 | Blocked By: 3,4,5

  **References**:
  - `internal/e2e/fullgame/{executor,run,runner,scheduler,diplomacy_scheduler,exploration_http,terminal,audit}.go`
  - `internal/e2e/onboarding/{public_runtime,http_client}.go`
  - `internal/gameplay/events/{events,http}.go`
  - `internal/runtime/runtime.go`

  **Acceptance Criteria**:
  - [ ] Executor has no scenario-ID behavior switch and accepts explicit deterministic policies/configuration.
  - [ ] Row 1 and Row 2 pass independently of `Validate(Catalog)` and each pass two clean runs plus restart.
  - [ ] AI command audit shows every AI mutation originated from the production AI gateway; human/puppet mutations originated from authenticated public clients.
  - [ ] Per-run deadlines and command/turn budgets fail with diagnostic transcript rather than hang.

  **QA Scenarios**:
  ```
  Scenario: Run both acceptance games through one generic executor
    Tool: Go testing + httptest/memorysql
    Steps: Supply Row 1 and Row 2 configs; run each clean twice and with Last-Event-ID restart.
    Expected: Both terminal naturally with stable transcripts/hashes and no scenario-ID branch.
    Evidence: .sisyphus/evidence/task-6-executor.txt

  Scenario: Budget, resume, and actor violations
    Tool: Go testing
    Steps: Force command budget exhaustion, stale SSE cursor, wrong actor, duplicated request ID, and restart during active turn.
    Expected: Stable bounded failure/resync/idempotent behavior; evidence is emitted; no goroutine/process hangs.
    Evidence: .sisyphus/evidence/task-6-executor-failure.txt
  ```

  **Commit**: YES | Message: `refactor(e2e): generalize public full-game execution` | Files: fullgame executor/transport/orchestration and tests

- [x] 7. Pass the complete pre-freeze production acceptance gate

  **Source review required**: for every remaining CTP2-specific production discrepancy discovered by the gate, inspect the pinned source before fixing it; if none, record that no further behavior inference was required.

  **What to do**:
  - Run and archive the full pre-freeze matrix: bridge/migrations, counts/controllers/civilizations, test-boundary scans, engine/service/world/AI/events, Row 1, Row 2, restart/replay, race, coverage, build, and protected-data checks.
  - Fix only genuine production defects using the mandatory RED–GREEN–REFACTOR protocol. Harness-only assertion/transport defects stay under `internal/e2e`.
  - Require zero known/provisional/disabled failures. Catalog’s unfinished 60 campaign rows are not part of this gate, but both pinned acceptance fixtures must have observed immutable hashes.
  - Tag the resulting evidence and commit boundary as the engine-freeze baseline.

  **Must NOT do**: freeze with red/skipped tests; include unobserved goldens; broaden production behavior beyond evidenced defects; run unbounded tests; allow wall-clock sleeps.

  **Recommended Agent Profile**:
  - Category: `unspecified-high` — broad acceptance and defect classification.
  - Skills: [`git-master`] — atomic fixes and baseline commit inspection.
  - Omitted: [`playwright`] — no browser.

  **Parallelization**: Can Parallel: NO | Wave 2 | Blocks: 8–14 | Blocked By: 2–6

  **References**:
  - Production candidates listed in Task 8 freeze manifest
  - `internal/e2e/fullgame/fullgame_test.go`
  - `internal/e2e/fullgame/validate.go`
  - `.github/workflows/coverage.yml`
  - `Taskfile.yml`

  **Acceptance Criteria**:
  - [ ] Focused/race/migration/import-boundary/protected-data commands all exit 0 with evidence.
  - [ ] Row 1 and Row 2 immutable acceptance artifacts are observed, reviewed, fixed, and repeat/restart identical.
  - [ ] No skipped, provisional, disabled, or known-red production acceptance test remains.
  - [ ] Repository build and bounded full package tests exit 0 before lock generation.

  **QA Scenarios**:
  ```
  Scenario: Full pre-freeze acceptance
    Tool: Bash + Go testing + ephemeral PostgreSQL
    Steps: Run the documented matrix from a clean committed tree with fixed resources/timeouts; aggregate exit codes and hashes.
    Expected: Every production gate exits 0; accepted hashes match repeats/restarts; migration container is removed.
    Evidence: .sisyphus/evidence/task-7-pre-freeze.txt

  Scenario: Deliberate regression detection
    Tool: Go testing in temporary worktree
    Steps: Temporarily perturb one deterministic expected value or hidden-state boundary in the temporary tree only; run the associated guard; discard temporary tree.
    Expected: Guard fails for the intended predicate, proving the gate is sensitive rather than vacuous.
    Evidence: .sisyphus/evidence/task-7-guard-sensitivity.txt
  ```

  **Commit**: YES | Message: `test(gameplay): establish engine freeze baseline` | Files: final acceptance tests/fixtures and any narrowly evidenced fixes

- [x] 8. Freeze production execution paths with a complete SHA-256 manifest and campaign allowlist

  **What to do**:
  - Generate and review `internal/e2e/freeze/engine-lock.json`, storing sorted relative paths and SHA-256 for:
    - all `internal/runtime/**` production files;
    - all `internal/gameplay/**` after testkit relocation;
    - production world files `{config,contracts,generator,models,recalculation,rng,serialization,snapshot,store}.go` that exist;
    - gameplay integration files `{create_game,endpoints,game_handler,helpers,models,player_endpoints,start_gameplay,world_contract_stubs}.go` that exist;
    - `main.go` and the four gameplay migrations.
  - Add a Go guard that rejects hash mismatch, missing files, undeclared new files beneath fully locked roots, duplicate manifest paths, and manifest entries outside the allowlist.
  - Add a production-import guard forbidding imports from `internal/e2e/**`.
  - Record the post-freeze campaign write allowlist: `internal/e2e/**` except lock/guard changes, synthetic fixtures, scripts/catalog data, `.github/workflows/**`, `Taskfile.yml`, `README.md`, and evidence. Any production-path need is a blocker requiring explicit user-approved freeze reopening.

  **Must NOT do**: rely on a Git ref, mutable environment baseline, or build tag; permit campaign tasks to update hashes; lock all `internal/game` or test/world fixture files; claim repository-local guard prevents malicious self-edit without required review.

  **Recommended Agent Profile**:
  - Category: `quick` — deterministic manifest/guard after boundary is decided.
  - Skills: [`git-master`] — verify exact freeze commit contents.
  - Omitted: [`playwright`] — static/Go guards.

  **Parallelization**: Can Parallel: NO | Wave 2 | Blocks: 9–14 | Blocked By: 7

  **References**:
  - `internal/runtime/`, `internal/gameplay/`
  - Selected `internal/world/*.go` and `internal/game/*.go` named above
  - Gameplay migrations named in Task 2
  - `internal/e2e/freeze/` — new lock/guard package
  - Oracle recommendation in planning session `ses_01278c23cffePQZ0Ump6PPi8m5`

  **Acceptance Criteria**:
  - [ ] Lock guard exits 0 on the accepted tree and enumerates every declared production path exactly once.
  - [ ] Temporary changed, added, deleted, duplicate, and out-of-scope manifest cases each fail with precise diagnostics.
  - [ ] Production import guard proves no production package imports `internal/e2e`.
  - [ ] Freeze commit contains the reviewed lock, guard/tests, CI invocation, and no production behavior change.

  **QA Scenarios**:
  ```
  Scenario: Accepted engine remains locked
    Tool: Go testing
    Steps: Enumerate roots, compare sorted manifest, recompute SHA-256, scan production imports.
    Expected: Exact path/hash equality and no forbidden import; exit 0.
    Evidence: .sisyphus/evidence/task-8-freeze.txt

  Scenario: Detect post-freeze mutation
    Tool: Go testing in temporary worktree
    Steps: Modify one locked file, add one file beneath internal/gameplay, delete one manifest file, and tamper one hash in separate subtests.
    Expected: Each case exits nonzero with changed/added/deleted/hash diagnostic; real worktree remains untouched.
    Evidence: .sisyphus/evidence/task-8-freeze-failure.txt
  ```

  **Commit**: YES | Message: `test(e2e): lock the accepted gameplay engine` | Files: `internal/e2e/freeze/**`, CI invocation only

- [ ] 9. Finalize catalog schema and Rows 01–12: conquest, elimination, and resignation

  **Source review required**: inspect pinned CTP2 victory/elimination/resignation behavior relevant to this family and record citations. Source review may change only campaign expectations; any engine discrepancy is a freeze-reopen blocker.

  **What to do**:
  - Refactor the checked-in 60-row catalog schema to explicit ordered player slots, imported civilization fixture IDs, seed, turn mode, map/terrain/rules fixture, human policy, autonomous-AI policy, terminal oracle, restart checkpoint, and immutable observed hashes.
  - Split row declarations with filesystem moves/extractions into disjoint files `catalog_rows_01_12.go`, `catalog_rows_13_24.go`, `catalog_rows_25_36.go`, `catalog_rows_37_48.go`, and `catalog_rows_49_60.go`; `catalog.go` owns schema/assembly only. Move every existing stable scenario ID into its assigned file before parallel work.
  - Keep exactly 60 stable scenario IDs. Complete Rows 01–12 for conquest/elimination/resignation across all four turn modes and representative 2–8 counts, ensuring this family includes one-human/rest-AI, all-AI, alternating, and test-controlled human policies.
  - Drive human actions by legal projection-derived policies; never prescribe AI commands or objectives. Use diagnostics only after actions for assertions.
  - Run every row twice clean and required rows with restart; record fixed artifacts manually after review.

  **Must NOT do**: modify any locked file or lock hash; retain composition fields; use hidden-state routing; skip rows; auto-update artifacts; make AI “build/capture X” scripts.

  **Recommended Agent Profile**:
  - Category: `deep` — deterministic multi-mode campaign authoring.
  - Skills: [`git-master`] — post-freeze allowlist and atomic family commit.
  - Omitted: [`playwright`] — HTTP/SSE runner.

  **Parallelization**: Can Parallel: NO | Wave 3A | Blocks: 10–14 | Blocked By: 8

  **References**:
  - `internal/e2e/fullgame/{catalog,validate,executor,run,runner,conquest}.go`
  - `internal/e2e/fullgame/fullgame_test.go`
  - `internal/e2e/freeze/engine-lock.json`
  - Existing Row 1 acceptance fixture from Tasks 5–7

  **Acceptance Criteria**:
  - [ ] Catalog remains exactly 60 unique IDs and validator rejects missing/extra/duplicate/provisional data.
  - [ ] Rows 01–12 naturally reach their declared terminal reasons and pass clean repeat/restart hash equality.
  - [ ] AI autonomy audit and hidden-state canary pass for every row.
  - [ ] Engine-lock guard exits 0 before and after the commit.

  **QA Scenarios**:
  ```
  Scenario: Conquest family completes deterministically
    Tool: Go test + public executor
    Steps: Run Rows 01–12 twice clean and designated restart checkpoints; collect terminal/state/event/schedule hashes.
    Expected: Every row reaches declared terminal reason and all repeated artifacts match exactly.
    Evidence: .sisyphus/evidence/task-9-conquest-family.txt

  Scenario: Invalid/provisional conquest row
    Tool: Go testing
    Steps: Remove slot, use hidden target/scripted AI instruction, omit hash, duplicate ID, or declare terminal inconsistent with state in table tests.
    Expected: Validator/audit rejects each case before acceptance; lock remains green.
    Evidence: .sisyphus/evidence/task-9-catalog-failure.txt
  ```

  **Commit**: YES | Message: `test(e2e): prove conquest and elimination campaigns` | Files: unlocked catalog/scripts/policies/tests for Rows 01–12

- [ ] 10. Complete Rows 13–24: diplomacy, alliances, trade, and diplomatic victory

  **Source review required**: inspect pinned diplomacy/contact/embassy/agreement/victory code for this family. Do not infer from community descriptions.

  **What to do**:
  - Implement projection-driven human diplomacy/trade policies and observation oracles for contact, embassy, proposal/response, agreement durability, trade, war/peace transitions, and diplomatic terminal result.
  - Let production AI decide autonomously whether/when to explore, establish contact/embassies, propose, accept, reject, or declare war. Tests assert outcomes appropriate to fixed seeds; they do not force AI actions.
  - Populate and verify Rows 13–24 across turn modes, counts, controller assignments, terrain/rules fixtures, and restart checkpoints.
  - Preserve Row 2 as a regression fixture within this family.

  **Must NOT do**: tell AI to create an embassy/proposal/agreement; inject contact; read hidden diplomacy state for action selection; modify frozen production files when a fixed seed produces a legitimate loss/rejection.

  **Recommended Agent Profile**:
  - Category: `deep` — autonomous diplomacy is long-horizon and stateful.
  - Skills: [`git-master`] — allowlisted campaign commit.
  - Omitted: [`playwright`] — transport-level tests.

  **Parallelization**: Can Parallel: YES | Wave 3B | Blocks: 14 | Blocked By: 9

  **References**:
  - `internal/e2e/fullgame/{diplomacy_scheduler,exploration,exploration_http,scheduler}.go`
  - `internal/e2e/fullgame/fullgame_test.go`
  - Existing Row 2 acceptance fixture
  - External pinned CTP2 diplomacy symbols recorded in Task 10 source evidence

  **Acceptance Criteria**:
  - [ ] Rows 13–24 pass two clean runs and designated restart runs with identical artifacts.
  - [ ] AI audit contains no runner-created AI submissions or objective-specific directives.
  - [ ] Agreement/contact/terminal observations are proven through public events/projections and post-action diagnostic audit.
  - [ ] Lock guard exits 0.

  **QA Scenarios**:
  ```
  Scenario: Autonomous diplomatic campaigns
    Tool: Go test + HTTP/SSE collector
    Steps: Run Rows 13–24 with fixed seeds; observe contacts/proposals/agreements and terminal results; repeat/restart.
    Expected: Declared observation/terminal oracles pass and deterministic artifacts match without directed AI actions.
    Evidence: .sisyphus/evidence/task-10-diplomacy-family.txt

  Scenario: Unauthorized or stale diplomacy command
    Tool: Go testing
    Steps: Submit proposal as wrong actor, respond twice, use stale revision, request hidden counterpart data, and command after terminal.
    Expected: Stable authorization/idempotency/conflict/redaction/terminal errors; no leaked fields or state mutation.
    Evidence: .sisyphus/evidence/task-10-diplomacy-failure.txt
  ```

  **Commit**: YES | Message: `test(e2e): prove autonomous diplomacy campaigns` | Files: unlocked diplomacy policies/scripts/catalog/tests for Rows 13–24

- [ ] 11. Complete Rows 25–36: cities, economy, buildings, wonders, science, and non-conquest victories

  **Source review required**: inspect pinned city production, economy, wonder, science/Gaia, and applicable victory behavior before defining oracles.

  **What to do**:
  - Extend deterministic human public policies for city founding/management, build queues, economy/public works, building/wonder completion, research progression, and applicable non-conquest victory prerequisites.
  - Configure synthetic rules records through the existing importer; do not embed retail costs/names/text.
  - Populate Rows 25–36 across fixed map/rules/turn/player dimensions. Production AI remains autonomous; human policy may pursue declared test goals using public state.
  - Assert durable queue/resource/research/wonder/terminal state and restart equality.

  **Must NOT do**: modify frozen rules/engine code; copy retail records; mutate resources directly; instruct AI to research/build a named objective; accept harness turn/time/command-cap exhaustion as success. Only a draw emitted by production rules is a valid draw terminal.

  **Recommended Agent Profile**:
  - Category: `deep` — long-running economic/state campaigns.
  - Skills: [`git-master`] — post-freeze commit guard.
  - Omitted: [`playwright`] — public API harness.

  **Parallelization**: Can Parallel: YES | Wave 3B | Blocks: 14 | Blocked By: 9

  **References**:
  - `internal/e2e/fullgame/scripts/city-economy-wonder.json`
  - `internal/e2e/fullgame/{scripts,scheduler,executor,audit}.go`
  - `internal/gameplay/rules/testdata/synthetic-ctp2/`
  - External pinned CTP2 city/economy/wonder/victory symbols recorded in Task 11 evidence

  **Acceptance Criteria**:
  - [ ] Rows 25–36 complete with declared durable economy/wonder/science/terminal oracles.
  - [ ] Synthetic importer path is identical to production import path and tracked-data guard exits 0.
  - [ ] Repeated and restart runs have identical queue, resource, research, event, state, and terminal hashes.
  - [ ] Lock guard exits 0.

  **QA Scenarios**:
  ```
  Scenario: Deterministic economy and victory progression
    Tool: Go test + public executor + diagnostic audit
    Steps: Run Rows 25–36 twice and at restart checkpoints; submit human public policies while AI runs normally.
    Expected: Declared production/research/wonder/terminal states occur durably and hashes match.
    Evidence: .sisyphus/evidence/task-11-economy-family.txt

  Scenario: Invalid economy command and missing source install
    Tool: Go testing
    Steps: Queue unavailable record, overspend, submit stale/duplicate command, and invoke retail import with absent path.
    Expected: Stable rejection/idempotency; no partial resource mutation; importer fails actionably without repository output.
    Evidence: .sisyphus/evidence/task-11-economy-failure.txt
  ```

  **Commit**: YES | Message: `test(e2e): prove economic and victory campaigns` | Files: unlocked scripts/catalog/policies/tests for Rows 25–36

- [ ] 12. Complete Rows 37–48: fog, movement, combat, special actions, and terraforming

  **Source review required**: inspect pinned visibility, movement/ZOC, combat, special-unit actions, and terraforming behavior represented by these synthetic rules.

  **What to do**:
  - Build deterministic human policies from visible frontier/history for exploration, movement/ZOC, army combat, special actions, and terraforming.
  - Populate Rows 37–48 using `fog-exploration-lifecycle`, `military-zoc-combat`, and `special-terraform` fixture families across modes/counts/controllers/terrain.
  - Prove fog memory survives restart, events/errors are viewer-filtered, combat/special outcomes are deterministic, and diagnostic post-audit agrees with public observations.
  - Retain hidden-state canaries in enemy units/cities/tiles and inspect raw serialized HTTP/SSE/AI/puppet bytes.

  **Must NOT do**: reveal canonical targets to policy; modify frozen visibility/combat/movement code; client-supply RNG/results; retry until a preferred random result; copy protected rules.

  **Recommended Agent Profile**:
  - Category: `deep` — visibility-sensitive deterministic campaigns.
  - Skills: [`git-master`] — lock-aware family commit.
  - Omitted: [`playwright`] — raw HTTP/SSE inspection is stronger.

  **Parallelization**: Can Parallel: YES | Wave 3B | Blocks: 14 | Blocked By: 9

  **References**:
  - `internal/e2e/fullgame/scripts/{fog-exploration-lifecycle,military-zoc-combat,special-terraform}.json`
  - `internal/e2e/fullgame/{exploration,movement,conquest,public_helpers,audit}.go`
  - External pinned CTP2 mechanics recorded in Task 12 source evidence

  **Acceptance Criteria**:
  - [ ] Rows 37–48 pass clean repeat/restart with identical visible histories, combat/special outcomes, events, and terminal artifacts.
  - [ ] Hidden canaries are absent from all raw public/AI/puppet bytes before visibility is legitimately gained.
  - [ ] Invalid movement/ZOC/special/terraform commands fail without revision/state leakage or partial mutation.
  - [ ] Lock guard exits 0.

  **QA Scenarios**:
  ```
  Scenario: Visibility-safe mechanics campaigns
    Tool: Go test + raw HTTP/SSE collector
    Steps: Run Rows 37–48 twice/restart; capture raw projections/events/errors and post-action audits.
    Expected: Mechanics and visible history are deterministic; no hidden canary leaks; declared terminal oracles pass.
    Evidence: .sisyphus/evidence/task-12-mechanics-family.txt

  Scenario: Illegal unseen or unavailable action
    Tool: Go testing
    Steps: Target unseen tile/unit, violate ZOC, reuse exhausted unit, invoke unavailable special/terraform record, and replay request ID with changed payload.
    Expected: Stable redacted errors; no state/event side effect; payload-bound idempotency rejects mismatch.
    Evidence: .sisyphus/evidence/task-12-mechanics-failure.txt
  ```

  **Commit**: YES | Message: `test(e2e): prove visibility and special-action campaigns` | Files: unlocked scripts/catalog/policies/tests for Rows 37–48

- [ ] 13. Complete Rows 49–60 and exhaustive lifecycle/disruption coverage

  **Source review required**: inspect pinned player-leave/AI takeover, turn-mode lifecycle, save/load, terminal, and applicable draw behavior before defining expectations.

  **What to do**:
  - Assign Rows 49–60 to player-count/controller/lifecycle coverage so every count 2–8 completes at least one full game; include explicit 5- and 7-player completion, all-human, all-AI, alternating, 2H/6AI, and 3H/2AI.
  - Add separate named disruption suites outside the 60 rows: all humans leave, two humans leave with deterministic AI takeover, disconnect/reconnect, process restart during active turn and immediately before terminal commit, stale/resumed SSE, duplicate/out-of-order commands, AI timeout/failure, terminal command rejection, production rules-defined draw behavior, and harness-cap failure behavior.
  - Use fake clock and deterministic barriers—not sleeps—to schedule disruptions. Human leave/takeover uses public endpoints; production AI then acts autonomously.
  - Verify 5-/7-player placement, diplomacy defaults, turn order, save/restart, and terminal completion, not merely lobby acceptance.

  **Must NOT do**: alter frozen code/lock; use test-only direct mutation to cause leave/takeover; count abort/crash/budget timeout as a valid full game; add spectators or matchmaking.

  **Recommended Agent Profile**:
  - Category: `deep` — concurrency, restart, and lifecycle matrix.
  - Skills: [`git-master`] — lock-aware atomic campaign/disruption commits.
  - Omitted: [`playwright`] — API/SSE lifecycle tests.

  **Parallelization**: Can Parallel: YES | Wave 3B | Blocks: 14 | Blocked By: 9

  **References**:
  - `internal/e2e/fullgame/{catalog,executor,run,scheduler,terminal,audit}.go`
  - `internal/e2e/onboarding/{onboarding,public_runtime,http_client}.go`
  - Existing lifecycle/fault requirements in superseded plan/notepads
  - External pinned CTP2 lifecycle symbols recorded in Task 13 evidence

  **Acceptance Criteria**:
  - [ ] Rows 49–60 complete deterministically and every player count 2–8 appears in at least one successful full game across the complete catalog.
  - [ ] Required controller boundary assignments and 5-/7-player placement/order/restart/terminal assertions pass.
  - [ ] Every named disruption suite has fixed checkpoint, expected status/event/state, bounded completion, and repeat equality.
  - [ ] No wall-clock sleeps, external DB, production test hook, or lock change exists.

  **QA Scenarios**:
  ```
  Scenario: Five- and seven-player lifecycle campaigns
    Tool: Go test + public executor
    Steps: Run designated 5- and 7-player Rows 49–60 twice and with active-turn restart; exercise alternating and 3H/2AI assignments.
    Expected: Deterministic placement, order, diplomacy defaults, AI autonomy, save/resume, and declared terminal result.
    Evidence: .sisyphus/evidence/task-13-player-count-family.txt

  Scenario: Deterministic disruption and failure handling
    Tool: Go test + fake clock/barriers
    Steps: Execute all named disconnect/leave/takeover/restart/SSE/duplicate/out-of-order/AI-failure/terminal-rejection cases.
    Expected: Each reaches its exact bounded oracle without deadlock, hidden leak, partial commit, or invalid terminal classification.
    Evidence: .sisyphus/evidence/task-13-disruptions.txt
  ```

  **Commit**: YES | Messages: `test(e2e): prove all player-count campaigns`; `test(e2e): cover gameplay lifecycle disruptions` | Files: unlocked catalog/policies/disruption tests for Rows 49–60

- [ ] 14. Aggregate the immutable catalog, CI gates, evidence, and Gameplay traceability

  **What to do**:
  - Merge the independently completed row-family edits without changing scenario IDs or accepted artifacts. Resolve overlapping `catalog.go` edits mechanically from the predeclared row ranges; rerun each family after merge.
  - Make `Validate(Catalog)` require exactly 60 enabled, non-provisional rows; all four turn modes; every count 2–8; declared terrain/map/rules/mechanic/victory coverage; explicit ordered slots; observed rules/state/event/schedule artifacts; and no composition fields.
  - Add bounded Taskfile/CI targets with these exact names: `test-gameplay-fast`, `test-gameplay-migrations`, `test-gameplay-race`, `test-gameplay-campaigns` (fixed shards 01–12, 13–24, 25–36, 37–48, 49–60), `test-gameplay-disruptions`, `test-gameplay-guards` (protected-data, production-import, replay, lock), and aggregate `test-gameplay-all`. Every PR runs `test-gameplay-all`; every target embeds the global inner/outer timeout and resource caps.
  - Update README Gameplay traceability only after linked tests pass. Document local user-owned rules import without copying records and document test-only diagnostics/puppet boundaries.
  - Run all commands from a clean committed tree; record exact environment and aggregate evidence.

  **Must NOT do**: edit frozen paths/lock; omit slow rows from PR CI; use unbounded parallelism/timeouts; mark README items complete without executable links; commit generated retail data or `.sisyphus/evidence`.

  **Recommended Agent Profile**:
  - Category: `unspecified-high` — cross-family aggregation and repository-wide verification.
  - Skills: [`git-master`] — merge row-range commits and produce isolated CI/docs commits.
  - Omitted: [`playwright`] — no browser deliverable.

  **Parallelization**: Can Parallel: NO | Wave 4 | Blocks: F1–F4 | Blocked By: 9–13

  **References**:
  - `internal/e2e/fullgame/{catalog,validate,fullgame_test}.go`
  - `.github/workflows/coverage.yml`
  - `Taskfile.yml`
  - `README.md:240-262` — Gameplay checklist/traceability target
  - `internal/e2e/freeze/engine-lock.json`
  - Task 9–13 evidence and stable row assignments

  **Acceptance Criteria**:
  - [ ] Catalog validator exits 0 for the exact accepted 60 rows and table tests reject every missing dimension/artifact/provisional/composition case.
  - [ ] `GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 20m task test-gameplay-all` exits 0 from clean HEAD and invokes every fixed shard/gate exactly once.
  - [ ] Every scenario passes twice clean; every designated row passes restart; no artifact differs.
  - [ ] README Gameplay items link to concrete automated tests and are checked only when those tests pass.
  - [ ] `GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m task test-gameplay-guards` exits 0 and proves path/hash equality, no undeclared files beneath locked roots, no production-to-E2E imports, and no protected-data signatures.

  **QA Scenarios**:
  ```
  Scenario: Complete all-PR gameplay gate
    Tool: Bash + Go testing + ephemeral PostgreSQL
    Steps: From clean HEAD run `GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 20m task test-gameplay-all`; aggregate exit code, row IDs, durations, hashes, coverage, lock/import/asset results.
    Expected: All commands exit 0; exactly 60 unique rows and every disruption suite execute; no external Supabase access or leaked container remains.
    Evidence: .sisyphus/evidence/task-14-all-pr-gate.txt

  Scenario: Validator and CI fail closed
    Tool: Go testing in temporary worktree
    Steps: Independently remove a row/hash/player count/turn mode, reintroduce composition, alter locked file, add protected fixture signature, and make a shard omit a row.
    Expected: The corresponding validator/lock/asset/shard-coverage guard fails precisely; real tree remains unchanged.
    Evidence: .sisyphus/evidence/task-14-fail-closed.txt
  ```

  **Commit**: YES | Messages: `test(e2e): enforce the complete gameplay catalog`; `ci: require deterministic gameplay campaigns`; `docs: link completed gameplay verification` | Files: unlocked catalog validator/tests, CI/Taskfile, README

## Final Verification Wave (MANDATORY — after ALL implementation tasks)
> Four review agents run in parallel without human intervention. After all four return APPROVE, present consolidated results and wait for explicit user approval before changing F1–F4 checkboxes or declaring completion. User approval is an administrative completion gate after verification, not part of verification execution. Rejection/feedback means fix, rerun all affected verification, present again, and wait.
- [ ] F1. Plan Compliance Audit — `oracle`
  - Verify every immutable TODO, source citation, evidence file, dependency, and acceptance command against this plan.
- [ ] F2. Code Quality and Security Review — `unspecified-high`
  - Review production/test boundaries, authorization, hidden-state leakage, deterministic concurrency, protected-data handling, and migration safety.
- [ ] F3. Real Manual QA — `unspecified-high`
  - Agent-execute representative 2-, 5-, 7-, and 8-player campaigns, restart/resume, terminal rejection, and failure paths through HTTP/SSE; no browser required.
- [ ] F4. Scope Fidelity and Freeze Audit — `deep`
  - Verify engine-lock hashes, post-freeze Git history, campaign allowlist, no production E2E imports/routes, no protected assets, and no out-of-scope mechanics.

## Commit Strategy
- Use repository-detected commit style and English language.
- Each implementation and its direct tests commit together; migration and migration tests commit together.
- Before each commit: `git diff --check`, staged-path allowlist assertion, clean-worktree patch verification, focused tests, and `go build ./...`.
- Never stage `.sisyphus/evidence`, unrelated user changes, blocked red files, or protected source data.
- Campaign commits after freeze must pass the engine-lock test before commit.

## Success Criteria
- Every Task 1–14 acceptance checkbox is green with agent-generated evidence.
- All counts 2–8 and arbitrary slot assignments satisfy the operational definitions.
- Every imported playable base civilization is supported without tracked retail data.
- Row 1, Row 2, all 60 fixed scenarios, and lifecycle/fault suites pass twice clean plus required restart checkpoints.
- Engine lock proves campaign work did not alter production execution code.
- `go build ./...`, bounded `go test ./...`, race checks, coverage, migration, asset, import-boundary, catalog, and freeze gates exit 0.
