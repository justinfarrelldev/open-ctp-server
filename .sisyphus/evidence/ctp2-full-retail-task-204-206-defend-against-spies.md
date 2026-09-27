# CTP2 Full Retail Tasks 204-206: DefendAgainstSpies counter-spy behavior

## Source evidence
- `unit.cdb` defines `DefendAgainstSpies` as a unit float bit field.
- Retail `Units.txt` gives counter-spy values such as `UNIT_CYBER_NINJA` with `DefendAgainstSpies 0.5`.
- Original `UnitData::StoppedBySpies` checks units on the target city cell, rolls `chance * 100`, and kills the acting spy if stopped before the city special action proceeds.

## Implementation
- Retail runtime now imports `DefendAgainstSpies` as `defend_against_spies` percent effects.
- City-target spy actions (`steal-technology`, `convert-city`, `investigate-city`, `plant-nuke`, `assassinate-ruler`) now roll target-city owner defenders before resolving the action.
- A stopped spy is removed, emits the relevant failed special-action event with a safe payload, and does not additionally mark the target city watchful.
- Enemy-visible unit projections continue to expose only public target marker effects, so hidden counter-spy odds are not leaked across the trust boundary.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailDefendAgainstSpiesStopsAndKillsCitySpy -count=1`
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/rules -count=1`
- `go test ./internal/gameplay/visibility -count=1`
