# CTP2 Full Retail Task 102: increase-HP wonder effects

## Source references
- `ctp2_source_code/ctp2_code/gs/gameobj/Player.cpp` applies `wonderutil_GetIncreaseHP` to all existing owner units when the wonder is built.
- `ctp2_source_code/ctp2_code/gs/gameobj/UnitData.cpp` creates units with `GetMaxHP() + wonderutil_GetIncreaseHP(owner built wonders)`.

## Implementation
- Wonder completion now applies `increase_hp` to existing owner units, increasing current and max health to preserve damage ratios safely for combat validation.
- Runtime unit production now includes owner wonder `increase_hp` bonuses for future units.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailIncreaseHPWonderRaisesExistingAndFutureUnitHealth -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Exact CTP2 healing interactions with forts/installations and wonder removal/destroyed rollback remain future work.
