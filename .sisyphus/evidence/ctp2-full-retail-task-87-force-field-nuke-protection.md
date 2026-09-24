# CTP2 Full Retail Task 87: force-field nuke protection

## Source references
- Retail building marker `ForceField` is a city protection effect.
- Nuclear city attacks should respect city protection systems.

## Implementation
- `plant-nuke` now treats city-local `force_field` as a nuke-blocking protection alongside `protect_from_nukes` and nation `eliminate_nukes`.
- The action still spends movement and emits a disclosure-safe failure event when blocked.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailPlantNukeDamagesCityAndHonorsProtection -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Full retail force-field combat protection, bombard interaction, and visual/diplomatic effects remain to be implemented.
