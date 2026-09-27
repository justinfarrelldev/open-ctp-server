# CTP2 Full Retail Tasks 317-323: religious conversion and indulgence branches

## Source notes
- Retail `GEV_ConvertCity` calls `CityData::ConvertTo(player, by)`; conversion is a religious state, not city ownership transfer.
- Retail indulgence sales branch on `c.IsConvertedTo()`:
  - unconverted: `UNCONVERTED_INDULGENCE_*`
  - converted to seller: `CONVERTED_INDULGENCE_*`
  - converted to another faith: `OTHER_FAITH_INDULGENCE_*`

## Implementation
- Imported converted/other-faith indulgence constants from `Const.txt` into retail/runtime constants.
- Added canonical city conversion fields (`converted_to_player_id`, `converted_by`).
- Convert-city now preserves city ownership and records conversion owner/source (`cleric` or `televangelist`).
- Indulgence-sales now selects unconverted, converted, or other-faith gold/happiness constants from runtime rules.
- Safe projections expose conversion state only to the city owner or the converting player.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -run 'TestRetail(ConvertCityHonorsProtectionAndDisclosure|IndulgenceSalesUsesConvertedAndOtherFaithConstants|IndulgenceSalesTransferGoldAndTimedHappiness)' -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/rules -count=1`
- `go test ./internal/gameplay/visibility -count=1`
