# CTP2 Full Retail Task 169: building maintenance deficits debit treasury

## Source evidence
- `CityData::PayWages` and gold maintenance paths in the retail source allow city trade/gold deficits to draw from player treasury.
- Building upkeep is subtracted from city trade/gold; when upkeep exceeds local gold, the remaining cost affects the owning player.

## Implementation
- Treasury advancement now computes city building-maintenance deficits from gross city gold without upkeep.
- Deficits debit the nation treasury in addition to unit gold upkeep.
- Treasury remains clamped at zero for safe canonical state.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailBuildingMaintenanceDeficitDebitsTreasury -count=1`
- `go test ./internal/gameplay/engine -count=1`
