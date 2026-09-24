# CTP2 Full Retail Task 64: remaining building marker effects

## Source references
- Retail `buildings.txt` includes marker/value fields such as `FoodVat`, `AllowGrunts`, `Brokerage`, `Capitol`, `Cathedral`, `DoubleTelevangelists`, `IsReligious`, `LowerPeaceMovement`, `NuclearPlant`, and `Television`.
- `building.cdb` defines these as gameplay-affecting building fields.

## Implementation
- Retail runtime projection now preserves these building marker/value fields as safe city effects.
- Production validation accepts the new effect kinds so runtime-completed retail buildings retain the metadata needed for later mechanics.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/rules -count=1`

## Remaining gaps
- Several marker effects still require full execution semantics (for example religious conversion modifiers, nuclear-plant/wonder interactions, TV/gold interactions, and FoodVat warning/pollution behavior).
