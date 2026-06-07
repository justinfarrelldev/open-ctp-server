- Use local `main` as the branch base for `game-server-world`; no rebase/reset/pull/push was needed for this preflight step.
- Treat missing world-config fields as an expected gap to close in T2/T7 rather than trying to infer them from existing game DTOs alone.
- Follow current route/handler/doc conventions from `internal/game` instead of inventing a new HTTP surface pattern for `/world`.
- Keep the T2 stubs intentionally non-functional: `internal/world/contracts.go` only defines request/response shapes plus placeholder handlers, and `internal/game/world_contract_stubs.go` only maps the currently-known host params so the red suite fails on missing world semantics rather than on compile errors.
- Lock the space-layer contract to a fixed JSON error payload in tests now, even before handler implementation, because that 404 is explicitly documented as a stable CTP2 placeholder behavior.
- Keep the new world harness pure with file-level import guards and immutable clone helpers; the existing contract stubs may still import HTTP/DB for red semantics, but the new support files must not.
- Keep the T3 fixtures as in-repo metadata under `internal/world/fixtures/` plus a minimal `GenerateWorldFromFixture` red contract, so future generator work can consume deterministic references without adding any runtime file I/O to the external CTP2 source tree.
- Reuse the package-level `WorldTile` contract type and align snapshot/test helpers with it; this keeps the red suite focused on missing generator parity instead of unrelated scaffold type drift.
- Use a bounded deterministic spawn-selection pass instead of unbounded retries: rank spawnable land tiles deterministically, require setup-radius separation, and fail with  after a fixed 64-attempt ceiling.
- Keep the existing  ready-state contract stub backed by the same pure generator path as fixtures, so T2/T3 coverage shares one canonical world-output implementation rather than diverging test-only scaffolds.
- Clarification for the previous note: the bounded failure path is ErrWorldGenerationRetryCap, and the /world ready-state contract stub now reuses the same pure generator path as fixture generation.

- T7 persists immutable world data in `world_snapshots` and mutable recalculation data in `world_turn_states`, both keyed by `game_id`; create uses one transaction with an existence check so duplicate worlds fail cleanly and partial inserts roll back.
- Treat continent IDs as true connected land components before spawn selection; future start-placement logic can depend on those IDs safely once gameplay integration arrives.
- Keep the retry cap at 64 bounded placement passes, but let each pass scan the full deterministically ordered candidate set so impossible worlds still fail predictably without rejecting valid supported maps too early.
- For T8, export a tx-composable `world.CreateInTx` helper and call it from `StartGame` instead of opening a second world-store transaction; that keeps initial world persistence and the `games.status=started` transition atomic.
- For T8, use `SELECT ... FOR UPDATE` on the game row during StartGame so concurrent start attempts serialize before generation and persistence, and the loser observes a clean setup-state conflict instead of racing into duplicate world writes.
- Kept the implementation untouched and added only contract-level coverage in `internal/game/world_contract_test.go`; this preserves the existing lobby chat surface while making the active-game chat claim explicit for review.
- Captured the verification as evidence files under `.sisyphus/evidence/` instead of broadening into new endpoints or StartGame/world-generation changes.
- For the R3 follow-up, keep seed generation outside NormalizeWorldGenerationInput and inside the StartGame flow so production uses a server-only start-time seed while tests can inject a fixed seed through the package seam without reintroducing public determinism.
- Reuse  for setup-only mutators instead of adding a second concurrency mechanism, so StartGame and pre-start setup writes share one row-lock serialization discipline.
- Reuse getGameByIDForUpdate(...) for setup-only mutators instead of adding a second concurrency mechanism, so StartGame and pre-start setup writes share one row-lock serialization discipline.
- Kept the persisted /world read surface in internal/game rather than replacing the pure internal/world contract stubs, so the HTTP endpoint can share existing game auth/membership helpers and preserve the pure world package as a storage/generation boundary.
- Registered the new /world routes in main.go with Go 1.23 ServeMux path patterns (GET /world/{game_id} and GET /world/{game_id}/space) and the existing lobbyGameLimiter, matching current game/lobby route conventions without introducing a new limiter class.

- For T11 pollution reads, return  when present, otherwise fall back to persisted , then the stored turn-row update timestamp; this keeps  deterministic and non-zero without mutating state on GET.

- For T11 pollution reads, return the persisted pollution UpdatedAt when present, otherwise fall back to LastRecalculatedAt and then the stored turn-row update timestamp so the endpoint stays deterministic and read-only even before recalculation work lands.
- T12 uses a per-game TurnRecalculationService lock plus deterministic formulas for ratings, pollution, recalculation_count, and recalculated timestamps so concurrent calls serialize at the service boundary without advancing broader gameplay systems.
