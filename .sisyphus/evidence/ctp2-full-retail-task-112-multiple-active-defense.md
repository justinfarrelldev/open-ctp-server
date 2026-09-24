# CTP2 Full Retail Task 112: multiple active-defense bombard interceptors

## Source references
- Retail active-defense units can protect bombardment targets within range; prior implementation preserved stats and used the first interceptor only.

## Implementation
- Bombardment now resolves active-defense interceptors in deterministic unit-ID order.
- Each eligible interceptor can fire until the bombarder is destroyed, after which bombard rounds are skipped.
- Active-defense hit counts continue to use the disclosure-safe aggregate bombard event payload.

## Tests
- `go test ./internal/gameplay/engine -run 'TestRetail(MultipleActiveDefenses|ActiveDefenseIntercepts)' -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Retail counter-bombard, ammunition/ROF nuances, and domain-specific active-defense masks still need deeper audit.
