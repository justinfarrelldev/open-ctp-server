# Client report implementation checklist

- [x] Read client ideal interaction topology and both reports completely.
- [x] Tighten both requirements documents for fog-safe ownership, optimistic presentation/rollback, full-state resync and hybrid play.
- [x] Create implementation plan and this live checklist.
- [ ] HTTP-01: RED dispatch/drift tests, GREEN exact registrations, refactor and verify.
- [ ] MAP-01/CITY-01/TRADE-01: RED visibility/memory/legacy/isolation tests, GREEN projections and persistence, refactor and verify.
- [ ] MAP-01: RED versioned world metadata/generation/handoff tests, GREEN implementation, refactor and verify.
- [ ] MAP-02/ECON-01: RED effective observation-safe facts and quote tests, GREEN implementation, refactor and verify.
- [ ] QUEUE-01: RED non-head cancellation regression, GREEN preservation, source-check head changes.
- [ ] QUEUE-02/03: RED atomic stable-entry edit/repeated unit tests, GREEN validated implementation, refactor and verify.
- [ ] Source-check distinct orders, persistent research goals and permitted city-art facts; implement supported contracts or document explicit unsupported decisions.
- [ ] Source-check retained movement and world/calendar/notification/history contracts; implement supported contracts or document explicit unsupported decisions.
- [ ] TURN-01/SYNC-01: RED hybrid phase, receipt/current snapshot and retry regressions; GREEN/refactor and verify.
- [ ] Document capabilities, wire semantics and remaining client integration work without marking client fixes complete.
- [ ] Full tests, vet, build, freeze lock, diff checks and unmodified commit hooks.
- [ ] Final security, maintainability and original-source review; reconcile both reports and checklist with actual results.

## Evidence

Implementation not yet started. Optional feature decisions are not completed features.
