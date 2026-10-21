# Implementation checklist

- [x] Inspect canonical commands, turn lifecycle, persistence and CTP2 source.
- [x] Record implementation scope and source-backed limits in plan.
- [x] Movement RED: destination continuation, exact membership, obstruction,
  manual/explicit cancellation, replay and turn-mode tests fail first.
- [x] Movement GREEN/REFACTOR: canonical orders and lifecycle, tests pass.
  Verified engine, visibility and HTTP packages; private reconnect state, generic
  obstacle/ZOC stop, manual overrides and all four scheduler modes. Source:
  ArmyData.cpp ExecuteMoveOrder/Move retains orders until destination completion.
  Added set/cancel routes and phase-aware coordinate intent advertisement.
- [x] Ruins RED: consumption/rewards, versioned generation, fog/restart tests.
- [x] Ruins GREEN/REFACTOR: authoritative villages-only rewards and generation.
  Gold/city/empty tests failed before implementation; unit creation tests caught
  upgrade-style zero movement and now verify full source-backed creation movement.
  Advance rewards use normal feat/age/research handling. Generation/fog tests
  failed before implementation and now pass with legacy terrain/spawn equality.
  Verified engine, world, game and visibility packages. Source: GoodyHuts.cpp/.h,
  Advances::GetMinPrerequisites, UnitData.cpp initial movement, risks.txt villages
  profile and Const.txt RUINS_BOX_WIDTH/HEIGHT=6, RUINS_CHANCE_PER_BOX=1.0.
- [x] Governors RED: enable/disable/priorities, queue continuity and privacy tests.
- [x] Governors GREEN/REFACTOR: pinned-list city production automation.
  Engine tests failed on unsupported command; projection tests failed on missing
  capability before wiring. Tests now pass for priorities, empty-queue continuity,
  manual progress preservation, disable, restart and conquest reset. Source:
  Governor::FillEmptyBuildQueues and released Building/WonderBuildLists; scope is
  city production, not original empire-wide population/army/slider automation.
- [x] Wire routes, legal actions/intents, snapshots/capabilities; regression tests.
  Engine, visibility and HTTP package suites pass.
- [x] Transactional service idempotency/reconnect coverage for new commands.
  Isolated SQL mocks verify set/cancel journeys, governor enablement and ruin moves
  persist authority and complete private receipt snapshots in one transaction;
  replay reads the original receipt without re-executing or repaying rewards.
  Exact HTTP payload dispatch/forged-field tests and cross-file route guards pass.
- [x] Update contracts/reports with implemented scope and remaining limitations.
  Published routes, payloads, owner-private persistence, geometric-route stops,
  villages-only probabilities/results and production-only governor boundaries.
- [x] Thorough source-backed code review and fix findings with regression tests.
  RED review regressions caught impossible-budget journeys, ZOC arrival mislabeled
  as blocked, alphabetically earlier future settlers selected as nomads and ruins
  retained under founded cities. GREEN fixes use real creation movement/HP bonuses,
  explicit starter identity, bounded prerequisite recursion and source city-removal
  semantics. Automatic journeys stop processing later orders on conquest victory.
  Full-suite RED also caught an E2E assumption that one battle ends conquest:
  villagers can leave additional hostile recruits without a city. The harness now
  pursues visible hostile assets, preserves per-unit exploration history and seeks
  public discovery boundaries, without hidden reads or larger turn/action budgets.
  Verified clean/clean/restart conquest equality and refreshed observed hashes;
  science events remain unchanged, with state hash updated for persisted ruin fog.
- [x] Full tests, vet, build, freeze guard, diff checks and atomic commits.
  `GOMAXPROCS=4 go test -timeout=16m -p=2 ./... -count=1` passes, including
  fullgame/retail autoplay and clean/clean/restart acceptance. Race tests pass for
  all gameplay packages plus game/world. `go vet ./...`, `go build ./...` and
  `git diff --check` pass. Freeze coverage grows from 64 to 69 reviewed production
  files; no prior entries removed. Explicit released-catalog governor/nomad tests
  execute real pinned building lists and UNIT_SETTLER. A focused terminal-boundary
  regression also verifies that conquest stops later automatic journeys.
  All commits use the unchanged pre-commit hook. No live service/client mutations.
