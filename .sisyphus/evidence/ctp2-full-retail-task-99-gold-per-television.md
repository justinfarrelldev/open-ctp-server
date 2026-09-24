# CTP2 Full Retail Task 99: gold per television wonder effect

## Source references
- Retail wonder field `GoldPerTelevision` grants gold from cities with television infrastructure.
- Retail building marker `Television` identifies eligible cities.

## Implementation
- City yield calculation now detects city-local `television` markers and owner/global `gold_per_television` effects.
- Eligible cities add the configured gold-per-television value to gold yield.
- The bonus composes with existing terrain gold, upkeep, crime, and percent effects.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailWonderGoldPerTelevisionUsesCityTelevision -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Full retail television/televangelist/religion interactions and international broadcasting effects remain to be implemented.
