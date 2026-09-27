# CTP2 Full Retail Tasks 288-291: incite-revolution special execution

## Implementation
- Added `incite-revolution` as an executable special-action command using retail `InciteRevolution` chance, elite-chance, and death-chance metadata.
- Successful execution spends special movement, sets the target city watchful, transfers the city to the acting player as the current safe rebellion abstraction, and emits only aggregate/disclosed event payloads.
- Failure uses deterministic special-action RNG and death-chance resolution through existing special-death handling.
- Safe projections, HTTP routes, and AI scoring expose/use the action only from owned capable units.

## Source notes
- Retail `ArmyData::InciteRevolution` schedules `GEV_InciteRevolutionUnit`; `UnitData::InciteRevolution` reads chance/elite/death chances, sets the city watchful, and calls city revolt logic. The current server abstraction resolves the revolt as actor ownership transfer pending a fuller rebel-civilisation model.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailInciteRevolutionUsesChanceAndChangesCityOwner -count=1`
- `go test ./internal/gameplay/visibility -run TestTask12ProjectionIncludesRulesAndLegalActionMetadata -count=1`
- `go test ./internal/gameplay/http -count=1`
- `go test ./internal/gameplay/ai -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/visibility -count=1`
