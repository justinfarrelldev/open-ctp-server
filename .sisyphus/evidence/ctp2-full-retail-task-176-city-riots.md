# CTP2 Full Retail Task 176: city riots

## Source evidence
- Retail `Const.txt` defines `RIOT_LEVEL 73`.
- `CityData::CheckRiot` marks low-happiness cities as rioting.
- Retail city production storage and trade collection paths zero production/trade for rioting cities.

## Implementation
- Canonical cities now persist a `rioting` flag.
- Turn recalculation marks explicitly low-happiness cities (`0 < happiness < 73`) as rioting.
- Rioting cities lose production, gold, and science yields for the turn.
- Owner city projections expose rioting status.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailLowHappinessCityRiotsAndLosesProductionGold -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/visibility -count=1`

## Remaining gaps
- Retail random riot chance and full default-happiness initialization remain future parity refinements.
