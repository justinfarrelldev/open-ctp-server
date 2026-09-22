# CTP2 Gameplay Systems Completion Branch

## TL;DR
> **Summary**: Duplicate the current dirty gameplay workspace to a dedicated branch/worktree, surgically roll Row 09 back to provisional status there, then pivot from row-by-row acceptance to completing broad CTP2 gameplay-system functionality with synthetic, source-cited content and executable client-readiness gates.
> **Deliverables**:
> - Branch/worktree duplication with recoverability evidence
> - Row 09 acceptance rollback without touching accepted Rows 01–07
> - Explicit accepted-row validation model that supports provisional rows during breadth work
> - Mechanic-family completion for units, advances/unlocks, cities/economy, diplomacy/trade/government scaffolds, terrain/public works/pollution, special actions, combat/movement/transport, AI usage, and public API/projections
> - Cross-system client-ready scenarios proving imported content → runtime compile → engine behavior → HTTP/SSE projection → AI/human legal use
> - Final fullgame/catalog revalidation after gameplay breadth is mature
> **Effort**: XL
> **Parallel**: YES - 4 waves
> **Critical Path**: Task 1 → 2 → 3 → 4 → 5–11 → 12 → 13 → 14 → F1–F4

## Context
### Original Request
Make a new plan. Do not use more than 2 subagents for research tasks. Duplicate the current work to a new branch. On that branch, carefully revert the Row 09 work that has already been done. Do not continue at Row 09; instead finish gameplay functionality: units, advances, and the rest of the CTP2 systems. User explicitly requested specialist reviews and Momus high-accuracy review.

### Interview Summary
- This is an architecture-level replan, not immediate implementation.
- Research subagents are capped at 2 total for this planning task; review agents are allowed. No successful research subagent output was used.
- Current branch is `game-server-world` with a large dirty/untracked gameplay workspace.
- The new execution branch must preserve current tracked modifications and untracked non-ignored gameplay files before any rollback.
- Row 09 currently passes as an accepted fullgame row, but the new branch must intentionally revert that acceptance and stop treating Row 09 as the next progress target.
- “Client-ready” is bounded to: deterministic server-side gameplay systems, public HTTP/SSE contracts, imported/synthetic rules coverage, restart-safe persistence, and basic legal AI use. It does **not** mean full retail CTP2 parity, polished UI, or strategically strong AI.

### Research Findings
- `internal/e2e/fullgame/validate.go` has `ValidateCatalogSchema(Catalog)` for 60 stable row shape and `Validate(Catalog)` for final immutable artifacts.
- `ValidateRows01_12Final(catalog[:12])` currently makes Rows 01–12 positional final, which conflicts with reverting Row 09 to provisional.
- `internal/e2e/fullgame/catalog_rows_01_12.go` currently has observed artifacts for Rows 01–07 and 09; Rows 08 and 10–60 are provisional.
- Row 09 rollback surface:
  - `internal/e2e/fullgame/catalog_rows_01_12.go`: Row 09 `observedArtifacts("account-1", "13f89849...", "8504b909...")`.
  - `internal/e2e/fullgame/fullgame_test.go`: `TestCatalogRowNineDeclaresSequentialScoreReachability` asserts Row 09 observed artifacts.
  - `internal/e2e/fullgame/task9_rows_test.go`: `TestRowNineObservedSequentialScoreRestartsAtCompletedTurnTwelve` asserts Row 09 winner/hash/AI commands.
- Rules importer currently recognizes `Units.txt`, `buildings.txt`, `Wonder.txt`, `terrain.txt`, `Const.txt`, `EndGameObjects.txt`, and `Civilisation.txt` via `internal/gameplay/rules/rules.go:CatalogFiles`.
- Runtime rules compile units/buildings/wonders/terrain/endgame mostly generically and require `UNIT_SYNTHETIC`; this is not complete CTP2 content.
- Supported manifest advertises movement/ZOC, unit actions, armies/transports, combat, city founding/production/upkeep, building/wonder/effects, special embassy/soothsaying/steal/terraform, terrain transformation, and victory evaluators.
- Engine command constants include unit, army, city, combat, special, terraform, policy, leave/takeover actions; breadth gaps remain around imported unit stats/capabilities, advances/prerequisites/unlocks, governments, trade, public works/roads, pollution effects, richer diplomacy, and AI use.

### Oracle Review (gaps addressed)
- Protect the dirty tree first with explicit branch/worktree preservation and recovery evidence.
- Treat Row 09 as a freeze revert, not a redesign: keep ID, seed, mode, slot count/order, world fixture, victory, budgets, and restart plan stable; remove observed-artifact freeze and Row 09-specific frozen assertions.
- Replace positional “Rows 01–12 are final” validation with an explicit accepted-set model or breadth-phase validation will deadlock.
- Stop using catalog row count as the main progress signal. Use mechanic-family gates.
- Expand with synthetic content only; do not copy retail data.
- Freeze and full campaign artifacts come after layered importer/runtime/engine/API/AI/cross-system gates pass.

### Metis Review (gaps addressed)
- Add executable branch preservation criteria before any branch or revert work.
- Define “client-ready” as ranked mechanic families with explicit deferrals.
- Include exact Row 09 rollback pass/fail matrix and impact checks.
- Include an accepted-row validation model before breadth work.
- Every mechanic family must have importer/runtime, engine, public API/projection, AI/basic usage, happy path, and failure/edge acceptance.
- Include at least one cross-system scenario where an advance unlocks content, production creates it, the engine uses it, HTTP/SSE exposes it, and AI can issue a legal command involving it.

## Work Objectives
### Core Objective
Create a safe gameplay-systems branch that preserves the current work, removes Row 09 acceptance as a misleading progress target, and completes server-side CTP2-style gameplay breadth to a deterministic, public-API, client-ready slice using source-cited synthetic content and strict TDD.

### Deliverables
- New branch/worktree: `gameplay-systems-completion`
- Recovery evidence proving current dirty tracked/untracked work was preserved before branching
- Row 09 rollback evidence proving only Row 09 acceptance state was demoted
- Explicit accepted-row registry replacing positional final validation
- Expanded rules import/runtime model for units, advances, buildings, wonders, terrain improvements, governments/trade scaffolds, and effects
- Engine behavior for major gameplay families with deterministic persistence/restart semantics
- Public HTTP/SSE/projection contracts for every included system
- AI planner/driver capable of basic legal use of each included major mechanic from public projection/history only
- Cross-system end-to-end gameplay scenarios focused on mechanic breadth rather than catalog row count
- Final evidence bundle and Momus-approved plan execution review

### Definition of Done (verifiable conditions with commands)
- Branch duplication evidence exists under `.sisyphus/evidence/ctp2-systems-task-1-branch-preservation.txt` and shows source and destination dirty-path inventories match.
- Row 09 rollback evidence exists under `.sisyphus/evidence/ctp2-systems-task-2-row09-rollback.txt` and proves Row 09 is provisional while Rows 01–07 remain accepted.
- Focused gameplay stack exits 0:
  `GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m go test -timeout=8m -p=2 ./internal/game ./internal/gameplay/... ./internal/world ./internal/runtime ./internal/e2e/memorysql ./internal/e2e/onboarding ./internal/e2e/fullgame -count=1`
- Full repository build exits 0:
  `GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m go build -p=2 ./...`
- Freeze guard exits 0 before and after gameplay-system work unless an explicit freeze-reopen task is included and reviewed:
  `GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 5m go test -timeout=4m -p=2 ./internal/e2e/freeze -run '^TestEngineLock$' -count=1`
- New mechanic-family tests exit 0 and include importer/runtime, engine, public HTTP/SSE/projection, AI, restart, and invalid-command cases.
- No test, script, catalog, or AI path uses hidden canonical state for action selection.
- `git diff --check` exits 0.
- Final verification wave F1–F4 all approve, then the user gives explicit administrative okay before completion is marked.

### Must Have
- Branch/worktree preservation before rollback or code changes.
- TDD RED → GREEN → REFACTOR evidence for every production behavior change.
- Pinned CTP2 source citations before adding CTP2-specific mechanics.
- Synthetic test records only; no protected retail data copied or generated into tracked files.
- Public HTTP/SSE and projection-driven tests for client-facing behavior.
- Restart equality for state, events, visible projections, and persisted content where relevant.
- Basic AI legality coverage for every included mechanic family.
- Explicit accepted-row registry; no positional assumption that Rows 01–12 are all final.

### Must NOT Have
- Continuing Row 09 as the next campaign target.
- Row 09 observed artifacts left accepted after rollback.
- Broad `git reset`, `git checkout .`, `git clean`, or stash operations that can destroy dirty work.
- Fabricated fullgame artifacts or success-by-budget-exhaustion.
- Retail CTP2 data committed to the repo.
- Direct DB/store mutation for gameplay actions in E2E tests.
- Hidden-state targeting by human harness, puppet, or AI.
- Strategic-AI competence claims beyond “basic legal use” unless separately implemented and verified.

## Verification Strategy
> ZERO HUMAN INTERVENTION - all verification is agent-executed.
- Test decision: TDD for production mechanics; tests-after only for pure plan/branch evidence; no unverified mechanics.
- QA policy: Every task has agent-executed happy and failure scenarios.
- Evidence: `.sisyphus/evidence/ctp2-systems-task-{N}-{slug}.txt` and source notes as `.md`.
- No Playwright required: this server scope has no browser UI.
- All commands unset Supabase/network DB variables unless a task explicitly uses an isolated ephemeral PostgreSQL container.

## Execution Strategy
### Parallel Execution Waves
> Target: 5-8 tasks per wave. <3 per wave (except final) = under-splitting.
> Extract shared dependencies as Wave-1 tasks for max parallelism.

Wave 1: Tasks 1–4 — branch preservation, Row 09 rollback, accepted-row validation model, source/content inventory
Wave 2: Tasks 5–10 — mechanic-family implementation gates in parallel after shared content model is ready
Wave 3: Tasks 11–13 — AI/public API/cross-system scenarios and freeze reconciliation
Wave 4: Task 14 — final evidence, CI, documentation, and readiness report

### Dependency Matrix (full, all tasks)
| Task | Depends On | Blocks |
|---|---|---|
| 1 | — | 2–14 |
| 2 | 1 | 3–14 |
| 3 | 2 | 4–14 |
| 4 | 3 | 5–13 |
| 5 | 4 | 11–13 |
| 6 | 4 | 11–13 |
| 7 | 4 | 11–13 |
| 8 | 4 | 11–13 |
| 9 | 4 | 11–13 |
| 10 | 4 | 11–13 |
| 11 | 5–10 | 12–14 |
| 12 | 5–11 | 13–14 |
| 13 | 5–12 | 14 |
| 14 | 1–13 | F1–F4 |

### Agent Dispatch Summary (wave → task count → categories)
- Wave 1: 4 tasks — `deep`, `quick`, `deep`, `deep`
- Wave 2: 6 tasks — `deep` × 6, parallel by mechanic family
- Wave 3: 3 tasks — `deep`, `unspecified-high`, `deep`
- Wave 4: 1 task — `unspecified-high`

## TODOs

- [x] 1. Duplicate the current dirty gameplay workspace to a protected systems branch

  **What to do**: Create a recovery snapshot before branching, then duplicate the current dirty tracked and untracked non-ignored gameplay workspace into branch/worktree `gameplay-systems-completion`. Preferred method: keep the original worktree untouched, create a sibling worktree from current `HEAD`, apply a tracked diff patch, then copy untracked non-ignored files listed by `git ls-files --others --exclude-standard`. Record source branch, HEAD, tracked diff hash, untracked file list hash, destination branch, and destination verification.
  **Must NOT do**: run `git reset`, `git checkout .`, `git clean`, destructive stash operations, or switch the existing dirty tree without first creating recovery artifacts.

  **Recommended Agent Profile**:
  - Category: `deep` - Reason: dirty-tree preservation and branch safety are high risk.
  - Skills: [`git-master`] - Required for safe branch/worktree operations.
  - Omitted: [`playwright`] - No browser surface.

  **Parallelization**: Can Parallel: NO | Wave 1 | Blocks: 2–14 | Blocked By: none

  **References**:
  - Branch state: `git branch --show-current` currently reported `game-server-world`.
  - Dirty inventory: `git status --short` shows broad tracked and untracked gameplay work.
  - Plan artifact: `.sisyphus/plans/ctp2-gameplay-systems-branch.md`.

  **Acceptance Criteria**:
  - [ ] `.sisyphus/evidence/ctp2-systems-task-1-branch-preservation.txt` records source HEAD, source branch, destination branch, tracked diff SHA-256, untracked file list SHA-256, and destination verification.
  - [ ] `git -C ../open-ctp-server-gameplay-systems branch --show-current` prints `gameplay-systems-completion` or the evidence records an equivalent branch path.
  - [ ] Destination `git diff --stat` matches source tracked diff stat before rollback.
  - [ ] Destination untracked non-ignored file inventory matches the source inventory, excluding explicitly recorded generated binaries such as `fullgame.test` only if they are reproducible.

  **QA Scenarios**:
  ```
  Scenario: Dirty workspace duplicated safely
    Tool: Bash
    Steps: Capture source `git status --short`, `git diff --binary`, and `git ls-files --others --exclude-standard`; create sibling worktree branch; apply patch; copy untracked non-ignored paths; compare inventories and diff stats.
    Expected: Source worktree remains unchanged; destination branch has matching gameplay work and recoverability evidence.
    Evidence: .sisyphus/evidence/ctp2-systems-task-1-branch-preservation.txt

  Scenario: Recovery snapshot works
    Tool: Bash
    Steps: In `/tmp/opencode`, create a disposable clone/worktree from source HEAD; apply the recorded patch and untracked archive/list.
    Expected: Disposable reconstruction reproduces the destination inventory; no destructive commands were used in the source tree.
    Evidence: .sisyphus/evidence/ctp2-systems-task-1-recovery.txt
  ```

  **Commit**: NO | Message: n/a | Files: evidence only until branch safety is confirmed

- [x] 2. Surgically demote Row 09 acceptance on the new branch

  **What to do**: On `gameplay-systems-completion`, revert only Row 09 acceptance/freeze state. Keep Row 09 ID, name, seed, mode, world fixture, slots, human policy, victory, script, budgets, and restart plan stable. Replace Row 09 `observedArtifacts(...)` with `ObservedArtifacts{}`. Rewrite `TestCatalogRowNineDeclaresSequentialScoreReachability` to assert Row 09 declaration shape and provisional observed artifacts. Delete or demote `TestRowNineObservedSequentialScoreRestartsAtCompletedTurnTwelve` so it is not an acceptance gate.
  **Must NOT do**: modify accepted Rows 01–07, Row 08 blocker status, Rows 10–60, executor scheduling, production gameplay code, or fullgame artifacts for unrelated rows.

  **Recommended Agent Profile**:
  - Category: `quick` - Reason: narrow test/catalog rollback.
  - Skills: [] - No special skill required unless committing.
  - Omitted: [`playwright`] - No UI.

  **Parallelization**: Can Parallel: NO | Wave 1 | Blocks: 3–14 | Blocked By: 1

  **References**:
  - Row 09 declaration: `internal/e2e/fullgame/catalog_rows_01_12.go` function `catalogRows01_12`.
  - Row 09 catalog guard: `internal/e2e/fullgame/fullgame_test.go` function `TestCatalogRowNineDeclaresSequentialScoreReachability`.
  - Row 09 acceptance test: `internal/e2e/fullgame/task9_rows_test.go` function `TestRowNineObservedSequentialScoreRestartsAtCompletedTurnTwelve`.
  - Current Row 09 hashes to remove: state `13f89849c047ab5ea09180597a3efe6e57f52cb1be674763182974585aa7dea3`, events `8504b909dc83ca8c63fded829e7e6a9532d87e86063fe5938619dc5e38bf9550`.

  **Acceptance Criteria**:
  - [ ] `grep -R '13f89849c047ab5ea09180597a3efe6e57f52cb1be674763182974585aa7dea3\|8504b909dc83ca8c63fded829e7e6a9532d87e86063fe5938619dc5e38bf9550\|TestRowNineObservedSequentialScoreRestartsAtCompletedTurnTwelve' internal/e2e/fullgame` returns no matches.
  - [ ] Row 09 still has ID `fullgame-09-sequential-score_2300`, sequential mode, small/cold/wetness-50 world, one human plus two AI slots, `score_2300`, `scripts/city-economy-wonder.json`, default budgets, and completed-turn-12 restart.
  - [ ] Rows 01–07 observed artifact literals are unchanged.
  - [ ] `go test ./internal/e2e/fullgame -run 'TestCatalogRow.*|TestCatalogRows01_12Cover|TestCatalogRejectsProvisionalScenarios' -count=1` exits 0 with Row 08 or Row 09 provisional behavior explicitly expected.

  **QA Scenarios**:
  ```
  Scenario: Row 09 demoted but shape retained
    Tool: Bash + Go test
    Steps: Run the grep above; run targeted catalog tests; inspect Row 09 declaration.
    Expected: Row 09 no longer has immutable artifacts or an acceptance test, but retains scenario shape for future use.
    Evidence: .sisyphus/evidence/ctp2-systems-task-2-row09-rollback.txt

  Scenario: Accepted Rows 01-07 preserved
    Tool: Bash + Go test
    Steps: Run Row 01-07 catalog artifact guards and compare literals against pre-rollback inventory.
    Expected: No accepted Row 01-07 state/event hash changes; rollback blast radius is Row 09-only.
    Evidence: .sisyphus/evidence/ctp2-systems-task-2-accepted-row-preservation.txt
  ```

  **Commit**: YES | Message: `test(e2e): demote row nine acceptance on systems branch` | Files: `internal/e2e/fullgame/catalog_rows_01_12.go`, `internal/e2e/fullgame/fullgame_test.go`, `internal/e2e/fullgame/task9_rows_test.go`, evidence

- [x] 3. Replace positional fullgame final validation with an explicit accepted-row registry

  **What to do**: Refactor final catalog validation so accepted rows are explicit by scenario ID, not positional `catalog[:12]`. Initial accepted set after Row 09 rollback must include Rows 01–07 only. Row 08 and Row 09 are provisional and must fail final validation if final-all mode is requested, but must not block mechanic-breadth development gates that validate schema and accepted rows only.
  **Must NOT do**: make provisional rows look accepted, delete the 60-row catalog, weaken schema coverage, or hide Row 08’s known blocker.

  **Recommended Agent Profile**:
  - Category: `deep` - Reason: validation semantics affect all later work.
  - Skills: [] - Git skill only if committing.
  - Omitted: [`playwright`] - No UI.

  **Parallelization**: Can Parallel: NO | Wave 1 | Blocks: 4–14 | Blocked By: 2

  **References**:
  - Current final validator: `internal/e2e/fullgame/validate.go` functions `Validate`, `ValidateRows01_12Final`, `ValidateCatalogSchema`.
  - Provisional test: `internal/e2e/fullgame/fullgame_test.go` function `TestCatalogRejectsProvisionalScenarios`.
  - Row declarations: `internal/e2e/fullgame/catalog_rows_01_12.go`.

  **Acceptance Criteria**:
  - [ ] New validation API has separate gates: schema validation, accepted-row validation, and final-all validation.
  - [ ] Accepted-row validation passes for Rows 01–07 and ignores provisional Row 08/09/10–60 only by explicit ID policy.
  - [ ] Final-all validation still fails with a precise provisional-row diagnostic.
  - [ ] Tests prove removing an accepted Row 01–07 artifact fails accepted-row validation.
  - [ ] Tests prove adding Row 09 back to accepted registry without artifacts fails.

  **QA Scenarios**:
  ```
  Scenario: Breadth-phase validation is not blocked by provisional rows
    Tool: Go test
    Steps: Run schema validation and accepted-row validation with Row 09 provisional.
    Expected: Both pass; final-all validation still fails on the first explicit provisional scenario.
    Evidence: .sisyphus/evidence/ctp2-systems-task-3-validation-model.txt

  Scenario: Accepted-row registry catches drift
    Tool: Go test
    Steps: In table tests, remove a Row 01 hash, corrupt a Row 02 winner, and add Row 09 to the accepted registry without artifacts.
    Expected: Each mutation fails with row-specific diagnostics.
    Evidence: .sisyphus/evidence/ctp2-systems-task-3-validation-failure.txt
  ```

  **Commit**: YES | Message: `test(e2e): track accepted full-game rows explicitly` | Files: `internal/e2e/fullgame/validate.go`, `internal/e2e/fullgame/*_test.go`, evidence

- [ ] 4. Build the CTP2 mechanic inventory and source-cited synthetic content contract

  **What to do**: Inventory CTP2 mechanics required for the bounded client-ready slice and map each to importer/runtime/engine/API/AI status. Inspect pinned CTP2 source before making behavior decisions. Define synthetic record formats for units, advances, buildings, wonders, terrain, terrain improvements, governments, trade routes/agreements, pollution sources, and special actions. Produce a machine-readable mechanic matrix used by tests.
  **Must NOT do**: copy retail data, translate protected names/text into fixtures, or claim unsupported systems are complete.

  **Recommended Agent Profile**:
  - Category: `deep` - Reason: source review and scope boundary definition.
  - Skills: [] - External source can be inspected with normal read/fetch tools by executor.
  - Omitted: [`playwright`] - No UI.

  **Parallelization**: Can Parallel: NO | Wave 1 | Blocks: 5–13 | Blocked By: 3

  **References**:
  - Import file list: `internal/gameplay/rules/rules.go` variable `CatalogFiles`.
  - Runtime compile: `internal/gameplay/rules/runtime.go` functions `CompileRuntime`, `compileRecord`.
  - Supported manifest: `internal/gameplay/rules/rules.go` function `SupportedManifest`.
  - Existing synthetic fixtures: `internal/gameplay/rules/testdata/synthetic-ctp2/`.
  - Pinned upstream: `civctp2/civctp2@0668dd199c2519ff8ab7d2871e7c7a89e21c8e52`.

  **Acceptance Criteria**:
  - [ ] `.sisyphus/evidence/ctp2-systems-task-4-source-inventory.md` cites exact upstream files/symbols inspected and implementation consequences.
  - [ ] Mechanic matrix distinguishes `included-now`, `deferred`, and `explicitly-excluded` systems.
  - [ ] Synthetic content contract covers all included systems without retail literals.
  - [ ] Protected-data scans pass against all new fixtures.
  - [ ] Tests fail closed if a mechanic is marked included without importer/runtime/engine/API/AI status.

  **QA Scenarios**:
  ```
  Scenario: Mechanic matrix is complete and bounded
    Tool: Go test + repository scan
    Steps: Load the matrix, assert required statuses for all included families, and scan fixtures for protected retail markers.
    Expected: Included systems have implementation gates; deferred/excluded systems are explicit; no protected data is present.
    Evidence: .sisyphus/evidence/ctp2-systems-task-4-mechanic-matrix.txt

  Scenario: Unsupported system cannot masquerade as complete
    Tool: Go test
    Steps: Add table-test cases with included mechanic missing runtime or API status.
    Expected: Validation fails with the missing layer named.
    Evidence: .sisyphus/evidence/ctp2-systems-task-4-matrix-failure.txt
  ```

  **Commit**: YES | Message: `test(gameplay): define source-cited mechanic readiness matrix` | Files: rules/e2e matrix fixtures/tests/evidence only

- [x] 5. Expand imported unit definitions into runtime unit capabilities

  **What to do**: Extend synthetic/imported unit records beyond `MaxMovePoints` into a deterministic runtime unit definition model: domains, movement, attack, defense, health, armor, ZOC, transport capacity, cargo compatibility, special effects, production cost/upkeep, prerequisites, obsolete-by, and visibility effects. Start-game setup must instantiate starter units from runtime definitions instead of hard-coded generic starter values where rules specify them.
  **Must NOT do**: copy retail unit stats, silently accept unsupported tags, or break existing `UNIT_SYNTHETIC` fixtures.

  **Recommended Agent Profile**:
  - Category: `deep` - Reason: importer/runtime/start-state/engine coupling.
  - Skills: [] - No browser needed.
  - Omitted: [`playwright`] - API and Go tests suffice.

  **Parallelization**: Can Parallel: YES | Wave 2 | Blocks: 11–13 | Blocked By: 4

  **References**:
  - Importer: `internal/gameplay/rules/rules.go` `CatalogFiles`, parser/validation functions.
  - Runtime: `internal/gameplay/rules/runtime.go` `RuntimeRules`, `RuntimeStarter`, `compileRecord`.
  - Start state: `internal/game/start_gameplay.go` starter-unit creation.
  - Engine unit behavior: `internal/gameplay/engine/units.go`.
  - Visibility: `internal/gameplay/visibility/visibility.go` unit projections.

  **Acceptance Criteria**:
  - [ ] RED tests prove current runtime drops at least domain/combat/transport/prerequisite metadata from synthetic unit records.
  - [ ] Runtime unit definitions preserve deterministic sorted IDs and reject malformed/unsupported capability tags.
  - [ ] Start-game creates units from runtime definitions and preserves restart equality.
  - [ ] Unit projections expose only client-safe capability fields needed to render legal actions.
  - [ ] Existing focused gameplay tests and Row 01–07 accepted-row validation remain green.

  **QA Scenarios**:
  ```
  Scenario: Imported unit capabilities drive gameplay
    Tool: Go test + httptest
    Steps: Import synthetic land/sea/transport/special units; start a game; inspect public projection; submit legal move/embark/disembark/army commands.
    Expected: Capabilities match runtime records; legal commands succeed; restart preserves unit stats/cargo/effects.
    Evidence: .sisyphus/evidence/ctp2-systems-task-5-units.txt

  Scenario: Invalid or hidden unit data rejected
    Tool: Go test
    Steps: Import malformed unit stats, unsupported required tags, invalid cargo references, hidden enemy-only metadata, and impossible domains.
    Expected: Import/runtime validation fails closed or projection redacts hidden fields; no partial state mutation.
    Evidence: .sisyphus/evidence/ctp2-systems-task-5-units-failure.txt
  ```

  **Commit**: YES | Message: `feat(gameplay): compile imported unit capabilities` | Files: rules/runtime/start/engine/visibility tests and implementation

- [x] 6. Implement advances, prerequisites, research progress, and unlock semantics

  **What to do**: Add synthetic `Advance`/technology import support and runtime definitions. Implement nation research state, research-point accrual, advance completion, prerequisite checks, obsolete-by effects, and unlock gates for units/buildings/wonders/special actions/victory prerequisites. Integrate with steal-technology and Gaia requirements without replacing existing accepted behavior.
  **Must NOT do**: fake a full retail tech tree, make all content always available, or let steal/research expose hidden target advances.

  **Recommended Agent Profile**:
  - Category: `deep` - Reason: cross-cutting rules, nation state, production gates, victory gates.
  - Skills: [] - No browser needed.
  - Omitted: [`playwright`] - API and engine tests suffice.

  **Parallelization**: Can Parallel: YES | Wave 2 | Blocks: 11–13 | Blocked By: 4

  **References**:
  - Nation state: `internal/gameplay/engine/engine.go` struct `Nation`.
  - City/effects: `internal/gameplay/engine/cities.go` functions `queueProduction`, `completeProduction`, `hasPrerequisite`.
  - Special steal: `internal/gameplay/engine/special.go` function `stealTechnology`.
  - Gaia victory: `internal/gameplay/engine/lifecycle.go` functions `gaiaPrerequisiteIDs`, `nationHasGaiaPrerequisites`.
  - Runtime rules: `internal/gameplay/rules/runtime.go`.

  **Acceptance Criteria**:
  - [x] Importer accepts synthetic advance records with prerequisites and rejects cycles/missing IDs.
  - [x] Research accrues deterministically at round completion and completes advances in stable order.
  - [x] Production/special/victory gates honor prerequisites and obsolete-by semantics.
  - [x] Public projection exposes actor-owned known advances/research progress only.
  - [x] Steal-technology uses public/legal targeting and deterministic RNG without leaking hidden tech lists.

  **QA Scenarios**:
  ```
  Scenario: Advance unlocks content end-to-end
    Tool: Go test + public HTTP
    Steps: Import synthetic advance A unlocking unit/building; run rounds until research completes; queue and complete unlocked content; inspect projection and restart.
    Expected: Content unavailable before A, available after A, persists across restart, and emits deterministic events.
    Evidence: .sisyphus/evidence/ctp2-systems-task-6-advances.txt

  Scenario: Invalid prerequisites and hidden tech access fail
    Tool: Go test
    Steps: Import cyclic/missing prerequisites; submit production before prerequisite; attempt steal against unseen/unauthorized target.
    Expected: Import or command fails closed with stable errors; no hidden advances leak.
    Evidence: .sisyphus/evidence/ctp2-systems-task-6-advances-failure.txt
  ```

  **Commit**: YES | Message: `feat(gameplay): model advances and unlock prerequisites` | Files: rules/runtime/engine/visibility/service/http tests and implementation

- [x] 7. Complete city economy, production, upkeep, buildings, wonders, and public works

  **What to do**: Extend city/nation economy from generic synthetic effects to a deterministic model for food, production, gold, science, public works, upkeep, building effects, wonder uniqueness/scope, queue validation, queue cancellation/refund policy, and starvation/maintenance failure behavior if included by the mechanic matrix.
  **Must NOT do**: mutate resources directly in tests, use wall-clock time, or copy retail costs/names.

  **Recommended Agent Profile**:
  - Category: `deep` - Reason: economic state is broad and persistence-sensitive.
  - Skills: [] - No UI.
  - Omitted: [`playwright`] - Server tests suffice.

  **Parallelization**: Can Parallel: YES | Wave 2 | Blocks: 11–13 | Blocked By: 4

  **References**:
  - City commands/effects: `internal/gameplay/engine/cities.go`.
  - Runtime production: `internal/gameplay/rules/runtime.go` `RuntimeProduction`, `RuntimeEffect`.
  - Existing tests: `internal/gameplay/engine/cities_task10_test.go`, `internal/gameplay/rules/runtime_test.go`.
  - Public API: `internal/gameplay/http/http.go`, `internal/gameplay/service/service.go`.

  **Acceptance Criteria**:
  - [x] Economy calculations are deterministic, sorted, and covered by canonical state hashes.
  - [x] Buildings/wonders apply scoped effects and enforce global/national/city uniqueness as specified by synthetic records.
  - [x] Queue/cancel/upkeep failure cases are stable and idempotent.
  - [x] Public projection shows enough city/economy data for clients to render legal choices without hidden fields.
  - [x] Restart equality covers city queues, resources, buildings, wonders, and public works.

  **QA Scenarios**:
  ```
  Scenario: City produces unlocked building and wonder
    Tool: Go test + public HTTP
    Steps: Found city, set production queue, advance rounds, complete building and wonder, verify effects/resources/projection, restart and compare.
    Expected: Resources and effects match synthetic runtime; duplicate wonder is rejected; state/events deterministic.
    Evidence: .sisyphus/evidence/ctp2-systems-task-7-economy.txt

  Scenario: Invalid economy commands fail atomically
    Tool: Go test
    Steps: Queue unavailable item, overspend public works, cancel nonexistent queue, duplicate request ID with changed payload, and trigger upkeep insufficiency case.
    Expected: Stable rejections or deterministic upkeep behavior; no partial mutation.
    Evidence: .sisyphus/evidence/ctp2-systems-task-7-economy-failure.txt
  ```

  **Commit**: YES | Message: `feat(gameplay): complete deterministic city economy` | Files: engine/rules/service/http/visibility tests and implementation

- [x] 8. Complete terrain, tile improvements, roads, public works, and pollution systems

  **What to do**: Extend terrain runtime and engine overlays to cover deterministic tile improvements: roads/movement modifiers, farms/mines or generic yield improvements, terraform lifecycle, pollution generation/cleanup, public works cost, and projection/event redaction. Tie world pollution read endpoints to gameplay state where appropriate.
  **Must NOT do**: conflate old world-scaffold pollution with gameplay pollution unless tests define the bridge, or expose hidden tile pollution before visibility.

  **Recommended Agent Profile**:
  - Category: `deep` - Reason: terrain affects movement, economy, visibility, and public projections.
  - Skills: [] - No UI.
  - Omitted: [`playwright`] - API tests suffice.

  **Parallelization**: Can Parallel: YES | Wave 2 | Blocks: 11–13 | Blocked By: 4

  **References**:
  - Terrain runtime: `internal/gameplay/rules/runtime.go` `RuntimeMovement`, `RuntimeTerrain`.
  - Terraform engine: `internal/gameplay/engine/special.go` `applySpecialCommand`, `advanceTerraform`, `EffectiveTerrain`.
  - Movement engine: `internal/gameplay/engine/units.go` `terrainCost`, `domainCanEnter`.
  - World pollution scaffold: `internal/world/recalculation.go`, `internal/game/world_ratings_pollution_endpoints.go`.

  **Acceptance Criteria**:
  - [x] Synthetic terrain-improvement records compile into runtime definitions.
  - [x] Roads/improvements modify movement/economy deterministically only after completion.
  - [x] Pollution is generated and cleaned according to synthetic rules and visible only when appropriate.
  - [x] Public works costs and cancellation/refund semantics are deterministic.
  - [x] Public projection, SSE, canonical state, and restart equality cover overlay lifecycles.

  **QA Scenarios**:
  ```
  Scenario: Public works improvement changes movement and yield
    Tool: Go test + public HTTP/SSE
    Steps: Queue road/improvement/terraform on visible tile, advance completion, move unit over improved path, recalculate city yield, restart.
    Expected: Movement/yield changes only after completion; events and projections match; hashes stable.
    Evidence: .sisyphus/evidence/ctp2-systems-task-8-terrain-public-works.txt

  Scenario: Hidden or unaffordable terrain actions fail
    Tool: Go test
    Steps: Target unseen tile, invalid terrain, occupied/duplicate overlay, insufficient public works, and stale revision.
    Expected: Stable rejection without hidden data leakage or partial mutation.
    Evidence: .sisyphus/evidence/ctp2-systems-task-8-terrain-failure.txt
  ```

  **Commit**: YES | Message: `feat(gameplay): add deterministic terrain improvements and pollution` | Files: rules/engine/world/visibility/http tests and implementation

- [ ] 9. Complete diplomacy, trade, governments, agreements, and relationship state

  **What to do**: Move beyond embassy-only diplomacy into deterministic relationship state: contact, war/peace, proposals/responses, treaties, trade routes or agreements, government/civic modifiers if included by the mechanic matrix, and World Union prerequisites. Public commands must validate actor authority and stale revisions; AI may use only public projections/history.
  **Must NOT do**: script AI diplomatic objectives, inject hidden contact, or make World Union a synthetic shortcut disconnected from public relationship state.

  **Recommended Agent Profile**:
  - Category: `deep` - Reason: diplomacy affects terminal conditions, visibility, AI, and multiplayer UX.
  - Skills: [] - No UI.
  - Omitted: [`playwright`] - API tests suffice.

  **Parallelization**: Can Parallel: YES | Wave 2 | Blocks: 11–13 | Blocked By: 4

  **References**:
  - Current diplomacy harness: `internal/e2e/fullgame/diplomacy_scheduler.go`.
  - Embassy special action: `internal/gameplay/engine/special.go`.
  - War command: `internal/gameplay/engine/engine.go` command constants and `applyPolicyCommand`/`applyValidatedMutation`.
  - Victory evaluator: `internal/gameplay/engine/lifecycle.go` `unionWinner`.
  - Existing Row 02/05 diplomacy fixtures in `internal/e2e/fullgame/catalog_rows_01_12.go`.

  **Acceptance Criteria**:
  - [ ] Relationship state is canonical, restart-stable, and viewer-filtered.
  - [ ] Public commands cover contact/proposal/response/trade/peace-war transitions included by the matrix.
  - [ ] World Union/diplomatic terminal conditions depend on durable relationship state.
  - [ ] AI demonstrates basic legal relationship action or legal closeout without hidden directives.
  - [ ] Unauthorized, duplicate, stale, impossible, and terminal diplomacy commands fail predictably.

  **QA Scenarios**:
  ```
  Scenario: Public diplomacy and trade lifecycle
    Tool: Go test + public HTTP/SSE
    Steps: Establish contact, send proposal, accept agreement/trade, change war/peace state, advance rounds, restart, and inspect public projections/events.
    Expected: Relationship/trade state persists, projection redacts hidden counterpart details, and terminal prerequisites are computed from durable state.
    Evidence: .sisyphus/evidence/ctp2-systems-task-9-diplomacy-trade.txt

  Scenario: Invalid diplomacy commands fail safely
    Tool: Go test
    Steps: Wrong actor proposal, duplicate response, stale revision, hidden target, trade without contact, agreement after terminal.
    Expected: Stable error status and no partial relationship mutation.
    Evidence: .sisyphus/evidence/ctp2-systems-task-9-diplomacy-trade-failure.txt
  ```

  **Commit**: YES | Message: `feat(gameplay): model deterministic diplomacy and trade` | Files: engine/service/http/visibility/ai/e2e tests and implementation

- [ ] 10. Complete movement, combat, armies, transports, and special-action interactions across domains

  **What to do**: Integrate expanded unit capabilities with movement and combat: land/sea/amphibious domains, transport capacity/cargo, army capacity/composition, ZOC, embark/disembark, ranged or special combat modifiers if included, deterministic combat RNG, soothsaying, capture, steal technology, and terraform interactions. Ensure public errors never leak hidden defenders or tiles.
  **Must NOT do**: client-supply RNG, retry until desired outcomes, or reveal hidden enemy positions.

  **Recommended Agent Profile**:
  - Category: `deep` - Reason: high interaction density and hidden-state risk.
  - Skills: [] - No UI.
  - Omitted: [`playwright`] - API tests suffice.

  **Parallelization**: Can Parallel: YES | Wave 2 | Blocks: 11–13 | Blocked By: 4

  **References**:
  - Unit engine: `internal/gameplay/engine/units.go`.
  - Combat engine: `internal/gameplay/engine/combat.go`.
  - Special engine: `internal/gameplay/engine/special.go`.
  - Visibility: `internal/gameplay/visibility/visibility.go`.
  - Current fullgame scripts: `internal/e2e/fullgame/scripts/military-zoc-combat.json`, `scripts/special-terraform.json`.

  **Acceptance Criteria**:
  - [ ] Domain/cargo/army legality is enforced from runtime unit definitions.
  - [ ] Combat outcomes are deterministic from canonical RNG and source-cited formulas/simplifications.
  - [ ] ZOC/path blocking and transport interactions work across restart.
  - [ ] Special actions integrate with advances/economy/visibility without hidden leaks.
  - [ ] Public projections expose legal action hints sufficient for a client to avoid avoidable invalid commands.

  **QA Scenarios**:
  ```
  Scenario: Multi-domain combat and transport lifecycle
    Tool: Go test + public HTTP/SSE
    Steps: Load transport with cargo, move across water, disembark, form army, enter ZOC, resolve combat, apply special action, restart.
    Expected: Every state transition is legal, deterministic, projected safely, and restart-stable.
    Evidence: .sisyphus/evidence/ctp2-systems-task-10-combat-transport.txt

  Scenario: Illegal hidden movement/combat rejected
    Tool: Go test
    Steps: Attack unseen defender, overfill transport, mix invalid army domain, move through enemy ZOC illegally, replay RNG-sensitive command with changed payload.
    Expected: Stable rejection or idempotency; no hidden defender/tile data in errors/events.
    Evidence: .sisyphus/evidence/ctp2-systems-task-10-combat-transport-failure.txt
  ```

  **Commit**: YES | Message: `feat(gameplay): complete movement combat and transports` | Files: rules/engine/visibility/http/e2e tests and implementation

- [ ] 11. Teach production AI basic legal use of each included gameplay family

  **What to do**: Expand `ai.PlanWithHistory` and `Driver.Drive` so production AI can choose legal public-projection-based actions for included mechanics: research selection/closeout, production queue, city founding/management, movement/exploration, embassy/diplomacy/trade, terrain improvements/public works, transport/combat when visible, and graceful policy closeout. Score candidates deterministically with stable tie-breakers and bounded history.
  **Must NOT do**: give AI hidden canonical state, row-specific scripts/objectives, direct service submissions outside `ExecuteAICommand`, or competence claims beyond basic legal use.

  **Recommended Agent Profile**:
  - Category: `deep` - Reason: AI touches every mechanic and can destabilize campaigns.
  - Skills: [] - No UI.
  - Omitted: [`playwright`] - AI tests and HTTP provenance suffice.

  **Parallelization**: Can Parallel: NO | Wave 3 | Blocks: 12–14 | Blocked By: 5–10

  **References**:
  - AI planner: `internal/gameplay/ai/ai.go` `PlanWithHistory`, `Driver.Drive`, `Compare`.
  - AI tests: `internal/gameplay/ai/ai_test.go`.
  - Service gateway: `internal/gameplay/service/service.go` `ExecuteAICommand`.
  - Visibility AI input: `internal/gameplay/visibility/visibility.go` `AIInput`.

  **Acceptance Criteria**:
  - [ ] AI has at least one deterministic legal candidate for every included mechanic when public prerequisites are visible.
  - [ ] AI chooses policy closeout instead of illegal repeated actions in simultaneous/hybrid planning.
  - [ ] AI candidate ordering is stable and covered by table tests.
  - [ ] AI submissions are observed only through production gateway provenance.
  - [ ] Hidden-state canary is absent from raw AI input bytes for every tested family.

  **QA Scenarios**:
  ```
  Scenario: AI legally exercises each mechanic family
    Tool: Go test + public observer
    Steps: Feed AI public projections for research, production, diplomacy, movement/combat, public works, and closeout situations; drive through service gateway.
    Expected: AI submits legal commands or zero-command closed status, with deterministic candidate order and observer counts.
    Evidence: .sisyphus/evidence/ctp2-systems-task-11-ai-legal-use.txt

  Scenario: AI cannot use hidden or impossible actions
    Tool: Go test
    Steps: Provide projections lacking visibility/prerequisites/resources and with canary hidden fields in canonical state.
    Expected: AI omits illegal candidates, uses safe closeout/pass fallback, and raw input contains no canary.
    Evidence: .sisyphus/evidence/ctp2-systems-task-11-ai-failure.txt
  ```

  **Commit**: YES | Message: `feat(ai): use gameplay systems from public projections` | Files: ai/visibility/service/e2e tests and implementation

- [ ] 12. Expose client-ready public API, projection, and SSE contracts for included systems

  **What to do**: Ensure clients can discover legal actions and render state for all included systems through authenticated public APIs and SSE only. Add/extend DTOs for unit definitions/capabilities, advances/research, city economy/queues, diplomacy/trade, terrain improvements/pollution, combat summaries, and terminal conditions. Regenerate docs only if public routes or schemas change.
  **Must NOT do**: add debug/test routes to production, expose hidden canonical state, or rely on diagnostics for client-facing data.

  **Recommended Agent Profile**:
  - Category: `unspecified-high` - Reason: API/schema integration and docs consistency.
  - Skills: [] - No UI.
  - Omitted: [`playwright`] - Contract tests use HTTP/SSE.

  **Parallelization**: Can Parallel: NO | Wave 3 | Blocks: 13–14 | Blocked By: 5–11

  **References**:
  - Runtime routes: `internal/runtime/runtime.go`.
  - Gameplay HTTP: `internal/gameplay/http/http.go`.
  - Service: `internal/gameplay/service/service.go`.
  - Events: `internal/gameplay/events/`.
  - Swagger docs: `docs/docs.go`, `docs/swagger.json`, `docs/swagger.yaml`.

  **Acceptance Criteria**:
  - [ ] Public projection includes client-safe render/action metadata for every included family.
  - [ ] SSE emits deterministic filtered events for command success/failure-relevant state transitions.
  - [ ] Unauthorized/nonmember/wrong-turn/stale/terminal API cases return stable status codes and no hidden data.
  - [ ] Generated docs, if changed, are reproducible and match route/schema tests.
  - [ ] Existing onboarding/setup routes still support 2–8 players and arbitrary ordered controllers.

  **QA Scenarios**:
  ```
  Scenario: Client can render and act from public data only
    Tool: Bash curl or Go httptest
    Steps: Start a synthetic multi-system game, fetch projection, derive one legal command per included family from projection fields, submit through HTTP, and consume SSE.
    Expected: Client-visible data is sufficient for legal actions; SSE/projection agree after every command.
    Evidence: .sisyphus/evidence/ctp2-systems-task-12-public-api.txt

  Scenario: Public API redaction and error stability
    Tool: Go httptest
    Steps: Fetch as nonmember, wrong player, hidden-tile viewer, stale revision, and terminal game.
    Expected: Stable status codes, no hidden canary, and no diagnostics/test-only route exposure.
    Evidence: .sisyphus/evidence/ctp2-systems-task-12-public-api-failure.txt
  ```

  **Commit**: YES | Message: `feat(api): expose client-ready gameplay projections` | Files: runtime/http/service/events/visibility/docs/tests

- [ ] 13. Prove cross-system client-ready gameplay scenarios before returning to campaign rows

  **What to do**: Add a small set of deterministic cross-system scenarios that are not Row 09 continuations and are not the 60-row campaign acceptance track. Each scenario must combine imported synthetic content, advances/unlocks, production, movement/combat or diplomacy, public API/SSE, AI legal action, restart, and terminal or durable objective verification. Use these as product-readiness gates before any future catalog-row expansion.
  **Must NOT do**: promote 60-row catalog artifacts, target Row 09 acceptance, or accept scripts that cannot be derived from public projection.

  **Recommended Agent Profile**:
  - Category: `deep` - Reason: true end-to-end product gate across mechanics.
  - Skills: [] - No browser.
  - Omitted: [`playwright`] - Public API/SSE tests are direct.

  **Parallelization**: Can Parallel: NO | Wave 3 | Blocks: 14 | Blocked By: 5–12

  **References**:
  - Public executor: `internal/e2e/fullgame/executor.go`, `internal/e2e/fullgame/acceptance_contract.go`.
  - Scripts/harness patterns: `internal/e2e/fullgame/scripts.go`, `internal/e2e/fullgame/scripts/*.json`.
  - API/onboarding runtime: `internal/e2e/onboarding/`, `internal/runtime/runtime.go`.
  - Accepted-row validation from Task 3.

  **Acceptance Criteria**:
  - [ ] At least three cross-system scenarios pass two clean runs and one restart run each.
  - [ ] One scenario proves advance unlock → production → unit/building use → public projection/SSE.
  - [ ] One scenario proves terrain/public works/pollution interaction with movement/economy.
  - [ ] One scenario proves diplomacy/trade/government or relationship state with AI legal participation.
  - [ ] Scenario success is based on durable objective/terminal state, not row count or budget exhaustion.

  **QA Scenarios**:
  ```
  Scenario: Multi-system deterministic client slice
    Tool: Go test + public HTTP/SSE collector
    Steps: Run the three cross-system scenarios twice clean and once with restart; compare state/event/projection hashes.
    Expected: Deterministic equality and durable objectives pass using only public commands.
    Evidence: .sisyphus/evidence/ctp2-systems-task-13-cross-system.txt

  Scenario: Harness cannot cheat
    Tool: Go test + raw byte scan
    Steps: Inject hidden canaries, attempt direct-store/direct-AI submission, and mutate a scenario to use hidden target IDs.
    Expected: Tests reject hidden/direct paths and canary bytes never appear in public/AI/puppet input.
    Evidence: .sisyphus/evidence/ctp2-systems-task-13-cross-system-failure.txt
  ```

  **Commit**: YES | Message: `test(e2e): prove client-ready gameplay systems slice` | Files: e2e fullgame/onboarding scripts/tests/evidence only unless API gaps require paired implementation

- [ ] 14. Reconcile freeze, CI, evidence, and readiness documentation

  **What to do**: After all mechanic-family and cross-system gates pass, update freeze manifest only for reviewed production-path changes, wire bounded CI commands for the systems slice, regenerate API docs if public schemas changed, and write a readiness report stating exactly what is supported, deferred, and excluded. The report must say that Row 09 was intentionally demoted and that campaign-row acceptance is a later track.
  **Must NOT do**: imply full retail CTP2 parity, hide provisional catalog rows, or update freeze hashes without review evidence.

  **Recommended Agent Profile**:
  - Category: `unspecified-high` - Reason: broad verification/evidence/docs integration.
  - Skills: [`git-master`] - Commit review and intended file staging.
  - Omitted: [`playwright`] - No UI unless added later.

  **Parallelization**: Can Parallel: NO | Wave 4 | Blocks: F1–F4 | Blocked By: 1–13

  **References**:
  - Freeze guard: `internal/e2e/freeze/engine-lock.json`, `internal/e2e/freeze/lock_test.go`.
  - CI: `.github/workflows/coverage.yml`, `Taskfile.yml`.
  - API docs: `docs/docs.go`, `docs/swagger.json`, `docs/swagger.yaml`.
  - Plan/evidence: `.sisyphus/plans/ctp2-gameplay-systems-branch.md`, `.sisyphus/evidence/`.

  **Acceptance Criteria**:
  - [ ] Focused gameplay stack, full build, freeze guard, docs generation check if applicable, protected-data scan, and `git diff --check` exit 0.
  - [ ] Freeze manifest changes, if any, correspond exactly to reviewed production changes from Tasks 5–12.
  - [ ] Readiness report lists included, deferred, and excluded CTP2 systems with test/evidence links.
  - [ ] CI includes bounded systems-slice commands and does not replace existing freeze/fullgame safeguards.
  - [ ] `ValidateCatalogSchema` and accepted-row validation pass; final-all catalog validation remains either passing only if all artifacts were later accepted or fail-closed with provisional rows documented.

  **QA Scenarios**:
  ```
  Scenario: Release-readiness evidence bundle
    Tool: Bash + Go test
    Steps: Run focused stack, build, freeze, protected-data scan, docs check, and accepted-row validation; collect command output with timestamps.
    Expected: All required gates pass and evidence maps every included mechanic to tests.
    Evidence: .sisyphus/evidence/ctp2-systems-task-14-readiness.txt

  Scenario: Freeze/doc/claim drift caught
    Tool: Go test + repository scan
    Steps: Mutate a locked file without hash update in temp worktree, introduce protected fixture text, and add unsupported-system claim to readiness report in table test.
    Expected: Freeze/protected/claim guards fail precisely; real worktree unchanged.
    Evidence: .sisyphus/evidence/ctp2-systems-task-14-readiness-failure.txt
  ```

  **Commit**: YES | Message: `docs(gameplay): document supported systems slice` | Files: freeze/CI/docs/evidence/readiness report and generated docs if required

## Final Verification Wave (MANDATORY — after ALL implementation tasks)
> 4 review agents run in PARALLEL. ALL must APPROVE. Present consolidated results to user and get explicit "okay" before completing.
> **Do NOT auto-proceed after verification. Wait for user's explicit approval before marking work complete.**
> **Never mark F1-F4 as checked before getting user's okay.** Rejection or user feedback -> fix -> re-run -> present again -> wait for okay.
- [ ] F1. Plan Compliance Audit — oracle
- [ ] F2. Code Quality Review — unspecified-high
- [ ] F3. Real Manual QA — unspecified-high (+ API/CLI checks; no Playwright unless UI is added)
- [ ] F4. Scope Fidelity Check — deep

## Commit Strategy
- First commit on new branch: branch-preservation evidence only if repository policy allows committing `.sisyphus/evidence`; otherwise leave evidence uncommitted and record it in final summary.
- Row 09 rollback commit: `test(e2e): demote row nine acceptance on systems branch`.
- Accepted-row validation commit: `test(e2e): track accepted full-game rows explicitly`.
- Mechanic-family commits: one family per commit, implementation and direct tests together.
- Final integration commit: `test(gameplay): prove client-ready systems slice`.
- Do not commit generated docs unless the task changed public API and regeneration is reproducible.

## Success Criteria
- The new branch exists and contains a recoverable copy of the current gameplay workspace.
- Row 09 is provisional again on that branch; accepted Rows 01–07 remain immutable and Row 08 remains a known blocker rather than hidden.
- Progress is measured by completed gameplay-system gates, not by catalog row acceptance count.
- Core client-visible gameplay systems have deterministic rules import/runtime, engine behavior, public HTTP/SSE/projection coverage, AI legality, restart equality, and failure-mode tests.
- Final reviewers approve compliance, quality, real QA, and scope fidelity.
