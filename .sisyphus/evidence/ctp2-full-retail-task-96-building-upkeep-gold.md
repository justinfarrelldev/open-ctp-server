# CTP2 Full Retail Task 96: building upkeep gold maintenance

## Source references
- Retail building `Upkeep` is gold maintenance, not a shield/production penalty.
- Existing production cost/shield values are separate from gold upkeep in retail data.

## Implementation
- City yield calculation now subtracts non-obsolete building upkeep from gold yield rather than production yield.
- Gold yield is clamped at zero after maintenance.
- Production yield is left unchanged by building upkeep.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailBuildingUpkeepReducesGoldNotProduction -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Full nation treasury/accounting, unit support, and bankruptcy/disband mechanics still need retail parity.
