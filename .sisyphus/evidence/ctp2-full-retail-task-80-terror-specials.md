# CTP2 Full Retail Task 80: bio/nano/plague special actions

## Source references
- Retail unit fields include `BioTerror`, `NanoTerror`, and `Plague` special city actions.
- Retail building fields include `ProtectFromBioAgents` and `ProtectFromNanoVirus` protection percentages.

## Implementation
- Added `bio-terror`, `nano-terror`, and `plague` commands with disclosure-safe events.
- Capable adjacent owned units spend movement to apply the action to enemy cities.
- Bio/nano actions use deterministic special-action RNG against city protection effects.
- Successful actions reduce target city population by one without dropping below population one.
- Owner legal-action projections include only visible adjacent legal special targets.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailBioNanoAndPlagueSpecialsAffectCityPopulation -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/visibility -count=1`

## Remaining gaps
- Exact retail plague duration, diplomatic consequences, casualty curves, and capture/escape side effects remain to be matched.
