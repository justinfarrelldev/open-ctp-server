# CTP2 Full Retail Task 84: plant nuke special action

## Source references
- Retail unit field `PlantNuke` marks units able to plant nuclear devices in enemy cities.
- Retail building/wonder fields include `ProtectFromNukes` and `EliminateNukes` protections.

## Implementation
- Added `plant-nuke` command and disclosure-safe success/failure events.
- Capable adjacent owned units spend movement to target enemy cities.
- City-local `protect_from_nukes` and target-nation `eliminate_nukes` block detonation.
- Successful detonations halve target city population (not below one) and add pollution to the city tile.
- Owner legal-action projections expose visible adjacent nuke targets for capable units.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailPlantNukeDamagesCityAndHonorsProtection -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/visibility -count=1`

## Remaining gaps
- Exact retail nuclear blast radius, unit/city destruction, diplomatic consequences, fallout, and global wonder effects remain to be matched.
