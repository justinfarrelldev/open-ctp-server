# CTP2 Full Retail Task 72: CanEntrench fortification legality

## Source references
- Retail `Units.txt` marks fortifiable military units with `CanEntrench`; many civilian/special units omit it.
- `unit.cdb` defines `CanEntrench` as a gameplay-affecting unit flag.

## Implementation
- Retail unit projection maps `CanEntrench` to a runtime `can_entrench` effect.
- Fortify command validation rejects retail-capability units that lack `can_entrench`, while preserving legacy synthetic/minimal-unit behavior.
- AI and safe action projections only offer fortify for units that pass the same projected fortification rule.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailCanEntrenchRestrictsFortify -count=1`
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/rules ./internal/gameplay/ai ./internal/gameplay/visibility -count=1`

## Remaining gaps
- Full retail entrenchment timing/bonus decay and UI messaging still need deeper parity review.
