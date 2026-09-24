# CTP2 Full Retail Task 133: sue-franchise own-city path

## Source evidence
- `ArmyData::Sue` routes same-owner city targets to `UNIT_ORDER_SUE_FRANCHISE` / `SueFranchise`.
- `ArmyData::SueFranchise` requires the target city to be owned by the lawyer's owner and already franchised.
- `RemoveFranchiseEvent` calls `SetFranchiseTurnsRemaining(TurnsToSueFranchise)`, and `CityData::SetFranchiseTurnsRemaining` removes the franchise when the remaining turns are less than one.

## Implementation
- `CommandSue` can now target the actor's own franchised city.
- The own-city path requires an external franchise marker, spends movement, and removes the external franchise marker.
- The prior enemy-city lawsuit path remains supported.
- Safe projections now expose the `sue` action for own visible cities with external franchise markers.

## Tests
- `go test ./internal/gameplay/visibility -run TestRetailSueFranchiseActionProjectsForOwnFranchisedCity -count=1`
- `go test ./internal/gameplay/engine -run 'TestRetailSue(CityAppliesLawsuit|FranchiseRemovesOwnCityExternalFranchise)' -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/visibility -count=1`

## Remaining gaps
- Lawsuit actions against units and configurable delayed franchise removal durations remain future work.
