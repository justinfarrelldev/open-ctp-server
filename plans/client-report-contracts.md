# Client report contract implementation

## Authority and scope

Implement the server requirements in `client-discovered-server-issue.md` and
`client-discovered-server-issues.md` under the client's ideal interaction topology.
Never disclose never-observed geography, including through ownership or previews.
Optimism, rollback animation and the resynchronizing banner are client presentation
requirements; the server must provide revisioned viewer-safe receipts/current reads.
Keep legacy persisted games and generation versions stable.

## Method

Use red → green → refactor for each behavior slice, recording tests and source
verification in `plans/checklists/client-report-contracts.md`. Use injected stores,
in-memory engine fixtures and isolated HTTP stubs. Do not call live services.
Make atomic commits with the existing hooks unchanged. Review original CTP2 source
before selecting gameplay semantics. Run targeted tests, full tests, vet, build and
freeze verification before closure.

## Ordered work

1. Tighten both reports: fog-safe ownership, provisional movement/reconciliation,
   full-state resync and explicit hybrid coverage.
2. HTTP-01: register every advertised executable action with exact payload fields;
   add dispatch and cross-layer route/field drift guards.
3. MAP-01 / CITY-01 / TRADE-01: preserve visible and remembered independent tile
   facts, city names/types and permitted public art facts; protect legacy unknowns,
   player isolation and receipts. Route accounting is endpoint-owner-only.
4. MAP-01 generation: versioned deterministic server river/goods generation,
   snapshot persistence and world-start handoff, retaining legacy generator output.
5. MAP-02 / ECON-01: engine-derived inspectable tile facts and last observations;
   separate revision-bound rush quotes from executable payloads; explicit forecast
   status/assumptions rather than invented values.
6. QUEUE-01/02/03: tail cancellation preserves head progress; stable entry identity,
   validated atomic edits and source-allowed repeated unit orders; backward-safe
   legacy cancellation and persistence.
7. ORDERS-01 / POLICY-01 / RESEARCH-01 / CITY-02 / MOVE-01 / PRESENTATION-01:
   source-check and implement supported contracts; explicitly document unsupported
   features rather than inventing culture, calendar, optimizer or governor mechanics.
   Publish public map bounds independently of discoveries. Retained orders, if
   implemented, must be server-owned and replay-safe, not client path retries.
8. TURN-01 / SYNC-01: hybrid phase/action and revisioned receipt/current snapshot
   regressions with persisted observations and idempotent replay.
9. Document wire contracts/capabilities and client decoder follow-up, update report
   acceptance status only for verified work, run full verification and final review.

Feature requests that explicitly require a product/rules decision must be listed
as unsupported, never advertised as executable or called implemented. Do not weaken
validation or visibility to satisfy existing client fallbacks. The checklist must
separate these decisions from implemented server defects.
