# CTP2 Full Retail Tasks 210-212: counter-bombard units

## Source evidence
- `unit.cdb` defines `CanCounterBombard`.
- Retail `Units.txt` marks units such as `UNIT_LEVIATHON` with `CanCounterBombard`.
- Original `UnitData::CanCounterBombard` checks the flag and bombard eligibility; `ArmyData` lets defending units counter-bombard the attacking stack after bombardment.

## Implementation
- Retail runtime imports `CanCounterBombard` as a `can_counter_bombard` unit effect.
- Bombard resolution now lets surviving eligible target defenders counter-bombard the attacking unit without spending normal movement.
- Counter-bombard uses existing firepower-aware bombard damage and domain eligibility, then reports safe `counter_bombard_hits` in the bombard event payload.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailCanCounterBombardReturnsFireAfterBombardment -count=1`
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/rules -count=1`
- `go test ./internal/gameplay/visibility -count=1`
