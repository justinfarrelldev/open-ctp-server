# CTP2 Full Retail Task 147: martial-law happiness

## Implementation
- Retail unit `ExertsMartialLaw` imports as an `exerts_martial_law` runtime effect.
- Retail government `MaxMartialLawUnits`, `MartialLawEffect`, and `MartialLawThreshold` import into runtime government definitions and are applied as government effects.
- City happiness now counts same-tile owned martial-law units, caps counted units by government, adds the government martial-law effect, and caps the boosted value at the government threshold when base happiness is below the threshold.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailMartialLawHappinessUsesGovernmentCapAndUnitMarkers -count=1`
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/rules -count=1`

## Remaining gaps
- Full riot/unrest and war-discontent happiness systems remain future government/economy parity work.
