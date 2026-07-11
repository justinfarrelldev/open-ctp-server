# Game Server World

## TL;DR
> **Summary**: Implement the README `World (/world)` section with CTP2-first deterministic world generation, persisted world state, world read/rating/pollution endpoints, turn recalculation, Swagger docs, README checklist sync, and rate limiting. Work must happen on a new `game-server-world` branch based on the current local `main`.
> **Deliverables**:
> - New `internal/world` package with pure deterministic generation/recalculation engine and tests.
> - World persistence/migrations for seed, generator version, host parameters, static map snapshot, mutable turn state, ratings, pollution, and player spawns.
> - Game start integration that generates/persists world once inside one transaction.
> - `/world` API surface with documented read endpoints, ratings, pollution, and space-layer 404 placeholder.
> - Swagger docs and README World bullets updated only after corresponding behavior is test-green.
> - Atomic commits and reviewer waves after each big item.
> **Effort**: XL
> **Parallel**: YES - 5 waves
> **Critical Path**: Branch preflight → endpoint/data contracts → generator fixtures → pure generation engine → persistence migration → start-game integration → endpoints/docs → turn recalculation → final verification

## Context
### Original Request
Implement items in the README `World (/world)` section using TDD red-green-refactor. Base work on local `main`, create and work on branch `game-server-world`, read top-level `.md` files first, keep README bullets and Swagger docs in sync, commit atomically after each finished feature, use reviewer waves after each big item, target Call to Power 2 first using `ctp2`/released source as reference, and keep world logic close to original/Spyroviper Edition while avoiding crash-causing bugs. User additionally clarified: world generation must use all host-provided setup parameters, including wetness.

### Interview Summary
- User confirmed local `main` is the intended base even though it is `main...origin/main [ahead 23]`.
- TDD with RED-GREEN-REFACTOR is mandatory.
- World generation must consume all host parameters, not only seed/size.
- Scope is the whole README World section in one plan; implementation still proceeds in atomic vertical slices.

### Metis Review (gaps addressed)
- Added contract-first task before coding to prevent endpoint semantics being invented mid-implementation.
- Locked defaults for vague README items: ratings/pollution are read-only derived summaries; recalculation triggers on game start and explicit turn-advance/recalc service calls only; impossible spawn placement fails deterministically instead of degrading silently; `/world/space` is a registered documented 404 with fixed JSON error; active-game chat via lobby is a verification/docs non-goal unless tests reveal missing behavior.
- Added explicit crash guards, concurrency serialization tests, README/Swagger sync gates, and reviewer waves after big items.

## Work Objectives
### Core Objective
Build an authoritative CTP2-first server-side world subsystem that deterministically generates and persists game worlds from host setup parameters, exposes documented `/world` endpoints, supports deterministic turn recalculation for documented world state, and updates project docs/checklists only when behavior is verified.

### Deliverables
- Branch `game-server-world` based on current local `main`.
- `internal/world` package with pure functions only for generation, spawn placement, ratings, pollution, and turn recalculation.
- Persistence schema/migrations for world snapshots and mutable turn state.
- Game start integration that generates the world exactly once and atomically transitions the game to started.
- `/world` routes registered in `main.go`, rate limited, documented in Swagger, and tested with `httptest`/`sqlmock`.
- README `World` checklist updated as bullets become complete.
- Atomic commits per vertical slice.

### Definition of Done (verifiable conditions with commands)
- `git branch --show-current` prints `game-server-world`.
- `task test` passes.
- `task test-coverage` passes and internal coverage remains `>80%`.
- `go test ./internal/world/...` passes.
- `go test ./internal/game/...` passes.
- `go test ./internal/... -run 'Test(StartGame|Recalc).*Serialization|TestWorld'` passes.
- Swagger output includes `/world`, ratings, pollution, and `/world/space` placeholder paths.
- README `World (/world)` bullets are checked only for features with passing tests and docs.
- Reviewer waves R1-R5 and final verification F1-F4 all approve before completion.

### Must Have
- TDD per task: write failing test first, implement minimum green path, refactor, then docs/README sync only if the slice completes a README bullet.
- Use `internal/game` patterns for HTTP handlers, helpers, sqlmock tests, auth/session checks, setup/started-state guards, and Swagger comments.
- Use `internal/world` for generator/domain logic with no DB/HTTP/global RNG/package-global mutable state.
- Persist seed, generator version, canonical player order, all host generation parameters including wetness, immutable static map, mutable turn state, pollution, ratings, and spawns.
- Deterministic RNG contract: same seed + same host params + same player order + same generator version produces byte-identical canonical world output.
- Crash guards: validate parameters, cap retries, bound-check coordinates, assert legal starts, serialize concurrent starts/recalcs, and fail explicitly.
- Runtime must not parse external CTP2 source/data files on request path.

### Must NOT Have (guardrails, AI slop patterns, scope boundaries)
- No AI turns, fog-of-war implementation, graphics/sound/client assets, map editor/import tools, admin/debug endpoints, CTP1 space mechanics beyond 404 placeholder, or undocumented balancing changes.
- No bug-for-bug crash parity; preserve behavior where safe, but replace unbounded retry/crash paths with deterministic errors.
- No new dependencies unless a task first proves stdlib/current repo patterns cannot satisfy the need.
- No README checkbox updates before tests and Swagger docs pass for that exact bullet.
- No broad refactors outside `main.go`, `internal/game`, `internal/world`, `docs`, migrations, tests, and README unless required by a failing test.

## Endpoint/Data Contract Defaults
> Executor must encode these contracts in tests before implementation. If existing code shape forces a different field name, update tests and Swagger consistently in the same task; do not silently change semantics.

| Capability | Contract |
|---|---|
| World generation | Triggered once by game start; request uses existing game setup plus persisted host world params. Response from start may remain existing shape but world state must be persisted. |
| World read | `GET /world/{game_id}` returns authenticated player's permitted generated world summary/current full server-visible world for now; future fog-of-war is explicitly out of scope. |
| Ratings | `GET /world/{game_id}/ratings` returns read-only derived categories: `overall`, `science`, `military`, `economic`, `pollution`, each deterministic from persisted world/player state. Values may be zero/default until future gameplay changes state, but schema must exist and be documented. |
| Pollution | `GET /world/{game_id}/pollution` returns read-only summary: `game_id`, `turn`, `global_pollution`, `tile_pollution_count`, `updated_at` or deterministic equivalent from persisted turn state. No mutation on GET. |
| Recalculation | Service-level function runs on game start and future turn-advance integration; for this World plan, expose no public mutation endpoint unless existing gameplay turn advancement requires it. Tests call service directly and integration point from game start. |
| Space layer | `GET /world/{game_id}/space` is registered, rate-limited, documented, and returns `404` with fixed JSON `{ "error": "space layer is not supported for CTP2 worlds" }`. |
| Active-game chat | No new `/game/chat` or `/world/chat` endpoint. Verify existing lobby chat docs/tests remain source of truth; README bullet can be checked only if an automated test or documentation assertion proves active games still use lobby chat. |
| Spawn failure | If requested parameters cannot produce legal non-overlapping starts within capped retries, game start fails deterministically with a documented error and no partial world/game-start persistence. |
| Invalid params | Invalid host world params return `400` or cause setup validation failure before generation; no panic, nil deref, or infinite loop. |

## Verification Strategy
> ZERO HUMAN INTERVENTION - all verification is agent-executed.
- Test decision: TDD RED-GREEN-REFACTOR using Go standard testing, `httptest`, `sqlmock`, and existing `task test`/`task test-coverage` commands.
- QA policy: Every task has agent-executed scenarios.
- Evidence: `.sisyphus/evidence/task-{N}-{slug}.{ext}`.
- Reviewer waves after each big item: contract/model, generation/persistence, start/spawn, endpoints/docs, turn recalculation/docs.

## Execution Strategy
### Parallel Execution Waves
> Target: 5-8 tasks per wave. <3 per wave (except final) = under-splitting.
> Extract shared dependencies as Wave-1 tasks for max parallelism.

Wave 1: Branch preflight, contract/model design, CTP2 fixture discovery, test harness scaffolding, host-parameter inventory.
Wave 2: Pure generation engine, spawn placement, crash guards, persistence migration/model.
Wave 3: Start-game integration, serialization, ratings/pollution calculation services, lobby-chat verification.
Wave 4: `/world` handlers, rate limits, Swagger docs, README sync, turn recalculation service.
Wave 5: Reviewer waves, coverage hardening, final verification.

### Dependency Matrix (full, all tasks)
- T1 blocks all implementation tasks.
- T2 blocks T5-T13.
- T3 blocks T5-T7 and T11.
- T4 blocks T5-T13.
- T5 blocks T6, T8, T10, T11, T13.
- T6 blocks T8 and T10.
- T7 blocks T8, T10, T11, T13.
- T8 blocks T9, T10, T11, T12, T14.
- T9 blocks T14.
- T10 blocks T14.
- T11 blocks T14.
- T12 blocks T14.
- T13 blocks T14.
- T14 blocks final verification.

### Agent Dispatch Summary (wave → task count → categories)
- Wave 1 → 4 tasks → git-master/quick, deep, quick, deep.
- Wave 2 → 4 tasks → deep, deep, deep, unspecified-high.
- Wave 3 → 4 tasks → deep, deep, quick, unspecified-high.
- Wave 4 → 2 tasks → unspecified-high, writing.
- Wave 5 → reviewer/final verification agents.

## TODOs
> Implementation + Test = ONE task. Never separate.
> EVERY task MUST have: Agent Profile + Parallelization + QA Scenarios.

- [x] T1. Branch Preflight and Worktree Safety

  **What to do**: Use git-master workflow. Confirm worktree state, confirm current branch is local `main`, create/switch to `game-server-world` from local `main`, and record preflight evidence. Do not rebase, reset, pull, push, or discard local commits. If `game-server-world` already exists, verify it points at/descends from local `main` before continuing.
  **Must NOT do**: Do not base on `origin/main`; do not alter files except evidence if needed; do not commit unrelated existing changes.

  **Recommended Agent Profile**:
  - Category: `quick` - Reason: bounded git preflight.
  - Skills: [`git-master`] - Required for branch creation and atomic commit discipline.
  - Omitted: [`frontend-ui-ux`] - No UI work.

  **Parallelization**: Can Parallel: NO | Wave 1 | Blocks: T2-T14 | Blocked By: none

  **References**:
  - Pattern: `CONTRIBUTING.md` - PR/test/format expectations.
  - Pattern: current observed state `main...origin/main [ahead 23]` - local `main` is confirmed branch base.
  - Pattern: repository branch naming - hyphenated feature branches.

  **Acceptance Criteria**:
  - [ ] `git branch --show-current` returns `game-server-world`.
  - [ ] `git merge-base --is-ancestor main HEAD` succeeds.
  - [ ] `git status --short --branch` shows branch `game-server-world` and no unexpected tracked changes from this task.

  **QA Scenarios**:
  ```
  Scenario: Correct branch base
    Tool: Bash
    Steps: Run `git branch --show-current && git merge-base --is-ancestor main HEAD && git status --short --branch`.
    Expected: branch is `game-server-world`; merge-base command exits 0; status has no unexpected changes.
    Evidence: .sisyphus/evidence/task-1-branch-preflight.txt

  Scenario: Existing branch safety
    Tool: Bash
    Steps: If branch already existed, run `git log --oneline --decorate --max-count=5` and `git merge-base --is-ancestor main HEAD`.
    Expected: no reset/rebase performed; branch descends from local main.
    Evidence: .sisyphus/evidence/task-1-existing-branch.txt
  ```

  **Commit**: NO | Message: n/a | Files: []

- [x] T2. Contract-First World API and Data Model Tests

  **What to do**: Before implementation, write failing tests and docs-only contract fixtures for world data/endpoint semantics. Define Go types for request/response contracts only as needed to make tests compile in red state. Contract must include generation inputs, all host params including wetness, world read response, ratings response, pollution response, space-layer 404 response, invalid params, duplicate start, missing world, and non-active/ended game behavior.
  **Must NOT do**: Do not implement generator logic, DB persistence, or handlers beyond compile stubs required for red tests.

  **Recommended Agent Profile**:
  - Category: `deep` - Reason: locks cross-cutting contracts and prevents semantic drift.
  - Skills: [] - Existing Go test patterns are enough.
  - Omitted: [`playwright`] - API/unit tests only.

  **Parallelization**: Can Parallel: NO | Wave 1 | Blocks: T5-T13 | Blocked By: T1

  **References**:
  - Pattern: `internal/game/create_game.go` - game setup defaults and CTP2 source breadcrumb style.
  - Pattern: `internal/game/swagger_models.go` - docs-only request payload patterns.
  - Pattern: `internal/game/endpoints.go` - Swagger annotation style.
  - Test: `internal/game/create_game_test.go` - handler/sqlmock style.
  - Test: `internal/ratelimit/ratelimit_test.go` - table-driven and parallel test style.

  **Acceptance Criteria**:
  - [ ] New red tests fail for missing world contracts with explicit expected schemas/status codes.
  - [ ] Tests assert wetness and every existing host generation parameter are present in normalized generation input.
  - [ ] Tests assert `/world/{game_id}/space` returns exact 404 JSON error.
  - [ ] Tests document that chat remains via lobby and no `/world/chat` or `/game/chat` route is added.

  **QA Scenarios**:
  ```
  Scenario: Contract tests are red for missing world implementation
    Tool: Bash
    Steps: Run `go test ./internal/... -run 'TestWorld.*Contract|TestGenerationConsumesAllHostParams|TestWorldSpaceLayerPlaceholder'` immediately after writing tests.
    Expected: tests fail because implementation/routes are missing, with failures matching missing world behavior rather than syntax errors.
    Evidence: .sisyphus/evidence/task-2-contract-red.txt

  Scenario: Host params contract includes wetness
    Tool: Bash
    Steps: Run `go test ./internal/world/... -run TestGenerationConsumesAllHostParams` after red tests are added.
    Expected: failure explicitly indicates wetness/host params are not yet consumed.
    Evidence: .sisyphus/evidence/task-2-host-params-red.txt
  ```

  **Commit**: YES | Message: `test(world): lock world api and data contracts` | Files: [`internal/world/*_test.go`, `internal/game/*world*_test.go`, optional contract stubs]

- [x] T3. CTP2 Reference Fixture and Parameter Inventory

  **What to do**: Inventory host-provided setup parameters from existing game setup models and CTP2 references. Create deterministic golden fixture definitions for at least three seeds: small/default map, wet map, dry map. Fixture metadata must cite source references and include expected invariants; exact tile/spawn outputs are required where feasible, and otherwise the task must generate reproducible canonical fixture data from the CTP2 reference path before implementation proceeds.
  **Must NOT do**: Do not parse external CTP2 files at runtime; do not copy crash-prone unbounded loops; do not mark parity README bullets complete from invariants alone if exact seed parity is still absent.

  **Recommended Agent Profile**:
  - Category: `deep` - Reason: source-reference and parity mapping.
  - Skills: [] - Read-only source mapping plus Go fixtures.
  - Omitted: [`git-master`] - No git operation beyond normal commit.

  **Parallelization**: Can Parallel: YES | Wave 1 | Blocks: T5-T7, T11 | Blocked By: T1, T2

  **References**:
  - External: `/home/justi/Projects/ctp2-src/ctp2_code/gs/world/wldgen.cpp` - world generation algorithms and start placement.
  - External: `/home/justi/Projects/ctp2-src/ctp2_code/gs/utility/gameinit.cpp` - world creation/bootstrap and initial unit placement.
  - External: `/home/justi/Projects/ctp2-src/ctp2_code/gs/world/WrldCont.cpp` - continent/chokepoint logic.
  - Pattern: `internal/game/create_game.go` - existing host setup/default model references.
  - README: `README.md` World section - parity and spawn bullets.

  **Acceptance Criteria**:
  - [ ] Fixture file records seed, generator version, map params, wetness, player order, expected invariants, and source references.
  - [ ] Tests fail red until generator consumes these fixtures.
  - [ ] Changing wetness in fixture changes normalized config and expected generation path.
  - [ ] Fixture docs state whether exact Spyroviper/CTP2 output is available or generated during implementation.

  **QA Scenarios**:
  ```
  Scenario: Fixture inventory is complete
    Tool: Bash
    Steps: Run `go test ./internal/world/... -run TestFixtureInventoryIncludesAllHostParams`.
    Expected: test fails before generator implementation but passes fixture completeness checks for seed, wetness, player order, and generator version.
    Evidence: .sisyphus/evidence/task-3-fixture-inventory.txt

  Scenario: No runtime CTP2 file dependency
    Tool: Bash
    Steps: Search new world code/tests for absolute CTP2 source reads on request/generation path.
    Expected: references exist only in comments/fixtures/test derivation notes, not production file I/O.
    Evidence: .sisyphus/evidence/task-3-no-runtime-ctp2-parse.txt
  ```

  **Commit**: YES | Message: `test(world): add ctp2 generation fixtures` | Files: [`internal/world/fixtures/*`, `internal/world/*fixture*_test.go`]

- [x] T4. World Package Skeleton and Test Harness

  **What to do**: Add `internal/world` package structure with compile stubs, canonical serialization helper, deterministic RNG wrapper interface/value, normalized config type, world snapshot type, mutable turn state type, and test helpers. Keep generator functions pure: explicit inputs, explicit outputs, no DB/HTTP/global mutable state.
  **Must NOT do**: Do not implement actual generation yet; do not add dependencies; do not wire routes.

  **Recommended Agent Profile**:
  - Category: `quick` - Reason: scaffolding and compile harness.
  - Skills: [] - Standard Go only.
  - Omitted: [`playwright`] - No UI.

  **Parallelization**: Can Parallel: YES | Wave 1 | Blocks: T5-T13 | Blocked By: T1, T2

  **References**:
  - Pattern: `internal/game/models.go` - model organization.
  - Pattern: `internal/game/helpers.go` - package helper organization; follow style but do not import DB/HTTP into generator.
  - Test: `internal/auth/auth_test.go` - core unit test style.

  **Acceptance Criteria**:
  - [ ] `go test ./internal/world/...` compiles and fails only on intentional red behavior tests.
  - [ ] Package has no imports from `net/http`, `database/sql`, `github.com/jmoiron/sqlx`, or route packages.
  - [ ] Canonical serialization produces stable bytes for identical fixtures.

  **QA Scenarios**:
  ```
  Scenario: Pure package boundary
    Tool: Bash
    Steps: Run `go test ./internal/world/...` and inspect imports in `internal/world`.
    Expected: tests compile; package has no DB/HTTP imports.
    Evidence: .sisyphus/evidence/task-4-package-boundary.txt

  Scenario: Canonical serialization stable
    Tool: Bash
    Steps: Run `go test ./internal/world/... -run TestCanonicalWorldSerializationStable`.
    Expected: identical snapshot serializes identically across repeated calls.
    Evidence: .sisyphus/evidence/task-4-serialization.txt
  ```

  **Commit**: YES | Message: `feat(world): add pure world package harness` | Files: [`internal/world/*`]

- [x] T5. Deterministic CTP2-First World Generation Engine

  **What to do**: Implement pure generation logic against T2/T3 tests. Consume normalized seed, generator version, canonical player order, and every host generation parameter including wetness. Use CTP2 source as behavioral reference while replacing crash-prone unbounded retries with capped deterministic attempts and explicit errors. Produce static map snapshot and initial mutable turn state.
  **Must NOT do**: Do not persist to DB, register routes, or mutate game status in this task.

  **Recommended Agent Profile**:
  - Category: `deep` - Reason: core deterministic algorithm with CTP2 reference behavior.
  - Skills: [] - Standard Go; no dependencies.
  - Omitted: [`frontend-ui-ux`] - No UI.

  **Parallelization**: Can Parallel: NO | Wave 2 | Blocks: T6, T8, T10, T11, T13 | Blocked By: T2, T3, T4

  **References**:
  - External: `/home/justi/Projects/ctp2-src/ctp2_code/gs/world/wldgen.cpp` - map generation and terrain/wetness behavior.
  - External: `/home/justi/Projects/ctp2-src/ctp2_code/gs/world/WrldCont.cpp` - continent/chokepoint behavior.
  - Fixture: `internal/world/fixtures/*` - deterministic seeds and expected outputs/invariants.

  **Acceptance Criteria**:
  - [ ] `TestWorldGenerationDeterministic` passes: same seed/params/player order produces identical canonical output.
  - [ ] `TestGenerationConsumesAllHostParams` passes and explicitly covers wetness.
  - [ ] `TestWorldGenerationRejectsInvalidParams` passes without panic.
  - [ ] `TestWorldGenerationCapsRetries` passes for impossible maps.

  **QA Scenarios**:
  ```
  Scenario: Deterministic generation
    Tool: Bash
    Steps: Run `go test ./internal/world/... -run TestWorldGenerationDeterministic -count=20`.
    Expected: all 20 runs pass with byte-identical canonical snapshots.
    Evidence: .sisyphus/evidence/task-5-deterministic-generation.txt

  Scenario: Invalid host params fail safely
    Tool: Bash
    Steps: Run `go test ./internal/world/... -run 'TestWorldGenerationRejectsInvalidParams|TestWorldGenerationCapsRetries'`.
    Expected: invalid/impossible inputs return deterministic errors; no panic/hang.
    Evidence: .sisyphus/evidence/task-5-crash-guards.txt
  ```

  **Commit**: YES | Message: `feat(world): add deterministic ctp2 generation engine` | Files: [`internal/world/*`, `internal/world/fixtures/*`]

- [x] T6. Deterministic Player Spawn Placement

  **What to do**: Implement spawn placement as a pure phase of generation. Respect canonical player order, CTP2 reference constraints, non-overlap, valid terrain, map bounds/wrapping rules, host params, and capped failure handling. Persistable spawn output must be stable and independently testable.
  **Must NOT do**: Do not silently degrade to invalid/overlapping spawns; do not randomize based on map iteration order.

  **Recommended Agent Profile**:
  - Category: `deep` - Reason: parity-sensitive placement and failure paths.
  - Skills: [] - Pure Go tests.
  - Omitted: [`playwright`] - No UI.

  **Parallelization**: Can Parallel: NO | Wave 2 | Blocks: T8, T10 | Blocked By: T5

  **References**:
  - External: `/home/justi/Projects/ctp2-src/ctp2_code/gs/world/wldgen.cpp` - player start routines.
  - External: `/home/justi/Projects/ctp2-src/ctp2_code/gs/utility/gameinit.cpp` - bootstrap and initial unit placement context.
  - README: `README.md` World bullet: players spawn in same locations as Spyroviper Edition.

  **Acceptance Criteria**:
  - [ ] `TestPlayerSpawnsAreValid` passes for all fixtures.
  - [ ] `TestPlayerSpawnsDeterministic` passes with `-count=20`.
  - [ ] `TestStartPlacementFailsWhenNoValidSpawn` passes and asserts exact error.
  - [ ] Exact seed parity tests pass for spawn fixtures where available; otherwise README spawn bullet remains unchecked.

  **QA Scenarios**:
  ```
  Scenario: Valid deterministic spawns
    Tool: Bash
    Steps: Run `go test ./internal/world/... -run 'TestPlayerSpawns(AreValid|Deterministic)' -count=20`.
    Expected: every player has exactly one non-overlapping in-bounds valid spawn; output stable across runs.
    Evidence: .sisyphus/evidence/task-6-spawns.txt

  Scenario: Spawn failure has no crash fallback
    Tool: Bash
    Steps: Run `go test ./internal/world/... -run TestStartPlacementFailsWhenNoValidSpawn`.
    Expected: deterministic typed error; no panic and no partial spawn list.
    Evidence: .sisyphus/evidence/task-6-spawn-failure.txt
  ```

  **Commit**: YES | Message: `feat(world): place deterministic player spawns` | Files: [`internal/world/*`, `internal/world/fixtures/*`]

- [x] T7. World Persistence Schema and Store

  **What to do**: Add migrations/store code for authoritative world state. Persist static world snapshot separately from mutable turn state; store seed, generator version, host params JSON, player order, spawns, ratings/pollution current state, and timestamps. Add rollback/clean-state considerations consistent with current Supabase/Postgres migration style. Store methods must be transaction-friendly and sqlmock-tested.
  **Must NOT do**: Do not prematurely normalize every tile into excessive future tables unless needed by tests; do not expose DB rows directly as API DTOs.

  **Recommended Agent Profile**:
  - Category: `unspecified-high` - Reason: DB schema, transactions, tests.
  - Skills: [] - Existing sqlmock patterns.
  - Omitted: [`frontend-ui-ux`] - No UI.

  **Parallelization**: Can Parallel: YES | Wave 2 | Blocks: T8, T10, T11, T13 | Blocked By: T2, T3, T4

  **References**:
  - Pattern: Supabase migration files under `supabase/` - schema/migration style.
  - Pattern: `internal/game/models.go` - game/player persistence models.
  - Test: `internal/game/create_game_test.go` - sqlmock DB expectation style.
  - Architecture: `ARCHITECTURE.md` - Supabase/Postgres.

  **Acceptance Criteria**:
  - [ ] Migration creates required world state structures and is idempotent in local test flow.
  - [ ] Store tests cover create, read, duplicate world prevention, missing world, and transaction rollback on error.
  - [ ] Static snapshot is immutable after creation; mutable state updates use separate store path.
  - [ ] Host params including wetness are persisted and recoverable exactly.

  **QA Scenarios**:
  ```
  Scenario: Store persists and reloads canonical world
    Tool: Bash
    Steps: Run `go test ./internal/world/... -run TestWorldStoreCreateAndRead`.
    Expected: seed, generator version, wetness, player order, static snapshot, mutable turn state, and spawns round-trip exactly.
    Evidence: .sisyphus/evidence/task-7-store-roundtrip.txt

  Scenario: Store rollback prevents partial world
    Tool: Bash
    Steps: Run `go test ./internal/world/... -run TestWorldStoreRollbackOnCreateFailure`.
    Expected: injected DB error rolls back; follow-up read reports missing world.
    Evidence: .sisyphus/evidence/task-7-store-rollback.txt
  ```

  **Commit**: YES | Message: `feat(world): persist world snapshots and turn state` | Files: [`supabase/*`, `internal/world/*store*`, `internal/world/*store*_test.go`]

- [x] T8. Start Game Integration and Serialization

  **What to do**: Integrate world generation/persistence into the existing game start flow. Use LSP/reference search before edits. Game start must load setup/host params, normalize generation input, generate world/spawns, persist world, and transition game status to started inside one transaction. Serialize duplicate/concurrent starts for the same game; if generation/spawn/persistence fails, leave game unstarted and no partial world state.
  **Must NOT do**: Do not regenerate world on reads; do not let duplicate start create duplicate world rows; do not change unrelated game setup behavior.

  **Recommended Agent Profile**:
  - Category: `deep` - Reason: lifecycle integration and transaction/concurrency safety.
  - Skills: [] - Existing Go/sqlmock patterns.
  - Omitted: [`playwright`] - HTTP tests via httptest.

  **Parallelization**: Can Parallel: NO | Wave 3 | Blocks: T9, T10, T11, T12, T14 | Blocked By: T5, T6, T7

  **References**:
  - Pattern: `internal/game/player_endpoints.go` - start-game/player lifecycle flow.
  - Pattern: `internal/game/helpers.go` - setup-state/auth/DB helpers.
  - Pattern: `main.go` - route registration and rate limiter style.
  - Test: `internal/game/create_game_test.go` - sqlmock handler style.

  **Acceptance Criteria**:
  - [ ] `TestStartGameGeneratesWorldOnce` passes.
  - [ ] `TestStartGamePersistsWorldInTransaction` passes.
  - [ ] `TestStartGameDoesNotStartOnWorldGenerationFailure` passes.
  - [ ] `TestStartGameSerialization` passes under concurrent calls.

  **QA Scenarios**:
  ```
  Scenario: Start generates exactly once
    Tool: Bash
    Steps: Run `go test ./internal/game/... -run TestStartGameGeneratesWorldOnce`.
    Expected: first start persists one world and starts game; duplicate start does not regenerate.
    Evidence: .sisyphus/evidence/task-8-start-once.txt

  Scenario: Concurrent start serialization
    Tool: Bash
    Steps: Run `go test ./internal/... -run TestStartGameSerialization -count=20`.
    Expected: no duplicate world state; one success or deterministic already-started response; no race-sensitive failures.
    Evidence: .sisyphus/evidence/task-8-start-serialization.txt
  ```

  **Commit**: YES | Message: `feat(world): generate world during game start` | Files: [`internal/game/*`, `internal/world/*`, tests]

- [x] T9. Lobby Chat Verification for Active Games

  **What to do**: Prove the README World chat bullet without adding duplicate chat endpoints. Add/adjust tests or documentation assertions showing active-game chat continues through existing lobby chat streams/endpoints and that no `/game/chat` or `/world/chat` route exists. Update README checkbox only if the automated evidence proves this behavior is already correct or after minimal code required to preserve it.
  **Must NOT do**: Do not create new chat endpoints; do not alter lobby chat behavior unless a failing test proves active-game state blocks it incorrectly.

  **Recommended Agent Profile**:
  - Category: `quick` - Reason: verification/non-goal enforcement.
  - Skills: [] - Existing lobby tests.
  - Omitted: [`frontend-ui-ux`] - No UI.

  **Parallelization**: Can Parallel: YES | Wave 3 | Blocks: T14 | Blocked By: T8

  **References**:
  - Pattern: `internal/lobby/stream_task7_test.go` - lobby chat/stream test style.
  - README: `README.md` World bullet: chats in active games via lobby.
  - Pattern: `main.go` - registered route surface.

  **Acceptance Criteria**:
  - [ ] Test confirms active-game chat uses lobby flow or existing docs/route behavior supports it.
  - [ ] Test confirms no `/game/chat` or `/world/chat` route is registered.
  - [ ] README chat bullet is checked only if automated test passes.

  **QA Scenarios**:
  ```
  Scenario: Active-game chat remains lobby-based
    Tool: Bash
    Steps: Run `go test ./internal/lobby/... ./internal/game/... -run TestActiveGameChatUsesLobby`.
    Expected: lobby chat works for active game context or documented route behavior; no game/world chat endpoint required.
    Evidence: .sisyphus/evidence/task-9-lobby-chat.txt

  Scenario: No duplicate chat route
    Tool: Bash
    Steps: Run route-surface test for `/game/chat` and `/world/chat`.
    Expected: both are absent/404 while lobby chat route remains tested.
    Evidence: .sisyphus/evidence/task-9-no-duplicate-chat.txt
  ```

  **Commit**: YES | Message: `test(world): verify active game chat uses lobby` | Files: [`internal/lobby/*_test.go`, `internal/game/*_test.go`, `README.md` if proven]

- [x] T10. World Read and Space Placeholder Endpoints

  **What to do**: Add `/world/{game_id}` and `/world/{game_id}/space` handlers, route registration in `main.go`, auth/session checks, game membership checks, started-state/missing-world handling, standard rate limiting, DTO mapping from persisted state, and tests. `/world/{game_id}/space` must return fixed 404 JSON for CTP2.
  **Must NOT do**: Do not expose raw DB rows; do not implement CTP1 space mechanics; do not regenerate world on GET.

  **Recommended Agent Profile**:
  - Category: `unspecified-high` - Reason: handler routing, auth, docs, DB reads.
  - Skills: [] - Existing httptest/sqlmock patterns.
  - Omitted: [`playwright`] - HTTP unit tests sufficient.

  **Parallelization**: Can Parallel: NO | Wave 4 | Blocks: T14 | Blocked By: T5, T6, T7, T8

  **References**:
  - Pattern: `main.go` - route and rate limiter registration.
  - Pattern: `internal/game/endpoints.go` - handler and Swagger annotation style.
  - Pattern: `internal/game/helpers.go` - auth, setup/state, JSON response helpers.
  - Test: `internal/health/health_check_test.go` - minimal httptest example.

  **Acceptance Criteria**:
  - [ ] `TestGetWorldReturnsGeneratedState` passes with `200` and exact JSON fields.
  - [ ] `TestGetWorldRequiresAuthAndMembership` passes.
  - [ ] `TestGetWorldMissingWorld` passes with exact error.
  - [ ] `TestWorldSpaceLayerPlaceholder` passes with `404` and `{ "error": "space layer is not supported for CTP2 worlds" }`.
  - [ ] World read and space routes are rate limited consistently with existing game/lobby route policy.

  **QA Scenarios**:
  ```
  Scenario: Read generated world
    Tool: Bash
    Steps: Run `go test ./internal/game/... -run TestGetWorldReturnsGeneratedState`.
    Expected: authenticated member receives persisted generated state; no generation call occurs.
    Evidence: .sisyphus/evidence/task-10-world-read.txt

  Scenario: Space placeholder
    Tool: Bash
    Steps: Run `go test ./internal/game/... -run TestWorldSpaceLayerPlaceholder`.
    Expected: route exists, returns 404 and exact JSON error, and appears in Swagger after docs task.
    Evidence: .sisyphus/evidence/task-10-space-placeholder.txt
  ```

  **Commit**: YES | Message: `feat(world): expose generated world endpoints` | Files: [`main.go`, `internal/game/*world*`, `internal/world/*`, tests]

- [x] T11. Ratings and Pollution Services and Endpoints

  **What to do**: Implement read-only deterministic ratings and pollution calculations from persisted world/player state. Add `/world/{game_id}/ratings` and `/world/{game_id}/pollution` handlers, DTOs, tests, and rate limiting. Ratings default categories are `overall`, `science`, `military`, `economic`, `pollution`; values may be deterministic zeros/defaults until later gameplay mutates state, but schema must be stable. Pollution response includes game id, turn, global pollution, tile pollution count, and deterministic update marker/time field consistent with repo style.
  **Must NOT do**: Do not mutate state on GET; do not invent full future simulation beyond README-visible summaries; do not expose per-tile pollution unless tests/contracts require it.

  **Recommended Agent Profile**:
  - Category: `deep` - Reason: derived state semantics and API tests.
  - Skills: [] - Standard Go tests.
  - Omitted: [`frontend-ui-ux`] - No UI.

  **Parallelization**: Can Parallel: YES | Wave 4 | Blocks: T14 | Blocked By: T5, T7, T8

  **References**:
  - External: `/home/justi/Projects/ctp2-src/ctp2_code/gs/world/WrldPoll.cpp` - pollution behavior reference.
  - External: `/home/justi/Projects/ctp2-src/ctp2_code/gs/gameobj/PlayerTurn.cpp` - turn-start pollution handling.
  - Pattern: `internal/game/endpoints.go` - endpoint style.
  - README: `README.md` World bullets for ratings and world pollution.

  **Acceptance Criteria**:
  - [ ] `TestWorldRatingsEndpoint` passes with exact status codes and JSON fields.
  - [ ] `TestWorldPollutionEndpoint` passes with exact status codes and JSON fields.
  - [ ] `TestWorldRatingsDeterministic` and `TestWorldPollutionDeterministic` pass.
  - [ ] GET requests do not change persisted world state.

  **QA Scenarios**:
  ```
  Scenario: Ratings endpoint returns stable schema
    Tool: Bash
    Steps: Run `go test ./internal/game/... -run TestWorldRatingsEndpoint`.
    Expected: response contains `overall`, `science`, `military`, `economic`, `pollution`; unauthorized/missing cases return documented errors.
    Evidence: .sisyphus/evidence/task-11-ratings.txt

  Scenario: Pollution endpoint is read-only
    Tool: Bash
    Steps: Run `go test ./internal/game/... -run TestWorldPollutionEndpoint` and `go test ./internal/world/... -run TestWorldPollutionReadOnly`.
    Expected: response matches contract and persisted state checksum is unchanged by GET.
    Evidence: .sisyphus/evidence/task-11-pollution.txt
  ```

  **Commit**: YES | Message: `feat(world): add ratings and pollution endpoints` | Files: [`main.go`, `internal/game/*world*`, `internal/world/*ratings*`, `internal/world/*pollution*`, tests]

- [x] T12. Turn Recalculation Service

  **What to do**: Implement deterministic world turn recalculation service for documented world fields: pollution, events placeholder state, city production placeholders only where existing data supports them, and mutable turn state updates. For now, trigger from game start initialization and provide service function for future turn-advance integration; do not add a public mutation endpoint unless existing gameplay turn advancement already has a route to call. Static map snapshot must remain unchanged.
  **Must NOT do**: Do not implement full GamePlay turn passing outside README World scope; do not add AI/city/unit systems beyond documented placeholders needed to keep state stable.

  **Recommended Agent Profile**:
  - Category: `deep` - Reason: deterministic state transition and future integration seam.
  - Skills: [] - Pure service tests.
  - Omitted: [`playwright`] - No UI.

  **Parallelization**: Can Parallel: YES | Wave 4 | Blocks: T14 | Blocked By: T8, T11

  **References**:
  - External: `/home/justi/Projects/ctp2-src/ctp2_code/gs/gameobj/PlayerTurn.cpp` - turn-start sequencing.
  - External: `/home/justi/Projects/ctp2-src/ctp2_code/gs/gameobj/PlayerEvent.cpp` - pollution turn callbacks.
  - External: `/home/justi/Projects/ctp2-src/ctp2_code/gs/world/WrldPoll.cpp` - pollution/world turn logic.
  - README: `README.md` World bullet: world recalculated each turn.

  **Acceptance Criteria**:
  - [ ] `TestTurnRecalculationDeterministic` passes with fixed seed/params.
  - [ ] `TestTurnRecalculationMutatesOnlyTurnState` passes: static map snapshot unchanged.
  - [ ] `TestRecalcRequiresExistingWorld` passes.
  - [ ] `TestRecalcSerialization` passes under concurrent calls.

  **QA Scenarios**:
  ```
  Scenario: Recalculation is deterministic and scoped
    Tool: Bash
    Steps: Run `go test ./internal/world/... -run 'TestTurnRecalculation(Deterministic|MutatesOnlyTurnState)' -count=20`.
    Expected: same input produces same mutable state; static snapshot checksum unchanged.
    Evidence: .sisyphus/evidence/task-12-recalc-deterministic.txt

  Scenario: Concurrent recalculation is serialized
    Tool: Bash
    Steps: Run `go test ./internal/... -run TestRecalcSerialization -count=20`.
    Expected: no double-advance/corruption; deterministic final turn state.
    Evidence: .sisyphus/evidence/task-12-recalc-serialization.txt
  ```

  **Commit**: YES | Message: `feat(world): recalculate world turn state` | Files: [`internal/world/*recalc*`, store updates, tests]

- [x] T13. Swagger Documentation and README Checklist Sync

  **What to do**: Add Swagger comments/models for all implemented world routes, regenerate docs using the repository's existing workflow, and update README `World (/world)` checkboxes only for behaviors fully completed by tests. Keep unchecked any bullet whose parity requirement lacks exact seed/automated evidence. Include path/status/request/response examples for world read, ratings, pollution, and space placeholder.
  **Must NOT do**: Do not check all World bullets by assumption; do not document routes that are not registered; do not leave generated docs stale.

  **Recommended Agent Profile**:
  - Category: `writing` - Reason: docs plus precise checklist synchronization.
  - Skills: [] - Existing Swagger/Husky workflow.
  - Omitted: [`frontend-ui-ux`] - No visual design.

  **Parallelization**: Can Parallel: NO | Wave 4 | Blocks: T14 | Blocked By: T10, T11, T12

  **References**:
  - Pattern: `docs/swagger.json`, `docs/swagger.yaml`, `docs/docs.go` - generated docs outputs.
  - Pattern: `internal/game/endpoints.go` - Swagger comment style.
  - Pattern: `internal/game/swagger_models.go` - docs-only models.
  - README: `README.md` World section.
  - CONTRIBUTING: `CONTRIBUTING.md` Husky/Swagger comments guidance.

  **Acceptance Criteria**:
  - [ ] Swagger docs include all implemented `/world` paths and exact error/status contracts.
  - [ ] README World checkboxes match completed tests only.
  - [ ] `task test` still passes after docs/README changes.
  - [ ] Generated docs diff contains no unrelated path removals.

  **QA Scenarios**:
  ```
  Scenario: Swagger includes world paths
    Tool: Bash
    Steps: Run repository Swagger generation command from Taskfile/Husky workflow, then verify docs contain `/world/{game_id}`, `/world/{game_id}/ratings`, `/world/{game_id}/pollution`, `/world/{game_id}/space`.
    Expected: all implemented paths present with correct status codes; no unrelated docs removed.
    Evidence: .sisyphus/evidence/task-13-swagger.txt

  Scenario: README checkboxes reflect test evidence
    Tool: Bash
    Steps: Compare README World section against passing test evidence list.
    Expected: only completed/tested bullets are checked; parity bullets remain unchecked if exact seed evidence is incomplete.
    Evidence: .sisyphus/evidence/task-13-readme-sync.txt
  ```

  **Commit**: YES | Message: `docs(world): sync swagger and readme checklist` | Files: [`README.md`, `docs/*`, Swagger model/comment files]

- [x] T14. Reviewer Waves and Coverage Hardening

  **What to do**: Run reviewer waves after each big item and fix findings before continuing. At the end, run full test/coverage commands, address gaps, and collect evidence. Reviewer waves: R1 contract/model after T2-T4; R2 generation/persistence after T5-T7; R3 start/spawn/chat after T8-T9; R4 endpoints/docs after T10-T13; R5 turn recalc/whole-world consistency after T12-T13.
  **Must NOT do**: Do not skip reviewer waves; do not mark final verification complete without user approval; do not hide intentionally deferred README bullets.

  **Recommended Agent Profile**:
  - Category: `unspecified-high` - Reason: hands-on QA and review coordination.
  - Skills: [] - May use review-work-like pattern, but execute wave-specific checks.
  - Omitted: [`frontend-ui-ux`] - No UI.

  **Parallelization**: Can Parallel: NO | Wave 5 | Blocks: Final Verification | Blocked By: T1-T13

  **References**:
  - Plan: this file - reviewer wave requirements.
  - Task evidence: `.sisyphus/evidence/*` - red/green/refactor and QA proof.
  - CI: `.github/workflows/coverage.yml` - coverage threshold behavior.

  **Acceptance Criteria**:
  - [ ] R1-R5 reviewer outputs are saved and all blocking findings fixed.
  - [ ] `task test` passes.
  - [ ] `task test-coverage` passes and reports `>80%` internal coverage.
  - [ ] `go test ./...` passes or any broader-package failure is explained with repo-existing reason and `go test ./internal/...` still passes.
  - [ ] Atomic commit log shows completed slices with conventional messages.

  **QA Scenarios**:
  ```
  Scenario: Full test and coverage gate
    Tool: Bash
    Steps: Run `task test`, `task test-coverage`, and `go test ./internal/...`.
    Expected: all pass; coverage >80%.
    Evidence: .sisyphus/evidence/task-14-test-coverage.txt

  Scenario: Reviewer wave findings closed
    Tool: Bash
    Steps: Inspect saved R1-R5 reviewer outputs and rerun affected tests after each fix.
    Expected: every blocking finding has a fix commit or explicit non-blocking rationale; no unresolved critical findings.
    Evidence: .sisyphus/evidence/task-14-reviewer-waves.txt
  ```

  **Commit**: YES | Message: `chore(world): complete reviewer wave fixes` | Files: [only files changed by fixes/evidence policy]

## Final Verification Wave (MANDATORY — after ALL implementation tasks)
> 4 review agents run in PARALLEL. ALL must APPROVE. Present consolidated results to user and get explicit "okay" before completing.
> **Do NOT auto-proceed after verification. Wait for user's explicit approval before marking work complete.**
> **Never mark F1-F4 as checked before getting user's okay.** Rejection or user feedback -> fix -> re-run -> present again -> wait for okay.
- [x] F1. Plan Compliance Audit — oracle
  - Verify every README World bullet is either completed with evidence or explicitly left unchecked with reason.
  - Verify TDD evidence exists for every task.
  - Verify branch is `game-server-world` and based on local `main`.
- [x] F2. Code Quality Review — unspecified-high
  - Check Go package boundaries, no global RNG/state, no request-path CTP2 file parsing, no broad unrelated refactors, no new dependencies without justification.
- [x] F3. Real Manual QA — unspecified-high (+ Bash/HTTP, no browser required unless Swagger UI is inspected)
  - Run health/docs checks locally, exercise `/world` endpoints with fixed test data, save exact request/response evidence.
- [x] F4. Scope Fidelity Check — deep
  - Confirm no AI, CTP1 space mechanics, fog-of-war, graphics/client work, or undocumented gameplay mechanics were introduced.

## Commit Strategy
- Use atomic conventional commits after each completed vertical slice.
- Required examples:
  - `chore(world): create game-server-world branch preflight evidence`
  - `test(world): lock endpoint and generation contracts`
  - `feat(world): add deterministic ctp2 generation engine`
  - `feat(world): persist generated world on game start`
  - `feat(world): expose world read endpoints`
  - `docs(world): sync swagger and readme checklist`
- Never commit failing tests except as temporary red-phase evidence if repository convention explicitly allows it; otherwise capture red output in `.sisyphus/evidence` before making green commit.
- Run `git status`, `git diff`, and `git log --oneline -10` before every commit; stage only intended files.

## Success Criteria
- Local branch `game-server-world` exists and contains atomic commits for completed slices.
- `task test` and `task test-coverage` pass.
- `/world` Swagger docs exist and match route behavior.
- README World bullets are synchronized with implemented behavior.
- Deterministic seed tests prove generation, spawn placement, all host params including wetness, and turn recalculation behavior.
- Reviewer waves and final verification approve, then user explicitly approves completion.
