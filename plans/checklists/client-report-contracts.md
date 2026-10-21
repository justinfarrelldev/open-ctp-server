# Client report implementation checklist

- [x] Read client ideal interaction topology and both reports completely.
- [x] Tighten both requirements documents for fog-safe ownership, optimistic presentation/rollback, full-state resync and hybrid play.
- [x] Create implementation plan and this live checklist.
- [x] HTTP-01: RED dispatch/drift tests, GREEN exact registrations, refactor and verify.
- [x] MAP-01/CITY-01/TRADE-01: RED visibility/memory/legacy/isolation tests, GREEN projections and persistence, refactor and verify.
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

- HTTP-01 RED: all 18 reported routes plus unregistered pillage returned 404; literal-action drift guard identified missing registrations.
- HTTP-01 GREEN/refactor: exact HTTP payload registrations (including both lawsuit targets), shared dispatch tests and forged-field rejection pass; `go test ./internal/gameplay/http -count=1` passes. Existing authenticated service still performs all gameplay validation. Source: `UnitData.cpp` CanInjoin/CanExpel, existing `special.go` payload decoders and engine tests. Freeze hash renewed for the reviewed adapter.
- MAP/CITY/TRADE RED: visible/stale rivers, goods, ownership, names/type/population and route accounting absent. GREEN/refactor: optional metadata distinguishes legacy unknown from explicit negatives; foreign city population is public only when observed, private economy remains protected; endpoint accounting limited to owners with currently visible endpoints. Hidden changes, re-observation, player isolation, never-explored neighbors, canonical round trips and mocked durable move receipts pass. Source: `gfx/tilesys/tiledraw.cpp:4616` DrawCityNames and `gs/world/UnseenCell.cpp:45–86` store observed environment/name/size. Targeted visibility/service tests pass.
- Optional feature decisions are not completed features.
