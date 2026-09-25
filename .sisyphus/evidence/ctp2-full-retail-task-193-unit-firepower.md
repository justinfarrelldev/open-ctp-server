# CTP2 Full Retail Task 193: unit firepower combat damage

## Source evidence
- `unit.cdb` defines required unit `Firepower`.
- Retail `Units.txt` assigns non-default firepower values, e.g. `UNIT_BATTLESHIP Firepower 4`.
- `UnitData.cpp` applies attacking and defending unit firepower when deducting hit points.

## Implementation
- Runtime unit and starter definitions now preserve firepower from synthetic and retail rules.
- Produced and starting units carry canonical `firepower`; owner projections expose it with other visible owned combat stats.
- Combat damage multiplies the existing deterministic damage floor by attacker firepower, treating legacy zero firepower as `1` for old states.

## Tests
- `go test ./internal/gameplay/engine -run 'TestRetailFirepowerScalesCombatDamage|TestProductionCompletionCopiesRuntimeRangedStats' -count=1`
- `go test ./internal/gameplay/rules -run TestRetailRuntime -count=1`
