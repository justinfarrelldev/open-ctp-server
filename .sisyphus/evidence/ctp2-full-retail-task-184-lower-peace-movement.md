# CTP2 Full Retail Task 184: lower peace movement city effects

## Source evidence
- `Happy.cpp::CalcPeaceMovement` applies `buildingutil_GetLowerPeaceMovement` to the computed peace movement value.
- Retail `buildings.txt` gives `IMPROVE_MOVIE_PALACE` `LowerPeaceMovement -0.2`.
- Runtime projection already imports this as `lower_peace_movement_percent`.

## Implementation
- City happiness now combines war-discontent and overseas deployment penalties as peace-movement centis.
- City-level `lower_peace_movement_percent` effects scale that combined penalty before integer happiness subtraction.
- Negative modifiers reduce the penalty; modifiers at or below -100% clamp the penalty to zero.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailLowerPeaceMovementReducesWarDiscontentPenalty -count=1`
- `go test ./internal/gameplay/engine -count=1`
