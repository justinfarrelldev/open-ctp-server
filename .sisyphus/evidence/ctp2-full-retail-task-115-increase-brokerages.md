# CTP2 Full Retail Task 115: IncreaseBrokerages wonder effect

## Source notes
- CTP2 `buildingutil_GetCommercePercent` increases commerce percent only for buildings flagged `Brokerage` by `wonderutil_GetIncreaseBrokerages(...) / 100.0`.
- Retail data has `WONDER_ELECTRONIC_COMMERCE_ACT IncreaseBrokerages 50` and brokerage buildings carry both `CommercePercent` and `Brokerage`.

## Implementation
- City gold yield now adds an extra gold-percent bonus for active brokerage buildings when the owner has an active `increase_brokerages` wonder effect.
- The bonus is based on each active brokerage building's `yield_gold_percent` value, not on unrelated non-brokerage commerce effects.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailIncreaseBrokeragesAmplifiesBrokerageCommerce -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- The exact retail fractional-percent rounding path remains integerized with the rest of the current yield model.
