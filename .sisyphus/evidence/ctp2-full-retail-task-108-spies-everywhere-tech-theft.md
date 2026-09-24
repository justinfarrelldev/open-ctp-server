# CTP2 Full Retail Task 108: spies-everywhere stops technology theft

## Source references
- `UnitData::StoppedBySpies` checks `wonderutil_GetSpiesEverywhere` for the target owner and uses a 50% spy-stop chance.
- `UnitData::StealTechnology` calls `StoppedBySpies` before resolving steal chance; stopped spies are killed and the action fails.

## Implementation
- Technology theft now checks target-owner `spies_everywhere` before steal chance resolution.
- The check uses the deterministic special-action RNG domain and a 50% threshold.
- A stopped spy is removed through canonical unit removal and the action emits the existing safe technology-failed event.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailSpiesEverywhereStopsTechnologyTheftAndKillsAgent -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Retail applies the same stopped-by-spies path to additional spy actions (investigate, incite/revolution, assassinate, nullify walls); those still need individual safe events and command handling.
