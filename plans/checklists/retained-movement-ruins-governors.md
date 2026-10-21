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
- [ ] Ruins RED: consumption/rewards, versioned generation, fog/restart tests.
- [ ] Ruins GREEN/REFACTOR: authoritative villages-only rewards and generation.
- [ ] Governors RED: enable/disable/priorities, queue continuity and privacy tests.
- [ ] Governors GREEN/REFACTOR: pinned-list city production automation.
- [ ] Wire routes, legal actions/intents, snapshots/capabilities; regression tests.
- [ ] Transactional service idempotency/reconnect coverage for new commands.
- [ ] Update contracts/reports with implemented scope and remaining limitations.
- [ ] Thorough source-backed code review and fix findings with regression tests.
- [ ] Full tests, vet, build, freeze guard, diff checks and atomic commits.
