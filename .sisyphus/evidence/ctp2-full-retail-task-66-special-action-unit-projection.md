# CTP2 Full Retail Task 66: special-action unit projection

## Source references
- Retail `Units.txt` defines special unit actions including `ConvertCities`, `CreateFranchise`, `InvestigateCity`, `PlantNuke`, `BioTerror`, `NanoTerror`, `Plague`, `AssasinateRuler`, `SlaveRaids`, `SettlerSlaveRaids`, `UndergroundRailway`, `CanInjoin`, `CanReform`, `CreateParks`, `CanExpel`, and `CanSue`.
- `unit.cdb` models these as gameplay-affecting unit capabilities with chance/death fields.

## Implementation
- Retail unit projection now preserves these special-action capabilities as safe runtime effects with chance values where available.
- Production validation accepts the new effect kinds so completed special units retain their retail capability metadata.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/rules -count=1`

## Remaining gaps
- Most newly projected special actions still need full command execution, protection checks, diplomatic consequences, and safe action visibility.
