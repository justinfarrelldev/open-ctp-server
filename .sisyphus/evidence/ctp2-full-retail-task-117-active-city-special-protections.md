# CTP2 Full Retail Task 117: active city special-action protections

## Implementation
- Special-action protection checks now use the city's active aggregate effects rather than only direct city markers.
- This makes active buildings/wonders participate in:
  - nuke/force-field protection,
  - bio/nano terror protection,
  - slavery protection,
  - conversion protection.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailSpecialProtectionsUseActiveCityBuildings -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Exact CTP2 spy capture/escape and diplomatic side-effects for protected failed actions still need deeper parity work.
