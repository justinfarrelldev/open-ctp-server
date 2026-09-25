# CTP2 Full Retail Task 181: no-pollution-unhappiness wonder

## Source evidence
- Retail `wonder.cdb` defines `NoPollutionUnhappiness`.
- `WonderRecord` exposes `GetNoPollutionUnhappiness`.
- `Happy.cpp::CalcPollution` sets pollution happiness impact to zero when `wonderutil_GetNoPollutionUnhappiness` is active.

## Implementation
- Retail wonder projection now maps `NoPollutionUnhappiness` to `no_pollution_unhappiness` marker effects when present.
- City pollution-unhappiness penalty now returns zero if the city owner has an active wonder with that effect.
- Effect validation recognizes the marker.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailNoPollutionUnhappinessWonderSuppressesPenalty -count=1`
- `go test ./internal/gameplay/rules -run TestRetailWonderNoPollutionUnhappinessProjectsMarker -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/rules -count=1`
