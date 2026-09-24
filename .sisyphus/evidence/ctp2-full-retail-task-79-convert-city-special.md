# CTP2 Full Retail Task 79: convert city special action

## Source references
- Retail unit field `ConvertCities` marks units able to convert enemy cities.
- Retail building/wonder fields include `PreventConversion` protection.

## Implementation
- Added `convert-city` command and city conversion success/failure events.
- Units with `special_convert_city` may convert adjacent enemy cities, spending movement.
- City-local `prevent_conversion_percent` consumes deterministic special-action RNG and can block conversion.
- Legal-action projection exposes convert-city actions only for visible adjacent enemy cities and capable owned units.
- Event payloads avoid disclosing target city identifiers or technology/state details.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailConvertCityHonorsProtectionAndDisclosure -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/visibility -count=1`

## Remaining gaps
- Exact retail conversion probability/cost/diplomacy and unit-capture side effects still need parity review.
