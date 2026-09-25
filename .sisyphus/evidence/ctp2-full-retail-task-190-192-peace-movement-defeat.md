# CTP2 Full Retail Tasks 190-192: peace-movement defeat state

## Source evidence
- `Player::RegisterLostUnits` classifies casualties as home or overseas using nearest city distance versus `AtHomeRadius`.
- `PlayerHappiness::CalcPeaceMovement` decays raw overseas/home defeat pressure with `OverseasDefeatDecay`/`HomeDefeatDecay`, adds new losses scaled by `OverseasDefeatCoef`/`HomeDefeatCoef`, then resets per-turn loss counters.
- `Happy::CalcPeaceMovement` applies `-overseas_defeat + home_defeat - overseas - war_discontent`, divided by city count and scaled by `LowerPeaceMovement` buildings.

## Implementation
- Added canonical `PeaceMovementState` to nations for raw defeat pressure and per-turn home/overseas loss counters.
- Combat casualty removal records each lost unit as home or overseas via nearest owned city and active at-home radius effects.
- Round recalculation advances defeat pressure using runtime government decay/coefficients, resets loss counters, and applies signed peace-movement happiness adjustments alongside existing overseas deployment and war-discontent penalties.

## Tests
- `go test ./internal/gameplay/engine -run 'Test(PeaceMovementDefeatStateAdjustsHappiness|CombatLossesClassifyPeaceMovementHomeAndOverseas|AdvancePeaceMovementAppliesDefeatDecayAndResetsLossCounters)' -count=1`
- `go test ./internal/gameplay/engine -count=1`
