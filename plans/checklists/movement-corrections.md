# Movement corrections

- [x] Inspect client screenshot/handoff and released movement source.
- [x] RED regression coverage for access, first moves, armies, restart, routing and privacy (failures observed before each fix).
- [x] GREEN versioned retail access and protected-city policy shared with planning.
- [x] GREEN persisted first-move admission/spend, creation/refresh/actions and planning/AI.
- [x] GREEN deterministic visible-knowledge routing, replanning and waiting.
- [x] One released-record-driven subtest for each of all 74 units; RED exposed fixed air cost, amphibious entry and stationary city records; GREEN after corrections.
- [x] Replace initial least-cost search with released UnitAstar A* at the user's request; verify capped costs and human-only DecayOrtho preference in regression tests.
- [x] Refactor and thoroughly review against released CTP2 mechanics; share sight/detection, per-member admission, costs and access; remove replaced greedy/dead helpers.
- [x] Final review verified Player::ResetMovement refreshes cargo too; add failing cargo refresh/planning regressions, implement corrected-pin cargo refresh, and rerun focused tests.
- [x] Preserve legacy projected route selection after full-suite acceptance exposed a version-zero exploration regression; rerun both previously failing public conquest tests successfully.
- [x] Full `go test -timeout=10m -p=2 ./... -count=1`, including public fullgame acceptance and freeze guard.
- [x] Race tests for engine, visibility and AI; `go vet ./...`, production build to `/tmp/open-ctp-server-movement`, and `git diff --check`.
- [x] Document legacy policy and live-game/client validation boundary in `docs/movement-corrections.md`; commit the coupled movement correction atomically with its regression tests, docs and reviewed freeze manifest using the unchanged hook.
