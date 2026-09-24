# CTP2 Full Retail Task 135: lawsuit unit targets

## Source evidence
- `ArmyData::Sue` targets adjacent/same-tile armies with `CanBeSued` units and emits `GEV_Lawsuit`.
- `LawsuitEvent` iterates units on the target cell and kills each unit whose DB record has `CanBeSued`.

## Implementation
- `SpecialActionPayload` now supports `target_unit_id` for `CommandSue` unit lawsuits.
- Unit lawsuit path requires an acting `special_sue` unit and an adjacent/same-tile enemy target with `can_be_sued`.
- All visible units on the target tile with `can_be_sued` are removed; non-suable guards remain.
- Event payload remains disclosure-safe and does not include target unit IDs.

## Tests
- `go test ./internal/gameplay/engine -run 'TestRetailSue(Unit|City|Franchise)' -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Safe action projection for unit lawsuits remains conservative because enemy unit marker effects are not projected.
