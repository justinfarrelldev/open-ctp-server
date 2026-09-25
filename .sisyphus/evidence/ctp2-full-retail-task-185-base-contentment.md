# CTP2 Full Retail Task 185: base contentment happiness

## Source evidence
- `Happy.cpp::CalcHappiness` starts city happiness with `CalcBase(p)`.
- `Happy.cpp::CalcBase` returns `Player::GetBaseContentment()`.
- Retail `DiffDB.txt` sets `BASE_CONTENTMENT 75` for released difficulties.

## Implementation
- City happiness now defaults to retail base contentment `75` before applying city/building/wonder/specialist/social/empire penalties and bonuses.
- Synthetic fixtures may still use an explicit `EffectHappiness` with ID `base` to override the base value for focused low/high-happiness scenarios.
- Fullgame expected state hashes were updated because canonical city happiness now stores retail base values instead of zero after recalculation.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailCityHappinessDefaultsToBaseContentment -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/e2e/freeze -count=1`
- `go test ./internal/e2e/fullgame -count=1`
