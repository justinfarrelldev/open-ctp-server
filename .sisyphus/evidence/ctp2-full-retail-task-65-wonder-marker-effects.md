# CTP2 Full Retail Task 65: wonder marker effects

## Source references
- Retail `Wonder.txt` contains gameplay-affecting fields such as `EliminateNukes`, `SpiesEverywhere`, `IncreaseBoatMovement`, `GlobalRadar`, `FreeSlaves`, `PreventConversion`, `GoldPerTelevision`, `StartGaiaController`, and related empire-wide markers.
- `wonder.cdb` defines these as typed wonder fields.

## Implementation
- Retail runtime projection now preserves the remaining wonder marker/value fields as safe runtime effects.
- Production validation accepts the new effect kinds so completed wonders can carry this metadata for subsequent mechanics.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/rules -count=1`

## Remaining gaps
- Many projected wonder markers still need full execution semantics (for example nuclear elimination, global radar, spy visibility, park-ranger unlocking, and Gaia/terminal-game flow).
