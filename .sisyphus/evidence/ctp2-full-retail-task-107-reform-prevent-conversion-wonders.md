# CTP2 Full Retail Task 107: reform-cities and prevent-conversion empire wonders

## Source references
- Retail `Wonder.txt` gives Gutenberg's Bible `ReformCities` and `PreventConversion`.
- `Player.cpp` applies `PreventConversion` on wonder completion to unconvert owner cities.
- Existing bounded reform command semantics clear injunction/crime city state.

## Implementation
- Wonder completion now applies `reform_cities` to all owner cities by clearing injunction markers and crime percent.
- `prevent_conversion_empire` now blocks city-conversion special actions against any city owned by the protected player, returning a safe conversion-failed event.

## Tests
- `go test ./internal/gameplay/engine -run 'TestRetail(ReformCities|PreventConversionEmpire)' -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- The full retail converted-city identity/state model is not represented yet, so wonder completion cannot unconvert pre-existing converted-city records beyond preventing future conversion.
