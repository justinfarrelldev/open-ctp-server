# CTP2 Full Retail Tasks 324-327: conversion gold income

## Source notes
- Retail `Const.txt` defines `CLERIC_CONVERSION_FACTOR` and `TELEVANGELIST_CONVERSION_FACTOR`.
- Retail `CityData` computes converted-city gold from trade using the cleric/televangelist factor, subtracts it from the city trade, and adds it to the converted-to player.

## Implementation
- Imported conversion factors into runtime constants as percent values.
- Treasury advancement now siphons converted-city gold to the converted-to nation before the owner receives taxable/science-split gold.
- Cleric conversions default to 20%; televangelist conversions default to 40%, with runtime constants overriding both.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -run TestRetailConversionSiphonsGoldToConvertedOwner -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/rules -count=1`
- `go test ./internal/gameplay/visibility -count=1`
