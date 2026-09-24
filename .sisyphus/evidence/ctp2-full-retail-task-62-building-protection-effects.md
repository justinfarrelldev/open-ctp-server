# CTP2 Full Retail Task 62: building protection effects

## Source references
- Retail `buildings.txt` includes city protection fields: `ProtectFromNukes`, `ProtectFromBioAgents`, `ProtectFromNanoVirus`, `PreventConversion`, `PreventSlavery`, and `ForceField`.
- `building.cdb` defines these as gameplay-affecting building bits/percent values.

## Implementation
- Retail building projection now emits safe runtime effects for nuclear, bio-agent, nano-virus, conversion, slavery, and force-field protection metadata.
- Production validation accepts the new effect kinds so completed retail buildings preserve these effects on city state for future special-action/combat resolution.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/rules -count=1`

## Remaining gaps
- The corresponding hostile actions (nuke planting, bio/nano terror, conversion, slavery) still need full server-side execution and protection rolls.
