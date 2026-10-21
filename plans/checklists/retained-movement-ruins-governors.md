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
- [ ] Governors RED: enable/disable/priorities, queue continuity and privacy tests.
- [ ] Governors GREEN/REFACTOR: pinned-list city production automation.
- [ ] Wire routes, legal actions/intents, snapshots/capabilities; regression tests.
- [ ] Transactional service idempotency/reconnect coverage for new commands.
- [ ] Update contracts/reports with implemented scope and remaining limitations.
- [ ] Thorough source-backed code review and fix findings with regression tests.
- [ ] Full tests, vet, build, freeze guard, diff checks and atomic commits.
