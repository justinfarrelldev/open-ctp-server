# CTP2 Full Retail Task 75: population-removing unit production

## Source references
- Retail `Units.txt` marks `UNIT_SETTLER` with `BuildingRemovesAPop`.
- `unit.cdb` defines `BuildingRemovesAPop` as a gameplay-affecting unit flag.

## Implementation
- Retail unit projection maps `BuildingRemovesAPop` to a runtime `building_removes_population` effect.
- Completing a unit with this effect reduces the producing city's population by one, without dropping below population one.
- AI/action fortify filtering treats this as retail capability metadata for `CanEntrench` gating.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailPopulationRemovingUnitsReduceCityPopulation -count=1`
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/rules ./internal/gameplay/ai ./internal/gameplay/visibility -count=1`

## Remaining gaps
- Retail legality around building pop-consuming units at minimum city size still needs implementation.
