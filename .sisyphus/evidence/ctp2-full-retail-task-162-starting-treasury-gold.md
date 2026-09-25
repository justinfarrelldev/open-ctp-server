# CTP2 Full Retail Task 162: starting treasury gold

## Implementation
- Gameplay startup now initializes each nation's treasury `gold` from the configured starting gold value.
- Existing public-works initialization is preserved for compatibility until public-works/material-pool generation is fully retail-modeled.

## Tests
- `go test ./internal/game -run TestInitialGameplayStateDerivesEntitiesFromSnapshot -count=1`
- `go test ./internal/game -count=1`

## Remaining gaps
- Public works should eventually be generated through the retail material/public-works economy rather than mirroring starting gold.
